# Preset: Production Analytics

## Use Case

Full-featured production stream for analytics pipelines, event sourcing, and high-reliability data ingestion. ON_DEMAND for zero capacity planning with comprehensive monitoring.

## What You Get

- **Capacity**: Auto-scaling (up to 200 MB/s write, 400 MB/s read)
- **Retention**: 7 days (168 hours) — enables reprocessing and late-arriving consumers
- **Encryption**: KMS (Kinesis-owned key)
- **Monitoring**: All 7 shard-level metrics enabled — full per-shard observability
- **Deletion safety**: Consumer deregistration on delete — prevents accidental orphaned consumers

## When to Use

- Production analytics and data pipelines
- Event sourcing architectures requiring replay capability
- Streams feeding multiple consumers (Lambda, Firehose, custom applications)
- High-reliability workloads where monitoring every shard matters

## Recommended Alarms

Pair this preset with CloudWatch alarms on:

- `IteratorAgeMilliseconds` > 3,600,000 (1 hour) — consumer falling behind
- `WriteProvisionedThroughputExceeded` > 0 — should never happen on ON_DEMAND but indicates edge cases
- `ReadProvisionedThroughputExceeded` > 0 — consumer is hitting shared read limits (consider enhanced fan-out)

## Cost

Variable (ON_DEMAND): ~$0.08/GB written + ~$0.04/GB read + extended retention + enhanced monitoring per shard.
