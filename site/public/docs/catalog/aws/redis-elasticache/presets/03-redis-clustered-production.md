---
title: "Redis Clustered Production"
description: "This preset creates a Cluster Mode Enabled Redis 7.1 deployment with 3 shards, 2 replicas per shard (9 total nodes), customer-managed KMS encryption, slow-log delivery to CloudWatch, SNS event..."
type: "preset"
rank: "03"
presetSlug: "03-redis-clustered-production"
componentSlug: "redis-elasticache"
componentTitle: "Redis ElastiCache"
provider: "aws"
icon: "package"
order: 3
---

# Redis Clustered Production

This preset creates a Cluster Mode Enabled Redis 7.1 deployment with 3 shards, 2 replicas per shard (9 total nodes), customer-managed KMS encryption, slow-log delivery to CloudWatch, SNS event notifications, and a 14-day snapshot retention policy. Designed for high-throughput, multi-terabyte production workloads.

## When to Use

- E-commerce product catalogs, inventory caches, and real-time pricing
- Large-scale session stores exceeding single-node memory limits (~113 GB)
- Write-heavy workloads benefiting from hash-slot partitioning across shards
- Applications using cluster-aware Redis clients (Lettuce, redis-py-cluster, ioredis)

## Key Configuration Choices

- **3 shards x 2 replicas** (`numNodeGroups: 3`, `replicasPerNodeGroup: 2`) — data partitioned across 3 primary nodes with 2 read replicas each; total 9 nodes for throughput and resilience
- **cache.r7g.xlarge** — memory-optimized Graviton3 instance; 26.32 GiB memory per node, ~237 GiB total usable
- **Customer-managed KMS** — encryption with your own key for compliance and key rotation control
- **Slow-log to CloudWatch** — JSON-formatted slow query logs for performance monitoring and alerting
- **SNS notifications** — real-time alerts on failover, maintenance, and configuration changes
- **14-day snapshot retention** — extended backup window for disaster recovery
- **Connection timeout** (`timeout: 300`) — automatically close idle connections after 5 minutes to reclaim resources

## Placeholders to Replace

| Placeholder | Description | Example |
|-------------|-------------|---------|
| `<kms-key-arn>` | Customer-managed KMS key ARN | `arn:aws:kms:us-east-1:123456789012:key/mrk-abc123` |
| `<private-subnet-id-az1>` | Private subnet AZ1 | `subnet-0a1b2c3d4e5f6g7h8` |
| `<private-subnet-id-az2>` | Private subnet AZ2 | `subnet-1a2b3c4d5e6f7g8h9` |
| `<private-subnet-id-az3>` | Private subnet AZ3 | `subnet-2a3b4c5d6e7f8g9h0` |
| `<security-group-id>` | SG allowing port 6379 | `sg-0123456789abcdef0` |
| `<cloudwatch-log-group-name>` | Log group for slow-log | `/aws/elasticache/my-redis-clustered` |
| `<sns-topic-arn>` | SNS topic for alerts | `arn:aws:sns:us-east-1:123456789012:infra-alerts` |

## Common Additions

- Add `userGroupIds` for Redis ACL multi-user authentication
- Add a second `logDeliveryConfigurations` entry for `engine-log`
- Add `dataTieringEnabled: true` with `cache.r6gd.*` nodes for 5x data per node
- Increase `numNodeGroups` for further horizontal scaling

## Related Presets

- **01-redis-single-node** — minimal dev/test setup
- **02-redis-ha-cluster** — non-clustered HA for simpler workloads under 113 GB
