# CloudflareLoadBalancer — Technical Notes

Cloudflare Load Balancer is a Global Server Load Balancing (GSLB) control plane in
Cloudflare's global network: it steers DNS responses and HTTP traffic across origin
pools based on real-time health, geography, and steering policy. You define
policies; Cloudflare runs them at the edge.

## The three-resource model

Cloudflare splits load balancing into three resources with different scopes, and
Planton mirrors that split exactly:

| Resource | Scope | Planton kind |
|---|---|---|
| Monitor (health check) | account | `CloudflareLoadBalancerMonitor` |
| Pool (origin group) | account | `CloudflareLoadBalancerPool` |
| Load Balancer (hostname + steering) | zone | `CloudflareLoadBalancer` |

Monitors and pools are **account-scoped and reusable** — one pool can back many
load balancers across zones, and one monitor can health-check many pools. Modeling
each as its own first-class kind (rather than bundling them) means a pool edit does
not churn the load balancers that reference it, and the same pool composes into
many traffic paths. This load balancer references pools by ID or `valueFrom`.

```
Monitor (account) -> Pool (account) -> LoadBalancer (zone) -> CloudflareDnsZone
        compute output --origin address valueFrom--> Pool
```

## Steering policies

`steeringPolicy` selects how a pool is chosen:

- **`off`** (default) — static failover over `defaultPools` (priority order).
- **`geo`** — route by `regionPools` / `countryPools` / `popPools`.
- **`random`** — weighted random selection (`randomSteering`).
- **`dynamic_latency`** — closest pool by measured round-trip time.
- **`proximity`** — closest pool by pool latitude/longitude.
- **`least_outstanding_requests` / `least_connections`** — load-aware selection.

### Active-passive failover

The canonical resilience pattern is counter-intuitive: set `steeringPolicy: off`
and order the pools in `defaultPools` by priority. Traffic flows to the first
healthy pool; `fallbackPool` is the pool of last resort served even if unhealthy.
When a higher-priority pool recovers, traffic fails back automatically.

## Health monitoring

A pool's `monitor` drives health. Cloudflare probes each origin from three data
centers per region (a quorum), and a pool is healthy while its healthy-origin count
stays at or above `minimumOrigins`. The minimum probe interval is plan-gated
(Pro 60s, Business 15s, Enterprise 10s), which sets the floor on failover time.

Common monitor pitfalls: origin firewalls blocking Cloudflare probe IPs; an
`expectedCodes` mismatch when the health endpoint returns a redirect (set
`followRedirects`); and HTTPS probes against self-signed origins (use a Cloudflare
Origin CA cert or `allowInsecure`).

## Session affinity

`sessionAffinity` (`cookie`, `ip_cookie`, `header`) pins a client to an origin —
required for stateful apps. Affinity is only meaningful on proxied (orange-cloud)
load balancers. `sessionAffinityAttributes` tunes drain duration, the header set,
cookie `samesite`/`secure` flags, and zero-downtime failover behavior.

## Geo-pool maps

`regionPools`, `countryPools`, and `popPools` map a geo code to an ordered list of
pools. They are modeled as a list of `{ code, poolIds[] }` entries (rather than a
raw map) because each entry's `poolIds` is an ordered list of pool references; the
IaC rebuilds the provider's `{ code => pool_ids }` map. Lookups fall back
PoP -> country -> region -> `defaultPools`.

## Pricing (context)

Load Balancing is a paid add-on: a base fee includes two origins, with per-origin
fees beyond that and a flat geo-routing add-on. Proxied load balancers incur no
significant per-query fee; gray-clouded (DNS-only) ones do.

## Parity

The Pulumi (`iac/pulumi/module`) and Terraform (`iac/tf`) modules are behaviorally
identical: same pool references, same enum/default handling (none/off omitted so the
provider applies its default), same geo-pool map construction, same session-affinity
and steering nested blocks, and the same three stack outputs — with
`load_balancer_cname_target` resolving to the hostname clients point DNS at.

## Outputs

- `load_balancer_id` — the load balancer ID.
- `load_balancer_dns_record_name` — the hostname.
- `load_balancer_cname_target` — the hostname clients CNAME to.
