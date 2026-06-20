---
title: "Mesh-Wide Trace Sampling"
description: "The canonical Telemetry resource: turn on distributed-tracing sampling for the whole mesh and (optionally) stamp every span with an operator-supplied tag. This is how you get traces flowing to your..."
type: "preset"
rank: "01"
presetSlug: "01-mesh-tracing-sampling"
componentSlug: "telemetry"
componentTitle: "Telemetry"
provider: "kubernetes"
icon: "package"
order: 1
---

# Mesh-Wide Trace Sampling

The canonical Telemetry resource: turn on distributed-tracing sampling for the whole mesh
and (optionally) stamp every span with an operator-supplied tag. This is how you get traces
flowing to your tracing backend without editing the mesh's `MeshConfig` -- you pick the
sample rate and tags declaratively.

## When to Use

- You have a tracing provider configured in `MeshConfig` (e.g. `zipkin`, `otel`) and want a
  consistent sample rate across all workloads.
- You want a constant dimension on every span (cluster name, region, environment) for
  filtering in your tracing UI.

## Key Configuration Choices

- **No `selector`/`target_refs`** -- placed in the mesh root namespace (`istio-system`),
  this is the mesh-wide default. Add a `selector` to scope it to specific workloads.
- **`tracing[].random_sampling_percentage`** -- the sample rate (0.00-100.00). Start low in
  production (1-10%) to bound overhead.
- **`tracing[].custom_tags`** -- add a hard-coded `literal`, an `environment` variable read
  by the sidecar, or a request `header` value to every span.

## Prerequisites

- The Istio CRDs are installed (`KubernetesIstioBaseCrds`).
- istiod is running with a tracing provider configured in `MeshConfig`, and the workloads
  have sidecars or are in the ambient mesh (`KubernetesIstio`).
- The target namespace exists (`KubernetesNamespace`).

## Placeholders to Replace

| Placeholder | Description |
|-------------|-------------|
| `<root-namespace>` | The Istio root/config namespace (usually `istio-system`). |
| `<sample-percentage>` | Trace sample rate, e.g. `10`. |
| `<tag-value>` | A literal value to stamp on every span (e.g. your cluster name). |
