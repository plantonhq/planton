# AWS Kafka Streaming

Provisions a production-ready Kafka event streaming platform on Amazon MSK with VPC network isolation, S3 for broker log offloading, and CloudWatch monitoring. Multi-AZ deployment with encryption enabled by default.

## Architecture

```
                       Producers / Consumers
                              │
                              ▼
                    ┌───────────────────┐
                    │   AwsMskCluster   │
                    │  (Kafka brokers)  │
                    │  3+ AZ spread     │
                    └───┬──────────┬────┘
                        │          │
             ┌──────────┘          └──────────┐
             ▼                                ▼
    ┌──────────────────┐           ┌─────────────────────┐
    │   AwsS3Bucket    │           │ AwsCloudwatchLogGroup│
    │ (broker logs)    │           │ (broker logs)        │
    └──────────────────┘           └─────────────────────┘

    ┌──────────────────┐           ┌───────────────────┐
    │     AwsVpc       │           │ AwsSecurityGroup  │
    │ (3-AZ network)   │           │ (Kafka ports)     │
    └──────────────────┘           └───────────────────┘
```

## Dependency Graph

```
Layer 0 (parallel):  AwsVpc, AwsS3Bucket, AwsCloudwatchLogGroup
Layer 1 (dep VPC):   AwsSecurityGroup
Layer 2 (dep all):   AwsMskCluster
```

## Included Cloud Resources

| Resource | Kind | Group | Condition | Purpose |
|----------|------|-------|-----------|---------|
| VPC | `AwsVpc` | network | Always | 3-AZ network with private subnets for brokers |
| Security Group | `AwsSecurityGroup` | network | Always | Kafka ports (9092, 9094, 9096, 9098) within VPC CIDR |
| S3 Bucket | `AwsS3Bucket` | storage | Always | Broker log offloading and data archival |
| CloudWatch Log Group | `AwsCloudwatchLogGroup` | monitoring | Always | Broker log streaming (30-day retention) |
| MSK Cluster | `AwsMskCluster` | messaging | Always | Managed Apache Kafka brokers |

## Parameters

| Parameter | Description | Default | Required |
|-----------|-------------|---------|----------|
| **Network** | | | |
| `availability_zone_1` | First AZ | `us-east-1a` | Yes |
| `availability_zone_2` | Second AZ | `us-east-1b` | Yes |
| `availability_zone_3` | Third AZ | `us-east-1c` | Yes |
| **MSK Cluster** | | | |
| `cluster_name` | MSK cluster name | `kafka-streaming` | Yes |
| `kafka_version` | Apache Kafka version | `3.6.0` | Yes |
| `broker_count` | Number of broker nodes | `3` | Yes |
| `broker_instance_type` | Broker instance type | `kafka.m5.large` | Yes |
| `ebs_volume_size_gib` | EBS volume per broker (GiB) | `100` | Yes |
| **Logging** | | | |
| `s3_bucket_name` | S3 bucket for broker logs | `kafka-streaming-logs` | Yes |

## Common Configurations

### Development (small cluster)

```yaml
broker_count: "3"
broker_instance_type: kafka.t3.small
ebs_volume_size_gib: "20"
```

### Production (high throughput)

```yaml
broker_count: "6"
broker_instance_type: kafka.m5.2xlarge
ebs_volume_size_gib: "500"
```

## Important Notes

- MSK requires **3 Availability Zones** for multi-AZ deployment. The `broker_count` should be a multiple of the AZ count (e.g., 3, 6, 9).
- Security group rules restrict Kafka ports to the VPC CIDR (10.0.0.0/16). Producers and consumers must run inside the VPC or use VPC peering / transit gateway.
- The S3 bucket is created for broker log archival. Configure MSK's logging settings to ship logs to S3 after deployment.
- Broker instance types starting with `kafka.t3` are burstable and suitable for development only. Use `kafka.m5` or larger for production workloads.
- EBS volumes are provisioned per broker. Total storage = `ebs_volume_size_gib` × `broker_count`.

---

© Planton. Licensed under [Apache-2.0](../../../LICENSE).
