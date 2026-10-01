# AwsNeptuneCluster

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

AwsNeptuneClusterSpec defines an Amazon Neptune graph database cluster
-- a managed graph database supporting property graphs (Apache
TinkerPop Gremlin, openCypher) and RDF (SPARQL) over shared cluster
storage.

The cluster is the shared-storage brain -- endpoints, backups,
encryption, and engine lifecycle live here. The compute that serves
queries is the `instances` list: each entry materializes as its own DB
instance inside the cluster (a writer plus any readers). Cluster
instances are sub-resources of exactly one cluster and are referenced
by nothing else, so they are folded into this spec rather than modeled
as a standalone kind; both IaC modules manage each entry as its own
provider resource keyed by name, so adding or removing a reader is an
in-place update, never a cluster replacement.

Neptune has no master username or password -- that is AWS's design, not
an omission: access is controlled by network reachability (security
groups) and, with iam_database_authentication_enabled, SigV4-signed
requests from IAM identities.

The cluster identifier comes from metadata.name. Security groups,
subnets, KMS keys, and IAM roles compose by reference -- this cluster
never creates or mutates resources that deserve to be their own nodes.

## Example

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsNeptuneCluster
metadata:
  name: awsneptunecluster-demo
spec:
  region: us-west-2
  subnetIds:
    - value: subnet-0a1b2c3d4e5f60001
    - value: subnet-0a1b2c3d4e5f60002
  storageEncrypted: true
  iamDatabaseAuthenticationEnabled: true
  skipFinalSnapshot: true
  serverlessV2Scaling:
    minCapacity: 1
    maxCapacity: 8
  instances:
    - name: writer
      instanceClass: db.serverless
  customEndpoints:
    - name: analytics
      endpointType: READER
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.subnetIds` | `[]string \| valueFrom` |  |  | AwsSubnet (`status.outputs.subnet_id`) |
| `spec.neptuneSubnetGroupName` | `string` |  |  |  |
| `spec.securityGroupIds` | `[]string \| valueFrom` |  |  | AwsSecurityGroup (`status.outputs.security_group_id`) |
| `spec.availabilityZones` | `[]string` |  |  |  |
| `spec.port` | `int32` |  |  |  |
| `spec.engineVersion` | `string` |  |  |  |
| `spec.storageType` | `string` |  |  |  |
| `spec.instances` | `[]AwsNeptuneClusterInstance` |  |  |  |
| `spec.instances[].name` | `string` | yes |  |  |
| `spec.instances[].instanceClass` | `string` | yes |  |  |
| `spec.instances[].promotionTier` | `int32` |  |  |  |
| `spec.instances[].availabilityZone` | `string` |  |  |  |
| `spec.instances[].publiclyAccessible` | `bool` |  |  |  |
| `spec.instances[].neptuneParameterGroupName` | `string` |  |  |  |
| `spec.instances[].autoMinorVersionUpgrade` | `bool` |  | `true` |  |
| `spec.instances[].preferredMaintenanceWindow` | `string` |  |  |  |
| `spec.serverlessV2Scaling` | `AwsNeptuneClusterServerlessV2Scaling` |  |  |  |
| `spec.serverlessV2Scaling.minCapacity` | `double` | yes |  |  |
| `spec.serverlessV2Scaling.maxCapacity` | `double` | yes |  |  |
| `spec.storageEncrypted` | `bool` |  | `true` |  |
| `spec.kmsKeyId` | `string \| valueFrom` |  |  | AwsKmsKey (`status.outputs.key_arn`) |
| `spec.iamDatabaseAuthenticationEnabled` | `bool` |  |  |  |
| `spec.iamRoles` | `[]string \| valueFrom` |  |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.backupRetentionPeriod` | `int32` |  |  |  |
| `spec.preferredBackupWindow` | `string` |  |  |  |
| `spec.preferredMaintenanceWindow` | `string` |  |  |  |
| `spec.copyTagsToSnapshot` | `bool` |  |  |  |
| `spec.skipFinalSnapshot` | `bool` |  |  |  |
| `spec.finalSnapshotIdentifier` | `string` |  |  |  |
| `spec.deletionProtection` | `bool` |  |  |  |
| `spec.enabledCloudwatchLogsExports` | `[]string` |  |  |  |
| `spec.snapshotIdentifier` | `string` |  |  |  |
| `spec.replicationSourceIdentifier` | `string` |  |  |  |
| `spec.globalClusterIdentifier` | `string` |  |  |  |
| `spec.neptuneClusterParameterGroupName` | `string` |  |  |  |
| `spec.parameters` | `[]AwsNeptuneClusterParameter` |  |  |  |
| `spec.parameters[].name` | `string` | yes |  |  |
| `spec.parameters[].value` | `string` | yes |  |  |
| `spec.parameters[].applyMethod` | `string` |  |  |  |
| `spec.neptuneInstanceParameterGroupName` | `string` |  |  |  |
| `spec.applyImmediately` | `bool` |  |  |  |
| `spec.allowMajorVersionUpgrade` | `bool` |  |  |  |
| `spec.customEndpoints` | `[]AwsNeptuneClusterCustomEndpoint` |  |  |  |
| `spec.customEndpoints[].name` | `string` | yes |  |  |
| `spec.customEndpoints[].endpointType` | `string` | yes |  |  |
| `spec.customEndpoints[].staticMembers` | `[]string` |  |  |  |
| `spec.customEndpoints[].excludedMembers` | `[]string` |  |  |  |
| `spec.instanceParameters` | `[]AwsNeptuneClusterParameter` |  |  |  |
| `spec.instanceParameters[].name` | `string` | yes |  |  |
| `spec.instanceParameters[].value` | `string` | yes |  |  |
| `spec.instanceParameters[].applyMethod` | `string` |  |  |  |

## Field Details

### spec.region

`string` · required

The AWS region the cluster is created in. Must match the region of
the subnets, security groups, and KMS keys it references.
Example: "us-west-2", "eu-west-1".

- rule: {"string":{"minLen":"1"}}

### spec.subnetIds

`[]string | valueFrom`

Subnets for the cluster's Neptune subnet group. Provide at least
two subnets in DISTINCT availability zones -- AWS rejects a subnet
group that covers fewer than two AZs. Reference AwsSubnet subnet_id
outputs or pass literal subnet IDs. The module manages the subnet
group itself (pure glue: a named list of subnets); alternatively
point neptune_subnet_group_name at an existing group.

- references: AwsSubnet (`status.outputs.subnet_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSubnet, name: <that resource's name>, fieldPath: status.outputs.subnet_id}} -- a bare string does not parse

### spec.neptuneSubnetGroupName

`string`

Name of an existing Neptune subnet group to place the cluster in,
instead of providing subnet_ids. Changing the subnet group replaces
the cluster.

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

Availability zones the cluster's storage is replicated across (at
most three). Leave empty and AWS picks the zones automatically --
the right call almost always, because this list is create-time-only
and a later change replaces the cluster.

- rule: {"repeated":{"maxItems":"3"}}

### spec.port

`int32`

The port the cluster accepts connections on. 0 keeps the AWS
default (8182 -- the Neptune convention). Create-time only --
changing the port replaces the cluster. Both engines also pin the
cluster's port on every instance resource so instance state always
converges with the cluster (instances have no port of their own --
they listen on the cluster's).

- rule: {"int32":{"lte":65535,"gte":0}}

### spec.engineVersion

`string`

The Neptune engine version, e.g. "1.4.5.1". Leave empty to let AWS
pick the current default version -- an empty pin never goes stale.
Minor upgrades apply in place; major upgrades additionally need
allow_major_version_upgrade and neptune_instance_parameter_group_name.

### spec.storageType

`string`

The storage I/O model: "" or "standard" (billed per I/O, the AWS
default) or "iopt1" (I/O-Optimized -- predictable pricing that pays
off for I/O-heavy workloads; requires engine version 1.3+ and is
switchable once per 30 days).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["standard","iopt1"]}}

### spec.instances

`[]AwsNeptuneClusterInstance`

The DB instances that serve this cluster's queries -- one writer
(lowest promotion tier) plus any number of readers. Each entry is
managed as its own provider resource keyed by `name`, so scaling
readers in and out never touches the cluster. Empty is only valid
for headless shapes that attach compute later: a snapshot restore,
a cross-region replica, or a cluster joined to a global cluster --
a regular cluster with no instances stores data but cannot serve a
single query.

### spec.instances[].name

`string` · required

Instance name, unique within the cluster. Becomes part of the AWS
instance identifier and the key both IaC engines manage the
provider resource by -- renaming an entry replaces that instance
(the others are untouched). Required.

- rule: {"required":true,"string":{"pattern":"^[a-z]([a-z0-9]|-[a-z0-9])*$"}}

### spec.instances[].instanceClass

`string` · required

The instance class: a provisioned class like "db.r6g.large", or
"db.serverless" for a Neptune Serverless instance that scales
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
for anything production-shaped. Create-time only.

### spec.instances[].neptuneParameterGroupName

`string`

The DB (instance-level) parameter group for this instance. Empty
keeps the engine default group.

### spec.instances[].autoMinorVersionUpgrade

`bool` · optional (explicit presence)

Apply minor engine version patches automatically during the
maintenance window. AWS defaults this to true; disable only when
patch timing must be controlled manually.

- default: `true`

### spec.instances[].preferredMaintenanceWindow

`string`

The weekly maintenance window for THIS instance in UTC, format
"ddd:hh24:mi-ddd:hh24:mi". Empty inherits scheduling from AWS --
stagger per-instance windows so readers never patch simultaneously.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^(mon|tue|wed|thu|fri|sat|sun):([01][0-9]|2[0-3]):[0-5][0-9]-(mon|tue|wed|thu|fri|sat|sun):([01][0-9]|2[0-3]):[0-5][0-9]$"}}

### spec.serverlessV2Scaling

`AwsNeptuneClusterServerlessV2Scaling`

Neptune Serverless capacity bounds. When set, every instance in
`instances` must use class "db.serverless" -- each such instance
scales independently within these bounds.

- rule: max_capacity must be greater than or equal to min_capacity

### spec.serverlessV2Scaling.minCapacity

`double` · required

Minimum NCUs, 1-128. The floor each serverless instance never
scales below -- also the idle cost floor (Neptune Serverless does
not pause to zero). Required.

- rule: {"required":true,"double":{"lte":128,"gte":1}}

### spec.serverlessV2Scaling.maxCapacity

`double` · required

Maximum NCUs, 1-128. The hard spend/performance ceiling per
instance. Required.

- rule: {"required":true,"double":{"lte":128,"gte":1}}

### spec.storageEncrypted

`bool` · optional (explicit presence)

Encrypt cluster storage at rest. Defaults to TRUE (secure by
default) -- and create-time only: an unencrypted cluster cannot be
encrypted later (requires a snapshot-restore migration). Set an
explicit false only where encryption cannot apply: restoring an
UNENCRYPTED snapshot, or joining a global cluster whose primary is
unencrypted (encryption must match the source in both cases).

- default: `true`

### spec.kmsKeyId

`string | valueFrom`

The KMS key for storage encryption when storage_encrypted is true.
Empty uses the AWS-managed aws/rds key. Reference an AwsKmsKey
key_arn output or pass a literal key ARN. Create-time only.

- references: AwsKmsKey (`status.outputs.key_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_arn}} -- a bare string does not parse

### spec.iamDatabaseAuthenticationEnabled

`bool`

Require SigV4-signed requests from IAM identities to query the
database -- Neptune's only credential mechanism (there is no master
username/password). Off relies purely on network reachability.
(The AWS default is off; an explicit false at create time is
indistinguishable from omitting it -- the provider only sends the
flag on later changes.)

### spec.iamRoles

`[]string | valueFrom`

IAM roles the cluster assumes for engine features that reach into
other AWS services (the Neptune bulk loader reading from S3,
Neptune ML with SageMaker). Reference AwsIamRole role_arn outputs
or pass literal ARNs -- the roles own their policies; this cluster
only associates them.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

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

### spec.copyTagsToSnapshot

`bool`

Copy the cluster's tags onto automated and manual snapshots.

### spec.skipFinalSnapshot

`bool`

Skip the final snapshot when the cluster is deleted. When false
(the safe default), final_snapshot_identifier must be set -- AWS
refuses to delete without knowing the snapshot name. Also forwarded
to each cluster instance's delete (instance-level final snapshots
do not apply to cluster members, but the flag keeps teardown intent
consistent).

### spec.finalSnapshotIdentifier

`string`

The name for the final snapshot taken on deletion. Required when
skip_final_snapshot is false. Letters, digits, and hyphens; no
double or trailing hyphen (the provider validates this at plan
time).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^[A-Za-z0-9]([A-Za-z0-9]|-[A-Za-z0-9])*$"}}

### spec.deletionProtection

`bool`

Refuse deletion of the cluster while enabled. Turn this on for
anything holding data you cannot recreate -- deletion then requires
an explicit two-step (disable, delete).

### spec.enabledCloudwatchLogsExports

`[]string`

Database log types to export to CloudWatch Logs: "audit"
(connection and query audit trail; requires the
neptune_enable_audit_log cluster parameter) and "slowquery"
(queries exceeding the slow-query threshold parameters).

- rule: {"repeated":{"unique":true}}

### spec.snapshotIdentifier

`string`

Restore the cluster from an existing cluster snapshot (name or ARN)
at create time. Create-time only.

### spec.replicationSourceIdentifier

`string`

Make this cluster a read replica of the given source cluster ARN.
Promote by clearing the field.

### spec.globalClusterIdentifier

`string`

Join a Neptune global database: the identifier of the global
cluster this cluster participates in. The first cluster joined
becomes the global writer; clusters added afterwards become
read-only secondaries. Letter-first; letters, digits, and hyphens;
no double or trailing hyphen.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"255","pattern":"^[A-Za-z]([A-Za-z0-9]|-[A-Za-z0-9])*$"}}

### spec.neptuneClusterParameterGroupName

`string`

The name of an existing cluster parameter group to use. Mutually
exclusive with `parameters` -- either bring your own group or let
the module manage one from inline parameters.

### spec.parameters

`[]AwsNeptuneClusterParameter`

Cluster-level engine parameters (e.g. "neptune_enable_audit_log",
"neptune_query_timeout"), managed as a dedicated parameter group
owned by this cluster (the group is glue -- a named parameter list
-- so it stays folded). Mutually exclusive with
neptune_cluster_parameter_group_name.

- rule: apply_method must be 'immediate' or 'pending-reboot' when set

### spec.parameters[].name

`string` · required

The parameter name (e.g. "neptune_enable_audit_log",
"neptune_query_timeout"). Required.

- rule: {"required":true}

### spec.parameters[].value

`string` · required

The parameter value. Required.

- rule: {"required":true}

### spec.parameters[].applyMethod

`string`

When the change lands: "immediate" (dynamic parameters apply now)
or "pending-reboot" (static parameters wait for the next instance
reboot). Empty defers to the provider default, which is
"pending-reboot" at the pinned provider version -- set "immediate"
explicitly when a dynamic parameter should land right away.

### spec.neptuneInstanceParameterGroupName

`string`

The name of the DB (instance-level) parameter group applied to
cluster instances DURING a major engine version upgrade. AWS
requires it when engine_version changes across a major version.

### spec.applyImmediately

`bool`

Apply modifications immediately instead of waiting for the next
maintenance window. Immediate changes can interrupt connections;
deferred changes wait quietly. AWS defaults to deferred. Applies to
cluster-scope changes AND to instance-scope changes (class resizes,
per-instance windows, parameter group switches) -- both engines
forward it to every cluster instance.

### spec.allowMajorVersionUpgrade

`bool`

Permit engine_version changes that cross a major version. Off (the
default) guards against an accidental major upgrade hidden in a
version bump.

### spec.customEndpoints

`[]AwsNeptuneClusterCustomEndpoint`

Custom cluster endpoints: stable DNS names over chosen subsets of
the cluster's instances (e.g. an analytics endpoint fronting only
the big readers). Each entry is managed as its own provider
resource keyed by `name`, so endpoints come and go without touching
the cluster. Per-endpoint addresses are exported in the
custom_endpoint_addresses output keyed by name.

- rule: static_members and excluded_members are mutually exclusive -- pin the member set or subtract from it, not both

### spec.customEndpoints[].name

`string` · required

Endpoint name, unique within the cluster (lowercase letters,
digits, hyphens; starts with a letter, no double or trailing
hyphen). Becomes the DNS-visible endpoint identifier and the key
both IaC engines manage the provider resource by -- renaming an
entry replaces that endpoint (the cluster is untouched). Required.

- rule: {"required":true,"string":{"pattern":"^[a-z]([a-z0-9]|-[a-z0-9])*$"}}

### spec.customEndpoints[].endpointType

`string` · required

Which instances the endpoint fronts: "READER" (reader instances
only), "WRITER" (the writer), or "ANY" (all instances). Required.

- rule: {"required":true,"string":{"in":["READER","WRITER","ANY"]}}

### spec.customEndpoints[].staticMembers

`[]string`

Pin the endpoint to exactly these instances, by their
spec.instances entry names. Mutually exclusive with
excluded_members. Both empty fronts every instance of the
endpoint's type.

### spec.customEndpoints[].excludedMembers

`[]string`

Front every instance of the endpoint's type EXCEPT these, by their
spec.instances entry names. Mutually exclusive with static_members.

### spec.instanceParameters

`[]AwsNeptuneClusterParameter`

DB (instance-level) engine parameters, managed as a dedicated
instance parameter group owned by this cluster and applied to every
instance that does not bring its own group (the group is glue -- a
named parameter list -- so it stays folded, mirroring `parameters`).
Mutually exclusive with neptune_instance_parameter_group_name.

- rule: apply_method must be 'immediate' or 'pending-reboot' when set

### spec.instanceParameters[].name

`string` · required

The parameter name (e.g. "neptune_enable_audit_log",
"neptune_query_timeout"). Required.

- rule: {"required":true}

### spec.instanceParameters[].value

`string` · required

The parameter value. Required.

- rule: {"required":true}

### spec.instanceParameters[].applyMethod

`string`

When the change lands: "immediate" (dynamic parameters apply now)
or "pending-reboot" (static parameters wait for the next instance
reboot). Empty defers to the provider default, which is
"pending-reboot" at the pinned provider version -- set "immediate"
explicitly when a dynamic parameter should land right away.

## Validation Rules

- `subnets_or_group`: provide at least two subnet_ids (distinct AZs) or an existing neptune_subnet_group_name
- `instances_required_unless_headless`: instances is required -- a cluster with no instances cannot serve queries; only snapshot restores, replicas, and global-cluster members may start headless and attach compute later
- `final_snapshot_id_required_when_not_skipping`: final_snapshot_identifier is required when skip_final_snapshot is false -- AWS refuses to delete the cluster without a final snapshot name
- `log_exports_valid`: enabled_cloudwatch_logs_exports must contain only 'audit' or 'slowquery' -- the two log types Neptune exports
- `serverless_instances_are_db_serverless`: every instance must use instance_class 'db.serverless' when serverless_v2_scaling is set -- provisioned and serverless instances cannot mix
- `db_serverless_requires_scaling`: serverless_v2_scaling is required when any instance uses class 'db.serverless' -- AWS rejects a serverless instance without capacity bounds on the cluster
- `own_parameters_xor_existing_group`: parameters and neptune_cluster_parameter_group_name are mutually exclusive -- manage parameters here or bring an existing group
- `parameters_require_engine_version`: inline parameters require a pinned engine_version -- the managed parameter group's family is derived from it and must match the running engine
- `major_upgrade_needs_instance_parameter_group`: allow_major_version_upgrade requires neptune_instance_parameter_group_name -- AWS applies it to the cluster's instances during the major version upgrade
- `own_instance_parameters_xor_existing_group`: instance_parameters and neptune_instance_parameter_group_name are mutually exclusive -- manage the instance parameter group here or bring an existing one
- `instance_parameters_require_engine_version`: inline instance_parameters require a pinned engine_version -- the managed parameter group's family is derived from it and must match the running engine
- `custom_endpoint_names_unique`: custom_endpoints[].name must be unique -- names key the provider resources
- `custom_endpoint_members_exist`: custom_endpoints member lists must name entries of spec.instances

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsNeptuneCluster, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.cluster_identifier` | `string` | The cluster identifier (e.g. "knowledge-graph"). |
| `status.outputs.arn` | `string` | The Amazon Resource Name of the cluster. |
| `status.outputs.cluster_resource_id` | `string` | The immutable cluster resource ID (cluster-...). Survives identifier renames -- the durable handle for CloudWatch dimensions and IAM database authentication policies. |
| `status.outputs.endpoint` | `string` | The writer endpoint -- send Gremlin/openCypher/SPARQL queries here for reads and writes. |
| `status.outputs.reader_endpoint` | `string` | The reader endpoint -- load-balances read-only queries across the cluster's reader instances. |
| `status.outputs.port` | `int32` | The port the cluster accepts connections on. |
| `status.outputs.hosted_zone_id` | `string` | The Route53 hosted zone ID of the cluster endpoints, for DNS alias records. |
| `status.outputs.engine_version_actual` | `string` | The resolved engine version actually running (meaningful when the spec leaves engine_version to the AWS default). |
| `status.outputs.neptune_subnet_group_name` | `string` | The name of the Neptune subnet group the cluster runs in. |
| `status.outputs.neptune_cluster_parameter_group_name` | `string` | The name of the cluster parameter group in use (module-managed or the referenced existing group). |
| `status.outputs.instance_endpoints` | `[]string` | Per-instance endpoints of the cluster's folded instances, ordered as declared in spec.instances. Empty for headless shapes (restores, replicas, and global-cluster members created without instances). |
| `status.outputs.custom_endpoint_addresses` | `map<string, string>` | Custom endpoint DNS addresses keyed by spec.custom_endpoints[].name (e.g. "analytics" -> "analytics.cluster-custom-...neptune.amazonaws.com"). |
| `status.outputs.neptune_instance_parameter_group_name` | `string` | The name of the module-managed instance parameter group created from spec.instance_parameters. Empty when instance parameters are not managed here. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.subnetIds` | AwsSubnet | `status.outputs.subnet_id` |
| `spec.securityGroupIds` | AwsSecurityGroup | `status.outputs.security_group_id` |
| `spec.kmsKeyId` | AwsKmsKey | `status.outputs.key_arn` |
| `spec.iamRoles` | AwsIamRole | `status.outputs.role_arn` |

## See Also

- [Overview](../README.md)
