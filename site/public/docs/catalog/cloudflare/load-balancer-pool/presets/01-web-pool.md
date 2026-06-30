---
title: "Preset: Web pool with two origins"
description: "A pool of two web origins health-checked by an HTTPS monitor, ready to attach to a load balancer's `default_pools`."
type: "preset"
rank: "01"
presetSlug: "01-web-pool"
componentSlug: "load-balancer-pool"
componentTitle: "Load Balancer Pool"
provider: "cloudflare"
icon: "package"
order: 1
---

# Preset: Web pool with two origins

A pool of two web origins health-checked by an HTTPS monitor, ready to attach to a
load balancer's `default_pools`.

## When to use

- You have two or more interchangeable web/API backends and want Cloudflare to
  balance across the healthy ones.

## Key choices

- `origins[].address`: literal IPs/hostnames, or `valueFrom` references to compute
  outputs so the pool tracks backend changes in the graph.
- `monitor`: reference a `CloudflareLoadBalancerMonitor` so origins are actively
  health-checked (omit only for always-healthy origins).
- `minimumOrigins`: keep at 1, or raise it to fail the pool over before it is fully
  drained.

## Placeholders

| Placeholder | Description |
|---|---|
| `<cloudflare-account-id>` | 32-character Cloudflare account ID |
| `<origin-1-address>` / `<origin-2-address>` | Origin IPs or hostnames |

## Attaching it to a load balancer

```yaml
defaultPools:
  - valueFrom:
      kind: CloudflareLoadBalancerPool
      name: web-pool
      fieldPath: status.outputs.pool_id
fallbackPool:
  valueFrom:
    kind: CloudflareLoadBalancerPool
    name: web-pool
    fieldPath: status.outputs.pool_id
```
