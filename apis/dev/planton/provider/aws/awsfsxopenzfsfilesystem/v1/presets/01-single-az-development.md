# Preset: Single-AZ Development

**Use case**: Development and testing environments where cost efficiency matters more than high availability or performance.

## Configuration

- **Deployment type**: SINGLE_AZ_1 — lowest-cost deployment option
- **Storage**: 64 GiB minimum — enough for development datasets
- **Throughput**: 64 MB/s — sufficient for development I/O patterns
- **Compression**: None (default) — keeps it simple for dev
- **Backups**: Disabled — dev environments typically don't need backups

## When to use

- Local development and testing
- Prototyping and proof-of-concept workloads
- Non-production environments with limited data
- Cost-sensitive workloads where data durability is not critical

## Cost considerations

SINGLE_AZ_1 with minimum storage and throughput is the lowest-cost FSx for OpenZFS option. No backup retention means no additional storage costs for snapshots.
