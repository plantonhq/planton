---
title: "Redis ElastiCache"
description: "Redis/Valkey ElastiCache deployment documentation"
icon: "package"
order: 100
componentName: "awsrediselasticache"
---

# AWS Redis ElastiCache

Deploys an AWS ElastiCache replication group running Redis or Valkey — a fully managed, sub-millisecond in-memory data store with automatic failover, encryption, snapshot persistence, and flexible topology options. ElastiCache is the managed caching layer for session stores, application caches, real-time leaderboards, and message brokers on AWS.

## What Gets Created

When you deploy an AwsRedisElasticache resource, OpenMCF provisions:

- **Replication Group** — a Redis or Valkey cluster in either non-clustered mode (1 primary + up to 5 read replicas) or clustered mode (multiple shards with configurable replicas per shard)
- **Subnet Group** — created automatically when `subnetIds` are provided, placing cluster nodes in the specified VPC subnets
- **Parameter Group** — created automatically when `parameters` are provided with a `parameterGroupFamily`, allowing custom engine tuning (e.g., maxmemory-policy, timeout)

## Prerequisites

- **AWS credentials** configured via environment variables or OpenMCF provider config
- **VPC subnets** in at least two Availability Zones for multi-AZ deployments
- **Security group** allowing inbound traffic on the Redis port (default 6379) from your application instances

## Quick Start

Create a file `redis.yaml`:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsRedisElasticache
metadata:
  name: my-redis
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AwsRedisElasticache.my-redis
spec:
  engine: redis
  engineVersion: "7.1"
  description: Application cache
  nodeType: cache.t3.micro
  numCacheClusters: 1
  atRestEncryptionEnabled: true
  transitEncryptionEnabled: true
```

Deploy:

```shell
openmcf apply -f redis.yaml
```

This creates a single-node Redis 7.1 cluster with encryption at rest and in transit.

## Configuration Reference

### Engine

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `engine` | `string` | Yes | `"redis"` or `"valkey"` |
| `engineVersion` | `string` | No | Engine version (e.g., `"7.1"`, `"7.0"`, `"6.2"`) |
| `description` | `string` | Yes | Human-readable description for the replication group |

### Node Configuration

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `nodeType` | `string` | Yes | — | Instance type (e.g., `"cache.t3.micro"`, `"cache.r7g.large"`) |
| `port` | `int32` | No | `6379` | Connection port. ForceNew. |

### Topology — Non-Clustered Mode

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `numCacheClusters` | `int32` | — | Total nodes (primary + replicas). Range: 1–6. Mutually exclusive with `numNodeGroups`. |

### Topology — Clustered Mode

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `numNodeGroups` | `int32` | — | Number of shards. Mutually exclusive with `numCacheClusters`. |
| `replicasPerNodeGroup` | `int32` | — | Replicas per shard (0–5). Requires `numNodeGroups`. |

### High Availability

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `automaticFailoverEnabled` | `bool` | `false` | Auto-failover to replica. Requires multi-node topology. |
| `multiAzEnabled` | `bool` | `false` | Spread replicas across AZs. Requires failover. |

### Networking

| Field | Type | Description |
|-------|------|-------------|
| `subnetIds` | `repeated StringValueOrRef` | VPC subnet IDs for subnet group creation. Can reference AwsVpc via `valueFrom`. |
| `securityGroupIds` | `repeated StringValueOrRef` | Security groups to attach. Can reference AwsSecurityGroup via `valueFrom`. |

### Encryption

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `atRestEncryptionEnabled` | `bool` | Recommended: `true` | Encrypt data on disk. ForceNew. |
| `transitEncryptionEnabled` | `bool` | Recommended: `true` | Encrypt data in transit (TLS). |
| `transitEncryptionMode` | `string` | — | `"preferred"` or `"required"`. Requires transit encryption. |
| `kmsKeyId` | `StringValueOrRef` | — | Customer-managed KMS key. ForceNew. Can reference AwsKmsKey. |

### Authentication

| Field | Type | Description |
|-------|------|-------------|
| `authToken` | `StringValueOrRef` | Redis AUTH password. Mutually exclusive with `userGroupIds`. |
| `userGroupIds` | `repeated string` | Redis ACL user group IDs. Mutually exclusive with `authToken`. |

### Maintenance and Snapshots

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `maintenanceWindow` | `string` | AWS-assigned | Weekly window (e.g., `"sun:05:00-sun:06:00"`). |
| `snapshotRetentionLimit` | `int32` | `0` | Days to retain snapshots (0–35). 0 disables. |
| `snapshotWindow` | `string` | AWS-assigned | Daily snapshot window (e.g., `"03:00-04:00"`). |
| `finalSnapshotIdentifier` | `string` | — | Final snapshot name on deletion. |
| `applyImmediately` | `bool` | `false` | Apply changes immediately vs. next maintenance window. |

### Parameters

| Field | Type | Description |
|-------|------|-------------|
| `parameterGroupFamily` | `string` | Required with `parameters` (e.g., `"redis7"`, `"valkey7"`). |
| `parameters` | `repeated AwsRedisElasticacheParameter` | Name/value pairs for custom engine tuning. |

### Logging

| Field | Type | Description |
|-------|------|-------------|
| `logDeliveryConfigurations` | `repeated AwsRedisElasticacheLogDeliveryConfig` | Up to 2 entries (slow-log, engine-log). |

### Advanced

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `notificationTopicArn` | `StringValueOrRef` | — | SNS topic for cluster events. Can reference AwsSnsTopic. |
| `autoMinorVersionUpgrade` | `bool` | `false` | Auto-apply minor version upgrades. |
| `dataTieringEnabled` | `bool` | `false` | Move cold data to SSD. r6gd node types only. ForceNew. |

## Examples

### HA Non-Clustered with Encryption

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsRedisElasticache
metadata:
  name: session-cache
spec:
  engine: redis
  engineVersion: "7.1"
  description: Session cache with HA
  nodeType: cache.r7g.large
  numCacheClusters: 3
  automaticFailoverEnabled: true
  multiAzEnabled: true
  atRestEncryptionEnabled: true
  transitEncryptionEnabled: true
  transitEncryptionMode: required
  subnetIds:
    - valueFrom:
        kind: AwsSubnet
        name: my-private-subnet-a
        fieldPath: status.outputs.subnet_id
    - valueFrom:
        kind: AwsSubnet
        name: my-private-subnet-b
        fieldPath: status.outputs.subnet_id
  securityGroupIds:
    - valueFrom:
        kind: AwsSecurityGroup
        name: redis-sg
        fieldPath: status.outputs.security_group_id
  snapshotRetentionLimit: 7
  snapshotWindow: "03:00-04:00"
  maintenanceWindow: "sun:05:00-sun:06:00"
```

### Clustered (Sharded) Production

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsRedisElasticache
metadata:
  name: product-catalog-cache
spec:
  engine: redis
  engineVersion: "7.1"
  description: Product catalog cache with sharding
  nodeType: cache.r7g.xlarge
  numNodeGroups: 3
  replicasPerNodeGroup: 2
  automaticFailoverEnabled: true
  multiAzEnabled: true
  atRestEncryptionEnabled: true
  transitEncryptionEnabled: true
  kmsKeyId:
    valueFrom:
      kind: AwsKmsKey
      name: redis-key
      fieldPath: status.outputs.key_arn
  parameterGroupFamily: redis7
  parameters:
    - name: maxmemory-policy
      value: volatile-lru
  logDeliveryConfigurations:
    - destinationType: cloudwatch-logs
      destination:
        value: /aws/elasticache/product-catalog
      logFormat: json
      logType: slow-log
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `replication_group_id` | `string` | Replication group identifier |
| `primary_endpoint_address` | `string` | Primary (writer) endpoint DNS |
| `reader_endpoint_address` | `string` | Reader endpoint for read replicas |
| `configuration_endpoint_address` | `string` | Cluster mode endpoint (empty if non-clustered) |
| `arn` | `string` | ARN of the replication group |
| `port` | `int32` | Connection port |
| `subnet_group_name` | `string` | Created subnet group name (if applicable) |
| `parameter_group_name` | `string` | Created parameter group name (if applicable) |

## Related Components

- [AwsVpc](/docs/catalog/aws/vpc) — provides VPC subnets for cluster placement
- [AwsSecurityGroup](/docs/catalog/aws/security-group) — controls network access to Redis endpoints
- [AwsKmsKey](/docs/catalog/aws/kms-key) — provides a customer-managed encryption key for at-rest encryption
- [AwsSnsTopic](/docs/catalog/aws/sns-topic) — receives cluster event notifications
