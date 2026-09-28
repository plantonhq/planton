package module

import (
	"strings"

	"github.com/pkg/errors"
	cloudflarednsrecordv1alpha1 "github.com/plantonhq/planton/catalog/cloudflare/cloudflarednsrecord/v1alpha1"
	"github.com/pulumi/pulumi-cloudflare/sdk/v6/go/cloudflare"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// dnsRecord provisions the Cloudflare DNS record and exports outputs.
func dnsRecord(
	ctx *pulumi.Context,
	locals *Locals,
	cloudflareProvider *cloudflare.Provider,
) (*cloudflare.DnsRecord, error) {
	spec := locals.CloudflareDnsRecord.Spec

	// Determine TTL (1 = automatic, or specified value).
	ttl := float64(1)
	if spec.Ttl > 0 {
		ttl = float64(spec.Ttl)
	}

	// Resolve zone_id from literal value or reference.
	zoneId := ""
	if spec.ZoneId != nil {
		zoneId = spec.ZoneId.GetValue()
	}

	recordType := strings.ToUpper(spec.Type.String())

	// Build the record arguments.
	recordArgs := &cloudflare.DnsRecordArgs{
		ZoneId:  pulumi.String(zoneId),
		Name:    pulumi.String(spec.Name),
		Type:    pulumi.String(recordType),
		Proxied: pulumi.Bool(spec.Proxied),
		Ttl:     pulumi.Float64(ttl),
	}

	// Simple record types carry their value in content -- a literal, or a
	// reference already resolved to its value -- and structured types use data.
	if content := spec.Content.GetValue(); content != "" {
		recordArgs.Content = pulumi.String(content)
	}
	if data := buildDnsRecordData(spec); data != nil {
		recordArgs.Data = data
	}

	// The provider schema marks top-level priority "Required for MX, SRV and
	// URI records". MX carries it in spec.priority; SRV/URI carry it inside
	// their structured data, and Cloudflare mirrors that value into the
	// top-level field on its own -- omitting it here drifts on refresh
	// (live-measured on the Terraform engine; same provider underneath).
	switch {
	case spec.Type == cloudflarednsrecordv1alpha1.CloudflareDnsRecordSpec_MX:
		recordArgs.Priority = pulumi.Float64(float64(spec.Priority))
	case spec.GetSrv() != nil:
		recordArgs.Priority = pulumi.Float64(float64(spec.GetSrv().Priority))
	case spec.GetUri() != nil:
		recordArgs.Priority = pulumi.Float64(float64(spec.GetUri().Priority))
	}

	if spec.Comment != "" {
		recordArgs.Comment = pulumi.String(spec.Comment)
	}

	if len(spec.Tags) > 0 {
		tags := make(pulumi.StringArray, 0, len(spec.Tags))
		for _, t := range spec.Tags {
			tags = append(tags, pulumi.String(t))
		}
		recordArgs.Tags = tags
	}

	if s := spec.Settings; s != nil {
		recordArgs.Settings = cloudflare.DnsRecordSettingsArgs{
			Ipv4Only:     pulumi.Bool(s.Ipv4Only),
			Ipv6Only:     pulumi.Bool(s.Ipv6Only),
			FlattenCname: pulumi.Bool(s.FlattenCname),
		}
	}

	if spec.PrivateRouting {
		recordArgs.PrivateRouting = pulumi.Bool(true)
	}

	// Create the DNS record using DnsRecord (Record is deprecated in v6).
	createdRecord, err := cloudflare.NewDnsRecord(
		ctx,
		strings.ToLower(locals.CloudflareDnsRecord.Metadata.Name),
		recordArgs,
		pulumi.Provider(cloudflareProvider),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create cloudflare dns record")
	}

	// Export required outputs. record_name is deliberately the DECLARED name:
	// Cloudflare normalizes names to the full FQDN on read, so exporting the
	// resource attribute flips the output after the first refresh (the
	// Terraform engine's apply-idempotency re-plan measured this live; both
	// engines keep the declared-value contract for parity).
	ctx.Export(OpRecordId, createdRecord.ID())
	ctx.Export(OpRecordName, pulumi.String(spec.Name))
	ctx.Export(OpRecordType, pulumi.String(recordType))
	ctx.Export(OpProxied, createdRecord.Proxied)
	ctx.Export(OpZoneId, createdRecord.ZoneId)

	return createdRecord, nil
}

// buildDnsRecordData translates the typed `data` oneof into the provider's flat
// data object, returning nil when the record is a simple (content) record.
func buildDnsRecordData(spec *cloudflarednsrecordv1alpha1.CloudflareDnsRecordSpec) cloudflare.DnsRecordDataPtrInput {
	f64 := func(v uint32) pulumi.Float64PtrInput { return pulumi.Float64Ptr(float64(v)) }

	switch {
	case spec.GetCaa() != nil:
		d := spec.GetCaa()
		return cloudflare.DnsRecordDataArgs{
			Flags: pulumi.Float64(float64(d.Flags)),
			Tag:   pulumi.String(d.Tag),
			Value: pulumi.String(d.Value),
		}
	case spec.GetCert() != nil:
		d := spec.GetCert()
		return cloudflare.DnsRecordDataArgs{
			Type:        f64(d.Type),
			KeyTag:      f64(d.KeyTag),
			Algorithm:   f64(d.Algorithm),
			Certificate: pulumi.String(d.Certificate),
		}
	case spec.GetDnskey() != nil:
		d := spec.GetDnskey()
		return cloudflare.DnsRecordDataArgs{
			Flags:     pulumi.Float64(float64(d.Flags)),
			Protocol:  f64(d.Protocol),
			Algorithm: f64(d.Algorithm),
			PublicKey: pulumi.String(d.PublicKey),
		}
	case spec.GetDs() != nil:
		d := spec.GetDs()
		return cloudflare.DnsRecordDataArgs{
			KeyTag:     f64(d.KeyTag),
			Algorithm:  f64(d.Algorithm),
			DigestType: f64(d.DigestType),
			Digest:     pulumi.String(d.Digest),
		}
	case spec.GetHttps() != nil:
		d := spec.GetHttps()
		return cloudflare.DnsRecordDataArgs{
			Priority: f64(d.Priority),
			Target:   pulumi.String(d.Target.GetValue()),
			Value:    pulumi.String(d.Value),
		}
	case spec.GetLoc() != nil:
		d := spec.GetLoc()
		return cloudflare.DnsRecordDataArgs{
			LatDirection:  pulumi.String(d.LatDirection),
			LatDegrees:    f64(d.LatDegrees),
			LatMinutes:    f64(d.LatMinutes),
			LatSeconds:    pulumi.Float64Ptr(d.LatSeconds),
			LongDirection: pulumi.String(d.LongDirection),
			LongDegrees:   f64(d.LongDegrees),
			LongMinutes:   f64(d.LongMinutes),
			LongSeconds:   pulumi.Float64Ptr(d.LongSeconds),
			Altitude:      pulumi.Float64Ptr(d.Altitude),
			Size:          pulumi.Float64Ptr(d.Size),
			PrecisionHorz: pulumi.Float64Ptr(d.PrecisionHorz),
			PrecisionVert: pulumi.Float64Ptr(d.PrecisionVert),
		}
	case spec.GetNaptr() != nil:
		d := spec.GetNaptr()
		return cloudflare.DnsRecordDataArgs{
			Flags:       pulumi.String(d.Flags),
			Order:       f64(d.Order),
			Preference:  f64(d.Preference),
			Service:     pulumi.String(d.Service),
			Regex:       pulumi.String(d.Regex),
			Replacement: pulumi.String(d.Replacement),
		}
	case spec.GetSmimea() != nil:
		d := spec.GetSmimea()
		return cloudflare.DnsRecordDataArgs{
			Usage:        f64(d.Usage),
			Selector:     f64(d.Selector),
			MatchingType: f64(d.MatchingType),
			Certificate:  pulumi.String(d.Certificate),
		}
	case spec.GetSrv() != nil:
		d := spec.GetSrv()
		return cloudflare.DnsRecordDataArgs{
			Priority: f64(d.Priority),
			Weight:   f64(d.Weight),
			Port:     f64(d.Port),
			Target:   pulumi.String(d.Target.GetValue()),
		}
	case spec.GetSshfp() != nil:
		d := spec.GetSshfp()
		return cloudflare.DnsRecordDataArgs{
			Algorithm:   f64(d.Algorithm),
			Type:        f64(d.Type),
			Fingerprint: pulumi.String(d.Fingerprint),
		}
	case spec.GetSvcb() != nil:
		d := spec.GetSvcb()
		return cloudflare.DnsRecordDataArgs{
			Priority: f64(d.Priority),
			Target:   pulumi.String(d.Target.GetValue()),
			Value:    pulumi.String(d.Value),
		}
	case spec.GetTlsa() != nil:
		d := spec.GetTlsa()
		return cloudflare.DnsRecordDataArgs{
			Usage:        f64(d.Usage),
			Selector:     f64(d.Selector),
			MatchingType: f64(d.MatchingType),
			Certificate:  pulumi.String(d.Certificate),
		}
	case spec.GetUri() != nil:
		d := spec.GetUri()
		return cloudflare.DnsRecordDataArgs{
			Priority: f64(d.Priority),
			Weight:   f64(d.Weight),
			Target:   pulumi.String(d.Target.GetValue()),
		}
	}
	return nil
}
