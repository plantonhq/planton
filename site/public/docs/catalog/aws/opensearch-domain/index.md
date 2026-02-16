---
title: "OpenSearch Domain"
description: "OpenSearch Domain deployment documentation"
icon: "package"
order: 100
componentName: "awsopensearchdomain"
---

# AWS OpenSearch Domain

Deploys an Amazon OpenSearch Service domain with configurable cluster topology, EBS storage, encryption at rest and in transit, optional VPC placement, fine-grained access control, UltraWarm and cold storage tiers, log publishing to CloudWatch, and Auto-Tune optimization.

## What Gets Created

When you deploy an AwsOpenSearchDomain resource, OpenMCF provisions:

- **OpenSearch Domain** — an `opensearch.Domain` with the specified engine version, cluster configuration, and tags derived from metadata (organization, environment, resource kind, resource ID)
- **EBS Volumes** — attached to each data node for index storage, with configurable volume type, size, IOPS, and throughput
- **Encryption at Rest** — optional KMS-based encryption of indices and automated snapshots, using the AWS-managed `aws/es` key or a customer-managed key
- **Node-to-Node Encryption** — optional TLS encryption for all inter-node traffic within the cluster
- **VPC Endpoints** (when `vpcOptions` is set) — ENIs deployed into specified subnets with security groups controlling inbound/outbound traffic on port 443; the domain is not publicly accessible
- **Fine-Grained Access Control** (when `advancedSecurityOptions.enabled` is true) — internal user database or IAM-based authentication with role-based index-level permissions
- **Log Publishing** (when `logPublishingOptions` is set) — CloudWatch Logs delivery for index slow logs, search slow logs, application logs, and audit logs
- **Auto-Tune** (when `autoTuneEnabled` is true) — automatic JVM heap, disk I/O, and performance setting optimization based on cluster metrics

## Prerequisites

- **AWS credentials** configured via environment variables or OpenMCF provider config
- **An OpenSearch or Elasticsearch engine version** string (e.g., `OpenSearch_2.11` or `Elasticsearch_7.10`)
- **(Optional) VPC subnets and security groups** for VPC-based deployment
- **(Optional) A KMS key ARN** for customer-managed encryption at rest
- **(Optional) CloudWatch Logs log group ARNs** for log publishing
- **(Optional) An ACM certificate ARN** for custom domain endpoints

## Quick Start

Create a file `opensearch-domain.yaml`:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsOpenSearchDomain
metadata:
  name: my-search
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AwsOpenSearchDomain.my-search
spec:
  engineVersion: "OpenSearch_2.11"
  clusterConfig:
    instanceType: t3.small.search
  ebsOptions:
    ebsEnabled: true
    volumeType: gp3
    volumeSize: 20
  encryptAtRestEnabled: true
  nodeToNodeEncryptionEnabled: true
```

Deploy:

```shell
openmcf apply -f opensearch-domain.yaml
```

This creates a single-node OpenSearch 2.11 domain with 20 GB gp3 storage, encryption at rest, and node-to-node encryption.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `engineVersion` | `string` | Engine version. Format: `OpenSearch_X.Y` or `Elasticsearch_X.Y`. Changing to an incompatible version forces domain recreation. | Must match `OpenSearch_X.Y` or `Elasticsearch_X.Y`. |
| `clusterConfig` | `object` | Cluster topology configuration. | Required. |
| `clusterConfig.instanceType` | `string` | Instance type for data nodes. Uses the `.search` suffix. | Must be set. Examples: `t3.small.search`, `r6g.large.search`. |
| `ebsOptions` | `object` | EBS volume configuration for data node storage. | Required. |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `clusterConfig.instanceCount` | `int32` | `1` | Number of data node instances. For zone-aware deployments, use a multiple of the availability zone count. |
| `clusterConfig.dedicatedMasterEnabled` | `bool` | `false` | Enable dedicated master nodes for cluster management. Recommended for production. |
| `clusterConfig.dedicatedMasterType` | `string` | — | Instance type for dedicated master nodes. Only used when `dedicatedMasterEnabled` is true. |
| `clusterConfig.dedicatedMasterCount` | `int32` | — | Number of dedicated master nodes. AWS recommends 3 for production. Only used when `dedicatedMasterEnabled` is true. |
| `clusterConfig.zoneAwarenessEnabled` | `bool` | `false` | Distribute data nodes and replicas across multiple Availability Zones. |
| `clusterConfig.availabilityZoneCount` | `int32` | — | Number of AZs (2 or 3). Only used when `zoneAwarenessEnabled` is true. |
| `clusterConfig.warmEnabled` | `bool` | `false` | Enable UltraWarm storage tier for infrequently accessed, read-only data. |
| `clusterConfig.warmType` | `string` | — | Instance type for UltraWarm nodes (e.g., `ultrawarm1.medium.search`). Only used when `warmEnabled` is true. |
| `clusterConfig.warmCount` | `int32` | — | Number of UltraWarm nodes (2–150). Only used when `warmEnabled` is true. |
| `clusterConfig.coldStorageEnabled` | `bool` | `false` | Enable cold storage backed by S3. Requires `warmEnabled` to be true. |
| `clusterConfig.multiAzWithStandbyEnabled` | `bool` | `false` | Enable Multi-AZ with Standby for 99.99% SLA. Requires 3 AZs and at least 3 data nodes. |
| `ebsOptions.ebsEnabled` | `bool` | `false` | Attach EBS volumes to data nodes. Required for most instance types. |
| `ebsOptions.volumeType` | `string` | — | EBS volume type: `gp3` (recommended), `gp2`, `io1`, or `standard`. |
| `ebsOptions.volumeSize` | `int32` | — | Size of each EBS volume in GB. Total storage = `volumeSize` × `instanceCount`. |
| `ebsOptions.iops` | `int32` | — | Provisioned IOPS. Only valid for `gp3` and `io1`. gp3 baseline: 3000. |
| `ebsOptions.throughput` | `int32` | — | Provisioned throughput in MiB/s. Only valid for `gp3`. Minimum: 125. |
| `encryptAtRestEnabled` | `bool` | `false` | Enable encryption at rest for indices and automated snapshots. |
| `kmsKeyId` | `string` | — | Customer-managed KMS key ARN or ID for at-rest encryption. ForceNew. Can reference `AwsKmsKey` via `valueFrom`. |
| `nodeToNodeEncryptionEnabled` | `bool` | `false` | Enable TLS encryption for all inter-node traffic. |
| `vpcOptions.subnetIds` | `string[]` | `[]` | Subnet IDs where OpenSearch deploys ENIs. For zone-aware domains, provide subnets in 2 or 3 AZs. ForceNew. Can reference `AwsVpc` via `valueFrom`. |
| `vpcOptions.securityGroupIds` | `string[]` | `[]` | Security group IDs controlling traffic to the domain. Must allow HTTPS (port 443). Can reference `AwsSecurityGroup` via `valueFrom`. |
| `domainEndpointOptions.enforceHttps` | `bool` | `true` | Require HTTPS for all traffic to the domain endpoint. |
| `domainEndpointOptions.tlsSecurityPolicy` | `string` | — | TLS policy for the HTTPS endpoint (e.g., `Policy-Min-TLS-1-2-PFS-2023-10`). |
| `domainEndpointOptions.customEndpointEnabled` | `bool` | `false` | Enable a custom domain endpoint instead of the AWS-generated endpoint. |
| `domainEndpointOptions.customEndpoint` | `string` | — | FQDN for the custom endpoint (e.g., `search.example.com`). Only used when `customEndpointEnabled` is true. |
| `domainEndpointOptions.customEndpointCertificateArn` | `string` | — | ACM certificate ARN for the custom endpoint. Can reference `AwsCertManagerCert` via `valueFrom`. |
| `advancedSecurityOptions.enabled` | `bool` | `false` | Enable fine-grained access control. ForceNew if disabling. |
| `advancedSecurityOptions.internalUserDatabaseEnabled` | `bool` | `false` | Enable the internal user database for managing users and roles in OpenSearch Dashboards. |
| `advancedSecurityOptions.masterUserArn` | `string` | — | IAM entity ARN designated as master user. Mutually exclusive with `masterUserName`. Can reference via `valueFrom`. |
| `advancedSecurityOptions.masterUserName` | `string` | — | Username for internal user database master user. Mutually exclusive with `masterUserArn`. |
| `advancedSecurityOptions.masterUserPassword` | `string` | — | Password for internal user database master user. Must be 8+ characters with uppercase, lowercase, digit, and special character. Can reference via `valueFrom`. |
| `logPublishingOptions` | `list` | `[]` | Up to 4 log publishing configurations, one per log type. |
| `logPublishingOptions[].logType` | `string` | — | Log type: `INDEX_SLOW_LOGS`, `SEARCH_SLOW_LOGS`, `ES_APPLICATION_LOGS`, or `AUDIT_LOGS`. |
| `logPublishingOptions[].cloudwatchLogGroupArn` | `string` | — | CloudWatch Logs log group ARN. Can reference via `valueFrom`. |
| `logPublishingOptions[].enabled` | `bool` | `true` | Whether this log publishing option is active. |
| `accessPolicies` | `object` | — | IAM-based access policy document (JSON object). Controls who can access the domain and its indices. |
| `autoTuneEnabled` | `bool` | `false` | Enable Auto-Tune to optimize JVM heap, disk I/O, and other settings. |
| `autoSoftwareUpdateEnabled` | `bool` | `false` | Enable automatic service software updates during the off-peak window. |
| `ipAddressType` | `string` | `ipv4` | IP address type: `ipv4` or `dualstack`. Changing from `dualstack` to `ipv4` forces recreation. |
| `advancedOptions` | `map<string,string>` | `{}` | Low-level key-value configuration options (e.g., `rest.action.multi.allow_explicit_index`). |

## Examples

### Production VPC Deployment with Dedicated Masters

A multi-AZ domain deployed into a VPC with dedicated master nodes, zone awareness across 3 AZs, and encryption enabled:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsOpenSearchDomain
metadata:
  name: prod-search
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsOpenSearchDomain.prod-search
spec:
  engineVersion: "OpenSearch_2.11"
  clusterConfig:
    instanceType: r6g.large.search
    instanceCount: 3
    dedicatedMasterEnabled: true
    dedicatedMasterType: r6g.large.search
    dedicatedMasterCount: 3
    zoneAwarenessEnabled: true
    availabilityZoneCount: 3
  ebsOptions:
    ebsEnabled: true
    volumeType: gp3
    volumeSize: 100
    iops: 3000
    throughput: 125
  encryptAtRestEnabled: true
  nodeToNodeEncryptionEnabled: true
  vpcOptions:
    subnetIds:
      - subnet-az1
      - subnet-az2
      - subnet-az3
    securityGroupIds:
      - sg-opensearch
  domainEndpointOptions:
    enforceHttps: true
    tlsSecurityPolicy: Policy-Min-TLS-1-2-PFS-2023-10
  autoTuneEnabled: true
```

### Fine-Grained Access Control with Internal User Database

A domain with FGAC enabled using the internal user database for authentication:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsOpenSearchDomain
metadata:
  name: fgac-search
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: staging.AwsOpenSearchDomain.fgac-search
spec:
  engineVersion: "OpenSearch_2.11"
  clusterConfig:
    instanceType: r6g.large.search
    instanceCount: 2
  ebsOptions:
    ebsEnabled: true
    volumeType: gp3
    volumeSize: 50
  encryptAtRestEnabled: true
  nodeToNodeEncryptionEnabled: true
  advancedSecurityOptions:
    enabled: true
    internalUserDatabaseEnabled: true
    masterUserName: admin
    masterUserPassword: "MyStr0ng!Pass"
  domainEndpointOptions:
    enforceHttps: true
```

### Analytics Domain with UltraWarm and Cold Storage

A domain with hot, warm, and cold storage tiers for cost-effective log analytics:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsOpenSearchDomain
metadata:
  name: analytics-search
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsOpenSearchDomain.analytics-search
spec:
  engineVersion: "OpenSearch_2.11"
  clusterConfig:
    instanceType: r6g.xlarge.search
    instanceCount: 3
    dedicatedMasterEnabled: true
    dedicatedMasterType: r6g.large.search
    dedicatedMasterCount: 3
    zoneAwarenessEnabled: true
    availabilityZoneCount: 3
    warmEnabled: true
    warmType: ultrawarm1.medium.search
    warmCount: 2
    coldStorageEnabled: true
  ebsOptions:
    ebsEnabled: true
    volumeType: gp3
    volumeSize: 200
    iops: 6000
    throughput: 250
  encryptAtRestEnabled: true
  kmsKeyId: arn:aws:kms:us-east-1:123456789012:key/abcd-1234-5678
  nodeToNodeEncryptionEnabled: true
  autoTuneEnabled: true
  autoSoftwareUpdateEnabled: true
```

### Using Foreign Key References

Reference other OpenMCF-managed resources instead of hardcoding IDs:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsOpenSearchDomain
metadata:
  name: ref-search
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsOpenSearchDomain.ref-search
spec:
  engineVersion: "OpenSearch_2.11"
  clusterConfig:
    instanceType: r6g.large.search
    instanceCount: 2
    zoneAwarenessEnabled: true
    availabilityZoneCount: 2
  ebsOptions:
    ebsEnabled: true
    volumeType: gp3
    volumeSize: 80
  encryptAtRestEnabled: true
  kmsKeyId:
    valueFrom:
      kind: AwsKmsKey
      name: search-key
      field: status.outputs.key_arn
  nodeToNodeEncryptionEnabled: true
  vpcOptions:
    subnetIds:
      - valueFrom:
          kind: AwsVpc
          name: main-vpc
          field: status.outputs.private_subnets[0].id
      - valueFrom:
          kind: AwsVpc
          name: main-vpc
          field: status.outputs.private_subnets[1].id
    securityGroupIds:
      - valueFrom:
          kind: AwsSecurityGroup
          name: opensearch-sg
          field: status.outputs.security_group_id
  domainEndpointOptions:
    enforceHttps: true
    customEndpointEnabled: true
    customEndpoint: search.example.com
    customEndpointCertificateArn:
      valueFrom:
        kind: AwsCertManagerCert
        name: search-cert
        field: status.outputs.certificate_arn
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `domain_id` | `string` | Unique identifier assigned to the domain by AWS |
| `domain_name` | `string` | Name of the domain (matches the ID derived from metadata) |
| `domain_arn` | `string` | Amazon Resource Name of the domain, used in IAM policies and cross-service permissions |
| `endpoint` | `string` | Domain endpoint for index, search, and data upload requests. For VPC domains this is a VPC endpoint; for public domains it is internet-accessible |
| `dashboard_endpoint` | `string` | OpenSearch Dashboards UI endpoint (`endpoint/_dashboards`) |

## Related Components

- [AwsVpc](/docs/catalog/aws/vpc) — provides subnets for VPC deployment
- [AwsSecurityGroup](/docs/catalog/aws/security-group) — controls network access to the domain
- [AwsKmsKey](/docs/catalog/aws/kms-key) — provides KMS keys for at-rest encryption
- [AwsCertManagerCert](/docs/catalog/aws/certificate) — provides ACM certificates for custom endpoints
- [AwsRoute53Zone](/docs/catalog/aws/route53-zone) — hosts DNS zones for CNAME records pointing to the domain
