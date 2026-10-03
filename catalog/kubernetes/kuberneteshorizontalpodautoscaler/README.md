# Kubernetes Horizontal Pod Autoscaler

## Overview

**KubernetesHorizontalPodAutoscaler** is a catalog kind that creates and manages Kubernetes HorizontalPodAutoscalers — automatic replica scaling driven by observed metrics — as first-class, declaratively managed resources. An HPA points at one scale target (`scale_target`) and adjusts its replica count between a floor (`min_replicas`) and a ceiling (`max_replicas`), driven by one or more metrics.

The kind covers the complete `autoscaling/v2` surface: resource utilization metrics (CPU/memory), per-container resource metrics, custom per-pod metrics, metrics on other objects, external metrics (queue depths, cloud load balancer QPS), and fine-grained scaling behavior — per-direction velocity policies and stabilization windows. There is nothing an upstream `autoscaling/v2` HPA can express that this spec cannot.

## Purpose

Fixed replica counts are always wrong twice a day: too few at peak, too many at trough. The HorizontalPodAutoscaler is the standard Kubernetes answer — a control loop that measures load and holds the fleet at the size the load demands. Once an HPA governs a workload, the workload's own `replicas` field becomes advisory: the autoscaler owns the count.

**Key value over raw manifests:**

- **Schema-level validation**: Every metric must carry exactly the source matching its declared type (a mismatch deploys a metric the controller silently ignores), every target exactly the value form matching its target type, floor ≤ ceiling, DaemonSet targets rejected, quantity and percentage format checks, and behavior contracts (a `disabled` direction cannot list policies) — all caught before anything reaches the cluster
- **Namespace and target by value or reference**: `spec.namespace` accepts a literal or a `KubernetesNamespace` reference; `scale_target.name` accepts a literal or a reference to a `KubernetesDeployment`'s exported name, so an infra chart deploys the workload and its autoscaler in one run
- **Deterministic defaults**: Both IaC modules apply the spec defaults (apps/v1 Deployment target, `min_replicas` 1) module-side and always send them explicitly, so the deployed object never depends on which engine applied it
- **Dual IaC support**: Both Pulumi and Terraform implementations with feature parity
- **Lifecycle management**: Integrated with Planton's deployment lifecycle for status tracking and outputs

## Metrics: Each Proposes, the Highest Wins

When several metrics are configured, each one independently proposes a replica count and the controller takes the **highest** — metrics OR together toward scale-out. A workload scaled on CPU *and* queue depth scales up when either is hot and scales down only when both are cold.

The five metric source families:

- **`resource`** — a pod resource (cpu/memory) averaged across the target's pods; requires metrics-server. The workhorse. CPU is the reliable scaling signal; memory rarely falls after scale-out, which makes it a poor scale-in driver
- **`container_resource`** — the same, for ONE named container in each pod — isolates the app container from sidecars that would skew the pod-level average
- **`pods`** — a custom metric exposed per pod (e.g. requests-per-second), averaged; requires a custom-metrics adapter
- **`object`** — a metric describing ONE other object (e.g. an Ingress's requests-per-second); requires a custom-metrics adapter
- **`external`** — a metric from outside the cluster (queue depth, cloud LB QPS); requires an external-metrics adapter (e.g. prometheus-adapter or KEDA)

When `metrics` is EMPTY, Kubernetes applies its default: 80% average CPU utilization — which requires the pods to declare CPU requests, because utilization is measured against requests.

### Target value forms

Each metric holds a target expressed in exactly one of three forms:

- **`utilization`** — a percentage of the pods' resource REQUESTS (resource/container_resource only)
- **`average_value`** — a quantity averaged across the target's pods (the usual form for pods/external metrics, e.g. "30 queue messages per pod")
- **`raw_value`** — a raw quantity compared to the metric directly (object/external metrics); this maps to the upstream API's `Value` target type

## Scaling Behavior

`behavior` tunes velocity per direction. Omit it for the Kubernetes defaults: scale UP fast (the higher of doubling or +4 pods per 15s, no stabilization), scale DOWN conservatively (to the recommendation's 300-second high-water mark). Each direction takes:

- **`stabilization_window_seconds`** (0–3600) — the flap damper: the safest recommendation in the window wins; 0 acts immediately
- **`policies`** — velocity caps: at most N pods or N percent of current replicas per period; `select_policy` (`max_change` default, `min_change`, or `disabled`) arbitrates between several. `disabled` turns a direction off entirely — e.g. freeze scale-down during an incident while leaving scale-up live

## Standalone Autoscaler vs the Workload's Built-in Block

Planton's KubernetesDeployment carries its own `availability.horizontal_pod_autoscaling` block. The boundary is:

- **Use the workload's built-in block** for simple CPU/memory autoscaling of a Planton Deployment's OWN replicas — the module manages the HPA with the workload's `replicas` as the floor, wired to the right target automatically.
- **Use this standalone kind** for scale targets a Planton workload kind does not manage — operator-managed Deployments, non-Planton workloads — and for the advanced `autoscaling/v2` surface: custom/object/external metrics, per-container metrics, and behavior tuning.

**Never point both at the same target.** Two controllers fighting over one replica count flaps the fleet.

## Essential Configuration Fields

### Required

- **`spec.name`**: The HorizontalPodAutoscaler name (DNS subdomain: lowercase alphanumeric, hyphens, dots, 1–253 chars)
- **`spec.scale_target`**: The workload whose replica count this autoscaler owns — `kind` (default `Deployment`; DaemonSets are rejected — one pod per node by definition), `api_version` (default `apps/v1`), and `name` (literal or reference to a `KubernetesDeployment`)
- **`spec.max_replicas`**: The replica ceiling — required, and the honest capacity conversation: what is the most this workload may cost?

### Common

- **`spec.namespace`**: Literal namespace name or reference to a KubernetesNamespace resource — MUST be the scale target's own namespace (an HPA cannot scale across namespaces). When omitted, the autoscaler lands in the cluster's `default` namespace
- **`spec.min_replicas`**: The replica floor (default 1; must be ≥ 1)
- **`spec.metrics`**: The metrics driving the decision; empty applies the Kubernetes default (80% average CPU utilization)
- **`spec.behavior`**: Per-direction velocity and stabilization tuning
- **`spec.labels`** / **`spec.annotations`**: Merged with standard Planton labels for tracking and governance

## Outputs

After deployment, the following outputs are available in `status.outputs`:

- **`horizontal_pod_autoscaler_name`**: The name of the HorizontalPodAutoscaler object as created in the cluster
- **`namespace`**: The namespace the autoscaler was created in
- **`scale_target`**: The scale target as `"Kind/name"` (e.g. `"Deployment/checkout"`)
- **`min_replicas`** / **`max_replicas`**: The configured floor and ceiling

## How It Works

This kind includes both **Pulumi** (Go) and **Terraform** (HCL) modules that:

1. Resolve the target namespace and scale-target name (literal values or resolved references)
2. Merge user labels and annotations with standard Planton tracking labels
3. Apply the spec defaults module-side — apps/v1 Deployment target, `min_replicas` 1 — and always send them explicitly
4. Create the `autoscaling/v2` HorizontalPodAutoscaler with the metrics and behavior mapped one-to-one; when the spec lists no metrics, the field is OMITTED so the API server applies its own default (sending an empty list would instead disable metric-driven scaling)
5. Export the autoscaler name, namespace, scale target, and replica bounds for downstream composition

Both IaC implementations have feature parity and follow identical logic.

## When to Use

Use **KubernetesHorizontalPodAutoscaler** when you need:

- Autoscaling for operator-managed or non-Planton workloads — anything exposing the `scale` subresource
- Queue-driven or traffic-driven scaling on custom, object, or external metrics
- Per-container resource metrics that exclude sidecar noise from the scaling signal
- Explicit scale-velocity control — stabilization windows, percent-per-minute caps, or freezing one direction

**Do NOT use** when:

- The target is a Planton Deployment and simple CPU/memory scaling suffices — use the workload's built-in `availability.horizontal_pod_autoscaling` block
- Another autoscaler (including the built-in block, or KEDA) already owns the target's replica count — never point two controllers at one target
- The workload cannot scale horizontally — DaemonSets (rejected by the schema), singletons, or anything whose replicas are not interchangeable
- You need scale-to-zero — that is feature-gated upstream and not modeled; use KEDA for event-driven scale-to-zero semantics

## Prerequisites

- **Kubernetes Cluster**: Access to a Kubernetes cluster
- **Credentials**: Kubernetes cluster credentials (kubeconfig)
- **metrics-server** for resource metrics (CPU/memory) — most managed clusters ship it; without it, resource metrics read as unavailable
- **A custom/external-metrics adapter** (prometheus-adapter, KEDA, a cloud adapter) for `pods`, `object`, and `external` metrics
- **Resource requests declared on the pods** for `utilization` targets — utilization is measured against requests
- **Namespace**: The target namespace must exist before creating the autoscaler (unless deploying to `default`, or creating the namespace in the same chart via a reference). The scale target itself is not required at creation — the controller reports it unresolved until the workload appears, the correct steady state for infra deployed ahead of the app

## Best Practices

1. **Scale on CPU, alert on memory**: CPU rises and falls with load; memory rarely falls after scale-out, making it a scale-in trap
2. **Set requests before setting utilization targets**: utilization is a percentage of requests — unset or wrong requests make the signal meaningless
3. **Treat `max_replicas` as a budget decision**: it is the answer to "what is the most this workload may cost", not a number to set generously and forget
4. **Use `container_resource` when sidecars skew the average**: a hot proxy sidecar can mask an idle app container and vice versa
5. **Lengthen scale-down stabilization for spiky traffic**: the 300-second default handles ordinary noise; queue-driven and bursty services often want 600s plus a percent-per-minute policy to bleed capacity gradually
6. **Let the HPA own the count**: once governed, stop setting `replicas` on the workload — it is advisory from then on and fighting it causes churn

## References

- [Horizontal Pod Autoscaling Documentation](https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/)
- [HorizontalPodAutoscaler Walkthrough](https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale-walkthrough/)
- [HorizontalPodAutoscaler API Reference](https://kubernetes.io/docs/reference/kubernetes-api/workload-resources/horizontal-pod-autoscaler-v2/)
- [Metrics APIs (metrics-server, custom, external)](https://kubernetes.io/docs/tasks/debug/debug-cluster/resource-metrics-pipeline/)

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
