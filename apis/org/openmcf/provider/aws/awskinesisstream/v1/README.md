# AwsKinesisStream

Deploys an [Amazon Kinesis Data Stream](https://docs.aws.amazon.com/streams/latest/dev/introduction.html) — a real-time data streaming service that captures gigabytes of data per second from hundreds of thousands of sources.

## When to Use

Use Kinesis Data Streams when you need:

- **Real-time data ingestion** — Click-streams, IoT telemetry, database change events, log streams
- **Event sourcing** — Durable, ordered event log with configurable retention (up to 365 days)
- **Fan-out to multiple consumers** — Multiple applications can independently read from the same stream
- **Streaming analytics** — Feed data into Kinesis Firehose, Lambda, or custom consumers for real-time processing

### Kinesis vs SQS vs SNS

| Feature | Kinesis | SQS | SNS |
|---------|---------|-----|-----|
| Model | Streaming (pull) | Queue (pull) | Pub/sub (push) |
| Ordering | Per-shard (partition key) | FIFO or best-effort | No ordering |
| Retention | 24h – 365 days | 1 min – 14 days | No retention |
| Consumers | Multiple, independent | Single consumer per message | Multiple subscribers |
| Replay | Yes (seek by timestamp/sequence) | No (once consumed, gone) | No |
| Throughput | 1 MB/s per shard (provisioned) or auto-scaling (on-demand) | Scales automatically | Scales automatically |

**Choose Kinesis** when you need ordered replay, multiple independent consumers, or high-throughput streaming. **Choose SQS** for task queues and decoupling. **Choose SNS** for fan-out notifications.

## Prerequisites

- AWS account with permissions to create Kinesis streams
- (Optional) AWS KMS key for encryption (see `AwsKmsKey`)

## Capacity Modes

### PROVISIONED

You specify the number of shards. Each shard provides:

- **Write**: 1 MB/s or 1,000 records/s (whichever is reached first)
- **Read**: 2 MB/s (shared across all consumers; 2 MB/s per consumer with enhanced fan-out)

Best for **predictable, steady workloads** where you can estimate throughput. You pay per shard-hour.

### ON_DEMAND

AWS automatically manages shard count to accommodate throughput. Supports:

- **Write**: Up to 200 MB/s
- **Read**: Up to 400 MB/s

Best for **variable or unpredictable workloads**. You pay per GB of data written and read. No capacity planning required.

## Spec Reference

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `stream_mode` | string | **Yes** | — | `"PROVISIONED"` or `"ON_DEMAND"` |
| `shard_count` | int32 | Conditional | — | Number of shards (required for PROVISIONED, forbidden for ON_DEMAND) |
| `retention_period_hours` | int32 | No | 24 | Data retention: 24–8760 hours (1 day to 365 days) |
| `kms_key_id` | StringValueOrRef | No | — | KMS key for encryption. Presence enables KMS encryption |
| `max_record_size_in_kib` | int32 | No | 1024 | Max record size: 1024–10240 KiB (1–10 MiB) |
| `shard_level_metrics` | repeated string | No | [] | Enhanced CloudWatch metrics per shard |
| `enforce_consumer_deletion` | bool | No | false | Auto-deregister consumers on stream deletion |

## Stack Outputs

| Output | Description |
|--------|-------------|
| `stream_arn` | ARN of the Kinesis stream (used by Firehose, Lambda, IAM policies) |
| `stream_name` | Name of the Kinesis stream (used for API calls) |

## Deliberate v1 Omissions

| Feature | Reason |
|---------|--------|
| Resource-based policy | Separate TF resource (`aws_kinesis_resource_policy`), <20% usage. Most streams use IAM. |
| Stream consumers (enhanced fan-out) | Independent lifecycle, ForceNew on name+stream_arn. Separate component recommended. |
| Warm throughput | Not available in pinned provider versions (AWS Native SDK only). |

## Related Resources

- **AwsKmsKey** — Encryption key referenced by `kms_key_id`
- **AwsKinesisFirehose** (R17) — Delivery stream that reads from this stream
- **AwsLambda** — Can be configured with Kinesis as an event source
- **AwsCloudwatchAlarm** — Monitor IteratorAgeMilliseconds for consumer lag alerts
