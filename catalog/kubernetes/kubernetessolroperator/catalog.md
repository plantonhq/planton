# Apache Solr Operator

Installs the Apache Solr Operator — the Apache Solr project's own operator for running SolrCloud on Kubernetes — from the official `solr-operator` Helm chart. The operator reconciles `SolrCloud` custom resources (declared with **Apache Solr**) into running Solr clusters with managed rolling updates, scaling with replica movement, and backup repositories.

This component installs and configures the **engine**. Solr clusters themselves are declared with **Apache Solr** resources — one per cluster — which this operator reconciles.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Helm release (solr-operator)** — the operator Deployment (with leader election between replicas), its RBAC, and ServiceAccount
- **The operator CRDs** — SolrCloud, SolrBackup, SolrPrometheusExporter, and ZookeeperCluster (for the bundled dependency), derived from the pinned chart and applied by the module itself outside the release, keyed by CRD name. Because the module owns the CRD lifecycle, a chart bump moves the CRDs with it, uninstalling the operator never cascade-deletes SolrCloud resources and their data (unless `crds.keepOnUninstall` is false), a reinstall re-adopts them, and a version below what the cluster's CRDs carry is refused before anything changes
- **The bundled zookeeper-operator** (default on) — the chart dependency that provisions PROVIDED ZooKeeper ensembles, so an **Apache Solr** resource works out of the box. Disable it only when a zookeeper-operator already runs in the cluster (its fixed-name cluster-scoped RBAC conflicts on a second install) or when every Solr cluster connects to an EXTERNAL ensemble
- **Kubernetes Namespace** — created only when `createNamespace` is `true`; otherwise the operator installs into an existing namespace

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** — an active connection in the Connect module with credentials for the target cluster.

### Kubernetes Cluster

- **No existing zookeeper-operator** — SolrCloud requires ZooKeeper, and with the defaults you need nothing: the bundled zookeeper-operator provisions provided ensembles. Check first whether the cluster already runs a zookeeper-operator — one per cluster, ever (its fixed-name cluster-scoped RBAC conflicts on a second install); if one exists, set `zookeeperOperator.install: false` with `useExisting: true`.

## Deploy

### Console

Open the deployment store, find **Apache Solr Operator**, and click **Deploy**. The creation wizard walks you through namespace placement, the chart pin, watch scope, the bundled zookeeper-operator, operator runtime, the mTLS client identity, resources and scheduling, image sourcing, and the Helm-values escape hatch. Start from the **Standard preset** in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesSolrOperator
metadata:
  name: solr-operator
  org: acme-corp
  env: dev
spec:
  namespace:
    value: solr-operator
  createNamespace: true
```

```shell
planton apply -f solr-operator.yaml
```

The defaults are the point: chart 0.9.1, the bundled zookeeper-operator installed, and a cluster-wide watch — every **Apache Solr** resource in any namespace gets reconciled, including its provided ZooKeeper ensemble. An Infra Job tracks the provisioning in real time.

### InfraChart

Compose the operator with a managed namespace:

```yaml
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: search-platform
      fieldPath: spec.name
  createNamespace: false
```

The InfraPipeline deploys the namespace first, then installs the operator into it.

## Key Configuration

These are the most important decisions when configuring an Apache Solr Operator install. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Chart pinning** — `chartVersion` defaults to **0.9.1**. Chart and operator versions move together (the operator image tag is `v<chart_version>`), and the chart version has **no `v` prefix** while the operator/CRD artifacts carry one. The operator is **pre-1.0**: minor versions can change the CRD API, so a version bump means applying the matching CRDs — which the deploy does, because the CRDs are derived from the pinned chart. Versions must exist as SERVED charts in the repository index; a version that is not published is refused with the remedy before anything is created.

**The bundled zookeeper-operator** — `zookeeperOperator.install` defaults true: the path that makes provided ZooKeeper ensembles work out of the box. Set `install: false` together with `useExisting: true` when the cluster already runs a zookeeper-operator (fixed-name cluster-scoped RBAC conflicts on a second install); plain `install: false` alone is the external-ensemble-only posture — a provided ensemble would then hang silently waiting on ZooKeeper.

**Watch scope** — by default the operator watches ALL namespaces: one install serves every **Apache Solr** resource on the cluster. Set `watchNamespaces` to fence it to an explicit list — the multi-tenant posture where several fenced installs share a cluster. The fence is silent on the outside (a Solr cluster declared beyond it is never reconciled, with no event pointing at the fence), and the watched namespaces need only exist by the time SolrCloud resources appear in them.

**mTLS client identity** — configure `mtls` only when **Apache Solr** resources enforce `clientAuth` on their TLS listeners; the operator then presents `mtls.clientCertSecret` (required within the block) when calling Solr. `insecureSkipVerify` defaults true because the operator dials pods by pod IP — addresses that rarely appear in certificate SANs; the CA trust still authenticates both ends. All the Secret fields hold NAMES, never certificate material.

**Operator runtime** — `replicas` (default 1) adds leader-elected warm standbys, never throughput; `leaderElectionEnabled` (default true) should be disabled only for a single-replica dev install; `metricsEnabled` (default true) exposes the operator's Prometheus endpoint. `resources` is empty by default — the chart sets none; the operator is lightweight.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **KubernetesNamespace** | `namespace` | `spec.name` |
| **KubernetesSecret** | `mtls.clientCertSecret` | `metadata.name` |
| **KubernetesSecret** | `mtls.caCertSecret` | `metadata.name` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `namespace` | The installation namespace | Composition, debugging |
| `release_name` | Helm release name (= metadata.name) | Helm-level operations |
| `deployment_name` | The Solr operator Deployment name | Monitoring, log collection |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Standard** — one cluster-wide operator plus the bundled zookeeper-operator on chart 0.9.1; the first thing to change is nothing. Start from the **Standard preset**.

**Existing zookeeper-operator** — the second-install posture: `install: false` + `useExisting: true` when another release already runs the dependency. Start from the **Existing zookeeper operator preset**.

**Namespace fenced** — an explicit `watchNamespaces` list, a warm-standby replica, and declared resources: the multi-tenant / governed-platform posture. Start from the **Namespace fenced preset**.

## Works With

- [**Apache Solr**](/infra-catalog/kubernetes-solr) — the SolrCloud clusters this operator reconciles; deploy the operator FIRST (it is the registered prerequisite).
- [**Kubernetes Namespace**](/infra-catalog/kubernetes-namespace) — reference a managed namespace to compose governance (quotas, pod-security labels) with the installation.
- [**Kubernetes Secret**](/infra-catalog/kubernetes-secret) — the mTLS client and CA certificate Secrets the operator presents when Solr clusters enforce clientAuth.
