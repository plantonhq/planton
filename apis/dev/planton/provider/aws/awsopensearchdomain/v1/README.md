# AwsOpenSearchDomain

Amazon OpenSearch Service domain (successor to Amazon Elasticsearch Service). Provisions a fully managed search and analytics engine with configurable cluster topology (data nodes, dedicated masters, UltraWarm, cold storage), EBS storage, encryption at rest and in transit, VPC or public deployment, fine-grained access control, log publishing, and Auto-Tune optimization.

## When to Use

- **Full-text search** — product catalogs, document search, autocomplete
- **Log analytics** — centralized logging with Kibana/OpenSearch Dashboards visualization
- **Application monitoring** — metrics, traces, and observability pipelines
- **SIEM** — security information and event management with audit trails
- **Analytics** — interactive exploration of large datasets with aggregations

## When NOT to Use

- **Simple key-value storage** — use DynamoDB or ElastiCache instead
- **Transactional databases** — use RDS, Aurora, or DocumentDB for ACID workloads
- **Data warehousing** — use Redshift or Athena for SQL-based analytics at petabyte scale
- **Message queuing** — use SQS, SNS, or EventBridge for event-driven architectures

## Prerequisites

- **AWS credentials** configured via environment variables or Planton provider config
- **Engine version** in `OpenSearch_X.Y` format (e.g., `OpenSearch_2.11`)
- **(Optional) VPC subnets and security groups** if deploying into a VPC for network isolation
- **(Optional) KMS key** for customer-managed encryption at rest
- **(Optional) CloudWatch log group ARNs** for log publishing
- **(Optional) ACM certificate** for custom domain endpoints

## Quick Start

Create a file `opensearch-domain.yaml`:

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsOpenSearchDomain
metadata:
  name: my-search
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.AwsOpenSearchDomain.my-search
spec:
  engineVersion: "OpenSearch_2.11"
  clusterConfig:
    instanceType: t3.small.search
  ebsOptions:
    ebsEnabled: true
    volumeType: gp3
    volumeSize: 10
  encryptAtRestEnabled: true
  nodeToNodeEncryptionEnabled: true
```

Deploy:

```shell
planton apply -f opensearch-domain.yaml
```

This creates a single-node OpenSearch domain with gp3 storage, encryption at rest, and node-to-node encryption enabled.

## Configuration Reference

### Engine

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `engineVersion` | `string` | Yes | OpenSearch or Elasticsearch engine version. Format: `OpenSearch_X.Y` (e.g., `OpenSearch_2.11`) or `Elasticsearch_X.Y` (e.g., `Elasticsearch_7.10`). |

### Cluster Configuration (`clusterConfig`)

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `instanceType` | `string` | — (required) | Instance type for data nodes. Examples: `t3.small.search` (dev), `r6g.large.search` (prod), `r6g.2xlarge.search` (high-memory). |
| `instanceCount` | `int32` | `1` | Number of data node instances. For zone-aware clusters, use a multiple of the AZ count. |
| `dedicatedMasterEnabled` | `bool` | `false` | Enable dedicated master nodes for cluster stability. Recommended for production. |
| `dedicatedMasterType` | `string` | — | Instance type for dedicated masters. Example: `r6g.large.search`. |
| `dedicatedMasterCount` | `int32` | `0` | Number of dedicated masters. AWS recommends 3 for quorum-based split-brain protection. |
| `zoneAwarenessEnabled` | `bool` | `false` | Distribute data and replicas across multiple Availability Zones. |
| `availabilityZoneCount` | `int32` | `0` | Number of AZs (2 or 3). Only used when `zoneAwarenessEnabled` is true. |
| `warmEnabled` | `bool` | `false` | Enable UltraWarm storage tier for infrequently accessed data. |
| `warmType` | `string` | — | UltraWarm instance type. Examples: `ultrawarm1.medium.search`, `ultrawarm1.large.search`. |
| `warmCount` | `int32` | `0` | Number of UltraWarm nodes (2-150). |
| `coldStorageEnabled` | `bool` | `false` | Enable cold storage backed by S3. Requires UltraWarm to be enabled. |
| `multiAzWithStandbyEnabled` | `bool` | `false` | Enable Multi-AZ with Standby for 99.99% SLA. Requires 3 AZs and 3+ data nodes. |

### EBS Storage (`ebsOptions`)

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `ebsEnabled` | `bool` | — (required) | Attach EBS volumes to data nodes. Required for most instance types. |
| `volumeType` | `string` | — | Volume type: `gp3` (recommended), `gp2`, `io1`, or `standard`. |
| `volumeSize` | `int32` | — | Size per volume in GB. Total storage = `volumeSize * instanceCount`. |
| `iops` | `int32` | `0` | Provisioned IOPS. Only valid for `gp3` (baseline 3000) and `io1`. |
| `throughput` | `int32` | `0` | Provisioned throughput in MiB/s. Only valid for `gp3` (baseline 125 MiB/s, min 125). |

### Encryption

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `encryptAtRestEnabled` | `bool` | `false` (recommended: `true`) | Encrypt indices and snapshots at rest. Uses AWS-managed `aws/es` key unless `kmsKeyId` is set. ForceNew when disabling. |
| `kmsKeyId` | `StringValueOrRef` | — | Customer-managed KMS key ARN for at-rest encryption. ForceNew. Can reference `AwsKmsKey`. |
| `nodeToNodeEncryptionEnabled` | `bool` | `false` (recommended: `true`) | TLS encryption for inter-node traffic. Strongly recommended. |

### Networking — VPC (`vpcOptions`)

| Field | Type | Description |
|-------|------|-------------|
| `subnetIds` | `StringValueOrRef[]` | Subnet IDs for domain ENIs. Match AZ count for zone-aware clusters. Can reference `AwsVpc`. |
| `securityGroupIds` | `StringValueOrRef[]` | Security groups controlling access. Must allow HTTPS (443). Can reference `AwsSecurityGroup`. |

### Domain Endpoint (`domainEndpointOptions`)

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `enforceHttps` | `bool` | `true` | Require HTTPS for all traffic. |
| `tlsSecurityPolicy` | `string` | — | TLS policy. Recommended: `Policy-Min-TLS-1-2-PFS-2023-10`. |
| `customEndpointEnabled` | `bool` | `false` | Use a custom domain (e.g., `search.example.com`). |
| `customEndpoint` | `string` | — | FQDN for the custom endpoint. |
| `customEndpointCertificateArn` | `StringValueOrRef` | — | ACM certificate ARN for the custom endpoint. Can reference `AwsCertManagerCert`. |

### Fine-Grained Access Control (`advancedSecurityOptions`)

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `enabled` | `bool` | `false` | Enable FGAC. ForceNew if disabling (cannot be turned off once enabled). |
| `internalUserDatabaseEnabled` | `bool` | `false` | Enable OpenSearch internal user database for username/password authentication. |
| `masterUserArn` | `StringValueOrRef` | — | IAM user/role ARN as master user. Mutually exclusive with `masterUserName`. |
| `masterUserName` | `string` | — | Internal DB master username. Mutually exclusive with `masterUserArn`. |
| `masterUserPassword` | `StringValueOrRef` | — | Internal DB master password. Min 8 chars with mixed case, digit, and special character. |

### Log Publishing (`logPublishingOptions`)

Repeated block, up to 4 entries (one per log type):

| Field | Type | Description |
|-------|------|-------------|
| `logType` | `string` | `INDEX_SLOW_LOGS`, `SEARCH_SLOW_LOGS`, `ES_APPLICATION_LOGS`, or `AUDIT_LOGS`. |
| `cloudwatchLogGroupArn` | `StringValueOrRef` | CloudWatch Logs log group ARN. |
| `enabled` | `bool` | Default `true`. Set `false` to temporarily disable. |

### Access Policies and Options

| Field | Type | Description |
|-------|------|-------------|
| `accessPolicies` | `Struct` | IAM-based access policy (serialized to JSON). Controls domain and index-level access. |
| `autoTuneEnabled` | `bool` | Enable Auto-Tune for JVM, disk I/O, and performance optimization. |
| `autoSoftwareUpdateEnabled` | `bool` | Allow AWS to apply mandatory and optional software updates. |
| `ipAddressType` | `string` | `ipv4` (default) or `dualstack` (IPv4 + IPv6). Changing from dualstack to ipv4 is ForceNew. |
| `advancedOptions` | `map<string, string>` | Low-level engine options (e.g., `rest.action.multi.allow_explicit_index`). |

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `domain_id` | `string` | Unique identifier assigned to the domain by AWS |
| `domain_name` | `string` | Name of the domain (matches metadata) |
| `domain_arn` | `string` | Amazon Resource Name for IAM policies and cross-service references |
| `endpoint` | `string` | Domain endpoint for index, search, and data upload requests |
| `dashboard_endpoint` | `string` | OpenSearch Dashboards UI endpoint |

## Related Components

- [AwsVpc](/docs/catalog/aws/vpc) — provides subnets for VPC deployment
- [AwsSecurityGroup](/docs/catalog/aws/security-group) — controls network access to the domain
- [AwsKmsKey](/docs/catalog/aws/kms-key) — provides KMS keys for at-rest encryption
- [AwsCertManagerCert](/docs/catalog/aws/certificate-manager-certificate) — provides ACM certificates for custom endpoints
- [AwsRoute53Zone](/docs/catalog/aws/route53-zone) — hosts DNS zones for CNAME records pointing to the domain endpoint

## References

- AWS OpenSearch Service: https://docs.aws.amazon.com/opensearch-service/latest/developerguide/
- OpenSearch Project: https://opensearch.org/
- Instance types: https://docs.aws.amazon.com/opensearch-service/latest/developerguide/supported-instance-types.html
- Best practices: https://docs.aws.amazon.com/opensearch-service/latest/developerguide/bp.html
