# Basic Kafka Cluster

A minimal 3-broker MSK cluster suitable for development and testing workloads.

## When to Use

- Development and testing environments
- Small-scale event streaming prototyping
- Teams getting started with managed Kafka on AWS

## Configuration Highlights

- **Instance type**: `kafka.t3.small` (burstable, cost-effective for low throughput)
- **Brokers**: 3 across 3 AZs (minimum for high availability)
- **Authentication**: SASL/IAM (recommended, no password management)
- **Encryption**: Defaults (TLS client-broker, in-cluster encryption enabled)
- **Storage**: AWS-managed defaults per instance type

## Cost Estimate

The cost drivers are the three kafka.t3.small brokers (billed hourly, the dominant line) plus their EBS storage. The verified figure for this preset lives in the kind's generated estimate at `catalog/_pricing/estimates/awsmskcluster.yaml` — computed from the pinned price book, never hand-typed here.

## Customization

- Upgrade `instanceType` to `kafka.m5.large` for production workloads
- Add `serverProperties` to tune replication factor and ISR settings
- Add `logging` for CloudWatch or S3 broker log delivery
