package module

import (
	"github.com/pkg/errors"
	scalewaycontainerregistryv1 "github.com/plantonhq/planton/apis/dev/planton/provider/scaleway/scalewaycontainerregistry/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/scaleway/pulumiscalewayprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the module entry point that provisions a Scaleway
// Container Registry namespace.
//
// This is a standalone resource (not composite): it wraps a single
// scaleway_registry_namespace resource. The registry namespace is an
// OCI-compliant container image registry with a Docker-compatible
// endpoint for push/pull operations.
func Resources(
	ctx *pulumi.Context,
	stackInput *scalewaycontainerregistryv1.ScalewayContainerRegistryStackInput,
) error {
	// 1. Prepare locals (metadata, tags, resolved references).
	locals := initializeLocals(ctx, stackInput)

	// 2. Create a Scaleway provider from the supplied credential.
	scalewayProvider, err := pulumiscalewayprovider.Get(
		ctx,
		stackInput.ProviderConfig,
	)
	if err != nil {
		return errors.Wrap(err, "failed to setup scaleway provider")
	}

	// 3. Create the registry namespace and export outputs.
	if err := registryNamespace(ctx, locals, scalewayProvider); err != nil {
		return errors.Wrap(err, "failed to create container registry namespace")
	}

	return nil
}
