// Package aa_e2e implements the E2E provider harness for Auth0,
// using the Auth0 Management API for resource verification.
package aa_e2e

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/pkg/errors"
	"github.com/plantonhq/openmcf/apis/org/openmcf/provider/auth0/aa_e2e/verify"
	"github.com/plantonhq/openmcf/e2e/framework/provider"
)

// Harness manages Auth0 E2E test lifecycle.
// Unlike Kubernetes (kind cluster), Auth0 is a SaaS API -- Setup validates
// credentials and Teardown is a no-op.
type Harness struct {
	client *ManagementClient

	// mu guards deployedIDs which is written by VerifyDeployed
	// and read by VerifyDestroyed.
	mu          sync.Mutex
	deployedIDs map[string]string // component -> resource ID
}

// NewHarness creates an Auth0 test harness.
// Credentials are read from AUTH0_DOMAIN, AUTH0_CLIENT_ID, AUTH0_CLIENT_SECRET env vars.
func NewHarness() *Harness {
	return &Harness{
		deployedIDs: make(map[string]string),
	}
}

// Setup authenticates with the Auth0 Management API and verifies connectivity.
func (h *Harness) Setup(ctx context.Context) error {
	domain := os.Getenv("AUTH0_DOMAIN")
	clientID := os.Getenv("AUTH0_CLIENT_ID")
	clientSecret := os.Getenv("AUTH0_CLIENT_SECRET")

	if domain == "" || clientID == "" || clientSecret == "" {
		return errors.New("AUTH0_DOMAIN, AUTH0_CLIENT_ID, and AUTH0_CLIENT_SECRET must be set")
	}

	fmt.Printf("  [auth0] Authenticating with tenant %s...\n", domain)

	client, err := NewManagementClient(domain, clientID, clientSecret)
	if err != nil {
		return errors.Wrap(err, "failed to create Auth0 Management API client")
	}

	if err := client.VerifyConnectivity(); err != nil {
		return errors.Wrap(err, "Auth0 Management API connectivity check failed")
	}

	h.client = client
	fmt.Printf("  [auth0] Authenticated and connected to %s\n", domain)
	return nil
}

// Teardown is a no-op for Auth0 (no infrastructure to destroy).
func (h *Harness) Teardown(ctx context.Context) error {
	return nil
}

// VerifyDeployed checks that the deployed Auth0 resource exists via the Management API.
// The resource ID is extracted from stack outputs and stored for VerifyDestroyed.
func (h *Harness) VerifyDeployed(ctx context.Context, component string, outputs map[string]interface{}) error {
	v, err := verify.GetVerifier(component)
	if err != nil {
		return err
	}

	id := extractResourceID(outputs)
	if id == "" {
		return errors.Errorf("no resource ID found in outputs for %s", component)
	}

	// Store for VerifyDestroyed
	h.mu.Lock()
	key := componentKey(ctx, component)
	h.deployedIDs[key] = id
	h.mu.Unlock()

	return v.VerifyExists(h.client, id)
}

// VerifyDestroyed confirms that the previously deployed Auth0 resource no longer exists.
func (h *Harness) VerifyDestroyed(ctx context.Context, component string) error {
	v, err := verify.GetVerifier(component)
	if err != nil {
		return err
	}

	h.mu.Lock()
	key := componentKey(ctx, component)
	id := h.deployedIDs[key]
	h.mu.Unlock()

	if id == "" {
		return errors.Errorf("no stored resource ID for %s -- VerifyDeployed may not have run", component)
	}

	return v.VerifyAbsent(h.client, id)
}

// extractResourceID pulls the Auth0 resource ID from stack outputs.
// Pulumi exports "id" for all Auth0 components; Terraform outputs vary
// but all include an "id" key.
func extractResourceID(outputs map[string]interface{}) string {
	if outputs == nil {
		return ""
	}
	if id, ok := outputs["id"]; ok {
		switch v := id.(type) {
		case string:
			return v
		default:
			return fmt.Sprintf("%v", v)
		}
	}
	return ""
}

// componentKey creates a unique lookup key combining the manifest path (from context)
// and component name, so concurrent tests for the same component type don't collide.
func componentKey(ctx context.Context, component string) string {
	if mp, ok := ctx.Value(provider.ManifestPathKey{}).(string); ok && mp != "" {
		return mp + "::" + component
	}
	return component
}
