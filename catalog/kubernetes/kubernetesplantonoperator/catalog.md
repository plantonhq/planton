# Planton Operator

Installs the Planton operator — the lifecycle manager that reconciles `PlantonPlatform` declarations into running self-hosted Planton platforms (control plane, console, identity, databases, secrets manager, in-cluster runner) — from the official `planton-operator` Helm chart (OCI, ghcr.io/plantonhq/charts). This component installs the MANAGER; the platforms themselves are declared with KubernetesPlantonPlatform resources, one per platform, all served by this one operator. One installation per cluster: the operator enforces that itself at startup, and the release name is fixed to `planton-operator`.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Kubernetes Namespace** — created only when `createNamespace` is `true` (`planton-operator` is the convention); otherwise installs into an existing namespace
- **Helm Release** (`planton-operator`) — the operator's manager Deployment, its ServiceAccount, the cluster-wide reconciliation ClusterRole/ClusterRoleBinding, and the namespaced leader-election Role/RoleBinding
- **CRDs** (`plantonplatforms.planton.ai`, `plantonidentityproviders.planton.ai`) — resources of the release, rendered from the operator's own source, so every `chartVersion` bump carries the matching schema; KEPT on uninstall by default (`crds.keepOnUninstall`) so removing the operator never cascade-deletes platform declarations

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** — an active connection in the Connect module with credentials for the target cluster (installing an operator is a cluster-admin act — the chart grants the manager its cluster-wide reconciliation RBAC). Map it as the default for your environment, or specify it explicitly when creating the Infra Component.

### Kubernetes Cluster

- **None beyond cluster-admin access** — the operator is self-contained. The platforms it deploys bring their own dependencies (declared per KubernetesPlantonPlatform).
- **No sibling operator** — the operator refuses to start beside another `planton-operator` install (its own startup guard) and names the remedy in its log. Adding more platforms to a cluster never needs a second operator.

## Deploy

### Console

Open the deployment store, find **Planton Operator**, and click **Deploy**. The creation wizard walks you through placement, the chart pin, operator runtime, image sourcing, and scheduling. Start from the **Standard** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesPlantonOperator
metadata:
  name: planton-operator
  org: acme-corp
  env: prod
spec:
  namespace:
    value: planton-operator
  createNamespace: true
```

```shell
planton apply -f planton-operator.yaml
```

This installs the manager with chart defaults (one replica, leader election on) and the two definitions the chart owns — no platform is deployed until one is declared. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to place the operator in a managed namespace:

```yaml
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: planton-operator
      fieldPath: spec.name
  createNamespace: false
```

The InfraPipeline creates the namespace first, then installs the operator into it. A KubernetesPlantonPlatform in the same chart declares its `depends_on` edge to the operator through `metadata.relationships` — no spec field consumes an operator output, so the edge is declared, not wired.

## Key Configuration

These are the most important decisions when configuring a Planton Operator install. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**One operator per cluster is self-enforced** — the operator scans for sibling operator Deployments at startup and refuses to start beside one, so the failure mode of a duplicate install is a crash-looping pod with a remedy in its log, never two managers fighting. It watches every namespace: adding platforms to the cluster means declaring more KubernetesPlantonPlatform resources, never installing a second operator.

**The chart pin moves the schema with it** — `chartVersion` defaults to the version this catalog release was validated against; pin a different published version for change control. The operator's definitions are resources of the release, so a `chartVersion` bump upgrades the operator and its schema together, and there is no second copy of the schema for anything to fall behind. Charts older than `0.8.0` do not own their definitions and are refused at plan time with the version to pin.

**`crds.install: false` is a hand-off, not a shortcut** — set it only when another release or a GitOps tool already owns the two definitions on this cluster; with them absent the operator cannot start. **`crds.keepOnUninstall: false` is the one destructive dial** — destroying the resource then deletes the definitions and, with them, every PlantonPlatform declaration and every platform behind them on the cluster. The default keeps them.

**Replicas are warm standbys** — extra replicas shorten failover of the OPERATOR itself through leader election; they add no reconciliation throughput. Leader election (chart default on) is required whenever `replicas` exceeds 1. The chart's resource defaults (requests 10m/256Mi, limits 500m/512Mi) comfortably serve several platforms — the heavy lifting happens in the platforms' own workloads, never in the operator.

**Destroy is safe by construction** — destroying this resource removes the manager and its RBAC, but the kept definitions mean platform declarations — and the running platforms behind them — survive, unmanaged, until an operator is reinstalled and adopts them. Platform deletion completes even without the operator (teardown is Kubernetes garbage collection of each declaration's owner-referenced objects), so platforms-then-operator ordering is hygiene, not a requirement.

**The Helm-values escape hatch has two forbidden knobs** — `helmValues` merges LAST over the typed fields (Helm `-f` semantics) for the chart surface beyond them (affinity, health-probe port, image pull policy). `nameOverride`/`fullnameOverride` are deliberately not honored: renaming the Deployment takes it out of the operator's own one-per-cluster startup guard's view. Never put secrets here.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **KubernetesNamespace** | `namespace` | `spec.name` |

### What This Kind Provides

This component's `status.outputs` only identify the installation — `namespace` (where the manager runs) and `release_name` (fixed to `planton-operator` by the singleton design). The operator has no per-platform surface to wire: KubernetesPlantonPlatform resources compose against the CRD it installs, and the deploy-order edge is declared through `metadata.relationships`, never through ValueFromRef.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Standard install** — the manager alone with chart defaults: one replica, leader election on, pinned chart, definitions kept on uninstall. The right first install for any cluster that will run self-hosted Planton platforms. Start from the **Standard** preset.

**HA operator** — two leader-elected replicas and raised resource headroom for clusters where several platforms ride one operator and reconcile latency after a node failure matters. The standby holds no work but takes the lease within seconds of the leader dying. Start from the **HA** preset.

## Works With

- [**Planton Platform**](/infra-catalog/kubernetes-planton-platform) — the platforms this operator reconciles, one declaration per platform; the operator is the hard prerequisite
- [**Kubernetes Namespace**](/infra-catalog/kubernetes-namespace) — provides the manager's home namespace when composed in an InfraChart
