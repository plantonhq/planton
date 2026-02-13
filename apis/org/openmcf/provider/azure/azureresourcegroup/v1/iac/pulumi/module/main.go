package module

import (
	"github.com/pkg/errors"
	azureresourcegroupv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/azure/azureresourcegroup/v1"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/core"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *azureresourcegroupv1.AzureResourceGroupStackInput) error {
	locals := initializeLocals(ctx, stackInput)
	azureProviderConfig := stackInput.ProviderConfig

	// Create azure provider using the credentials from the input
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

	spec := locals.AzureResourceGroup.Spec

	// Create the Resource Group
	rg, err := core.NewResourceGroup(ctx,
		spec.Name,
		&core.ResourceGroupArgs{
			Name:     pulumi.String(spec.Name),
			Location: pulumi.String(spec.Region),
			Tags:     pulumi.ToStringMap(locals.AzureTags),
		},
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create resource group %s", spec.Name)
	}

	// Export stack outputs
	ctx.Export(OpResourceGroupId, rg.ID())
	ctx.Export(OpResourceGroupName, rg.Name)
	ctx.Export(OpRegion, pulumi.String(spec.Region))

	return nil
}
