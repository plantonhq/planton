// Package provider defines the harness interface for E2E test providers.
// Each cloud provider implements this interface to manage test infrastructure
// lifecycle and resource verification.
package provider

import "context"

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

	// VerifyDeployed checks that resources created by a component are present and healthy.
	VerifyDeployed(ctx context.Context, component string, outputs map[string]interface{}) error

	// VerifyDestroyed confirms that resources have been removed after destroy.
	VerifyDestroyed(ctx context.Context, component string) error
}

// ComponentTestContext holds runtime information passed between test phases.
type ComponentTestContext struct {
	// Component is the lowercase component name (e.g., "kubernetesnamespace").
	Component string

	// Provider is the provider name (e.g., "kubernetes", "aws").
	Provider string

	// Engine is the IaC engine ("pulumi" or "terraform").
	Engine string

	// ModuleDir is the absolute path to the component's IaC module directory.
	ModuleDir string

	// ManifestPath is the absolute path to the component's hack/manifest.yaml.
	ManifestPath string

	// StackName is the unique Pulumi stack name for this test run.
	StackName string

	// BackendURL is the Pulumi backend URL (file-based for E2E).
	BackendURL string

	// StackInputFilePath is the path to the generated stack-input YAML.
	StackInputFilePath string

	// Outputs holds stack outputs after deployment.
	Outputs map[string]interface{}
}
