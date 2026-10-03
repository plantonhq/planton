# Kubernetes HorizontalPodAutoscaler

Deploys a Kubernetes HorizontalPodAutoscaler carrying the full autoscaling/v2 surface — resource utilization, per-container resources, custom per-pod metrics, object metrics, and external metrics (queue depths, cloud LB QPS) driving a workload's replica count between a floor and a ceiling, with per-direction scaling behavior. The standalone autoscaler is the right shape for scale targets a Planton workload kind does not manage and for the advanced v2 surface; simple CPU/memory autoscaling of a Planton Deployment belongs in the workload's own availability block instead.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Kubernetes HorizontalPodAutoscaler** -- an autoscaling/v2 HPA in the specified namespace targeting the scale workload, with the replica bounds, metric list, and optional behavior tuning
- **Kubernetes Labels** -- resource metadata labels (resource name, kind, organization, environment) applied automatically for tracking

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** -- an active connection in the Connect module with kubeconfig credentials for the target cluster. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline kubeconfig authentication.

### Kubernetes Cluster

- **A metrics source per family** -- CPU/memory metrics need metrics-server (deploy the Kubernetes MetricsServer component); pods/object metrics need a custom-metrics adapter (prometheus-adapter); external metrics need an external-metrics adapter (KEDA builds on this path).
- The target workload must live in the SAME namespace and expose the scale subresource. Never combine with the workload's own built-in autoscaling.

## Deploy

### Console

Open the deployment store, find **Kubernetes HorizontalPodAutoscaler**, and click **Deploy**. The creation wizard walks you through placement, the scale target, the replica bounds, the metric list, and the per-direction behavior tuning. Start from the **CPU Autoscale** preset for the workhorse case or **Queue Driven** for worker fleets in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesHorizontalPodAutoscaler
metadata:
  name: checkout-hpa
  org: acme-corp
  env: prod
spec:
  name: checkout-hpa
  namespace:
    value: backend-services
  scaleTarget:
    name:
      value: checkout
  minReplicas: 2
  maxReplicas: 10
  metrics:
    - type: resource
      resource:
        name: cpu
        target:
          type: utilization
          averageUtilization: 60
```

```shell
planton apply -f hpa.yaml
```

This holds the `checkout` Deployment's average CPU at 60% of requests, between 2 and 10 replicas. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, wire the autoscaler to its workload by reference so it deploys after the Deployment it governs:

```yaml
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: backend-namespace
      fieldPath: spec.name
  scaleTarget:
    name:
      valueFrom:
        kind: KubernetesDeployment
        name: checkout-deployment
        fieldPath: status.outputs.deployment_name
  maxReplicas: 10
```

The InfraPipeline deploys the namespace and the Deployment first, then creates the autoscaler against them.

## Key Configuration

These are the most important decisions when configuring a Kubernetes HorizontalPodAutoscaler. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**One controller per replica count** -- Once this HPA governs the target, the workload's own `replicas` is advisory. Pointing this AND the workload's built-in autoscaling at the same target flaps the fleet.

**The ceiling is the capacity conversation** -- `maxReplicas` is required and deliberate: an autoscaler without a ceiling is a blank check to the metric driving it. The floor (default 1) is what survives quiet hours.

**Metrics OR toward scale-out** -- Each configured metric proposes a replica count and the HIGHEST wins. With no metrics, Kubernetes applies its default: 80% average CPU utilization (pods must declare CPU requests — utilization measures against them).

**Pick the honest signal** -- CPU falls when load falls (the reliable driver); memory rarely does. Container-resource metrics isolate the app container from sidecars; external metrics ("30 queue messages per pod") are the honest signal for workers.

**Behavior is the flap damper** -- Defaults scale up fast and down conservatively (300-second stabilization). Tune per direction: lengthen the scale-down window for spiky traffic, add a percent policy to bleed capacity gradually, or DISABLE scale-down during an incident.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **KubernetesNamespace** | `namespace` | `spec.name` |
| **KubernetesDeployment** | `scaleTarget.name` | `status.outputs.deployment_name` |

The namespace must be the target's own (omitted means the cluster's `default` namespace); other scale-target kinds are referenced by their exported name output explicitly.

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `horizontal_pod_autoscaler_name` | The name of the created HPA | `kubectl describe hpa` decision auditing |
| `namespace` | The namespace where the HPA was created | Verifying target co-location |
| `scale_target` | The governed workload | Auditing controller ownership |
| `min_replicas` | The replica floor | Capacity audits |
| `max_replicas` | The replica ceiling | Capacity audits |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**CPU workhorse** -- 60% utilization between sane bounds. Start from the **CPU Autoscale** preset.

**Sidecar-proof scaling** -- Measure only the app container. Start from the **Container Isolated** preset.

**Queue-driven workers** -- One pod per N ready messages via an external metric. Start from the **Queue Driven** preset.

**Tuned velocity** -- Longer scale-down stabilization with gradual percent bleed. Start from the **Behavior Tuned** preset.

## Works With

- [**Kubernetes Deployment**](/infra-catalog/kubernetes-deployment) -- the default scale target, referenced declaratively by its exported name.
- [**Metrics Server**](/infra-catalog/kubernetes-metrics-server) -- the prerequisite for CPU/memory metrics.
- [**KEDA**](/infra-catalog/kubernetes-keda) -- event-driven scale-to-zero semantics built on the external-metrics path this HPA consumes.
- [**Kubernetes PodDisruptionBudget**](/infra-catalog/kubernetes-pod-disruption-budget) -- bounds how fast maintenance may shrink what this autoscaler grew.
