// Package aa_e2e implements the E2E provider harness for Cloudflare, using
// the Cloudflare v4 REST API for resource verification. Cloudflare is a SaaS
// control plane -- Setup validates credentials and Teardown is a no-op; both
// IaC engines authenticate through the same ambient CLOUDFLARE_API_TOKEN the
// harness requires, so a lane that passes Setup deploys with the credentials
// verification will read with.
package aa_e2e

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/pkg/errors"
	"github.com/plantonhq/planton/catalog/cloudflare/aa_e2e/verify"
	"github.com/plantonhq/planton/e2e/framework/provider"
)

// Environment variables the harness reads at Setup. The PLANTON_E2E_ prefix
// on the account ID makes it referenceable from committed manifests via the
// ${E2E_ENV:...} token (the runner only expands that prefix); the API token
// deliberately has no such alias -- credentials never belong in a manifest.
const (
	// EnvAPIToken is the scoped API token both engines and the harness use.
	EnvAPIToken = "CLOUDFLARE_API_TOKEN"
	// EnvAccountID is the Cloudflare account under test. Account-scoped
	// manifests reference it as ${E2E_ENV:PLANTON_E2E_CLOUDFLARE_ACCOUNT_ID}.
	EnvAccountID = "PLANTON_E2E_CLOUDFLARE_ACCOUNT_ID"
)

// Harness manages the Cloudflare E2E test lifecycle.
type Harness struct {
	client *Client

	// mu guards deployedOutputs, written by VerifyDeployed and read by
	// VerifyDestroyed.
	mu sync.Mutex
	// deployedOutputs stores each kind's full string-ified
	// outputs rather than a single ID: Cloudflare identities are compound
	// (zone_id + the resource's own id), and VerifyDestroyed receives no
	// outputs of its own.
	deployedOutputs map[string]map[string]string
}

// NewHarness creates a Cloudflare test harness. Credentials are read from
// the environment at Setup (EnvAPIToken, EnvAccountID).
func NewHarness() *Harness {
	return &Harness{
		deployedOutputs: make(map[string]map[string]string),
	}
}

// Setup validates credentials against the live API: the token through
// Cloudflare's side-effect-free verify endpoint, then visibility of the
// configured account. No infrastructure is created.
func (h *Harness) Setup(ctx context.Context) error {
	apiToken := os.Getenv(EnvAPIToken)
	accountID := os.Getenv(EnvAccountID)
	if apiToken == "" || accountID == "" {
		return errors.Errorf("%s and %s must be set", EnvAPIToken, EnvAccountID)
	}

	fmt.Printf("  [cloudflare] Verifying API token and account %s...\n", accountID)

	client := NewClient(apiToken, accountID)
	if err := client.VerifyConnectivity(ctx); err != nil {
		return errors.Wrap(err, "Cloudflare API connectivity check failed")
	}

	h.client = client
	fmt.Printf("  [cloudflare] Token active, account %s reachable\n", accountID)
	return nil
}

// Teardown is a no-op for Cloudflare (no test infrastructure to destroy;
// resource cleanup is the DESTROY phase's job, verified per kind).
func (h *Harness) Teardown(ctx context.Context) error {
	return nil
}

// VerifyDeployed checks that the deployed Cloudflare resource exists via the
// REST API. The full output map is stored for VerifyDestroyed.
func (h *Harness) VerifyDeployed(ctx context.Context, kindDir string, outputs map[string]interface{}) error {
	v, err := verify.GetVerifier(kindDir)
	if err != nil {
		return err
	}

	stringOutputs := stringifyOutputs(outputs)
	if stringOutputs[v.IDOutputKey()] == "" {
		return errors.Errorf("no %q found in outputs for %s", v.IDOutputKey(), kindDir)
	}

	h.mu.Lock()
	h.deployedOutputs[kindKey(ctx, kindDir)] = stringOutputs
	h.mu.Unlock()

	return v.VerifyExists(ctx, h.client, stringOutputs)
}

// VerifyDestroyed confirms that the previously deployed Cloudflare resource
// no longer exists.
func (h *Harness) VerifyDestroyed(ctx context.Context, kindDir string) error {
	v, err := verify.GetVerifier(kindDir)
	if err != nil {
		return err
	}

	h.mu.Lock()
	stringOutputs := h.deployedOutputs[kindKey(ctx, kindDir)]
	h.mu.Unlock()

	if stringOutputs == nil {
		return errors.Errorf("no stored outputs for %s -- VerifyDeployed may not have run", kindDir)
	}

	return v.VerifyAbsent(ctx, h.client, stringOutputs)
}

// stringifyOutputs flattens raw outputs to strings, which is what the
// path-template verifiers consume.
func stringifyOutputs(outputs map[string]interface{}) map[string]string {
	result := make(map[string]string, len(outputs))
	for key, value := range outputs {
		switch v := value.(type) {
		case string:
			result[key] = v
		case nil:
			result[key] = ""
		default:
			result[key] = fmt.Sprintf("%v", v)
		}
	}
	return result
}

// kindKey creates a unique lookup key combining the manifest path (from
// context) and kind name, so concurrent tests for the same kind
// type don't collide.
func kindKey(ctx context.Context, kindDir string) string {
	if mp, ok := ctx.Value(provider.ManifestPathKey{}).(string); ok && mp != "" {
		return mp + "::" + kindDir
	}
	return kindDir
}
