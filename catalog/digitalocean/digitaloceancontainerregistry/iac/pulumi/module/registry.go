package module

import (
	"github.com/pkg/errors"
	"github.com/plantonhq/planton/catalog/digitalocean"
	dosdk "github.com/pulumi/pulumi-digitalocean/sdk/v4/go/digitalocean"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// registry provisions the DigitalOcean Container Registry, optionally mints
// Docker credentials for it, and exports outputs.
//
// A DigitalOcean account holds exactly ONE container registry, and registry
// names are globally unique across ALL DigitalOcean accounts. Name and region
// are create-only; only the subscription tier can change after creation.
func registry(
	ctx *pulumi.Context,
	locals *Locals,
	digitalOceanProvider *dosdk.Provider,
) (*dosdk.ContainerRegistry, error) {
	spec := locals.DigitalOceanContainerRegistry.Spec

	// The proto enum value names ARE the DigitalOcean tier slugs.
	registryArgs := &dosdk.ContainerRegistryArgs{
		Name:                 pulumi.String(spec.Name),
		SubscriptionTierSlug: pulumi.String(spec.SubscriptionTier.String()),
	}

	// Region is Optional+Computed at the provider: unset stays nil so
	// DigitalOcean picks the region and reports the chosen slug back.
	if spec.Region != digitalocean.DigitalOceanRegion_digital_ocean_region_unspecified {
		registryArgs.Region = pulumi.StringPtr(spec.Region.String())
	}

	createdRegistry, err := dosdk.NewContainerRegistry(
		ctx,
		"registry",
		registryArgs,
		pulumi.Provider(digitalOceanProvider),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create digitalocean container registry")
	}

	ctx.Export(OpRegistryName, createdRegistry.Name)
	ctx.Export(OpRegion, createdRegistry.Region)
	ctx.Export(OpServerUrl, createdRegistry.ServerUrl)
	ctx.Export(OpEndpoint, createdRegistry.Endpoint)

	// Docker credentials, minted only when the spec asks for them. Neither
	// knob is recoverable from the DigitalOcean API afterwards (the API only
	// ever returns a freshly minted credential), and import of this resource
	// is defective at the current provider pin -- see the kind's import map.
	if creds := spec.DockerCredentials; creds != nil {
		credentialsArgs := &dosdk.ContainerRegistryDockerCredentialsArgs{
			RegistryName: createdRegistry.Name,
			Write:        pulumi.BoolPtr(creds.Write),
		}
		// Unset defers to the provider default: the API maximum (~50 years).
		if creds.ExpirySeconds != nil {
			credentialsArgs.ExpirySeconds = pulumi.IntPtr(int(*creds.ExpirySeconds))
		}

		createdCredentials, err := dosdk.NewContainerRegistryDockerCredentials(
			ctx,
			"credentials",
			credentialsArgs,
			pulumi.Provider(digitalOceanProvider),
		)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create digitalocean container registry docker credentials")
		}

		// The SDK does not flag DockerCredentials as secret, so wrap it
		// explicitly -- the token must never ship as a plaintext output.
		ctx.Export(OpDockerCredentials, pulumi.ToSecret(createdCredentials.DockerCredentials))
		ctx.Export(OpCredentialExpirationTime, createdCredentials.CredentialExpirationTime)
	} else {
		// Both provisioners export the full outputs contract; unconfigured
		// credentials surface as empty strings, identically on both engines.
		ctx.Export(OpDockerCredentials, pulumi.ToSecret(pulumi.String("")))
		ctx.Export(OpCredentialExpirationTime, pulumi.String(""))
	}

	return createdRegistry, nil
}
