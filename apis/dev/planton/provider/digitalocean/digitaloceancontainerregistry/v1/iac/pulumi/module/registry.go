package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-digitalocean/sdk/v4/go/digitalocean"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// registry provisions the DigitalOcean Container Registry and exports outputs.
func registry(
	ctx *pulumi.Context,
	locals *Locals,
	digitalOceanProvider *digitalocean.Provider,
) (*digitalocean.ContainerRegistry, error) {

	// 1. Get subscription tier directly from enum (values match DigitalOcean API slugs).
	tierSlug := "starter" // default to starter tier
	if locals.DigitalOceanContainerRegistry.Spec.SubscriptionTier != 0 {
		tierSlug = locals.DigitalOceanContainerRegistry.Spec.SubscriptionTier.String()
	}

	// 2. Build the resource arguments from proto fields.
	registryArgs := &digitalocean.ContainerRegistryArgs{
		Name:                 pulumi.String(locals.DigitalOceanContainerRegistry.Spec.Name),
		SubscriptionTierSlug: pulumi.String(tierSlug),
		Region:               pulumi.StringPtr(locals.DigitalOceanContainerRegistry.Spec.Region.String()),
	}

	// 3. Create the Container Registry.
	createdRegistry, err := digitalocean.NewContainerRegistry(
		ctx,
		"registry",
		registryArgs,
		pulumi.Provider(digitalOceanProvider),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create digitalocean container registry")
	}

	// 4. Warn if garbage collection flag is set (provider limitation).
	if locals.DigitalOceanContainerRegistry.Spec.GarbageCollectionEnabled {
		ctx.Log.Warn(
			"garbage_collection_enabled is set but currently not supported by the DigitalOcean provider; ignoring.",
			&pulumi.LogArgs{},
		)
	}

	// 5. Export stack outputs.
	ctx.Export(OpRegistryName, createdRegistry.Name)
	ctx.Export(OpRegion, createdRegistry.Region)
	ctx.Export(OpServerUrl, pulumi.Sprintf("registry.digitalocean.com/%s", createdRegistry.Name))

	return createdRegistry, nil
}
