# Production HA

This preset creates a 2-shard MemoryDB cluster with 2 replicas per shard (6 total nodes), daily snapshots, a tuned eviction policy, and active defragmentation. Production-ready for session stores, user profiles, and other latency-sensitive durable workloads.

## When to Use

- Session stores for web applications requiring sub-millisecond reads with durability
- User profile stores, shopping carts, and leaderboards
- Rate limiting and throttling systems with persistent state
- Any production workload needing both Redis performance and database-grade durability

## Key Configuration Choices

- **2 shards x 2 replicas** (`numShards: 2`, `numReplicasPerShard: 2`) — data partitioned across 2 shards with 2 read replicas each; total 6 nodes for throughput and resilience
- **db.r7g.large** — memory-optimized Graviton3 instance; 13.07 GiB per node
- **Custom ACL** — replace `<acl-name>` with your MemoryDB ACL for fine-grained authentication
- **Daily snapshots** — 7-day retention with a 03:00–04:00 UTC window
- **Active defragmentation** — reclaims memory from fragmented allocations automatically
- **volatile-lru eviction** — evicts least-recently-used keys with a TTL first, preserving persistent keys

## Placeholders to Replace

| Placeholder | Description | Example |
|-------------|-------------|---------|
| `<acl-name>` | MemoryDB ACL with configured users | `my-prod-acl` |
| `<private-subnet-id-az1>` | Private subnet in first AZ | `subnet-0a1b2c3d4e5f6g7h8` |
| `<private-subnet-id-az2>` | Private subnet in second AZ | `subnet-1a2b3c4d5e6f7g8h9` |
| `<security-group-id>` | SG allowing port 6379 from app instances | `sg-0123456789abcdef0` |

Alternatively, replace literal values with `valueFrom` references:

```yaml
subnetIds:
  - valueFrom:
      kind: AwsVpc
      name: my-vpc
      fieldPath: status.outputs.private_subnets.[0].id
  - valueFrom:
      kind: AwsVpc
      name: my-vpc
      fieldPath: status.outputs.private_subnets.[1].id
```

## Common Additions

- Add `kmsKeyId` for customer-managed encryption keys
- Add `snsTopicArn` for SNS alerts on cluster events
- Increase `numShards` to 4+ for higher write throughput
- Add more `parameters` for advanced Redis tuning (timeout, tcp-keepalive)

## Related Presets

- **01-dev-single-shard** — minimal dev/test setup with open-access ACL
- **03-high-throughput** — large-scale cluster with data tiering for massive datasets
