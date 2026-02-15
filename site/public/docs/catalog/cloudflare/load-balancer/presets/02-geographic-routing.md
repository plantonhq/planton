---
title: "Geographic Routing"
description: "Multiple origins with steering=geo: Cloudflare routes clients to the geographically nearest healthy origin. Use for multi-region deployments where latency matters (e.g., US, EU, APAC)."
type: "preset"
rank: "02"
presetSlug: "02-geographic-routing"
componentSlug: "load-balancer"
componentTitle: "Load Balancer"
provider: "cloudflare"
icon: "package"
order: 2
---

# Geographic Routing

Multiple origins with steering=geo: Cloudflare routes clients to the geographically nearest healthy origin. Use for multi-region deployments where latency matters (e.g., US, EU, APAC).

## When to Use

- Multi-region backends (US, EU, Asia)
- Latency-sensitive applications with regional origins
- Global traffic distribution by user location

## Key Configuration Choices

- **steeringPolicy: geo** (`steeringPolicy: geo`) -- Route by client geography to nearest origin.
- **origins** (`origins`) -- One per region; address can be IP or hostname.
- **zoneId** (`zoneId`) -- Zone ID; use value wrapper or reference to CloudflareDnsZone.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|-------------|-------------|---------------|
| `<cloudflare-zone-id>` | Zone ID containing hostname | CloudflareDnsZone status.outputs.zone_id |
| `www.example.com` | Hostname for the load balancer | Your application FQDN |
| `us-origin.example.com`, etc. | Origin hostnames or IPs per region | Your regional origin servers |

## Related Presets

- **01-active-passive-failover** -- Use for failover instead of geo-routing
- **03-weighted-ab-testing** -- Use for weighted traffic split
