package module

import (
	"github.com/pkg/errors"
	azureprivatednsrecordv1alpha1 "github.com/plantonhq/planton/catalog/azure/azureprivatednsrecord/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/privatedns"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources creates one DNS record set in an Azure PRIVATE DNS zone.
//
// The record type is whichever typed payload the spec carries (validation
// guarantees exactly one), so exactly one branch below runs. The zone
// reference arrives as its ARM id and is split into (resource group, zone
// name) in locals -- this SDK's record resources address the zone by its
// segments, the Terraform provider by the id; both write the same ARM
// object.
//
// Private DNS has no alias records (a public-DNS concept) and supports
// exactly these seven types -- no CAA, no NS (private zones cannot
// delegate subdomains).
func Resources(ctx *pulumi.Context, iacInput *azureprivatednsrecordv1alpha1.AzurePrivateDnsRecordIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the Azure provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzurePrivateDnsRecord.Spec

	// The platform materializes the proto default (300) before the module
	// runs; the presence guard is a same-value safety net for direct
	// iac-input paths, never a different fallback.
	ttl := 300
	if spec.TtlSeconds != nil {
		ttl = int(*spec.TtlSeconds)
	}

	// Exactly one record resource materializes; both its ARM id and its
	// fqdn flatten onto the same outputs regardless of type.
	var recordId pulumi.StringOutput
	var fqdn pulumi.StringOutput

	switch {
	case len(spec.A) > 0:
		created, err := privatedns.NewARecord(ctx, "main", &privatedns.ARecordArgs{
			Name:              pulumi.String(spec.Name),
			ZoneName:          pulumi.String(locals.ZoneName),
			ResourceGroupName: pulumi.String(locals.ResourceGroupName),
			Ttl:               pulumi.Int(ttl),
			Records:           pulumi.ToStringArray(stringValues(spec.A)),
			Tags:              pulumi.ToStringMap(locals.AzureTags),
		}, pulumi.Provider(azureProvider))
		if err != nil {
			return errors.Wrapf(err, "failed to create a record %s", spec.Name)
		}
		recordId = created.ID().ToStringOutput()
		fqdn = created.Fqdn

	case len(spec.Aaaa) > 0:
		created, err := privatedns.NewAAAARecord(ctx, "main", &privatedns.AAAARecordArgs{
			Name:              pulumi.String(spec.Name),
			ZoneName:          pulumi.String(locals.ZoneName),
			ResourceGroupName: pulumi.String(locals.ResourceGroupName),
			Ttl:               pulumi.Int(ttl),
			Records:           pulumi.ToStringArray(stringValues(spec.Aaaa)),
			Tags:              pulumi.ToStringMap(locals.AzureTags),
		}, pulumi.Provider(azureProvider))
		if err != nil {
			return errors.Wrapf(err, "failed to create aaaa record %s", spec.Name)
		}
		recordId = created.ID().ToStringOutput()
		fqdn = created.Fqdn

	case spec.Cname != nil && spec.Cname.GetValue() != "":
		// cname is a StringValueOrRef; the platform resolves valueFrom
		// references before the module runs, so GetValue() is the
		// resolved literal target hostname.
		created, err := privatedns.NewCnameRecord(ctx, "main", &privatedns.CnameRecordArgs{
			Name:              pulumi.String(spec.Name),
			ZoneName:          pulumi.String(locals.ZoneName),
			ResourceGroupName: pulumi.String(locals.ResourceGroupName),
			Ttl:               pulumi.Int(ttl),
			Record:            pulumi.String(spec.Cname.GetValue()),
			Tags:              pulumi.ToStringMap(locals.AzureTags),
		}, pulumi.Provider(azureProvider))
		if err != nil {
			return errors.Wrapf(err, "failed to create cname record %s", spec.Name)
		}
		recordId = created.ID().ToStringOutput()
		fqdn = created.Fqdn

	case len(spec.Mx) > 0:
		// Each entry carries its own preference, so multi-server mail
		// setups (10 primary / 20 secondary) express exactly.
		mxRecords := make(privatedns.MxRecordRecordArray, 0, len(spec.Mx))
		for _, entry := range spec.Mx {
			mxRecords = append(mxRecords, &privatedns.MxRecordRecordArgs{
				Preference: pulumi.Int(int(entry.GetPreference())),
				Exchange:   pulumi.String(entry.Exchange.GetValue()),
			})
		}
		created, err := privatedns.NewMxRecord(ctx, "main", &privatedns.MxRecordArgs{
			Name:              pulumi.String(spec.Name),
			ZoneName:          pulumi.String(locals.ZoneName),
			ResourceGroupName: pulumi.String(locals.ResourceGroupName),
			Ttl:               pulumi.Int(ttl),
			Records:           mxRecords,
			Tags:              pulumi.ToStringMap(locals.AzureTags),
		}, pulumi.Provider(azureProvider))
		if err != nil {
			return errors.Wrapf(err, "failed to create mx record %s", spec.Name)
		}
		recordId = created.ID().ToStringOutput()
		fqdn = created.Fqdn

	case len(spec.Ptr) > 0:
		created, err := privatedns.NewPTRRecord(ctx, "main", &privatedns.PTRRecordArgs{
			Name:              pulumi.String(spec.Name),
			ZoneName:          pulumi.String(locals.ZoneName),
			ResourceGroupName: pulumi.String(locals.ResourceGroupName),
			Ttl:               pulumi.Int(ttl),
			Records:           pulumi.ToStringArray(stringValues(spec.Ptr)),
			Tags:              pulumi.ToStringMap(locals.AzureTags),
		}, pulumi.Provider(azureProvider))
		if err != nil {
			return errors.Wrapf(err, "failed to create ptr record %s", spec.Name)
		}
		recordId = created.ID().ToStringOutput()
		fqdn = created.Fqdn

	case len(spec.Srv) > 0:
		srvRecords := make(privatedns.SRVRecordRecordArray, 0, len(spec.Srv))
		for _, entry := range spec.Srv {
			srvRecords = append(srvRecords, &privatedns.SRVRecordRecordArgs{
				Priority: pulumi.Int(int(entry.GetPriority())),
				Weight:   pulumi.Int(int(entry.GetWeight())),
				Port:     pulumi.Int(int(entry.GetPort())),
				Target:   pulumi.String(entry.Target.GetValue()),
			})
		}
		created, err := privatedns.NewSRVRecord(ctx, "main", &privatedns.SRVRecordArgs{
			Name:              pulumi.String(spec.Name),
			ZoneName:          pulumi.String(locals.ZoneName),
			ResourceGroupName: pulumi.String(locals.ResourceGroupName),
			Ttl:               pulumi.Int(ttl),
			Records:           srvRecords,
			Tags:              pulumi.ToStringMap(locals.AzureTags),
		}, pulumi.Provider(azureProvider))
		if err != nil {
			return errors.Wrapf(err, "failed to create srv record %s", spec.Name)
		}
		recordId = created.ID().ToStringOutput()
		fqdn = created.Fqdn

	case len(spec.Txt) > 0:
		// Each value is a StringValueOrRef; the platform resolves
		// valueFrom references before the module runs, so GetValue() is
		// the resolved literal. Azure caps each value at 1,024
		// characters (ARM enforces).
		txtRecords := make(privatedns.TxtRecordRecordArray, 0, len(spec.Txt))
		for _, value := range spec.Txt {
			txtRecords = append(txtRecords, &privatedns.TxtRecordRecordArgs{
				Value: pulumi.String(value.GetValue()),
			})
		}
		created, err := privatedns.NewTxtRecord(ctx, "main", &privatedns.TxtRecordArgs{
			Name:              pulumi.String(spec.Name),
			ZoneName:          pulumi.String(locals.ZoneName),
			ResourceGroupName: pulumi.String(locals.ResourceGroupName),
			Ttl:               pulumi.Int(ttl),
			Records:           txtRecords,
			Tags:              pulumi.ToStringMap(locals.AzureTags),
		}, pulumi.Provider(azureProvider))
		if err != nil {
			return errors.Wrapf(err, "failed to create txt record %s", spec.Name)
		}
		recordId = created.ID().ToStringOutput()
		fqdn = created.Fqdn

	default:
		// Unreachable: spec validation requires exactly one payload.
		return errors.New("no record payload present in spec")
	}

	ctx.Export(OpRecordId, recordId)
	ctx.Export(OpFqdn, fqdn)

	return nil
}

// stringValues resolves each value-or-reference entry to the string the
// provider receives (references are resolved to values before the module runs).
func stringValues(entries []*foreignkeyv1.StringValueOrRef) []string {
	values := make([]string, 0, len(entries))
	for _, entry := range entries {
		values = append(values, entry.GetValue())
	}
	return values
}
