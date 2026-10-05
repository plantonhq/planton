# GCP Health Check

Deploys a Compute Engine health check (`google_compute_health_check` / `google_compute_region_health_check`) — the probe that decides which backends receive load-balancer traffic and which instances a managed instance group auto-heals.

## What Gets Created

When you deploy a GcpHealthCheck resource, Planton provisions exactly one of:

- **Global health check** — when `region` is empty; referenced by global backend services (global external HTTP(S) load balancers)
- **Regional health check** — when `region` is set; referenced by regional backend services and regional managed instance groups

Both scopes expose an identical probe surface in GCP — only `sourceRegions` (global-only) differs — which is why they are one kind with a scope switch rather than two kinds.

## Prerequisites

- **GCP credentials** configured via environment variables or Planton provider config
- **An existing GCP project** — referenced via `projectId` (or the provider's default project)
- **IAM permissions** — see [`iac/permissions.yaml`](iac/permissions.yaml) for the least-privilege permission set the deploying principal needs

## Quick Start

Create a file `health-check.yaml`:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpHealthCheck
metadata:
  name: web-backend-probe
spec:
  projectId:
    value: my-gcp-project-123
  description: Probes the web tier's /healthz endpoint
  http:
    portSpecification: USE_SERVING_PORT
    requestPath: /healthz
```

Deploy:

```shell
planton apply -f health-check.yaml
```

This creates a global HTTP health check probing every backend's serving port at `/healthz` on the default cadence (every 5 s, 5 s timeout, 2 successes up / 2 failures down).

## Configuration Reference

### Protocol (exactly one required)

| Block | Probe | Typical use |
|-------|-------|-------------|
| `http` | HTTP GET on `requestPath` | Serverless NEGs, plaintext web tiers, anything behind an internal LB |
| `https` | TLS + HTTP GET | Backends that reject plaintext |
| `http2` | HTTP/2-over-TLS GET | HTTP/2-only backends |
| `tcp` | TCP connect (+ optional byte exchange) | Databases, custom TCP protocols |
| `ssl` | TLS handshake (+ optional byte exchange) | TLS services without HTTP |
| `grpc` | gRPC health service (plaintext) | gRPC microservices |
| `grpcTls` | gRPC health service over TLS | TLS gRPC microservices |

Each HTTP-family block accepts `host`, `port`, `portName`, `portSpecification`, `proxyHeader`, `requestPath`, `response`; TCP/SSL swap `requestPath` for a `request` payload; gRPC blocks take `grpcServiceName` + port fields. Ports left unset use GCP's protocol defaults (HTTP/TCP 80, HTTPS/HTTP2/SSL 443).

`portSpecification` picks how the probe port is chosen:

- `USE_FIXED_PORT` (default) — the numeric `port`
- `USE_NAMED_PORT` — the instance group's named port in `portName` (not supported by `grpcTls`)
- `USE_SERVING_PORT` — whatever port the backend actually serves on; the right choice for serverless NEGs and most instance-group backends

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `projectId` | `StringValueOrRef` | provider default project | GCP project that owns the check. Can reference a GcpProject resource. Immutable. |
| `healthCheckName` | `string` | `metadata.name` | Cloud-side name (RFC1035). Immutable. |
| `region` | `string` | `""` (global) | Set for a REGIONAL health check. Immutable. |
| `description` | `string` | `""` | What the check probes and which backends rely on it. |
| `checkIntervalSec` | `int` | `5` | Seconds between probes from each prober. |
| `timeoutSec` | `int` | `5` | Seconds to wait for a response. Must not exceed `checkIntervalSec`. |
| `healthyThreshold` | `int` | `2` | Consecutive successes to mark healthy. |
| `unhealthyThreshold` | `int` | `2` | Consecutive failures to mark unhealthy. |
| `enableLogging` | `bool` | `false` | Log every health state transition to Cloud Logging. |
| `sourceRegions` | `list(string)` | `[]` | Global-only: probe from exactly 3 named regions. Requires `checkIntervalSec` ≥ 30, HTTP/HTTPS/TCP protocol, no proxy header. |

## Tuning Failover

Failure detection time ≈ `checkIntervalSec` × `unhealthyThreshold` (default ≈ 10 s). Tightening either speeds up failover at the cost of more probe traffic and more sensitivity to blips; `healthyThreshold` guards the other direction, keeping flapping backends out of rotation until they prove stable. Point HTTP probes at a cheap, dependency-free endpoint — a `/healthz` that touches your database turns database latency into load-balancer failovers.

## Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `self_link` | `string` | Self-link URI — the value backend services reference in their `health_checks` list |
| `health_check_name` | `string` | Name of the health check in GCP |
| `type` | `string` | Probe protocol GCP computed (HTTP, HTTPS, HTTP2, TCP, SSL, GRPC, GRPC_TLS) |
| `region` | `string` | Region of a regional check; empty for global |

## Deployment Methods

Planton supports two deployment methods:

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md) for Pulumi-specific deployment instructions.

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md) for Terraform-specific deployment instructions.

## Important Notes

- **Scope is permanent**: `region` (and the global/regional choice it encodes) is immutable — a health check cannot move between scopes. Regional backend services can only reference health checks in their own region; global backend services only global ones.
- **Immutability**: `healthCheckName` and `projectId` are ForceNew — changing either destroys and recreates the check, briefly breaking every backend service referencing the old `self_link`. All probe knobs update in place.
- **Firewall rules are separate**: probes originate from Google's prober ranges (`35.191.0.0/16`, `130.211.0.0/22` for most load balancers) — instance-group backends need an ingress firewall rule admitting them (a GcpFirewallRule), or every backend shows unhealthy. Serverless backends need nothing.
- **`sourceRegions` disables auto-healing use**: a check probing from named regions cannot be attached to managed-instance-group auto-healing.
- **gRPC-with-TLS**: the `grpcTls` probe is a preview-stage surface on the current provider line; the modules select the beta provider so it is available without a retrofit.

## Related Kinds

- [GcpCloudArmorPolicy](/docs/catalog/gcp/gcpcloudarmorpolicy) — protects the backends this check keeps in rotation
- [GcpFirewallRule](/docs/catalog/gcp/gcpfirewallrule) — admits Google's prober ranges to instance-group backends
- [GcpProject](/docs/catalog/gcp/gcpproject) — provides the GCP project that owns the check

## Additional Resources

- [Health checks overview](https://cloud.google.com/load-balancing/docs/health-check-concepts)
- [Creating health checks](https://cloud.google.com/load-balancing/docs/health-checks)

## Support

For issues, questions, or contributions, please refer to the Planton documentation or open an issue in the repository.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
