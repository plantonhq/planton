package module

import (
	"fmt"

	"github.com/pkg/errors"
	cloudflarednszonev1alpha1 "github.com/plantonhq/planton/catalog/cloudflare/cloudflarednszone/v1alpha1"
	"github.com/pulumi/pulumi-cloudflare/sdk/v6/go/cloudflare"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// records creates DNS records within the zone and returns their
// Cloudflare-assigned ids keyed by the same name-type-index key the tofu
// module uses for for_each -- the keyed record_ids output import
// recipes derive per-record import IDs from.
func records(
	ctx *pulumi.Context,
	zone *cloudflare.Zone,
	recordsList []*cloudflarednszonev1alpha1.CloudflareDnsZoneRecord,
	cloudflareProvider *cloudflare.Provider,
) (pulumi.StringMap, error) {
	recordIds := pulumi.StringMap{}
	for idx, record := range recordsList {
		// Include index to ensure uniqueness when multiple records have same name and type
		resourceName := fmt.Sprintf("%s-%s-%d", record.Name, record.Type.String(), idx)

		// The provider requires ttl >= 1 (1 = automatic); the proto's unset 0
		// maps to automatic.
		ttl := float64(1)
		if record.Ttl > 0 {
			ttl = float64(record.Ttl)
		}

		recordArgs := &cloudflare.DnsRecordArgs{
			ZoneId: zone.ID(),
			Name:   pulumi.String(record.Name),
			Type:   pulumi.String(record.Type.String()),
			Ttl:    pulumi.Float64(ttl),
		}

		// Simple record types carry their value in content -- a literal, or a
		// reference already resolved to its value -- and structured types use data.
		if content := record.Content.GetValue(); content != "" {
			recordArgs.Content = pulumi.String(content)
		}
		if data := buildRecordData(record); data != nil {
			recordArgs.Data = data
		}

		// proxied is only applicable to A, AAAA, and CNAME records
		if record.Type == cloudflarednszonev1alpha1.CloudflareDnsZoneRecord_A ||
			record.Type == cloudflarednszonev1alpha1.CloudflareDnsZoneRecord_AAAA ||
			record.Type == cloudflarednszonev1alpha1.CloudflareDnsZoneRecord_CNAME {
			recordArgs.Proxied = pulumi.Bool(record.Proxied)
		}

		// The provider schema marks top-level priority "Required for MX, SRV
		// and URI records". MX carries it in the record's priority field;
		// SRV/URI carry it inside their structured data, and Cloudflare
		// mirrors that value into the top-level field on its own -- omitting
		// the mirror drifts on refresh forever (live-measured on the sibling
		// record kind's SRV lane). HTTPS/SVCB carry priority only in data.
		switch {
		case record.Type == cloudflarednszonev1alpha1.CloudflareDnsZoneRecord_MX:
			recordArgs.Priority = pulumi.Float64Ptr(float64(record.Priority))
		case record.GetSrv() != nil:
			recordArgs.Priority = pulumi.Float64Ptr(float64(record.GetSrv().Priority))
		case record.GetUri() != nil:
			recordArgs.Priority = pulumi.Float64Ptr(float64(record.GetUri().Priority))
		}

		// comment for the DNS record
		if record.Comment != "" {
			recordArgs.Comment = pulumi.String(record.Comment)
		}

		if len(record.Tags) > 0 {
			tags := make(pulumi.StringArray, 0, len(record.Tags))
			for _, t := range record.Tags {
				tags = append(tags, pulumi.String(t))
			}
			recordArgs.Tags = tags
		}

		if s := record.Settings; s != nil {
			recordArgs.Settings = cloudflare.DnsRecordSettingsArgs{
				Ipv4Only:     pulumi.Bool(s.Ipv4Only),
				Ipv6Only:     pulumi.Bool(s.Ipv6Only),
				FlattenCname: pulumi.Bool(s.FlattenCname),
			}
		}

		if record.PrivateRouting {
			recordArgs.PrivateRouting = pulumi.Bool(true)
		}

		createdRecord, err := cloudflare.NewDnsRecord(
			ctx,
			resourceName,
			recordArgs,
			pulumi.Provider(cloudflareProvider),
			pulumi.DependsOn([]pulumi.Resource{zone}),
		)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to create dns record %s", resourceName)
		}
		recordIds[resourceName] = createdRecord.ID()
	}
	return recordIds, nil
}

// buildRecordData translates the typed `data` oneof into the provider's flat
// data object, returning nil when the record is a simple (content) record.
func buildRecordData(record *cloudflarednszonev1alpha1.CloudflareDnsZoneRecord) cloudflare.DnsRecordDataPtrInput {
	f64 := func(v uint32) pulumi.Float64PtrInput { return pulumi.Float64Ptr(float64(v)) }

	switch {
	case record.GetCaa() != nil:
		d := record.GetCaa()
		return cloudflare.DnsRecordDataArgs{
			Flags: pulumi.Float64(float64(d.Flags)),
			Tag:   pulumi.String(d.Tag),
			Value: pulumi.String(d.Value),
		}
	case record.GetCert() != nil:
		d := record.GetCert()
		return cloudflare.DnsRecordDataArgs{
			Type:        f64(d.Type),
			KeyTag:      f64(d.KeyTag),
			Algorithm:   f64(d.Algorithm),
			Certificate: pulumi.String(d.Certificate),
		}
	case record.GetDnskey() != nil:
		d := record.GetDnskey()
		return cloudflare.DnsRecordDataArgs{
			Flags:     pulumi.Float64(float64(d.Flags)),
			Protocol:  f64(d.Protocol),
			Algorithm: f64(d.Algorithm),
			PublicKey: pulumi.String(d.PublicKey),
		}
	case record.GetDs() != nil:
		d := record.GetDs()
		return cloudflare.DnsRecordDataArgs{
			KeyTag:     f64(d.KeyTag),
			Algorithm:  f64(d.Algorithm),
			DigestType: f64(d.DigestType),
			Digest:     pulumi.String(d.Digest),
		}
	case record.GetHttps() != nil:
		d := record.GetHttps()
		return cloudflare.DnsRecordDataArgs{
			Priority: f64(d.Priority),
			Target:   pulumi.String(d.Target.GetValue()),
			Value:    pulumi.String(d.Value),
		}
	case record.GetLoc() != nil:
		d := record.GetLoc()
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
	case record.GetNaptr() != nil:
		d := record.GetNaptr()
		return cloudflare.DnsRecordDataArgs{
			Flags:       pulumi.String(d.Flags),
			Order:       f64(d.Order),
			Preference:  f64(d.Preference),
			Service:     pulumi.String(d.Service),
			Regex:       pulumi.String(d.Regex),
			Replacement: pulumi.String(d.Replacement),
		}
	case record.GetSmimea() != nil:
		d := record.GetSmimea()
		return cloudflare.DnsRecordDataArgs{
			Usage:        f64(d.Usage),
			Selector:     f64(d.Selector),
			MatchingType: f64(d.MatchingType),
			Certificate:  pulumi.String(d.Certificate),
		}
	case record.GetSrv() != nil:
		d := record.GetSrv()
		return cloudflare.DnsRecordDataArgs{
			Priority: f64(d.Priority),
			Weight:   f64(d.Weight),
			Port:     f64(d.Port),
			Target:   pulumi.String(d.Target.GetValue()),
		}
	case record.GetSshfp() != nil:
		d := record.GetSshfp()
		return cloudflare.DnsRecordDataArgs{
			Algorithm:   f64(d.Algorithm),
			Type:        f64(d.Type),
			Fingerprint: pulumi.String(d.Fingerprint),
		}
	case record.GetSvcb() != nil:
		d := record.GetSvcb()
		return cloudflare.DnsRecordDataArgs{
			Priority: f64(d.Priority),
			Target:   pulumi.String(d.Target.GetValue()),
			Value:    pulumi.String(d.Value),
		}
	case record.GetTlsa() != nil:
		d := record.GetTlsa()
		return cloudflare.DnsRecordDataArgs{
			Usage:        f64(d.Usage),
			Selector:     f64(d.Selector),
			MatchingType: f64(d.MatchingType),
			Certificate:  pulumi.String(d.Certificate),
		}
	case record.GetUri() != nil:
		d := record.GetUri()
		return cloudflare.DnsRecordDataArgs{
			Priority: f64(d.Priority),
			Weight:   f64(d.Weight),
			Target:   pulumi.String(d.Target.GetValue()),
		}
	}
	return nil
}
