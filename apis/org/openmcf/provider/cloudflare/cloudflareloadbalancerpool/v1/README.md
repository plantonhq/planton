# CloudflareLoadBalancerPool

Provision an account-scoped pool of origin servers for Cloudflare Load Balancing.
A pool groups origins, health-checks them via a referenced
`CloudflareLoadBalancerMonitor`, and is selected by one or more zone-scoped
`CloudflareLoadBalancer`s (as default, fallback, or geo-routed pools).

## Why a standalone pool

Cloudflare v5 makes pools **account-scoped and reusable**: the same pool can back
many load balancers across different zones, with an independent lifecycle. Editing
a pool's origins should not churn the load balancers that reference it — so a pool
is its own resource, referenced by ID or `valueFrom`.

## Requirements

- **Load Balancing add-on**: Cloudflare Load Balancing is a paid account add-on and
  must be enabled on the account first, or the entire Load Balancing API returns `403`.
- **API token**: requires **Account → Load Balancing: Monitors and Pools → Edit**
  (pools are account-scoped).
- **Origins must be globally routable** when a `monitor` is attached — Cloudflare
  rejects reserved / non-routable addresses (e.g. RFC 5737 documentation ranges like
  `203.0.113.x`) with health monitoring enabled.
- **`checkRegions` is capped by plan tier** — exceeding the allowed number of probe
  regions fails validation; leave it empty to health-check from every region.

## Quick start

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareLoadBalancerPool
metadata:
  name: web-pool
spec:
  accountId: 0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d
  name: web-pool
  origins:
    - name: origin-1
      address:
        value: 203.0.113.10
  monitor:
    valueFrom:
      kind: CloudflareLoadBalancerMonitor
      name: web-https
      fieldPath: status.outputs.monitor_id
```

## Configuration reference

| Field | Required | Description |
|---|---|---|
| `accountId` | yes | 32-char Cloudflare account ID |
| `name` | yes | Short pool name (alphanumeric, `-`, `_`) |
| `origins[]` | yes | Origin servers (at least one) |
| `origins[].name` | yes | Origin name |
| `origins[].address` | yes | IP/hostname, or a reference to a compute output |
| `origins[].weight` | no | Relative weight 0.0–1.0 (default 1) |
| `origins[].enabled` | no | Enable the origin (default true) |
| `origins[].port` | no | Upstream port (0 = protocol default) |
| `origins[].hostHeader` | no | Host header override |
| `origins[].virtualNetworkId` | no | Virtual-network ID for internal addresses |
| `origins[].flattenCname` | no | Flatten CNAME origins to A/AAAA (default true) |
| `monitor` | no | Monitor ID/reference (origins always healthy if omitted) |
| `checkRegions[]` | no | Regions to check from (empty = everywhere) |
| `enabled` | no | Enable the pool (default true) |
| `minimumOrigins` | no | Healthy origins required to serve (default 1) |
| `latitude` / `longitude` | no | Coordinates for proximity steering |
| `loadShedding` | no | Shed a percentage of traffic (drain) |
| `originSteering` | no | Origin-selection policy |
| `notificationFilter` | no | Filter health notifications |

## Outputs

| Output | Description |
|---|---|
| `pool_id` | The pool ID (referenced by a load balancer's pool lists) |
| `pool_name` | The pool name |

## Composition

```
Monitor (account) -> Pool.monitor -> LoadBalancer.default_pools (zone)
```

Origins can reference any compute resource's output (e.g. an instance public IP),
so a pool wires backends into the resource graph.

## Related components

- `CloudflareLoadBalancerMonitor` — referenced by `monitor`.
- `CloudflareLoadBalancer` — selects this pool.
