# GcpCloudSql

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

GcpCloudSqlSpec defines a Cloud SQL instance (`google_sql_database_instance`) —
a fully managed MySQL, PostgreSQL, or SQL Server database server.

One GcpCloudSql resource is one instance: a primary, or a read replica of
another instance (set master_instance_name). Databases and users inside an
instance are separate composable resources — GcpCloudSqlDatabase and
GcpCloudSqlUser — referencing this instance by its instance_name output.

Connectivity is explicit: public IPv4 (optionally restricted by authorized
networks), private IP inside a VPC (requires a service networking
connection — compose GcpGlobalAddress + GcpServiceNetworkingConnection on
the network first), or Private Service Connect. At least one path must be
enabled when the network block is present; omitting the block defaults to
public IPv4 reachable only through the Cloud SQL Auth Proxy / connectors
(no authorized networks).

## Example

```yaml
# Exercises the deep instance surface offline: engine/version/tier, disk,
# explicit connectivity with an authorized network, HA with backups + PITR
# (with an explicit retention unit), a final backup on delete, maintenance
# scheduling, Query Insights, flags, the ExecuteSql API posture, both
# delete guards, and the teardown policy.
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpCloudSql
metadata:
  name: hack-postgres
spec:
  # project_id omitted — falls back to the provider's default project.
  instanceName: hack-postgres
  region: us-central1
  databaseEngine: POSTGRESQL
  databaseVersion: POSTGRES_16
  tier: db-custom-2-7680
  edition: ENTERPRISE
  availabilityType: REGIONAL
  disk:
    type: PD_SSD
    sizeGb: 20
    autoResize: true
    autoResizeLimit: 100
  network:
    ipv4Enabled: true
    authorizedNetworks:
      - value: 203.0.113.0/24
        name: office
    sslMode: ENCRYPTED_ONLY
  backup:
    enabled: true
    startTime: "03:00"
    pointInTimeRecoveryEnabled: true
    transactionLogRetentionDays: 7
    retainedBackups: 14
    retentionUnit: COUNT
  finalBackup:
    enabled: true
    retentionDays: 7
    description: decommission snapshot
  dataApiAccess: DISALLOW_DATA_API
  maintenanceWindow:
    day: 7
    hour: 3
    updateTrack: stable
  insightsConfig:
    queryInsightsEnabled: true
    recordApplicationTags: true
  databaseFlags:
    max_connections: "200"
  deletionProtection: false
  deletionProtectionEnabled: false
  deletionPolicy: DELETE
  rootPassword: HackPassword123!
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.instanceName` | `string` | yes |  |  |
| `spec.region` | `string` | yes |  |  |
| `spec.databaseEngine` | `string` | yes |  |  |
| `spec.databaseVersion` | `string` | yes |  |  |
| `spec.tier` | `string` | yes |  |  |
| `spec.edition` | `string` |  | `ENTERPRISE` |  |
| `spec.availabilityType` | `string` |  | `ZONAL` |  |
| `spec.activationPolicy` | `string` |  | `ALWAYS` |  |
| `spec.disk` | `GcpCloudSqlDisk` |  |  |  |
| `spec.disk.type` | `string` |  | `PD_SSD` |  |
| `spec.disk.sizeGb` | `int32` |  | `10` |  |
| `spec.disk.autoResize` | `bool` |  | `true` |  |
| `spec.disk.autoResizeLimit` | `int32` |  |  |  |
| `spec.disk.provisionedIops` | `int64` |  |  |  |
| `spec.disk.provisionedThroughput` | `int64` |  |  |  |
| `spec.network` | `GcpCloudSqlNetwork` |  |  |  |
| `spec.network.privateNetwork` | `string \| valueFrom` |  |  | GcpVpcNetwork (`status.outputs.network_id`) |
| `spec.network.ipv4Enabled` | `bool` |  |  |  |
| `spec.network.authorizedNetworks` | `[]GcpCloudSqlAuthorizedNetwork` |  |  |  |
| `spec.network.authorizedNetworks[].value` | `string` | yes |  |  |
| `spec.network.authorizedNetworks[].name` | `string` |  |  |  |
| `spec.network.authorizedNetworks[].expirationTime` | `string` |  |  |  |
| `spec.network.allocatedIpRange` | `string` |  |  |  |
| `spec.network.enablePrivatePathForGoogleCloudServices` | `bool` |  |  |  |
| `spec.network.sslMode` | `string` |  |  |  |
| `spec.network.serverCaMode` | `string` |  |  |  |
| `spec.network.serverCaPool` | `string` |  |  |  |
| `spec.network.customSubjectAlternativeNames` | `[]string` |  |  |  |
| `spec.network.psc` | `GcpCloudSqlPscConfig` |  |  |  |
| `spec.network.psc.enabled` | `bool` |  |  |  |
| `spec.network.psc.allowedConsumerProjects` | `[]string` |  |  |  |
| `spec.network.psc.networkAttachmentUri` | `string` |  |  |  |
| `spec.network.psc.autoConnections` | `[]GcpCloudSqlPscAutoConnection` |  |  |  |
| `spec.network.psc.autoConnections[].consumerNetwork` | `string` | yes |  |  |
| `spec.network.psc.autoConnections[].consumerServiceProjectId` | `string` |  |  |  |
| `spec.network.psc.autoDnsEnabled` | `bool` |  |  |  |
| `spec.network.psc.writeEndpointDnsEnabled` | `bool` |  |  |  |
| `spec.network.psc.autoConnectionPolicyEnabled` | `bool` |  |  |  |
| `spec.network.serverCertificateRotationMode` | `string` |  |  |  |
| `spec.locationPreference` | `GcpCloudSqlLocationPreference` |  |  |  |
| `spec.locationPreference.zone` | `string` |  |  |  |
| `spec.locationPreference.secondaryZone` | `string` |  |  |  |
| `spec.backup` | `GcpCloudSqlBackup` |  |  |  |
| `spec.backup.enabled` | `bool` |  |  |  |
| `spec.backup.startTime` | `string` |  |  |  |
| `spec.backup.location` | `string` |  |  |  |
| `spec.backup.binaryLogEnabled` | `bool` |  |  |  |
| `spec.backup.pointInTimeRecoveryEnabled` | `bool` |  |  |  |
| `spec.backup.transactionLogRetentionDays` | `int32` |  |  |  |
| `spec.backup.retainedBackups` | `int32` |  |  |  |
| `spec.backup.retentionUnit` | `string` |  |  |  |
| `spec.maintenanceWindow` | `GcpCloudSqlMaintenanceWindow` |  |  |  |
| `spec.maintenanceWindow.day` | `int32` | yes |  |  |
| `spec.maintenanceWindow.hour` | `int32` |  |  |  |
| `spec.maintenanceWindow.updateTrack` | `string` |  |  |  |
| `spec.denyMaintenancePeriod` | `GcpCloudSqlDenyMaintenancePeriod` |  |  |  |
| `spec.denyMaintenancePeriod.startDate` | `string` | yes |  |  |
| `spec.denyMaintenancePeriod.endDate` | `string` | yes |  |  |
| `spec.denyMaintenancePeriod.time` | `string` | yes |  |  |
| `spec.insightsConfig` | `GcpCloudSqlInsightsConfig` |  |  |  |
| `spec.insightsConfig.queryInsightsEnabled` | `bool` |  |  |  |
| `spec.insightsConfig.queryStringLength` | `int32` |  |  |  |
| `spec.insightsConfig.recordApplicationTags` | `bool` |  |  |  |
| `spec.insightsConfig.recordClientAddress` | `bool` |  |  |  |
| `spec.insightsConfig.queryPlansPerMinute` | `int32` |  |  |  |
| `spec.insightsConfig.enhancedQueryInsightsEnabled` | `bool` |  |  |  |
| `spec.passwordValidationPolicy` | `GcpCloudSqlPasswordValidationPolicy` |  |  |  |
| `spec.passwordValidationPolicy.enablePasswordPolicy` | `bool` |  |  |  |
| `spec.passwordValidationPolicy.minLength` | `int32` |  |  |  |
| `spec.passwordValidationPolicy.complexity` | `string` |  |  |  |
| `spec.passwordValidationPolicy.reuseInterval` | `int32` |  |  |  |
| `spec.passwordValidationPolicy.disallowUsernameSubstring` | `bool` |  |  |  |
| `spec.passwordValidationPolicy.passwordChangeInterval` | `string` |  |  |  |
| `spec.dataCacheEnabled` | `bool` |  |  |  |
| `spec.connectionPooling` | `GcpCloudSqlConnectionPooling` |  |  |  |
| `spec.connectionPooling.enabled` | `bool` |  |  |  |
| `spec.connectionPooling.flags` | `map<string, string>` |  |  |  |
| `spec.databaseFlags` | `map<string, string>` |  |  |  |
| `spec.threadsPerCore` | `int32` |  |  |  |
| `spec.timeZone` | `string` |  |  |  |
| `spec.collation` | `string` |  |  |  |
| `spec.sqlServerAuditConfig` | `GcpCloudSqlSqlServerAuditConfig` |  |  |  |
| `spec.sqlServerAuditConfig.bucket` | `string` |  |  |  |
| `spec.sqlServerAuditConfig.retentionInterval` | `string` |  |  |  |
| `spec.sqlServerAuditConfig.uploadInterval` | `string` |  |  |  |
| `spec.activeDirectory` | `GcpCloudSqlActiveDirectoryConfig` |  |  |  |
| `spec.activeDirectory.domain` | `string` | yes |  |  |
| `spec.activeDirectory.mode` | `string` |  |  |  |
| `spec.activeDirectory.dnsServers` | `[]string` |  |  |  |
| `spec.activeDirectory.adminCredentialSecretName` | `string` |  |  |  |
| `spec.activeDirectory.organizationalUnit` | `string` |  |  |  |
| `spec.connectorEnforcement` | `string` |  |  |  |
| `spec.enableGoogleMlIntegration` | `bool` |  |  |  |
| `spec.enableDataplexIntegration` | `bool` |  |  |  |
| `spec.encryptionKeyName` | `string \| valueFrom` |  |  | GcpKmsKey (`status.outputs.key_id`) |
| `spec.deletionProtection` | `bool` |  |  |  |
| `spec.deletionProtectionEnabled` | `bool` |  |  |  |
| `spec.retainBackupsOnDelete` | `bool` |  |  |  |
| `spec.masterInstanceName` | `string \| valueFrom` |  |  | GcpCloudSql (`status.outputs.instance_name`) |
| `spec.replicaConfiguration` | `GcpCloudSqlReplicaConfiguration` |  |  |  |
| `spec.replicaConfiguration.failoverTarget` | `bool` |  |  |  |
| `spec.replicaConfiguration.cascadableReplica` | `bool` |  |  |  |
| `spec.replicaConfiguration.username` | `string` |  |  |  |
| `spec.replicaConfiguration.password` | `string` (sensitive) |  |  |  |
| `spec.replicaConfiguration.caCertificate` | `string` |  |  |  |
| `spec.replicaConfiguration.clientCertificate` | `string` |  |  |  |
| `spec.replicaConfiguration.clientKey` | `string` (sensitive) |  |  |  |
| `spec.replicaConfiguration.dumpFilePath` | `string` |  |  |  |
| `spec.replicaConfiguration.connectRetryInterval` | `int32` |  |  |  |
| `spec.replicaConfiguration.masterHeartbeatPeriod` | `int32` |  |  |  |
| `spec.replicaConfiguration.sslCipher` | `string` |  |  |  |
| `spec.replicaConfiguration.verifyServerCertificate` | `bool` |  |  |  |
| `spec.rootPassword` | `string` (sensitive) | yes |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |
| `spec.instanceType` | `string` |  |  |  |
| `spec.nodeCount` | `int32` |  |  |  |
| `spec.readPoolAutoScale` | `GcpCloudSqlReadPoolAutoScale` |  |  |  |
| `spec.readPoolAutoScale.enabled` | `bool` |  |  |  |
| `spec.readPoolAutoScale.minNodeCount` | `int32` |  |  |  |
| `spec.readPoolAutoScale.maxNodeCount` | `int32` |  |  |  |
| `spec.readPoolAutoScale.disableScaleIn` | `bool` |  |  |  |
| `spec.readPoolAutoScale.scaleInCooldownSeconds` | `int32` |  |  |  |
| `spec.readPoolAutoScale.scaleOutCooldownSeconds` | `int32` |  |  |  |
| `spec.readPoolAutoScale.targetMetrics` | `[]GcpCloudSqlReadPoolTargetMetric` |  |  |  |
| `spec.readPoolAutoScale.targetMetrics[].metric` | `string` | yes |  |  |
| `spec.readPoolAutoScale.targetMetrics[].targetValue` | `double` |  |  |  |
| `spec.clone` | `GcpCloudSqlClone` |  |  |  |
| `spec.clone.sourceInstanceName` | `string` | yes |  |  |
| `spec.clone.sourceProject` | `string` |  |  |  |
| `spec.clone.pointInTime` | `string` |  |  |  |
| `spec.clone.preferredZone` | `string` |  |  |  |
| `spec.clone.databaseNames` | `[]string` |  |  |  |
| `spec.clone.allocatedIpRange` | `string` |  |  |  |
| `spec.clone.sourceInstanceDeletionTime` | `string` |  |  |  |
| `spec.restoreBackupContext` | `GcpCloudSqlRestoreBackupContext` |  |  |  |
| `spec.restoreBackupContext.backupRunId` | `int64` |  |  |  |
| `spec.restoreBackupContext.instanceId` | `string` |  |  |  |
| `spec.restoreBackupContext.project` | `string` |  |  |  |
| `spec.pointInTimeRestoreContext` | `GcpCloudSqlPointInTimeRestoreContext` |  |  |  |
| `spec.pointInTimeRestoreContext.datasource` | `string` | yes |  |  |
| `spec.pointInTimeRestoreContext.pointInTime` | `string` | yes |  |  |
| `spec.pointInTimeRestoreContext.targetInstance` | `string` |  |  |  |
| `spec.pointInTimeRestoreContext.region` | `string` |  |  |  |
| `spec.pointInTimeRestoreContext.preferredZone` | `string` |  |  |  |
| `spec.pointInTimeRestoreContext.allocatedIpRange` | `string` |  |  |  |
| `spec.backupdrBackup` | `string` |  |  |  |
| `spec.maintenanceVersion` | `string` |  |  |  |
| `spec.replicaNames` | `[]string` |  |  |  |
| `spec.failoverDrReplicaName` | `string` |  |  |  |
| `spec.autoUpgradeEnabled` | `bool` |  |  |  |
| `spec.dataApiAccess` | `string` |  |  |  |
| `spec.finalBackup` | `GcpCloudSqlFinalBackup` |  |  |  |
| `spec.finalBackup.enabled` | `bool` |  |  |  |
| `spec.finalBackup.retentionDays` | `int32` |  |  |  |
| `spec.finalBackup.description` | `string` |  |  |  |
| `spec.entraId` | `GcpCloudSqlEntraIdConfig` |  |  |  |
| `spec.entraId.applicationId` | `string` | yes |  |  |
| `spec.entraId.tenantId` | `string` | yes |  |  |
| `spec.switchTransactionLogsToCloudStorageEnabled` | `bool` |  |  |  |
| `spec.includeReplicasForMajorVersionUpgrade` | `bool` |  |  |  |
| `spec.enforceNewSqlNetworkArchitecture` | `bool` |  |  |  |
| `spec.replicationLagMaxSeconds` | `int32` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

The GCP project in which to create this Cloud SQL instance.
Can be a literal project ID or a reference to a GcpProject resource.
If omitted, the provider's default project is used.
Example: "my-prod-project-123"

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.instanceName

`string` · required

Name of the Cloud SQL instance in GCP. Immutable.
After deletion, the name stays reserved for about one week and cannot be
reused — plan instance naming with that soft-delete window in mind.
Example: "orders-db-prod"

- rule: {"required":true,"string":{"maxLen":"98","pattern":"^[a-z]([-a-z0-9]*[a-z0-9])?$"}}

### spec.region

`string` · required

The GCP region hosting the instance (e.g. "us-central1"). Immutable:
an instance cannot move between regions in place.

- rule: {"required":true,"string":{"pattern":"^[a-z]+-[a-z]+[0-9]+$"}}

### spec.databaseEngine

`string` · required

The database engine family. Drives engine-specific validation: the
database_version prefix, PITR vs binary-log semantics, and the SQL
Server-only surface (time_zone, collation, audit, Active Directory).

- rule: database_engine must be MYSQL, POSTGRESQL, or SQLSERVER
- rule: {"required":true}

### spec.databaseVersion

`string` · required

The exact engine version, e.g. "MYSQL_8_0", "POSTGRES_16",
"SQLSERVER_2022_STANDARD". In-place major version upgrades are supported
by the API (no destroy/recreate), so this field is mutable — but always
take a backup before upgrading.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.tier

`string` · required

Machine type for the instance, e.g. "db-f1-micro" (shared-core,
dev/test), "db-custom-4-15360" (4 vCPU / 15 GB), or an Enterprise Plus
performance tier like "db-perf-optimized-N-4". Mutable: changing tier
resizes the instance in place (with a restart).

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.edition

`string` · optional (explicit presence)

Cloud SQL edition. ENTERPRISE is the standard tier (99.95% SLA on HA).
ENTERPRISE_PLUS adds a 99.99% HA SLA, the data cache, near-zero-downtime
maintenance, and 35-day transaction log retention. Mutable: edition
upgrades happen in place.

- default: `ENTERPRISE`
- rule: edition must be ENTERPRISE or ENTERPRISE_PLUS

### spec.availabilityType

`string` · optional (explicit presence)

ZONAL runs a single instance in one zone. REGIONAL enables high
availability: a standby in a second zone with automatic failover.
REGIONAL requires automated backups (and binary logs on MySQL).
Mutable: HA can be enabled or disabled in place.

- default: `ZONAL`
- rule: availability_type must be ZONAL or REGIONAL

### spec.activationPolicy

`string` · optional (explicit presence)

When the instance is activated. ALWAYS keeps it running (default);
NEVER stops the instance (storage is retained and billed — the
stop/start lever without destroying data); ON_DEMAND is legacy
first-generation behavior.

- default: `ALWAYS`
- rule: activation_policy must be ALWAYS, NEVER, or ON_DEMAND

### spec.disk

`GcpCloudSqlDisk`

Data disk configuration. If omitted: 10 GB PD_SSD with auto-resize.

- rule: HYPERDISK_BALANCED disks require size_gb of at least 20
- rule: provisioned_iops and provisioned_throughput apply to HYPERDISK_BALANCED disks only

### spec.disk.type

`string` · optional (explicit presence)

Disk type: PD_SSD (default, general purpose), PD_HDD (cheaper, slower —
dev/archive only), or HYPERDISK_BALANCED (min 20 GB). Immutable.

- default: `PD_SSD`
- rule: disk type must be PD_SSD, PD_HDD, or HYPERDISK_BALANCED

### spec.disk.sizeGb

`int32` · optional (explicit presence)

Disk size in GB (10–65536). Can grow in place; can NEVER shrink —
shrinking requires replacing the instance. With auto_resize enabled,
GCP grows the disk past this value as data accumulates.

- default: `10`
- rule: {"int32":{"lte":65536,"gte":10}}

### spec.disk.autoResize

`bool` · optional (explicit presence)

Automatically grow the disk as it approaches capacity. Enabled by
default — running a database out of disk is an outage.

- default: `true`

### spec.disk.autoResizeLimit

`int32` · optional (explicit presence)

Upper bound in GB for automatic growth (0 = no limit). The brake that
stops a runaway workload from growing a disk — and a bill — without
bound.

- rule: {"int32":{"gte":0}}

### spec.disk.provisionedIops

`int64` · optional (explicit presence)

HYPERDISK_BALANCED only: provisioned I/O operations per second for the
data disk — the IOPS dial decoupled from disk size.

- rule: {"int64":{"gte":"1"}}

### spec.disk.provisionedThroughput

`int64` · optional (explicit presence)

HYPERDISK_BALANCED only: provisioned throughput in MiB/s for the data
disk — the bandwidth dial decoupled from disk size.

- rule: {"int64":{"gte":"1"}}

### spec.network

`GcpCloudSqlNetwork`

Connectivity configuration: public IPv4, private VPC IP, and/or Private
Service Connect. If omitted, the instance gets a public IPv4 address
with NO authorized networks — reachable only through the Cloud SQL Auth
Proxy or connectors (IAM-authenticated), which is a safe default.

- rule: at least one connectivity path must be enabled: ipv4_enabled, private_network, or psc.enabled
- rule: authorized_networks apply to the public IP — set ipv4_enabled to true
- rule: allocated_ip_range applies to private IP — set private_network
- rule: enable_private_path_for_google_cloud_services applies to private IP — set private_network
- rule: server_ca_pool is required when server_ca_mode is CUSTOMER_MANAGED_CAS_CA
- rule: server_ca_pool applies only when server_ca_mode is CUSTOMER_MANAGED_CAS_CA
- rule: custom_subject_alternative_names apply only when server_ca_mode is CUSTOMER_MANAGED_CAS_CA
- rule: server_certificate_rotation_mode AUTOMATIC_ROTATION_DURING_MAINTENANCE requires server_ca_mode GOOGLE_MANAGED_CAS_CA or CUSTOMER_MANAGED_CAS_CA

### spec.network.privateNetwork

`string | valueFrom`

The VPC network for private IP connectivity, in
projects/{project}/global/networks/{network} form. Accepts a literal or
a reference to a GcpVpcNetwork resource. Setting this ENABLES private IP.
The network must already have a service networking connection (compose
GcpGlobalAddress with purpose VPC_PEERING + GcpServiceNetworkingConnection)
or instance creation fails. Can be set or changed in place, but never
removed — removing it forces instance replacement.

- references: GcpVpcNetwork (`status.outputs.network_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpVpcNetwork, name: <that resource's name>, fieldPath: status.outputs.network_id}} -- a bare string does not parse

### spec.network.ipv4Enabled

`bool`

Whether the instance gets a public IPv4 address. With no
authorized_networks, a public IP is reachable only through the Cloud
SQL Auth Proxy or connectors (IAM-authenticated) — a safe pattern.

### spec.network.authorizedNetworks

`[]GcpCloudSqlAuthorizedNetwork`

CIDR ranges allowed to connect DIRECTLY to the public IP. Prefer the
Auth Proxy over widening this list; never add 0.0.0.0/0 to a production
instance.

### spec.network.authorizedNetworks[].value

`string` · required

The CIDR range, e.g. "203.0.113.0/24".

- rule: {"required":true,"string":{"pattern":"^([0-9]{1,3}\\.){3}[0-9]{1,3}/[0-9]{1,2}$"}}

### spec.network.authorizedNetworks[].name

`string`

Display label for this entry in the console.

### spec.network.authorizedNetworks[].expirationTime

`string`

RFC 3339 timestamp after which this entry stops being honored —
built-in expiry for temporary access grants.

### spec.network.allocatedIpRange

`string`

Name of a specific allocated IP range (a GcpGlobalAddress with purpose
VPC_PEERING) from which the private IP is assigned. If empty, GCP picks
any range on the service networking connection.

### spec.network.enablePrivatePathForGoogleCloudServices

`bool`

Allows Google Cloud services (e.g. BigQuery federated queries) to reach
this instance over its private IP path.

### spec.network.sslMode

`string`

TLS posture for DIRECT connections. ALLOW_UNENCRYPTED_AND_ENCRYPTED
(default), ENCRYPTED_ONLY (reject plaintext), or
TRUSTED_CLIENT_CERTIFICATE_REQUIRED (mutual TLS with client certs).
Connector/Auth Proxy traffic is always encrypted regardless.

- rule: ssl_mode must be empty, ALLOW_UNENCRYPTED_AND_ENCRYPTED, ENCRYPTED_ONLY, or TRUSTED_CLIENT_CERTIFICATE_REQUIRED

### spec.network.serverCaMode

`string`

Which certificate authority signs the server certificate.
GOOGLE_MANAGED_INTERNAL_CA (default, per-instance CA),
GOOGLE_MANAGED_CAS_CA (Google-managed CA hierarchy in CA Service), or
CUSTOMER_MANAGED_CAS_CA (your own CA pool — set server_ca_pool).
Immutable after creation.

- rule: server_ca_mode must be empty, GOOGLE_MANAGED_INTERNAL_CA, GOOGLE_MANAGED_CAS_CA, or CUSTOMER_MANAGED_CAS_CA

### spec.network.serverCaPool

`string`

The CA Service CA pool (full resource path) that signs the server
certificate when server_ca_mode is CUSTOMER_MANAGED_CAS_CA.

### spec.network.customSubjectAlternativeNames

`[]string`

Additional DNS names embedded in the server certificate (customer-
managed CA only) — lets clients validate the cert against your own
hostnames instead of the instance IP.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","repeated":{"unique":true,"items":{"string":{"minLen":"1"}}}}

### spec.network.psc

`GcpCloudSqlPscConfig`

Private Service Connect: expose the instance as a PSC service
attachment that consumer VPCs connect to via PSC endpoints — private
connectivity WITHOUT VPC peering. The PSC alternative to
private_network.

- rule: PSC settings apply only when psc.enabled is true

### spec.network.psc.enabled

`bool`

Whether PSC connectivity is enabled for this instance. Immutable.

### spec.network.psc.allowedConsumerProjects

`[]string`

Consumer projects allowed to create PSC endpoints to this instance.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","repeated":{"unique":true,"items":{"string":{"minLen":"1"}}}}

### spec.network.psc.networkAttachmentUri

`string`

Network attachment (full resource path) for outbound connectivity from
the instance into a consumer VPC (used by outbound features such as
external replication over PSC).

### spec.network.psc.autoConnections

`[]GcpCloudSqlPscAutoConnection`

PSC endpoints GCP creates automatically in the listed consumer
networks — endpoint provisioning without consumer-side IaC.

### spec.network.psc.autoConnections[].consumerNetwork

`string` · required

The consumer VPC network (full resource path) in which GCP creates the
PSC endpoint.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.network.psc.autoConnections[].consumerServiceProjectId

`string`

The project owning the consumer network (defaults to the network's
project).

### spec.network.psc.autoDnsEnabled

`bool`

Automatically create DNS records for the PSC endpoints in the
consumer networks — clients resolve the instance by name instead of
tracking endpoint IPs.

### spec.network.psc.writeEndpointDnsEnabled

`bool`

Enterprise Plus only: also create a DNS record for the PSA write
endpoint, so clients follow the primary across switchovers by name.

### spec.network.psc.autoConnectionPolicyEnabled

`bool` · optional (explicit presence)

Whether Cloud SQL also creates a Service Connection Policy for the
auto_connections above, so the consumer networks need no separately
authored policy before the automatic endpoints can be provisioned.
Leave unset to keep the API's own default; sent only when set because
the API fills the value itself.

### spec.network.serverCertificateRotationMode

`string`

Controls automatic rotation of the server certificate.
NO_AUTOMATIC_ROTATION (default) or
AUTOMATIC_ROTATION_DURING_MAINTENANCE (requires a CA Service
server_ca_mode: GOOGLE_MANAGED_CAS_CA or CUSTOMER_MANAGED_CAS_CA).

- rule: server_certificate_rotation_mode must be empty, NO_AUTOMATIC_ROTATION, or AUTOMATIC_ROTATION_DURING_MAINTENANCE

### spec.locationPreference

`GcpCloudSqlLocationPreference`

Preferred zone placement for the primary (and the standby on REGIONAL
instances). If omitted, GCP picks zones automatically.

### spec.locationPreference.zone

`string`

Preferred zone for the primary, e.g. "us-central1-a". Must be in the
instance's region.

### spec.locationPreference.secondaryZone

`string`

Preferred zone for the standby of a REGIONAL instance. Must differ from
zone.

### spec.backup

`GcpCloudSqlBackup`

Automated backup configuration. Strongly recommended for anything that
holds real data; required for REGIONAL availability and for creating
read replicas.

- rule: backup settings (start_time, location, retention, binary logs, PITR) require backup.enabled to be true

### spec.backup.enabled

`bool`

Whether daily automated backups run. The foundation for PITR, read
replicas, and REGIONAL high availability.

### spec.backup.startTime

`string`

Start of the daily backup window in "HH:MM" (UTC). If empty, GCP
assigns a window.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^([0-1][0-9]|2[0-3]):[0-5][0-9]$"}}

### spec.backup.location

`string`

Multi-region or region where backups are stored, e.g. "us" or
"us-central1". If empty, GCP picks the closest multi-region.

### spec.backup.binaryLogEnabled

`bool`

MySQL only: write-ahead binary logging — MySQL's mechanism for
point-in-time recovery and the prerequisite for MySQL replicas and HA.

### spec.backup.pointInTimeRecoveryEnabled

`bool`

PostgreSQL / SQL Server only: point-in-time recovery via write-ahead /
transaction logs. Lets you restore to any second inside the log
retention window — the defense against bad migrations and accidental
deletes.

### spec.backup.transactionLogRetentionDays

`int32` · optional (explicit presence)

Days of transaction logs retained for PITR: 1–7 (ENTERPRISE) or 1–35
(ENTERPRISE_PLUS). Default 7.

- rule: {"int32":{"lte":35,"gte":1}}

### spec.backup.retainedBackups

`int32` · optional (explicit presence)

Number of daily backups retained (default 7). Older backups are pruned
automatically.

- rule: {"int32":{"gte":1}}

### spec.backup.retentionUnit

`string`

The unit retained_backups counts in. COUNT (the default and the only
unit the API defines) retains that many most-recent backups.

- rule: retention_unit must be empty or COUNT

### spec.maintenanceWindow

`GcpCloudSqlMaintenanceWindow`

One-hour weekly window in which GCP may restart the instance to apply
updates. Without it, maintenance can happen at any time.

### spec.maintenanceWindow.day

`int32` · required

Day of week: 1 (Monday) through 7 (Sunday).

- rule: {"required":true,"int32":{"lte":7,"gte":1}}

### spec.maintenanceWindow.hour

`int32` · optional (explicit presence)

Hour of day 0–23 (UTC) at which the window opens. 0 is midnight.

- rule: {"int32":{"lte":23,"gte":0}}

### spec.maintenanceWindow.updateTrack

`string`

Release cadence: "canary" (about one week after notification), "stable"
(about two weeks), or "week5" (about five weeks — maximum notice).

- rule: update_track must be empty, canary, stable, or week5

### spec.denyMaintenancePeriod

`GcpCloudSqlDenyMaintenancePeriod`

A date range during which maintenance is denied (e.g. an end-of-year
freeze). At most 90 days per period.

### spec.denyMaintenancePeriod.startDate

`string` · required

First denied date, "yyyy-mm-dd" (specific year) or "mm-dd" (recurs
annually).

- rule: {"required":true,"string":{"pattern":"^([0-9]{4}-)?[0-9]{2}-[0-9]{2}$"}}

### spec.denyMaintenancePeriod.endDate

`string` · required

Last denied date, same format as start_date.

- rule: {"required":true,"string":{"pattern":"^([0-9]{4}-)?[0-9]{2}-[0-9]{2}$"}}

### spec.denyMaintenancePeriod.time

`string` · required

Time of day (UTC) at which the deny period starts and ends, "HH:mm:SS",
e.g. "00:00:00".

- rule: {"required":true,"string":{"pattern":"^[0-2][0-9]:[0-5][0-9]:[0-5][0-9]$"}}

### spec.insightsConfig

`GcpCloudSqlInsightsConfig`

Query Insights: per-query performance telemetry in the console.
Negligible overhead for most workloads — enable it in production.

- rule: insights settings apply only when query_insights_enabled is true

### spec.insightsConfig.queryInsightsEnabled

`bool`

Whether Query Insights collects per-query performance telemetry.

### spec.insightsConfig.queryStringLength

`int32` · optional (explicit presence)

Maximum captured query text length in bytes (default 1024). Raise it
when long analytical queries get truncated in the console.

- rule: {"int32":{"lte":4500,"gte":256}}

### spec.insightsConfig.recordApplicationTags

`bool`

Record application tags (e.g. from sqlcommenter) with each query.

### spec.insightsConfig.recordClientAddress

`bool`

Record the client IP address with each query.

### spec.insightsConfig.queryPlansPerMinute

`int32` · optional (explicit presence)

Sampled execution plans captured per minute across all queries
(0 disables plan sampling; default 5).

- rule: {"int32":{"lte":20,"gte":0}}

### spec.insightsConfig.enhancedQueryInsightsEnabled

`bool`

Enhanced Query Insights: longer retention and finer-grained query
telemetry than the standard tier.

### spec.passwordValidationPolicy

`GcpCloudSqlPasswordValidationPolicy`

Password complexity/rotation policy enforced by the instance for
built-in database users.

### spec.passwordValidationPolicy.enablePasswordPolicy

`bool`

Master switch for the policy. The other fields take effect only while
this is true.

### spec.passwordValidationPolicy.minLength

`int32` · optional (explicit presence)

Minimum password length.

- rule: {"int32":{"gte":0}}

### spec.passwordValidationPolicy.complexity

`string`

COMPLEXITY_DEFAULT requires a mix of lower/upper case, numbers, and
non-alphanumeric characters.

- rule: complexity must be empty or COMPLEXITY_DEFAULT

### spec.passwordValidationPolicy.reuseInterval

`int32` · optional (explicit presence)

Number of previous passwords that cannot be reused.

- rule: {"int32":{"gte":0}}

### spec.passwordValidationPolicy.disallowUsernameSubstring

`bool`

Disallow the username as a substring of the password.

### spec.passwordValidationPolicy.passwordChangeInterval

`string`

PostgreSQL only: minimum interval between password changes, as a
duration string, e.g. "3600s".

- rule: password_change_interval must be a seconds duration string such as 3600s

### spec.dataCacheEnabled

`bool`

Enables the data cache (local SSD read caching). Enterprise Plus only;
delivers up to 4x read throughput for cache-friendly workloads.

### spec.connectionPooling

`GcpCloudSqlConnectionPooling`

Managed connection pooling (built-in pooler in front of the engine).
Reduces connection-storm pressure without deploying PgBouncer/ProxySQL.

### spec.connectionPooling.enabled

`bool`

Whether managed connection pooling is enabled.

### spec.connectionPooling.flags

`map<string, string>`

Pooler tuning flags (name → value), e.g. {"max_client_connections":
"1000"}. Permitted flags are engine-specific and validated by the API.

### spec.databaseFlags

`map<string, string>`

Engine configuration flags, e.g. {"max_connections": "500"} or
{"cloudsql.iam_authentication": "on"} (required on PostgreSQL before
creating IAM-type users). Flag names and permitted values are
engine-specific and validated by the API at deploy time.

### spec.threadsPerCore

`int32` · optional (explicit presence)

SQL Server only: number of threads per physical core (1 or 2).
Tuning lever for SQL Server licensing/performance trade-offs.

- rule: {"int32":{"lte":2,"gte":1}}

### spec.timeZone

`string`

SQL Server only: server time zone, e.g. "Pacific Standard Time".
Immutable in practice — changing it forces a maintenance restart.

### spec.collation

`string`

SQL Server only: server-level collation, e.g.
"SQL_Latin1_General_CP1_CI_AS". Immutable (set at create time).
MySQL/PostgreSQL collation is configured per database on
GcpCloudSqlDatabase instead.

### spec.sqlServerAuditConfig

`GcpCloudSqlSqlServerAuditConfig`

SQL Server only: SQLServer Audit — writes audit files to a GCS bucket.

### spec.sqlServerAuditConfig.bucket

`string`

Destination bucket, e.g. "gs://my-audit-bucket". The instance's service
account (service_account_email output) needs write access to it.

- rule: bucket must be a gs:// URI such as gs://my-audit-bucket

### spec.sqlServerAuditConfig.retentionInterval

`string`

How long generated audit files are kept, e.g. "86400s" (1 day).

- rule: retention_interval must be a seconds duration string such as 86400s

### spec.sqlServerAuditConfig.uploadInterval

`string`

How often audit files are uploaded to the bucket, e.g. "1800s".

- rule: upload_interval must be a seconds duration string such as 1800s

### spec.activeDirectory

`GcpCloudSqlActiveDirectoryConfig`

SQL Server only: the Active Directory the instance joins for Windows
authentication — a Managed Microsoft AD domain or (with mode
CUSTOMER_MANAGED_ACTIVE_DIRECTORY) a self-managed AD reached through
the listed domain controllers.

- rule: dns_servers, admin_credential_secret_name, and organizational_unit apply to CUSTOMER_MANAGED_ACTIVE_DIRECTORY mode only

### spec.activeDirectory.domain

`string` · required

The AD domain to join, e.g. "ad.example.com". With the default
(managed) mode this is a Managed Microsoft AD domain in the project.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.activeDirectory.mode

`string`

MANAGED_ACTIVE_DIRECTORY (default) joins a Managed Microsoft AD
domain; CUSTOMER_MANAGED_ACTIVE_DIRECTORY joins a self-managed AD
bootstrapped through dns_servers and the admin credential.

- rule: mode must be empty, MANAGED_ACTIVE_DIRECTORY, or CUSTOMER_MANAGED_ACTIVE_DIRECTORY

### spec.activeDirectory.dnsServers

`[]string`

Customer-managed AD only: domain controller IPv4 addresses used to
bootstrap the join.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","repeated":{"unique":true,"items":{"string":{"pattern":"^([0-9]{1,3}\\.){3}[0-9]{1,3}$"}}}}

### spec.activeDirectory.adminCredentialSecretName

`string`

Customer-managed AD only: the Secret Manager secret
(projects/{project}/secrets/{secret}) holding the AD administrator
credential used to join the domain. A secret NAME, not the credential
itself.

### spec.activeDirectory.organizationalUnit

`string`

Customer-managed AD only: the organizational unit distinguished name
(full hierarchical path) the instance's computer account joins under.

### spec.connectorEnforcement

`string`

Connection-path enforcement. REQUIRED rejects all direct connections,
admitting only Cloud SQL connectors / Auth Proxy traffic (which is
always TLS-encrypted and IAM-authenticated). NOT_REQUIRED (default)
admits direct connections too.

- rule: connector_enforcement must be empty, NOT_REQUIRED, or REQUIRED

### spec.enableGoogleMlIntegration

`bool`

Enables Vertex AI integration (e.g. ML predictions from SQL via
ml_integration). PostgreSQL and MySQL.

### spec.enableDataplexIntegration

`bool`

Enables Dataplex integration for data cataloging/lineage.

### spec.encryptionKeyName

`string | valueFrom`

Customer-managed encryption key (CMEK) for the instance's storage.
Accepts a full crypto key path
(projects/.../locations/.../keyRings/.../cryptoKeys/...) or a reference
to a GcpKmsKey resource. The key MUST be in the same region as the
instance. Immutable: CMEK cannot be added or changed after creation.

- references: GcpKmsKey (`status.outputs.key_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_id}} -- a bare string does not parse

### spec.deletionProtection

`bool`

Engine-side delete guard: when true, both IaC engines refuse to destroy
the instance (the plan/preview fails) until this is set back to false.
Protects against a bad manifest or an accidental destroy.

### spec.deletionProtectionEnabled

`bool`

API-side delete guard: when true, GCP itself rejects instance deletion
from EVERY surface — console, gcloud, API, and IaC. The strongest
protection; set both guards on production instances.

### spec.retainBackupsOnDelete

`bool`

When true, automated backups (and transaction logs for PITR) are
retained after the instance is deleted — the recovery path for
"deleted the instance, need the data back".

### spec.masterInstanceName

`string | valueFrom`

Makes this instance a READ REPLICA of the named primary instance.
Accepts the primary's instance name or a reference to a GcpCloudSql
resource. Immutable — an existing primary cannot be converted in place.
The primary must have automated backups enabled (and binary logs on
MySQL). To promote a replica into a standalone primary, set
instance_type to CLOUD_SQL_INSTANCE and clear this field and
replica_configuration in the same change (the instance restarts).

- references: GcpCloudSql (`status.outputs.instance_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpCloudSql, name: <that resource's name>, fieldPath: status.outputs.instance_name}} -- a bare string does not parse

### spec.replicaConfiguration

`GcpCloudSqlReplicaConfiguration`

Replica behavior (failover target, cascading) and — for replicas of an
external source — the replication channel credentials. Only meaningful
together with master_instance_name.

### spec.replicaConfiguration.failoverTarget

`bool`

Designates this replica as the failover target promoted if the primary
fails. MySQL only (legacy HA); not supported for PostgreSQL — prefer
availability_type REGIONAL on the primary for modern HA.

### spec.replicaConfiguration.cascadableReplica

`bool`

SQL Server only: allows this cross-region replica to carry replicas of
its own (cascading replication).

### spec.replicaConfiguration.username

`string`

Replication username on the external source.

### spec.replicaConfiguration.password

`string` · sensitive

Replication password on the external source. Write-only replication
channel material — never exported in outputs.

### spec.replicaConfiguration.caCertificate

`string`

PEM certificate of the external source's CA — public trust material,
not a secret.

### spec.replicaConfiguration.clientCertificate

`string`

PEM client certificate for mutual TLS with the external source —
public handshake material, not a secret.

### spec.replicaConfiguration.clientKey

`string` · sensitive

PEM private key matching client_certificate. Secret key material.

### spec.replicaConfiguration.dumpFilePath

`string`

Path to a SQL dump file in GCS (gs://...) used to seed the replica.

- rule: dump_file_path must be a gs:// URI such as gs://bucket/dump.sql.gz

### spec.replicaConfiguration.connectRetryInterval

`int32` · optional (explicit presence)

Seconds between connection retries to the source (MySQL only).

- rule: {"int32":{"gte":0}}

### spec.replicaConfiguration.masterHeartbeatPeriod

`int32` · optional (explicit presence)

Interval in milliseconds between replication heartbeats (MySQL only).

- rule: {"int32":{"gte":0}}

### spec.replicaConfiguration.sslCipher

`string`

Permitted ciphers for the replication channel TLS (MySQL only).

### spec.replicaConfiguration.verifyServerCertificate

`bool`

Whether to verify the external source's server certificate against
ca_certificate (MySQL only).

### spec.rootPassword

`string` · required · sensitive

Initial password for the engine's default admin user ("root" on MySQL,
"postgres" on PostgreSQL, "sqlserver" on SQL Server — where it is
REQUIRED). Write-only in GCP: never readable back from the API and
never exported in outputs. Create additional users as first-class
GcpCloudSqlUser resources.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"minLen":"8"}}

### spec.deletionPolicy

`string`

Engine-side teardown behavior. "DELETE" (default) destroys the
instance; "PREVENT" fails any plan that would destroy it; "ABANDON"
removes it from IaC management while leaving it running in GCP.
Distinct from deletion_protection (which blocks the destroy) — ABANDON
is the lever for handing an instance over to out-of-band management.

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

### spec.instanceType

`string`

The instance's role. Usually derived by GCP (primary vs replica);
set explicitly for two workflows: READ_POOL_INSTANCE turns the
resource into a read pool (with node_count / read_pool_auto_scale),
and CLOUD_SQL_INSTANCE promotes an existing read replica into a
standalone primary (clear master_instance_name and
replica_configuration in the same change).

- rule: instance_type must be empty, CLOUD_SQL_INSTANCE, READ_REPLICA_INSTANCE, READ_POOL_INSTANCE, or ON_PREMISES_INSTANCE

### spec.nodeCount

`int32` · optional (explicit presence)

Read pools only: number of nodes serving reads behind the pool's
single endpoint. Read-only while read_pool_auto_scale is enabled (the
autoscaler owns it).

- rule: {"int32":{"gte":1}}

### spec.readPoolAutoScale

`GcpCloudSqlReadPoolAutoScale`

Read pools only: automatic node-count scaling between min and max
bounds, driven by target metrics.

- rule: max_node_count must be greater than or equal to min_node_count

### spec.readPoolAutoScale.enabled

`bool`

Whether auto scaling is active. While enabled, node_count is owned by
the autoscaler and read-only.

### spec.readPoolAutoScale.minNodeCount

`int32` · optional (explicit presence)

Lower bound of pool nodes; scale-in never goes below it.

- rule: {"int32":{"gte":1}}

### spec.readPoolAutoScale.maxNodeCount

`int32` · optional (explicit presence)

Upper bound of pool nodes; scale-out never exceeds it.

- rule: {"int32":{"gte":1}}

### spec.readPoolAutoScale.disableScaleIn

`bool`

Disables scale-in entirely: the pool only ever grows automatically.

### spec.readPoolAutoScale.scaleInCooldownSeconds

`int32` · optional (explicit presence)

Cooldown in seconds after a scale-in before another may run.

- rule: {"int32":{"gte":0}}

### spec.readPoolAutoScale.scaleOutCooldownSeconds

`int32` · optional (explicit presence)

Cooldown in seconds after a scale-out before another may run.

- rule: {"int32":{"gte":0}}

### spec.readPoolAutoScale.targetMetrics

`[]GcpCloudSqlReadPoolTargetMetric`

Metrics the autoscaler steers by; nodes are added or removed to hold
each metric at its target value.

### spec.readPoolAutoScale.targetMetrics[].metric

`string` · required

Metric name, e.g. a CPU utilization metric.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.readPoolAutoScale.targetMetrics[].targetValue

`double`

Target value for the metric.

### spec.clone

`GcpCloudSqlClone`

Creates this instance as a CLONE of another instance — a full copy of
its data at a point in time. A create-time source: changing it on an
existing instance triggers a new clone operation.

### spec.clone.sourceInstanceName

`string` · required

Name of the source instance to clone.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.clone.sourceProject

`string`

Project of the source instance, for cross-project clones. Defaults to
this instance's project.

### spec.clone.pointInTime

`string`

RFC 3339 timestamp to clone from (point-in-time clone). Requires PITR
(or binary logs on MySQL) on the source.

### spec.clone.preferredZone

`string`

PostgreSQL point-in-time clones only: the zone the clone lands in.
Defaults to the source's zone.

### spec.clone.databaseNames

`[]string`

SQL Server point-in-time clones only: clone just the named databases.
Empty clones all databases.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","repeated":{"unique":true,"items":{"string":{"minLen":"1"}}}}

### spec.clone.allocatedIpRange

`string`

Private-IP clones: the allocated IP range name (RFC 1035) the clone's
private IP is drawn from.

### spec.clone.sourceInstanceDeletionTime

`string`

Cloning a DELETED instance: the RFC 3339 timestamp of the source's
deletion.

### spec.restoreBackupContext

`GcpCloudSqlRestoreBackupContext`

Restores a specific backup run into this instance. An imperative
restore trigger expressed declaratively: adding or changing the block
runs the restore after the instance exists.

### spec.restoreBackupContext.backupRunId

`int64`

The backup run ID to restore.

- rule: {"int64":{"gte":"1"}}

### spec.restoreBackupContext.instanceId

`string`

The instance the backup was taken from. Defaults to this instance.

### spec.restoreBackupContext.project

`string`

The full project ID of the source instance.

### spec.pointInTimeRestoreContext

`GcpCloudSqlPointInTimeRestoreContext`

Restores this instance from a Backup and DR datasource to a point in
time. Like restore_backup_context, adding or changing the block
triggers the restore.

### spec.pointInTimeRestoreContext.datasource

`string` · required

The Backup and DR datasource URI to restore from.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.pointInTimeRestoreContext.pointInTime

`string` · required

RFC 3339 timestamp to restore to.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.pointInTimeRestoreContext.targetInstance

`string`

Name of the target instance receiving the restore.

### spec.pointInTimeRestoreContext.region

`string`

Region of the target instance, e.g. "us-central1".

### spec.pointInTimeRestoreContext.preferredZone

`string`

The zone the restored instance lands in. Defaults to the source
primary's zone.

### spec.pointInTimeRestoreContext.allocatedIpRange

`string`

Private-IP restores: the allocated IP range name (RFC 1035) the
restored instance's private IP is drawn from.

### spec.backupdrBackup

`string`

Restores from a Backup and DR backup (the backup's full resource
name). The backup must be in active state. Adding or changing this
triggers the restore.

### spec.maintenanceVersion

`string`

Pins the instance's maintenance (patch) version. Cannot be set at
creation; updating it restarts the instance. Values older than the
running version are ignored by the API.

### spec.replicaNames

`[]string`

Declares the instance's read replicas by name from the primary's side.
Most compositions leave this to GCP (replicas declare their primary
via master_instance_name instead).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","repeated":{"unique":true,"items":{"string":{"minLen":"1"}}}}

### spec.failoverDrReplicaName

`string`

MySQL/PostgreSQL disaster recovery: names this primary's DR replica
("project:instance" or plain instance name), enabling switchover /
replica failover between regions.

### spec.autoUpgradeEnabled

`bool`

MySQL 8.0 only: opts the instance into automatic minor-version
upgrades. The database_version must be an eligible MYSQL_8_0 minor.

### spec.dataApiAccess

`string`

Whether the ExecuteSql API may connect to this instance.
DISALLOW_DATA_API (default) rejects it; ALLOW_DATA_API admits it —
on private-IP instances this allows authorized users to reach the
instance from the public internet through the API.

- rule: data_api_access must be empty, ALLOW_DATA_API, or DISALLOW_DATA_API

### spec.finalBackup

`GcpCloudSqlFinalBackup`

A final backup taken automatically when the instance is deleted — the
safety net that survives the teardown itself.

### spec.finalBackup.enabled

`bool`

Whether a final backup is taken on delete.

### spec.finalBackup.retentionDays

`int32` · optional (explicit presence)

Days the final backup is retained.

- rule: {"int32":{"gte":1}}

### spec.finalBackup.description

`string`

Description recorded on the final backup.

### spec.entraId

`GcpCloudSqlEntraIdConfig`

SQL Server only: Microsoft Entra ID (Azure AD) authentication for the
instance.

### spec.entraId.applicationId

`string` · required

The Entra ID application (client) ID.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.entraId.tenantId

`string` · required

The Entra ID tenant (directory) ID.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.switchTransactionLogsToCloudStorageEnabled

`bool`

Opt-in that lets the instance move point-in-time-recovery transaction
logs from the data disk to Cloud Storage, freeing disk space and
allowing longer transaction-log retention windows. An input-only
instruction: Cloud SQL acts on it but never stores it, so it is sent
exactly as written and never read back.

### spec.includeReplicasForMajorVersionUpgrade

`bool`

Opt-in that upgrades this primary's read replicas in place, together
with the primary, when database_version moves to a new major version.
Without it a major-version upgrade leaves replicas on the old version
to be upgraded (or recreated) separately. Input-only: consulted only
during a major-version upgrade and never stored by the API.

### spec.enforceNewSqlNetworkArchitecture

`bool` · optional (explicit presence)

Irreversible opt-in to Cloud SQL's new network architecture for an
instance created in a project that predates it (projects created after
August 2021 already use it). Required before features such as PSC
and outbound network attachments on those older projects. Once true
it cannot be set back to false. Leave unset to let Cloud SQL report
the instance's current architecture; sent only when set because the
API fills the value itself.

### spec.replicationLagMaxSeconds

`int32` · optional (explicit presence)

Read replicas only: the replication lag, in seconds, beyond which the
replica recreates itself. The lag must persist for at least five
minutes before recreation triggers. Between 300 (five minutes) and
31536000 (one year). Leave unset for no automatic recreation; sent
only when set because the API fills the value itself.

- rule: {"int32":{"lte":31536000,"gte":300}}

## Validation Rules

- `database_version_matches_engine`: database_version must match database_engine: MYSQL_* for MYSQL, POSTGRES_* for POSTGRESQL, SQLSERVER_* for SQLSERVER
- `sqlserver_requires_root_password`: root_password is required for SQL Server instances — the engine cannot bootstrap without an initial 'sqlserver' user password
- `pitr_requires_postgres_or_sqlserver`: point_in_time_recovery_enabled applies to POSTGRESQL and SQLSERVER only — MySQL point-in-time recovery is enabled via backup.binary_log_enabled instead
- `binary_log_requires_mysql`: backup.binary_log_enabled applies to MYSQL only — PostgreSQL and SQL Server use point_in_time_recovery_enabled instead
- `regional_requires_backups`: availability_type REGIONAL (high availability) requires automated backups — set backup.enabled to true
- `regional_mysql_requires_binary_log`: a REGIONAL (high availability) MySQL instance requires backup.binary_log_enabled to be true
- `data_cache_requires_enterprise_plus`: data_cache_enabled requires edition ENTERPRISE_PLUS — the data cache is an Enterprise Plus feature
- `time_zone_requires_sqlserver`: time_zone applies to SQL Server instances only
- `collation_requires_sqlserver`: collation (server-level) applies to SQL Server instances only — MySQL and PostgreSQL set collation per database
- `threads_per_core_requires_sqlserver`: threads_per_core applies to SQL Server instances only
- `audit_config_requires_sqlserver`: sql_server_audit_config applies to SQL Server instances only
- `active_directory_requires_sqlserver`: active_directory applies to SQL Server instances only
- `entra_id_requires_sqlserver`: entra_id applies to SQL Server instances only
- `read_pool_requires_read_pool_instance`: node_count and read_pool_auto_scale apply to read pools only — set instance_type to READ_POOL_INSTANCE
- `final_backup_description_requires_enabled`: final_backup.description applies only when final_backup.enabled is true
- `password_change_interval_requires_postgres`: password_validation_policy.password_change_interval applies to POSTGRESQL instances only
- `replica_configuration_requires_master`: replica_configuration is only meaningful on a read replica — set master_instance_name to the primary instance
- `secondary_zone_requires_regional`: location_preference.secondary_zone applies only when availability_type is REGIONAL

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpCloudSql, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.instance_name` | `string` | Name of the Cloud SQL instance. The composition key: GcpCloudSqlDatabase and GcpCloudSqlUser reference an instance by this name, and a read replica's master_instance_name resolves to it. |
| `status.outputs.connection_name` | `string` | Full connection name in the format project:region:instance — what the Cloud SQL Auth Proxy, connectors, and serverless integrations consume. |
| `status.outputs.private_ip` | `string` | Private IP address of the instance (empty unless private_network is configured). |
| `status.outputs.public_ip` | `string` | Public IPv4 address of the instance (empty unless ipv4_enabled). |
| `status.outputs.self_link` | `string` | GCP resource self link for the Cloud SQL instance. |
| `status.outputs.service_account_email` | `string` | The Google-managed service account this instance runs as. Grant it GCS access to enable imports/exports and SQL Server audit uploads. |
| `status.outputs.dns_name` | `string` | DNS name of the instance (populated for PSC-enabled instances). |
| `status.outputs.psc_service_attachment_link` | `string` | The PSC service attachment link consumers target with PSC endpoints (populated only when psc.enabled is true). |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.network.privateNetwork` | GcpVpcNetwork | `status.outputs.network_id` |
| `spec.encryptionKeyName` | GcpKmsKey | `status.outputs.key_id` |
| `spec.masterInstanceName` | GcpCloudSql | `status.outputs.instance_name` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| GcpCloudRun | `spec.volumes[].cloudSqlInstance.instances` | `status.outputs.connection_name` |
| GcpCloudRunJob | `spec.template.volumes[].cloudSqlInstance.instances` | `status.outputs.connection_name` |
| GcpCloudSql | `spec.masterInstanceName` | `status.outputs.instance_name` |
| GcpCloudSqlDatabase | `spec.instance` | `status.outputs.instance_name` |
| GcpCloudSqlUser | `spec.instance` | `status.outputs.instance_name` |

## See Also

- [Overview](../README.md)
