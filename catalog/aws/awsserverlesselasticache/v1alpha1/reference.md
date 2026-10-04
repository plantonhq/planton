# AwsServerlessElasticache

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

AwsServerlessElasticacheSpec defines the desired configuration for an AWS
ElastiCache Serverless cache.

ElastiCache Serverless removes all node management. There are no instance
types, no replica counts, no parameter groups, no maintenance windows.
AWS automatically scales compute (measured in ElastiCache Processing Units,
ECPU) and storage (measured in GB) within the limits you configure.

This kind supports all three ElastiCache engines:

- **Redis** — in-memory data store with persistence, replication, and
  fine-grained access control via Redis ACL user groups.

- **Valkey** — open-source Redis-compatible engine with the same feature set.

- **Memcached** — volatile key-value cache with no persistence, no
  replication, and no authentication. Simpler and cheaper for pure caching.

Engine-specific features (snapshots, user groups) are guarded by CEL
validations and only valid for Redis/Valkey.

Notes:
- `kms_key_id` and `subnet_ids` are ForceNew — changing them destroys and
  recreates the cache. Design encryption and networking choices upfront.
- Credentials, region, and deployment workflow live outside this spec in
  IaC inputs.

## Example

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsServerlessElasticache
metadata:
  name: test-serverless-cache
  org: test-org
  env: dev
  id: test-serverless-cache-dev
spec:
  region: us-west-2
  engine: redis
  majorEngineVersion: "7"
  description: Test serverless Redis cache
  dataStorageMaxGb: 10
  ecpuMax: 5000
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.engine` | `string` | yes |  |  |
| `spec.majorEngineVersion` | `string` |  |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.dataStorageMaxGb` | `int32` |  |  |  |
| `spec.dataStorageMinGb` | `int32` |  |  |  |
| `spec.ecpuMax` | `int32` |  |  |  |
| `spec.ecpuMin` | `int32` |  |  |  |
| `spec.subnetIds` | `[]string \| valueFrom` |  |  | AwsSubnet (`status.outputs.subnet_id`) |
| `spec.securityGroupIds` | `[]string \| valueFrom` |  |  | AwsSecurityGroup (`status.outputs.security_group_id`) |
| `spec.networkType` | `string` |  |  |  |
| `spec.kmsKeyId` | `string \| valueFrom` |  |  | AwsKmsKey (`status.outputs.key_arn`) |
| `spec.dailySnapshotTime` | `string` |  |  |  |
| `spec.snapshotRetentionLimit` | `int32` |  |  |  |
| `spec.snapshotArnsToRestore` | `[]string` |  |  |  |
| `spec.userGroupId` | `string \| valueFrom` |  |  | AwsElasticacheUserGroup (`status.outputs.user_group_id`) |

## Field Details

### spec.region

`string` · required

The AWS region where the resource will be created.
Example: "us-west-2", "eu-west-1"

- rule: {"string":{"minLen":"1"}}

### spec.engine

`string` · required

Cache engine to use. Values: "redis", "valkey", "memcached".
Switching between redis and valkey is an in-place update. Switching
to/from memcached forces recreation.

- rule: {"required":true}

### spec.majorEngineVersion

`string`

Major engine version. Examples: "7", "8" for Redis/Valkey; "1.6" for
Memcached. Leave empty to use the provider default for the chosen engine.

### spec.description

`string`

Human-readable description of the serverless cache.

### spec.dataStorageMaxGb

`int32`

Maximum data storage in GB. AWS auto-scales storage up to this limit.
Range: 1–5000. Leave as 0 to use the AWS default for the engine.

- rule: {"int32":{"lte":5000,"gte":0}}

### spec.dataStorageMinGb

`int32`

Minimum data storage in GB. AWS guarantees at least this capacity is
always provisioned. Range: 1–5000. Leave as 0 to use the AWS default.

- rule: {"int32":{"lte":5000,"gte":0}}

### spec.ecpuMax

`int32`

Maximum ElastiCache Processing Units per second. AWS auto-scales compute
up to this limit. Range: 1000–15000000. Leave as 0 for AWS default.

- rule: {"int32":{"lte":15000000,"gte":0}}

### spec.ecpuMin

`int32`

Minimum ElastiCache Processing Units per second. AWS guarantees at least
this compute capacity. Range: 1000–15000000. Leave as 0 for AWS default.

- rule: {"int32":{"lte":15000000,"gte":0}}

### spec.subnetIds

`[]string | valueFrom`

Subnet IDs for the serverless cache's VPC endpoint. The cache creates
VPC endpoints in these subnets. ForceNew — changing this destroys and
recreates the cache.

- references: AwsSubnet (`status.outputs.subnet_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSubnet, name: <that resource's name>, fieldPath: status.outputs.subnet_id}} -- a bare string does not parse

### spec.securityGroupIds

`[]string | valueFrom`

VPC security groups to attach to the serverless cache endpoint.
Controls network-level access.

- references: AwsSecurityGroup (`status.outputs.security_group_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSecurityGroup, name: <that resource's name>, fieldPath: status.outputs.security_group_id}} -- a bare string does not parse

### spec.networkType

`string`

IP addressing for the cache's VPC endpoints. Values: "ipv4" (default),
"ipv6", "dual_stack". ForceNew — changing the network type destroys and
recreates the cache. Dual-stack requires subnets with both IPv4 and
IPv6 CIDRs.

### spec.kmsKeyId

`string | valueFrom`

Customer-managed KMS key ARN for at-rest encryption. When set, ElastiCache
Serverless uses this key instead of the AWS-managed key. ForceNew —
changing this destroys and recreates the cache.

- references: AwsKmsKey (`status.outputs.key_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_arn}} -- a bare string does not parse

### spec.dailySnapshotTime

`string`

Daily automatic snapshot time in UTC. Format: "HH:mm" (e.g., "05:00").
Only valid for Redis/Valkey engines. Memcached has no persistence.

### spec.snapshotRetentionLimit

`int32`

Number of days to retain automatic snapshots. Range: 0–35. 0 disables
snapshots. Only valid for Redis/Valkey engines.

- rule: {"int32":{"lte":35,"gte":0}}

### spec.snapshotArnsToRestore

`[]string`

ARNs of existing ElastiCache snapshots to seed the new cache from — the
migration path from a node-based Redis/Valkey cluster to serverless
(snapshot the cluster, restore here). Create-time-only; only valid for
Redis/Valkey engines.

### spec.userGroupId

`string | valueFrom`

RBAC user group controlling fine-grained access
(AwsElasticacheUser/AwsElasticacheUserGroup). Serverless caches accept
exactly one group. Only valid for Redis/Valkey engines — Memcached has
no authentication mechanism.

- references: AwsElasticacheUserGroup (`status.outputs.user_group_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsElasticacheUserGroup, name: <that resource's name>, fieldPath: status.outputs.user_group_id}} -- a bare string does not parse

## Validation Rules

- `engine_valid_values`: engine must be 'redis', 'valkey', or 'memcached'
- `data_storage_min_max`: data_storage_min_gb must not exceed data_storage_max_gb when both are set
- `ecpu_min_max`: ecpu_min must not exceed ecpu_max when both are set
- `snapshot_time_engine_guard`: daily_snapshot_time is only valid for redis or valkey engines
- `snapshot_retention_engine_guard`: snapshot_retention_limit is only valid for redis or valkey engines
- `user_group_engine_guard`: user_group_id is only valid for redis or valkey engines — Memcached has no authentication mechanism
- `snapshot_restore_engine_guard`: snapshot_arns_to_restore is only valid for redis or valkey engines — Memcached has no persistence to restore
- `network_type_valid_values`: network_type must be 'ipv4', 'ipv6', or 'dual_stack' when set
- `data_storage_min_floor`: data_storage_min_gb must be at least 1 when set
- `data_storage_max_floor`: data_storage_max_gb must be at least 1 when set
- `ecpu_min_floor`: ecpu_min must be at least 1000 when set
- `ecpu_max_floor`: ecpu_max must be at least 1000 when set

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsServerlessElasticache, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.arn` | `string` | The Amazon Resource Name of the serverless cache. Used in IAM policies and cross-service permissions. |
| `status.outputs.endpoint_address` | `string` | The primary connection endpoint DNS address. All clients connect here for read-write operations. |
| `status.outputs.endpoint_port` | `int32` | The port of the primary connection endpoint. |
| `status.outputs.reader_endpoint_address` | `string` | The reader endpoint DNS address. Distributes read traffic for Redis/Valkey engines. Empty for Memcached (no read replicas). |
| `status.outputs.reader_endpoint_port` | `int32` | The port of the reader endpoint. |
| `status.outputs.full_engine_version` | `string` | The full engine version string (e.g., "7.1.0"). Useful for confirming the exact version deployed when only major_engine_version was specified. |
| `status.outputs.name` | `string` | The name of the serverless cache. Matches the metadata ID used during creation. Useful for downstream references and data source lookups. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.subnetIds` | AwsSubnet | `status.outputs.subnet_id` |
| `spec.securityGroupIds` | AwsSecurityGroup | `status.outputs.security_group_id` |
| `spec.kmsKeyId` | AwsKmsKey | `status.outputs.key_arn` |
| `spec.userGroupId` | AwsElasticacheUserGroup | `status.outputs.user_group_id` |

## See Also

- [Overview](../README.md)
