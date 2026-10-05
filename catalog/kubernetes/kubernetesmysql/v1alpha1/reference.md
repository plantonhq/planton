# KubernetesMysql

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `kubernetes.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

**KubernetesMysqlSpec** declares a production-grade MySQL cluster
reconciled by the Percona Operator for MySQL based on Percona XtraDB
Cluster (KubernetesPerconaMysqlOperator must be on the cluster). The
spec renders a `pxc.percona.com/v1` PerconaXtraDBCluster custom
resource: Galera SYNCHRONOUS multi-primary replication (every node
holds the full dataset; a committed transaction is certified on every
node — losing a node loses no data), automated recovery, HAProxy or
ProxySQL query routing, scheduled XtraBackup backups with
point-in-time recovery, TLS, and declarative users.

CLUSTER SIZE: Galera needs a quorum — 3 nodes is the production
shape (5 for more read capacity). Sizes below 3 lose quorum-based
safety and are rejected by the operator unless the `unsafe` block
explicitly opts in (the single-node development posture).

APPLICATIONS CONNECT THROUGH THE PROXY, never a database pod: the
proxy layer (HAProxy by default) routes writes to one healthy node,
detects failures, and re-routes without client changes. The chart of
services: `<name>-haproxy` (writes, port 3306; reads via 3307) — or
`<name>-proxysql` when ProxySQL is chosen.

NAMING CONTRACT: every object derives from `metadata.name` — pods
(`<name>-pxc-0..N`), Services (`<name>-pxc`, `<name>-haproxy` /
`<name>-proxysql`), and the system-users Secret (`<name>-secrets`,
operator-generated passwords for root and the internal accounts).

EXPOSURE IS COMPOSED, never embedded: the cluster is in-cluster
plumbing reachable at the exported `kube_endpoint`. To reach it from
outside, compose a first-class exposure kind — the proxy's
`expose_primary` service type/annotations exist for the managed-cloud
LoadBalancer recipes, documented per environment.

DELIBERATELY NOT MODELED (reachable via the operator's own upstream
surface where noted, recorded in the research doc):
cross-cluster replication channels, PMM client integration, sidecar
containers, Vault keyring encryption, hostPath/emptyDir storage.

## Example

```yaml
# Full-surface manifest for offline module proofs (tofu validate/plan and
# pulumi preview). Exercises every typed block coherently: a three-node
# Galera cluster with HAProxy, tuned my.cnf, declarative users with
# declared passwords, S3-compatible (MinIO-style) backups with declared
# access keys and a schedule, PITR, cert-manager TLS alt names, hard
# anti-affinity with tolerations, log-collector resources, and image pull
# secrets. Both engines must render identical CRs + credential Secrets
# from it.
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesMysql
metadata:
  name: test-mysql
spec:
  namespace:
    value: test-mysql-ns
  create_namespace: true
  instances: 3
  image_name: percona/percona-xtradb-cluster:8.4.8-8.1
  storage:
    size: 20Gi
    storage_class:
      value: fast-ssd
  resources:
    limits:
      cpu: 2000m
      memory: 4Gi
    requests:
      cpu: 500m
      memory: 1Gi
  mysql_config: |
    [mysqld]
    max_connections=200
  auto_recovery: true
  proxy:
    haproxy:
      replicas: 3
      resources:
        requests:
          cpu: 100m
          memory: 256Mi
        limits:
          cpu: 500m
          memory: 512Mi
      expose_replicas:
        enabled: true
        only_readers: true
  tls:
    sans:
      - mysql.example.com
  users:
    - name: appuser
      dbs:
        - appdb
      hosts:
        - "%"
      grants:
        - SELECT
        - INSERT
        - UPDATE
        - DELETE
      password: initial-app-password
    - name: readonly
      dbs:
        - appdb
      hosts:
        - "%"
      grants:
        - SELECT
      password: readonly-password
  backup:
    storages:
      - name: minio
        verify_tls: false
        s3:
          bucket: mysql-backups
          region: minio
          prefix: test-mysql
          endpoint_url: http://minio.minio-system.svc:9000
          force_path_style: true
          access_keys:
            access_key_id: minio-access-key
            secret_access_key: minio-secret-key
      - name: pitr-store
        verify_tls: false
        s3:
          bucket: mysql-pitr
          region: minio
          endpoint_url: http://minio.minio-system.svc:9000
          force_path_style: true
          access_keys:
            access_key_id: minio-access-key
            secret_access_key: minio-secret-key
    schedules:
      - name: nightly
        schedule: "0 2 * * *"
        storage_name: minio
        keep: 7
        delete_from_storage: true
    pitr:
      enabled: true
      storage_name: pitr-store
      time_between_uploads: 60
  scheduling:
    anti_affinity_topology_key: topology.kubernetes.io/zone
    node_selector:
      workload: database
    tolerations:
      - key: dedicated
        operator: Equal
        value: database
        effect: NoSchedule
    priority_class_name: database-critical
  pod_disruption_budget:
    max_unavailable: 1
  log_collector:
    enabled: true
    resources:
      requests:
        cpu: 50m
        memory: 64Mi
  update_strategy: SmartUpdate
  image_pull_secrets:
    - ghcr-pull-secret
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.namespace` | `string \| valueFrom` | yes |  | KubernetesNamespace (`spec.name`) |
| `spec.createNamespace` | `bool` |  |  |  |
| `spec.imageName` | `string` |  |  |  |
| `spec.instances` | `int32` |  | `3` |  |
| `spec.storage` | `KubernetesMysqlStorage` | yes |  |  |
| `spec.storage.size` | `string` | yes |  |  |
| `spec.storage.storageClass` | `string \| valueFrom` |  |  | KubernetesStorageClass (`status.outputs.storage_class_name`) |
| `spec.resources` | `ContainerResources` |  |  |  |
| `spec.resources.limits` | `CpuMemory` |  |  |  |
| `spec.resources.limits.cpu` | `string` |  |  |  |
| `spec.resources.limits.memory` | `string` |  |  |  |
| `spec.resources.requests` | `CpuMemory` |  |  |  |
| `spec.resources.requests.cpu` | `string` |  |  |  |
| `spec.resources.requests.memory` | `string` |  |  |  |
| `spec.mysqlConfig` | `string` |  |  |  |
| `spec.autoRecovery` | `bool` |  | `true` |  |
| `spec.proxy` | `KubernetesMysqlProxy` |  |  |  |
| `spec.proxy.haproxy` | `KubernetesMysqlHaproxy` |  |  |  |
| `spec.proxy.haproxy.replicas` | `int32` |  | `3` |  |
| `spec.proxy.haproxy.resources` | `ContainerResources` |  |  |  |
| `spec.proxy.haproxy.resources.limits` | `CpuMemory` |  |  |  |
| `spec.proxy.haproxy.resources.limits.cpu` | `string` |  |  |  |
| `spec.proxy.haproxy.resources.limits.memory` | `string` |  |  |  |
| `spec.proxy.haproxy.resources.requests` | `CpuMemory` |  |  |  |
| `spec.proxy.haproxy.resources.requests.cpu` | `string` |  |  |  |
| `spec.proxy.haproxy.resources.requests.memory` | `string` |  |  |  |
| `spec.proxy.haproxy.config` | `string` |  |  |  |
| `spec.proxy.haproxy.exposePrimary` | `KubernetesMysqlProxyService` |  |  |  |
| `spec.proxy.haproxy.exposePrimary.type` | `string` |  | `ClusterIP` |  |
| `spec.proxy.haproxy.exposePrimary.annotations` | `map<string, string>` |  |  |  |
| `spec.proxy.haproxy.exposeReplicas` | `KubernetesMysqlHaproxyReplicasService` |  |  |  |
| `spec.proxy.haproxy.exposeReplicas.enabled` | `bool` |  | `true` |  |
| `spec.proxy.haproxy.exposeReplicas.onlyReaders` | `bool` |  |  |  |
| `spec.proxy.haproxy.exposeReplicas.type` | `string` |  | `ClusterIP` |  |
| `spec.proxy.haproxy.exposeReplicas.annotations` | `map<string, string>` |  |  |  |
| `spec.proxy.proxysql` | `KubernetesMysqlProxysql` |  |  |  |
| `spec.proxy.proxysql.replicas` | `int32` |  | `3` |  |
| `spec.proxy.proxysql.resources` | `ContainerResources` |  |  |  |
| `spec.proxy.proxysql.resources.limits` | `CpuMemory` |  |  |  |
| `spec.proxy.proxysql.resources.limits.cpu` | `string` |  |  |  |
| `spec.proxy.proxysql.resources.limits.memory` | `string` |  |  |  |
| `spec.proxy.proxysql.resources.requests` | `CpuMemory` |  |  |  |
| `spec.proxy.proxysql.resources.requests.cpu` | `string` |  |  |  |
| `spec.proxy.proxysql.resources.requests.memory` | `string` |  |  |  |
| `spec.proxy.proxysql.storage` | `KubernetesMysqlStorage` | yes |  |  |
| `spec.proxy.proxysql.storage.size` | `string` | yes |  |  |
| `spec.proxy.proxysql.storage.storageClass` | `string \| valueFrom` |  |  | KubernetesStorageClass (`status.outputs.storage_class_name`) |
| `spec.proxy.proxysql.config` | `string` |  |  |  |
| `spec.proxy.proxysql.exposePrimary` | `KubernetesMysqlProxyService` |  |  |  |
| `spec.proxy.proxysql.exposePrimary.type` | `string` |  | `ClusterIP` |  |
| `spec.proxy.proxysql.exposePrimary.annotations` | `map<string, string>` |  |  |  |
| `spec.tls` | `KubernetesMysqlTls` |  |  |  |
| `spec.tls.enabled` | `bool` |  | `true` |  |
| `spec.tls.issuer` | `string \| valueFrom` |  |  | KubernetesClusterIssuer (`metadata.name`) |
| `spec.tls.issuerKind` | `string` |  | `ClusterIssuer` |  |
| `spec.tls.sans` | `[]string` |  |  |  |
| `spec.users` | `[]KubernetesMysqlUser` |  |  |  |
| `spec.users[].name` | `string` | yes |  |  |
| `spec.users[].dbs` | `[]string` |  |  |  |
| `spec.users[].hosts` | `[]string` |  |  |  |
| `spec.users[].grants` | `[]string` |  |  |  |
| `spec.users[].withGrantOption` | `bool` |  |  |  |
| `spec.users[].password` | `string` (sensitive) |  |  |  |
| `spec.backup` | `KubernetesMysqlBackup` |  |  |  |
| `spec.backup.storages` | `[]KubernetesMysqlBackupStorage` | yes |  |  |
| `spec.backup.storages[].name` | `string` | yes |  |  |
| `spec.backup.storages[].s3` | `KubernetesMysqlS3Storage` |  |  |  |
| `spec.backup.storages[].s3.bucket` | `string` | yes |  |  |
| `spec.backup.storages[].s3.region` | `string` |  |  |  |
| `spec.backup.storages[].s3.prefix` | `string` |  |  |  |
| `spec.backup.storages[].s3.endpointUrl` | `string` |  |  |  |
| `spec.backup.storages[].s3.forcePathStyle` | `bool` |  |  |  |
| `spec.backup.storages[].s3.accessKeys` | `KubernetesMysqlS3AccessKeys` | yes |  |  |
| `spec.backup.storages[].s3.accessKeys.accessKeyId` | `string` | yes |  |  |
| `spec.backup.storages[].s3.accessKeys.secretAccessKey` | `string` (sensitive) | yes |  |  |
| `spec.backup.storages[].azure` | `KubernetesMysqlAzureStorage` |  |  |  |
| `spec.backup.storages[].azure.container` | `string` | yes |  |  |
| `spec.backup.storages[].azure.prefix` | `string` |  |  |  |
| `spec.backup.storages[].azure.endpointUrl` | `string` |  |  |  |
| `spec.backup.storages[].azure.storageAccount` | `string` | yes |  |  |
| `spec.backup.storages[].azure.accessKey` | `string` (sensitive) | yes |  |  |
| `spec.backup.storages[].pvc` | `KubernetesMysqlPvcStorage` |  |  |  |
| `spec.backup.storages[].pvc.volume` | `KubernetesMysqlStorage` | yes |  |  |
| `spec.backup.storages[].pvc.volume.size` | `string` | yes |  |  |
| `spec.backup.storages[].pvc.volume.storageClass` | `string \| valueFrom` |  |  | KubernetesStorageClass (`status.outputs.storage_class_name`) |
| `spec.backup.storages[].verifyTls` | `bool` |  | `true` |  |
| `spec.backup.schedules` | `[]KubernetesMysqlBackupSchedule` |  |  |  |
| `spec.backup.schedules[].name` | `string` | yes |  |  |
| `spec.backup.schedules[].schedule` | `string` | yes |  |  |
| `spec.backup.schedules[].storageName` | `string` | yes |  |  |
| `spec.backup.schedules[].keep` | `int32` |  |  |  |
| `spec.backup.schedules[].deleteFromStorage` | `bool` |  | `true` |  |
| `spec.backup.pitr` | `KubernetesMysqlPitr` |  |  |  |
| `spec.backup.pitr.enabled` | `bool` |  |  |  |
| `spec.backup.pitr.storageName` | `string` |  |  |  |
| `spec.backup.pitr.timeBetweenUploads` | `int32` |  | `60` |  |
| `spec.scheduling` | `KubernetesMysqlScheduling` |  |  |  |
| `spec.scheduling.antiAffinityTopologyKey` | `string` |  |  |  |
| `spec.scheduling.nodeSelector` | `map<string, string>` |  |  |  |
| `spec.scheduling.tolerations` | `[]WorkloadToleration` |  |  |  |
| `spec.scheduling.tolerations[].key` | `string` |  |  |  |
| `spec.scheduling.tolerations[].operator` | `string` |  |  |  |
| `spec.scheduling.tolerations[].value` | `string` |  |  |  |
| `spec.scheduling.tolerations[].effect` | `string` |  |  |  |
| `spec.scheduling.tolerations[].tolerationSeconds` | `int64` |  |  |  |
| `spec.scheduling.priorityClassName` | `string` |  |  |  |
| `spec.podDisruptionBudget` | `KubernetesMysqlPodDisruptionBudget` |  |  |  |
| `spec.podDisruptionBudget.maxUnavailable` | `int32` |  |  |  |
| `spec.podDisruptionBudget.minAvailable` | `int32` |  |  |  |
| `spec.logCollector` | `KubernetesMysqlLogCollector` |  |  |  |
| `spec.logCollector.enabled` | `bool` |  | `true` |  |
| `spec.logCollector.resources` | `ContainerResources` |  |  |  |
| `spec.logCollector.resources.limits` | `CpuMemory` |  |  |  |
| `spec.logCollector.resources.limits.cpu` | `string` |  |  |  |
| `spec.logCollector.resources.limits.memory` | `string` |  |  |  |
| `spec.logCollector.resources.requests` | `CpuMemory` |  |  |  |
| `spec.logCollector.resources.requests.cpu` | `string` |  |  |  |
| `spec.logCollector.resources.requests.memory` | `string` |  |  |  |
| `spec.updateStrategy` | `string` |  | `SmartUpdate` |  |
| `spec.unsafe` | `KubernetesMysqlUnsafe` |  |  |  |
| `spec.unsafe.clusterSize` | `bool` |  |  |  |
| `spec.unsafe.tls` | `bool` |  |  |  |
| `spec.unsafe.proxySize` | `bool` |  |  |  |
| `spec.unsafe.backupIfUnhealthy` | `bool` |  |  |  |
| `spec.pause` | `bool` |  |  |  |
| `spec.imagePullSecrets` | `[]string` |  |  |  |

## Field Details

### spec.namespace

`string | valueFrom` · required

Namespace to create the MySQL cluster in. The reconciling operator
must watch this namespace (the default operator posture watches its
OWN namespace — install the operator there, or widen its watch).
Accepts a literal namespace name or a reference to a
KubernetesNamespace resource.

- references: KubernetesNamespace (`spec.name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesNamespace, name: <that resource's name>, fieldPath: spec.name}} -- a bare string does not parse

### spec.createNamespace

`bool`

When true, the namespace is created (with the standard Planton
governance labels) before the cluster and deleted with the
resource. When false, the namespace must already exist.

### spec.imageName

`string`

Percona XtraDB Cluster container image, tag form (e.g.
"percona/percona-xtradb-cluster:8.4.8-8.1" — MySQL 8.4). Empty =
the module's default for the pinned operator. This is how the MySQL
version is chosen; changing it on a live cluster performs a
SmartUpdate rolling upgrade.

### spec.instances

`int32` · optional (explicit presence)

Number of database nodes. 3 is the production shape (Galera
quorum); 5 adds read capacity. 1 is a development posture that
REQUIRES unsafe.cluster_size.

- default: `3`
- rule: {"int32":{"gte":1}}

### spec.storage

`KubernetesMysqlStorage` · required

Storage for every database node. Required: the operator provisions
one PVC per node and grows are applied in place (never shrinks).

- rule: {"required":true}

### spec.storage.size

`string` · required

Volume size as a Kubernetes quantity (e.g. "10Gi"). Grows are
applied to live PVCs; shrinks are rejected.

- rule: size must be a Kubernetes quantity like '10Gi' or '500Mi'
- rule: {"required":true}

### spec.storage.storageClass

`string | valueFrom`

StorageClass for the PVCs. Empty = the cluster's default class.
Accepts a literal class name or a reference to a
KubernetesStorageClass resource.

- references: KubernetesStorageClass (`status.outputs.storage_class_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesStorageClass, name: <that resource's name>, fieldPath: status.outputs.storage_class_name}} -- a bare string does not parse

### spec.resources

`ContainerResources`

CPU/memory for every database node pod. Empty = no requests/limits
(fine for evaluation, not production).

### spec.resources.limits

`CpuMemory`

The resource limits for the container.
Specify the maximum amount of CPU and memory that the container can use.

### spec.resources.limits.cpu

`string`

### spec.resources.limits.memory

`string`

### spec.resources.requests

`CpuMemory`

The resource requests for the container.
Specify the minimum amount of CPU and memory that the container is guaranteed.

### spec.resources.requests.cpu

`string`

### spec.resources.requests.memory

`string`

### spec.mysqlConfig

`string`

Extra my.cnf configuration appended to the operator's defaults
(`[mysqld]` sections and friends, verbatim). Galera/wsrep and SST
essentials are operator-managed — override them only when you know
the interaction.

### spec.autoRecovery

`bool` · optional (explicit presence)

Automatic full-cluster-crash recovery: when every node is down
(node-pool replacement, full AZ outage), the operator finds the
node with the newest data and bootstraps the cluster from it.
Upstream default: true.

- default: `true`

### spec.proxy

`KubernetesMysqlProxy`

The query-routing proxy in front of the database nodes. Omitted =
HAProxy with 3 replicas (the upstream default posture). Exactly one
proxy flavor may be enabled.

### spec.proxy.haproxy

`KubernetesMysqlHaproxy`

HAProxy — the upstream default: TCP routing, writes to one
healthy node (port 3306), reads load-balanced (port 3307).
Lighter than ProxySQL; no query awareness.

### spec.proxy.haproxy.replicas

`int32` · optional (explicit presence)

HAProxy replicas. Upstream default: 3 (2+ for availability of the
write path). 1 REQUIRES unsafe.proxy_size.

- default: `3`
- rule: {"int32":{"gte":1}}

### spec.proxy.haproxy.resources

`ContainerResources`

CPU/memory for each HAProxy pod.

### spec.proxy.haproxy.resources.limits

`CpuMemory`

The resource limits for the container.
Specify the maximum amount of CPU and memory that the container can use.

### spec.proxy.haproxy.resources.limits.cpu

`string`

### spec.proxy.haproxy.resources.limits.memory

`string`

### spec.proxy.haproxy.resources.requests

`CpuMemory`

The resource requests for the container.
Specify the minimum amount of CPU and memory that the container is guaranteed.

### spec.proxy.haproxy.resources.requests.cpu

`string`

### spec.proxy.haproxy.resources.requests.memory

`string`

### spec.proxy.haproxy.config

`string`

Extra haproxy.cfg content replacing the operator's global/defaults
sections (verbatim; the upstream default configuration is the
baseline to copy from).

### spec.proxy.haproxy.exposePrimary

`KubernetesMysqlProxyService`

The PRIMARY (write) Service — `<name>-haproxy`. Type and per-cloud
annotations for the managed-LoadBalancer recipes.

### spec.proxy.haproxy.exposePrimary.type

`string` · optional (explicit presence)

Service type: ClusterIP (default), NodePort, or LoadBalancer.

- default: `ClusterIP`
- rule: service type must be ClusterIP, NodePort, or LoadBalancer

### spec.proxy.haproxy.exposePrimary.annotations

`map<string, string>`

Service annotations — the per-cloud LoadBalancer recipe surface.

### spec.proxy.haproxy.exposeReplicas

`KubernetesMysqlHaproxyReplicasService`

The REPLICAS (read) Service — `<name>-haproxy-replicas`, port
3307 load-balanced across all nodes. Omitted = enabled (the
upstream default).

### spec.proxy.haproxy.exposeReplicas.enabled

`bool` · optional (explicit presence)

Render the replicas Service. Upstream default: true.

- default: `true`

### spec.proxy.haproxy.exposeReplicas.onlyReaders

`bool`

Route reads to REPLICA nodes only, keeping the write node free of
read traffic. Upstream default: false (reads hit every node).

### spec.proxy.haproxy.exposeReplicas.type

`string` · optional (explicit presence)

Service type: ClusterIP (default), NodePort, or LoadBalancer.

- default: `ClusterIP`
- rule: service type must be ClusterIP, NodePort, or LoadBalancer

### spec.proxy.haproxy.exposeReplicas.annotations

`map<string, string>`

Service annotations (per-cloud LoadBalancer recipes).

### spec.proxy.proxysql

`KubernetesMysqlProxysql`

ProxySQL — SQL-aware routing: query rules, read/write split at
the statement level, a stateful configuration database on its own
volume. The heavier, more capable choice.

### spec.proxy.proxysql.replicas

`int32` · optional (explicit presence)

ProxySQL replicas. Upstream default: 3. 1 REQUIRES
unsafe.proxy_size.

- default: `3`
- rule: {"int32":{"gte":1}}

### spec.proxy.proxysql.resources

`ContainerResources`

CPU/memory for each ProxySQL pod.

### spec.proxy.proxysql.resources.limits

`CpuMemory`

The resource limits for the container.
Specify the maximum amount of CPU and memory that the container can use.

### spec.proxy.proxysql.resources.limits.cpu

`string`

### spec.proxy.proxysql.resources.limits.memory

`string`

### spec.proxy.proxysql.resources.requests

`CpuMemory`

The resource requests for the container.
Specify the minimum amount of CPU and memory that the container is guaranteed.

### spec.proxy.proxysql.resources.requests.cpu

`string`

### spec.proxy.proxysql.resources.requests.memory

`string`

### spec.proxy.proxysql.storage

`KubernetesMysqlStorage` · required

Storage for ProxySQL's own configuration database. Required —
ProxySQL is stateful.

- rule: {"required":true}

### spec.proxy.proxysql.storage.size

`string` · required

Volume size as a Kubernetes quantity (e.g. "10Gi"). Grows are
applied to live PVCs; shrinks are rejected.

- rule: size must be a Kubernetes quantity like '10Gi' or '500Mi'
- rule: {"required":true}

### spec.proxy.proxysql.storage.storageClass

`string | valueFrom`

StorageClass for the PVCs. Empty = the cluster's default class.
Accepts a literal class name or a reference to a
KubernetesStorageClass resource.

- references: KubernetesStorageClass (`status.outputs.storage_class_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesStorageClass, name: <that resource's name>, fieldPath: status.outputs.storage_class_name}} -- a bare string does not parse

### spec.proxy.proxysql.config

`string`

Extra proxysql.cnf content replacing the operator's defaults
(verbatim).

### spec.proxy.proxysql.exposePrimary

`KubernetesMysqlProxyService`

The client-facing Service — `<name>-proxysql`. Type and per-cloud
annotations for the managed-LoadBalancer recipes.

### spec.proxy.proxysql.exposePrimary.type

`string` · optional (explicit presence)

Service type: ClusterIP (default), NodePort, or LoadBalancer.

- default: `ClusterIP`
- rule: service type must be ClusterIP, NodePort, or LoadBalancer

### spec.proxy.proxysql.exposePrimary.annotations

`map<string, string>`

Service annotations — the per-cloud LoadBalancer recipe surface.

### spec.tls

`KubernetesMysqlTls`

TLS for client and replication traffic. Omitted = enabled with
operator-generated certificates (the upstream default). Point
issuer at a cert-manager (Cluster)Issuer for an
organization-trusted chain; disabling TLS entirely REQUIRES
unsafe.tls.

### spec.tls.enabled

`bool` · optional (explicit presence)

Serve TLS. Upstream default: true (operator-generated
certificates). Disabling REQUIRES unsafe.tls — a plaintext MySQL
wire is a development posture.

- default: `true`

### spec.tls.issuer

`string | valueFrom`

cert-manager issuer for the cluster's certificates instead of the
operator's self-generated ones — the organization-trust seam.
References a KubernetesClusterIssuer (set issuer_kind to "Issuer"
for a namespaced KubernetesIssuer).

- references: KubernetesClusterIssuer (`metadata.name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesClusterIssuer, name: <that resource's name>, fieldPath: metadata.name}} -- a bare string does not parse

### spec.tls.issuerKind

`string` · optional (explicit presence)

Kind of the cert-manager issuer named above: "ClusterIssuer" (the
default) or "Issuer" (must live in the database's namespace).

- default: `ClusterIssuer`
- rule: issuer_kind must be ClusterIssuer or Issuer

### spec.tls.sans

`[]string`

Additional DNS SANs for the generated certificates (external
hostnames the database is reached at through composed exposure).

### spec.users

`[]KubernetesMysqlUser`

Declarative application users: the operator creates them, keeps
their grants reconciled, and manages their password Secrets.

- rule: each user needs a distinct name

### spec.users[].name

`string` · required

Username.

- rule: {"required":true}

### spec.users[].dbs

`[]string`

Databases the grants below apply to. Empty = server-wide grants
(*.*) — administrative users only.

### spec.users[].hosts

`[]string`

MySQL hosts clause per grant (e.g. "%" for anywhere — the upstream
default).

### spec.users[].grants

`[]string`

Privileges to grant (SELECT, INSERT, UPDATE, DELETE, ALL, ...).
Empty = the operator's default (ALL on the listed dbs).

### spec.users[].withGrantOption

`bool`

GRANT OPTION: the user can grant its own privileges onward.

### spec.users[].password

`string` · sensitive

Password, materialized as a Kubernetes Secret the operator watches
(`<cluster>-user-<name>` unless the operator generates one) —
rotating the value rotates the database password. Empty = the
operator generates a password into the same Secret.

### spec.backup

`KubernetesMysqlBackup`

Scheduled XtraBackup backups (and point-in-time recovery) to
S3/S3-compatible/Azure-Blob stores or a PVC. Omitted = no backups
(a deliberate choice to make, not a default to forget).

### spec.backup.storages

`[]KubernetesMysqlBackupStorage` · required

Backup storage destinations, referenced by schedules and PITR via
their names.

- rule: each backup storage needs a distinct name
- rule: {"repeated":{"minItems":"1"}}

### spec.backup.storages[].name

`string` · required

Name schedules and PITR reference this storage by.

- rule: storage name must be a lowercase DNS label (letters, numbers, hyphens)
- rule: {"required":true}

### spec.backup.storages[].s3

`KubernetesMysqlS3Storage`

AWS S3 — or ANY S3-compatible store (MinIO, Ceph RGW, GCS via
its S3 interoperability endpoint, ...) via the endpoint_url
override.

### spec.backup.storages[].s3.bucket

`string` · required

Bucket name.

- rule: {"required":true}

### spec.backup.storages[].s3.region

`string`

AWS region of the bucket. Required for real S3; S3-compatible
stores use their expected value (MinIO accepts any).

### spec.backup.storages[].s3.prefix

`string`

Key prefix inside the bucket (a folder for this cluster's
backups).

### spec.backup.storages[].s3.endpointUrl

`string`

S3-COMPATIBLE ARM: endpoint URL of the store (e.g.
http://minio.minio-system.svc:9000 for in-cluster MinIO,
https://storage.googleapis.com for GCS interoperability). Empty =
real AWS S3.

- rule: endpoint_url must be an http(s) URL (e.g. http://minio.minio-system.svc:9000)

### spec.backup.storages[].s3.forcePathStyle

`bool`

Force path-style addressing (bucket in the path, not the
subdomain) — required by MinIO and most self-hosted stores.

### spec.backup.storages[].s3.accessKeys

`KubernetesMysqlS3AccessKeys` · required

Access keys, materialized as a Kubernetes Secret the backup jobs
read. Required: XtraBackup's S3 uploads authenticate with keys
(for AWS, mint scoped keys for a backup-only IAM user).

- rule: {"required":true}

### spec.backup.storages[].s3.accessKeys.accessKeyId

`string` · required

Access key ID — the public identifier of the key pair, not a
secret. For MinIO this is the access key / username.

- rule: {"required":true}

### spec.backup.storages[].s3.accessKeys.secretAccessKey

`string` · required · sensitive

Secret access key (for MinIO: the secret key / password).

- rule: {"required":true}

### spec.backup.storages[].azure

`KubernetesMysqlAzureStorage`

Azure Blob Storage.

### spec.backup.storages[].azure.container

`string` · required

Blob container name.

- rule: {"required":true}

### spec.backup.storages[].azure.prefix

`string`

Key prefix inside the container.

### spec.backup.storages[].azure.endpointUrl

`string`

Endpoint URL override (sovereign clouds, Azurite). Empty = the
account's default blob endpoint.

### spec.backup.storages[].azure.storageAccount

`string` · required

Storage-account name.

- rule: {"required":true}

### spec.backup.storages[].azure.accessKey

`string` · required · sensitive

Storage-account access key, materialized as a Kubernetes Secret
the backup jobs read.

- rule: {"required":true}

### spec.backup.storages[].pvc

`KubernetesMysqlPvcStorage`

A PersistentVolumeClaim next to the cluster — backups without an
object store. No PITR and no off-cluster durability; prefer an
object store for production.

### spec.backup.storages[].pvc.volume

`KubernetesMysqlStorage` · required

The backup volume.

- rule: {"required":true}

### spec.backup.storages[].pvc.volume.size

`string` · required

Volume size as a Kubernetes quantity (e.g. "10Gi"). Grows are
applied to live PVCs; shrinks are rejected.

- rule: size must be a Kubernetes quantity like '10Gi' or '500Mi'
- rule: {"required":true}

### spec.backup.storages[].pvc.volume.storageClass

`string | valueFrom`

StorageClass for the PVCs. Empty = the cluster's default class.
Accepts a literal class name or a reference to a
KubernetesStorageClass resource.

- references: KubernetesStorageClass (`status.outputs.storage_class_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesStorageClass, name: <that resource's name>, fieldPath: status.outputs.storage_class_name}} -- a bare string does not parse

### spec.backup.storages[].verifyTls

`bool` · optional (explicit presence)

Verify the store's TLS certificate. Upstream default: true —
disable only for self-signed test endpoints.

- default: `true`

### spec.backup.schedules

`[]KubernetesMysqlBackupSchedule`

Scheduled backups (rendered into the cluster's backup schedule —
the operator runs one XtraBackup job per tick).

- rule: each backup schedule needs a distinct name

### spec.backup.schedules[].name

`string` · required

Schedule name.

- rule: schedule name must be a lowercase DNS label (letters, numbers, hyphens)
- rule: {"required":true}

### spec.backup.schedules[].schedule

`string` · required

Standard FIVE-field cron expression (e.g. "0 2 * * *" = daily at
02:00).

- rule: schedule is a five-field cron expression — e.g. '0 2 * * *' for daily at 02:00
- rule: {"required":true}

### spec.backup.schedules[].storageName

`string` · required

Name of the declared storage this schedule writes to.

- rule: {"required":true}

### spec.backup.schedules[].keep

`int32` · optional (explicit presence)

How many backups this schedule keeps. Older ones are pruned.

- rule: {"int32":{"gte":1}}

### spec.backup.schedules[].deleteFromStorage

`bool` · optional (explicit presence)

Also delete pruned backups FROM THE STORE (not just the cluster's
records). Upstream default when retention is set: true.

- default: `true`

### spec.backup.pitr

`KubernetesMysqlPitr`

Point-in-time recovery: continuously upload binlog events so a
restore can land between backups. The referenced storage should be
DEDICATED to PITR (the upstream recommendation — never share it
with base backups).

- rule: PITR needs storage_name — the declared backup storage binlogs are shipped to

### spec.backup.pitr.enabled

`bool`

Continuously ship binlogs to the storage below.

### spec.backup.pitr.storageName

`string`

Name of the declared storage binlogs land in. DEDICATE a storage
to PITR (never the base-backup storage — the upstream
requirement).

### spec.backup.pitr.timeBetweenUploads

`int32` · optional (explicit presence)

Seconds between binlog uploads. Upstream default: 60 — the
recovery-point objective for a total-cluster loss.

- default: `60`
- rule: {"int32":{"gte":1}}

### spec.scheduling

`KubernetesMysqlScheduling`

Where database pods run: anti-affinity spreading, node selection,
tolerations, and scheduling priority.

### spec.scheduling.antiAffinityTopologyKey

`string`

Topology key the operator's anti-affinity spreads database nodes
across. Upstream default: kubernetes.io/hostname (one node per
host); use topology.kubernetes.io/zone to spread across zones, or
"none" to disable anti-affinity (development only).

### spec.scheduling.nodeSelector

`map<string, string>`

Node selector for the database pods.

### spec.scheduling.tolerations

`[]WorkloadToleration`

Tolerations for the database pods.

### spec.scheduling.tolerations[].key

`string`

Taint key to tolerate. Empty key with operator "Exists" tolerates every taint.

### spec.scheduling.tolerations[].operator

`string`

How key/value match: "Equal" (default — value must match too) or "Exists"
(key presence alone matches).

- rule: Toleration operator must be either "Equal" or "Exists"

### spec.scheduling.tolerations[].value

`string`

Taint value to match when operator is "Equal".

### spec.scheduling.tolerations[].effect

`string`

Which taint effect is tolerated: "NoSchedule", "PreferNoSchedule", or
"NoExecute". Empty tolerates all effects for the key.

- rule: Toleration effect must be one of "NoSchedule", "PreferNoSchedule", or "NoExecute"

### spec.scheduling.tolerations[].tolerationSeconds

`int64` · optional (explicit presence)

For "NoExecute" taints only: how many seconds already-running pods stay bound
after the taint appears. Unset means tolerate forever.

### spec.scheduling.priorityClassName

`string`

PriorityClass for the database pods — databases should outlive
stateless workloads under node pressure.

### spec.podDisruptionBudget

`KubernetesMysqlPodDisruptionBudget`

PodDisruptionBudget for the database nodes. Omitted = the upstream
default (max one node down to voluntary disruptions).

- rule: declare at most one PDB bound — max_unavailable or min_available

### spec.podDisruptionBudget.maxUnavailable

`int32`

Maximum database pods down during voluntary disruptions. Upstream
default: 1 — never raise it above (instances - quorum). Mutually
exclusive with min_available.

- rule: {"int32":{"gte":0}}

### spec.podDisruptionBudget.minAvailable

`int32`

Minimum database pods that must stay up. Mutually exclusive with
max_unavailable.

- rule: {"int32":{"gte":0}}

### spec.logCollector

`KubernetesMysqlLogCollector`

The log-collector sidecar (fluent-bit) shipping mysqld logs.
Omitted = enabled (the upstream default).

### spec.logCollector.enabled

`bool` · optional (explicit presence)

Run the sidecar. Upstream default: true.

- default: `true`

### spec.logCollector.resources

`ContainerResources`

CPU/memory for the sidecar.

### spec.logCollector.resources.limits

`CpuMemory`

The resource limits for the container.
Specify the maximum amount of CPU and memory that the container can use.

### spec.logCollector.resources.limits.cpu

`string`

### spec.logCollector.resources.limits.memory

`string`

### spec.logCollector.resources.requests

`CpuMemory`

The resource requests for the container.
Specify the minimum amount of CPU and memory that the container is guaranteed.

### spec.logCollector.resources.requests.cpu

`string`

### spec.logCollector.resources.requests.memory

`string`

### spec.updateStrategy

`string` · optional (explicit presence)

How updates roll across the cluster: SmartUpdate (the operator
orders restarts safely — the upstream default), RollingUpdate
(StatefulSet semantics), or OnDelete.

- default: `SmartUpdate`
- rule: update_strategy must be SmartUpdate, RollingUpdate, or OnDelete

### spec.unsafe

`KubernetesMysqlUnsafe`

Explicit opt-in to unsafe topologies and postures the operator
otherwise REJECTS. Development conveniences — never production.

### spec.unsafe.clusterSize

`bool`

Allow fewer than 3 database nodes (single-node development
clusters).

### spec.unsafe.tls

`bool`

Allow disabling TLS (tls.enabled false).

### spec.unsafe.proxySize

`bool`

Allow fewer than 2 proxy replicas.

### spec.unsafe.backupIfUnhealthy

`bool`

Allow backups to run against an unhealthy cluster.

### spec.pause

`bool`

Pause the cluster: scale everything to zero while keeping the data
volumes (and resume by flipping back).

### spec.imagePullSecrets

`[]string`

Names of image-pull secrets (in the cluster's namespace) for
pulling images from a private registry.

## Validation Rules

- `spec.instances_quorum_or_unsafe`: fewer than 3 nodes loses Galera quorum safety — the operator rejects it unless unsafe.cluster_size explicitly opts in (development only)
- `spec.tls_disabled_or_unsafe`: disabling TLS is a plaintext development posture — the operator rejects it unless unsafe.tls explicitly opts in
- `spec.proxy_replicas_or_unsafe`: fewer than 2 proxy replicas loses write-path availability — the operator rejects it unless unsafe.proxy_size explicitly opts in (development only)

## Outputs

Reference an output from another manifest as `valueFrom: {kind: KubernetesMysql, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.namespace` | `string` | Namespace the cluster runs in. |
| `status.outputs.cluster_name` | `string` | The PerconaXtraDBCluster resource name (`metadata.name`) — every operator-created object derives from it. |
| `status.outputs.primary_service` | `string` | The WRITE Service applications connect to: `<name>-haproxy` (port 3306) or `<name>-proxysql` depending on the chosen proxy. |
| `status.outputs.replicas_service` | `string` | The READ Service (`<name>-haproxy-replicas`, port 3307) — HAProxy with the replicas Service enabled; empty otherwise. |
| `status.outputs.kube_endpoint` | `string` | In-cluster endpoint of the write path: `<primary_service>.<namespace>.svc.cluster.local:3306`. |
| `status.outputs.port_forward_command` | `string` | kubectl port-forward one-liner for reaching the database from a workstation. |
| `status.outputs.root_password_secret` | `KubernetesSecretKey` | The Kubernetes Secret key holding the root password (the operator-managed `<name>-secrets` system-users Secret, key "root"). |
| `status.outputs.root_password_secret.name` | `string` | The name of the Kubernetes Secret. |
| `status.outputs.root_password_secret.key` | `string` | The key within the Kubernetes Secret. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.namespace` | KubernetesNamespace | `spec.name` |
| `spec.storage.storageClass` | KubernetesStorageClass | `status.outputs.storage_class_name` |
| `spec.proxy.proxysql.storage.storageClass` | KubernetesStorageClass | `status.outputs.storage_class_name` |
| `spec.tls.issuer` | KubernetesClusterIssuer | `metadata.name` |
| `spec.backup.storages[].pvc.volume.storageClass` | KubernetesStorageClass | `status.outputs.storage_class_name` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| KubernetesAirflow | `spec.database.mysql.host` | `status.outputs.primary_service` |
| KubernetesAirflow | `spec.database.mysql.passwordSecret.secretName` | `status.outputs.root_password_secret.name` |
| KubernetesJupyterHub | `spec.hub.database.mysql.host` | `status.outputs.primary_service` |
| KubernetesJupyterHub | `spec.hub.database.mysql.passwordSecret.secretName` | `status.outputs.root_password_secret.name` |
| KubernetesMlflow | `spec.backendStore.mysql.host` | `status.outputs.primary_service` |
| KubernetesMlflow | `spec.backendStore.mysql.passwordSecret.secretName` | `status.outputs.root_password_secret.name` |
| KubernetesOpenFga | `spec.datastore.mysql.host` | `status.outputs.primary_service` |
| KubernetesTemporal | `spec.database.mysql.host` | `status.outputs.primary_service` |
| KubernetesTemporal | `spec.database.mysql.passwordSecret.secretName` | `status.outputs.root_password_secret.name` |
| KubernetesTemporal | `spec.database.visibility.mysql.host` | `status.outputs.primary_service` |
| KubernetesTemporal | `spec.database.visibility.mysql.passwordSecret.secretName` | `status.outputs.root_password_secret.name` |
| KubernetesTrino | `spec.catalogs.mysql[].host` | `status.outputs.primary_service` |
| KubernetesTrino | `spec.catalogs.mysql[].passwordSecret.secretName` | `status.outputs.root_password_secret.name` |

## See Also

- [Overview](../README.md)
