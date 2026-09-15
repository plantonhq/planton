# KubernetesClickHouse

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `kubernetes.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

**KubernetesClickHouseSpec** declares one ClickHouse cluster — the
columnar OLAP database — as a `ClickHouseInstallation` (CHI) custom
resource reconciled by the Altinity ClickHouse operator (declare the
operator with KubernetesAltinityOperator; it is a registry
prerequisite of this kind). The operator renders every shard×replica
host as its own single-pod StatefulSet with generated ClickHouse
configuration mounted from ConfigMaps.

TOPOLOGY: `shards` × `replicas` hosts. Shards split the data for
parallel query processing (Distributed-engine tables fan queries out
across shards); replicas within a shard hold copies of the same data
via ReplicatedMergeTree engines. Replication (replicas > 1) and
`ON CLUSTER` DDL both require a coordination service (ClickHouse
Keeper or ZooKeeper) — see `coordination`.

NAMING CONTRACT (operator patterns, read from the operator source at
the pinned release): the cluster-wide client Service is
`clickhouse-<name>` (ClusterIP), the per-cluster Service is
`cluster-<name>-<cluster_name>`, and every host's StatefulSet and
headless Service are `chi-<name>-<cluster_name>-<shard>-<replica>`.
Kubernetes caps Service names at 63 characters, so keep
`metadata.name` within 48 characters with the default `cluster_name`
("main"); a longer `cluster_name` shrinks that budget one-for-one.

EXPOSURE: no ingress resources are created here. All generated
Services are ClusterIP (the operator's own default — verified in the
operator source); compose external exposure from first-class kinds
(KubernetesIngress, Gateway API kinds) over the exported
`service_name`, or set `service_annotations` for in-cluster
service-mesh/LB annotations.

SERVER CONFIGURATION LAYERS: fully typed fields cover topology,
storage, users, and placement. ClickHouse's own configuration
vocabulary (hundreds of server settings, per-profile settings, quota
intervals, raw config-file drop-ins) is passed through the CHI's own
path-keyed maps — `settings`, `profiles`, `quotas`, `files` — where
keys use `/`-separated XML paths exactly as the upstream CRD defines
them. Those maps are the upstream's native model, not an escape
hatch bolted on top of one.

## Example

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesClickHouse
metadata:
  name: clickhouse-dev
spec:
  namespace:
    value: clickhouse-dev
  createNamespace: true
  version: "25.3"
  shards: 1
  replicas: 2
  diskSize: 10Gi
  resources:
    requests:
      cpu: 250m
      memory: 1Gi
    limits:
      cpu: "1"
      memory: 2Gi
  users:
    - name: app
      password:
        value: Ch4ngeMe-app
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.namespace` | `string \| valueFrom` | yes |  | KubernetesNamespace (`spec.name`) |
| `spec.createNamespace` | `bool` |  |  |  |
| `spec.version` | `string` | yes |  |  |
| `spec.image` | `ContainerImage` |  |  |  |
| `spec.image.repo` | `string` |  |  |  |
| `spec.image.tag` | `string` |  |  |  |
| `spec.image.pullSecretName` | `string` |  |  |  |
| `spec.clusterName` | `string` |  | `main` |  |
| `spec.shards` | `int32` |  | `1` |  |
| `spec.replicas` | `int32` |  | `1` |  |
| `spec.resources` | `ContainerResources` |  |  |  |
| `spec.resources.limits` | `CpuMemory` |  |  |  |
| `spec.resources.limits.cpu` | `string` |  |  |  |
| `spec.resources.limits.memory` | `string` |  |  |  |
| `spec.resources.requests` | `CpuMemory` |  |  |  |
| `spec.resources.requests.cpu` | `string` |  |  |  |
| `spec.resources.requests.memory` | `string` |  |  |  |
| `spec.diskSize` | `string` | yes |  |  |
| `spec.storageClass` | `string \| valueFrom` |  |  | KubernetesStorageClass (`status.outputs.storage_class_name`) |
| `spec.logDiskSize` | `string` |  |  |  |
| `spec.retainVolumesOnDelete` | `bool` |  |  |  |
| `spec.coordination` | `KubernetesClickHouseCoordination` |  |  |  |
| `spec.coordination.type` | `enum` |  |  |  |
| `spec.coordination.keeper` | `KubernetesClickHouseManagedKeeper` |  |  |  |
| `spec.coordination.keeper.replicas` | `int32` |  | `3` |  |
| `spec.coordination.keeper.resources` | `ContainerResources` |  |  |  |
| `spec.coordination.keeper.resources.limits` | `CpuMemory` |  |  |  |
| `spec.coordination.keeper.resources.limits.cpu` | `string` |  |  |  |
| `spec.coordination.keeper.resources.limits.memory` | `string` |  |  |  |
| `spec.coordination.keeper.resources.requests` | `CpuMemory` |  |  |  |
| `spec.coordination.keeper.resources.requests.cpu` | `string` |  |  |  |
| `spec.coordination.keeper.resources.requests.memory` | `string` |  |  |  |
| `spec.coordination.keeper.diskSize` | `string` |  | `10Gi` |  |
| `spec.coordination.keeper.storageClass` | `string \| valueFrom` |  |  | KubernetesStorageClass (`status.outputs.storage_class_name`) |
| `spec.coordination.external` | `KubernetesClickHouseExternalCoordination` |  |  |  |
| `spec.coordination.external.nodes` | `[]KubernetesClickHouseCoordinationNode` |  |  |  |
| `spec.coordination.external.nodes[].host` | `string` | yes |  |  |
| `spec.coordination.external.nodes[].port` | `int32` |  | `2181` |  |
| `spec.coordination.external.root` | `string` |  |  |  |
| `spec.coordination.external.identity` | `string` (sensitive) |  |  |  |
| `spec.users` | `[]KubernetesClickHouseUser` |  |  |  |
| `spec.users[].name` | `string` | yes |  |  |
| `spec.users[].password` | `string \| valueFrom` (sensitive) | yes |  |  |
| `spec.users[].profile` | `string` |  |  |  |
| `spec.users[].quota` | `string` |  |  |  |
| `spec.users[].networks` | `[]string` |  |  |  |
| `spec.users[].grants` | `[]string` |  |  |  |
| `spec.users[].accessManagement` | `bool` |  |  |  |
| `spec.users[].settings` | `map<string, string>` |  |  |  |
| `spec.profiles` | `[]KubernetesClickHouseNamedSettings` |  |  |  |
| `spec.profiles[].name` | `string` | yes |  |  |
| `spec.profiles[].settings` | `map<string, string>` |  |  |  |
| `spec.quotas` | `[]KubernetesClickHouseNamedSettings` |  |  |  |
| `spec.quotas[].name` | `string` | yes |  |  |
| `spec.quotas[].settings` | `map<string, string>` |  |  |  |
| `spec.settings` | `map<string, string>` |  |  |  |
| `spec.files` | `map<string, string>` |  |  |  |
| `spec.autoInterNodeSecret` | `bool` |  | `true` |  |
| `spec.spreadReplicasAcrossNodes` | `bool` |  |  |  |
| `spec.pdbMaxUnavailable` | `int32` |  | `1` |  |
| `spec.nodeSelector` | `map<string, string>` |  |  |  |
| `spec.tolerations` | `[]WorkloadToleration` |  |  |  |
| `spec.tolerations[].key` | `string` |  |  |  |
| `spec.tolerations[].operator` | `string` |  |  |  |
| `spec.tolerations[].value` | `string` |  |  |  |
| `spec.tolerations[].effect` | `string` |  |  |  |
| `spec.tolerations[].tolerationSeconds` | `int64` |  |  |  |
| `spec.serviceAnnotations` | `map<string, string>` |  |  |  |
| `spec.stopped` | `bool` |  |  |  |
| `spec.imagePullSecrets` | `[]string` |  |  |  |

## Field Details

### spec.namespace

`string | valueFrom` · required

Namespace to deploy the cluster into. Accepts a literal namespace
name or a reference to a KubernetesNamespace resource. The
ClickHouseInstallation is namespaced; the Altinity operator must
be watching this namespace (see the operator kind's
`watch_namespaces`).

- references: KubernetesNamespace (`spec.name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesNamespace, name: <that resource's name>, fieldPath: spec.name}} -- a bare string does not parse

### spec.createNamespace

`bool`

When true, the namespace is created (with the standard Planton
governance labels) before deploying and deleted with the resource.
When false, the namespace must already exist.

### spec.version

`string` · required

ClickHouse server version tag, e.g. "25.3" (an LTS line) or
"24.8". Resolves to the `clickhouse/clickhouse-server` image
unless `image` overrides the repository. Always pin a version:
the operator's built-in fallback is the `latest` tag, which makes
cluster upgrades happen implicitly on pod restarts.

- rule: {"required":true}

### spec.image

`ContainerImage`

Override the ClickHouse server image (air-gap / private-mirror
path). `repo` empty = "clickhouse/clickhouse-server"; `tag` empty
= `version`.

### spec.image.repo

`string`

The repository of the image (e.g., "gcr.io/project/image").

### spec.image.tag

`string`

The tag of the image (e.g., "latest" or "1.0.0").

### spec.image.pullSecretName

`string`

The name of the image pull secret for private image repositories.

### spec.clusterName

`string` · optional (explicit presence)

Logical ClickHouse cluster name — the entry in `remote_servers`
that Distributed tables and `ON CLUSTER '<name>'` DDL statements
target. Also a segment of every generated child name (see the
naming contract above), which is why the operator's CRD caps it at
15 characters. Default: "main".

Verified live: the in-server cluster definition propagates through
mounted config and can LAG the installation reaching Completed —
`ON CLUSTER` DDL initiated in that window silently executes on the
subset of hosts the initiator can see (and still returns success).
Before running distributed DDL right after a deploy or topology
change, confirm `SELECT count() FROM system.clusters WHERE cluster
= '<name>'` reports every declared host.

- default: `main`
- rule: {"string":{"pattern":"^[a-z0-9]([-a-z0-9]{0,13}[a-z0-9])?$"}}

### spec.shards

`int32` · optional (explicit presence)

Number of shards. Each shard holds a disjoint slice of the data;
queries against Distributed-engine tables run on all shards in
parallel. 1 (the default) means all data on every host — scale
this only for datasets or write rates a single shard cannot carry.

- default: `1`
- rule: {"int32":{"gte":1}}

### spec.replicas

`int32` · optional (explicit presence)

Replicas per shard. Each replica is a full copy of its shard's
data, kept in sync through ReplicatedMergeTree — which requires
`coordination`. 1 (the default) means no redundancy: a lost volume
loses that shard's data. Production: 2–3.

- default: `1`
- rule: {"int32":{"gte":1}}

### spec.resources

`ContainerResources`

CPU and memory for each ClickHouse host container. Empty = no
requests/limits (schedulable anywhere, no protection). ClickHouse
is memory-hungry under analytical load: give production hosts at
least 4Gi and set `max_server_memory_usage_to_ram_ratio` (via
`settings`) if the limit is tight.

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

### spec.diskSize

`string` · required

Size of the persistent data volume for EACH host (e.g. "100Gi").
Rendered as the data VolumeClaimTemplate mounted at
/var/lib/clickhouse. Kubernetes cannot shrink PVCs, and expanding
requires a storage class that allows it — plan for growth.

- rule: {"required":true,"string":{"pattern":"^\\d+(\\.\\d+)?\\s?(Ki|Mi|Gi|Ti|Pi|Ei|K|M|G|T|P|E)$"}}

### spec.storageClass

`string | valueFrom`

Storage class for the data volumes. Accepts a literal name or a
reference to a KubernetesStorageClass resource. Empty = the
cluster's default storage class.

- references: KubernetesStorageClass (`status.outputs.storage_class_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesStorageClass, name: <that resource's name>, fieldPath: status.outputs.storage_class_name}} -- a bare string does not parse

### spec.logDiskSize

`string`

Size of a SEPARATE persistent volume for server logs (e.g.
"10Gi"), mounted at /var/log/clickhouse-server. Empty (default) =
logs live on the container filesystem and vanish with the pod.
Useful when log retention matters and no log shipper is in place.

- rule: log_disk_size must be a Kubernetes quantity like 10Gi or 512Mi

### spec.retainVolumesOnDelete

`bool`

Keep the data volumes when the resource (or a host) is deleted.
Maps to the operator's PVC reclaim policy: false (the operator
default) deletes PVCs with their StatefulSet; true retains them,
so a re-created cluster with the same name re-attaches the data.
Retained PVCs are never garbage-collected — deleting them becomes
a manual operation.

### spec.coordination

`KubernetesClickHouseCoordination`

Coordination service for replication and distributed DDL.
UNSET (recommended): the module deploys a managed ClickHouse
Keeper (3 nodes) automatically whenever the topology needs one
(replicas > 1 or shards > 1) and none otherwise. Set explicitly to
size the managed Keeper, point at external coordination, or opt
out entirely.

- rule: external coordination (external_keeper / external_zookeeper) requires external.nodes with at least one host entry

### spec.coordination.type

`enum`

Coordination flavor. See the enum values for semantics.

Allowed values (use exactly as shown):

- `unspecified` -- Auto: managed Keeper when the topology needs coordination (replicas > 1 or shards > 1), none otherwise.
- `managed_keeper` -- Module-managed ClickHouse Keeper: a ClickHouseKeeperInstallation reconciled by the same operator, wired to the cluster through the CHI's native keeper reference (the operator resolves the endpoints itself). Keeper is ClickHouse's own Raft-based ZooKeeper replacement — markedly lighter than ZooKeeper (no JVM), protocol-compatible, and the upstream-recommended default.
- `external_keeper` -- Existing ClickHouse Keeper ensemble reachable at `external.nodes` — for shared coordination infrastructure.
- `external_zookeeper` -- Existing ZooKeeper ensemble reachable at `external.nodes` — for legacy or Kafka-shared ZooKeeper deployments.
- `none` -- No coordination. Valid only for single-replica topologies; multi-shard single-replica clusters lose `ON CLUSTER` DDL (run DDL per shard instead) but Distributed queries keep working.

### spec.coordination.keeper

`KubernetesClickHouseManagedKeeper`

Managed Keeper sizing. Only read when the effective type is
managed_keeper; empty = 3 replicas with operator-default
resources and a 10Gi volume.

### spec.coordination.keeper.replicas

`int32` · optional (explicit presence)

Keeper ensemble size. Raft quorum needs an odd count: 1 (dev — no
fault tolerance), 3 (production — survives one node loss), or 5
(survives two). Default 3.

- default: `3`
- rule: {"int32":{"in":[1,3,5]}}

### spec.coordination.keeper.resources

`ContainerResources`

CPU and memory per Keeper pod. Keeper is deliberately light
(C++, no JVM): 100m/256Mi requests with 500m/1Gi limits carry most
clusters. Empty = operator defaults.

### spec.coordination.keeper.resources.limits

`CpuMemory`

The resource limits for the container.
Specify the maximum amount of CPU and memory that the container can use.

### spec.coordination.keeper.resources.limits.cpu

`string`

### spec.coordination.keeper.resources.limits.memory

`string`

### spec.coordination.keeper.resources.requests

`CpuMemory`

The resource requests for the container.
Specify the minimum amount of CPU and memory that the container is guaranteed.

### spec.coordination.keeper.resources.requests.cpu

`string`

### spec.coordination.keeper.resources.requests.memory

`string`

### spec.coordination.keeper.diskSize

`string` · optional (explicit presence)

Persistent volume per Keeper pod for the coordination log and
snapshots (metadata only, not table data). Default "10Gi".

- default: `10Gi`
- rule: {"string":{"pattern":"^\\d+(\\.\\d+)?\\s?(Ki|Mi|Gi|Ti|Pi|Ei|K|M|G|T|P|E)$"}}

### spec.coordination.keeper.storageClass

`string | valueFrom`

Storage class for the Keeper volumes. Empty = the cluster's
default storage class.

- references: KubernetesStorageClass (`status.outputs.storage_class_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesStorageClass, name: <that resource's name>, fieldPath: status.outputs.storage_class_name}} -- a bare string does not parse

### spec.coordination.external

`KubernetesClickHouseExternalCoordination`

External coordination endpoints. Required when type is
external_keeper or external_zookeeper.

### spec.coordination.external.nodes

`[]KubernetesClickHouseCoordinationNode`

Ensemble nodes. List every member for client-side failover.

### spec.coordination.external.nodes[].host

`string` · required

DNS name or IP of the node, e.g.
"zk-0.zk-headless.zoo.svc.cluster.local".

- rule: {"required":true}

### spec.coordination.external.nodes[].port

`int32` · optional (explicit presence)

Client port. Default 2181 — the standard for both ZooKeeper and
ClickHouse Keeper.

- default: `2181`
- rule: {"int32":{"lte":65535,"gte":1}}

### spec.coordination.external.root

`string`

Optional root znode path under which this cluster stores its
replication and DDL metadata — set it when several ClickHouse
clusters share one ensemble (e.g. "/clickhouse/prod-analytics").

### spec.coordination.external.identity

`string` · sensitive

Optional digest-auth credentials for the ensemble in
"user:password" form (ZooKeeper ACL identity). Delivered into the
generated config verbatim.

### spec.users

`[]KubernetesClickHouseUser`

ClickHouse users to provision, each with a password delivered
through a Kubernetes Secret (never plaintext in the CHI). The
built-in `default` user stays operator-managed: passwordless but
network-restricted to the cluster's own pods (the operator
generates the IP allowlist) — create named users for every real
client. KNOW THIS (upstream-documented): secret-sourced passwords
reach ClickHouse through pod environment variables, so rotating
the Secret alone does not re-render config — bump any spec field
(the operator re-reconciles on CHI change) to roll a rotation out.

### spec.users[].name

`string` · required

User name (becomes the XML section name and the Secret key).

- rule: {"required":true,"string":{"pattern":"^[a-zA-Z0-9_]([a-zA-Z0-9_-]{0,61}[a-zA-Z0-9_])?$"}}

### spec.users[].password

`string | valueFrom` · required · sensitive

Password. Accepts a literal value or a reference to another
resource's output. Required — passwordless named users are not
provisioned by this kind.

- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.users[].profile

`string`

Settings profile this user runs under (a name from `profiles`, or
a profile defined via `files`/upstream defaults such as
"default").

### spec.users[].quota

`string`

Quota applied to this user (a name from `quotas`).

### spec.users[].networks

`[]string`

Networks the user may connect from, as IPs/CIDRs (e.g.
"10.0.0.0/8") or "::/0" for anywhere. KNOW THIS (verified live):
empty is NOT ClickHouse's own any-network default — the operator
normalizes a user declared without networks to a restrictive
fence (the cluster's own pods by host_regexp, plus ::1 and
127.0.0.1), and ClickHouse reports the network rejection as
"Authentication failed: password is incorrect, or there is no
user with such name" — indistinguishable from a wrong password.
Port-forwarded connections arrive as localhost and slip through
the fence, which is why a smoke test can pass while every
in-cluster client fails. Declare networks explicitly for every
user a workload connects as (e.g. "0.0.0.0/0" and "::/0" when the
password is the gate).

### spec.users[].grants

`[]string`

SQL GRANT statements executed for the user at config render, e.g.
"GRANT SELECT ON analytics.*". The declarative alternative to
running GRANTs by hand; requires `access_management` on an admin
user only for runtime-managed grants, not for these.

Two access behaviors verified live (ClickHouse 25.3): a user
declared with NO grants receives ClickHouse's default,
UNRESTRICTED access (config-file user semantics) — declare grants
to constrain a user, never to widen one. And once grants are
declared, distributed DDL (`... ON CLUSTER`) additionally requires
"GRANT CLUSTER ON *.*" — CREATE on the target database alone is
rejected with ACCESS_DENIED for ON CLUSTER statements.

### spec.users[].accessManagement

`bool`

Allow this user to manage users and grants at runtime via SQL
(CREATE USER / GRANT). Reserve for administrative users.

### spec.users[].settings

`map<string, string>`

Per-user setting overrides. Keys are `/`-separated paths within
the user's XML section, e.g. "max_memory_usage" = "10000000000".

### spec.profiles

`[]KubernetesClickHouseNamedSettings`

Settings profiles: named bundles of query-level settings users
reference by profile name (e.g. a "readonly" profile with
`readonly: "1"`). Keys are `/`-separated setting paths within the
profile, exactly as the upstream CRD's `profiles` section takes
them.

### spec.profiles[].name

`string` · required

Bundle name (profile name or quota name) referenced from `users`.

- rule: {"required":true,"string":{"pattern":"^[a-zA-Z0-9_]([a-zA-Z0-9_-]{0,61}[a-zA-Z0-9_])?$"}}

### spec.profiles[].settings

`map<string, string>`

Path-keyed values within the bundle, exactly as the upstream CRD
takes them (profiles: setting paths; quotas: "interval/…" paths).

### spec.quotas

`[]KubernetesClickHouseNamedSettings`

Quotas: named resource-consumption limits users reference by quota
name. Keys are `/`-separated paths within the quota, e.g.
"interval/duration" = "3600", "interval/queries" = "10000".

### spec.quotas[].name

`string` · required

Bundle name (profile name or quota name) referenced from `users`.

- rule: {"required":true,"string":{"pattern":"^[a-zA-Z0-9_]([a-zA-Z0-9_-]{0,61}[a-zA-Z0-9_])?$"}}

### spec.quotas[].settings

`map<string, string>`

Path-keyed values within the bundle, exactly as the upstream CRD
takes them (profiles: setting paths; quotas: "interval/…" paths).

### spec.settings

`map<string, string>`

Server-level settings rendered into config.d for every host. Keys
are `/`-separated XML paths exactly as the upstream CRD's
`settings` section takes them, e.g.
"compression/case/method" = "zstd",
"merge_tree/max_suspicious_broken_parts" = "5",
"max_concurrent_queries" = "200".

### spec.files

`map<string, string>`

Raw configuration file drop-ins: file name → full file content.
Rendered into the generated ConfigMaps alongside the operator's
own files. A file name may carry the upstream's placement prefix
({common}, {users}, {hosts}) to choose between config.d, users.d
and conf.d — unprefixed names land in config.d.

### spec.autoInterNodeSecret

`bool` · optional (explicit presence)

Secure distributed queries between this cluster's own hosts with
an operator-generated shared secret (the CHI `secret.auto`
mechanism). Default true; rendered only when the topology has more
than one host. Disable only for ClickHouse versions below 20.10,
which predate the mechanism.

- default: `true`

### spec.spreadReplicasAcrossNodes

`bool`

Never schedule two replicas of the same shard on the same
Kubernetes node (the operator's ShardAntiAffinity pod
distribution). Off by default so single-node dev clusters
schedule; turn on in production — co-located replicas make
replication pointless against node loss.

### spec.pdbMaxUnavailable

`int32` · optional (explicit presence)

How many of this cluster's pods may be voluntarily evicted at
once (PodDisruptionBudget maxUnavailable, one PDB per cluster,
operator-managed). Default 1.

- default: `1`
- rule: {"int32":{"gte":0}}

### spec.nodeSelector

`map<string, string>`

Node selector for every ClickHouse host pod.

### spec.tolerations

`[]WorkloadToleration`

Tolerations for every ClickHouse host pod.

### spec.tolerations[].key

`string`

Taint key to tolerate. Empty key with operator "Exists" tolerates every taint.

### spec.tolerations[].operator

`string`

How key/value match: "Equal" (default — value must match too) or "Exists"
(key presence alone matches).

- rule: Toleration operator must be either "Equal" or "Exists"

### spec.tolerations[].value

`string`

Taint value to match when operator is "Equal".

### spec.tolerations[].effect

`string`

Which taint effect is tolerated: "NoSchedule", "PreferNoSchedule", or
"NoExecute". Empty tolerates all effects for the key.

- rule: Toleration effect must be one of "NoSchedule", "PreferNoSchedule", or "NoExecute"

### spec.tolerations[].tolerationSeconds

`int64` · optional (explicit presence)

For "NoExecute" taints only: how many seconds already-running pods stay bound
after the taint appears. Unset means tolerate forever.

### spec.serviceAnnotations

`map<string, string>`

Annotations for the cluster-wide client Service `clickhouse-<name>`
(internal LB or service-mesh annotations). The Service stays
ClusterIP — compose external exposure from first-class kinds.

### spec.stopped

`bool`

Stop the cluster without losing data (the CHI `stop` verb): true
scales every host StatefulSet to zero — pods and Services go away,
every PVC stays — and false brings the same data back. The
declarative pause switch for expensive dev/staging clusters.

### spec.imagePullSecrets

`[]string`

Names of image-pull secrets (in the deployment namespace) for
pulling images from a private mirror.

## Validation Rules

- `spec.coordination.required_for_replication`: replicas > 1 requires coordination — ReplicatedMergeTree cannot sync without ClickHouse Keeper or ZooKeeper; leave coordination unset to get a managed Keeper automatically, or configure it explicitly (type none is only valid for single-replica topologies)

## Outputs

Reference an output from another manifest as `valueFrom: {kind: KubernetesClickHouse, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.namespace` | `string` | namespace the cluster runs in. |
| `status.outputs.chi_name` | `string` | name of the ClickHouseInstallation resource (= metadata.name). |
| `status.outputs.cluster_name` | `string` | logical ClickHouse cluster name (the `ON CLUSTER` / remote_servers target), e.g. main. |
| `status.outputs.service_name` | `string` | name of the cluster-wide client Service covering all hosts (operator naming contract: clickhouse-<name>). |
| `status.outputs.tcp_endpoint` | `string` | in-cluster native-protocol endpoint (clickhouse-client, drivers), e.g. clickhouse-analytics.data.svc.cluster.local:9000 |
| `status.outputs.http_endpoint` | `string` | in-cluster HTTP interface endpoint (curl, JDBC/ODBC over HTTP), e.g. http://clickhouse-analytics.data.svc.cluster.local:8123 |
| `status.outputs.auth_secret_name` | `string` | name of the module-managed Secret holding the provisioned users' passwords (one key per user name), e.g. <name>-clickhouse-auth. Empty when no users are declared. |
| `status.outputs.keeper_name` | `string` | name of the managed ClickHouseKeeperInstallation resource, e.g. <name>-keeper. Empty when coordination is external or none. |
| `status.outputs.keeper_service_name` | `string` | name of the managed Keeper's client Service (operator naming contract: keeper-<keeper_name>). Empty when coordination is external or none. |
| `status.outputs.port_forward_command` | `string` | command to port-forward the HTTP interface to a developer laptop, e.g. kubectl port-forward svc/clickhouse-analytics -n data 8123:8123 |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.namespace` | KubernetesNamespace | `spec.name` |
| `spec.storageClass` | KubernetesStorageClass | `status.outputs.storage_class_name` |
| `spec.coordination.keeper.storageClass` | KubernetesStorageClass | `status.outputs.storage_class_name` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| KubernetesSignoz | `spec.clickhouse.host` | `status.outputs.service_name` |
| KubernetesSignoz | `spec.clickhouse.clusterName` | `status.outputs.cluster_name` |
| KubernetesSignoz | `spec.clickhouse.passwordSecret.secretName` | `status.outputs.auth_secret_name` |

## See Also

- [Overview](../README.md)
