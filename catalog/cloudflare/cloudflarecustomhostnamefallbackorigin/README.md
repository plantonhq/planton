# CloudflareCustomHostnameFallbackOrigin

Set the **default origin** for a Cloudflare for SaaS zone — the backend that all of
the zone's custom hostnames route to unless a hostname overrides it. It is a
zone-level singleton (one per zone) and a prerequisite for serving traffic to any
`CloudflareCustomHostname` in the zone.

## When to use

- Once per SaaS zone, before onboarding customer custom hostnames.

## Quick start

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareCustomHostnameFallbackOrigin
metadata:
  name: saas-fallback-origin
spec:
  zoneId:
    valueFrom:
      kind: CloudflareDnsZone
      name: saas-zone
      fieldPath: status.outputs.zone_id
  origin:
    value: origin.helpdesk.io
```

## Configuration reference

| Field | Required | Description |
|---|---|---|
| `zoneId` | yes | The SaaS zone ID (literal or `CloudflareDnsZone` reference) |
| `origin` | yes | The fallback origin hostname (a record within the SaaS zone) |

## Outputs

| Output | Description |
|---|---|
| `created_at` | Creation timestamp |
| `updated_at` | Last-update timestamp |
| `zone_id` | The SaaS zone this singleton belongs to (its API identity IS the zone) |

There is no `status` output: deployment is asynchronous (`pending_deployment` → `active`), and a point-in-time phase is never a stable output. There is no `errors` output for the same reason — Cloudflare fills and clears that diagnostic list asynchronously (measured live on the sibling custom hostname). Read both from the Cloudflare API or dashboard.

## Related kinds

- `CloudflareCustomHostname` — per-customer hostnames that route to this origin.
- `CloudflareDnsZone` — the SaaS zone.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
