# Cloudflare Load Balancer Pool

Group origin servers into a reusable, health-checked pool that load balancers
steer traffic to.

## What Gets Created

- A `cloudflare_load_balancer_pool` (account-scoped) with its origins, optional
  monitor reference, check regions, load-shedding, origin steering, and
  notification filters.

## Prerequisites

- A Cloudflare account ID.
- Optionally, a `CloudflareLoadBalancerMonitor` to health-check the origins.

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
