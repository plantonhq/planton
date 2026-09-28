package module

import (
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"strconv"

	"github.com/pkg/errors"
	azurednsrecordv1alpha1 "github.com/plantonhq/planton/catalog/azure/azurednsrecord/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/pulumiazureprovider"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/dns"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// caaTagWireValues maps the spec's CAA tag enum value names to Azure's
// wire vocabulary -- the provider validates the lowercase form
// case-sensitively.
var caaTagWireValues = map[string]string{
	"ISSUE":        "issue",
	"ISSUEWILD":    "issuewild",
	"IODEF":        "iodef",
	"CONTACTEMAIL": "contactemail",
}

// Resources creates one DNS record set in an Azure public DNS zone.
//
// The record type is whichever typed payload the spec carries (validation
// guarantees exactly one), so exactly one branch below runs. Azure's
// management plane addresses record sets by (resource group, zone name,
// type, record name) -- there is no ARM-id addressing mode for record
// sets on either engine.
//
// Alias records (A/AAAA/CNAME only): when the payload carries
// target_resource_id instead of literal values, Azure keeps the answer in
// sync with the referenced resource -- no drift window when a Public IP's
// address changes, and a way to point the zone APEX at an Azure resource
// where DNS itself forbids CNAME.
func Resources(ctx *pulumi.Context, stackInput *azurednsrecordv1alpha1.AzureDnsRecordStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	// Build the Azure provider from the stack input via the shared builder, which resolves
	// the right credential mechanism (static client secret, keyless web identity, or ambient chain).
	azureProvider, err := pulumiazureprovider.Get(ctx, stackInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create azure provider")
	}

	spec := locals.AzureDnsRecord.Spec

	// The platform materializes the proto default (300) before the module
	// runs; the presence guard is a same-value safety net for direct
	// stack-input paths, never a different fallback.
	ttl := 300
	if spec.TtlSeconds != nil {
		ttl = int(*spec.TtlSeconds)
	}

	// Exactly one record resource materializes; both its ARM id and its
	// fqdn flatten onto the same stack outputs regardless of type.
	var recordId pulumi.StringOutput
	var fqdn pulumi.StringOutput

	switch {
	case spec.A != nil:
		args := &dns.ARecordArgs{
			Name:              pulumi.String(spec.Name),
			ZoneName:          pulumi.String(locals.ZoneName),
			ResourceGroupName: pulumi.String(locals.ResourceGroupName),
			Ttl:               pulumi.Int(ttl),
			Tags:              pulumi.ToStringMap(locals.AzureTags),
		}
		// Spec validation guarantees exactly one of addresses/alias; the
		// unused argument stays nil so the provider never sees an
		// empty-but-present value.
		if len(spec.A.Addresses) > 0 {
			args.Records = pulumi.ToStringArray(stringValues(spec.A.Addresses))
		}
		if spec.A.TargetResourceId != nil && spec.A.TargetResourceId.GetValue() != "" {
			args.TargetResourceId = pulumi.String(spec.A.TargetResourceId.GetValue())
		}
		created, err := dns.NewARecord(ctx, "main", args, pulumi.Provider(azureProvider))
		if err != nil {
			return errors.Wrapf(err, "failed to create a record %s", spec.Name)
		}
		recordId = created.ID().ToStringOutput()
		fqdn = created.Fqdn

	case spec.Aaaa != nil:
		args := &dns.AaaaRecordArgs{
			Name:              pulumi.String(spec.Name),
			ZoneName:          pulumi.String(locals.ZoneName),
			ResourceGroupName: pulumi.String(locals.ResourceGroupName),
			Ttl:               pulumi.Int(ttl),
			Tags:              pulumi.ToStringMap(locals.AzureTags),
		}
		if len(spec.Aaaa.Addresses) > 0 {
			args.Records = pulumi.ToStringArray(stringValues(spec.Aaaa.Addresses))
		}
		if spec.Aaaa.TargetResourceId != nil && spec.Aaaa.TargetResourceId.GetValue() != "" {
			args.TargetResourceId = pulumi.String(spec.Aaaa.TargetResourceId.GetValue())
		}
		created, err := dns.NewAaaaRecord(ctx, "main", args, pulumi.Provider(azureProvider))
		if err != nil {
			return errors.Wrapf(err, "failed to create aaaa record %s", spec.Name)
		}
		recordId = created.ID().ToStringOutput()
		fqdn = created.Fqdn

	case spec.Cname != nil:
		args := &dns.CNameRecordArgs{
			Name:              pulumi.String(spec.Name),
			ZoneName:          pulumi.String(locals.ZoneName),
			ResourceGroupName: pulumi.String(locals.ResourceGroupName),
			Ttl:               pulumi.Int(ttl),
			Tags:              pulumi.ToStringMap(locals.AzureTags),
		}
		// value is a StringValueOrRef; the platform resolves valueFrom
		// references before the module runs, so GetValue() is the
		// resolved literal (e.g. a Front Door endpoint's host_name).
		if spec.Cname.Value != nil && spec.Cname.Value.GetValue() != "" {
			args.Record = pulumi.String(spec.Cname.Value.GetValue())
		}
		if spec.Cname.TargetResourceId != nil && spec.Cname.TargetResourceId.GetValue() != "" {
			args.TargetResourceId = pulumi.String(spec.Cname.TargetResourceId.GetValue())
		}
		created, err := dns.NewCNameRecord(ctx, "main", args, pulumi.Provider(azureProvider))
		if err != nil {
			return errors.Wrapf(err, "failed to create cname record %s", spec.Name)
		}
		recordId = created.ID().ToStringOutput()
		fqdn = created.Fqdn

	case len(spec.Mx) > 0:
		// Each entry carries its own preference, so multi-server mail
		// setups (10 primary / 20 secondary) express exactly. The
		// provider's preference field is string-typed; the spec's integer
		// is converted here.
		mxRecords := make(dns.MxRecordRecordArray, 0, len(spec.Mx))
		for _, entry := range spec.Mx {
			mxRecords = append(mxRecords, &dns.MxRecordRecordArgs{
				Preference: pulumi.String(strconv.Itoa(int(entry.GetPreference()))),
				Exchange:   pulumi.String(entry.Exchange.GetValue()),
			})
		}
		created, err := dns.NewMxRecord(ctx, "main", &dns.MxRecordArgs{
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

	case len(spec.Srv) > 0:
		srvRecords := make(dns.SrvRecordRecordArray, 0, len(spec.Srv))
		for _, entry := range spec.Srv {
			srvRecords = append(srvRecords, &dns.SrvRecordRecordArgs{
				Priority: pulumi.Int(int(entry.GetPriority())),
				Weight:   pulumi.Int(int(entry.GetWeight())),
				Port:     pulumi.Int(int(entry.GetPort())),
				Target:   pulumi.String(entry.Target.GetValue()),
			})
		}
		created, err := dns.NewSrvRecord(ctx, "main", &dns.SrvRecordArgs{
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

	case len(spec.Caa) > 0:
		caaRecords := make(dns.CaaRecordRecordArray, 0, len(spec.Caa))
		for _, entry := range spec.Caa {
			caaRecords = append(caaRecords, &dns.CaaRecordRecordArgs{
				Flags: pulumi.Int(int(entry.GetFlags())),
				Tag:   pulumi.String(caaTagWireValues[entry.Tag.String()]),
				Value: pulumi.String(entry.Value),
			})
		}
		created, err := dns.NewCaaRecord(ctx, "main", &dns.CaaRecordArgs{
			Name:              pulumi.String(spec.Name),
			ZoneName:          pulumi.String(locals.ZoneName),
			ResourceGroupName: pulumi.String(locals.ResourceGroupName),
			Ttl:               pulumi.Int(ttl),
			Records:           caaRecords,
			Tags:              pulumi.ToStringMap(locals.AzureTags),
		}, pulumi.Provider(azureProvider))
		if err != nil {
			return errors.Wrapf(err, "failed to create caa record %s", spec.Name)
		}
		recordId = created.ID().ToStringOutput()
		fqdn = created.Fqdn

	case len(spec.Txt) > 0:
		// Values up to 4096 characters are legal: the provider
		// transparently splits each into the 254-character strings DNS
		// requires and reassembles them on read. Each value is a
		// StringValueOrRef; the platform resolves valueFrom references
		// before the module runs (e.g. a Front Door custom domain's
		// validation_token), so GetValue() is the resolved literal.
		txtRecords := make(dns.TxtRecordRecordArray, 0, len(spec.Txt))
		for _, value := range spec.Txt {
			txtRecords = append(txtRecords, &dns.TxtRecordRecordArgs{
				Value: pulumi.String(value.GetValue()),
			})
		}
		created, err := dns.NewTxtRecord(ctx, "main", &dns.TxtRecordArgs{
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

	case len(spec.Ns) > 0:
		// Delegates a CHILD subdomain to another zone's name servers.
		// The zone's own apex NS records are Azure-managed.
		created, err := dns.NewNsRecord(ctx, "main", &dns.NsRecordArgs{
			Name:              pulumi.String(spec.Name),
			ZoneName:          pulumi.String(locals.ZoneName),
			ResourceGroupName: pulumi.String(locals.ResourceGroupName),
			Ttl:               pulumi.Int(ttl),
			Records:           pulumi.ToStringArray(stringValues(spec.Ns)),
			Tags:              pulumi.ToStringMap(locals.AzureTags),
		}, pulumi.Provider(azureProvider))
		if err != nil {
			return errors.Wrapf(err, "failed to create ns record %s", spec.Name)
		}
		recordId = created.ID().ToStringOutput()
		fqdn = created.Fqdn

	case len(spec.Ptr) > 0:
		created, err := dns.NewPtrRecord(ctx, "main", &dns.PtrRecordArgs{
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
