# GcpHealthCheck

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

GcpHealthCheckSpec defines a Compute Engine health check — the probe that
decides which backends receive load-balancer traffic and which instances a
managed instance group auto-heals. Health checks are the leaf of the load
balancing family: every backend service references one, and a single check
is commonly shared by many backend services.

One kind covers both scopes. Leave `region` empty for a GLOBAL health check
(used by global external load balancers and backend services); set it for a
REGIONAL health check (used by regional backend services and regional
managed instance groups). The two scopes expose an identical probe surface
in GCP — only `source_regions` (global-only) differs.

The probe protocol is an exclusive choice (http | https | http2 | tcp |
ssl | grpc | grpc_tls) fixed per check; pick the protocol your backends
actually speak on the health-check port, which is not necessarily the
serving protocol (a TCP check on an HTTPS service is valid but shallow).

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpHealthCheck
metadata:
  name: my-sample-health-check
spec:
  # GCP project that owns the health check.
  # Omit to use the provider's default project.
  projectId:
    value: my-gcp-project-123

  # Cloud-side name; omit to default to metadata.name.
  healthCheckName: web-backend-probe

  # Leave region empty for a GLOBAL health check (global external LBs);
  # set it (e.g. us-central1) for a REGIONAL one.

  description: Probes the web tier's /healthz endpoint

  # Probe cadence: every 5s, 5s timeout, 2 successes up / 2 failures down.
  checkIntervalSec: 5
  timeoutSec: 5
  healthyThreshold: 2
  unhealthyThreshold: 2

  # Log every health state transition while tuning thresholds.
  enableLogging: true

  # Exactly one protocol block. USE_SERVING_PORT probes whatever port the
  # backend actually serves on — the right choice for most backends.
  http:
    portSpecification: USE_SERVING_PORT
    requestPath: /healthz

  # Delete the check on destroy (GCP's default, made explicit; applies to
  # whichever scope — global or regional — the check was created in). The
  # API rejects the delete while a backend service still references it.
  deletionPolicy: DELETE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.healthCheckName` | `string` |  |  |  |
| `spec.region` | `string` |  |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.checkIntervalSec` | `int32` |  | `5` |  |
| `spec.timeoutSec` | `int32` |  | `5` |  |
| `spec.healthyThreshold` | `int32` |  | `2` |  |
| `spec.unhealthyThreshold` | `int32` |  | `2` |  |
| `spec.enableLogging` | `bool` |  |  |  |
| `spec.sourceRegions` | `[]string` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |
| `spec.http` | `GcpHealthCheckHttp` |  |  |  |
| `spec.http.host` | `string` |  |  |  |
| `spec.http.port` | `int32` |  |  |  |
| `spec.http.portName` | `string` |  |  |  |
| `spec.http.portSpecification` | `string` |  |  |  |
| `spec.http.proxyHeader` | `string` |  |  |  |
| `spec.http.requestPath` | `string` |  |  |  |
| `spec.http.response` | `string` |  |  |  |
| `spec.https` | `GcpHealthCheckHttps` |  |  |  |
| `spec.https.host` | `string` |  |  |  |
| `spec.https.port` | `int32` |  |  |  |
| `spec.https.portName` | `string` |  |  |  |
| `spec.https.portSpecification` | `string` |  |  |  |
| `spec.https.proxyHeader` | `string` |  |  |  |
| `spec.https.requestPath` | `string` |  |  |  |
| `spec.https.response` | `string` |  |  |  |
| `spec.http2` | `GcpHealthCheckHttp2` |  |  |  |
| `spec.http2.host` | `string` |  |  |  |
| `spec.http2.port` | `int32` |  |  |  |
| `spec.http2.portName` | `string` |  |  |  |
| `spec.http2.portSpecification` | `string` |  |  |  |
| `spec.http2.proxyHeader` | `string` |  |  |  |
| `spec.http2.requestPath` | `string` |  |  |  |
| `spec.http2.response` | `string` |  |  |  |
| `spec.tcp` | `GcpHealthCheckTcp` |  |  |  |
| `spec.tcp.port` | `int32` |  |  |  |
| `spec.tcp.portName` | `string` |  |  |  |
| `spec.tcp.portSpecification` | `string` |  |  |  |
| `spec.tcp.proxyHeader` | `string` |  |  |  |
| `spec.tcp.request` | `string` |  |  |  |
| `spec.tcp.response` | `string` |  |  |  |
| `spec.ssl` | `GcpHealthCheckSsl` |  |  |  |
| `spec.ssl.port` | `int32` |  |  |  |
| `spec.ssl.portName` | `string` |  |  |  |
| `spec.ssl.portSpecification` | `string` |  |  |  |
| `spec.ssl.proxyHeader` | `string` |  |  |  |
| `spec.ssl.request` | `string` |  |  |  |
| `spec.ssl.response` | `string` |  |  |  |
| `spec.grpc` | `GcpHealthCheckGrpc` |  |  |  |
| `spec.grpc.grpcServiceName` | `string` |  |  |  |
| `spec.grpc.port` | `int32` |  |  |  |
| `spec.grpc.portName` | `string` |  |  |  |
| `spec.grpc.portSpecification` | `string` |  |  |  |
| `spec.grpcTls` | `GcpHealthCheckGrpcTls` |  |  |  |
| `spec.grpcTls.grpcServiceName` | `string` |  |  |  |
| `spec.grpcTls.port` | `int32` |  |  |  |
| `spec.grpcTls.portSpecification` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

The GCP project that owns the health check.
Can be a literal project ID or a reference to a GcpProject resource.
If omitted, the provider's default project is used.
Immutable: changing it destroys and recreates the health check.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.healthCheckName

`string`

Name of the health check in GCP. Must be 1-63 characters: lowercase
letters, digits, and hyphens; must start with a letter and end with a
letter or digit. If not specified, defaults to metadata.name.
Immutable: changing it destroys and recreates the health check, briefly
breaking every backend service that references the old self_link.

- rule: health_check_name must be RFC1035-compliant: 1-63 lowercase letters, digits, or hyphens; must start with a letter and end with a letter or digit

### spec.region

`string`

Region for a REGIONAL health check (e.g. "us-central1"), used by regional
backend services and regional managed instance groups. Leave empty for a
GLOBAL health check — the right scope for global external load balancers.
Immutable: a health check cannot move between scopes or regions.

- rule: region must be a valid GCP region name such as us-central1, or empty for a global health check

### spec.description

`string`

What this health check probes and which backends rely on it — write it
for the operator debugging failover behavior later. Mutable.

- rule: {"string":{"maxLen":"2048"}}

### spec.checkIntervalSec

`int32` · optional (explicit presence)

Seconds between probe attempts from each prober (default 5). Lower values
detect failures faster at the cost of more probe traffic; the effective
detection time also depends on unhealthy_threshold. Mutable.

- default: `5`
- rule: {"int32":{"gt":0}}

### spec.timeoutSec

`int32` · optional (explicit presence)

Seconds to wait for a probe response before counting the attempt as a
failure (default 5). Must not exceed check_interval_sec — GCP rejects a
timeout longer than the interval. Mutable.

- default: `5`
- rule: {"int32":{"gt":0}}

### spec.healthyThreshold

`int32` · optional (explicit presence)

Consecutive successes required to mark a backend healthy again
(default 2). Higher values prevent flapping backends from re-entering
rotation too quickly. Mutable.

- default: `2`
- rule: {"int32":{"gt":0}}

### spec.unhealthyThreshold

`int32` · optional (explicit presence)

Consecutive failures required to mark a backend unhealthy (default 2).
Failure detection time ≈ check_interval_sec × unhealthy_threshold — tune
both together when tightening failover. Mutable.

- default: `2`
- rule: {"int32":{"gt":0}}

### spec.enableLogging

`bool`

Export a log entry on every health status change. Off by default —
enable it while tuning thresholds or debugging flapping backends, and
consider the log volume on large backend fleets. Mutable.

### spec.sourceRegions

`[]string`

GLOBAL checks only: probe from exactly 3 specific GCP regions instead of
Google's default prober set, so a regional outage cannot flip global
health verdicts. Constraints enforced by GCP: exactly 3 regions; only
HTTP, HTTPS, and TCP protocols; check_interval_sec at least 30;
proxy_header must be NONE; TCP request payload unsupported; a check with
source_regions cannot be used by managed-instance-group auto-healing.

- rule: source_regions must list exactly 3 GCP regions (or be omitted to use Google's default probers)
- rule: {"repeated":{"items":{"string":{"pattern":"^[a-z]([-a-z0-9]*[a-z0-9])?$"}}}}

### spec.deletionPolicy

`string`

What happens to the health check in GCP when this resource is destroyed.
Applies to whichever scope the check was created in (global or regional).
  "DELETE"  -- (GCP's default when unset) the health check is deleted;
               any backend service still referencing it makes the delete
               fail on the API side, so tear consumers down first
  "PREVENT" -- destroy FAILS; protects a probe that many backend
               services may share
  "ABANDON" -- the health check is removed from management but keeps
               probing in GCP (free at rest; clean it up manually)

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

### spec.http

`GcpHealthCheckHttp`

Probe with an HTTP GET (effective default port 80). The workhorse for
serverless NEGs, instance groups serving plaintext HTTP, and anything
behind an internal load balancer.

- rule: with USE_NAMED_PORT, set port_name and leave port unset
- rule: with USE_SERVING_PORT, leave both port and port_name unset — the backend's serving port is probed

### spec.http.host

`string`

Value of the Host header in the probe request. Empty uses the IP of the
backend being probed — set it when backends route by virtual host.

### spec.http.port

`int32`

TCP port for the probe when port_specification is USE_FIXED_PORT or
unset. Omit to use the protocol default (80). Must be 1-65535.

- rule: port must be between 1 and 65535

### spec.http.portName

`string`

Instance-group named port to probe (only with USE_NAMED_PORT). Named
ports let each instance group map the logical port to its own number.

### spec.http.portSpecification

`string`

How the probe port is chosen: USE_FIXED_PORT (the `port` field, the
default), USE_NAMED_PORT (`port_name` on the instance group), or
USE_SERVING_PORT (the backend's own serving port — the right choice for
serverless NEGs and most instance-group backends).

- rule: port_specification must be USE_FIXED_PORT, USE_NAMED_PORT, or USE_SERVING_PORT

### spec.http.proxyHeader

`string`

Prepend a PROXY protocol v1 header to the probe (PROXY_V1) so backends
behind a proxy see the original client info. Default NONE.

- rule: proxy_header must be NONE or PROXY_V1

### spec.http.requestPath

`string`

Request path of the probe GET (default "/"). Point it at a cheap,
dependency-free endpoint (e.g. /healthz) — a path that touches databases
turns their latency into load-balancer failovers.

### spec.http.response

`string`

If set, the response body must START WITH this ASCII string for the
probe to pass — a guard against a wrong service answering 200 on the
probed port.

### spec.https

`GcpHealthCheckHttps`

Probe with an HTTPS GET (effective default port 443). Requires the
backend to present a certificate; use when backends redirect or reject
plaintext.

- rule: with USE_NAMED_PORT, set port_name and leave port unset
- rule: with USE_SERVING_PORT, leave both port and port_name unset — the backend's serving port is probed

### spec.https.host

`string`

Value of the Host header in the probe request. Empty uses the IP of the
backend being probed — set it when backends route or present
certificates by virtual host.

### spec.https.port

`int32`

TCP port for the probe when port_specification is USE_FIXED_PORT or
unset. Omit to use the protocol default (443). Must be 1-65535.

- rule: port must be between 1 and 65535

### spec.https.portName

`string`

Instance-group named port to probe (only with USE_NAMED_PORT).

### spec.https.portSpecification

`string`

How the probe port is chosen: USE_FIXED_PORT (the `port` field, the
default), USE_NAMED_PORT (`port_name` on the instance group), or
USE_SERVING_PORT (the backend's own serving port).

- rule: port_specification must be USE_FIXED_PORT, USE_NAMED_PORT, or USE_SERVING_PORT

### spec.https.proxyHeader

`string`

Prepend a PROXY protocol v1 header to the probe (PROXY_V1). Default NONE.

- rule: proxy_header must be NONE or PROXY_V1

### spec.https.requestPath

`string`

Request path of the probe GET (default "/"). Point it at a cheap,
dependency-free endpoint (e.g. /healthz).

### spec.https.response

`string`

If set, the response body must START WITH this ASCII string for the
probe to pass.

### spec.http2

`GcpHealthCheckHttp2`

Probe with an HTTP/2 GET (effective default port 443). For backends
that only speak HTTP/2 over TLS.

- rule: with USE_NAMED_PORT, set port_name and leave port unset
- rule: with USE_SERVING_PORT, leave both port and port_name unset — the backend's serving port is probed

### spec.http2.host

`string`

Value of the :authority pseudo-header in the probe request. Empty uses
the IP of the backend being probed.

### spec.http2.port

`int32`

TCP port for the probe when port_specification is USE_FIXED_PORT or
unset. Omit to use the protocol default (443). Must be 1-65535.

- rule: port must be between 1 and 65535

### spec.http2.portName

`string`

Instance-group named port to probe (only with USE_NAMED_PORT).

### spec.http2.portSpecification

`string`

How the probe port is chosen: USE_FIXED_PORT (the `port` field, the
default), USE_NAMED_PORT (`port_name` on the instance group), or
USE_SERVING_PORT (the backend's own serving port).

- rule: port_specification must be USE_FIXED_PORT, USE_NAMED_PORT, or USE_SERVING_PORT

### spec.http2.proxyHeader

`string`

Prepend a PROXY protocol v1 header to the probe (PROXY_V1). Default NONE.

- rule: proxy_header must be NONE or PROXY_V1

### spec.http2.requestPath

`string`

Request path of the probe GET (default "/").

### spec.http2.response

`string`

If set, the response body must START WITH this ASCII string for the
probe to pass.

### spec.tcp

`GcpHealthCheckTcp`

Probe by opening a TCP connection (effective default port 80),
optionally exchanging an ASCII request/response pair. The cheapest
liveness signal when no HTTP endpoint exists.

- rule: with USE_NAMED_PORT, set port_name and leave port unset
- rule: with USE_SERVING_PORT, leave both port and port_name unset — the backend's serving port is probed

### spec.tcp.port

`int32`

TCP port for the probe when port_specification is USE_FIXED_PORT or
unset. Omit to use the protocol default (80). Must be 1-65535.

- rule: port must be between 1 and 65535

### spec.tcp.portName

`string`

Instance-group named port to probe (only with USE_NAMED_PORT).

### spec.tcp.portSpecification

`string`

How the probe port is chosen: USE_FIXED_PORT (the `port` field, the
default), USE_NAMED_PORT (`port_name` on the instance group), or
USE_SERVING_PORT (the backend's own serving port).

- rule: port_specification must be USE_FIXED_PORT, USE_NAMED_PORT, or USE_SERVING_PORT

### spec.tcp.proxyHeader

`string`

Prepend a PROXY protocol v1 header to the probe (PROXY_V1). Default NONE.

- rule: proxy_header must be NONE or PROXY_V1

### spec.tcp.request

`string`

ASCII string to send after the connection opens — pair with `response`
to verify an application-level banner instead of bare connectivity.

### spec.tcp.response

`string`

If set, the first bytes the backend sends must START WITH this ASCII
string for the probe to pass.

### spec.ssl

`GcpHealthCheckSsl`

Probe by completing a TLS handshake (effective default port 443),
optionally exchanging an ASCII request/response pair after the
handshake.

- rule: with USE_NAMED_PORT, set port_name and leave port unset
- rule: with USE_SERVING_PORT, leave both port and port_name unset — the backend's serving port is probed

### spec.ssl.port

`int32`

TCP port for the probe when port_specification is USE_FIXED_PORT or
unset. Omit to use the protocol default (443). Must be 1-65535.

- rule: port must be between 1 and 65535

### spec.ssl.portName

`string`

Instance-group named port to probe (only with USE_NAMED_PORT).

### spec.ssl.portSpecification

`string`

How the probe port is chosen: USE_FIXED_PORT (the `port` field, the
default), USE_NAMED_PORT (`port_name` on the instance group), or
USE_SERVING_PORT (the backend's own serving port).

- rule: port_specification must be USE_FIXED_PORT, USE_NAMED_PORT, or USE_SERVING_PORT

### spec.ssl.proxyHeader

`string`

Prepend a PROXY protocol v1 header to the probe (PROXY_V1). Default NONE.

- rule: proxy_header must be NONE or PROXY_V1

### spec.ssl.request

`string`

ASCII string to send after the TLS handshake completes.

### spec.ssl.response

`string`

If set, the first bytes the backend sends must START WITH this ASCII
string for the probe to pass.

### spec.grpc

`GcpHealthCheckGrpc`

Probe the standard gRPC health checking service
(grpc.health.v1.Health/Check) over plaintext.

- rule: with USE_NAMED_PORT, set port_name and leave port unset
- rule: with USE_SERVING_PORT, leave both port and port_name unset — the backend's serving port is probed

### spec.grpc.grpcServiceName

`string`

The gRPC service name passed to Health/Check. Empty probes the server's
OVERALL health; set it to probe one service on a multi-service server.
The backend's health implementation must recognize the same string.

- rule: {"string":{"maxLen":"1024"}}

### spec.grpc.port

`int32`

TCP port for the probe when port_specification is USE_FIXED_PORT or
unset. Must be 1-65535.

- rule: port must be between 1 and 65535

### spec.grpc.portName

`string`

Instance-group named port to probe (only with USE_NAMED_PORT).

### spec.grpc.portSpecification

`string`

How the probe port is chosen: USE_FIXED_PORT (the `port` field, the
default), USE_NAMED_PORT (`port_name` on the instance group), or
USE_SERVING_PORT (the backend's own serving port).

- rule: port_specification must be USE_FIXED_PORT, USE_NAMED_PORT, or USE_SERVING_PORT

### spec.grpcTls

`GcpHealthCheckGrpcTls`

Probe the standard gRPC health checking service over TLS.

- rule: with USE_SERVING_PORT, leave port unset — the backend's serving port is probed

### spec.grpcTls.grpcServiceName

`string`

The gRPC service name passed to Health/Check. Empty probes the server's
OVERALL health.

- rule: {"string":{"maxLen":"1024"}}

### spec.grpcTls.port

`int32`

TCP port for the probe when port_specification is USE_FIXED_PORT or
unset. Must be 1-65535 when set. Optional — the provider does not
require it; set it explicitly for a deterministic probe target.

- rule: port must be between 1 and 65535

### spec.grpcTls.portSpecification

`string`

How the probe port is chosen: USE_FIXED_PORT (the `port` field, the
default) or USE_SERVING_PORT (the backend's own serving port).
USE_NAMED_PORT is not supported for gRPC-with-TLS probes.

- rule: port_specification must be USE_FIXED_PORT or USE_SERVING_PORT — named ports are not supported for gRPC-with-TLS health checks

## Validation Rules

- `timeout_within_interval`: timeout_sec must not exceed check_interval_sec — GCP rejects a probe timeout longer than the probe interval
- `source_regions_global_only`: source_regions is only supported on global health checks — remove it or clear region
- `source_regions_protocols`: source_regions supports only HTTP, HTTPS, and TCP health checks
- `source_regions_min_interval`: with source_regions set, check_interval_sec must be at least 30 seconds
- `source_regions_no_proxy_header_or_request`: with source_regions set, proxy_header must be NONE and the TCP request payload must be empty

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpHealthCheck, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.self_link` | `string` | Self-link URI of the health check. This is the value backend services reference in their health_checks list — the composition handle for the load balancing family. Global: https://www.googleapis.com/compute/v1/projects/{project}/global/healthChecks/{name} Regional: https://www.googleapis.com/compute/v1/projects/{project}/regions/{region}/healthChecks/{name} |
| `status.outputs.health_check_name` | `string` | Name of the health check as it exists in GCP. |
| `status.outputs.type` | `string` | The probe protocol GCP computed from the configured block (HTTP, HTTPS, HTTP2, TCP, SSL, GRPC, or GRPC_TLS). |
| `status.outputs.region` | `string` | Region of a regional health check; empty for a global one. Downstream composition can use this to confirm scope compatibility (regional backend services require a health check in their own region). |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| GcpBackendService | `spec.healthCheck` | `status.outputs.self_link` |
| GcpComputeMig | `spec.autoHealing.healthCheck` | `status.outputs.self_link` |
| GcpDnsRecord | `spec.routingPolicy.healthCheck` | `status.outputs.self_link` |

## See Also

- [Overview](../README.md)
