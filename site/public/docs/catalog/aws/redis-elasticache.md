---
title: "Redis ElastiCache"
description: "Redis/Valkey ElastiCache deployment documentation"
icon: "package"
order: 100
componentName: "awsrediselasticache"
---

# AWS Redis ElastiCache

Deploys an AWS ElastiCache replication group running Redis or Valkey — a fully managed, sub-millisecond in-memory data store with automatic failover, encryption, snapshot persistence, and flexible topology options. Supports both non-clustered (single primary + read replicas) and clustered (sharded) modes for workloads ranging from simple session caches to multi-terabyte real-time data stores.

## What Gets Created

When you deploy an AwsRedisElasticache resource, OpenMCF provisions:

- **Replication Group** — a Redis or Valkey cluster in non-clustered mode (1 primary + up to 5 replicas) or clustered mode (multiple shards with replicas per shard)
- **Subnet Group** — created automatically when `subnetIds` are provided
- **Parameter Group** — created automatically when `parameters` are provided with a `parameterGroupFamily`

## Prerequisites

- **AWS credentials** configured via environment variables or OpenMCF provider config
- **VPC subnets** in at least two AZs for multi-AZ deployments
- **Security group** allowing inbound traffic on the Redis port (default 6379)

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

## Configuration Reference

### Engine

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `engine` | `string` | Yes | `"redis"` or `"valkey"` |
| `engineVersion` | `string` | No | Engine version (e.g., `"7.1"`) |
| `description` | `string` | Yes | Human-readable description |

### Topology

| Field | Type | Description |
|-------|------|-------------|
| `nodeType` | `string` | Instance type (e.g., `"cache.r7g.large"`) |
| `numCacheClusters` | `int32` | Total nodes for non-clustered mode (1–6) |
| `numNodeGroups` | `int32` | Shard count for clustered mode |
| `replicasPerNodeGroup` | `int32` | Replicas per shard (0–5) |

### High Availability

| Field | Type | Description |
|-------|------|-------------|
| `automaticFailoverEnabled` | `bool` | Auto-failover to replica on primary failure |
| `multiAzEnabled` | `bool` | Spread replicas across Availability Zones |

### Encryption

| Field | Type | Description |
|-------|------|-------------|
| `atRestEncryptionEnabled` | `bool` | Encrypt data on disk (ForceNew) |
| `transitEncryptionEnabled` | `bool` | Encrypt data in transit (TLS) |
| `transitEncryptionMode` | `string` | `"preferred"` or `"required"` |
| `kmsKeyId` | `StringValueOrRef` | Customer-managed KMS key (ForceNew) |

### Authentication

| Field | Type | Description |
|-------|------|-------------|
| `authToken` | `StringValueOrRef` | Redis AUTH password. Mutually exclusive with `userGroupIds`. |
| `userGroupIds` | `repeated string` | Redis ACL user group IDs. Mutually exclusive with `authToken`. |

## Examples

### HA with Encryption

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
        kind: AwsVpc
        name: my-vpc
        fieldPath: status.outputs.private_subnets.[0].id
    - valueFrom:
        kind: AwsVpc
        name: my-vpc
        fieldPath: status.outputs.private_subnets.[1].id
  securityGroupIds:
    - valueFrom:
        kind: AwsSecurityGroup
        name: redis-sg
        fieldPath: status.outputs.security_group_id
```

## Stack Outputs

| Output | Type | Description |
|--------|------|-------------|
| `replication_group_id` | `string` | Replication group identifier |
| `primary_endpoint_address` | `string` | Primary (writer) endpoint DNS |
| `reader_endpoint_address` | `string` | Reader endpoint for read replicas |
| `configuration_endpoint_address` | `string` | Cluster mode endpoint |
| `arn` | `string` | ARN of the replication group |
| `port` | `int32` | Connection port |

## Related Components

- [AwsVpc](/docs/catalog/aws/vpc) — provides VPC subnets for cluster placement
- [AwsSecurityGroup](/docs/catalog/aws/security-group) — controls network access to Redis endpoints
- [AwsKmsKey](/docs/catalog/aws/kms-key) — customer-managed encryption key
- [AwsSnsTopic](/docs/catalog/aws/sns-topic) — receives cluster event notifications
- [AwsMemcachedElasticache](/docs/catalog/aws/memcached-elasticache) — Memcached variant (coming soon)
