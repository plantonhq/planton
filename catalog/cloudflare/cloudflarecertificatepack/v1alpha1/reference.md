# CloudflareCertificatePack

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `cloudflare.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

CloudflareCertificatePackSpec orders an advanced edge certificate for a zone:
a publicly-trusted TLS certificate, provisioned and auto-renewed by Cloudflare,
that covers the hostnames you list (beyond the free Universal SSL certificate).
Use it when you need a specific certificate authority, multiple/longer-lived
certs, or coverage for hostnames Universal SSL does not include.

Most attributes are immutable: changing them re-orders (replaces) the pack.
Allowed-value sets are validated with CEL using the provider's exact strings so
the modules pass them through verbatim with no lossy enum mapping.

## Example

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareCertificatePack
metadata:
  name: test-cert-pack
spec:
  zone_id:
    value: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  certificateAuthority: google
  type: advanced
  validationMethod: txt
  validityDays: 90
  hosts:
    - example.com
    - "*.example.com"
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.zoneId` | `string \| valueFrom` | yes |  | CloudflareDnsZone (`status.outputs.zone_id`) |
| `spec.certificateAuthority` | `string` | yes |  |  |
| `spec.type` | `string` |  | `advanced` |  |
| `spec.validationMethod` | `string` | yes |  |  |
| `spec.validityDays` | `int64` | yes |  |  |
| `spec.hosts` | `[]string` | yes |  |  |
| `spec.cloudflareBranding` | `bool` |  |  |  |

## Field Details

### spec.zoneId

`string | valueFrom` · required

The Cloudflare Zone ID the certificate is ordered for. A literal value or a
reference to a CloudflareDnsZone resource (its status.outputs.zone_id).

- references: CloudflareDnsZone (`status.outputs.zone_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: CloudflareDnsZone, name: <that resource's name>, fieldPath: status.outputs.zone_id}} -- a bare string does not parse

### spec.certificateAuthority

`string` · required

The certificate authority that issues the certificate: "google",
"lets_encrypt", or "ssl_com". CA-specific restrictions apply (see Cloudflare's
certificate-authorities reference).

- rule: certificate_authority must be one of google, lets_encrypt, ssl_com
- rule: {"required":true}

### spec.type

`string` · optional (explicit presence)

The certificate pack type. Only "advanced" is supported. Defaults to
"advanced".

- default: `advanced`
- rule: type must be advanced

### spec.validationMethod

`string` · required

The domain control validation (DCV) method: "txt", "http", or "email". For a
zone using Cloudflare's nameservers, "txt" validation is completed
automatically.

- rule: validation_method must be one of txt, http, email
- rule: {"required":true}

### spec.validityDays

`int64` · required

How long the certificate is valid, in days: 14, 30, 90, or 365.

- rule: validity_days must be one of 14, 30, 90, 365
- rule: {"required":true}

### spec.hosts

`[]string` · required

The hostnames the certificate covers. Must include the zone apex, may not
exceed 50 hosts, and may not be empty (e.g. "example.com" and "*.example.com").

- rule: {"repeated":{"minItems":"1","maxItems":"50","items":{"string":{"minLen":"1"}}}}

### spec.cloudflareBranding

`bool`

Add Cloudflare branding to the order, which uses a subdomain of
sni.cloudflaressl.com as the Common Name. Defaults to false.

## Outputs

Reference an output from another manifest as `valueFrom: {kind: CloudflareCertificatePack, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.certificate_pack_id` | `string` | The certificate pack identifier. |
| `status.outputs.zone_id` | `string` | The Cloudflare Zone ID the pack was ordered in. A pack's API identity is (zone_id, certificate_pack_id), so downstream consumers -- verification tooling, imports, chart blocks composing on the pack -- need the zone alongside the pack's own id. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.zoneId` | CloudflareDnsZone | `status.outputs.zone_id` |

## See Also

- [Overview](../README.md)
