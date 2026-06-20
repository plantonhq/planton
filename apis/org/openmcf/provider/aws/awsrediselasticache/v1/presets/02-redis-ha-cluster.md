# Redis HA Cluster

This preset creates a 3-node Redis 7.1 cluster (1 primary + 2 read replicas) with automatic failover, multi-AZ deployment, encryption, daily snapshots, and a tuned eviction policy. Production-ready for most caching workloads.

## When to Use

- Session stores for web applications requiring sub-millisecond reads
- Application caches where data loss on node failure is unacceptable
- Rate limiting and leaderboard systems needing consistent availability
- Any production workload that needs HA without horizontal sharding

## Key Configuration Choices

- **3 nodes** (`numCacheClusters: 3`) — 1 primary + 2 replicas provides read scaling and failover capacity
- **Automatic failover** — if the primary fails, ElastiCache promotes a replica within seconds
- **Multi-AZ** — replicas are spread across Availability Zones for AZ-level resilience
- **TLS required** (`transitEncryptionMode: required`) — all connections must use TLS; no plaintext fallback
- **Daily snapshots** — 7-day retention with a 03:00–04:00 UTC window (low-traffic period)
- **volatile-lru eviction** — evicts least-recently-used keys with a TTL first, preserving persistent keys

## Placeholders to Replace

| Placeholder | Description | Example |
|-------------|-------------|---------|
| `<private-subnet-id-az1>` | Private subnet in first AZ | `subnet-0a1b2c3d4e5f6g7h8` |
| `<private-subnet-id-az2>` | Private subnet in second AZ | `subnet-1a2b3c4d5e6f7g8h9` |
| `<security-group-id>` | SG allowing port 6379 from app instances | `sg-0123456789abcdef0` |

Alternatively, replace literal values with `valueFrom` references:

```yaml
subnetIds:
  - valueFrom:
      kind: AwsSubnet
      name: my-private-subnet-a
      fieldPath: status.outputs.subnet_id
  - valueFrom:
      kind: AwsSubnet
      name: my-private-subnet-b
      fieldPath: status.outputs.subnet_id
```

## Common Additions

- Add `kmsKeyId` for customer-managed encryption keys
- Add `authToken` for Redis AUTH password protection
- Add `logDeliveryConfigurations` for slow-log monitoring
- Add `notificationTopicArn` for SNS alerts on failover and maintenance events

## Related Presets

- **01-redis-single-node** — minimal dev/test setup
- **03-redis-clustered-production** — horizontally sharded cluster for large-scale workloads
