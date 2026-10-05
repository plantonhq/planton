# Cloudflare Health Check

## Overview

`CloudflareHealthcheck` is one standalone health check: Cloudflare probes an origin address on a schedule and records healthy/unhealthy status, with notifications available through Cloudflare's alerting. Standalone health checks need NO load balancer -- they are the monitoring-only sibling of the load-balancer monitor.

Health checks are a paid zone feature (Pro plans and above include a small allotment; Business and Enterprise more). Cloudflare enforces the plan gate at create -- the API, not this spec, is the wall.

## Key Features

- **Standalone probes** -- no load balancer required; use `CloudflareLoadBalancerMonitor` when a pool consumes the result
- **Three types** -- `HTTP`, `HTTPS`, `TCP` (a deliberate tightening; the provider accepts any string and rejects bad values only at the API)
- **Config matches type** -- `http_config` only for HTTP/HTTPS; `tcp_config` only for TCP
- **Paid zone feature** -- Pro+; the API enforces the allotment

## Use Cases

**Ideal for:**

- Watching an origin that is not behind a Cloudflare load balancer
- An HTTP probe on `/health` with expected `2xx`
- A TCP handshake check on a non-HTTP port

**Not ideal for:**

- A monitor a load-balancer pool consumes -- that is `CloudflareLoadBalancerMonitor`
- Steering traffic based on the result -- that is `CloudflareLoadBalancer`

## API Specification

### Required Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `zone_id` | StringValueOrRef | Yes | The zone the health check belongs to. Can reference a `CloudflareDnsZone` via `value_from` (defaults to `status.outputs.zone_id`). |
| `address` | string | Yes | The origin being probed: a hostname or IP address. |
| `name` | string | Yes | Short name shown in the dashboard and alerts. |

### Optional Fields

| Field | Type | Description |
|-------|------|-------------|
| `type` | string | `HTTP` (default), `HTTPS`, or `TCP`. |
| `check_regions` | list | Regions to probe from. Unset lets Cloudflare pick. `ALL_REGIONS` is Enterprise only. |
| `consecutive_fails` | int32 | Failed probes before unhealthy (Cloudflare default: 1). |
| `consecutive_successes` | int32 | Successful probes before healthy again (Cloudflare default: 1). |
| `interval` | int32 | Seconds between probes (Cloudflare default: 60). |
| `retries` | int32 | Immediate retries on timeout (Cloudflare default: 2). |
| `timeout` | int32 | Probe timeout in seconds (Cloudflare default: 5). |
| `suspended` | bool | Pause probing without deleting the check. |
| `http_config` | object | HTTP/HTTPS probe details. Only valid when `type` is HTTP or HTTPS. The provider's header map is `header`; this spec wraps values as `headers`. |
| `tcp_config` | object | TCP probe details. Only valid when `type` is TCP. |

### Outputs

| Field | Description |
|-------|-------------|
| `healthcheck_id` | The created health check's ID |
| `zone_id` | The zone the health check belongs to |

## Example Manifests

An HTTP probe against `example.com/health`:

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareHealthcheck
metadata:
  name: origin-http
spec:
  zone_id:
    value: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  name: origin-http
  address: example.com
  type: HTTP
  http_config:
    path: /health
    expected_codes:
      - "200"
```

## Destroy Semantics

Destroy is a real delete. The probe stops and its history is removed. Suspending (`suspended: true`) is the reversible alternative when you want the ID and history to stay.

## Related Resources

- **CloudflareLoadBalancerMonitor** -- the account-scoped monitor a load-balancer pool consumes; do not mix the two
- **CloudflareDnsZone** -- `zone_id` foreign key; the zone plan is what gates create

## Further Reading

For operational judgment -- the Pro+ wall, type/config pairing, and the API-token auth defect class -- see GUIDE.md.

## References

- [Cloudflare Health Checks](https://developers.cloudflare.com/health-checks/)

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
