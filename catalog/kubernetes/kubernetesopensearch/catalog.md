# OpenSearch

Deploys an OpenSearch cluster — the Apache-2.0 search and analytics engine (a drop-in replacement for the Elasticsearch APIs at the 7.10 fork line, with its own 2.x/3.x feature line since). The cluster is declared as an `OpenSearchCluster` custom resource reconciled by the OpenSearch Kubernetes Operator, which manages the full lifecycle: node StatefulSets per pool, cluster bootstrap, TLS, the security plugin's admin bootstrap, safe rolling upgrades, and an optional OpenSearch Dashboards console.

The topology is yours to declare: node pools carry roles (`cluster_manager`, `data`, `ingest`, …), counts, storage, and sizing. The smallest working dev shape is one all-roles pool with 2 replicas — a single manager-eligible replica cannot survive the operator's bootstrap handoff (verified live).

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Kubernetes Namespace** — created only when `createNamespace` is `true`; otherwise deploys into an existing namespace
- **OpenSearchCluster custom resource** — the only object the module itself creates in the namespace; the operator derives everything else from it: one StatefulSet per node pool, the cluster Services, the generated TLS Secrets (a CA plus per-layer certificates for node-to-node and client traffic — the default posture), the `<name>-admin-password` bootstrap Secret, and (when enabled) the Dashboards Deployment
- **Snapshot repositories and keystore entries** — registered on the cluster at startup when declared in `snapshotRepositories` and `keystore`

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** — an active connection in the Connect module with credentials for the target cluster.
- **OpenSearch Operator** — an **OpenSearch Operator** resource must be running and watching the target namespace (its default watch is cluster-wide). Deploy it first.

### Kubernetes Cluster

- **Kernel tuning** — OpenSearch needs `vm.max_map_count` raised; the default-on privileged init container handles it. On clusters that forbid privileged init containers, tune nodes out of band and disable the dial.
- **Prometheus Operator CRDs** — only if you enable monitoring; a missing ServiceMonitor CRD fails reconciliation.

## Deploy

### Console

Open the deployment store, find **OpenSearch**, and click **Deploy**. The creation wizard walks you through namespace placement, the engine version, the node-pool topology (with the manager-quorum floor held live), engine settings, node runtime, the TLS/security posture, the optional Dashboards console, monitoring, backups, and air-gap sourcing. Start from the **Dev minimal preset** in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesOpenSearch
metadata:
  name: dev-opensearch
  org: acme-corp
  env: prod
spec:
  namespace:
    value: dev-opensearch
  createNamespace: true
  version: 2.19.4
  nodePools:
    - name: all
      replicas: 2
      roles:
        - cluster_manager
        - data
        - ingest
      jvm: -Xmx1G -Xms1G
      diskSize: 10Gi
```

```shell
planton apply -f opensearch-cluster.yaml
```

This creates a two-node all-roles cluster (the smallest shape that survives the operator's bootstrap handoff) with operator-generated TLS on both layers and a 10Gi PVC per node. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to place the cluster in a managed namespace and pin its storage to a composed StorageClass:

```yaml
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: search
      fieldPath: spec.name
  version: 2.19.4
  nodePools:
    - name: data
      replicas: 3
      roles:
        - cluster_manager
        - data
        - ingest
      persistence:
        pvc:
          storageClass:
            valueFrom:
              kind: KubernetesStorageClass
              name: fast-ssd
              fieldPath: status.outputs.storage_class_name
```

The InfraPipeline creates the namespace and StorageClass first, then provisions the cluster against them. Reference the exported `dashboards_service_name` from an Ingress or Gateway API route for team access.

## Key Configuration

These are the most important decisions when configuring an OpenSearch cluster. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The manager floor** — a lone 1-replica `cluster_manager` pool is stranded by the operator's temporary bootstrap manager (every write returns `cluster_manager_not_discovered`). Declare at least 2 manager-eligible replicas; production runs 3 dedicated managers.

**The admin password is the image's demo credential** — the operator does not generate a random admin password at this release. The bootstrapped `<name>-admin-password` Secret holds the well-known demo pair: rotate it through the security API immediately after install, or bring a custom security config, before anything real uses the cluster.

**Storage is per-pool and durable by default** — a PVC per node on the default StorageClass. emptyDir is for throwaway experiments only; heap should be about half the container memory (`jvm` sets it; the rest is OS page cache).

**Version bumps are day-2 friendly** — the operator rolls nodes one at a time with drain ordering. Check its compatibility table before a major-line jump, and keep an enabled Dashboards aligned with the engine version.

**Snapshot credentials live in the keystore** — repository settings pass verbatim into the CR (never put credentials there), and a cloud repository type needs its plugin (`repository-s3`, …) in the node plugins list.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **KubernetesNamespace** | `namespace` | `spec.name` |
| **KubernetesStorageClass** (optional, per pool) | `nodePools[].persistence.pvc.storageClass` | `status.outputs.storage_class_name` |
| **KubernetesCertificate** (bring-your-own TLS) | `security.transportTls.secret` | `status.outputs.secret_name` |
| **KubernetesCertificate** (bring-your-own TLS) | `security.httpTls.secret` | `status.outputs.secret_name` |
| **KubernetesSecret** (optional) | `security.transportTls.caSecret` | `metadata.name` |
| **KubernetesSecret** (custom security config) | `security.config.securityConfigSecret` | `metadata.name` |
| **KubernetesCertificate** (optional) | `dashboards.tls.secret` | `status.outputs.secret_name` |
| **KubernetesSecret** (custom security config) | `dashboards.opensearchCredentialsSecret` | `metadata.name` |
| **KubernetesSecret** (custom security config) | `monitoring.monitoringUserSecret` | `metadata.name` |
| **KubernetesSecret** (per entry) | `keystore[].secret` | `metadata.name` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `http_endpoint` | In-cluster https API endpoint | Application clients, log shippers, Dashboards |
| `admin_credentials_secret_name` | The bootstrapped admin Secret (username/password) — empty when a custom security config replaces the bootstrap | Client authentication |
| `service_name` | The cluster's main Service | Ingress/Gateway composition, monitoring scrape targets |
| `dashboards_service_name` | The Dashboards Service — empty when Dashboards is disabled | Ingress/Gateway exposure for the team |
| `dashboards_endpoint` | In-cluster Dashboards endpoint — empty when Dashboards is disabled | Wiring internal links and health checks |
| `port_forward_command` | Copy-paste developer access | Local exploration |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Dev minimal** — one all-roles pool with 2 replicas (the smallest shape that survives the bootstrap handoff), generated TLS on both layers. Start from the **Dev minimal preset**.

**Production cluster** — dedicated manager and data pools, sized heaps, durable storage, disruption budgets. Start from the **Production cluster preset**.

**S3 snapshots** — the production shape plus the `repository-s3` plugin, keystore-loaded credentials, and a registered S3 snapshot repository. Start from the **S3 snapshots preset**.

## Works With

- [**OpenSearch Operator**](/infra-catalog/kubernetes-open-search-operator) — the engine that reconciles this cluster; deploy it first
- [**Kubernetes Namespace**](/infra-catalog/kubernetes-namespace) — provides the namespace the cluster runs in
- [**Kubernetes StorageClass**](/infra-catalog/kubernetes-storage-class) — pins per-pool volume classes by reference
- [**Cert Manager Certificate**](/infra-catalog/kubernetes-certificate) — cert-manager-issued TLS for bring-your-own postures
- [**Kubernetes Secret**](/infra-catalog/kubernetes-secret) — keystore sources, custom security config, credentials
- [**Kubernetes Ingress**](/infra-catalog/kubernetes-ingress) — real exposure for the API and Dashboards
- [**kube-prometheus-stack**](/infra-catalog/kubernetes-kube-prometheus-stack) — the ServiceMonitor consumer when monitoring is enabled
