---
title: "Geographic Routing"
description: "A monitor, two regional pools (US and EU), and a load balancer with `steeringPolicy: geo` and `regionPools` mapping regions to pools. Users are routed to the nearest healthy region, falling back to..."
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

A monitor, two regional pools (US and EU), and a load balancer with
`steeringPolicy: geo` and `regionPools` mapping regions to pools. Users are routed
to the nearest healthy region, falling back to `defaultPools` when no mapping
matches.

## When to Use

- Multi-region deployments where latency matters
- Data-residency routing (serve EU users from EU origins)

## Key Configuration Choices

- **`steeringPolicy: geo`** — route by `regionPools`/`countryPools`/`popPools`.
- **`regionPools`** — map each region code (e.g. `WNAM`, `WEU`) to its pool.
- **`fallbackPool`** — the pool of last resort if every mapped pool is unhealthy.
- One reusable monitor health-checks both pools.

## Placeholders to Replace

| Placeholder | Description |
|---|---|
| `<cloudflare-account-id>` | Account that owns the monitor and pools |
| `<cloudflare-zone-id>` | Zone containing the hostname |
| `<www-subdomain>.<your-domain.com>` | Load balancer hostname |
| `<us-east-origin-ip-or-hostname>` / `<eu-west-origin-ip-or-hostname>` | Regional origin addresses |

## Related Presets

- **01-active-passive-failover** — simple primary/backup failover
- **03-weighted-ab-testing** — split traffic across pools by weight
