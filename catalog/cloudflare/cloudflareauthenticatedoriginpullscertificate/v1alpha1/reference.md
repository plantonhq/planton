# CloudflareAuthenticatedOriginPullsCertificate

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `cloudflare.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

CloudflareAuthenticatedOriginPullsCertificateSpec uploads the client
certificate Cloudflare presents to YOUR origin when Authenticated Origin
Pulls is enabled. Self-signed certificates are the normal case -- the
origin, not the public, validates this certificate. The scope decides the
blast radius: a zone upload replaces Cloudflare's shared certificate for
the whole zone; a hostname upload is referenced per hostname through
CloudflareAuthenticatedOriginPulls associations.

Rotation is replacement: upload a new certificate (a new resource or a
changed certificate value), re-point consumers, then destroy the old
upload. Never rotate only the private key against the same certificate --
the zone-scoped API silently ignores a key-only change (a measured
provider defect at v5.23.0), so key and certificate always change together.

## Example

```yaml
# A complete, protovalidate-valid CloudflareAuthenticatedOriginPullsCertificate
# example: a zone-scoped upload replacing Cloudflare's shared client
# certificate for the whole zone. The PEM values are placeholders --
# self-signed pairs are the normal case (the origin validates them, not the
# public). Rotate certificate and key TOGETHER: the zone-scoped API silently
# ignores a key-only change.
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareAuthenticatedOriginPullsCertificate
metadata:
  name: zone-client-certificate
spec:
  zone_id:
    value: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  scope: zone
  certificate: |
    -----BEGIN CERTIFICATE-----
    REPLACE_WITH_CLIENT_CERTIFICATE_PEM
    -----END CERTIFICATE-----
  private_key:
    value: |
      -----BEGIN PRIVATE KEY-----
      REPLACE_WITH_PRIVATE_KEY_PEM
      -----END PRIVATE KEY-----
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.zoneId` | `string \| valueFrom` | yes |  | CloudflareDnsZone (`status.outputs.zone_id`) |
| `spec.scope` | `string` |  | `zone` |  |
| `spec.certificate` | `string` | yes |  |  |
| `spec.privateKey` | `string \| valueFrom` (sensitive) | yes |  |  |

## Field Details

### spec.zoneId

`string | valueFrom` · required

The zone the certificate is uploaded to.
When using value_from, defaults to CloudflareDnsZone kind and status.outputs.zone_id field path.

- references: CloudflareDnsZone (`status.outputs.zone_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: CloudflareDnsZone, name: <that resource's name>, fieldPath: status.outputs.zone_id}} -- a bare string does not parse

### spec.scope

`string` · optional (explicit presence)

The upload's scope (Cloudflare's two distinct API surfaces): zone
replaces the zone-wide client certificate; hostname uploads a
certificate for per-hostname associations to reference.

- default: `zone`
- rule: {"string":{"in":["zone","hostname"]}}

### spec.certificate

`string` · required

The client certificate in PEM form. Must be a LEAF certificate --
Cloudflare rejects CA-flagged uploads (400 code 1412 "Missing leaf
certificate.", measured 2026-08-28; a self-signed cert is fine, but
mint it with basicConstraints CA:FALSE). Keep the PEM byte-stable
(trailing newline included) -- a formatting-only change still replaces
the upload.

- rule: {"required":true}

### spec.privateKey

`string | valueFrom` · required · sensitive

The certificate's private key in PEM form. Provide a managed-secret
reference; the platform resolves it just-in-time at deploy and never
stores it in plaintext. The API never returns the key, and on the
hostname surface it forces replacement -- an imported resource cannot
restore it, so the first post-import apply replaces the upload under a
new certificate id (re-point associations; see the import catalog).

- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

## Outputs

Reference an output from another manifest as `valueFrom: {kind: CloudflareAuthenticatedOriginPullsCertificate, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.certificate_id` | `string` | The ID of the uploaded certificate -- what per-hostname associations reference. |
| `status.outputs.zone_id` | `string` | The zone the certificate belongs to. |
| `status.outputs.expires_on` | `string` | When the certificate expires (RFC3339). |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.zoneId` | CloudflareDnsZone | `status.outputs.zone_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| CloudflareAuthenticatedOriginPulls | `spec.hostnameAssociations[].certificateId` | `status.outputs.certificate_id` |

## See Also

- [Overview](../README.md)
