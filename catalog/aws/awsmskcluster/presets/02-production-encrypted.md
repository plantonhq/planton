# Production Encrypted Kafka Cluster

A production-grade MSK cluster with customer-managed KMS encryption, tiered storage,
comprehensive monitoring, and hardened Kafka server properties.

## When to Use

- Production workloads requiring enterprise-grade security
- Compliance environments mandating customer-managed encryption keys
- High-throughput event streaming with cost-optimized tiered storage
- Teams using Prometheus for Kafka observability

## Configuration Highlights

- **Instance type**: `kafka.m7g.xlarge` (Graviton, optimal price-performance)
- **Brokers**: 6 across 3 AZs (2 per AZ for high throughput and rebalancing headroom)
- **Authentication**: SASL/IAM (no credentials to rotate)
- **Encryption**: Customer-managed KMS key, TLS client-broker, in-cluster TLS
- **Storage**: 1 TB EBS per broker with TIERED mode (hot on EBS, warm on S3)
- **Server properties**: Auto-create disabled, RF=3, ISR=2, 12 default partitions, 7-day retention
- **Logging**: CloudWatch Logs for broker diagnostics
- **Monitoring**: PER_TOPIC_PER_BROKER metrics + Prometheus JMX and Node exporters

## Infra Chart Composition

This preset uses `valueFrom` references to compose with:
- **AwsSubnet** (broker placement)
- **AwsSecurityGroup** (attached to the broker network interfaces; Kafka/ZooKeeper ingress rules live on this first-class node)
- **AwsKmsKey** (encryption at rest)
- **AwsCloudwatchLogGroup** (broker log destination)

## Cost Estimate

The cost drivers are the six kafka.m7g.xlarge brokers (billed hourly, the dominant line) plus the 6 TB of EBS storage and the tiered-storage S3 tier — significantly cheaper per GB than keeping everything on local EBS. The verified figure for this preset lives in the kind's generated estimate at `catalog/_pricing/estimates/awsmskcluster.yaml` — computed from the pinned price book, never hand-typed here.

## Customization

- Add `firehose` and `s3` logging for multi-destination log delivery
- Open VPN CIDR ranges by adding ingress rules on the referenced security group
- Add SASL/SCRAM authentication for non-AWS clients
