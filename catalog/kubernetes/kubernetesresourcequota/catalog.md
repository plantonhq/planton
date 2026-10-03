# Kubernetes ResourceQuota

Deploys a Kubernetes ResourceQuota with an optional companion LimitRange — two objects, one governance story: how much may this namespace consume in total, and what does a workload get when it doesn't say? Pairing the two is the load-bearing decision: a compute quota without container defaults makes the API reject every pod that omits requests and limits.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Kubernetes ResourceQuota** -- the aggregate caps (compute, storage, object counts) with optional scope filtering, in the specified namespace
- **Kubernetes LimitRange** (when limit defaults are set) -- the per-object defaults and bounds, sharing this resource's name
- **Kubernetes Labels** -- resource metadata labels (resource name, kind, organization, environment) applied automatically for tracking

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** -- an active connection in the Connect module with kubeconfig credentials for the target cluster. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline kubeconfig authentication.

### Kubernetes Cluster

- The target namespace must already exist (the module does not create it).
- For simple T-shirt-size governance on namespaces this platform creates, prefer the Kubernetes Namespace component's resource profiles — this kind is the full-fidelity instrument.

## Deploy

### Console

Open the deployment store, find **Kubernetes ResourceQuota**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Team Namespace Governed** preset — the safe caps-plus-defaults pairing — in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesResourceQuota
metadata:
  name: team-quota
  org: acme-corp
  env: prod
spec:
  name: team-quota
  namespace:
    value: backend-services
  hard:
    requests.cpu: "10"
    limits.memory: 40Gi
  limitDefaults:
    - type: container
      defaultRequest:
        cpu: 100m
        memory: 128Mi
      defaultLimit:
        cpu: 500m
        memory: 512Mi
```

```shell
planton apply -f resourcequota.yaml
```

This caps the namespace's compute AND gives silent containers sane defaults — the pairing that keeps naive workloads deployable. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, wire the quota to its Planton-managed namespace:

```yaml
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: backend-namespace
      fieldPath: spec.name
  hard:
    requests.cpu: "10"
    limits.memory: 40Gi
```

The InfraPipeline deploys the namespace first, then applies the quota to it.

## Key Configuration

These are the most important decisions when configuring a Kubernetes ResourceQuota. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The caps use upstream vocabulary** -- Compute (`requests.cpu`, `limits.memory`), storage (`requests.storage`, `persistentvolumeclaims`, per-class variants), and object counts (`pods`, `services.loadbalancers`, the generic `count/<resource>.<group>` form) all meter through one map.

**The rejection trap** -- Once a quota caps compute, the API REJECTS pods that omit those requests/limits. A compute quota WITHOUT container defaults breaks naive pod creation in the namespace — setting both together is the safe pattern.

**Scopes filter what is metered** -- A pod must match ALL listed scopes to count. Best Effort quotas can only meter pod counts (such pods have no compute); the priority-class scope with In/NotIn values budgets tiers separately.

**Defaults live on containers** -- Only Container limit items carry default request/limit (that is where requests live); pod and claim items take only min/max/ratio bounds. A default limit alone makes Kubernetes copy it into the request — Guaranteed-QoS containers.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **KubernetesNamespace** | `namespace` | `spec.name` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `resource_quota_name` | The name of the created ResourceQuota | `kubectl describe resourcequota` usage-vs-caps checks |
| `namespace` | The governed namespace | Verifying scope |
| `limit_range_name` | The companion LimitRange's name (when defaults are set) | Auditing the defaults arm |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**The governed team namespace** -- Compute caps + container defaults, the livable pairing. Start from the **Team Namespace Governed** preset.

**Object-count discipline** -- Cap load balancers, services, and runaway job counts without touching compute. Start from the **Object Count Caps** preset.

**Best-effort containment** -- At most N pods with no requests/limits, ever. Start from the **BestEffort Guard** preset.

## Works With

- [**Kubernetes Namespace**](/infra-catalog/kubernetes-namespace) -- reference the namespace so infra charts create it and this quota in dependency order; prefer its built-in resource profiles for simple T-shirt sizing.
- [**Kubernetes PriorityClass**](/infra-catalog/kubernetes-priority-class) -- a priority-class-scoped quota budgets one tier's consumption.
