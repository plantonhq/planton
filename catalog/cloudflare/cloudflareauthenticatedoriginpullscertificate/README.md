# Cloudflare Authenticated Origin Pulls Certificate

## Overview

`CloudflareAuthenticatedOriginPullsCertificate` uploads the client certificate Cloudflare presents to YOUR origin when Authenticated Origin Pulls is enabled. Self-signed certificates are the normal case -- the origin, not the public, validates this certificate.

The scope decides the blast radius: a `zone` upload replaces Cloudflare's shared certificate for the whole zone; a `hostname` upload is referenced per hostname through `CloudflareAuthenticatedOriginPulls` associations. Free-plan feature.

## Key Features

- **Two upload surfaces, one kind** -- `scope: zone` (zone-wide replacement) or `scope: hostname` (per-hostname material)
- **Self-signed is normal** -- your origin's trust store validates it
- **Chart-ready output** -- `certificate_id` is what association rows reference
- **Asynchronous lifecycle** -- deployment and deletion settle through pending states

## Use Cases

**Ideal for:**

- Replacing Cloudflare's shared zone certificate with your own client certificate
- Per-hostname client certificates consumed by `CloudflareAuthenticatedOriginPulls` associations
- Rotating origin-pull credentials on your own calendar

**Not ideal for:**

- Visitor-facing certificates -- that is `CloudflareCustomSslCertificate`
- Account-level CA trust material -- that is `CloudflareMtlsCertificate`

## API Specification

### Required Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `zone_id` | StringValueOrRef | Yes | The zone the certificate is uploaded to. Can reference a `CloudflareDnsZone` via `value_from` (defaults to `status.outputs.zone_id`). |
| `certificate` | string | Yes | The client certificate in PEM form. Keep it byte-stable. |
| `private_key` | StringValueOrRef (sensitive) | Yes | The private key in PEM form. Provide a managed-secret reference; the API never returns it. |

### Optional Fields

| Field | Type | Description |
|-------|------|-------------|
| `scope` | string | `zone` (default) or `hostname` -- selects which Cloudflare upload surface receives the certificate. |

### Outputs

| Field | Description |
|-------|-------------|
| `certificate_id` | The uploaded certificate's ID -- what per-hostname associations reference |
| `zone_id` | The zone the certificate belongs to |
| `expires_on` | When the certificate expires (RFC3339) |

Deployment status is deliberately not an output: deployment and deletion are asynchronous (pending_deployment to active seconds after create), so a point-in-time phase would flip on the first refresh and re-plan forever. Read it from the Cloudflare API or dashboard.

## Example Manifest

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareAuthenticatedOriginPullsCertificate
metadata:
  name: app-client-certificate
spec:
  zone_id:
    value: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  scope: hostname
  certificate: |
    -----BEGIN CERTIFICATE-----
    ...
    -----END CERTIFICATE-----
  private_key:
    value: ${secrets-group/prod-aop/app-client-key}
```

## Destroy Semantics

Destroy is a real delete on both surfaces, settling asynchronously (the API answers 200 with `pending_deletion`/`deleted` before the record goes). Associations referencing a deleted hostname certificate stop authenticating -- revert or re-point them first.

## Rotation

Rotation is replacement. Never rotate ONLY the private key against the same certificate: the zone-scoped API silently ignores a key-only change (a measured provider defect at v5.23.0) -- key and certificate always change together.

## Related Resources

- **CloudflareAuthenticatedOriginPulls** -- the enablement whose association rows reference this kind's `certificate_id`
- **CloudflareMtlsCertificate** -- the account-level CA that validates per-hostname client certificates
- **CloudflareDnsZone** -- `zone_id` foreign key

## Further Reading

For operational judgment -- the key-rotation defect, scope blast radius, async deletion -- see GUIDE.md.

## References

- [Cloudflare Authenticated Origin Pulls setup](https://developers.cloudflare.com/ssl/origin-configuration/authenticated-origin-pull/set-up/)

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
