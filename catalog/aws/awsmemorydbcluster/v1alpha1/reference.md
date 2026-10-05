# AwsMemorydbCluster

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

AwsMemorydbClusterSpec defines the desired configuration for an Amazon MemoryDB
cluster — a fully managed, Redis/Valkey-compatible, durable in-memory database.

Unlike ElastiCache (which is an ephemeral cache), MemoryDB uses a Multi-AZ
distributed transaction log to provide data durability with microsecond reads
and single-digit millisecond writes. Use MemoryDB when you need a
Redis-compatible primary database, not just a caching layer.

Topology is always sharded: `num_shards` controls partitions, and
`num_replicas_per_shard` controls read replicas within each shard. There is no
"clustered vs non-clustered" mode distinction. Both dials scale in place.

Authentication is ACL-based — MemoryDB's only auth model. Every cluster
attaches exactly one Access Control List: reference an AwsMemorydbAcl for
per-application least-privilege users, or the built-in "open-access" ACL
(by literal value) for development. The cluster's name is taken from
`metadata.name` (max 40 characters, AWS-enforced at create).

MemoryDB always encrypts data at rest. `kms_key_arn` optionally specifies a
customer-managed KMS key; without it, the AWS-managed key is used.

Create-time-immutable (ForceNew) fields — changing any of these destroys
and recreates the cluster: `port`, `tls_enabled`, `kms_key_arn`,
`data_tiering`, `auto_minor_version_upgrade`, `network_type`,
`multi_region_cluster_name`, the subnet-group choice, and the snapshot
restore sources. Everything else (node type, shard/replica counts, ACL,
engine and version, parameter group, windows, SNS topic) updates in place.

Credentials, region, and deployment workflow live outside this spec in
IaC inputs.

## Example

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsMemorydbCluster
metadata:
  name: test-memorydb
  id: test-memorydb
  org: test-org
  env: dev
spec:
  region: us-west-2
  engine: valkey
  engineVersion: "7.2"
  description: Full-surface MemoryDB hack manifest
  nodeType: db.r7g.large
  port: 6379
  numShards: 2
  numReplicasPerShard: 1
  aclName:
    value: payments-env-acl
  subnetIds:
    - value: subnet-0a1b2c3d4e5f00001
    - value: subnet-0a1b2c3d4e5f00002
  securityGroupIds:
    - value: sg-0a1b2c3d4e5f00001
  networkType: dual_stack
  ipDiscovery: ipv6
  tlsEnabled: true
  kmsKeyArn:
    value: arn:aws:kms:us-west-2:123456789012:key/abc-123
  maintenanceWindow: sun:05:00-sun:06:00
  snapshotRetentionLimit: 7
  snapshotWindow: 03:00-04:00
  finalSnapshotName: test-memorydb-final
  parameterGroupFamily: memorydb_valkey7
  parameters:
    - name: activedefrag
      value: "yes"
  snsTopicArn:
    value: arn:aws:sns:us-west-2:123456789012:memorydb-events
  autoMinorVersionUpgrade: true
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.engine` | `string` | yes |  |  |
| `spec.engineVersion` | `string` |  |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.nodeType` | `string` | yes |  |  |
| `spec.port` | `int32` |  | `6379` |  |
| `spec.numShards` | `int32` |  | `1` |  |
| `spec.numReplicasPerShard` | `int32` |  | `1` |  |
| `spec.aclName` | `string \| valueFrom` | yes |  | AwsMemorydbAcl (`status.outputs.acl_name`) |
| `spec.subnetIds` | `[]string \| valueFrom` |  |  | AwsSubnet (`status.outputs.subnet_id`) |
| `spec.subnetGroupName` | `string` |  |  |  |
| `spec.securityGroupIds` | `[]string \| valueFrom` |  |  | AwsSecurityGroup (`status.outputs.security_group_id`) |
| `spec.networkType` | `string` |  |  |  |
| `spec.ipDiscovery` | `string` |  |  |  |
| `spec.tlsEnabled` | `bool` |  | `true` |  |
| `spec.kmsKeyArn` | `string \| valueFrom` |  |  | AwsKmsKey (`status.outputs.key_arn`) |
| `spec.maintenanceWindow` | `string` |  |  |  |
| `spec.snapshotRetentionLimit` | `int32` |  |  |  |
| `spec.snapshotWindow` | `string` |  |  |  |
| `spec.finalSnapshotName` | `string` |  |  |  |
| `spec.snapshotArns` | `[]string` |  |  |  |
| `spec.snapshotName` | `string` |  |  |  |
| `spec.parameterGroupFamily` | `string` |  |  |  |
| `spec.parameters` | `[]AwsMemorydbClusterParameter` |  |  |  |
| `spec.parameters[].name` | `string` | yes |  |  |
| `spec.parameters[].value` | `string` | yes |  |  |
| `spec.parameterGroupName` | `string` |  |  |  |
| `spec.multiRegionClusterName` | `string` |  |  |  |
| `spec.snsTopicArn` | `string \| valueFrom` |  |  | AwsSnsTopic (`status.outputs.topic_arn`) |
| `spec.autoMinorVersionUpgrade` | `bool` |  | `true` |  |
| `spec.dataTiering` | `bool` |  |  |  |

## Field Details

### spec.region

`string` · required

The AWS region where the resource will be created.
Example: "us-west-2", "eu-west-1"

- rule: {"string":{"minLen":"1"}}

### spec.engine

`string` · required

Database engine to run. Redis OSS is the long-standing choice; Valkey is
the open-source, Linux-Foundation-governed fork with lower per-node
pricing on AWS. Values: "redis", "valkey". Updates in place — AWS
supports switching a Redis cluster to Valkey (never the reverse:
downgrades are not supported).

- rule: {"required":true}

### spec.engineVersion

`string`

Engine version to deploy. Examples: "7.1", "7.0", "6.2" for Redis;
"7.2", "7.3" for Valkey. Leave empty to let AWS pick the default for
the engine. Upgrades apply in place; downgrades are not supported.
One-way once applied: removing the field keeps the running version
(the provider adopts the live value rather than reverting) — change
it by naming the new version explicitly.

### spec.description

`string`

Human-readable description shown in the AWS console. Leave empty for
none — the modules always send an explicit value so the two IaC engines
never inject their own differing "Managed by ..." defaults.

### spec.nodeType

`string` · required

MemoryDB node type. Determines CPU, memory, and network capacity of every
node in the cluster. Examples: "db.t4g.small" (dev),
"db.r7g.large" (production), "db.r6gd.xlarge" (required for data
tiering). Updates in place — AWS performs a rolling vertical scale.

- rule: {"required":true}

### spec.port

`int32` · optional (explicit presence)

Port on which the cluster accepts connections. Default: 6379.
ForceNew — changing it destroys and recreates the cluster.

- default: `6379`
- rule: {"int32":{"lte":65535,"gte":1}}

### spec.numShards

`int32` · optional (explicit presence)

Number of shards (data partitions) in the cluster. Each shard holds a
portion of the keyspace and has its own primary. Default: 1. Scales in
place (resharding redistributes slots online).

- default: `1`
- rule: {"int32":{"gte":1}}

### spec.numReplicasPerShard

`int32` · optional (explicit presence)

Number of read replicas per shard. Range: 0–5. Default: 1 (each shard
has 1 primary + 1 replica = 2 nodes). Multi-AZ durability comes from the
transaction log even at 0 replicas, but failover is fastest with at
least one replica. Scales in place.

- default: `1`
- rule: {"int32":{"lte":5,"gte":0}}

### spec.aclName

`string | valueFrom` · required

The Access Control List the cluster authenticates against — MemoryDB's
only authentication model. Reference an AwsMemorydbAcl for
per-application users, or set the literal value "open-access" (the
built-in allow-everything ACL, no authentication) for development.
Required: stating open access explicitly is deliberate — an invisible
default that grants unauthenticated access has no place in a manifest.
Updates in place. Note AWS's own coupling: a cluster with `tls_enabled:
false` only accepts the "open-access" ACL (rejected at create otherwise).

- references: AwsMemorydbAcl (`status.outputs.acl_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsMemorydbAcl, name: <that resource's name>, fieldPath: status.outputs.acl_name}} -- a bare string does not parse

### spec.subnetIds

`[]string | valueFrom`

Subnet IDs for a module-managed MemoryDB subnet group. Provide subnets in
at least two AZs for multi-AZ resilience; a subnet group named after the
cluster is created from them. Mutually exclusive with
`subnet_group_name`. When BOTH are omitted, AWS falls back to the
account's "default" subnet group — which only exists in accounts with a
default VPC, so production manifests set one of the two arms.

- references: AwsSubnet (`status.outputs.subnet_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSubnet, name: <that resource's name>, fieldPath: status.outputs.subnet_id}} -- a bare string does not parse

### spec.subnetGroupName

`string`

Name of an existing MemoryDB subnet group to place the cluster in —
the bring-your-own arm, mutually exclusive with `subnet_ids`. ForceNew:
the subnet-group choice is fixed at create time.

### spec.securityGroupIds

`[]string | valueFrom`

VPC security groups to attach to the cluster nodes. Controls
network-level access to the MemoryDB endpoint. Updates in place, with
one AWS quirk: once a cluster has at least one security group, the set
can never be emptied again — only swapped.

- references: AwsSecurityGroup (`status.outputs.security_group_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSecurityGroup, name: <that resource's name>, fieldPath: status.outputs.security_group_id}} -- a bare string does not parse

### spec.networkType

`string`

IP address type for the cluster's network. Values: "ipv4" (default),
"ipv6", "dual_stack". ForceNew — the network type is fixed at create
time. The subnets in the subnet group must support the chosen type.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["ipv4","ipv6","dual_stack"]}}

### spec.ipDiscovery

`string`

How cluster discovery commands (CLUSTER SLOTS / CLUSTER SHARDS) report
node addresses to clients. Values: "ipv4" (default), "ipv6". Setting
"ipv6" requires `network_type` "ipv6" or "dual_stack" — on a dual-stack
cluster this is the dial that moves client traffic onto IPv6. Updates
in place.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["ipv4","ipv6"]}}

### spec.tlsEnabled

`bool` · optional (explicit presence)

Enable TLS for in-transit encryption on all client connections. Default
true — and the provider enforces that default itself, so omitting the
field is identical to sending true; explicit false is the only way to
disable. When false, AWS only accepts the "open-access" ACL (no
authentication without encryption). ForceNew — changing it destroys and
recreates the cluster. IAM-authenticated users also require TLS.

- default: `true`

### spec.kmsKeyArn

`string | valueFrom`

Customer-managed KMS key for at-rest encryption. MemoryDB always
encrypts data at rest; this optionally substitutes your own key for the
AWS-managed one. ForceNew — choose the key at create time.

- references: AwsKmsKey (`status.outputs.key_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_arn}} -- a bare string does not parse

### spec.maintenanceWindow

`string`

Weekly maintenance window in UTC, minimum 60 minutes.
Format: "ddd:hh24:mi-ddd:hh24:mi". Example: "sun:05:00-sun:06:00".
Leave empty for an AWS-assigned window. One-way once applied: removing
the field keeps the current window rather than reverting to an
AWS-assigned one — change it by naming a new window explicitly.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^(mon|tue|wed|thu|fri|sat|sun):([01][0-9]|2[0-3]):[0-5][0-9]-(mon|tue|wed|thu|fri|sat|sun):([01][0-9]|2[0-3]):[0-5][0-9]$"}}

### spec.snapshotRetentionLimit

`int32`

Number of days to retain automatic snapshots. 0 disables automatic
snapshots. Range: 0–35. Updates in place.

- rule: {"int32":{"lte":35,"gte":0}}

### spec.snapshotWindow

`string`

Daily snapshot window in UTC. Format: "hh24:mi-hh24:mi".
Example: "05:00-09:00". Leave empty for an AWS-assigned window.
One-way once applied: removing the field keeps the current window
rather than reverting to an AWS-assigned one — change it by naming a
new window explicitly.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^([01][0-9]|2[0-3]):[0-5][0-9]-([01][0-9]|2[0-3]):[0-5][0-9]$"}}

### spec.finalSnapshotName

`string`

Name of the final snapshot to create when the cluster is deleted. If not
provided, the cluster's data is gone when the cluster is. Consumed only
at delete time — it never affects the running cluster. Format: 1–255
lowercase alphanumeric or hyphen characters, no consecutive hyphens, no
trailing hyphen (the provider rejects violations at plan time; the rule
below fails them at manifest time, before any engine runs).

- rule: final_snapshot_name must be 1-255 lowercase alphanumeric or hyphen characters, with no consecutive hyphens and no trailing hyphen
- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

### spec.snapshotArns

`[]string`

ARN(s) of RDB snapshot files stored in S3 to seed the new cluster from
(the offline-migration path from self-managed Redis). ForceNew — only
read at cluster creation. Mutually exclusive with snapshot_name. Each
entry must be an S3 object ARN (AWS's CreateCluster contract; the
provider itself accepts any ARN shape — a recorded looseness) and must
not contain commas (an AWS API constraint).

- rule: {"repeated":{"items":{"string":{"pattern":"^arn:aws[a-zA-Z-]*:s3:::[^,]+$"}}}}

### spec.snapshotName

`string`

Name of a MemoryDB snapshot to restore from. ForceNew — only read at
cluster creation. Mutually exclusive with snapshot_arns.

### spec.parameterGroupFamily

`string`

Parameter group family for the module-managed parameter group. Required
when `parameters` is provided. Examples: "memorydb_redis7",
"memorydb_valkey7", "memorydb_redis6".

### spec.parameters

`[]AwsMemorydbClusterParameter`

Custom engine parameters to apply via a module-managed parameter group
(named after the cluster). Common examples: activedefrag,
maxmemory-policy. Mutually exclusive with `parameter_group_name`.
Removing an entry resets that parameter to its family default.

### spec.parameters[].name

`string` · required

Parameter name (e.g., "activedefrag", "maxmemory-policy").

- rule: {"required":true}

### spec.parameters[].value

`string` · required

Parameter value (e.g., "yes", "volatile-lru").

- rule: {"required":true}

### spec.parameterGroupName

`string`

Name of an existing MemoryDB parameter group to attach — the
bring-your-own arm, mutually exclusive with the folded `parameters`
list. Leave both empty for the family default parameter group.
Updates in place (AWS waits for the group to be in-sync).

### spec.multiRegionClusterName

`string`

Name of an existing MemoryDB multi-region cluster to join, making this
regional cluster one of its members (active-active writes across
regions). The multi-region cluster is created outside this resource
(its name is server-generated with a chosen suffix); joining is
ForceNew. Leave empty for a single-region cluster.

### spec.snsTopicArn

`string | valueFrom`

SNS topic for cluster event notifications (failover, maintenance,
scaling, configuration changes). Updates in place; clearing it disables
notifications.

- references: AwsSnsTopic (`status.outputs.topic_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSnsTopic, name: <that resource's name>, fieldPath: status.outputs.topic_arn}} -- a bare string does not parse

### spec.autoMinorVersionUpgrade

`bool` · optional (explicit presence)

Automatically apply minor engine version upgrades during maintenance
windows. Default: true — and the provider enforces that default itself,
so omitting the field is identical to sending true; explicit false is
the only way to opt out. ForceNew — AWS fixes this posture at create
time.

- default: `true`

### spec.dataTiering

`bool`

Enable data tiering — automatically moves less-frequently-accessed data
to local SSD for cost efficiency with large datasets. Only available on
db.r6gd.* node types. ForceNew — cannot be changed after creation.

## Validation Rules

- `engine_valid_values`: engine must be 'redis' or 'valkey'
- `subnet_arms_mutual_exclusion`: provide either subnet_ids (module-managed subnet group) or subnet_group_name (existing group), not both
- `parameter_arms_mutual_exclusion`: provide either parameters (module-managed parameter group) or parameter_group_name (existing group), not both
- `parameters_require_family`: parameter_group_family is required when parameters are provided
- `snapshot_restore_mutual_exclusion`: snapshot_arns and snapshot_name are mutually exclusive; choose one restore source
- `ip_discovery_requires_ipv6_network`: ip_discovery 'ipv6' requires network_type 'ipv6' or 'dual_stack'

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsMemorydbCluster, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.cluster_endpoint_address` | `string` | The DNS address of the cluster endpoint. Applications connect here for read-write operations. MemoryDB exposes a single cluster endpoint that handles slot discovery and routing internally. |
| `status.outputs.cluster_endpoint_port` | `int32` | The port of the cluster endpoint. |
| `status.outputs.cluster_arn` | `string` | The Amazon Resource Name of the MemoryDB cluster. Used in IAM policies and cross-service permissions. |
| `status.outputs.cluster_name` | `string` | The name of the MemoryDB cluster. Matches metadata.name. |
| `status.outputs.engine_patch_version` | `string` | The actual engine patch version running on the cluster (e.g., "7.1.0.20"). May differ from the requested engine_version due to automatic patching. |
| `status.outputs.subnet_group_name` | `string` | The name of the subnet group the cluster is placed in — the module-managed group (when subnet_ids were provided), the referenced existing group (when subnet_group_name was set), or empty when the cluster fell back to the account's default group. |
| `status.outputs.parameter_group_name` | `string` | The name of the parameter group attached to the cluster — the module-managed group (when parameters were provided), the referenced existing group (when parameter_group_name was set), or empty when the cluster runs on the family default. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.aclName` | AwsMemorydbAcl | `status.outputs.acl_name` |
| `spec.subnetIds` | AwsSubnet | `status.outputs.subnet_id` |
| `spec.securityGroupIds` | AwsSecurityGroup | `status.outputs.security_group_id` |
| `spec.kmsKeyArn` | AwsKmsKey | `status.outputs.key_arn` |
| `spec.snsTopicArn` | AwsSnsTopic | `status.outputs.topic_arn` |

## See Also

- [Overview](../README.md)
