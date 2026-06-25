# CloudflareLoadBalancer

A zone-scoped Cloudflare Load Balancer: Global Server Load Balancing (GSLB) that
attaches a DNS hostname to one or more **pools** and steers traffic across them
with health-aware failover, geo-routing, weighted distribution, and session
affinity.

## The three resources

Cloudflare models load balancing as three resources with distinct scopes and
lifecycles, and OpenMCF mirrors that with three composable kinds:

| Kind | Scope | Role |
|---|---|---|
| `CloudflareLoadBalancerMonitor` | account | Health check that probes origins |
| `CloudflareLoadBalancerPool` | account | A group of origins, health-checked by a monitor |
| `CloudflareLoadBalancer` | zone | Attaches a hostname to pools and steers traffic |

Pools and monitors are **account-scoped and reusable** — one pool can back many
load balancers, and one monitor can health-check many pools. This load balancer
references pools by ID or `valueFrom`; define the pools and monitor as their own
resources.

## Quick start

```yaml
# 1. A health check (account-scoped, reusable)
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareLoadBalancerMonitor
metadata:
  name: web-https
spec:
  accountId: 0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d
  type: https
  path: /health
  expectedCodes: "2xx"
---
# 2. A pool of origins, health-checked by the monitor
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareLoadBalancerPool
metadata:
  name: web-pool
spec:
  accountId: 0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d
  name: web-pool
  origins:
    - name: primary
      address:
        value: 203.0.113.10
    - name: secondary
      address:
        value: 198.51.100.20
  monitor:
    valueFrom:
      kind: CloudflareLoadBalancerMonitor
      name: web-https
      fieldPath: status.outputs.monitor_id
---
# 3. The load balancer (zone-scoped) referencing the pool
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareLoadBalancer
metadata:
  name: api-lb
spec:
  hostname: api.example.com
  zoneId:
    valueFrom:
      kind: CloudflareDnsZone
      name: example-zone
      fieldPath: status.outputs.zone_id
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
  proxied: true
  sessionAffinity: cookie
  steeringPolicy: off
```

## Configuration reference

| Field | Required | Description |
|---|---|---|
| `hostname` | yes | DNS hostname for the load balancer |
| `zoneId` | yes | Zone that owns the hostname (ID or `CloudflareDnsZone` ref) |
| `defaultPools[]` | yes | Ordered pools by failover priority (pool IDs/refs) |
| `fallbackPool` | yes | Pool of last resort when all others are unhealthy |
| `proxied` | no | Proxy through Cloudflare (orange cloud); default false |
| `enabled` | no | Enable the load balancer; default true |
| `steeringPolicy` | no | `off` (default), `geo`, `random`, `dynamic_latency`, `proximity`, `least_outstanding_requests`, `least_connections` |
| `sessionAffinity` | no | `none` (default), `cookie`, `ip_cookie`, `header` |
| `sessionAffinityTtl` | no | Affinity session expiry (seconds) |
| `sessionAffinityAttributes` | no | Drain, headers, cookie flags, zero-downtime failover |
| `ttl` | no | DNS TTL (gray-clouded only) |
| `regionPools` / `countryPools` / `popPools` | no | Geo code -> ordered pool list |
| `adaptiveRouting` | no | Zero-downtime failover across pools |
| `locationStrategy` | no | Location steering for non-proxied requests |
| `randomSteering` | no | Pool weights for random/least-* policies |

## Steering policies

- **`off`** (default): static failover over `defaultPools` (priority order).
- **`geo`**: route by `regionPools` / `countryPools` / `popPools`.
- **`random`**: weighted random selection (`randomSteering`).
- **`dynamic_latency`**: closest pool by round-trip time.
- **`proximity`**: closest pool by pool latitude/longitude.
- **`least_outstanding_requests`** / **`least_connections`**: load-aware selection.

## Outputs

| Output | Description |
|---|---|
| `load_balancer_id` | The load balancer ID |
| `load_balancer_dns_record_name` | The load balancer hostname |
| `load_balancer_cname_target` | The hostname clients point their DNS at |

## Composition

```
Monitor (account) -> Pool (account) -> LoadBalancer (zone) -> CloudflareDnsZone
```

Define a monitor and pool once, then reference the pool from any number of load
balancers across zones. Origins can themselves reference compute outputs, so the
whole traffic path is expressible as a dependency graph.

## Related components

- `CloudflareLoadBalancerPool` — referenced by `defaultPools`/`fallbackPool`/geo maps.
- `CloudflareLoadBalancerMonitor` — referenced by a pool's `monitor`.
- `CloudflareDnsZone` — provides `zoneId`.
