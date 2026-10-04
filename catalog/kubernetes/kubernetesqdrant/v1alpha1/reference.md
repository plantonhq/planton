# KubernetesQdrant

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `kubernetes.planton.dev/v1alpha1`

**KubernetesQdrantSpec** deploys Qdrant — the catalog's vector
database (Apache-2.0), the search engine behind RAG, semantic
search and agent-memory architectures — from the official `qdrant`
Helm chart (https://qdrant.github.io/qdrant-helm).

GRAIN: one release = one Qdrant cluster (a StatefulSet).
Distributed mode is always on (the chart's default): pod 0
bootstraps the Raft consensus and every further replica joins over
the p2p port — scaling is a `replicas` change. Collection-level
properties (shards, replication factor, quantization) are DATA
operations declared per collection through the Qdrant API, not
deployment configuration.

SECURITY: API keys are OFF upstream and off by default here —
declare `api_key` (read-write) and optionally `read_only_api_key`
for anything beyond a private dev cluster. Keys are materialized
as Secrets and wired through the chart's existing-Secret contract;
key material never lands in rendered Helm values. `tls` enables
HTTPS/gRPC-TLS on the client listeners from a certificate Secret —
the cert-manager seam.

STORAGE: vectors and the write-ahead log live on the storage
volume. If you use snapshots (backup or shard transfers), declare
the separate `snapshots` volume — upstream recommends sizing it
like the main volume so a large snapshot cannot crash the node.

EXPOSURE: the service stays ClusterIP; expose via first-class
kinds (KubernetesIngress, Gateway API kinds) over the exported
service handle. REST is 6333, gRPC 6334; SDKs default to gRPC.

The typed fields below cover the chart's meaningful configuration
surface; `helm_values` remains as the escape hatch for chart
values beyond them (merged last, Helm `-f` semantics, identical on
both engines) — engine config under `config:` (collection
defaults, optimizer/WAL tuning, p2p TLS), probes, PDB, sidecars,
service-type changes — a safety valve, never the primary
interface.

## Example

```yaml
# Full-surface offline-proof manifest: exercises a 3-node Raft cluster, a
# generated read-write API key plus an existing-Secret read-only key, the
# certificate-Secret TLS seam, dual persistence (storage + a separate
# snapshots volume on a cold class), scheduling, the ServiceMonitor toggle
# (off), an image override, and an engine-config escape-hatch entry — so
# the offline tofu plan and pulumi preview proofs cover the full typed
# surface. Placeholder values; never applied to a real cluster.
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesQdrant
metadata:
  name: qdrant-hack
spec:
  namespace:
    value: qdrant-hack
  createNamespace: true
  chartVersion: 1.18.2
  replicas: 3
  resources:
    requests:
      cpu: 500m
      memory: 2Gi
    limits:
      cpu: "2"
      memory: 8Gi
  storage:
    size: 50Gi
    storageClass:
      value: fast-ssd
  snapshots:
    size: 50Gi
    storageClass:
      value: cold-storage
  apiKey:
    generate: true
  readOnlyApiKey:
    existingSecret:
      name: qdrant-hack-ro-key
      key: api-key
  tls:
    secret:
      value: qdrant-hack-tls
  scheduling:
    nodeSelector:
      kubernetes.io/os: linux
    tolerations:
      - key: dedicated
        operator: Equal
        value: vector
        effect: NoSchedule
    podAntiAffinity: true
    priorityClassName: system-cluster-critical
  serviceMonitorEnabled: false
  image:
    repository: docker.io/qdrant/qdrant
    useUnprivileged: true
  helmValues: |
    config:
      storage:
        performance:
          max_search_threads: 4
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.namespace` | `string \| valueFrom` | yes |  | KubernetesNamespace (`spec.name`) |
| `spec.createNamespace` | `bool` |  |  |  |
| `spec.chartVersion` | `string` |  | `1.18.2` |  |
| `spec.replicas` | `int32` |  | `1` |  |
| `spec.resources` | `ContainerResources` |  |  |  |
| `spec.resources.limits` | `CpuMemory` |  |  |  |
| `spec.resources.limits.cpu` | `string` |  |  |  |
| `spec.resources.limits.memory` | `string` |  |  |  |
| `spec.resources.requests` | `CpuMemory` |  |  |  |
| `spec.resources.requests.cpu` | `string` |  |  |  |
| `spec.resources.requests.memory` | `string` |  |  |  |
| `spec.storage` | `KubernetesQdrantDataVolume` |  |  |  |
| `spec.storage.size` | `string` |  | `10Gi` |  |
| `spec.storage.storageClass` | `string \| valueFrom` |  |  | KubernetesStorageClass (`status.outputs.storage_class_name`) |
| `spec.snapshots` | `KubernetesQdrantSnapshotsVolume` |  |  |  |
| `spec.snapshots.size` | `string` |  | `10Gi` |  |
| `spec.snapshots.storageClass` | `string \| valueFrom` |  |  | KubernetesStorageClass (`status.outputs.storage_class_name`) |
| `spec.apiKey` | `KubernetesQdrantApiKey` |  |  |  |
| `spec.apiKey.generate` | `bool` |  |  |  |
| `spec.apiKey.existingSecret` | `KubernetesQdrantSecretKeyRef` |  |  |  |
| `spec.apiKey.existingSecret.name` | `string` | yes |  |  |
| `spec.apiKey.existingSecret.key` | `string` | yes |  |  |
| `spec.readOnlyApiKey` | `KubernetesQdrantApiKey` |  |  |  |
| `spec.readOnlyApiKey.generate` | `bool` |  |  |  |
| `spec.readOnlyApiKey.existingSecret` | `KubernetesQdrantSecretKeyRef` |  |  |  |
| `spec.readOnlyApiKey.existingSecret.name` | `string` | yes |  |  |
| `spec.readOnlyApiKey.existingSecret.key` | `string` | yes |  |  |
| `spec.tls` | `KubernetesQdrantTls` |  |  |  |
| `spec.tls.secret` | `string \| valueFrom` | yes |  | KubernetesCertificate (`status.outputs.secret_name`) |
| `spec.scheduling` | `KubernetesQdrantScheduling` |  |  |  |
| `spec.scheduling.nodeSelector` | `map<string, string>` |  |  |  |
| `spec.scheduling.tolerations` | `[]WorkloadToleration` |  |  |  |
| `spec.scheduling.tolerations[].key` | `string` |  |  |  |
| `spec.scheduling.tolerations[].operator` | `string` |  |  |  |
| `spec.scheduling.tolerations[].value` | `string` |  |  |  |
| `spec.scheduling.tolerations[].effect` | `string` |  |  |  |
| `spec.scheduling.tolerations[].tolerationSeconds` | `int64` |  |  |  |
| `spec.scheduling.podAntiAffinity` | `bool` |  |  |  |
| `spec.scheduling.priorityClassName` | `string` |  |  |  |
| `spec.serviceMonitorEnabled` | `bool` |  |  |  |
| `spec.image` | `KubernetesQdrantImage` |  |  |  |
| `spec.image.repository` | `string` |  |  |  |
| `spec.image.tag` | `string` |  |  |  |
| `spec.image.useUnprivileged` | `bool` |  |  |  |
| `spec.helmValues` | `string` |  |  |  |

## Field Details

### spec.namespace

`string | valueFrom` · required

Namespace to install into. Accepts a literal namespace name or a
reference to a KubernetesNamespace resource.

- references: KubernetesNamespace (`spec.name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesNamespace, name: <that resource's name>, fieldPath: spec.name}} -- a bare string does not parse

### spec.createNamespace

`bool`

When true, the namespace is created (with the standard Planton
governance labels) before installing and deleted with the
resource. When false, the namespace must already exist.

### spec.chartVersion

`string` · optional (explicit presence)

Helm chart version to install (e.g. "1.18.2" — the chart version
tracks the Qdrant release it ships). Versions must exist as
SERVED charts in the repository index
(https://qdrant.github.io/qdrant-helm).

- default: `1.18.2`

### spec.replicas

`int32` · optional (explicit presence)

Cluster size. 1 (default) is a single node; any higher count
forms a Raft cluster over the p2p port. Collections replicate
per their own replication_factor — adding pods alone does not
copy data.

- default: `1`
- rule: {"int32":{"gte":1}}

### spec.resources

`ContainerResources`

Container resources. Size memory for the vectors you plan to
hold — Qdrant keeps hot segments and indexes in RAM.

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

### spec.storage

`KubernetesQdrantDataVolume`

Storage volume PER pod (vectors, payloads, WAL). Empty = 10Gi on
the cluster's default StorageClass.

### spec.storage.size

`string` · optional (explicit presence)

Volume size as a Kubernetes quantity (e.g. "10Gi").

- default: `10Gi`
- rule: size must be a Kubernetes quantity like '10Gi' or '500Mi'

### spec.storage.storageClass

`string | valueFrom`

StorageClass for the PVC. Accepts a literal class name or a
reference to a KubernetesStorageClass resource. Empty = the
cluster's default class.

- references: KubernetesStorageClass (`status.outputs.storage_class_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesStorageClass, name: <that resource's name>, fieldPath: status.outputs.storage_class_name}} -- a bare string does not parse

### spec.snapshots

`KubernetesQdrantSnapshotsVolume`

Separate snapshots volume PER pod. Declare when using snapshots
or snapshot shard transfers — upstream recommends sizing it like
`storage` so a big snapshot cannot fill the data volume. Empty =
snapshots land on the storage volume.

### spec.snapshots.size

`string` · optional (explicit presence)

Volume size as a Kubernetes quantity. Size like the storage
volume (upstream guidance). Empty = "10Gi".

- default: `10Gi`
- rule: size must be a Kubernetes quantity like '10Gi' or '500Mi'

### spec.snapshots.storageClass

`string | valueFrom`

StorageClass for the snapshots PVC (cold storage is a good fit).
Accepts a literal class name or a KubernetesStorageClass
reference. Empty = the cluster's default class.

- references: KubernetesStorageClass (`status.outputs.storage_class_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesStorageClass, name: <that resource's name>, fieldPath: status.outputs.storage_class_name}} -- a bare string does not parse

### spec.apiKey

`KubernetesQdrantApiKey`

Read-write API key. Empty = the listeners accept unauthenticated
requests (upstream default — private dev clusters only).

### spec.apiKey.generate

`bool`

Generate a random key. The chart generates it once at first
install (stable across upgrades) and keeps it in its own
`<name>-apikey` Secret — key `api-key` (read-write) /
`read-only-api-key`. The key never appears in rendered Helm
values; the Secret name lands in the outputs.

### spec.apiKey.existingSecret

`KubernetesQdrantSecretKeyRef`

Read the key from an existing Secret. The Secret must exist
BEFORE the install — the chart resolves it at template time.

### spec.apiKey.existingSecret.name

`string` · required

Secret name (a reference to an existing Kubernetes Secret, never
secret material).

- rule: {"required":true}

### spec.apiKey.existingSecret.key

`string` · required

Key within the Secret.

- rule: {"required":true}

### spec.readOnlyApiKey

`KubernetesQdrantApiKey`

Read-only API key — hand this one to query-only consumers.
Requires `api_key` (an unauthenticated cluster with a read-only
key protects nothing — writes would stay open).

### spec.readOnlyApiKey.generate

`bool`

Generate a random key. The chart generates it once at first
install (stable across upgrades) and keeps it in its own
`<name>-apikey` Secret — key `api-key` (read-write) /
`read-only-api-key`. The key never appears in rendered Helm
values; the Secret name lands in the outputs.

### spec.readOnlyApiKey.existingSecret

`KubernetesQdrantSecretKeyRef`

Read the key from an existing Secret. The Secret must exist
BEFORE the install — the chart resolves it at template time.

### spec.readOnlyApiKey.existingSecret.name

`string` · required

Secret name (a reference to an existing Kubernetes Secret, never
secret material).

- rule: {"required":true}

### spec.readOnlyApiKey.existingSecret.key

`string` · required

Key within the Secret.

- rule: {"required":true}

### spec.tls

`KubernetesQdrantTls`

TLS on the client listeners (REST 6333 + gRPC 6334) from a
certificate Secret — the cert-manager seam. Empty = plaintext
in-cluster (compose TLS at the exposure layer). Inter-node p2p
TLS is a separate upstream surface (`config.cluster.p2p`) and
rides `helm_values`.

### spec.tls.secret

`string | valueFrom` · required

Existing TLS Secret (tls.crt + tls.key — the standard Kubernetes
TLS Secret shape cert-manager produces). Accepts a literal name
or a KubernetesCertificate reference. The modules mount it and
point the engine's `service.enable_tls` + cert paths at it.

- references: KubernetesCertificate (`status.outputs.secret_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesCertificate, name: <that resource's name>, fieldPath: status.outputs.secret_name}} -- a bare string does not parse

### spec.scheduling

`KubernetesQdrantScheduling`

Scheduling for the Qdrant pods.

### spec.scheduling.nodeSelector

`map<string, string>`

Node selector for the Qdrant pods.

### spec.scheduling.tolerations

`[]WorkloadToleration`

Tolerations for the Qdrant pods.

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

### spec.scheduling.podAntiAffinity

`bool`

Spread pods across nodes with required pod anti-affinity —
meaningful from 2 replicas up (a node loss then takes one
member, not the quorum). Chart default: none.

### spec.scheduling.priorityClassName

`string`

Priority class name for the Qdrant pods.

### spec.serviceMonitorEnabled

`bool`

Create a ServiceMonitor for Prometheus scraping of /metrics
(requires the Prometheus Operator CRDs). Chart default: false.

### spec.image

`KubernetesQdrantImage`

Override the Qdrant image (air-gap path). Empty = the chart's
official `qdrant/qdrant` at the chart's appVersion.

### spec.image.repository

`string`

Image repository including registry, e.g.
"my.registry.com/qdrant/qdrant". Empty = "docker.io/qdrant/qdrant".

### spec.image.tag

`string`

Image tag. Empty = the chart's appVersion for the pinned
chart_version.

### spec.image.useUnprivileged

`bool`

Run the unprivileged image variant for restricted Pod Security
Standards environments — the chart appends `-unprivileged` to the
image TAG (e.g. `qdrant/qdrant:v1.18.2-unprivileged`) and skips
the root-owned volume-ownership init container.

### spec.helmValues

`string`

Escape hatch: additional chart values as a YAML document, merged
LAST over everything the typed fields render (Helm `-f`
semantics, identical on both engines). For the chart surface
beyond the typed fields (the `config:` engine document, probes,
PDB, sidecars, additional volumes, service tweaks, ...) — never
the substitute for them. Do not put secrets here; key material
belongs in `api_key`/`read_only_api_key`.

## Validation Rules

- `spec.read_only_key_requires_api_key`: read_only_api_key requires api_key — with no read-write key the cluster is unauthenticated and a read-only key protects nothing

## Outputs

Reference an output from another manifest as `valueFrom: {kind: KubernetesQdrant, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.namespace` | `string` | namespace the cluster runs in. |
| `status.outputs.release_name` | `string` | Helm release name (= metadata.name). |
| `status.outputs.service_name` | `string` | name of the main Qdrant Service (http 6333, grpc 6334). |
| `status.outputs.http_endpoint` | `string` | in-cluster REST endpoint, e.g. http://main.vector.svc.cluster.local:6333 |
| `status.outputs.grpc_endpoint` | `string` | in-cluster gRPC endpoint SDKs default to, e.g. main.vector.svc.cluster.local:6334 |
| `status.outputs.api_key_secret_name` | `string` | name of the chart-owned Secret holding the API key material — `<name>-apikey`, key `api-key` for the read-write key (populated for the generate arm and the existing-secret arm alike). Empty when unauthenticated. |
| `status.outputs.read_only_api_key_secret_name` | `string` | name of the Secret holding the read-only API key — the same chart-owned `<name>-apikey`, key `read-only-api-key`. Empty when not declared. |
| `status.outputs.port_forward_command` | `string` | command to port-forward the REST port to a developer laptop, e.g. kubectl port-forward svc/main -n vector 6333:6333 |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.namespace` | KubernetesNamespace | `spec.name` |
| `spec.storage.storageClass` | KubernetesStorageClass | `status.outputs.storage_class_name` |
| `spec.snapshots.storageClass` | KubernetesStorageClass | `status.outputs.storage_class_name` |
| `spec.tls.secret` | KubernetesCertificate | `status.outputs.secret_name` |

## See Also

- [Overview](../README.md)
