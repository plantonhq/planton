---
title: "Athena Workgroup"
description: "Athena Workgroup deployment documentation"
icon: "package"
order: 100
componentName: "awsathenaworkgroup"
---

# AWS Athena Workgroup

Deploys an Amazon Athena workgroup with configurable query result storage, server-side encryption, per-query cost controls, and optional Apache Spark execution support. The workgroup enforces governance settings so individual queries cannot override result locations or encryption policies.

## What Gets Created

When you deploy an AwsAthenaWorkgroup resource, OpenMCF provisions:

- **Athena Workgroup** — an `aws_athena_workgroup` resource with the specified name, configuration enforcement, engine version, and cost controls
- **Result Configuration** — created only when `resultConfiguration` is set, directs query output to the specified S3 location with optional encryption and ACL settings
- **Engine Version** — created only when `selectedEngineVersion` is set, pins the workgroup to a specific Athena or PySpark engine version

## Prerequisites

- **AWS credentials** configured via environment variables or OpenMCF provider config
- **An S3 bucket** for storing query results (if setting `resultConfiguration.outputLocation`)
- **An AWS Glue Data Catalog** with databases and tables defined over your S3 data sources
- **A KMS key ARN** if using SSE_KMS or CSE_KMS encryption for query results
- **An IAM execution role** only if creating a Spark workgroup (standard SQL workgroups do not need one)

## Quick Start

Create a file `athena-workgroup.yaml`:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsAthenaWorkgroup
metadata:
  name: analytics
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AwsAthenaWorkgroup.analytics
spec:
  region: us-east-1
  resultConfiguration:
    outputLocation: "s3://my-athena-results/analytics/"
```

Deploy:

```shell
openmcf apply -f athena-workgroup.yaml
```

This creates an Athena workgroup named `analytics` with query results stored in S3, configuration enforcement enabled (default), and CloudWatch metrics published (default).

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `region` | `string` | AWS region where the workgroup will be created (e.g., `us-east-1`, `eu-west-1`). | Required; non-empty |

However, most practical deployments set at least `resultConfiguration.outputLocation`.

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `resultConfiguration` | `object` | — | Query result storage and encryption settings |
| `resultConfiguration.outputLocation` | `string` | — | S3 URI where query results are stored (e.g., `s3://bucket/prefix/`) |
| `resultConfiguration.encryptionOption` | `string` | — | `SSE_S3`, `SSE_KMS`, or `CSE_KMS` |
| `resultConfiguration.kmsKeyArn` | `string` | — | KMS key ARN for SSE_KMS/CSE_KMS. Can reference AwsKmsKey resource via `valueFrom` |
| `resultConfiguration.expectedBucketOwner` | `string` | — | AWS account ID for cross-account S3 buckets |
| `resultConfiguration.s3AclOption` | `string` | — | `BUCKET_OWNER_FULL_CONTROL` for cross-account result ownership |
| `bytesScannedCutoffPerQuery` | `int64` | `0` (no limit) | Max bytes a query can scan. Must be 0 or >= 10485760 (10 MB) |
| `enforceWorkgroupConfiguration` | `bool` | `true` | Lock settings so queries cannot override them |
| `publishCloudwatchMetricsEnabled` | `bool` | `true` | Publish query metrics to CloudWatch |
| `requesterPaysEnabled` | `bool` | `false` | Requester pays for S3 data access |
| `enableMinimumEncryptionConfiguration` | `bool` | `false` | Require at least SSE_S3 for all query results |
| `selectedEngineVersion` | `string` | `AUTO` | Athena engine version (`Athena engine version 3`, `PySpark engine version 3`, or `AUTO`) |
| `forceDestroy` | `bool` | `false` | Delete named queries and prepared statements on workgroup destroy |
| `executionRole` | `string` | — | IAM role ARN for Spark workgroups. Can reference AwsIamRole resource via `valueFrom` |

## Examples

### Basic SQL Workgroup

A minimal workgroup directing query results to S3 with all governance defaults.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsAthenaWorkgroup
metadata:
  name: analytics-team
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: analytics
    pulumi.openmcf.org/stack.name: dev.AwsAthenaWorkgroup.analytics-team
spec:
  region: us-east-1
  resultConfiguration:
    outputLocation: "s3://my-athena-results/analytics-team/"
```

### Cost-Controlled with SSE_S3

Workgroup with a 10 GB per-query scan limit and enforced minimum encryption.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsAthenaWorkgroup
metadata:
  name: data-science
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: data
    pulumi.openmcf.org/stack.name: prod.AwsAthenaWorkgroup.data-science
spec:
  region: us-east-1
  resultConfiguration:
    outputLocation: "s3://data-science-results/queries/"
    encryptionOption: SSE_S3
  bytesScannedCutoffPerQuery: 10737418240
  enableMinimumEncryptionConfiguration: true
```

### Production KMS-Encrypted with valueFrom

Production workgroup with SSE_KMS encryption referencing a KMS key from another OpenMCF resource.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsAthenaWorkgroup
metadata:
  name: prod-analytics
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: acme
    pulumi.openmcf.org/project: analytics
    pulumi.openmcf.org/stack.name: prod.AwsAthenaWorkgroup.prod-analytics
spec:
  region: us-east-1
  resultConfiguration:
    outputLocation: "s3://prod-athena-results/queries/"
    encryptionOption: SSE_KMS
    kmsKeyArn:
      valueFrom:
        kind: AwsKmsKey
        name: analytics-encryption-key
        fieldPath: status.outputs.key_arn
  bytesScannedCutoffPerQuery: 53687091200
  enforceWorkgroupConfiguration: true
  publishCloudwatchMetricsEnabled: true
  enableMinimumEncryptionConfiguration: true
  selectedEngineVersion: "Athena engine version 3"
```

## Stack Outputs

| Output | Type | Description |
|--------|------|-------------|
| `workgroup_arn` | `string` | ARN of the Athena workgroup, used for IAM policies and cross-service references |
| `workgroup_name` | `string` | Name of the workgroup, used in Athena API calls (`StartQueryExecution`, etc.) |
| `effective_engine_version` | `string` | Actual engine version in use (resolved from `selectedEngineVersion` or `AUTO`) |

## Related Components

- [AWS S3 Bucket](/docs/catalog/aws/s3-bucket) — S3 bucket for query result storage
- [AWS KMS Key](/docs/catalog/aws/kms-key) — Customer-managed encryption for query results
- [AWS IAM Role](/docs/catalog/aws/iam-role) — Execution role for Spark workgroups
- [AWS CloudWatch Log Group](/docs/catalog/aws/cloudwatch-log-group) — Query execution logging
