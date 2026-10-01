package module

import (
	"github.com/pkg/errors"
	azurednszonev1alpha1 "github.com/plantonhq/planton/catalog/azure/azurednszone/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/dns"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *azurednszonev1alpha1.AzureDnsZoneStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	// Build the Azure provider from the stack input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, stackInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureDnsZone.Spec

	// Lifecycle notes worth knowing before operating this resource:
	// - Tags update IN PLACE; the zone's name is its ARM identity, so
	//   renaming replaces the zone, every record in it, AND the assigned
	//   name-server set -- breaking the registrar delegation until it is
	//   updated.
	// - Public DNS zones are global resources -- no location/region.
	// - The zone is deliberately just the zone: records are declared
	//   through AzureDnsRecord resources referencing this zone's
	//   zone_name output, one resource per record set.
	// - Creating the zone does NOT make it authoritative: the domain
	//   resolves through it only once the nameservers output is
	//   configured at the registrar (or as parent-zone NS records for
	//   subdomain delegation).
	zoneArgs := &dns.ZoneArgs{
		Name:              pulumi.String(spec.ZoneName),
		ResourceGroupName: pulumi.String(locals.ResourceGroupName),
		Tags:              pulumi.ToStringMap(locals.AzureTags),
	}

	// Omitted means Azure creates its standard SOA record; the block is
	// only sent when the spec customizes it. Unset timers fall back to
	// Azure's defaults so a partially-specified block and Azure's own
	// values deploy identically on both engines. The SOA host name is
	// never sent -- Azure owns it and rejects changes.
	if spec.SoaRecord != nil {
		soaArgs := &dns.ZoneSoaRecordArgs{
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
		if spec.SoaRecord.SerialNumber != nil {
			soaArgs.SerialNumber = pulumi.Int(int(*spec.SoaRecord.SerialNumber))
		}
		if spec.SoaRecord.Ttl != nil {
			soaArgs.Ttl = pulumi.Int(int(*spec.SoaRecord.Ttl))
		}
		if len(spec.SoaRecord.Tags) > 0 {
			soaArgs.Tags = pulumi.ToStringMap(spec.SoaRecord.Tags)
		}
		zoneArgs.SoaRecord = soaArgs
	}

	createdZone, err := dns.NewZone(ctx,
		spec.ZoneName,
		zoneArgs,
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create dns zone %s", spec.ZoneName)
	}

	// Export stack outputs from the created resource. zone_name (with
	// resource_group_name) is the join key AzureDnsRecord resources
	// address record sets through; zone_id is the ARM-id seam for kinds
	// that watch the zone as a whole (Front Door custom-domain
	// validation, AKS web-app routing); nameservers is the registrar
	// delegation handoff.
	ctx.Export(OpZoneId, createdZone.ID())
	ctx.Export(OpZoneName, createdZone.Name)
	ctx.Export(OpResourceGroupName, createdZone.ResourceGroupName)
	ctx.Export(OpNameservers, createdZone.NameServers)
	ctx.Export(OpMaxNumberOfRecordSets, createdZone.MaxNumberOfRecordSets)

	return nil
}
