# AwsRdsCluster

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

AwsRdsClusterSpec defines an RDS DB cluster: an Aurora MySQL/PostgreSQL
cluster (provisioned or Serverless v2), an Aurora Serverless v1 cluster,
or a Multi-AZ RDS cluster for the community mysql/postgres engines.

The cluster is the shared-storage brain -- endpoints, credentials,
backups, encryption, and engine lifecycle live here. The compute that
serves queries is the `instances` list: each entry materializes as its
own DB instance inside the cluster (a writer plus any readers). Cluster
instances are sub-resources of exactly one cluster and are referenced by
nothing else, so they are folded into this spec rather than modeled as a
standalone kind; both IaC modules manage each entry as its own provider
resource keyed by name, so adding or removing a reader is an in-place
update, never a cluster replacement. Aurora Serverless v1
(engine_mode: serverless) and Multi-AZ RDS clusters
(db_cluster_instance_class) are the two shapes where AWS itself manages
the compute and `instances` stays empty.

The cluster identifier comes from metadata.name. Security groups,
subnets, KMS keys, and IAM roles compose by reference -- this cluster
never creates or mutates resources that deserve to be their own nodes.

## Example

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsRdsCluster
metadata:
  name: awsrdscluster-demo
spec:
  region: us-west-2
  subnetIds:
    - value: subnet-0a1b2c3d4e5f60001
    - value: subnet-0a1b2c3d4e5f60002
  engine: aurora-postgresql
  masterUsername: hackadmin
  manageMasterUserPassword: true
  storageEncrypted: true
  skipFinalSnapshot: true
  autoMinorVersionUpgrade: true
  # Kerberos authentication through an AWS Managed Microsoft AD.
  domain: d-9267012345
  domainIamRoleName: rds-directory-service-role
  serverlessV2Scaling:
    minCapacity: 0
    maxCapacity: 1
  instances:
    - name: writer
      instanceClass: db.serverless
      copyTagsToSnapshot: true
      preferredMaintenanceWindow: sun:05:00-sun:06:00
  # Engine feature roles, one association per entry (feature-scoped).
  iamRoles:
    - role:
        value: arn:aws:iam::123456789012:role/aurora-s3-export
      featureName: s3Export
  # A stable analytics endpoint over the chosen instances.
  customEndpoints:
    - name: analytics
      type: READER
      excludedMembers:
        - writer
  # Compliance-grade audit feed: every database event to a KMS-encrypted
  # Kinesis stream (the stream name surfaces as an output).
  activityStream:
    mode: async
    kmsKeyId:
      value: arn:aws:kms:us-west-2:123456789012:key/1234abcd-12ab-34cd-56ef-1234567890ab
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.subnetIds` | `[]string \| valueFrom` |  |  | AwsSubnet (`status.outputs.subnet_id`) |
| `spec.dbSubnetGroupName` | `string` |  |  |  |
| `spec.securityGroupIds` | `[]string \| valueFrom` |  |  | AwsSecurityGroup (`status.outputs.security_group_id`) |
| `spec.availabilityZones` | `[]string` |  |  |  |
| `spec.networkType` | `string` |  |  |  |
| `spec.port` | `int32` |  |  |  |
| `spec.engine` | `string` | yes |  |  |
| `spec.engineVersion` | `string` |  |  |  |
| `spec.engineMode` | `string` |  |  |  |
| `spec.engineLifecycleSupport` | `string` |  |  |  |
| `spec.instances` | `[]AwsRdsClusterInstance` |  |  |  |
| `spec.instances[].name` | `string` | yes |  |  |
| `spec.instances[].instanceClass` | `string` | yes |  |  |
| `spec.instances[].promotionTier` | `int32` |  |  |  |
| `spec.instances[].availabilityZone` | `string` |  |  |  |
| `spec.instances[].publiclyAccessible` | `bool` |  |  |  |
| `spec.instances[].dbParameterGroupName` | `string` |  |  |  |
| `spec.instances[].autoMinorVersionUpgrade` | `bool` |  | `true` |  |
| `spec.instances[].performanceInsightsEnabled` | `bool` |  |  |  |
| `spec.instances[].monitoringInterval` | `int32` |  |  |  |
| `spec.instances[].caCertIdentifier` | `string` |  |  |  |
| `spec.instances[].copyTagsToSnapshot` | `bool` |  |  |  |
| `spec.instances[].performanceInsightsKmsKeyId` | `string \| valueFrom` |  |  | AwsKmsKey (`status.outputs.key_arn`) |
| `spec.instances[].performanceInsightsRetentionPeriod` | `int32` |  |  |  |
| `spec.instances[].preferredBackupWindow` | `string` |  |  |  |
| `spec.instances[].preferredMaintenanceWindow` | `string` |  |  |  |
| `spec.instances[].monitoringRoleArn` | `string \| valueFrom` |  |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.instances[].applyImmediately` | `bool` |  |  |  |
| `spec.serverlessV2Scaling` | `AwsRdsClusterServerlessV2Scaling` |  |  |  |
| `spec.serverlessV2Scaling.minCapacity` | `double` |  |  |  |
| `spec.serverlessV2Scaling.maxCapacity` | `double` | yes |  |  |
| `spec.serverlessV2Scaling.secondsUntilAutoPause` | `int32` |  |  |  |
| `spec.serverlessV1Scaling` | `AwsRdsClusterServerlessV1Scaling` |  |  |  |
| `spec.serverlessV1Scaling.autoPause` | `bool` |  | `true` |  |
| `spec.serverlessV1Scaling.minCapacity` | `int32` |  |  |  |
| `spec.serverlessV1Scaling.maxCapacity` | `int32` |  |  |  |
| `spec.serverlessV1Scaling.secondsUntilAutoPause` | `int32` |  |  |  |
| `spec.serverlessV1Scaling.secondsBeforeTimeout` | `int32` |  |  |  |
| `spec.serverlessV1Scaling.timeoutAction` | `string` |  |  |  |
| `spec.dbClusterInstanceClass` | `string` |  |  |  |
| `spec.allocatedStorageGb` | `int32` |  |  |  |
| `spec.iops` | `int32` |  |  |  |
| `spec.storageType` | `string` |  |  |  |
| `spec.databaseName` | `string` |  |  |  |
| `spec.masterUsername` | `string` |  |  |  |
| `spec.manageMasterUserPassword` | `bool` |  | `true` |  |
| `spec.masterUserSecretKmsKeyId` | `string \| valueFrom` |  |  | AwsKmsKey (`status.outputs.key_arn`) |
| `spec.masterPassword` | `string` (sensitive) |  |  |  |
| `spec.storageEncrypted` | `bool` |  | `true` |  |
| `spec.kmsKeyId` | `string \| valueFrom` |  |  | AwsKmsKey (`status.outputs.key_arn`) |
| `spec.backupRetentionPeriod` | `int32` |  |  |  |
| `spec.preferredBackupWindow` | `string` |  |  |  |
| `spec.preferredMaintenanceWindow` | `string` |  |  |  |
| `spec.copyTagsToSnapshot` | `bool` |  |  |  |
| `spec.deleteAutomatedBackups` | `bool` |  | `true` |  |
| `spec.skipFinalSnapshot` | `bool` |  |  |  |
| `spec.finalSnapshotIdentifier` | `string` |  |  |  |
| `spec.deletionProtection` | `bool` |  |  |  |
| `spec.backtrackWindowSeconds` | `int32` |  |  |  |
| `spec.iamDatabaseAuthenticationEnabled` | `bool` |  |  |  |
| `spec.iamRoles` | `[]AwsRdsClusterIamRole` |  |  |  |
| `spec.iamRoles[].role` | `string \| valueFrom` | yes |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.iamRoles[].featureName` | `string` |  |  |  |
| `spec.enableHttpEndpoint` | `bool` |  |  |  |
| `spec.enabledCloudwatchLogsExports` | `[]string` |  |  |  |
| `spec.performanceInsightsEnabled` | `bool` |  |  |  |
| `spec.performanceInsightsKmsKeyId` | `string \| valueFrom` |  |  | AwsKmsKey (`status.outputs.key_arn`) |
| `spec.performanceInsightsRetentionPeriod` | `int32` |  |  |  |
| `spec.monitoringInterval` | `int32` |  |  |  |
| `spec.monitoringRoleArn` | `string \| valueFrom` |  |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.databaseInsightsMode` | `string` |  |  |  |
| `spec.snapshotIdentifier` | `string` |  |  |  |
| `spec.restoreToPointInTime` | `AwsRdsClusterRestoreToPointInTime` |  |  |  |
| `spec.restoreToPointInTime.sourceClusterIdentifier` | `string` |  |  |  |
| `spec.restoreToPointInTime.sourceClusterResourceId` | `string` |  |  |  |
| `spec.restoreToPointInTime.restoreToTime` | `string` |  |  |  |
| `spec.restoreToPointInTime.useLatestRestorableTime` | `bool` |  |  |  |
| `spec.restoreToPointInTime.restoreType` | `string` |  |  |  |
| `spec.replicationSourceIdentifier` | `string` |  |  |  |
| `spec.sourceRegion` | `string` |  |  |  |
| `spec.globalClusterIdentifier` | `string` |  |  |  |
| `spec.enableGlobalWriteForwarding` | `bool` |  |  |  |
| `spec.enableLocalWriteForwarding` | `bool` |  |  |  |
| `spec.dbClusterParameterGroupName` | `string` |  |  |  |
| `spec.parameters` | `[]AwsRdsClusterParameter` |  |  |  |
| `spec.parameters[].name` | `string` | yes |  |  |
| `spec.parameters[].value` | `string` | yes |  |  |
| `spec.parameters[].applyMethod` | `string` |  |  |  |
| `spec.dbInstanceParameterGroupName` | `string` |  |  |  |
| `spec.caCertificateIdentifier` | `string` |  |  |  |
| `spec.applyImmediately` | `bool` |  |  |  |
| `spec.allowMajorVersionUpgrade` | `bool` |  |  |  |
| `spec.domain` | `string` |  |  |  |
| `spec.domainIamRoleName` | `string` |  |  |  |
| `spec.autoMinorVersionUpgrade` | `bool` |  | `true` |  |
| `spec.s3Import` | `AwsRdsClusterS3Import` |  |  |  |
| `spec.s3Import.bucketName` | `string` | yes |  |  |
| `spec.s3Import.bucketPrefix` | `string` |  |  |  |
| `spec.s3Import.ingestionRole` | `string \| valueFrom` | yes |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.s3Import.sourceEngine` | `string` | yes |  |  |
| `spec.s3Import.sourceEngineVersion` | `string` | yes |  |  |
| `spec.customEndpoints` | `[]AwsRdsClusterCustomEndpoint` |  |  |  |
| `spec.customEndpoints[].name` | `string` | yes |  |  |
| `spec.customEndpoints[].type` | `string` | yes |  |  |
| `spec.customEndpoints[].staticMembers` | `[]string` |  |  |  |
| `spec.customEndpoints[].excludedMembers` | `[]string` |  |  |  |
| `spec.activityStream` | `AwsRdsClusterActivityStream` |  |  |  |
| `spec.activityStream.mode` | `string` | yes |  |  |
| `spec.activityStream.kmsKeyId` | `string \| valueFrom` | yes |  | AwsKmsKey (`status.outputs.key_arn`) |
| `spec.activityStream.engineNativeAuditFieldsIncluded` | `bool` |  |  |  |

## Field Details

### spec.region

`string` · required

The AWS region the cluster is created in. Must match the region of
the subnets, security groups, and KMS keys it references.
Example: "us-west-2", "eu-west-1".

- rule: {"string":{"minLen":"1"}}

### spec.subnetIds

`[]string | valueFrom`

Subnets for the cluster's DB subnet group. Provide at least two
subnets in DISTINCT availability zones -- AWS rejects a subnet group
that covers fewer than two AZs. Reference AwsSubnet subnet_id outputs
or pass literal subnet IDs. The module manages the subnet group
itself (pure glue: a named list of subnets); alternatively point
db_subnet_group_name at an existing group.

- references: AwsSubnet (`status.outputs.subnet_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSubnet, name: <that resource's name>, fieldPath: status.outputs.subnet_id}} -- a bare string does not parse

### spec.dbSubnetGroupName

`string`

Name of an existing DB subnet group to place the cluster in, instead
of providing subnet_ids. Changing the subnet group replaces the
cluster.

### spec.securityGroupIds

`[]string | valueFrom`

Security groups attached to the cluster. Empty uses the VPC's
default security group (the AWS default). Reference AwsSecurityGroup
security_group_id outputs or pass literal SG IDs -- database ingress
rules belong on the referenced AwsSecurityGroup node, never inside
this cluster.

- references: AwsSecurityGroup (`status.outputs.security_group_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSecurityGroup, name: <that resource's name>, fieldPath: status.outputs.security_group_id}} -- a bare string does not parse

### spec.availabilityZones

`[]string`

Availability zones Aurora replicates storage across. Leave empty and
AWS picks three zones automatically -- the right call almost always,
because this list is create-time-only and a later change replaces the
cluster. Not applicable to Multi-AZ RDS clusters (AWS chooses).

### spec.networkType

`string`

The network stack of the cluster: "IPV4" (AWS default when unset) or
"DUAL" for dual-stack IPv4+IPv6. Requires subnets with IPv6 CIDRs
for "DUAL".

### spec.port

`int32`

The port the cluster accepts connections on. 0 keeps the engine
default (3306 for MySQL-family, 5432 for PostgreSQL-family).

- rule: {"int32":{"lte":65535,"gte":0}}

### spec.engine

`string` · required

The database engine. Aurora engines ("aurora-mysql",
"aurora-postgresql") use shared cluster storage with `instances`
compute; community engines ("mysql", "postgres") create a Multi-AZ
RDS cluster and require db_cluster_instance_class +
allocated_storage_gb + iops. Changing the engine replaces the
cluster.

- rule: {"required":true,"string":{"in":["aurora-mysql","aurora-postgresql","mysql","postgres"]}}

### spec.engineVersion

`string`

The engine version, e.g. "8.0.mysql_aurora.3.08.0" (Aurora MySQL)
or "16.4" (Aurora PostgreSQL). Leave empty to let AWS pick the
engine's current default version -- an empty pin never goes stale.
Minor upgrades apply in place; major upgrades additionally need
allow_major_version_upgrade.

### spec.engineMode

`string`

How the engine provisions compute. Empty or "provisioned" (the AWS
default) covers both classic provisioned instances AND Aurora
Serverless v2 (Serverless v2 is provisioned mode + a
serverless_v2_scaling block + "db.serverless" instances).
"serverless" selects the legacy Aurora Serverless v1 engine mode,
where AWS owns the compute and serverless_v1_scaling applies.
Changing the mode replaces the cluster.

### spec.engineLifecycleSupport

`string`

Extended support posture when the engine version leaves standard
support: "open-source-rds-extended-support" (AWS default -- paid
extended support kicks in automatically) or
"open-source-rds-extended-support-disabled" (the cluster must be
upgraded before end of standard support; opts out of the extra
cost).

### spec.instances

`[]AwsRdsClusterInstance`

The DB instances that serve this cluster's queries -- one writer
(lowest promotion tier) plus any number of readers. Each entry is
managed as its own provider resource keyed by `name`, so scaling
readers in and out never touches the cluster. Empty is only valid
for the two shapes where AWS owns the compute: Aurora Serverless v1
(engine_mode "serverless") and Multi-AZ RDS clusters
(db_cluster_instance_class set) -- an Aurora provisioned cluster
with no instances stores data but cannot serve a single query.

- rule: monitoring_interval must be 0 (disabled), 1, 5, 10, 15, 30, or 60 seconds
- rule: performance_insights_retention_period must be 7, 731, or a multiple of 31 (month granularity) when set

### spec.instances[].name

`string` · required

Instance name, unique within the cluster. Becomes part of the AWS
instance identifier and the key both IaC engines manage the
provider resource by -- renaming an entry replaces that instance
(the others are untouched). Required.

- rule: {"required":true,"string":{"pattern":"^[a-z][a-z0-9-]*$"}}

### spec.instances[].instanceClass

`string` · required

The instance class: a provisioned class like "db.r6g.large", or
"db.serverless" for an Aurora Serverless v2 instance that scales
within the cluster's serverless_v2_scaling bounds. Required.

- rule: {"required":true,"string":{"pattern":"^db\\."}}

### spec.instances[].promotionTier

`int32`

Failover priority, 0-15 (lower is promoted first). Give the
largest readers tier 0/1 so a failover lands on capacity that can
absorb the write load. 0 is the AWS default.

- rule: {"int32":{"lte":15,"gte":0}}

### spec.instances[].availabilityZone

`string`

Pin the instance to one availability zone. Empty lets AWS place it
-- preferred, since AWS spreads instances across the cluster's
zones automatically. Create-time only.

### spec.instances[].publiclyAccessible

`bool`

Give the instance a public IP. Requires public subnets; keep false
for anything production-shaped.

### spec.instances[].dbParameterGroupName

`string`

The DB (instance-level) parameter group for this instance. Empty
keeps the engine default group.

### spec.instances[].autoMinorVersionUpgrade

`bool` · optional (explicit presence)

Apply minor engine version patches automatically during the
maintenance window. AWS defaults this to true; disable only when
patch timing must be controlled manually.

- default: `true`

### spec.instances[].performanceInsightsEnabled

`bool` · optional (explicit presence)

Per-instance Performance Insights override. Unset inherits the
cluster-level setting.

### spec.instances[].monitoringInterval

`int32`

Enhanced Monitoring granularity for this instance in seconds: 1,
5, 10, 15, 30, or 60. 0 disables. Uses the cluster's
monitoring_role_arn.

### spec.instances[].caCertIdentifier

`string`

The CA certificate bundle for this instance (e.g.
"rds-ca-rsa2048-g1"). Empty keeps the AWS default.

### spec.instances[].copyTagsToSnapshot

`bool`

Copy this instance's tags onto its snapshots.

### spec.instances[].performanceInsightsKmsKeyId

`string | valueFrom`

The KMS key encrypting this instance's Performance Insights data.
Empty uses the AWS default. Reference an AwsKmsKey key_arn output
or pass a literal key ARN. Cannot change after Performance
Insights is first enabled on the instance.

- references: AwsKmsKey (`status.outputs.key_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_arn}} -- a bare string does not parse

### spec.instances[].performanceInsightsRetentionPeriod

`int32`

Days of Performance Insights history for THIS instance: 7 (free
tier), 731 (2 years), or any multiple of 31 in between. 0 inherits
the cluster-level setting (or the AWS default of 7).

### spec.instances[].preferredBackupWindow

`string`

The daily backup window for THIS instance in UTC, format
"hh24:mi-hh24:mi". Empty inherits the cluster's window -- stagger
per-instance windows when backups must not contend.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^([01][0-9]|2[0-3]):[0-5][0-9]-([01][0-9]|2[0-3]):[0-5][0-9]$"}}

### spec.instances[].preferredMaintenanceWindow

`string`

The weekly maintenance window for THIS instance in UTC, format
"ddd:hh24:mi-ddd:hh24:mi". Empty inherits scheduling from AWS --
stagger per-instance windows so readers never patch
simultaneously.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^(mon|tue|wed|thu|fri|sat|sun):([01][0-9]|2[0-3]):[0-5][0-9]-(mon|tue|wed|thu|fri|sat|sun):([01][0-9]|2[0-3]):[0-5][0-9]$"}}

### spec.instances[].monitoringRoleArn

`string | valueFrom`

The IAM role Enhanced Monitoring publishes through for THIS
instance (needs the AmazonRDSEnhancedMonitoringRole managed
policy). Empty uses the cluster's monitoring_role_arn. Reference
an AwsIamRole role_arn output or pass a literal ARN.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.instances[].applyImmediately

`bool`

Apply modifications to THIS instance immediately instead of
waiting for its next maintenance window. AWS defaults to deferred.

### spec.serverlessV2Scaling

`AwsRdsClusterServerlessV2Scaling`

Aurora Serverless v2 capacity bounds. Requires provisioned engine
mode and instances of class "db.serverless" -- each such instance
scales independently within these bounds. min_capacity 0 enables
automatic pause (scale-to-zero) after seconds_until_auto_pause of
idleness.

- rule: max_capacity must be greater than or equal to min_capacity
- rule: seconds_until_auto_pause must be between 300 (5 minutes) and 86400 (1 day) when set
- rule: seconds_until_auto_pause only applies when min_capacity is 0 (auto-pause enabled)

### spec.serverlessV2Scaling.minCapacity

`double`

Minimum ACUs, 0-256 in 0.5 steps. 0 enables automatic pause: the
instance suspends after seconds_until_auto_pause of idleness and
costs nothing while paused (storage still billed) -- resumed on the
next connection in ~15 seconds. Dev/test clusters want 0; latency-
sensitive production wants >= 0.5 to never pause.

- rule: {"double":{"lte":256,"gte":0}}

### spec.serverlessV2Scaling.maxCapacity

`double` · required

Maximum ACUs, 1-256. The hard spend/performance ceiling per
instance. Required.

- rule: {"required":true,"double":{"lte":256,"gte":1}}

### spec.serverlessV2Scaling.secondsUntilAutoPause

`int32`

Seconds of idleness before an instance auto-pauses, 300-86400.
Only meaningful when min_capacity is 0. 0 keeps the AWS default
(300 -- five minutes).

### spec.serverlessV1Scaling

`AwsRdsClusterServerlessV1Scaling`

Aurora Serverless v1 autoscaling configuration. Only valid with
engine_mode "serverless" (the legacy serverless offering) -- prefer
Serverless v2 for new designs.

- rule: max_capacity must be greater than or equal to min_capacity when both are set
- rule: seconds_until_auto_pause must be between 300 and 86400 when set
- rule: seconds_before_timeout must be between 60 and 600 when set
- rule: timeout_action must be 'RollbackCapacityChange' or 'ForceApplyCapacityChange' when set

### spec.serverlessV1Scaling.autoPause

`bool` · optional (explicit presence)

Pause compute after seconds_until_auto_pause of idleness. AWS
defaults this to true -- the cost model Serverless v1 exists for.

- default: `true`

### spec.serverlessV1Scaling.minCapacity

`int32`

Minimum ACUs (whole units, engine-specific valid set). 0 keeps the
AWS default (1).

### spec.serverlessV1Scaling.maxCapacity

`int32`

Maximum ACUs (whole units). 0 keeps the AWS default (16).

### spec.serverlessV1Scaling.secondsUntilAutoPause

`int32`

Seconds of idleness before pausing, 300-86400. 0 keeps the AWS
default (300).

### spec.serverlessV1Scaling.secondsBeforeTimeout

`int32`

Seconds a scaling operation waits for a safe scaling point before
timeout_action applies, 60-600. 0 keeps the AWS default (300).

### spec.serverlessV1Scaling.timeoutAction

`string`

What happens when scaling times out: "RollbackCapacityChange" (AWS
default -- keep current capacity) or "ForceApplyCapacityChange"
(scale anyway, dropping connections that block it).

### spec.dbClusterInstanceClass

`string`

The instance class for a Multi-AZ RDS cluster (community
mysql/postgres engines), e.g. "db.m6gd.large". Setting it selects
the Multi-AZ cluster shape: AWS manages one writer and two readers
internally, `instances` stays empty, and allocated_storage_gb + iops
are required.

### spec.allocatedStorageGb

`int32`

Provisioned storage in GiB for a Multi-AZ RDS cluster. Not
applicable to Aurora (Aurora storage grows automatically).

- rule: {"int32":{"gte":0}}

### spec.iops

`int32`

Provisioned IOPS for a Multi-AZ RDS cluster (required by AWS for
io1/io2/gp3 cluster storage). Not applicable to Aurora.

- rule: {"int32":{"gte":0}}

### spec.storageType

`string`

The storage type. Aurora engines: "" (standard, billed per I/O) or
"aurora-iopt1" (I/O-Optimized, up to ~40% cheaper for I/O-heavy
workloads, switchable once per 30 days). Multi-AZ RDS clusters:
"io1", "io2", or "gp3".

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["aurora","aurora-iopt1","io1","io2","gp3"]}}

### spec.databaseName

`string`

The name of the initial database AWS creates in the cluster. Empty
creates no database (create one from SQL later). Create-time only.

### spec.masterUsername

`string`

The master username. Required for a brand-new cluster -- AWS has no
default and rejects a blank value at CreateDBCluster. Only clusters
that inherit credentials from a source (snapshot restore,
point-in-time restore, a replication source, or joining an existing
global database as a secondary) leave it empty. Avoid the engine's
reserved names (e.g. "rdsadmin"). Create-time only -- changing it
replaces the cluster.

### spec.manageMasterUserPassword

`bool`

Let AWS manage the master password in Secrets Manager: AWS
generates it, stores it, rotates it on schedule, and no secret ever
touches this manifest or the IaC state. The managed secret's ARN is
exported as the master_user_secret_arn output. Mutually exclusive
with master_password -- and the recommended posture.

- default: `true`

### spec.masterUserSecretKmsKeyId

`string | valueFrom`

The KMS key that encrypts the AWS-managed master-user secret (only
meaningful with manage_master_user_password). Empty uses the
account's default aws/secretsmanager key. Reference an AwsKmsKey
key_arn output or pass a literal key ARN.

- references: AwsKmsKey (`status.outputs.key_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_arn}} -- a bare string does not parse

### spec.masterPassword

`string` · sensitive

The master password, supplied directly. Stored in IaC state --
prefer manage_master_user_password, which keeps the secret in
Secrets Manager entirely. Mutually exclusive with
manage_master_user_password.

### spec.storageEncrypted

`bool`

Encrypt cluster storage at rest. Strongly recommended -- and
create-time only: an unencrypted cluster cannot be encrypted later
(requires a snapshot-restore migration).

- default: `true`

### spec.kmsKeyId

`string | valueFrom`

The KMS key for storage encryption when storage_encrypted is true.
Empty uses the AWS-managed aws/rds key. Reference an AwsKmsKey
key_arn output or pass a literal key ARN. Create-time only.

- references: AwsKmsKey (`status.outputs.key_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_arn}} -- a bare string does not parse

### spec.backupRetentionPeriod

`int32`

Days automated backups are retained, 1-35. 0 keeps the AWS default
(1 day). Aurora backups are continuous -- this window bounds
point-in-time recovery, so production clusters typically want 7+.

- rule: {"int32":{"lte":35,"gte":0}}

### spec.preferredBackupWindow

`string`

The daily backup window in UTC, format "hh24:mi-hh24:mi" (e.g.
"04:00-05:00"). Empty lets AWS assign one. Must not overlap the
maintenance window.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^([01][0-9]|2[0-3]):[0-5][0-9]-([01][0-9]|2[0-3]):[0-5][0-9]$"}}

### spec.preferredMaintenanceWindow

`string`

The weekly maintenance window in UTC, format
"ddd:hh24:mi-ddd:hh24:mi" (e.g. "sun:05:00-sun:06:00"). Empty lets
AWS assign one.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^(mon|tue|wed|thu|fri|sat|sun):([01][0-9]|2[0-3]):[0-5][0-9]-(mon|tue|wed|thu|fri|sat|sun):([01][0-9]|2[0-3]):[0-5][0-9]$"}}

### spec.copyTagsToSnapshot

`bool`

Copy the cluster's tags onto automated and manual snapshots.

### spec.deleteAutomatedBackups

`bool` · optional (explicit presence)

Remove automated backups immediately when the cluster is deleted.
AWS defaults this to true; set false to retain the backups for the
remainder of their retention window after deletion -- the last line
of defense against a mistaken teardown.

- default: `true`

### spec.skipFinalSnapshot

`bool`

Skip the final snapshot when the cluster is deleted. When false
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

Refuse deletion of the cluster while enabled. Turn this on for
anything holding data you cannot recreate -- deletion then requires
an explicit two-step (disable, delete).

### spec.backtrackWindowSeconds

`int32`

Aurora MySQL backtrack window in seconds, 0-259200 (72 hours).
Backtrack rewinds the cluster in place (no restore, no new
endpoint) -- the fastest "undo" for fat-fingered writes. 0 disables.
Aurora MySQL only; enabling on an existing cluster is not supported
by AWS.

- rule: {"int32":{"lte":259200,"gte":0}}

### spec.iamDatabaseAuthenticationEnabled

`bool`

Map IAM identities to database users -- connect with short-lived
IAM auth tokens instead of passwords.

### spec.iamRoles

`[]AwsRdsClusterIamRole`

IAM roles the cluster assumes for engine features that reach into
other AWS services (S3 import/export, Lambda invocation, Comprehend
/SageMaker for ML functions). Each entry associates one role,
optionally linked to a specific engine feature by feature_name --
the roles own their policies; this cluster only associates them.
Both IaC modules manage each entry as its own role-association
resource (never the cluster's inline role list, which cannot carry
feature names and conflicts with association resources), so roles
attach and detach without touching the cluster.

### spec.iamRoles[].role

`string | valueFrom` · required

The IAM role to associate. Reference an AwsIamRole role_arn output
or pass a literal role ARN. The role owns its policies; the
cluster only assumes it. The role's trust policy MUST allow
rds.amazonaws.com to assume it -- AWS validates that server-side at
association time and rejects the call with InvalidParameterValue
("IAM role ARN value is invalid or does not include the required
permissions") otherwise; no plan-time check catches it. Required.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.iamRoles[].featureName

`string`

The engine feature the role is linked to (e.g. "s3Import",
"s3Export", "Lambda", "SageMaker", "Comprehend"). Empty associates
the role without a feature link; AWS requires the name whenever
the role powers a specific engine capability. Changing it replaces
the association (the cluster is untouched).

### spec.enableHttpEndpoint

`bool`

Enable the RDS Data API: SQL over HTTPS with IAM auth, no
persistent connections -- the natural fit for Lambda and other
connection-averse callers. Aurora PostgreSQL and Aurora MySQL
(Serverless v2 / provisioned), plus Serverless v1.

### spec.enabledCloudwatchLogsExports

`[]string`

Database log types to export to CloudWatch Logs. MySQL family
("aurora-mysql", "mysql"): "audit", "error", "general", "slowquery".
PostgreSQL family ("aurora-postgresql", "postgres"): "postgresql",
"upgrade". Both families also accept "iam-db-auth-error", and
Multi-AZ RDS clusters accept "instance".

- rule: {"repeated":{"unique":true}}

### spec.performanceInsightsEnabled

`bool`

Cluster-level Performance Insights (per-query performance
telemetry). On Aurora, per-instance settings in `instances` can
override this. Free at the default 7-day retention.

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

Days of Performance Insights history: 7 (free tier), 731 (2 years),
or any multiple of 31 in between. 0 keeps the AWS default (7).

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

### spec.snapshotIdentifier

`string`

Restore the cluster from an existing cluster snapshot (name or ARN)
at create time. Create-time only; mutually exclusive with
restore_to_point_in_time.

### spec.restoreToPointInTime

`AwsRdsClusterRestoreToPointInTime`

Clone or restore this cluster from another cluster's continuous
backup at create time -- point-in-time recovery as a first-class
create shape. Create-time only; mutually exclusive with
snapshot_identifier.

- rule: exactly one of source_cluster_identifier or source_cluster_resource_id must be set
- rule: exactly one of restore_to_time or use_latest_restorable_time must be set
- rule: restore_type must be 'full-copy' or 'copy-on-write' when set

### spec.restoreToPointInTime.sourceClusterIdentifier

`string`

The source cluster identifier (name or ARN). Exactly one of
source_cluster_identifier or source_cluster_resource_id must be
set.

### spec.restoreToPointInTime.sourceClusterResourceId

`string`

The source cluster's immutable resource ID (cluster_resource_id
output) -- survives identifier renames and points at deleted
clusters' retained backups. Exactly one of the two source fields
must be set.

### spec.restoreToPointInTime.restoreToTime

`string`

The UTC timestamp to restore to, RFC3339 (e.g.
"2026-07-01T09:45:00Z"). Mutually exclusive with
use_latest_restorable_time.

### spec.restoreToPointInTime.useLatestRestorableTime

`bool`

Restore to the most recent recoverable moment. Mutually exclusive
with restore_to_time.

### spec.restoreToPointInTime.restoreType

`string`

"full-copy" (independent storage, AWS default) or "copy-on-write"
(an Aurora fast clone -- shares storage with the source and only
pays for divergence; ideal for prod-data staging environments).

### spec.replicationSourceIdentifier

`string`

Make this cluster a cross-region (or cross-account) read replica of
the given source cluster ARN. Promote by clearing the field.

### spec.sourceRegion

`string`

The region of the replication source, required by AWS when creating
an encrypted cross-region replica (it scopes the KMS re-encryption).
Create-time only.

### spec.globalClusterIdentifier

`string`

Join an Aurora Global Database: the identifier of the
aws_rds_global_cluster this cluster participates in. The first
cluster joined becomes the global writer; clusters added afterwards
become read-only secondaries.

### spec.enableGlobalWriteForwarding

`bool`

Let secondary-region endpoints accept writes and forward them to
the global writer (Aurora Global Database only). Apps in secondary
regions get a single connection string for reads AND writes.

### spec.enableLocalWriteForwarding

`bool`

Let reader instances in THIS cluster accept writes and forward them
to the writer -- one endpoint for the whole cluster without a
client-side split. Aurora MySQL 3.04+ / Aurora PostgreSQL 16.4+.

### spec.dbClusterParameterGroupName

`string`

The name of an existing cluster parameter group to use. Mutually
exclusive with `parameters` -- either bring your own group or let
the module manage one from inline parameters.

### spec.parameters

`[]AwsRdsClusterParameter`

Cluster-level engine parameters, managed as a dedicated parameter
group owned by this cluster (the group is glue -- a named parameter
list -- so it stays folded). Mutually exclusive with
db_cluster_parameter_group_name.

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

### spec.dbInstanceParameterGroupName

`string`

The name of the DB (instance-level) parameter group applied to
cluster instances DURING a major engine version upgrade. Only
consulted when engine_version changes across a major version.

### spec.caCertificateIdentifier

`string`

The CA certificate bundle for the cluster's instances (e.g.
"rds-ca-rsa2048-g1"). Empty keeps the AWS default bundle.

### spec.applyImmediately

`bool`

Apply modifications immediately instead of waiting for the next
maintenance window. Immediate changes can interrupt connections
(e.g. scaling); deferred changes wait quietly. AWS defaults to
deferred.

### spec.allowMajorVersionUpgrade

`bool`

Permit engine_version changes that cross a major version. Off (the
default) guards against an accidental major upgrade hidden in a
version bump.

### spec.domain

`string`

Join the cluster to an AWS Managed Microsoft AD directory (d-...)
for Kerberos authentication (Aurora MySQL and Aurora PostgreSQL).
Pairs with domain_iam_role_name. (Self-managed AD is an
instance-kind shape -- clusters only support the managed
directory.)

### spec.domainIamRoleName

`string`

The name of the IAM role RDS uses to join the managed directory
(needs the AmazonRDSDirectoryServiceAccess managed policy).
Required with domain.

### spec.autoMinorVersionUpgrade

`bool` · optional (explicit presence)

Apply minor engine version patches to the cluster automatically
during the maintenance window. AWS defaults this to true; disable
only when patch timing must be controlled manually. Per-instance
auto_minor_version_upgrade in `instances` overrides this for that
instance.

- default: `true`

### spec.s3Import

`AwsRdsClusterS3Import`

Create the cluster by restoring a Percona XtraBackup stored in S3
-- the on-ramp for migrating a self-managed MySQL database into
Aurora MySQL without a logical dump. Create-time only; mutually
exclusive with snapshot_identifier and restore_to_point_in_time.

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
cluster S3 restores. Required.

- rule: {"required":true,"string":{"in":["mysql"]}}

### spec.s3Import.sourceEngineVersion

`string` · required

The version of the source engine the backup was taken from (e.g.
"8.0"). Required.

- rule: {"required":true}

### spec.customEndpoints

`[]AwsRdsClusterCustomEndpoint`

Custom cluster endpoints -- stable DNS names scoped to a chosen
subset of the cluster's instances (e.g. an analytics endpoint over
the big readers). Each entry is managed as its own provider
resource keyed by `name`, so endpoints come and go without
touching the cluster.

- rule: static_members and excluded_members are mutually exclusive -- pin the member set or subtract from it, not both

### spec.customEndpoints[].name

`string` · required

Endpoint name, unique within the cluster (lowercase letters,
digits, hyphens; starts with a letter). Becomes the DNS-visible
endpoint identifier and the key both IaC engines manage the
provider resource by -- renaming an entry replaces that endpoint
(the cluster is untouched). Required.

- rule: {"required":true,"string":{"pattern":"^[a-z][a-z0-9-]*$"}}

### spec.customEndpoints[].type

`string` · required

Which instances the endpoint fronts: "READER" (reader instances
only) or "ANY" (all instances). Required.

- rule: {"required":true,"string":{"in":["READER","ANY"]}}

### spec.customEndpoints[].staticMembers

`[]string`

Pin the endpoint to exactly these instances, by their
spec.instances entry names. Mutually exclusive with
excluded_members. Both empty fronts every instance of the
endpoint's type.

### spec.customEndpoints[].excludedMembers

`[]string`

Front every instance of the endpoint's type EXCEPT these, by their
spec.instances entry names. Mutually exclusive with
static_members.

### spec.activityStream

`AwsRdsClusterActivityStream`

Stream every audited database event to a dedicated Kinesis stream
(a Database Activity Stream), encrypted with the given KMS key --
the compliance-grade audit feed consumed by GuardDuty RDS
Protection and partner SIEMs. Aurora engines only. The stream's
Kinesis name is exported as the activity_stream_kinesis_stream_name
output.

### spec.activityStream.mode

`string` · required

Delivery mode: "sync" (the database blocks until the event is
durably recorded -- audit-grade guarantees at a latency cost) or
"async" (the database never waits; rare event loss is possible).
Required.

- rule: {"required":true,"string":{"in":["sync","async"]}}

### spec.activityStream.kmsKeyId

`string | valueFrom` · required

The KMS key that encrypts the activity stream. Reference an
AwsKmsKey key_arn output or pass a literal key ARN. Required.

- references: AwsKmsKey (`status.outputs.key_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_arn}} -- a bare string does not parse

### spec.activityStream.engineNativeAuditFieldsIncluded

`bool`

Also capture the engine's native audit fields in the stream
(engine-version dependent; leave false unless the consumer needs
them).

## Validation Rules

- `subnets_or_group`: provide at least two subnet_ids (distinct AZs) or an existing db_subnet_group_name
- `password_xor_managed`: master_password cannot be set when manage_master_user_password is true -- pick one password strategy
- `master_username_required_unless_derived`: master_username is required for a new cluster -- AWS rejects a blank username; only snapshot/point-in-time restores, replicas, and global-database secondaries inherit credentials from their source
- `final_snapshot_id_required_when_not_skipping`: final_snapshot_identifier is required when skip_final_snapshot is false -- AWS refuses to delete the cluster without a final snapshot name
- `engine_mode_valid`: engine_mode must be 'provisioned' or 'serverless' when set (Serverless v2 uses provisioned mode + serverless_v2_scaling)
- `serverless_mode_is_aurora_only`: engine_mode 'serverless' (Aurora Serverless v1) requires an Aurora engine
- `v1_scaling_requires_serverless_mode`: serverless_v1_scaling only applies when engine_mode is 'serverless'
- `v2_scaling_requires_provisioned_mode`: serverless_v2_scaling requires provisioned engine mode (Serverless v2 is provisioned mode with db.serverless instances), not engine_mode 'serverless'
- `instances_not_with_serverless_v1`: instances cannot be set with engine_mode 'serverless' -- Aurora Serverless v1 manages compute itself
- `instances_not_with_multi_az_cluster`: instances cannot be set with db_cluster_instance_class -- a Multi-AZ RDS cluster manages its writer and readers itself
- `multi_az_cluster_is_community_engine`: db_cluster_instance_class (Multi-AZ RDS cluster) requires engine 'mysql' or 'postgres' -- Aurora clusters size compute through instances
- `community_engine_is_multi_az_cluster`: engine 'mysql' or 'postgres' creates a Multi-AZ RDS cluster and requires db_cluster_instance_class, allocated_storage_gb, and iops
- `aurora_storage_types`: Aurora engines use storage_type '' (standard) or 'aurora-iopt1'; io1/io2/gp3 are Multi-AZ RDS cluster storage types
- `multi_az_cluster_storage_types`: Multi-AZ RDS clusters (mysql/postgres) use storage_type 'io1', 'io2', or 'gp3'
- `backtrack_is_aurora_mysql_only`: backtrack_window_seconds is an Aurora MySQL feature
- `log_exports_match_engine_family`: MySQL-family log types are audit/error/general/slowquery (plus iam-db-auth-error and, on Multi-AZ RDS clusters, instance); PostgreSQL-family are postgresql/upgrade (plus iam-db-auth-error and instance)
- `pi_retention_valid`: performance_insights_retention_period must be 7, 731, or a multiple of 31 (month granularity) when set
- `monitoring_interval_valid`: monitoring_interval must be 0 (disabled), 1, 5, 10, 15, 30, or 60 seconds
- `monitoring_role_required_with_interval`: monitoring_role_arn is required when monitoring_interval is set -- Enhanced Monitoring publishes through that role
- `database_insights_mode_valid`: database_insights_mode must be 'standard' or 'advanced' when set
- `one_create_source`: snapshot_identifier, restore_to_point_in_time, and s3_import are mutually exclusive create-time sources
- `s3_import_is_aurora_mysql`: s3_import (Percona XtraBackup restore) is an Aurora MySQL feature -- engine must be 'aurora-mysql'
- `domain_pair_together`: domain and domain_iam_role_name are required together -- the managed-directory join needs both
- `activity_stream_is_aurora_only`: activity_stream (Database Activity Streams) is an Aurora feature -- Multi-AZ RDS clusters (mysql/postgres) do not support it
- `own_parameters_xor_existing_group`: parameters and db_cluster_parameter_group_name are mutually exclusive -- manage parameters here or bring an existing group
- `parameters_require_engine_version`: inline parameters require a pinned engine_version -- the managed parameter group's family is derived from it and must match the running engine
- `network_type_valid`: network_type must be 'IPV4' or 'DUAL' when set
- `engine_lifecycle_support_valid`: engine_lifecycle_support must be 'open-source-rds-extended-support' or 'open-source-rds-extended-support-disabled' when set

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsRdsCluster, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.cluster_identifier` | `string` | The cluster identifier (e.g. "orders-db"). |
| `status.outputs.arn` | `string` | The Amazon Resource Name of the cluster. |
| `status.outputs.cluster_resource_id` | `string` | The immutable cluster resource ID (cluster-...). Survives identifier renames -- the durable handle for point-in-time restores and CloudWatch dimensions. |
| `status.outputs.endpoint` | `string` | The writer endpoint -- connect here for reads and writes. |
| `status.outputs.reader_endpoint` | `string` | The reader endpoint -- load-balances connections across the cluster's reader instances. |
| `status.outputs.port` | `int32` | The port the cluster accepts connections on. |
| `status.outputs.hosted_zone_id` | `string` | The Route53 hosted zone ID of the cluster endpoints, for DNS alias records. |
| `status.outputs.engine_version_actual` | `string` | The resolved engine version actually running (meaningful when the spec leaves engine_version to the AWS default). |
| `status.outputs.master_user_secret_arn` | `string` | The ARN of the AWS-managed master-user secret in Secrets Manager. Populated only when manage_master_user_password is true -- the handle applications use to fetch credentials at runtime. |
| `status.outputs.db_subnet_group_name` | `string` | The name of the DB subnet group the cluster runs in. |
| `status.outputs.db_cluster_parameter_group_name` | `string` | The name of the cluster parameter group in use (module-managed or the referenced existing group). |
| `status.outputs.instance_endpoints` | `[]string` | Per-instance endpoints of the cluster's folded instances, ordered as declared in spec.instances. Empty for Aurora Serverless v1 and Multi-AZ RDS clusters, where AWS owns the compute. |
| `status.outputs.custom_endpoints` | `[]AwsRdsClusterCustomEndpointOutput` | The custom cluster endpoints declared in spec.custom_endpoints -- one entry per endpoint, carrying its DNS name for charts to wire into consumers. |
| `status.outputs.custom_endpoints[].name` | `string` | The endpoint's spec.custom_endpoints entry name. |
| `status.outputs.custom_endpoints[].endpoint` | `string` | The DNS name of the custom endpoint -- connect here to reach the endpoint's member subset. |
| `status.outputs.activity_stream_kinesis_stream_name` | `string` | The name of the Kinesis stream receiving the Database Activity Stream (aws-rds-das-...). Populated only when spec.activity_stream is set -- the handle audit consumers attach to. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.subnetIds` | AwsSubnet | `status.outputs.subnet_id` |
| `spec.securityGroupIds` | AwsSecurityGroup | `status.outputs.security_group_id` |
| `spec.instances[].performanceInsightsKmsKeyId` | AwsKmsKey | `status.outputs.key_arn` |
| `spec.instances[].monitoringRoleArn` | AwsIamRole | `status.outputs.role_arn` |
| `spec.masterUserSecretKmsKeyId` | AwsKmsKey | `status.outputs.key_arn` |
| `spec.kmsKeyId` | AwsKmsKey | `status.outputs.key_arn` |
| `spec.iamRoles[].role` | AwsIamRole | `status.outputs.role_arn` |
| `spec.performanceInsightsKmsKeyId` | AwsKmsKey | `status.outputs.key_arn` |
| `spec.monitoringRoleArn` | AwsIamRole | `status.outputs.role_arn` |
| `spec.s3Import.ingestionRole` | AwsIamRole | `status.outputs.role_arn` |
| `spec.activityStream.kmsKeyId` | AwsKmsKey | `status.outputs.key_arn` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AwsAppSyncApi | `spec.datasources[].relationalDatabase.dbClusterIdentifier` | `status.outputs.cluster_identifier` |
| AwsBedrockKnowledgeBase | `spec.storage.rds.resourceArn` | `status.outputs.arn` |
| AwsRdsProxy | `spec.target.dbClusterIdentifier` | `status.outputs.cluster_identifier` |

## See Also

- [Overview](../README.md)
