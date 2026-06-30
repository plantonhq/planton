# Cloudflare Email Routing Rule

Route incoming mail for a zone: match by recipient (or all) and drop, forward, or
hand to an Email Worker.

## What Gets Created

- A `cloudflare_email_routing_rule` on the zone.

## Prerequisites

- Email Routing enabled on the zone (a `CloudflareEmailRoutingZone`).
- Verified destination addresses for `forward` actions.

## Configuration Reference

**Required**

- `zoneId` — zone ID or a reference to a `CloudflareDnsZone`.
- `matchers` — one or more match patterns.
- `action` — drop / forward / worker.

**Optional**

- `name`, `enabled` (default true), `priority` (default 0).

## Stack Outputs

| Output | Description |
|---|---|
| `rule_id` | The routing rule identifier |
| `zone_id` | The zone the rule belongs to |

## Related Components

- CloudflareEmailRoutingZone
- CloudflareEmailRoutingAddress
- CloudflareWorker
