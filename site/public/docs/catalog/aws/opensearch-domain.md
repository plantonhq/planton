---
title: "OpenSearch Domain"
description: "OpenSearch Domain deployment documentation"
icon: "package"
order: 100
componentName: "awsopensearchdomain"
---

# AWS OpenSearch Domain

Deploys an Amazon OpenSearch Service domain (successor to Amazon Elasticsearch Service) with configurable cluster topology, EBS storage, encryption at rest and in transit, VPC or public deployment, fine-grained access control, UltraWarm and cold storage tiers, log publishing, and Auto-Tune optimization.

## What Gets Created

When you deploy an AwsOpenSearchDomain resource, OpenMCF provisions:

- **OpenSearch Domain** — an `opensearch.Domain` with the specified engine version, cluster configuration, encryption settings, and access control policies
- **EBS Volumes** — gp3/gp2/io1 volumes attached to each data node for index storage, with configurable IOPS and throughput
- **VPC Endpoints** (when `vpcOptions` is set) — ENIs deployed into specified subnets, with security groups controlling access on port 443
- **Fine-Grained Access Control** (when `advancedSecurityOptions` is enabled) — internal user database or IAM-based authentication with role-based index-level permissions
- **Log Publishing** (when `logPublishingOptions` is set) — CloudWatch Logs delivery for slow logs, application logs, and audit logs

## Prerequisites

- **AWS credentials** configured via environment variables or OpenMCF provider config
- **An OpenSearch engine version** (e.g., `OpenSearch_2.11`)
- **(Optional) VPC subnets and security groups** for VPC deployment
- **(Optional) A KMS key ARN** for customer-managed encryption at rest
- **(Optional) CloudWatch log group ARNs** for log publishing
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
    volumeSize: 10
  encryptAtRestEnabled: true
  nodeToNodeEncryptionEnabled: true
```

Deploy:

```shell
openmcf apply -f opensearch-domain.yaml
```

This creates a single-node OpenSearch domain with gp3 storage, encryption at rest, and node-to-node encryption.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `engineVersion` | `string` | Engine version to deploy. | Must match `OpenSearch_X.Y` or `Elasticsearch_X.Y` format. |
| `clusterConfig.instanceType` | `string` | Instance type for data nodes. | Must be set. Examples: `t3.small.search`, `r6g.large.search`. |
| `ebsOptions.ebsEnabled` | `bool` | Attach EBS volumes to data nodes. | Required for most instance types. |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `clusterConfig.instanceCount` | `int32` | `1` | Number of data node instances. |
| `clusterConfig.dedicatedMasterEnabled` | `bool` | `false` | Enable dedicated master nodes. |
| `clusterConfig.zoneAwarenessEnabled` | `bool` | `false` | Distribute across multiple AZs. |
| `clusterConfig.warmEnabled` | `bool` | `false` | Enable UltraWarm storage tier. |
| `clusterConfig.coldStorageEnabled` | `bool` | `false` | Enable cold storage (requires warm). |
| `ebsOptions.volumeType` | `string` | — | `gp3` (recommended), `gp2`, `io1`, `standard`. |
| `ebsOptions.volumeSize` | `int32` | — | Size per volume in GB. |
| `encryptAtRestEnabled` | `bool` | `false` | Encrypt data at rest. |
| `nodeToNodeEncryptionEnabled` | `bool` | `false` | TLS for inter-node traffic. |
| `vpcOptions` | object | — | VPC deployment (subnets + security groups). ForceNew. |
| `advancedSecurityOptions` | object | — | Fine-grained access control. ForceNew if disabling. |
| `logPublishingOptions` | list | `[]` | CloudWatch log publishing (up to 4 types). |
| `autoTuneEnabled` | `bool` | `false` | Auto-Tune performance optimization. |

## Examples

See [examples.md](../../awsopensearchdomain/v1/examples.md) for complete YAML manifests covering:
- Minimal single-node (public, no VPC)
- Production VPC with dedicated masters
- FGAC with internal user database
- FGAC with IAM master user
- Analytics with warm + cold storage
- Custom endpoint with ACM certificate
- Log publishing to CloudWatch

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `domain_id` | `string` | Unique identifier assigned to the domain by AWS |
| `domain_name` | `string` | Name of the domain |
| `domain_arn` | `string` | Amazon Resource Name for IAM policies |
| `endpoint` | `string` | Domain endpoint for search and indexing requests |
| `dashboard_endpoint` | `string` | OpenSearch Dashboards UI endpoint |

## Related Components

- [AwsVpc](/docs/catalog/aws/vpc) — provides subnets for VPC deployment
- [AwsSecurityGroup](/docs/catalog/aws/security-group) — controls network access to the domain
- [AwsKmsKey](/docs/catalog/aws/kms-key) — provides KMS keys for at-rest encryption
- [AwsCertManagerCert](/docs/catalog/aws/certificate-manager-certificate) — provides ACM certificates for custom endpoints
- [AwsRoute53Zone](/docs/catalog/aws/route53-zone) — hosts DNS zones for CNAME records
