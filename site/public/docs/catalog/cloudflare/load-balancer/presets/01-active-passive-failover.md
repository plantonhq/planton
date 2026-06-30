---
title: "Active-Passive Failover"
description: "A monitor, a single pool with a primary and secondary origin, and a load balancer with `steeringPolicy: off`. Traffic goes to the first healthy origin; if it fails health checks, it fails over to the..."
type: "preset"
rank: "01"
presetSlug: "01-active-passive-failover"
componentSlug: "load-balancer"
componentTitle: "Load Balancer"
provider: "cloudflare"
icon: "package"
order: 1
---

# Active-Passive Failover

A monitor, a single pool with a primary and secondary origin, and a load balancer
with `steeringPolicy: off`. Traffic goes to the first healthy origin; if it fails
health checks, it fails over to the next. Proxied through Cloudflare for DDoS
protection and CDN.

## When to Use

- Primary + backup servers (e.g., main site, standby)
- DR failover when the primary becomes unhealthy
- Simple active-passive redundancy

## Key Configuration Choices

- **`steeringPolicy: off`** — static failover; the first healthy origin gets all traffic.
- **`proxied: true`** — recommended; traffic flows through Cloudflare.
- **Monitor `path`/`expectedCodes`** — tune the health check to a real endpoint.
- **`zoneId`** — a value or a reference to a `CloudflareDnsZone`.

## Placeholders to Replace

| Placeholder | Description |
|---|---|
| `<cloudflare-account-id>` | Account that owns the monitor and pool |
| `<cloudflare-zone-id>` | Zone containing the hostname |
| `<app-subdomain>.<your-domain.com>` | Load balancer hostname |
| `192.0.2.1`, `192.0.2.2` | Primary and secondary origin addresses |

## Related Presets

- **02-geographic-routing** — route by geography across regional pools
- **03-weighted-ab-testing** — split traffic across pools by weight
