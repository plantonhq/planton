# AwsKinesisStreamConsumer — Architecture Reference

## Overview

An Amazon Kinesis enhanced fan-out consumer is a named, registered consumer that receives data from a Kinesis Data Stream via a dedicated throughput channel. Unlike standard consumers that share the shard's 2 MB/s read capacity, each enhanced fan-out consumer gets its own 2 MB/s pipe per shard, delivered via HTTP/2 push (SubscribeToShard API).

## Consumer Model

### Standard Consumers (GetRecords)

Standard consumers poll the stream using the GetRecords API. All standard consumers for a shard share the same 2 MB/s read throughput:

- Up to 5 GetRecords calls per shard per second
- Each call returns up to 10 MB (or 10,000 records)
- Effective throughput: 2 MB/s per shard shared across all consumers
- Consumer applications use shard iterators to track position

With multiple standard consumers, each gets a fraction of the 2 MB/s. For example, 5 consumers sharing a shard each effectively get ~400 KB/s.

### Enhanced Fan-Out Consumers (SubscribeToShard)

Enhanced fan-out consumers register with the stream and receive push-based delivery:

- Dedicated 2 MB/s per shard per consumer (not shared)
- Push delivery via HTTP/2 long-lived connections
- ~70ms propagation delay (vs ~200ms for polling)
- Automatic shard-level data delivery without polling loops
- Up to 20 consumers per stream (soft limit, can be increased via AWS Support)

### When Enhanced Fan-Out Matters

Enhanced fan-out is justified when:

1. **3+ consumers** on the same stream — shared throughput becomes a bottleneck
2. **Sub-100ms latency** requirements — push delivery is inherently faster than polling
3. **Guaranteed throughput** — each consumer needs predictable, dedicated bandwidth
4. **Lambda integration** — AWS Lambda event source mappings with enhanced fan-out provide the lowest latency path

### When Standard Consumers Suffice

Standard consumers are sufficient when:

1. **1-2 consumers** — ample shared throughput
2. **Cost sensitivity** — no per-consumer-shard-hour charge
3. **Higher latency acceptable** — 200ms+ is fine for the use case
4. **Simple KCL applications** — the Kinesis Client Library handles standard consumers well

## Resource Lifecycle

### Registration

When you deploy an AwsKinesisStreamConsumer, the IaC module calls `RegisterStreamConsumer`:

1. AWS validates the consumer name is unique within the stream
2. The consumer enters `CREATING` status
3. AWS provisions the dedicated throughput channel (typically 10-30 seconds)
4. The consumer transitions to `ACTIVE` status
5. The consumer ARN is exported (includes the creation timestamp)

### Immutability

Both inputs are **ForceNew**:

- **Consumer name** (from `metadata.name`) — Cannot be renamed. Changing the name deletes the old consumer and creates a new one with a new ARN.
- **Stream ARN** — A consumer is bound to exactly one stream. Moving a consumer to a different stream requires deletion and re-creation.

This immutability means the `consumer_arn` is stable for the lifetime of the resource. Downstream references (Lambda event source mappings, application configurations) do not break unless the consumer is explicitly deleted or replaced.

### Deregistration

Deleting the resource calls `DeregisterStreamConsumer`:

1. The consumer enters `DELETING` status
2. AWS removes the dedicated throughput channel (typically 10-30 seconds)
3. Active SubscribeToShard connections are terminated
4. The consumer ARN becomes invalid

**Important**: If a consumer is actively being used by a Lambda event source mapping or application, deleting the consumer will cause those integrations to fail. Ensure downstream dependencies are removed or updated first.

## Lambda Event Source Mapping Integration

The most common use of enhanced fan-out consumers is with AWS Lambda:

1. Deploy an `AwsKinesisStreamConsumer` resource
2. Note the `consumer_arn` from outputs
3. Create a Lambda event source mapping with:
   - `EventSourceArn`: the stream ARN
   - `ConsumerArn`: the consumer ARN from step 2
   - `StartingPosition`: `LATEST` or `TRIM_HORIZON`

Lambda then receives records via push delivery with dedicated throughput.

**Key benefit**: Multiple Lambda functions can each have their own consumer, processing the stream independently at full shard throughput without competing with each other.

## Cost Model

Enhanced fan-out pricing (in addition to the base stream cost):

| Component | Price | Unit |
|-----------|-------|------|
| Consumer registration | $0.015 | per consumer-shard-hour |
| Data retrieval | $0.013 | per GB retrieved |

### Cost Example

A stream with 4 shards and 3 enhanced fan-out consumers, processing 100 GB/day:

- **Consumer-shard-hours**: 3 consumers x 4 shards x 730 hours = 8,760 consumer-shard-hours
- **Consumer cost**: 8,760 x $0.015 = **$131.40/month**
- **Retrieval cost**: 3 consumers x 100 GB x 30 days x $0.013 = **$117.00/month**
- **Total fan-out cost**: ~$248/month (in addition to base stream cost)

### Cost Optimization

- Only register consumers that truly need dedicated throughput
- Use standard consumers (GetRecords) for non-latency-sensitive workloads
- Consider the break-even: enhanced fan-out becomes cost-effective at ~3 consumers where shared throughput would be the bottleneck

## Security

### IAM Permissions

Consumers require these IAM permissions:

- `kinesis:SubscribeToShard` — Required for push-based delivery
- `kinesis:DescribeStreamConsumer` — Required to check consumer status
- `kinesis:RegisterStreamConsumer` — Required for the IaC module to create the consumer
- `kinesis:DeregisterStreamConsumer` — Required for the IaC module to delete the consumer

### Encryption

Enhanced fan-out consumers inherit the stream's encryption settings. If the stream uses KMS encryption, consumers transparently decrypt data without additional configuration. The consumer's IAM principal needs `kms:Decrypt` permission on the stream's KMS key.

## Limits

| Limit | Value | Type |
|-------|-------|------|
| Consumers per stream | 20 | Soft (can be increased) |
| SubscribeToShard calls per consumer per shard | 1 per second | Hard |
| Consumer name length | 1-128 characters | Hard |
| Consumer name characters | Letters, digits, hyphens, underscores, periods | Hard |

## Naming Constraints

- Consumer names must be unique within a stream
- 1-128 characters
- Allowed: letters, digits, hyphens, underscores, periods
- **ForceNew**: Consumer name cannot be changed after registration
