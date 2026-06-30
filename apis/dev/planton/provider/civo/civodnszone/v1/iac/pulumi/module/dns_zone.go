package module

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-civo/sdk/v2/go/civo"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// dnsZone provisions the Civo DNS domain plus associated records
// and exports stack outputs.
func dnsZone(
	ctx *pulumi.Context,
	locals *Locals,
	civoProvider *civo.Provider,
) (*civo.DnsDomainName, error) {
	// 1. Create the DNS zone (domain).
	createdDomain, err := civo.NewDnsDomainName(
		ctx,
		"dns_zone",
		&civo.DnsDomainNameArgs{
			Name: pulumi.String(locals.CivoDnsZone.Spec.DomainName),
		},
		pulumi.Provider(civoProvider),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create Civo DNS domain")
	}

	// 2. Create DNS records — one Record per value (simple mapping).
	for recIdx, rec := range locals.CivoDnsZone.Spec.Records {
		ttl := int(rec.TtlSeconds)
		if ttl == 0 {
			ttl = 3600
		}

		for valIdx, val := range rec.Values {
			resourceName := fmt.Sprintf("%s-%d-%d", rec.Name, recIdx, valIdx)

			// Note: StringValueOrRef has multiple fields; here we assume the literal lives in 'Value'.
			_, err := civo.NewDnsDomainRecord(
				ctx,
				resourceName,
				&civo.DnsDomainRecordArgs{
					DomainId: createdDomain.ID(),
					Name:     pulumi.String(rec.Name),
					Type:     pulumi.String(rec.Type.String()),
					Value:    pulumi.String(val.GetValue()),
					Ttl:      pulumi.Int(ttl),
				},
				pulumi.Provider(civoProvider),
			)
			if err != nil {
				return nil, errors.Wrapf(err, "failed to create DNS record %s", resourceName)
			}
		}
	}

	// 3. Export stack outputs.
	ctx.Export(OpZoneName, pulumi.String(locals.CivoDnsZone.Spec.DomainName))
	ctx.Export(OpZoneId, createdDomain.ID())
	ctx.Export(
		OpNameServers,
		pulumi.StringArray{
			pulumi.String("ns0.civo.com"),
			pulumi.String("ns1.civo.com"),
			pulumi.String("ns2.civo.com"),
		},
	)

	return createdDomain, nil
}
