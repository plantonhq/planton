# GCP Health Check

Deploys a Compute Engine health check — the probe that decides which backends receive load-balancer traffic and which managed-instance-group VMs get auto-healed. One kind covers both scopes: leave `region` empty for a GLOBAL check (global backend services and external Application Load Balancers) or set it for a REGIONAL one (regional backend services and MIG auto-healing). The probe protocol is an exclusive choice — HTTP, HTTPS, HTTP/2, TCP, SSL, gRPC, or gRPC with TLS — each with its own knobs, and the probe configuration stays mutable in place: retuning the check is its normal day-2 life.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Compute Engine Health Check** -- global or regional, probing with the selected protocol at the configured cadence
- **Probe configuration** -- the request path / banner exchange / gRPC service name, the port choice (fixed, named, or serving port), timing dials, and optional health-transition logging
- **Compute Engine API enablement** -- `compute.googleapis.com` enabled in the target project (never disabled on destroy)

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with credentials for the target GCP project. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### GCP Project

- **A GCP project** where the check will be created. Provide the project ID directly or reference a GcpProject Infra Component via ValueFromRef.
- **A firewall rule for instance-group backends** -- probes originate from Google's ranges `35.191.0.0/16` and `130.211.0.0/22`; without an allow rule every probe fails and the whole service drains. Serverless NEG backends need no firewall work.

## Deploy

### Console

Open the deployment store, find **GCP Health Check**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **HTTP Probe on the Serving Port** preset in the [Presets](#presets) tab to pre-populate the workhorse probe.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpHealthCheck
metadata:
  name: web-serving-hc
  org: acme-corp
  env: prod
spec:
  projectId:
    value: "acme-prod-12345"
  description: Probes the web tier's health endpoint on its serving port
  http:
    portSpecification: USE_SERVING_PORT
    requestPath: /healthz
```

```shell
planton apply -f health-check.yaml
```

This creates the workhorse probe: an HTTP GET of `/healthz` against whatever port each backend actually serves on — the right default for serverless NEGs and most instance groups. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the check to a GCP project deployed in the same InfraPipeline:

```yaml
spec:
  projectId:
    valueFrom:
      kind: GcpProject
      name: production-project
      fieldPath: status.outputs.project_id
```

The InfraPipeline resolves the dependency graph, deploys the project first, then provisions the check — and a downstream GcpBackendService references its `self_link` output in its `healthCheck` field.

## Key Configuration

These are the most important decisions when configuring a health check. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Probe protocol** -- Exactly one of `http`, `https`, `http2`, `tcp`, `ssl`, `grpc`, or `grpcTls`. Pick the protocol your backends actually speak on the probed port — not necessarily the serving protocol. HTTP-family probes GET a request path and can require a response prefix; TCP/SSL probes verify connectivity with an optional ASCII banner exchange; gRPC probes call the standard `grpc.health.v1.Health/Check` service. Mutable — retuning the probe is the check's day-2 life.

**Port choice** -- `USE_FIXED_PORT` (the default) probes the `port` field or the protocol default; `USE_NAMED_PORT` probes the instance group's named port; `USE_SERVING_PORT` probes whatever the backend serves on — the right choice for serverless NEGs. Named ports are not supported for gRPC-with-TLS probes.

**Timing dials** -- `checkIntervalSec`, `timeoutSec`, `healthyThreshold`, `unhealthyThreshold` — all optional with GCP defaults 5/5/2/2. Detection time ≈ interval × unhealthy threshold. The timeout must never exceed the interval. All mutable in place.

**Serving scope** -- Leave `region` empty for a GLOBAL check; set it for a REGIONAL one. The scope must match the consumer, and a check never moves between scopes. Immutable.

**Source regions** -- Global checks only: probe from exactly 3 named GCP regions instead of Google's default prober set, so one regional outage cannot flip global verdicts. Strict constraints: HTTP/HTTPS/TCP probes only, interval ≥ 30s, proxy header NONE, no TCP request payload, and no MIG auto-healing.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** | `projectId` | `status.outputs.project_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `self_link` | Self-link URI of the health check | GcpBackendService `healthCheck` field; GcpDnsRecord routing-policy health checks |
| `health_check_name` | Name as it exists in GCP | Audit, fleet inventory |
| `type` | The probe protocol GCP computed (HTTP/HTTPS/HTTP2/TCP/SSL/GRPC/GRPC_TLS) | Scope verification |
| `region` | Region of a regional check (empty for global) | Scope verification |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**HTTP serving port** -- The workhorse: GET a dedicated health endpoint on whatever port each backend serves. Start from the **HTTP Probe on the Serving Port** preset.

**Regional TCP** -- The cheapest liveness signal for a regional backend service fronting a TCP tier (databases, brokers). Start from the **Regional TCP Probe** preset.

**gRPC service** -- Probe one service's health on a multi-service gRPC server via the standard health protocol. Start from the **gRPC Health Service Probe** preset.

## Works With

- [**GCP Project**](/infra-catalog/gcp-project) -- provides the GCP project where the check is created
- [**GCP Backend Service**](/infra-catalog/gcp-backend-service) -- consumes the check's `self_link` in its `healthCheck` field
- [**GCP DNS Record**](/infra-catalog/gcp-dns-record) -- consumes the check's `self_link` for health-checked DNS routing policies
- [**GCP Region Network Endpoint Group**](/infra-catalog/gcp-region-network-endpoint-group) -- serverless backends behind the same backend service manage their own health; the check covers the non-serverless tiers
