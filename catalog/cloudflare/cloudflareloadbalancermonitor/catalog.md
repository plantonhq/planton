# Cloudflare Load Balancer Monitor

Deploys a reusable Cloudflare Load Balancing health monitor. A monitor probes the origins inside a load balancer pool and decides whether each origin (and the pool) is healthy. Monitors are account-scoped and reusable -- many pools can reference the same monitor -- and a monitor has no knowledge of the pools that use it, so its lifecycle is deliberately independent of theirs.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Load Balancer Monitor** -- a health check of the configured protocol (HTTP, HTTPS, TCP, UDP/ICMP, ICMP ping, or SMTP) with the chosen probe configuration, timing, and health-flip thresholds

## Before You Deploy

### Planton Setup

- **Cloudflare Provider Connection** -- an active connection in the Connect module with a Cloudflare API token that has Load Balancing edit access. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline API token authentication.

### Cloudflare Account

- **Load Balancing add-on** -- monitors, pools, and load balancers all ride the paid Load Balancing add-on; without it every call returns `403`. Enable the subscription before deploying any of the three.
- **Reachable origins** -- the origins this monitor will eventually check (via a pool) must be reachable from Cloudflare's probing regions over the chosen protocol.

## Deploy

### Console

Open the deployment store, find **Cloudflare Load Balancer Monitor**, and click **Deploy**. The creation wizard walks you through account and protocol selection, the protocol-specific probe configuration, and the timing thresholds. Start from the **HTTPS web health check** preset in the [Presets](#presets) tab for an application-layer check.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareLoadBalancerMonitor
metadata:
  name: web-https
  org: acme-corp
  env: prod
spec:
  accountId: "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4"
  type: https
  path: /healthz
  expectedCodes: "2xx"
  method: GET
  interval: 60
  timeout: 5
  retries: 2
  headers:
    - name: Host
      values:
        - app.example.com
```

```shell
planton apply -f cloudflare-load-balancer-monitor.yaml
```

This creates an HTTPS monitor that probes `/healthz` on each origin and marks it healthy on a `2xx` response. An Infra Job tracks the provisioning in real time.

## Key Configuration

These are the most important decisions when configuring a monitor. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Protocol (`type`)** -- determines what a check measures and which fields matter. `http`/`https` validate the application response (path, status codes, body, headers) and catch "up but broken" origins. `tcp`, `udp_icmp`, and `smtp` confirm a port accepts connections and require `port > 0` -- a tcp monitor with no port fails validation, while the HTTP knobs are ignored. `icmp_ping` checks raw reachability only. Prefer an application-layer check whenever clients use HTTP(S).

**Expected codes (`expectedCodes`)** -- for HTTP(S), the response code or range (`200`, `2xx`, `200-299`) that marks an origin healthy. The endpoint must signal health through its HTTP status, not just the body. Cloudflare requires `expectedCodes` or `expectedBody` on every http/https monitor -- creation fails with a 400 (code 1002) when both are empty.

**Timing and thresholds** -- `interval`, `timeout`, `retries`, `consecutiveUp`, and `consecutiveDown` trade failover speed for stability; effective detection time is roughly interval × consecutive-down. Zero means "Cloudflare's default" (60s / 5s / 2): the module omits zeroed fields so the server default applies. After an import, the API reports those defaults as numbers, so a post-import plan may show a diff on `port` or the `consecutive*` fields even though nothing operationally changed.

**Deletion order matters** -- pools reference monitors, not the other way around. Deleting a monitor a pool still names will fail or leave the pool unmonitored, at which point its origins look permanently healthy. Delete pools first, or clear the pool's `monitor` field, then the monitor.

## Outputs and Dependencies

### What This Kind Consumes

This kind has no foreign-key dependencies -- a monitor is defined entirely by its own probe configuration.

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `monitor_id` | The Cloudflare-assigned identifier of the monitor | Referenced by a CloudflareLoadBalancerPool's `monitor` field |

## Common Patterns

**HTTPS web health check** -- an HTTPS monitor probing `/healthz` for a `2xx` response, with a `Host` header for virtual-hosted origins. Use for web/API pools where application-level health matters. Start from the **HTTPS web health check** preset.

**TCP port check** -- a TCP monitor that confirms a port accepts connections, for non-HTTP origins such as databases or message brokers. Use when only connection-level health is needed. Start from the **TCP port health check** preset.

**One monitor, many pools** -- because monitors are account-scoped and reusable, define one health check per service contract (e.g. "web tier answers `/healthz` with 2xx") and point every regional pool at it, so a probe change rolls out everywhere at once.

## Works With

- [**Cloudflare Load Balancer Pool**](/infra-catalog/cloudflare-load-balancer-pool) -- references this monitor (via `monitor`) to health-check its origins
- [**Cloudflare Load Balancer**](/infra-catalog/cloudflare-load-balancer) -- steers traffic across the pools this monitor keeps honest
