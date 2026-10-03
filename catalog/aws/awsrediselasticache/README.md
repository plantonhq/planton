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

When you deploy an AwsRedisElasticache resource, Planton provisions:

- **Replication Group** — a Redis or Valkey cluster in either non-clustered mode (1 primary + up to 5 read replicas) or clustered mode (multiple shards with configurable replicas per shard). The AWS identifier is `metadata.name` (create-time immutable).
- **Subnet Group** — created automatically when `subnetIds` are provided, or use `subnetGroupName` to attach an existing subnet group
- **Parameter Group** — created automatically when `parameters` are provided with a `parameterGroupFamily`, or use `parameterGroupName` to attach an existing parameter group

## Prerequisites

- **AWS credentials** configured via environment variables or Planton provider config
- **VPC subnets** in at least two Availability Zones for multi-AZ deployments
- **Security group** allowing inbound traffic on the Redis port (default 6379) from your application instances

## Quick Start

Create a file `redis.yaml`:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsRedisElasticache
metadata:
  name: my-redis
  annotations:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.AwsRedisElasticache.my-redis
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
planton apply -f redis.yaml
```

This creates a single-node Redis 7.1 cluster with encryption at rest and in transit.

## Configuration Reference

### Engine

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `engine` | `string` | Yes* | `"redis"` or `"valkey"`. Leave empty when joining a global datastore (inherited from primary). |
| `engineVersion` | `string` | No | Engine version (e.g., `"7.1"`, `"7.0"`, `"6.2"`). Leave empty when joining a global datastore. |
| `description` | `string` | Yes | Human-readable description for the replication group |

### Node Configuration

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `nodeType` | `string` | Yes* | — | Instance type (e.g., `"cache.t3.micro"`, `"cache.r7g.large"`). Leave empty when joining a global datastore. |
| `port` | `int32` | No | `6379` | Connection port. ForceNew. |

### Topology — Non-Clustered Mode

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `numCacheClusters` | `int32` | — | Total nodes (primary + replicas). Range: 1–6. Mutually exclusive with `numNodeGroups`. |
| `preferredCacheClusterAzs` | `repeated string` | — | AZ placement per cache cluster (non-clustered). Length must match `numCacheClusters`. Mutually exclusive with `nodeGroupConfigurations`. |

### Topology — Clustered Mode

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `numNodeGroups` | `int32` | — | Number of shards. Mutually exclusive with `numCacheClusters`. |
| `replicasPerNodeGroup` | `int32` | — | Replicas per shard (0–5). Requires `numNodeGroups`. |
| `nodeGroupConfigurations` | `repeated object` | — | Per-shard AZ placement, replica count, and slot assignment. Requires `numNodeGroups`. |

### High Availability

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `automaticFailoverEnabled` | `bool` | `false` | Auto-failover to replica. Requires multi-node topology. |
| `multiAzEnabled` | `bool` | `false` | Spread replicas across AZs. Requires failover. |
| `durability` | `string` | — | Write-acknowledgement mode: `default`, `async`, `sync`, `disabled`. Valkey 9.0+ clustered only. ForceNew. |

### Global Datastore

| Field | Type | Description |
|-------|------|-------------|
| `globalReplicationGroupId` | `string` | Join an existing global replication group as secondary. Inherits engine, topology, encryption, and parameters from primary. ForceNew. |

### Networking

| Field | Type | Description |
|-------|------|-------------|
| `subnetIds` | `repeated StringValueOrRef` | VPC subnet IDs for subnet group creation. Mutually exclusive with `subnetGroupName`. |
| `subnetGroupName` | `string` | Existing ElastiCache subnet group (bring-your-own). ForceNew. |
| `securityGroupIds` | `repeated StringValueOrRef` | Security groups to attach. Can reference AwsSecurityGroup via `valueFrom`. |
| `networkType` | `string` | IP addressing: `ipv4`, `ipv6`, `dual_stack`. ForceNew. |
| `ipDiscovery` | `string` | DNS discovery address family: `ipv4` or `ipv6`. Meaningful with dual-stack. |

### Encryption

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `atRestEncryptionEnabled` | `bool` | Unset (AWS default); recommended: `true` | Encrypt data on disk. Presence matters: leave unset to let AWS apply its engine default; must stay unset when joining a global datastore. ForceNew. |
| `transitEncryptionEnabled` | `bool` | Unset (AWS default: off); recommended: `true` | Encrypt data in transit (TLS). Presence matters: must stay unset when joining a global datastore. |
| `transitEncryptionMode` | `string` | — | `"preferred"` or `"required"`. Requires transit encryption. |
| `kmsKeyId` | `StringValueOrRef` | — | Customer-managed KMS key. ForceNew. Can reference AwsKmsKey. |

### Authentication

| Field | Type | Description |
|-------|------|-------------|
| `authToken` | `StringValueOrRef` | Redis AUTH password. Mutually exclusive with `userGroupIds`. Requires transit encryption. |
| `authTokenUpdateStrategy` | `string` | How token changes roll out: `ROTATE` and `SET` require `authToken`; `DELETE` removes AUTH entirely (the AUTH-to-ACL migration step) and requires `authToken` to be absent. |
| `userGroupIds` | `repeated StringValueOrRef` | Redis ACL user groups via `AwsElasticacheUserGroup`. Mutually exclusive with `authToken`. |

### Restore (Create-Time)

| Field | Type | Description |
|-------|------|-------------|
| `snapshotArns` | `repeated string` | S3 ARNs of RDB files to seed from. Mutually exclusive with `snapshotName`. ForceNew. |
| `snapshotName` | `string` | Existing ElastiCache snapshot to restore from. Mutually exclusive with `snapshotArns`. ForceNew. |

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
| `parameters` | `repeated AwsRedisElasticacheParameter` | Name/value pairs for custom engine tuning. Mutually exclusive with `parameterGroupName`. |
| `parameterGroupName` | `string` | Existing parameter group (bring-your-own). Cluster mode requires `.cluster.on` family. |

### Logging

| Field | Type | Description |
|-------|------|-------------|
| `logDeliveryConfigurations` | `repeated AwsRedisElasticacheLogDeliveryConfig` | Up to 2 entries (slow-log, engine-log). |

### Advanced

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `notificationTopicArn` | `StringValueOrRef` | — | SNS topic for cluster events. Can reference AwsSnsTopic. |
| `autoMinorVersionUpgrade` | `bool` | Unset (AWS default: on) | Auto-apply minor version upgrades. Leave unset to keep AWS's enabled-by-default; set `false` to pin the running minor version. |
| `dataTieringEnabled` | `bool` | `false` | Move cold data to SSD. r6gd node types only. ForceNew. |
| `clusterMode` | `string` | — | Migration setting: `enabled`, `compatible`, `disabled`. Online path from non-clustered to clustered. |

\* Required unless joining a global datastore via `globalReplicationGroupId`.

## Examples

### HA Non-Clustered with Encryption

```yaml
apiVersion: aws.planton.dev/v1alpha1
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
apiVersion: aws.planton.dev/v1alpha1
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

## Outputs

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

## Related Kinds

- [AwsVpc](/docs/catalog/aws/vpc) — provides VPC subnets for cluster placement
- [AwsSecurityGroup](/docs/catalog/aws/security-group) — controls network access to Redis endpoints
- [AwsKmsKey](/docs/catalog/aws/kms-key) — provides a customer-managed encryption key for at-rest encryption
- [AwsSnsTopic](/docs/catalog/aws/sns-topic) — receives cluster event notifications
- [AwsElasticacheUser](/docs/catalog/aws/awselasticacheuser) — RBAC identity (WHO / WHAT)
- [AwsElasticacheUserGroup](/docs/catalog/aws/awselasticacheusergroup) — collects users; referenced in `userGroupIds`

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
