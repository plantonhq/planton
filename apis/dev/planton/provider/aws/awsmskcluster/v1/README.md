# AwsMskCluster

Amazon MSK (Managed Streaming for Apache Kafka) cluster resource for Planton. Provisions a fully managed Kafka cluster on AWS with configurable broker topology, encryption, authentication, logging, monitoring, and networking. MSK handles broker infrastructure, ZooKeeper coordination, and storage so teams can focus on producing and consuming streaming data.

## When to use

- You need a managed Apache Kafka cluster on AWS without operating broker infrastructure.
- Streaming workloads require durable, ordered, partitioned event logs (event sourcing, CQRS, change data capture, real-time analytics).
- You want IAM-native authentication for Kafka clients instead of managing passwords or certificates.
- Your architecture requires multi-AZ, encrypted-at-rest, encrypted-in-transit Kafka with CloudWatch/S3 logging out of the box.

## Prerequisites

| Prerequisite | Why | Planton Resource |
|---|---|---|
| VPC with private subnets in 2+ AZs | Brokers are placed in subnets; count must be a multiple of subnet count | `AwsVpc` |
| Security groups (optional) | Control which clients can reach Kafka (9092-9098) and ZooKeeper (2181-2182) ports | `AwsSecurityGroup` |
| KMS key (optional) | Customer-managed encryption at rest for EBS volumes | `AwsKmsKey` |
| CloudWatch log group (optional) | Destination for broker log delivery | `AwsCloudwatchLogGroup` |
| Kinesis Firehose delivery stream (optional) | Destination for broker log delivery | `AwsKinesisFirehose` |
| S3 bucket (optional) | Destination for broker log delivery | `AwsS3Bucket` |
| ACM Private CA (optional) | Required for mTLS client authentication | (external) |

## API envelope

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsMskCluster
metadata:
  name: <resource-id>
spec: { ... }
```

## Spec fields reference

### Core

| Field | Type | Required | Default | Description |
|---|---|---|---|---|
| `kafkaVersion` | string | **yes** | — | Apache Kafka version (`3.6.0`, `3.5.1`, `3.4.0`, `2.8.1`, etc.). Upgrades via rolling restart; downgrades force replacement. |
| `numberOfBrokerNodes` | int32 | **yes** | — | Total broker count. Must be a multiple of the subnet count for even AZ distribution. |
| `instanceType` | string | **yes** | — | Broker instance type. Standard: `kafka.m5.large`–`kafka.m5.4xlarge`. Graviton: `kafka.m7g.large`, `kafka.m7g.xlarge`. Dev: `kafka.t3.small`. |

### Networking

| Field | Type | Required | Default | Description |
|---|---|---|---|---|
| `subnetIds` | list(StringValueOrRef) | **yes** (≥1) | — | VPC subnets for broker placement. Supports `value` or `valueFrom` (AwsVpc). **ForceNew**. |
| `securityGroupIds` | list(StringValueOrRef) | no | [] | Source security groups. When set (with `vpcId`), creates a managed SG with ingress on ports 9092-9098 and 2181-2182. |
| `allowedCidrBlocks` | list(string) | no | [] | IPv4 CIDRs allowed to reach brokers. Same managed-SG behavior as `securityGroupIds`. |
| `associateSecurityGroupIds` | list(StringValueOrRef) | no | [] | Existing SGs attached directly alongside the managed SG. **ForceNew**. |
| `vpcId` | StringValueOrRef | conditional | — | VPC for the managed SG. Required when `securityGroupIds` or `allowedCidrBlocks` are set. |
| `publicAccessType` | string | no | `""` | `DISABLED` (default) or `SERVICE_PROVIDED_EIPS`. Public access requires SASL auth + TLS. |

### Storage

| Field | Type | Required | Default | Description |
|---|---|---|---|---|
| `ebsVolumeSizeGib` | int32 | no | instance default | EBS volume per broker (1-16384 GiB). |
| `provisionedThroughputEnabled` | bool | no | false | Enable provisioned EBS throughput. Requires `kafka.m5.4xlarge`+ and ≥10 GiB EBS. |
| `provisionedThroughputMbs` | int32 | conditional | — | Provisioned throughput (250-2375 MiB/s). Required when `provisionedThroughputEnabled` is true. |
| `storageMode` | string | no | `LOCAL` | `LOCAL` (all data on EBS) or `TIERED` (warm data offloaded to S3). Tiered requires Kafka 2.8.2.tiered+. |

### Encryption

| Field | Type | Required | Default | Description |
|---|---|---|---|---|
| `kmsKeyArn` | StringValueOrRef | no | aws/msk service key | KMS key for at-rest encryption. **ForceNew**. |
| `clientBrokerEncryption` | string | no | `TLS` | In-transit encryption between clients and brokers: `TLS`, `TLS_PLAINTEXT`, or `PLAINTEXT`. |
| `inClusterEncryption` | bool | no | true | TLS encryption between brokers. **ForceNew**. |

### Authentication

Nested message `authentication` with:

| Field | Type | Default | Description |
|---|---|---|---|
| `saslIamEnabled` | bool | false | SASL/IAM auth (port 9098). Recommended for most workloads. |
| `saslScramEnabled` | bool | false | SASL/SCRAM-SHA-512 auth (port 9096). Secrets in AWS Secrets Manager. |
| `tlsEnabled` | bool | false | Mutual TLS auth (port 9094). Requires `tlsCertificateAuthorityArns`. |
| `tlsCertificateAuthorityArns` | list(StringValueOrRef) | [] | ACM PCA ARNs for mTLS certificate validation. |
| `unauthenticated` | bool | false | Allow unauthenticated connections. Not recommended for production. |

### Configuration

| Field | Type | Required | Default | Description |
|---|---|---|---|---|
| `configurationArn` | string | no | — | ARN of an external MSK Configuration. Mutually exclusive with `serverProperties`. |
| `configurationRevision` | int32 | conditional | — | Revision of the external config. Required when `configurationArn` is set (≥1). |
| `serverProperties` | map(string, string) | no | {} | Inline Kafka server.properties overrides. Creates an MSK Configuration resource. Mutually exclusive with `configurationArn`. |

### Logging

Nested message `logging` with three destination sub-messages:

| Destination | Fields | Description |
|---|---|---|
| `cloudwatchLogs` | `enabled`, `logGroup` (StringValueOrRef) | Deliver broker logs to CloudWatch Logs. |
| `firehose` | `enabled`, `deliveryStream` (StringValueOrRef) | Deliver broker logs to Kinesis Data Firehose. |
| `s3` | `enabled`, `bucket` (StringValueOrRef), `prefix` | Deliver broker logs to S3. |

All three destinations can be enabled simultaneously.

### Monitoring

| Field | Type | Default | Description |
|---|---|---|---|
| `enhancedMonitoring` | string | `DEFAULT` | CloudWatch metrics level: `DEFAULT`, `PER_BROKER`, `PER_TOPIC_PER_BROKER`, `PER_TOPIC_PER_PARTITION`. |
| `jmxExporterEnabled` | bool | false | Prometheus JMX Exporter on port 11001. |
| `nodeExporterEnabled` | bool | false | Prometheus Node Exporter on port 11002. |

## Output fields reference

| Output | Type | Description |
|---|---|---|
| `cluster_arn` | string | ARN of the MSK cluster. |
| `cluster_name` | string | Human-readable cluster name. |
| `cluster_uuid` | string | UUID extracted from the cluster ARN. |
| `current_version` | string | Cluster version string (changes after each modification). |
| `bootstrap_brokers` | string | Plaintext broker endpoints (port 9092). Empty when TLS-only. |
| `bootstrap_brokers_tls` | string | TLS broker endpoints (port 9094). |
| `bootstrap_brokers_sasl_iam` | string | SASL/IAM broker endpoints (port 9098). |
| `bootstrap_brokers_sasl_scram` | string | SASL/SCRAM broker endpoints (port 9096). |
| `bootstrap_brokers_public_tls` | string | Public TLS broker endpoints (when public access enabled). |
| `bootstrap_brokers_public_sasl_iam` | string | Public SASL/IAM broker endpoints. |
| `bootstrap_brokers_public_sasl_scram` | string | Public SASL/SCRAM broker endpoints. |
| `zookeeper_connect_string` | string | ZooKeeper plaintext endpoints. |
| `zookeeper_connect_string_tls` | string | ZooKeeper TLS endpoints. |
| `security_group_id` | string | Managed security group ID (if created). |
| `configuration_arn` | string | Inline MSK Configuration ARN (if created from `serverProperties`). |

## Examples

### Minimal cluster

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsMskCluster
metadata:
  name: dev-kafka
spec:
  kafkaVersion: "3.6.0"
  numberOfBrokerNodes: 3
  instanceType: kafka.t3.small
  subnetIds:
    - value: subnet-0aaa1111
    - value: subnet-0bbb2222
    - value: subnet-0ccc3333
  authentication:
    saslIamEnabled: true
```

### Production-ready cluster

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsMskCluster
metadata:
  name: prod-events
spec:
  kafkaVersion: "3.6.0"
  numberOfBrokerNodes: 6
  instanceType: kafka.m7g.xlarge
  subnetIds:
    - valueFrom:
        kind: AwsSubnet
        name: production-private-subnet-a
        fieldPath: status.outputs.subnet_id
    - valueFrom:
        kind: AwsSubnet
        name: production-private-subnet-b
        fieldPath: status.outputs.subnet_id
    - valueFrom:
        kind: AwsSubnet
        name: production-private-subnet-c
        fieldPath: status.outputs.subnet_id
  vpcId:
    valueFrom:
      kind: AwsVpc
      name: production-vpc
      fieldPath: status.outputs.vpc_id
  securityGroupIds:
    - valueFrom:
        kind: AwsSecurityGroup
        name: kafka-clients-sg
        fieldPath: status.outputs.security_group_id
  ebsVolumeSizeGib: 1000
  storageMode: TIERED
  kmsKeyArn:
    valueFrom:
      kind: AwsKmsKey
      name: platform-encryption-key
      fieldPath: status.outputs.key_arn
  clientBrokerEncryption: TLS
  inClusterEncryption: true
  authentication:
    saslIamEnabled: true
  serverProperties:
    auto.create.topics.enable: "false"
    default.replication.factor: "3"
    min.insync.replicas: "2"
    num.partitions: "12"
    log.retention.hours: "168"
  logging:
    cloudwatchLogs:
      enabled: true
      logGroup:
        valueFrom:
          kind: AwsCloudwatchLogGroup
          name: kafka-broker-logs
          fieldPath: status.outputs.log_group_name
  enhancedMonitoring: PER_TOPIC_PER_BROKER
  jmxExporterEnabled: true
  nodeExporterEnabled: true
```

## Related resources

| Resource | Relationship |
|---|---|
| `AwsVpc` | Provides subnets and VPC ID for broker placement and managed security group. |
| `AwsSecurityGroup` | Source security groups for managed ingress rules. |
| `AwsKmsKey` | Customer-managed KMS key for at-rest encryption. |
| `AwsCloudwatchLogGroup` | Destination for broker log delivery. |
| `AwsKinesisFirehose` | Destination for broker log delivery via Firehose. |
| `AwsS3Bucket` | Destination for broker log delivery to S3. |

## Cross-field validations

The spec enforces three cross-field validations at the protobuf level:

1. **Provisioned throughput requires MiB/s** — `provisionedThroughputMbs` must be 250-2375 when `provisionedThroughputEnabled` is true.
2. **Configuration mutual exclusion** — `configurationArn` and `serverProperties` cannot both be set.
3. **Configuration revision required** — `configurationRevision` (≥1) is required when `configurationArn` is set.

## Deliberately omitted features

The following MSK features are **not** covered by this v1 API. They may be added in future versions:

| Feature | Reason |
|---|---|
| MSK Serverless | Different provisioning model; would be a separate resource kind. |
| SCRAM secret association | Requires `aws_msk_scram_secret_association` after cluster creation; better modeled as a lifecycle operation. |
| VPC connectivity (multi-VPC) | Requires `aws_msk_vpc_connection`; cross-VPC plumbing is a separate concern. |
| MSK Replicator | Cross-region/cross-cluster replication is a separate resource. |
| Cluster rebalancing (CRUISE_CONTROL) | v2 candidate; requires additional partition assignment configuration. |

## How it works

Planton provisions the MSK cluster via Pulumi or Terraform modules defined in this repository. The API contract is protobuf-based (`api.proto`, `spec.proto`, `stack_outputs.proto`) and stack execution is orchestrated by the platform using the `AwsMskClusterStackInput` (includes provider credentials and target resource).

## References

- [Amazon MSK Documentation](https://docs.aws.amazon.com/msk/latest/developerguide/what-is-msk.html)
- [MSK Best Practices](https://docs.aws.amazon.com/msk/latest/developerguide/bestpractices.html)
- [MSK Pricing](https://aws.amazon.com/msk/pricing/)
- [Apache Kafka Documentation](https://kafka.apache.org/documentation/)
