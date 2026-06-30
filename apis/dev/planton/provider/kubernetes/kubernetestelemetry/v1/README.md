# Kubernetes Telemetry

Provision an Istio `Telemetry` resource -- a namespaced configuration of **how telemetry
is generated** for the workloads it selects: distributed-tracing sampling and span tags,
metrics dimensions and toggles, and access-log providers and filters. It does not deploy a
telemetry backend; it tunes what the mesh's sidecars/waypoints emit and to which providers
(declared in the mesh's `MeshConfig`).

A Telemetry resource is hierarchical: a resource with no `selector`/`target_refs` in the
mesh root namespace is the mesh-wide default; a namespace-scoped one overrides it for that
namespace; a workload-scoped one (via `selector` or `target_refs`) overrides it for the
selected workloads.

## What Gets Created

- A namespaced `telemetry.istio.io/v1` `Telemetry` custom resource.
- Optional `selector` **or** `target_refs` (at most one), plus any of `tracing`,
  `metrics`, and `access_logging`, scoped to this resource's namespace.

## How it relates to the other Istio resources

- **Telemetry** decides *what observability signals* are produced (traces, metrics, logs).
- **DestinationRule / VirtualService / ServiceEntry** shape *traffic*; Telemetry observes it.
- Providers named here (e.g. `prometheus`, `zipkin`, `envoy`, `otel`) must be defined as
  extension providers in the mesh's `MeshConfig`.

## Prerequisites

- Istio CRDs installed on the cluster (`KubernetesIstioBaseCrds`).
- A running Istio control plane, istiod (`KubernetesIstio`), to apply the configuration.
  The resource applies with only the CRDs present, but it only affects telemetry where
  istiod and a sidecar (or ambient/waypoint) data plane are running.
- The target namespace (`KubernetesNamespace`).

## Quick Start

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesTelemetry
metadata:
  name: mesh-default
spec:
  namespace:
    value: istio-system
  tracing:
    - random_sampling_percentage: 10
```

```bash
planton apply -f telemetry.yaml
```

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `namespace` | reference | Namespace the Telemetry resource is created in. |

### Optional Fields

| Field | Type | Description |
|-------|------|-------------|
| `selector.match_labels` | map | Pods/VMs (by label) this config applies to. Matched by istiod; not a foreign key. Mutually exclusive with `target_refs`. |
| `target_refs` | list | Attach to specific resources (Gateway/Service/ServiceEntry) instead of a label selector. Max 16. Mutually exclusive with `selector`. |
| `tracing` | list | Tracing behavior: `match`, `providers`, `random_sampling_percentage`, `disable_span_reporting`, `custom_tags`, `enable_istio_tags`. |
| `metrics` | list | Metrics behavior: `providers`, `overrides` (per-metric enable/disable + tag dimensions), `reporting_interval`. |
| `access_logging` | list | Access-log behavior: `match`, `providers`, `disabled`, `filter` (CEL). |

### `tracing[]`

| Field | Type | Description |
|-------|------|-------------|
| `match.mode` | string | `CLIENT_AND_SERVER` / `CLIENT` / `SERVER`. |
| `providers[].name` | string | Tracing provider(s) from MeshConfig (e.g. `zipkin`). |
| `random_sampling_percentage` | number | 0.00-100.00 (0.01% increments). |
| `disable_span_reporting` | bool | Stop reporting spans (context propagation continues). |
| `custom_tags` | map | Tag name -> one of `literal`, `environment`, or `header`. |
| `enable_istio_tags` | bool | Include Istio-specific span tags (default true). |

### `metrics[]`

| Field | Type | Description |
|-------|------|-------------|
| `providers[].name` | string | Metrics provider(s) from MeshConfig (e.g. `prometheus`). |
| `overrides[].match` | object | `metric` (a standard metric) **or** `custom_metric`, plus `mode`. |
| `overrides[].disabled` | bool | Turn the matched metric(s) off. |
| `overrides[].tag_overrides` | map | Tag name -> `{ operation: UPSERT/REMOVE, value: <CEL> }`. |
| `reporting_interval` | string | TCP metrics report interval (duration, >= 1ms; default `5s`). |

### `access_logging[]`

| Field | Type | Description |
|-------|------|-------------|
| `match.mode` | string | `CLIENT_AND_SERVER` / `CLIENT` / `SERVER`. |
| `providers[].name` | string | Logging provider(s) from MeshConfig (e.g. `envoy`). |
| `disabled` | bool | Turn access logging off for the matched workloads. |
| `filter.expression` | string | CEL selecting which requests/connections are logged. |

## Examples

### Mesh-wide 10% trace sampling with a custom tag

```yaml
spec:
  namespace:
    value: istio-system
  tracing:
    - random_sampling_percentage: 10
      custom_tags:
        cluster:
          literal:
            value: prod-us-east
```

### Add/remove Prometheus metric dimensions for a namespace

```yaml
spec:
  namespace:
    value: bookinfo
  metrics:
    - providers:
        - name: prometheus
      overrides:
        - tag_overrides:
            request_host:
              operation: UPSERT
              value: request.host
            response_code:
              operation: REMOVE
```

### Disable access logging for selected workloads

```yaml
spec:
  namespace:
    value: bookinfo
  selector:
    match_labels:
      app: ratings
  access_logging:
    - disabled: true
```

## Composing in Infra Charts

`namespace` is a `StringValueOrRef`, so it can reference a `KubernetesNamespace` output via
`valueFrom`. The `selector.match_labels` and `target_refs` are **not** foreign keys --
istiod resolves them at runtime, so they create no automatic DAG edge. To order this
Telemetry relative to the workloads or gateways it observes, declare the dependency on
`metadata.relationships`:

```yaml
metadata:
  name: ratings-tracing
  relationships:
    - kind: KubernetesDeployment
      name: ratings
      type: depends_on
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: bookinfo-ns
      fieldPath: spec.name
  selector:
    match_labels:
      app: ratings
  tracing:
    - random_sampling_percentage: 50
```

See `docs/README.md` for the full composability rationale.

## Stack Outputs

| Output | Description |
|--------|-------------|
| `telemetry_name` | Name of the created Telemetry resource (equals metadata.name). |
| `namespace` | Namespace the Telemetry resource was created in. |

## Related Components

- [Kubernetes Istio](kubernetesistio)
- [Kubernetes Istio Base CRDs](kubernetesistiobasecrds)
- [Kubernetes Namespace](kubernetesnamespace)
