package iacrunner

import (
	"github.com/plantonhq/planton/pkg/iac/provisioner"
	"github.com/plantonhq/planton/pkg/iac/stackinput/providerdetect"
	"github.com/plantonhq/planton/pkg/iac/stackinput/stackinputproviderconfig"
	"google.golang.org/protobuf/proto"
)

// Context holds all resolved inputs needed for IaC command execution.
type Context struct {
	// ManifestPath is the path to the resolved manifest file
	ManifestPath string

	// ManifestObject is the loaded and validated manifest proto message
	ManifestObject proto.Message

	// StackInputFilePath is the original stack input file path (if provided via --stack-input)
	StackInputFilePath string

	// ProviderConfig contains the unified provider configuration
	ProviderConfig *stackinputproviderconfig.ProviderConfig

	// DetectionResult contains the provider detection result from the manifest
	DetectionResult *providerdetect.DetectionResult

	// ProvisionerType indicates which IaC provisioner to use (Pulumi, Tofu, Terraform)
	ProvisionerType provisioner.ProvisionerType

	// KubeContext is the kubectl context to use for Kubernetes deployments
	KubeContext string

	// ModuleDir is the directory containing the provisioner module
	ModuleDir string

	// ValueOverrides contains key=value pairs for manifest field overrides
	ValueOverrides map[string]string

	// ModuleVersion is the specific version to checkout for IaC modules
	ModuleVersion string

	// NoCleanup indicates whether to keep workspace copy after execution
	NoCleanup bool

	// ShowDiff indicates whether to show detailed resource diffs
	ShowDiff bool

	// CleanupFuncs contains functions to run after execution for cleanup
	CleanupFuncs []func()
}

// Cleanup runs all registered cleanup functions.
func (c *Context) Cleanup() {
	for _, fn := range c.CleanupFuncs {
		fn()
	}
}

// AddCleanupFunc registers a cleanup function to run after execution.
func (c *Context) AddCleanupFunc(fn func()) {
	c.CleanupFuncs = append(c.CleanupFuncs, fn)
}
