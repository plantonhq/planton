# Stream Reference (valueFrom)

## Use Case

Register an enhanced fan-out consumer with an Planton-managed Kinesis stream using a `valueFrom` reference. The platform resolves the stream ARN at deployment time and creates a dependency edge in the Infra Chart DAG, ensuring the stream is provisioned before the consumer.

## What You Get

- **Dedicated throughput**: 2 MB/s per shard, independent of other consumers
- **Push delivery**: ~70ms propagation delay via HTTP/2 (SubscribeToShard)
- **Dependency wiring**: Automatic deployment ordering via the Infra Chart DAG
- **No hardcoded ARNs**: Stream ARN resolved from the referenced resource's outputs

## When to Use

- Production deployments where the stream is also managed by Planton
- Infra Chart composition with AwsKinesisStream as a dependency
- Multi-consumer setups where each consumer is a separate resource referencing the same stream

## Cost

Enhanced fan-out bills per consumer-shard-hour plus per GB retrieved, so the base cost scales linearly with the stream's shard count — a 4-shard stream pays four shard-hour meters around the clock before any data retrieval.
