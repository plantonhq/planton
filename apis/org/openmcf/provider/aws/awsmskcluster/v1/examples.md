# AwsMskCluster Examples

Realistic deployment examples for the AwsMskCluster resource. Each example is a complete manifest ready for customization.

---

## 1. Minimal cluster — 3 brokers with SASL/IAM

The smallest production-viable cluster. Three brokers across three AZs with IAM authentication and default encryption.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsMskCluster
metadata:
  name: dev-events
spec:
  kafkaVersion: "3.6.0"
  numberOfBrokerNodes: 3
  instanceType: kafka.t3.small
  subnetIds:
    - value: subnet-0a1b2c3d4e5f60001
    - value: subnet-0a1b2c3d4e5f60002
    - value: subnet-0a1b2c3d4e5f60003
  authentication:
    saslIamEnabled: true
```

**Key points:**
- `kafka.t3.small` is suitable for development and low-throughput staging.
- SASL/IAM is the recommended default — no password management required.
- Default encryption: TLS for client-broker, TLS for inter-broker, AWS-managed KMS for at-rest.
- EBS volume size defaults to the instance-type-specific value.

---

## 2. Production encrypted with KMS + TIERED storage

High-throughput production cluster with customer-managed encryption, tiered storage for cost optimization, and Graviton instances.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsMskCluster
metadata:
  name: prod-orderstream
spec:
  kafkaVersion: "3.6.0"
  numberOfBrokerNodes: 6
  instanceType: kafka.m7g.xlarge
  subnetIds:
    - value: subnet-prod-useast1a
    - value: subnet-prod-useast1b
    - value: subnet-prod-useast1c
  vpcId:
    value: vpc-0abc123def456789
  securityGroupIds:
    - value: sg-0orderservice001
    - value: sg-0analyticsservice002
  ebsVolumeSizeGib: 2000
  provisionedThroughputEnabled: true
  provisionedThroughputMbs: 500
  storageMode: TIERED
  kmsKeyArn:
    value: arn:aws:kms:us-east-1:111122223333:key/mrk-abc123def456
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
    log.retention.bytes: "107374182400"
    replica.fetch.max.bytes: "10485760"
    message.max.bytes: "10485760"
  enhancedMonitoring: PER_TOPIC_PER_BROKER
  jmxExporterEnabled: true
  nodeExporterEnabled: true
```

**Key points:**
- 6 brokers across 3 AZs (2 per AZ) for high availability and throughput.
- `kafka.m7g.xlarge` (Graviton) provides better price-performance than `m5.xlarge`.
- Provisioned throughput at 500 MiB/s per broker for consistent write performance.
- `TIERED` storage offloads cold data to S3 — reduces EBS costs for long-retention topics.
- `min.insync.replicas: "2"` with `default.replication.factor: "3"` ensures durability.
- `PER_TOPIC_PER_BROKER` monitoring gives per-topic visibility without the cost of per-partition metrics.

---

## 3. Full logging — all three destinations

Cluster with broker logs delivered simultaneously to CloudWatch Logs, Kinesis Data Firehose, and S3.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsMskCluster
metadata:
  name: audit-kafka
spec:
  kafkaVersion: "3.6.0"
  numberOfBrokerNodes: 3
  instanceType: kafka.m5.large
  subnetIds:
    - value: subnet-audit-az1
    - value: subnet-audit-az2
    - value: subnet-audit-az3
  ebsVolumeSizeGib: 500
  clientBrokerEncryption: TLS
  inClusterEncryption: true
  authentication:
    saslIamEnabled: true
  logging:
    cloudwatchLogs:
      enabled: true
      logGroup:
        value: /aws/msk/audit-kafka
    firehose:
      enabled: true
      deliveryStream:
        value: msk-broker-logs-to-splunk
    s3:
      enabled: true
      bucket:
        value: company-msk-audit-logs
      prefix: audit-kafka/broker-logs/
  enhancedMonitoring: PER_BROKER
```

**Key points:**
- CloudWatch Logs for real-time alerting and dashboards.
- Firehose for forwarding to third-party SIEM (Splunk, Datadog, etc.).
- S3 for long-term archival and compliance.
- `prefix` organizes S3 objects by cluster name.
- All three destinations are independent — each can be enabled/disabled without affecting others.

---

## 4. Multi-auth — IAM + SCRAM + mTLS

Cluster supporting all three authentication methods simultaneously for diverse client populations.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsMskCluster
metadata:
  name: platform-kafka
spec:
  kafkaVersion: "3.6.0"
  numberOfBrokerNodes: 6
  instanceType: kafka.m5.xlarge
  subnetIds:
    - value: subnet-platform-az1
    - value: subnet-platform-az2
    - value: subnet-platform-az3
  vpcId:
    value: vpc-0platform123456
  securityGroupIds:
    - value: sg-0awsservices001
  allowedCidrBlocks:
    - "10.100.0.0/16"
    - "10.200.0.0/16"
  ebsVolumeSizeGib: 1000
  clientBrokerEncryption: TLS
  inClusterEncryption: true
  authentication:
    saslIamEnabled: true
    saslScramEnabled: true
    tlsEnabled: true
    tlsCertificateAuthorityArns:
      - value: arn:aws:acm-pca:us-east-1:111122223333:certificate-authority/ca-abc123
  serverProperties:
    auto.create.topics.enable: "false"
    default.replication.factor: "3"
    min.insync.replicas: "2"
    num.partitions: "6"
    log.retention.hours: "336"
  enhancedMonitoring: PER_TOPIC_PER_BROKER
  jmxExporterEnabled: true
  nodeExporterEnabled: true
```

**Key points:**
- **SASL/IAM (port 9098):** For AWS-native services — Lambda, ECS, EKS pods with IRSA.
- **SASL/SCRAM (port 9096):** For external partners or non-AWS clients using username/password.
- **mTLS (port 9094):** For on-premise services with X.509 certificates signed by a private CA.
- `allowedCidrBlocks` provides network-level access from two VPC peering ranges.
- After deployment, SCRAM secrets must be associated via `aws_msk_scram_secret_association` (separate operation).

---

## 5. With valueFrom references — infrastructure chart pattern

Production cluster wired to other OpenMCF resources using `valueFrom` foreign-key references. This is the recommended pattern for infrastructure charts where resources reference each other by name.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsMskCluster
metadata:
  name: analytics-kafka
spec:
  kafkaVersion: "3.6.0"
  numberOfBrokerNodes: 9
  instanceType: kafka.m7g.2xlarge
  subnetIds:
    - valueFrom:
        kind: AwsVpc
        name: analytics-vpc
        fieldPath: status.outputs.private_subnets.[0].id
    - valueFrom:
        kind: AwsVpc
        name: analytics-vpc
        fieldPath: status.outputs.private_subnets.[1].id
    - valueFrom:
        kind: AwsVpc
        name: analytics-vpc
        fieldPath: status.outputs.private_subnets.[2].id
  vpcId:
    valueFrom:
      kind: AwsVpc
      name: analytics-vpc
      fieldPath: status.outputs.vpc_id
  securityGroupIds:
    - valueFrom:
        kind: AwsSecurityGroup
        name: analytics-consumers-sg
        fieldPath: status.outputs.security_group_id
    - valueFrom:
        kind: AwsSecurityGroup
        name: data-pipeline-sg
        fieldPath: status.outputs.security_group_id
  ebsVolumeSizeGib: 4000
  storageMode: TIERED
  kmsKeyArn:
    valueFrom:
      kind: AwsKmsKey
      name: analytics-encryption-key
      fieldPath: status.outputs.key_arn
  clientBrokerEncryption: TLS
  inClusterEncryption: true
  authentication:
    saslIamEnabled: true
  serverProperties:
    auto.create.topics.enable: "false"
    default.replication.factor: "3"
    min.insync.replicas: "2"
    num.partitions: "24"
    log.retention.hours: "72"
    log.retention.bytes: "-1"
  logging:
    cloudwatchLogs:
      enabled: true
      logGroup:
        valueFrom:
          kind: AwsCloudwatchLogGroup
          name: analytics-kafka-logs
          fieldPath: status.outputs.log_group_name
    s3:
      enabled: true
      bucket:
        valueFrom:
          kind: AwsS3Bucket
          name: analytics-log-archive
          fieldPath: status.outputs.bucket_name
      prefix: kafka/analytics-cluster/
  enhancedMonitoring: PER_TOPIC_PER_PARTITION
  jmxExporterEnabled: true
  nodeExporterEnabled: true
```

**Key points:**
- `valueFrom` references resolve at deployment time from other resources' outputs.
- No hardcoded subnet IDs, VPC IDs, security group IDs, KMS ARNs, or log destinations.
- 9 brokers (3 per AZ) for high-throughput analytics workloads.
- `PER_TOPIC_PER_PARTITION` gives the most granular metrics for partition-level alerting.
- `log.retention.bytes: "-1"` means unlimited retention by size (time-based only).

---

## 6. With inline server_properties and monitoring

Focused example demonstrating Kafka tuning parameters and full observability configuration.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsMskCluster
metadata:
  name: tuned-kafka
spec:
  kafkaVersion: "3.5.1"
  numberOfBrokerNodes: 3
  instanceType: kafka.m5.2xlarge
  subnetIds:
    - value: subnet-tune-az1
    - value: subnet-tune-az2
    - value: subnet-tune-az3
  ebsVolumeSizeGib: 500
  clientBrokerEncryption: TLS
  inClusterEncryption: true
  authentication:
    saslIamEnabled: true
  serverProperties:
    auto.create.topics.enable: "false"
    default.replication.factor: "3"
    min.insync.replicas: "2"
    num.partitions: "6"
    log.retention.hours: "720"
    log.retention.bytes: "53687091200"
    log.segment.bytes: "536870912"
    log.cleanup.policy: "delete"
    compression.type: "lz4"
    replica.fetch.max.bytes: "10485760"
    message.max.bytes: "10485760"
    max.incremental.fetch.session.cache.slots: "2000"
    num.replica.fetchers: "4"
    num.io.threads: "8"
    num.network.threads: "5"
    socket.send.buffer.bytes: "1048576"
    socket.receive.buffer.bytes: "1048576"
  enhancedMonitoring: PER_TOPIC_PER_BROKER
  jmxExporterEnabled: true
  nodeExporterEnabled: true
```

**Key points:**
- `serverProperties` creates an inline MSK Configuration with Kafka server.properties overrides.
- Tuning parameters cover retention, compression, replication fetch size, I/O threads, and network buffers.
- `compression.type: "lz4"` provides a good balance of CPU cost and compression ratio.
- `num.replica.fetchers: "4"` increases parallelism for inter-broker replication.
- JMX Exporter (port 11001) and Node Exporter (port 11002) enable Prometheus-based monitoring.
- `PER_TOPIC_PER_BROKER` metrics are sufficient for most production alerting without the cost of per-partition metrics.
