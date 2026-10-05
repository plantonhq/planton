# Percona Operator for MySQL

Installs the Percona Operator for MySQL — based on Percona XtraDB Cluster — on any Kubernetes cluster from the official `pxc-operator` Helm chart. The operator reconciles `PerconaXtraDBCluster` custom resources into highly available MySQL clusters: Galera synchronous multi-primary replication with automated failover, HAProxy or ProxySQL query routing, scheduled XtraBackup backups with point-in-time recovery, and TLS. This component installs and configures the ENGINE; the databases themselves are declared with `KubernetesMysql` Infra Components — one per MySQL cluster — which this operator reconciles.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Kubernetes Namespace** -- created only when `createNamespace` is `true`; otherwise deploys into an existing namespace
- **Helm Release** -- installs the `pxc-operator` chart, which creates:
  - Deployment running the operator controller (replicas beyond one are leader-elected warm standbys)
  - RBAC scoped to the watch posture: namespace-scoped Roles by default, ClusterRoles for a cluster-wide operator
  - Custom Resource Definitions for `PerconaXtraDBCluster`, `PerconaXtraDBClusterBackup`, and `PerconaXtraDBClusterRestore` -- installed on first install, never upgraded or deleted by Helm
- **Kubernetes Labels** -- resource metadata labels (resource name, kind, organization, environment) applied automatically for tracking

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** -- an active connection in the Connect module with kubeconfig credentials for the target Kubernetes cluster. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline kubeconfig authentication.

### Kubernetes Cluster

- **Kubernetes 1.21+** with support for Custom Resource Definitions and Helm-based deployments.
- **A storage class** available for persistent volumes, required by the MySQL clusters the operator will manage.

## Deploy

### Console

Open the deployment store, find **Percona Operator for MySQL**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and the spec fields — including the watch-scope decision that determines which namespaces' databases this operator serves. Start from the **Standard** preset in the [Presets](#presets) tab to pre-populate a working configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesPerconaMysqlOperator
metadata:
  name: pxc-operator
  org: acme-corp
  env: prod
spec:
  namespace:
    value: "mysql-prod"
  createNamespace: true
  chartVersion: "1.20.0"
  maxConcurrentReconciles: 5
  disableTelemetry: true
```

```shell
planton apply -f percona-mysql-operator.yaml
```

This installs the operator into the `mysql-prod` namespace, watching that namespace only (the upstream default) — `KubernetesMysql` resources declared there are reconciled immediately. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the operator to a namespace managed by another Infra Component:

```yaml
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: mysql-namespace
      fieldPath: spec.name
  createNamespace: false
```

The InfraPipeline deploys the namespace first, then installs the operator into it.

## Key Configuration

These are the most important decisions when configuring the operator. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Watch scope** -- By default the operator watches ITS OWN namespace only (the upstream posture: databases live beside their operator). Set `watch.clusterWide: true` for one operator managing databases in every namespace, or `watch.namespaces` for a fenced set — the two are mutually exclusive. A `KubernetesMysql` declared in a namespace the operator does not watch is silently never reconciled: it stays Pending forever with no error anywhere.

**Chart version** -- Chart and operator versions move together for this chart; the chart pin governs. Pin deliberately: editing the pin re-runs the release with the new chart, which IS the upgrade. Helm never upgrades the CRDs it installed — apply a new release's CRDs yourself when its release notes call for it.

**Reconcile throughput** -- `maxConcurrentReconciles` (chart default 1) is the throughput dial: how many database clusters reconcile in parallel. `replicas` adds leader-elected warm standbys for the operator itself — faster operator failover, zero extra throughput.

**Backup engine** -- `s3WorkersLimit` (chart default 10) sizes the S3 upload/delete worker pool shared by EVERY database's backups; `xtrabackupSidecar` runs XtraBackup inside the database pods instead of separate jobs.

**Leader election** -- On by default (`leaderElection.enabled` unset means enabled). The lease/renew/retry durations (defaults 60s/40s/10s) trade operator-failover speed against API-server traffic.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **KubernetesNamespace** | `namespace` | `spec.name` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `namespace` | Namespace the operator runs in | Placing `KubernetesMysql` databases beside their operator (the default watch scope) |
| `release_name` | Helm release name of the operator | Identifying the installation in cluster tooling |

The operator has no per-database surface of its own: each MySQL cluster is a `KubernetesMysql` resource composing against the CRDs this installation provides.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Operator beside its databases** -- The upstream default: install the operator into the namespace where its `KubernetesMysql` clusters will live, keeping the own-namespace watch scope. Strongest per-team isolation. Start from the **Standard** preset.

**One cluster-wide engine** -- A single operator in a platform namespace with `watch.clusterWide: true`, serving every team's databases. Raise `maxConcurrentReconciles` for fleets.

## Works With

- [**Kubernetes Namespace**](/infra-catalog/kubernetes-namespace) -- provides the namespace for the operator deployment
- [**MySQL**](/infra-catalog/kubernetes-mysql) -- the databases this operator reconciles, one Infra Component per MySQL cluster
