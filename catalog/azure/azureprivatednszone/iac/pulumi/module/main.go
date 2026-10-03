package module

import (
	"github.com/pkg/errors"
	azureprivatednszonev1alpha1 "github.com/plantonhq/planton/catalog/azure/azureprivatednszone/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/privatedns"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, iacInput *azureprivatednszonev1alpha1.AzurePrivateDnsZoneIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzurePrivateDnsZone.Spec

	// Lifecycle notes worth knowing before operating this resource:
	// - Tags update IN PLACE; the zone's name is its ARM identity, so
	//   renaming replaces the zone AND every record in it. The SOA record
	//   is written at creation and cannot be customized afterwards.
	// - Private DNS zones are global resources -- no location/region.
	// - The zone is deliberately just the zone: which networks can resolve
	//   it is declared through AzurePrivateDnsZoneVirtualNetworkLink
	//   resources referencing this zone's zone_id output, one per network.
	//   A zone with no links answers nobody.
	zoneArgs := &privatedns.ZoneArgs{
		Name:              pulumi.String(spec.Name),
		ResourceGroupName: pulumi.String(locals.ResourceGroupName),
		Tags:              pulumi.ToStringMap(locals.AzureTags),
	}

	// Omitted means Azure creates its standard SOA record; the block is
	// only sent when the spec customizes it. Unset timers fall back to
	// Azure's defaults so a partially-specified block and Azure's own
	// values deploy identically on both engines.
	if spec.SoaRecord != nil {
		soaArgs := &privatedns.ZoneSoaRecordArgs{
			Email: pulumi.String(spec.SoaRecord.Email),
		}
		if spec.SoaRecord.ExpireTime != nil {
			soaArgs.ExpireTime = pulumi.Int(int(*spec.SoaRecord.ExpireTime))
		}
		if spec.SoaRecord.MinimumTtl != nil {
			soaArgs.MinimumTtl = pulumi.Int(int(*spec.SoaRecord.MinimumTtl))
		}
		if spec.SoaRecord.RefreshTime != nil {
			soaArgs.RefreshTime = pulumi.Int(int(*spec.SoaRecord.RefreshTime))
		}
		if spec.SoaRecord.RetryTime != nil {
			soaArgs.RetryTime = pulumi.Int(int(*spec.SoaRecord.RetryTime))
		}
		if spec.SoaRecord.Ttl != nil {
			soaArgs.Ttl = pulumi.Int(int(*spec.SoaRecord.Ttl))
		}
		if len(spec.SoaRecord.Tags) > 0 {
			soaArgs.Tags = pulumi.ToStringMap(spec.SoaRecord.Tags)
		}
		zoneArgs.SoaRecord = soaArgs
	}

	createdZone, err := privatedns.NewZone(ctx,
		spec.Name,
		zoneArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create private dns zone %s", spec.Name)
	}

	// Export outputs from the created resource. zone_id is the join
	// key virtual network links, private endpoints, and VNet-integrated
	// databases attach through; resource_group_name is echoed for tooling
	// that addresses records by zone name + resource group.
	ctx.Export(OpZoneId, createdZone.ID())
	ctx.Export(OpZoneName, createdZone.Name)
	ctx.Export(OpResourceGroupName, createdZone.ResourceGroupName)

	return nil
}
