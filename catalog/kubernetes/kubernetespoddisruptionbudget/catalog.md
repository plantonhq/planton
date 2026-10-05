# Kubernetes PodDisruptionBudget

Deploys a Kubernetes PodDisruptionBudget — insurance against maintenance. The budget limits how many of a set of pods may be taken down VOLUNTARILY at once (node drains, cluster upgrades, autoscaler consolidation); the eviction API refuses to breach it. The budget carries a pod selector, exactly one availability bound, and the unhealthy-pod eviction policy.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Kubernetes PodDisruptionBudget** -- a policy/v1 budget in the specified namespace carrying the pod selector, exactly one availability bound (min available OR max unavailable), and the unhealthy-pod eviction policy
- **Kubernetes Labels** -- resource metadata labels (resource name, kind, organization, environment) applied automatically for tracking

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** -- an active connection in the Connect module with kubeconfig credentials for the target cluster. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline kubeconfig authentication.

### Kubernetes Cluster

- The target namespace must already exist (the module does not create it).
- For a Planton Deployment or StatefulSet's OWN pods, prefer the workload's built-in disruption-budget block — it derives the selector automatically. This standalone kind is for pods a Planton workload does not manage.

## Deploy

### Console

Open the deployment store, find **Kubernetes PodDisruptionBudget**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Protect Workload** preset — the one-at-a-time rolling-maintenance contract — in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesPodDisruptionBudget
metadata:
  name: checkout-pdb
  org: acme-corp
  env: prod
spec:
  name: checkout-pdb
  namespace:
    value: backend-services
  selector:
    matchLabels:
      app: checkout
  maxUnavailable: "1"
```

```shell
planton apply -f pdb.yaml
```

This lets cluster maintenance move at most ONE `checkout` pod at a time — drains negotiate, they never surprise. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to place the budget in a namespace managed alongside it:

```yaml
spec:
  name: checkout-pdb
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: backend-services
      fieldPath: spec.name
  selector:
    matchLabels:
      app: checkout
  maxUnavailable: "1"
```

The InfraPipeline creates the namespace first, then the budget inside it — the selector matches the workload's pods at runtime, so no ordering edge to the workload is needed.

## Key Configuration

These are the most important decisions when configuring a Kubernetes PodDisruptionBudget. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Selection by the label contract** -- Every Planton workload stamps the `app` label (set to its name) on its pods, so `matchLabels: {app: checkout}` protects exactly that workload. An EMPTY selector protects ALL pods in the namespace — legal, explicit, and rarely what you want.

**Exactly one bound** -- `maxUnavailable` (preferred for workloads that scale — it tracks replica count) or `minAvailable` (an absolute floor that goes stale as replicas change). The API rejects both together, and a budget with neither protects nothing.

**The unsatisfiable-budget trap** -- `maxUnavailable: "0"` or `minAvailable: "100%"` blocks ALL voluntary evictions including node drains; a single-replica workload behind such a budget wedges every cluster upgrade until someone relaxes it.

**The crash-loop wedge** -- A budget counts HEALTHY (Ready) pods. Under the conservative default, a crash-looping app can block drains forever; **Always Allow** lets not-ready pods be evicted regardless — the practical choice for workloads that can crash-loop.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **KubernetesNamespace** (optional) | `namespace` | `spec.name` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `pod_disruption_budget_name` | The name of the created budget | `kubectl get pdb` allowed-disruptions checks |
| `namespace` | The namespace the budget protects | Verifying scope |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Rolling-maintenance contract** -- One workload, one pod movable at a time. Start from the **Protect Workload** preset.

**Percentage tier** -- At most 25% of a scaled fleet down at once. Start from the **Tier Percentage** preset.

**Crash-loop tolerant** -- Always Allow eviction of not-ready pods so a broken app never wedges upgrades. Start from the **Crashloop Tolerant** preset.

## Works With

- [**Kubernetes Namespace**](/infra-catalog/kubernetes-namespace) -- reference the namespace so Infra Charts create it and this budget in dependency order
- [**Kubernetes Deployment**](/infra-catalog/kubernetes-deployment) -- its built-in disruption-budget block covers its own pods; this standalone kind covers operator-managed replicas and everything else, matched via their `app` label
