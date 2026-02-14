package module

import (
	"github.com/pkg/errors"
	azurefrontdoorprofilev1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/azure/azurefrontdoorprofile/v1"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *azurefrontdoorprofilev1.AzureFrontDoorProfileStackInput) error {
	locals := initializeLocals(ctx, stackInput)
	azureProviderConfig := stackInput.ProviderConfig

	// Create azure provider using the credentials from the input.
	azureProvider, err := azure.NewProvider(ctx,
		"azure",
		&azure.ProviderArgs{
			ClientId:       pulumi.String(azureProviderConfig.ClientId),
			ClientSecret:   pulumi.String(azureProviderConfig.ClientSecret),
			SubscriptionId: pulumi.String(azureProviderConfig.SubscriptionId),
			TenantId:       pulumi.String(azureProviderConfig.TenantId),
		})
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	// Create the Front Door profile.
	profile, err := createProfile(ctx, locals, azureProvider)
	if err != nil {
		return err
	}

	// Create endpoints.
	endpoints, err := createEndpoints(ctx, locals, azureProvider, profile)
	if err != nil {
		return err
	}

	// Create origin groups.
	originGroups, err := createOriginGroups(ctx, locals, azureProvider, profile)
	if err != nil {
		return err
	}

	// Create origins for each origin group.
	origins, err := createAllOrigins(ctx, azureProvider, originGroups)
	if err != nil {
		return err
	}

	// Create routes.
	if err := createRoutes(ctx, locals, azureProvider, endpoints, originGroups, origins); err != nil {
		return err
	}

	// Export stack outputs.
	ctx.Export(OpProfileId, profile.ID())
	ctx.Export(OpProfileName, profile.Name)
	ctx.Export(OpResourceGuid, profile.ResourceGuid)

	endpointIds := pulumi.StringMap{}
	endpointHostnames := pulumi.StringMap{}
	for name, ep := range endpoints {
		endpointIds[name] = ep.Resource.ID().ToStringOutput()
		endpointHostnames[name] = ep.Resource.HostName
	}
	ctx.Export(OpEndpointIds, endpointIds)
	ctx.Export(OpEndpointHostnames, endpointHostnames)

	return nil
}
