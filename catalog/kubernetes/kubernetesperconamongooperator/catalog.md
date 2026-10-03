# Percona Operator for MongoDB

Installs the Percona Operator for MongoDB on any Kubernetes cluster from the official `psmdb-operator` Helm chart. The operator is the ENGINE: it reconciles `PerconaServerMongoDB` custom resources into highly available MongoDB deployments — replica sets with automated failover, sharded clusters, scheduled backups with point-in-time recovery via Percona Backup for MongoDB, TLS, and user management. The databases themselves are declared with [KubernetesMongodb](/infra-catalog/kubernetes-mongodb) resources, one per MongoDB cluster.

By the upstream default the operator watches **its own namespace only** — databases live beside their operator. Widen the watch to cluster-wide, or fence it to a named namespace set, from the spec.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Kubernetes Namespace** -- created only when `createNamespace` is `true`; otherwise deploys into an existing namespace
- **Helm Release** -- installs the pinned `psmdb-operator` chart, which creates:
  - Deployment running the operator controller (replicas are leader-elected warm standbys; reconcile throughput is the `maxConcurrentReconciles` dial)
  - RBAC scoped to the watch posture: namespace-scoped Roles for the own-namespace default, ClusterRole/ClusterRoleBinding for cluster-wide
  - Custom Resource Definitions for `PerconaServerMongoDB`, `PerconaServerMongoDBBackup`, and `PerconaServerMongoDBRestore` -- installed on first install, never upgraded or deleted by Helm
- **Kubernetes Labels** -- resource metadata labels (resource name, kind, organization, environment) applied automatically for tracking

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** -- an active connection in the Connect module with kubeconfig credentials for the target Kubernetes cluster. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline kubeconfig authentication.

### Kubernetes Cluster

- **A storage class** available for persistent volumes, required by the MongoDB clusters the operator will manage (not by the operator itself).

## Deploy

### Console

Open the deployment store, find **Percona Operator for MongoDB**, and click **Deploy**. The creation wizard walks you through placement, the chart pin, the watch scope (with the silently-never-reconciled trap taught inline), operator runtime, logging and telemetry, image sourcing, scheduling, and the Helm-values escape hatch. Start from the **Standard** preset in the [Presets](#presets) tab to pre-populate a working configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesPerconaMongoOperator
metadata:
  name: psmdb-operator
  org: acme-corp
  env: prod
spec:
  namespace:
    value: "mongodb"
  createNamespace: true
  chartVersion: "1.22.0"
  disableTelemetry: true
```

```shell
planton apply -f psmdb-operator.yaml
```

This installs the operator into the `mongodb` namespace with the default own-namespace watch — KubernetesMongodb resources declared in `mongodb` are reconciled; databases anywhere else need a wider watch. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the operator to a namespace managed by another Infra Component:

```yaml
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: mongodb-namespace
      fieldPath: spec.name
  createNamespace: false
```

The InfraPipeline deploys the namespace first, then installs the operator into it.

## Key Configuration

These are the most important decisions when configuring the operator. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Watch scope** -- The sharpest operational decision. Omitted, the operator watches ITS OWN namespace only (the upstream default): a KubernetesMongodb declared in a namespace the operator does not watch is **silently never reconciled** — no error anywhere, it waits forever. Set `watch.clusterWide: true` for one operator managing databases in every namespace, or `watch.namespaces` for a fenced set (the two are mutually exclusive by the spec's own validation).

**Chart pin and the CRD lifecycle** -- `chartVersion` governs both the chart and the operator (they move together for this chart). The CRDs follow a different lifecycle: installed on first install, never upgraded or deleted by Helm. A version upgrade runs new operator code against the EXISTING CRDs — apply the new release's CRDs yourself when an upgrade's release notes call for it. Uninstalling the release never cascade-deletes the database clusters.

**Runtime sizing** -- `replicas` are leader-elected warm standbys for the operator itself (control-plane availability, never throughput); `maxConcurrentReconciles` is the throughput dial for control planes managing many databases. `resources` defaults to the chart's posture: no requests or limits.

**Telemetry** -- Percona's anonymous version/feature pings to `check.percona.com` are on by the chart default; set `disableTelemetry: true` for air-gapped or compliance-bound clusters.

**Escape hatch** -- `helmValues` merges LAST over everything the typed fields render (Helm `-f` semantics, identical on both engines) — for the chart surface beyond the typed fields, never a substitute for them, and never a home for secrets.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **KubernetesNamespace** | `namespace` | `spec.name` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `namespace` | Namespace the operator runs in | Placing KubernetesMongodb databases beside their operator (the default watch posture) |
| `release_name` | Helm release name of the operator (`metadata.name`) | Identifying the installation in cluster tooling |

The operator has no per-database surface of its own: KubernetesMongodb resources compose against the CRDs it installs, not against output wiring.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Standard operator deployment** -- Own-namespace watch, pinned chart, telemetry disabled, structured logs. Databases are declared in the operator's namespace. Start from the **Standard** preset.

**Platform-wide engine** -- Set `watch.clusterWide: true` and install once (commonly in a dedicated namespace): application teams declare KubernetesMongodb resources in their own namespaces and the one operator reconciles them all.

## Works With

- [**Kubernetes Namespace**](/infra-catalog/kubernetes-namespace) -- provides the namespace for the operator deployment
- [**MongoDB**](/infra-catalog/kubernetes-mongodb) -- the databases this engine reconciles; declare them in a namespace the operator watches
