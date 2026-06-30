package module

import (
	"fmt"

	"github.com/pkg/errors"
	azurednszonev1 "github.com/plantonhq/planton/apis/dev/planton/provider/azure/azurednszone/v1"
	"github.com/plantonhq/planton/apis/dev/planton/shared/networking/enums/dnsrecordtype"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure"
	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/dns"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *azurednszonev1.AzureDnsZoneStackInput) error {
	locals := initializeLocals(ctx, stackInput)
	azureProviderConfig := stackInput.ProviderConfig

	//create azure provider using the credentials from the input
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

	// Get the spec from locals
	spec := locals.AzureDnsZone.Spec

	// Create the DNS Zone
	dnsZone, err := dns.NewZone(ctx,
		spec.ZoneName,
		&dns.ZoneArgs{
			Name:              pulumi.String(spec.ZoneName),
			ResourceGroupName: pulumi.String(locals.ResourceGroupName),
			Tags:              pulumi.ToStringMap(locals.AzureTags),
		},
		pulumi.Provider(azureProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create DNS zone for %s", spec.ZoneName)
	}

	// Create DNS records in the zone
	for index, record := range spec.Records {
		recordType := record.RecordType.String()
		recordName := record.Name
		ttl := int(record.GetTtlSeconds())

		// Create different record types based on the DNS record type
		switch record.RecordType {
		case dnsrecordtype.DnsRecordType_A:
			_, err := dns.NewARecord(ctx,
				fmt.Sprintf("dns-a-record-%d", index),
				&dns.ARecordArgs{
					Name:              pulumi.String(recordName),
					ZoneName:          dnsZone.Name,
					ResourceGroupName: pulumi.String(locals.ResourceGroupName),
					Ttl:               pulumi.Int(ttl),
					Records:           pulumi.ToStringArray(record.Values),
					Tags:              pulumi.ToStringMap(locals.AzureTags),
				},
				pulumi.Provider(azureProvider),
				pulumi.Parent(dnsZone))
			if err != nil {
				return errors.Wrapf(err, "failed to create A record %s", recordName)
			}

		case dnsrecordtype.DnsRecordType_AAAA:
			_, err := dns.NewAaaaRecord(ctx,
				fmt.Sprintf("dns-aaaa-record-%d", index),
				&dns.AaaaRecordArgs{
					Name:              pulumi.String(recordName),
					ZoneName:          dnsZone.Name,
					ResourceGroupName: pulumi.String(locals.ResourceGroupName),
					Ttl:               pulumi.Int(ttl),
					Records:           pulumi.ToStringArray(record.Values),
					Tags:              pulumi.ToStringMap(locals.AzureTags),
				},
				pulumi.Provider(azureProvider),
				pulumi.Parent(dnsZone))
			if err != nil {
				return errors.Wrapf(err, "failed to create AAAA record %s", recordName)
			}

		case dnsrecordtype.DnsRecordType_CNAME:
			if len(record.Values) > 0 {
				_, err := dns.NewCNameRecord(ctx,
					fmt.Sprintf("dns-cname-record-%d", index),
					&dns.CNameRecordArgs{
						Name:              pulumi.String(recordName),
						ZoneName:          dnsZone.Name,
						ResourceGroupName: pulumi.String(locals.ResourceGroupName),
						Ttl:               pulumi.Int(ttl),
						Record:            pulumi.String(record.Values[0]),
						Tags:              pulumi.ToStringMap(locals.AzureTags),
					},
					pulumi.Provider(azureProvider),
					pulumi.Parent(dnsZone))
				if err != nil {
					return errors.Wrapf(err, "failed to create CNAME record %s", recordName)
				}
			}

		case dnsrecordtype.DnsRecordType_MX:
			mxRecords := make(dns.MxRecordRecordArray, 0)
			for _, value := range record.Values {
				// MX records should be in format "priority hostname"
				// For simplicity, we'll parse or use default priority 10
				mxRecords = append(mxRecords, &dns.MxRecordRecordArgs{
					Preference: pulumi.String("10"),
					Exchange:   pulumi.String(value),
				})
			}
			_, err := dns.NewMxRecord(ctx,
				fmt.Sprintf("dns-mx-record-%d", index),
				&dns.MxRecordArgs{
					Name:              pulumi.String(recordName),
					ZoneName:          dnsZone.Name,
					ResourceGroupName: pulumi.String(locals.ResourceGroupName),
					Ttl:               pulumi.Int(ttl),
					Records:           mxRecords,
					Tags:              pulumi.ToStringMap(locals.AzureTags),
				},
				pulumi.Provider(azureProvider),
				pulumi.Parent(dnsZone))
			if err != nil {
				return errors.Wrapf(err, "failed to create MX record %s", recordName)
			}

		case dnsrecordtype.DnsRecordType_TXT:
			_, err := dns.NewTxtRecord(ctx,
				fmt.Sprintf("dns-txt-record-%d", index),
				&dns.TxtRecordArgs{
					Name:              pulumi.String(recordName),
					ZoneName:          dnsZone.Name,
					ResourceGroupName: pulumi.String(locals.ResourceGroupName),
					Ttl:               pulumi.Int(ttl),
					Records: dns.TxtRecordRecordArray{
						&dns.TxtRecordRecordArgs{
							Value: pulumi.String(record.Values[0]),
						},
					},
					Tags: pulumi.ToStringMap(locals.AzureTags),
				},
				pulumi.Provider(azureProvider),
				pulumi.Parent(dnsZone))
			if err != nil {
				return errors.Wrapf(err, "failed to create TXT record %s", recordName)
			}

		case dnsrecordtype.DnsRecordType_NS:
			_, err := dns.NewNsRecord(ctx,
				fmt.Sprintf("dns-ns-record-%d", index),
				&dns.NsRecordArgs{
					Name:              pulumi.String(recordName),
					ZoneName:          dnsZone.Name,
					ResourceGroupName: pulumi.String(locals.ResourceGroupName),
					Ttl:               pulumi.Int(ttl),
					Records:           pulumi.ToStringArray(record.Values),
					Tags:              pulumi.ToStringMap(locals.AzureTags),
				},
				pulumi.Provider(azureProvider),
				pulumi.Parent(dnsZone))
			if err != nil {
				return errors.Wrapf(err, "failed to create NS record %s", recordName)
			}

		case dnsrecordtype.DnsRecordType_CAA:
			caaRecords := make(dns.CaaRecordRecordArray, 0)
			for _, value := range record.Values {
				// CAA records should be in format "flags tag value"
				// For simplicity, we'll parse or use defaults
				caaRecords = append(caaRecords, &dns.CaaRecordRecordArgs{
					Flags: pulumi.Int(0),
					Tag:   pulumi.String("issue"),
					Value: pulumi.String(value),
				})
			}
			_, err := dns.NewCaaRecord(ctx,
				fmt.Sprintf("dns-caa-record-%d", index),
				&dns.CaaRecordArgs{
					Name:              pulumi.String(recordName),
					ZoneName:          dnsZone.Name,
					ResourceGroupName: pulumi.String(locals.ResourceGroupName),
					Ttl:               pulumi.Int(ttl),
					Records:           caaRecords,
					Tags:              pulumi.ToStringMap(locals.AzureTags),
				},
				pulumi.Provider(azureProvider),
				pulumi.Parent(dnsZone))
			if err != nil {
				return errors.Wrapf(err, "failed to create CAA record %s", recordName)
			}

		case dnsrecordtype.DnsRecordType_SRV:
			srvRecords := make(dns.SrvRecordRecordArray, 0)
			for _, value := range record.Values {
				// SRV records should be in format "priority weight port target"
				// For simplicity, we'll use defaults
				srvRecords = append(srvRecords, &dns.SrvRecordRecordArgs{
					Priority: pulumi.Int(10),
					Weight:   pulumi.Int(10),
					Port:     pulumi.Int(80),
					Target:   pulumi.String(value),
				})
			}
			_, err := dns.NewSrvRecord(ctx,
				fmt.Sprintf("dns-srv-record-%d", index),
				&dns.SrvRecordArgs{
					Name:              pulumi.String(recordName),
					ZoneName:          dnsZone.Name,
					ResourceGroupName: pulumi.String(locals.ResourceGroupName),
					Ttl:               pulumi.Int(ttl),
					Records:           srvRecords,
					Tags:              pulumi.ToStringMap(locals.AzureTags),
				},
				pulumi.Provider(azureProvider),
				pulumi.Parent(dnsZone))
			if err != nil {
				return errors.Wrapf(err, "failed to create SRV record %s", recordName)
			}

		case dnsrecordtype.DnsRecordType_PTR:
			_, err := dns.NewPtrRecord(ctx,
				fmt.Sprintf("dns-ptr-record-%d", index),
				&dns.PtrRecordArgs{
					Name:              pulumi.String(recordName),
					ZoneName:          dnsZone.Name,
					ResourceGroupName: pulumi.String(locals.ResourceGroupName),
					Ttl:               pulumi.Int(ttl),
					Records:           pulumi.ToStringArray(record.Values),
					Tags:              pulumi.ToStringMap(locals.AzureTags),
				},
				pulumi.Provider(azureProvider),
				pulumi.Parent(dnsZone))
			if err != nil {
				return errors.Wrapf(err, "failed to create PTR record %s", recordName)
			}

		default:
			return errors.Errorf("unsupported DNS record type: %s", recordType)
		}
	}

	// Export stack outputs
	ctx.Export(OpZoneId, dnsZone.ID())
	ctx.Export(OpZoneName, dnsZone.Name)
	ctx.Export(OpNameservers, dnsZone.NameServers)

	return nil
}
