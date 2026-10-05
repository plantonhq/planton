# AwsRedisElasticache

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

AwsRedisElasticacheSpec defines the desired configuration for an AWS
ElastiCache replication group running Redis or Valkey.

This component provisions a fully managed, in-memory data store optimized
for sub-millisecond latency. It supports two topology modes:

- **Non-clustered** (Cluster Mode Disabled): one primary node with up to 5
  read replicas. Use `num_cache_clusters` to set the total node count
  (primary + replicas). Suitable for workloads under ~113 GB that benefit
  from read scaling.

- **Clustered** (Cluster Mode Enabled): data is partitioned across multiple
  shards, each with a primary and optional replicas. Use `num_node_groups`
  (shard count) and `replicas_per_node_group`. Suitable for multi-terabyte
  datasets and write scaling.

Specify exactly one of `num_cache_clusters` or `num_node_groups` to select
the mode — unless the group joins a global datastore (see
`global_replication_group_id`), where topology and engine settings are
inherited from the primary.

The replication group's AWS identifier is taken from `metadata.name` —
create-time immutable, so renaming means replacement.

Notes:
- `at_rest_encryption_enabled`, `kms_key_id`, `port`, `network_type`, and
  the restore sources (`snapshot_arns`/`snapshot_name`) are ForceNew in
  AWS — changing them destroys and recreates the cluster. Design
  encryption, networking, and restore choices upfront.
- `auth_token` and `user_group_ids` are mutually exclusive authentication
  methods; user groups (RBAC) are AWS's recommended production model.
- Credentials, region, and deployment workflow live outside this spec in
  IaC inputs.

## Example

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsRedisElasticache
metadata:
  name: test-redis
  org: test-org
  env: dev
  id: test-redis-dev
spec:
  region: us-west-2
  engine: redis
  engineVersion: "7.1"
  description: Test Redis cluster for development
  nodeType: cache.t3.micro
  numCacheClusters: 1
  # Presence-typed encryption and upgrade flags: unset omits the argument
  # (AWS decides, and a global-datastore secondary MUST leave them unset);
  # explicit values pin the choice.
  atRestEncryptionEnabled: true
  transitEncryptionEnabled: true
  autoMinorVersionUpgrade: true
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.engine` | `string` |  |  |  |
| `spec.engineVersion` | `string` |  |  |  |
| `spec.description` | `string` | yes |  |  |
| `spec.nodeType` | `string` |  |  |  |
| `spec.port` | `int32` |  | `6379` |  |
| `spec.numCacheClusters` | `int32` |  |  |  |
| `spec.preferredCacheClusterAzs` | `[]string` |  |  |  |
| `spec.numNodeGroups` | `int32` |  |  |  |
| `spec.replicasPerNodeGroup` | `int32` |  |  |  |
| `spec.nodeGroupConfigurations` | `[]AwsRedisElasticacheNodeGroupConfiguration` |  |  |  |
| `spec.nodeGroupConfigurations[].nodeGroupId` | `string` |  |  |  |
| `spec.nodeGroupConfigurations[].primaryAvailabilityZone` | `string` |  |  |  |
| `spec.nodeGroupConfigurations[].replicaAvailabilityZones` | `[]string` |  |  |  |
| `spec.nodeGroupConfigurations[].replicaCount` | `int32` |  |  |  |
| `spec.nodeGroupConfigurations[].slots` | `string` |  |  |  |
| `spec.automaticFailoverEnabled` | `bool` |  |  |  |
| `spec.multiAzEnabled` | `bool` |  |  |  |
| `spec.durability` | `string` |  |  |  |
| `spec.globalReplicationGroupId` | `string` |  |  |  |
| `spec.subnetIds` | `[]string \| valueFrom` |  |  | AwsSubnet (`status.outputs.subnet_id`) |
| `spec.subnetGroupName` | `string` |  |  |  |
| `spec.securityGroupIds` | `[]string \| valueFrom` |  |  | AwsSecurityGroup (`status.outputs.security_group_id`) |
| `spec.networkType` | `string` |  |  |  |
| `spec.ipDiscovery` | `string` |  |  |  |
| `spec.atRestEncryptionEnabled` | `bool` |  | `true` |  |
| `spec.transitEncryptionEnabled` | `bool` |  | `true` |  |
| `spec.transitEncryptionMode` | `string` |  |  |  |
| `spec.kmsKeyId` | `string \| valueFrom` |  |  | AwsKmsKey (`status.outputs.key_arn`) |
| `spec.authToken` | `string \| valueFrom` (sensitive) |  |  |  |
| `spec.authTokenUpdateStrategy` | `string` |  |  |  |
| `spec.userGroupIds` | `[]string \| valueFrom` |  |  | AwsElasticacheUserGroup (`status.outputs.user_group_id`) |
| `spec.snapshotArns` | `[]string` |  |  |  |
| `spec.snapshotName` | `string` |  |  |  |
| `spec.maintenanceWindow` | `string` |  |  |  |
| `spec.snapshotRetentionLimit` | `int32` |  |  |  |
| `spec.snapshotWindow` | `string` |  |  |  |
| `spec.finalSnapshotIdentifier` | `string` |  |  |  |
| `spec.applyImmediately` | `bool` |  |  |  |
| `spec.parameterGroupFamily` | `string` |  |  |  |
| `spec.parameters` | `[]AwsRedisElasticacheParameter` |  |  |  |
| `spec.parameters[].name` | `string` | yes |  |  |
| `spec.parameters[].value` | `string` | yes |  |  |
| `spec.parameterGroupName` | `string` |  |  |  |
| `spec.logDeliveryConfigurations` | `[]AwsRedisElasticacheLogDeliveryConfig` |  |  |  |
| `spec.logDeliveryConfigurations[].destinationType` | `string` | yes |  |  |
| `spec.logDeliveryConfigurations[].destination` | `string \| valueFrom` | yes |  |  |
| `spec.logDeliveryConfigurations[].logFormat` | `string` | yes |  |  |
| `spec.logDeliveryConfigurations[].logType` | `string` | yes |  |  |
| `spec.notificationTopicArn` | `string \| valueFrom` |  |  | AwsSnsTopic (`status.outputs.topic_arn`) |
| `spec.autoMinorVersionUpgrade` | `bool` |  |  |  |
| `spec.dataTieringEnabled` | `bool` |  |  |  |
| `spec.clusterMode` | `string` |  |  |  |

## Field Details

### spec.region

`string` · required

The AWS region where the resource will be created.
Example: "us-west-2", "eu-west-1"

- rule: {"string":{"minLen":"1"}}

### spec.engine

`string`

Cache engine to use. Redis is the dominant choice; Valkey is the
open-source Redis-compatible alternative. Values: "redis", "valkey".
Redis <-> Valkey switches apply in place (Valkey is protocol-compatible).
Required — unless the group joins a global datastore, where the engine
is inherited from the primary and must be left empty.

### spec.engineVersion

`string`

Engine version to deploy. Examples: "7.1", "7.0", "6.2" for Redis;
"7.2", "8.0" for Valkey. Redis 6+ and Valkey use major.minor ("7.1");
Redis 5 and earlier use full three-part versions ("5.0.6"). Leave empty
to use the provider default.
Must be left empty when joining a global datastore (inherited).

### spec.description

`string` · required

Human-readable description for the replication group. Required by AWS.

- rule: {"required":true}

### spec.nodeType

`string`

ElastiCache node type. Determines CPU, memory, and network capacity.
Examples: "cache.t3.micro" (dev), "cache.r7g.large" (production),
"cache.r6gd.xlarge" (data tiering). Required — unless the group joins a
global datastore, where the node type is inherited from the primary and
must be left empty.

### spec.port

`int32` · optional (explicit presence)

Port on which the cluster accepts connections. Default: 6379.
This is a ForceNew attribute — changing it destroys and recreates the cluster.

- default: `6379`
- rule: {"int32":{"lte":65535,"gte":1}}

### spec.numCacheClusters

`int32`

Total number of cache clusters (nodes) in the replication group. This includes
the primary and all read replicas. For example, 3 means 1 primary + 2 replicas.
Range: 1–6 — AWS's CreateReplicationGroup contract caps a non-clustered
group at 1 primary + 5 replicas (the Terraform provider stopped
validating this cap in 6.35.0; the spec deliberately mirrors AWS's
contract, not the provider's looseness).
Mutually exclusive with `num_node_groups`.

### spec.preferredCacheClusterAzs

`[]string`

Preferred Availability Zones for the cache clusters of a NON-clustered
group, in creation order (first entry hosts the primary). When provided,
the list length must match `num_cache_clusters`. For per-shard placement
in clustered mode use `node_group_configurations` instead — the two are
mutually exclusive.

### spec.numNodeGroups

`int32`

Number of node groups (shards) for Cluster Mode Enabled. Each shard holds a
partition of the keyspace. Mutually exclusive with `num_cache_clusters`,
and forbidden when joining a global datastore (the primary defines the
shard layout).

### spec.replicasPerNodeGroup

`int32`

Number of read replicas per shard. Range: 0–5 — AWS's
CreateReplicationGroup contract ("Valid values are 0 to 5"; the
Terraform provider stopped validating the ceiling in 6.35.0 — the spec
deliberately mirrors AWS's contract, not the provider's looseness).
Only valid when `num_node_groups` is set.

- rule: {"int32":{"lte":5,"gte":0}}

### spec.nodeGroupConfigurations

`[]AwsRedisElasticacheNodeGroupConfiguration`

Per-shard placement for Cluster Mode Enabled — pin each shard's primary
and replicas to specific Availability Zones, control its replica count,
or assign its keyspace slots. Most clustered deployments leave this
empty and let AWS spread shards; reach for it when data locality or an
AZ-aligned client topology demands explicit placement. Requires
`num_node_groups`; mutually exclusive with `preferred_cache_cluster_azs`.
Changing explicit shard placement after creation replaces the group.

### spec.nodeGroupConfigurations[].nodeGroupId

`string`

Identifier for the shard. 1–4 digits (e.g. "0001"). Determines ordering;
when omitted AWS assigns sequential ids.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^\\d{1,4}$"}}

### spec.nodeGroupConfigurations[].primaryAvailabilityZone

`string`

Availability Zone hosting the shard's primary node.

### spec.nodeGroupConfigurations[].replicaAvailabilityZones

`[]string`

Availability Zones for the shard's replicas, in order. The list length
should match replica_count when both are set.

### spec.nodeGroupConfigurations[].replicaCount

`int32`

Number of replicas in this shard. Overrides the group-level
replicas_per_node_group for this shard. Range: 0–5.

- rule: {"int32":{"lte":5,"gte":0}}

### spec.nodeGroupConfigurations[].slots

`string`

Keyspace slots owned by this shard, as a range or list expression
(e.g. "0-5461"). Leave empty for AWS's even distribution — set only
when migrating an existing slot layout.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^[\\d,-]+$"}}

### spec.automaticFailoverEnabled

`bool`

Enable automatic failover to a read replica if the primary fails.
Requires `num_cache_clusters >= 2` (non-clustered) or `num_node_groups > 0`
(clustered mode, where failover is always on).

### spec.multiAzEnabled

`bool`

Deploy replicas across multiple Availability Zones for resilience against
AZ-level failures. Requires `automatic_failover_enabled` to be true.

### spec.durability

`string`

Durability mode for the replication group — controls how writes are
acknowledged relative to replica propagation. Values: "default",
"async", "sync", "disabled". Requires Cluster Mode Enabled
(`num_node_groups`) and engine Valkey 9.0 or later; leave empty
everywhere else. "sync" trades write latency for zero-data-loss
failovers. ForceNew — changing it replaces the group.

### spec.globalReplicationGroupId

`string`

ID of an existing global replication group (Aurora-style cross-region
replication for ElastiCache) this group joins as a SECONDARY. The
secondary inherits engine, engine version, node type, encryption
settings, and shard layout from the global primary — leave `engine`,
`engine_version`, `node_type`, `num_node_groups`, the encryption fields,
the parameter-group fields, and the restore sources empty when set.
ForceNew — a group cannot change datastore membership in place.
Global replication groups themselves are created outside this component;
this field is the join path.

### spec.subnetIds

`[]string | valueFrom`

Subnet IDs for the ElastiCache subnet group. Provide subnets in at least two
AZs for multi-AZ deployments. A subnet group is created automatically from
these subnets. Mutually exclusive with `subnet_group_name`.

- references: AwsSubnet (`status.outputs.subnet_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSubnet, name: <that resource's name>, fieldPath: status.outputs.subnet_id}} -- a bare string does not parse

### spec.subnetGroupName

`string`

Name of an EXISTING ElastiCache subnet group to place the cluster in,
instead of building one from `subnet_ids`. Bring-your-own for
organizations that manage subnet groups centrally. ForceNew — changing
the subnet group replaces the cluster.

### spec.securityGroupIds

`[]string | valueFrom`

VPC security groups to attach to the cluster nodes. Controls network-level
access to the Redis/Valkey endpoint.

- references: AwsSecurityGroup (`status.outputs.security_group_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSecurityGroup, name: <that resource's name>, fieldPath: status.outputs.security_group_id}} -- a bare string does not parse

### spec.networkType

`string`

IP addressing for the cluster's network. Values: "ipv4" (default),
"ipv6", "dual_stack". ForceNew — changing the network type replaces the
cluster. Dual-stack requires subnets with both IPv4 and IPv6 CIDRs.

### spec.ipDiscovery

`string`

Which address family DNS discovery returns to clients. Values: "ipv4",
"ipv6". Only meaningful alongside a dual-stack `network_type` (a
single-stack cluster has nothing to choose); updates in place, letting
clients migrate address families without replacing the cluster.

### spec.atRestEncryptionEnabled

`bool` · optional (explicit presence)

Enable encryption at rest for data stored on disk and in snapshots.
Presence matters: leave unset to let AWS apply its engine default,
set true/false to pin it explicitly. Must be left UNSET when joining a
global datastore — the setting is inherited from the primary, and the
provider rejects the argument's presence alongside
global_replication_group_id. ForceNew — changing it destroys and
recreates the cluster.

- default: `true`

### spec.transitEncryptionEnabled

`bool` · optional (explicit presence)

Enable encryption in transit (TLS) for all client connections and
replication traffic. Strongly recommended for production; required for
AUTH tokens and IAM-authenticated RBAC users. Presence matters: leave
unset to let AWS apply its default (disabled), set true/false to pin it
explicitly. Must be left UNSET when joining a global datastore — the
setting is inherited from the primary, and the provider rejects the
argument's presence alongside global_replication_group_id.

- default: `true`

### spec.transitEncryptionMode

`string`

TLS enforcement mode. "preferred" allows both TLS and non-TLS connections
(useful during migration); "required" enforces TLS for all connections.
Only valid when `transit_encryption_enabled` is true.

### spec.kmsKeyId

`string | valueFrom`

Customer-managed KMS key for at-rest encryption. When set, ElastiCache uses
this key instead of the AWS-managed key. ForceNew — changing this destroys
and recreates the cluster.

- references: AwsKmsKey (`status.outputs.key_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_arn}} -- a bare string does not parse

### spec.authToken

`string | valueFrom` · sensitive

Redis AUTH token (password) for client authentication — the legacy
single-shared-credential model. Requires `transit_encryption_enabled` to
be true. 16–128 printable characters. Mutually exclusive with
`user_group_ids`; prefer RBAC user groups for new deployments.

- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.authTokenUpdateStrategy

`string`

How an auth-token CHANGE is applied to the running cluster. Values:
"ROTATE" (old and new tokens both work until the rotation completes —
zero-downtime), "SET" (the new token replaces the old immediately),
"DELETE" (remove the token entirely and turn AUTH off — the migration
step from AUTH to RBAC user groups). ROTATE and SET require
`auth_token`; DELETE requires it to be ABSENT (the provider rejects a
token alongside DELETE — you are removing it).

### spec.userGroupIds

`[]string | valueFrom`

RBAC user groups controlling fine-grained access. Each group carries
users with specific command and key permissions
(AwsElasticacheUser/AwsElasticacheUserGroup) — AWS's recommended
production authentication model. Mutually exclusive with `auth_token`.

- references: AwsElasticacheUserGroup (`status.outputs.user_group_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsElasticacheUserGroup, name: <that resource's name>, fieldPath: status.outputs.user_group_id}} -- a bare string does not parse

### spec.snapshotArns

`[]string`

S3 ARNs of RDB snapshot files to seed the new replication group from —
the migration path from self-managed Redis (upload the RDB to S3, point
here). ForceNew and create-time-only; mutually exclusive with
`snapshot_name`.

### spec.snapshotName

`string`

Name of an existing ElastiCache snapshot to restore into the new
replication group — the clone-from-backup path. ForceNew and
create-time-only; mutually exclusive with `snapshot_arns`.

### spec.maintenanceWindow

`string`

Weekly maintenance window in UTC. Format: "ddd:hh24:mi-ddd:hh24:mi".
Example: "sun:05:00-sun:06:00". Leave empty for AWS-assigned default.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^(mon|tue|wed|thu|fri|sat|sun):([01][0-9]|2[0-3]):[0-5][0-9]-(mon|tue|wed|thu|fri|sat|sun):([01][0-9]|2[0-3]):[0-5][0-9]$"}}

### spec.snapshotRetentionLimit

`int32`

Number of days to retain automatic snapshots before deletion. 0 disables
automatic snapshots. Range: 0–35.

- rule: {"int32":{"lte":35,"gte":0}}

### spec.snapshotWindow

`string`

Daily snapshot window in UTC. Format: "hh24:mi-hh24:mi".
Example: "03:00-04:00". Leave empty for AWS-assigned default.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^([01][0-9]|2[0-3]):[0-5][0-9]-([01][0-9]|2[0-3]):[0-5][0-9]$"}}

### spec.finalSnapshotIdentifier

`string`

Identifier for the final snapshot taken when the cluster is deleted. If not
provided, no final snapshot is created.

### spec.applyImmediately

`bool`

Apply changes immediately instead of waiting for the next maintenance window.
May cause brief downtime for some operations.

### spec.parameterGroupFamily

`string`

Parameter group family for custom parameters. Required when `parameters` is
provided. Examples: "redis7", "redis6.x", "valkey7", "valkey8".

### spec.parameters

`[]AwsRedisElasticacheParameter`

Custom cache parameters to apply via a managed parameter group. Common
examples: maxmemory-policy, timeout, tcp-keepalive. Mutually exclusive
with `parameter_group_name`.

### spec.parameters[].name

`string` · required

Parameter name (e.g., "maxmemory-policy", "timeout").

- rule: {"required":true}

### spec.parameters[].value

`string` · required

Parameter value (e.g., "volatile-lru", "300").

- rule: {"required":true}

### spec.parameterGroupName

`string`

Name of an EXISTING parameter group to use instead of managing
parameters here. Bring-your-own for organizations that share one tuned
group across many caches. Mutually exclusive with `parameters`.
Note: Cluster Mode Enabled requires a family ".cluster.on" group.

### spec.logDeliveryConfigurations

`[]AwsRedisElasticacheLogDeliveryConfig`

Log delivery configurations for slow-log and/or engine-log. At most 2
entries — one per log type. Logs can be delivered to CloudWatch Logs or
Kinesis Data Firehose.

- rule: destination_type must be 'cloudwatch-logs' or 'kinesis-firehose'
- rule: log_format must be 'text' or 'json'
- rule: log_type must be 'slow-log' or 'engine-log'

### spec.logDeliveryConfigurations[].destinationType

`string` · required

Type of destination. Values: "cloudwatch-logs", "kinesis-firehose".

- rule: {"required":true}

### spec.logDeliveryConfigurations[].destination

`string | valueFrom` · required

Destination identifier. For CloudWatch Logs: the log group name.
For Kinesis Firehose: the delivery stream name.

- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.logDeliveryConfigurations[].logFormat

`string` · required

Log serialization format. Values: "text", "json".

- rule: {"required":true}

### spec.logDeliveryConfigurations[].logType

`string` · required

Type of log to deliver. Values: "slow-log" (commands exceeding slowlog
threshold), "engine-log" (engine-level diagnostic output).

- rule: {"required":true}

### spec.notificationTopicArn

`string | valueFrom`

SNS topic ARN for cluster event notifications (failover, maintenance,
configuration changes, etc.).

- references: AwsSnsTopic (`status.outputs.topic_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSnsTopic, name: <that resource's name>, fieldPath: status.outputs.topic_arn}} -- a bare string does not parse

### spec.autoMinorVersionUpgrade

`bool` · optional (explicit presence)

Automatically apply minor engine version upgrades during maintenance
windows. AWS enables this by default: leave unset to keep the default,
set false to pin the running minor version explicitly, set true to pin
the opt-in. Presence matters — unset is forwarded to AWS as "decide",
never as false.

### spec.dataTieringEnabled

`bool`

Enable data tiering — automatically moves less-frequently-accessed data to
SSD storage for up to 5x more data per node. Only available on r6gd node
types. ForceNew — cannot be changed after creation.

### spec.clusterMode

`string`

Cluster-mode migration setting. Values: "enabled", "compatible",
"disabled". "compatible" runs a non-clustered group in cluster-mode-
compatible form so clients can migrate to the cluster protocol before
the topology actually shards — the online migration path from
non-clustered to clustered. Leave empty to let the topology fields
decide (the common case).

## Validation Rules

- `engine_required_or_inherited`: set engine to 'redis' or 'valkey' — or leave it empty only when joining a global datastore (global_replication_group_id), where the primary's engine is inherited
- `engine_version_inherited_with_global`: leave engine_version empty when joining a global datastore — the primary's engine version is inherited
- `engine_version_format`: engine_version format is invalid — Redis uses '5.0.6'-style below 6, '6.x' or '7.1'-style from 6; Valkey uses '7.2'-style
- `node_type_required_or_inherited`: node_type is required — unless joining a global datastore (global_replication_group_id), where the primary's node type is inherited
- `topology_mode_selection`: specify either num_cache_clusters (non-clustered) or num_node_groups (clustered), not both and not neither — a global-datastore secondary may set only num_cache_clusters
- `num_cache_clusters_range`: num_cache_clusters must be between 1 and 6 when set (1 primary + up to 5 read replicas)
- `replicas_requires_node_groups`: replicas_per_node_group requires num_node_groups to be set
- `preferred_azs_match_node_count`: preferred_cache_cluster_azs must list one AZ per cache cluster (its length must equal num_cache_clusters)
- `node_group_configs_require_clustered`: node_group_configurations require clustered mode (num_node_groups) — use preferred_cache_cluster_azs for non-clustered placement
- `shard_placement_mutual_exclusion`: node_group_configurations and preferred_cache_cluster_azs are mutually exclusive placement controls — one is per-shard, the other per-cache-cluster
- `failover_requires_multi_node`: automatic_failover_enabled requires num_cache_clusters >= 2 or clustered mode (num_node_groups > 0)
- `multi_az_requires_failover`: multi_az_enabled requires automatic_failover_enabled to be true
- `durability_valid_values`: durability must be 'default', 'async', 'sync', or 'disabled' when set
- `durability_requires_clustered`: durability requires Cluster Mode Enabled (num_node_groups) — it is a per-shard write-acknowledgement setting, available on Valkey 9.0 or later
- `cluster_mode_valid_values`: cluster_mode must be 'enabled', 'compatible', or 'disabled' when set
- `network_type_valid_values`: network_type must be 'ipv4', 'ipv6', or 'dual_stack' when set
- `ip_discovery_valid_values`: ip_discovery must be 'ipv4' or 'ipv6' when set
- `auth_mutual_exclusion`: auth_token and user_group_ids are mutually exclusive; choose one authentication method (user groups are the recommended RBAC model)
- `auth_strategy_valid_values`: auth_token_update_strategy must be 'ROTATE', 'SET', or 'DELETE' when set
- `auth_strategy_requires_token`: auth_token_update_strategy ROTATE/SET require auth_token; DELETE removes AUTH and requires auth_token to be absent
- `transit_mode_requires_encryption`: transit_encryption_mode requires transit_encryption_enabled to be explicitly true
- `transit_mode_valid_values`: transit_encryption_mode must be 'preferred' or 'required' when set
- `restore_sources_mutual_exclusion`: snapshot_arns and snapshot_name are alternative restore sources — seed from S3 RDB files or from an ElastiCache snapshot, not both
- `restore_forbidden_with_global`: a global-datastore secondary receives its data from the primary — snapshot_arns and snapshot_name cannot be combined with global_replication_group_id
- `encryption_inherited_with_global`: a global-datastore secondary inherits encryption from the primary — leave at_rest_encryption_enabled, transit_encryption_enabled, and transit_encryption_mode unset when joining
- `parameter_group_inherited_with_global`: a global-datastore secondary inherits its parameter group from the primary — leave parameters, parameter_group_family, and parameter_group_name unset when joining
- `subnet_arms_mutual_exclusion`: subnet_ids and subnet_group_name are mutually exclusive — build a subnet group from subnets or bring an existing group
- `parameter_arms_mutual_exclusion`: parameters and parameter_group_name are mutually exclusive — manage parameters here or bring an existing group
- `parameters_require_family`: parameter_group_family is required when parameters are provided
- `log_delivery_max_two`: at most 2 log delivery configurations are allowed (one per log_type)
- `log_delivery_unique_types`: each log_type (slow-log, engine-log) can only appear once in log_delivery_configurations

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsRedisElasticache, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.replication_group_id` | `string` | The identifier of the replication group. Used in AWS CLI/API calls and as a reference in other ElastiCache operations. |
| `status.outputs.primary_endpoint_address` | `string` | The primary (writer) endpoint DNS name. Applications connect here for read-write operations. For non-clustered mode, this is the single primary node endpoint. |
| `status.outputs.reader_endpoint_address` | `string` | The reader endpoint DNS name. Distributes read traffic across all read replicas using DNS round-robin. Use this endpoint for read-heavy workloads to offload the primary. Empty for single-node deployments. |
| `status.outputs.configuration_endpoint_address` | `string` | The configuration endpoint for Cluster Mode Enabled. Redis clients that support cluster mode use this endpoint for automatic slot discovery and routing. Empty when Cluster Mode is disabled. |
| `status.outputs.arn` | `string` | The Amazon Resource Name of the replication group. Used in IAM policies and cross-service permissions. |
| `status.outputs.port` | `int32` | The port on which the cluster accepts connections. |
| `status.outputs.subnet_group_name` | `string` | The name of the ElastiCache subnet group associated with this cluster. Only populated when `subnet_ids` were provided and a subnet group was created by the module. |
| `status.outputs.parameter_group_name` | `string` | The name of the custom parameter group associated with this cluster. Only populated when `parameters` were provided and a parameter group was created by the module. |
| `status.outputs.engine_version_actual` | `string` | The engine version actually running, as resolved by AWS. Useful for confirming the exact version deployed when `engine_version` was left empty or given as a major version only. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.subnetIds` | AwsSubnet | `status.outputs.subnet_id` |
| `spec.securityGroupIds` | AwsSecurityGroup | `status.outputs.security_group_id` |
| `spec.kmsKeyId` | AwsKmsKey | `status.outputs.key_arn` |
| `spec.userGroupIds` | AwsElasticacheUserGroup | `status.outputs.user_group_id` |
| `spec.notificationTopicArn` | AwsSnsTopic | `status.outputs.topic_arn` |

## See Also

- [Overview](../README.md)
