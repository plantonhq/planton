# GcpRedisCluster

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

GcpRedisClusterSpec describes a Memorystore for Redis Cluster
(`google_redis_cluster`): Google's fully managed, sharded Redis with
OSS Cluster protocol compatibility, horizontal scale across shards and
replicas, and Private Service Connect networking.

Choose this over GcpRedisInstance (the legacy single-instance Memorystore
for Redis, reached over VPC peering) when the keyspace or throughput
outgrows one node, when zero-downtime resizing matters, or when the
clients speak the Cluster protocol. Choose GcpMemorystoreInstance when the
engine should be Valkey rather than Redis.

Behavior an operator must know:

  - Immutable after creation: cluster_name, region, authorization_mode,
    transit_encryption_mode, zone_distribution_config, and the seed
    sources (gcs_source / managed_backup_source). Everything else --
    shard_count, replica_count, node_type, redis_configs, kms_key,
    persistence, backups, maintenance, the replication role -- updates in
    place; shard and replica changes resize the cluster without downtime.

  - Connectivity is Private Service Connect only. With psc_configs set,
    service connectivity automation places the endpoints, and it acts
    only when a GcpServiceConnectionPolicy for the gcp-memorystore-redis
    class exists on that network in this region -- deploy the policy
    first or creation fails with a connectivity error. With psc_configs
    empty, the cluster publishes service attachments (see the
    *_service_attachment outputs) and consumers build their own
    forwarding rules, registered through GcpRedisClusterEndpointSet.

  - Creating a cluster is a long-running operation (typically ten to
    fifteen minutes; more for many shards). Budget deploy windows for it.

  - Every node bills by the hour from creation, serving or idle; the
    smallest shape is one REDIS_SHARED_CORE_NANO shard with no replicas.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpRedisCluster
metadata:
  name: orders-cache
spec:
  projectId:
    value: my-gcp-project
  clusterName: orders-cache
  region: us-central1
  # Three shards with one replica each: six nodes, every shard survives a
  # node failure.
  shardCount: 3
  replicaCount: 1
  nodeType: REDIS_HIGHMEM_MEDIUM
  redisConfigs:
    maxmemory-policy: allkeys-lru
  # Google places the Private Service Connect endpoints in this VPC; a
  # GcpServiceConnectionPolicy for gcp-memorystore-redis must exist on it
  # in this region first.
  pscConfigs:
    - network:
        valueFrom:
          kind: GcpVpcNetwork
          name: prod-vpc
          fieldPath: status.outputs.network_id
  # Both modes are immutable -- decided at creation.
  authorizationMode: AUTH_MODE_IAM_AUTH
  transitEncryptionMode: TRANSIT_ENCRYPTION_MODE_SERVER_AUTHENTICATION
  serverCaMode: SERVER_CA_MODE_GOOGLE_MANAGED_PER_INSTANCE_CA
  # Every write reaches disk within a second.
  persistenceConfig:
    mode: AOF
    aofConfig:
      appendFsync: EVERYSEC
  zoneDistributionConfig:
    mode: MULTI_ZONE
  # Daily backups at 02:00 UTC kept 35 days; maintenance Sundays from 03:00.
  automatedBackupConfig:
    startHour: 2
    retention: "3024000s"
  maintenancePolicy:
    weeklyMaintenanceWindow:
      day: SUNDAY
      hour: 3
  labels:
    tier: production
  deletionProtectionEnabled: true
  deletionPolicy: PREVENT
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.clusterName` | `string` |  |  |  |
| `spec.region` | `string` | yes |  |  |
| `spec.shardCount` | `int32` | yes |  |  |
| `spec.replicaCount` | `int32` |  |  |  |
| `spec.nodeType` | `string` |  |  |  |
| `spec.redisConfigs` | `map<string, string>` |  |  |  |
| `spec.pscConfigs` | `[]GcpRedisClusterPscConfig` |  |  |  |
| `spec.pscConfigs[].network` | `string \| valueFrom` | yes |  | GcpVpcNetwork (`status.outputs.network_id`) |
| `spec.authorizationMode` | `string` |  |  |  |
| `spec.transitEncryptionMode` | `string` |  |  |  |
| `spec.serverCaMode` | `string` |  |  |  |
| `spec.serverCaPool` | `string` |  |  |  |
| `spec.kmsKey` | `string \| valueFrom` |  |  | GcpKmsKey (`status.outputs.key_id`) |
| `spec.persistenceConfig` | `GcpRedisClusterPersistenceConfig` |  |  |  |
| `spec.persistenceConfig.mode` | `string` |  |  |  |
| `spec.persistenceConfig.rdbConfig` | `GcpRedisClusterRdbConfig` |  |  |  |
| `spec.persistenceConfig.rdbConfig.rdbSnapshotPeriod` | `string` |  |  |  |
| `spec.persistenceConfig.rdbConfig.rdbSnapshotStartTime` | `string` |  |  |  |
| `spec.persistenceConfig.aofConfig` | `GcpRedisClusterAofConfig` |  |  |  |
| `spec.persistenceConfig.aofConfig.appendFsync` | `string` |  |  |  |
| `spec.zoneDistributionConfig` | `GcpRedisClusterZoneDistributionConfig` |  |  |  |
| `spec.zoneDistributionConfig.mode` | `string` |  |  |  |
| `spec.zoneDistributionConfig.zone` | `string` |  |  |  |
| `spec.maintenancePolicy` | `GcpRedisClusterMaintenancePolicy` |  |  |  |
| `spec.maintenancePolicy.weeklyMaintenanceWindow` | `GcpRedisClusterMaintenanceWindow` | yes |  |  |
| `spec.maintenancePolicy.weeklyMaintenanceWindow.day` | `string` | yes |  |  |
| `spec.maintenancePolicy.weeklyMaintenanceWindow.hour` | `int32` |  |  |  |
| `spec.automatedBackupConfig` | `GcpRedisClusterAutomatedBackupConfig` |  |  |  |
| `spec.automatedBackupConfig.startHour` | `int32` |  |  |  |
| `spec.automatedBackupConfig.retention` | `string` | yes |  |  |
| `spec.crossClusterReplicationConfig` | `GcpRedisClusterCrossClusterReplicationConfig` |  |  |  |
| `spec.crossClusterReplicationConfig.clusterRole` | `string` |  |  |  |
| `spec.crossClusterReplicationConfig.primaryCluster` | `GcpRedisClusterPrimaryCluster` |  |  |  |
| `spec.crossClusterReplicationConfig.primaryCluster.cluster` | `string \| valueFrom` | yes |  | GcpRedisCluster (`status.outputs.name`) |
| `spec.crossClusterReplicationConfig.secondaryClusters` | `[]GcpRedisClusterSecondaryCluster` |  |  |  |
| `spec.crossClusterReplicationConfig.secondaryClusters[].cluster` | `string \| valueFrom` | yes |  | GcpRedisCluster (`status.outputs.name`) |
| `spec.gcsSource` | `GcpRedisClusterGcsSource` |  |  |  |
| `spec.gcsSource.uris` | `[]string` | yes |  |  |
| `spec.managedBackupSource` | `GcpRedisClusterManagedBackupSource` |  |  |  |
| `spec.managedBackupSource.backup` | `string` | yes |  |  |
| `spec.labels` | `map<string, string>` |  |  |  |
| `spec.deletionProtectionEnabled` | `bool` |  | `true` |  |
| `spec.maintenanceVersion` | `string` |  |  |  |
| `spec.aclPolicy` | `string` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

The GCP project the cluster is created in: a literal project ID or a
GcpProject reference. If omitted, the provider's default project is
used.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.clusterName

`string`

The cluster's name in GCP. Defaults to metadata.name. 1-63
characters: lowercase letters, digits, and hyphens, starting with a
letter and ending alphanumeric. Immutable.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"63","pattern":"^[a-z]([-a-z0-9]*[a-z0-9])?$"}}

### spec.region

`string` · required

The region the cluster lives in (e.g. "us-central1"). Immutable.

- rule: {"required":true,"string":{"pattern":"^[a-z]+-[a-z]+[0-9]+$"}}

### spec.shardCount

`int32` · required

Number of shards. Each shard owns a slice of the keyspace and is at
least one node; Google caps shards per cluster at 250. Resizes in
place without downtime -- Google rebalances slots across the new
shard count.

- rule: {"required":true,"int32":{"lte":250,"gte":1}}

### spec.replicaCount

`int32`

Replica nodes per shard (0-5). Replicas serve reads and take over on
primary failure; 0 (Google's default) means a shard outage loses that
shard's data until it recovers. Resizes in place. Always sent, so the
manifest value is authoritative on both engines.

- rule: {"int32":{"lte":5,"gte":0}}

### spec.nodeType

`string`

Node shape for every node in the cluster -- the per-node hourly rate
and the memory each shard holds:
  REDIS_SHARED_CORE_NANO -- shared-core, ~1.4 GB (dev and test)
  REDIS_STANDARD_SMALL   -- ~6.5 GB, balanced
  REDIS_HIGHCPU_MEDIUM   -- ~13 GB, compute-leaning
  REDIS_HIGHMEM_MEDIUM   -- ~13 GB, memory-leaning
  REDIS_STANDARD_LARGE   -- ~13 GB, balanced
  REDIS_HIGHMEM_XLARGE   -- ~58 GB
  REDIS_HIGHMEM_2XLARGE  -- ~116 GB
If unset, Google picks REDIS_HIGHMEM_MEDIUM. Updates in place.

- rule: node_type must be one of: REDIS_SHARED_CORE_NANO, REDIS_STANDARD_SMALL, REDIS_HIGHCPU_MEDIUM, REDIS_HIGHMEM_MEDIUM, REDIS_STANDARD_LARGE, REDIS_HIGHMEM_XLARGE, REDIS_HIGHMEM_2XLARGE

### spec.redisConfigs

`map<string, string>`

Native Redis configuration parameters Google lets a cluster tune, as
key-value pairs (e.g. "maxmemory-policy": "allkeys-lru",
"notify-keyspace-events": "Ex"). Only the parameters in Google's
supported subset are accepted. Updates in place.

### spec.pscConfigs

`[]GcpRedisClusterPscConfig`

Consumer networks for Google-managed Private Service Connect
endpoints. Google supports one entry today. Leave empty to publish
service attachments only and register hand-built connections through
GcpRedisClusterEndpointSet. Updates in place.

- rule: {"repeated":{"maxItems":"1"}}

### spec.pscConfigs[].network

`string | valueFrom` · required

Consumer VPC network the automatic endpoints land in. The API takes
the relative resource path (projects/{project}/global/networks/{name});
a GcpVpcNetwork reference resolves to its network_id output, which is
already in that form.

- references: GcpVpcNetwork (`status.outputs.network_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpVpcNetwork, name: <that resource's name>, fieldPath: status.outputs.network_id}} -- a bare string does not parse

### spec.authorizationMode

`string`

How clients authenticate:
  AUTH_MODE_DISABLED -- no authentication (Google's default; the
                        network is the boundary)
  AUTH_MODE_IAM_AUTH -- clients present an IAM access token as the
                        Redis password; access is governed by IAM
Immutable. Sent explicitly on both engines so the choice never
depends on a provider default.

- rule: authorization_mode must be AUTH_MODE_DISABLED or AUTH_MODE_IAM_AUTH

### spec.transitEncryptionMode

`string`

TLS for client connections:
  TRANSIT_ENCRYPTION_MODE_DISABLED              -- plaintext (Google's
                                                   default)
  TRANSIT_ENCRYPTION_MODE_SERVER_AUTHENTICATION -- TLS; clients verify
                                                   the server's
                                                   certificate
Immutable. Sent explicitly on both engines.

- rule: transit_encryption_mode must be TRANSIT_ENCRYPTION_MODE_DISABLED or TRANSIT_ENCRYPTION_MODE_SERVER_AUTHENTICATION

### spec.serverCaMode

`string`

Which certificate authority signs the server certificate when TLS is
on:
  SERVER_CA_MODE_GOOGLE_MANAGED_PER_INSTANCE_CA -- a Google CA unique
                                                   to this cluster
                                                   (Google's default)
  SERVER_CA_MODE_GOOGLE_MANAGED_SHARED_CA       -- a Google CA shared
                                                   across clusters, so
                                                   a fleet trusts one
                                                   CA
  SERVER_CA_MODE_CUSTOMER_MANAGED_CAS_CA        -- your own CA pool in
                                                   Certificate
                                                   Authority Service
                                                   (server_ca_pool)
Meaningful only with TRANSIT_ENCRYPTION_MODE_SERVER_AUTHENTICATION.

- rule: server_ca_mode must be SERVER_CA_MODE_GOOGLE_MANAGED_PER_INSTANCE_CA, SERVER_CA_MODE_GOOGLE_MANAGED_SHARED_CA, or SERVER_CA_MODE_CUSTOMER_MANAGED_CAS_CA

### spec.serverCaPool

`string`

The Certificate Authority Service pool that signs the server
certificate under SERVER_CA_MODE_CUSTOMER_MANAGED_CAS_CA, as
projects/{project}/locations/{region}/caPools/{pool}.

- rule: server_ca_pool must be empty or of the form projects/{project}/locations/{region}/caPools/{pool}

### spec.kmsKey

`string | valueFrom`

Customer-managed encryption key (CMEK) for data at rest: a full
crypto key ID (projects/*/locations/*/keyRings/*/cryptoKeys/*) or a
GcpKmsKey reference. The key must be in the cluster's region and the
Memorystore service agent needs encrypter/decrypter on it. If unset,
Google-managed keys encrypt the data.

- references: GcpKmsKey (`status.outputs.key_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_id}} -- a bare string does not parse

### spec.persistenceConfig

`GcpRedisClusterPersistenceConfig`

Whether and how data survives a restart. Unset means in-memory only.

- rule: rdb_config is only set when mode is RDB
- rule: aof_config is only set when mode is AOF

### spec.persistenceConfig.mode

`string`

Persistence mode:
  DISABLED -- in-memory only (Google's default); a restart loses data
  RDB      -- periodic point-in-time snapshots (rdb_config)
  AOF      -- append-only write log (aof_config)

- rule: mode must be DISABLED, RDB, or AOF

### spec.persistenceConfig.rdbConfig

`GcpRedisClusterRdbConfig`

Snapshot schedule. Only meaningful when mode is RDB.

### spec.persistenceConfig.rdbConfig.rdbSnapshotPeriod

`string`

How often RDB snapshots are taken. If unset, Google defaults to
TWENTY_FOUR_HOURS.

- rule: rdb_snapshot_period must be ONE_HOUR, SIX_HOURS, TWELVE_HOURS, or TWENTY_FOUR_HOURS

### spec.persistenceConfig.rdbConfig.rdbSnapshotStartTime

`string`

RFC 3339 timestamp the snapshot schedule is anchored to (the first
snapshot happens at this time, then every rdb_snapshot_period). If
unset, Google anchors the schedule at creation.

### spec.persistenceConfig.aofConfig

`GcpRedisClusterAofConfig`

Flush policy. Only meaningful when mode is AOF.

### spec.persistenceConfig.aofConfig.appendFsync

`string`

How often the AOF buffer is flushed to disk (Redis appendfsync):
  NO       -- the OS decides (fastest; risks the last seconds of writes)
  EVERYSEC -- once per second (Google's default; the usual balance)
  ALWAYS   -- on every write (strongest durability, slowest)

- rule: append_fsync must be NO, EVERYSEC, or ALWAYS

### spec.zoneDistributionConfig

`GcpRedisClusterZoneDistributionConfig`

How nodes spread across the region's zones. Unset means MULTI_ZONE.
Immutable.

- rule: zone is required when mode is SINGLE_ZONE
- rule: zone is only set when mode is SINGLE_ZONE

### spec.zoneDistributionConfig.mode

`string`

MULTI_ZONE (Google's default) spreads primaries and replicas across
zones so a zonal outage keeps the cluster serving; SINGLE_ZONE puts
every node in `zone` for the lowest latency at the cost of zonal
failure.

- rule: mode must be MULTI_ZONE or SINGLE_ZONE

### spec.zoneDistributionConfig.zone

`string`

The zone every node lives in (e.g. "us-central1-a"). Required for
SINGLE_ZONE; must be empty for MULTI_ZONE.

### spec.maintenancePolicy

`GcpRedisClusterMaintenancePolicy`

The weekly window Google may apply maintenance in. Unset lets Google
choose.

### spec.maintenancePolicy.weeklyMaintenanceWindow

`GcpRedisClusterMaintenanceWindow` · required

The weekly window.

- rule: {"required":true}

### spec.maintenancePolicy.weeklyMaintenanceWindow.day

`string` · required

Day of the week (UTC).

- rule: {"required":true,"string":{"in":["MONDAY","TUESDAY","WEDNESDAY","THURSDAY","FRIDAY","SATURDAY","SUNDAY"]}}

### spec.maintenancePolicy.weeklyMaintenanceWindow.hour

`int32`

Hour of day (0-23, UTC) the one-hour window starts.

- rule: {"int32":{"lte":23,"gte":0}}

### spec.automatedBackupConfig

`GcpRedisClusterAutomatedBackupConfig`

Daily backups into the cluster's managed backup collection. Unset
means no automated backups (on-demand backups stay available through
the console and gcloud).

### spec.automatedBackupConfig.startHour

`int32`

Hour of day (0-23, UTC) the daily backup starts.

- rule: {"int32":{"lte":23,"gte":0}}

### spec.automatedBackupConfig.retention

`string` · required

How long backups are kept, as a seconds duration between one day
("86400s") and 365 days ("31536000s"), e.g. "3024000s" for 35 days.

- rule: {"required":true,"string":{"pattern":"^[0-9]+s$"}}

### spec.crossClusterReplicationConfig

`GcpRedisClusterCrossClusterReplicationConfig`

Cross-region disaster recovery: make this cluster a PRIMARY that
replicates to secondaries in other regions, or a SECONDARY that
replicates from a primary. Unset (or role NONE) is a standalone
cluster.

- rule: a SECONDARY cluster must name its primary through primary_cluster
- rule: primary_cluster is only set when cluster_role is SECONDARY
- rule: secondary_clusters is only set when cluster_role is PRIMARY

### spec.crossClusterReplicationConfig.clusterRole

`string`

This cluster's role:
  NONE      -- not replicating across regions (Google's default)
  PRIMARY   -- serves writes and replicates to the secondaries
  SECONDARY -- read-only replica of primary_cluster

- rule: cluster_role must be NONE, PRIMARY, or SECONDARY

### spec.crossClusterReplicationConfig.primaryCluster

`GcpRedisClusterPrimaryCluster`

The primary this cluster replicates from. Required for SECONDARY;
must be unset otherwise.

### spec.crossClusterReplicationConfig.primaryCluster.cluster

`string | valueFrom` · required

Full resource path of the primary
(projects/{project}/locations/{region}/clusters/{cluster}). A
reference resolves to another GcpRedisCluster's name output.

- references: GcpRedisCluster (`status.outputs.name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpRedisCluster, name: <that resource's name>, fieldPath: status.outputs.name}} -- a bare string does not parse

### spec.crossClusterReplicationConfig.secondaryClusters

`[]GcpRedisClusterSecondaryCluster`

The secondaries replicating from this cluster. Only for PRIMARY.

### spec.crossClusterReplicationConfig.secondaryClusters[].cluster

`string | valueFrom` · required

Full resource path of the secondary
(projects/{project}/locations/{region}/clusters/{cluster}). A
reference resolves to another GcpRedisCluster's name output.

- references: GcpRedisCluster (`status.outputs.name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpRedisCluster, name: <that resource's name>, fieldPath: status.outputs.name}} -- a bare string does not parse

### spec.gcsSource

`GcpRedisClusterGcsSource`

Seed the new cluster from RDB files in Cloud Storage at creation.
Mutually exclusive with managed_backup_source. Immutable.

### spec.gcsSource.uris

`[]string` · required

Cloud Storage URIs of the RDB files (gs://bucket/path.rdb). The
Memorystore service agent needs read access to the objects.

- rule: {"repeated":{"minItems":"1","items":{"cel":[{"id":"gcs_uri_format","message":"each URI must be a Cloud Storage path starting with gs://","expression":"this.startsWith('gs://')"}]}}}

### spec.managedBackupSource

`GcpRedisClusterManagedBackupSource`

Seed the new cluster from a managed backup at creation. Mutually
exclusive with gcs_source. Immutable.

### spec.managedBackupSource.backup

`string` · required

Full resource path of the backup
(projects/{project}/locations/{region}/backupCollections/{collection}/backups/{backup});
backup collections are listed in another cluster's backup_collection
output.

- rule: {"required":true}

### spec.labels

`map<string, string>`

User labels on the cluster, merged beneath Planton's platform
attribution labels (platform keys win on conflict).

### spec.deletionProtectionEnabled

`bool` · optional (explicit presence)

Whether deletion protection is on. Defaults to true (Google's own
posture): destroying the cluster fails until this is set to false.
Both engines send the value explicitly so a manifest that never
mentions it behaves the same everywhere. Updates in place.

- default: `true`

### spec.maintenanceVersion

`string`

Self-service maintenance: setting this to a newer version from the
cluster's available maintenance versions applies the update on your
schedule instead of waiting for Google's rollout. Update-only and
forward-only; leave unset to follow Google's rollout.

### spec.aclPolicy

`string`

A Memorystore for Redis Cluster ACL policy attached to the cluster:
Redis ACL rules (users, key patterns, allowed commands) authored once
and shared by clusters in the same region, as
projects/{project}/locations/{region}/aclPolicies/{policy}. Leave
empty for the built-in default ACL. Updates in place; the cluster's
is_acl_policy_in_sync status reports when the rules have reached
every node.

- rule: acl_policy must be empty or a full resource name of the form projects/{project}/locations/{region}/aclPolicies/{policy}

### spec.deletionPolicy

`string`

What happens to the cluster when this resource is destroyed
(evaluated after deletion_protection_enabled allows the destroy):
  "" / "DELETE" -- the cluster is deleted and its data lost
  "PREVENT"     -- destroy fails; a second guard for a cache whose
                   loss would stampede the backing store
  "ABANDON"     -- the cluster leaves management but keeps running
                   (and billing) in GCP

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Validation Rules

- `at_most_one_seed_source`: gcs_source and managed_backup_source are mutually exclusive -- choose one seed source
- `server_ca_pool_requires_customer_managed_cas`: server_ca_pool requires server_ca_mode SERVER_CA_MODE_CUSTOMER_MANAGED_CAS_CA
- `server_ca_mode_requires_tls`: server_ca_mode requires transit_encryption_mode TRANSIT_ENCRYPTION_MODE_SERVER_AUTHENTICATION

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpRedisCluster, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.name` | `string` | Full resource path of the cluster (projects/{project}/locations/{region}/clusters/{cluster}). The composition key: a SECONDARY's primary_cluster reference and a GcpRedisClusterEndpointSet's cluster reference resolve to it. |
| `status.outputs.uid` | `string` | Server-generated unique identifier, stable across updates. |
| `status.outputs.state` | `string` | The cluster's lifecycle state at the end of provisioning: CREATING, READY, UPDATING, DELETING, or SUSPENDED. |
| `status.outputs.discovery_endpoint_address` | `string` | IP address of the Google-placed discovery endpoint -- the address a Cluster-protocol client bootstraps from. Empty when psc_configs is unset (consumers build their own endpoints). |
| `status.outputs.discovery_endpoint_port` | `int32` | Port of the discovery endpoint (6379 unless configured otherwise). 0 when psc_configs is unset. |
| `status.outputs.discovery_service_attachment` | `string` | Service attachment a consumer's own forwarding rule targets to reach the DISCOVERY endpoint (projects/{project}/regions/{region}/serviceAttachments/{id}). The handle GcpRedisClusterEndpointSet registers a discovery connection against. |
| `status.outputs.primary_service_attachment` | `string` | Service attachment for the PRIMARY endpoint (writes). Empty when Google publishes no separate primary attachment for this cluster shape. |
| `status.outputs.reader_service_attachment` | `string` | Service attachment for the READER endpoint (replica reads). Empty when the cluster has no replicas. |
| `status.outputs.size_gb` | `int32` | Redis memory across the whole cluster in GB, as Google reports it. |
| `status.outputs.shard_count` | `int32` | The shard count in effect after provisioning. |
| `status.outputs.replica_count` | `int32` | The replica count per shard in effect after provisioning. |
| `status.outputs.backup_collection` | `string` | Full resource path of the managed backup collection Google keeps for this cluster once automated backups are configured -- the source of managed_backup_source paths for seeding new clusters. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.pscConfigs[].network` | GcpVpcNetwork | `status.outputs.network_id` |
| `spec.kmsKey` | GcpKmsKey | `status.outputs.key_id` |
| `spec.crossClusterReplicationConfig.primaryCluster.cluster` | GcpRedisCluster | `status.outputs.name` |
| `spec.crossClusterReplicationConfig.secondaryClusters[].cluster` | GcpRedisCluster | `status.outputs.name` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| GcpRedisCluster | `spec.crossClusterReplicationConfig.primaryCluster.cluster` | `status.outputs.name` |
| GcpRedisCluster | `spec.crossClusterReplicationConfig.secondaryClusters[].cluster` | `status.outputs.name` |
| GcpRedisClusterEndpointSet | `spec.cluster` | `status.outputs.name` |
| GcpRedisClusterEndpointSet | `spec.endpoints[].connections[].serviceAttachment` | `status.outputs.discovery_service_attachment` |

## See Also

- [Overview](../README.md)
