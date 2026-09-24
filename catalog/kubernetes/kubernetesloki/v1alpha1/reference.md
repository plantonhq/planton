# KubernetesLoki

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `kubernetes.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

**KubernetesLokiSpec** deploys Grafana Loki — the log-aggregation
system that indexes log LABELS (not content) and stores compressed log
chunks in object storage — from the official `loki` Helm chart
(https://grafana-community.github.io/helm-charts).

HOW LOGS GET IN: Loki stores logs; something must ship them. Deploy a
KubernetesOtelCollector in daemonset mode with the cluster-logs
pipeline (its presets carry it) pointed at this component's
`gateway_endpoint` — or push from any OTLP/Loki-HTTP client. Grafana
reads them back: point a KubernetesGrafana datasource of type `loki`
at the same endpoint.

MODES: `monolithic` (default) runs every Loki target in one
StatefulSet — right for single-node clusters, dev environments and
small production volumes. `simple_scalable` splits into write/read/
backend tiers that scale independently and REQUIRES object storage.
The chart's microservices ("Distributed") mode and its transitional
migration modes are deliberately not modeled — by the time a
deployment needs per-component microservices, it deserves a dedicated
operations posture, and mode migrations are operational verbs.

STORAGE DOCTRINE (mirrors the chart's own validation): `filesystem`
keeps chunks on a PersistentVolume — honest ONLY for a single
monolithic replica; more than one replica, or any simple_scalable
tier, REQUIRES an object-storage backend (s3/gcs/azure). The
s3-compatible arm (endpoint + path-style) composes with an in-cluster
KubernetesSeaweedFs. The chart's bundled MinIO subchart is deprecated
by the chart itself and is never enabled by this component.

SCHEMA: Loki requires a `schema_config` naming the index schema and
its start date — upstream makes every user hand-author it. This
component derives it (TSDB, schema v13, the object store matching
your storage backend) so a new install never writes one. The
`schema_from_date` override exists solely for IMPORTING clusters
whose existing schema started on a real date.

TENANCY: single-tenant by default — pushes and queries need no
X-Scope-OrgID header, so the composed Grafana datasource and
collector pipelines work with one line of wiring. This deliberately
diverges from the chart's multi-tenant-on default, where every
client must send a tenant header. Enable `multi_tenancy` for tenant
isolation: every client then sends its X-Scope-OrgID header, and the
gateway gates access with HTTP basic auth (it authenticates —
clients still declare their own tenant header).

NAMING: keep the resource name at 40 characters or fewer. The chart
builds child names like `<name>-backend-headless` and truncates the
COMPOSED name at 63 characters — a longer resource name would corrupt
the naming contract the outputs promise. The modules pin the chart's
fullname to the resource name and fail loudly over the budget.

EXPOSURE: everything stays ClusterIP; the nginx gateway is the one
front door (push + query). Expose it via first-class kinds
(KubernetesIngress, Gateway API kinds) over the exported handles.

The typed fields cover the chart's meaningful surface; `helm_values`
remains the escape hatch (merged last, Helm `-f` semantics, identical
on both engines) for the long tail — per-component overrides, bloom
filters, ruler storage tuning, zone-aware rollouts — a safety valve,
never the primary interface.

## Example

```yaml
# Full-surface offline-proof manifest: simple-scalable tiers on an
# S3-compatible object store (credentials via Secret + env expansion),
# derived-schema override, retention with the compactor, ingestion limits,
# the gateway's multi-tenant basic-auth path (bcrypt hashes), tuned caches,
# the ruler firing at an Alertmanager, ServiceMonitor, a private-mirror
# image registry with a pull secret, scheduling, and an escape-hatch entry
# — so the offline tofu plan and pulumi preview proofs cover the full typed
# surface incl. the split/combined image forms. Placeholder values; never
# applied to a real cluster.
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesLoki
metadata:
  name: loki-hack
spec:
  namespace:
    value: loki-hack
  createNamespace: true
  chartVersion: 18.5.4
  simpleScalable:
    writeReplicas: 3
    readReplicas: 3
    backendReplicas: 3
    diskSize: 20Gi
    storageClass:
      value: gp3
    resources:
      requests:
        cpu: 100m
        memory: 256Mi
      limits:
        cpu: "1"
        memory: 1Gi
  storage:
    s3:
      bucket: loki-chunks
      rulerBucket: loki-ruler
      endpoint: http://objects-filer.storage.svc.cluster.local:8333
      forcePathStyle: true
      insecure: true
      credentials:
        accessKeyIdSecret:
          name: loki-s3
          key: access-key-id
        secretAccessKeySecret:
          name: loki-s3
          key: secret-access-key
  retentionPeriod: 744h
  limits:
    ingestionRateMb: 8
    ingestionBurstSizeMb: 12
    maxGlobalStreamsPerUser: 10000
    maxQueryLookback: 720h
  multiTenancy:
    enabled: true
    tenants:
      - name: team-a
        passwordHash: "$2y$10$7O40CaY1yz7fu9O24k2/u.ct/wELYHRBsn25v/7AyuQ8E8hrLqpva"
  gateway:
    replicas: 2
    resources:
      requests:
        cpu: 50m
        memory: 64Mi
  caching:
    chunksCacheMemoryMb: 256
    resultsCacheMemoryMb: 128
  ruler:
    enabled: true
    alertmanagerUrl:
      value: http://monitoring-kube-prometheus-alertmanager.observability.svc.cluster.local:9093
  serviceMonitorEnabled: true
  imageRegistry: mirror.example.com
  imagePullSecrets:
    - mirror-pull
  scheduling:
    nodeSelector:
      kubernetes.io/os: linux
    tolerations:
      - key: dedicated
        operator: Equal
        value: observability
        effect: NoSchedule
    priorityClassName: system-cluster-critical
  helmValues: |
    loki:
      podLabels:
        team: observability
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.namespace` | `string \| valueFrom` | yes |  | KubernetesNamespace (`spec.name`) |
| `spec.createNamespace` | `bool` |  |  |  |
| `spec.chartVersion` | `string` |  | `18.5.4` |  |
| `spec.monolithic` | `KubernetesLokiMonolithic` |  |  |  |
| `spec.monolithic.replicas` | `int32` |  | `1` |  |
| `spec.monolithic.diskSize` | `string` |  | `10Gi` |  |
| `spec.monolithic.storageClass` | `string \| valueFrom` |  |  | KubernetesStorageClass (`status.outputs.storage_class_name`) |
| `spec.monolithic.resources` | `ContainerResources` |  |  |  |
| `spec.monolithic.resources.limits` | `CpuMemory` |  |  |  |
| `spec.monolithic.resources.limits.cpu` | `string` |  |  |  |
| `spec.monolithic.resources.limits.memory` | `string` |  |  |  |
| `spec.monolithic.resources.requests` | `CpuMemory` |  |  |  |
| `spec.monolithic.resources.requests.cpu` | `string` |  |  |  |
| `spec.monolithic.resources.requests.memory` | `string` |  |  |  |
| `spec.simpleScalable` | `KubernetesLokiSimpleScalable` |  |  |  |
| `spec.simpleScalable.writeReplicas` | `int32` |  | `3` |  |
| `spec.simpleScalable.readReplicas` | `int32` |  | `3` |  |
| `spec.simpleScalable.backendReplicas` | `int32` |  | `3` |  |
| `spec.simpleScalable.diskSize` | `string` |  | `10Gi` |  |
| `spec.simpleScalable.storageClass` | `string \| valueFrom` |  |  | KubernetesStorageClass (`status.outputs.storage_class_name`) |
| `spec.simpleScalable.resources` | `ContainerResources` |  |  |  |
| `spec.simpleScalable.resources.limits` | `CpuMemory` |  |  |  |
| `spec.simpleScalable.resources.limits.cpu` | `string` |  |  |  |
| `spec.simpleScalable.resources.limits.memory` | `string` |  |  |  |
| `spec.simpleScalable.resources.requests` | `CpuMemory` |  |  |  |
| `spec.simpleScalable.resources.requests.cpu` | `string` |  |  |  |
| `spec.simpleScalable.resources.requests.memory` | `string` |  |  |  |
| `spec.storage` | `KubernetesLokiStorage` |  |  |  |
| `spec.storage.filesystem` | `KubernetesLokiFilesystemStorage` |  |  |  |
| `spec.storage.s3` | `KubernetesLokiS3Storage` |  |  |  |
| `spec.storage.s3.bucket` | `string` | yes |  |  |
| `spec.storage.s3.rulerBucket` | `string` |  |  |  |
| `spec.storage.s3.region` | `string` |  |  |  |
| `spec.storage.s3.endpoint` | `string` |  |  |  |
| `spec.storage.s3.forcePathStyle` | `bool` |  |  |  |
| `spec.storage.s3.insecure` | `bool` |  |  |  |
| `spec.storage.s3.credentials` | `KubernetesLokiObjectStoreCredentials` |  |  |  |
| `spec.storage.s3.credentials.accessKeyIdSecret` | `KubernetesLokiSecretKeyRef` | yes |  |  |
| `spec.storage.s3.credentials.accessKeyIdSecret.name` | `string` | yes |  |  |
| `spec.storage.s3.credentials.accessKeyIdSecret.key` | `string` | yes |  |  |
| `spec.storage.s3.credentials.secretAccessKeySecret` | `KubernetesLokiSecretKeyRef` | yes |  |  |
| `spec.storage.s3.credentials.secretAccessKeySecret.name` | `string` | yes |  |  |
| `spec.storage.s3.credentials.secretAccessKeySecret.key` | `string` | yes |  |  |
| `spec.storage.gcs` | `KubernetesLokiGcsStorage` |  |  |  |
| `spec.storage.gcs.bucket` | `string` | yes |  |  |
| `spec.storage.gcs.rulerBucket` | `string` |  |  |  |
| `spec.storage.gcs.serviceAccountKeySecret` | `KubernetesLokiSecretKeyRef` |  |  |  |
| `spec.storage.gcs.serviceAccountKeySecret.name` | `string` | yes |  |  |
| `spec.storage.gcs.serviceAccountKeySecret.key` | `string` | yes |  |  |
| `spec.storage.azure` | `KubernetesLokiAzureStorage` |  |  |  |
| `spec.storage.azure.accountName` | `string` | yes |  |  |
| `spec.storage.azure.container` | `string` | yes |  |  |
| `spec.storage.azure.rulerContainer` | `string` |  |  |  |
| `spec.storage.azure.accountKeySecret` | `KubernetesLokiSecretKeyRef` |  |  |  |
| `spec.storage.azure.accountKeySecret.name` | `string` | yes |  |  |
| `spec.storage.azure.accountKeySecret.key` | `string` | yes |  |  |
| `spec.schemaFromDate` | `string` |  |  |  |
| `spec.retentionPeriod` | `string` |  |  |  |
| `spec.limits` | `KubernetesLokiLimits` |  |  |  |
| `spec.limits.ingestionRateMb` | `int32` |  |  |  |
| `spec.limits.ingestionBurstSizeMb` | `int32` |  |  |  |
| `spec.limits.maxGlobalStreamsPerUser` | `int32` |  |  |  |
| `spec.limits.maxQueryLookback` | `string` |  |  |  |
| `spec.multiTenancy` | `KubernetesLokiMultiTenancy` |  |  |  |
| `spec.multiTenancy.enabled` | `bool` |  |  |  |
| `spec.multiTenancy.tenants` | `[]KubernetesLokiTenant` |  |  |  |
| `spec.multiTenancy.tenants[].name` | `string` | yes |  |  |
| `spec.multiTenancy.tenants[].passwordHash` | `string` | yes |  |  |
| `spec.multiTenancy.existingHtpasswdSecret` | `string` |  |  |  |
| `spec.gateway` | `KubernetesLokiGateway` |  |  |  |
| `spec.gateway.enabled` | `bool` |  | `true` |  |
| `spec.gateway.replicas` | `int32` |  | `1` |  |
| `spec.gateway.resources` | `ContainerResources` |  |  |  |
| `spec.gateway.resources.limits` | `CpuMemory` |  |  |  |
| `spec.gateway.resources.limits.cpu` | `string` |  |  |  |
| `spec.gateway.resources.limits.memory` | `string` |  |  |  |
| `spec.gateway.resources.requests` | `CpuMemory` |  |  |  |
| `spec.gateway.resources.requests.cpu` | `string` |  |  |  |
| `spec.gateway.resources.requests.memory` | `string` |  |  |  |
| `spec.caching` | `KubernetesLokiCaching` |  |  |  |
| `spec.caching.chunksCacheEnabled` | `bool` |  | `true` |  |
| `spec.caching.chunksCacheMemoryMb` | `int32` |  |  |  |
| `spec.caching.resultsCacheEnabled` | `bool` |  | `true` |  |
| `spec.caching.resultsCacheMemoryMb` | `int32` |  |  |  |
| `spec.canaryEnabled` | `bool` |  | `true` |  |
| `spec.ruler` | `KubernetesLokiRuler` |  |  |  |
| `spec.ruler.enabled` | `bool` |  |  |  |
| `spec.ruler.alertmanagerUrl` | `string \| valueFrom` |  |  | KubernetesKubePrometheusStack (`status.outputs.alertmanager_endpoint`) |
| `spec.serviceMonitorEnabled` | `bool` |  |  |  |
| `spec.usageReporting` | `bool` |  | `false` |  |
| `spec.imageRegistry` | `string` |  |  |  |
| `spec.imagePullSecrets` | `[]string` |  |  |  |
| `spec.scheduling` | `KubernetesLokiScheduling` |  |  |  |
| `spec.scheduling.nodeSelector` | `map<string, string>` |  |  |  |
| `spec.scheduling.tolerations` | `[]WorkloadToleration` |  |  |  |
| `spec.scheduling.tolerations[].key` | `string` |  |  |  |
| `spec.scheduling.tolerations[].operator` | `string` |  |  |  |
| `spec.scheduling.tolerations[].value` | `string` |  |  |  |
| `spec.scheduling.tolerations[].effect` | `string` |  |  |  |
| `spec.scheduling.tolerations[].tolerationSeconds` | `int64` |  |  |  |
| `spec.scheduling.priorityClassName` | `string` |  |  |  |
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
governance labels) before installing and deleted with the resource.
When false, the namespace must already exist.

### spec.chartVersion

`string` · optional (explicit presence)

Helm chart version to install (e.g. "18.5.4" — chart 18.5.4 pairs
with Loki 3.7.4). Versions must exist as SERVED charts in the
repository index (https://grafana-community.github.io/helm-charts).

- default: `18.5.4`

### spec.monolithic

`KubernetesLokiMonolithic`

Every Loki target in one StatefulSet. The default (a
single-replica instance on filesystem storage when nothing is
declared).

### spec.monolithic.replicas

`int32` · optional (explicit presence)

Number of replicas. 1 is the filesystem-storage ceiling (the chart
refuses more without object storage); 2–3 with object storage give
HA ingest and query.

- default: `1`
- rule: {"int32":{"gte":1}}

### spec.monolithic.diskSize

`string` · optional (explicit presence)

Size of the persistent volume PER replica (e.g. "10Gi"). With
filesystem storage this volume holds ALL chunks and indexes — size
it for your full retention window. With object storage it holds
only the WAL and in-flight chunks.

- default: `10Gi`
- rule: {"string":{"pattern":"^\\d+(\\.\\d+)?\\s?(Ki|Mi|Gi|Ti|Pi|Ei|K|M|G|T|P|E)$"}}

### spec.monolithic.storageClass

`string | valueFrom`

Storage class for the volumes. Accepts a literal name or a
reference to a KubernetesStorageClass resource. Empty = the
cluster's default class.

- references: KubernetesStorageClass (`status.outputs.storage_class_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesStorageClass, name: <that resource's name>, fieldPath: status.outputs.storage_class_name}} -- a bare string does not parse

### spec.monolithic.resources

`ContainerResources`

CPU and memory for the Loki container. Empty = no requests/limits
(the chart default). Loki's memory scales with active streams and
query concurrency.

### spec.monolithic.resources.limits

`CpuMemory`

The resource limits for the container.
Specify the maximum amount of CPU and memory that the container can use.

### spec.monolithic.resources.limits.cpu

`string`

### spec.monolithic.resources.limits.memory

`string`

### spec.monolithic.resources.requests

`CpuMemory`

The resource requests for the container.
Specify the minimum amount of CPU and memory that the container is guaranteed.

### spec.monolithic.resources.requests.cpu

`string`

### spec.monolithic.resources.requests.memory

`string`

### spec.simpleScalable

`KubernetesLokiSimpleScalable`

Write/read/backend tiers scaling independently. REQUIRES an
object-storage backend (the chart refuses filesystem here — reads
and writes rendezvous in the object store, not on a shared disk).

### spec.simpleScalable.writeReplicas

`int32` · optional (explicit presence)

Write tier (distributor + ingester; a StatefulSet with a WAL
volume). 3 is the chart's default and the replication-factor-safe
floor.

- default: `3`
- rule: {"int32":{"gte":1}}

### spec.simpleScalable.readReplicas

`int32` · optional (explicit presence)

Read tier (query-frontend + querier; a stateless Deployment that
reads from object storage).

- default: `3`
- rule: {"int32":{"gte":1}}

### spec.simpleScalable.backendReplicas

`int32` · optional (explicit presence)

Backend tier (compactor, ruler, index-gateway, scheduler; a
StatefulSet with a working volume).

- default: `3`
- rule: {"int32":{"gte":1}}

### spec.simpleScalable.diskSize

`string` · optional (explicit presence)

Size of the persistent volume PER write/backend replica (e.g.
"10Gi" — WAL and compaction workspace; chunks live in the object
store).

- default: `10Gi`
- rule: {"string":{"pattern":"^\\d+(\\.\\d+)?\\s?(Ki|Mi|Gi|Ti|Pi|Ei|K|M|G|T|P|E)$"}}

### spec.simpleScalable.storageClass

`string | valueFrom`

Storage class for the write/backend volumes. Empty = the cluster's
default class.

- references: KubernetesStorageClass (`status.outputs.storage_class_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesStorageClass, name: <that resource's name>, fieldPath: status.outputs.storage_class_name}} -- a bare string does not parse

### spec.simpleScalable.resources

`ContainerResources`

CPU and memory applied to each tier's containers. Empty = no
requests/limits (the chart default). Size the write tier first —
it holds the ingest path's memory.

### spec.simpleScalable.resources.limits

`CpuMemory`

The resource limits for the container.
Specify the maximum amount of CPU and memory that the container can use.

### spec.simpleScalable.resources.limits.cpu

`string`

### spec.simpleScalable.resources.limits.memory

`string`

### spec.simpleScalable.resources.requests

`CpuMemory`

The resource requests for the container.
Specify the minimum amount of CPU and memory that the container is guaranteed.

### spec.simpleScalable.resources.requests.cpu

`string`

### spec.simpleScalable.resources.requests.memory

`string`

### spec.storage

`KubernetesLokiStorage`

Where chunks and indexes live. Empty = filesystem (a
PersistentVolume) — honest only for a single monolithic replica;
declare an object-storage backend for anything bigger.

### spec.storage.filesystem

`KubernetesLokiFilesystemStorage`

Chunks and indexes on the pod's PersistentVolume. Single
monolithic replica only — the chart itself refuses every other
shape on filesystem storage.

### spec.storage.s3

`KubernetesLokiS3Storage`

Amazon S3 or any S3-compatible store. For in-cluster storage
point `endpoint` at a KubernetesSeaweedFs S3 endpoint with
`force_path_style: true`.

### spec.storage.s3.bucket

`string` · required

Bucket for chunks and indexes (must exist; Loki does not create
it). Also serves ruler state unless `ruler_bucket` is set.

- rule: {"required":true}

### spec.storage.s3.rulerBucket

`string`

Separate bucket for ruler state. Empty = `bucket`.

### spec.storage.s3.region

`string`

AWS region (e.g. "us-east-1"). Required for Amazon S3; usually
empty for S3-compatible endpoints.

### spec.storage.s3.endpoint

`string`

Custom endpoint for S3-COMPATIBLE stores (e.g.
"http://objects-filer.storage.svc.cluster.local:8333" for an
in-cluster KubernetesSeaweedFs). Empty = Amazon S3.

### spec.storage.s3.forcePathStyle

`bool`

Use path-style addressing (bucket in the path, not the hostname).
Required by most S3-compatible endpoints, including SeaweedFS.

### spec.storage.s3.insecure

`bool`

Allow plain-HTTP endpoints. Only for in-cluster stores; never over
a network you do not own.

### spec.storage.s3.credentials

`KubernetesLokiObjectStoreCredentials`

Static credentials. Empty = the pod's ambient identity (IRSA on
EKS — the recommended keyless posture). When declared, the modules
inject them as environment variables from a Secret — they never
land in the rendered Loki config.

### spec.storage.s3.credentials.accessKeyIdSecret

`KubernetesLokiSecretKeyRef` · required

Existing Secret holding the access key ID.

- rule: {"required":true}

### spec.storage.s3.credentials.accessKeyIdSecret.name

`string` · required

Secret name.

- rule: {"required":true}

### spec.storage.s3.credentials.accessKeyIdSecret.key

`string` · required

Key within the Secret.

- rule: {"required":true}

### spec.storage.s3.credentials.secretAccessKeySecret

`KubernetesLokiSecretKeyRef` · required

Existing Secret holding the secret access key.

- rule: {"required":true}

### spec.storage.s3.credentials.secretAccessKeySecret.name

`string` · required

Secret name.

- rule: {"required":true}

### spec.storage.s3.credentials.secretAccessKeySecret.key

`string` · required

Key within the Secret.

- rule: {"required":true}

### spec.storage.gcs

`KubernetesLokiGcsStorage`

Google Cloud Storage. With no key declared the pod's ambient
identity is used (GKE workload identity — the keyless posture).

### spec.storage.gcs.bucket

`string` · required

Bucket for chunks and indexes. Also serves ruler state unless
`ruler_bucket` is set.

- rule: {"required":true}

### spec.storage.gcs.rulerBucket

`string`

Separate bucket for ruler state. Empty = `bucket`.

### spec.storage.gcs.serviceAccountKeySecret

`KubernetesLokiSecretKeyRef`

Existing Secret holding a service-account JSON key. Empty = the
pod's ambient identity (GKE workload identity — the recommended
keyless posture). When declared, the key is mounted and referenced
via GOOGLE_APPLICATION_CREDENTIALS — never inlined in config.

### spec.storage.gcs.serviceAccountKeySecret.name

`string` · required

Secret name.

- rule: {"required":true}

### spec.storage.gcs.serviceAccountKeySecret.key

`string` · required

Key within the Secret.

- rule: {"required":true}

### spec.storage.azure

`KubernetesLokiAzureStorage`

Azure Blob Storage.

### spec.storage.azure.accountName

`string` · required

Storage-account name.

- rule: {"required":true}

### spec.storage.azure.container

`string` · required

Blob container for chunks and indexes. Also serves ruler state
unless `ruler_container` is set.

- rule: {"required":true}

### spec.storage.azure.rulerContainer

`string`

Separate container for ruler state. Empty = `container`.

### spec.storage.azure.accountKeySecret

`KubernetesLokiSecretKeyRef`

Existing Secret holding the storage-account key. Empty = federated
workload identity (AKS — the recommended keyless posture; the
modules set the chart's federated-token switch).

### spec.storage.azure.accountKeySecret.name

`string` · required

Secret name.

- rule: {"required":true}

### spec.storage.azure.accountKeySecret.key

`string` · required

Key within the Secret.

- rule: {"required":true}

### spec.schemaFromDate

`string`

Override the derived schema start date (RFC3339 date, e.g.
"2020-10-24"). ONLY for importing an existing Loki whose schema
began on a real date — the date must be on or before the first
ingested log's day, and CHANGING it on a running install corrupts
queries over old data. New installs never set this.

- rule: schema_from_date must be a date like '2024-04-01'

### spec.retentionPeriod

`string`

How long logs are kept before the compactor deletes them, as a
Prometheus common-model duration in hours or days (e.g. "744h" or
"31d"). Empty = keep everything forever (Loki's own default) —
object-storage costs grow unbounded; production installs should
set this. Deletion is asynchronous: the compactor marks and later
sweeps, so objects outlive the period by the delete delay — any
bucket lifecycle policy must expire LATER than this period, never
earlier.

- rule: retention_period must be a duration in hours or days like '744h' or '31d'

### spec.limits

`KubernetesLokiLimits`

Ingestion and query limits — the knobs that stop one noisy tenant
or one bad query from taking the install down.

### spec.limits.ingestionRateMb

`int32` · optional (explicit presence)

Per-tenant ingestion rate in MB/s. Empty = Loki's default (4).

- rule: {"int32":{"gte":1}}

### spec.limits.ingestionBurstSizeMb

`int32` · optional (explicit presence)

Per-tenant ingestion burst in MB. Empty = Loki's default (6).

- rule: {"int32":{"gte":1}}

### spec.limits.maxGlobalStreamsPerUser

`int32` · optional (explicit presence)

Maximum number of active streams per tenant. Empty = Loki's
default (a high-cardinality label set explodes stream counts —
this is the fence).

- rule: {"int32":{"gte":1}}

### spec.limits.maxQueryLookback

`string`

How far back queries may look, as a Prometheus duration (e.g.
"720h"). Empty = unlimited. Pair with `retention_period` — queries
beyond retention return nothing anyway.

- rule: max_query_lookback must be a Prometheus duration like '720h' or '30d'

### spec.multiTenancy

`KubernetesLokiMultiTenancy`

Tenant isolation. Empty = single-tenant (see the TENANCY note
above). When enabled, every push/query must carry an X-Scope-OrgID
header and the gateway enforces basic auth for the declared
tenants.

- rule: declaring tenants or an htpasswd Secret requires multi_tenancy.enabled: true — without it the gateway enforces no auth and they would silently do nothing
- rule: tenants and existing_htpasswd_secret are mutually exclusive — declare the tenant list OR bring your own htpasswd Secret, not both

### spec.multiTenancy.enabled

`bool`

Require an X-Scope-OrgID tenant header on every push and query.
KNOW THIS: basic auth at the gateway AUTHENTICATES clients; it
does not set their tenant — each client still sends the
X-Scope-OrgID header naming the tenant it writes to or reads from
(the collector's `headers` setting, Grafana's datasource httpHeader
jsonData).

### spec.multiTenancy.tenants

`[]KubernetesLokiTenant`

Tenants gated at the gateway with HTTP basic auth. The chart
builds the gateway's htpasswd file from these entries. Requires
`enabled: true`. Mutually exclusive with
`existing_htpasswd_secret`.

### spec.multiTenancy.tenants[].name

`string` · required

Tenant name — the basic-auth username. By convention clients use
the same value as their X-Scope-OrgID tenant header.

- rule: {"required":true}

### spec.multiTenancy.tenants[].passwordHash

`string` · required

bcrypt htpasswd HASH of the tenant's password (generate with
`htpasswd -nbBC10 <name> <password>` and take the part after the
colon). A one-way hash, safe to carry in a manifest — the actual
password never appears anywhere. The chart writes these into the
gateway's htpasswd Secret.

- rule: password_hash must be a bcrypt htpasswd hash (starts with $2y$ or $2b$ — generate with `htpasswd -nbBC10 <name> <password>`), never a plaintext password
- rule: {"required":true}

### spec.multiTenancy.existingHtpasswdSecret

`string`

Name of an existing Secret carrying a ready-made `.htpasswd` key —
the bring-your-own-credentials arm for teams that manage htpasswd
material themselves. Mutually exclusive with `tenants`. Requires
`enabled: true`.

### spec.gateway

`KubernetesLokiGateway`

The nginx gateway — the single front door that routes pushes and
queries to the right internal target in every mode. On by default;
disable only when clients address the internal services directly
(single-tenant monolithic only — the exported endpoints assume the
gateway).

### spec.gateway.enabled

`bool` · optional (explicit presence)

Deploy the gateway. Default true (the exported endpoints route
through it).

- default: `true`

### spec.gateway.replicas

`int32` · optional (explicit presence)

Number of gateway replicas.

- default: `1`
- rule: {"int32":{"gte":1}}

### spec.gateway.resources

`ContainerResources`

CPU and memory for the gateway container. Empty = no
requests/limits (nginx is light).

### spec.gateway.resources.limits

`CpuMemory`

The resource limits for the container.
Specify the maximum amount of CPU and memory that the container can use.

### spec.gateway.resources.limits.cpu

`string`

### spec.gateway.resources.limits.memory

`string`

### spec.gateway.resources.requests

`CpuMemory`

The resource requests for the container.
Specify the minimum amount of CPU and memory that the container is guaranteed.

### spec.gateway.resources.requests.cpu

`string`

### spec.gateway.resources.requests.memory

`string`

### spec.caching

`KubernetesLokiCaching`

The memcached-based caches the chart deploys alongside Loki. Both
on by default (the chart's grain — queries without the results
cache re-scan object storage on every dashboard refresh). The
DEFAULT SIZES ARE PRODUCTION-SCALE: unset, the chunks cache alone
requests 9830Mi of memory (verified live) — on a small cluster the
pod never schedules and the atomic install rolls the whole release
back, so small clusters must size chunks_cache_memory_mb down.

### spec.caching.chunksCacheEnabled

`bool` · optional (explicit presence)

Deploy the chunks cache (dedicated memcached; keeps hot chunks out
of object-store reads). Default true.

- default: `true`

### spec.caching.chunksCacheMemoryMb

`int32` · optional (explicit presence)

Memory allocated to the chunks-cache memcached in MB. Empty = the
chart default (8192), sized for production log volume: the chart
requests container memory at 1.2× this value, so the default
renders a 9830Mi request that NEVER SCHEDULES on a node with less
than ~10Gi allocatable — the pod stays Pending and the atomic
install rolls the whole release back after its timeout (verified
live). Set this explicitly on any small or dev cluster (128–1024
is plenty for light query loads); the biggest memory consumer
after Loki itself.

- rule: {"int32":{"gte":64}}

### spec.caching.resultsCacheEnabled

`bool` · optional (explicit presence)

Deploy the query-results cache. Default true — without it every
dashboard refresh re-runs its queries end to end.

- default: `true`

### spec.caching.resultsCacheMemoryMb

`int32` · optional (explicit presence)

Memory allocated to the results-cache memcached in MB. Empty = the
chart default (1024).

- rule: {"int32":{"gte":64}}

### spec.canaryEnabled

`bool` · optional (explicit presence)

The Loki canary — a DaemonSet that continuously writes and reads
test log lines through the full pipeline, turning silent log loss
into a visible metric. On by default (the chart's grain).

- default: `true`

### spec.ruler

`KubernetesLokiRuler`

The ruler — evaluates alerting/recording rules over logs (LogQL)
and fires alerts to an Alertmanager. Off by default.

### spec.ruler.enabled

`bool`

Evaluate LogQL alerting/recording rules. Rules are discovered from
ConfigMaps labeled `loki_rule: "1"` (the sidecar contract — the
same pattern Grafana dashboards use).

### spec.ruler.alertmanagerUrl

`string | valueFrom`

Alertmanager URL to fire alerts at. Accepts a literal URL or a
reference to a KubernetesKubePrometheusStack resource (its
Alertmanager endpoint) — the one-line wiring into the cluster's
alerting.

- references: KubernetesKubePrometheusStack (`status.outputs.alertmanager_endpoint`)
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesKubePrometheusStack, name: <that resource's name>, fieldPath: status.outputs.alertmanager_endpoint}} -- a bare string does not parse

### spec.serviceMonitorEnabled

`bool`

Render ServiceMonitor objects for Prometheus discovery. Requires
the monitoring.coreos.com CRDs on the cluster (deploy
KubernetesKubePrometheusStack first).

### spec.usageReporting

`bool` · optional (explicit presence)

Send anonymous usage statistics about this install to Grafana Labs.
Default false — this component deliberately diverges from Loki's
report-by-default so no data leaves the cluster without an explicit
opt-in.

- default: `false`

### spec.imageRegistry

`string`

Registry that replaces the registry part of EVERY image this
component pulls (loki, gateway nginx, memcached caches, canary,
rules sidecar) — the air-gap/private-mirror path. Empty = each
image's upstream registry (docker.io). KNOW THIS: the chart mixes
split registry+repository image values with bare library images
(memcached) — the modules translate this override correctly for
both forms.

### spec.imagePullSecrets

`[]string`

Names of existing image-pull Secrets applied to every workload
(chart `imagePullSecrets`). The Secrets must already exist in the
namespace.

### spec.scheduling

`KubernetesLokiScheduling`

Scheduling applied to the Loki workloads (monolithic StatefulSet or
the write/read/backend tiers). The gateway, caches and canary keep
the chart's own scheduling; steer them via `helm_values`.

### spec.scheduling.nodeSelector

`map<string, string>`

Node selector for the Loki pods.

### spec.scheduling.tolerations

`[]WorkloadToleration`

Tolerations for the Loki pods.

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

Priority class name for the Loki pods.

### spec.helmValues

`string`

Escape hatch: additional chart values as a YAML document, merged
LAST over everything the typed fields render (Helm `-f` semantics,
identical on both engines). For the chart surface beyond the typed
fields (bloom filters, pattern ingester, zone-aware rollouts,
per-component overrides, ruler storage) — never the substitute for
them. Do not put secrets here; object-storage credentials belong in
the typed secret-reference fields, which keep them out of the
rendered config.

## Validation Rules

- `spec.mode.simple_scalable.requires_object_storage`: simple_scalable mode requires an object-storage backend (s3, gcs or azure) — the write/read/backend tiers rendezvous in the object store, so filesystem storage cannot serve them (this mirrors the chart's own validation)
- `spec.mode.monolithic.replicas.require_object_storage`: more than one monolithic replica requires an object-storage backend (s3, gcs or azure) — replicas cannot share a filesystem volume (this mirrors the chart's own validation)

## Outputs

Reference an output from another manifest as `valueFrom: {kind: KubernetesLoki, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.namespace` | `string` | namespace Loki runs in. |
| `status.outputs.release_name` | `string` | Helm release name (= metadata.name). The modules pin the chart's fullname to it, so every child name below derives from it. |
| `status.outputs.gateway_service` | `string` | name of the gateway Service (`<name>-gateway`, port 80) — the one front door for pushes and queries in every mode. Empty when the gateway is disabled. |
| `status.outputs.gateway_endpoint` | `string` | in-cluster endpoint of the gateway — the URL log shippers push to and Grafana `loki` datasources read from, e.g. http://logs-gateway.observability.svc.cluster.local |
| `status.outputs.otlp_push_endpoint` | `string` | in-cluster OTLP log-ingest endpoint (the gateway's `/otlp` route) — point a KubernetesOtelCollector otlphttp exporter here, e.g. http://logs-gateway.observability.svc.cluster.local/otlp |
| `status.outputs.loki_service` | `string` | name of the Loki HTTP Service (`<name>`, port 3100) — the direct internal API behind the gateway (monolithic) or the read tier's entry (simple_scalable). |
| `status.outputs.port_forward_command` | `string` | command to port-forward the gateway to a developer laptop, e.g. kubectl port-forward svc/logs-gateway -n observability 3100:80 |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.namespace` | KubernetesNamespace | `spec.name` |
| `spec.monolithic.storageClass` | KubernetesStorageClass | `status.outputs.storage_class_name` |
| `spec.simpleScalable.storageClass` | KubernetesStorageClass | `status.outputs.storage_class_name` |
| `spec.ruler.alertmanagerUrl` | KubernetesKubePrometheusStack | `status.outputs.alertmanager_endpoint` |

## See Also

- [Overview](../README.md)
