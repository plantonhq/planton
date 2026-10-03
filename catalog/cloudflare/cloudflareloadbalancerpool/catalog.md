# Cloudflare Load Balancer Pool

Deploys a reusable Cloudflare Load Balancing origin pool. A pool groups origin servers, health-checks them via a referenced monitor, and is selected by one or more zone-scoped load balancers as default, fallback, or geo-routed pools. Pools are account-scoped and have an independent lifecycle, so one pool can serve several load balancers at once.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Load Balancer Pool** -- a pool containing the declared origins (with per-origin weight, port, host-header, and enabled state), linked to a health monitor, with optional check-region restriction, load shedding, origin steering, and notification filters

## Before You Deploy

### Planton Setup

- **Cloudflare Provider Connection** -- an active connection in the Connect module with a Cloudflare API token that has Load Balancing edit access. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline API token authentication.

### Cloudflare Account

- **Load Balancing add-on** -- Cloudflare Load Balancing is a paid feature and must be enabled on your account before deploying.
- **A health monitor** -- create a CloudflareLoadBalancerMonitor first (or reference an existing monitor ID) so the pool's origins are actively health-checked. Origins in a monitored pool must be globally routable addresses: Cloudflare rejects reserved/documentation ranges with a 400 (code 1002) the moment a monitor is attached.
- **Reachable origins** -- each origin address (IP or unproxied hostname) must be reachable for health checks and traffic. Internal addresses require a virtual-network ID.

## Deploy

### Console

Open the deployment store, find **Cloudflare Load Balancer Pool**, and click **Deploy**. The creation wizard walks you through account and name, the origins builder, the monitor reference and check regions, and optional advanced tuning. Start from the **Web pool with two origins** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareLoadBalancerPool
metadata:
  name: web-pool
  org: acme-corp
  env: prod
spec:
  accountId: "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4"
  name: web-pool
  origins:
    - name: origin-1
      address:
        value: 203.0.113.10
      weight: 1
    - name: origin-2
      address:
        value: 203.0.113.11
      weight: 1
  monitor:
    valueFrom:
      kind: CloudflareLoadBalancerMonitor
      name: web-https
      fieldPath: status.outputs.monitor_id
  minimumOrigins: 1
```

```shell
planton apply -f cloudflare-load-balancer-pool.yaml
```

This creates a two-origin pool health-checked by a referenced monitor. An Infra Job tracks the provisioning in real time.

### InfraChart

Deploy the monitor and pool together and wire the pool to the monitor with ValueFromRef:

```yaml
spec:
  monitor:
    valueFrom:
      kind: CloudflareLoadBalancerMonitor
      name: web-https
      fieldPath: status.outputs.monitor_id
```

The InfraPipeline resolves the dependency graph, deploys the monitor first, then provisions the pool with the resolved monitor ID. A load balancer then references this pool's `status.outputs.pool_id`.

## Key Configuration

These are the most important decisions when configuring a pool. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Origins** -- The servers behind the pool. Each origin's `address` can be a literal IP/unproxied hostname or a `valueFrom` reference to another resource's output, so the pool tracks backends as they are recreated. `weight` (0.0-1.0) drives weighted steering; `port` and `hostHeader` tune the connection.

**Monitor** -- The `monitor` foreign key references a CloudflareLoadBalancerMonitor. Without a monitor, every origin is treated as permanently healthy and traffic keeps flowing to failed origins -- always attach one in production.

**Minimum origins** -- The pool fails over when fewer than `minimumOrigins` origins are healthy (default 1). Raise it to fail a pool over before it is fully drained rather than overloading a lone survivor.

**Check regions** -- Restrict which Cloudflare regions run the health checks. Leave empty to check from every data center, or pin to regions near your origins to reflect real client geography and reduce probe traffic.

**Advanced tuning** -- Optional: proximity coordinates (for geo/proximity steering), load shedding (drain a pool gracefully), origin steering policy (random, hash, least-connections, least-outstanding-requests), and health-notification filters.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **CloudflareLoadBalancerMonitor** | `monitor` | `status.outputs.monitor_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `pool_id` | The Cloudflare-assigned identifier of the pool | Referenced by a CloudflareLoadBalancer's `default_pools`, `fallback_pool`, or geo-pool maps |

`status.outputs` also echoes `pool_name` back, but it is the value you set in the spec -- downstream resources reference pools by `pool_id`.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Web pool with two origins** -- Two interchangeable web origins health-checked by an HTTPS monitor, ready to attach to a load balancer's default pools. Use for standard balanced backends. Start from the **Web pool with two origins** preset.

**Geo-located pool** -- A regional pool tagged with latitude/longitude and a least-connections origin policy, for proximity or geo steering across regional backends. Start from the **Geo-located pool (proximity steering)** preset.

## Works With

- [**Cloudflare Load Balancer Monitor**](/infra-catalog/cloudflare-load-balancer-monitor) -- health-checks this pool's origins
- [**Cloudflare Load Balancer**](/infra-catalog/cloudflare-load-balancer) -- references this pool as a default, fallback, or geo-routed pool
