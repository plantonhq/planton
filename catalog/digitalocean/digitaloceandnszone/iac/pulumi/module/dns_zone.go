package module

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-digitalocean/sdk/v4/go/digitalocean"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// dnsZone provisions the DigitalOcean domain plus its managed DNS records and
// exports stack outputs.
func dnsZone(
	ctx *pulumi.Context,
	locals *Locals,
	digitalOceanProvider *digitalocean.Provider,
) (*digitalocean.Domain, error) {
	spec := locals.DigitalOceanDnsZone.Spec

	domainArgs := &digitalocean.DomainArgs{
		Name: pulumi.String(spec.DomainName),
	}
	// ip_address is a create-only convenience that seeds an initial apex A
	// record DigitalOcean never tracks afterwards — prefer declaring records.
	if spec.IpAddress != "" {
		domainArgs.IpAddress = pulumi.StringPtr(spec.IpAddress)
	}

	createdDomain, err := digitalocean.NewDomain(
		ctx,
		"dns_zone",
		domainArgs,
		pulumi.Provider(digitalOceanProvider),
		// ip_address is applied at creation ONLY and never read back by the
		// API, and the provider marks it ForceNew. Left unguarded, adopting an
		// existing zone whose manifest still carries it -- or editing it later
		// -- would plan a destroy-and-recreate of the whole zone, taking every
		// record and the domain's resolution with it. The seed record it
		// created keeps living in the zone regardless, so later changes are
		// ignored here, exactly as the Terraform module's
		// lifecycle.ignore_changes does; the spec field comment tells manifest
		// authors the same.
		pulumi.IgnoreChanges([]string{"ipAddress"}),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create digitalocean domain")
	}

	// One DigitalOcean record per record value (same name and type). The
	// per-type fields carry the spec's presence semantics: unset stays
	// unset, matching the provider's own GetOk-guarded request building.
	// Record ids are collected under the SAME key the Terraform module uses
	// for its for_each (<record name>-<record index>-<value index>) -- the
	// import recipes resolve per-record ids through these keys, so the two
	// engines must agree on them.
	//
	// Records are created ONE AT A TIME through an explicit dependency
	// chain: DigitalOcean's DNS API deadlocks on concurrent record writes to
	// one domain ("422 Error 1213 (40001): Deadlock found when trying to
	// get lock" -- live-verified on 1 in 8 parallel creates, fresh or
	// settled domain alike), and the provider does not retry 422s. The
	// chain also serializes the destroy, which walks it in reverse. The
	// Terraform module serializes the same writes with a one-request-per-
	// second client limit, the only lever a for_each resource offers.
	recordIds := pulumi.StringMap{}
	var previousRecord pulumi.Resource
	for recIdx, rec := range spec.Records {
		for valIdx, val := range rec.Values {
			// The spec's shared enum value names ARE the DigitalOcean record
			// types (A, AAAA, CNAME, ...), so the type wires through directly.
			recordArgs := &digitalocean.DnsRecordArgs{
				Domain: createdDomain.Name,
				Name:   pulumi.String(rec.Name),
				Type:   pulumi.String(rec.Type.String()),
				Value:  pulumi.String(val.GetValue()),
			}

			// 0 means unset: the ttl attribute is then Computed and
			// DigitalOcean applies its default (1800 seconds).
			if rec.TtlSeconds > 0 {
				recordArgs.Ttl = pulumi.IntPtr(int(rec.TtlSeconds))
			}
			if rec.Priority != nil {
				recordArgs.Priority = pulumi.IntPtr(int(*rec.Priority))
			}
			if rec.Weight != nil {
				recordArgs.Weight = pulumi.IntPtr(int(*rec.Weight))
			}
			if rec.Port != nil {
				recordArgs.Port = pulumi.IntPtr(int(*rec.Port))
			}
			if rec.Flags != nil {
				recordArgs.Flags = pulumi.IntPtr(int(*rec.Flags))
			}
			if rec.Tag != "" {
				recordArgs.Tag = pulumi.StringPtr(rec.Tag)
			}

			resourceName := fmt.Sprintf("%s-%d-%d", rec.Name, recIdx, valIdx)
			recordOpts := []pulumi.ResourceOption{pulumi.Provider(digitalOceanProvider)}
			if previousRecord != nil {
				recordOpts = append(recordOpts, pulumi.DependsOn([]pulumi.Resource{previousRecord}))
			}
			createdRecord, err := digitalocean.NewDnsRecord(
				ctx,
				resourceName,
				recordArgs,
				recordOpts...,
			)
			if err != nil {
				return nil, errors.Wrapf(err, "failed to create dns record %s", resourceName)
			}
			previousRecord = createdRecord
			recordIds[resourceName] = createdRecord.ID().ToStringOutput()
		}
	}

	ctx.Export(OpZoneName, createdDomain.Name)
	ctx.Export(OpZoneId, createdDomain.ID())
	// DigitalOcean's authoritative name servers are a fixed platform-wide set
	// the API does not return per zone.
	ctx.Export(OpNameServers, pulumi.StringArray{
		pulumi.String("ns1.digitalocean.com"),
		pulumi.String("ns2.digitalocean.com"),
		pulumi.String("ns3.digitalocean.com"),
	})
	// The SDK renames the provider's urn attribute to domainUrn.
	ctx.Export(OpUrn, createdDomain.DomainUrn)
	ctx.Export(OpRecordIds, recordIds)

	return createdDomain, nil
}
