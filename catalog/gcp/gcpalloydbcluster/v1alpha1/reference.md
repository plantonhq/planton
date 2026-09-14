# GcpAlloydbCluster

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

GcpAlloydbClusterSpec defines the configuration for an AlloyDB cluster
with a bundled primary instance.

AlloyDB is Google Cloud's fully managed, PostgreSQL-compatible database
designed for demanding enterprise workloads requiring high performance,
high availability, and strong consistency. It combines the familiarity
of PostgreSQL with Google's infrastructure for superior throughput and
lower latency compared to standard PostgreSQL.

This component bundles a cluster (the logical container with backup,
encryption, and networking configuration) with a primary instance (the
compute node that serves queries). A cluster without a primary instance
cannot serve any traffic, which is why they are bundled together.

Networking uses VPC peering via Private Service Access. The cluster's
network field references a VPC that must have Private Service Access
configured (compose GcpGlobalAddress with VPC_PEERING purpose +
GcpServiceNetworkingConnection on the target VPC).

Important behavioral notes:

  - The cluster_name, location, network, and kms_key_name fields are
    immutable after creation. Changing them requires recreating the cluster.

  - The primary_instance.instance_id field is immutable after creation.

  - AlloyDB supports both automated periodic backups and continuous
    backup (WAL streaming for point-in-time recovery). Both can be
    configured independently with separate CMEK keys.

  - Connectivity is exactly one of PSA (network_config.network) or PSC
    (psc_config.psc_enabled). The provider enforces this mutual exclusion.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpAlloydbCluster
metadata:
  name: test-alloydb
  org: test-org
  env: dev
spec:
  projectId:
    value: my-gcp-project
  clusterName: test-alloydb-cluster
  location: us-central1
  network:
    value: projects/my-gcp-project/global/networks/default
  databaseVersion: POSTGRES_15
  displayName: Test AlloyDB Cluster
  # The destroy guard defaults TRUE; the canonical example shows the
  # disarmed posture a test cluster wants (production keeps the default).
  deletionProtection: false
  deletionPolicy: DEFAULT
  labels:
    team: data-platform
  primaryInstance:
    instanceId: test-primary
    cpuCount: 2
    availabilityType: ZONAL
    deletionPolicy: DELETE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.clusterName` | `string` | yes |  |  |
| `spec.location` | `string` | yes |  |  |
| `spec.network` | `string \| valueFrom` |  |  | GcpVpcNetwork (`status.outputs.network_id`) |
| `spec.pscConfig` | `GcpAlloydbClusterPscConfig` |  |  |  |
| `spec.pscConfig.pscEnabled` | `bool` |  |  |  |
| `spec.clusterType` | `string` |  | `PRIMARY` |  |
| `spec.secondaryConfig` | `GcpAlloydbClusterSecondaryConfig` |  |  |  |
| `spec.secondaryConfig.primaryClusterName` | `string` | yes |  |  |
| `spec.allocatedIpRange` | `string` |  |  |  |
| `spec.databaseVersion` | `string` |  |  |  |
| `spec.displayName` | `string` |  |  |  |
| `spec.initialUser` | `GcpAlloydbClusterInitialUser` |  |  |  |
| `spec.initialUser.password` | `string` (sensitive) | yes |  |  |
| `spec.initialUser.user` | `string` |  |  |  |
| `spec.automatedBackupPolicy` | `GcpAlloydbClusterAutomatedBackupPolicy` |  |  |  |
| `spec.automatedBackupPolicy.enabled` | `bool` |  |  |  |
| `spec.automatedBackupPolicy.backupWindow` | `string` |  |  |  |
| `spec.automatedBackupPolicy.location` | `string` |  |  |  |
| `spec.automatedBackupPolicy.quantityBasedRetentionCount` | `int32` |  |  |  |
| `spec.automatedBackupPolicy.timeBasedRetentionPeriod` | `string` |  |  |  |
| `spec.automatedBackupPolicy.weeklySchedule` | `GcpAlloydbClusterBackupSchedule` |  |  |  |
| `spec.automatedBackupPolicy.weeklySchedule.daysOfWeek` | `[]string` |  |  |  |
| `spec.automatedBackupPolicy.weeklySchedule.startHour` | `int32` |  |  |  |
| `spec.automatedBackupPolicy.encryptionKmsKeyName` | `string \| valueFrom` |  |  | GcpKmsKey (`status.outputs.key_id`) |
| `spec.automatedBackupPolicy.labels` | `map<string, string>` |  |  |  |
| `spec.continuousBackupConfig` | `GcpAlloydbClusterContinuousBackupConfig` |  |  |  |
| `spec.continuousBackupConfig.enabled` | `bool` |  |  |  |
| `spec.continuousBackupConfig.recoveryWindowDays` | `int32` |  |  |  |
| `spec.continuousBackupConfig.encryptionKmsKeyName` | `string \| valueFrom` |  |  | GcpKmsKey (`status.outputs.key_id`) |
| `spec.kmsKeyName` | `string \| valueFrom` |  |  | GcpKmsKey (`status.outputs.key_id`) |
| `spec.maintenanceWindow` | `GcpAlloydbClusterMaintenanceWindow` |  |  |  |
| `spec.maintenanceWindow.day` | `string` | yes |  |  |
| `spec.maintenanceWindow.startHour` | `int32` |  |  |  |
| `spec.primaryInstance` | `GcpAlloydbClusterPrimaryInstance` | yes |  |  |
| `spec.primaryInstance.instanceId` | `string` | yes |  |  |
| `spec.primaryInstance.cpuCount` | `int32` |  |  |  |
| `spec.primaryInstance.machineType` | `string` |  |  |  |
| `spec.primaryInstance.availabilityType` | `string` |  |  |  |
| `spec.primaryInstance.databaseFlags` | `map<string, string>` |  |  |  |
| `spec.primaryInstance.displayName` | `string` |  |  |  |
| `spec.primaryInstance.queryInsightsConfig` | `GcpAlloydbClusterQueryInsightsConfig` |  |  |  |
| `spec.primaryInstance.queryInsightsConfig.queryPlansPerMinute` | `int32` |  |  |  |
| `spec.primaryInstance.queryInsightsConfig.queryStringLength` | `int32` |  |  |  |
| `spec.primaryInstance.queryInsightsConfig.recordApplicationTags` | `bool` |  |  |  |
| `spec.primaryInstance.queryInsightsConfig.recordClientAddress` | `bool` |  |  |  |
| `spec.primaryInstance.requireConnectors` | `bool` |  |  |  |
| `spec.primaryInstance.sslMode` | `string` |  |  |  |
| `spec.primaryInstance.activationPolicy` | `string` |  |  |  |
| `spec.primaryInstance.annotations` | `map<string, string>` |  |  |  |
| `spec.primaryInstance.gceZone` | `string` |  |  |  |
| `spec.primaryInstance.connectionPoolConfig` | `GcpAlloydbClusterConnectionPoolConfig` |  |  |  |
| `spec.primaryInstance.connectionPoolConfig.enabled` | `bool` |  |  |  |
| `spec.primaryInstance.connectionPoolConfig.flags` | `map<string, string>` |  |  |  |
| `spec.primaryInstance.enablePublicIp` | `bool` |  |  |  |
| `spec.primaryInstance.enableOutboundPublicIp` | `bool` |  |  |  |
| `spec.primaryInstance.authorizedExternalNetworks` | `[]GcpAlloydbClusterAuthorizedExternalNetwork` |  |  |  |
| `spec.primaryInstance.authorizedExternalNetworks[].cidrRange` | `string` | yes |  |  |
| `spec.primaryInstance.allocatedIpRangeOverride` | `string` |  |  |  |
| `spec.primaryInstance.pscInstanceConfig` | `GcpAlloydbClusterPscInstanceConfig` |  |  |  |
| `spec.primaryInstance.pscInstanceConfig.allowedConsumerProjects` | `[]string` |  |  |  |
| `spec.primaryInstance.pscInstanceConfig.pscAutoConnections` | `[]GcpAlloydbClusterPscAutoConnection` |  |  |  |
| `spec.primaryInstance.pscInstanceConfig.pscAutoConnections[].consumerNetwork` | `string` |  |  |  |
| `spec.primaryInstance.pscInstanceConfig.pscAutoConnections[].consumerProject` | `string` |  |  |  |
| `spec.primaryInstance.pscInstanceConfig.pscInterfaceConfigs` | `[]GcpAlloydbClusterPscInterfaceConfig` |  |  |  |
| `spec.primaryInstance.pscInstanceConfig.pscInterfaceConfigs[].networkAttachmentResource` | `string` |  |  |  |
| `spec.primaryInstance.deletionPolicy` | `string` |  |  |  |
| `spec.annotations` | `map<string, string>` |  |  |  |
| `spec.subscriptionType` | `string` |  |  |  |
| `spec.skipAwaitMajorVersionUpgrade` | `bool` |  |  |  |
| `spec.labels` | `map<string, string>` |  |  |  |
| `spec.dataplexConfig` | `GcpAlloydbClusterDataplexConfig` |  |  |  |
| `spec.dataplexConfig.enabled` | `bool` | yes |  |  |
| `spec.restoreBackupSource` | `GcpAlloydbClusterRestoreBackupSource` |  |  |  |
| `spec.restoreBackupSource.backupName` | `string` | yes |  |  |
| `spec.restoreContinuousBackupSource` | `GcpAlloydbClusterRestoreContinuousBackupSource` |  |  |  |
| `spec.restoreContinuousBackupSource.cluster` | `string \| valueFrom` | yes |  | GcpAlloydbCluster (`status.outputs.cluster_id`) |
| `spec.restoreContinuousBackupSource.pointInTime` | `string` | yes |  |  |
| `spec.restoreBackupdrBackupSource` | `GcpAlloydbClusterRestoreBackupdrBackupSource` |  |  |  |
| `spec.restoreBackupdrBackupSource.backup` | `string` | yes |  |  |
| `spec.restoreBackupdrPitrSource` | `GcpAlloydbClusterRestoreBackupdrPitrSource` |  |  |  |
| `spec.restoreBackupdrPitrSource.dataSource` | `string` | yes |  |  |
| `spec.restoreBackupdrPitrSource.pointInTime` | `string` | yes |  |  |
| `spec.deletionProtection` | `bool` |  | `true` |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

GCP project where the AlloyDB cluster will be created.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.clusterName

`string` · required

Name of the AlloyDB cluster. This becomes the GCP resource cluster_id.
Must start with a lowercase letter, can contain lowercase letters,
numbers, and hyphens, and must end with a lowercase letter or number.
Maximum 63 characters. Immutable after creation.

- rule: {"required":true,"string":{"minLen":"2","maxLen":"63","pattern":"^[a-z][a-z0-9-]{0,61}[a-z0-9]$"}}

### spec.location

`string` · required

GCP region where the cluster will be deployed (e.g., "us-central1").
Immutable after creation.

- rule: {"required":true}

### spec.network

`string | valueFrom`

VPC network for Private Service Access connectivity, as the relative
resource path "projects/{project}/global/networks/{network}" (the
AlloyDB API rejects full https:// self-link URLs). The VPC must have
Private Service Access configured (compose GcpGlobalAddress +
GcpServiceNetworkingConnection). Mutually exclusive with PSC.
Immutable after creation.

- references: GcpVpcNetwork (`status.outputs.network_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpVpcNetwork, name: <that resource's name>, fieldPath: status.outputs.network_id}} -- a bare string does not parse

### spec.pscConfig

`GcpAlloydbClusterPscConfig`

Private Service Connect configuration. When psc_enabled is true, network
must be unset — the cluster is reachable only through PSC endpoints.

### spec.pscConfig.pscEnabled

`bool`

When true, the cluster is reachable only through PSC endpoints.

### spec.clusterType

`string` · optional (explicit presence)

Cluster role. PRIMARY (default) is a writable primary; SECONDARY is a
cross-region DR replica that follows a primary cluster.

- default: `PRIMARY`
- rule: cluster_type must be PRIMARY or SECONDARY

### spec.secondaryConfig

`GcpAlloydbClusterSecondaryConfig`

Required when cluster_type is SECONDARY — names the primary cluster.

### spec.secondaryConfig.primaryClusterName

`string` · required

Full resource name of the primary cluster this secondary follows.

- rule: {"string":{"minLen":"1"}}

### spec.allocatedIpRange

`string`

Name of the allocated IP range for Private Service Access.
When set, the cluster uses this specific IP range for its private
connectivity instead of an auto-allocated range. This is common in
enterprise setups where IP ranges are pre-planned.

### spec.databaseVersion

`string`

PostgreSQL major version for the cluster.
Supported values: "POSTGRES_14", "POSTGRES_15", "POSTGRES_16".
If not specified, GCP selects the latest stable version.

- rule: database_version must be POSTGRES_14, POSTGRES_15, or POSTGRES_16

### spec.displayName

`string`

Human-readable display name for the cluster.

### spec.initialUser

`GcpAlloydbClusterInitialUser`

Initial database user created during cluster provisioning.
If not specified, no initial user is created. Access must then be
configured via AlloyDB Auth Proxy with IAM authentication.

### spec.initialUser.password

`string` · required · sensitive

Password for the initial user. Must be at least 8 characters.
This value is sensitive and should be handled accordingly.

- rule: {"required":true,"string":{"minLen":"8"}}

### spec.initialUser.user

`string`

Username for the initial user. If not specified, defaults to "postgres"
per GCP AlloyDB conventions.

### spec.automatedBackupPolicy

`GcpAlloydbClusterAutomatedBackupPolicy`

Automated backup policy for periodic snapshot backups.
When not specified, GCP uses its default policy (enabled, daily, 14-day retention).

- rule: only one of quantity_based_retention_count or time_based_retention_period may be set

### spec.automatedBackupPolicy.enabled

`bool`

Whether automated backups are enabled. Set to false to explicitly
disable automated backups.

### spec.automatedBackupPolicy.backupWindow

`string`

Length of the time window during which a backup can be taken.
Duration in seconds with 's' suffix, e.g., "3600s" (1 hour).
Default: "3600s".

- rule: backup_window must be a duration in seconds (e.g., '3600s')

### spec.automatedBackupPolicy.location

`string`

GCP region where backups will be stored. If not specified,
backups are stored in the same region as the cluster.

### spec.automatedBackupPolicy.quantityBasedRetentionCount

`int32`

Number of backups to retain. Mutually exclusive with
time_based_retention_period.

### spec.automatedBackupPolicy.timeBasedRetentionPeriod

`string`

How long to retain backups. Duration in seconds with 's' suffix,
e.g., "1209600s" (14 days). Mutually exclusive with
quantity_based_retention_count.

- rule: time_based_retention_period must be a duration in seconds (e.g., '1209600s')

### spec.automatedBackupPolicy.weeklySchedule

`GcpAlloydbClusterBackupSchedule`

Weekly schedule defining when backups are taken.

### spec.automatedBackupPolicy.weeklySchedule.daysOfWeek

`[]string`

Days of the week to take backups.
If not specified, GCP defaults to daily backups.

- rule: {"repeated":{"unique":true,"items":{"string":{"in":["MONDAY","TUESDAY","WEDNESDAY","THURSDAY","FRIDAY","SATURDAY","SUNDAY"]}}}}

### spec.automatedBackupPolicy.weeklySchedule.startHour

`int32`

Hour of day (0-23 UTC) to start backups. GCP's TimeOfDay structure
is simplified here since minutes/seconds/nanos are always zero for
AlloyDB backup schedules.

- rule: {"int32":{"lte":23,"gte":0}}

### spec.automatedBackupPolicy.encryptionKmsKeyName

`string | valueFrom`

Cloud KMS key for encrypting automated backups. If not specified,
backups use Google-managed encryption. When using a different key
from the cluster's encryption, this enables independent backup
encryption lifecycle management.

- references: GcpKmsKey (`status.outputs.key_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_id}} -- a bare string does not parse

### spec.automatedBackupPolicy.labels

`map<string, string>`

Labels applied to every backup created by this policy — how backup
storage costs are attributed and backup sets are filtered in the
console (distinct from the cluster's own labels).

### spec.continuousBackupConfig

`GcpAlloydbClusterContinuousBackupConfig`

Continuous backup configuration for point-in-time recovery (PITR).
Enabled by default with a 14-day recovery window.

### spec.continuousBackupConfig.enabled

`bool`

Whether continuous backup is enabled. Defaults to true.
Set to false only if you do not need point-in-time recovery.

### spec.continuousBackupConfig.recoveryWindowDays

`int32`

Number of days for which continuous backup data is retained,
enabling PITR within this window. Range: 1-35. Default: 14.

- rule: recovery_window_days must be between 1 and 35

### spec.continuousBackupConfig.encryptionKmsKeyName

`string | valueFrom`

Cloud KMS key for encrypting continuous backup data. If not specified,
continuous backups use Google-managed encryption.

- references: GcpKmsKey (`status.outputs.key_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_id}} -- a bare string does not parse

### spec.kmsKeyName

`string | valueFrom`

Cloud KMS key for encrypting the cluster's data at rest (CMEK).
Format: projects/{project}/locations/{location}/keyRings/{keyRing}/cryptoKeys/{key}
If not specified, data is encrypted with Google-managed keys.
Immutable after creation.

- references: GcpKmsKey (`status.outputs.key_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_id}} -- a bare string does not parse

### spec.maintenanceWindow

`GcpAlloydbClusterMaintenanceWindow`

Preferred maintenance window for system updates.

### spec.maintenanceWindow.day

`string` · required

Day of the week for maintenance.

- rule: {"required":true,"string":{"in":["MONDAY","TUESDAY","WEDNESDAY","THURSDAY","FRIDAY","SATURDAY","SUNDAY"]}}

### spec.maintenanceWindow.startHour

`int32`

Hour of day (0-23, UTC) when the maintenance window starts.

- rule: {"int32":{"lte":23,"gte":0}}

### spec.primaryInstance

`GcpAlloydbClusterPrimaryInstance` · required

Primary instance configuration. Required -- a cluster without a
primary instance cannot serve queries.

- rule: {"required":true}
- rule: only one of cpu_count or machine_type may be set
- rule: authorized_external_networks requires enable_public_ip
- rule: gce_zone can only be set on ZONAL instances — GCP rejects it when availability_type is REGIONAL (the default)

### spec.primaryInstance.instanceId

`string` · required

ID for the primary instance. This becomes the GCP resource name.
Must start with a lowercase letter, can contain lowercase letters,
numbers, and hyphens, and must end with a lowercase letter or number.
Maximum 63 characters. Immutable after creation.

- rule: {"required":true,"string":{"minLen":"2","maxLen":"63","pattern":"^[a-z][a-z0-9-]{0,61}[a-z0-9]$"}}

### spec.primaryInstance.cpuCount

`int32`

Number of CPUs for the instance. Valid values: 2, 4, 8, 16, 32, 64, 96, 128.
GCP selects the appropriate machine family automatically.
Mutually exclusive with machine_type.

### spec.primaryInstance.machineType

`string`

Explicit machine type (e.g., "n2-highmem-4", "c4a-highmem-4-lssd").
Use this for advanced scenarios where you need a specific machine family.
Mutually exclusive with cpu_count.

### spec.primaryInstance.availabilityType

`string`

Availability type controlling the placement of the instance.
ZONAL: single-zone deployment (lower cost, single zone of failure).
REGIONAL: multi-zone deployment with automatic failover (recommended
for production). Default: REGIONAL when unset.

- rule: availability_type must be ZONAL, REGIONAL, or AVAILABILITY_TYPE_UNSPECIFIED

### spec.primaryInstance.databaseFlags

`map<string, string>`

PostgreSQL database flags as key-value pairs.
These correspond to PostgreSQL server parameters (e.g.,
"max_connections", "work_mem", "shared_buffers").
See GCP AlloyDB documentation for supported flags.

### spec.primaryInstance.displayName

`string`

Human-readable display name for the primary instance.

### spec.primaryInstance.queryInsightsConfig

`GcpAlloydbClusterQueryInsightsConfig`

Query insights configuration for performance monitoring.
If not specified, GCP uses default query insights settings
(enabled with sensible defaults).

### spec.primaryInstance.queryInsightsConfig.queryPlansPerMinute

`int32`

Number of query execution plans captured per minute.
Range: 0-20. Default: 5. Set to 0 to disable plan capture.

- rule: {"int32":{"lte":20,"gte":0}}

### spec.primaryInstance.queryInsightsConfig.queryStringLength

`int32`

Maximum length of the query string stored in insights.
Range: 256-4500. 0 means unset — GCP applies its default (1024).
Longer strings help debug complex queries but use more storage.

- rule: query_string_length must be between 256 and 4500 (or 0 for the GCP default)

### spec.primaryInstance.queryInsightsConfig.recordApplicationTags

`bool`

Whether to record application tags set via
pg_stat_statements.track_activity_query_size.
Useful for tagging queries by application or feature.

### spec.primaryInstance.queryInsightsConfig.recordClientAddress

`bool`

Whether to record the client IP address for each query.
Useful for identifying which application instances generate load.

### spec.primaryInstance.requireConnectors

`bool`

Whether to require the AlloyDB Auth Proxy or AlloyDB Language Connectors
for all connections. When true, direct IP connections are rejected.
This enforces IAM-based authentication for all database access.

### spec.primaryInstance.sslMode

`string`

SSL mode for client connections.
ENCRYPTED_ONLY: all connections must use TLS (recommended for production).
ALLOW_UNENCRYPTED_AND_ENCRYPTED: both TLS and plaintext allowed.

- rule: ssl_mode must be ENCRYPTED_ONLY or ALLOW_UNENCRYPTED_AND_ENCRYPTED

### spec.primaryInstance.activationPolicy

`string`

Instance activation: ALWAYS keeps the primary running (the default
posture); NEVER stops it. Flipping ALWAYS→NEVER→ALWAYS is the
stop/start lever — a stopped primary keeps its configuration and
storage but serves nothing and stops billing for compute. Stop read
pool instances before stopping the primary.

- rule: activation_policy must be ALWAYS, NEVER, or ACTIVATION_POLICY_UNSPECIFIED

### spec.primaryInstance.annotations

`map<string, string>`

Unstructured metadata stored on the primary instance (annotations, not
labels — not used for billing filtering). Mutable in place.

### spec.primaryInstance.gceZone

`string`

Pin a ZONAL primary to a specific Compute Engine zone (e.g.
"us-central1-a"). Only valid when availability_type is ZONAL — GCP
rejects it on REGIONAL instances; leave empty to let GCP pick a zone
with available capacity. Mutable: changing it live-migrates the
primary to the new zone.

### spec.primaryInstance.connectionPoolConfig

`GcpAlloydbClusterConnectionPoolConfig`

AlloyDB managed connection pooling on the primary (built-in pooler).
Mutable in place.

### spec.primaryInstance.connectionPoolConfig.enabled

`bool`

Turn managed connection pooling on or off. Mutable in place.

### spec.primaryInstance.connectionPoolConfig.flags

`map<string, string>`

Pooler flags, keyed by flag name WITHOUT the "connection-pooling-"
prefix and with underscores instead of dashes (GCP's documented
convention for this provider surface): e.g. the flag
"connection-pooling-pool-mode" is set as key "pool_mode". Only
applied while enabled is true.

### spec.primaryInstance.enablePublicIp

`bool`

Enable a public IP on the primary instance. Pair with
authorized_external_networks to control who may reach it.

### spec.primaryInstance.enableOutboundPublicIp

`bool`

Enable outbound public IP for the primary instance.

### spec.primaryInstance.authorizedExternalNetworks

`[]GcpAlloydbClusterAuthorizedExternalNetwork`

CIDR ranges allowed to reach the primary's public IP. Requires
enable_public_ip.

### spec.primaryInstance.authorizedExternalNetworks[].cidrRange

`string` · required

- rule: {"string":{"minLen":"1"}}

### spec.primaryInstance.allocatedIpRangeOverride

`string`

Draw the primary's private IPs from a specific Private Service Access
allocated range (RFC 1035 name) instead of the range the cluster uses.
Immutable: changing it destroys and recreates the primary instance.

- rule: allocated_ip_range_override must be an RFC 1035 range name (1-63 chars: [a-z]([-a-z0-9]*[a-z0-9])?)

### spec.primaryInstance.pscInstanceConfig

`GcpAlloydbClusterPscInstanceConfig`

Private Service Connect configuration for the primary instance —
meaningful only on PSC clusters (psc_config.psc_enabled).

### spec.primaryInstance.pscInstanceConfig.allowedConsumerProjects

`[]string`

Consumer project numbers allowed to create PSC endpoints.

### spec.primaryInstance.pscInstanceConfig.pscAutoConnections

`[]GcpAlloydbClusterPscAutoConnection`

PSC service automation connections.

### spec.primaryInstance.pscInstanceConfig.pscAutoConnections[].consumerNetwork

`string`

Consumer network, e.g. "projects/vpc-host/global/networks/default".

### spec.primaryInstance.pscInstanceConfig.pscAutoConnections[].consumerProject

`string`

Consumer project ID (not project number).

### spec.primaryInstance.pscInstanceConfig.pscInterfaceConfigs

`[]GcpAlloydbClusterPscInterfaceConfig`

PSC interfaces for outbound connectivity (0 or 1 supported by AlloyDB).

### spec.primaryInstance.pscInstanceConfig.pscInterfaceConfigs[].networkAttachmentResource

`string`

Network attachment resource in the consumer project.

### spec.primaryInstance.deletionPolicy

`string`

What happens to the PRIMARY INSTANCE in GCP when this resource is
destroyed (the cluster has its own deletion_policy at the spec root).
  "DELETE"  -- (GCP's default when unset) the primary is deleted
  "PREVENT" -- destroy FAILS while the primary exists
  "ABANDON" -- the primary is removed from management but keeps
               running (and billing) in GCP

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

### spec.annotations

`map<string, string>`

Unstructured metadata stored on the cluster (annotations, not labels —
not used for billing filtering).

### spec.subscriptionType

`string`

Billing subscription tier: STANDARD (default) or TRIAL. A TRIAL cluster
converts to STANDARD when trial credits end.

- rule: subscription_type must be TRIAL or STANDARD

### spec.skipAwaitMajorVersionUpgrade

`bool`

When true, a database_version change returns immediately instead of
waiting for the in-place major-version upgrade to finish. The upgrade
continues server-side; use for very large clusters where the wait
exceeds sane IaC timeouts.

### spec.labels

`map<string, string>`

User-defined labels on the cluster and its bundled primary instance
(cost attribution, team ownership, environment tagging). Merged with
the platform's attribution labels; on key conflicts the platform
labels win. Mutable in place.

### spec.dataplexConfig

`GcpAlloydbClusterDataplexConfig`

Dataplex Universal Catalog integration (automatic metadata discovery).
GCP enables it by default when this block is absent; set
enabled: false to opt out explicitly.

### spec.dataplexConfig.enabled

`bool` · required · optional (explicit presence)

Whether Dataplex integration is enabled for the cluster. Mutable.
Required inside dataplex_config: declaring the block takes the
integration under management, and the switch says which way.

- rule: {"required":true}

### spec.restoreBackupSource

`GcpAlloydbClusterRestoreBackupSource`

Seed the new cluster from an AlloyDB backup. At most one restore
source may be set; all restore sources are create-time only (changing
one destroys and recreates the cluster).

### spec.restoreBackupSource.backupName

`string` · required

Full resource name of the source backup, e.g.
"projects/{project}/locations/{location}/backups/{backup}".

- rule: {"required":true}

### spec.restoreContinuousBackupSource

`GcpAlloydbClusterRestoreContinuousBackupSource`

Seed the new cluster by point-in-time recovery from a source cluster's
continuous backup stream. At most one restore source may be set;
create-time only.

### spec.restoreContinuousBackupSource.cluster

`string | valueFrom` · required

The source cluster to restore from — its full resource name
"projects/{project}/locations/{location}/clusters/{cluster}", or a
reference to a GcpAlloydbCluster resource. Restore provenance, not
physical placement: the new cluster is seeded FROM the source, never
contained IN it (containment_exempt).

- references: GcpAlloydbCluster (`status.outputs.cluster_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpAlloydbCluster, name: <that resource's name>, fieldPath: status.outputs.cluster_id}} -- a bare string does not parse

### spec.restoreContinuousBackupSource.pointInTime

`string` · required

The point in time to restore to, in RFC 3339 format (e.g.
"2026-08-01T12:00:00Z"). Must fall inside the source cluster's
continuous-backup recovery window.

- rule: {"required":true}

### spec.restoreBackupdrBackupSource

`GcpAlloydbClusterRestoreBackupdrBackupSource`

Seed the new cluster from a Backup and DR Service backup. At most one
restore source may be set; create-time only.

### spec.restoreBackupdrBackupSource.backup

`string` · required

Full resource name of the Backup and DR backup, in the format
"projects/{project}/locations/{location}/backupVaults/{vault}/dataSources/{dataSource}/backups/{backup}".

- rule: {"required":true}

### spec.restoreBackupdrPitrSource

`GcpAlloydbClusterRestoreBackupdrPitrSource`

Seed the new cluster by point-in-time recovery through the Backup and
DR Service. At most one restore source may be set; create-time only.

### spec.restoreBackupdrPitrSource.dataSource

`string` · required

Full resource name of the Backup and DR data source, in the format
"projects/{project}/locations/{location}/backupVaults/{vault}/dataSources/{dataSource}".

- rule: {"required":true}

### spec.restoreBackupdrPitrSource.pointInTime

`string` · required

The point in time to restore to, in RFC 3339 format.

- rule: {"required":true}

### spec.deletionProtection

`bool` · optional (explicit presence)

Client-side destroy guard: while true (GCP's default), any destroy —
including the platform's own teardown flows — FAILS until this field
is flipped to false and applied. Both engines always send the value
explicitly, so the spec is the single source of truth. Note the
ordering quirk in the provider: deletion_policy ABANDON is evaluated
BEFORE this guard, so abandoning a protected cluster still works.

- default: `true`

### spec.deletionPolicy

`string`

What happens to the CLUSTER in GCP when this resource is destroyed
(the bundled primary has its own deletion_policy under
primary_instance). AlloyDB clusters use a different value set from
most GCP resources:
  "DEFAULT" -- (GCP's default when unset) the cluster is deleted;
               the API rejects the delete while any instance other
               than the bundled primary still exists
  "FORCE"   -- the cluster AND every instance still in it are
               deleted — required when destroying a SECONDARY
               cluster that has a secondary instance
  "PREVENT" -- destroy FAILS; the strongest guard, evaluated even
               before deletion_protection
  "ABANDON" -- the cluster is removed from management but keeps
               running (and billing) in GCP; bypasses
               deletion_protection

- rule: deletion_policy must be one of: DEFAULT, FORCE, PREVENT, ABANDON

## Validation Rules

- `connectivity_psa_xor_psc`: set network for Private Service Access OR enable psc_config.psc_enabled — not both, not neither
- `secondary_requires_secondary_config`: cluster_type SECONDARY requires secondary_config.primary_cluster_name
- `at_most_one_restore_source`: at most one restore source may be set: restore_backup_source, restore_continuous_backup_source, restore_backupdr_backup_source, or restore_backupdr_pitr_source

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpAlloydbCluster, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.cluster_id` | `string` | Fully qualified cluster resource name. Format: projects/{project}/locations/{location}/clusters/{cluster} Used by downstream resources (read pool instances, backups) that reference this cluster. |
| `status.outputs.cluster_name` | `string` | Short name of the cluster (same as the spec's cluster_name input). Useful for display, logging, and human-readable references. |
| `status.outputs.primary_instance_ip` | `string` | Private IP address of the primary instance. This is the primary connection endpoint for applications. Applications connect to this IP on port 5432 (PostgreSQL default). |
| `status.outputs.primary_instance_name` | `string` | Fully qualified primary instance resource name. Format: projects/{project}/locations/{location}/clusters/{cluster}/instances/{instance} Used for AlloyDB Auth Proxy connections and monitoring. |
| `status.outputs.database_version` | `string` | Computed database engine version (e.g., "POSTGRES_15"). Reflects the actual version running, which may differ from the requested version if GCP selected a default. |
| `status.outputs.state` | `string` | Current state of the cluster (e.g., "READY", "CREATING", "MAINTENANCE"). |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.network` | GcpVpcNetwork | `status.outputs.network_id` |
| `spec.automatedBackupPolicy.encryptionKmsKeyName` | GcpKmsKey | `status.outputs.key_id` |
| `spec.continuousBackupConfig.encryptionKmsKeyName` | GcpKmsKey | `status.outputs.key_id` |
| `spec.kmsKeyName` | GcpKmsKey | `status.outputs.key_id` |
| `spec.restoreContinuousBackupSource.cluster` | GcpAlloydbCluster | `status.outputs.cluster_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| GcpAlloydbCluster | `spec.restoreContinuousBackupSource.cluster` | `status.outputs.cluster_id` |
| GcpAlloydbInstance | `spec.cluster` | `status.outputs.cluster_id` |
| GcpAlloydbUser | `spec.cluster` | `status.outputs.cluster_id` |

## See Also

- [Overview](../README.md)
