# AWS CloudWatch Log Group

Deploys an AWS CloudWatch Logs log group with configurable retention policy, optional KMS encryption, and log group class selection. The log group serves as a centralized destination for application logs, service logs, and operational data, and is referenced by many other AWS components including Step Functions, API Gateway, and OpenSearch.

## What Gets Created

- **CloudWatch Log Group** — a container for log streams with the specified retention, encryption, and class settings

## Prerequisites

- An AWS account with credentials configured in the stack input
- An AwsKmsKey resource if enabling customer-managed encryption

## Quick Start

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsCloudwatchLogGroup
metadata:
  name: app-logs
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.AwsCloudwatchLogGroup.app-logs
spec:
  region: us-west-2
  retentionInDays: 30
```

```shell
planton apply -f log-group.yaml
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
apiVersion: aws.planton.dev/v1
kind: AwsCloudwatchLogGroup
metadata:
  name: app-logs
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: acme
    pulumi.planton.dev/project: platform
    pulumi.planton.dev/stack.name: dev.AwsCloudwatchLogGroup.app-logs
spec:
  region: us-west-2
  retentionInDays: 30
```

### Encrypted Production Log Group

A log group with 90-day retention and KMS encryption for compliance workloads:

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsCloudwatchLogGroup
metadata:
  name: prod-app-logs
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: acme
    pulumi.planton.dev/project: platform
    pulumi.planton.dev/stack.name: prod.AwsCloudwatchLogGroup.prod-app-logs
spec:
  region: us-west-2
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
apiVersion: aws.planton.dev/v1
kind: AwsCloudwatchLogGroup
metadata:
  name: vpc-flow-logs
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: acme
    pulumi.planton.dev/project: networking
    pulumi.planton.dev/stack.name: prod.AwsCloudwatchLogGroup.vpc-flow-logs
spec:
  region: us-west-2
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

- [AWS KMS Key](/docs/catalog/aws/awskmskey) — Customer-managed encryption key for log data
- [AWS Step Function](/docs/catalog/aws/awsstepfunction) — Uses log group ARN for execution logging
- [AWS HTTP API Gateway](/docs/catalog/aws/awshttpapigateway) — Uses log group ARN for access logging
- [AWS OpenSearch Domain](/docs/catalog/aws/awsopensearchdomain) — Uses log group ARN for slow logs, app logs, and audit logs
