package module

import (
	"github.com/pkg/errors"
	azureresourcegroupv1 "github.com/plantonhq/planton/apis/dev/planton/provider/azure/azureresourcegroup/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/core"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *azureresourcegroupv1.AzureResourceGroupStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	// Build the Azure provider from the stack input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, stackInput.ProviderConfig)
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
