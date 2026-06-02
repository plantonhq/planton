# Prometheus Metric Dimensions

Customize the dimensions (labels) Istio attaches to its Prometheus metrics for a namespace
or workload: add high-value tags (request host, method) and drop noisy/high-cardinality
ones (response code). This is how you keep your metrics useful and your Prometheus bill
under control without touching the mesh's `MeshConfig`.

## When to Use

- You want extra dimensions on Istio's standard metrics for better slicing in Grafana.
- A dimension (often `response_code` or `source_*`) is exploding cardinality and you want it
  removed from specific metrics.

## Key Configuration Choices

- **`metrics[].providers[].name: prometheus`** -- target the Prometheus provider configured
  in `MeshConfig`.
- **`overrides[].match.metric`** -- one of the standard metrics (`REQUEST_COUNT`,
  `REQUEST_DURATION`, ...) or omit `match` to apply to all. `mode` (`CLIENT`/`SERVER`)
  narrows by traffic direction.
- **`overrides[].tag_overrides`** -- per tag, `operation: UPSERT` with a CEL `value` to
  add/update a dimension, or `operation: REMOVE` to drop it. `value` must be set for UPSERT
  and must be absent for REMOVE.

## Prerequisites

- The Istio CRDs are installed (`KubernetesIstioBaseCrds`).
- istiod is running with the Prometheus provider configured in `MeshConfig`, and the
  workloads have sidecars or are in the ambient mesh (`KubernetesIstio`).
- The target namespace exists (`KubernetesNamespace`).

## Placeholders to Replace

| Placeholder | Description |
|-------------|-------------|
| `<namespace>` | Namespace the override applies to (e.g. `bookinfo`). |

Scope to specific workloads by adding a `selector.match_labels`.
