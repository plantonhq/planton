# KubernetesKeda Guide

The judgment this guide carries: KEDA is the autoscaling lane for
REAL-WORLD signals and for scale-to-zero — the one thing plain HPA
cannot do. It is the engine only; the scaling declarations are per-workload
resources deployed beside the workloads they scale.

## When KEDA over a plain HPA

Reach for KEDA when the scale signal is not CPU/memory: queue depth,
stream lag, database rows, cron schedules, cloud metrics (70+ scalers).
And specifically when a workload should idle at ZERO replicas until work
arrives — plain HPA floors at one, KEDA scales to zero and back. Under
the hood KEDA drives an HPA and serves the external-metrics API the HPA
controller reads; the three-lane comparison is the
[HorizontalPodAutoscaler guide](../kuberneteshorizontalpodautoscaler/GUIDE.md).

## Engine here, declarations beside the workload

This kind installs the operator. The scaling rules — ScaledObject,
ScaledJob, TriggerAuthentication — are KEDA custom resources deployed
per workload (via KubernetesManifest today), in the workload's namespace.
Installing KEDA alone scales nothing; a complete proposal names the
ScaledObjects too. Diagram consequence: Manifest-carried ScaledObjects
render as opaque manifest nodes, so the event-scaling topology (which
queue drives which workload) is invisible in the rendered architecture —
state it in the proposal's prose.

## Once per cluster — the external-metrics singleton

KEDA registers the cluster-wide `external.metrics.k8s.io` APIService, and
Kubernetes allows exactly ONE external-metrics provider — a second KEDA
(or another adapter claiming that API) fights it. Release name fixed;
one install, shared-cluster chart, own namespace
([operator-prerequisite pattern](../../_patterns/operator-prerequisite.md)
singleton posture; namespace the
[namespace-ownership pattern](../../_patterns/namespace-ownership.md)'s
sole-tenant case). Note the collision risk: a cluster cannot run KEDA and
another external-metrics adapter at once.

## Pairs well with

- KubernetesHorizontalPodAutoscaler — the lane comparison; KEDA drives an
  HPA underneath.
- The workloads its ScaledObjects target — deployed beside them.
