package module

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-digitalocean/sdk/v4/go/digitalocean"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// dnsZone provisions the DigitalOcean domain plus all DNS records
// and exports stack outputs.
func dnsZone(
	ctx *pulumi.Context,
	locals *Locals,
	digitalOceanProvider *digitalocean.Provider,
) (*digitalocean.Domain, error) {
	// 1. Create the DNS zone (Domain).
	domainArgs := &digitalocean.DomainArgs{
		Name: pulumi.String(locals.DigitalOceanDnsZone.Spec.DomainName),
	}

	createdDomain, err := digitalocean.NewDomain(
		ctx,
		"dns_zone",
		domainArgs,
		pulumi.Provider(digitalOceanProvider),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create digitalocean domain")
	}

	// 2. Create DNS records—one Record per value (simple mapping).
	for recIdx, rec := range locals.DigitalOceanDnsZone.Spec.Records {
		ttl := int(rec.TtlSeconds)
		if ttl == 0 {
			ttl = 3600
		}

		for valIdx, val := range rec.Values {
			resourceName := fmt.Sprintf("%s-%d-%d", rec.Name, recIdx, valIdx)

			// Build the DnsRecordArgs with basic fields
			recordArgs := &digitalocean.DnsRecordArgs{
				Domain: createdDomain.Name,
				Name:   pulumi.String(rec.Name),
				Type:   pulumi.String(rec.Type.String()),
				Value:  pulumi.String(val.GetValue()),
				Ttl:    pulumi.Int(ttl),
			}

			// Add priority for MX and SRV records
			if rec.Type.String() == "MX" || rec.Type.String() == "SRV" {
				recordArgs.Priority = pulumi.Int(int(rec.Priority))
			}

			// Add weight and port for SRV records
			if rec.Type.String() == "SRV" {
				recordArgs.Weight = pulumi.Int(int(rec.Weight))
				recordArgs.Port = pulumi.Int(int(rec.Port))
			}

			// Add flags and tag for CAA records
			if rec.Type.String() == "CAA" {
				recordArgs.Flags = pulumi.Int(int(rec.Flags))
				recordArgs.Tag = pulumi.String(rec.Tag)
			}

			// Note: StringValueOrRef has multiple fields; we assume 'Value' carries the literal.
			createdDnsRecord, err := digitalocean.NewDnsRecord(
				ctx,
				resourceName,
				recordArgs,
				pulumi.Provider(digitalOceanProvider),
			)
			if err != nil {
				return nil, errors.Wrapf(err, "failed to create dns record %s", resourceName)
			}

			// Lint‑appeasing reference so the variable prefix rule is met.
			_ = createdDnsRecord
		}
	}

	// 3. Export stack outputs.
	ctx.Export(OpZoneName, pulumi.String(locals.DigitalOceanDnsZone.Spec.DomainName))
	ctx.Export(OpZoneId, createdDomain.ID())
	ctx.Export(
		OpNameServers,
		pulumi.StringArray{
			pulumi.String("ns1.digitalocean.com"),
			pulumi.String("ns2.digitalocean.com"),
			pulumi.String("ns3.digitalocean.com"),
		},
	)

	return createdDomain, nil
}
