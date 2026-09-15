# KubernetesSeaweedFs

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `kubernetes.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

**KubernetesSeaweedFsSpec** deploys SeaweedFS — the catalog's
in-cluster S3-compatible object store (Apache-2.0) — from the
official `seaweedfs` Helm chart
(https://seaweedfs.github.io/seaweedfs/helm).

TOPOLOGY: SeaweedFS separates metadata from data. `master` servers
coordinate the cluster and assign file ids, `volume` servers store
the actual blobs, and the `filer` provides the file/bucket
namespace plus the S3 API. Each tier scales independently; the
defaults (1/1/1 on PVCs) give a working single-node store, and
3 masters + N volumes + 2 filers is the HA shape.

S3: the S3 gateway is ON by default — this kind exists to serve
S3. It runs embedded on the filer pods unless `s3.dedicated` is
declared (then a separate Deployment scales the API independently
of metadata). Auth is ON by default: the chart materializes an
admin and a read-only credential pair in the
`<name>-s3-secret` Secret (stable across upgrades, kept on
uninstall) — the stack outputs point at it. Buckets declared in
`s3.buckets` are created by the chart's post-install hook.

STORAGE: the chart's out-of-the-box storage is hostPath (bare-metal
grain); this kind deliberately maps every data volume to a
PersistentVolumeClaim — declare sizes and (optionally) a
StorageClass per tier. Master and filer metadata are small; volume
servers hold the object bytes.

EXPOSURE: everything stays ClusterIP — exposure composes from
first-class kinds (KubernetesIngress, Gateway API kinds) over the
exported service handles.

The typed fields below cover the chart's meaningful configuration
surface; `helm_values` remains as the escape hatch for chart
values beyond them (merged last, Helm `-f` semantics, identical on
both engines) — per-tier scheduling, the SFTP arm, the maintenance
worker, mTLS (`global.seaweedfs.enableSecurity`, requires
cert-manager), external filer stores (Postgres/MySQL via
`filer.extraEnvironmentVars` + `filer.secretExtraEnvironmentVars`),
and the all-in-one dev mode ride there — a safety valve, never the
primary interface.

## Example

```yaml
# Full-surface offline-proof manifest: exercises an HA topology (3 masters,
# 3 volume servers with replication 001), per-tier PVCs with explicit
# StorageClasses, a leveldb volume index, filer extra env + volume-data
# encryption, an authenticated DEDICATED S3 gateway with typed buckets (TTL,
# versioning, object lock, anonymous read) and a virtual-host domain, the
# authenticated admin console with persistence, an image override, and an
# escape-hatch SFTP arm — so the offline tofu plan and pulumi preview proofs
# cover the full typed surface. Placeholder values; never applied to a real
# cluster.
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesSeaweedFs
metadata:
  name: seaweedfs-hack
spec:
  namespace:
    value: seaweedfs-hack
  createNamespace: true
  chartVersion: 4.40.0
  master:
    replicas: 3
    dataVolume:
      size: 5Gi
      storageClass:
        value: fast-ssd
    volumeSizeLimitMb: 1000
    resources:
      requests:
        cpu: 100m
        memory: 256Mi
      limits:
        cpu: "1"
        memory: 512Mi
  volume:
    replicas: 3
    dataVolume:
      size: 100Gi
      storageClass:
        value: fast-ssd
    maxVolumes: 0
    indexMode: leveldb
    minFreeSpacePercent: 5
    resources:
      requests:
        cpu: 250m
        memory: 512Mi
      limits:
        cpu: "2"
        memory: 2Gi
  filer:
    replicas: 1
    dataVolume:
      size: 10Gi
      storageClass:
        value: fast-ssd
    encryptVolumeData: true
    extraEnvironmentVars:
      WEED_FILER_OPTIONS_RECURSIVE_DELETE: "true"
    resources:
      requests:
        cpu: 250m
        memory: 512Mi
      limits:
        cpu: "1"
        memory: 1Gi
  s3:
    enabled: true
    enableAuth: true
    buckets:
      - name: artifacts
        ttl: 30d
        versioning: true
      - name: public-assets
        anonymousRead: true
      - name: compliance-archive
        objectLock: true
    domainName: s3.example.internal
    dedicated:
      replicas: 2
      resources:
        requests:
          cpu: 250m
          memory: 256Mi
        limits:
          cpu: "1"
          memory: 512Mi
  replication: "001"
  admin:
    enabled: true
    dataVolume:
      size: 10Gi
  serviceMonitorEnabled: false
  image:
    registry: mirror.example.com
    repository: chrislusf/seaweedfs
    tag: "4.40"
  helmValues: |
    sftp:
      enabled: false
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.namespace` | `string \| valueFrom` | yes |  | KubernetesNamespace (`spec.name`) |
| `spec.createNamespace` | `bool` |  |  |  |
| `spec.chartVersion` | `string` |  | `4.40.0` |  |
| `spec.master` | `KubernetesSeaweedFsMaster` |  |  |  |
| `spec.master.replicas` | `int32` |  | `1` |  |
| `spec.master.dataVolume` | `KubernetesSeaweedFsDataVolume` |  |  |  |
| `spec.master.dataVolume.size` | `string` |  |  |  |
| `spec.master.dataVolume.storageClass` | `string \| valueFrom` |  |  | KubernetesStorageClass (`status.outputs.storage_class_name`) |
| `spec.master.volumeSizeLimitMb` | `int32` |  | `1000` |  |
| `spec.master.resources` | `ContainerResources` |  |  |  |
| `spec.master.resources.limits` | `CpuMemory` |  |  |  |
| `spec.master.resources.limits.cpu` | `string` |  |  |  |
| `spec.master.resources.limits.memory` | `string` |  |  |  |
| `spec.master.resources.requests` | `CpuMemory` |  |  |  |
| `spec.master.resources.requests.cpu` | `string` |  |  |  |
| `spec.master.resources.requests.memory` | `string` |  |  |  |
| `spec.volume` | `KubernetesSeaweedFsVolume` |  |  |  |
| `spec.volume.replicas` | `int32` |  | `1` |  |
| `spec.volume.dataVolume` | `KubernetesSeaweedFsDataVolume` |  |  |  |
| `spec.volume.dataVolume.size` | `string` |  |  |  |
| `spec.volume.dataVolume.storageClass` | `string \| valueFrom` |  |  | KubernetesStorageClass (`status.outputs.storage_class_name`) |
| `spec.volume.maxVolumes` | `int32` |  |  |  |
| `spec.volume.indexMode` | `string` |  |  |  |
| `spec.volume.minFreeSpacePercent` | `int32` |  | `1` |  |
| `spec.volume.resources` | `ContainerResources` |  |  |  |
| `spec.volume.resources.limits` | `CpuMemory` |  |  |  |
| `spec.volume.resources.limits.cpu` | `string` |  |  |  |
| `spec.volume.resources.limits.memory` | `string` |  |  |  |
| `spec.volume.resources.requests` | `CpuMemory` |  |  |  |
| `spec.volume.resources.requests.cpu` | `string` |  |  |  |
| `spec.volume.resources.requests.memory` | `string` |  |  |  |
| `spec.filer` | `KubernetesSeaweedFsFiler` |  |  |  |
| `spec.filer.replicas` | `int32` |  | `1` |  |
| `spec.filer.dataVolume` | `KubernetesSeaweedFsDataVolume` |  |  |  |
| `spec.filer.dataVolume.size` | `string` |  |  |  |
| `spec.filer.dataVolume.storageClass` | `string \| valueFrom` |  |  | KubernetesStorageClass (`status.outputs.storage_class_name`) |
| `spec.filer.encryptVolumeData` | `bool` |  |  |  |
| `spec.filer.extraEnvironmentVars` | `map<string, string>` |  |  |  |
| `spec.filer.resources` | `ContainerResources` |  |  |  |
| `spec.filer.resources.limits` | `CpuMemory` |  |  |  |
| `spec.filer.resources.limits.cpu` | `string` |  |  |  |
| `spec.filer.resources.limits.memory` | `string` |  |  |  |
| `spec.filer.resources.requests` | `CpuMemory` |  |  |  |
| `spec.filer.resources.requests.cpu` | `string` |  |  |  |
| `spec.filer.resources.requests.memory` | `string` |  |  |  |
| `spec.s3` | `KubernetesSeaweedFsS3` |  |  |  |
| `spec.s3.enabled` | `bool` |  | `true` |  |
| `spec.s3.enableAuth` | `bool` |  | `true` |  |
| `spec.s3.existingConfigSecret` | `string` |  |  |  |
| `spec.s3.buckets` | `[]KubernetesSeaweedFsS3Bucket` |  |  |  |
| `spec.s3.buckets[].name` | `string` | yes |  |  |
| `spec.s3.buckets[].anonymousRead` | `bool` |  |  |  |
| `spec.s3.buckets[].ttl` | `string` |  |  |  |
| `spec.s3.buckets[].objectLock` | `bool` |  |  |  |
| `spec.s3.buckets[].versioning` | `bool` |  |  |  |
| `spec.s3.domainName` | `string` |  |  |  |
| `spec.s3.dedicated` | `KubernetesSeaweedFsS3Dedicated` |  |  |  |
| `spec.s3.dedicated.replicas` | `int32` |  | `1` |  |
| `spec.s3.dedicated.resources` | `ContainerResources` |  |  |  |
| `spec.s3.dedicated.resources.limits` | `CpuMemory` |  |  |  |
| `spec.s3.dedicated.resources.limits.cpu` | `string` |  |  |  |
| `spec.s3.dedicated.resources.limits.memory` | `string` |  |  |  |
| `spec.s3.dedicated.resources.requests` | `CpuMemory` |  |  |  |
| `spec.s3.dedicated.resources.requests.cpu` | `string` |  |  |  |
| `spec.s3.dedicated.resources.requests.memory` | `string` |  |  |  |
| `spec.replication` | `string` |  |  |  |
| `spec.admin` | `KubernetesSeaweedFsAdmin` |  |  |  |
| `spec.admin.enabled` | `bool` |  |  |  |
| `spec.admin.existingAuthSecret` | `string` |  |  |  |
| `spec.admin.dataVolume` | `KubernetesSeaweedFsDataVolume` |  |  |  |
| `spec.admin.dataVolume.size` | `string` |  |  |  |
| `spec.admin.dataVolume.storageClass` | `string \| valueFrom` |  |  | KubernetesStorageClass (`status.outputs.storage_class_name`) |
| `spec.admin.resources` | `ContainerResources` |  |  |  |
| `spec.admin.resources.limits` | `CpuMemory` |  |  |  |
| `spec.admin.resources.limits.cpu` | `string` |  |  |  |
| `spec.admin.resources.limits.memory` | `string` |  |  |  |
| `spec.admin.resources.requests` | `CpuMemory` |  |  |  |
| `spec.admin.resources.requests.cpu` | `string` |  |  |  |
| `spec.admin.resources.requests.memory` | `string` |  |  |  |
| `spec.serviceMonitorEnabled` | `bool` |  |  |  |
| `spec.image` | `KubernetesSeaweedFsImage` |  |  |  |
| `spec.image.registry` | `string` |  |  |  |
| `spec.image.repository` | `string` |  |  |  |
| `spec.image.tag` | `string` |  |  |  |
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

Helm chart version to install (e.g. "4.40.0" — chart versions
track SeaweedFS releases; chart 4.40.0 ships appVersion 4.40).
Versions must exist as SERVED charts in the repository index
(https://seaweedfs.github.io/seaweedfs/helm).

- default: `4.40.0`

### spec.master

`KubernetesSeaweedFsMaster`

Master tier — cluster coordination and volume assignment.
Empty = 1 master on a 5Gi PVC.

### spec.master.replicas

`int32` · optional (explicit presence)

Master replicas. Masters form a Raft quorum — use 1 (dev) or an
odd count (3 for HA); an even count wastes a node without adding
failure tolerance.

- default: `1`
- rule: {"int32":{"gte":1}}

### spec.master.dataVolume

`KubernetesSeaweedFsDataVolume`

Metadata volume. Master state is small — the 5Gi default is
generous.

### spec.master.dataVolume.size

`string`

Volume size as a Kubernetes quantity (e.g. "30Gi"). Empty = the
tier's default (master 5Gi, volume 30Gi, filer 10Gi, admin
10Gi).

- rule: size must be a Kubernetes quantity like '30Gi' or '500Mi'

### spec.master.dataVolume.storageClass

`string | valueFrom`

StorageClass for the PVC. Accepts a literal class name or a
reference to a KubernetesStorageClass resource. Empty = the
cluster's default class.

- references: KubernetesStorageClass (`status.outputs.storage_class_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesStorageClass, name: <that resource's name>, fieldPath: status.outputs.storage_class_name}} -- a bare string does not parse

### spec.master.volumeSizeLimitMb

`int32` · optional (explicit presence)

Per-volume-file size limit in MB before the master assigns a new
volume (the chart/upstream default 1000 suits mixed workloads;
raise for large-object stores).

- default: `1000`
- rule: {"int32":{"gte":1}}

### spec.master.resources

`ContainerResources`

Container resources for master pods.

### spec.master.resources.limits

`CpuMemory`

The resource limits for the container.
Specify the maximum amount of CPU and memory that the container can use.

### spec.master.resources.limits.cpu

`string`

### spec.master.resources.limits.memory

`string`

### spec.master.resources.requests

`CpuMemory`

The resource requests for the container.
Specify the minimum amount of CPU and memory that the container is guaranteed.

### spec.master.resources.requests.cpu

`string`

### spec.master.resources.requests.memory

`string`

### spec.volume

`KubernetesSeaweedFsVolume`

Volume tier — the servers that store object bytes. Empty =
1 volume server with a 30Gi data PVC. Size this tier for your
data; it is the only tier that grows with stored bytes.

### spec.volume.replicas

`int32` · optional (explicit presence)

Volume server replicas. With `replication` set, must be enough
servers/racks/DCs to place every copy.

- default: `1`
- rule: {"int32":{"gte":1}}

### spec.volume.dataVolume

`KubernetesSeaweedFsDataVolume`

Data volume PER volume-server pod — where object bytes live.
Empty = 30Gi on the cluster's default StorageClass.

### spec.volume.dataVolume.size

`string`

Volume size as a Kubernetes quantity (e.g. "30Gi"). Empty = the
tier's default (master 5Gi, volume 30Gi, filer 10Gi, admin
10Gi).

- rule: size must be a Kubernetes quantity like '30Gi' or '500Mi'

### spec.volume.dataVolume.storageClass

`string | valueFrom`

StorageClass for the PVC. Accepts a literal class name or a
reference to a KubernetesStorageClass resource. Empty = the
cluster's default class.

- references: KubernetesStorageClass (`status.outputs.storage_class_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesStorageClass, name: <that resource's name>, fieldPath: status.outputs.storage_class_name}} -- a bare string does not parse

### spec.volume.maxVolumes

`int32`

Max SeaweedFS volume files per server. 0 (default) auto-sizes
from free disk — the right answer almost always.

- rule: {"int32":{"gte":0}}

### spec.volume.indexMode

`string`

Needle-index memory mode: memory (fastest, index rebuilt from
disk on start), leveldb, leveldbMedium, or leveldbLarge (least
memory, for very large stores). Empty = upstream default
(memory).

- rule: index_mode must be memory, leveldb, leveldbMedium, or leveldbLarge

### spec.volume.minFreeSpacePercent

`int32` · optional (explicit presence)

Mark all volumes read-only when free disk drops below this
percentage (upstream default 1).

- default: `1`
- rule: {"int32":{"lte":100,"gte":0}}

### spec.volume.resources

`ContainerResources`

Container resources for volume-server pods.

### spec.volume.resources.limits

`CpuMemory`

The resource limits for the container.
Specify the maximum amount of CPU and memory that the container can use.

### spec.volume.resources.limits.cpu

`string`

### spec.volume.resources.limits.memory

`string`

### spec.volume.resources.requests

`CpuMemory`

The resource requests for the container.
Specify the minimum amount of CPU and memory that the container is guaranteed.

### spec.volume.resources.requests.cpu

`string`

### spec.volume.resources.requests.memory

`string`

### spec.filer

`KubernetesSeaweedFsFiler`

Filer tier — the file/bucket namespace and the S3 API host.
Empty = 1 filer on a 10Gi PVC (embedded leveldb metadata store —
the chart default; external stores ride `helm_values`).

### spec.filer.replicas

`int32` · optional (explicit presence)

Filer replicas. Multiple filers share state through the metadata
store; the embedded leveldb default is per-pod, so run 1 filer
unless you wire a shared external store (Postgres/MySQL via
`helm_values`).

- default: `1`
- rule: {"int32":{"gte":1}}

### spec.filer.dataVolume

`KubernetesSeaweedFsDataVolume`

Metadata volume (the embedded leveldb store and filer logs).
Empty = 10Gi on the cluster's default StorageClass.

### spec.filer.dataVolume.size

`string`

Volume size as a Kubernetes quantity (e.g. "30Gi"). Empty = the
tier's default (master 5Gi, volume 30Gi, filer 10Gi, admin
10Gi).

- rule: size must be a Kubernetes quantity like '30Gi' or '500Mi'

### spec.filer.dataVolume.storageClass

`string | valueFrom`

StorageClass for the PVC. Accepts a literal class name or a
reference to a KubernetesStorageClass resource. Empty = the
cluster's default class.

- references: KubernetesStorageClass (`status.outputs.storage_class_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesStorageClass, name: <that resource's name>, fieldPath: status.outputs.storage_class_name}} -- a bare string does not parse

### spec.filer.encryptVolumeData

`bool`

Encrypt object data on the volume servers as it is written
through this filer (filer -encryptVolumeData). Ciphertext lands
on disk; keys stay in filer metadata.

### spec.filer.extraEnvironmentVars

`map<string, string>`

Extra environment variables for the filer, exactly as SeaweedFS
expects them (WEED_* keys — this is upstream's configuration
surface for filer stores and options, e.g.
"WEED_FILER_OPTIONS_RECURSIVE_DELETE": "true"). Values here are
plain text; credential-bearing variables belong in a Secret
wired through `filer.secretExtraEnvironmentVars` in
`helm_values` (secretKeyRef entries — references, not material).

### spec.filer.resources

`ContainerResources`

Container resources for filer pods.

### spec.filer.resources.limits

`CpuMemory`

The resource limits for the container.
Specify the maximum amount of CPU and memory that the container can use.

### spec.filer.resources.limits.cpu

`string`

### spec.filer.resources.limits.memory

`string`

### spec.filer.resources.requests

`CpuMemory`

The resource requests for the container.
Specify the minimum amount of CPU and memory that the container is guaranteed.

### spec.filer.resources.requests.cpu

`string`

### spec.filer.resources.requests.memory

`string`

### spec.s3

`KubernetesSeaweedFsS3`

S3 gateway. Empty = enabled with auth, embedded on the filer —
the component's reason to exist. Declare explicitly to add
buckets, wire an existing credential config, or split the
gateway into its own Deployment.

### spec.s3.enabled

`bool` · optional (explicit presence)

Serve the S3 API. Component default: true (unset renders as
enabled — this kind IS the catalog's S3 store); set false
explicitly for a pure filer/POSIX deployment.

- default: `true`

### spec.s3.enableAuth

`bool` · optional (explicit presence)

Require S3 credentials. Component default: true — the chart
materializes admin + read-only credential pairs in the
`<name>-s3-secret` Secret (generated once, stable across
upgrades, kept on uninstall; surfaced in the stack outputs).
False serves an OPEN in-cluster S3 endpoint — dev only.

- default: `true`

### spec.s3.existingConfigSecret

`string`

Name of an existing Secret carrying the full S3 identity config
under the `seaweedfs_s3_config` key (the chart's contract — an
inline JSON identities document). Declares every user/credential
yourself; when set, the chart generates nothing.

### spec.s3.buckets

`[]KubernetesSeaweedFsS3Bucket`

Buckets created at install/upgrade by the chart's post-install
hook. Removing an entry later does NOT delete the bucket (or its
data) — bucket deletion is a data operation, never an IaC one.

### spec.s3.buckets[].name

`string` · required

Bucket name (S3 naming: 3-63 chars, lowercase letters, digits,
dots, hyphens; starts and ends alphanumeric).

- rule: bucket name must be 3-63 lowercase letters, digits, dots or hyphens, starting and ending alphanumeric
- rule: {"required":true}

### spec.s3.buckets[].anonymousRead

`bool`

Allow unauthenticated reads on this bucket (public objects
behind your ingress). Auth still applies to writes.

### spec.s3.buckets[].ttl

`string`

Object time-to-live, SeaweedFS TTL syntax: 1-255 followed by
m|h|d|w|M|y (e.g. "7d"). Empty = objects live forever.

- rule: ttl is a SeaweedFS TTL like "7d" — 1-255 followed by one of m, h, d, w, M, y

### spec.s3.buckets[].objectLock

`bool`

Enable S3 Object Lock on the bucket — IRREVERSIBLE and forces
versioning (WORM compliance workloads).

### spec.s3.buckets[].versioning

`bool`

Enable S3 versioning on the bucket.

### spec.s3.domainName

`string`

Host-name suffix for virtual-hosted-style requests
({bucket}.{domain_name}). Empty = path-style only (the
in-cluster norm).

### spec.s3.dedicated

`KubernetesSeaweedFsS3Dedicated`

Run the S3 gateway as its OWN Deployment instead of embedded on
the filer pods — scale the API independently of metadata.
Empty = embedded on the filer (the chart's default install
shape).

### spec.s3.dedicated.replicas

`int32` · optional (explicit presence)

Gateway replicas.

- default: `1`
- rule: {"int32":{"gte":1}}

### spec.s3.dedicated.resources

`ContainerResources`

Container resources for gateway pods.

### spec.s3.dedicated.resources.limits

`CpuMemory`

The resource limits for the container.
Specify the maximum amount of CPU and memory that the container can use.

### spec.s3.dedicated.resources.limits.cpu

`string`

### spec.s3.dedicated.resources.limits.memory

`string`

### spec.s3.dedicated.resources.requests

`CpuMemory`

The resource requests for the container.
Specify the minimum amount of CPU and memory that the container is guaranteed.

### spec.s3.dedicated.resources.requests.cpu

`string`

### spec.s3.dedicated.resources.requests.memory

`string`

### spec.replication

`string`

Cross-tier replication placement as the SeaweedFS "XYZ" code:
X replicas in other data centers, Y in other racks, Z in other
servers of the same rack (e.g. "001" = one extra copy on another
server, needs >= 2 volume servers; "010" needs rack topology).
Empty = no replication ("000"). Setting this flips the chart's
`global.seaweedfs.enableReplication` and overrides master and
filer placement together — the copies must fit the declared
volume topology or writes fail.

- rule: replication is the SeaweedFS XYZ placement code — three digits, e.g. "001"

### spec.admin

`KubernetesSeaweedFsAdmin`

Admin UI (the SeaweedFS management console: cluster state,
volumes, buckets, maintenance). Off by default. When enabled
without an existing credential Secret, the modules generate one
(`<name>-admin-auth`) — the console is never left open.

### spec.admin.enabled

`bool`

Run the admin console.

### spec.admin.existingAuthSecret

`string`

Name of an existing Secret with the console credentials under
the `user` and `password` keys. Empty = the modules generate
`<name>-admin-auth` (user "admin", random password) — the
console is never installed without authentication.

### spec.admin.dataVolume

`KubernetesSeaweedFsDataVolume`

Persist console configuration and maintenance state. Empty =
in-memory only (state lost on restart — fine for inspection-only
use).

### spec.admin.dataVolume.size

`string`

Volume size as a Kubernetes quantity (e.g. "30Gi"). Empty = the
tier's default (master 5Gi, volume 30Gi, filer 10Gi, admin
10Gi).

- rule: size must be a Kubernetes quantity like '30Gi' or '500Mi'

### spec.admin.dataVolume.storageClass

`string | valueFrom`

StorageClass for the PVC. Accepts a literal class name or a
reference to a KubernetesStorageClass resource. Empty = the
cluster's default class.

- references: KubernetesStorageClass (`status.outputs.storage_class_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesStorageClass, name: <that resource's name>, fieldPath: status.outputs.storage_class_name}} -- a bare string does not parse

### spec.admin.resources

`ContainerResources`

Container resources for the console pod.

### spec.admin.resources.limits

`CpuMemory`

The resource limits for the container.
Specify the maximum amount of CPU and memory that the container can use.

### spec.admin.resources.limits.cpu

`string`

### spec.admin.resources.limits.memory

`string`

### spec.admin.resources.requests

`CpuMemory`

The resource requests for the container.
Specify the minimum amount of CPU and memory that the container is guaranteed.

### spec.admin.resources.requests.cpu

`string`

### spec.admin.resources.requests.memory

`string`

### spec.serviceMonitorEnabled

`bool`

Create ServiceMonitors for Prometheus scraping on every enabled
tier (requires the Prometheus Operator CRDs). Chart default:
false.

### spec.image

`KubernetesSeaweedFsImage`

Override the SeaweedFS image (air-gap path). Empty = the chart's
official `chrislusf/seaweedfs` at the chart's appVersion.

### spec.image.registry

`string`

Image registry, e.g. "my.private.registry.com". Empty = Docker
Hub.

### spec.image.repository

`string`

Image repository. Empty = "chrislusf/seaweedfs" (the official
image).

### spec.image.tag

`string`

Image tag. Empty = the chart's appVersion for the pinned
chart_version.

### spec.helmValues

`string`

Escape hatch: additional chart values as a YAML document, merged
LAST over everything the typed fields render (Helm `-f`
semantics, identical on both engines). For the chart surface
beyond the typed fields (per-tier scheduling/tolerations, probes,
SFTP, the maintenance worker, mTLS, external filer stores, COSI,
the all-in-one mode, ...) — never the substitute for them. Do
not put secrets here; credential material belongs in Secrets
(see `s3.existing_config_secret`, `admin.existing_auth_secret`).

## Outputs

Reference an output from another manifest as `valueFrom: {kind: KubernetesSeaweedFs, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.namespace` | `string` | namespace the store runs in. |
| `status.outputs.release_name` | `string` | Helm release name (= metadata.name). |
| `status.outputs.s3_endpoint` | `string` | in-cluster S3 endpoint clients point their SDKs at, e.g. http://main-s3.object-store.svc.cluster.local:8333 (the filer service when the gateway is embedded, the dedicated gateway service when s3.dedicated is set). Empty when the S3 gateway is disabled. |
| `status.outputs.s3_credentials_secret_name` | `string` | name of the Secret holding the S3 credentials — keys admin_access_key_id / admin_secret_access_key / read_access_key_id / read_secret_access_key (the chart-generated `<name>-s3-secret`, or the referenced existing config secret). Empty when auth is disabled. |
| `status.outputs.filer_service_name` | `string` | name of the filer Service (file namespace HTTP API, port 8888). |
| `status.outputs.master_service_name` | `string` | name of the master Service (cluster coordination, port 9333). |
| `status.outputs.admin_endpoint` | `string` | in-cluster admin console endpoint, e.g. http://main-admin.object-store.svc.cluster.local:23646. Empty when the console is disabled. |
| `status.outputs.admin_auth_secret_name` | `string` | name of the Secret holding the admin-console credentials (keys user / password). Empty when the console is disabled. |
| `status.outputs.port_forward_command` | `string` | command to port-forward the S3 endpoint to a developer laptop, e.g. kubectl port-forward svc/main-s3 -n object-store 8333:8333 |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.namespace` | KubernetesNamespace | `spec.name` |
| `spec.master.dataVolume.storageClass` | KubernetesStorageClass | `status.outputs.storage_class_name` |
| `spec.volume.dataVolume.storageClass` | KubernetesStorageClass | `status.outputs.storage_class_name` |
| `spec.filer.dataVolume.storageClass` | KubernetesStorageClass | `status.outputs.storage_class_name` |
| `spec.admin.dataVolume.storageClass` | KubernetesStorageClass | `status.outputs.storage_class_name` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| KubernetesArgoWorkflows | `spec.artifactRepository.s3.endpoint` | `status.outputs.s3_endpoint` |
| KubernetesArgoWorkflows | `spec.artifactRepository.s3.credentialsSecret.secretName` | `status.outputs.s3_credentials_secret_name` |
| KubernetesFlinkDeployment | `spec.state.s3.endpoint` | `status.outputs.s3_endpoint` |
| KubernetesFlinkDeployment | `spec.state.s3.accessKeySecret.name` | `status.outputs.s3_credentials_secret_name` |
| KubernetesFlinkDeployment | `spec.state.s3.secretKeySecret.name` | `status.outputs.s3_credentials_secret_name` |
| KubernetesHarbor | `spec.storage.s3.endpoint` | `status.outputs.s3_endpoint` |
| KubernetesMlflow | `spec.artifactStore.s3Compatible.endpoint` | `status.outputs.s3_endpoint` |
| KubernetesMlflow | `spec.artifactStore.s3Compatible.credentialsSecret.secretName` | `status.outputs.s3_credentials_secret_name` |
| KubernetesOpenBao | `spec.backup.objectStore.s3.endpointUrl` | `status.outputs.s3_endpoint` |

## See Also

- [Overview](../README.md)
