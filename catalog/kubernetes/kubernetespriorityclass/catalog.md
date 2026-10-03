# Kubernetes PriorityClass

Deploys a cluster-scoped Kubernetes PriorityClass — one step of the workload importance ladder. Pods reference the class by name; the scheduler places higher-priority pods first when capacity is scarce and, unless preemption is disabled, evicts lower-priority pods to make room. The class carries the priority value, the preemption policy, and the optional cluster-wide default flag.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Kubernetes PriorityClass** -- a cluster-scoped scheduling.k8s.io/v1 PriorityClass carrying the priority value, preemption policy, optional global-default flag, and description
- **Kubernetes Labels** -- resource metadata labels (resource name, kind, organization, environment) applied automatically for tracking

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** -- an active connection in the Connect module with kubeconfig credentials for the target cluster. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline kubeconfig authentication.

### Kubernetes Cluster

- No namespace needed -- PriorityClasses are cluster-scoped; one class serves pods in every namespace.
- Check for an existing global default (`kubectl get priorityclass`) before claiming the default here -- a cluster should have exactly one.

## Deploy

### Console

Open the deployment store, find **Kubernetes PriorityClass**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Critical Services** preset for a revenue-path tier or **Preemptable Batch** for interruptible work in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesPriorityClass
metadata:
  name: critical
  org: acme-corp
  env: prod
spec:
  name: critical
  value: 1000000
  description: Revenue-path services. Use only for workloads whose downtime pages someone.
```

```shell
planton apply -f priorityclass.yaml
```

This creates a preempting class at value 1,000,000 that pods opt into via `priorityClassName: critical`. An Infra Job tracks the provisioning in real time.

## Key Configuration

These are the most important decisions when configuring a Kubernetes PriorityClass. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The value is the ladder position** -- Higher schedules (and preempts) ahead of lower; only the ORDER matters, so leave generous gaps between steps. User classes stay at or below 1,000,000,000 (the range above belongs to Kubernetes system classes); negative values make an always-preemptable tier. The value is immutable -- changing it replaces the class.

**Preemption policy** -- The default (preempt lower priority) is what critical service tiers want: pending pods evict lower-priority pods to fit. **Never Preempt** keeps the queue-jumping benefit without evicting anything running -- the right policy for high-priority batch work.

**Global default** -- Pods that name no class run at priority 0 unless one class claims the cluster-wide default. Exactly one class should claim it; with several, Kubernetes uses the smallest such value. Changing the default never re-prioritizes existing pods.

**The reserved prefix** -- Names beginning with `system-` belong to Kubernetes' built-in classes (`system-cluster-critical`, `system-node-critical`) and are rejected.

## Outputs and Dependencies

### What This Kind Consumes

This kind has no foreign-key dependencies -- it is cluster-scoped and references nothing.

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `priority_class_name` | The name of the created PriorityClass | Pod specs' `priorityClassName` (Planton workloads expose it under pod scheduling) |
| `value` | The priority integer pods of this class receive | Auditing the ladder |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**The three-step ladder** -- `critical` (1,000,000, preempting) for revenue-path services, `standard` (1,000, the global default) for everything unmarked, and `batch` (-100, never-preempt) for interruptible work. Start from the **Critical Services**, **Standard Default**, and **Preemptable Batch** presets.

## Works With

- [**Kubernetes Deployment**](/infra-catalog/kubernetes-deployment) and [**Kubernetes StatefulSet**](/infra-catalog/kubernetes-stateful-set) -- pods opt into the class via `priorityClassName` in their pod scheduling configuration
- [**Kubernetes ResourceQuota**](/infra-catalog/kubernetes-resource-quota) -- a priority-class-scoped quota budgets how much a tier may consume, so the critical tier can neither starve nor be starved
