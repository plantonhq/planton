# GcpRedisInstance

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

GcpRedisInstanceSpec defines the configuration for a Google Cloud Memorystore
for Redis instance.

Memorystore for Redis provides a fully managed, in-memory data store backed
by the Redis protocol. It is used for caching, session management, real-time
analytics, rate limiting, leaderboards, and pub/sub messaging.

Instances come in two tiers:
  - BASIC: standalone single-node instance, no replication, no SLA
  - STANDARD_HA: primary node with automatic failover to a replica in a
    different zone, 99.9% availability SLA

Important behavioral notes:

  - The instance_name, tier, connect_mode, transit_encryption_mode,
    authorized_network, reserved_ip_range, location_id,
    alternative_location_id, and customer_managed_key fields are immutable
    after creation. Changing them requires replacing the instance.

  - When auth_enabled is true, GCP generates a random AUTH string that is
    rotated automatically. The current AUTH string is exported in outputs.

  - Read replicas are only available with STANDARD_HA tier and require
    read_replicas_mode to be set to READ_REPLICAS_ENABLED.

  - With connect_mode PRIVATE_SERVICE_ACCESS, the VPC must already carry a
    private services access connection (GcpServiceNetworkingConnection with
    a reserved GcpGlobalAddress range) before the instance is created.

## Example

```yaml
# Exercises the deep instance surface offline: HA tier with pinned zones,
# read replicas with the scale-out range, AUTH + TLS, tuned Redis configs,
# RDB persistence with an anchored snapshot schedule, a to-the-minute
# maintenance window with a description, user labels, and both destroy
# guards (protection on, policy DELETE).
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpRedisInstance
metadata:
  name: hack-redis
spec:
  # project_id omitted — falls back to the provider's default project.
  instanceName: hack-redis
  region: us-central1
  tier: STANDARD_HA
  memorySizeGb: 5
  redisVersion: REDIS_7_2
  displayName: Hack Redis
  locationId: us-central1-a
  alternativeLocationId: us-central1-b
  connectMode: DIRECT_PEERING
  reservedIpRange: 10.118.0.0/29
  secondaryIpRange: auto
  authEnabled: true
  transitEncryptionMode: SERVER_AUTHENTICATION
  redisConfigs:
    maxmemory-policy: allkeys-lru
    notify-keyspace-events: Ex
  maintenanceWindow:
    day: SUNDAY
    hour: 3
    minute: 30
    description: post-midnight window, after the nightly batch completes
  readReplicasMode: READ_REPLICAS_ENABLED
  replicaCount: 2
  persistenceConfig:
    persistenceMode: RDB
    rdbSnapshotPeriod: TWELVE_HOURS
    rdbSnapshotStartTime: "2026-01-01T03:00:00Z"
  labels:
    team: platform
    cost-center: engineering
  # Production posture: destroy is a deliberate two-step (flip protection
  # to false, apply, destroy); the policy then deletes the instance.
  deletionProtection: true
  deletionPolicy: DELETE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.instanceName` | `string` | yes |  |  |
| `spec.region` | `string` | yes |  |  |
| `spec.tier` | `string` | yes |  |  |
| `spec.memorySizeGb` | `int32` | yes |  |  |
| `spec.redisVersion` | `string` |  |  |  |
| `spec.displayName` | `string` |  |  |  |
| `spec.locationId` | `string` |  |  |  |
| `spec.alternativeLocationId` | `string` |  |  |  |
| `spec.authorizedNetwork` | `string \| valueFrom` |  |  | GcpVpcNetwork (`status.outputs.network_self_link`) |
| `spec.connectMode` | `string` |  |  |  |
| `spec.reservedIpRange` | `string` |  |  |  |
| `spec.secondaryIpRange` | `string` |  |  |  |
| `spec.authEnabled` | `bool` |  |  |  |
| `spec.transitEncryptionMode` | `string` |  |  |  |
| `spec.redisConfigs` | `map<string, string>` |  |  |  |
| `spec.maintenanceWindow` | `GcpRedisInstanceMaintenanceWindow` |  |  |  |
| `spec.maintenanceWindow.day` | `string` | yes |  |  |
| `spec.maintenanceWindow.hour` | `int32` |  |  |  |
| `spec.maintenanceWindow.minute` | `int32` |  |  |  |
| `spec.maintenanceWindow.description` | `string` |  |  |  |
| `spec.maintenanceVersion` | `string` |  |  |  |
| `spec.readReplicasMode` | `string` |  |  |  |
| `spec.replicaCount` | `int32` |  |  |  |
| `spec.persistenceConfig` | `GcpRedisInstancePersistenceConfig` |  |  |  |
| `spec.persistenceConfig.persistenceMode` | `string` | yes |  |  |
| `spec.persistenceConfig.rdbSnapshotPeriod` | `string` |  |  |  |
| `spec.persistenceConfig.rdbSnapshotStartTime` | `string` |  |  |  |
| `spec.customerManagedKey` | `string \| valueFrom` |  |  | GcpKmsKey (`status.outputs.key_id`), GcpKmsKeyHandle (`status.outputs.kms_key`) |
| `spec.labels` | `map<string, string>` |  |  |  |
| `spec.deletionProtection` | `bool` |  | `true` |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

GCP project where the Redis instance will be created.
If not specified, the provider's default project is used.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.instanceName

`string` · required

Name of the Redis instance. This becomes the GCP resource name.
Must start with a lowercase letter, contain only lowercase letters, numbers,
and hyphens, and end with a lowercase letter or number. Maximum 40 characters.
Immutable after creation.

- rule: {"required":true,"string":{"minLen":"2","maxLen":"40","pattern":"^[a-z][a-z0-9-]{0,38}[a-z0-9]$"}}

### spec.region

`string` · required

GCP region where the instance will be deployed (e.g., "us-central1").

- rule: {"required":true}

### spec.tier

`string` · required

Service tier controlling availability and replication.
BASIC: standalone instance, no replication, no SLA.
STANDARD_HA: primary + replica with automatic failover, 99.9% SLA.
Immutable after creation.

- rule: {"required":true,"string":{"in":["BASIC","STANDARD_HA"]}}

### spec.memorySizeGb

`int32` · required

Memory size in GiB for the Redis instance. This is the total memory
available for storing data. Minimum 1 GiB for BASIC; the GCP API requires
at least 5 GiB for STANDARD_HA and for enabling read replicas.

- rule: {"required":true,"int32":{"gte":1}}

### spec.redisVersion

`string`

Redis engine version (e.g., "REDIS_7_0", "REDIS_7_2", "REDIS_6_X").
If not specified, the latest supported version is used. Upgrades apply
in place; a version downgrade replaces the instance.

### spec.displayName

`string`

Human-readable display name for the instance.

### spec.locationId

`string`

Zone within the region where the instance will be placed.
For STANDARD_HA, this is the primary zone. GCP automatically selects
a different zone for the replica unless alternative_location_id pins it.
If not specified, GCP picks a zone. Immutable after creation.

### spec.alternativeLocationId

`string`

Zone for the STANDARD_HA replica. Only applicable to STANDARD_HA tier;
must differ from location_id. Pinning both zones matters when co-locating
the cache with zonal workloads (e.g. keeping the replica in the same zone
as a standby application stack to bound cross-zone latency after
failover). If not specified, GCP picks a different zone automatically.
Immutable after creation.

### spec.authorizedNetwork

`string | valueFrom`

VPC network to which the instance is connected.
If not specified, the default network is used.
Immutable after creation.

- references: GcpVpcNetwork (`status.outputs.network_self_link`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpVpcNetwork, name: <that resource's name>, fieldPath: status.outputs.network_self_link}} -- a bare string does not parse

### spec.connectMode

`string`

How the instance connects to the VPC network.
DIRECT_PEERING: VPC peering (default). Simpler setup.
PRIVATE_SERVICE_ACCESS: uses the network's private services access
connection. Required for Shared VPC and lets the instance consume an
address range you allocated (compose GcpGlobalAddress +
GcpServiceNetworkingConnection on the network first).
Immutable after creation.

- rule: connect_mode must be DIRECT_PEERING or PRIVATE_SERVICE_ACCESS

### spec.reservedIpRange

`string`

CIDR range of internal addresses reserved for this instance.
For DIRECT_PEERING: a /29 block (e.g., "10.0.0.0/29"), unique and
non-overlapping with existing subnets; if not specified, GCP selects an
unused /29 automatically. For PRIVATE_SERVICE_ACCESS: the NAME of an
allocated address range on the private services access connection
(a GcpGlobalAddress with purpose VPC_PEERING).
Immutable after creation.

### spec.secondaryIpRange

`string`

Additional IP range for node placement. Required when enabling read
replicas on an EXISTING instance (the original /29 has no room for the
extra nodes). For DIRECT_PEERING: a /28 CIDR or "auto". For
PRIVATE_SERVICE_ACCESS: the name of an allocated address range on the
private services access connection, or "auto". Mutable — this is the
field you set when scaling an in-place instance out to read replicas.

### spec.authEnabled

`bool`

Whether Redis AUTH is enabled. When true, clients must provide
the AUTH string (exported in outputs) to connect.
AUTH provides an additional layer of security beyond network controls.

### spec.transitEncryptionMode

`string`

TLS encryption mode for client-to-server traffic.
DISABLED: no encryption (default).
SERVER_AUTHENTICATION: clients verify the server's identity via TLS;
pair with the server_ca_certs output, which carries the CA
certificates clients must trust.
Immutable after creation.

- rule: transit_encryption_mode must be DISABLED or SERVER_AUTHENTICATION

### spec.redisConfigs

`map<string, string>`

Redis configuration parameters as key-value pairs.
See https://cloud.google.com/memorystore/docs/redis/reference/rest/v1/projects.locations.instances#Instance.FIELDS.redis_configs
for the list of supported parameters (e.g., "maxmemory-policy", "notify-keyspace-events").

### spec.maintenanceWindow

`GcpRedisInstanceMaintenanceWindow`

Weekly maintenance window. If not specified, GCP schedules maintenance
at its discretion.

### spec.maintenanceWindow.day

`string` · required

Day of the week for the maintenance window.

- rule: {"required":true,"string":{"in":["MONDAY","TUESDAY","WEDNESDAY","THURSDAY","FRIDAY","SATURDAY","SUNDAY"]}}

### spec.maintenanceWindow.hour

`int32`

Hour of day (0-23, UTC) when the maintenance window starts.

- rule: {"int32":{"lte":23,"gte":0}}

### spec.maintenanceWindow.minute

`int32`

Minute of the hour (0-59, UTC) when the maintenance window starts.
Combined with hour, this pins the window start to the exact minute —
useful for coordinating with maintenance windows of dependent systems
(e.g. start Redis maintenance 30 minutes after the database's window).

- rule: {"int32":{"lte":59,"gte":0}}

### spec.maintenanceWindow.description

`string`

Human-readable description of what this maintenance policy is for
(e.g. "post-midnight window, after the nightly batch completes").
Maximum 512 characters — the API rejects longer descriptions.

- rule: {"string":{"maxLen":"512"}}

### spec.maintenanceVersion

`string`

Self-service maintenance version. Setting this to a newer available
version triggers the maintenance update on your schedule instead of
waiting for GCP's rollout — the lever for applying a security patch
immediately. Leave unset to follow GCP's automatic rollout.

### spec.readReplicasMode

`string`

Read replica mode. Can only be set at creation time.
READ_REPLICAS_DISABLED (default): no read endpoint, no scaling.
READ_REPLICAS_ENABLED: read endpoint provided, instance can scale replicas.
Only available with STANDARD_HA tier.

- rule: read_replicas_mode must be READ_REPLICAS_DISABLED or READ_REPLICAS_ENABLED

### spec.replicaCount

`int32`

Number of read replicas. Valid range is 1-5 when read_replicas_mode is
READ_REPLICAS_ENABLED and tier is STANDARD_HA.

### spec.persistenceConfig

`GcpRedisInstancePersistenceConfig`

Persistence configuration for RDB snapshots.

- rule: rdb_snapshot_period is required when persistence_mode is RDB
- rule: rdb_snapshot_start_time is only meaningful when persistence_mode is RDB

### spec.persistenceConfig.persistenceMode

`string` · required

Persistence mode. DISABLED turns off persistence entirely.
RDB enables periodic RDB snapshots.

- rule: {"required":true,"string":{"in":["DISABLED","RDB"]}}

### spec.persistenceConfig.rdbSnapshotPeriod

`string`

How often RDB snapshots are taken. Required when persistence_mode is RDB.

- rule: rdb_snapshot_period must be ONE_HOUR, SIX_HOURS, TWELVE_HOURS, or TWENTY_FOUR_HOURS

### spec.persistenceConfig.rdbSnapshotStartTime

`string`

Date and time the first snapshot was/will be attempted, to which all
future snapshots align. RFC3339 UTC "Zulu" format (e.g.
"2014-10-02T15:01:23Z"). Anchoring the schedule lets you place snapshot
I/O in a low-traffic window instead of wherever instance creation time
happened to fall. If not provided, GCP uses the creation time.

- rule: rdb_snapshot_start_time must be an RFC3339 UTC timestamp like 2014-10-02T15:01:23Z

### spec.customerManagedKey

`string | valueFrom`

Cloud KMS key for customer-managed encryption at rest (CMEK).
Format: projects/{project}/locations/{location}/keyRings/{keyRing}/cryptoKeys/{key}
If not specified, data is encrypted with Google-managed keys.
Immutable after creation.

- references: GcpKmsKey (`status.outputs.key_id`), GcpKmsKeyHandle (`status.outputs.kms_key`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_id}} -- a bare string does not parse

### spec.labels

`map<string, string>`

User-defined labels to organize and track the instance, for cost
attribution and fleet queries. Merged beneath Planton's platform
attribution labels (platform keys win on conflict).

### spec.deletionProtection

`bool` · optional (explicit presence)

Whether deletion protection is enabled. When true (the default —
matching GCP's safety posture for stateful stores), destroying the
instance fails until this is explicitly set to false. Both IaC
engines send the value explicitly so destroy behavior is identical
regardless of engine.

- default: `true`

### spec.deletionPolicy

`string`

Deletion policy for the instance — what happens when this resource
is destroyed (evaluated only after deletion_protection allows the
destroy at all):
  ""        -- same as "DELETE" (provider default)
  "DELETE"  -- the instance is deleted; all in-memory data is lost
  "PREVENT" -- destroy FAILS; a second, independent guard for a
               cache whose loss would stampede the backing store
  "ABANDON" -- the instance is removed from management but left
               running (and billing) in GCP with its data intact

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Validation Rules

- `read_replicas_require_ha_tier`: read_replicas_mode READ_REPLICAS_ENABLED requires tier STANDARD_HA
- `replica_count_requires_ha_with_read_replicas`: replica_count (1-5) requires tier STANDARD_HA with read_replicas_mode READ_REPLICAS_ENABLED
- `alternative_location_requires_ha_tier`: alternative_location_id is only applicable to STANDARD_HA tier
- `alternative_location_must_differ`: alternative_location_id must be a different zone from location_id

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpRedisInstance, name: <resource-name>, fieldPath: status.outputs.<output>}`. A sensitive output is a secret the resource generates: on Planton it is kept in the organization's secret store and the output holds a `$secret/` reference, so feed it only to a sensitive field.

| Output | Type | Description |
|---|---|---|
| `status.outputs.host` | `string` | Hostname or IP address of the primary Redis endpoint. Clients connect to this address for read and write operations. |
| `status.outputs.port` | `int32` | Port number of the primary Redis endpoint (typically 6379). |
| `status.outputs.read_endpoint` | `string` | Hostname or IP address of the read replica endpoint. Only populated when tier is STANDARD_HA with read replicas enabled. Clients can direct read-only traffic here to reduce load on the primary. |
| `status.outputs.read_endpoint_port` | `int32` | Port number of the read replica endpoint. Only populated when tier is STANDARD_HA with read replicas enabled. |
| `status.outputs.current_location_id` | `string` | Zone where the Redis primary is currently running. For STANDARD_HA, this may change after a failover event. |
| `status.outputs.auth_string` | `string` (sensitive) | Redis AUTH string for client authentication. Only populated when auth_enabled is true. This value is generated and rotated by GCP automatically. Treat as a secret -- do not log or expose in UIs. |
| `status.outputs.server_ca_certs` | `[]string` | PEM-encoded CA certificates protecting the server endpoint. Only populated when transit_encryption_mode is SERVER_AUTHENTICATION. Clients must install these as trust anchors to complete the TLS handshake — this is the material every TLS-enabled consumer needs. |
| `status.outputs.persistence_iam_identity` | `string` | Cloud IAM identity (format "serviceAccount:<email>") used by import and export operations. Grant this identity access to Cloud Storage buckets used for RDB import/export. May change on CMEK rotation. |
| `status.outputs.effective_reserved_ip_range` | `string` | The CIDR range actually in use by the instance. Populated whether reserved_ip_range was set explicitly or auto-selected by GCP — the value to consult when planning non-overlapping address space. |
| `status.outputs.instance_name` | `string` | Name of the Redis instance in GCP — the identity handle API callers and automation use to address the instance. |
| `status.outputs.region` | `string` | Region hosting the instance (plain region name, e.g. "us-central1") — regional kinds export it so API callers can construct the instance's locations path without re-deriving it from a zone. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.authorizedNetwork` | GcpVpcNetwork | `status.outputs.network_self_link` |
| `spec.customerManagedKey` | GcpKmsKey | `status.outputs.key_id` |
| `spec.customerManagedKey` | GcpKmsKeyHandle | `status.outputs.kms_key` |

## See Also

- [Overview](../README.md)
