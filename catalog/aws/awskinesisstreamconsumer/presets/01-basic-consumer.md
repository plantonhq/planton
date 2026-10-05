# Basic Consumer

## Use Case

Register an enhanced fan-out consumer with an existing Kinesis stream using a direct ARN. Suitable for quick setup when the stream ARN is known and not managed by Planton.

## What You Get

- **Dedicated throughput**: 2 MB/s per shard, independent of other consumers
- **Push delivery**: ~70ms propagation delay via HTTP/2 (SubscribeToShard)
- **Immutable binding**: Consumer name and stream ARN are fixed after creation

## When to Use

- Development and testing against an existing stream
- Standalone consumer registration without Infra Chart composition
- Quick prototyping of enhanced fan-out patterns

## Cost

Enhanced fan-out bills per consumer-shard-hour plus per GB retrieved. The base cost scales with the stream's shard count — every shard the stream carries is billed for this consumer around the clock, before any data retrieval.
