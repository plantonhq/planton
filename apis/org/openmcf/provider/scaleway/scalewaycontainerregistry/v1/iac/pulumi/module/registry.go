package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	scalewayv2 "github.com/pulumiverse/pulumi-scaleway/sdk/go/scaleway"
	"github.com/pulumiverse/pulumi-scaleway/sdk/go/scaleway/registry"
)

// registryNamespace provisions the Scaleway Container Registry namespace
// and exports stack outputs.
//
// Uses the registry.NewNamespace function from the scaleway/registry
// subpackage (the preferred API path, replacing the deprecated top-level
// scaleway.NewRegistryNamespace).
func registryNamespace(
	ctx *pulumi.Context,
	locals *Locals,
	scalewayProvider *scalewayv2.Provider,
) error {
	spec := locals.ScalewayContainerRegistry.Spec

	// Build namespace arguments from spec fields.
	//
	// NOTE: The Scaleway Pulumi SDK's registry.NamespaceArgs does not
	// expose a Tags field (unlike most other Scaleway resources). Tags
	// are applied in the Terraform module but cannot be set via Pulumi.
	// This is a known limitation of the pulumiverse SDK v1.43.0.
	namespaceArgs := &registry.NamespaceArgs{
		Name:     pulumi.StringPtr(locals.ScalewayContainerRegistry.Metadata.Name),
		Region:   pulumi.StringPtr(spec.Region),
		IsPublic: pulumi.BoolPtr(spec.IsPublic),
	}

	// Description is optional -- only set if provided.
	if spec.Description != "" {
		namespaceArgs.Description = pulumi.StringPtr(spec.Description)
	}

	// Create the registry namespace.
	createdNamespace, err := registry.NewNamespace(
		ctx,
		"registry",
		namespaceArgs,
		pulumi.Provider(scalewayProvider),
	)
	if err != nil {
		return errors.Wrap(err, "failed to create scaleway container registry namespace")
	}

	// Export stack outputs.
	ctx.Export(OpNamespaceId, createdNamespace.ID())
	ctx.Export(OpEndpoint, createdNamespace.Endpoint)
	ctx.Export(OpNamespaceName, createdNamespace.Name)
	ctx.Export(OpRegion, pulumi.String(spec.Region))

	return nil
}
