# Cloudflare mTLS Certificate

## Overview

`CloudflareMtlsCertificate` uploads a certificate to the account-level mTLS certificate store. A CA certificate (`ca: true`) is what Authenticated Origin Pulls rows, zone TLS CA associations, and Workers mTLS bindings reference to validate client certificates; a leaf certificate (`ca: false`) is presented BY Cloudflare as a client. Self-signed CAs are the normal case -- this store holds YOUR trust material, not publicly trusted certificates.

Every field is create-only at the API: any change replaces the upload and the certificate ID changes.

## Key Features

- **Account-scoped trust store** -- one upload, referenced by many zone-level consumers
- **CA or leaf** -- `ca: true` for validating clients; `ca: false` (with a private key) for Cloudflare presenting the certificate itself
- **Self-signed is normal** -- your infrastructure validates these, not the public
- **Free** -- no plan gate on the store itself

## Use Cases

**Ideal for:**

- The CA behind per-hostname Authenticated Origin Pulls (referenced through `CloudflareZoneTlsSettings` CA associations)
- The CA a Workers mTLS binding presents to an origin
- Rotating client-auth trust material on your own calendar

**Not ideal for:**

- Visitor-facing zone certificates -- that is `CloudflareCustomSslCertificate`
- The AOP client certificate itself -- that is `CloudflareAuthenticatedOriginPullsCertificate` (zone-scoped uploads)

## API Specification

### Required Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `account_id` | string | Yes | The Cloudflare account (32-hex). |
| `ca` | bool | Yes | CA certificate (true) or leaf (false). Must be stated explicitly. |
| `certificates` | string | Yes | The certificate (or CA chain) in PEM form. Keep it byte-stable. |

### Optional Fields

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | Display name in the dashboard. |
| `private_key` | StringValueOrRef (sensitive) | Only when Cloudflare must present this certificate itself (leaf). CA uploads validating clients carry no key. The API never returns it. |

### Outputs

| Field | Description |
|-------|-------------|
| `certificate_id` | The uploaded certificate's ID -- what consumers reference |
| `expires_on` | When the certificate expires (RFC3339) |
| `serial_number` | The certificate's serial number |

## Example Manifest

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareMtlsCertificate
metadata:
  name: origin-pull-ca
spec:
  account_id: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  name: origin-pull-ca
  ca: true
  certificates: |
    -----BEGIN CERTIFICATE-----
    ...
    -----END CERTIFICATE-----
```

## Destroy Semantics

Destroy is a real delete. Consumers referencing the certificate ID (CA associations, Workers bindings) lose their trust anchor -- re-point them BEFORE destroying.

## Rotation

Every field is create-only: rotate by uploading a new certificate (a new resource or a changed value forces replacement), re-pointing consumers at the new `certificate_id`, then destroying the old upload.

## Related Resources

- **CloudflareZoneTlsSettings** -- CA hostname associations reference this kind's `certificate_id`
- **CloudflareAuthenticatedOriginPulls** -- the zone-level enablement this trust material serves
- **CloudflareWorker** -- mTLS bindings reference uploads from this store

## Further Reading

For operational judgment -- replace-on-rotate discipline, CA vs leaf, the consumer re-point order -- see GUIDE.md.

## References

- [Cloudflare mTLS](https://developers.cloudflare.com/ssl/client-certificates/)

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
