package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	scalewayv2 "github.com/pulumiverse/pulumi-scaleway/sdk/go/scaleway"
	"github.com/pulumiverse/pulumi-scaleway/sdk/go/scaleway/domain"
)

// dnsRecordTypeMap maps the local RecordType enum string values
// to the record type strings expected by the Scaleway API.
// Includes all 13 Scaleway-supported record types.
var dnsRecordTypeMap = map[string]string{
	"A":     "A",
	"AAAA":  "AAAA",
	"ALIAS": "ALIAS",
	"CAA":   "CAA",
	"CNAME": "CNAME",
	"DNAME": "DNAME",
	"MX":    "MX",
	"NS":    "NS",
	"PTR":   "PTR",
	"SOA":   "SOA",
	"SRV":   "SRV",
	"TXT":   "TXT",
	"TLSA":  "TLSA",
}

// dnsRecord provisions a single Scaleway DNS record and exports
// stack outputs.
//
// Uses domain.NewRecord from the scaleway/domain subpackage
// (the preferred API path in the pulumiverse SDK v1.43.0).
func dnsRecord(
	ctx *pulumi.Context,
	locals *Locals,
	scalewayProvider *scalewayv2.Provider,
) error {
	spec := locals.ScalewayDnsRecord.Spec

	// Resolve StringValueOrRef fields.
	zoneName := spec.ZoneName.GetValue()
	data := spec.Data.GetValue()

	// Resolve record type from the proto enum to the Scaleway API string.
	recordType := spec.Type.String()
	if mapped, ok := dnsRecordTypeMap[recordType]; ok {
		recordType = mapped
	}

	// Determine TTL -- use 3600 default if not specified.
	ttl := int(spec.Ttl)
	if ttl == 0 {
		ttl = 3600
	}

	// Build record arguments.
	recordArgs := &domain.RecordArgs{
		DnsZone: pulumi.String(zoneName),
		Name:    pulumi.String(spec.Name),
		Type:    pulumi.String(recordType),
		Data:    pulumi.String(data),
		Ttl:     pulumi.IntPtr(ttl),
	}

	// Set priority for MX and SRV records.
	if spec.Priority > 0 {
		recordArgs.Priority = pulumi.IntPtr(int(spec.Priority))
	}

	// NOTE: keep_empty_zone is a Terraform-only feature. The Pulumi
	// Scaleway SDK does not expose this field on RecordArgs. The
	// Terraform module handles it; the Pulumi module cannot.

	// Create the DNS record.
	createdRecord, err := domain.NewRecord(
		ctx,
		"dns_record",
		recordArgs,
		pulumi.Provider(scalewayProvider),
	)
	if err != nil {
		return errors.Wrap(err, "failed to create scaleway dns record")
	}

	// Export stack outputs.
	ctx.Export(OpRecordId, createdRecord.ID())
	ctx.Export(OpFqdn, createdRecord.Fqdn)

	return nil
}
