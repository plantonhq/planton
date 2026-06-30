package module

import (
	"fmt"

	"github.com/pkg/errors"
	cloudflarednszonev1 "github.com/plantonhq/planton/apis/dev/planton/provider/cloudflare/cloudflarednszone/v1"
	"github.com/pulumi/pulumi-cloudflare/sdk/v6/go/cloudflare"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// records creates DNS records within the zone.
func records(
	ctx *pulumi.Context,
	zone *cloudflare.Zone,
	recordsList []*cloudflarednszonev1.CloudflareDnsZoneRecord,
	cloudflareProvider *cloudflare.Provider,
) error {
	for idx, record := range recordsList {
		// Include index to ensure uniqueness when multiple records have same name and type
		resourceName := fmt.Sprintf("%s-%s-%d", record.Name, record.Type.String(), idx)

		ttl := float64(1)
		if record.Ttl > 0 {
			ttl = float64(record.Ttl)
		}

		recordArgs := &cloudflare.DnsRecordArgs{
			ZoneId:  zone.ID(),
			Name:    pulumi.String(record.Name),
			Type:    pulumi.String(record.Type.String()),
			Content: pulumi.String(record.Content),
			Ttl:     pulumi.Float64(ttl),
		}

		// proxied is only applicable to A, AAAA, and CNAME records
		if record.Type == cloudflarednszonev1.CloudflareDnsZoneRecord_A ||
			record.Type == cloudflarednszonev1.CloudflareDnsZoneRecord_AAAA ||
			record.Type == cloudflarednszonev1.CloudflareDnsZoneRecord_CNAME {
			recordArgs.Proxied = pulumi.Bool(record.Proxied)
		}

		// priority is only used for MX and SRV records
		if record.Type == cloudflarednszonev1.CloudflareDnsZoneRecord_MX ||
			record.Type == cloudflarednszonev1.CloudflareDnsZoneRecord_SRV {
			recordArgs.Priority = pulumi.Float64Ptr(float64(record.Priority))
		}

		// comment for the DNS record
		if record.Comment != "" {
			recordArgs.Comment = pulumi.String(record.Comment)
		}

		_, err := cloudflare.NewDnsRecord(
			ctx,
			resourceName,
			recordArgs,
			pulumi.Provider(cloudflareProvider),
			pulumi.DependsOn([]pulumi.Resource{zone}),
		)
		if err != nil {
			return errors.Wrapf(err, "failed to create dns record %s", resourceName)
		}
	}
	return nil
}
