# AwsMskCluster — Technical Reference

Comprehensive technical documentation for the AwsMskCluster deployment component in Planton. This document covers architecture, networking, authentication, encryption, storage, monitoring, logging, configuration, cost, limits, security, common patterns, and the v2 roadmap.

---

## Table of Contents

1. [Kafka Architecture on MSK](#kafka-architecture-on-msk)
2. [Networking Model](#networking-model)
3. [Authentication Deep Dive](#authentication-deep-dive)
4. [Encryption Model](#encryption-model)
5. [Storage Model](#storage-model)
6. [Monitoring Model](#monitoring-model)
7. [Logging Destinations](#logging-destinations)
8. [MSK Configuration](#msk-configuration)
9. [Cost Model](#cost-model)
10. [Service Limits](#service-limits)
11. [Security Model](#security-model)
12. [Common Patterns](#common-patterns)
13. [v2 Roadmap](#v2-roadmap)

---

## Kafka Architecture on MSK

Amazon MSK provisions and manages Apache Kafka clusters with the following architecture:

### Brokers

- Each broker is an EC2 instance running the Apache Kafka broker process.
- Brokers are distributed across Availability Zones using the subnets provided in `subnetIds`.
- The total `numberOfBrokerNodes` must be a multiple of the number of subnets so that each AZ gets the same number of brokers.
- Each broker has a dedicated EBS volume for log segment storage.
- Broker instance types determine memory, CPU, and maximum number of partitions:
  - `kafka.t3.small` — 2 vCPU, 2 GiB RAM. Development only.
  - `kafka.m5.large` — 2 vCPU, 8 GiB RAM. Low-medium production workloads.
  - `kafka.m5.xlarge` — 4 vCPU, 16 GiB RAM. Medium production workloads.
  - `kafka.m5.2xlarge` — 8 vCPU, 32 GiB RAM. High throughput.
  - `kafka.m5.4xlarge` — 16 vCPU, 64 GiB RAM. Provisioned throughput eligible.
  - `kafka.m7g.large` / `kafka.m7g.xlarge` — Graviton instances with better price-performance.

### ZooKeeper

- MSK provisions and manages a ZooKeeper ensemble automatically.
- ZooKeeper nodes are not visible in the customer's VPC — they run in MSK-managed infrastructure.
- ZooKeeper is used for controller election, topic metadata, consumer group coordination (older protocols), and configuration management.
- Plaintext access on port 2181, TLS access on port 2182.
- ZooKeeper connect strings are exported as outputs (`zookeeper_connect_string`, `zookeeper_connect_string_tls`).
- AWS is moving toward KRaft mode (ZooKeeper-less Kafka) in newer Kafka versions.

### Partitions

- Topics are divided into partitions. Each partition is an ordered, immutable sequence of records.
- Each partition has a leader broker and zero or more follower (replica) brokers.
- The `default.replication.factor` server property controls how many replicas are created for new topics.
- `min.insync.replicas` controls how many replicas must acknowledge a write for it to be considered committed.
- AWS enforces partition limits per broker based on instance type (see [Service Limits](#service-limits)).

---

## Networking Model

### VPC Placement

MSK clusters are deployed entirely within a customer VPC:

- Brokers are placed in the subnets specified by `subnetIds`.
- All subnets must be in the same VPC.
- AWS strongly recommends private subnets (no internet gateway route).
- The ENIs (Elastic Network Interfaces) created by MSK are visible in the customer's VPC.

### Availability Zone Distribution

- Brokers are distributed round-robin across the provided subnets.
- For a 6-broker cluster with 3 subnets: 2 brokers per AZ.
- For a 3-broker cluster with 3 subnets: 1 broker per AZ (minimum for production).
- **Constraint:** `numberOfBrokerNodes` must be a multiple of the subnet count.

### Security Groups

The Planton component supports three security group patterns:

1. **Managed security group from source SGs (`securityGroupIds`):** Creates an SG with ingress rules allowing TCP from each source SG on Kafka ports (9092-9098) and ZooKeeper ports (2181-2182). Requires `vpcId`.

2. **Managed security group from CIDRs (`allowedCidrBlocks`):** Creates an SG with ingress rules allowing TCP from each CIDR on the same ports. Requires `vpcId`.

3. **Direct attachment (`associateSecurityGroupIds`):** Existing SGs attached directly to broker ENIs. No managed SG creation.

All three can be combined. The managed SG (if created) is included alongside any `associateSecurityGroupIds`.

**Important:** The security group list in `broker_node_group_info` is **ForceNew** in the AWS provider. Adding or removing security groups after cluster creation forces cluster replacement.

### Port Assignments

| Port | Protocol | Auth Method |
|------|----------|-------------|
| 9092 | Plaintext | Unauthenticated |
| 9094 | TLS | mTLS or unauthenticated over TLS |
| 9096 | SASL/SCRAM over TLS | SCRAM-SHA-512 |
| 9098 | SASL/IAM over TLS | AWS IAM |
| 2181 | Plaintext | ZooKeeper |
| 2182 | TLS | ZooKeeper |
| 11001 | HTTP | Prometheus JMX Exporter |
| 11002 | HTTP | Prometheus Node Exporter |

---

## Authentication Deep Dive

MSK supports multiple authentication methods that can be enabled simultaneously.

### SASL/IAM (Port 9098)

- **Recommended for most workloads.**
- Clients authenticate using AWS IAM credentials (access keys, role assumption, instance profiles, IRSA for EKS).
- No password rotation or certificate management required.
- Authorization is done via IAM policies (`kafka-cluster:*` actions).
- Client library requirement: `aws-msk-iam-auth` library (Java) or equivalent for other languages.
- Connection string: `bootstrap_brokers_sasl_iam` output.

**IAM policy example:**
```json
{
  "Effect": "Allow",
  "Action": [
    "kafka-cluster:Connect",
    "kafka-cluster:DescribeCluster",
    "kafka-cluster:AlterCluster",
    "kafka-cluster:DescribeTopic",
    "kafka-cluster:CreateTopic",
    "kafka-cluster:WriteData",
    "kafka-cluster:ReadData",
    "kafka-cluster:AlterGroup",
    "kafka-cluster:DescribeGroup"
  ],
  "Resource": [
    "arn:aws:kafka:us-east-1:111122223333:cluster/prod-events/*",
    "arn:aws:kafka:us-east-1:111122223333:topic/prod-events/*",
    "arn:aws:kafka:us-east-1:111122223333:group/prod-events/*"
  ]
}
```

### SASL/SCRAM-SHA-512 (Port 9096)

- Username/password authentication stored in AWS Secrets Manager.
- Useful for non-AWS clients that cannot use IAM.
- Secrets must follow the naming pattern `AmazonMSK_*` and be stored in the same region.
- Secret association with the cluster is done via `aws_msk_scram_secret_association` (not modeled in v1 — see [v2 Roadmap](#v2-roadmap)).
- Connection string: `bootstrap_brokers_sasl_scram` output.

### Mutual TLS / mTLS (Port 9094)

- Clients present X.509 certificates signed by a trusted ACM Private Certificate Authority.
- The trusted CAs are specified in `authentication.tlsCertificateAuthorityArns`.
- Certificate-based identity enables fine-grained ACLs based on the certificate's Distinguished Name.
- Higher operational overhead: certificates must be provisioned, rotated, and revoked.
- Connection string: `bootstrap_brokers_tls` output.

### Unauthenticated (Port 9092)

- No client authentication. Access controlled entirely by network-level security (VPC + security groups).
- Suitable for development environments or tightly controlled VPCs.
- **Not recommended for production.**

### Multi-Auth Considerations

When multiple auth methods are enabled:
- Each method listens on its own port.
- Clients connect to the port corresponding to their auth method.
- The `bootstrap_brokers_*` outputs reflect which endpoints are available.
- A single cluster can serve IAM clients (port 9098), SCRAM clients (port 9096), and mTLS clients (port 9094) simultaneously.

---

## Encryption Model

### At Rest (EBS Volumes)

- All MSK clusters encrypt data at rest using AWS KMS.
- Default: AWS-managed `aws/msk` service key (no cost, no management).
- Custom: Specify `kmsKeyArn` for customer-managed key (CMK). Enables key rotation policies and cross-account access.
- **ForceNew:** Changing the KMS key forces cluster replacement.

### In Transit — Client-to-Broker

Controlled by `clientBrokerEncryption`:

| Value | Behavior | Ports Available |
|-------|----------|-----------------|
| `TLS` (default) | All client traffic encrypted | 9094, 9096, 9098 |
| `TLS_PLAINTEXT` | Both encrypted and plaintext | 9092, 9094, 9096, 9098 |
| `PLAINTEXT` | All client traffic unencrypted | 9092 |

### In Transit — Inter-Broker

Controlled by `inClusterEncryption`:

- `true` (default): All data replicated between brokers is TLS-encrypted.
- `false`: Inter-broker traffic is plaintext.
- **ForceNew:** Changing this forces cluster replacement.
- **Recommendation:** Always enable in production.

---

## Storage Model

### EBS Volumes

- Each broker has a dedicated EBS volume (gp3 by default).
- Size: 1 GiB to 16,384 GiB (16 TiB) per broker.
- Default size is instance-type-specific if `ebsVolumeSizeGib` is not set.
- EBS volumes can be expanded online (no downtime) but cannot be shrunk.

### Provisioned Throughput

- Available on `kafka.m5.4xlarge` and larger instances with ≥10 GiB EBS.
- Provides dedicated throughput (250–2,375 MiB/s) per broker.
- Useful for write-heavy workloads that exceed baseline EBS throughput.
- Enabled via `provisionedThroughputEnabled` and `provisionedThroughputMbs`.

### Tiered Storage

- `storageMode: TIERED` enables tiered storage.
- Hot (recent) data stays on broker EBS volumes.
- Warm (older) data is automatically offloaded to S3 by MSK.
- Reduces EBS costs for topics with long retention periods (days/weeks/months).
- Consumers read seamlessly from both tiers — no application changes required.
- Requires Kafka 2.8.2.tiered or later.
- Tiered storage has its own pricing (per-GB-month for S3 tier + retrieval costs).

---

## Monitoring Model

### CloudWatch Metrics

Controlled by `enhancedMonitoring`:

| Level | Metrics | Use Case |
|-------|---------|----------|
| `DEFAULT` | Cluster and topic level | Basic monitoring, low cost |
| `PER_BROKER` | + Per-broker metrics | Identify hot brokers |
| `PER_TOPIC_PER_BROKER` | + Per-topic per-broker | Production recommended |
| `PER_TOPIC_PER_PARTITION` | + Per-partition metrics | Deep debugging, highest cost |

Key CloudWatch metrics:
- `BytesInPerSec` / `BytesOutPerSec` — throughput per broker.
- `UnderReplicatedPartitions` — indicator of broker or network issues.
- `OfflinePartitionsCount` — critical: means a partition has no leader.
- `CpuUser` / `CpuSystem` — broker CPU utilization.
- `KafkaDataLogsDiskUsed` — EBS utilization percentage.
- `MemoryUsed` — broker heap utilization.

### Prometheus Exporters

Two Prometheus-compatible exporters can be enabled:

**JMX Exporter (`jmxExporterEnabled`, port 11001):**
- Exposes JVM metrics (heap, GC, threads) and Kafka broker-internal metrics.
- Scrape endpoint: `http://<broker-ip>:11001/metrics`.
- Provides finer-grained metrics than CloudWatch (e.g., request handler pool utilization, log flush latency).

**Node Exporter (`nodeExporterEnabled`, port 11002):**
- Exposes host-level metrics: CPU, memory, disk I/O, network I/O.
- Scrape endpoint: `http://<broker-ip>:11002/metrics`.
- Standard Prometheus Node Exporter metrics format.

Both exporters are accessible from within the VPC only. Security groups must allow inbound TCP on ports 11001/11002 from Prometheus scrapers.

---

## Logging Destinations

Broker logs can be delivered to one, two, or all three destinations simultaneously.

### CloudWatch Logs

- Delivers broker logs to a CloudWatch Logs group.
- Enables real-time log search, metric filters, and CloudWatch Alarms.
- Best for: operational alerting, dashboards, short-term analysis.
- Cost: per-GB ingested + per-GB stored.

### Kinesis Data Firehose

- Delivers broker logs to a Firehose delivery stream.
- Firehose can transform and deliver to S3, Redshift, OpenSearch, Splunk, Datadog, etc.
- Best for: forwarding to third-party analytics/SIEM platforms.
- Cost: Firehose per-GB pricing + destination costs.

### S3

- Delivers broker logs directly to an S3 bucket.
- Optional `prefix` organizes objects by cluster/date.
- Best for: long-term archival, compliance, batch analysis with Athena/Spark.
- Cost: S3 storage pricing (typically cheapest for archival).

---

## MSK Configuration

### Server Properties

MSK Configuration holds Apache Kafka `server.properties` overrides. In Planton, there are two ways to configure:

1. **Inline (`serverProperties` map):** Creates an MSK Configuration resource automatically. Each key-value pair becomes a line in the properties file. The configuration is versioned (each update creates a new revision).

2. **External (`configurationArn` + `configurationRevision`):** References a pre-existing MSK Configuration created outside Planton.

These are mutually exclusive — the spec enforces this via cross-field validation.

### Common Tuning Parameters

| Property | Default | Recommended | Purpose |
|----------|---------|-------------|---------|
| `auto.create.topics.enable` | true | **false** | Prevent accidental topic creation |
| `default.replication.factor` | 1 | **3** | Ensure durability across AZs |
| `min.insync.replicas` | 1 | **2** | Require 2 acks for committed writes |
| `num.partitions` | 1 | 6-24 | Default partition count for new topics |
| `log.retention.hours` | 168 (7d) | workload-specific | Time-based retention |
| `log.retention.bytes` | -1 | workload-specific | Size-based retention per partition |
| `log.segment.bytes` | 1073741824 | 536870912 (512 MB) | Smaller segments = faster cleanup |
| `log.cleanup.policy` | delete | delete or compact | Compaction for changelog topics |
| `compression.type` | producer | lz4 or zstd | Broker-side compression |
| `message.max.bytes` | 1048588 | workload-specific | Maximum message size |
| `replica.fetch.max.bytes` | 1048576 | match message.max.bytes | Replication must handle max message |
| `num.io.threads` | 8 | 8-16 | I/O thread pool size |
| `num.network.threads` | 3 | 5-8 | Network request handler pool size |
| `num.replica.fetchers` | 1 | 2-4 | Parallelism for replication |

### Configuration Lifecycle

- Inline configurations are immutable once created. Updates create a new revision.
- Applying a new configuration revision triggers a rolling restart of brokers.
- Some properties are **static** (require broker restart); others are **dynamic** (applied live).
- MSK validates properties before applying — invalid properties cause the update to fail.

---

## Cost Model

MSK pricing has several components:

### Broker Hours

- Charged per broker-hour based on instance type.
- Example: `kafka.m5.large` is approximately $0.21/hour per broker.
- Graviton (`kafka.m7g.*`) instances are typically 10-20% cheaper than equivalent `m5` instances.

### EBS Storage

- Charged per GB-month provisioned.
- Example: approximately $0.10/GB-month for gp3 volumes.
- 6 brokers × 1000 GiB = 6000 GiB × $0.10 = $600/month.

### Provisioned Throughput

- Charged per MiB/s-hour when `provisionedThroughputEnabled` is true.
- Only available on large instance types.

### Tiered Storage

- When `storageMode: TIERED`:
  - Storage cost for data in the S3 tier (per GB-month, cheaper than EBS).
  - Retrieval cost when consumers read from the S3 tier.
  - Reduces EBS costs significantly for long-retention topics.

### Data Transfer

- Data transfer within the same AZ: free.
- Data transfer across AZs: standard AWS cross-AZ pricing (~$0.01/GB each way).
- A 3-AZ cluster with replication factor 3 generates cross-AZ traffic for 2 of 3 replicas.

### CloudWatch Metrics

- Higher `enhancedMonitoring` levels generate more custom metrics.
- `PER_TOPIC_PER_PARTITION` can generate thousands of metrics, each billed at CloudWatch custom metric rates.
- **Recommendation:** Use `PER_TOPIC_PER_BROKER` unless partition-level visibility is required.

### Logging

- CloudWatch Logs: per-GB ingestion + retention.
- Firehose: per-GB processed.
- S3: standard storage pricing.

---

## Service Limits

Key AWS service limits for MSK (as of early 2026):

| Limit | Default | Adjustable |
|-------|---------|------------|
| Clusters per account per region | 30 | Yes |
| Brokers per cluster | 30 | No |
| Minimum brokers per cluster | 1 | No |
| Maximum partitions per broker (m5.large) | 1,000 | No |
| Maximum partitions per broker (m5.4xlarge) | 4,000 | No |
| Maximum partitions per cluster | 30,000 | No |
| Maximum client connections per broker | 3,000+ (instance-dependent) | No |
| Maximum message size | 10 MiB (configurable) | No |
| Maximum EBS volume size | 16,384 GiB (16 TiB) | No |
| Maximum provisioned throughput | 2,375 MiB/s per broker | No |
| Subnets per cluster | 2-3 (ForceNew) | No |
| Security groups per cluster | 5 | No |
| MSK Configurations per account | 100 | Yes |
| ZooKeeper nodes | Managed (not configurable) | No |

---

## Security Model

### IAM Policies for SASL/IAM

With SASL/IAM, authorization is enforced via IAM policies using the `kafka-cluster:*` action namespace:

- `kafka-cluster:Connect` — connect to the cluster.
- `kafka-cluster:DescribeCluster` / `AlterCluster` — cluster metadata.
- `kafka-cluster:CreateTopic` / `DescribeTopic` / `DeleteTopic` — topic management.
- `kafka-cluster:WriteData` / `ReadData` — produce/consume.
- `kafka-cluster:AlterGroup` / `DescribeGroup` — consumer group management.

Resources follow the pattern:
- `arn:aws:kafka:REGION:ACCOUNT:cluster/CLUSTER-NAME/*`
- `arn:aws:kafka:REGION:ACCOUNT:topic/CLUSTER-NAME/*/TOPIC-NAME`
- `arn:aws:kafka:REGION:ACCOUNT:group/CLUSTER-NAME/*/GROUP-NAME`

### SCRAM Secrets

- Stored in AWS Secrets Manager with prefix `AmazonMSK_`.
- Secret value must be JSON: `{"username": "...", "password": "..."}`.
- The secret must be encrypted with a customer-managed KMS key (not the default `aws/secretsmanager` key).
- Secrets are associated with the cluster via `aws_msk_scram_secret_association`.

### ACM Private CA for mTLS

- Create a private CA in AWS Certificate Manager Private CA.
- Issue client certificates from the private CA.
- Add the CA ARN to `authentication.tlsCertificateAuthorityArns`.
- Kafka ACLs use the certificate's Distinguished Name for authorization.
- CA and certificate rotation must be managed externally.

### Network Security

- **VPC isolation:** Brokers are only reachable from within the VPC (unless public access is enabled).
- **Security groups:** Control which IP ranges and services can reach broker ports.
- **Private subnets:** No direct internet access to brokers.
- **Public access (`SERVICE_PROVIDED_EIPS`):** Requires SASL/IAM or SASL/SCRAM + TLS encryption. AWS assigns Elastic IPs to brokers.

---

## Common Patterns

### Event-Driven Architecture

MSK as the central event bus for microservices:
- Services publish domain events to Kafka topics.
- Consumers subscribe to topics for async processing.
- Use SASL/IAM for per-service authorization.
- Set `default.replication.factor: 3` and `min.insync.replicas: 2` for durability.
- Tiered storage for event replay over long time windows.

### CQRS (Command Query Responsibility Segregation)

- Commands are published to Kafka topics.
- Read-side projections consume events and build materialized views.
- Kafka's ordering guarantees (per partition) ensure consistent projections.
- Compacted topics (`log.cleanup.policy: compact`) for latest-state snapshots.

### Change Data Capture (CDC)

- Debezium connectors capture database changes and publish to Kafka.
- MSK acts as the durable, replayable change log.
- Downstream systems (data lakes, search indexes, caches) consume CDC events.
- High partition counts and tiered storage for large CDC volumes.

### Real-Time Analytics Pipeline

- Application events → Kafka → Firehose/Lambda/Flink → data warehouse/S3.
- MSK provides backpressure handling and replay capability.
- `PER_TOPIC_PER_BROKER` monitoring for pipeline health visibility.

### Multi-Tenant Event Platform

- Multiple auth methods for different tenant types.
- IAM for internal AWS services.
- SCRAM for external partners.
- mTLS for high-security tenants.
- Topic naming conventions for tenant isolation.
- IAM policies restrict per-tenant topic access.

---

## v2 Roadmap

Features under consideration for future versions of the AwsMskCluster API:

### VPC Connectivity

- `aws_msk_vpc_connection` resource for cross-VPC access.
- Enables consumers in different VPCs to access the cluster without VPC peering.
- Would be modeled as a nested spec or companion resource.

### SCRAM Secret Management

- `aws_msk_scram_secret_association` to manage SCRAM secrets as part of the cluster lifecycle.
- Automatic secret creation in Secrets Manager.
- Secret rotation integration.

### MSK Serverless

- Separate resource kind (`AwsMskServerlessCluster`) for auto-scaling, pay-per-use Kafka.
- Different API surface — no broker count, instance type, or EBS configuration.
- Serverless supports only SASL/IAM auth.

### Cluster Rebalancing

- `CRUISE_CONTROL` integration for automatic partition rebalancing.
- Useful after broker additions or when partition distribution is uneven.

### Additional Features Under Evaluation

- **MSK Replicator** — cross-region or cross-cluster replication.
- **Cluster policies** — resource-based policies for cross-account access.
- **Custom plugins** — MSK Connect connector management.
- **KRaft mode** — ZooKeeper-less clusters for newer Kafka versions.
