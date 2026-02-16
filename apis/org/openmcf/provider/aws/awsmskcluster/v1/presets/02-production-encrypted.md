# Preset: Production Encrypted Kafka Cluster

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
- **AwsVpc** (subnets and VPC ID)
- **AwsSecurityGroup** (client access control)
- **AwsKmsKey** (encryption at rest)
- **AwsCloudwatchLogGroup** (broker log destination)

## Cost Estimate

Approximately $1.40/hr for 6 x kafka.m7g.xlarge brokers (~$1,000/month) plus 6 TB EBS storage
and tiered storage S3 costs (significantly lower than equivalent local-only storage).

## Customization

- Add `firehose` and `s3` logging for multi-destination log delivery
- Add `allowedCidrBlocks` for VPN-based access patterns
- Add SASL/SCRAM authentication for non-AWS clients
