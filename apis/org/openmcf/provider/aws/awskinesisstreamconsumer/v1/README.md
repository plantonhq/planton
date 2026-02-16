# AwsKinesisStreamConsumer

Registers an [Amazon Kinesis enhanced fan-out consumer](https://docs.aws.amazon.com/streams/latest/dev/enhanced-consumers.html) for a Kinesis Data Stream, providing dedicated 2 MB/s read throughput per shard independent of all other consumers.

## When to Use

Use an AwsKinesisStreamConsumer when you need:

- **Dedicated throughput** — Each registered consumer gets 2 MB/s per shard, not shared with other readers
- **Low-latency delivery** — Push-based delivery via HTTP/2 (SubscribeToShard) with ~70ms propagation delay
- **Multiple independent readers** — Several applications reading from the same stream without contention
- **Lambda with enhanced fan-out** — AWS Lambda event source mappings configured for dedicated throughput

### Standard Consumers vs Enhanced Fan-Out

| Feature | Standard (GetRecords) | Enhanced Fan-Out (SubscribeToShard) |
|---------|----------------------|--------------------------------------|
| Throughput | 2 MB/s per shard (shared) | 2 MB/s per shard per consumer (dedicated) |
| Delivery | Pull (polling) | Push (HTTP/2) |
| Latency | ~200ms | ~70ms |
| Max consumers | 5 per stream (soft) | 20 per stream (soft) |
| Cost | Included with stream | ~$0.015/consumer-shard-hour + $0.013/GB |

**Choose enhanced fan-out** when you have 3+ consumers on the same stream, need sub-100ms latency, or want guaranteed throughput per consumer. **Choose standard** when cost is a priority, you have 1-2 consumers, and 200ms latency is acceptable.

## Prerequisites

- An existing Kinesis Data Stream (see `AwsKinesisStream`)
- AWS account with permissions to register stream consumers (`kinesis:RegisterStreamConsumer`)

## Spec Reference

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `stream_arn` | StringValueOrRef | **Yes** | ARN of the parent Kinesis Data Stream. ForceNew. |

The consumer name is derived from `metadata.name` and cannot be changed after creation (ForceNew).

## Stack Outputs

| Output | Description |
|--------|-------------|
| `consumer_arn` | ARN of the consumer — used for Lambda event source mappings with enhanced fan-out |
| `consumer_name` | Name of the consumer |
| `stream_arn` | ARN of the parent stream (echoed back for convenience) |
| `creation_timestamp` | RFC3339 timestamp of consumer registration |

## Immutability

Both inputs (`name` from metadata and `stream_arn`) are **ForceNew**. Any change to either value causes the consumer to be deregistered and re-registered. There are no mutable configuration options — this resource is effectively create-or-replace.

## Deliberate v1 Omissions

This component has no omissions. The `aws_kinesis_stream_consumer` resource has only two inputs (`name` and `stream_arn`), both of which are exposed. Tags are managed via metadata labels.

## Related Resources

- **AwsKinesisStream** — The parent stream this consumer registers with
- **AwsLambda** — Can be configured with a consumer ARN for enhanced fan-out event source mapping
- **AwsIamRole** — Provides `kinesis:SubscribeToShard` and `kinesis:DescribeStreamConsumer` permissions
