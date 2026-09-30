# KubernetesOpenSearch

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `kubernetes.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

**KubernetesOpenSearchSpec** declares an OpenSearch cluster — the
Apache-2.0 search and analytics engine (drop-in replacement for the
Elasticsearch APIs at the 7.10 fork line, with its own 2.x/3.x
feature line since) — as an `OpenSearchCluster` custom resource
reconciled by the OpenSearch Kubernetes Operator
(KubernetesOpenSearchOperator, the registry prerequisite).

The operator manages the full cluster lifecycle: node StatefulSets
per pool, cluster bootstrap, TLS (generated or provided), the
security plugin's admin bootstrap, safe rolling upgrades and drain
ordering, and an optional OpenSearch Dashboards deployment (the
Kibana-role console — a section of the same custom resource, not a
separate component).

TOPOLOGY: `node_pools` is the cluster's shape — every pool declares
its roles (cluster_manager, data, ingest, ...). The smallest real
cluster is one pool with all roles and 3 replicas; a 1-replica
all-roles pool works for development. Storage is per-pool:
PVC-backed by default (survives pod loss), emptyDir for
throwaway data.

SECURITY: declare `security` with generated TLS (the scenarios'
posture) so the operator issues real certificates for both layers.
The HTTP API serves https in EVERY posture — even without a security
block the image's DEMO security config is active (verified in the
operator source). KNOW THIS: the operator does not generate a random
admin password at this release — the bootstrapped credentials in the
`<name>-admin-password` Secret are the image's well-known demo
admin credentials. Fine inside a private cluster for development;
for anything real, bring a custom `security.config` (your own
internal_users.yml and admin credentials) or rotate the admin
password through the security API immediately after install.
Clients read credentials from the Secret named in the stack
outputs — no credential ever appears in this spec unless you bring
your own security config.

EXPOSURE: the cluster and Dashboards Services are ClusterIP by
design; exposure composes from first-class kinds (KubernetesIngress,
Gateway API kinds) referencing the exported service handles, or via
the Dashboards service type for a quick LoadBalancer.

## Example

```yaml
# Full-surface manifest for offline module proofs (tofu validate/plan and
# pulumi preview). Exercises every typed block coherently: a three-pool
# topology (dedicated cluster managers, PVC-backed data nodes with a hard
# PDB, emptyDir coordinators with a percentage PDB), cluster- and pool-level
# additional config, bootstrap tuning, generated transport/HTTP TLS plus a
# custom security config, Dashboards with TLS and a LoadBalancer service,
# Prometheus monitoring, keystore loads, an S3 snapshot repository (with the
# matching repository-s3 plugin), an additional certificate volume, a custom
# image and pull secrets. Both engines must render an identical
# OpenSearchCluster CR from it.
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesOpenSearch
metadata:
  name: os-hack
spec:
  namespace:
    value: os-hack-ns
  createNamespace: true
  version: 2.19.4
  httpPort: 9201
  nodePools:
    - name: managers
      replicas: 3
      roles:
        - cluster_manager
      resources:
        limits:
          cpu: 1000m
          memory: 2Gi
        requests:
          cpu: 500m
          memory: 2Gi
      jvm: -Xmx1G -Xms1G
      diskSize: 10Gi
      persistence:
        pvc:
          storageClass:
            value: fast-ssd
      pdb:
        enable: true
        minAvailable: "2"
    - name: data
      replicas: 3
      roles:
        - data
        - ingest
      resources:
        limits:
          cpu: 2000m
          memory: 8Gi
        requests:
          cpu: 1000m
          memory: 8Gi
      jvm: -Xmx4G -Xms4G
      diskSize: 100Gi
      persistence:
        pvc:
          storageClass:
            value: fast-ssd
      additionalConfig:
        indices.memory.index_buffer_size: "20%"
      nodeSelector:
        workload: search
      tolerations:
        - key: dedicated
          operator: Equal
          value: search
          effect: NoSchedule
      pdb:
        enable: true
        maxUnavailable: 25%
    - name: coordinators
      replicas: 2
      roles:
        - remote_cluster_client
      persistence:
        emptyDir:
          sizeLimit: 5Gi
  additionalConfig:
    action.auto_create_index: "false"
    cluster.routing.allocation.disk.watermark.low: 85%
  serviceAnnotations:
    service.beta.kubernetes.io/aws-load-balancer-internal: "true"
  setVmMaxMapCount: true
  drainDataNodes: true
  pluginsList:
    - repository-s3
  bootstrap:
    resources:
      limits:
        cpu: 500m
        memory: 1Gi
      requests:
        cpu: 250m
        memory: 1Gi
    jvm: -Xmx512M -Xms512M
    additionalConfig:
      cluster.initial_cluster_manager_nodes: os-hack-bootstrap-0
  security:
    transportTls:
      generate: true
      perNode: true
    httpTls:
      generate: false
      secret:
        value: os-hack-http-cert
    config:
      securityConfigSecret:
        value: os-hack-securityconfig
      adminSecret:
        value: os-hack-admin-cert
      adminCredentialsSecret:
        value: os-hack-admin-creds
  dashboards:
    enabled: true
    replicas: 2
    version: 2.19.4
    resources:
      limits:
        cpu: 500m
        memory: 1Gi
      requests:
        cpu: 250m
        memory: 512Mi
    tls:
      enable: true
      generate: true
    basePath: /dashboards
    additionalConfig:
      opensearch_security.multitenancy.enabled: "true"
    opensearchCredentialsSecret:
      value: os-hack-dashboards-creds
    service:
      type: LoadBalancer
      loadBalancerSourceRanges:
        - 10.0.0.0/8
    pluginsList:
      - https://example.com/dashboards-plugin.zip
  monitoring:
    enabled: true
    scrapeInterval: 30s
    monitoringUserSecret:
      value: os-hack-monitoring-creds
    pluginUrl: https://example.com/prometheus-exporter-2.19.4.0.zip
  keystore:
    - secret:
        value: os-hack-s3-keys
      keyMappings:
        accessKey: s3.client.default.access_key
        secretKey: s3.client.default.secret_key
  snapshotRepositories:
    - name: nightly-backups
      type: s3
      settings:
        bucket: os-hack-snapshots
        region: us-east-1
        base_path: os-hack
  additionalVolumes:
    - name: extra-ca
      path: /usr/share/opensearch/config/extra-ca
      subPath: ca.crt
      secretName: os-hack-extra-ca
      restartPods: true
    - name: custom-log4j
      path: /usr/share/opensearch/config/log4j2.properties.d
      configMapName: os-hack-log4j
  image:
    repo: registry.example.com/opensearch/opensearch
    tag: 2.19.4-custom
  imagePullSecrets:
    - registry-pull-secret
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.namespace` | `string \| valueFrom` | yes |  | KubernetesNamespace (`spec.name`) |
| `spec.createNamespace` | `bool` |  |  |  |
| `spec.version` | `string` | yes |  |  |
| `spec.httpPort` | `int32` |  | `9200` |  |
| `spec.nodePools` | `[]KubernetesOpenSearchNodePool` | yes |  |  |
| `spec.nodePools[].name` | `string` | yes |  |  |
| `spec.nodePools[].replicas` | `int32` | yes |  |  |
| `spec.nodePools[].roles` | `[]string` | yes |  |  |
| `spec.nodePools[].resources` | `ContainerResources` |  |  |  |
| `spec.nodePools[].resources.limits` | `CpuMemory` |  |  |  |
| `spec.nodePools[].resources.limits.cpu` | `string` |  |  |  |
| `spec.nodePools[].resources.limits.memory` | `string` |  |  |  |
| `spec.nodePools[].resources.requests` | `CpuMemory` |  |  |  |
| `spec.nodePools[].resources.requests.cpu` | `string` |  |  |  |
| `spec.nodePools[].resources.requests.memory` | `string` |  |  |  |
| `spec.nodePools[].jvm` | `string` |  |  |  |
| `spec.nodePools[].diskSize` | `string` |  |  |  |
| `spec.nodePools[].persistence` | `KubernetesOpenSearchPersistence` |  |  |  |
| `spec.nodePools[].persistence.pvc` | `KubernetesOpenSearchPvc` |  |  |  |
| `spec.nodePools[].persistence.pvc.storageClass` | `string \| valueFrom` |  |  | KubernetesStorageClass (`status.outputs.storage_class_name`) |
| `spec.nodePools[].persistence.emptyDir` | `KubernetesOpenSearchEmptyDir` |  |  |  |
| `spec.nodePools[].persistence.emptyDir.sizeLimit` | `string` |  |  |  |
| `spec.nodePools[].additionalConfig` | `map<string, string>` |  |  |  |
| `spec.nodePools[].nodeSelector` | `map<string, string>` |  |  |  |
| `spec.nodePools[].tolerations` | `[]WorkloadToleration` |  |  |  |
| `spec.nodePools[].tolerations[].key` | `string` |  |  |  |
| `spec.nodePools[].tolerations[].operator` | `string` |  |  |  |
| `spec.nodePools[].tolerations[].value` | `string` |  |  |  |
| `spec.nodePools[].tolerations[].effect` | `string` |  |  |  |
| `spec.nodePools[].tolerations[].tolerationSeconds` | `int64` |  |  |  |
| `spec.nodePools[].pdb` | `KubernetesOpenSearchPdb` |  |  |  |
| `spec.nodePools[].pdb.enable` | `bool` |  |  |  |
| `spec.nodePools[].pdb.minAvailable` | `string` |  |  |  |
| `spec.nodePools[].pdb.maxUnavailable` | `string` |  |  |  |
| `spec.additionalConfig` | `map<string, string>` |  |  |  |
| `spec.serviceAnnotations` | `map<string, string>` |  |  |  |
| `spec.setVmMaxMapCount` | `bool` |  | `true` |  |
| `spec.drainDataNodes` | `bool` |  |  |  |
| `spec.pluginsList` | `[]string` |  |  |  |
| `spec.bootstrap` | `KubernetesOpenSearchBootstrap` |  |  |  |
| `spec.bootstrap.resources` | `ContainerResources` |  |  |  |
| `spec.bootstrap.resources.limits` | `CpuMemory` |  |  |  |
| `spec.bootstrap.resources.limits.cpu` | `string` |  |  |  |
| `spec.bootstrap.resources.limits.memory` | `string` |  |  |  |
| `spec.bootstrap.resources.requests` | `CpuMemory` |  |  |  |
| `spec.bootstrap.resources.requests.cpu` | `string` |  |  |  |
| `spec.bootstrap.resources.requests.memory` | `string` |  |  |  |
| `spec.bootstrap.jvm` | `string` |  |  |  |
| `spec.bootstrap.additionalConfig` | `map<string, string>` |  |  |  |
| `spec.security` | `KubernetesOpenSearchSecurity` |  |  |  |
| `spec.security.transportTls` | `KubernetesOpenSearchTlsTransport` |  |  |  |
| `spec.security.transportTls.generate` | `bool` |  | `true` |  |
| `spec.security.transportTls.perNode` | `bool` |  | `true` |  |
| `spec.security.transportTls.secret` | `string \| valueFrom` |  |  | KubernetesCertificate (`status.outputs.secret_name`) |
| `spec.security.transportTls.caSecret` | `string \| valueFrom` |  |  | KubernetesSecret (`status.outputs.secret_name`) |
| `spec.security.transportTls.nodesDn` | `[]string` |  |  |  |
| `spec.security.transportTls.adminDn` | `[]string` |  |  |  |
| `spec.security.httpTls` | `KubernetesOpenSearchTlsHttp` |  |  |  |
| `spec.security.httpTls.generate` | `bool` |  | `true` |  |
| `spec.security.httpTls.secret` | `string \| valueFrom` |  |  | KubernetesCertificate (`status.outputs.secret_name`) |
| `spec.security.config` | `KubernetesOpenSearchSecurityConfig` |  |  |  |
| `spec.security.config.securityConfigSecret` | `string \| valueFrom` |  |  | KubernetesSecret (`status.outputs.secret_name`) |
| `spec.security.config.adminSecret` | `string \| valueFrom` |  |  | KubernetesSecret (`status.outputs.secret_name`) |
| `spec.security.config.adminCredentialsSecret` | `string \| valueFrom` |  |  | KubernetesSecret (`status.outputs.secret_name`) |
| `spec.dashboards` | `KubernetesOpenSearchDashboards` |  |  |  |
| `spec.dashboards.enabled` | `bool` |  |  |  |
| `spec.dashboards.replicas` | `int32` |  | `1` |  |
| `spec.dashboards.version` | `string` |  |  |  |
| `spec.dashboards.resources` | `ContainerResources` |  |  |  |
| `spec.dashboards.resources.limits` | `CpuMemory` |  |  |  |
| `spec.dashboards.resources.limits.cpu` | `string` |  |  |  |
| `spec.dashboards.resources.limits.memory` | `string` |  |  |  |
| `spec.dashboards.resources.requests` | `CpuMemory` |  |  |  |
| `spec.dashboards.resources.requests.cpu` | `string` |  |  |  |
| `spec.dashboards.resources.requests.memory` | `string` |  |  |  |
| `spec.dashboards.tls` | `KubernetesOpenSearchDashboardsTls` |  |  |  |
| `spec.dashboards.tls.enable` | `bool` |  |  |  |
| `spec.dashboards.tls.generate` | `bool` |  | `true` |  |
| `spec.dashboards.tls.secret` | `string \| valueFrom` |  |  | KubernetesCertificate (`status.outputs.secret_name`) |
| `spec.dashboards.basePath` | `string` |  |  |  |
| `spec.dashboards.additionalConfig` | `map<string, string>` |  |  |  |
| `spec.dashboards.opensearchCredentialsSecret` | `string \| valueFrom` |  |  | KubernetesSecret (`status.outputs.secret_name`) |
| `spec.dashboards.service` | `KubernetesOpenSearchDashboardsService` |  |  |  |
| `spec.dashboards.service.type` | `string` |  | `ClusterIP` |  |
| `spec.dashboards.service.loadBalancerSourceRanges` | `[]string` |  |  |  |
| `spec.dashboards.pluginsList` | `[]string` |  |  |  |
| `spec.monitoring` | `KubernetesOpenSearchMonitoring` |  |  |  |
| `spec.monitoring.enabled` | `bool` |  |  |  |
| `spec.monitoring.scrapeInterval` | `string` |  |  |  |
| `spec.monitoring.monitoringUserSecret` | `string \| valueFrom` |  |  | KubernetesSecret (`status.outputs.secret_name`) |
| `spec.monitoring.pluginUrl` | `string` |  |  |  |
| `spec.keystore` | `[]KubernetesOpenSearchKeystoreValue` |  |  |  |
| `spec.keystore[].secret` | `string \| valueFrom` | yes |  | KubernetesSecret (`status.outputs.secret_name`) |
| `spec.keystore[].keyMappings` | `map<string, string>` |  |  |  |
| `spec.snapshotRepositories` | `[]KubernetesOpenSearchSnapshotRepo` |  |  |  |
| `spec.snapshotRepositories[].name` | `string` | yes |  |  |
| `spec.snapshotRepositories[].type` | `string` | yes |  |  |
| `spec.snapshotRepositories[].settings` | `map<string, string>` |  |  |  |
| `spec.additionalVolumes` | `[]KubernetesOpenSearchAdditionalVolume` |  |  |  |
| `spec.additionalVolumes[].name` | `string` | yes |  |  |
| `spec.additionalVolumes[].path` | `string` | yes |  |  |
| `spec.additionalVolumes[].subPath` | `string` |  |  |  |
| `spec.additionalVolumes[].secretName` | `string` |  |  |  |
| `spec.additionalVolumes[].configMapName` | `string` |  |  |  |
| `spec.additionalVolumes[].restartPods` | `bool` |  |  |  |
| `spec.image` | `ContainerImage` |  |  |  |
| `spec.image.repo` | `string` |  |  |  |
| `spec.image.tag` | `string` |  |  |  |
| `spec.image.pullSecretName` | `string` |  |  |  |
| `spec.imagePullSecrets` | `[]string` |  |  |  |

## Field Details

### spec.namespace

`string | valueFrom` · required

Namespace to create the cluster in. Accepts a literal namespace
name or a reference to a KubernetesNamespace resource.

- references: KubernetesNamespace (`spec.name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesNamespace, name: <that resource's name>, fieldPath: spec.name}} -- a bare string does not parse

### spec.createNamespace

`bool`

When true, the namespace is created (with the standard Planton
governance labels) before the cluster and deleted with the
resource. When false, the namespace must already exist.

### spec.version

`string` · required

OpenSearch version to run, e.g. "2.19.4". Must be a published
opensearchproject/opensearch image tag. Check the operator's
compatibility table before pinning a major line — the pinned
stable operator supports OpenSearch 2.19.x through the 3.x line.

- rule: {"required":true}

### spec.httpPort

`int32` · optional (explicit presence)

HTTP API port. Operator default: 9200. Changing it changes every
endpoint the cluster advertises.

- default: `9200`
- rule: {"int32":{"lte":65535,"gte":1}}

### spec.nodePools

`[]KubernetesOpenSearchNodePool` · required

Node pools — the cluster topology. At least one pool; every pool
carries its own roles, sizing and storage. Pool names become part
of pod/StatefulSet names (`<cluster>-<pool>`), so keep them short
and DNS-safe.

MANAGER FLOOR: the cluster needs at least TWO
cluster_manager-eligible replicas in total (three recommended).
Verified live: the operator seeds every new cluster through a
temporary bootstrap manager and deletes it once the cluster
initializes — a single-manager cluster is stranded by that handoff
(the quorum never re-forms; every write returns
cluster_manager_not_discovered) while two or more managers hold a
majority through it. The smallest working dev shape is one
all-roles pool with 2 replicas.

- rule: a single cluster_manager-eligible replica cannot survive the operator's bootstrap handoff (verified live) — declare at least 2 manager-eligible replicas in total
- rule: {"repeated":{"minItems":"1"}}

### spec.nodePools[].name

`string` · required

Pool name (becomes `<cluster>-<name>` in pod names). Keep short,
lowercase, DNS-safe.

- rule: {"required":true,"string":{"maxLen":"20","pattern":"^[a-z0-9]([-a-z0-9]*[a-z0-9])?$"}}

### spec.nodePools[].replicas

`int32` · required

Number of nodes in this pool.

- rule: {"required":true,"int32":{"gte":1}}

### spec.nodePools[].roles

`[]string` · required

OpenSearch roles for the pool's nodes: cluster_manager, data,
ingest, ml, remote_cluster_client, search (underscore forms — a
dashed "cluster-manager" is NOT a role and fails only at node
startup, never at apply: the CRD deliberately leaves this open for
new upstream roles, so typos travel all the way to the pod). At
least one. Every cluster needs at least one pool carrying
cluster_manager and one carrying data (one pool can carry both).

- rule: {"repeated":{"minItems":"1"}}

### spec.nodePools[].resources

`ContainerResources`

Node container resources. Empty = operator defaults. Size memory
together with `jvm` — the JVM heap should be about half the
container memory.

### spec.nodePools[].resources.limits

`CpuMemory`

The resource limits for the container.
Specify the maximum amount of CPU and memory that the container can use.

### spec.nodePools[].resources.limits.cpu

`string`

### spec.nodePools[].resources.limits.memory

`string`

### spec.nodePools[].resources.requests

`CpuMemory`

The resource requests for the container.
Specify the minimum amount of CPU and memory that the container is guaranteed.

### spec.nodePools[].resources.requests.cpu

`string`

### spec.nodePools[].resources.requests.memory

`string`

### spec.nodePools[].jvm

`string`

JVM options for the pool's nodes, e.g. "-Xmx1G -Xms1G". Operator
default: 512M heap. Set heap to roughly half the container
memory request.

### spec.nodePools[].diskSize

`string`

Per-node data volume size, e.g. "30Gi". Operator default: 30Gi.
Applies to the PVC storage arm; ignored for emptyDir.

### spec.nodePools[].persistence

`KubernetesOpenSearchPersistence`

Data storage backing. Empty = a PVC per node on the cluster's
default StorageClass (the safe default — data survives pod
loss).

### spec.nodePools[].persistence.pvc

`KubernetesOpenSearchPvc`

PVC-backed storage (the durable default).

### spec.nodePools[].persistence.pvc.storageClass

`string | valueFrom`

StorageClass for the pool's PVCs. Accepts a literal class name
or a reference to a KubernetesStorageClass resource. Empty = the
cluster's default class.

- references: KubernetesStorageClass (`status.outputs.storage_class_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesStorageClass, name: <that resource's name>, fieldPath: status.outputs.storage_class_name}} -- a bare string does not parse

### spec.nodePools[].persistence.emptyDir

`KubernetesOpenSearchEmptyDir`

Node-local emptyDir — data is LOST when the pod leaves the
node. Only for caches and throwaway experiments.

### spec.nodePools[].persistence.emptyDir.sizeLimit

`string`

Optional cap on the emptyDir size, e.g. "10Gi". Empty = no limit.

### spec.nodePools[].additionalConfig

`map<string, string>`

Extra opensearch.yml entries for THIS pool only (wins over the
cluster-level additional_config on key conflicts).

### spec.nodePools[].nodeSelector

`map<string, string>`

Node selector for the pool's pods.

### spec.nodePools[].tolerations

`[]WorkloadToleration`

Tolerations for the pool's pods.

### spec.nodePools[].tolerations[].key

`string`

Taint key to tolerate. Empty key with operator "Exists" tolerates every taint.

### spec.nodePools[].tolerations[].operator

`string`

How key/value match: "Equal" (default — value must match too) or "Exists"
(key presence alone matches).

- rule: Toleration operator must be either "Equal" or "Exists"

### spec.nodePools[].tolerations[].value

`string`

Taint value to match when operator is "Equal".

### spec.nodePools[].tolerations[].effect

`string`

Which taint effect is tolerated: "NoSchedule", "PreferNoSchedule", or
"NoExecute". Empty tolerates all effects for the key.

- rule: Toleration effect must be one of "NoSchedule", "PreferNoSchedule", or "NoExecute"

### spec.nodePools[].tolerations[].tolerationSeconds

`int64` · optional (explicit presence)

For "NoExecute" taints only: how many seconds already-running pods stay bound
after the taint appears. Unset means tolerate forever.

### spec.nodePools[].pdb

`KubernetesOpenSearchPdb`

PodDisruptionBudget for the pool. Declare at most one bound.

- rule: declare min_available or max_unavailable, not both

### spec.nodePools[].pdb.enable

`bool`

Create the PDB.

### spec.nodePools[].pdb.minAvailable

`string`

Minimum pods that must stay available — an integer ("2") or a
percentage ("50%"). Mutually exclusive with max_unavailable.

### spec.nodePools[].pdb.maxUnavailable

`string`

Maximum pods that may be unavailable — an integer ("1") or a
percentage ("25%"). Mutually exclusive with min_available.

### spec.additionalConfig

`map<string, string>`

Extra entries for opensearch.yml, applied to ALL pools (a pool's
own additional_config wins on key conflicts). Values are strings
exactly as opensearch.yml expects them. Cluster-formation,
network and TLS keys are operator-owned — overriding them here
fights the reconciler.

### spec.serviceAnnotations

`map<string, string>`

Annotations for the Services the operator creates (the
cloud-controller injection surface — internal LB annotations and
similar recipes ride here).

### spec.setVmMaxMapCount

`bool` · optional (explicit presence)

Run a privileged init container that raises the node's
vm.max_map_count to the value OpenSearch requires. Default: true —
most distributions ship a lower kernel default and OpenSearch pods
crash-loop without it. Disable only when node kernels are already
tuned (or privileged init containers are forbidden — then tune
nodes out of band).

- default: `true`

### spec.drainDataNodes

`bool`

Drain data from a node before stopping it during rolling
operations. Operator default: false (restarts rely on replica
recovery). Enable for large data nodes where recovering from
replicas is slower than draining.

### spec.pluginsList

`[]string`

OpenSearch plugins to install into EVERY node at STARTUP
(`opensearch-plugin install` runs in the init phase). Plugin
installs download from the internet at pod start unless the
plugin ships in the image — verify a plugin's availability for
the pinned version before declaring it; a failing install
crash-loops the pod. For air-gapped clusters bake plugins into a
custom image instead.

### spec.bootstrap

`KubernetesOpenSearchBootstrap`

Bootstrap pod tuning. The operator runs a temporary bootstrap
node to form the initial cluster quorum; these knobs size it.
Rarely needed — the defaults serve.

### spec.bootstrap.resources

`ContainerResources`

Bootstrap container resources. Empty = operator defaults.

### spec.bootstrap.resources.limits

`CpuMemory`

The resource limits for the container.
Specify the maximum amount of CPU and memory that the container can use.

### spec.bootstrap.resources.limits.cpu

`string`

### spec.bootstrap.resources.limits.memory

`string`

### spec.bootstrap.resources.requests

`CpuMemory`

The resource requests for the container.
Specify the minimum amount of CPU and memory that the container is guaranteed.

### spec.bootstrap.resources.requests.cpu

`string`

### spec.bootstrap.resources.requests.memory

`string`

### spec.bootstrap.jvm

`string`

JVM options for the bootstrap node, e.g. "-Xmx512M -Xms512M".

### spec.bootstrap.additionalConfig

`map<string, string>`

Extra opensearch.yml entries for the bootstrap node only.
Defaults to the cluster-level additional_config.

### spec.security

`KubernetesOpenSearchSecurity`

Security plugin and TLS posture. Empty = the operator generates
a CA and per-layer certificates and bootstraps the security
plugin with an admin user (the recommended default).

### spec.security.transportTls

`KubernetesOpenSearchTlsTransport`

Transport-layer (node-to-node) TLS. Empty = operator-generated
CA and per-node certificates (the default posture).

### spec.security.transportTls.generate

`bool` · optional (explicit presence)

Let the operator generate the CA and node certificates.
Default: true. Set false only when providing certificates via
`secret`/`ca_secret`.

- default: `true`

### spec.security.transportTls.perNode

`bool` · optional (explicit presence)

Issue one certificate per node (hostname-pinned) instead of a
shared certificate. Component default: true (the stronger
posture); the operator's OWN default is a single shared
certificate — the modules always render this field explicitly, so
the component default governs.

- default: `true`

### spec.security.transportTls.secret

`string | valueFrom`

Existing TLS Secret (ca.crt, tls.key, tls.crt) to use when
generate=false. Accepts a literal name or a reference to a
KubernetesCertificate's secret — the cert-manager seam.

- references: KubernetesCertificate (`status.outputs.secret_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesCertificate, name: <that resource's name>, fieldPath: status.outputs.secret_name}} -- a bare string does not parse

### spec.security.transportTls.caSecret

`string | valueFrom`

Separate Secret holding the CA (ca.crt; with generate=true and
this set, the existing CA signs the generated node certs — then
it must also hold ca.key).

- references: KubernetesSecret (`status.outputs.secret_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesSecret, name: <that resource's name>, fieldPath: status.outputs.secret_name}} -- a bare string does not parse

### spec.security.transportTls.nodesDn

`[]string`

Certificate DNs allowed as cluster nodes — REQUIRED when
providing your own certificates (generate=false); the security
plugin rejects inter-node connections from unlisted DNs.

### spec.security.transportTls.adminDn

`[]string`

Certificate DNs granted admin access (securityadmin.sh runs).
Used with provided certificates.

### spec.security.httpTls

`KubernetesOpenSearchTlsHttp`

HTTP-layer (client-facing) TLS. Empty = operator-generated
certificates. Clients connect with https and the CA from the
cluster's CA secret.

### spec.security.httpTls.generate

`bool` · optional (explicit presence)

Let the operator generate the HTTP certificates. Default: true.

- default: `true`

### spec.security.httpTls.secret

`string | valueFrom`

Existing TLS Secret (ca.crt, tls.key, tls.crt) when
generate=false. Accepts a literal name or a KubernetesCertificate
reference — the cert-manager seam.

- references: KubernetesCertificate (`status.outputs.secret_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesCertificate, name: <that resource's name>, fieldPath: status.outputs.secret_name}} -- a bare string does not parse

### spec.security.config

`KubernetesOpenSearchSecurityConfig`

Bring-your-own security-plugin configuration. Only for advanced
setups that replace the operator's bootstrap (custom realms,
OIDC, fine-grained roles seeded at install). When set, ALL THREE
secrets are typically required by the operator: the security
config itself, an admin client certificate, and admin
credentials for the operator's own API access.

### spec.security.config.securityConfigSecret

`string | valueFrom`

Secret holding the security plugin YAML files (config.yml,
internal_users.yml, roles.yml, ...).

- references: KubernetesSecret (`status.outputs.secret_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesSecret, name: <that resource's name>, fieldPath: status.outputs.secret_name}} -- a bare string does not parse

### spec.security.config.adminSecret

`string | valueFrom`

TLS Secret with an admin client certificate (tls.key, tls.crt,
ca.crt) for securityadmin.sh. Required when transport
certificates are user-provided.

- references: KubernetesSecret (`status.outputs.secret_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesSecret, name: <that resource's name>, fieldPath: status.outputs.secret_name}} -- a bare string does not parse

### spec.security.config.adminCredentialsSecret

`string | valueFrom`

Secret with `username`/`password` fields the OPERATOR uses for
its own API calls (drain coordination, health checks). Required
with a custom security config — the operator's bootstrapped
credentials do not exist in that case.

- references: KubernetesSecret (`status.outputs.secret_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesSecret, name: <that resource's name>, fieldPath: status.outputs.secret_name}} -- a bare string does not parse

### spec.dashboards

`KubernetesOpenSearchDashboards`

OpenSearch Dashboards — the web console. Disabled unless
declared with enabled=true.

### spec.dashboards.enabled

`bool`

Deploy Dashboards alongside the cluster.

### spec.dashboards.replicas

`int32` · optional (explicit presence)

Dashboards replica count. Default: 1.

- default: `1`
- rule: {"int32":{"gte":1}}

### spec.dashboards.version

`string`

Dashboards version. Empty = the cluster's `version` (keep them
aligned — Dashboards refuses mismatched clusters).

### spec.dashboards.resources

`ContainerResources`

Dashboards container resources. Empty = operator defaults.

### spec.dashboards.resources.limits

`CpuMemory`

The resource limits for the container.
Specify the maximum amount of CPU and memory that the container can use.

### spec.dashboards.resources.limits.cpu

`string`

### spec.dashboards.resources.limits.memory

`string`

### spec.dashboards.resources.requests

`CpuMemory`

The resource requests for the container.
Specify the minimum amount of CPU and memory that the container is guaranteed.

### spec.dashboards.resources.requests.cpu

`string`

### spec.dashboards.resources.requests.memory

`string`

### spec.dashboards.tls

`KubernetesOpenSearchDashboardsTls`

Serve Dashboards over HTTPS with an operator-generated (or
provided) certificate.

### spec.dashboards.tls.enable

`bool`

Serve HTTPS.

### spec.dashboards.tls.generate

`bool` · optional (explicit presence)

Generate the certificate. Default true when enabled; set false
to provide one via `secret`.

- default: `true`

### spec.dashboards.tls.secret

`string | valueFrom`

Existing TLS Secret when generate=false. Accepts a literal name
or a KubernetesCertificate reference.

- references: KubernetesCertificate (`status.outputs.secret_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesCertificate, name: <that resource's name>, fieldPath: status.outputs.secret_name}} -- a bare string does not parse

### spec.dashboards.basePath

`string`

Base path when running behind a path-rewriting reverse proxy
(e.g. "/dashboards").

### spec.dashboards.additionalConfig

`map<string, string>`

Extra entries for opensearch_dashboards.yml.

### spec.dashboards.opensearchCredentialsSecret

`string | valueFrom`

Secret with `username`/`password` Dashboards uses to reach the
cluster — required ONLY with a custom security config (the
operator wires its bootstrapped credentials otherwise).

- references: KubernetesSecret (`status.outputs.secret_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesSecret, name: <that resource's name>, fieldPath: status.outputs.secret_name}} -- a bare string does not parse

### spec.dashboards.service

`KubernetesOpenSearchDashboardsService`

Dashboards Service exposure. Default ClusterIP — prefer
composing exposure from KubernetesIngress / Gateway API kinds.

### spec.dashboards.service.type

`string` · optional (explicit presence)

Service type: ClusterIP (default), NodePort, or LoadBalancer.

- default: `ClusterIP`
- rule: service type must be ClusterIP, NodePort, or LoadBalancer

### spec.dashboards.service.loadBalancerSourceRanges

`[]string`

CIDR allow-list for a LoadBalancer service.

### spec.dashboards.pluginsList

`[]string`

Dashboards plugins to install at startup (same install-time
caveats as the cluster's plugins_list).

### spec.monitoring

`KubernetesOpenSearchMonitoring`

Prometheus monitoring via the Aiven prometheus-exporter plugin.
Requires the Prometheus Operator's ServiceMonitor CRD on the
cluster (compose with the metrics stack) — enabling it without
that CRD fails reconciliation.

### spec.monitoring.enabled

`bool`

Enable the monitoring plugin + ServiceMonitor. Requires the
Prometheus Operator CRDs on the cluster.

### spec.monitoring.scrapeInterval

`string`

Scrape interval, e.g. "30s". Empty = operator default.

### spec.monitoring.monitoringUserSecret

`string | valueFrom`

Secret with `username`/`password` for the metrics scrape when a
custom security config is in play. Empty = the operator wires
its bootstrapped monitoring access.

- references: KubernetesSecret (`status.outputs.secret_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesSecret, name: <that resource's name>, fieldPath: status.outputs.secret_name}} -- a bare string does not parse

### spec.monitoring.pluginUrl

`string`

Override the download URL of the aiven prometheus-exporter
plugin (air-gap path). Empty = the operator derives it from the
cluster version.

### spec.keystore

`[]KubernetesOpenSearchKeystoreValue`

Entries to load into the OpenSearch keystore before startup —
the safe home for snapshot-repository credentials and other
secure settings. Each entry references an existing Secret; keys
map 1:1 into the keystore unless key_mappings renames them.

### spec.keystore[].secret

`string | valueFrom` · required

Existing Secret with the keystore entries.

- references: KubernetesSecret (`status.outputs.secret_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesSecret, name: <that resource's name>, fieldPath: status.outputs.secret_name}} -- a bare string does not parse

### spec.keystore[].keyMappings

`map<string, string>`

Rename Secret keys into keystore keys (secret key -> keystore
key). Empty = keys map 1:1.

### spec.snapshotRepositories

`[]KubernetesOpenSearchSnapshotRepo`

Snapshot repositories to register on the cluster (the backup
surface). `type` and `settings` pass through to the OpenSearch
snapshot API — e.g. type "s3" with settings {bucket: ...,
region: ...} requires the repository-s3 plugin (declare it in
plugins_list) and credentials in the keystore (or an
instance/workload identity on the nodes).

### spec.snapshotRepositories[].name

`string` · required

Repository name (referenced by snapshot/restore API calls).

- rule: {"required":true}

### spec.snapshotRepositories[].type

`string` · required

Repository type: "s3", "gcs", "azure", "fs", ... — the matching
repository plugin must be on the nodes (plugins_list or a custom
image).

- rule: {"required":true}

### spec.snapshotRepositories[].settings

`map<string, string>`

Type-specific settings exactly as the OpenSearch snapshot API
expects them (bucket, region, base_path, endpoint, ...).
Credentials belong in the KEYSTORE (or node identity), never
here.

### spec.additionalVolumes

`[]KubernetesOpenSearchAdditionalVolume`

Additional volumes projected into ALL cluster pods (certificates,
custom configuration files, plugin config). Curated sources:
Secret or ConfigMap.

- rule: declare exactly one of secret_name or config_map_name

### spec.additionalVolumes[].name

`string` · required

Volume name.

- rule: {"required":true}

### spec.additionalVolumes[].path

`string` · required

Mount path inside the containers.

- rule: {"required":true}

### spec.additionalVolumes[].subPath

`string`

Mount only this item of the source instead of the whole volume.

### spec.additionalVolumes[].secretName

`string`

Name of the Secret to project (a reference, never secret
material). Exactly one of secret_name / config_map_name.

### spec.additionalVolumes[].configMapName

`string`

Name of the ConfigMap to project.

### spec.additionalVolumes[].restartPods

`bool`

Restart cluster pods when the projected content changes.

### spec.image

`ContainerImage`

Override the OpenSearch node image (air-gap / custom-plugin
path). Empty = opensearchproject/opensearch at `version`. The
image's pull_secret_name, when set, joins `image_pull_secrets` in
the rendered cluster (deduplicated) — a private image travels
with its own pull secret.

### spec.image.repo

`string`

The repository of the image (e.g., "gcr.io/project/image").

### spec.image.tag

`string`

The tag of the image (e.g., "latest" or "1.0.0").

### spec.image.pullSecretName

`string`

The name of the image pull secret for private image repositories.

### spec.imagePullSecrets

`[]string`

Names of image-pull secrets (in the cluster namespace) for
pulling images from a private registry.

## Outputs

Reference an output from another manifest as `valueFrom: {kind: KubernetesOpenSearch, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.namespace` | `string` | namespace the cluster runs in. |
| `status.outputs.cluster_name` | `string` | name of the OpenSearchCluster resource (= metadata.name). |
| `status.outputs.service_name` | `string` | name of the cluster's main Service (all nodes; = metadata.name — the operator names the Service after the cluster's serviceName, which the modules pin to the resource name). |
| `status.outputs.http_endpoint` | `string` | in-cluster HTTP API endpoint, e.g. https://main.search.svc.cluster.local:9200 (https because the operator serves TLS on the HTTP layer by default). |
| `status.outputs.admin_credentials_secret_name` | `string` | name of the operator-generated Secret holding the admin credentials (fields username/password), e.g. main-admin-password. Empty when a custom security config replaces the operator bootstrap. |
| `status.outputs.dashboards_service_name` | `string` | name of the Dashboards Service (e.g. main-dashboards). Empty when dashboards are not enabled. |
| `status.outputs.dashboards_endpoint` | `string` | in-cluster Dashboards endpoint, e.g. http://main-dashboards.search.svc.cluster.local:5601. Empty when dashboards are not enabled. |
| `status.outputs.port_forward_command` | `string` | command to port-forward the cluster's HTTP API to a developer laptop, e.g. kubectl port-forward svc/main -n search 9200:9200 |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.namespace` | KubernetesNamespace | `spec.name` |
| `spec.nodePools[].persistence.pvc.storageClass` | KubernetesStorageClass | `status.outputs.storage_class_name` |
| `spec.security.transportTls.secret` | KubernetesCertificate | `status.outputs.secret_name` |
| `spec.security.transportTls.caSecret` | KubernetesSecret | `status.outputs.secret_name` |
| `spec.security.httpTls.secret` | KubernetesCertificate | `status.outputs.secret_name` |
| `spec.security.config.securityConfigSecret` | KubernetesSecret | `status.outputs.secret_name` |
| `spec.security.config.adminSecret` | KubernetesSecret | `status.outputs.secret_name` |
| `spec.security.config.adminCredentialsSecret` | KubernetesSecret | `status.outputs.secret_name` |
| `spec.dashboards.tls.secret` | KubernetesCertificate | `status.outputs.secret_name` |
| `spec.dashboards.opensearchCredentialsSecret` | KubernetesSecret | `status.outputs.secret_name` |
| `spec.monitoring.monitoringUserSecret` | KubernetesSecret | `status.outputs.secret_name` |
| `spec.keystore[].secret` | KubernetesSecret | `status.outputs.secret_name` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| KubernetesAirflow | `spec.logging.elasticsearch.host` | `status.outputs.service_name` |
| KubernetesAirflow | `spec.logging.elasticsearch.passwordSecret.secretName` | `status.outputs.admin_credentials_secret_name` |
| KubernetesAirflow | `spec.logging.opensearch.host` | `status.outputs.service_name` |
| KubernetesAirflow | `spec.logging.opensearch.passwordSecret.secretName` | `status.outputs.admin_credentials_secret_name` |

## See Also

- [Overview](../README.md)
