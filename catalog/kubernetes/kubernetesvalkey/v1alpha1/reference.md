# KubernetesValkey

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `kubernetes.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

**KubernetesValkeySpec** deploys Valkey — the Linux Foundation's
Redis-compatible in-memory data store (the open-source successor every
Redis client library speaks natively) — from the OFFICIAL Valkey Helm
chart (`valkey` at https://valkey.io/valkey-helm/).

TWO TOPOLOGIES, chosen by the presence of the `replication` block:

- STANDALONE (default): one instance backed by an optional persistent
  volume. The right shape for caches and development.
- PRIMARY/REPLICA: one primary plus N replicas with streaming
  replication and a dedicated read Service. Persistence is REQUIRED in
  this mode (replicas full-sync from disk). Note what replication is
  NOT: automated failover. The chart does not ship Sentinel yet — if
  the primary pod dies, Kubernetes restarts it and replicas re-attach,
  but no replica is promoted. Durability through a primary restart
  comes from PERSISTENCE (append-only file + volume), not promotion.

NAMING CONTRACT: the Helm release and chart fullname are pinned to
`metadata.name`, so several Valkey instances coexist in one cluster.
The chart renders `<name>` (the write Service), `<name>-headless`
(pod discovery), and — in replication mode — `<name>-read` (load
balances reads across all pods).

AUTHENTICATION IS DECLARED, NEVER DEFAULTED; PASSWORDS ARE MINTED, NEVER
REQUIRED: the chart ships with auth OFF (anyone who can reach the
Service can read and write). Declare ACL users to turn it on. A user
declared without a password gets one the module generates; a declared
password (a managed-secret reference) is used as given. Either way the
credentials are materialized as a Kubernetes Secret (`<name>-auth`, one
key per username) the chart consumes, reported as `password_secret`;
they never appear in rendered values.

DURABILITY IS MODULE-OWNED: Valkey's persistence directives
(`appendonly`, `save`, `maxmemory`, ...) live in valkey.conf, which
the chart accepts only as one raw string. The typed `config` block
renders that string deterministically on both engines; the block's
`extra_directives` field appends anything beyond the typed fields.

EXPOSURE IS COMPOSED, never embedded: the store is in-cluster plumbing
reachable at the exported `kube_endpoint`. To reach it from outside,
compose a first-class exposure kind — this component never creates
one. (The service block's type/annotations exist for the LoadBalancer
arm of managed-cloud recipes, documented per environment.)

The typed fields below cover the chart's meaningful configuration
surface; `helm_values` remains as the escape hatch for chart values
beyond them (merged last, Helm `-f` semantics, identical on both
engines) — a safety valve, never the primary interface.

## Example

```yaml
# Full-surface offline-proof manifest: exercises replication (with a
# dedicated replication user, write-safety bounds, diskless sync, and an
# annotated read Service), ACL auth (default + replication + read-only
# users), TLS with mutual authentication, the module-rendered valkey.conf
# block, metrics with a ServiceMonitor, scheduling, the PodDisruptionBudget,
# image override, pull secrets, and the helm_values escape hatch — so the
# offline tofu plan and pulumi preview proofs cover the full typed surface.
# Placeholder values; never applied to a real cluster.
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesValkey
metadata:
  name: hack-valkey
spec:
  namespace:
    value: hack-valkey
  createNamespace: true
  chartVersion: 0.11.0
  image:
    registry: registry.example.com
    repository: mirrors/valkey/valkey
    tag: 9.1.1
  replication:
    replicas: 2
    persistence:
      size: 5Gi
      storageClass:
        value: fast-ssd
    replicationUser: replicator
    disklessSync: true
    minReplicasToWrite: 1
    minReplicasMaxLag: 10
    readService:
      enabled: true
      type: LoadBalancer
      annotations:
        service.beta.kubernetes.io/aws-load-balancer-internal: "true"
  config:
    appendOnly: true
    rdbSavePoints:
      - "900 1"
      - "300 10"
    maxMemory: 512mb
    maxMemoryPolicy: noeviction
    extraDirectives: |
      tcp-keepalive 60
      latency-monitor-threshold 100
  auth:
    users:
      - name: default
        password: hack-placeholder-admin
      - name: replicator
        password: hack-placeholder-repl
        permissions: "+psync +replconf +ping"
      - name: reporting
        password: hack-placeholder-read
        permissions: "~* -@all +@read +ping +info"
  tls:
    enabled: true
    certificateSecret:
      value: hack-valkey-cert
    requireClientCertificate: true
  service:
    type: ClusterIP
    port: 6380
    annotations:
      example.org/purpose: shared-cache
  resources:
    requests:
      cpu: 100m
      memory: 512Mi
    limits:
      cpu: "1"
      memory: 1Gi
  metrics:
    enabled: true
    serviceMonitorEnabled: true
  scheduling:
    nodeSelector:
      kubernetes.io/os: linux
    tolerations:
      - key: dedicated
        operator: Equal
        value: cache
        effect: NoSchedule
    priorityClassName: system-cluster-critical
  podDisruptionBudget:
    enabled: true
    maxUnavailable: 1
  logLevel: verbose
  imagePullSecrets:
    - registry-pull-secret
  helmValues: |
    commonLabels:
      example.org/owner: platform
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.namespace` | `string \| valueFrom` | yes |  | KubernetesNamespace (`spec.name`) |
| `spec.createNamespace` | `bool` |  |  |  |
| `spec.chartVersion` | `string` |  | `0.11.0` |  |
| `spec.image` | `KubernetesValkeyImage` |  |  |  |
| `spec.image.registry` | `string` |  |  |  |
| `spec.image.repository` | `string` |  |  |  |
| `spec.image.tag` | `string` |  |  |  |
| `spec.replication` | `KubernetesValkeyReplication` |  |  |  |
| `spec.replication.replicas` | `int32` |  | `2` |  |
| `spec.replication.persistence` | `KubernetesValkeyPersistence` | yes |  |  |
| `spec.replication.persistence.size` | `string` | yes |  |  |
| `spec.replication.persistence.storageClass` | `string \| valueFrom` |  |  | KubernetesStorageClass (`status.outputs.storage_class_name`) |
| `spec.replication.persistence.keepOnUninstall` | `bool` |  |  |  |
| `spec.replication.replicationUser` | `string` |  | `default` |  |
| `spec.replication.disklessSync` | `bool` |  |  |  |
| `spec.replication.minReplicasToWrite` | `int32` |  |  |  |
| `spec.replication.minReplicasMaxLag` | `int32` |  | `10` |  |
| `spec.replication.readService` | `KubernetesValkeyReadService` |  |  |  |
| `spec.replication.readService.enabled` | `bool` |  | `true` |  |
| `spec.replication.readService.type` | `string` |  | `ClusterIP` |  |
| `spec.replication.readService.annotations` | `map<string, string>` |  |  |  |
| `spec.persistence` | `KubernetesValkeyPersistence` |  |  |  |
| `spec.persistence.size` | `string` | yes |  |  |
| `spec.persistence.storageClass` | `string \| valueFrom` |  |  | KubernetesStorageClass (`status.outputs.storage_class_name`) |
| `spec.persistence.keepOnUninstall` | `bool` |  |  |  |
| `spec.config` | `KubernetesValkeyConfig` |  |  |  |
| `spec.config.appendOnly` | `bool` |  |  |  |
| `spec.config.rdbSavePoints` | `[]string` |  |  |  |
| `spec.config.snapshotsDisabled` | `bool` |  |  |  |
| `spec.config.maxMemory` | `string` |  |  |  |
| `spec.config.maxMemoryPolicy` | `string` |  |  |  |
| `spec.config.extraDirectives` | `string` |  |  |  |
| `spec.auth` | `KubernetesValkeyAuth` |  |  |  |
| `spec.auth.users` | `[]KubernetesValkeyAclUser` | yes |  |  |
| `spec.auth.users[].name` | `string` | yes |  |  |
| `spec.auth.users[].password` | `string` (sensitive) |  |  |  |
| `spec.auth.users[].permissions` | `string` |  | `~* &* +@all` |  |
| `spec.tls` | `KubernetesValkeyTls` |  |  |  |
| `spec.tls.enabled` | `bool` |  |  |  |
| `spec.tls.certificateSecret` | `string \| valueFrom` |  |  | KubernetesCertificate (`status.outputs.secret_name`) |
| `spec.tls.requireClientCertificate` | `bool` |  |  |  |
| `spec.service` | `KubernetesValkeyService` |  |  |  |
| `spec.service.type` | `string` |  | `ClusterIP` |  |
| `spec.service.port` | `int32` |  | `6379` |  |
| `spec.service.annotations` | `map<string, string>` |  |  |  |
| `spec.resources` | `ContainerResources` |  |  |  |
| `spec.resources.limits` | `CpuMemory` |  |  |  |
| `spec.resources.limits.cpu` | `string` |  |  |  |
| `spec.resources.limits.memory` | `string` |  |  |  |
| `spec.resources.requests` | `CpuMemory` |  |  |  |
| `spec.resources.requests.cpu` | `string` |  |  |  |
| `spec.resources.requests.memory` | `string` |  |  |  |
| `spec.metrics` | `KubernetesValkeyMetrics` |  |  |  |
| `spec.metrics.enabled` | `bool` |  |  |  |
| `spec.metrics.serviceMonitorEnabled` | `bool` |  |  |  |
| `spec.scheduling` | `KubernetesValkeyScheduling` |  |  |  |
| `spec.scheduling.nodeSelector` | `map<string, string>` |  |  |  |
| `spec.scheduling.tolerations` | `[]WorkloadToleration` |  |  |  |
| `spec.scheduling.tolerations[].key` | `string` |  |  |  |
| `spec.scheduling.tolerations[].operator` | `string` |  |  |  |
| `spec.scheduling.tolerations[].value` | `string` |  |  |  |
| `spec.scheduling.tolerations[].effect` | `string` |  |  |  |
| `spec.scheduling.tolerations[].tolerationSeconds` | `int64` |  |  |  |
| `spec.scheduling.priorityClassName` | `string` |  |  |  |
| `spec.podDisruptionBudget` | `KubernetesValkeyPodDisruptionBudget` |  |  |  |
| `spec.podDisruptionBudget.enabled` | `bool` |  |  |  |
| `spec.podDisruptionBudget.maxUnavailable` | `int32` |  |  |  |
| `spec.podDisruptionBudget.minAvailable` | `int32` |  |  |  |
| `spec.logLevel` | `string` |  | `notice` |  |
| `spec.imagePullSecrets` | `[]string` |  |  |  |
| `spec.helmValues` | `string` |  |  |  |

## Field Details

### spec.namespace

`string | valueFrom` · required

Namespace to deploy Valkey into. Accepts a literal namespace name or
a reference to a KubernetesNamespace resource.

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

Helm chart version to install (e.g. "0.11.0", which ships Valkey
9.1.1 — chart and app versions move separately; the chart pin
governs). Pin deliberately; upgrades re-run the release with the new
chart.

- default: `0.11.0`

### spec.image

`KubernetesValkeyImage`

Valkey container image override (registry mirrors, a specific Valkey
version). Empty = the chart default (docker.io/valkey/valkey at the
chart's app version).

### spec.image.registry

`string`

Image registry (e.g. "my-mirror.example.com"). Empty = docker.io.

### spec.image.repository

`string`

Image repository. Empty = "valkey/valkey".

### spec.image.tag

`string`

Image tag (a Valkey version, e.g. "9.1.1"). Empty = the chart's app
version.

### spec.replication

`KubernetesValkeyReplication`

Primary/replica replication. Omitted = standalone (one instance).
Present = one primary plus `replicas` replicas, a read Service, and
REQUIRED persistence (replicas bootstrap by syncing the primary's
dataset — an ephemeral primary would replicate an empty set after
every restart).

### spec.replication.replicas

`int32` · optional (explicit presence)

Number of REPLICAS (total pods = replicas + 1 primary). Chart
default: 2.

- default: `2`
- rule: {"int32":{"gte":1}}

### spec.replication.persistence

`KubernetesValkeyPersistence` · required

Persistent storage for EVERY pod (primary and replicas). Required:
replication without persistence would replicate an empty dataset
after any primary restart.

- rule: {"required":true}

### spec.replication.persistence.size

`string` · required

Volume size as a Kubernetes quantity (e.g. "5Gi").

- rule: size must be a Kubernetes quantity like '5Gi' or '500Mi'
- rule: {"required":true}

### spec.replication.persistence.storageClass

`string | valueFrom`

StorageClass for the volume. Empty = the cluster's default class.
Accepts a literal class name or a reference to a
KubernetesStorageClass resource.

- references: KubernetesStorageClass (`status.outputs.storage_class_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesStorageClass, name: <that resource's name>, fieldPath: status.outputs.storage_class_name}} -- a bare string does not parse

### spec.replication.persistence.keepOnUninstall

`bool`

STANDALONE mode only: keep the PVC (and the dataset on it) when the
release is uninstalled. Chart default: false — the data dies with
the resource.

### spec.replication.replicationUser

`string` · optional (explicit presence)

ACL user replicas authenticate to the primary with. Only meaningful
with auth enabled, and the user must exist in auth.users with
replication permissions (+psync +replconf +ping). Chart default:
"default".

- default: `default`

### spec.replication.disklessSync

`bool`

Diskless replication: replicas sync directly from the primary's
memory instead of an on-disk RDB snapshot — faster full syncs when
the primary's disk is slow, at the cost of primary memory during
the transfer.

### spec.replication.minReplicasToWrite

`int32`

Write safety: the primary refuses writes unless at least this many
replicas are connected and in sync (min-replicas-to-write). 0 (the
chart default) disables the check. Set 1+ so a partitioned primary
stops accepting writes that replicas would never see.

- rule: {"int32":{"gte":0}}

### spec.replication.minReplicasMaxLag

`int32` · optional (explicit presence)

Maximum replication lag (seconds) before a replica no longer counts
toward min_replicas_to_write. Chart default: 10.

- default: `10`
- rule: {"int32":{"gte":1}}

### spec.replication.readService

`KubernetesValkeyReadService`

The read Service (`<name>-read`, load balancing reads across ALL
pods — replicas and primary). Enabled by the chart by default in
replication mode.

### spec.replication.readService.enabled

`bool` · optional (explicit presence)

Render the read Service. Chart default: true.

- default: `true`

### spec.replication.readService.type

`string` · optional (explicit presence)

Service type: ClusterIP (default), NodePort, or LoadBalancer.

- default: `ClusterIP`
- rule: read service type must be ClusterIP, NodePort, or LoadBalancer

### spec.replication.readService.annotations

`map<string, string>`

Service annotations (per-cloud LoadBalancer recipes).

### spec.persistence

`KubernetesValkeyPersistence`

Persistent storage for the dataset. STANDALONE mode: optional —
omitted means a pure in-memory cache that starts empty after every
pod restart. REPLICATION mode: required (declared inside the
replication block instead of here).

### spec.persistence.size

`string` · required

Volume size as a Kubernetes quantity (e.g. "5Gi").

- rule: size must be a Kubernetes quantity like '5Gi' or '500Mi'
- rule: {"required":true}

### spec.persistence.storageClass

`string | valueFrom`

StorageClass for the volume. Empty = the cluster's default class.
Accepts a literal class name or a reference to a
KubernetesStorageClass resource.

- references: KubernetesStorageClass (`status.outputs.storage_class_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesStorageClass, name: <that resource's name>, fieldPath: status.outputs.storage_class_name}} -- a bare string does not parse

### spec.persistence.keepOnUninstall

`bool`

STANDALONE mode only: keep the PVC (and the dataset on it) when the
release is uninstalled. Chart default: false — the data dies with
the resource.

### spec.config

`KubernetesValkeyConfig`

The valkey.conf directives that decide durability and memory
behavior, rendered by the module into the chart's config string.

- rule: rdb_save_points and snapshots_disabled are mutually exclusive — declare snapshot points or disable snapshots, not both

### spec.config.appendOnly

`bool`

Append-only-file persistence (`appendonly yes`): every write is
logged and replayed on restart — the durability posture that makes
a pod restart lossless (paired with a persistent volume). Without
it, restarts recover only the last RDB snapshot (or nothing, if
snapshots are disabled too).

### spec.config.rdbSavePoints

`[]string`

RDB snapshot points as standard `save` directives (e.g.
"900 1" = snapshot after 900s if at least 1 key changed). Empty =
the server's built-in default schedule. To DISABLE snapshots
entirely, set snapshots_disabled instead.

- rule: {"repeated":{"items":{"cel":[{"id":"spec.config.rdb_save_point_format","message":"each RDB save point is '<seconds> <changes>' — two positive integers, e.g. '900 1'","expression":"this.matches('^[0-9]+ [0-9]+$')"}]}}}

### spec.config.snapshotsDisabled

`bool`

Disable RDB snapshots (`save \"\"`). Mutually exclusive with
rdb_save_points. For pure caches (no persistence volume) and
AOF-only postures.

### spec.config.maxMemory

`string`

Memory ceiling for the dataset as a Valkey size (e.g. "256mb",
"1gb"). Empty = unlimited (the server default) — the pod's memory
LIMIT then becomes the only bound, and hitting it is an OOM kill,
not an eviction. Always set this for caches.

- rule: max_memory is a Valkey size like '256mb' or '1gb'

### spec.config.maxMemoryPolicy

`string`

What happens at max_memory: noeviction (writes fail — the server
default, right for durable stores), allkeys-lru / volatile-lru,
allkeys-lfu / volatile-lfu, allkeys-random / volatile-random, or
volatile-ttl (right for caches).

- rule: max_memory_policy must be one of noeviction, allkeys-lru, volatile-lru, allkeys-lfu, volatile-lfu, allkeys-random, volatile-random, volatile-ttl

### spec.config.extraDirectives

`string`

Additional raw valkey.conf directives appended verbatim after the
typed ones (one per line). The in-config escape hatch — for
directives beyond the typed fields, never a replacement for them.

### spec.auth

`KubernetesValkeyAuth`

ACL authentication. Omitted = AUTH OFF (the chart default — anyone
with network reach has full access; acceptable only behind
NetworkPolicies or for development). Declare users to require
credentials; each user's password — declared, or generated by the
module when left empty — lands in the `<name>-auth` Secret.

### spec.auth.users

`[]KubernetesValkeyAclUser` · required

ACL users to create. MUST include the "default" user — without an
explicit default user, unauthenticated clients retain full access
and enabling auth is meaningless (the chart's own warning). A user
needs only a name: the module generates a password for every user
declared without one. Passwords are materialized as the
`<name>-auth` Kubernetes Secret (one key per username) that the
chart consumes via usersExistingSecret.

- rule: each ACL user needs a distinct name
- rule: auth requires the 'default' user to be declared — without it, unauthenticated clients keep full access
- rule: {"repeated":{"minItems":"1"}}

### spec.auth.users[].name

`string` · required

Username ("default" is the user unauthenticated clients and plain
AUTH <password> map to).

- rule: {"required":true}

### spec.auth.users[].password

`string` · sensitive

Password. Optional: leave it empty and the module generates one
(32 letters and digits, no symbols — the chart's init script and
client URLs never have to quote it) and keeps it stable across
re-applies; declare one (as a managed-secret reference) to bring
your own. Either way it lands ONLY in the `<name>-auth` Secret
under this user's name — never plaintext in rendered chart values —
and clients read it from there (the `password_secret` output names
the `default` user's key).

### spec.auth.users[].permissions

`string` · optional (explicit presence)

ACL permission rule string. Empty = full access ("~* &* +@all" —
right for the default admin user). Examples: read-only
"~* -@all +@read +ping +info"; replication user
"+psync +replconf +ping".

- default: `~* &* +@all`

### spec.tls

`KubernetesValkeyTls`

TLS for client connections. Requires an existing kubernetes.io/tls
Secret — the cert-manager seam: issue a KubernetesCertificate and
wire its secret here.

- rule: TLS requires certificate_secret — the kubernetes.io/tls Secret holding the server certificate
- rule: require_client_certificate only applies when TLS is enabled

### spec.tls.enabled

`bool`

Serve TLS. Requires the certificate Secret below.

### spec.tls.certificateSecret

`string | valueFrom`

Name of a kubernetes.io/tls Secret with the server certificate,
key, and CA. Accepts a literal Secret name or a reference to a
KubernetesCertificate resource (the cert-manager seam).

- references: KubernetesCertificate (`status.outputs.secret_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesCertificate, name: <that resource's name>, fieldPath: status.outputs.secret_name}} -- a bare string does not parse

### spec.tls.requireClientCertificate

`bool`

Require clients to present a certificate (mutual TLS).

### spec.service

`KubernetesValkeyService`

The write Service (named `<name>`). ClusterIP by default; the
LoadBalancer arm carries per-cloud annotations (NLB, internal LB,
...) for managed clusters.

### spec.service.type

`string` · optional (explicit presence)

Service type: ClusterIP (default), NodePort, or LoadBalancer.

- default: `ClusterIP`
- rule: service type must be ClusterIP, NodePort, or LoadBalancer

### spec.service.port

`int32` · optional (explicit presence)

Service port. Chart default: 6379.

- default: `6379`
- rule: {"int32":{"lte":65535,"gte":1}}

### spec.service.annotations

`map<string, string>`

Service annotations — the per-cloud LoadBalancer recipe surface
(internal LBs, NLB, source ranges via the chart fields, ...).

### spec.resources

`ContainerResources`

CPU/memory for the Valkey container. Empty = no requests/limits
(the chart default — fine for evaluation, not production). Size
memory ABOVE `config.max_memory`: Valkey needs headroom for
replication buffers and fork-based persistence.

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

### spec.metrics

`KubernetesValkeyMetrics`

Prometheus metrics via the redis_exporter sidecar the chart ships.

- rule: a ServiceMonitor scrapes the exporter sidecar — enable metrics with it

### spec.metrics.enabled

`bool`

Run the redis_exporter sidecar and its metrics Service.

### spec.metrics.serviceMonitorEnabled

`bool`

Create a ServiceMonitor for the Prometheus operator. Requires the
Prometheus operator CRDs on the cluster — the release FAILS to
install without them.

### spec.scheduling

`KubernetesValkeyScheduling`

Where the pods run: node selection, tolerations, and scheduling
priority.

### spec.scheduling.nodeSelector

`map<string, string>`

Node selector for the pods.

### spec.scheduling.tolerations

`[]WorkloadToleration`

Tolerations for the pods.

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

PriorityClass for the pods — a shared cache should outlive
stateless workloads under node pressure.

### spec.podDisruptionBudget

`KubernetesValkeyPodDisruptionBudget`

PodDisruptionBudget for voluntary disruptions. The chart only
renders it in REPLICATION mode (a single standalone pod cannot
usefully budget disruptions).

- rule: declare at most one PDB bound — max_unavailable or min_available

### spec.podDisruptionBudget.enabled

`bool`

Render the PDB.

### spec.podDisruptionBudget.maxUnavailable

`int32`

Maximum pods down during voluntary disruptions. Chart default: 1.
Mutually exclusive with min_available.

- rule: {"int32":{"gte":0}}

### spec.podDisruptionBudget.minAvailable

`int32`

Minimum pods that must stay up. Takes precedence over
max_unavailable in the chart — declare exactly one here.

- rule: {"int32":{"gte":0}}

### spec.logLevel

`string` · optional (explicit presence)

Server log verbosity: debug, verbose, notice (the upstream
default), or warning.

- default: `notice`
- rule: log_level must be one of debug, verbose, notice, warning

### spec.imagePullSecrets

`[]string`

Names of image-pull secrets (in the namespace) for pulling the
Valkey image from a private mirror.

### spec.helmValues

`string`

Escape hatch: additional chart values as a YAML document, merged
LAST over everything the typed fields render (Helm `-f` semantics,
identical on both engines). For the chart surface beyond the typed
fields (probes, security contexts, topology spread, network policy,
extra volumes, ...) — never the substitute for them. Do not put
secrets here.

## Validation Rules

- `spec.replication_persistence_placement`: in replication mode, persistence is declared INSIDE the replication block (it applies to every pod) — the top-level persistence field is for standalone mode only

## Outputs

Reference an output from another manifest as `valueFrom: {kind: KubernetesValkey, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.namespace` | `string` | Namespace the instance runs in. |
| `status.outputs.service` | `string` | The WRITE Service name (`<metadata.name>`). In replication mode it always targets the primary's role; standalone, it targets the one instance. |
| `status.outputs.read_service` | `string` | The READ Service name (`<metadata.name>-read`) — replication mode with the read service enabled; empty otherwise. Load balances reads across all pods. |
| `status.outputs.headless_service` | `string` | The headless Service name (`<metadata.name>-headless`) for direct pod discovery. REPLICATION MODE ONLY — the chart renders no headless Service for a standalone instance (a Deployment, not a StatefulSet), so this is empty standalone. |
| `status.outputs.kube_endpoint` | `string` | In-cluster endpoint of the write Service: `<service>.<namespace>.svc.cluster.local:<port>`. |
| `status.outputs.port_forward_command` | `string` | kubectl port-forward one-liner for reaching the store from a workstation. |
| `status.outputs.username` | `string` | The ACL username applications authenticate with ("default" when auth is declared; empty when auth is off). |
| `status.outputs.password_secret` | `KubernetesSecretKey` | The Kubernetes Secret key holding that user's password (the module-materialized `<metadata.name>-auth` Secret). Unset when auth is off. |
| `status.outputs.password_secret.name` | `string` | The name of the Kubernetes Secret. |
| `status.outputs.password_secret.key` | `string` | The key within the Kubernetes Secret. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.namespace` | KubernetesNamespace | `spec.name` |
| `spec.replication.persistence.storageClass` | KubernetesStorageClass | `status.outputs.storage_class_name` |
| `spec.persistence.storageClass` | KubernetesStorageClass | `status.outputs.storage_class_name` |
| `spec.tls.certificateSecret` | KubernetesCertificate | `status.outputs.secret_name` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| KubernetesAirflow | `spec.broker.valkey.host` | `status.outputs.service` |
| KubernetesAirflow | `spec.broker.valkey.passwordSecret.secretName` | `status.outputs.password_secret.name` |
| KubernetesArgocd | `spec.redis.external.host` | `status.outputs.service` |
| KubernetesHarbor | `spec.cache.external.addr` | `status.outputs.kube_endpoint` |
| KubernetesRayCluster | `spec.gcsFaultTolerance.redisAddress` | `status.outputs.kube_endpoint` |
| KubernetesRayCluster | `spec.gcsFaultTolerance.redisPasswordSecret.name` | `status.outputs.password_secret.name` |
| KubernetesSuperset | `spec.cache.host` | `status.outputs.service` |
| KubernetesSuperset | `spec.cache.passwordSecret.secretName` | `status.outputs.password_secret.name` |

## See Also

- [Overview](../README.md)
