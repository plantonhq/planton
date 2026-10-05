# CloudflareHealthcheck

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `cloudflare.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

CloudflareHealthcheckSpec defines one standalone health check: Cloudflare
probes an origin address on a schedule and records healthy/unhealthy status,
with notifications available through Cloudflare's alerting. Standalone health
checks need NO load balancer -- they are the monitoring-only sibling of the
load-balancer monitor (use CloudflareLoadBalancerMonitor when a load balancer
consumes the result; use this kind to watch an origin).

Health checks are a paid zone feature (Pro plans and above include a small
allotment; Business and Enterprise more). Cloudflare enforces the plan gate at
create -- the API, not this spec, is the wall.

## Example

```yaml
# A complete, protovalidate-valid CloudflareHealthcheck example: an HTTPS
# probe with response assertions and a Host header. Health checks are a paid
# zone feature (Pro and above); the plan gate is Cloudflare's, at create.
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareHealthcheck
metadata:
  name: origin-https-probe
spec:
  zone_id:
    value: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  name: origin-https
  address: origin.example.com
  type: HTTPS
  check_regions:
    - WEU
    - ENAM
  consecutive_fails: 2
  consecutive_successes: 2
  interval: 60
  retries: 2
  timeout: 5
  http_config:
    method: GET
    path: /healthz
    port: 443
    expected_codes:
      - "200"
    expected_body: ok
    follow_redirects: false
    allow_insecure: false
    headers:
      Host:
        values:
          - origin.example.com
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.zoneId` | `string \| valueFrom` | yes |  | CloudflareDnsZone (`status.outputs.zone_id`) |
| `spec.name` | `string` | yes |  |  |
| `spec.address` | `string` | yes |  |  |
| `spec.type` | `string` |  | `HTTP` |  |
| `spec.checkRegions` | `[]string` |  |  |  |
| `spec.consecutiveFails` | `int32` |  |  |  |
| `spec.consecutiveSuccesses` | `int32` |  |  |  |
| `spec.interval` | `int32` |  |  |  |
| `spec.retries` | `int32` |  |  |  |
| `spec.timeout` | `int32` |  |  |  |
| `spec.suspended` | `bool` |  |  |  |
| `spec.httpConfig` | `CloudflareHealthcheckHttpConfig` |  |  |  |
| `spec.httpConfig.method` | `string` |  | `GET` |  |
| `spec.httpConfig.path` | `string` |  |  |  |
| `spec.httpConfig.port` | `int32` |  |  |  |
| `spec.httpConfig.expectedCodes` | `[]string` |  |  |  |
| `spec.httpConfig.expectedBody` | `string` |  |  |  |
| `spec.httpConfig.followRedirects` | `bool` |  |  |  |
| `spec.httpConfig.allowInsecure` | `bool` |  |  |  |
| `spec.httpConfig.headers` | `map<string, CloudflareHealthcheckHeaderValues>` |  |  |  |
| `spec.httpConfig.headers.*.values` | `[]string` | yes |  |  |
| `spec.tcpConfig` | `CloudflareHealthcheckTcpConfig` |  |  |  |
| `spec.tcpConfig.method` | `string` |  | `connection_established` |  |
| `spec.tcpConfig.port` | `int32` |  |  |  |

## Field Details

### spec.zoneId

`string | valueFrom` · required

The zone the health check belongs to.
When using value_from, defaults to CloudflareDnsZone kind and status.outputs.zone_id field path.

- references: CloudflareDnsZone (`status.outputs.zone_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: CloudflareDnsZone, name: <that resource's name>, fieldPath: status.outputs.zone_id}} -- a bare string does not parse

### spec.name

`string` · required

A short name for the health check (shown in the dashboard and alerts).

- rule: {"string":{"minLen":"1"}}

### spec.address

`string` · required

The origin being probed: a hostname or IP address. Cloudflare validates
this is a real, publicly routable origin at create: documentation-range
addresses (TEST-NET, e.g. 203.0.113.x) are rejected with 400 code 1002
"origin address is invalid" (measured live 2026-08-27). The origin does
not need to answer the probe -- it needs to be routable.

- rule: {"required":true}

### spec.type

`string` · optional (explicit presence)

The probe protocol. Cloudflare's own schema accepts any string here and
rejects bad values only at the API -- this wall (HTTP, HTTPS, TCP) is a
deliberate tightening to the documented protocol set.

- default: `HTTP`
- rule: {"string":{"in":["HTTP","HTTPS","TCP"]}}

### spec.checkRegions

`[]string`

Regions to probe from. Unset lets Cloudflare pick a default region --
BUT declare it explicitly on Terraform-managed checks: Cloudflare echoes
its chosen default (measured live 2026-08-27: ["WNAM"] on a Pro zone)
into an attribute the provider models optional-not-computed, so an
omitted list re-plans a no-op update forever (the declare-it-or-drift
class; upstream modeling defect at v5.23.0).
ALL_REGIONS (Enterprise only) probes from everywhere.

- rule: {"repeated":{"items":{"string":{"in":["WNAM","ENAM","WEU","EEU","NSAM","SSAM","OC","ME","NAF","SAF","IN","SEAS","NEAS","ALL_REGIONS"]}}}}

### spec.consecutiveFails

`int32` · optional (explicit presence)

Consecutive failed probes before the origin is marked unhealthy
(Cloudflare's default: 1).

- rule: {"int32":{"gt":0}}

### spec.consecutiveSuccesses

`int32` · optional (explicit presence)

Consecutive successful probes before the origin is marked healthy again
(Cloudflare's default: 1).

- rule: {"int32":{"gt":0}}

### spec.interval

`int32` · optional (explicit presence)

Seconds between probes (Cloudflare's default: 60). Shorter intervals give
faster detection at the cost of more origin traffic; plan limits govern the
minimum.

- rule: {"int32":{"gt":0}}

### spec.retries

`int32` · optional (explicit presence)

Retries attempted (immediately, not on the interval) when a probe times out
before it counts as failed (Cloudflare's default: 2).

- rule: {"int32":{"gt":0}}

### spec.timeout

`int32` · optional (explicit presence)

Probe timeout in seconds (Cloudflare's default: 5).

- rule: {"int32":{"gt":0}}

### spec.suspended

`bool` · optional (explicit presence)

Pause probing without deleting the check's configuration or history.

### spec.httpConfig

`CloudflareHealthcheckHttpConfig`

HTTP/HTTPS probe details. Only valid when type is HTTP or HTTPS.

### spec.httpConfig.method

`string` · optional (explicit presence)

The HTTP method (Cloudflare's default: GET).

- default: `GET`
- rule: {"string":{"in":["GET","HEAD"]}}

### spec.httpConfig.path

`string` · optional (explicit presence)

The path probed (Cloudflare's default: /).

### spec.httpConfig.port

`int32` · optional (explicit presence)

The port probed (Cloudflare's default: 80 -- set 443 explicitly for HTTPS
checks on the standard port).

- rule: {"int32":{"lte":65535,"gte":1}}

### spec.httpConfig.expectedCodes

`[]string`

Response codes counted as healthy. Accepts exact codes ("200") and
wildcard families ("2xx"). Unset means Cloudflare's default of 200.

### spec.httpConfig.expectedBody

`string`

A substring the response body must contain to count as healthy. Cloudflare
reads only the first 10 KiB of the body.

### spec.httpConfig.followRedirects

`bool` · optional (explicit presence)

Follow up to 5 redirects before evaluating the response.

### spec.httpConfig.allowInsecure

`bool` · optional (explicit presence)

Accept the origin's TLS certificate without validating it (self-signed
origins). Only meaningful for HTTPS probes.

### spec.httpConfig.headers

`map<string, CloudflareHealthcheckHeaderValues>`

Extra request headers, each with one or more values (e.g. a Host header for
name-based virtual hosts). The map value is a list wrapper because one
header can carry multiple values.

### spec.httpConfig.headers.*.values

`[]string` · required

The header's values (a single-value header is a one-element list).

- rule: {"repeated":{"minItems":"1"}}

### spec.tcpConfig

`CloudflareHealthcheckTcpConfig`

TCP probe details. Only valid when type is TCP.

### spec.tcpConfig.method

`string` · optional (explicit presence)

The probe method. Cloudflare supports only connection_established (a
successful TCP handshake counts as healthy).

- default: `connection_established`
- rule: {"string":{"in":["connection_established"]}}

### spec.tcpConfig.port

`int32` · optional (explicit presence)

The port probed (Cloudflare's default: 80).

- rule: {"int32":{"lte":65535,"gte":1}}

## Validation Rules

- `spec.tcp_config_matches_type`: tcp_config is only valid when type is TCP -- for HTTP/HTTPS checks use http_config
- `spec.http_config_matches_type`: http_config is only valid when type is HTTP or HTTPS (the default type is HTTP) -- for TCP checks use tcp_config

## Outputs

Reference an output from another manifest as `valueFrom: {kind: CloudflareHealthcheck, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.healthcheck_id` | `string` | The ID of the created health check. |
| `status.outputs.zone_id` | `string` | The zone the health check belongs to. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.zoneId` | CloudflareDnsZone | `status.outputs.zone_id` |

## See Also

- [Overview](../README.md)
