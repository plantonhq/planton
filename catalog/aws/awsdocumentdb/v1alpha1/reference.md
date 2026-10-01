# AwsDocumentDb

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

AwsDocumentDbSpec defines an Amazon DocumentDB (with MongoDB
compatibility) cluster -- a managed document database that speaks the
MongoDB 4.0/5.0 wire protocol over shared cluster storage.

The cluster is the shared-storage brain -- endpoints, credentials,
backups, encryption, and engine lifecycle live here. The compute that
serves queries is the `instances` list: each entry materializes as its
own DB instance inside the cluster (a writer plus any readers). Cluster
instances are sub-resources of exactly one cluster and are referenced by
nothing else, so they are folded into this spec rather than modeled as a
standalone kind; both IaC modules manage each entry as its own provider
resource keyed by name, so adding or removing a reader is an in-place
update, never a cluster replacement.

The cluster identifier comes from metadata.name. Security groups,
subnets, and KMS keys compose by reference -- this cluster never creates
or mutates resources that deserve to be their own nodes.

## Example

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsDocumentDb
metadata:
  name: awsdocumentdb-demo
spec:
  region: us-west-2
  subnetIds:
    - value: subnet-0a1b2c3d4e5f60001
    - value: subnet-0a1b2c3d4e5f60002
  masterUsername: hackadmin
  manageMasterUserPassword: true
  storageEncrypted: true
  skipFinalSnapshot: true
  instances:
    - name: writer
      instanceClass: db.t4g.medium
      # Defer the CA-rotation restart to the next maintenance action
      # instead of restarting immediately (unset keeps the AWS default:
      # restart on rotation).
      certificateRotationRestart: false
      applyImmediately: true
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
| `spec.engineVersion` | `string` |  |  |  |
| `spec.storageType` | `string` |  |  |  |
| `spec.instances` | `[]AwsDocumentDbInstance` |  |  |  |
| `spec.instances[].name` | `string` | yes |  |  |
| `spec.instances[].instanceClass` | `string` | yes |  |  |
| `spec.instances[].promotionTier` | `int32` |  |  |  |
| `spec.instances[].availabilityZone` | `string` |  |  |  |
| `spec.instances[].autoMinorVersionUpgrade` | `bool` |  | `true` |  |
| `spec.instances[].performanceInsightsEnabled` | `bool` |  |  |  |
| `spec.instances[].performanceInsightsKmsKeyId` | `string \| valueFrom` |  |  | AwsKmsKey (`status.outputs.key_arn`) |
| `spec.instances[].preferredMaintenanceWindow` | `string` |  |  |  |
| `spec.instances[].caCertIdentifier` | `string` |  |  |  |
| `spec.instances[].copyTagsToSnapshot` | `bool` |  |  |  |
| `spec.instances[].certificateRotationRestart` | `bool` |  | `true` |  |
| `spec.instances[].applyImmediately` | `bool` |  |  |  |
| `spec.serverlessV2Scaling` | `AwsDocumentDbServerlessV2Scaling` |  |  |  |
| `spec.serverlessV2Scaling.minCapacity` | `double` | yes |  |  |
| `spec.serverlessV2Scaling.maxCapacity` | `double` | yes |  |  |
| `spec.masterUsername` | `string` |  |  |  |
| `spec.manageMasterUserPassword` | `bool` |  | `true` |  |
| `spec.masterPassword` | `string` (sensitive) |  |  |  |
| `spec.storageEncrypted` | `bool` |  | `true` |  |
| `spec.kmsKeyId` | `string \| valueFrom` |  |  | AwsKmsKey (`status.outputs.key_arn`) |
| `spec.backupRetentionPeriod` | `int32` |  |  |  |
| `spec.preferredBackupWindow` | `string` |  |  |  |
| `spec.preferredMaintenanceWindow` | `string` |  |  |  |
| `spec.skipFinalSnapshot` | `bool` |  |  |  |
| `spec.finalSnapshotIdentifier` | `string` |  |  |  |
| `spec.deletionProtection` | `bool` |  |  |  |
| `spec.enabledCloudwatchLogsExports` | `[]string` |  |  |  |
| `spec.snapshotIdentifier` | `string` |  |  |  |
| `spec.restoreToPointInTime` | `AwsDocumentDbRestoreToPointInTime` |  |  |  |
| `spec.restoreToPointInTime.sourceClusterIdentifier` | `string` | yes |  |  |
| `spec.restoreToPointInTime.restoreToTime` | `string` |  |  |  |
| `spec.restoreToPointInTime.useLatestRestorableTime` | `bool` |  |  |  |
| `spec.restoreToPointInTime.restoreType` | `string` |  |  |  |
| `spec.globalClusterIdentifier` | `string` |  |  |  |
| `spec.dbClusterParameterGroupName` | `string` |  |  |  |
| `spec.parameters` | `[]AwsDocumentDbParameter` |  |  |  |
| `spec.parameters[].name` | `string` | yes |  |  |
| `spec.parameters[].value` | `string` | yes |  |  |
| `spec.parameters[].applyMethod` | `string` |  |  |  |
| `spec.applyImmediately` | `bool` |  |  |  |
| `spec.allowMajorVersionUpgrade` | `bool` |  |  |  |

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

Availability zones the cluster's storage is replicated across. Leave
empty and AWS picks three zones automatically -- the right call
almost always, because this list is create-time-only and a later
change replaces the cluster.

### spec.networkType

`string`

The network stack of the cluster: "IPV4" (AWS default when unset) or
"DUAL" for dual-stack IPv4+IPv6. Requires subnets with IPv6 CIDRs
for "DUAL".

### spec.port

`int32`

The port the cluster accepts connections on. 0 keeps the AWS default
(27017 -- the MongoDB convention). DocumentDB rejects ports below
1150. Create-time only -- changing the port replaces the cluster.

### spec.engineVersion

`string`

The DocumentDB engine version, e.g. "5.0.0". Leave empty to let AWS
pick the current default version -- an empty pin never goes stale.
Minor upgrades apply in place; major upgrades additionally need
allow_major_version_upgrade.

### spec.storageType

`string`

The storage I/O model: "" or "standard" (billed per I/O, the AWS
default) or "iopt1" (I/O-Optimized -- predictable pricing that pays
off for I/O-heavy workloads; switchable once per 30 days).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["standard","iopt1"]}}

### spec.instances

`[]AwsDocumentDbInstance`

The DB instances that serve this cluster's queries -- one writer
(lowest promotion tier) plus any number of readers. Each entry is
managed as its own provider resource keyed by `name`, so scaling
readers in and out never touches the cluster. Empty is only valid
for headless shapes that attach compute later: a snapshot or
point-in-time restore, or a cluster joined to a global cluster --
a regular cluster with no instances stores data but cannot serve a
single query.

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
"db.serverless" for a DocumentDB Serverless instance that scales
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

### spec.instances[].autoMinorVersionUpgrade

`bool` · optional (explicit presence)

Apply minor engine version patches automatically during the
maintenance window. AWS defaults this to true; disable only when
patch timing must be controlled manually.

- default: `true`

### spec.instances[].performanceInsightsEnabled

`bool`

Per-instance Performance Insights (per-query performance
telemetry). Free at the default 7-day retention. DocumentDB scopes
this to the instance -- there is no cluster-level setting.

### spec.instances[].performanceInsightsKmsKeyId

`string | valueFrom`

The KMS key encrypting this instance's Performance Insights data.
Empty uses the AWS default. Reference an AwsKmsKey key_arn output
or pass a literal key ARN.

- references: AwsKmsKey (`status.outputs.key_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_arn}} -- a bare string does not parse

### spec.instances[].preferredMaintenanceWindow

`string`

The weekly maintenance window for THIS instance in UTC, format
"ddd:hh24:mi-ddd:hh24:mi". Empty inherits scheduling from AWS --
stagger per-instance windows so readers never patch simultaneously.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^(mon|tue|wed|thu|fri|sat|sun):([01][0-9]|2[0-3]):[0-5][0-9]-(mon|tue|wed|thu|fri|sat|sun):([01][0-9]|2[0-3]):[0-5][0-9]$"}}

### spec.instances[].caCertIdentifier

`string`

The CA certificate bundle for this instance (e.g.
"rds-ca-rsa2048-g1"). Empty keeps the AWS default.

### spec.instances[].copyTagsToSnapshot

`bool`

Copy this instance's tags onto its snapshots.

### spec.instances[].certificateRotationRestart

`bool` · optional (explicit presence)

Restart the instance when its CA certificate rotates. Unset keeps
the AWS default (true -- rotation restarts the instance so the new
certificate is served immediately); false defers the restart to
the next maintenance action, for workloads that cannot absorb an
unscheduled restart. Tri-state on purpose: only an explicit value
is sent to AWS.

- default: `true`

### spec.instances[].applyImmediately

`bool`

Apply modifications to THIS instance immediately instead of
waiting for its next maintenance window. AWS defaults to deferred.

### spec.serverlessV2Scaling

`AwsDocumentDbServerlessV2Scaling`

DocumentDB Serverless capacity bounds. When set, every instance in
`instances` must use class "db.serverless" -- each such instance
scales independently within these bounds. Adding or modifying this
block is an in-place update; REMOVING it from a live cluster
replaces the cluster (AWS cannot switch a cluster off serverless).

- rule: max_capacity must be greater than or equal to min_capacity
- rule: min_capacity and max_capacity must be multiples of 0.5 -- AWS only accepts half-step DCU values

### spec.serverlessV2Scaling.minCapacity

`double` · required

Minimum DCUs, 0.5-256 in half-step multiples. The floor each
serverless instance never scales below -- also the idle cost floor
(DocumentDB Serverless does not pause to zero). Required.

- rule: {"required":true,"double":{"lte":256,"gte":0.5}}

### spec.serverlessV2Scaling.maxCapacity

`double` · required

Maximum DCUs, 1-256 in half-step multiples. The hard
spend/performance ceiling per instance. Required.

- rule: {"required":true,"double":{"lte":256,"gte":1}}

### spec.masterUsername

`string`

The master username. Required for a brand-new cluster -- AWS has no
default and rejects a blank value at CreateDBCluster. Only clusters
that inherit credentials from a source (snapshot restore,
point-in-time restore, or joining a global cluster) leave it empty.
Create-time only -- changing it replaces the cluster.

### spec.manageMasterUserPassword

`bool`

Let AWS manage the master password in Secrets Manager: AWS
generates it, stores it, rotates it on schedule, and no secret ever
touches this manifest or the IaC state. The managed secret's ARN is
exported as the master_user_secret_arn output. Mutually exclusive
with master_password -- and the recommended posture.

- default: `true`

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
(1 day). Backups are continuous -- this window bounds point-in-time
recovery, so production clusters typically want 7+.

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

### spec.enabledCloudwatchLogsExports

`[]string`

Database log types to export to CloudWatch Logs: "audit" (DDL and
authentication events; requires the audit_logs cluster parameter)
and "profiler" (slow-operation profiling; requires the profiler
cluster parameters).

- rule: {"repeated":{"unique":true}}

### spec.snapshotIdentifier

`string`

Restore the cluster from an existing cluster snapshot (name or ARN)
at create time. Create-time only; mutually exclusive with
restore_to_point_in_time.

### spec.restoreToPointInTime

`AwsDocumentDbRestoreToPointInTime`

Clone or restore this cluster from another cluster's continuous
backup at create time -- point-in-time recovery as a first-class
create shape. Create-time only; mutually exclusive with
snapshot_identifier.

- rule: exactly one of restore_to_time or use_latest_restorable_time must be set
- rule: restore_type must be 'full-copy' or 'copy-on-write' when set

### spec.restoreToPointInTime.sourceClusterIdentifier

`string` · required

The source cluster identifier (name or ARN). Required.

- rule: {"required":true}

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
(a fast clone -- shares storage with the source and only pays for
divergence; ideal for prod-data staging environments).

### spec.globalClusterIdentifier

`string`

Join a DocumentDB global cluster: the identifier of the
aws_docdb_global_cluster this cluster participates in. The first
cluster joined becomes the global writer; clusters added afterwards
become read-only secondaries that inherit credentials from the
primary.

### spec.dbClusterParameterGroupName

`string`

The name of an existing cluster parameter group to use. Mutually
exclusive with `parameters` -- either bring your own group or let
the module manage one from inline parameters.

### spec.parameters

`[]AwsDocumentDbParameter`

Cluster-level engine parameters (e.g. "audit_logs", "tls",
"ttl_monitor"), managed as a dedicated parameter group owned by this
cluster (the group is glue -- a named parameter list -- so it stays
folded). Mutually exclusive with db_cluster_parameter_group_name.

- rule: apply_method must be 'immediate' or 'pending-reboot' when set

### spec.parameters[].name

`string` · required

The parameter name (e.g. "audit_logs", "tls",
"ttl_monitor"). Required.

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

### spec.applyImmediately

`bool`

Apply modifications immediately instead of waiting for the next
maintenance window. Immediate changes can interrupt connections;
deferred changes wait quietly. AWS defaults to deferred.

### spec.allowMajorVersionUpgrade

`bool`

Permit engine_version changes that cross a major version. Off (the
default) guards against an accidental major upgrade hidden in a
version bump.

## Validation Rules

- `subnets_or_group`: provide at least two subnet_ids (distinct AZs) or an existing db_subnet_group_name
- `port_range`: port must be between 1150 and 65535 when set -- DocumentDB rejects ports below 1150; 0 keeps the AWS default (27017)
- `password_xor_managed`: master_password cannot be set when manage_master_user_password is true -- pick one password strategy
- `master_username_required_unless_derived`: master_username is required for a new cluster -- AWS rejects a blank username; only snapshot/point-in-time restores and global-cluster members inherit credentials from their source
- `instances_required_unless_headless`: instances is required -- a cluster with no instances cannot serve queries; only snapshot/point-in-time restores and global-cluster members may start headless and attach compute later
- `final_snapshot_id_required_when_not_skipping`: final_snapshot_identifier is required when skip_final_snapshot is false -- AWS refuses to delete the cluster without a final snapshot name
- `log_exports_valid`: enabled_cloudwatch_logs_exports must contain only 'audit' or 'profiler' -- the two log types DocumentDB exports
- `network_type_valid`: network_type must be 'IPV4' or 'DUAL' when set
- `serverless_instances_are_db_serverless`: every instance must use instance_class 'db.serverless' when serverless_v2_scaling is set -- provisioned and serverless instances cannot mix
- `db_serverless_requires_scaling`: serverless_v2_scaling is required when any instance uses class 'db.serverless' -- AWS rejects a serverless instance without capacity bounds on the cluster
- `snapshot_xor_point_in_time`: snapshot_identifier and restore_to_point_in_time are mutually exclusive create-time restore sources
- `own_parameters_xor_existing_group`: parameters and db_cluster_parameter_group_name are mutually exclusive -- manage parameters here or bring an existing group
- `parameters_require_engine_version`: inline parameters require a pinned engine_version -- the managed parameter group's family is derived from it and must match the running engine

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsDocumentDb, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.cluster_identifier` | `string` | The cluster identifier (e.g. "orders-docdb"). |
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
| `status.outputs.instance_endpoints` | `[]string` | Per-instance endpoints of the cluster's folded instances, ordered as declared in spec.instances. Empty for headless shapes (restores and global-cluster members created without instances). |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.subnetIds` | AwsSubnet | `status.outputs.subnet_id` |
| `spec.securityGroupIds` | AwsSecurityGroup | `status.outputs.security_group_id` |
| `spec.instances[].performanceInsightsKmsKeyId` | AwsKmsKey | `status.outputs.key_arn` |
| `spec.kmsKeyId` | AwsKmsKey | `status.outputs.key_arn` |

## See Also

- [Overview](../README.md)
