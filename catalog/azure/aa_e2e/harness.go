// Package aa_e2e implements the E2E provider harness for Azure. Like AWS (and
// unlike Kubernetes' local kind cluster), Azure is a real cloud subscription:
// Setup validates that the ambient Azure credential chain can reach the
// subscription, and resource verification runs through the Azure SDK.
//
// Credentials are intentionally NOT plumbed through the IaC input. The E2E
// framework builds every IaC input with a nil provider config, so the IaC
// modules resolve credentials from the ambient chain -- for Pulumi, the shared
// pulumiazureprovider builder falls back to the SDK default chain; for Terraform,
// the empty `provider "azurerm" { features {} }` block reads the ARM_* env vars.
// That chain is populated keylessly (an `az` CLI login locally, or a GitHub
// Actions OIDC federation in CI), so no static secret is ever stored on disk.
//
// The one Azure-specific requirement: azurerm does not infer the
// subscription, so Setup exports ARM_SUBSCRIPTION_ID (and ARM_TENANT_ID) into the
// process environment, which the framework forwards to both engines.
package aa_e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/pkg/errors"
	"github.com/plantonhq/planton/catalog/azure/aa_e2e/verify"
	"github.com/plantonhq/planton/e2e/framework/provider"
)

// armScope is the Azure Resource Manager token audience. Acquiring a token for it
// is the zero-permission credential probe (the Azure analog of AWS
// sts:GetCallerIdentity): it proves the ambient chain can mint an ARM token
// without requiring any RBAC role assignment.
const armScope = "https://management.azure.com/.default"

// Harness manages the Azure E2E test lifecycle.
type Harness struct {
	cred           *azidentity.DefaultAzureCredential
	subscriptionID string

	// mu guards deployed, written by VerifyDeployed and read by VerifyDestroyed.
	mu       sync.Mutex
	deployed map[string]deployedResource
}

// deployedResource records what VerifyDeployed observed so VerifyDestroyed can
// re-probe the same resource. Azure resource verification is subscription-scoped
// (not region-scoped like AWS), so only the identifier is retained.
type deployedResource struct {
	id string
}

// NewHarness creates an Azure test harness. Credentials come from the ambient
// chain (see the package doc); none are passed here.
func NewHarness() *Harness {
	return &Harness{deployed: make(map[string]deployedResource)}
}

// Setup resolves the target subscription and tenant, exports them so both IaC
// engines agree on which subscription to deploy into, loads the ambient credential
// chain, and confirms it can mint an ARM token (zero-permission, side-effect-free).
func (h *Harness) Setup(ctx context.Context) error {
	subscriptionID := firstNonEmpty(
		os.Getenv("E2E_AZURE_SUBSCRIPTION_ID"), os.Getenv("ARM_SUBSCRIPTION_ID"))
	tenantID := firstNonEmpty(
		os.Getenv("E2E_AZURE_TENANT_ID"), os.Getenv("ARM_TENANT_ID"))

	// Fall back to the `az` CLI login for anything not pinned via env vars.
	if subscriptionID == "" || tenantID == "" {
		acct, err := azAccountShow(ctx)
		if err != nil {
			return errors.Wrap(err, "no subscription/tenant in the environment and "+
				"`az account show` failed (locally: `az login`; in CI: the OIDC federation step)")
		}
		if subscriptionID == "" {
			subscriptionID = acct.ID
		}
		if tenantID == "" {
			tenantID = acct.TenantID
		}
	}

	if subscriptionID == "" {
		return errors.New("could not resolve an Azure subscription id " +
			"(set E2E_AZURE_SUBSCRIPTION_ID / ARM_SUBSCRIPTION_ID or run `az login`)")
	}

	// Export for both engines: azurerm requires an explicit subscription, and
	// the empty provider block reads ARM_* from the environment.
	if err := os.Setenv("ARM_SUBSCRIPTION_ID", subscriptionID); err != nil {
		return errors.Wrap(err, "failed to export ARM_SUBSCRIPTION_ID")
	}
	if tenantID != "" {
		if err := os.Setenv("ARM_TENANT_ID", tenantID); err != nil {
			return errors.Wrap(err, "failed to export ARM_TENANT_ID")
		}
	}

	// Also export the subscription under the manifest env-token prefix, so a
	// scenario can compose subscription-scoped ARM IDs (e.g. a data-plane
	// asset's full resource ID inside a fixture workspace) via
	// ${E2E_ENV:PLANTON_E2E_AZURE_SUBSCRIPTION_ID} without hardcoding one
	// account's identifier into a committed manifest.
	if err := os.Setenv("PLANTON_E2E_AZURE_SUBSCRIPTION_ID", subscriptionID); err != nil {
		return errors.Wrap(err, "failed to export PLANTON_E2E_AZURE_SUBSCRIPTION_ID")
	}

	// Skip per-apply Azure Resource Provider registration for the ephemeral test
	// run. The pulumi-azure classic engine auto-registers a broad set of resource
	// providers (Microsoft.Compute, .Storage, ...) at init, firing the
	// registrations concurrently; Azure serialises subscription-level writes and
	// returns 409 ConflictingConcurrentWriteNotAllowed on a subscription whose
	// providers are not yet registered. Registration is a one-time subscription
	// bootstrap, orthogonal to whether a module's resource is created correctly --
	// the contract E2E actually validates -- so the harness opts out for the test
	// environment. pulumi-azure classic reads this env var when the provider arg
	// is unset; azurerm itself needs nothing (its
	// `resource_provider_registrations` default is "none", so it ignores the
	// var harmlessly). Respect an explicit override if the caller already set it.
	if os.Getenv("ARM_SKIP_PROVIDER_REGISTRATION") == "" {
		if err := os.Setenv("ARM_SKIP_PROVIDER_REGISTRATION", "true"); err != nil {
			return errors.Wrap(err, "failed to export ARM_SKIP_PROVIDER_REGISTRATION")
		}
	}

	// Key Vault data-plane bootstrap: keys and certificates are data-plane
	// objects, and even a subscription Owner cannot create one without an
	// explicit data-plane grant. Ensure (idempotently) that the signed-in
	// test principal holds "Key Vault Administrator" at the subscription
	// scope -- a one-time test-subscription bootstrap in the same class as
	// the resource-provider-registration opt-out above. Best-effort: when
	// the signed-in principal cannot be resolved (e.g. a service-principal
	// login where `az ad signed-in-user` does not apply), the run proceeds
	// and any data-plane scenario fails attributably with a 403.
	ensureKeyVaultDataPlaneGrant(ctx, subscriptionID)

	// Front Door BYO-certificate bootstrap: Front Door reads Key Vault
	// with Microsoft's own service principal (the
	// "Microsoft.AzureFrontDoor-Cdn" enterprise application), which (a)
	// must exist in the tenant and (b) must hold vault read access
	// before a Front Door secret can deploy. Same one-time bootstrap
	// class and best-effort semantics as the grant above: on failure the
	// run proceeds and any BYO-certificate scenario fails attributably
	// with an access-denied error naming the vault.
	ensureFrontDoorKeyVaultGrant(ctx, subscriptionID)

	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return errors.Wrap(err, "failed to load the ambient Azure credential chain "+
			"(locally: `az login`; in CI: the GitHub OIDC federation step)")
	}

	if _, err := cred.GetToken(ctx, policy.TokenRequestOptions{Scopes: []string{armScope}}); err != nil {
		return errors.Wrap(err, "Azure credential validation failed (could not acquire an ARM token); "+
			"no usable credentials in the ambient chain")
	}

	fmt.Printf("  [azure] authenticated for subscription %s (tenant %s)\n", subscriptionID, tenantID)

	h.cred = cred
	h.subscriptionID = subscriptionID
	return nil
}

// Teardown is a no-op. Each scenario destroys its own resources in the DESTROY
// phase and confirms removal in VERIFY-CLN; cross-run orphans (tagged
// managed-by=planton-e2e) are reclaimed by the scheduled janitor in CI, not here.
func (h *Harness) Teardown(ctx context.Context) error {
	return nil
}

// VerifyDeployed confirms the kind's resource exists via its registered
// verifier, using the identifier carried in the outputs.
func (h *Harness) VerifyDeployed(ctx context.Context, kindDir string, outputs map[string]interface{}) error {
	v, err := verify.GetVerifier(kindDir)
	if err != nil {
		return err
	}

	id := stringOutput(outputs, v.IDOutputKey())
	if id == "" {
		return errors.Errorf("no %q in outputs for %s -- cannot verify", v.IDOutputKey(), kindDir)
	}

	h.mu.Lock()
	h.deployed[kindKey(ctx, kindDir)] = deployedResource{id: id}
	h.mu.Unlock()

	return v.VerifyExists(ctx, h.cred, h.subscriptionID, id)
}

// VerifyRuntimeFailureCause implements the framework's optional
// RuntimeCauseVerifier capability (expected-runtime-failure lanes): it
// dispatches to the kind's verifier, which must itself opt in via the local
// verify.RuntimeCauseVerifier interface, reusing the resource id VERIFY-RES
// stored (the phase is guaranteed to have run first).
func (h *Harness) VerifyRuntimeFailureCause(ctx context.Context, tc *provider.KindTestContext, cause string) error {
	v, err := verify.GetVerifier(tc.Kind)
	if err != nil {
		return err
	}
	rcv, ok := v.(verify.RuntimeCauseVerifier)
	if !ok {
		return errors.Errorf("kind %q's verifier does not implement verify.RuntimeCauseVerifier -- the scenario expects a runtime failure cause (%s) it cannot pin", tc.Kind, cause)
	}

	h.mu.Lock()
	res := h.deployed[kindKey(ctx, tc.Kind)]
	h.mu.Unlock()
	if res.id == "" {
		return errors.Errorf("no stored resource id for %s -- VERIFY-RES may not have run", tc.Kind)
	}
	return rcv.VerifyRuntimeFailureCause(ctx, h.cred, h.subscriptionID, res.id, cause)
}

// VerifyDestroyed confirms the previously deployed resource no longer exists.
func (h *Harness) VerifyDestroyed(ctx context.Context, kindDir string) error {
	v, err := verify.GetVerifier(kindDir)
	if err != nil {
		return err
	}

	h.mu.Lock()
	res := h.deployed[kindKey(ctx, kindDir)]
	h.mu.Unlock()

	if res.id == "" {
		return errors.Errorf("no stored resource id for %s -- VerifyDeployed may not have run", kindDir)
	}
	return v.VerifyAbsent(ctx, h.cred, h.subscriptionID, res.id)
}

// keyVaultDataPlaneRole is the built-in role granting full Key Vault
// data-plane access (keys, secrets, certificates) on RBAC-mode vaults.
const keyVaultDataPlaneRole = "Key Vault Administrator"

// ensureKeyVaultDataPlaneGrant checks that the signed-in principal holds the
// Key Vault data-plane role at the subscription scope and creates the grant
// when missing. Subscription-scoped (rather than per-run, per-vault) is
// deliberate: RBAC data-plane grants take minutes to propagate, so a
// per-ephemeral-vault grant would make every key/certificate scenario race
// its own authorization -- the standing grant propagates once and every
// fresh vault the tests create is covered immediately.
func ensureKeyVaultDataPlaneGrant(ctx context.Context, subscriptionID string) {
	oidOut, err := exec.CommandContext(ctx, "az", "ad", "signed-in-user", "show",
		"--query", "id", "--output", "tsv").Output()
	if err != nil {
		fmt.Printf("  [azure] note: could not resolve the signed-in principal "+
			"(non-user login?); skipping the Key Vault data-plane grant check: %v\n", err)
		return
	}
	objectID := strings.TrimSpace(string(oidOut))
	if objectID == "" {
		return
	}

	scope := "/subscriptions/" + subscriptionID
	listOut, err := exec.CommandContext(ctx, "az", "role", "assignment", "list",
		"--assignee", objectID, "--role", keyVaultDataPlaneRole, "--scope", scope,
		"--query", "[].id", "--output", "tsv").Output()
	if err == nil && strings.TrimSpace(string(listOut)) != "" {
		return // grant already in place
	}

	if out, err := exec.CommandContext(ctx, "az", "role", "assignment", "create",
		"--assignee-object-id", objectID, "--assignee-principal-type", "User",
		"--role", keyVaultDataPlaneRole, "--scope", scope,
		"--output", "none").CombinedOutput(); err != nil {
		fmt.Printf("  [azure] note: could not create the Key Vault data-plane grant "+
			"(data-plane scenarios may 403): %v: %s\n", err, strings.TrimSpace(string(out)))
		return
	}
	fmt.Printf("  [azure] granted %q to the test principal at %s "+
		"(one-time bootstrap; data-plane propagation may take a minute)\n",
		keyVaultDataPlaneRole, scope)
}

// frontDoorAppId is the well-known application id of Azure Front Door's
// service principal (the "Microsoft.AzureFrontDoor-Cdn" enterprise
// application) -- the identity Front Door uses to READ customer Key
// Vaults for bring-your-own TLS certificates. The id is a Microsoft
// constant, identical in every tenant.
const frontDoorAppId = "205478c0-bd83-4e1b-a9d6-db63a3e1e1c8"

// frontDoorKeyVaultRole is the built-in role granting read access to
// vault secret material -- what Front Door needs to fetch a wrapped
// certificate (certificates' key material is read through the secret
// face) on RBAC-mode vaults.
const frontDoorKeyVaultRole = "Key Vault Secrets User"

// ensureFrontDoorKeyVaultGrant ensures Azure Front Door's own service
// principal exists in the tenant (first use of Front Door BYO
// certificates in a tenant requires instantiating it) and holds vault
// read access at the subscription scope. Subscription-scoped for the
// same reason as the test principal's Key Vault grant: data-plane RBAC
// propagates in minutes, so per-ephemeral-vault grants would race
// authorization on every scenario.
func ensureFrontDoorKeyVaultGrant(ctx context.Context, subscriptionID string) {
	// Resolve (or instantiate) the Front Door service principal. `az ad sp
	// create` fails when the principal already exists, so resolve first.
	oidOut, err := exec.CommandContext(ctx, "az", "ad", "sp", "show",
		"--id", frontDoorAppId, "--query", "id", "--output", "tsv").Output()
	objectID := strings.TrimSpace(string(oidOut))
	if err != nil || objectID == "" {
		if out, createErr := exec.CommandContext(ctx, "az", "ad", "sp", "create",
			"--id", frontDoorAppId, "--output", "none").CombinedOutput(); createErr != nil {
			fmt.Printf("  [azure] note: could not instantiate the Front Door service principal "+
				"(BYO-certificate scenarios may fail): %v: %s\n", createErr, strings.TrimSpace(string(out)))
			return
		}
		oidOut, err = exec.CommandContext(ctx, "az", "ad", "sp", "show",
			"--id", frontDoorAppId, "--query", "id", "--output", "tsv").Output()
		if err != nil {
			fmt.Printf("  [azure] note: could not resolve the Front Door service principal "+
				"after creating it (BYO-certificate scenarios may fail): %v\n", err)
			return
		}
		objectID = strings.TrimSpace(string(oidOut))
	}
	if objectID == "" {
		return
	}

	scope := "/subscriptions/" + subscriptionID
	listOut, err := exec.CommandContext(ctx, "az", "role", "assignment", "list",
		"--assignee", objectID, "--role", frontDoorKeyVaultRole, "--scope", scope,
		"--query", "[].id", "--output", "tsv").Output()
	if err == nil && strings.TrimSpace(string(listOut)) != "" {
		return // grant already in place
	}

	if out, err := exec.CommandContext(ctx, "az", "role", "assignment", "create",
		"--assignee-object-id", objectID, "--assignee-principal-type", "ServicePrincipal",
		"--role", frontDoorKeyVaultRole, "--scope", scope,
		"--output", "none").CombinedOutput(); err != nil {
		fmt.Printf("  [azure] note: could not grant the Front Door service principal "+
			"vault read access (BYO-certificate scenarios may fail): %v: %s\n",
			err, strings.TrimSpace(string(out)))
		return
	}
	fmt.Printf("  [azure] granted %q to the Front Door service principal at %s "+
		"(one-time bootstrap; data-plane propagation may take a minute)\n",
		frontDoorKeyVaultRole, scope)
}

// azAccount is the subset of `az account show` output the harness consumes.
type azAccount struct {
	ID       string `json:"id"`
	TenantID string `json:"tenantId"`
}

// azAccountShow reads the active subscription/tenant from the local Azure CLI.
func azAccountShow(ctx context.Context) (*azAccount, error) {
	out, err := exec.CommandContext(ctx, "az", "account", "show", "--output", "json").Output()
	if err != nil {
		return nil, errors.Wrap(err, "`az account show` failed")
	}
	acct := &azAccount{}
	if err := json.Unmarshal(out, acct); err != nil {
		return nil, errors.Wrap(err, "failed to parse `az account show` output")
	}
	return acct, nil
}

// stringOutput reads a string-valued output, tolerating non-string scalars.
func stringOutput(outputs map[string]interface{}, key string) string {
	if outputs == nil {
		return ""
	}
	if v, ok := outputs[key]; ok {
		if s, isStr := v.(string); isStr {
			return s
		}
		return fmt.Sprintf("%v", v)
	}
	return ""
}

// kindKey combines the manifest path (from context) with the kind name
// so concurrent scenarios of the same kind do not collide in the map.
func kindKey(ctx context.Context, kindDir string) string {
	if mp, ok := ctx.Value(provider.ManifestPathKey{}).(string); ok && mp != "" {
		return mp + "::" + kindDir
	}
	return kindDir
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}
