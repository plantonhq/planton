# CloudflareLoadBalancerMonitor

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `cloudflare.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

CloudflareLoadBalancerMonitorSpec defines an account-scoped health monitor for
Cloudflare Load Balancing. A monitor probes the origins inside a
CloudflareLoadBalancerPool and decides whether each origin (and the pool) is
healthy. Monitors are reusable: many pools can reference the same monitor.

## Example

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareLoadBalancerMonitor
metadata:
  name: test-monitor
spec:
  accountId: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
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

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.accountId` | `string` | yes |  |  |
| `spec.type` | `enum` |  | `http` |  |
| `spec.description` | `string` |  |  |  |
| `spec.path` | `string` |  |  |  |
| `spec.expectedCodes` | `string` |  | `2xx` |  |
| `spec.expectedBody` | `string` |  |  |  |
| `spec.method` | `string` |  |  |  |
| `spec.headers` | `[]CloudflareLoadBalancerMonitorHeader` |  |  |  |
| `spec.headers[].name` | `string` | yes |  |  |
| `spec.headers[].values` | `[]string` | yes |  |  |
| `spec.port` | `int32` |  |  |  |
| `spec.interval` | `int32` |  | `60` |  |
| `spec.timeout` | `int32` |  | `5` |  |
| `spec.retries` | `int32` |  | `2` |  |
| `spec.consecutiveUp` | `int32` |  |  |  |
| `spec.consecutiveDown` | `int32` |  |  |  |
| `spec.followRedirects` | `bool` |  |  |  |
| `spec.allowInsecure` | `bool` |  |  |  |
| `spec.probeZone` | `string` |  |  |  |

## Field Details

### spec.accountId

`string` · required

The Cloudflare account ID that owns this monitor. Must match the account of
every pool that references it.

- rule: {"required":true,"string":{"len":"32","pattern":"^[0-9a-fA-F]{32}$"}}

### spec.type

`enum`

The health-check protocol. Defaults to http when unspecified.

- default: `http`
- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `monitor_type_unspecified` -- Unspecified — the module treats this as "http" (the Cloudflare default).
- `http` -- HTTP health check (application-layer; supports path/codes/body/headers).
- `https` -- HTTPS health check (application-layer over TLS).
- `tcp` -- TCP connection check (transport-layer; requires a port).
- `udp_icmp` -- UDP check via ICMP (requires a port).
- `icmp_ping` -- ICMP echo (ping) reachability check.
- `smtp` -- SMTP health check (requires a port).

### spec.description

`string`

Human-readable description shown in the Cloudflare dashboard.

### spec.path

`string`

The endpoint path to health-check against (e.g. "/healthz"). Only valid for
http/https monitors.

### spec.expectedCodes

`string`

The expected HTTP response code or range that marks an origin healthy
(e.g. "200", "2xx", "200-299"). Only valid for http/https monitors — and
for those types Cloudflare REQUIRES this or expected_body (the API answers
400 code 1002 when both are empty; measured live).

- default: `2xx`

### spec.expectedBody

`string`

A case-insensitive substring that must appear in the response body for the
origin to be considered healthy. Only valid for http/https monitors; either
this or expected_codes must be set for those types (Cloudflare 400/1002).

### spec.method

`string`

The HTTP method for the health check. Leave empty to use the protocol
default ("GET" for http/https, "connection_established" for tcp). Only
meaningful for http/https monitors.

### spec.headers

`[]CloudflareLoadBalancerMonitorHeader`

HTTP request headers to send with the health check, keyed by header name
(each may carry multiple values). Setting a "Host" header is recommended.
Only valid for http/https monitors.

### spec.headers[].name

`string` · required

The header name (e.g. "Host").

- rule: {"required":true}

### spec.headers[].values

`[]string` · required

The header value(s). Cloudflare currently supports a single Host override
per monitor, but the API models each header as a list of values.

- rule: {"repeated":{"minItems":"1"}}

### spec.port

`int32`

The port to connect to. Required for tcp, udp_icmp, and smtp monitors; for
http/https set it only when using a non-standard port (defaults: 80/443).

- rule: port must be 0 (protocol default) or between 1 and 65535

### spec.interval

`int32`

Seconds between health checks. Leave 0 for the Cloudflare default (60).
Shorter intervals improve failover time but increase origin load.

- default: `60`
- rule: interval must be 0 (default) or a positive number of seconds

### spec.timeout

`int32`

Seconds to wait before a single probe is considered failed. Leave 0 for the
Cloudflare default (5).

- default: `5`
- rule: timeout must be 0 (default) or a positive number of seconds

### spec.retries

`int32`

Number of immediate retries after a failed probe before marking the origin
unhealthy. Leave 0 for the Cloudflare default (2).

- default: `2`
- rule: retries must be 0 (default) or a positive number

### spec.consecutiveUp

`int32`

Consecutive passing checks required to mark a previously-unhealthy origin
healthy. Leave 0 to use Cloudflare's behavior.

- rule: consecutive_up must be 0 or a positive number

### spec.consecutiveDown

`int32`

Consecutive failing checks required to mark a healthy origin unhealthy.
Leave 0 to use Cloudflare's behavior.

- rule: consecutive_down must be 0 or a positive number

### spec.followRedirects

`bool`

Follow redirects returned by the origin. Only valid for http/https monitors.

### spec.allowInsecure

`bool`

Skip TLS certificate validation when probing over HTTPS. Only valid for
https monitors; use with care.

### spec.probeZone

`string`

Probe as if from this zone (emulates the zone while probing). Only valid for
http/https monitors.

## Validation Rules

- `monitor.port_required_for_l4`: a port is required for tcp, udp_icmp, and smtp monitors
- `monitor.http_expectation_required`: expected_codes or expected_body is required for http and https monitors

## Outputs

Reference an output from another manifest as `valueFrom: {kind: CloudflareLoadBalancerMonitor, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.monitor_id` | `string` | The Cloudflare-assigned identifier of the monitor. A CloudflareLoadBalancerPool references this value via its `monitor` field. |
| `status.outputs.monitor_type` | `string` | The health-check protocol of the monitor (echoed for convenience). |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| CloudflareLoadBalancerPool | `spec.monitor` | `status.outputs.monitor_id` |

## See Also

- [Overview](../README.md)
