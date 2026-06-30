# AwsServerlessElasticache

Provisions an AWS ElastiCache Serverless cache — a fully managed, auto-scaling in-memory data store that eliminates all node management. AWS automatically scales compute and storage within configurable limits.

## When to Use

| Use Case | Component |
|---|---|
| Serverless, pay-per-use caching with zero node management | **AwsServerlessElasticache** (this) |
| Provisioned Redis/Valkey with explicit node types and topology control | [AwsRedisElasticache](../awsrediselasticache/v1/) |
| Provisioned Memcached with explicit node types and cluster sizing | [AwsMemcachedElasticache](../awsmemcachedelasticache/v1/) |

Choose serverless when:
- You want zero infrastructure management (no node types, no replica counts, no parameter tuning)
- Your workload has variable or unpredictable traffic patterns
- You prefer pay-per-use billing over reserved capacity
- You're prototyping or running development environments

Choose provisioned when:
- You need fine-grained control over instance types and topology
- You have predictable, steady-state workloads where reserved capacity is cheaper
- You need custom parameter groups for engine tuning
- You need features not available in serverless (e.g., Global Datastore, data tiering)

## Supported Engines

| Engine | Description | Snapshots | User Groups | Reader Endpoint |
|---|---|---|---|---|
| **redis** | In-memory data store with persistence and replication | Yes | Yes | Yes |
| **valkey** | Open-source Redis-compatible engine | Yes | Yes | Yes |
| **memcached** | Volatile key-value cache, no persistence | No | No | No |

## Spec Fields

### Engine

| Field | Type | Required | Description |
|---|---|---|---|
| `engine` | string | **Yes** | Cache engine: `redis`, `valkey`, or `memcached` |
| `major_engine_version` | string | No | Major version (e.g., `7`, `8`). AWS default if empty |
| `description` | string | No | Human-readable description |

### Scaling Limits — Data Storage

| Field | Type | Required | Range | Description |
|---|---|---|---|---|
| `data_storage_max_gb` | int32 | No | 1–5000 | Maximum data storage in GB |
| `data_storage_min_gb` | int32 | No | 1–5000 | Minimum guaranteed data storage in GB |

### Scaling Limits — Compute (ECPU)

| Field | Type | Required | Range | Description |
|---|---|---|---|---|
| `ecpu_max` | int32 | No | 1000–15000000 | Maximum ElastiCache Processing Units per second |
| `ecpu_min` | int32 | No | 1000–15000000 | Minimum guaranteed ECPU per second |

### Networking

| Field | Type | Required | Description |
|---|---|---|---|
| `subnet_ids` | repeated StringValueOrRef | No | VPC subnet IDs. ForceNew |
| `security_group_ids` | repeated StringValueOrRef | No | VPC security group IDs |

### Encryption

| Field | Type | Required | Description |
|---|---|---|---|
| `kms_key_id` | StringValueOrRef | No | Customer-managed KMS key ARN. ForceNew |

### Snapshots (Redis/Valkey Only)

| Field | Type | Required | Range | Description |
|---|---|---|---|---|
| `daily_snapshot_time` | string | No | — | UTC time in `HH:mm` format |
| `snapshot_retention_limit` | int32 | No | 0–35 | Days to retain automatic snapshots |

### Authentication (Redis/Valkey Only)

| Field | Type | Required | Description |
|---|---|---|---|
| `user_group_id` | string | No | Redis ACL user group ID |

## Stack Outputs

| Output | Type | Description |
|---|---|---|
| `arn` | string | Amazon Resource Name |
| `endpoint_address` | string | Primary connection endpoint DNS |
| `endpoint_port` | int32 | Primary connection port |
| `reader_endpoint_address` | string | Reader endpoint DNS (empty for Memcached) |
| `reader_endpoint_port` | int32 | Reader endpoint port |
| `full_engine_version` | string | Full version string (e.g., `7.1.0`) |
| `name` | string | Serverless cache name |

## Validation Rules

- `engine` must be `redis`, `valkey`, or `memcached`
- `data_storage_min_gb` must not exceed `data_storage_max_gb` when both set
- `ecpu_min` must not exceed `ecpu_max` when both set
- `daily_snapshot_time`, `snapshot_retention_limit`, and `user_group_id` are only valid for `redis` or `valkey`
- ECPU values must be at least 1000 when set (non-zero)
- Data storage values must be at least 1 GB when set (non-zero)

## ForceNew Fields

These fields cannot be changed after creation — modifying them destroys and recreates the cache:

- `kms_key_id`
- `subnet_ids`
- Engine changes to/from `memcached` (switching between `redis` and `valkey` is in-place)

## Prerequisites

- An AWS account with ElastiCache Serverless availability in your target region
- (Optional) A VPC with private subnets for VPC-based deployment
- (Optional) A KMS key for customer-managed encryption
- (Optional) A Redis ACL user group for fine-grained access control
