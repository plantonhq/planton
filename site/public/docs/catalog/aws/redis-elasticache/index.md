---
title: "Redis ElastiCache"
description: "Redis ElastiCache deployment documentation"
icon: "package"
order: 100
componentName: "awsrediselasticache"
---

# AWS Redis ElastiCache

Deploys an AWS ElastiCache replication group running Redis or Valkey, supporting both non-clustered mode (single primary with up to 5 read replicas) and clustered mode (data partitioned across multiple shards with optional replicas per shard). The component manages an optional subnet group, an optional custom parameter group, encryption, authentication, logging, and snapshot configuration.

## What Gets Created

When you deploy an AwsRedisElasticache resource, OpenMCF provisions:

- **ElastiCache Replication Group** — an `aws_elasticache_replication_group` running Redis or Valkey with the specified topology, node type, and engine version
- **Subnet Group** — created only when `subnetIds` are provided, places cluster nodes in the specified VPC subnets
- **Custom Parameter Group** — created only when `parameters` and `parameterGroupFamily` are provided, applies engine parameter overrides (e.g., `maxmemory-policy`, `timeout`)

## Prerequisites

- **AWS credentials** configured via environment variables or OpenMCF provider config
- **VPC subnets** for in-VPC deployments — provide at least two subnets in different Availability Zones for multi-AZ
- **A security group** allowing inbound traffic on the Redis port (default 6379)
- **A KMS key** if using customer-managed encryption at rest
- **An ACM certificate or TLS-capable client** if enabling transit encryption

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
  region: us-west-2
  engine: redis
  engineVersion: "7.1"
  description: Development Redis cache
  nodeType: cache.t3.micro
  numCacheClusters: 1
  subnetIds:
    - subnet-0a1b2c3d4e5f00001
    - subnet-0a1b2c3d4e5f00002
  securityGroupIds:
    - sg-0a1b2c3d4e5f00001
```

Deploy:

```shell
openmcf apply -f redis.yaml
```

This creates a single-node Redis 7.1 cluster (non-clustered mode) in the specified subnets.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `region` | `string` | AWS region where the ElastiCache cluster will be created. Example: `us-west-2`, `eu-west-1`. | Required |
| `engine` | `string` | Cache engine. Values: `redis`, `valkey`. | Must be `redis` or `valkey` |
| `description` | `string` | Human-readable description for the replication group. | Required by AWS |
| `nodeType` | `string` | ElastiCache node type determining CPU, memory, and network capacity. Examples: `cache.t3.micro`, `cache.r7g.large`, `cache.r6gd.xlarge`. | Required |
| `numCacheClusters` | `int` | Total node count (primary + replicas) for non-clustered mode. Mutually exclusive with `numNodeGroups`. | 1–6 when set |
| `numNodeGroups` | `int` | Shard count for clustered mode. Mutually exclusive with `numCacheClusters`. | Must be > 0 when set |

Exactly one of `numCacheClusters` or `numNodeGroups` must be provided to select the topology mode.

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `engineVersion` | `string` | Provider default | Engine version. Examples: `7.1`, `7.0`, `6.2` for Redis; `7.2` for Valkey. |
| `port` | `int` | `6379` | Port for client connections. **ForceNew** — changing this destroys and recreates the cluster. Range: 1–65535. |
| `replicasPerNodeGroup` | `int` | `0` | Read replicas per shard. Only valid when `numNodeGroups` is set. Range: 0–5. |
| `automaticFailoverEnabled` | `bool` | `false` | Promote a replica to primary on failure. Requires `numCacheClusters >= 2` or clustered mode. |
| `multiAzEnabled` | `bool` | `false` | Spread replicas across Availability Zones. Requires `automaticFailoverEnabled` to be `true`. |
| `subnetIds` | `StringValueOrRef[]` | `[]` | Subnet IDs for the ElastiCache subnet group. Provide subnets in ≥ 2 AZs for multi-AZ. Can reference `AwsVpc` via `valueFrom`. |
| `securityGroupIds` | `StringValueOrRef[]` | `[]` | VPC security groups attached to cluster nodes. Can reference `AwsSecurityGroup` via `valueFrom`. |
| `atRestEncryptionEnabled` | `bool` | `false` | Encrypt data on disk and in snapshots. **ForceNew** — changing this destroys and recreates the cluster. Recommended: `true`. |
| `transitEncryptionEnabled` | `bool` | `false` | Encrypt all client and replication traffic with TLS. Recommended: `true`. |
| `transitEncryptionMode` | `string` | — | TLS enforcement mode. Values: `preferred` (allows non-TLS), `required` (TLS only). Requires `transitEncryptionEnabled`. |
| `kmsKeyId` | `StringValueOrRef` | — | Customer-managed KMS key ARN for at-rest encryption. **ForceNew**. Can reference `AwsKmsKey` via `valueFrom`. |
| `authToken` | `StringValueOrRef` | — | Redis AUTH password (16–128 printable chars). Requires `transitEncryptionEnabled`. Mutually exclusive with `userGroupIds`. |
| `userGroupIds` | `string[]` | `[]` | Redis ACL user group IDs for fine-grained access control. Mutually exclusive with `authToken`. |
| `maintenanceWindow` | `string` | AWS default | Weekly maintenance window in UTC. Format: `ddd:hh24:mi-ddd:hh24:mi`. Example: `sun:05:00-sun:06:00`. |
| `snapshotRetentionLimit` | `int` | `0` | Days to retain automatic snapshots. `0` disables snapshots. Range: 0–35. |
| `snapshotWindow` | `string` | AWS default | Daily snapshot window in UTC. Format: `hh24:mi-hh24:mi`. Example: `03:00-04:00`. |
| `finalSnapshotIdentifier` | `string` | — | Name for the final snapshot taken on deletion. If omitted, no final snapshot is created. |
| `applyImmediately` | `bool` | `false` | Apply changes immediately instead of during the next maintenance window. May cause brief downtime. |
| `parameterGroupFamily` | `string` | — | Parameter group family. Required when `parameters` is provided. Examples: `redis7`, `redis6.x`, `valkey7`. |
| `parameters` | `object[]` | `[]` | Custom cache parameters applied via a managed parameter group. |
| `parameters[].name` | `string` | — | Parameter name (e.g., `maxmemory-policy`, `timeout`). Required. |
| `parameters[].value` | `string` | — | Parameter value (e.g., `volatile-lru`, `300`). Required. |
| `logDeliveryConfigurations` | `object[]` | `[]` | Log delivery configs. At most 2 entries — one per log type. |
| `logDeliveryConfigurations[].destinationType` | `string` | — | Destination type. Values: `cloudwatch-logs`, `kinesis-firehose`. Required. |
| `logDeliveryConfigurations[].destination` | `StringValueOrRef` | — | Destination identifier (log group name or delivery stream name). Required. |
| `logDeliveryConfigurations[].logFormat` | `string` | — | Serialization format. Values: `text`, `json`. Required. |
| `logDeliveryConfigurations[].logType` | `string` | — | Log type. Values: `slow-log`, `engine-log`. Required. |
| `notificationTopicArn` | `StringValueOrRef` | — | SNS topic ARN for cluster event notifications. Can reference `AwsSnsTopic` via `valueFrom`. |
| `autoMinorVersionUpgrade` | `bool` | `false` | Automatically apply minor engine version upgrades during maintenance windows. |
| `dataTieringEnabled` | `bool` | `false` | Move less-frequently-accessed data to SSD. Only on `r6gd` node types. **ForceNew**. |

## Examples

### Non-Clustered with Encryption and Failover

A 3-node Redis cluster (1 primary + 2 replicas) with encryption and automatic failover across multiple AZs:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsRedisElasticache
metadata:
  name: session-cache
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsRedisElasticache.session-cache
spec:
  region: us-west-2
  engine: redis
  engineVersion: "7.1"
  description: Session cache with HA
  nodeType: cache.r7g.large
  numCacheClusters: 3
  automaticFailoverEnabled: true
  multiAzEnabled: true
  subnetIds:
    - subnet-private-az1
    - subnet-private-az2
    - subnet-private-az3
  securityGroupIds:
    - sg-redis-prod
  atRestEncryptionEnabled: true
  transitEncryptionEnabled: true
  transitEncryptionMode: required
  snapshotRetentionLimit: 7
  snapshotWindow: "03:00-04:00"
  maintenanceWindow: "sun:05:00-sun:06:00"
```

### Clustered Mode with Custom Parameters

A sharded Redis cluster with 3 shards and 2 replicas per shard, custom parameter overrides, and slow-log delivery to CloudWatch:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsRedisElasticache
metadata:
  name: analytics-cache
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsRedisElasticache.analytics-cache
spec:
  region: us-west-2
  engine: redis
  engineVersion: "7.1"
  description: Sharded analytics cache
  nodeType: cache.r7g.xlarge
  numNodeGroups: 3
  replicasPerNodeGroup: 2
  automaticFailoverEnabled: true
  multiAzEnabled: true
  subnetIds:
    - subnet-private-az1
    - subnet-private-az2
    - subnet-private-az3
  securityGroupIds:
    - sg-redis-analytics
  atRestEncryptionEnabled: true
  transitEncryptionEnabled: true
  parameterGroupFamily: redis7
  parameters:
    - name: maxmemory-policy
      value: volatile-lru
    - name: timeout
      value: "300"
  logDeliveryConfigurations:
    - destinationType: cloudwatch-logs
      destination: /aws/elasticache/analytics-cache/slow-log
      logFormat: json
      logType: slow-log
  applyImmediately: true
```

### Valkey with Data Tiering and Foreign Key References

A Valkey cluster using `r6gd` nodes for data tiering, referencing other OpenMCF-managed resources:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsRedisElasticache
metadata:
  name: tiered-cache
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsRedisElasticache.tiered-cache
spec:
  region: us-west-2
  engine: valkey
  engineVersion: "7.2"
  description: Valkey cache with data tiering
  nodeType: cache.r6gd.xlarge
  numCacheClusters: 3
  automaticFailoverEnabled: true
  multiAzEnabled: true
  dataTieringEnabled: true
  subnetIds:
    - valueFrom:
        kind: AwsVpc
        name: main-vpc
        field: status.outputs.private_subnets[0].id
    - valueFrom:
        kind: AwsVpc
        name: main-vpc
        field: status.outputs.private_subnets[1].id
  securityGroupIds:
    - valueFrom:
        kind: AwsSecurityGroup
        name: redis-sg
        field: status.outputs.security_group_id
  atRestEncryptionEnabled: true
  kmsKeyId:
    valueFrom:
      kind: AwsKmsKey
      name: data-key
      field: status.outputs.key_arn
  transitEncryptionEnabled: true
  transitEncryptionMode: required
  notificationTopicArn:
    valueFrom:
      kind: AwsSnsTopic
      name: infra-alerts
      field: status.outputs.topic_arn
  snapshotRetentionLimit: 14
  finalSnapshotIdentifier: tiered-cache-final
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `replication_group_id` | `string` | Identifier of the replication group, used in AWS CLI/API calls |
| `primary_endpoint_address` | `string` | Primary (writer) endpoint DNS name for read-write operations |
| `reader_endpoint_address` | `string` | Reader endpoint DNS name distributing reads across replicas. Empty for single-node deployments. |
| `configuration_endpoint_address` | `string` | Configuration endpoint for Cluster Mode Enabled clients. Empty when Cluster Mode is disabled. |
| `arn` | `string` | Amazon Resource Name of the replication group |
| `port` | `int` | Port on which the cluster accepts connections |
| `subnet_group_name` | `string` | Name of the created subnet group. Only populated when `subnetIds` were provided. |
| `parameter_group_name` | `string` | Name of the created parameter group. Only populated when `parameters` were provided. |

## Related Components

- [AwsVpc](/docs/catalog/aws/vpc) — provides subnets for cluster placement
- [AwsSecurityGroup](/docs/catalog/aws/security-group) — controls network-level access to the Redis/Valkey endpoint
- [AwsKmsKey](/docs/catalog/aws/kms-key) — provides a customer-managed key for at-rest encryption
- [AwsSnsTopic](/docs/catalog/aws/sns-topic) — receives cluster event notifications
