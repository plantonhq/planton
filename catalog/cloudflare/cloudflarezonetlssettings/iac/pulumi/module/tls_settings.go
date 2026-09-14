package module

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-cloudflare/sdk/v6/go/cloudflare"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// onOff maps the spec's managed booleans to the per-hostname settings API's
// on/off strings.
func onOff(b bool) string {
	if b {
		return "on"
	}
	return "off"
}

// sanitizeHostname makes a hostname safe for embedding in a resource name.
func sanitizeHostname(hostname string) string {
	return strings.NewReplacer(".", "-", "*", "wildcard").Replace(hostname)
}

// tlsSettings emits one API object per managed TLS setting.
func tlsSettings(
	ctx *pulumi.Context,
	locals *Locals,
	cloudflareProvider *cloudflare.Provider,
) error {
	spec := locals.CloudflareZoneTlsSettings.Spec
	zoneId := pulumi.String(spec.ZoneId.GetValue())

	// Universal SSL issuance. NO delete at Cloudflare: destroy abandons the
	// last-applied value. Disabling can make proxied hostnames unreachable over
	// HTTPS -- the spec carries the warning.
	if spec.UniversalSslEnabled != nil {
		if _, err := cloudflare.NewUniversalSslSetting(
			ctx,
			"universal_ssl_setting",
			&cloudflare.UniversalSslSettingArgs{
				ZoneId:  zoneId,
				Enabled: pulumi.Bool(*spec.UniversalSslEnabled),
			},
			pulumi.Provider(cloudflareProvider),
		); err != nil {
			return errors.Wrap(err, "failed to create universal ssl setting")
		}
	}

	// Total TLS. NO delete at Cloudflare. The certificates' validity period is
	// fixed by Cloudflare (90 days) and deliberately not modeled.
	if spec.TotalTls != nil {
		totalTlsArgs := &cloudflare.TotalTlsArgs{
			ZoneId:  zoneId,
			Enabled: pulumi.Bool(spec.TotalTls.GetEnabled()),
		}
		if spec.TotalTls.CertificateAuthority != nil {
			totalTlsArgs.CertificateAuthority = pulumi.StringPtr(*spec.TotalTls.CertificateAuthority)
		}
		if _, err := cloudflare.NewTotalTls(
			ctx,
			"total_tls",
			totalTlsArgs,
			pulumi.Provider(cloudflareProvider),
		); err != nil {
			return errors.Wrap(err, "failed to create total tls")
		}
	}

	// Automatic origin TLS key exchange. NO delete at Cloudflare.
	if spec.AutoOriginTlsKex != nil {
		if _, err := cloudflare.NewZoneAutoOriginTlsKex(
			ctx,
			"auto_origin_tls_kex",
			&cloudflare.ZoneAutoOriginTlsKexArgs{
				ZoneId:  zoneId,
				Enabled: pulumi.Bool(*spec.AutoOriginTlsKex),
			},
			pulumi.Provider(cloudflareProvider),
		); err != nil {
			return errors.Wrap(err, "failed to create auto origin tls kex")
		}
	}

	// Origin TLS compliance modes. Real delete: destroying the resource clears
	// the compliance requirement. The module never sends an empty list -- the
	// spec's at-least-one contract keeps this arm unmanaged when the list is
	// empty.
	if len(spec.OriginTlsComplianceModes) > 0 {
		if _, err := cloudflare.NewOriginTlsComplianceModes(
			ctx,
			"origin_tls_compliance_modes",
			&cloudflare.OriginTlsComplianceModesArgs{
				ZoneId: zoneId,
				Values: pulumi.ToStringArray(spec.OriginTlsComplianceModes),
			},
			pulumi.Provider(cloudflareProvider),
		); err != nil {
			return errors.Wrap(err, "failed to create origin tls compliance modes")
		}
	}

	// Per-hostname TLS overrides: each set attribute of each row is its own API
	// object keyed by (setting_id, hostname). Resource names embed both so
	// editing one row never churns another row's resources. Real delete:
	// destroy removes the overrides.
	for _, row := range spec.HostnameSettings {
		hostKey := sanitizeHostname(row.Hostname)
		if row.MinTlsVersion != nil {
			if _, err := cloudflare.NewHostnameTlsSetting(
				ctx,
				"hostname_tls_min_tls_version_"+hostKey,
				&cloudflare.HostnameTlsSettingArgs{
					ZoneId:    zoneId,
					SettingId: pulumi.String("min_tls_version"),
					Hostname:  pulumi.String(row.Hostname),
					Value:     pulumi.Any(*row.MinTlsVersion),
				},
				pulumi.Provider(cloudflareProvider),
			); err != nil {
				return errors.Wrapf(err, "failed to create min_tls_version override for %q", row.Hostname)
			}
		}
		if row.Http2 != nil {
			if _, err := cloudflare.NewHostnameTlsSetting(
				ctx,
				"hostname_tls_http2_"+hostKey,
				&cloudflare.HostnameTlsSettingArgs{
					ZoneId:    zoneId,
					SettingId: pulumi.String("http2"),
					Hostname:  pulumi.String(row.Hostname),
					Value:     pulumi.Any(onOff(*row.Http2)),
				},
				pulumi.Provider(cloudflareProvider),
			); err != nil {
				return errors.Wrapf(err, "failed to create http2 override for %q", row.Hostname)
			}
		}
		if len(row.Ciphers) > 0 {
			if _, err := cloudflare.NewHostnameTlsSetting(
				ctx,
				"hostname_tls_ciphers_"+hostKey,
				&cloudflare.HostnameTlsSettingArgs{
					ZoneId:    zoneId,
					SettingId: pulumi.String("ciphers"),
					Hostname:  pulumi.String(row.Hostname),
					Value:     pulumi.Any(row.Ciphers),
				},
				pulumi.Provider(cloudflareProvider),
			); err != nil {
				return errors.Wrapf(err, "failed to create ciphers override for %q", row.Hostname)
			}
		}
	}

	// Certificate-authority hostname associations. NO delete at Cloudflare:
	// destroy abandons the last-applied association lists. A row without an
	// mTLS certificate manages the zone's managed-CA list; a row with one
	// manages that certificate's list (the resource name embeds the id).
	for _, association := range spec.CaHostnameAssociations {
		resourceName := "ca_hostname_association_managed"
		associationArgs := &cloudflare.CertificateAuthoritiesHostnameAssociationsArgs{
			ZoneId:    zoneId,
			Hostnames: pulumi.ToStringArray(association.Hostnames),
		}
		if association.MtlsCertificateId != nil && association.MtlsCertificateId.GetValue() != "" {
			mtlsCertificateId := association.MtlsCertificateId.GetValue()
			resourceName = "ca_hostname_association_" + mtlsCertificateId
			associationArgs.MtlsCertificateId = pulumi.StringPtr(mtlsCertificateId)
		}
		if _, err := cloudflare.NewCertificateAuthoritiesHostnameAssociations(
			ctx,
			resourceName,
			associationArgs,
			pulumi.Provider(cloudflareProvider),
		); err != nil {
			return errors.Wrap(err, "failed to create certificate authority hostname associations")
		}
	}

	return nil
}
