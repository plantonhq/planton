# Cloudflare Authenticated Origin Pulls

## Overview

`CloudflareAuthenticatedOriginPulls` manages a zone's Authenticated Origin Pulls enablement: the zone-wide toggle that makes Cloudflare present a client certificate when pulling from your origin, and the per-hostname associations that pin specific hostnames to specific uploaded certificates.

The origin server completes the story by requiring and validating that client certificate -- Cloudflare's side alone protects nothing. Free-plan feature.

## Key Features

- **Zone-wide toggle** -- one switch enabling client-certificate presentation for the whole zone (`zone_enabled`)
- **Per-hostname pinning** -- associations mapping hostnames to uploaded client certificates (`CloudflareAuthenticatedOriginPullsCertificate`)
- **Independent surfaces** -- manage the toggle, the associations, or both; unset means "leave it alone"
- **Chart-ready wiring** -- association rows reference the certificate kind's output by default

## Use Cases

**Ideal for:**

- Ensuring your origin only answers connections that genuinely come from Cloudflare
- Different client certificates per hostname (multi-tenant origins)
- Completing the mTLS story alongside `CloudflareMtlsCertificate` CAs

**Not ideal for:**

- Uploading the client certificates themselves -- that is `CloudflareAuthenticatedOriginPullsCertificate`
- Validating VISITOR client certificates -- that is API Shield mTLS, a different surface

## API Specification

### Required Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `zone_id` | StringValueOrRef | Yes | The zone whose AOP surface is managed. Can reference a `CloudflareDnsZone` via `value_from` (defaults to `status.outputs.zone_id`). |

At least one of `zone_enabled` / `hostname_associations` must be set -- an empty spec manages nothing.

### Optional Fields

| Field | Type | Description |
|-------|------|-------------|
| `zone_enabled` | bool | The zone-wide toggle. Unset = not managed; set = asserted. |
| `hostname_associations` | list | Rows of `{hostname, certificate_id, enabled}`. `certificate_id` is REQUIRED and references a `CloudflareAuthenticatedOriginPullsCertificate` -- Cloudflare rejects an association write without a certificate id (400 code 1404), even when zone-level certificate material exists. `enabled` unset means active -- both engines send true, because Cloudflare treats null as "void the association". |

### Outputs

| Field | Description |
|-------|-------------|
| `zone_id` | The zone whose AOP surface is managed (the surface is zone-singleton shaped) |

## Example Manifest

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareAuthenticatedOriginPulls
metadata:
  name: www-origin-pulls
spec:
  zone_id:
    value: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  zone_enabled: true
  hostname_associations:
    - hostname: app.example.com
      certificate_id:
        valueFrom:
          kind: CloudflareAuthenticatedOriginPullsCertificate
          name: app-client-certificate
```

## Destroy Semantics

Neither surface deletes: the zone-wide toggle has NO delete at Cloudflare (destroy abandons the live value), and an association is removed by a revert write (`enabled: null`) issued from state. Disable explicitly before destroying when the zone outlives this resource.

## Related Resources

- **CloudflareAuthenticatedOriginPullsCertificate** -- the uploaded client certificates association rows reference
- **CloudflareMtlsCertificate** -- account-level CA material for per-hostname validation
- **CloudflareDnsZone** -- `zone_id` foreign key

## Further Reading

For operational judgment -- the no-op destroy class, the revert-write semantics, the origin-side half of the story -- see GUIDE.md.

## References

- [Cloudflare Authenticated Origin Pulls](https://developers.cloudflare.com/ssl/origin-configuration/authenticated-origin-pull/)

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
