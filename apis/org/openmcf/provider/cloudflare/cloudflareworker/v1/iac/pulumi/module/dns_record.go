package module

import (
	"github.com/pkg/errors"
	cloudfl "github.com/pulumi/pulumi-cloudflare/sdk/v6/go/cloudflare"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// createDnsRecord creates an A record for the worker hostname if DNS is configured.
// The record is created with proxy (orange cloud) enabled so requests hit Cloudflare edge.
func createDnsRecord(
	ctx *pulumi.Context,
	locals *Locals,
	cloudflareProvider *cloudfl.Provider,
	zoneId pulumi.StringOutput,
) (*cloudfl.DnsRecord, error) {

	// Check if DNS configuration is provided and enabled
	if locals.CloudflareWorker.Spec.Dns == nil || !locals.CloudflareWorker.Spec.Dns.Enabled {
		// No DNS configuration or explicitly disabled
		return nil, nil
	}

	dns := locals.CloudflareWorker.Spec.Dns

	// Validate hostname is provided
	if dns.Hostname == "" {
		return nil, errors.New("dns.hostname is required when dns is enabled")
	}

	// Create an AAAA record pointed at a dummy address. The address is never used
	// because the record is proxied (orange cloud) and the Worker handles all
	// requests at the edge.
	recordArgs := &cloudfl.DnsRecordArgs{
		ZoneId:  zoneId.ToStringOutput(),
		Name:    pulumi.String(dns.Hostname),
		Type:    pulumi.String("AAAA"),
		Content: pulumi.String("100::"), // Dummy IPv6 - not used due to proxying
		Proxied: pulumi.Bool(true),      // Orange cloud - routes through Cloudflare
		Ttl:     pulumi.Float64(1),      // TTL is automatic when proxied
	}

	createdRecord, err := cloudfl.NewDnsRecord(
		ctx,
		"dns-record",
		recordArgs,
		pulumi.Provider(cloudflareProvider),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create cloudflare dns record")
	}

	return createdRecord, nil
}
