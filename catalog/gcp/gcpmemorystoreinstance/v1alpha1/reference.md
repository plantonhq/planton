# GcpMemorystoreInstance

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

GcpMemorystoreInstanceSpec defines the configuration for a Google Cloud
Memorystore instance.

Memorystore is a fully managed, in-memory data store that supports the
Valkey protocol (Redis-compatible). It provides sub-millisecond latency
for caching, session management, real-time analytics, leaderboards, and
pub/sub messaging.

Unlike the legacy Memorystore for Redis API (modeled as GcpRedisInstance),
this new-generation API offers:
  - Native sharding via configurable shard_count
  - Predefined node types (SHARED_CORE_NANO through HIGHMEM_2XLARGE)
  - Private Service Connect (PSC) networking instead of VPC peering
  - Both RDB and AOF persistence modes
  - Automated backups with configurable retention
  - Cross-region replication for disaster recovery
  - CLUSTER and CLUSTER_DISABLED (standalone) modes

Important behavioral notes:

  - The instance_name, location, mode, authorization_mode,
    transit_encryption_mode, kms_key, zone_distribution_config,
    psc_auto_connections, and the seed sources (gcs_source /
    managed_backup_source) are immutable after creation. Changing them
    requires replacing the instance.

  - PSC networking is the only connectivity option, and it is driven by
    service connectivity automation: a GcpServiceConnectionPolicy for
    the gcp-memorystore service class must exist on the network in this
    region BEFORE the instance is created, or creation fails with a
    connectivity error.

  - Node memory is determined by the node_type, not by an explicit
    memory_size_gb field. The actual memory per node is reported in
    stack outputs.

  - Instance creation is a long-running operation — the provider's
    default create timeout is 60 minutes, and multi-shard instances
    commonly run tens of minutes. Budget deploy windows accordingly; a
    create that appears stalled at the 20-minute mark is normal, not
    hung.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpMemorystoreInstance
metadata:
  name: my-test-memorystore
spec:
  # GCP project for the instance. Replace with your project ID.
  projectId:
    value: my-gcp-project-123

  instanceName: my-test-memorystore
  location: us-central1
  shardCount: 1

  # Standalone mode — compatible with any Valkey/Redis client.
  mode: CLUSTER_DISABLED
  nodeType: SHARED_CORE_NANO

  # PSC endpoint in the consumer VPC. A GcpServiceConnectionPolicy for
  # the gcp-memorystore class must exist on this network in this region
  # before the instance is created. Replace with your project/network.
  pscAutoConnections:
    - network:
        value: projects/my-gcp-project-123/global/networks/my-vpc
      projectId:
        value: my-gcp-project-123

  # Weekly maintenance window (UTC), starting on the hour — the API
  # supports no finer granularity. Stagger it against the systems this
  # cache fronts.
  maintenancePolicy:
    weeklyMaintenanceWindow:
      day: SUNDAY
      hour: 3

  # Allow hack-manifest teardown without a two-step disable; the policy
  # then deletes the instance.
  deletionProtectionEnabled: false
  deletionPolicy: DELETE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.instanceName` | `string` | yes |  |  |
| `spec.location` | `string` | yes |  |  |
| `spec.shardCount` | `int32` | yes |  |  |
| `spec.mode` | `string` |  |  |  |
| `spec.nodeType` | `string` |  |  |  |
| `spec.engineVersion` | `string` |  |  |  |
| `spec.engineConfigs` | `map<string, string>` |  |  |  |
| `spec.replicaCount` | `int32` |  |  |  |
| `spec.pscAutoConnections` | `[]GcpMemorystoreInstancePscAutoConnection` |  |  |  |
| `spec.pscAutoConnections[].network` | `string \| valueFrom` | yes |  | GcpVpcNetwork (`status.outputs.network_id`) |
| `spec.pscAutoConnections[].projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.authorizationMode` | `string` |  |  |  |
| `spec.transitEncryptionMode` | `string` |  |  |  |
| `spec.kmsKey` | `string \| valueFrom` |  |  | GcpKmsKey (`status.outputs.key_id`) |
| `spec.persistenceConfig` | `GcpMemorystoreInstancePersistenceConfig` |  |  |  |
| `spec.persistenceConfig.mode` | `string` | yes |  |  |
| `spec.persistenceConfig.rdbConfig` | `GcpMemorystoreInstanceRdbConfig` |  |  |  |
| `spec.persistenceConfig.rdbConfig.rdbSnapshotPeriod` | `string` | yes |  |  |
| `spec.persistenceConfig.rdbConfig.rdbSnapshotStartTime` | `string` |  |  |  |
| `spec.persistenceConfig.aofConfig` | `GcpMemorystoreInstanceAofConfig` |  |  |  |
| `spec.persistenceConfig.aofConfig.appendFsync` | `string` | yes |  |  |
| `spec.zoneDistributionConfig` | `GcpMemorystoreInstanceZoneDistributionConfig` |  |  |  |
| `spec.zoneDistributionConfig.mode` | `string` | yes |  |  |
| `spec.zoneDistributionConfig.zone` | `string` |  |  |  |
| `spec.maintenancePolicy` | `GcpMemorystoreInstanceMaintenancePolicy` |  |  |  |
| `spec.maintenancePolicy.weeklyMaintenanceWindow` | `GcpMemorystoreInstanceMaintenanceWindow` | yes |  |  |
| `spec.maintenancePolicy.weeklyMaintenanceWindow.day` | `string` | yes |  |  |
| `spec.maintenancePolicy.weeklyMaintenanceWindow.hour` | `int32` |  |  |  |
| `spec.automatedBackupConfig` | `GcpMemorystoreInstanceAutomatedBackupConfig` |  |  |  |
| `spec.automatedBackupConfig.startHour` | `int32` |  |  |  |
| `spec.automatedBackupConfig.retention` | `string` | yes |  |  |
| `spec.crossInstanceReplicationConfig` | `GcpMemorystoreInstanceCrossInstanceReplicationConfig` |  |  |  |
| `spec.crossInstanceReplicationConfig.instanceRole` | `string` | yes |  |  |
| `spec.crossInstanceReplicationConfig.primaryInstance` | `GcpMemorystoreInstancePrimaryInstance` |  |  |  |
| `spec.crossInstanceReplicationConfig.primaryInstance.instance` | `string \| valueFrom` | yes |  | GcpMemorystoreInstance (`status.outputs.name`) |
| `spec.crossInstanceReplicationConfig.secondaryInstances` | `[]GcpMemorystoreInstanceSecondaryInstance` |  |  |  |
| `spec.crossInstanceReplicationConfig.secondaryInstances[].instance` | `string \| valueFrom` | yes |  | GcpMemorystoreInstance (`status.outputs.name`) |
| `spec.gcsSource` | `GcpMemorystoreInstanceGcsSource` |  |  |  |
| `spec.gcsSource.uris` | `[]string` | yes |  |  |
| `spec.managedBackupSource` | `GcpMemorystoreInstanceManagedBackupSource` |  |  |  |
| `spec.managedBackupSource.backup` | `string` | yes |  |  |
| `spec.labels` | `map<string, string>` |  |  |  |
| `spec.deletionProtectionEnabled` | `bool` |  | `true` |  |
| `spec.serverCaMode` | `string` |  |  |  |
| `spec.serverCaPool` | `string` |  |  |  |
| `spec.maintenanceVersion` | `string` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |
| `spec.aclPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

GCP project where the Memorystore instance will be created.
Can be a literal project ID or a reference to a GcpProject resource.
If omitted, the provider's default project is used.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.instanceName

`string` · required

Name of the Memorystore instance. This becomes the GCP resource name.
Must start with a lowercase letter, contain only lowercase letters,
numbers, and hyphens, and end with a lowercase letter or number.
4-63 characters. Immutable after creation.

- rule: {"required":true,"string":{"minLen":"4","maxLen":"63","pattern":"^[a-z][a-z0-9-]{2,61}[a-z0-9]$"}}

### spec.location

`string` · required

GCP region where the instance will be deployed (e.g., "us-central1").
Immutable after creation.

- rule: {"required":true,"string":{"pattern":"^[a-z]+-[a-z]+[0-9]+$"}}

### spec.shardCount

`int32` · required

Number of shards for the instance. Each shard handles a portion of
the keyspace. Minimum 1 shard.

For CLUSTER mode: multiple shards distribute data across nodes.
For CLUSTER_DISABLED mode: typically 1 shard (single primary).

- rule: {"required":true,"int32":{"gte":1}}

### spec.mode

`string`

Instance mode controlling cluster topology.
CLUSTER: sharded mode with native cluster protocol support.
  Clients must use cluster-aware drivers.
CLUSTER_DISABLED: standalone mode with a single primary endpoint.
  Compatible with any Valkey/Redis client.
Immutable after creation.

- rule: mode must be CLUSTER or CLUSTER_DISABLED

### spec.nodeType

`string`

Predefined node type determining CPU and memory per node.
Shared-core and custom (burstable, dev/test tiers, smallest first):
  SHARED_CORE_NANO, CUSTOM_PICO, CUSTOM_MICRO, CUSTOM_MINI.
Dedicated-core (production tiers):
  STANDARD_SMALL, STANDARD_LARGE — balanced CPU:memory;
  HIGHCPU_MEDIUM — compute-leaning;
  HIGHMEM_MEDIUM, HIGHMEM_XLARGE, HIGHMEM_2XLARGE — memory-leaning,
  for large keyspaces.
If not specified, GCP selects a default.

- rule: node_type must be one of: SHARED_CORE_NANO, CUSTOM_PICO, CUSTOM_MICRO, CUSTOM_MINI, STANDARD_SMALL, STANDARD_LARGE, HIGHCPU_MEDIUM, HIGHMEM_MEDIUM, HIGHMEM_XLARGE, HIGHMEM_2XLARGE

### spec.engineVersion

`string`

Engine version (e.g., "VALKEY_8_0", "VALKEY_7_2").
If not specified, the latest supported version is used.

### spec.engineConfigs

`map<string, string>`

Engine configuration parameters as key-value pairs.
See Valkey/Redis configuration reference for supported parameters
(e.g., "maxmemory-policy", "notify-keyspace-events").

### spec.replicaCount

`int32`

Number of read replicas per shard (0-5). Default: 0 (no replicas).
Replicas provide read scaling and automatic failover.

### spec.pscAutoConnections

`[]GcpMemorystoreInstancePscAutoConnection`

Private Service Connect (PSC) endpoints for VPC connectivity.
Each entry creates a PSC endpoint in the specified consumer VPC,
allowing applications in that VPC to reach the instance.

A GcpServiceConnectionPolicy for the gcp-memorystore service class
must exist on each network in this region before the instance is
created — the connectivity automation refuses to place endpoints
without it.

At least one PSC connection is recommended for the instance to be
reachable. Multiple connections enable cross-project or multi-VPC access.
Immutable after creation.

### spec.pscAutoConnections[].network

`string | valueFrom` · required

Consumer VPC network where the PSC endpoint will be created.
The API requires the relative resource path
(projects/{project_id}/global/networks/{network_id}) — full https://
self-link URLs are rejected, so the reference resolves to the
GcpVpcNetwork's network_id output, which is already in that form.

- references: GcpVpcNetwork (`status.outputs.network_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpVpcNetwork, name: <that resource's name>, fieldPath: status.outputs.network_id}} -- a bare string does not parse

### spec.pscAutoConnections[].projectId

`string | valueFrom`

Consumer project ID where the PSC endpoint will be created.
Usually the same project as the Memorystore instance, but can differ
for cross-project connectivity. If omitted, both engines resolve the
provider's effective project — the endpoint lands next to the
instance, which is the common case.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.authorizationMode

`string`

Authentication mode for client connections.
AUTH_DISABLED: no authentication required (default).
IAM_AUTH: clients authenticate using GCP IAM credentials.
Immutable after creation.

- rule: authorization_mode must be AUTH_DISABLED or IAM_AUTH

### spec.transitEncryptionMode

`string`

TLS encryption mode for client-to-server traffic.
TRANSIT_ENCRYPTION_DISABLED: no encryption (default).
SERVER_AUTHENTICATION: clients verify the server's identity via TLS.
Immutable after creation.

- rule: transit_encryption_mode must be TRANSIT_ENCRYPTION_DISABLED or SERVER_AUTHENTICATION

### spec.kmsKey

`string | valueFrom`

Cloud KMS key for customer-managed encryption at rest (CMEK).
Format: projects/{project}/locations/{location}/keyRings/{keyRing}/cryptoKeys/{key}
If not specified, data is encrypted with Google-managed keys.
Immutable after creation.

- references: GcpKmsKey (`status.outputs.key_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_id}} -- a bare string does not parse

### spec.persistenceConfig

`GcpMemorystoreInstancePersistenceConfig`

Persistence configuration for data durability.
Controls whether and how data is written to disk.

- rule: rdb_config is required when persistence mode is RDB
- rule: aof_config is required when persistence mode is AOF

### spec.persistenceConfig.mode

`string` · required

Persistence mode.
DISABLED: no persistence, data is in-memory only.
RDB: periodic point-in-time snapshots.
AOF: append-only file logging every write.

- rule: {"required":true,"string":{"in":["DISABLED","RDB","AOF"]}}

### spec.persistenceConfig.rdbConfig

`GcpMemorystoreInstanceRdbConfig`

RDB snapshot configuration. Required when mode is RDB.

### spec.persistenceConfig.rdbConfig.rdbSnapshotPeriod

`string` · required

How often RDB snapshots are taken.

- rule: {"required":true,"string":{"in":["ONE_HOUR","SIX_HOURS","TWELVE_HOURS","TWENTY_FOUR_HOURS"]}}

### spec.persistenceConfig.rdbConfig.rdbSnapshotStartTime

`string`

Optional RFC3339 timestamp for when to start the first snapshot.
If not specified, GCP picks an appropriate time.

### spec.persistenceConfig.aofConfig

`GcpMemorystoreInstanceAofConfig`

AOF configuration. Required when mode is AOF.

### spec.persistenceConfig.aofConfig.appendFsync

`string` · required

How often the AOF buffer is flushed to disk.
NEVER: OS decides (best performance, risk of data loss on crash).
EVERY_SEC: flush once per second (good balance).
ALWAYS: flush on every write (strongest durability, lowest performance).

- rule: {"required":true,"string":{"in":["NEVER","EVERY_SEC","ALWAYS"]}}

### spec.zoneDistributionConfig

`GcpMemorystoreInstanceZoneDistributionConfig`

Zone distribution configuration.
Controls how nodes are spread across availability zones.
Immutable after creation.

- rule: zone is required when mode is SINGLE_ZONE

### spec.zoneDistributionConfig.mode

`string` · required

Zone distribution mode.
MULTI_ZONE: nodes spread across multiple zones for high availability (default).
SINGLE_ZONE: all nodes in a single zone for lowest latency.

- rule: {"required":true,"string":{"in":["MULTI_ZONE","SINGLE_ZONE"]}}

### spec.zoneDistributionConfig.zone

`string`

Zone for SINGLE_ZONE mode (e.g., "us-central1-a").
Required when mode is SINGLE_ZONE. Ignored for MULTI_ZONE.

### spec.maintenancePolicy

`GcpMemorystoreInstanceMaintenancePolicy`

Maintenance policy for scheduled maintenance windows.

### spec.maintenancePolicy.weeklyMaintenanceWindow

`GcpMemorystoreInstanceMaintenanceWindow` · required

Weekly maintenance window schedule.

- rule: {"required":true}

### spec.maintenancePolicy.weeklyMaintenanceWindow.day

`string` · required

Day of the week for the maintenance window.

- rule: {"required":true,"string":{"in":["MONDAY","TUESDAY","WEDNESDAY","THURSDAY","FRIDAY","SATURDAY","SUNDAY"]}}

### spec.maintenancePolicy.weeklyMaintenanceWindow.hour

`int32`

Hour of day (0-23, UTC) when the maintenance window starts. The window
always starts on the hour — the API supports no finer granularity.

- rule: {"int32":{"lte":23,"gte":0}}

### spec.automatedBackupConfig

`GcpMemorystoreInstanceAutomatedBackupConfig`

Automated backup configuration.
When configured, GCP takes daily backups at the specified hour
and retains them for the specified duration.

### spec.automatedBackupConfig.startHour

`int32`

Hour of day (0-23, UTC) when the daily backup starts.

- rule: {"int32":{"lte":23,"gte":0}}

### spec.automatedBackupConfig.retention

`string` · required

Backup retention duration in seconds.
Minimum: 86400s (1 day). Maximum: 31536000s (365 days).
Example: "3024000s" for 35 days.

- rule: {"required":true,"string":{"pattern":"^[0-9]+s$"}}

### spec.crossInstanceReplicationConfig

`GcpMemorystoreInstanceCrossInstanceReplicationConfig`

Cross-region replication for disaster recovery: make this instance a
PRIMARY replicating to secondaries in other regions, or a SECONDARY
continuously replicating from a primary. Omit (or role NONE) for a
standalone instance.

- rule: a SECONDARY instance must reference its primary via primary_instance
- rule: primary_instance is only set when instance_role is SECONDARY
- rule: secondary_instances is only set when instance_role is PRIMARY

### spec.crossInstanceReplicationConfig.instanceRole

`string` · required

This instance's role in the replication topology.
NONE: not participating in cross-instance replication.
PRIMARY: serves writes; replicates to the listed secondaries.
SECONDARY: read-only replica of primary_instance.

- rule: {"required":true,"string":{"in":["NONE","PRIMARY","SECONDARY"]}}

### spec.crossInstanceReplicationConfig.primaryInstance

`GcpMemorystoreInstancePrimaryInstance`

The primary this instance replicates from. Required when
instance_role is SECONDARY; must be unset otherwise.

### spec.crossInstanceReplicationConfig.primaryInstance.instance

`string | valueFrom` · required

Full resource path of the primary instance
(projects/{project}/locations/{location}/instances/{instance}).
A reference resolves to another GcpMemorystoreInstance's name output.

- references: GcpMemorystoreInstance (`status.outputs.name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpMemorystoreInstance, name: <that resource's name>, fieldPath: status.outputs.name}} -- a bare string does not parse

### spec.crossInstanceReplicationConfig.secondaryInstances

`[]GcpMemorystoreInstanceSecondaryInstance`

The secondaries replicating from this instance. Set when
instance_role is PRIMARY; must be empty otherwise.

### spec.crossInstanceReplicationConfig.secondaryInstances[].instance

`string | valueFrom` · required

Full resource path of the secondary instance
(projects/{project}/locations/{location}/instances/{instance}).
A reference resolves to another GcpMemorystoreInstance's name output.

- references: GcpMemorystoreInstance (`status.outputs.name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpMemorystoreInstance, name: <that resource's name>, fieldPath: status.outputs.name}} -- a bare string does not parse

### spec.gcsSource

`GcpMemorystoreInstanceGcsSource`

Seed the new instance's data from RDB files in Cloud Storage at
creation time. Mutually exclusive with managed_backup_source.
Immutable: seeding only happens at creation.

### spec.gcsSource.uris

`[]string` · required

Cloud Storage URIs of RDB files to import (gs://bucket/path.rdb).
The Memorystore service agent needs read access to the objects.

- rule: {"repeated":{"minItems":"1","items":{"cel":[{"id":"gcs_uri_format","message":"each URI must be a Cloud Storage path starting with gs://","expression":"this.startsWith('gs://')"}]}}}

### spec.managedBackupSource

`GcpMemorystoreInstanceManagedBackupSource`

Seed the new instance's data from an existing managed backup at
creation time. Mutually exclusive with gcs_source.
Immutable: seeding only happens at creation.

### spec.managedBackupSource.backup

`string` · required

Full resource path of the backup to restore from
(projects/{project}/locations/{location}/backupCollections/{collection}/backups/{backup}).

- rule: {"required":true}

### spec.labels

`map<string, string>`

User-defined labels to organize and track the instance. Merged
beneath Planton's platform attribution labels (platform keys win on
conflict).

### spec.deletionProtectionEnabled

`bool` · optional (explicit presence)

Whether deletion protection is enabled. When true (the default —
matching GCP's safety posture), destroying the instance fails until
this is explicitly set to false. Both IaC engines send the value
explicitly so destroy behavior is identical regardless of engine.

- default: `true`

### spec.serverCaMode

`string`

Server certificate authority mode for the TLS-enabled instance —
which CA signs the server certificate clients verify:
  ""                             -- GCP default (GOOGLE_MANAGED_PER_INSTANCE_CA)
  "GOOGLE_MANAGED_PER_INSTANCE_CA" -- a Google-managed CA unique to
                                      this instance
  "GOOGLE_MANAGED_SHARED_CA"       -- a Google-managed CA shared
                                      across instances (clients trust
                                      one CA for a whole fleet)
  "CUSTOMER_MANAGED_CAS_CA"        -- your own CA pool in Certificate
                                      Authority Service (pair with
                                      server_ca_pool)
Meaningful with transit_encryption_mode SERVER_AUTHENTICATION.
Immutable after creation.

- rule: server_ca_mode must be GOOGLE_MANAGED_PER_INSTANCE_CA, GOOGLE_MANAGED_SHARED_CA, or CUSTOMER_MANAGED_CAS_CA

### spec.serverCaPool

`string`

The Certificate Authority Service CA pool that signs the server
certificate when server_ca_mode is CUSTOMER_MANAGED_CAS_CA.
Format: projects/{project}/locations/{region}/caPools/{caPoolId}.
Immutable after creation.

### spec.maintenanceVersion

`string`

Self-service maintenance version. Setting this to a newer available
version triggers the maintenance update on your schedule instead of
waiting for GCP's rollout — the lever for applying a security patch
immediately. Only settable as an UPDATE to an existing instance, and
only forward (downgrades are rejected). Leave unset to follow GCP's
automatic rollout.

### spec.deletionPolicy

`string`

Deletion policy for the instance — what happens when this resource
is destroyed (evaluated only after deletion_protection_enabled allows
the destroy at all):
  ""        -- same as "DELETE" (provider default)
  "DELETE"  -- the instance is deleted; all in-memory data is lost
  "PREVENT" -- destroy FAILS; a second, independent guard for a
               cache whose loss would stampede the backing store
  "ABANDON" -- the instance is removed from management but left
               running (and billing) in GCP with its data intact

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

### spec.aclPolicy

`string`

The Memorystore ACL policy attached to the instance: a set of
Valkey ACL rules (users, key patterns, allowed commands) authored once
and shared across instances in the same region. Leave empty for the
instance's built-in default ACL (the "default" user with full access,
gated only by auth_enabled). Full resource name:
projects/{project}/locations/{region}/aclPolicies/{aclPolicyId}.
Mutable: attaching or swapping a policy is an in-place update; the
instance's is_acl_policy_in_sync status reports when the new rules
have propagated to every node.

- rule: acl_policy must be empty or a full resource name of the form projects/{project}/locations/{region}/aclPolicies/{aclPolicyId}

## Validation Rules

- `at_most_one_seed_source`: gcs_source and managed_backup_source are mutually exclusive — choose one seed source
- `server_ca_pool_requires_customer_managed_cas`: server_ca_pool requires server_ca_mode CUSTOMER_MANAGED_CAS_CA

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpMemorystoreInstance, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.discovery_address` | `string` | IP address of the instance's discovery endpoint. Clients connect to this address for cluster topology discovery and command routing. Works for both CLUSTER and CLUSTER_DISABLED modes. Extracted from the instance's PSC auto-created endpoint connections. |
| `status.outputs.discovery_port` | `int32` | Port of the instance's discovery endpoint (typically 6379). |
| `status.outputs.instance_uid` | `string` | Server-generated unique identifier for the instance. Stable across updates, useful for tracking and correlation. |
| `status.outputs.node_size_gb` | `double` | Memory size per node in GB, determined by the chosen node_type. Useful for capacity planning and monitoring. |
| `status.outputs.name` | `string` | Full resource path of the instance (projects/{project}/locations/{location}/instances/{instance}). The composition key for cross-instance replication: a SECONDARY's primary_instance reference resolves to this value. |
| `status.outputs.backup_collection` | `string` | Full resource path of the backup collection GCP maintains for this instance once automated backups are configured — where managed_backup_source paths for seeding new instances come from. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.pscAutoConnections[].network` | GcpVpcNetwork | `status.outputs.network_id` |
| `spec.pscAutoConnections[].projectId` | GcpProject | `status.outputs.project_id` |
| `spec.kmsKey` | GcpKmsKey | `status.outputs.key_id` |
| `spec.crossInstanceReplicationConfig.primaryInstance.instance` | GcpMemorystoreInstance | `status.outputs.name` |
| `spec.crossInstanceReplicationConfig.secondaryInstances[].instance` | GcpMemorystoreInstance | `status.outputs.name` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| GcpMemorystoreInstance | `spec.crossInstanceReplicationConfig.primaryInstance.instance` | `status.outputs.name` |
| GcpMemorystoreInstance | `spec.crossInstanceReplicationConfig.secondaryInstances[].instance` | `status.outputs.name` |

## See Also

- [Overview](../README.md)
