# Kubernetes Priority Class

## Overview

**KubernetesPriorityClass** is a catalog kind that creates and manages Kubernetes PriorityClasses — the steps of the cluster's workload importance ladder — as first-class, declaratively managed resources. Pods reference a class by name (the shared workload pod spec's `priority_class_name`); the scheduler places higher-priority pods first when capacity is scarce and — unless preemption is disabled — EVICTS lower-priority pods to make room for a higher-priority pod that cannot otherwise schedule.

The kind covers the complete `scheduling.k8s.io/v1` PriorityClass surface: the priority value, the global-default flag, the human description, and the preemption policy.

## Purpose

Without priority classes, every pod is equally important (priority 0): when a cluster runs out of room, whatever pod happens to arrive next simply stays Pending — a revenue-path API server waits behind a batch job that got there first. PriorityClasses are the standard mechanism for changing that: they rank workloads so scarce capacity goes to what matters, and (optionally) evict what matters less.

**Key value over raw manifests:**

- **Schema-level validation**: The user-value ceiling (at most 1,000,000,000), the reserved `system-` name prefix, and enum-checked preemption policy — all caught before anything reaches the cluster
- **Deterministic preemption policy**: Both IaC modules always send `preemptionPolicy` explicitly (applying the server default `PreemptLowerPriority` when the spec omits it), so the deployed object never depends on which engine applied it
- **Safe value changes**: The priority value is immutable upstream; both modules force delete-before-replace on value change, avoiding the name collision that cluster-unique PriorityClass names would otherwise cause
- **Dual IaC support**: Both Pulumi and Terraform implementations with feature parity
- **Lifecycle management**: Integrated with Planton's deployment lifecycle for status tracking and outputs

## The Importance Ladder

A typical ladder is small and deliberate:

- **`critical`** (value 1000000, preempting) — revenue-path services that must schedule even if something else has to move
- **`standard`** (value 1000, the global default) — everything unmarked
- **`batch`** (value -100, non-preempting) — work that should never displace services and yields to everything

Built-in classes `system-cluster-critical` and `system-node-critical` sit above the user-definable range (their values exceed one billion) and belong to Kubernetes itself — the `system-` name prefix is reserved.

Two scheduler behaviors flow from the value:

- **Queue ordering**: among Pending pods, higher priority schedules first
- **Preemption**: when a higher-priority pod cannot schedule, the scheduler evicts lower-priority pods to make room — unless the class sets `preemption_policy: never`, in which case its pods jump the queue but never evict anything already running (the right policy for high-priority batch work)

## The Global Default

`global_default: true` makes a class the cluster-wide default for pods that name NO priority class (pods that would otherwise get priority 0). Two rules matter:

- **Only one class should be the global default.** When several claim it, Kubernetes uses the SMALLEST such value — rarely what anyone intended
- Changing the default never re-prioritizes existing pods; it applies only to pods created afterwards

## Quick Start

Create a file `priority-class.yaml` — the critical tier of the ladder:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesPriorityClass
metadata:
  name: critical-services
spec:
  name: critical-services
  value: 1000000
  description: Revenue-path services; preempts lower tiers under capacity pressure.
```

Deploy:

```shell
planton apply -f priority-class.yaml
```

Any pod that sets `priority_class_name: critical-services` now schedules ahead of lower-priority pods and, under pressure, evicts them to make room (preemption defaults to `preempt_lower_priority`).

## Essential Configuration Fields

### Required

- **`spec.name`**: The PriorityClass name — the value pods reference in `priority_class_name`. DNS subdomain, and must not use the reserved `system-` prefix
- **`spec.value`**: The priority integer (higher schedules and preempts ahead of lower). User classes must stay at or below 1,000,000,000; negative values are valid and useful for always-preemptable tiers. IMMUTABLE after creation — changing it replaces the class

### Common

- **`spec.global_default`**: Whether this class applies to pods that name no class; keep exactly one per cluster
- **`spec.description`**: Human guidance on when to use this class, surfaced by `kubectl describe priorityclass` — write it for the next engineer choosing a class
- **`spec.preemption_policy`**: `preempt_lower_priority` (default) or `never`
- **`spec.labels`** / **`spec.annotations`**: Merged with standard Planton labels for tracking and governance

## Outputs

After deployment, the following outputs are available in `status.outputs`:

- **`priority_class_name`**: The name of the PriorityClass object as created in the cluster — the composition handle workload pod specs reference
- **`value`**: The priority integer pods of this class receive

## How It Works

This kind includes both **Pulumi** (Go) and **Terraform** (HCL) modules that:

1. Merge user labels and annotations with standard Planton tracking labels
2. Resolve the preemption policy — the explicit value, or the server default `PreemptLowerPriority` — and always submit it explicitly
3. Create the `scheduling.k8s.io/v1` PriorityClass with the value, global-default flag, and description (`global_default` is also sent explicitly, so both engines submit identical objects)
4. Handle value changes as delete-before-replace — the value is immutable upstream, and PriorityClass names are cluster-unique
5. Export the class name and value for downstream composition

Both IaC implementations have feature parity and follow identical logic.

## When to Use

Use **KubernetesPriorityClass** when you need:

- A deliberate importance ladder for a shared cluster — critical / standard / batch tiers
- Guaranteed scheduling for revenue-path services under capacity pressure, at the expense of less important pods
- High-priority batch work that jumps the queue without evicting running services (`preemption_policy: never`)
- A cluster-wide default priority for unmarked pods
- Priority-class names for `KubernetesResourceQuota`'s `priority_class` scope selector to budget against

**Do NOT use** when:

- You want to influence WHERE pods run — priority orders scheduling and preemption, not placement; use node selectors, affinity, and taints
- You want runtime CPU/memory precedence — priority does not change resource allocation on a node; QoS classes and requests/limits do
- The cluster is single-tenant with ample headroom — a ladder with one step is bookkeeping without benefit

## Prerequisites

- **Kubernetes Cluster**: Access to a Kubernetes cluster (PriorityClass is `scheduling.k8s.io/v1`, GA since Kubernetes 1.14 — available everywhere)
- **Credentials**: Kubernetes cluster credentials (kubeconfig)
- **Cluster-scoped permissions**: PriorityClasses are cluster-scoped objects; the credentials must be allowed to create them

## Best Practices

1. **Keep the ladder small and documented**: three to five classes with clear descriptions; every extra step multiplies the preemption interactions to reason about
2. **Exactly one global default per cluster**: audit with `kubectl get priorityclass` before setting `global_default: true` — several defaults resolve to the smallest value
3. **Give batch tiers `never` preemption and negative values**: they jump the queue when idle capacity exists, yield to everything under pressure, and never evict services
4. **Leave gaps between values**: 1000-unit spacing (or more) lets future tiers slot in without renumbering — and renumbering is a replace, not an update
5. **Pair critical tiers with PodDisruptionBudgets**: preemption respects PDBs on a best-effort basis; the victims of your critical tier deserve eviction budgets

## References

- [Pod Priority and Preemption Documentation](https://kubernetes.io/docs/concepts/scheduling-eviction/pod-priority-preemption/)
- [PriorityClass API Reference](https://kubernetes.io/docs/reference/kubernetes-api/workload-resources/priority-class-v1/)
- [Non-preempting Priority Classes](https://kubernetes.io/docs/concepts/scheduling-eviction/pod-priority-preemption/#non-preempting-priority-class)
- [Pod Disruption Budgets](https://kubernetes.io/docs/concepts/workloads/pods/disruptions/#pod-disruption-budgets)

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
