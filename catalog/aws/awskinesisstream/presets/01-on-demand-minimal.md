# On-Demand Minimal

## Use Case

The simplest possible Kinesis stream for development, prototyping, or variable-throughput workloads. AWS manages all capacity automatically.

## What You Get

- **Capacity**: Auto-scaling (up to 200 MB/s write, 400 MB/s read)
- **Retention**: 24 hours (AWS default)
- **Encryption**: None
- **Monitoring**: Stream-level metrics only

## When to Use

- Development and testing environments
- New projects where throughput is unknown
- Bursty or unpredictable workloads
- Getting started with Kinesis

## Cost

Pay-per-use: ON_DEMAND streams bill per GB written and per GB read, with writes the pricier of the two. No idle cost when no data is flowing. The verified figure for this preset lives in the kind's generated estimate at `catalog/_pricing/estimates/awskinesisstream.yaml` — computed from the pinned price book, never hand-typed here.
