# High-Throughput with Data Tiering

This preset creates a 4-shard MemoryDB cluster with 2 replicas per shard (12 total nodes), data tiering for cost-efficient cold data management, customer-managed KMS encryption, SNS notifications, and a 14-day snapshot retention. Designed for high-throughput workloads with large datasets where not all data is frequently accessed.

## When to Use

- Large-scale analytics stores, time-series data, and event logs
- E-commerce product catalogs with millions of items
- Session stores exceeding single-shard memory limits
- Workloads where 20-30% of data is "hot" and the rest is infrequently accessed

## Key Configuration Choices

- **4 shards x 2 replicas** (`numShards: 4`, `numReplicasPerShard: 2`) — 12 total nodes; data partitioned across 4 shards for write throughput, 2 replicas per shard for read scaling and resilience
- **db.r6gd.xlarge** — data tiering-capable Graviton2 instance with local NVMe SSD; 26.32 GiB memory + SSD storage per node
- **Data tiering** (`dataTiering: true`) — automatically moves less-frequently-accessed data to SSD, achieving up to 5x more data capacity per node at lower cost
- **Customer-managed KMS** — encryption with your own key for compliance and key rotation control
- **SNS notifications** — real-time alerts on cluster events
- **14-day snapshot retention** — extended backup window for disaster recovery
- **Connection timeout** (`timeout: 300`) — close idle connections after 5 minutes to reclaim resources

## Placeholders to Replace

| Placeholder | Description | Example |
|-------------|-------------|---------|
| `<acl-name>` | MemoryDB ACL with configured users | `analytics-acl` |
| `<kms-key-arn>` | Customer-managed KMS key ARN | `arn:aws:kms:us-east-1:123456789012:key/mrk-abc123` |
| `<private-subnet-id-az1>` | Private subnet AZ1 | `subnet-0a1b2c3d4e5f6g7h8` |
| `<private-subnet-id-az2>` | Private subnet AZ2 | `subnet-1a2b3c4d5e6f7g8h9` |
| `<private-subnet-id-az3>` | Private subnet AZ3 | `subnet-2a3b4c5d6e7f8g9h0` |
| `<security-group-id>` | SG allowing port 6379 | `sg-0123456789abcdef0` |
| `<sns-topic-arn>` | SNS topic for alerts | `arn:aws:sns:us-east-1:123456789012:infra-alerts` |

## Common Additions

- Increase `numShards` to 8+ for further horizontal scaling
- Add additional `parameters` for workload-specific tuning
- Add `finalSnapshotName` for on-delete backup protection
- Remove `dataTiering` and switch to `db.r7g.*` nodes if all data is frequently accessed

## Related Presets

- **01-dev-single-shard** — minimal dev/test setup
- **02-production-ha** — standard production setup without data tiering
