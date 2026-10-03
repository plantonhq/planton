# MySQL

Declares a production-grade MySQL cluster reconciled by the Percona Operator for MySQL (KubernetesPerconaMysqlOperator must be on the cluster, watching this namespace). The spec renders a `pxc.percona.com/v1` PerconaXtraDBCluster custom resource: Galera SYNCHRONOUS multi-primary replication — every node holds the full dataset, a committed transaction is certified on every node, losing a node loses no data — with automated recovery, HAProxy or ProxySQL query routing, scheduled XtraBackup backups with point-in-time recovery, TLS, and declarative users. One resource per MySQL cluster; applications connect through the proxy Service, never a database pod.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **PerconaXtraDBCluster** (`pxc.percona.com/v1`, named `metadata.name`) -- the Galera cluster; the operator derives database pods (`<name>-pxc-0..N`), the proxy Services (`<name>-haproxy` / `<name>-proxysql`, plus `<name>-haproxy-replicas` for reads), and the system-users Secret (`<name>-secrets`) from it
- **Credential Secrets** -- every declared user password and backup access key materializes as a Kubernetes Secret the operator and backup jobs read; never plaintext in the rendered resource
- **PersistentVolumeClaims** -- one per database node (grow-only), plus ProxySQL's own configuration volume when that flavor is chosen
- **Namespace** (optional) -- created with standard governance labels when `createNamespace` is true

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** -- an active connection in the Connect module with credentials for the target cluster. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.

### Kubernetes Cluster

- **The Percona MySQL operator** (KubernetesPerconaMysqlOperator) on the cluster — it must WATCH the database's namespace: the default operator posture watches its OWN namespace only (install the operator there, or widen its watch). A cluster in an unwatched namespace is silently never reconciled.
- For backups: an S3/S3-compatible store (with access keys — XtraBackup authenticates with keys), Azure Blob, or a PVC. Dedicate one storage to PITR when enabled.
- For organization-trusted TLS: a cert-manager ClusterIssuer or Issuer referenced by `tls.issuer`.

## Deploy

### Console

Open the deployment store, find **MySQL**, and click **Deploy**. The creation wizard walks you through placement (and the operator-watch decision), the Galera size and MySQL version, grow-only storage, my.cnf, the HAProxy-vs-ProxySQL routing decision, TLS, declarative users, the backup arc (stores, five-field cron schedules, PITR), availability, scheduling, and the explicit unsafe opt-ins. Start from the **Production HA preset** in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesMysql
metadata:
  name: orders-db
  org: acme-corp
  env: prod
spec:
  namespace:
    value: orders
  createNamespace: true
  instances: 3
  storage:
    size: 100Gi
  users:
    - name: app
      dbs: [orders]
      hosts: ["%"]
      grants: [SELECT, INSERT, UPDATE, DELETE]
      password: $secret/orders-db-app-password
```

```shell
planton apply -f mysql.yaml
```

The operator brings up three Galera nodes and three HAProxy replicas; applications connect at the exported `kube_endpoint` as `app` (or as root, from the operator-managed `orders-db-secrets` Secret). An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire placement and storage to resources managed by other Infra Components:

```yaml
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: orders-namespace
      fieldPath: spec.name
  instances: 3
  storage:
    size: 100Gi
    storageClass:
      valueFrom:
        kind: KubernetesStorageClass
        name: fast-ssd
        fieldPath: status.outputs.storage_class_name
```

The InfraPipeline deploys the referenced resources first, then provisions the MySQL cluster against them.

## Key Configuration

These are the most important decisions when configuring MySQL. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Cluster size is quorum math** -- 3 nodes is the production shape (a majority survives any single failure or drain); 5 adds read capacity. Sizes below 3 lose quorum safety and are REJECTED by the operator unless `unsafe.clusterSize` explicitly opts in — the single-node development posture.

**The version is the image tag** -- `imageName` (e.g. `percona/percona-xtradb-cluster:8.4.8-8.1` = MySQL 8.4) is how the MySQL version is chosen; changing it on a live cluster performs a SmartUpdate rolling upgrade, replicas first, the write node last.

**Applications connect through the proxy** -- never a database pod. HAProxy (the default: TCP routing, writes on 3306, reads on 3307) or ProxySQL (SQL-aware statement-level routing with its own required stateful volume). Omitting the block means HAProxy with 3 replicas; fewer than 2 proxy replicas requires the unsafe opt-in.

**Storage is grow-only, times N** -- every Galera node holds the full dataset, so the size provisions one PVC per node. Grows apply to live PVCs; shrinks are rejected.

**Backups are a deliberate choice** -- omitted = no backups. Declare named stores (S3/S3-compatible with REQUIRED access keys, Azure Blob, or a PVC — no PITR and no off-cluster durability on a PVC), five-field cron schedules referencing stores by name, and PITR shipping binlogs to a DEDICATED store (never share it with base backups); `timeBetweenUploads` is the recovery-point objective.

**TLS is on by default** -- operator-generated certificates encrypt client and replication traffic out of the box; point `tls.issuer` at a cert-manager (Cluster)Issuer for a chain your organization trusts. Disabling TLS entirely requires `unsafe.tls`.

**Users as configuration** -- declared users are created and kept reconciled by the operator, their passwords materialized as watched Secrets (rotate the referenced org secret and the database password rotates). Empty `dbs` means server-wide grants — administrative users only.

**The unsafe block is recorded consent** -- sub-quorum sizes, disabled TLS, and single-proxy postures are operator-REJECTED unless explicitly opted in. Development conveniences; review them at every environment promotion.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **KubernetesNamespace** | `namespace` | `spec.name` |
| **KubernetesStorageClass** | `storage.storageClass` (also ProxySQL's and the PVC backup volume's) | `status.outputs.storage_class_name` |
| **KubernetesClusterIssuer** | `tls.issuer` (or a namespaced KubernetesIssuer via `issuerKind`) | `metadata.name` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `namespace` | Namespace the cluster runs in | Debugging and composition |
| `cluster_name` | The PerconaXtraDBCluster name (equals `metadata.name`) | Every operator-created object derives from it |
| `primary_service` | The WRITE Service (`<name>-haproxy`, port 3306, or `<name>-proxysql`) | Where applications send writes |
| `replicas_service` | The READ Service (`<name>-haproxy-replicas`, port 3307); empty unless the HAProxy replicas Service is enabled | Read-heavy application paths |
| `kube_endpoint` | `<primary_service>.<namespace>.svc.cluster.local:3306` | The connection string applications consume |
| `port_forward_command` | kubectl port-forward one-liner | Reaching the database from a workstation |
| `root_password_secret` | `{name, key}` of the operator-managed root credential (`<name>-secrets`, key `root`) | Mounting the root credential without copying its value |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Production HA** -- three Galera nodes with sized storage, three HAProxy replicas with the read Service, nightly XtraBackup to S3 with a dedicated PITR store, zone anti-affinity, and cert-manager TLS. Start from the **Production HA preset**.

**Development single node** -- one node with `unsafe.clusterSize`, one proxy replica with `unsafe.proxySize`: a working MySQL in a fraction of the footprint, none of the guarantees.

## Works With

- [**Percona Operator for MySQL**](/infra-catalog/kubernetes-percona-mysql-operator) -- the prerequisite engine: it must be on the cluster and watching this namespace before the database can reconcile
- [**Kubernetes Namespace**](/infra-catalog/kubernetes-namespace) -- the placement target; databases conventionally live beside their operator
- [**Kubernetes StorageClass**](/infra-catalog/kubernetes-storage-class) -- pins the volumes to an SSD-backed or provisioned-IOPS class
- [**Cert Manager Cluster Issuer**](/infra-catalog/kubernetes-cluster-issuer) -- the cert-manager trust seam for organization-verified TLS
- [**Kubernetes Deployment**](/infra-catalog/kubernetes-deployment) -- the applications that consume the exported endpoint and the user credential Secrets
