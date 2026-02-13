package module

import (
	"fmt"

	"github.com/pkg/errors"
	azurednsrecordv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/azure/azurednsrecord/v1"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/dns"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *azurednsrecordv1.AzureDnsRecordStackInput) error {
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

	spec := locals.AzureDnsRecord.Spec
	recordType := spec.Type
	recordName := locals.RecordName
	zoneName := locals.ZoneName
	resourceGroup := locals.ResourceGroupName
	ttl := locals.TTL

	var recordId pulumi.IDOutput
	var fqdn pulumi.StringOutput

	// Create the appropriate DNS record based on record type
	switch recordType {
	case azurednsrecordv1.AzureDnsRecordSpec_A:
		record, err := dns.NewARecord(ctx,
			"dns-a-record",
			&dns.ARecordArgs{
				Name:              pulumi.String(recordName),
				ZoneName:          pulumi.String(zoneName),
				ResourceGroupName: pulumi.String(resourceGroup),
				Ttl:               pulumi.Int(ttl),
				Records:           pulumi.ToStringArray(spec.Values),
				Tags:              pulumi.ToStringMap(locals.AzureTags),
			},
			pulumi.Provider(azureProvider))
		if err != nil {
			return errors.Wrapf(err, "failed to create A record %s", recordName)
		}
		recordId = record.ID()
		fqdn = record.Fqdn

	case azurednsrecordv1.AzureDnsRecordSpec_AAAA:
		record, err := dns.NewAaaaRecord(ctx,
			"dns-aaaa-record",
			&dns.AaaaRecordArgs{
				Name:              pulumi.String(recordName),
				ZoneName:          pulumi.String(zoneName),
				ResourceGroupName: pulumi.String(resourceGroup),
				Ttl:               pulumi.Int(ttl),
				Records:           pulumi.ToStringArray(spec.Values),
				Tags:              pulumi.ToStringMap(locals.AzureTags),
			},
			pulumi.Provider(azureProvider))
		if err != nil {
			return errors.Wrapf(err, "failed to create AAAA record %s", recordName)
		}
		recordId = record.ID()
		fqdn = record.Fqdn

	case azurednsrecordv1.AzureDnsRecordSpec_CNAME:
		if len(spec.Values) == 0 {
			return errors.New("CNAME record requires at least one value")
		}
		record, err := dns.NewCNameRecord(ctx,
			"dns-cname-record",
			&dns.CNameRecordArgs{
				Name:              pulumi.String(recordName),
				ZoneName:          pulumi.String(zoneName),
				ResourceGroupName: pulumi.String(resourceGroup),
				Ttl:               pulumi.Int(ttl),
				Record:            pulumi.String(spec.Values[0]),
				Tags:              pulumi.ToStringMap(locals.AzureTags),
			},
			pulumi.Provider(azureProvider))
		if err != nil {
			return errors.Wrapf(err, "failed to create CNAME record %s", recordName)
		}
		recordId = record.ID()
		fqdn = record.Fqdn

	case azurednsrecordv1.AzureDnsRecordSpec_MX:
		mxRecords := make(dns.MxRecordRecordArray, 0)
		for _, value := range spec.Values {
			mxRecords = append(mxRecords, &dns.MxRecordRecordArgs{
				Preference: pulumi.String(fmt.Sprintf("%d", locals.MxPriority)),
				Exchange:   pulumi.String(value),
			})
		}
		record, err := dns.NewMxRecord(ctx,
			"dns-mx-record",
			&dns.MxRecordArgs{
				Name:              pulumi.String(recordName),
				ZoneName:          pulumi.String(zoneName),
				ResourceGroupName: pulumi.String(resourceGroup),
				Ttl:               pulumi.Int(ttl),
				Records:           mxRecords,
				Tags:              pulumi.ToStringMap(locals.AzureTags),
			},
			pulumi.Provider(azureProvider))
		if err != nil {
			return errors.Wrapf(err, "failed to create MX record %s", recordName)
		}
		recordId = record.ID()
		fqdn = record.Fqdn

	case azurednsrecordv1.AzureDnsRecordSpec_TXT:
		txtRecords := make(dns.TxtRecordRecordArray, 0)
		for _, value := range spec.Values {
			txtRecords = append(txtRecords, &dns.TxtRecordRecordArgs{
				Value: pulumi.String(value),
			})
		}
		record, err := dns.NewTxtRecord(ctx,
			"dns-txt-record",
			&dns.TxtRecordArgs{
				Name:              pulumi.String(recordName),
				ZoneName:          pulumi.String(zoneName),
				ResourceGroupName: pulumi.String(resourceGroup),
				Ttl:               pulumi.Int(ttl),
				Records:           txtRecords,
				Tags:              pulumi.ToStringMap(locals.AzureTags),
			},
			pulumi.Provider(azureProvider))
		if err != nil {
			return errors.Wrapf(err, "failed to create TXT record %s", recordName)
		}
		recordId = record.ID()
		fqdn = record.Fqdn

	case azurednsrecordv1.AzureDnsRecordSpec_NS:
		record, err := dns.NewNsRecord(ctx,
			"dns-ns-record",
			&dns.NsRecordArgs{
				Name:              pulumi.String(recordName),
				ZoneName:          pulumi.String(zoneName),
				ResourceGroupName: pulumi.String(resourceGroup),
				Ttl:               pulumi.Int(ttl),
				Records:           pulumi.ToStringArray(spec.Values),
				Tags:              pulumi.ToStringMap(locals.AzureTags),
			},
			pulumi.Provider(azureProvider))
		if err != nil {
			return errors.Wrapf(err, "failed to create NS record %s", recordName)
		}
		recordId = record.ID()
		fqdn = record.Fqdn

	case azurednsrecordv1.AzureDnsRecordSpec_SRV:
		srvRecords := make(dns.SrvRecordRecordArray, 0)
		for _, value := range spec.Values {
			// SRV records should be in format "priority weight port target"
			// For simplicity, we'll use defaults
			srvRecords = append(srvRecords, &dns.SrvRecordRecordArgs{
				Priority: pulumi.Int(10),
				Weight:   pulumi.Int(10),
				Port:     pulumi.Int(80),
				Target:   pulumi.String(value),
			})
		}
		record, err := dns.NewSrvRecord(ctx,
			"dns-srv-record",
			&dns.SrvRecordArgs{
				Name:              pulumi.String(recordName),
				ZoneName:          pulumi.String(zoneName),
				ResourceGroupName: pulumi.String(resourceGroup),
				Ttl:               pulumi.Int(ttl),
				Records:           srvRecords,
				Tags:              pulumi.ToStringMap(locals.AzureTags),
			},
			pulumi.Provider(azureProvider))
		if err != nil {
			return errors.Wrapf(err, "failed to create SRV record %s", recordName)
		}
		recordId = record.ID()
		fqdn = record.Fqdn

	case azurednsrecordv1.AzureDnsRecordSpec_CAA:
		caaRecords := make(dns.CaaRecordRecordArray, 0)
		for _, value := range spec.Values {
			caaRecords = append(caaRecords, &dns.CaaRecordRecordArgs{
				Flags: pulumi.Int(0),
				Tag:   pulumi.String("issue"),
				Value: pulumi.String(value),
			})
		}
		record, err := dns.NewCaaRecord(ctx,
			"dns-caa-record",
			&dns.CaaRecordArgs{
				Name:              pulumi.String(recordName),
				ZoneName:          pulumi.String(zoneName),
				ResourceGroupName: pulumi.String(resourceGroup),
				Ttl:               pulumi.Int(ttl),
				Records:           caaRecords,
				Tags:              pulumi.ToStringMap(locals.AzureTags),
			},
			pulumi.Provider(azureProvider))
		if err != nil {
			return errors.Wrapf(err, "failed to create CAA record %s", recordName)
		}
		recordId = record.ID()
		fqdn = record.Fqdn

	case azurednsrecordv1.AzureDnsRecordSpec_PTR:
		record, err := dns.NewPtrRecord(ctx,
			"dns-ptr-record",
			&dns.PtrRecordArgs{
				Name:              pulumi.String(recordName),
				ZoneName:          pulumi.String(zoneName),
				ResourceGroupName: pulumi.String(resourceGroup),
				Ttl:               pulumi.Int(ttl),
				Records:           pulumi.ToStringArray(spec.Values),
				Tags:              pulumi.ToStringMap(locals.AzureTags),
			},
			pulumi.Provider(azureProvider))
		if err != nil {
			return errors.Wrapf(err, "failed to create PTR record %s", recordName)
		}
		recordId = record.ID()
		fqdn = record.Fqdn

	default:
		return errors.Errorf("unsupported DNS record type: %s", recordType.String())
	}

	// Export stack outputs
	ctx.Export(OpRecordId, recordId)
	ctx.Export(OpFqdn, fqdn)

	return nil
}
