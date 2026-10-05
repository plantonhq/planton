# CloudflareCustomHostname

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `cloudflare.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

CloudflareCustomHostnameSpec attaches a customer's own domain to a Cloudflare
for SaaS zone (the "SaaS zone"). It extends Cloudflare's edge — TLS termination,
caching, WAF — onto a hostname your customer owns (e.g. "support.acme.com"),
with a per-customer certificate that Cloudflare provisions and auto-renews. The
customer points their hostname at the SaaS zone via CNAME and proves control via
the ownership-verification records exported in the outputs.

Traffic for the custom hostname is routed to the zone's default origin (the
CloudflareCustomHostnameFallbackOrigin) unless `custom_origin_server` overrides
it for this hostname.

## Example

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareCustomHostname
metadata:
  name: test-custom-hostname
spec:
  zone_id:
    value: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  hostname: support.acme.com
  ssl:
    method: txt
    type: dv
    settings:
      minTlsVersion: "1.2"
      http2: "on"
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.zoneId` | `string \| valueFrom` | yes |  | CloudflareDnsZone (`status.outputs.zone_id`) |
| `spec.hostname` | `string` | yes |  |  |
| `spec.customOriginServer` | `string \| valueFrom` |  |  |  |
| `spec.customOriginSni` | `string` |  |  |  |
| `spec.customMetadata` | `map<string, string>` |  |  |  |
| `spec.ssl` | `CloudflareCustomHostnameSsl` |  |  |  |
| `spec.ssl.bundleMethod` | `string` |  | `ubiquitous` |  |
| `spec.ssl.certificateAuthority` | `string` |  |  |  |
| `spec.ssl.cloudflareBranding` | `bool` |  |  |  |
| `spec.ssl.method` | `string` |  |  |  |
| `spec.ssl.type` | `string` |  | `dv` |  |
| `spec.ssl.wildcard` | `bool` |  |  |  |
| `spec.ssl.customCertificate` | `string` |  |  |  |
| `spec.ssl.customCsrId` | `string` |  |  |  |
| `spec.ssl.customKey` | `string` (sensitive) |  |  |  |
| `spec.ssl.customCertBundle` | `[]CloudflareCustomHostnameSslCustomCertBundle` |  |  |  |
| `spec.ssl.customCertBundle[].customCertificate` | `string` | yes |  |  |
| `spec.ssl.customCertBundle[].customKey` | `string` (sensitive) | yes |  |  |
| `spec.ssl.settings` | `CloudflareCustomHostnameSslSettings` |  |  |  |
| `spec.ssl.settings.ciphers` | `[]string` |  |  |  |
| `spec.ssl.settings.earlyHints` | `string` |  |  |  |
| `spec.ssl.settings.http2` | `string` |  |  |  |
| `spec.ssl.settings.minTlsVersion` | `string` |  |  |  |
| `spec.ssl.settings.tls13` | `string` |  |  |  |

## Field Details

### spec.zoneId

`string | valueFrom` · required

The Cloudflare Zone ID of the SaaS zone this custom hostname is added to. A
literal value or a reference to a CloudflareDnsZone resource (its
status.outputs.zone_id).

- references: CloudflareDnsZone (`status.outputs.zone_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: CloudflareDnsZone, name: <that resource's name>, fieldPath: status.outputs.zone_id}} -- a bare string does not parse

### spec.hostname

`string` · required

The customer's hostname to onboard, e.g. "support.acme.com". This is the
customer's own (external) domain, so it is a literal value, not a reference.

- rule: {"required":true,"string":{"minLen":"1","maxLen":"255"}}

### spec.customOriginServer

`string | valueFrom`

Override the origin this specific custom hostname routes to. A backend
endpoint — a literal hostname or a reference to another resource's output
(e.g. a load balancer hostname). When omitted, the zone's fallback origin is
used.

- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.customOriginSni

`string`

The SNI hostname sent to the custom origin during the TLS handshake. Can be a
valid subdomain of the SaaS zone, the custom origin server name, or the literal
":request_host_header:" to forward the request Host. Not configurable with a
default/fallback origin.

### spec.customMetadata

`map<string, string>`

Arbitrary key/value metadata stored with the custom hostname (e.g. a tenant id),
available to Workers and rules.

### spec.ssl

`CloudflareCustomHostnameSsl`

SSL/TLS settings for the custom hostname's certificate. When omitted, Cloudflare
provisions a DV certificate with sensible defaults.

### spec.ssl.bundleMethod

`string` · optional (explicit presence)

Certificate bundling strategy: "ubiquitous" (the default; widest client
compatibility), "optimal" (shortest chain), or "force" (verify but do not
modify the chain).

- default: `ubiquitous`
- rule: bundle_method must be one of ubiquitous, optimal, force

### spec.ssl.certificateAuthority

`string`

The certificate authority that issues the certificate: "digicert", "google",
"lets_encrypt", or "ssl_com". CA selection is Enterprise-gated: on any other
plan the API rejects an explicit value with 400 code 1459 "Certificate
Authority selection is only available on an Enterprise plan" (measured live
2026-08-29). Leave empty on non-Enterprise accounts -- Cloudflare then
assigns a CA at random per create (measured: ssl_com and google on
consecutive creates). Because that server pick can never be mirrored in
config, both IaC modules ignore post-create drift on this field; an
in-place CHANGE of an already-issued hostname's CA is therefore not
applied -- recreate the hostname to change CA (Enterprise).

- rule: certificate_authority must be empty or one of digicert, google, lets_encrypt, ssl_com

### spec.ssl.cloudflareBranding

`bool`

Add Cloudflare branding (a sni.cloudflaressl.com subdomain as the Common Name).

### spec.ssl.method

`string`

The domain control validation method used to prove control of the hostname:
"http", "txt", or "email". Leave empty to use the provider default.

- rule: method must be empty or one of http, txt, email

### spec.ssl.type

`string` · optional (explicit presence)

The level of validation: only "dv" (domain validation) is supported. Defaults
to "dv".

- default: `dv`
- rule: type must be dv

### spec.ssl.wildcard

`bool`

Issue a wildcard certificate that also covers a one-level subdomain of the
hostname (Enterprise feature).

### spec.ssl.customCertificate

`string`

A user-uploaded certificate (PEM) to use instead of a Cloudflare-issued one
(Enterprise feature). Public certificate material, not a secret; the private
key is custom_key.

### spec.ssl.customCsrId

`string`

The identifier of a previously-uploaded custom CSR to use (Enterprise feature).
An identifier, not a secret.

### spec.ssl.customKey

`string` · sensitive

The private key (PEM) for `custom_certificate` (Enterprise feature). Sensitive.

### spec.ssl.customCertBundle

`[]CloudflareCustomHostnameSslCustomCertBundle`

One or two custom certificate/key pairs to upload (Enterprise feature),
supporting mixed RSA/ECDSA serving.

- rule: {"repeated":{"maxItems":"2"}}

### spec.ssl.customCertBundle[].customCertificate

`string` · required

The certificate (PEM). Public material, not a secret; the private key is
custom_key.

- rule: {"required":true}

### spec.ssl.customCertBundle[].customKey

`string` · required · sensitive

The private key (PEM) for the certificate. Sensitive.

- rule: {"required":true}

### spec.ssl.settings

`CloudflareCustomHostnameSslSettings`

Fine-grained TLS termination settings for the custom hostname.

### spec.ssl.settings.ciphers

`[]string`

An allowlist of ciphers for TLS termination, in BoringSSL format. Leave empty
to use Cloudflare's defaults.

### spec.ssl.settings.earlyHints

`string`

Whether Early Hints is enabled: "on" or "off". Leave empty for the default.

- rule: early_hints must be empty or one of on, off

### spec.ssl.settings.http2

`string`

Whether HTTP/2 is enabled: "on" or "off". Leave empty for the default.

- rule: http2 must be empty or one of on, off

### spec.ssl.settings.minTlsVersion

`string`

The minimum TLS version: "1.0", "1.1", "1.2", or "1.3". Leave empty for the
default.

- rule: min_tls_version must be empty or one of 1.0, 1.1, 1.2, 1.3

### spec.ssl.settings.tls13

`string`

Whether TLS 1.3 is enabled: "on" or "off". Leave empty for the default.

- rule: tls_1_3 must be empty or one of on, off

## Outputs

Reference an output from another manifest as `valueFrom: {kind: CloudflareCustomHostname, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.custom_hostname_id` | `string` | The custom hostname identifier. |
| `status.outputs.ownership_verification_name` | `string` | The DNS record NAME the customer must create for ownership verification. |
| `status.outputs.ownership_verification_type` | `string` | The DNS record TYPE for ownership verification (typically "txt"). |
| `status.outputs.ownership_verification_value` | `string` | The DNS record VALUE for ownership verification. |
| `status.outputs.ownership_verification_http_url` | `string` | The URL the customer can serve for HTTP-based ownership verification. |
| `status.outputs.ownership_verification_http_body` | `string` | The body the customer must serve at the HTTP verification URL. |
| `status.outputs.created_at` | `string` | RFC3339 timestamp of when the custom hostname was created. |
| `status.outputs.zone_id` | `string` | The Cloudflare Zone ID the hostname was onboarded onto. A custom hostname's API identity is (zone_id, custom_hostname_id), so downstream consumers -- verification tooling, imports, chart blocks composing on the hostname -- need the zone alongside the hostname's own id. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.zoneId` | CloudflareDnsZone | `status.outputs.zone_id` |

## See Also

- [Overview](../README.md)
