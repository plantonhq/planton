# Provisioned Encrypted

## Use Case

A provisioned stream with predictable capacity, KMS encryption using the Kinesis-owned key, and 48-hour retention for basic reprocessing.

## What You Get

- **Capacity**: 2 shards (2 MB/s write, 4 MB/s read)
- **Retention**: 48 hours
- **Encryption**: KMS (Kinesis-owned key — no additional KMS cost)
- **Monitoring**: Stream-level metrics only

## When to Use

- Staging environments with steady, predictable throughput
- Workloads processing sensitive data that requires encryption
- Streams where you want to control and predict costs
- When you know your throughput fits within 2 shards (2 MB/s write)

## Cost

The cost drivers are the two provisioned shards (billed per shard-hour, running whether data flows or not) plus the extended-retention charge for hours 25-48. The verified figure for this preset lives in the kind's generated estimate at `catalog/_pricing/estimates/awskinesisstream.yaml` — computed from the pinned price book, never hand-typed here.
