# AwsRedisElasticache — Architecture and Design

## Overview

AwsRedisElasticache provisions a managed ElastiCache replication group running Redis or Valkey. The component abstracts the complexity of subnet groups, parameter groups, encryption, and topology selection into a single declarative resource while preserving full access to production-critical features.

## Why Redis/Valkey Only

ElastiCache supports three engines: Redis, Valkey, and Memcached. This component covers **Redis and Valkey only** because they share the same Terraform resource (`aws_elasticache_replication_group`) and identical configuration surface. Memcached uses a fundamentally different resource (`aws_elasticache_cluster`) with a different topology model (no replication, no persistence, no encryption at rest, no AUTH). A separate `AwsMemcachedElasticache` component covers Memcached.

## Topology Modes

### Non-Clustered (Cluster Mode Disabled)

Set `numCacheClusters` (1–6). One primary handles all writes; additional nodes are read replicas that receive asynchronous replication. The primary endpoint handles writes; the reader endpoint load-balances reads across replicas.

**Best for:** workloads under ~113 GB, read-heavy patterns, simple client configuration.

### Clustered (Cluster Mode Enabled)

Set `numNodeGroups` (shard count) and optionally `replicasPerNodeGroup` (0–5). Data is hash-slotted across shards. Each shard has its own primary and replicas. The configuration endpoint enables automatic slot discovery for cluster-aware clients.

**Best for:** multi-terabyte datasets, write-heavy patterns, horizontal scaling needs.

### Topology Selection CEL

The spec enforces exactly one topology via CEL: `(numCacheClusters > 0) != (numNodeGroups > 0)`. This XOR prevents ambiguous configurations.

## Bundled Resources

Following the AwsRdsCluster pattern, the IaC modules conditionally create supporting resources:

- **Subnet Group**: Created when `subnetIds` are provided. Name is sanitized from the resource metadata ID to meet AWS naming constraints (lowercase, alphanumeric, hyphens).
- **Parameter Group**: Created when `parameters` are provided with a `parameterGroupFamily`. Uses `name_prefix` to avoid naming collisions during updates.

## Encryption Architecture

- **At Rest**: Enabled via `atRestEncryptionEnabled`. Uses AWS-managed key by default; `kmsKeyId` overrides with a customer-managed KMS key. Both are ForceNew — changing them destroys and recreates the cluster.
- **In Transit**: Enabled via `transitEncryptionEnabled`. The `transitEncryptionMode` field controls enforcement: `"preferred"` allows both TLS and non-TLS connections (for migration); `"required"` enforces TLS for all connections.

## Authentication

Two mutually exclusive methods:

1. **AUTH Token** (`authToken`): A single password (16–128 chars) that all clients must provide. Simple but shared-secret model. Requires transit encryption.
2. **User Groups** (`userGroupIds`): Redis ACL with fine-grained user permissions (specific commands, key patterns). More secure for multi-tenant or multi-service access patterns.

## Infra Chart Composability

### Inputs (StringValueOrRef)

| Field | Default Reference |
|-------|-------------------|
| `subnetIds` | `AwsSubnet.status.outputs.subnet_id` |
| `securityGroupIds` | `AwsSecurityGroup.status.outputs.security_group_id` |
| `kmsKeyId` | `AwsKmsKey.status.outputs.key_arn` |
| `notificationTopicArn` | `AwsSnsTopic.status.outputs.topic_arn` |

### Outputs (for downstream)

| Output | Downstream Use |
|--------|---------------|
| `primary_endpoint_address` + `port` | Application connection config |
| `reader_endpoint_address` | Read-only connection pool |
| `configuration_endpoint_address` | Cluster-aware client config |
| `arn` | IAM policies, metric dimensions |

### Typical DAG Position

```
Layer 0: AwsVpc
Layer 1: AwsSecurityGroup, AwsKmsKey
Layer 2: AwsRedisElasticache  ← this component
Layer 3: Application configs, AwsLambda event triggers
```

## Deliberately Omitted (v1)

| Feature | Reason |
|---------|--------|
| Global Replication Group | Cross-region, conflicts with 10+ fields, <5% of deployments |
| Node Group Configuration | Per-shard AZ/slot tuning — AWS auto-distributes |
| ElastiCache Serverless | Fundamentally different resource and config model |
| Memcached | Different TF resource, different topology, no replication |
| Restore from snapshot | `snapshot_arns`/`snapshot_name` can be added in v2 |
| IPv6 / dual-stack | `network_type`/`ip_discovery` — niche for v1 |

## References

- [AWS ElastiCache for Redis User Guide](https://docs.aws.amazon.com/AmazonElastiCache/latest/red-ug/WhatIs.html)
- [ElastiCache Replication Group API](https://docs.aws.amazon.com/AmazonElastiCache/latest/APIReference/API_CreateReplicationGroup.html)
- [Redis Cluster Mode Explained](https://docs.aws.amazon.com/AmazonElastiCache/latest/red-ug/Replication.Redis-RedisCluster.html)
- [ElastiCache Parameter Groups](https://docs.aws.amazon.com/AmazonElastiCache/latest/red-ug/ParameterGroups.html)
