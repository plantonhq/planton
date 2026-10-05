package module

import (
	"github.com/pkg/errors"
	azurevirtualwanv1alpha1 "github.com/plantonhq/planton/catalog/azure/azurevirtualwan/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/network"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azurevirtualwanv1alpha1.AzureVirtualWanIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureVirtualWan.Spec

	// Create the Virtual WAN -- the free, lightweight umbrella object of
	// Azure's managed hub-and-spoke networking. Hubs (and the gateways
	// on them) are separate resources that reference this WAN's ID; ARM
	// refuses to delete a WAN that still has hubs.
	createdVirtualWan, err := network.NewVirtualWan(ctx,
		spec.Name,
		&network.VirtualWanArgs{
			Name:                 pulumi.String(spec.Name),
			Location:             pulumi.String(spec.Region),
			ResourceGroupName:    pulumi.String(locals.ResourceGroupName),
			DisableVpnEncryption: pulumi.Bool(spec.DisableVpnEncryption),
			// ARM defaults branch-to-branch transit ON -- most of the point
			// of a Virtual WAN; the optional field's default mirrors it.
			AllowBranchToBranchTraffic:     pulumi.Bool(optionalBool(spec.AllowBranchToBranchTraffic, true)),
			Office365LocalBreakoutCategory: pulumi.String(office365BreakoutWireValue(spec.Office365LocalBreakoutCategory)),
			// "Standard" is the full-mesh tier and ARM's default; "Basic"
			// is the constrained legacy tier (upgradeable, never
			// downgradeable).
			Type: pulumi.String(optionalString(spec.Type, "Standard")),
			Tags: pulumi.ToStringMap(locals.AzureTags),
		},
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create virtual wan %s", spec.Name)
	}

	ctx.Export(OpVirtualWanId, createdVirtualWan.ID())
	ctx.Export(OpVirtualWanName, createdVirtualWan.Name)

	return nil
}
