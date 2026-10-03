# Cloudflare Health Check

Deploys a standalone Cloudflare health check: a scheduled probe against an origin address that records healthy/unhealthy status, with no load balancer required. It is the monitoring-only sibling of the load-balancer monitor — use Cloudflare Load Balancer Monitor when a pool consumes the result to drive failover, and this kind to watch an origin. Health checks are a paid zone feature (Pro plans and above include a small allotment), with the plan gate enforced by the API at create. The probe protocol is `HTTP`, `HTTPS`, or `TCP`, and the matching config block is validated against it.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Health Check** — one `cloudflare_healthcheck` on the zone, carrying the probe target, protocol, regions, thresholds, and exactly one of `http_config` or `tcp_config` depending on `type`. The unused block is never sent — both are computed upstream, and sending the wrong one reads back as drift.

Destroy is a real delete: the probe stops and its history is removed. Set `suspended: true` to pause probing without losing the ID.

## Before You Deploy

### Planton Setup

- **Cloudflare Provider Connection** — an active connection in the Connect module whose API token carries **Zone → Health Checks → Edit**. Prefer a scoped API token over the global API key and do not mix the two in one process — a token-only environment can fail with an opaque 403 rather than "wrong credential type".
- **Planton Runner** — required when using Runner-based credential delivery. Not needed for inline credential authentication.

### Cloudflare Account

- **A zone on a Pro plan or above** for `zoneId` — free zones are rejected at the API (402/403). Checks beyond the plan's allotment are a zone-plan/add-on decision billed on the zone.
- **An origin reachable from Cloudflare's probe regions** — the check is created even if the origin never answers; its status just sits unhealthy.

## Deploy

### Console

Open the deployment store, find **Cloudflare Health Check**, and click **Deploy**. The creation wizard walks you through environment and connection configuration, the target zone, the probe address and protocol, thresholds and regions, and the protocol-specific probe settings. Start from the **HTTP origin probe** preset in the [Presets](#presets) tab to pre-populate a working configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareHealthcheck
metadata:
  name: origin-http
  org: acme-corp
  env: prod
spec:
  zoneId:
    value: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  name: origin-http
  address: acme.com
  type: HTTP
  httpConfig:
    path: /health
    expectedCodes:
      - "200"
```

```shell
planton apply -f healthcheck.yaml
```

This probes `acme.com/health` over HTTP every 60 seconds (Cloudflare's default interval) and marks the origin unhealthy after one failed probe. An Infra Job tracks the provisioning in real time.

### InfraChart

When the zone is deployed in the same InfraPipeline, wire the reference with ValueFromRef:

```yaml
spec:
  zoneId:
    valueFrom:
      kind: CloudflareDnsZone
      name: acme-zone
      fieldPath: status.outputs.zone_id
  name: origin-tcp
  address: origin.acme.com
  type: TCP
  tcpConfig:
    port: 5432
```

The InfraPipeline resolves the dependency graph, deploys the zone first, then creates the health check on the resolved zone ID.

## Key Configuration

These are the most important decisions when configuring a health check. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**This is not a load-balancer monitor** — standalone health checks and load-balancer monitors do not share IDs, APIs, or import formats. If you are about to paste this component's ID into a load-balancer pool, you are on the wrong kind — that is Cloudflare Load Balancer Monitor.

**The config block must match the type** — `httpConfig` is only valid on HTTP/HTTPS, `tcpConfig` only on TCP, and the spec rejects the wrong pairing at validation instead of letting it surface as apply failure or refresh drift. The `type` wall itself (HTTP, HTTPS, TCP) is a deliberate tightening — Cloudflare's schema accepts any string and rejects bad values only at the API.

**HTTPS needs an explicit port** — Cloudflare's default probe port is 80 even when `type` is HTTPS. Set `httpConfig.port: 443` explicitly for HTTPS checks on the standard port. Use `allowInsecure` only for self-signed origins, and pair HTTPS probing with an origin that presents a resolvable chain otherwise.

**Detection speed versus origin load** — `interval` (default 60s), `consecutiveFails`, and `consecutiveSuccesses` (both default 1) trade detection latency against probe traffic and flappiness. A single failed probe marking the origin unhealthy is aggressive; raise `consecutiveFails` for origins with occasional slow responses. Plan limits govern the minimum interval.

**Host headers for virtual hosts** — `httpConfig.headers` maps header names to `{values: [...]}` lists; set a `Host` header when the origin serves name-based virtual hosts, or the probe hits the default site. The User-Agent header cannot be overridden.

**Suspend instead of delete** — `suspended: true` pauses probing while keeping the check's configuration, ID, and history. Destroy removes all three.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **CloudflareDnsZone** | `zoneId` | `status.outputs.zone_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `healthcheck_id` | The created health check's ID | Wiring health-check status alerts in a Cloudflare Notification Policy |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**HTTP origin probe** — a GET against a dedicated `/health` path expecting `200`, on an origin that is not behind a Cloudflare load balancer. Start from the **HTTP origin probe** preset.

**TCP port probe** — `type: TCP` with `tcpConfig.port` against a database or non-HTTP service; a successful TCP handshake counts as healthy. The right shape when there is no HTTP endpoint to ask.

**Alert-wired check** — pair the check with a Cloudflare Notification Policy on health-check status changes, so an unhealthy origin pages someone instead of just coloring a dashboard.

## Works With

- [**Cloudflare DNS Zone**](/infra-catalog/cloudflare-dns-zone) — the zone the check belongs to; its plan gates create
- [**Cloudflare Load Balancer Monitor**](/infra-catalog/cloudflare-load-balancer-monitor) — the other health-check family, consumed by load-balancer pools to drive failover
- [**Cloudflare Notification Policy**](/infra-catalog/cloudflare-notification-policy) — alerting on the check's healthy/unhealthy transitions
