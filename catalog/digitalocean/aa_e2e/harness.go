// Package aa_e2e implements the E2E provider harness for DigitalOcean. Like
// AWS (and unlike Kubernetes' local kind cluster), DigitalOcean is a real
// cloud account: Setup validates that the ambient API token can reach the
// account, and resource verification runs through the DigitalOcean REST API
// via godo.
//
// Credentials are intentionally NOT plumbed through the stack input. The E2E
// framework builds every stack input with a nil provider config, so both IaC
// engines resolve credentials from the environment -- DIGITALOCEAN_TOKEN for
// the API (the exact variable the terraform provider and the pulumi bridge
// read), plus SPACES_ACCESS_KEY_ID / SPACES_SECRET_ACCESS_KEY for Spaces
// bucket lanes (Spaces is an S3-compatible credential plane the API token
// cannot reach). The harness reads the same variables, so one credential set
// feeds deploys, destroys, and verification.
package aa_e2e

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/digitalocean/godo"
	"github.com/pkg/errors"
	"github.com/plantonhq/planton/catalog/digitalocean/aa_e2e/verify"
	"github.com/plantonhq/planton/e2e/framework/provider"
)

// Harness manages the DigitalOcean E2E test lifecycle.
type Harness struct {
	client *godo.Client

	// mu guards deployed, written by VerifyDeployed and read by VerifyDestroyed.
	mu       sync.Mutex
	deployed map[string]deployedResource
}

// deployedResource records what VerifyDeployed observed so VerifyDestroyed can
// re-probe the same resource after the DESTROY phase, when stack outputs are
// no longer available.
type deployedResource struct {
	id      string
	outputs map[string]interface{}
}

// NewHarness creates a DigitalOcean test harness. Credentials come from the
// environment (see the package doc); none are passed here.
func NewHarness() *Harness {
	return &Harness{deployed: make(map[string]deployedResource)}
}

// Setup builds a godo client from the ambient API token and confirms it
// resolves to a usable account via GET /v2/account (read-only,
// side-effect-free). DIGITALOCEAN_TOKEN is the canonical variable;
// DIGITALOCEAN_ACCESS_TOKEN is accepted as the provider's documented
// alternate. Spaces keys are deliberately NOT validated here -- only bucket
// lanes need them, and their absence must not block every other kind's lane.
func (h *Harness) Setup(ctx context.Context) error {
	token := firstNonEmpty(os.Getenv("DIGITALOCEAN_TOKEN"), os.Getenv("DIGITALOCEAN_ACCESS_TOKEN"))
	if token == "" {
		return errors.New("no DigitalOcean credential in the environment: set DIGITALOCEAN_TOKEN " +
			"(a scoped API token from the DigitalOcean control panel)")
	}

	client := godo.NewFromToken(token)
	account, _, err := client.Account.Get(ctx)
	if err != nil {
		return errors.Wrap(err, "DigitalOcean credential validation failed (GET /v2/account); "+
			"the token is invalid, expired, or lacks read scope")
	}

	fmt.Printf("  [digitalocean] authenticated as %s (uuid %s, status %s)\n",
		account.Email, account.UUID, account.Status)

	h.client = client
	return nil
}

// Teardown is a no-op. Each scenario destroys its own resources in the DESTROY
// phase and confirms removal in VERIFY-CLN; cross-run orphans are reclaimed by
// scheduled cleanup, not here.
func (h *Harness) Teardown(ctx context.Context) error {
	return nil
}

// VerifyDeployed confirms the component's resource exists via its registered
// verifier, using the resource id carried in the stack outputs.
func (h *Harness) VerifyDeployed(ctx context.Context, component string, outputs map[string]interface{}) error {
	v, err := verify.GetVerifier(component)
	if err != nil {
		return err
	}

	id := verify.StringOutput(outputs, v.IDOutputKey())
	if id == "" {
		return errors.Errorf("no %q in outputs for %s -- cannot verify", v.IDOutputKey(), component)
	}

	h.mu.Lock()
	h.deployed[componentKey(ctx, component)] = deployedResource{id: id, outputs: outputs}
	h.mu.Unlock()

	if ov, ok := v.(verify.OutputsVerifier); ok {
		return ov.VerifyExistsFromOutputs(ctx, h.client, outputs)
	}
	return v.VerifyExists(ctx, h.client, id)
}

// absentPollTimeout bounds how long VerifyDestroyed keeps re-probing a
// resource the API still answers for after a successful destroy, and
// absentPollInterval spaces the probes. DigitalOcean's read-after-delete is
// eventually consistent: the IaC destroy returns once the DELETE is accepted,
// and a GET issued within a second or two can still answer 200 (measured on
// volumes -- destroy done in ~2s, the probe 1s later saw the volume, 404 a
// few seconds after). Sixty seconds is far beyond the observed lag for the
// fast classes and short enough that a genuinely leaked resource still fails
// the lane promptly; the slow classes (clusters, databases) delete
// synchronously in the provider and never reach this poll.
const (
	absentPollTimeout  = 60 * time.Second
	absentPollInterval = 2 * time.Second
)

// VerifyDestroyed confirms the previously deployed resource no longer exists.
//
// Only a StillExistsError is retried (the resource is simply not gone YET);
// every other error -- credentials, rate limits, a broken lookup -- fails the
// phase on the first probe, so an API problem can never be polled into a
// timeout that reads like a leak.
func (h *Harness) VerifyDestroyed(ctx context.Context, component string) error {
	v, err := verify.GetVerifier(component)
	if err != nil {
		return err
	}

	h.mu.Lock()
	res := h.deployed[componentKey(ctx, component)]
	h.mu.Unlock()

	if res.id == "" && res.outputs == nil {
		return errors.Errorf("no stored resource id for %s -- VerifyDeployed may not have run", component)
	}

	probe := func() error {
		if ov, ok := v.(verify.OutputsVerifier); ok {
			return ov.VerifyAbsentFromOutputs(ctx, h.client, res.outputs)
		}
		return v.VerifyAbsent(ctx, h.client, res.id)
	}

	deadline := time.Now().Add(absentPollTimeout)
	for attempt := 1; ; attempt++ {
		err := probe()
		if err == nil {
			if attempt > 1 {
				fmt.Printf("  [verify] %s absent after %d probes (DigitalOcean read-after-delete lag)\n", component, attempt)
			}
			return nil
		}
		if !verify.IsStillExists(err) || time.Now().After(deadline) {
			return err
		}
		select {
		case <-time.After(absentPollInterval):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// componentKey combines the manifest path (from context) with the component name
// so concurrent scenarios of the same component type do not collide in the map.
func componentKey(ctx context.Context, component string) string {
	if mp, ok := ctx.Value(provider.ManifestPathKey{}).(string); ok && mp != "" {
		return mp + "::" + component
	}
	return component
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
