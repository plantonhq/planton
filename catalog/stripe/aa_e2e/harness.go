// Package aa_e2e implements the E2E provider harness for Stripe, using Stripe's REST API to
// verify what the modules created, changed and left behind.
package aa_e2e

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/pkg/errors"
	"github.com/plantonhq/planton/catalog/stripe/aa_e2e/verify"
	"github.com/plantonhq/planton/e2e/framework/provider"
	"github.com/plantonhq/planton/pkg/iac/provider/stripe/stripekey"
)

const (
	// envAPIKey is the dedicated test sandbox's key, the same variable the provider reads.
	envAPIKey = "STRIPE_API_KEY"
	// envAccount optionally names a Connect account to act on, as the provider reads it.
	envAccount = "STRIPE_ACCOUNT"
)

// Harness manages the Stripe E2E lifecycle. Stripe is a SaaS API: Setup proves the key belongs
// to a test sandbox and authenticates, and Teardown has nothing to remove.
//
// The key check is the harness's own. A lane passes the modules no provider config -- OpenTofu
// reads STRIPE_API_KEY from the environment -- so the credential loader's mode guard never runs
// here, and a live key would otherwise apply every scenario to a live account.
type Harness struct {
	client *Client

	// mu guards deployedIDs and deployedChildren, written by VerifyDeployed and read by
	// VerifyDestroyed.
	mu               sync.Mutex
	deployedIDs      map[string]string            // manifest path + component -> Stripe id
	deployedChildren map[string]map[string]string // manifest path + component -> child key -> Stripe id
}

// NewHarness creates a Stripe harness; credentials are read in Setup.
func NewHarness() *Harness {
	return &Harness{deployedIDs: make(map[string]string), deployedChildren: make(map[string]map[string]string)}
}

// Setup refuses any key that is not a test-mode key, before the first API call, then proves the
// key authenticates. A refusal fails TestMain, so no lane runs.
func (h *Harness) Setup(ctx context.Context) error {
	apiKey := os.Getenv(envAPIKey)
	if apiKey == "" {
		return errors.Errorf("%s must be set to the dedicated test sandbox's key", envAPIKey)
	}
	if err := stripekey.RequireTestMode(apiKey); err != nil {
		return err
	}

	client := NewClient(apiKey, os.Getenv(envAccount))
	if err := client.VerifyConnectivity(); err != nil {
		return err
	}
	h.client = client
	fmt.Println("  [stripe] Authenticated with the test sandbox")
	return nil
}

// Teardown is a no-op: the harness creates nothing of its own.
func (h *Harness) Teardown(ctx context.Context) error {
	return nil
}

// VerifyDeployed checks the object the module created, reading its id from the outputs, and
// stores the id for VerifyDestroyed. A kind that folds children (a meter's alerts, a product's
// feature links, a Radar list's items) reports their ids as a map output, and each child is
// checked and stored too.
func (h *Harness) VerifyDeployed(ctx context.Context, component string, outputs map[string]interface{}) error {
	v, err := verify.GetVerifier(component)
	if err != nil {
		return err
	}
	id, _ := outputs["id"].(string)
	if id == "" {
		return errors.Errorf("no id found in the outputs of %s", component)
	}
	var children map[string]string
	cv, folds := v.(verify.ChildVerifier)
	if folds {
		if children, err = childIDs(outputs[cv.ChildOutput()]); err != nil {
			return errors.Wrapf(err, "%s: output %s", component, cv.ChildOutput())
		}
	}

	key := componentKey(ctx, component)
	h.mu.Lock()
	h.deployedIDs[key] = id
	h.deployedChildren[key] = children
	h.mu.Unlock()

	if err := v.VerifyExists(h.client, id); err != nil {
		return err
	}
	if folds {
		return cv.VerifyChildrenExist(h.client, id, children)
	}
	return nil
}

// VerifyDestroyed checks what destroy left, which differs by kind: a deleted endpoint is gone,
// while a deactivated configuration is still there, inactive.
func (h *Harness) VerifyDestroyed(ctx context.Context, component string) error {
	v, err := verify.GetVerifier(component)
	if err != nil {
		return err
	}
	key := componentKey(ctx, component)
	h.mu.Lock()
	id, stored := h.deployedIDs[key]
	children := h.deployedChildren[key]
	h.mu.Unlock()
	if !stored {
		return errors.Errorf("no stored id for %s -- VerifyDeployed may not have run", component)
	}
	if err := v.VerifyDestroyed(h.client, id); err != nil {
		return err
	}
	if cv, folds := v.(verify.ChildVerifier); folds {
		return cv.VerifyChildrenDestroyed(h.client, id, children)
	}
	return nil
}

// childIDs reads a map output of child ids. A kind that declared no children reports an empty
// map, or no output at all.
func childIDs(output interface{}) (map[string]string, error) {
	ids := map[string]string{}
	switch m := output.(type) {
	case nil:
	case map[string]string:
		for k, id := range m {
			ids[k] = id
		}
	case map[string]interface{}:
		for k, raw := range m {
			id, ok := raw.(string)
			if !ok || id == "" {
				return nil, errors.Errorf("child %q has no id (got %v)", k, raw)
			}
			ids[k] = id
		}
	default:
		return nil, errors.Errorf("want a map of child ids, got %T", output)
	}
	return ids, nil
}

// componentKey combines the manifest path from the context with the component, so scenarios of
// one kind running side by side never share an id.
func componentKey(ctx context.Context, component string) string {
	if mp, ok := ctx.Value(provider.ManifestPathKey{}).(string); ok && mp != "" {
		return mp + "::" + component
	}
	return component
}
