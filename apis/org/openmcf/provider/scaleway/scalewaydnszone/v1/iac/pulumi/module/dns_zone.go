package module

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	scalewayv2 "github.com/pulumiverse/pulumi-scaleway/sdk/go/scaleway"
	"github.com/pulumiverse/pulumi-scaleway/sdk/go/scaleway/domain"
)

// dnsRecordTypeMap maps the local RecordType enum string values
// to the uppercase record type strings expected by the Scaleway API.
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

// dnsZone provisions the Scaleway DNS zone, creates inline DNS records,
// and exports stack outputs.
//
// Uses the domain.NewZone and domain.NewRecord functions from the
// scaleway/domain subpackage (the preferred API path in the pulumiverse
// SDK v1.43.0).
func dnsZone(
	ctx *pulumi.Context,
	locals *Locals,
	scalewayProvider *scalewayv2.Provider,
) error {
	spec := locals.ScalewayDnsZone.Spec

	// ── 1. Create the DNS zone ──────────────────────────────────────

	zoneArgs := &domain.ZoneArgs{
		Domain:    pulumi.String(spec.Domain),
		Subdomain: pulumi.String(spec.Subdomain),
	}

	createdZone, err := domain.NewZone(
		ctx,
		"dns_zone",
		zoneArgs,
		pulumi.Provider(scalewayProvider),
	)
	if err != nil {
		return errors.Wrap(err, "failed to create scaleway dns zone")
	}

	// ── 2. Create inline DNS records ────────────────────────────────

	for idx, rec := range spec.Records {
		ttl := int(rec.Ttl)
		if ttl == 0 {
			ttl = 3600
		}

		// Resolve the record type from the proto enum to the Scaleway
		// API string. The proto enum uses uppercase names (A, AAAA, etc.)
		// which happen to match Scaleway's API directly.
		recordType := rec.Type.String()
		if mapped, ok := dnsRecordTypeMap[recordType]; ok {
			recordType = mapped
		}

		// Build a unique resource name from the record name and index.
		// Use the record name if available, otherwise use "apex" for
		// root records.
		recordLabel := rec.Name
		if recordLabel == "" || recordLabel == "@" {
			recordLabel = "apex"
		}
		resourceName := fmt.Sprintf("record-%s-%d", strings.ReplaceAll(recordLabel, ".", "-"), idx)

		recordArgs := &domain.RecordArgs{
			DnsZone: createdZone.ID().ApplyT(func(_ string) string {
				// Use the computed zone name rather than the zone ID
				// because scaleway_domain_record expects the zone name
				// (e.g., "example.com"), not the Terraform resource ID.
				return locals.ZoneName
			}).(pulumi.StringOutput),
			Name: pulumi.String(rec.Name),
			Type: pulumi.String(recordType),
			Data: pulumi.String(rec.Data.GetValue()),
			Ttl:  pulumi.IntPtr(ttl),
		}

		// Set priority for MX and SRV records.
		if rec.Priority > 0 {
			recordArgs.Priority = pulumi.IntPtr(int(rec.Priority))
		}

		createdRecord, err := domain.NewRecord(
			ctx,
			resourceName,
			recordArgs,
			pulumi.Provider(scalewayProvider),
			pulumi.DependsOn([]pulumi.Resource{createdZone}),
		)
		if err != nil {
			return errors.Wrapf(err, "failed to create dns record %s", resourceName)
		}

		// Reference to satisfy linter.
		_ = createdRecord
	}

	// ── 3. Export stack outputs ──────────────────────────────────────

	ctx.Export(OpZoneName, pulumi.String(locals.ZoneName))

	ctx.Export(OpNameServers, createdZone.Ns)
	ctx.Export(OpNameServersDefault, createdZone.NsDefaults)
	ctx.Export(OpNameServersMaster, createdZone.NsMasters)
	ctx.Export(OpStatus, createdZone.Status)

	return nil
}
