# AwsKinesisStream — Architecture Reference

## Overview

Amazon Kinesis Data Streams is a serverless streaming data service that captures, processes, and stores data streams at any scale. It sits at the core of real-time data architectures, serving as a durable, ordered, replayable event log.

## Data Model

### Streams, Shards, and Records

A Kinesis stream is composed of one or more **shards**. Each shard is an ordered sequence of data records with a unique identifier.

**Data records** consist of:

- **Partition key** — Determines which shard receives the record. Records with the same partition key go to the same shard, guaranteeing ordering within a partition.
- **Sequence number** — Assigned by Kinesis when the record is written. Unique within a shard and monotonically increasing.
- **Data blob** — The payload (up to 1 MiB by default, configurable up to 10 MiB). Kinesis treats data as opaque bytes.

### Shard Capacity

Each shard provides fixed throughput:

| Direction | Throughput | Records |
|-----------|-----------|---------|
| Write | 1 MB/s | 1,000 records/s |
| Read (shared) | 2 MB/s | 5 GetRecords calls/s |
| Read (enhanced fan-out) | 2 MB/s per consumer | Push-based (SubscribeToShard) |

Total stream capacity scales linearly with shard count.

## Capacity Modes

### PROVISIONED Mode

You explicitly set the shard count. Scaling requires an `UpdateShardCount` API call, which uses uniform scaling — AWS evenly redistributes data across the new shard count.

**Scaling constraints:**

- Can scale up to 2x the current shard count per operation
- Can scale down to 50% of the current shard count per operation
- Maximum 10 scaling operations per rolling 24-hour period
- New shards take a few minutes to become active

**Cost model:** Per shard-hour. Predictable monthly cost = shard_count × hours × price_per_shard_hour.

### ON_DEMAND Mode

AWS automatically manages shards. The service scales to accommodate up to 200 MB/s write and 400 MB/s read throughput without any manual intervention.

**Scaling behavior:**

- Scales up within seconds in response to sustained throughput increase
- Scales down gradually (conservative — avoids oscillation)
- You do not see or control the shard count directly

**Cost model:** Per GB of data written and per GB of data read. More cost-effective for bursty or unpredictable workloads.

### Mode Switching

You can switch between PROVISIONED and ON_DEMAND at any time (not ForceNew). When switching from ON_DEMAND to PROVISIONED, the current shard count is preserved and you take over management.

## Data Retention

Records are accessible for a configurable **retention period** after being written:

- **Minimum**: 24 hours (default)
- **Maximum**: 8,760 hours (365 days)

Retention can be increased or decreased at any time. Extended retention incurs additional per-shard-hour cost.

**Use cases for extended retention:**

- **Reprocessing**: Replay events from a specific timestamp after fixing a consumer bug
- **Late-arriving consumers**: New applications can read historical data
- **Compliance**: Regulatory requirements for data availability windows

## Encryption

### Server-Side Encryption (SSE)

Kinesis supports transparent server-side encryption using AWS KMS:

- **No encryption** (`encryption_type = "NONE"`) — Data stored in plaintext
- **KMS encryption** (`encryption_type = "KMS"`) — Data encrypted at rest using a KMS key

When using KMS encryption:

- **Customer-managed key**: Full control over key rotation, policies, and deletion
- **Kinesis-owned key** (`alias/aws/kinesis`): Managed by AWS, no additional KMS cost
- Encryption/decryption is transparent to producers and consumers (no code changes)
- Encryption can be enabled or disabled after stream creation (not ForceNew)

**Important**: KMS encryption adds latency (~100-200μs per operation) and incurs KMS API costs.

## Enhanced Monitoring

By default, Kinesis provides **stream-level** CloudWatch metrics (aggregated across all shards). For production workloads, you can enable **shard-level** metrics for granular visibility.

### Available Shard-Level Metrics

| Metric | Description | Key Use |
|--------|-------------|---------|
| `IncomingBytes` | Bytes written per shard per period | Capacity planning |
| `IncomingRecords` | Records written per shard per period | Throughput analysis |
| `OutgoingBytes` | Bytes read per shard per period | Consumer throughput |
| `OutgoingRecords` | Records read per shard per period | Consumer throughput |
| `WriteProvisionedThroughputExceeded` | Write throttle events per shard | Hot shard detection |
| `ReadProvisionedThroughputExceeded` | Read throttle events per shard | Consumer bottleneck |
| `IteratorAgeMilliseconds` | Age of the oldest record in a shard's iterator | Consumer lag (critical) |

### Critical Metric: IteratorAgeMilliseconds

This is the most important metric for operational monitoring. It measures how far behind a consumer is from the head of the stream. A rising `IteratorAgeMilliseconds` indicates:

- Consumer is not keeping up with the write rate
- Consumer has crashed and is not processing
- Shard is "hot" (too much data for a single consumer)

**Recommended alarm**: `IteratorAgeMilliseconds > retention_period * 0.5` — consumer is at risk of losing data.

## Consumer Model

### Standard Consumers (GetRecords)

- Pull-based: Consumers call `GetRecords` to fetch data
- **Shared 2 MB/s per shard** across all consumers
- Up to 5 `GetRecords` calls per shard per second
- Limited to 5 registered consumers per stream (soft limit)
- Best for low consumer counts (1-3)

### Enhanced Fan-Out Consumers (SubscribeToShard)

- Push-based: Kinesis pushes data to consumers via HTTP/2
- **Dedicated 2 MB/s per shard per consumer** (not shared)
- Up to 20 registered consumers per stream
- ~70ms propagation delay (vs ~200ms for polling)
- Additional cost per consumer-shard-hour

Enhanced fan-out consumers are managed as separate resources (`aws_kinesis_stream_consumer`), which is why they are not bundled into this component.

## Security Best Practices

1. **Enable KMS encryption** for sensitive data streams (PII, financial, healthcare)
2. **Use IAM policies** to control who can write to and read from the stream
3. **Enable CloudTrail** for Kinesis API audit logging
4. **Use VPC endpoints** to keep Kinesis traffic within your VPC (no internet transit)
5. **Rotate KMS keys** using automatic key rotation (annually)

## Cost Considerations

### PROVISIONED Mode

- Shard-hour: ~$0.015/shard/hour (~$10.80/shard/month)
- Extended retention: ~$0.020/shard/hour for hours 25-168, ~$0.013/shard/hour for hours 169+
- Enhanced fan-out: ~$0.015/consumer-shard/hour

### ON_DEMAND Mode

- Data written: ~$0.08/GB
- Data read: ~$0.04/GB
- Extended retention: Same as provisioned

### Cost Optimization Tips

- Use ON_DEMAND for <10 MB/s or bursty workloads
- Use PROVISIONED for steady >10 MB/s (often cheaper)
- Set retention to minimum acceptable (24h for most streaming use cases)
- Use enhanced fan-out only when you have >3 consumers or need <100ms latency

## Naming Constraints

- Stream names are unique per AWS account per region
- 1-128 characters
- Allowed: letters, digits, hyphens, underscores, periods
- **ForceNew**: Stream name cannot be changed after creation
