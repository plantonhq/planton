package module

import (
	scalewayprovider "github.com/plantonhq/planton/apis/dev/planton/provider/scaleway"
	scalewaycontainerregistryv1 "github.com/plantonhq/planton/apis/dev/planton/provider/scaleway/scalewaycontainerregistry/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles resolved values from the stack input for use throughout
// the module.
//
// NOTE: Unlike most other Scaleway resource modules, there is no
// ScalewayTags field here. Scaleway Container Registry namespaces do
// not support tags in either the Pulumi SDK or the Terraform provider.
// The namespace name (from metadata.name) is the primary identifier.
type Locals struct {
	ScalewayProviderConfig    *scalewayprovider.ScalewayProviderConfig
	ScalewayContainerRegistry *scalewaycontainerregistryv1.ScalewayContainerRegistry
}

// initializeLocals copies stack-input fields into the Locals struct.
//
// Unlike other Scaleway resource modules, this does not build a tag
// slice because Scaleway Container Registry namespaces do not support
// tags in the API. Standard Planton metadata tags are not applied.
func initializeLocals(_ *pulumi.Context, stackInput *scalewaycontainerregistryv1.ScalewayContainerRegistryStackInput) *Locals {
	return &Locals{
		ScalewayContainerRegistry: stackInput.Target,
		ScalewayProviderConfig:    stackInput.ProviderConfig,
	}
}
