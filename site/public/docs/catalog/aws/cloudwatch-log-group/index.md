---
title: "CloudWatch Log Group"
description: "CloudWatch Log Group deployment documentation"
icon: "package"
order: 100
componentName: "awscloudwatchloggroup"
---

# AWS CloudWatch Log Group

Deploys an AWS CloudWatch Logs log group with configurable retention policy, optional KMS encryption, and log group class selection. The log group serves as a centralized destination for application logs, service logs, and operational data, and is referenced by many other AWS components including Step Functions, API Gateway, and OpenSearch.

## What Gets Created

- **CloudWatch Log Group** — a container for log streams with the specified retention, encryption, and class settings

## Prerequisites

- An AWS account with credentials configured in the stack input
- An AwsKmsKey resource if enabling customer-managed encryption

## Quick Start

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsCloudwatchLogGroup
metadata:
  name: app-logs
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AwsCloudwatchLogGroup.app-logs
spec:
  retentionInDays: 30
```

```shell
openmcf apply -f log-group.yaml
```

This creates a STANDARD class log group with 30-day retention using default AWS encryption.

## Configuration Reference

### Required Fields

No fields are strictly required. An empty spec creates a STANDARD class log group with indefinite retention and default encryption.

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `retentionInDays` | int | 0 (never expire) | Days to retain log events. Must be one of: 0, 1, 3, 5, 7, 14, 30, 60, 90, 120, 150, 180, 365, 400, 545, 731, 1096, 1827, 2192, 2557, 2922, 3288, 3653. Recommended default: 30. |
| `kmsKeyId` | StringValueOrRef | — | KMS key ARN for encrypting log data at rest. Can reference an AwsKmsKey resource via `valueFrom`. |
| `logGroupClass` | string | STANDARD | Log group class. Valid values: `STANDARD`, `INFREQUENT_ACCESS`, `DELIVERY`. ForceNew — changing requires replacing the log group. |
| `deletionProtectionEnabled` | bool | false | When true, prevents accidental deletion of the log group. Note: not yet implemented in IaC modules due to provider version limitations. |

**Validation rules:**
- `retentionInDays` must be one of the AWS-allowed discrete values listed above
- `logGroupClass` must be `STANDARD`, `INFREQUENT_ACCESS`, or `DELIVERY` when set
- `retentionInDays` must not be set when `logGroupClass` is `DELIVERY` (AWS manages retention for Delivery log groups)

## Examples

### Standard 30-Day Retention

A general-purpose log group for application logging:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsCloudwatchLogGroup
metadata:
  name: app-logs
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: acme
    pulumi.openmcf.org/project: platform
    pulumi.openmcf.org/stack.name: dev.AwsCloudwatchLogGroup.app-logs
spec:
  retentionInDays: 30
```

### Encrypted Production Log Group

A log group with 90-day retention and KMS encryption for compliance workloads:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsCloudwatchLogGroup
metadata:
  name: prod-app-logs
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: acme
    pulumi.openmcf.org/project: platform
    pulumi.openmcf.org/stack.name: prod.AwsCloudwatchLogGroup.prod-app-logs
spec:
  retentionInDays: 90
  kmsKeyId:
    valueFrom:
      kind: AwsKmsKey
      name: log-encryption-key
      fieldPath: status.outputs.key_arn
```

### Infrequent Access for High-Volume Logs

A cost-optimized log group for VPC flow logs or CDN access logs:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsCloudwatchLogGroup
metadata:
  name: vpc-flow-logs
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: acme
    pulumi.openmcf.org/project: networking
    pulumi.openmcf.org/stack.name: prod.AwsCloudwatchLogGroup.vpc-flow-logs
spec:
  retentionInDays: 365
  logGroupClass: INFREQUENT_ACCESS
  kmsKeyId:
    valueFrom:
      kind: AwsKmsKey
      name: infra-key
      fieldPath: status.outputs.key_arn
```

## Stack Outputs

| Output | Type | Description |
|--------|------|-------------|
| `log_group_arn` | string | The ARN of the CloudWatch Log Group. Used by downstream resources (Step Functions, API Gateway, OpenSearch) via `valueFrom`. |
| `log_group_name` | string | The name of the CloudWatch Log Group. Used by services that reference log groups by name (ElastiCache, ECS). |

## Related Components

- [AWS KMS Key](/docs/catalog/aws/kms-key) — Customer-managed encryption key for log data
- [AWS Step Function](/docs/catalog/aws/step-functions) — Uses log group ARN for execution logging
- [AWS HTTP API Gateway](/docs/catalog/aws/http-api-gateway) — Uses log group ARN for access logging
- [AWS OpenSearch Domain](/docs/catalog/aws/opensearch-domain) — Uses log group ARN for slow logs, app logs, and audit logs
