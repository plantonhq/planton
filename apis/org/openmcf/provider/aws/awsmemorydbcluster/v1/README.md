---
title: "MemoryDB Cluster"
description: "Amazon MemoryDB for Redis/Valkey deployment documentation"
icon: "package"
order: 100
componentName: "awsmemorydbcluster"
---

# AWS MemoryDB Cluster

Deploys an Amazon MemoryDB cluster — a fully managed, Redis-compatible, durable in-memory database with microsecond reads and single-digit millisecond writes. MemoryDB combines the performance of in-memory data stores with the durability of a Multi-AZ transactional log, making it suitable as a primary database for applications that require both speed and data persistence.

## MemoryDB vs ElastiCache

| Aspect | MemoryDB | ElastiCache |
|--------|----------|-------------|
| **Purpose** | Durable primary database | Ephemeral caching layer |
| **Durability** | Always durable (Multi-AZ transaction log) | Ephemeral (snapshots optional) |
| **Write latency** | Single-digit milliseconds | Sub-millisecond |
| **Read latency** | Microseconds | Sub-millisecond |
| **Use when** | Data loss is unacceptable | Losing cached data is tolerable |

**Choose MemoryDB** for session stores, user profiles, leaderboards, and any workload where you need Redis performance with database-grade durability. **Choose ElastiCache** for pure caching layers where the source of truth lives elsewhere.

## What Gets Created

When you deploy an AwsMemorydbCluster resource, OpenMCF provisions:

- **MemoryDB Cluster** — a sharded Redis/Valkey cluster with configurable shards and replicas per shard, always-on encryption at rest, and optional TLS for in-transit encryption
- **Subnet Group** — created automatically when `subnetIds` are provided, placing cluster nodes in the specified VPC subnets
- **Parameter Group** — created automatically when `parameters` are provided with a `parameterGroupFamily`, allowing custom engine tuning

## Prerequisites

- **AWS credentials** configured via environment variables or OpenMCF provider config
- **VPC subnets** in at least two Availability Zones for production deployments
- **Security group** allowing inbound traffic on the Redis port (default 6379) from your application instances
- **MemoryDB ACL** — use the built-in `open-access` for development, or create a custom ACL with users for production authentication

## Quick Start

Create a file `memorydb.yaml`:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsMemorydbCluster
metadata:
  name: my-memorydb
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AwsMemorydbCluster.my-memorydb
spec:
  engine: redis
  engineVersion: "7.1"
  nodeType: db.t4g.small
  numShards: 1
  numReplicasPerShard: 0
  aclName: open-access
  tlsEnabled: true
```

Deploy:

```shell
openmcf apply -f memorydb.yaml
```

This creates a single-shard, single-node MemoryDB cluster with TLS encryption, suitable for development and testing.

## Configuration Reference

### Engine

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `engine` | `string` | Yes | `"redis"` or `"valkey"` |
| `engineVersion` | `string` | No | Engine version (e.g., `"7.1"`, `"7.0"`, `"6.2"`) |
| `description` | `string` | No | Human-readable cluster description |

### Node Configuration

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `nodeType` | `string` | Yes | — | Instance type (e.g., `"db.t4g.small"`, `"db.r7g.large"`) |
| `port` | `int32` | No | `6379` | Connection port. ForceNew. |

### Topology

MemoryDB always uses a sharded architecture. Each shard holds a portion of the keyspace and has one primary plus zero or more replicas.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `numShards` | `int32` | `1` | Number of shards (data partitions). Min: 1. |
| `numReplicasPerShard` | `int32` | `1` | Replicas per shard (0–5). 0 means no replicas. |

### Authentication

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `aclName` | `string` | `"open-access"` | MemoryDB ACL name. Required. `"open-access"` disables auth. |

When `tlsEnabled` is `false`, `aclName` must be `"open-access"` (AWS constraint).

### Networking

| Field | Type | Description |
|-------|------|-------------|
| `subnetIds` | `repeated StringValueOrRef` | VPC subnet IDs for subnet group creation. Can reference AwsVpc via `valueFrom`. |
| `securityGroupIds` | `repeated StringValueOrRef` | Security groups to attach. Can reference AwsSecurityGroup via `valueFrom`. |

### Encryption

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `tlsEnabled` | `bool` | `true` | Enable TLS for in-transit encryption. ForceNew. |
| `kmsKeyId` | `StringValueOrRef` | — | Customer-managed KMS key for at-rest encryption. ForceNew. Can reference AwsKmsKey. |

MemoryDB always encrypts data at rest. The `kmsKeyId` field optionally provides a customer-managed key; without it, the AWS-managed key is used.

### Maintenance and Snapshots

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `maintenanceWindow` | `string` | AWS-assigned | Weekly window (e.g., `"sun:05:00-sun:06:00"`). |
| `snapshotRetentionLimit` | `int32` | `0` | Days to retain snapshots (0–35). 0 disables. |
| `snapshotWindow` | `string` | AWS-assigned | Daily snapshot window (e.g., `"05:00-09:00"`). |
| `finalSnapshotName` | `string` | — | Final snapshot name on deletion. |

### Restore from Snapshot

| Field | Type | Description |
|-------|------|-------------|
| `snapshotArns` | `repeated string` | S3 ARNs of RDB files to restore from. ForceNew. Mutually exclusive with `snapshotName`. |
| `snapshotName` | `string` | Named snapshot to restore from. ForceNew. Mutually exclusive with `snapshotArns`. |

### Parameters

| Field | Type | Description |
|-------|------|-------------|
| `parameterGroupFamily` | `string` | Required with `parameters` (e.g., `"memorydb_redis7"`). |
| `parameters` | `repeated AwsMemorydbClusterParameter` | Name/value pairs for custom engine tuning. |

### Advanced

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `snsTopicArn` | `StringValueOrRef` | — | SNS topic for cluster events. Can reference AwsSnsTopic. |
| `autoMinorVersionUpgrade` | `bool` | `true` | Auto-apply minor version upgrades. |
| `dataTiering` | `bool` | `false` | Move cold data to SSD. db.r6gd.* node types only. ForceNew. |

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `cluster_endpoint_address` | `string` | Cluster endpoint DNS address |
| `cluster_endpoint_port` | `int32` | Cluster endpoint port |
| `cluster_arn` | `string` | ARN of the MemoryDB cluster |
| `cluster_name` | `string` | Name of the cluster |
| `engine_patch_version` | `string` | Actual engine patch version running |
| `subnet_group_name` | `string` | Created subnet group name (if applicable) |
| `parameter_group_name` | `string` | Created parameter group name (if applicable) |

## Related Components

- [AwsVpc](/docs/catalog/aws/vpc) — provides VPC subnets for cluster placement
- [AwsSecurityGroup](/docs/catalog/aws/security-group) — controls network access to MemoryDB endpoints
- [AwsKmsKey](/docs/catalog/aws/kms-key) — provides a customer-managed encryption key
- [AwsSnsTopic](/docs/catalog/aws/sns-topic) — receives cluster event notifications
- [AwsRedisElasticache](/docs/catalog/aws/redis-elasticache) — ephemeral caching alternative when durability is not required
