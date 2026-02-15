# Preset: Basic Consumer

## Use Case

Register an enhanced fan-out consumer with an existing Kinesis stream using a direct ARN. Suitable for quick setup when the stream ARN is known and not managed by OpenMCF.

## What You Get

- **Dedicated throughput**: 2 MB/s per shard, independent of other consumers
- **Push delivery**: ~70ms propagation delay via HTTP/2 (SubscribeToShard)
- **Immutable binding**: Consumer name and stream ARN are fixed after creation

## When to Use

- Development and testing against an existing stream
- Standalone consumer registration without infra chart composition
- Quick prototyping of enhanced fan-out patterns

## Cost

~$0.015/consumer-shard-hour + $0.013/GB retrieved. For a 4-shard stream: ~$43.80/month base consumer cost (before data retrieval).
