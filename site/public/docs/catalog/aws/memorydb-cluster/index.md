---
title: "MemoryDB Cluster"
description: "MemoryDB Cluster deployment documentation"
icon: "package"
order: 100
componentName: "awsmemorydbcluster"
---

# AWS MemoryDB Cluster

Deploys an Amazon MemoryDB cluster — a fully managed, Redis-compatible, durable in-memory database with microsecond reads, single-digit millisecond writes, and Multi-AZ durability via a distributed transaction log. The component provisions the cluster with optional subnet group and parameter group management, ACL-based authentication, and always-on encryption at rest.

## What Gets Created

When you deploy an AwsMemorydbCluster resource, OpenMCF provisions:

- **MemoryDB Cluster** — a `memorydb.Cluster` resource with configurable shards and replicas per shard, TLS encryption, ACL-based authentication, and optional data tiering for cost-efficient large datasets
- **Subnet Group** — created only when `subnetIds` are provided, placing cluster nodes in the specified VPC subnets
- **Parameter Group** — created only when `parameters` are provided with a `parameterGroupFamily`, enabling custom engine tuning (e.g., activedefrag, maxmemory-policy)

## Prerequisites

- **AWS credentials** configured via environment variables or OpenMCF provider config
- **VPC subnets** in at least two Availability Zones for production deployments
- **A security group** allowing inbound traffic on port 6379 (default) from your application instances
- **A MemoryDB ACL** — use the built-in `open-access` for development; create a custom ACL with users via AWS console/CLI for production authentication

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
  region: us-east-1
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

This creates a single-shard, single-node MemoryDB cluster with Redis 7.1, TLS encryption, and no authentication.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `region` | `string` | AWS region where the MemoryDB cluster will be created (e.g., `us-west-2`, `eu-west-1`). | Required; non-empty |
| `engine` | `string` | Cache engine: `"redis"` or `"valkey"` | Must be `redis` or `valkey` |
| `nodeType` | `string` | Instance type determining CPU, memory, and network capacity (e.g., `"db.t4g.small"`, `"db.r7g.large"`, `"db.r6gd.xlarge"` for data tiering) | Required, non-empty |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `engineVersion` | `string` | Provider default | Engine version (e.g., `"7.1"`, `"7.0"`, `"6.2"`) |
| `description` | `string` | — | Human-readable cluster description |
| `port` | `int32` | `6379` | Connection port. ForceNew. Range: 1–65535. |
| `numShards` | `int32` | `1` | Number of shards (data partitions). Min: 1. |
| `numReplicasPerShard` | `int32` | `1` | Replicas per shard. Range: 0–5. |
| `aclName` | `string` | `"open-access"` | MemoryDB ACL name. Must be `"open-access"` when `tlsEnabled` is `false`. |
| `subnetIds` | `StringValueOrRef[]` | `[]` | VPC subnet IDs for subnet group creation. Can reference AwsVpc via `valueFrom`. |
| `securityGroupIds` | `StringValueOrRef[]` | `[]` | Security groups to attach. Can reference AwsSecurityGroup via `valueFrom`. |
| `tlsEnabled` | `bool` | `true` | Enable TLS for in-transit encryption. ForceNew. |
| `kmsKeyId` | `StringValueOrRef` | — | Customer-managed KMS key ARN for at-rest encryption. ForceNew. Can reference AwsKmsKey. |
| `maintenanceWindow` | `string` | AWS-assigned | Weekly window in UTC: `"ddd:hh24:mi-ddd:hh24:mi"`. |
| `snapshotRetentionLimit` | `int32` | `0` | Days to retain automatic snapshots (0–35). 0 disables. |
| `snapshotWindow` | `string` | AWS-assigned | Daily snapshot window in UTC: `"hh24:mi-hh24:mi"`. |
| `finalSnapshotName` | `string` | — | Final snapshot name on cluster deletion. |
| `snapshotArns` | `string[]` | `[]` | S3 ARNs of RDB files to restore from. ForceNew. Mutually exclusive with `snapshotName`. |
| `snapshotName` | `string` | — | Named snapshot to restore from. ForceNew. Mutually exclusive with `snapshotArns`. |
| `parameterGroupFamily` | `string` | — | Required when `parameters` are provided (e.g., `"memorydb_redis7"`). |
| `parameters` | `AwsMemorydbClusterParameter[]` | `[]` | Name/value pairs for engine parameter tuning. |
| `snsTopicArn` | `StringValueOrRef` | — | SNS topic for cluster event notifications. Can reference AwsSnsTopic. |
| `autoMinorVersionUpgrade` | `bool` | `true` | Automatically apply minor engine version upgrades. |
| `dataTiering` | `bool` | `false` | Move cold data to SSD. Only available on `db.r6gd.*` node types. ForceNew. |

## Examples

### Development Cluster

A minimal single-node cluster for local development:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsMemorydbCluster
metadata:
  name: dev-memorydb
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AwsMemorydbCluster.dev-memorydb
spec:
  region: us-east-1
  engine: redis
  engineVersion: "7.1"
  nodeType: db.t4g.small
  numShards: 1
  numReplicasPerShard: 0
  aclName: open-access
```

### Production HA with VPC and Snapshots

Multi-shard cluster with replicas, custom ACL, VPC placement, and daily snapshots:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsMemorydbCluster
metadata:
  name: session-store
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsMemorydbCluster.session-store
spec:
  region: us-east-1
  engine: redis
  engineVersion: "7.1"
  description: Production session store
  nodeType: db.r7g.large
  numShards: 2
  numReplicasPerShard: 2
  aclName: prod-acl
  subnetIds:
    - subnet-0a1b2c3d4e5f00001
    - subnet-0a1b2c3d4e5f00002
  securityGroupIds:
    - sg-0123456789abcdef0
  snapshotRetentionLimit: 7
  snapshotWindow: "03:00-04:00"
  maintenanceWindow: "sun:05:00-sun:06:00"
  parameterGroupFamily: memorydb_redis7
  parameters:
    - name: activedefrag
      value: "yes"
```

### Full-Featured with Foreign Key References

Production cluster using cross-resource references for VPC, security group, KMS, and SNS:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsMemorydbCluster
metadata:
  name: analytics-store
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsMemorydbCluster.analytics-store
spec:
  region: us-east-1
  engine: redis
  engineVersion: "7.1"
  description: High-throughput analytics store
  nodeType: db.r6gd.xlarge
  numShards: 4
  numReplicasPerShard: 2
  aclName: analytics-acl
  dataTiering: true
  subnetIds:
    - valueFrom:
        kind: AwsSubnet
        name: main-private-subnet-a
        fieldPath: status.outputs.subnet_id
    - valueFrom:
        kind: AwsSubnet
        name: main-private-subnet-b
        fieldPath: status.outputs.subnet_id
  securityGroupIds:
    - valueFrom:
        kind: AwsSecurityGroup
        name: memorydb-sg
        fieldPath: status.outputs.security_group_id
  kmsKeyId:
    valueFrom:
      kind: AwsKmsKey
      name: memorydb-key
      fieldPath: status.outputs.key_arn
  snapshotRetentionLimit: 14
  snapshotWindow: "02:00-03:00"
  maintenanceWindow: "wed:04:00-wed:05:00"
  parameterGroupFamily: memorydb_redis7
  parameters:
    - name: activedefrag
      value: "yes"
    - name: maxmemory-policy
      value: volatile-lru
  snsTopicArn:
    valueFrom:
      kind: AwsSnsTopic
      name: infra-alerts
      fieldPath: status.outputs.topic_arn
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `cluster_endpoint_address` | `string` | DNS address of the cluster endpoint for client connections |
| `cluster_endpoint_port` | `int32` | Port of the cluster endpoint |
| `cluster_arn` | `string` | ARN of the MemoryDB cluster |
| `cluster_name` | `string` | Name of the cluster (matches `metadata.id`) |
| `engine_patch_version` | `string` | Actual engine patch version running on the cluster |
| `subnet_group_name` | `string` | Name of the created subnet group (empty if not created) |
| `parameter_group_name` | `string` | Name of the created parameter group (empty if not created) |

## Related Components

- [AwsVpc](/docs/catalog/aws/vpc) — provides VPC subnets for cluster placement
- [AwsSecurityGroup](/docs/catalog/aws/security-group) — controls network access to MemoryDB endpoints
- [AwsKmsKey](/docs/catalog/aws/kms-key) — provides a customer-managed key for at-rest encryption
- [AwsSnsTopic](/docs/catalog/aws/sns-topic) — receives cluster event notifications
- [AwsRedisElasticache](/docs/catalog/aws/redis-elasticache) — ephemeral caching alternative when durability is not needed
