---
title: "Load Balancer Pool"
description: "Load Balancer Pool deployment documentation"
icon: "package"
order: 100
componentName: "cloudflareloadbalancerpool"
---

# Cloudflare Load Balancer Pool

Group origin servers into a reusable, health-checked pool that load balancers
steer traffic to.

## What Gets Created

- A `cloudflare_load_balancer_pool` (account-scoped) with its origins, optional
  monitor reference, check regions, load-shedding, origin steering, and
  notification filters.

## Prerequisites

- A Cloudflare account ID.
- The **Load Balancing add-on** enabled on the account (paid add-on) — otherwise the
  Load Balancing API returns `403`.
- An API token with **Account → Load Balancing: Monitors and Pools → Edit** (pools are
  account-scoped).
- Optionally, a `CloudflareLoadBalancerMonitor` to health-check the origins. When a
  monitor is attached, origin addresses must be **globally routable** (Cloudflare
  rejects reserved / non-routable IPs), and `checkRegions` is **capped by plan tier**.

## Configuration Reference

**Required**

- `accountId`, `name`, `origins[]` (each with `name` + `address`).

**Optional**

- `origins[]`: `weight`, `enabled`, `port`, `hostHeader`, `virtualNetworkId`,
  `flattenCname`.
- `monitor`, `checkRegions`, `enabled`, `minimumOrigins`, `latitude`, `longitude`,
  `loadShedding`, `originSteering`, `notificationFilter`.

## Stack Outputs

| Output | Description |
|---|---|
| `pool_id` | The pool ID |
| `pool_name` | The pool name |

## Related Components

- CloudflareLoadBalancerMonitor
- CloudflareLoadBalancer
