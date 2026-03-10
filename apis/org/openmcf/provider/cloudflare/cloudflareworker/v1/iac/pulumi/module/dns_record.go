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
) (*cloudfl.Record, error) {

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

	// Create A record with a dummy IP (100.0.0.1)
	// The IP doesn't matter because the Worker will handle all requests at the edge
	// The key is that the record must be proxied (orange cloud) so traffic hits Cloudflare
	recordArgs := &cloudfl.RecordArgs{
		ZoneId:  zoneId.ToStringOutput(),
		Name:    pulumi.String(dns.Hostname),
		Type:    pulumi.String("A"),
		Content: pulumi.String("100.0.0.1"), // Dummy IP - not used due to proxying
		Proxied: pulumi.Bool(true),          // Orange cloud - routes through Cloudflare
		Comment: pulumi.String("Managed by Planton - Routes to Cloudflare Worker"),
		Ttl:     pulumi.Float64(1), // TTL is automatic when proxied
	}

	createdRecord, err := cloudfl.NewRecord(
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
