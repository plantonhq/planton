# Cloudflare Zone TLS Settings

## Overview

`CloudflareZoneTlsSettings` is a resource for managing a Cloudflare zone's edge TLS posture: Universal SSL certificate issuance, Total TLS per-hostname certificates, automatic origin TLS key exchange, origin TLS compliance modes, per-hostname TLS overrides, and certificate-authority hostname associations.

Each of these is a separate zone-scoped object at Cloudflare, but they answer the same operational question -- "how does TLS behave at the edge of this zone?" -- so this kind manages them together. A field left unset is NOT MANAGED: the module never sends it and the zone keeps whatever value it already has. That makes it safe to manage only the one setting you care about.

The sharpest edge in this kind: setting `universal_ssl_enabled` to false stops Universal SSL certificate issuance for the zone. If no other certificate covers a proxied hostname, that hostname becomes UNREACHABLE over HTTPS. Only disable Universal SSL when dedicated, custom, or Total TLS certificates already cover every proxied hostname.

## Key Features

- **One kind, six TLS surfaces**: Universal SSL, Total TLS, auto origin TLS key exchange, origin TLS compliance modes, per-hostname overrides, and CA hostname associations, all in one spec
- **Unset means untouched**: any field you leave out of the manifest is never sent to Cloudflare, so partial management is the default, not a special mode
- **Per-hostname precision**: raise `min_tls_version`, toggle `http2`, or pin `ciphers` for individual hostnames without changing the zone-wide settings
- **Reference support**: `zone_id` can reference a `CloudflareDnsZone` resource, and `mtls_certificate_id` can reference a `CloudflareMtlsCertificate` resource, keeping dependencies in the deployment graph

## Use Cases

**Ideal for:**
- Enabling Total TLS so deep subdomains (which Universal SSL's wildcard does not cover) get their own certificates
- Enforcing TLS 1.3 on an API hostname while the rest of the zone stays on the zone-wide minimum
- Requiring FIPS or post-quantum hybrid compliance on Cloudflare-to-origin connections
- Restricting which hostnames a certificate authority may issue for, zone-wide or per mTLS certificate

**Not ideal for:**
- Zone-wide TLS settings like the zone's minimum TLS version or Always Use HTTPS (use `CloudflareZoneSettings` instead)
- Ordering advanced edge certificates (use `CloudflareCertificatePack` instead)
- TLS for hostnames outside the zone, such as SaaS vanity domains (use `CloudflareCustomHostname` instead)

## API Specification

### CloudflareZoneTlsSettingsSpec

At least one TLS setting must be configured -- a resource that manages nothing would deploy nothing.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `zone_id` | StringValueOrRef | Yes | The zone whose TLS settings are managed. Accepts a literal zone ID or a reference to a `CloudflareDnsZone` resource. |
| `universal_ssl_enabled` | bool (optional) | No | Universal SSL certificate issuance for the zone. Setting false STOPS issuance and can make proxied hostnames unreachable over HTTPS unless other certificates cover them. No delete at Cloudflare. |
| `total_tls` | message | No | Total TLS: automatically issue individual certificates for every proxied hostname, including deep subdomains Universal SSL's wildcard does not cover. `enabled` (bool) plus optional `certificate_authority` (`google`, `lets_encrypt`, `ssl_com`). Requires the zone's Advanced Certificate Manager subscription (401 code 1450 without it). No delete at Cloudflare. |
| `auto_origin_tls_kex` | bool (optional) | No | Automatic origin TLS key exchange: let Cloudflare negotiate the strongest key exchange the origin supports. No delete at Cloudflare. |
| `origin_tls_compliance_modes` | repeated string | No | Compliance modes required on Cloudflare-to-origin connections. Cloudflare documents `fips` and `pqh`; unknown strings pass through unvalidated as an open vocabulary. Real delete at Cloudflare. |
| `hostname_settings[]` | repeated message | No | Per-hostname TLS overrides. Each row targets one `hostname` (required) and sets any of `min_tls_version` (`1.0`-`1.3`), `http2` (bool), or `ciphers` (BoringSSL format). At least one override per row. Requires the zone's Advanced Certificate Manager subscription (401 code 1450 without it). Real delete at Cloudflare. |
| `ca_hostname_associations[]` | repeated message | No | CA hostname associations: `hostnames` (at least one) plus optional `mtls_certificate_id` (StringValueOrRef to a `CloudflareMtlsCertificate`). Without a certificate ID the row manages the zone's managed-CA list; with one it manages that certificate's hostname list. No delete at Cloudflare. |

### Outputs

After successful deployment, the following outputs are available:

| Field | Description |
|-------|-------------|
| `zone_id` | The zone ID the TLS settings belong to. TLS settings are a zone-scoped singleton with no resource ID of their own -- the zone is the identity. |

## How It Works

Each TLS setting in the spec maps to its own zone-scoped API object at Cloudflare, and the IaC module emits each object only when the manifest manages it:

1. **Universal SSL, Total TLS, and auto origin TLS key exchange** are singletons, created only when their field is set
2. **Origin TLS compliance modes** is a singleton, created only when the list is non-empty (the module never sends an empty list)
3. **Per-hostname overrides** fan out into one API object per (setting, hostname) pair, keyed by hostname so editing one row never churns another row's resources
4. **CA hostname associations** create one object per row, keyed by the mTLS certificate ID (or `managed` for the zone-wide row)

## Destroy Behavior

The six surfaces split into two destroy classes, and knowing which is which matters:

- **Real delete** (2): `hostname_settings` and `origin_tls_compliance_modes`. Destroying the resource removes the overrides and clears the compliance requirement; hostnames fall back to zone-wide settings.
- **No delete** (4): `universal_ssl_enabled`, `total_tls`, `auto_origin_tls_kex`, and `ca_hostname_associations` have no delete at Cloudflare. Destroy drops them from state and abandons the last-applied values -- the zone keeps whatever was last written.

If you need to revert a no-delete setting, write the value you want before destroying, not after.

## The Universal SSL Warning

`universal_ssl_enabled: false` is the one field in this kind that can take a site down. It stops certificate issuance for the zone, and any proxied hostname not covered by another certificate (dedicated, custom, or Total TLS) becomes unreachable over HTTPS. Disable it only as a deliberate step in a migration to other certificates, and confirm coverage first.

## Related Resources

- **Cloudflare DNS Zone**: the zone whose TLS posture this resource manages, referenced by `zone_id`
- **Cloudflare Zone Settings**: zone-wide settings like minimum TLS version and Always Use HTTPS
- **Cloudflare Certificate Pack**: advanced edge certificates, often the coverage that makes disabling Universal SSL safe
- **Cloudflare Custom Hostname**: TLS for SaaS vanity hostnames outside this zone

## Further Reading

For operational judgment on destroy classes, provider quirks, and when to touch Universal SSL, see GUIDE.md.

## References

- [Universal SSL](https://developers.cloudflare.com/ssl/edge-certificates/universal-ssl/)
- [Total TLS](https://developers.cloudflare.com/ssl/edge-certificates/additional-options/total-tls/)
- [Minimum TLS Version](https://developers.cloudflare.com/ssl/edge-certificates/additional-options/minimum-tls/)
- [Cipher Suites](https://developers.cloudflare.com/ssl/edge-certificates/additional-options/cipher-suites/)

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
