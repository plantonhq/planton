# Kubernetes Telemetry

Provision an Istio `Telemetry` resource -- the mesh primitive that controls *how
observability signals are generated* for selected workloads: trace sampling and custom
span tags, metric dimensions and toggles, and access-log providers and filters.

## What Gets Created

- A namespaced `telemetry.istio.io/v1` `Telemetry` custom resource.
- Optional `selector` or `target_refs`, plus any of `tracing`, `metrics`, and
  `access_logging`.

## Prerequisites

- Istio CRDs installed on the cluster (`KubernetesIstioBaseCrds`).
- A running Istio control plane, istiod (`KubernetesIstio`), to apply the configuration.
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
| `selector.match_labels` | map | Pods/VMs the config applies to; matched by istiod, not a foreign key. Mutually exclusive with `target_refs`. |
| `target_refs` | list | Attach to Gateway/Service/ServiceEntry instead of a selector (max 16). |
| `tracing` | list | Sampling, providers, custom span tags, span-reporting toggles. |
| `metrics` | list | Providers, per-metric overrides (enable/disable + tag dimensions), reporting interval. |
| `access_logging` | list | Providers, enable/disable, CEL filter. |

## Stack Outputs

| Output | Description |
|--------|-------------|
| `telemetry_name` | Name of the created Telemetry resource (equals metadata.name). |
| `namespace` | Namespace the Telemetry resource was created in. |

## Related Components

- [Kubernetes Istio](kubernetesistio)
- [Kubernetes Istio Base CRDs](kubernetesistiobasecrds)
- [Kubernetes Namespace](kubernetesnamespace)
