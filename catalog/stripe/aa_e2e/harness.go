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

	// mu guards deployedIDs, written by VerifyDeployed and read by VerifyDestroyed.
	mu          sync.Mutex
	deployedIDs map[string]string // manifest path + component -> Stripe id
}

// NewHarness creates a Stripe harness; credentials are read in Setup.
func NewHarness() *Harness {
	return &Harness{deployedIDs: make(map[string]string)}
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
// stores the id for VerifyDestroyed.
func (h *Harness) VerifyDeployed(ctx context.Context, component string, outputs map[string]interface{}) error {
	v, err := verify.GetVerifier(component)
	if err != nil {
		return err
	}
	id, _ := outputs["id"].(string)
	if id == "" {
		return errors.Errorf("no id found in the outputs of %s", component)
	}

	h.mu.Lock()
	h.deployedIDs[componentKey(ctx, component)] = id
	h.mu.Unlock()

	return v.VerifyExists(h.client, id)
}

// VerifyDestroyed checks what destroy left, which differs by kind: a deleted endpoint is gone,
// while a deactivated configuration is still there, inactive.
func (h *Harness) VerifyDestroyed(ctx context.Context, component string) error {
	v, err := verify.GetVerifier(component)
	if err != nil {
		return err
	}
	h.mu.Lock()
	id, stored := h.deployedIDs[componentKey(ctx, component)]
	h.mu.Unlock()
	if !stored {
		return errors.Errorf("no stored id for %s -- VerifyDeployed may not have run", component)
	}
	return v.VerifyDestroyed(h.client, id)
}

// componentKey combines the manifest path from the context with the component, so scenarios of
// one kind running side by side never share an id.
func componentKey(ctx context.Context, component string) string {
	if mp, ok := ctx.Value(provider.ManifestPathKey{}).(string); ok && mp != "" {
		return mp + "::" + component
	}
	return component
}
