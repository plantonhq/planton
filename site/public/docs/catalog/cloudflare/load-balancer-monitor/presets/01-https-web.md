---
title: "Preset: HTTPS web health check"
description: "An HTTPS monitor that probes `/healthz` on each origin and expects a 2xx response, suitable for a web/API pool behind a Cloudflare Load Balancer."
type: "preset"
rank: "01"
presetSlug: "01-https-web"
componentSlug: "load-balancer-monitor"
componentTitle: "Load Balancer Monitor"
provider: "cloudflare"
icon: "package"
order: 1
---

# Preset: HTTPS web health check

An HTTPS monitor that probes `/healthz` on each origin and expects a 2xx response,
suitable for a web/API pool behind a Cloudflare Load Balancer.

## When to use

- Origins serve HTTP(S) and expose a health endpoint.
- You want application-level health (status code / body), not just reachability.

## Key choices

- `path` / `expectedCodes`: tune to your health endpoint (e.g. `/healthz` -> `2xx`).
- `headers`: set a `Host` header so virtual-hosted origins route the probe correctly.
- `interval` / `timeout` / `retries`: leave at 60 / 5 / 2 unless you need faster
  failover (shorter intervals increase origin load).

## Placeholders

| Placeholder | Description |
|---|---|
| `<cloudflare-account-id>` | 32-character Cloudflare account ID |
| `<origin-host>` | The Host header value origins expect |

## Referencing it from a pool

```yaml
monitor:
  valueFrom:
    kind: CloudflareLoadBalancerMonitor
    name: web-https
    fieldPath: status.outputs.monitor_id
```
