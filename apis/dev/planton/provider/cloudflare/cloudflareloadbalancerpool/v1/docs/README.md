# CloudflareLoadBalancerPool — Technical Notes

A pool is the origin-grouping half of Cloudflare Load Balancing: account-scoped,
reusable, and referenced by zone-scoped load balancers.

## Composition

```
Monitor (account) --monitor--> Pool (account) --pool id--> LoadBalancer (zone)
        compute output --address valueFrom--> Pool.origins[]
```

- `monitor` is a `StringValueOrRef` to a `CloudflareLoadBalancerMonitor` — define a
  health check once, reuse it across pools.
- Each origin `address` is a `StringValueOrRef` with no fixed `default_kind`, so it
  accepts a literal IP/hostname or a reference to any resource's output (a compute
  instance's public IP, an upstream load balancer hostname, etc.).
- `pool_id` is consumed by `CloudflareLoadBalancer` (`default_pools`,
  `fallback_pool`, and the geo-pool maps).

## Defaults and the 0/null convention

Numeric "tuning" fields and the boolean toggles that default to true follow the
codebase convention: the proto marks the toggles `optional` (so unset is
distinguishable) and leaves a `recommended_default`; the IaC passes the value
through and lets the provider apply its own default when unset — origin `weight` 1,
`enabled` true, `flatten_cname` true, pool `enabled` true, `minimum_origins` 1.

## Origin header

Cloudflare supports a single `Host` override per origin. The spec models this as a
flat `host_header` string; both engines translate it to the provider's
`header { host = [<value>] }` shape.

## Parity

The Pulumi and Terraform modules are behaviorally identical: same origin
construction (including the host-header translation), same monitor/check-region
handling, same load-shedding / origin-steering / notification-filter mapping, and
the same `pool_id` / `pool_name` outputs.
