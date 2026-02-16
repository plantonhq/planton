---
title: "MSK Cluster"
description: "MSK Cluster deployment documentation"
icon: "package"
order: 100
componentName: "awsmskcluster"
---

# AWS MSK Cluster

Deploys an Amazon MSK (Managed Streaming for Apache Kafka) cluster with configurable broker nodes, multi-method authentication (SASL/IAM, SASL/SCRAM, mTLS), encryption at rest and in transit, inline Kafka configuration management, and broker log delivery to CloudWatch Logs, Kinesis Data Firehose, and S3. The component creates a managed security group with Kafka and ZooKeeper port rules when ingress sources are specified.

## What Gets Created

When you deploy an AwsMskCluster resource, OpenMCF provisions:

- **MSK Cluster** — an `aws_msk_cluster` resource with the specified number of broker nodes distributed across subnets, configured with the requested Kafka version, instance type, authentication methods, encryption settings, and monitoring level
- **Security Group** — created only when `securityGroupIds` or `allowedCidrBlocks` are provided; opens ports 9092-9098 (Kafka broker protocols) and 2181-2182 (ZooKeeper) for the specified source security groups and CIDR ranges, with unrestricted egress
- **MSK Configuration** — created only when `serverProperties` is provided; holds Apache Kafka server.properties overrides (e.g., replication factor, min ISR, auto-create topics) and is associated with the cluster

## Prerequisites

- **AWS credentials** configured via environment variables or OpenMCF provider config
- **At least one VPC subnet** for broker placement; three subnets across distinct Availability Zones recommended for production
- **A VPC ID** if specifying `securityGroupIds` or `allowedCidrBlocks` (required for managed security group creation)
- **A KMS key ARN** if using customer-managed encryption at rest
- **An ACM Private CA ARN** if enabling mutual TLS (mTLS) authentication
- **A CloudWatch Log Group** if enabling CloudWatch broker log delivery
- **A Kinesis Data Firehose delivery stream** if enabling Firehose broker log delivery
- **An S3 bucket** if enabling S3 broker log delivery

## Quick Start

Create a file `msk.yaml`:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsMskCluster
metadata:
  name: my-kafka
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AwsMskCluster.my-kafka
spec:
  kafkaVersion: "3.6.0"
  numberOfBrokerNodes: 3
  instanceType: kafka.t3.small
  subnetIds:
    - subnet-0a1b2c3d4e5f00001
    - subnet-0a1b2c3d4e5f00002
    - subnet-0a1b2c3d4e5f00003
  authentication:
    saslIamEnabled: true
```

Deploy:

```shell
openmcf apply -f msk.yaml
```

This creates a 3-broker MSK cluster with SASL/IAM authentication across three subnets, TLS encryption enabled by default.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `kafkaVersion` | `string` | Apache Kafka version (e.g., "3.6.0", "3.5.1"). Downgrades force cluster replacement. | Required |
| `numberOfBrokerNodes` | `int` | Total broker nodes. Must be a multiple of the number of subnets for even AZ distribution. | Required, >= 1 |
| `instanceType` | `string` | Broker EC2 instance type (e.g., "kafka.m5.large", "kafka.m7g.xlarge", "kafka.t3.small"). | Required |
| `subnetIds` | `StringValueOrRef[]` | VPC subnets for broker placement. ForceNew. Can reference AwsVpc via `valueFrom`. | Minimum 1 item |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `securityGroupIds` | `StringValueOrRef[]` | `[]` | Source security groups for managed SG ingress rules. Can reference AwsSecurityGroup via `valueFrom`. |
| `allowedCidrBlocks` | `string[]` | `[]` | IPv4 CIDR ranges for managed SG ingress rules. Must be valid CIDR notation. |
| `associateSecurityGroupIds` | `StringValueOrRef[]` | `[]` | Existing security groups attached directly to the cluster. ForceNew: changes force replacement. |
| `vpcId` | `StringValueOrRef` | — | VPC for managed security group creation. Required when `securityGroupIds` or `allowedCidrBlocks` are set. |
| `ebsVolumeSizeGib` | `int` | AWS default | EBS volume size per broker in GiB. Range: 1-16384. |
| `provisionedThroughputEnabled` | `bool` | `false` | Enable provisioned EBS throughput. Requires large instance types and `ebsVolumeSizeGib` >= 10. |
| `provisionedThroughputMbs` | `int` | — | Provisioned throughput in MiB/s per broker. Range: 250-2375. Required when `provisionedThroughputEnabled` is `true`. |
| `storageMode` | `string` | — | `LOCAL` or `TIERED`. Tiered offloads warm data to S3 for cost optimization. |
| `kmsKeyArn` | `StringValueOrRef` | AWS-managed key | KMS key for at-rest encryption. ForceNew. Can reference AwsKmsKey via `valueFrom`. |
| `clientBrokerEncryption` | `string` | `TLS` | Client-broker encryption: `TLS`, `TLS_PLAINTEXT`, or `PLAINTEXT`. |
| `inClusterEncryption` | `bool` | `true` | Inter-broker TLS encryption. ForceNew. |
| `authentication` | `object` | — | Client authentication configuration. See below. |
| `authentication.saslIamEnabled` | `bool` | `false` | Enable SASL/IAM authentication (port 9098). Recommended. |
| `authentication.saslScramEnabled` | `bool` | `false` | Enable SASL/SCRAM-SHA-512 authentication (port 9096). |
| `authentication.tlsEnabled` | `bool` | `false` | Enable mutual TLS authentication (port 9094). |
| `authentication.tlsCertificateAuthorityArns` | `StringValueOrRef[]` | `[]` | ACM Private CA ARNs for mTLS. Required when `tlsEnabled` is `true`. |
| `authentication.unauthenticated` | `bool` | `false` | Allow unauthenticated connections. Not recommended for production. |
| `configurationArn` | `string` | — | ARN of an external MSK Configuration. Mutually exclusive with `serverProperties`. |
| `configurationRevision` | `int` | — | Revision of external configuration. Required when `configurationArn` is set. >= 1. |
| `serverProperties` | `map<string,string>` | `{}` | Inline Kafka server.properties overrides. Creates an MSK Configuration resource. Mutually exclusive with `configurationArn`. |
| `logging.cloudwatchLogs.enabled` | `bool` | `false` | Enable CloudWatch Logs delivery. |
| `logging.cloudwatchLogs.logGroup` | `StringValueOrRef` | — | CloudWatch Log Group name. Required when enabled. Can reference AwsCloudwatchLogGroup via `valueFrom`. |
| `logging.firehose.enabled` | `bool` | `false` | Enable Firehose delivery. |
| `logging.firehose.deliveryStream` | `StringValueOrRef` | — | Firehose delivery stream name. Required when enabled. Can reference AwsKinesisFirehose via `valueFrom`. |
| `logging.s3.enabled` | `bool` | `false` | Enable S3 delivery. |
| `logging.s3.bucket` | `StringValueOrRef` | — | S3 bucket name. Required when enabled. Can reference AwsS3Bucket via `valueFrom`. |
| `logging.s3.prefix` | `string` | — | Optional S3 key prefix for log objects. |
| `enhancedMonitoring` | `string` | `DEFAULT` | CloudWatch metrics level: `DEFAULT`, `PER_BROKER`, `PER_TOPIC_PER_BROKER`, `PER_TOPIC_PER_PARTITION`. |
| `jmxExporterEnabled` | `bool` | `false` | Enable Prometheus JMX Exporter (port 11001). |
| `nodeExporterEnabled` | `bool` | `false` | Enable Prometheus Node Exporter (port 11002). |
| `publicAccessType` | `string` | — | `DISABLED` or `SERVICE_PROVIDED_EIPS`. Public access requires SASL/IAM or SASL/SCRAM with TLS. |

## Examples

### Production Cluster with IAM Auth and KMS

A 6-broker cluster with customer-managed encryption, tiered storage, and CloudWatch monitoring:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsMskCluster
metadata:
  name: prod-kafka
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsMskCluster.prod-kafka
spec:
  kafkaVersion: "3.6.0"
  numberOfBrokerNodes: 6
  instanceType: kafka.m7g.xlarge
  subnetIds:
    - subnet-az1
    - subnet-az2
    - subnet-az3
  ebsVolumeSizeGib: 1000
  storageMode: TIERED
  kmsKeyArn: arn:aws:kms:us-east-1:123456789012:key/mrk-abc123
  authentication:
    saslIamEnabled: true
  serverProperties:
    auto.create.topics.enable: "false"
    default.replication.factor: "3"
    min.insync.replicas: "2"
  logging:
    cloudwatchLogs:
      enabled: true
      logGroup: /aws/msk/prod-kafka
  enhancedMonitoring: PER_TOPIC_PER_BROKER
  jmxExporterEnabled: true
  nodeExporterEnabled: true
```

### Multi-Authentication Cluster

All three authentication methods enabled for mixed client populations:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsMskCluster
metadata:
  name: multi-auth-kafka
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsMskCluster.multi-auth-kafka
spec:
  kafkaVersion: "3.6.0"
  numberOfBrokerNodes: 3
  instanceType: kafka.m5.large
  subnetIds:
    - subnet-az1
    - subnet-az2
    - subnet-az3
  authentication:
    saslIamEnabled: true
    saslScramEnabled: true
    tlsEnabled: true
    tlsCertificateAuthorityArns:
      - arn:aws:acm-pca:us-east-1:123456789012:certificate-authority/abc-12345
```

### Full Logging Configuration

Broker logs delivered to all three destinations simultaneously:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsMskCluster
metadata:
  name: logged-kafka
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsMskCluster.logged-kafka
spec:
  kafkaVersion: "3.6.0"
  numberOfBrokerNodes: 3
  instanceType: kafka.m5.large
  subnetIds:
    - subnet-az1
    - subnet-az2
    - subnet-az3
  authentication:
    saslIamEnabled: true
  logging:
    cloudwatchLogs:
      enabled: true
      logGroup: /aws/msk/logged-kafka
    firehose:
      enabled: true
      deliveryStream: msk-logs-to-s3
    s3:
      enabled: true
      bucket: my-msk-audit-logs
      prefix: broker-logs/
```

### Using Foreign Key References

Reference other OpenMCF-managed resources instead of hardcoding IDs:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsMskCluster
metadata:
  name: ref-kafka
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsMskCluster.ref-kafka
spec:
  kafkaVersion: "3.6.0"
  numberOfBrokerNodes: 6
  instanceType: kafka.m7g.xlarge
  subnetIds:
    - valueFrom:
        kind: AwsVpc
        name: production-vpc
        fieldPath: status.outputs.private_subnets.[0].id
    - valueFrom:
        kind: AwsVpc
        name: production-vpc
        fieldPath: status.outputs.private_subnets.[1].id
    - valueFrom:
        kind: AwsVpc
        name: production-vpc
        fieldPath: status.outputs.private_subnets.[2].id
  vpcId:
    valueFrom:
      kind: AwsVpc
      name: production-vpc
      fieldPath: status.outputs.vpc_id
  securityGroupIds:
    - valueFrom:
        kind: AwsSecurityGroup
        name: kafka-clients
        fieldPath: status.outputs.security_group_id
  kmsKeyArn:
    valueFrom:
      kind: AwsKmsKey
      name: platform-key
      fieldPath: status.outputs.key_arn
  authentication:
    saslIamEnabled: true
  logging:
    cloudwatchLogs:
      enabled: true
      logGroup:
        valueFrom:
          kind: AwsCloudwatchLogGroup
          name: kafka-logs
          fieldPath: status.outputs.log_group_name
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `cluster_arn` | `string` | ARN of the MSK cluster, used in IAM policies and event source mappings |
| `cluster_name` | `string` | Name of the MSK cluster |
| `cluster_uuid` | `string` | UUID extracted from the cluster ARN |
| `current_version` | `string` | Cluster version string, required for update operations |
| `bootstrap_brokers` | `string` | Comma-separated plaintext broker endpoints (port 9092). Empty when `clientBrokerEncryption` is `TLS`. |
| `bootstrap_brokers_tls` | `string` | Comma-separated TLS broker endpoints (port 9094) |
| `bootstrap_brokers_sasl_iam` | `string` | Comma-separated SASL/IAM broker endpoints (port 9098). Populated when `saslIamEnabled` is `true`. |
| `bootstrap_brokers_sasl_scram` | `string` | Comma-separated SASL/SCRAM broker endpoints (port 9096). Populated when `saslScramEnabled` is `true`. |
| `bootstrap_brokers_public_tls` | `string` | Comma-separated public TLS endpoints. Populated when `publicAccessType` is `SERVICE_PROVIDED_EIPS`. |
| `bootstrap_brokers_public_sasl_iam` | `string` | Comma-separated public SASL/IAM endpoints |
| `bootstrap_brokers_public_sasl_scram` | `string` | Comma-separated public SASL/SCRAM endpoints |
| `zookeeper_connect_string` | `string` | Comma-separated ZooKeeper plaintext endpoints |
| `zookeeper_connect_string_tls` | `string` | Comma-separated ZooKeeper TLS endpoints |
| `security_group_id` | `string` | ID of the managed security group. Only set when `securityGroupIds` or `allowedCidrBlocks` are provided. |
| `configuration_arn` | `string` | ARN of the inline MSK Configuration. Only set when `serverProperties` is provided. |

## Related Components

- [AwsVpc](/docs/catalog/aws/vpc) — provides subnets for broker placement and VPC ID for managed security group
- [AwsSecurityGroup](/docs/catalog/aws/security-group) — controls network access to Kafka and ZooKeeper ports
- [AwsKmsKey](/docs/catalog/aws/kms-key) — provides customer-managed encryption key for data at rest
- [AwsCloudwatchLogGroup](/docs/catalog/aws/cloudwatch-log-group) — receives broker logs via CloudWatch Logs integration
- [AwsKinesisFirehose](/docs/catalog/aws/kinesis-firehose) — receives broker logs for analytics pipeline delivery
