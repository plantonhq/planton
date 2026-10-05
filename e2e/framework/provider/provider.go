// Package provider defines the harness interface for E2E test providers.
// Each cloud provider implements this interface to manage test infrastructure
// lifecycle and resource verification.
package provider

import (
	"context"
	"testing"
)

// ManifestPathKey is the context key used to pass the manifest path to provider harnesses
// so they can dynamically parse resource names and namespaces for verification.
type ManifestPathKey struct{}

// Harness manages the lifecycle of test infrastructure for a specific provider.
// For Kubernetes this means a kind cluster; for AWS it means credential validation
// and resource cleanup; for GCP it means project-scoped verification, etc.
type Harness interface {
	// Setup creates or validates the provider's test infrastructure.
	// For Kubernetes, this creates a kind cluster.
	// For cloud providers, this validates credentials and connectivity.
	Setup(ctx context.Context) error

	// Teardown destroys the provider's test infrastructure.
	// For Kubernetes, this deletes the kind cluster.
	Teardown(ctx context.Context) error

	// VerifyDeployed checks that resources created by a kind are present and healthy.
	VerifyDeployed(ctx context.Context, kindDir string, outputs map[string]interface{}) error

	// VerifyDestroyed confirms that resources have been removed after destroy.
	VerifyDestroyed(ctx context.Context, kindDir string) error
}

// KindTestContext holds runtime information passed between test phases.
type KindTestContext struct {
	// Kind is the lowercase kind name (e.g., "kubernetesnamespace").
	Kind string

	// Provider is the provider name (e.g., "kubernetes", "aws").
	Provider string

	// Engine is the IaC engine ("pulumi" or "terraform").
	Engine string

	// ModuleDir is the absolute path to the kind's IaC module directory.
	ModuleDir string

	// ManifestPath is the absolute path to the kind's hack/manifest.yaml.
	ManifestPath string

	// StackName is the unique Pulumi stack name for this test run.
	StackName string

	// BackendURL is the Pulumi backend URL (file-based for E2E).
	BackendURL string

	// IacInputFilePath is the path to the generated iac-input YAML.
	IacInputFilePath string

	// Outputs holds raw outputs after deployment (map[string]interface{}).
	Outputs map[string]interface{}

	// FlatOutputs holds the flattened string-keyed outputs after outputs.Flatten().
	// Populated during the VERIFY-OUT phase.
	FlatOutputs map[string]string

	// TransformedOutputs holds the typed Outputs proto after outputs.Transform().
	// Stored as interface{} to avoid importing proto in this package.
	// The runner package type-asserts to proto.Message when needed.
	TransformedOutputs interface{}

	// RepoRoot is the absolute path to the planton repository root. Passed through
	// to test helpers that need to locate checked-in assets (e.g. a function
	// source tree under catalog/.../e2e/fixtures/) or resolve module paths.
	RepoRoot string

	// RunID is the unique test run identifier, used for stack naming.
	RunID string

	// T is the Go test handle, required by Terratest for logging.
	// Populated from the test function's *testing.T.
	T testing.TB

	// TerraformOpts holds the Terratest terraform.Options configured during
	// VALIDATE phase. Stored as interface{} to avoid importing Terratest in the
	// provider package; the runner package type-asserts to *terraform.Options.
	TerraformOpts interface{}

	// IdentityProviderConfig, when set, is the provider configuration file the
	// lane deploys with in place of the kind's fixture: the identity an
	// IdentityProvisioner created for a scenario that declares
	// planton.dev/e2e-identity. Bound by the runner on every manifest binding
	// (the first deploy and the lifecycle lanes' second acts alike).
	IdentityProviderConfig string

	// TerraformWorkDir is the temp directory containing the TF module copy.
	// Cleaned up after the test completes.
	TerraformWorkDir string

	// TerraformCleanup removes the temporary TF working directory.
	TerraformCleanup func()

	// AssertApplyIdempotency adds an IDEMPOTENCY phase after DEPLOY that
	// re-plans the just-applied configuration and fails on any pending
	// change. Populated by the provider test entrypoint from the provider
	// E2E profile's assert_apply_idempotency field; applies only to the
	// kind under test, never to prerequisite fixtures.
	AssertApplyIdempotency bool
}
