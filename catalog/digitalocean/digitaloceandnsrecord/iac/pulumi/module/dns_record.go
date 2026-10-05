package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-digitalocean/sdk/v4/go/digitalocean"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// dnsRecord provisions a single DigitalOcean DNS record and exports
// outputs. The per-type fields (priority/weight/port/flags/tag) carry the
// spec's presence semantics: unset stays unset, matching the provider's own
// GetOk-guarded request building. Spec CEL rules already guarantee the fields
// each record type requires are present.
func dnsRecord(
	ctx *pulumi.Context,
	locals *Locals,
	digitalOceanProvider *digitalocean.Provider,
) error {
	spec := locals.DigitalOceanDnsRecord.Spec

	// The orchestrator resolves valueFrom references to literal values before
	// the module runs, so GetValue() always carries the final string here.
	// The spec's enum value names ARE the DigitalOcean record types.
	recordArgs := &digitalocean.DnsRecordArgs{
		Domain: pulumi.String(spec.Domain.GetValue()),
		Name:   pulumi.String(spec.Name),
		Type:   pulumi.String(spec.Type.String()),
		Value:  pulumi.String(spec.Value.GetValue()),
	}

	// When unset, the ttl attribute is Computed: DigitalOcean applies its
	// default (1800 seconds) and the applied value reads back into state.
	if spec.TtlSeconds != nil {
		recordArgs.Ttl = pulumi.IntPtr(int(*spec.TtlSeconds))
	}
	if spec.Priority != nil {
		recordArgs.Priority = pulumi.IntPtr(int(*spec.Priority))
	}
	if spec.Weight != nil {
		recordArgs.Weight = pulumi.IntPtr(int(*spec.Weight))
	}
	if spec.Port != nil {
		recordArgs.Port = pulumi.IntPtr(int(*spec.Port))
	}
	if spec.Flags != nil {
		recordArgs.Flags = pulumi.IntPtr(int(*spec.Flags))
	}
	if spec.Tag != "" {
		recordArgs.Tag = pulumi.StringPtr(spec.Tag)
	}

	createdRecord, err := digitalocean.NewDnsRecord(
		ctx,
		"dns_record",
		recordArgs,
		pulumi.Provider(digitalOceanProvider),
	)
	if err != nil {
		return errors.Wrap(err, "failed to create digitalocean dns record")
	}

	// Outputs come from the created resource (never recomputed locally) so
	// both engines export identical values: fqdn is the provider's computed
	// hostname and ttl carries the API default when the spec left it unset.
	ctx.Export(OpRecordId, createdRecord.ID())
	ctx.Export(OpHostname, createdRecord.Fqdn)
	ctx.Export(OpRecordType, createdRecord.Type)
	ctx.Export(OpDomain, createdRecord.Domain)
	ctx.Export(OpTtlSeconds, createdRecord.Ttl)

	return nil
}
