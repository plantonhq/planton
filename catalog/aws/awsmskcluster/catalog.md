# AWS MSK Cluster

Deploys a managed Apache Kafka cluster on Amazon MSK with configurable broker count and instance families (standard, Graviton, and Express brokers), EBS storage with optional tiered storage and provisioned throughput, KMS encryption at rest, TLS/SASL authentication, public access and multi-VPC PrivateLink connectivity, inline Kafka server properties, broker log delivery to CloudWatch Logs, Firehose, and S3, and Prometheus-compatible monitoring. Kafka topics can be declared with the cluster and managed through the MSK topic API — no client connectivity or credential setup at deploy time — and subnets, security groups, KMS keys, and log destinations all wire by reference.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **MSK Cluster** -- a managed Kafka cluster with the specified broker count, instance type, and Kafka version, distributed across subnets in multiple Availability Zones
- **Inline MSK Configuration** -- created only when `serverProperties` entries are provided; applies Apache Kafka `server.properties` overrides (replication factor, min ISR, partition count, log retention)
- **Connectivity Updates** -- public access (`publicAccessType`) and multi-VPC PrivateLink (`vpcConnectivity`) are applied by AWS as follow-up updates after the cluster creates
- **Cluster Policy** -- attached only when `clusterPolicy` is provided; the resource-based IAM policy (a structured document in the spec) that grants cross-account principals PrivateLink connection rights
- **Kafka Topics** -- one per `topics` entry, managed through the MSK topic API with no Kafka client or bootstrap connectivity needed; keyed by name, exported as the `topic_arns` output map
- **Broker Log Delivery** -- configurable delivery to CloudWatch Logs, Kinesis Data Firehose, and S3 (all three can be enabled simultaneously)
- **AWS Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) applied automatically for tracking and governance

Network ingress is composed, never embedded: brokers attach the referenced `securityGroupIds` directly, and the ingress rules that open the Kafka ports (9092 plaintext, 9094 TLS, 9096 SASL/SCRAM, 9098 SASL/IAM) live on those first-class AwsSecurityGroup resources.

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** -- an active connection in the Connect module with credentials for the target AWS account. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or cross-account trust authentication modes.

### AWS Account

- **At least two subnets** in distinct Availability Zones. The number of broker nodes must be a multiple of the number of subnets. Reference AwsSubnet Infra Components via ValueFromRef or provide subnet IDs directly. Subnets are immutable after creation.
- **Security groups** attached to the broker network interfaces. Reference AwsSecurityGroup Infra Components or provide security group IDs. Adding or removing groups after creation forces cluster replacement.
- **A KMS key** (optional) for encrypting data at rest on broker EBS volumes. The KMS key is immutable after creation.
- **A CloudWatch log group, Firehose stream, or S3 bucket** (optional) for broker log delivery.
- **SCRAM secrets** (optional) when SASL/SCRAM is enabled: Secrets Manager secrets named with the `AmazonMSK_` prefix and encrypted with a customer-managed KMS key.

## Deploy

### Console

Open the deployment store, find **AWS MSK Cluster**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Basic Kafka Cluster** preset in the [Presets](#presets) tab to pre-populate a working configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsMskCluster
metadata:
  name: event-bus
  org: acme-corp
  env: prod
spec:
  region: us-west-2
  kafkaVersion: "3.6.0"
  numberOfBrokerNodes: 3
  instanceType: kafka.m5.large
  subnetIds:
    - value: "subnet-0a1b2c3d4e5f00001"
    - value: "subnet-0a1b2c3d4e5f00002"
    - value: "subnet-0a1b2c3d4e5f00003"
  securityGroupIds:
    - value: "sg-0123456789abcdef0"
  authentication:
    saslIamEnabled: true
```

```shell
planton apply -f msk-cluster.yaml
```

This creates a 3-broker Kafka 3.6.0 cluster attached to the referenced security group, with SASL/IAM authentication, TLS-only client encryption (the AWS default), default EBS storage, and the AWS-managed encryption key. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the MSK cluster to subnets, a security group, and a KMS key deployed in the same InfraPipeline:

```yaml
spec:
  subnetIds:
    - valueFrom:
        kind: AwsSubnet
        name: stream-subnet-a
        fieldPath: status.outputs.subnet_id
    - valueFrom:
        kind: AwsSubnet
        name: stream-subnet-b
        fieldPath: status.outputs.subnet_id
    - valueFrom:
        kind: AwsSubnet
        name: stream-subnet-c
        fieldPath: status.outputs.subnet_id
  securityGroupIds:
    - valueFrom:
        kind: AwsSecurityGroup
        name: kafka-clients-sg
        fieldPath: status.outputs.security_group_id
  kmsKeyArn:
    valueFrom:
      kind: AwsKmsKey
      name: platform-encryption-key
      fieldPath: status.outputs.key_arn
```

The InfraPipeline resolves the dependency graph, deploys the subnets, security group, and KMS key first, then provisions the MSK cluster with the resolved values.

## Key Configuration

These are the most important decisions when configuring an MSK cluster. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Broker family and sizing** -- Set `instanceType` to a Graviton type (e.g., `kafka.m7g.xlarge`) for better price-performance, an Express type (`express.m7g.large` and up) for AWS-managed storage with faster scaling and intelligent rebalancing (`rebalancingStatus`), or `kafka.t3.small` for development only. Set `numberOfBrokerNodes` as a multiple of the number of subnets for even AZ distribution. Instance type changes and broker additions roll in place; the broker count never decreases.

**Authentication** -- Enable `saslIamEnabled` for IAM-based authentication (recommended for most workloads -- no password management required). Add `saslScramEnabled` for non-AWS clients using username/password and associate the credentials via `scramSecretArns`. Enable `tlsEnabled` for mutual TLS with `tlsCertificateAuthorityArns` naming the trusted private CAs. Multiple methods can be enabled simultaneously.

**Connectivity beyond the VPC** -- `vpcConnectivity` offers PrivateLink access to clients in other VPCs or accounts; each offered scheme must also be enabled in `authentication`, and cross-account consumers additionally need the `clusterPolicy` granting `kafka:CreateVpcConnection`. `publicAccessType: SERVICE_PROVIDED_EIPS` assigns public IPs -- it requires real client authentication (unauthenticated off) and TLS-only client encryption, and AWS applies it as a follow-up update after creation. `networkType: DUAL` adds IPv6 (one-way).

**Encryption** -- Keep `clientBrokerEncryption` at `TLS` (the default, and the only posture that allows public access). `inClusterEncryption` (broker-to-broker TLS) and `kmsKeyArn` (at-rest) are immutable after creation.

**Server properties** -- Use `serverProperties` to override Kafka `server.properties` (e.g., `default.replication.factor: "3"`, `min.insync.replicas: "2"`, `auto.create.topics.enable: "false"`), or pin an externally managed MSK Configuration via `configurationArn` + `configurationRevision` -- never both.

**Declared topics** -- `topics` entries deploy WITH the cluster through the MSK topic API: no Kafka client, bootstrap connectivity, or credentials at deploy time. Each entry is keyed by name (adding or removing one never churns the others); partition counts can only grow in place, while name and replication factor changes replace the topic. Pair with `auto.create.topics.enable: "false"` so the declared contract topics are the only ones applications can rely on. Deleting an entry deletes the topic and its data (requires `delete.topic.enable=true`, MSK's default).

**Storage** -- Set `ebsVolumeSizeGib` for per-broker EBS volume size (1-16384 GiB; grows in place). Enable `storageMode: TIERED` to offload warm data to low-cost storage (one-way). Enable provisioned throughput on `kafka.m5.4xlarge`+ for guaranteed per-broker EBS bandwidth. Express brokers manage their own storage -- none of these apply.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsSubnet** | `subnetIds` | `status.outputs.subnet_id` |
| **AwsSecurityGroup** (optional) | `securityGroupIds` | `status.outputs.security_group_id` |
| **AwsKmsKey** (optional) | `kmsKeyArn` | `status.outputs.key_arn` |
| **AwsCloudwatchLogGroup** (optional) | `logging.cloudwatchLogs.logGroup` | `status.outputs.log_group_name` |
| **AwsKinesisFirehose** (optional) | `logging.firehose.deliveryStream` | `status.outputs.delivery_stream_name` |
| **AwsS3Bucket** (optional) | `logging.s3.bucket` | `status.outputs.bucket_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `cluster_arn` | Amazon Resource Name of the cluster | IAM policies, Lambda event source mappings, the cluster policy |
| `cluster_name` | Cluster name | CloudWatch alarm dimensions, application configuration |
| `cluster_uuid` | Unique identifier extracted from the ARN | Fine-grained IAM resource patterns |
| `bootstrap_brokers_tls` | TLS broker endpoints (port 9094) | Application connection strings with TLS encryption |
| `bootstrap_brokers_sasl_iam` | SASL/IAM broker endpoints (port 9098) | Application connection strings with IAM authentication |
| `bootstrap_brokers_sasl_scram` | SASL/SCRAM broker endpoints (port 9096) | Application connection strings with password authentication |
| `topic_arns` | Map of declared topic name to topic ARN | IAM policies scoping producers and consumers to exactly the contract topics |

Every connectivity surface exports its own bootstrap variant alongside these: `bootstrap_brokers` (plaintext, populated only when plaintext is enabled), the `bootstrap_brokers_public_*` endpoints on a public cluster, and the `bootstrap_brokers_vpc_connectivity_*` endpoints over PrivateLink — pick the one matching how the client reaches the cluster. `zookeeper_connect_string` / `zookeeper_connect_string_tls` are empty on KRaft-mode clusters, and `current_version` / `configuration_arn` track update state rather than feeding downstream wiring.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Basic development cluster** -- 3 brokers on `kafka.t3.small` with SASL/IAM authentication. Minimal cost for development and testing workloads. Start from the **Basic Kafka Cluster** preset.

**Production encrypted cluster** -- 6 brokers on `kafka.m7g.xlarge` with tiered storage, KMS encryption, TLS client-broker encryption, per-topic-per-broker monitoring, Prometheus exporters, and hardened server properties (replication factor 3, min ISR 2). Start from the **Production Encrypted Kafka Cluster** preset.

**Multi-auth with full logging** -- 3 brokers with SASL/IAM, SASL/SCRAM, and mutual TLS authentication enabled simultaneously. Broker logs delivered to CloudWatch Logs, Firehose, and S3. Suitable for enterprises with diverse client populations and audit requirements. Start from the **Multi-Authentication with Full Logging** preset.

**Declared contract topics** -- the cluster and its topics deploy as one unit: an event stream, its dead-letter queue, and a compacted snapshot topic declared in `topics`, with `auto.create.topics.enable` off so a typo'd topic name fails loudly instead of materializing an accidental single-replica topic. The `topic_arns` output scopes producer/consumer IAM policies per topic. Start from the **Cluster with Declared Contract Topics** preset.

## Works With

- [**AWS Subnet**](/infra-catalog/aws-subnet) -- provides the subnets brokers are placed in across Availability Zones
- [**AWS Security Group**](/infra-catalog/aws-security-group) -- provides the security groups attached to broker network interfaces
- [**AWS KMS Key**](/infra-catalog/aws-kms-key) -- provides a customer-managed key for EBS volume encryption at rest
- [**AWS CloudWatch Log Group**](/infra-catalog/aws-cloudwatch-log-group) -- provides the log group for broker log delivery
- [**AWS Kinesis Firehose**](/infra-catalog/aws-kinesis-firehose) -- provides a delivery stream for broker log delivery
- [**AWS S3 Bucket**](/infra-catalog/aws-s3-bucket) -- provides an S3 bucket for broker log delivery
- [**AWS Lambda Event Source Mapping**](/infra-catalog/aws-lambda-event-source-mapping) -- consumes the cluster ARN to trigger Lambda functions from Kafka topics
