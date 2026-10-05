package module

import (
	"github.com/pkg/errors"
	azureresourcegroupv1alpha1 "github.com/plantonhq/planton/catalog/azure/azureresourcegroup/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/core"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azureresourcegroupv1alpha1.AzureResourceGroupIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient
	// chain). preventDeletionIfContainsResources is flipped off so destroy of a Planton
	// resource group means ARM-delete the group (matching `az group delete`): Azure
	// services leave nested furniture this module never created (Application Insights'
	// "Smart Detection" action group is the first live-caught case) and the provider
	// default would strand the group. Mirrors the Terraform module's features block.
	azureProvider, err := pulumiazureprovider.GetWithFeatures(ctx, iacInput.ProviderConfig,
		azure.ProviderFeaturesArgs{
			ResourceGroup: azure.ProviderFeaturesResourceGroupArgs{
				PreventDeletionIfContainsResources: pulumi.Bool(false),
			},
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

	// Export outputs
	ctx.Export(OpResourceGroupId, rg.ID())
	ctx.Export(OpResourceGroupName, rg.Name)
	ctx.Export(OpRegion, pulumi.String(spec.Region))

	return nil
}
