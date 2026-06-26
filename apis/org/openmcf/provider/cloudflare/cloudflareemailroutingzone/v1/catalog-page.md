# Cloudflare Email Routing (Zone)

Enable Email Routing on a zone and (optionally) configure its catch-all rule.

## What Gets Created

- `cloudflare_email_routing_settings` (enables routing; provisions required DNS).
- `cloudflare_email_routing_catch_all` (folded; only when `catchAll` is set).
- `cloudflare_email_routing_dns` (only when `lockDnsRecords` is true).

## Prerequisites

- A Cloudflare zone (a `CloudflareDnsZone`).

## Configuration Reference

**Required**

- `zoneId` — zone ID, or a reference to a `CloudflareDnsZone`.

**Optional**

- `catchAll` — folded catch-all rule.
- `lockDnsRecords` — lock the Email Routing DNS records.

## Stack Outputs

| Output | Description |
|---|---|
| `zone_id` | The zone ID |
| `enabled` | Whether Email Routing is enabled |
| `status` | Configuration status |
| `name` | The zone's domain name |

## Related Components

- CloudflareDnsZone
- CloudflareEmailRoutingRule
- CloudflareEmailRoutingAddress
