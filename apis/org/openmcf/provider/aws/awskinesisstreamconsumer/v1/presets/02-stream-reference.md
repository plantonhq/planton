# Preset: Stream Reference (valueFrom)

## Use Case

Register an enhanced fan-out consumer with an OpenMCF-managed Kinesis stream using a `valueFrom` reference. The platform resolves the stream ARN at deployment time and creates a dependency edge in the infra chart DAG, ensuring the stream is provisioned before the consumer.

## What You Get

- **Dedicated throughput**: 2 MB/s per shard, independent of other consumers
- **Push delivery**: ~70ms propagation delay via HTTP/2 (SubscribeToShard)
- **Dependency wiring**: Automatic deployment ordering via the infra chart DAG
- **No hardcoded ARNs**: Stream ARN resolved from the referenced resource's outputs

## When to Use

- Production deployments where the stream is also managed by OpenMCF
- Infra chart composition with AwsKinesisStream as a dependency
- Multi-consumer setups where each consumer is a separate resource referencing the same stream

## Cost

~$0.015/consumer-shard-hour + $0.013/GB retrieved. For a 4-shard stream: ~$43.80/month base consumer cost (before data retrieval).
