# CloudflareCustomHostname

Onboard a customer's own domain onto a Cloudflare for SaaS zone. This extends
Cloudflare's edge — TLS termination, caching, WAF — onto a hostname your customer
owns (e.g. `support.acme.com`), with a per-customer certificate that Cloudflare
provisions and auto-renews. The customer CNAMEs their hostname to your SaaS zone and
proves control via the ownership-verification records this component exports.

## When to use

- You run a SaaS platform and want customers to use the product on their own
  white-label/vanity domains with valid HTTPS.

## Quick start

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareCustomHostname
metadata:
  name: acme-vanity-domain
spec:
  zoneId:
    valueFrom:
      kind: CloudflareDnsZone
      name: saas-zone
      fieldPath: status.outputs.zone_id
  hostname: support.acme.com
  ssl:
    method: txt
    type: dv
```

## Configuration reference

| Field | Required | Description |
|---|---|---|
| `zoneId` | yes | The SaaS zone ID (literal or `CloudflareDnsZone` reference) |
| `hostname` | yes | The customer's hostname to onboard |
| `customOriginServer` | no | Override the origin for this hostname (backend endpoint) |
| `customOriginSni` | no | SNI sent to the custom origin |
| `customMetadata` | no | Arbitrary key/value metadata |
| `ssl` | no | Certificate + TLS termination settings (see catalog page) |

## Outputs

| Output | Description |
|---|---|
| `custom_hostname_id` | The custom hostname identifier |
| `ownership_verification_name` / `_type` / `_value` | The DNS record the customer adds to verify control |
| `ownership_verification_http_url` / `_http_body` | The HTTP alternative for verification |
| `created_at` | Creation timestamp |
| `zone_id` | The SaaS zone the hostname was onboarded onto (its API identity is `zone_id` + `custom_hostname_id`) |

There is no `status` output: activation is asynchronous (`pending` → `pending_validation` → `active`), and a point-in-time phase is never a stable output. There is no `verification_errors` output for the same reason: Cloudflare populates that list asynchronously after apply (measured live: "zone is not active yet" appears seconds post-create) and clears it on activation. Read both from the Cloudflare API or dashboard.

## Prerequisites

This kind requires the zone to have a fallback origin
(`CloudflareCustomHostnameFallbackOrigin`) configured. Express that dependency with
a `metadata.relationships` `depends_on` edge in an Infra Chart.

## Related kinds

- `CloudflareCustomHostnameFallbackOrigin` — the zone's default origin (prerequisite).
- `CloudflareDnsZone` — the SaaS zone.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
