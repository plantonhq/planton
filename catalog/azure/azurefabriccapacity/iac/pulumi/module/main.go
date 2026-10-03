package module

import (
	"github.com/pkg/errors"
	azurefabriccapacityv1alpha1 "github.com/plantonhq/planton/catalog/azure/azurefabriccapacity/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/fabric"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azurefabriccapacityv1alpha1.AzureFabricCapacityIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureFabricCapacity.Spec

	// Create the Fabric capacity -- a running capacity bills PER HOUR
	// from the moment it exists; the F-SKU scales up and down in place.
	// "Fabric" is the SKU tier's only legal value at v5 (deliberately
	// not part of the spec); the administrators list is required
	// non-empty by the spec (Azure rejects a capacity created with
	// none, and clearing it on a running capacity is a lockout, not a
	// configuration). Create and delete are polled long-running
	// operations.
	createdCapacity, err := fabric.NewCapacity(ctx,
		locals.AzureFabricCapacity.Metadata.Name,
		&fabric.CapacityArgs{
			Name:                  pulumi.String(spec.Name),
			ResourceGroupName:     pulumi.String(spec.ResourceGroup.GetValue()),
			Location:              pulumi.String(spec.Region),
			AdministrationMembers: pulumi.ToStringArray(spec.AdministrationMembers),
			Sku: &fabric.CapacitySkuArgs{
				Name: pulumi.String(spec.SkuName),
				Tier: pulumi.String("Fabric"),
			},
			Tags: pulumi.ToStringMap(locals.AzureTags),
		},
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create fabric capacity %s",
			locals.AzureFabricCapacity.Metadata.Name)
	}

	ctx.Export(OpFabricCapacityId, createdCapacity.ID())
	ctx.Export(OpFabricCapacityName, createdCapacity.Name)

	return nil
}
