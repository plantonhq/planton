# Cloudflare Load Balancer

Deploys a zone-scoped Cloudflare Load Balancer that attaches a DNS hostname to one
or more account-scoped pools and steers traffic across them with health-aware
failover, geo-routing, weighted distribution, and session affinity.

## What Gets Created

- A `cloudflare_load_balancer` bound to the hostname and zone, referencing the
  given `defaultPools` (and `fallbackPool`), with the configured steering,
  session affinity, geo-pool maps, adaptive routing, and location strategy.

Pools and monitors are **separate, reusable resources** — create them with
`CloudflareLoadBalancerPool` and `CloudflareLoadBalancerMonitor` and reference the
pools here by ID or `valueFrom`.

## Prerequisites

- Cloudflare Load Balancing add-on enabled on the account (paid feature) — otherwise
  the Load Balancing API returns `403`.
- An API token with **Zone → Load Balancers → Edit** (zone-scoped; distinct from the
  account-level "Load Balancers Account" permission), plus
  **Account → Load Balancing: Monitors and Pools → Edit** for the pools/monitors it
  references, and the target zone in the token's Zone Resources scope.
- An existing Cloudflare DNS zone (literal zone ID or a `CloudflareDnsZone` ref).
- One or more `CloudflareLoadBalancerPool` resources (each optionally health-checked
  by a `CloudflareLoadBalancerMonitor`).

## Quick Start

```yaml
apiVersion: cloudflare.planton.dev/v1
kind: CloudflareLoadBalancer
metadata:
  name: my-lb
spec:
  hostname: app.example.com
  zoneId:
    value: 0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d
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
  steeringPolicy: off
```

## Configuration Reference

### Required

| Field | Type | Description |
|---|---|---|
| `hostname` | `string` | DNS hostname the load balancer serves |
| `zoneId` | `StringValueOrRef` | Zone ID, or a `CloudflareDnsZone` reference |
| `defaultPools` | `list<StringValueOrRef>` | Ordered pools by failover priority |
| `fallbackPool` | `StringValueOrRef` | Pool of last resort |

### Optional

| Field | Default | Description |
|---|---|---|
| `proxied` | `false` | Route through Cloudflare (orange cloud) |
| `enabled` | `true` | Enable the load balancer |
| `steeringPolicy` | `off` | `off`, `geo`, `random`, `dynamic_latency`, `proximity`, `least_outstanding_requests`, `least_connections` |
| `sessionAffinity` | `none` | `none`, `cookie`, `ip_cookie`, `header` |
| `sessionAffinityTtl` | — | Affinity session expiry (seconds) |
| `sessionAffinityAttributes` | — | Drain, headers, cookie flags, zero-downtime failover |
| `ttl` | — | DNS TTL (gray-clouded only) |
| `regionPools` / `countryPools` / `popPools` | — | Geo code -> ordered pool list |
| `adaptiveRouting` | — | Zero-downtime failover across pools |
| `locationStrategy` | — | Location steering for non-proxied requests |
| `randomSteering` | — | Pool weights for random/least-* policies |

## Stack Outputs

| Output | Description |
|---|---|
| `load_balancer_id` | The load balancer ID |
| `load_balancer_dns_record_name` | The load balancer hostname |
| `load_balancer_cname_target` | The hostname clients point their DNS at |

## Related Components

- [CloudflareLoadBalancerPool](/docs/catalog/cloudflare/cloudflareloadbalancerpool) — the pools this load balancer references
- [CloudflareLoadBalancerMonitor](/docs/catalog/cloudflare/cloudflareloadbalancermonitor) — health checks the pools use
- [CloudflareDnsZone](/docs/catalog/cloudflare/cloudflarednszone) — provides `zoneId`
