# Cloudflare Load Balancer

Deploys a zone-scoped Cloudflare Load Balancer that attaches a DNS hostname to account-scoped origin pools and steers traffic across them with health-aware failover, geo-routing, weighted distribution, session affinity, and per-request traffic rules. Its zone and pool references wire to CloudflareDnsZone and CloudflareLoadBalancerPool resources via ValueFromRef, so the whole traffic path resolves as one dependency graph in an InfraPipeline.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Load Balancer** -- the zone-scoped load balancer bound to the specified hostname, referencing your pools (default, fallback, and optional geo maps), with the configured proxy, steering policy, session affinity, adaptive routing, and traffic rules

Pools and health monitors are separate, reusable Infra Components (`CloudflareLoadBalancerPool`, `CloudflareLoadBalancerMonitor`) with account scope and independent lifecycles -- one pool can back many load balancers. Reference them by ID or ValueFromRef.

## Before You Deploy

### Planton Setup

- **Cloudflare Provider Connection** -- an active connection in the Connect module with a Cloudflare API token that has **Zone -> Load Balancers -> Edit** (zone-scoped) plus **Account -> Load Balancing: Monitors and Pools -> Edit** access. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline API token authentication.

### Cloudflare Account

- **Load Balancing add-on** -- Cloudflare Load Balancing is a paid account add-on and must be enabled first, or the entire Load Balancing API returns `403`.
- **An existing Cloudflare DNS zone** -- the hostname must belong to a zone you control. Provide the zone ID directly or reference a CloudflareDnsZone resource via ValueFromRef.
- **At least one pool** -- deploy a `CloudflareLoadBalancerPool` (and usually a `CloudflareLoadBalancerMonitor` to health-check it) before or alongside the load balancer.

## Deploy

### Console

Open the deployment store, find **Cloudflare Load Balancer**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Active-Passive Failover** preset in the [Presets](#presets) tab to pre-populate a working configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareLoadBalancer
metadata:
  name: api-lb
  org: acme-corp
  env: prod
spec:
  hostname: api.example.com
  zoneId:
    value: "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4"
  defaultPools:
    - value: "17b5962d775c646f3f9725cbc7a53df4"
  fallbackPool:
    value: "17b5962d775c646f3f9725cbc7a53df4"
  proxied: true
```

```shell
planton apply -f cloudflare-load-balancer.yaml
```

This creates a proxied load balancer for `api.example.com` over one pool with static failover. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying alongside the zone and pools, use ValueFromRef to wire the whole traffic path in one InfraPipeline:

```yaml
spec:
  zoneId:
    valueFrom:
      kind: CloudflareDnsZone
      name: prod-zone
      fieldPath: status.outputs.zone_id
  defaultPools:
    - valueFrom:
        kind: CloudflareLoadBalancerPool
        name: web-pool
        fieldPath: status.outputs.pool_id
```

The InfraPipeline resolves the dependency graph, deploys the zone and pool first, then provisions the load balancer with the resolved IDs.

## Key Configuration

These are the most important decisions when configuring a load balancer. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Steering policy** -- Controls how traffic is distributed across pools. `off` (default) uses static failover in `defaultPools` order. `geo` routes by `regionPools`/`countryPools`/`popPools`. `random` distributes by `randomSteering` weights for A/B testing or canary deployments. `dynamic_latency`, `proximity`, and the `least_*` policies select pools by measured latency, location, or load.

**Session affinity** -- Set `sessionAffinity` to `cookie` to pin returning clients to the same origin using a Cloudflare-managed cookie (`ip_cookie` and `header` variants available). Tune drain, cookie flags, and zero-downtime failover via `sessionAffinityAttributes`. Leave as `none` (default) for stateless workloads.

**Traffic rules** -- `rules[]` matches requests with condition expressions and, per rule, either overrides any subset of the steering surface (pools, policy, affinity, TTLs, geo maps) or answers directly at the edge with a fixed response (e.g. a 503 maintenance page). Rules run in `priority` order; when no rule sets a priority, list order decides.

**Fallback pool** -- `fallbackPool` is the pool of last resort when every other pool is unhealthy. Cloudflare serves it regardless of its own health state.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **CloudflareDnsZone** | `zoneId` | `status.outputs.zone_id` |
| **CloudflareLoadBalancerPool** | `defaultPools[]`, `fallbackPool`, geo pool maps, rule override pools | `status.outputs.pool_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `load_balancer_id` | The unique identifier of the created Cloudflare Load Balancer | Monitoring dashboards, management automation |
| `load_balancer_dns_record_name` | The hostname DNS record associated with the load balancer | DNS verification, application endpoint configuration |
| `load_balancer_cname_target` | The canonical CNAME target that the hostname resolves to (Cloudflare endpoint) | External DNS configuration, CNAME record setup |
| `zone_id` | The Cloudflare zone that owns the load balancer | Import tooling, zone-scoped automation |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Active-passive failover** -- Two pools with static steering. All traffic goes to the first healthy pool; if it fails health checks, traffic fails over to the second. Use for primary/backup architectures and DR failover. Start from the **Active-Passive Failover** preset.

**Geographic routing** -- Multiple pools with `geo` steering. Cloudflare routes clients to the configured pool for their region or country for lowest latency. Use for multi-region deployments serving US, EU, and APAC traffic. Start from the **Geographic Routing** preset.

**Weighted A/B testing** -- Two or more pools with `random` steering and different weights. Traffic is distributed proportionally (e.g., 70/30 split) for A/B tests, canary deployments, or gradual rollouts. Start from the **Weighted A/B Testing** preset.

**Traffic rules** -- Per-path or per-header routing decisions: send `/api` traffic to a dedicated pool, disable session affinity for stateless endpoints, or serve a fixed maintenance response for a path -- all on one load balancer. Start from the **Traffic Rules** preset.

## Works With

- [**Cloudflare DNS Zone**](/infra-catalog/cloudflare-dns-zone) -- provides the zone ID that determines which domain the load balancer hostname belongs to
- [**Cloudflare Load Balancer Pool**](/infra-catalog/cloudflare-load-balancer-pool) -- the account-scoped origin pools the load balancer steers across
- [**Cloudflare Load Balancer Monitor**](/infra-catalog/cloudflare-load-balancer-monitor) -- health-checks the pools' origins
