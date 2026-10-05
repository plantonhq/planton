# Cloudflare Custom SSL Certificate

## Overview

`CloudflareCustomSslCertificate` uploads a bring-your-own TLS certificate to a zone: Cloudflare presents this certificate to visitors instead of a Universal SSL or Advanced certificate. The certificate must be issued by a certificate authority on Cloudflare's trust list -- self-signed certificates are rejected at the API.

Custom certificates are a Business/Enterprise zone feature. Cloudflare enforces the plan gate at create -- the API, not this spec, is the wall.

## Key Features

- **Bring your own certificate** -- full control over issuer, subject, and key when compliance demands it
- **SNI or legacy** -- `sni_custom` (modern clients, multiple uploads) or `legacy_custom` (every TLS client, one slot per zone)
- **Chain bundling** -- `ubiquitous` (default, maximum compatibility), `optimal`, or `force` (chain exactly as uploaded)
- **Private-key geo control** -- a `policy` expression or a coarse `geo_restrictions` label (`us`, `eu`, `highest_security`)
- **Staging first** -- `deploy: staging` validates the upload on Cloudflare's staging network before production

## Use Cases

**Ideal for:**

- Compliance regimes that mandate a specific issuer or an EV/OV certificate
- Keeping the private key out of specific geographies (`geo_restrictions`, `policy`)
- Zones whose certificate lifecycle is managed by an external PKI team

**Not ideal for:**

- Ordinary zones -- Universal SSL and `CloudflareCertificatePack` (Advanced) renew themselves; custom certificates are YOUR renewal burden
- Origin-facing client certificates -- that is `CloudflareAuthenticatedOriginPullsCertificate`

## API Specification

### Required Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `zone_id` | StringValueOrRef | Yes | The zone the certificate is uploaded to. Can reference a `CloudflareDnsZone` via `value_from` (defaults to `status.outputs.zone_id`). |
| `certificate` | string | Yes | The certificate in PEM form, publicly trusted, covering the zone's hostnames. Keep it byte-stable -- a formatting-only change still replaces the upload. |
| `private_key` | StringValueOrRef (sensitive) | Yes | The private key in PEM form. Provide a managed-secret reference; the API never returns it. |

### Optional Fields

| Field | Type | Description |
|-------|------|-------------|
| `type` | string | `legacy_custom` (default) or `sni_custom`. Changing it on an existing upload replaces it. |
| `bundle_method` | string | `ubiquitous` (default), `optimal`, or `force`. |
| `policy` | string | Geo-policy expression for the private key (e.g. `(country: US) or (region: EU)`). |
| `geo_restrictions` | object | Coarse geo label: `us`, `eu`, or `highest_security`. |
| `custom_csr_id` | string | The Cloudflare-generated CSR this certificate was issued from, when the CSR flow was used. |
| `deploy` | string | `staging` or `production`. Staging is a Business/Enterprise validation surface. |

### Outputs

| Field | Description |
|-------|-------------|
| `certificate_id` | The uploaded certificate's ID |
| `zone_id` | The zone the certificate belongs to |
| `expires_on` | When the certificate expires (RFC3339) |

Deployment status is deliberately not an output: deployment is asynchronous (pending before active), so a point-in-time phase would flip on the first refresh and re-plan forever. Read it from the Cloudflare API or dashboard.

## Example Manifest

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareCustomSslCertificate
metadata:
  name: www-custom-certificate
spec:
  zone_id:
    value: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  certificate: |
    -----BEGIN CERTIFICATE-----
    ...
    -----END CERTIFICATE-----
  private_key:
    value: ${secrets-group/prod-tls/www-key}
  type: sni_custom
```

## Destroy Semantics

Destroy is a real delete (deployment states settle asynchronously). The zone falls back to its Universal SSL / Advanced certificate -- have one active before destroying, or visitors get handshake failures.

## Rotation

Rotation is replacement: changing the certificate or private key destroys and re-creates the upload, and the certificate ID changes. Cloudflare serves the previous certificate until the replacement deploys, so rotation is not an outage. Certificate priority is NOT manageable at provider v5.23.0 (the field is read-only).

## Related Resources

- **CloudflareCertificatePack** -- Cloudflare-managed (Advanced) certificates that renew themselves
- **CloudflareZoneTlsSettings** -- the zone's TLS posture (minimum TLS, Universal SSL toggle)
- **CloudflareDnsZone** -- `zone_id` foreign key

## Further Reading

For operational judgment -- the Business plan wall, the trust-list rejection class, rotation discipline -- see GUIDE.md.

## References

- [Cloudflare Custom Certificates](https://developers.cloudflare.com/ssl/edge-certificates/custom-certificates/)

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
