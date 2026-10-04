# CloudflareCustomSslCertificate

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `cloudflare.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

CloudflareCustomSslCertificateSpec uploads a bring-your-own TLS certificate
to a zone (Cloudflare calls this a custom SSL certificate): Cloudflare
presents this certificate to visitors instead of a Universal SSL or
Advanced certificate. The certificate must be issued by a certificate
authority on Cloudflare's trust list -- self-signed certificates are
rejected at the API. Custom certificates are a Business/Enterprise zone
feature; Cloudflare enforces the plan gate at create -- the API, not this
spec, is the wall.

Rotation is replacement: changing the certificate or private key
destroys and re-creates the upload (the certificate id changes). Cloudflare
serves the previous certificate until the replacement deploys, so rotation
is not an outage, but anything referencing the old certificate id must
follow the new one.

## Example

```yaml
# A complete, protovalidate-valid CloudflareCustomSslCertificate example: an
# SNI certificate upload with an explicit bundle method and a US private-key
# geo restriction. The PEM values are placeholders -- a real upload needs a
# certificate issued by a publicly trusted CA covering the zone's hostnames.
# Custom certificates are a Business/Enterprise zone feature; the plan gate
# is Cloudflare's, at create.
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareCustomSslCertificate
metadata:
  name: www-custom-certificate
spec:
  zone_id:
    value: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  certificate: |
    -----BEGIN CERTIFICATE-----
    REPLACE_WITH_CERTIFICATE_PEM
    -----END CERTIFICATE-----
  private_key:
    value: |
      -----BEGIN PRIVATE KEY-----
      REPLACE_WITH_PRIVATE_KEY_PEM
      -----END PRIVATE KEY-----
  type: sni_custom
  bundle_method: ubiquitous
  geo_restrictions:
    label: us
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.zoneId` | `string \| valueFrom` | yes |  | CloudflareDnsZone (`status.outputs.zone_id`) |
| `spec.certificate` | `string` | yes |  |  |
| `spec.privateKey` | `string \| valueFrom` (sensitive) | yes |  |  |
| `spec.type` | `string` |  | `legacy_custom` |  |
| `spec.bundleMethod` | `string` |  | `ubiquitous` |  |
| `spec.policy` | `string` |  |  |  |
| `spec.geoRestrictions` | `CloudflareCustomSslCertificateGeoRestrictions` |  |  |  |
| `spec.geoRestrictions.label` | `string` |  |  |  |
| `spec.customCsrId` | `string` |  |  |  |
| `spec.deploy` | `string` |  |  |  |

## Field Details

### spec.zoneId

`string | valueFrom` · required

The zone the certificate is uploaded to.
When using value_from, defaults to CloudflareDnsZone kind and status.outputs.zone_id field path.

- references: CloudflareDnsZone (`status.outputs.zone_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: CloudflareDnsZone, name: <that resource's name>, fieldPath: status.outputs.zone_id}} -- a bare string does not parse

### spec.certificate

`string` · required

The certificate in PEM form, including any intermediate chain the chosen
bundle method expects. Must cover the zone's hostnames and be issued by a
publicly trusted certificate authority. Keep the PEM byte-stable
(trailing newline included) -- a formatting-only change still replaces
the upload.

- rule: {"required":true}

### spec.privateKey

`string | valueFrom` · required · sensitive

The certificate's private key in PEM form. Provide a managed-secret
reference; the platform resolves it just-in-time at deploy and never
stores it in plaintext. The API never returns the key, so an imported or
refreshed resource re-asserts it from configuration.

- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.type

`string` · optional (explicit presence)

The SNI class of the upload (Cloudflare's default: legacy_custom).
legacy_custom works on every TLS client but consumes the zone's single
legacy slot; sni_custom requires SNI-capable clients (all modern
browsers) and allows multiple uploads. Changing this on an existing
upload replaces it.

- default: `legacy_custom`
- rule: {"string":{"in":["legacy_custom","sni_custom"]}}

### spec.bundleMethod

`string` · optional (explicit presence)

How Cloudflare bundles the certificate chain (Cloudflare's default:
ubiquitous). ubiquitous maximizes older-client compatibility; optimal
prefers the shortest modern chain; force uses the chain exactly as
uploaded.

- default: `ubiquitous`
- rule: {"string":{"in":["ubiquitous","optimal","force"]}}

### spec.policy

`string`

A geo-policy expression restricting which data centers hold the private
key (for example "(country: US) or (region: EU)"). The API accepts this
as "policy" and returns the parsed form in a separate read-only field;
both engines send exactly what is written here.

### spec.geoRestrictions

`CloudflareCustomSslCertificateGeoRestrictions`

Coarse-grained private-key geo restriction: a single label instead of a
policy expression. us and eu keep the key in that region; highest_security
keeps it only in data centers meeting Cloudflare's highest security
requirements.

### spec.geoRestrictions.label

`string` · optional (explicit presence)

The restriction label.

- rule: {"string":{"in":["us","eu","highest_security"]}}

### spec.customCsrId

`string`

The ID of a Cloudflare-generated custom CSR this certificate was issued
from, when the CSR flow was used. A plain ID string -- CSR generation is
its own Cloudflare surface.

### spec.deploy

`string` · optional (explicit presence)

Deploy the certificate to Cloudflare's staging network first
(staging) or straight to production (Cloudflare's default). The staging
network is a Business/Enterprise validation surface; the value is
write-only bookkeeping on the upload.

- rule: {"string":{"in":["staging","production"]}}

## Outputs

Reference an output from another manifest as `valueFrom: {kind: CloudflareCustomSslCertificate, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.certificate_id` | `string` | The ID of the uploaded certificate. |
| `status.outputs.zone_id` | `string` | The zone the certificate belongs to. |
| `status.outputs.expires_on` | `string` | When the certificate expires (RFC3339). |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.zoneId` | CloudflareDnsZone | `status.outputs.zone_id` |

## See Also

- [Overview](../README.md)
