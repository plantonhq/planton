# Kubernetes Horizontal Pod Autoscaler - Pulumi Module

## Overview

This Pulumi module creates and manages a Kubernetes `autoscaling/v2` HorizontalPodAutoscaler. It supports the complete v2 surface: the scale target reference, replica floor and ceiling, all five metric source families (resource, container_resource, pods, object, external) with all three target value forms (utilization, value, average value), and per-direction scaling behavior with stabilization windows and velocity policies.

## Architecture

```
iac/pulumi/
├── main.go                      # Entrypoint: loads IaC input, calls module
├── Pulumi.yaml                  # Pulumi project configuration
├── Makefile                     # Make targets for preview/up/down/refresh
└── module/
    ├── main.go                  # Orchestrator: provider init, resource creation, output export
    ├── locals.go                # Derived values: labels, annotations, namespace default, resolved target and floor, enum→API-string maps
    ├── horizontalpodautoscaler.go # Creates kubernetes.autoscaling.v2.HorizontalPodAutoscaler resource
    └── outputs.go               # Exports horizontal_pod_autoscaler_name, namespace, scale_target, min_replicas, max_replicas
```

## How It Works

1. **IaC Input Loading**: The entrypoint loads `KubernetesHorizontalPodAutoscalerIacInput` from Pulumi config
2. **Locals Initialization**: `locals.go` computes:
   - Standard Planton labels merged with user labels (identity keys cannot be overridden)
   - User annotations
   - The target namespace (foreign-key references are pre-resolved; falls back to `default` when omitted)
   - The resolved scale target with the spec defaults applied (`apps/v1`, `Deployment`); the target name reference is pre-resolved to a literal
   - The resolved replica floor with the Kubernetes default (1) applied
3. **Provider Creation**: Kubernetes provider is initialized from `provider_config`
4. **HorizontalPodAutoscaler Creation**: A single `autoscaling/v2` HPA is created with the scale target, replica bounds, metrics, and behavior
5. **Output Export**: Autoscaler name, namespace, scale target (`Kind/name`), and replica bounds are exported as outputs

## Semantics Preserved by the Module

- **`minReplicas` is always sent explicitly** — with the Kubernetes default (1) applied module-side, so the Pulumi and Terraform modules submit byte-identical objects for the same manifest
- **An empty metrics list is OMITTED, not sent** — the API server then applies its own default (80% average CPU utilization); sending an empty list would instead disable metric-driven scaling
- **Each metric maps to exactly one API metric source** — the spec's CEL rules guarantee exactly the source matching each metric's declared type is present, and exactly the value form matching each target's type
- **Enum names map to the API strings** — `raw_value` → `Value` (the enum is named `raw_value` only because generated code cannot use the bare word "value"), `max_change`/`min_change`/`disabled` → `Max`/`Min`/`Disabled`, `pods`/`percent` → `Pods`/`Percent`
- **`selectPolicy` is always sent in behavior blocks** — with the API default (`Max`) applied module-side

## Field Mapping

| Spec Field | HorizontalPodAutoscaler Field | Notes |
|------------|-------------------------------|-------|
| `scale_target` | `spec.scaleTargetRef` | Defaults `apps/v1` / `Deployment` applied; name pre-resolved from references |
| `min_replicas` / `max_replicas` | `spec.minReplicas` / `spec.maxReplicas` | Floor default 1, always explicit |
| `metrics[]` | `spec.metrics[]` | Types mapped to `Resource`/`ContainerResource`/`Pods`/`Object`/`External`; empty list omitted |
| `metrics[].*.target` | `target` | `utilization` → `Utilization` + `averageUtilization`; `raw_value` → `Value` + `value`; `average_value` → `AverageValue` + `averageValue` |
| `metric.match_labels` | `metric.selector.matchLabels` | Scopes the series the metrics adapter reads |
| `behavior.scale_up` / `scale_down` | `spec.behavior.scaleUp` / `scaleDown` | `selectPolicy` always explicit; policies mapped one-to-one |

## Usage

```bash
# Preview changes
make preview manifest=../../e2e/manifest.yaml

# Deploy
make up manifest=../../e2e/manifest.yaml

# Destroy
make down manifest=../../e2e/manifest.yaml
```

## Debug

```bash
# Build the module
go build ./module/...

# Build the entrypoint
go build .
```

> **Note**: The HPA reads metrics; it does not produce them. Resource metrics require metrics-server; `pods`/`object`/`external` metrics require a custom/external-metrics adapter (e.g. prometheus-adapter, KEDA). The scale target's existence is not required at creation — the controller reports it unresolved until the workload appears.
