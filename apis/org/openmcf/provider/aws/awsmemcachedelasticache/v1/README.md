---
title: AwsMemcachedElasticache
description: Provision and manage AWS ElastiCache Memcached clusters with consistent, declarative configuration.
icon: aws
order: 252
componentName: AwsMemcachedElasticache
---

# AWS Memcached ElastiCache

Provision a fully managed, distributed Memcached cluster on AWS ElastiCache. Memcached is a high-performance, in-memory key-value store designed for simple caching workloads that demand sub-millisecond latency and horizontal scalability.

## When to Use Memcached vs Redis

| Criterion | Memcached | Redis (AwsRedisElasticache) |
|-----------|-----------|---------------------------|
| **Data model** | Simple key-value only | Rich data structures (strings, hashes, lists, sets, sorted sets, streams) |
| **Persistence** | None — volatile cache only | Snapshots, AOF, replication |
| **Replication** | None — each key lives on one node | Primary-replica with automatic failover |
| **Authentication** | None — security via network isolation | AUTH token or Redis ACL user groups |
| **Encryption at rest** | Not supported | Supported with optional KMS key |
| **Scaling** | Horizontal: 1–40 nodes | Vertical (resize) or horizontal (cluster mode sharding) |
| **Multi-threading** | Yes — efficient on multi-core instances | Single-threaded (per shard) |

**Choose Memcached when** you need a simple, high-throughput distributed cache and do not require persistence, replication, or authentication. Common use cases: session caching, HTML fragment caching, database query result caching, API response caching.

**Choose Redis when** you need data persistence, replication, failover, complex data structures, or authentication.

## What Gets Created

- **ElastiCache Cluster** — Memcached cluster with 1–40 nodes
- **Subnet Group** (conditional) — created when `subnetIds` are provided
- **Parameter Group** (conditional) — created when custom `parameters` are provided with a `parameterGroupFamily`

## Prerequisites

- An AWS account with ElastiCache permissions
- (Recommended) A VPC with private subnets and a security group configured for Memcached traffic (port 11211)

## Quick Start

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsMemcachedElasticache
metadata:
  name: my-cache
spec:
  engineVersion: "1.6.22"
  nodeType: cache.t3.micro
  numCacheNodes: 1
```

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `engineVersion` | string | Memcached engine version (e.g., `1.6.22`, `1.6.17`, `1.5.16`) |
| `nodeType` | string | ElastiCache node type (e.g., `cache.t3.micro`, `cache.r7g.large`) |

### Node Configuration

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `numCacheNodes` | int32 | 1 | Number of cache nodes (1–40) |
| `azMode` | string | — | AZ distribution: `single-az` or `cross-az` (cross-az requires > 1 node) |
| `port` | int32 | 11211 | Port for client connections (ForceNew) |
| `preferredAvailabilityZones` | list | — | AZ placement per node (length must match `numCacheNodes`) |

### Encryption

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `transitEncryptionEnabled` | bool | false | Enable TLS encryption in transit (requires engine >= 1.6.12) |

### Networking

| Field | Type | Description |
|-------|------|-------------|
| `subnetIds` | list (StringValueOrRef) | Subnet IDs for the ElastiCache subnet group |
| `securityGroupIds` | list (StringValueOrRef) | VPC security group IDs for access control |

### Parameters

| Field | Type | Description |
|-------|------|-------------|
| `parameterGroupFamily` | string | Parameter group family (e.g., `memcached1.6`). Required when `parameters` is set |
| `parameters` | list | Custom parameter name/value pairs |

### Maintenance

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `maintenanceWindow` | string | — | Weekly maintenance window (e.g., `sun:05:00-sun:06:00`) |
| `applyImmediately` | bool | false | Apply changes immediately instead of waiting for maintenance window |
| `autoMinorVersionUpgrade` | bool | — | Auto-upgrade minor engine versions during maintenance |

### Notifications

| Field | Type | Description |
|-------|------|-------------|
| `notificationTopicArn` | StringValueOrRef | SNS topic ARN for cluster event notifications |

## Stack Outputs

| Output | Type | Description |
|--------|------|-------------|
| `cluster_id` | string | ElastiCache cluster identifier |
| `cluster_address` | string | Auto-discovery DNS name (without port) |
| `configuration_endpoint` | string | Full endpoint (address:port) for client auto-discovery |
| `arn` | string | Cluster ARN |
| `port` | int32 | Connection port |
| `subnet_group_name` | string | Created subnet group name (if any) |
| `parameter_group_name` | string | Created parameter group name (if any) |

## Important Operational Notes

- **Node type changes force recreation** — Memcached does not support vertical scaling. Changing `nodeType` destroys the cluster and creates a new one, losing all cached data.
- **No data persistence** — Memcached is a volatile cache. All data is lost on node failure, restart, or scaling events.
- **Security is network-only** — Memcached has no authentication. Always deploy in a VPC with properly configured security groups to restrict access.
- **Transit encryption** — Only available on engine version 1.6.12+. Earlier versions will fail at the AWS API level if TLS is enabled.

## Related Components

- **AwsRedisElasticache** — Redis/Valkey caching with replication, persistence, and authentication
- **AwsVpc** — VPC for network isolation (referenced by `subnetIds`)
- **AwsSecurityGroup** — Security groups for access control (referenced by `securityGroupIds`)
- **AwsSnsTopic** — SNS topics for event notifications (referenced by `notificationTopicArn`)
