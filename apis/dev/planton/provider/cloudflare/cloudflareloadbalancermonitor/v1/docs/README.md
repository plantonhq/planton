# CloudflareLoadBalancerMonitor — Technical Notes

A monitor is the health-checking half of Cloudflare Load Balancing. It is
account-scoped and reusable across pools.

## How monitors fit together

```
Monitor (account)  ->  Pool.monitor (account)  ->  LoadBalancer.default_pools (zone)
```

A pool references exactly one monitor (via `monitor`). Cloudflare runs the monitor
against each origin in the pool from the pool's `check_regions` (or everywhere) and
marks origins up/down according to `consecutive_up`/`consecutive_down`, `retries`,
and `timeout`. When healthy origins fall below the pool's `minimum_origins`, the
pool fails over.

## Protocol-specific fields

- **http / https**: `path`, `expected_codes`, `expected_body`, `method`,
  `headers` (set a `Host` header so the origin routes correctly), `follow_redirects`,
  `probe_zone`. `allow_insecure` applies to https only.
- **tcp / udp_icmp / smtp**: require a `port`. Application-layer fields are ignored.
- **icmp_ping**: pure reachability; no port, no application-layer fields.

The proto enforces "port required for tcp/udp_icmp/smtp" with a message-level CEL
rule; the IaC omits any numeric tuning knob left at 0 so the provider applies its
own default (interval 60, timeout 5, retries 2).

## Outputs

- `monitor_id` — referenced by `CloudflareLoadBalancerPool.monitor`.
- `monitor_type` — the configured protocol.

## Parity

The Pulumi (`iac/pulumi/module`) and Terraform (`iac/tf`) modules are behaviorally
identical: same account scoping, same default handling (0 -> provider default),
same header map shape, and the same two stack outputs.
