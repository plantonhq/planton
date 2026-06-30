package module

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-digitalocean/sdk/v4/go/digitalocean"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// dnsRecord provisions a single DigitalOcean DNS record and exports stack outputs.
func dnsRecord(
	ctx *pulumi.Context,
	locals *Locals,
	digitalOceanProvider *digitalocean.Provider,
) error {
	spec := locals.DigitalOceanDnsRecord.Spec

	// Extract domain and value from StringValueOrRef
	domain := spec.Domain.GetValue()
	value := spec.Value.GetValue()

	// Determine TTL - use default of 1800 if not specified
	ttl := int(spec.GetTtlSeconds())
	if ttl == 0 {
		ttl = 1800
	}

	// Build the DnsRecordArgs
	recordArgs := &digitalocean.DnsRecordArgs{
		Domain: pulumi.String(domain),
		Name:   pulumi.String(spec.Name),
		Type:   pulumi.String(spec.Type.String()),
		Value:  pulumi.String(value),
		Ttl:    pulumi.Int(ttl),
	}

	// Add priority for MX and SRV records
	if spec.Type.String() == "MX" || spec.Type.String() == "SRV" {
		recordArgs.Priority = pulumi.Int(int(spec.Priority))
	}

	// Add weight and port for SRV records
	if spec.Type.String() == "SRV" {
		recordArgs.Weight = pulumi.Int(int(spec.Weight))
		recordArgs.Port = pulumi.Int(int(spec.Port))
	}

	// Add flags and tag for CAA records
	if spec.Type.String() == "CAA" {
		recordArgs.Flags = pulumi.Int(int(spec.Flags))
		recordArgs.Tag = pulumi.String(spec.Tag)
	}

	// Create the DNS record
	createdRecord, err := digitalocean.NewDnsRecord(
		ctx,
		"dns_record",
		recordArgs,
		pulumi.Provider(digitalOceanProvider),
	)
	if err != nil {
		return errors.Wrap(err, "failed to create digitalocean dns record")
	}

	// Build hostname based on record name
	hostname := spec.Name
	if hostname == "@" {
		hostname = domain
	} else {
		hostname = fmt.Sprintf("%s.%s", spec.Name, domain)
	}

	// Export stack outputs
	ctx.Export(OpRecordId, createdRecord.ID())
	ctx.Export(OpHostname, pulumi.String(hostname))
	ctx.Export(OpRecordType, pulumi.String(spec.Type.String()))
	ctx.Export(OpDomain, pulumi.String(domain))
	ctx.Export(OpTtlSeconds, pulumi.Int(ttl))

	return nil
}
