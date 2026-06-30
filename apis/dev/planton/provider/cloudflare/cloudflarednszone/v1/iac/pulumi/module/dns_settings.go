package module

import (
	"github.com/pkg/errors"
	cloudflarednszonev1 "github.com/plantonhq/planton/apis/dev/planton/provider/cloudflare/cloudflarednszone/v1"
	"github.com/pulumi/pulumi-cloudflare/sdk/v6/go/cloudflare"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// dnsSettings applies zone-wide DNS settings. Unset numeric/string fields are
// left nil so Cloudflare's defaults remain in effect.
func dnsSettings(
	ctx *pulumi.Context,
	resourceName string,
	zone *cloudflare.Zone,
	spec *cloudflarednszonev1.CloudflareDnsZoneDnsSettings,
	cloudflareProvider *cloudflare.Provider,
) error {
	args := &cloudflare.ZoneDnsSettingsArgs{
		ZoneId:             zone.ID(),
		FlattenAllCnames:   pulumi.Bool(spec.FlattenAllCnames),
		FoundationDns:      pulumi.Bool(spec.FoundationDns),
		MultiProvider:      pulumi.Bool(spec.MultiProvider),
		SecondaryOverrides: pulumi.Bool(spec.SecondaryOverrides),
	}

	if spec.NsTtl > 0 {
		args.NsTtl = pulumi.Float64(float64(spec.NsTtl))
	}
	if zm := zoneModeString(spec.ZoneMode); zm != "" {
		args.ZoneMode = pulumi.String(zm)
	}

	if s := spec.Soa; s != nil {
		soa := cloudflare.ZoneDnsSettingsSoaArgs{}
		if s.Expire > 0 {
			soa.Expire = pulumi.Float64(float64(s.Expire))
		}
		if s.MinTtl > 0 {
			soa.MinTtl = pulumi.Float64(float64(s.MinTtl))
		}
		if s.Mname != "" {
			soa.Mname = pulumi.String(s.Mname)
		}
		if s.Refresh > 0 {
			soa.Refresh = pulumi.Float64(float64(s.Refresh))
		}
		if s.Retry > 0 {
			soa.Retry = pulumi.Float64(float64(s.Retry))
		}
		if s.Rname != "" {
			soa.Rname = pulumi.String(s.Rname)
		}
		if s.Ttl > 0 {
			soa.Ttl = pulumi.Float64(float64(s.Ttl))
		}
		args.Soa = soa
	}

	if n := spec.Nameservers; n != nil {
		ns := cloudflare.ZoneDnsSettingsNameserversArgs{}
		if n.NsSet > 0 {
			ns.NsSet = pulumi.Int(int(n.NsSet))
		}
		if n.Type != "" {
			ns.Type = pulumi.String(n.Type)
		}
		args.Nameservers = ns
	}

	if id := spec.InternalDns; id != nil && id.ReferenceZoneId != nil && id.ReferenceZoneId.GetValue() != "" {
		args.InternalDns = cloudflare.ZoneDnsSettingsInternalDnsArgs{
			ReferenceZoneId: pulumi.String(id.ReferenceZoneId.GetValue()),
		}
	}

	if _, err := cloudflare.NewZoneDnsSettings(
		ctx,
		resourceName+"-dns-settings",
		args,
		pulumi.Provider(cloudflareProvider),
	); err != nil {
		return errors.Wrap(err, "failed to apply zone dns settings")
	}
	return nil
}

// zoneModeString maps the ZoneMode enum to the provider's string value, returning
// "" for the unspecified zero value.
func zoneModeString(m cloudflarednszonev1.CloudflareDnsZoneSpec_ZoneMode) string {
	if m == cloudflarednszonev1.CloudflareDnsZoneSpec_zone_mode_unspecified {
		return ""
	}
	return m.String()
}
