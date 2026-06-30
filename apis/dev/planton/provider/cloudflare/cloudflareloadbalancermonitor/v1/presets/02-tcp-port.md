# Preset: TCP port health check

A TCP monitor that checks whether a port accepts connections — suitable for
non-HTTP origins (databases, message brokers, custom TCP services) behind a
Cloudflare Spectrum / TCP load balancer.

## When to use

- Origins speak a non-HTTP protocol; you only need connection-level health.

## Key choices

- `port`: required for TCP monitors (e.g. 5432 for PostgreSQL, 6379 for Redis).
- Application-layer fields (`path`, `expectedCodes`, headers) do not apply.

## Placeholders

| Placeholder | Description |
|---|---|
| `<cloudflare-account-id>` | 32-character Cloudflare account ID |

## Referencing it from a pool

```yaml
monitor:
  valueFrom:
    kind: CloudflareLoadBalancerMonitor
    name: db-tcp
    fieldPath: status.outputs.monitor_id
```
