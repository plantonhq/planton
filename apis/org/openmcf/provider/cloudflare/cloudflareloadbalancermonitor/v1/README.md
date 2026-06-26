# CloudflareLoadBalancerMonitor

Provision an account-scoped health monitor for Cloudflare Load Balancing. A
monitor periodically probes the origins inside a `CloudflareLoadBalancerPool` and
decides whether each origin (and the pool) is healthy, driving failover.

## Why a standalone monitor

In Cloudflare's v5 model, monitors are **account-scoped and reusable**: one monitor
can health-check the origins of many pools. Modeling it as its own resource (rather
than burying it inside a load balancer) lets you define a health check once and
reference it from every pool that should share it.

## Requirements

- **Load Balancing add-on**: Cloudflare Load Balancing is a paid account add-on and
  must be enabled on the account first. Until it is, the entire Load Balancing API
  (read and write) returns `403`.
- **API token**: requires **Account → Load Balancing: Monitors and Pools → Edit**
  (monitors are account-scoped).

## Quick start

```yaml
apiVersion: cloudflare.openmcf.org/v1
kind: CloudflareLoadBalancerMonitor
metadata:
  name: web-https
spec:
  accountId: 0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d
  type: https
  path: /healthz
  expectedCodes: "2xx"
  interval: 60
  timeout: 5
  retries: 2
```

## Configuration reference

| Field | Required | Description |
|---|---|---|
| `accountId` | yes | 32-char Cloudflare account ID |
| `type` | no | `http` (default), `https`, `tcp`, `udp_icmp`, `icmp_ping`, `smtp` |
| `description` | no | Human-readable description |
| `path` | no | Endpoint path to check (http/https only) |
| `expectedCodes` | no | Expected response code/range, e.g. `2xx` (http/https only) |
| `expectedBody` | no | Substring that must appear in the body (http/https only) |
| `method` | no | HTTP method (default `GET`; tcp uses `connection_established`) |
| `headers` | no | Request headers (name -> values); set a `Host` header (http/https only) |
| `port` | conditional | Required for `tcp`, `udp_icmp`, `smtp`; optional for http/https |
| `interval` | no | Seconds between checks (default 60) |
| `timeout` | no | Seconds before a probe fails (default 5) |
| `retries` | no | Immediate retries before marking unhealthy (default 2) |
| `consecutiveUp` / `consecutiveDown` | no | Consecutive checks to flip health state |
| `followRedirects` | no | Follow origin redirects (http/https only) |
| `allowInsecure` | no | Skip TLS verification (https only) |
| `probeZone` | no | Emulate this zone while probing (http/https only) |

## Outputs

| Output | Description |
|---|---|
| `monitor_id` | The monitor ID (referenced by a pool's `monitor`) |
| `monitor_type` | The health-check protocol |

## Related components

- `CloudflareLoadBalancerPool` — references this monitor via `monitor`.
- `CloudflareLoadBalancer` — selects pools health-checked by this monitor.
