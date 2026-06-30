---
title: "Weighted A/B Testing"
description: "A monitor, two pools (control and variant), and a load balancer with `steeringPolicy: random` plus `randomSteering` weights. Cloudflare selects a pool at random in proportion to the configured..."
type: "preset"
rank: "03"
presetSlug: "03-weighted-ab-testing"
componentSlug: "load-balancer"
componentTitle: "Load Balancer"
provider: "cloudflare"
icon: "package"
order: 3
---

# Weighted A/B Testing

A monitor, two pools (control and variant), and a load balancer with
`steeringPolicy: random` plus `randomSteering` weights. Cloudflare selects a pool
at random in proportion to the configured weights — ideal for canary releases and
A/B experiments.

## When to Use

- Canary rollouts: send a small share of traffic to a new version
- A/B experiments across two backends

## Key Configuration Choices

- **`steeringPolicy: random`** — weighted random pool selection.
- **`randomSteering.defaultWeight`** — base weight applied to pools not listed in
  `poolWeights`; set per-pool weights via `randomSteering.poolWeights` (keyed by
  pool ID) for finer control.
- **`defaultPools`** lists both pools; `fallbackPool` is the safe control pool.

## Placeholders to Replace

| Placeholder | Description |
|---|---|
| `<cloudflare-account-id>` | Account that owns the monitor and pools |
| `<cloudflare-zone-id>` | Zone containing the hostname |
| `<app-subdomain>.<your-domain.com>` | Load balancer hostname |
| `<control-origin-ip-or-hostname>` / `<variant-origin-ip-or-hostname>` | Origin addresses |

## Related Presets

- **01-active-passive-failover** — simple primary/backup failover
- **02-geographic-routing** — route by geography across regional pools
