# Kubernetes MySQL

## When NOT to Use This

**The operator must already be on the cluster.** This component
declares a database cluster; KubernetesPerconaMysqlOperator installs
the ENGINE that reconciles it. The default operator posture watches its
OWN namespace — install the operator in the database's namespace, or
widen its watch. Deploy the operator first, databases after.

Also not the right kind when:

- **You want the operator itself** — installing and configuring the
  Percona Operator for MySQL is KubernetesPerconaMysqlOperator; this
  component is one XtraDB Cluster it manages.
- **You want asynchronous / read-replica MySQL semantics** — this is
  Galera SYNCHRONOUS multi-primary replication: every node holds the
  full dataset and a committed transaction is certified on every node
  before the client sees success. That buys zero-data-loss node
  failure at the cost of write latency bounded by the slowest node.
  If you want classic async primary/replica with lag-tolerant read
  scaling, this component's replication model is the wrong one.
- **You want a managed cloud database** — use the host cloud
  provider's managed-database kinds; this kind is for running
  MySQL ON the Kubernetes cluster itself.
- **You want external exposure baked in** — this component never
  creates a LoadBalancer or a route. The cluster is in-cluster
  plumbing reachable at the exported `kube_endpoint`; to reach it from
  outside, compose a first-class exposure kind against the exported
  service names. (The proxy's `expose_primary` type/annotations exist
  for the managed-cloud LoadBalancer recipes.)
- **You need a PXC surface the spec deliberately leaves out** —
  cross-cluster replication channels, PMM client integration, sidecar
  containers, Vault keyring encryption, hostPath/emptyDir storage (see
  the research doc). Those are reachable today by declaring the raw
  custom resource through KubernetesManifest.

## Overview

**KubernetesMysql** declares one production-grade MySQL cluster
reconciled by the Percona Operator for MySQL based on Percona XtraDB
Cluster. The spec renders a `pxc.percona.com/v1` PerconaXtraDBCluster
custom resource — so one resource carries the whole database story:
Galera synchronous multi-primary replication (losing a node loses no
committed data), automated full-cluster-crash recovery, HAProxy or
ProxySQL query routing, scheduled XtraBackup backups with
point-in-time recovery, TLS, and declarative users.

**Cluster size**: Galera needs a quorum — 3 nodes is the production
shape (5 for more read capacity). Fewer than 3 loses quorum-based
safety and is rejected by the operator unless `unsafe.cluster_size`
explicitly opts in (the single-node development posture).

**The naming contract**: every object the operator creates derives
from `metadata.name` — database pods (`<name>-pxc-0..N`), the proxy
Services (`<name>-haproxy` / `<name>-haproxy-replicas`, or
`<name>-proxysql`), and the system-users Secret (`<name>-secrets`,
operator-generated passwords for root and the internal accounts — key
`root` is the admin password). Applications connect through the PROXY
Service, never a database pod: the proxy routes writes to one healthy
node, detects failures, and re-routes without client changes.

**Key design points:**

- **The proxy is a oneof** — HAProxy (the default: TCP routing, writes
  on port 3306, reads load-balanced on 3307 through the replicas
  Service) or ProxySQL (SQL-aware routing with query rules and
  statement-level read/write split; stateful — it requires its own
  volume). Exactly one flavor.
- **TLS is on by default** — operator-generated certificates out of
  the box; point `tls.issuer` at a cert-manager (Cluster)Issuer for an
  organization-trusted chain. Disabling TLS entirely REQUIRES
  `unsafe.tls` — a plaintext MySQL wire is a development posture.
- **Users are declarative** — the operator creates them, keeps grants
  reconciled, and manages their password Secrets
  (`<name>-user-<username>`); rotating the Secret rotates the database
  password. Secrets never appear inline in the rendered resource.
- **Backups are XtraBackup + PITR** — named storages
  (S3/S3-compatible, Azure Blob, or a PVC; declared credentials
  materialize as `<name>-backup-<storage>` Secrets), five-field cron
  schedules with retention, and point-in-time recovery shipping
  binlogs to a DEDICATED storage (never shared with base backups).
- **The version is the image** — `image_name` chooses the MySQL
  version; changing it on a live cluster performs a SmartUpdate
  rolling upgrade (the operator orders restarts safely).
- **Exposure is composed, never embedded** — no ingress block exists
  in the spec.

## Essential Configuration Fields

### Required

- **`spec.namespace`**: namespace for the cluster — literal or a
  KubernetesNamespace reference; the operator must watch it
- **`spec.storage.size`**: volume size for every database node (one
  PVC per node; grows are applied in place — shrinks are rejected)

### Common

- **`spec.instances`**: database nodes — 3 is the production shape
  (Galera quorum), 5 adds read capacity, 1 is development-only and
  requires `unsafe.cluster_size` (default 3)
- **`spec.image_name`**: the Percona XtraDB Cluster image, tag form
  (e.g. `percona/percona-xtradb-cluster:8.4.8-8.1` — MySQL 8.4); empty
  = the module's default for the pinned operator
- **`spec.resources`**: CPU/memory for every database node pod
- **`spec.mysql_config`**: extra my.cnf appended to the operator's
  defaults — Galera/wsrep and SST essentials are operator-managed;
  override them only when you know the interaction
- **`spec.proxy`**: `haproxy` (replicas — default 3, 1 requires
  `unsafe.proxy_size`; `expose_primary` / `expose_replicas` Service
  knobs, `only_readers` to keep the write node free of reads) or
  `proxysql` (replicas, REQUIRED `storage` — ProxySQL is stateful)
- **`spec.tls`**: `enabled` (default true), `issuer` / `issuer_kind`
  (the cert-manager seam — a KubernetesClusterIssuer reference, or
  `Issuer` for a namespaced one), `sans` for external hostnames the
  database is reached at through composed exposure
- **`spec.users`**: declarative application users — `dbs`, `hosts`,
  `grants` (empty = ALL on the listed dbs), `with_grant_option`, and a
  password (empty = operator-generated into the same Secret)
- **`spec.backup`**: named `storages` (S3 with `endpoint_url` +
  `force_path_style` for MinIO-style stores, Azure Blob, or a PVC —
  PVC means no PITR and no off-cluster durability), `schedules`
  (five-field cron, `keep` retention, `delete_from_storage`), and
  `pitr` (binlog shipping every `time_between_uploads` seconds to a
  dedicated storage)
- **`spec.scheduling`**: `anti_affinity_topology_key`
  (`kubernetes.io/hostname` upstream default — one node per host;
  `topology.kubernetes.io/zone` to spread across zones), node
  selector, tolerations, priority class
- **`spec.pod_disruption_budget`**: one of `max_unavailable` /
  `min_available` — never allow more than (instances − quorum) down
- **`spec.auto_recovery`**: full-cluster-crash recovery — the operator
  finds the node with the newest data and bootstraps from it (default
  true)
- **`spec.update_strategy`**: `SmartUpdate` (default) /
  `RollingUpdate` / `OnDelete`
- **`spec.unsafe`**: explicit opt-in to postures the operator
  otherwise rejects — `cluster_size`, `tls`, `proxy_size`,
  `backup_if_unhealthy`; development only
- **`spec.pause`**: scale everything to zero, keep the volumes
- **`spec.log_collector`**: the fluent-bit sidecar shipping mysqld
  logs (default enabled)

## Outputs

| Output | Purpose |
|---|---|
| `namespace` | Namespace the cluster runs in |
| `cluster_name` | Name of the PerconaXtraDBCluster resource (equals `metadata.name`) — every derived object is prefixed with it |
| `primary_service` | The write Service applications connect to — `<name>-haproxy` (port 3306) or `<name>-proxysql` per the chosen proxy |
| `replicas_service` | The read Service (`<name>-haproxy-replicas`, port 3307) — HAProxy with the replicas Service enabled; empty otherwise |
| `kube_endpoint` | In-cluster endpoint of the write path (`<primary_service>.<namespace>.svc.cluster.local:3306`) |
| `port_forward_command` | Port-forward command for workstation access when no exposure is composed |
| `root_password_secret` | `{name, key}` of the root password — the operator-managed `<name>-secrets` Secret, key `root` |

## Composing in Infra Charts

- **`spec.namespace`** is a foreign key (default kind
  KubernetesNamespace, field path `spec.name`);
  **`storage.storage_class`** references a KubernetesStorageClass
  (`status.outputs.storage_class_name`); **`tls.issuer`** references a
  KubernetesClusterIssuer (`metadata.name`) — the cert-manager seam.
- **Applications consume the outputs**: `kube_endpoint` as the
  connection host, `root_password_secret` (or a declared user's
  `<name>-user-<username>` Secret) as env-from references — the
  credential rides the operator-managed Secret, never the manifest.
- **Exposure composes, never embeds**: a KubernetesService of type
  LoadBalancer (or a TCP route on a Gateway) targets the
  `primary_service` output; put externally reachable hostnames in
  `tls.sans` so the certificates cover them.
- **The operator is a cluster prerequisite**, not a reference: deploy
  KubernetesPerconaMysqlOperator first — in the database's namespace,
  or with its watch widened to cover it.

## Examples

### Development (single node)

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesMysql
metadata:
  name: dev-db
spec:
  namespace:
    value: dev-db
  create_namespace: true
  instances: 1
  storage:
    size: 10Gi
  unsafe:
    cluster_size: true # 1 node loses Galera quorum safety
```

### Production (3 nodes, declared user, S3 backups + PITR)

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesMysql
metadata:
  name: orders-db
spec:
  namespace:
    value: orders
  create_namespace: true
  instances: 3
  storage:
    size: 100Gi
  resources:
    requests:
      cpu: "1"
      memory: 2Gi
    limits:
      cpu: "2"
      memory: 4Gi
  users:
    - name: app
      dbs: [orders]
      hosts: ["%"]
      grants: [SELECT, INSERT, UPDATE, DELETE]
      password: <set-a-strong-password>
  backup:
    storages:
      - name: base
        s3:
          bucket: acme-mysql-backups
          region: us-west-2
          prefix: orders-db
          access_keys:
            access_key_id: AKIAEXAMPLE
            secret_access_key: <backup-user-secret-key>
      - name: binlogs # PITR gets its OWN storage
        s3:
          bucket: acme-mysql-backups
          region: us-west-2
          prefix: orders-db-pitr
          access_keys:
            access_key_id: AKIAEXAMPLE
            secret_access_key: <backup-user-secret-key>
    schedules:
      - name: nightly
        schedule: "0 2 * * *" # five-field cron
        storage_name: base
        keep: 7
    pitr:
      enabled: true
      storage_name: binlogs
  scheduling:
    anti_affinity_topology_key: topology.kubernetes.io/zone
```

### SQL-aware routing (ProxySQL instead of HAProxy)

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesMysql
metadata:
  name: reports-db
spec:
  namespace:
    value: reports
  create_namespace: true
  instances: 3
  storage:
    size: 50Gi
  proxy:
    proxysql:
      replicas: 3
      storage:
        size: 2Gi # ProxySQL is stateful
```

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
