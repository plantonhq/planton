# AwsRdsInstance

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

AwsRdsInstanceSpec defines a single RDS DB instance: a standalone
database server (postgres, mysql, mariadb, oracle-*, sqlserver-*) with
its own EBS-backed storage -- optionally Multi-AZ with a synchronous
standby, or a read replica of another instance.

This is the classic single-node RDS shape. For Aurora's shared-storage
clusters (and Multi-AZ DB clusters of mysql/postgres), use
AwsRdsCluster instead -- cluster members are modeled there, never here.

The instance identifier comes from metadata.name. Security groups,
subnets, KMS keys, and the monitoring role compose by reference --
database ingress rules belong on the referenced AwsSecurityGroup node,
never inside this instance.

## Example

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsRdsInstance
metadata:
  name: awsrdsinstance-demo
spec:
  region: us-west-2
  subnetIds:
    - value: subnet-0a1b2c3d4e5f60001
    - value: subnet-0a1b2c3d4e5f60002
  engine: postgres
  engineVersion: "16.4"
  instanceClass: db.t4g.micro
  allocatedStorageGb: 20
  storageType: gp3
  storageEncrypted: true
  username: hackadmin
  manageMasterUserPassword: true
  skipFinalSnapshot: true
  # Opt out of paid extended support: upgrade before standard support ends.
  engineLifecycleSupport: open-source-rds-extended-support-disabled
  # Instance-owned parameter group from inline parameters (the family is
  # derived from engine + engineVersion).
  parameters:
    - name: rds.force_ssl
      value: "1"
      applyMethod: immediate
  # Engine feature roles, one association per entry (feature_name is
  # required on instance associations).
  iamRoles:
    - role:
        value: arn:aws:iam::123456789012:role/rds-s3-export
      featureName: s3Export
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.subnetIds` | `[]string \| valueFrom` |  |  | AwsSubnet (`status.outputs.subnet_id`) |
| `spec.dbSubnetGroupName` | `string` |  |  |  |
| `spec.securityGroupIds` | `[]string \| valueFrom` |  |  | AwsSecurityGroup (`status.outputs.security_group_id`) |
| `spec.engine` | `string` |  |  |  |
| `spec.engineVersion` | `string` |  |  |  |
| `spec.instanceClass` | `string` | yes |  |  |
| `spec.allocatedStorageGb` | `int32` |  |  |  |
| `spec.maxAllocatedStorageGb` | `int32` |  |  |  |
| `spec.storageType` | `string` |  |  |  |
| `spec.iops` | `int32` |  |  |  |
| `spec.storageThroughput` | `int32` |  |  |  |
| `spec.dedicatedLogVolume` | `bool` |  |  |  |
| `spec.storageEncrypted` | `bool` |  | `true` |  |
| `spec.kmsKeyId` | `string \| valueFrom` |  |  | AwsKmsKey (`status.outputs.key_arn`) |
| `spec.dbName` | `string` |  |  |  |
| `spec.username` | `string` |  |  |  |
| `spec.manageMasterUserPassword` | `bool` |  | `true` |  |
| `spec.masterUserSecretKmsKeyId` | `string \| valueFrom` |  |  | AwsKmsKey (`status.outputs.key_arn`) |
| `spec.password` | `string` (sensitive) |  |  |  |
| `spec.port` | `int32` |  |  |  |
| `spec.multiAz` | `bool` |  |  |  |
| `spec.availabilityZone` | `string` |  |  |  |
| `spec.publiclyAccessible` | `bool` |  |  |  |
| `spec.networkType` | `string` |  |  |  |
| `spec.replicateSourceDb` | `string` |  |  |  |
| `spec.replicaMode` | `string` |  |  |  |
| `spec.snapshotIdentifier` | `string` |  |  |  |
| `spec.restoreToPointInTime` | `AwsRdsInstanceRestoreToPointInTime` |  |  |  |
| `spec.restoreToPointInTime.sourceDbInstanceIdentifier` | `string` |  |  |  |
| `spec.restoreToPointInTime.sourceDbiResourceId` | `string` |  |  |  |
| `spec.restoreToPointInTime.sourceDbInstanceAutomatedBackupsArn` | `string` |  |  |  |
| `spec.restoreToPointInTime.restoreTime` | `string` |  |  |  |
| `spec.restoreToPointInTime.useLatestRestorableTime` | `bool` |  |  |  |
| `spec.backupRetentionPeriod` | `int32` |  |  |  |
| `spec.backupWindow` | `string` |  |  |  |
| `spec.maintenanceWindow` | `string` |  |  |  |
| `spec.copyTagsToSnapshot` | `bool` |  |  |  |
| `spec.deleteAutomatedBackups` | `bool` |  | `true` |  |
| `spec.skipFinalSnapshot` | `bool` |  |  |  |
| `spec.finalSnapshotIdentifier` | `string` |  |  |  |
| `spec.deletionProtection` | `bool` |  |  |  |
| `spec.iamDatabaseAuthenticationEnabled` | `bool` |  |  |  |
| `spec.enabledCloudwatchLogsExports` | `[]string` |  |  |  |
| `spec.performanceInsightsEnabled` | `bool` |  |  |  |
| `spec.performanceInsightsKmsKeyId` | `string \| valueFrom` |  |  | AwsKmsKey (`status.outputs.key_arn`) |
| `spec.performanceInsightsRetentionPeriod` | `int32` |  |  |  |
| `spec.monitoringInterval` | `int32` |  |  |  |
| `spec.monitoringRoleArn` | `string \| valueFrom` |  |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.databaseInsightsMode` | `string` |  |  |  |
| `spec.parameterGroupName` | `string` |  |  |  |
| `spec.optionGroupName` | `string` |  |  |  |
| `spec.activeDirectory` | `AwsRdsInstanceActiveDirectory` |  |  |  |
| `spec.activeDirectory.domain` | `string` |  |  |  |
| `spec.activeDirectory.domainIamRoleName` | `string` |  |  |  |
| `spec.activeDirectory.domainFqdn` | `string` |  |  |  |
| `spec.activeDirectory.domainOu` | `string` |  |  |  |
| `spec.activeDirectory.domainAuthSecretArn` | `string` |  |  |  |
| `spec.activeDirectory.domainDnsIps` | `[]string` |  |  |  |
| `spec.licenseModel` | `string` |  |  |  |
| `spec.characterSetName` | `string` |  |  |  |
| `spec.ncharCharacterSetName` | `string` |  |  |  |
| `spec.timezone` | `string` |  |  |  |
| `spec.caCertIdentifier` | `string` |  |  |  |
| `spec.blueGreenUpdateEnabled` | `bool` |  |  |  |
| `spec.autoMinorVersionUpgrade` | `bool` |  | `true` |  |
| `spec.allowMajorVersionUpgrade` | `bool` |  |  |  |
| `spec.applyImmediately` | `bool` |  |  |  |
| `spec.engineLifecycleSupport` | `string` |  |  |  |
| `spec.upgradeStorageConfig` | `bool` |  |  |  |
| `spec.s3Import` | `AwsRdsInstanceS3Import` |  |  |  |
| `spec.s3Import.bucketName` | `string` | yes |  |  |
| `spec.s3Import.bucketPrefix` | `string` |  |  |  |
| `spec.s3Import.ingestionRole` | `string \| valueFrom` | yes |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.s3Import.sourceEngine` | `string` | yes |  |  |
| `spec.s3Import.sourceEngineVersion` | `string` | yes |  |  |
| `spec.iamRoles` | `[]AwsRdsInstanceIamRole` |  |  |  |
| `spec.iamRoles[].role` | `string \| valueFrom` | yes |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.iamRoles[].featureName` | `string` | yes |  |  |
| `spec.parameters` | `[]AwsRdsInstanceParameter` |  |  |  |
| `spec.parameters[].name` | `string` | yes |  |  |
| `spec.parameters[].value` | `string` | yes |  |  |
| `spec.parameters[].applyMethod` | `string` |  |  |  |
| `spec.options` | `[]AwsRdsInstanceOption` |  |  |  |
| `spec.options[].optionName` | `string` | yes |  |  |
| `spec.options[].optionSettings` | `[]AwsRdsInstanceOptionSetting` |  |  |  |
| `spec.options[].optionSettings[].name` | `string` | yes |  |  |
| `spec.options[].optionSettings[].value` | `string` | yes |  |  |
| `spec.options[].port` | `int32` |  |  |  |
| `spec.options[].version` | `string` |  |  |  |
| `spec.options[].vpcSecurityGroupMemberships` | `[]string \| valueFrom` |  |  | AwsSecurityGroup (`status.outputs.security_group_id`) |

## Field Details

### spec.region

`string` · required

The AWS region the instance is created in. Must match the region of
the subnets, security groups, and KMS keys it references.
Example: "us-west-2", "eu-west-1".

- rule: {"string":{"minLen":"1"}}

### spec.subnetIds

`[]string | valueFrom`

Subnets for the instance's DB subnet group. Provide at least two
subnets in DISTINCT availability zones -- AWS requires the subnet
group to cover two AZs even for a single-AZ instance. Reference
AwsSubnet subnet_id outputs or pass literal subnet IDs. The module
manages the subnet group itself (pure glue); alternatively point
db_subnet_group_name at an existing group.

- references: AwsSubnet (`status.outputs.subnet_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSubnet, name: <that resource's name>, fieldPath: status.outputs.subnet_id}} -- a bare string does not parse

### spec.dbSubnetGroupName

`string`

Name of an existing DB subnet group to place the instance in,
instead of providing subnet_ids.

### spec.securityGroupIds

`[]string | valueFrom`

Security groups attached to the instance. Empty uses the VPC's
default security group (the AWS default). Reference AwsSecurityGroup
security_group_id outputs or pass literal SG IDs.

- references: AwsSecurityGroup (`status.outputs.security_group_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSecurityGroup, name: <that resource's name>, fieldPath: status.outputs.security_group_id}} -- a bare string does not parse

### spec.engine

`string`

The database engine: "postgres", "mysql", "mariadb", "oracle-ee",
"oracle-se2", "sqlserver-ex", "sqlserver-web", "sqlserver-se",
"sqlserver-ee", and license-included/CDB variants. Required for a
new instance (a read replica inherits it from the source). Changing
the engine replaces the instance.

### spec.engineVersion

`string`

The engine version, e.g. "16.4" (postgres) or "8.0.39" (mysql).
Leave empty to let AWS pick the engine's current default version --
an empty pin never goes stale. Minor upgrades apply in place; major
upgrades additionally need allow_major_version_upgrade.

### spec.instanceClass

`string` · required

The instance class (compute size), e.g. "db.t4g.micro",
"db.m6g.large". Required.

- rule: {"string":{"minLen":"1","pattern":"^db\\."}}

### spec.allocatedStorageGb

`int32`

Provisioned storage in GiB. Required for a new instance (read
replicas and snapshot restores inherit the source's storage).
Growing it applies in place; shrinking requires a new instance.

- rule: {"int32":{"gte":0}}

### spec.maxAllocatedStorageGb

`int32`

Storage autoscaling ceiling in GiB. When set above
allocated_storage_gb, RDS grows storage automatically as the
database approaches capacity -- the cheap insurance against
disk-full outages. 0 disables autoscaling.

- rule: {"int32":{"gte":0}}

### spec.storageType

`string`

The EBS storage type: "gp3" (the modern default -- baseline
3000 IOPS/125 MiB/s, independently tunable), "gp2" (legacy
burst-credit SSD), "io1"/"io2" (provisioned-IOPS for
latency-critical workloads), or "standard" (magnetic, legacy).
Empty keeps the AWS default.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["gp2","gp3","io1","io2","standard"]}}

### spec.iops

`int32`

Provisioned IOPS. Required for io1/io2; optional for gp3 to raise
it above the 3000 baseline. 0 keeps the storage type's default.

- rule: {"int32":{"gte":0}}

### spec.storageThroughput

`int32`

Storage throughput in MiB/s, gp3 only, to raise it above the 125
baseline. 0 keeps the default.

- rule: {"int32":{"gte":0}}

### spec.dedicatedLogVolume

`bool`

Use a dedicated EBS volume for database logs instead of sharing the
data volume -- steadier I/O for audit-heavy or WAL-heavy workloads.

### spec.storageEncrypted

`bool`

Encrypt instance storage at rest. Strongly recommended -- and
create-time only: an unencrypted instance cannot be encrypted later
(requires a snapshot-restore migration). Read replicas inherit the
source's encryption.

- default: `true`

### spec.kmsKeyId

`string | valueFrom`

The KMS key for storage encryption when storage_encrypted is true.
Empty uses the AWS-managed aws/rds key. Reference an AwsKmsKey
key_arn output or pass a literal key ARN. Create-time only.

- references: AwsKmsKey (`status.outputs.key_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_arn}} -- a bare string does not parse

### spec.dbName

`string`

The name of the initial database AWS creates. Empty creates no
database. Create-time only. Not supported by SQL Server.

### spec.username

`string`

The master username. Required for a brand-new instance -- AWS has
no default and rejects a blank value at CreateDBInstance. Only
instances that inherit credentials from a source (a read replica,
a snapshot restore, or a point-in-time restore) leave it empty.
Avoid the engine's reserved names (e.g. "rdsadmin"). Create-time
only -- changing it replaces the instance.

### spec.manageMasterUserPassword

`bool`

Let AWS manage the master password in Secrets Manager: AWS
generates it, stores it, rotates it on schedule, and no secret ever
touches this manifest or the IaC state. The managed secret's ARN is
exported as the master_user_secret_arn output. Mutually exclusive
with password -- and the recommended posture.

- default: `true`

### spec.masterUserSecretKmsKeyId

`string | valueFrom`

The KMS key that encrypts the AWS-managed master-user secret (only
meaningful with manage_master_user_password). Empty uses the
account's default aws/secretsmanager key. Reference an AwsKmsKey
key_arn output or pass a literal key ARN.

- references: AwsKmsKey (`status.outputs.key_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_arn}} -- a bare string does not parse

### spec.password

`string` · sensitive

The master password, supplied directly. Stored in IaC state --
prefer manage_master_user_password, which keeps the secret in
Secrets Manager entirely. Mutually exclusive with
manage_master_user_password.

### spec.port

`int32`

The port the instance accepts connections on. 0 keeps the engine
default (5432 postgres, 3306 mysql/mariadb, 1521 oracle, 1433
sqlserver).

- rule: {"int32":{"lte":65535,"gte":0}}

### spec.multiAz

`bool`

Deploy a synchronous standby replica in a second AZ with automatic
failover -- the single most important availability knob on a
production instance. (For a Multi-AZ CLUSTER with readable
standbys, use AwsRdsCluster with a community engine instead.)

### spec.availabilityZone

`string`

Pin a single-AZ instance to one availability zone. Empty lets AWS
place it. Cannot be combined with multi_az.

### spec.publiclyAccessible

`bool`

Give the instance a public IP. Requires public subnets; keep false
for anything production-shaped.

### spec.networkType

`string`

The network stack: "IPV4" (AWS default when unset) or "DUAL" for
dual-stack IPv4+IPv6.

### spec.replicateSourceDb

`string`

Make this instance a read replica of the given source: a source
instance identifier (same region) or ARN (cross-region). Engine,
storage, and credentials are inherited from the source. Promote to
a standalone instance by clearing the field.

### spec.replicaMode

`string`

Replica behavior, Oracle only: "open-read-only" (a queryable
replica) or "mounted" (a running-but-closed disaster-recovery
target). Empty keeps the AWS default (open-read-only).

### spec.snapshotIdentifier

`string`

Restore from an existing DB snapshot (name or ARN) at create time.
Create-time only; mutually exclusive with restore_to_point_in_time
and replicate_source_db.

### spec.restoreToPointInTime

`AwsRdsInstanceRestoreToPointInTime`

Restore from another instance's continuous backup at create time --
point-in-time recovery as a first-class create shape. Create-time
only; mutually exclusive with snapshot_identifier and
replicate_source_db.

- rule: exactly one of source_db_instance_identifier, source_dbi_resource_id, or source_db_instance_automated_backups_arn must be set
- rule: exactly one of restore_time or use_latest_restorable_time must be set

### spec.restoreToPointInTime.sourceDbInstanceIdentifier

`string`

The source instance identifier. Exactly one source field must be
set.

### spec.restoreToPointInTime.sourceDbiResourceId

`string`

The source instance's immutable resource ID (resource_id output) --
survives identifier renames and points at deleted instances'
retained backups.

### spec.restoreToPointInTime.sourceDbInstanceAutomatedBackupsArn

`string`

The ARN of a retained automated backup to restore from (for
instances already deleted).

### spec.restoreToPointInTime.restoreTime

`string`

The UTC timestamp to restore to, RFC3339 (e.g.
"2026-07-01T09:45:00Z"). Mutually exclusive with
use_latest_restorable_time.

### spec.restoreToPointInTime.useLatestRestorableTime

`bool`

Restore to the most recent recoverable moment. Mutually exclusive
with restore_time.

### spec.backupRetentionPeriod

`int32`

Days automated backups are retained, 0-35. 0 disables automated
backups (and point-in-time recovery) -- production instances want
7+. Note: a MySQL-family read-replica source needs retention > 0.

- rule: {"int32":{"lte":35,"gte":0}}

### spec.backupWindow

`string`

The daily backup window in UTC, format "hh24:mi-hh24:mi" (e.g.
"04:00-05:00"). Empty lets AWS assign one. Must not overlap the
maintenance window.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^([01][0-9]|2[0-3]):[0-5][0-9]-([01][0-9]|2[0-3]):[0-5][0-9]$"}}

### spec.maintenanceWindow

`string`

The weekly maintenance window in UTC, format
"ddd:hh24:mi-ddd:hh24:mi" (e.g. "sun:05:00-sun:06:00"). Empty lets
AWS assign one.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^(mon|tue|wed|thu|fri|sat|sun):([01][0-9]|2[0-3]):[0-5][0-9]-(mon|tue|wed|thu|fri|sat|sun):([01][0-9]|2[0-3]):[0-5][0-9]$"}}

### spec.copyTagsToSnapshot

`bool`

Copy the instance's tags onto automated and manual snapshots.

### spec.deleteAutomatedBackups

`bool` · optional (explicit presence)

Remove automated backups immediately when the instance is deleted.
AWS defaults this to true; set false to retain the backups for the
remainder of their retention window after deletion -- the last line
of defense against a mistaken teardown.

- default: `true`

### spec.skipFinalSnapshot

`bool`

Skip the final snapshot when the instance is deleted. When false
(the safe default), final_snapshot_identifier must be set -- AWS
refuses to delete without knowing the snapshot name.

### spec.finalSnapshotIdentifier

`string`

The name for the final snapshot taken on deletion. Required when
skip_final_snapshot is false. Must start with a letter, contain
only letters, numbers, and hyphens, with no consecutive or
trailing hyphens -- AWS's snapshot-identifier rules.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^[A-Za-z][0-9A-Za-z]*(-[0-9A-Za-z]+)*$"}}

### spec.deletionProtection

`bool`

Refuse deletion of the instance while enabled. Turn this on for
anything holding data you cannot recreate.

### spec.iamDatabaseAuthenticationEnabled

`bool`

Map IAM identities to database users -- connect with short-lived
IAM auth tokens instead of passwords. MySQL and PostgreSQL engines.

### spec.enabledCloudwatchLogsExports

`[]string`

Database log types to export to CloudWatch Logs. Valid types vary
by engine: postgres exports "postgresql"/"upgrade"/"iam-db-auth-error";
mysql/mariadb export "audit"/"error"/"general"/"slowquery"/
"iam-db-auth-error"; oracle exports "alert"/"audit"/"listener"/
"trace"/"oemagent"; sqlserver exports "agent"/"error".

- rule: {"repeated":{"unique":true,"items":{"string":{"in":["agent","alert","audit","diag.log","error","general","iam-db-auth-error","listener","notify.log","oemagent","postgresql","slowquery","trace","upgrade"]}}}}

### spec.performanceInsightsEnabled

`bool`

Performance Insights: per-query performance telemetry. Free at the
default 7-day retention -- worth enabling on almost everything.

### spec.performanceInsightsKmsKeyId

`string | valueFrom`

The KMS key encrypting Performance Insights data. Empty uses the
AWS default. Reference an AwsKmsKey key_arn output or pass a
literal key ARN. Cannot change after Performance Insights is first
enabled.

- references: AwsKmsKey (`status.outputs.key_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_arn}} -- a bare string does not parse

### spec.performanceInsightsRetentionPeriod

`int32`

Days of Performance Insights history: 7 (free tier), 731 (2
years), or any multiple of 31 in between. 0 keeps the AWS default
(7).

### spec.monitoringInterval

`int32`

Enhanced Monitoring granularity in seconds: 1, 5, 10, 15, 30, or
60. 0 disables (the AWS default). OS-level metrics (CPU per
process, memory, disk) streamed to CloudWatch Logs -- requires
monitoring_role_arn.

### spec.monitoringRoleArn

`string | valueFrom`

The IAM role Enhanced Monitoring publishes through (needs the
AmazonRDSEnhancedMonitoringRole managed policy). Required by AWS
when monitoring_interval is set. Reference an AwsIamRole role_arn
output or pass a literal ARN.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.databaseInsightsMode

`string`

CloudWatch Database Insights tier: "standard" (free, included) or
"advanced" (paid fleet-level analysis; requires Performance
Insights with 465+ day retention). Empty keeps the AWS default
(standard).

### spec.parameterGroupName

`string`

The name of an existing DB parameter group to associate. Mutually
exclusive with `parameters` -- either bring your own group or let
the module manage one from inline parameters. Empty (with
`parameters` also empty) keeps the engine default group.

### spec.optionGroupName

`string`

The name of an existing option group to associate. Mutually
exclusive with `options` -- either bring your own group or let the
module manage one from inline options. Empty (with `options` also
empty) keeps the engine default group.

### spec.activeDirectory

`AwsRdsInstanceActiveDirectory`

Microsoft Active Directory domain join, for Windows/Kerberos
authentication (SQL Server, and Kerberos auth on MySQL/PostgreSQL/
Oracle).

- rule: use either the AWS-managed shape (domain + domain_iam_role_name) or the self-managed shape (domain_fqdn/domain_ou/domain_auth_secret_arn/domain_dns_ips), not both
- rule: domain and domain_iam_role_name are required together for the AWS-managed directory shape
- rule: self-managed AD requires exactly two domain_dns_ips

### spec.activeDirectory.domain

`string`

The directory ID of an AWS Managed Microsoft AD (d-...). Pairs with
domain_iam_role_name.

### spec.activeDirectory.domainIamRoleName

`string`

The IAM role RDS uses to join the AWS-managed directory (needs the
AmazonRDSDirectoryServiceAccess managed policy). Required with
domain.

### spec.activeDirectory.domainFqdn

`string`

The fully qualified domain name of a self-managed Active Directory
(e.g. "corp.example.com").

### spec.activeDirectory.domainOu

`string`

The organizational unit DN within the self-managed AD where the
computer account is created.

### spec.activeDirectory.domainAuthSecretArn

`string`

The ARN of the Secrets Manager secret holding the self-managed AD
join credentials. An ARN reference, not the secret itself.

### spec.activeDirectory.domainDnsIps

`[]string`

Exactly two DNS server IPs inside the self-managed AD.

### spec.licenseModel

`string`

The license model, for engines that carry one. Per AWS's contract:
MySQL/MariaDB use "general-public-license"; PostgreSQL uses
"postgresql-license"; Oracle uses "bring-your-own-license" or
"license-included"; SQL Server uses "license-included" or
"bring-your-own-media"; Db2 uses "bring-your-own-license" or
"marketplace-license". Empty keeps the engine default.

### spec.characterSetName

`string`

The character set for Oracle and SQL Server instances (e.g.
"AL32UTF8"). Create-time only. Empty keeps the engine default.

### spec.ncharCharacterSetName

`string`

The national character set (NCHAR) for Oracle instances. Create-time
only. Empty keeps the engine default.

### spec.timezone

`string`

The time zone, SQL Server only (e.g. "GMT Standard Time").
Create-time only.

### spec.caCertIdentifier

`string`

The CA certificate bundle for the instance (e.g.
"rds-ca-rsa2048-g1"). Empty keeps the AWS default bundle.

### spec.blueGreenUpdateEnabled

`bool`

Use RDS Blue/Green Deployments for updates: RDS provisions a
synchronized green copy, applies the change there, and switches
over in under a minute -- near-zero-downtime engine upgrades and
parameter changes. MySQL, MariaDB, and PostgreSQL; not compatible
with read replicas of this instance being modified in the same
operation.

### spec.autoMinorVersionUpgrade

`bool` · optional (explicit presence)

Apply minor engine version patches automatically during the
maintenance window. AWS defaults this to true; disable only when
patch timing must be controlled manually.

- default: `true`

### spec.allowMajorVersionUpgrade

`bool`

Permit engine_version changes that cross a major version. Off (the
default) guards against an accidental major upgrade hidden in a
version bump.

### spec.applyImmediately

`bool`

Apply modifications immediately instead of waiting for the next
maintenance window. Immediate changes can interrupt connections;
deferred changes wait quietly. AWS defaults to deferred.

### spec.engineLifecycleSupport

`string`

Extended support posture when the engine version leaves standard
support: "open-source-rds-extended-support" (AWS default -- paid
extended support kicks in automatically) or
"open-source-rds-extended-support-disabled" (the instance must be
upgraded before end of standard support; opts out of the extra
cost).

### spec.upgradeStorageConfig

`bool`

Upgrade the storage file system configuration on a read replica or
snapshot restore, when the source still runs the older 32-bit file
system. One-way and create/restore-scoped.

### spec.s3Import

`AwsRdsInstanceS3Import`

Create the instance by restoring a Percona XtraBackup stored in S3
-- the on-ramp for migrating a self-managed MySQL database into
RDS without a logical dump. Create-time only; mutually exclusive
with replicate_source_db, snapshot_identifier, and
restore_to_point_in_time.

### spec.s3Import.bucketName

`string` · required

The S3 bucket holding the backup files. Required.

- rule: {"required":true}

### spec.s3Import.bucketPrefix

`string`

The key prefix of the backup files within the bucket. Empty reads
from the bucket root.

### spec.s3Import.ingestionRole

`string | valueFrom` · required

The IAM role RDS assumes to read the backup from S3. Reference an
AwsIamRole role_arn output or pass a literal role ARN. Required.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.s3Import.sourceEngine

`string` · required

The engine of the source backup. AWS accepts only "mysql" for
instance S3 restores. Required.

- rule: {"required":true,"string":{"in":["mysql"]}}

### spec.s3Import.sourceEngineVersion

`string` · required

The version of the source engine the backup was taken from (e.g.
"8.0"). Required.

- rule: {"required":true}

### spec.iamRoles

`[]AwsRdsInstanceIamRole`

IAM roles the instance assumes for engine features that reach into
other AWS services (e.g. S3 import/export, Lambda invocation).
Each entry associates one role to one named engine feature -- the
roles own their policies; this instance only associates them. Both
IaC modules manage each entry as its own role-association
resource, so roles attach and detach without touching the
instance.

### spec.iamRoles[].role

`string | valueFrom` · required

The IAM role to associate. Reference an AwsIamRole role_arn output
or pass a literal role ARN. The role owns its policies; the
instance only assumes it. The role's trust policy MUST allow
rds.amazonaws.com to assume it -- AWS validates that server-side at
association time and rejects the call with InvalidParameterValue
("IAM role ARN value is invalid or does not include the required
permissions") otherwise; no plan-time check catches it. Required.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.iamRoles[].featureName

`string` · required

The engine feature the role is linked to (e.g. "s3Import",
"s3Export", "Lambda", "S3_INTEGRATION" -- the valid set is
engine-specific). Required -- AWS rejects an instance role
association without a feature name. Changing it replaces the
association (the instance is untouched).

- rule: {"required":true}

### spec.parameters

`[]AwsRdsInstanceParameter`

Instance-level engine parameters, managed as a dedicated DB
parameter group owned by this instance (the group is glue -- a
named parameter list -- so it stays folded). Mutually exclusive
with parameter_group_name.

- rule: apply_method must be 'immediate' or 'pending-reboot' when set

### spec.parameters[].name

`string` · required

The parameter name (e.g. "max_connections",
"rds.force_ssl"). Required.

- rule: {"required":true}

### spec.parameters[].value

`string` · required

The parameter value. Required.

- rule: {"required":true}

### spec.parameters[].applyMethod

`string`

When the change lands: "immediate" (AWS default -- dynamic
parameters apply now) or "pending-reboot" (static parameters wait
for the next instance reboot).

### spec.options

`[]AwsRdsInstanceOption`

Engine options (e.g. Oracle TDE/OEM, SQL Server native backup),
managed as a dedicated option group owned by this instance (the
group is glue -- a named option list -- so it stays folded).
Mutually exclusive with option_group_name.

### spec.options[].optionName

`string` · required

The option name (e.g. "TDE", "OEM", "SQLSERVER_BACKUP_RESTORE").
Required.

- rule: {"required":true}

### spec.options[].optionSettings

`[]AwsRdsInstanceOptionSetting`

Settings the option exposes, as name/value pairs (e.g. the
IAM_ROLE_ARN setting of SQLSERVER_BACKUP_RESTORE).

### spec.options[].optionSettings[].name

`string` · required

The setting name. Required.

- rule: {"required":true}

### spec.options[].optionSettings[].value

`string` · required

The setting value. Required.

- rule: {"required":true}

### spec.options[].port

`int32`

The port the option listens on, for options that open one (e.g.
OEM's 1158). 0 omits the port.

- rule: {"int32":{"lte":65535,"gte":0}}

### spec.options[].version

`string`

The option version, for options that carry one.

### spec.options[].vpcSecurityGroupMemberships

`[]string | valueFrom`

Security groups granting access to the option's port. Reference
AwsSecurityGroup security_group_id outputs or pass literal SG IDs.

- references: AwsSecurityGroup (`status.outputs.security_group_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSecurityGroup, name: <that resource's name>, fieldPath: status.outputs.security_group_id}} -- a bare string does not parse

## Validation Rules

- `subnets_or_group`: provide at least two subnet_ids (distinct AZs) or an existing db_subnet_group_name
- `engine_required_unless_derived`: engine is required unless the instance derives it from a source (replicate_source_db, snapshot_identifier, or restore_to_point_in_time)
- `storage_required_unless_derived`: allocated_storage_gb is required unless storage is inherited from a source (replicate_source_db, snapshot_identifier, or restore_to_point_in_time)
- `password_xor_managed`: password cannot be set when manage_master_user_password is true -- pick one password strategy
- `username_required_unless_derived`: username is required for a new instance -- AWS rejects a blank username; only read replicas and snapshot/point-in-time restores inherit credentials from their source
- `final_snapshot_id_required_when_not_skipping`: final_snapshot_identifier is required when skip_final_snapshot is false -- AWS refuses to delete the instance without a final snapshot name
- `multi_az_excludes_availability_zone`: availability_zone cannot be pinned on a Multi-AZ instance -- AWS places the primary and standby itself
- `one_create_source`: replicate_source_db, snapshot_identifier, restore_to_point_in_time, and s3_import are mutually exclusive create sources
- `s3_import_is_mysql`: s3_import (Percona XtraBackup restore) is a MySQL feature -- engine must be 'mysql'
- `charset_conflicts_with_create_sources`: character_set_name only applies to a brand-new instance -- it cannot be combined with replicate_source_db, snapshot_identifier, restore_to_point_in_time, or s3_import
- `timezone_not_with_s3_import`: timezone (SQL Server) cannot be combined with s3_import (a MySQL restore)
- `replica_inherits_credentials`: username, password, and manage_master_user_password cannot be set on a read replica -- credentials are inherited from the source
- `replica_mode_requires_replica`: replica_mode only applies to a read replica (replicate_source_db set)
- `replica_mode_valid`: replica_mode must be 'open-read-only' or 'mounted' when set
- `blue_green_not_with_replica`: blue_green_update_enabled cannot be combined with replicate_source_db -- Blue/Green manages replicas itself
- `max_storage_above_allocated`: max_allocated_storage_gb must exceed allocated_storage_gb to enable storage autoscaling (0 disables)
- `iops_requires_provisioned_storage_type`: iops applies to io1, io2, or gp3 storage
- `throughput_is_gp3_only`: storage_throughput only applies to gp3 storage
- `pi_retention_valid`: performance_insights_retention_period must be 7, 731, or a multiple of 31 (month granularity) when set
- `monitoring_interval_valid`: monitoring_interval must be 0 (disabled), 1, 5, 10, 15, 30, or 60 seconds
- `monitoring_role_required_with_interval`: monitoring_role_arn is required when monitoring_interval is set -- Enhanced Monitoring publishes through that role
- `database_insights_mode_valid`: database_insights_mode must be 'standard' or 'advanced' when set
- `network_type_valid`: network_type must be 'IPV4' or 'DUAL' when set
- `license_model_valid`: license_model must be one of 'license-included', 'bring-your-own-license', 'general-public-license', 'postgresql-license', 'marketplace-license', or 'bring-your-own-media' when set
- `engine_lifecycle_support_valid`: engine_lifecycle_support must be 'open-source-rds-extended-support' or 'open-source-rds-extended-support-disabled' when set
- `own_parameters_xor_existing_group`: parameters and parameter_group_name are mutually exclusive -- manage parameters here or bring an existing group
- `parameters_require_engine_and_version`: inline parameters require engine and a pinned engine_version -- the managed parameter group's family is derived from them and must match the running engine
- `own_options_xor_existing_group`: options and option_group_name are mutually exclusive -- manage options here or bring an existing group
- `options_require_engine_and_version`: inline options require engine and a pinned engine_version -- the managed option group's engine name and major version are derived from them and must match the running engine

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsRdsInstance, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.instance_identifier` | `string` | The instance identifier (e.g. "orders-db"). |
| `status.outputs.arn` | `string` | The Amazon Resource Name of the instance. |
| `status.outputs.resource_id` | `string` | The immutable DB instance resource ID (db-...). Survives identifier renames -- the durable handle for point-in-time restores, IAM auth policies, and CloudWatch dimensions. |
| `status.outputs.endpoint` | `string` | The connection endpoint in "address:port" form. |
| `status.outputs.address` | `string` | The DNS address of the instance (endpoint without the port). |
| `status.outputs.port` | `int32` | The port the instance accepts connections on. |
| `status.outputs.hosted_zone_id` | `string` | The Route53 hosted zone ID of the endpoint, for DNS alias records. |
| `status.outputs.engine_version_actual` | `string` | The resolved engine version actually running (meaningful when the spec leaves engine_version to the AWS default). |
| `status.outputs.master_user_secret_arn` | `string` | The ARN of the AWS-managed master-user secret in Secrets Manager. Populated only when manage_master_user_password is true -- the handle applications use to fetch credentials at runtime. |
| `status.outputs.db_subnet_group_name` | `string` | The name of the DB subnet group the instance runs in. |
| `status.outputs.db_parameter_group_name` | `string` | The name of the DB parameter group in use (module-managed from spec.parameters or the referenced existing group; empty when the engine default group applies). |
| `status.outputs.option_group_name` | `string` | The name of the option group in use (module-managed from spec.options or the referenced existing group; empty when the engine default group applies). |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.subnetIds` | AwsSubnet | `status.outputs.subnet_id` |
| `spec.securityGroupIds` | AwsSecurityGroup | `status.outputs.security_group_id` |
| `spec.kmsKeyId` | AwsKmsKey | `status.outputs.key_arn` |
| `spec.masterUserSecretKmsKeyId` | AwsKmsKey | `status.outputs.key_arn` |
| `spec.performanceInsightsKmsKeyId` | AwsKmsKey | `status.outputs.key_arn` |
| `spec.monitoringRoleArn` | AwsIamRole | `status.outputs.role_arn` |
| `spec.s3Import.ingestionRole` | AwsIamRole | `status.outputs.role_arn` |
| `spec.iamRoles[].role` | AwsIamRole | `status.outputs.role_arn` |
| `spec.options[].vpcSecurityGroupMemberships` | AwsSecurityGroup | `status.outputs.security_group_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AwsRdsProxy | `spec.target.dbInstanceIdentifier` | `status.outputs.instance_identifier` |

## See Also

- [Overview](../README.md)
