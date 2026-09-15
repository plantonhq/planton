# CloudflareZoneTlsSettings

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `cloudflare.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

CloudflareZoneTlsSettingsSpec manages a zone's edge TLS posture: Universal SSL
issuance, Total TLS per-hostname certificates, automatic origin TLS key
exchange, origin TLS compliance modes, per-hostname TLS overrides, and
certificate-authority hostname associations.

A field left unset is NOT MANAGED: the module never sends it and the zone keeps
its current value. Universal SSL, Total TLS, auto origin TLS key exchange, and
CA hostname associations have NO DELETE at Cloudflare -- destroy abandons the
last-applied values. Per-hostname settings and origin TLS compliance modes have
real deletes: destroying the resource removes those overrides.

The sharpest edge in this kind: setting universal_ssl_enabled to false stops
Universal SSL certificate issuance for the zone. If no other certificate covers
a proxied hostname, that hostname becomes UNREACHABLE over HTTPS. Only disable
Universal SSL when dedicated, custom, or Total TLS certificates already cover
every proxied hostname.

## Example

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareZoneTlsSettings
metadata:
  name: test-zone-tls-settings
spec:
  zone_id:
    value: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  total_tls:
    enabled: true
    certificate_authority: google
  auto_origin_tls_kex: true
  origin_tls_compliance_modes:
    - fips
  hostname_settings:
    - hostname: api.example.com
      min_tls_version: "1.3"
      http2: true
    - hostname: legacy.example.com
      min_tls_version: "1.0"
      ciphers:
        - "ECDHE-RSA-AES128-GCM-SHA256"
  ca_hostname_associations:
    - hostnames:
        - mtls.example.com
      mtls_certificate_id:
        value: "8d773bb3-90b2-4bd3-9d33-9c3ba33d5b9a"
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.zoneId` | `string \| valueFrom` | yes |  | CloudflareDnsZone (`status.outputs.zone_id`) |
| `spec.universalSslEnabled` | `bool` |  |  |  |
| `spec.totalTls` | `CloudflareZoneTlsSettingsTotalTls` |  |  |  |
| `spec.totalTls.enabled` | `bool` | yes |  |  |
| `spec.totalTls.certificateAuthority` | `string` |  |  |  |
| `spec.autoOriginTlsKex` | `bool` |  |  |  |
| `spec.originTlsComplianceModes` | `[]string` |  |  |  |
| `spec.hostnameSettings` | `[]CloudflareZoneTlsSettingsHostnameSetting` |  |  |  |
| `spec.hostnameSettings[].hostname` | `string` | yes |  |  |
| `spec.hostnameSettings[].minTlsVersion` | `string` |  |  |  |
| `spec.hostnameSettings[].http2` | `bool` |  |  |  |
| `spec.hostnameSettings[].ciphers` | `[]string` |  |  |  |
| `spec.caHostnameAssociations` | `[]CloudflareZoneTlsSettingsCaHostnameAssociation` |  |  |  |
| `spec.caHostnameAssociations[].hostnames` | `[]string` | yes |  |  |
| `spec.caHostnameAssociations[].mtlsCertificateId` | `string \| valueFrom` |  |  | CloudflareMtlsCertificate (`status.outputs.certificate_id`) |

## Field Details

### spec.zoneId

`string | valueFrom` · required

The zone whose TLS settings are managed.
When using value_from, defaults to CloudflareDnsZone kind and status.outputs.zone_id field path.

- references: CloudflareDnsZone (`status.outputs.zone_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: CloudflareDnsZone, name: <that resource's name>, fieldPath: status.outputs.zone_id}} -- a bare string does not parse

### spec.universalSslEnabled

`bool` · optional (explicit presence)

Universal SSL certificate issuance for the zone. Setting false STOPS issuance
and can make proxied hostnames unreachable over HTTPS unless other
certificates cover them -- see the message-level warning above. No delete at
Cloudflare: destroy abandons the last-applied value.

### spec.totalTls

`CloudflareZoneTlsSettingsTotalTls`

Total TLS: automatically issue individual certificates for every proxied
hostname in the zone, including deep subdomains Universal SSL's wildcard does
not cover. ENTITLEMENT-GATED: the zone must carry an Advanced Certificate
Manager subscription or every write answers 401 code 1450 (measured at
v5.23.0 on Free and Pro zones alike -- ACM is a per-zone add-on, not a plan
tier). No delete at Cloudflare: destroy abandons the last-applied value.

### spec.totalTls.enabled

`bool` · required · optional (explicit presence)

Whether Total TLS issues per-hostname certificates for the zone. Required
inside total_tls: declaring the block takes Total TLS under management,
and the switch says which way.

- rule: {"required":true}

### spec.totalTls.certificateAuthority

`string` · optional (explicit presence)

The certificate authority issuing Total TLS certificates. Unset lets
Cloudflare choose. The certificates' validity period is fixed by Cloudflare
(90 days) and is not configurable.

- rule: {"string":{"in":["google","lets_encrypt","ssl_com"]}}

### spec.autoOriginTlsKex

`bool` · optional (explicit presence)

Automatic origin TLS key exchange: let Cloudflare negotiate the strongest key
exchange the origin supports on origin-facing TLS. ACTIVE ZONES ONLY: on a
pending (undelegated) zone the write answers 400 code 1000 with the
misleading message "Invalid zone identifier" -- the zone id is fine, the
setting simply does not exist until the zone activates (measured at
v5.23.0). No delete at Cloudflare: destroy abandons the last-applied value.

### spec.originTlsComplianceModes

`[]string`

Origin TLS compliance modes required on Cloudflare-to-origin connections.
Cloudflare documents fips and pqh (post-quantum hybrid) and may add values;
unknown strings pass through to the API unvalidated, deliberately -- treat the
values as an open vocabulary. Real delete at Cloudflare: destroying the
resource (or ceasing to manage the list) clears the compliance requirement
via the module's destroy; the module never sends an empty list.

### spec.hostnameSettings

`[]CloudflareZoneTlsSettingsHostnameSetting`

Per-hostname TLS overrides within the zone. Each row targets one hostname and
sets any of the three overridable settings; each set setting becomes its own
per-hostname override object at Cloudflare. ENTITLEMENT-GATED: the zone must
carry an Advanced Certificate Manager subscription or every override write
answers 401 code 1450 (measured at v5.23.0 on Free and Pro zones alike --
ACM is a per-zone add-on, not a plan tier). Real delete: destroying the
resource removes the overrides and the hostnames fall back to zone-wide
settings.

- rule: set at least one of min_tls_version, http2, or ciphers on the hostname row -- a row that overrides nothing manages nothing

### spec.hostnameSettings[].hostname

`string` · required

The hostname the overrides apply to. Changing the hostname replaces the
override objects (the hostname is part of their API identity).

- rule: {"required":true}

### spec.hostnameSettings[].minTlsVersion

`string` · optional (explicit presence)

Minimum TLS version for this hostname, overriding the zone-wide setting.

- rule: {"string":{"in":["1.0","1.1","1.2","1.3"]}}

### spec.hostnameSettings[].http2

`bool` · optional (explicit presence)

HTTP/2 support for this hostname, overriding the zone-wide setting.

### spec.hostnameSettings[].ciphers

`[]string`

TLS cipher allowlist for this hostname in BoringSSL format, overriding the
zone-wide ciphers.

### spec.caHostnameAssociations

`[]CloudflareZoneTlsSettingsCaHostnameAssociation`

Certificate-authority hostname associations: restrict which hostnames a
certificate authority may issue for. A row WITHOUT mtls_certificate_id
manages the zone's managed-CA association list; a row WITH it manages the
hostname list of that mTLS certificate. No delete at Cloudflare: destroy
abandons the last-applied associations.

### spec.caHostnameAssociations[].hostnames

`[]string` · required

The hostnames the certificate authority may issue for. An explicit empty
list is not sendable -- provide at least one hostname per managed row.

- rule: {"repeated":{"minItems":"1"}}

### spec.caHostnameAssociations[].mtlsCertificateId

`string | valueFrom`

The mTLS certificate whose hostname associations this row manages. Leave
unset to manage the zone's managed-CA association list instead. Accepts a
literal certificate id or a reference to a CloudflareMtlsCertificate
resource's certificate_id output.
When using value_from, defaults to CloudflareMtlsCertificate kind and
status.outputs.certificate_id field path.

- references: CloudflareMtlsCertificate (`status.outputs.certificate_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: CloudflareMtlsCertificate, name: <that resource's name>, fieldPath: status.outputs.certificate_id}} -- a bare string does not parse

## Validation Rules

- `spec.at_least_one_setting`: configure at least one TLS setting -- a CloudflareZoneTlsSettings resource that manages nothing would deploy nothing

## Outputs

Reference an output from another manifest as `valueFrom: {kind: CloudflareZoneTlsSettings, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.zone_id` | `string` | The zone ID the TLS settings belong to (the singleton's identity, and the pass-through for downstream resource references). |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.zoneId` | CloudflareDnsZone | `status.outputs.zone_id` |
| `spec.caHostnameAssociations[].mtlsCertificateId` | CloudflareMtlsCertificate | `status.outputs.certificate_id` |

## See Also

- [Overview](../README.md)
