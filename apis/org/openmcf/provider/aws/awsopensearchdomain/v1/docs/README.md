# Deploying Amazon OpenSearch Service: From Search Clusters to Production Infrastructure as Code

## Introduction

Amazon OpenSearch Service — the successor to Amazon Elasticsearch Service — is AWS's fully managed platform for search, log analytics, application monitoring, and observability. Unlike a database where you query known keys, OpenSearch inverts the problem: you index documents and then search them using full-text queries, aggregations, and analytics. This architectural distinction is fundamental to understanding when and how to deploy OpenSearch effectively.

The landscape of search infrastructure has changed dramatically. What once required Elasticsearch experts hand-tuning JVM heap sizes, shard counts, and cluster topologies can now be expressed as declarative configuration. But the path from "spin up a domain in the console" to "production-ready IaC" is littered with misconfigurations, cost surprises, and availability pitfalls — many of them unique to OpenSearch's distributed architecture.

This document maps the OpenSearch deployment landscape: cluster architecture, data tiers, networking, security, storage, and how OpenMCF abstracts complexity while preserving the power needed for production workloads.

## OpenSearch Architecture: The Building Blocks

### Data Nodes — The Workhorses

Data nodes store your indices and execute search and indexing operations. Every OpenSearch domain has at least one data node. The instance type you choose determines CPU, memory, and baseline I/O capacity. The attached EBS volume determines how much data each node can hold.

**Instance type selection matters more than you think.** A `t3.small.search` is fine for development, but production workloads need instances with enough memory to hold hot data in the filesystem cache. The general rule: your data nodes should have enough RAM to cache 50-80% of your hot index data. For search-heavy workloads, the `r6g` family (memory-optimized, Graviton2) offers the best price-performance.

**Instance count determines throughput and resilience.** A single data node has no redundancy — if it fails, your domain is unavailable until the node recovers. With 2+ nodes, OpenSearch distributes primary and replica shards across nodes, so losing one node doesn't lose data. For zone-aware clusters, use a data node count that's a multiple of your AZ count (e.g., 3 nodes across 3 AZs, or 6 nodes across 3 AZs).

### Dedicated Master Nodes — Cluster Stability

Dedicated master nodes handle cluster management: tracking which shards live on which nodes, managing index state, and coordinating cluster-wide operations like snapshot creation. They don't store data or serve search requests.

**Why dedicate separate nodes?** In small clusters, data nodes double as master-eligible nodes. Under heavy indexing or search load, the master responsibilities compete with data operations. This can cause cluster instability — delayed shard allocation, split-brain scenarios, or unresponsive cluster state. Dedicated masters isolate these critical operations.

**AWS recommends 3 dedicated masters for production.** Three provides quorum-based split-brain protection: if one master fails, the remaining two can elect a new active master. Using an even number (2 or 4) risks split-brain where neither partition can reach quorum.

**Master node sizing is lighter than data nodes.** Masters don't need large EBS volumes or massive RAM. A `r6g.large.search` is sufficient for most clusters up to ~50 data nodes and ~1000 indices.

### UltraWarm Nodes — Cost-Effective Read-Only Storage

UltraWarm is OpenSearch's warm storage tier for data that is accessed infrequently but must remain queryable. UltraWarm uses S3-backed storage instead of EBS, dramatically reducing per-GB costs while keeping data searchable.

**How UltraWarm works:** When you migrate an index from hot storage to UltraWarm, OpenSearch moves the index data to S3 and caches frequently accessed portions locally. Queries still work, but with higher latency than hot storage. UltraWarm indices are read-only — you cannot index new documents into them.

**Typical use case:** Log analytics where you actively search the last 7-14 days of data (hot tier) but need the last 90 days queryable for investigations (warm tier). The hot tier uses `r6g.xlarge.search` nodes with gp3 EBS; the warm tier uses `ultrawarm1.medium.search` nodes.

**Node count:** UltraWarm requires at least 2 nodes (for availability). The maximum is 150 nodes. Each `ultrawarm1.medium.search` node supports up to 1.5 TB; `ultrawarm1.large.search` supports up to 20 TB.

### Cold Storage — Archive at Minimal Cost

Cold storage is the lowest-cost tier, backed entirely by S3. Cold indices are detached from the cluster and don't consume compute resources. To query cold data, you must first attach the index back to the warm tier.

**Cold storage requires UltraWarm to be enabled.** The data lifecycle is: Hot -> Warm -> Cold. You cannot skip warm and go directly to cold. Cold storage is ideal for data retention policies where data must be retained for years but is rarely accessed (compliance, audit trails).

## Data Tiers: Hot -> Warm -> Cold Lifecycle

OpenSearch's tiered storage model mirrors how organizations use data over time:

| Tier | Storage | Cost | Latency | Write | Use Case |
|------|---------|------|---------|-------|----------|
| **Hot** | EBS (gp3/io1) | Highest | Lowest | Read/Write | Active indexing and frequent queries |
| **Warm** | S3-backed (UltraWarm) | ~80% less | Moderate | Read-only | Infrequent queries, log retention |
| **Cold** | S3 (detached) | Lowest | Must attach first | Read-only (after attach) | Archival, compliance retention |

**Index State Management (ISM) policies** automate this lifecycle. An ISM policy can automatically move an index from hot to warm after 14 days, and from warm to cold after 90 days. This isn't configured in the OpenMCF spec (it's an OpenSearch-level setting), but enabling warm and cold tiers in the spec makes ISM policies possible.

**Cost example:** Storing 1 TB on `r6g.large.search` data nodes with gp3 EBS costs roughly $250-350/month. The same 1 TB on UltraWarm costs roughly $25/month. On cold storage, it's approximately $3/month. For log analytics retaining 90 days of data at 10 GB/day, the tiered approach saves 70-80% compared to keeping everything on hot storage.

## Networking: Public vs VPC Deployment

### Public Domains

When `vpcOptions` is not set, the domain gets a publicly accessible endpoint. Anyone who can resolve the DNS name can reach the domain's HTTPS endpoint. Access control relies entirely on:

1. **Resource-based access policies** — IAM policies attached to the domain controlling who can call which APIs
2. **Fine-grained access control (FGAC)** — Role-based access to specific indices, documents, and fields
3. **IP-based access policies** — Restricting access to specific source IP ranges

**When public is acceptable:** Development/test environments, public-facing search APIs (with FGAC), or when VPC deployment adds complexity without clear benefit. Always enable FGAC and enforce HTTPS on public domains.

### VPC Domains

When `vpcOptions` is configured, OpenSearch deploys Elastic Network Interfaces (ENIs) into the specified subnets. The domain is accessible only from within the VPC (or via VPN/peering/Transit Gateway). Security groups control which sources can reach the domain on port 443.

**Critical: VPC configuration is ForceNew.** Changing from public to VPC, VPC to public, or modifying subnets destroys and recreates the domain. Plan your networking model upfront.

**Subnet selection for zone-aware domains:** Provide subnets in the same number of AZs as your `availabilityZoneCount`. For a 3-AZ domain, provide 3 subnets in 3 different AZs. OpenSearch places ENIs in each subnet.

**Security group best practice:** Allow inbound HTTPS (TCP 443) from your application servers, bastion hosts, or VPN CIDR ranges. Avoid allowing 0.0.0.0/0 even within a VPC — apply the principle of least privilege.

### Networking Recommendation

For production workloads, **always use VPC deployment**. The combination of VPC isolation + security groups + FGAC provides defense in depth. Public deployment is reserved for:
- Development domains where simplicity outweighs security
- Public-facing search APIs where VPC adds latency for internet clients

## Encryption

### At-Rest Encryption

Encryption at rest protects data stored on disk (EBS volumes and S3-backed storage). When `encryptAtRestEnabled` is true, OpenSearch encrypts:
- All data on EBS volumes
- Automated snapshots
- UltraWarm and cold storage data
- Metadata about indices

**AWS-managed key vs customer-managed key:** By default, encryption uses the AWS-managed `aws/es` key. For compliance scenarios requiring key rotation control or cross-account access, specify a customer-managed KMS key via `kmsKeyId`.

**Important: `kmsKeyId` is ForceNew.** The encryption key cannot be changed after domain creation. Choose your key strategy upfront.

### Node-to-Node Encryption

When `nodeToNodeEncryptionEnabled` is true, all traffic between nodes in the cluster is encrypted using TLS. This prevents eavesdropping on inter-node communication (shard replication, cluster state updates, search coordination).

**Recommendation:** Always enable both `encryptAtRestEnabled` and `nodeToNodeEncryptionEnabled` for production. The performance impact is negligible on modern instance types with hardware-accelerated encryption.

### Enforce HTTPS

The `domainEndpointOptions.enforceHttps` field (default: true) ensures clients must use HTTPS to communicate with the domain endpoint. Combined with `tlsSecurityPolicy: "Policy-Min-TLS-1-2-PFS-2023-10"`, this enforces modern TLS standards.

## Access Control

### Resource-Based Policies

The `accessPolicies` field accepts an IAM policy document controlling who can call OpenSearch APIs. For VPC domains, this works alongside security groups. For public domains, this is the primary perimeter control.

**Common pattern:** Allow all authenticated IAM principals within your AWS account:

```json
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Principal": { "AWS": "arn:aws:iam::123456789012:root" },
    "Action": "es:*",
    "Resource": "arn:aws:es:us-east-1:123456789012:domain/my-domain/*"
  }]
}
```

### Fine-Grained Access Control (FGAC)

FGAC provides index-level, document-level, and field-level security. When enabled via `advancedSecurityOptions`, you can:

- **Control index access** — Allow team A to read `logs-*` indices but not `finance-*`
- **Document-level security** — Filter search results so users only see documents matching certain criteria
- **Field-level security** — Hide sensitive fields (e.g., PII) from certain roles
- **Internal user database** — Create users and roles directly in OpenSearch Dashboards
- **IAM integration** — Map IAM roles to OpenSearch roles for seamless AWS authentication

**Two authentication models:**

1. **Internal user database** (`internalUserDatabaseEnabled: true`): Create a `masterUserName`/`masterUserPassword`. Users authenticate with OpenSearch credentials. Simpler to set up, suitable for small teams.

2. **IAM master user** (`masterUserArn`): Designate an IAM role as the master user. All authentication goes through IAM. Better for organizations with centralized identity management.

**Once enabled, FGAC cannot be disabled** without recreating the domain. This is a one-way switch — plan accordingly.

## EBS Storage

### Volume Types

| Type | IOPS | Throughput | Use Case |
|------|------|------------|----------|
| `gp3` | Baseline 3000, configurable up to 16000 | Baseline 125 MiB/s, configurable up to 1000 MiB/s | **Recommended for all workloads.** Predictable performance with independent IOPS and throughput scaling. |
| `gp2` | Burstable, 3 IOPS/GB | Up to 250 MiB/s | Legacy. No reason to choose over gp3 for new deployments. |
| `io1` | Provisioned, up to 64000 | Up to 1000 MiB/s | Extreme IOPS requirements (rare for OpenSearch). Expensive. |
| `standard` | Magnetic | Low | Development only. Not recommended for any production use. |

**gp3 is the default recommendation.** It provides consistent baseline performance (3000 IOPS, 125 MiB/s) and allows you to independently increase IOPS and throughput without increasing volume size. This is a significant advantage over gp2, where IOPS scales with volume size.

### Sizing Guidance

Total storage per node = `volumeSize`. Total cluster storage = `volumeSize * instanceCount`.

**Rule of thumb for sizing:**
- Source data size + 10% for OpenSearch metadata overhead
- Multiply by number of replicas (typically 1 replica = 2x source data)
- Add 20% headroom for merges and temporary spikes
- Divide by number of data nodes

**Example:** 500 GB of source data with 1 replica = 1000 GB. Add 20% headroom = 1200 GB. With 3 data nodes: 400 GB per node. Set `volumeSize: 400`.

### IOPS and Throughput Considerations

For most search workloads, the gp3 baseline (3000 IOPS, 125 MiB/s) is sufficient. Increase IOPS for:
- Heavy indexing workloads (bulk ingestion of logs, metrics)
- Domains with frequent force-merge operations
- Clusters with many small shards and concurrent searches

Increase throughput for:
- Large aggregations that scan significant portions of data
- Shard recovery after node replacement

## Cluster Sizing Guidance

### Small Workloads (Development, Proof of Concept)

```
Data: < 50 GB
Queries: < 100 QPS
```

- 1 data node: `t3.small.search` or `t3.medium.search`
- No dedicated masters
- gp3, 10-50 GB
- No zone awareness

### Medium Workloads (Production Search, Moderate Analytics)

```
Data: 50 GB - 1 TB
Queries: 100-1000 QPS
```

- 3 data nodes: `r6g.large.search` or `r6g.xlarge.search`
- 3 dedicated masters: `r6g.large.search`
- Zone awareness: 3 AZs
- gp3, 100-500 GB per node

### Large Workloads (Heavy Analytics, SIEM, Log Aggregation)

```
Data: 1 TB - 10+ TB
Queries: 1000+ QPS or heavy aggregations
```

- 6-12+ data nodes: `r6g.2xlarge.search` or `r6g.4xlarge.search`
- 3 dedicated masters: `r6g.xlarge.search`
- Zone awareness: 3 AZs
- gp3, 500-2000 GB per node, increased IOPS/throughput
- UltraWarm + cold storage for data lifecycle
- FGAC with audit logging

## Monitoring and Logging

### Log Types

| Log Type | Description | When to Enable |
|----------|-------------|----------------|
| `INDEX_SLOW_LOGS` | Indexing operations exceeding the slow log threshold | When troubleshooting indexing performance |
| `SEARCH_SLOW_LOGS` | Search queries exceeding the slow log threshold | When troubleshooting query latency |
| `ES_APPLICATION_LOGS` | OpenSearch application and error logs | Always in production — captures errors, warnings, and cluster events |
| `AUDIT_LOGS` | Fine-grained access control audit trail | Compliance and security monitoring (requires FGAC) |

**To enable log publishing**, provide CloudWatch Logs log group ARNs in `logPublishingOptions`. Each log type requires its own log group. The CloudWatch log group must have a resource policy allowing OpenSearch to write to it.

**Slow log thresholds** are configured at the index level in OpenSearch (not in the spec). Common settings:

```json
PUT /my-index/_settings
{
  "index.search.slowlog.threshold.query.warn": "10s",
  "index.search.slowlog.threshold.query.info": "5s",
  "index.indexing.slowlog.threshold.index.warn": "10s"
}
```

### Key CloudWatch Metrics

Beyond log publishing, OpenSearch Service publishes metrics to CloudWatch automatically:
- `ClusterStatus.green/yellow/red` — cluster health
- `CPUUtilization` — data node CPU usage
- `JVMMemoryPressure` — JVM heap usage (critical: alert at 80%)
- `FreeStorageSpace` — available EBS storage per node
- `SearchLatency` — average search latency
- `IndexingLatency` — average indexing latency
- `SearchRate` / `IndexingRate` — operations per second

Set CloudWatch alarms on `JVMMemoryPressure` (>80%), `FreeStorageSpace` (<20%), and `ClusterStatus.red`.

## Auto-Tune Optimization

When `autoTuneEnabled` is true, OpenSearch Service automatically adjusts:
- **JVM heap size** — optimizes based on memory pressure patterns
- **Queue sizes** — adjusts search and indexing thread pool queues
- **Cache sizes** — tunes filesystem cache allocation
- **Circuit breaker settings** — prevents out-of-memory errors

Auto-Tune monitors cluster performance and applies changes during the maintenance window (or immediately for non-disruptive changes). For new deployments, enabling Auto-Tune is recommended — it handles tuning that previously required deep Elasticsearch expertise.

## Infra Chart Composability

### Inputs via StringValueOrRef

Several spec fields accept `StringValueOrRef`, enabling runtime composition with other OpenMCF resources:

- `kmsKeyId` — reference an `AwsKmsKey` output (`status.outputs.key_arn`)
- `vpcOptions.subnetIds` — reference `AwsVpc` outputs (`status.outputs.private_subnets.[*].id`)
- `vpcOptions.securityGroupIds` — reference `AwsSecurityGroup` outputs (`status.outputs.security_group_id`)
- `advancedSecurityOptions.masterUserArn` — reference `AwsIamRole` outputs (`status.outputs.role_arn`)
- `domainEndpointOptions.customEndpointCertificateArn` — reference `AwsCertManagerCert` outputs (`status.outputs.certificate_arn`)
- `logPublishingOptions[].cloudwatchLogGroupArn` — reference log group ARNs

This allows building complete infrastructure stacks where the VPC, security groups, KMS keys, and OpenSearch domain are all declared as separate resources with automatic dependency resolution.

### Outputs

The domain exports identifiers and endpoints that downstream resources can consume:

| Output | Downstream Use |
|--------|---------------|
| `domain_arn` | IAM policies granting access to the domain |
| `endpoint` | Application configuration for search/indexing clients |
| `dashboard_endpoint` | Proxy configuration or user documentation |

### DAG Role

In a typical infrastructure DAG, an OpenSearch domain sits downstream of networking (VPC, security groups) and encryption (KMS keys), and upstream of applications that need the search endpoint. Example dependency chain:

```
AwsVpc -> AwsSecurityGroup -> AwsKmsKey -> AwsOpenSearchDomain -> Application Config
```

## Deliberately Omitted Features

The following OpenSearch Service features are **not** included in the current spec. This is intentional — they add complexity without benefiting the majority of users, or they are better managed outside the infrastructure layer:

### Amazon Cognito Authentication
Cognito integration for OpenSearch Dashboards authentication is complex to configure (requires Cognito User Pool, Identity Pool, and IAM roles) and is less common than FGAC with the internal user database or IAM-based authentication. Users who need Cognito can use `advancedOptions` or manage it outside the OpenMCF spec.

### AWS IAM Identity Center (SSO)
SAML-based SSO integration is an organization-level concern typically managed at the identity provider level, not per-domain. The FGAC internal user database or IAM master user patterns cover most authentication needs.

### AI/ML Features
OpenSearch's ML Commons plugin (anomaly detection, k-NN search, semantic search) are index-level features configured within OpenSearch itself, not at the infrastructure provisioning layer. They don't affect domain creation.

### Node Options (node_options)
The `node_options` field for per-node-type configuration is a newer, rarely-used feature. The standard `clusterConfig` fields cover all common topologies.

### Snapshot Options (snapshot_options)
Automated snapshot hour configuration is a legacy setting. Modern OpenSearch Service handles automated snapshots without user intervention.

### Off-Peak Window
The off-peak window configuration is managed automatically by AWS and rarely needs explicit configuration.

## Conclusion

OpenSearch Service is a powerful but complex managed service. The domain spec captures the essential 80% of configuration — cluster topology, storage, encryption, networking, access control, and observability — while leaving advanced tuning to the OpenSearch engine itself.

For most deployments, the pattern is straightforward:
1. Choose your deployment model (public vs VPC)
2. Size your cluster (instance type, count, zone awareness)
3. Enable encryption (at-rest, node-to-node, enforce HTTPS)
4. Configure access control (FGAC with internal DB or IAM)
5. Enable monitoring (log publishing, Auto-Tune)
6. Optionally add warm/cold tiers for cost-effective data retention

The presets provided with this component encode these patterns as copy-and-customize starting points, from single-node development to production clusters with tiered storage. Infrastructure as Code transforms OpenSearch deployment from tribal knowledge into repeatable, auditable, and reviewable configuration.
