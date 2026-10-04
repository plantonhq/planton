# Kubernetes Pod Disruption Budget

## Overview

**KubernetesPodDisruptionBudget** is a catalog kind that creates and manages Kubernetes PodDisruptionBudgets — the availability floor for voluntary disruptions — as first-class, declaratively managed resources. A PodDisruptionBudget selects a set of pods with `selector` and declares how many of them may be taken down at once during node drains, cluster upgrades, and autoscaler consolidation.

The kind covers the complete `policy/v1` PodDisruptionBudgetSpec surface: exact-match and set-based label selectors, absolute and percentage availability bounds (`min_available` / `max_unavailable`), and the unhealthy-pod eviction policy. There is nothing an upstream PodDisruptionBudget can express that this spec cannot.

## Purpose

Kubernetes routinely moves pods on purpose: an administrator drains a node for maintenance, the cluster autoscaler consolidates underused nodes, an upgrade rolls through the fleet. Without a budget, nothing stops all replicas of a service from being evicted at the same moment. A PodDisruptionBudget makes the eviction API refuse any eviction that would breach the declared floor — the drain waits, retries, and proceeds only as replacement pods become ready.

**Key value over raw manifests:**

- **Schema-level validation**: Exactly one availability bound enforced (the API rejects both, and a budget with neither protects nothing), int-or-percent format checks on the bounds, selector operator contracts (`In`/`NotIn` require values, `Exists`/`DoesNotExist` forbid them), and a required selector — all caught before anything reaches the cluster
- **Namespace by value or reference**: `spec.namespace` accepts a literal name or a reference to a `KubernetesNamespace` resource, so an Infra Chart can create the namespace and its budgets in one run
- **Deterministic unhealthy-pod policy**: The Pulumi module always submits `unhealthyPodEvictionPolicy` explicitly with the server default applied, so the deployed object never depends on server-side defaulting
- **Dual IaC support**: Both Pulumi and Terraform implementations with feature parity (one documented exception below)
- **Lifecycle management**: Integrated with Planton's deployment lifecycle for status tracking and outputs

## Voluntary vs Involuntary Disruptions

The single most important thing to understand about PodDisruptionBudgets: **they govern only voluntary disruptions.**

- **Voluntary** disruptions go through the eviction API — `kubectl drain`, cluster upgrades, autoscaler scale-down, descheduler moves. The budget can refuse these.
- **Involuntary** failures — node crashes, kernel panics, OOM kills, pod deletion with `kubectl delete pod` — never consult the budget. A PDB is not a replacement for running enough replicas.

A budget also cannot conjure availability that does not exist: `min_available: "1"` on a single-replica workload blocks every drain touching that pod, because evicting it would leave zero available. Budgets express the floor; replica count provides the headroom above it.

## Standalone Budget vs the Workload's Built-in Block

Planton workload kinds (Deployment, StatefulSet) carry their own `availability.pod_disruption_budget` block. The boundary is:

- **Use the workload's built-in block** for a Planton Deployment's or StatefulSet's OWN pods — it derives the selector automatically from the workload's labels and cannot drift from them.
- **Use this standalone kind** for pods a Planton workload kind does not manage: operator-managed pods (a database operator's replicas), non-Planton deployments, or selector-level coverage spanning several workloads (e.g. one budget over a whole tier).

**Never point both at the same pods.** Kubernetes fails evictions for any pod covered by more than one budget, which wedges drains rather than protecting anything.

## Selecting Pods

`spec.selector` is required, and the empty selector is meaningful: a selector that is present but has no `match_labels` and no `match_expressions` protects **ALL pods in the namespace**. That shape must be declared explicitly — in `policy/v1`, a budget with NO selector matches no pods at all and is always a mistake, which is why the schema requires the field.

To protect one Planton workload's pods, match on its `app` label: every Planton workload kind stamps `app: <workload-metadata-name>` on its pods as immutable selection identity, and exports the full label set as its `selector_labels` output. Set-based `match_expressions` cover what exact matches cannot — e.g. `tier In [web, api]` for one budget spanning several workloads.

## Choosing the Bound

Set exactly one of:

- **`min_available`** — the floor: at least this many selected pods (absolute `"2"` or percentage `"50%"` of desired replicas) must remain available. Percentages round UP when computing the floor, making them stricter. `"100%"` blocks ALL voluntary evictions — including node drains — so use it only for pods that must never move.
- **`max_unavailable`** — the ceiling: at most this many selected pods may be down at once. Percentages round UP, giving evictions more room. `"0"` blocks all voluntary evictions. For workloads that scale, prefer `max_unavailable` — it tracks the replica count where an absolute `min_available` floor goes stale.

## Unhealthy Pods and Wedged Drains

`unhealthy_pod_eviction_policy` decides how the budget treats pods that are running but not ready:

- **`if_healthy_budget`** (the default): not-yet-ready pods may be evicted only while the healthy count meets the budget — the most conservative behavior.
- **`always_allow`**: not-yet-ready pods may ALWAYS be evicted. This prevents a crash-looping application from wedging node drains forever — a crash-looping pod never becomes healthy, so under the default policy its budget can block a drain indefinitely. The practical choice for budgets over workloads that can crash-loop.

> **Engine note**: `always_allow` deploys via the **Pulumi** provisioner only. The Terraform kubernetes provider cannot express `unhealthyPodEvictionPolicy`, and the Terraform module fails the plan with a precondition rather than silently deploying the default where `always_allow` was requested.

## Essential Configuration Fields

### Required

- **`spec.name`**: The PodDisruptionBudget name (DNS subdomain: lowercase alphanumeric, hyphens, dots, 1–253 chars)
- **`spec.selector`**: Which pods the budget protects; the explicit empty selector protects all pods in the namespace
- **Exactly one of `spec.min_available` / `spec.max_unavailable`**: The availability bound, absolute or percentage

### Common

- **`spec.namespace`**: Literal namespace name or reference to a KubernetesNamespace resource. A budget governs only pods in its own namespace. When omitted, the budget lands in the cluster's `default` namespace
- **`spec.unhealthy_pod_eviction_policy`**: `if_healthy_budget` (default) or `always_allow`
- **`spec.labels`** / **`spec.annotations`**: Merged with standard Planton labels for tracking and governance

## Outputs

After deployment, the following outputs are available in `status.outputs`:

- **`pod_disruption_budget_name`**: The name of the PodDisruptionBudget object as created in the cluster
- **`namespace`**: The namespace the budget was created in

A budget has no runtime handles of its own beyond identity — the eviction API enforces it by selector.

## How It Works

This kind includes both **Pulumi** (Go) and **Terraform** (HCL) modules that:

1. Resolve the target namespace (literal value or resolved reference)
2. Merge user labels and annotations with standard Planton tracking labels
3. Create the `policy/v1` PodDisruptionBudget with the selector always rendered (an empty selector block is the "all pods" wire form) and the chosen availability bound mapped as IntOrString (numeric strings as counts, `%`-suffixed as percentages)
4. Export the budget name and namespace for downstream composition

Both IaC implementations have feature parity with one deliberate exception: the Terraform kubernetes provider cannot express `unhealthyPodEvictionPolicy`, so the Terraform module rejects `always_allow` at plan time while the Pulumi module always sends the field explicitly.

## When to Use

Use **KubernetesPodDisruptionBudget** when you need:

- Availability guarantees during node drains and cluster upgrades for pods no Planton workload kind manages
- Protecting operator-managed pods (database operators, message-broker operators) whose operators do not create budgets themselves
- One budget spanning several workloads via a shared label (tier-level protection)
- Explicit control of the unhealthy-pod eviction policy for crash-loop-prone workloads

**Do NOT use** when:

- The pods belong to a Planton Deployment or StatefulSet — use the workload's built-in `availability.pod_disruption_budget` block, which derives the selector automatically
- Another budget already covers the same pods — multiple budgets on one pod make evictions fail
- You are trying to survive involuntary failures — budgets only govern the eviction API; run more replicas instead

## Prerequisites

- **Kubernetes Cluster**: Access to a Kubernetes cluster
- **Credentials**: Kubernetes cluster credentials (kubeconfig)
- **Namespace**: The target namespace must exist before creating the budget (unless deploying to `default`, or creating the namespace in the same chart via a reference)
- **Selected pods with headroom**: The budget can only refuse evictions; the workload's replica count must leave room above the declared floor for drains to proceed

## Best Practices

1. **Give every drain a way to succeed**: Ensure the bound leaves at least one evictable pod at steady state (e.g. `min_available: "1"` with 2+ replicas, or `max_unavailable: "1"`). A budget that permits zero disruptions turns every node drain into a stuck operation
2. **Prefer `max_unavailable` for scaling workloads**: it tracks replica count; an absolute `min_available` floor written for 3 replicas is far too loose at 10 and blocking at 2
3. **One budget per set of pods**: overlapping budgets fail evictions instead of combining
4. **Select workloads by the `app` label**: `match_labels: {app: <workload-name>}` composes with every Planton workload kind's label contract
5. **Set `always_allow` on crash-loop-prone workloads**: a running-but-never-ready pod under the default policy can block drains indefinitely; remember this arm requires the Pulumi provisioner

## References

- [Kubernetes Disruptions Documentation](https://kubernetes.io/docs/concepts/workloads/pods/disruptions/)
- [Specifying a Disruption Budget](https://kubernetes.io/docs/tasks/run-application/configure-pdb/)
- [PodDisruptionBudget API Reference](https://kubernetes.io/docs/reference/kubernetes-api/policy-resources/pod-disruption-budget-v1/)
- [Unhealthy Pod Eviction Policy](https://kubernetes.io/docs/tasks/run-application/configure-pdb/#unhealthy-pod-eviction-policy)

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
