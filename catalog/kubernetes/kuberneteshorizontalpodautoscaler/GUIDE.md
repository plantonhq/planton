# KubernetesHorizontalPodAutoscaler Guide

The judgment this guide carries: there are THREE ways to autoscale pods
on this platform, and the choice is about the SIGNAL you scale on — CPU
and memory are one lane, and reaching for a standalone HPA when the
inline block would do is the most common over-complication.

## The three lanes

1. **The workload's inline block** — KubernetesDeployment's
   `availability.horizontalPodAutoscaling` (and its siblings). CPU/memory
   targets, zero extra nodes, selector derived automatically. The right
   default; the [Deployment guide](../kubernetesdeployment/GUIDE.md)
   covers when the inline block suffices.
2. **This standalone kind** — the full autoscaling/v2 surface: custom
   per-pod metrics, metrics on other objects, external metrics (queue
   depth, cloud LB QPS), and fine-grained scaling BEHAVIOR (per-direction
   velocity, stabilization windows). Reach for it when scaling on
   anything beyond CPU/memory or when the behavior policies matter.
   Never pair it with an inline block on the same workload — two
   controllers fighting one replica count — and leave the target's
   `availability.replicas` unset, so the workload's own applies never
   reset the count this HPA chose.
3. **[KubernetesKeda](../kuberneteskeda/GUIDE.md)** — event-driven
   scaling on real-world signals (70+ sources) AND scale-to-zero, which
   plain HPA cannot do. When the trigger is a queue that should idle at
   zero, KEDA is the lane (it drives an HPA underneath).

## Metrics need a source, or nothing scales

Every lane reads metrics from somewhere: CPU/memory require
[KubernetesMetricsServer](../kubernetesmetricsserver/GUIDE.md) on
the cluster (absent it, the HPA reports no metrics and never scales —
its guide is the silent-failure home); custom/external metrics need an
adapter (prometheus-adapter, or KEDA serving the external-metrics API).
An autoscaler proposed without its metric source is inert.

## On the diagram

A standalone HPA references its target workload (`scaleTarget` is a
foreign key the platform draws); an inline block adds no node. When the
autoscaling is load-bearing to the architecture, the standalone kind's
visible edge is a point in its favor.

## Pairs well with

- KubernetesMetricsServer — the CPU/memory source every resource-metric
  HPA needs.
- KubernetesKeda — the event-driven / scale-to-zero lane.
- The target workload — referenced by `scaleTarget`.
