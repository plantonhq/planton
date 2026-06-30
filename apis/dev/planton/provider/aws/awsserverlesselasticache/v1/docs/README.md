# AwsServerlessElasticache — Architecture and Design

## Overview

ElastiCache Serverless is a consumption-based mode of AWS ElastiCache that removes
all infrastructure decisions. There are no instance types to choose, no replica
counts to set, no parameter groups to tune, and no maintenance windows to manage.
You specify an engine (Redis, Valkey, or Memcached), optional scaling bounds, and
AWS handles everything else.

This component wraps the `aws_elasticache_serverless_cache` Terraform resource
(and the corresponding `elasticache.ServerlessCache` Pulumi resource).

## How Serverless Scaling Works

### ElastiCache Processing Units (ECPU)

ECPU is the compute dimension. Each ElastiCache command consumes a number of ECPUs
based on the data transferred and the complexity of the operation:

- Simple GET/SET on small keys: ~1 ECPU
- Large key operations (>1 KB): proportionally more ECPUs
- Multi-key operations (MGET, pipeline): sum of individual costs

AWS auto-scales compute capacity between `ecpu_min` and `ecpu_max`. If neither is
set, AWS uses engine-specific defaults:

| Engine | Default Min ECPU | Default Max ECPU |
|--------|-----------------|-----------------|
| Redis/Valkey | 1000 | 5000 |
| Memcached | 1000 | 5000 |

### Data Storage

Data storage is measured in GB and auto-scales between `data_storage_min_gb` and
`data_storage_max_gb`. AWS defaults vary by engine:

| Engine | Default Min GB | Default Max GB |
|--------|---------------|---------------|
| Redis/Valkey | 1 | 5 |
| Memcached | 1 | 5 |

You are billed for the actual storage used, not the maximum. The minimum sets a
baseline that AWS pre-provisions for consistent performance.

## Serverless vs. Provisioned — When to Choose Which

| Dimension | Serverless | Provisioned (Redis/Memcached) |
|---|---|---|
| **Management** | Zero — no nodes, no sizing, no maintenance | Full control — you pick instance types and topology |
| **Scaling** | Automatic within bounds | Manual — change node count or instance type |
| **Cost model** | Pay per ECPU + GB used | Pay per node-hour (reserved or on-demand) |
| **Latency** | Consistent sub-millisecond | Consistent sub-millisecond |
| **Parameter tuning** | Not available | Full parameter group control |
| **Engine support** | Redis, Valkey, Memcached | Redis, Valkey (replication group) or Memcached (cluster) |
| **Global replication** | Not available | Available (Redis Global Datastore) |
| **Data tiering** | Not available | Available (r6gd node types) |

**Rule of thumb:** Start with serverless for new workloads. Move to provisioned when
you need parameter tuning, data tiering, global replication, or cost optimization
for steady-state workloads.

## Engine Comparison

### Redis (engine: "redis")

- Full-featured in-memory data store
- Persistence via automatic snapshots
- Redis ACL user groups for fine-grained access control
- Reader endpoint for read scaling
- Data structures: strings, lists, sets, sorted sets, hashes, streams, etc.

### Valkey (engine: "valkey")

- Open-source, Redis-compatible fork maintained by the Linux Foundation
- Same feature set as Redis in serverless mode
- Switching between redis and valkey is an in-place operation (no recreation)

### Memcached (engine: "memcached")

- Simple key-value volatile cache
- No persistence — all data is in-memory only
- No authentication — security relies on VPC network isolation
- No reader endpoint — all operations go through the primary endpoint
- Best for: simple caching where you don't need persistence, data types, or auth

## Networking Architecture

ElastiCache Serverless creates VPC endpoints in your specified subnets. Traffic
between your application and the cache never leaves the AWS network.

```
Application (EC2/ECS/Lambda)
       │
       ▼
Security Group ──► ElastiCache VPC Endpoints (in your subnets)
       │
       ▼
ElastiCache Serverless (managed by AWS)
```

If no `subnet_ids` are specified, AWS creates the cache in a default configuration
that is accessible from the VPC.

## Encryption

ElastiCache Serverless **always encrypts data in transit and at rest**. This is
non-optional (unlike provisioned clusters). The only choice is the encryption key:

- **AWS-managed key** (default): No `kms_key_id` needed
- **Customer-managed key**: Provide a KMS key ARN via `kms_key_id`

The KMS key is ForceNew — changing it after creation destroys and recreates the cache.

## Snapshots and Persistence (Redis/Valkey Only)

Serverless Redis/Valkey supports automatic daily snapshots:

- `daily_snapshot_time`: UTC time for the snapshot window (e.g., "03:00")
- `snapshot_retention_limit`: Number of days to keep snapshots (0–35)

Memcached has no persistence mechanism — snapshots are not available.

## Authentication (Redis/Valkey Only)

Serverless Redis/Valkey supports Redis ACL user groups via `user_group_id`. This
provides fine-grained command-level and key-pattern access control.

Memcached has no authentication. Access control relies entirely on VPC security
groups (network-level isolation).

## Deliberate v1 Omissions

- **`snapshot_arns_to_restore`**: ForceNew, disaster recovery use case. Not included
  for v1 to keep the spec lean. Can be added in v2 if there's demand.
- **Parameter groups**: Serverless caches do not support custom parameter groups.
  AWS manages engine configuration internally.
- **Maintenance windows**: AWS manages maintenance for serverless caches. There is
  no user-configurable maintenance window.
- **Multi-AZ configuration**: Serverless caches are inherently multi-AZ. There is
  no user toggle — AWS distributes endpoints across AZs automatically.
