# AWS Athena Workgroup

Deploys an Amazon Athena workgroup that isolates query execution, enforces cost controls, and manages result storage for SQL analytics against S3 data. The workgroup supports configurable result encryption (SSE-S3, SSE-KMS, CSE-KMS), per-query data scan limits, engine version pinning, and optional Spark execution roles. Every KMS key field and the Spark execution role accept ValueFromRef wiring, so an encrypted workgroup composes with AwsKmsKey and AwsIamRole resources in the same InfraChart.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Athena Workgroup** -- a query execution environment with configurable result storage location, encryption settings, cost controls, engine version selection, and CloudWatch metrics publishing
- **AWS Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) applied automatically for tracking and governance

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** -- an active connection in the Connect module with credentials for the target AWS account. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or cross-account trust authentication modes.

### AWS Account

- **An S3 bucket** for query result storage. Configure the S3 URI in `resultConfiguration.outputLocation` (e.g., `s3://my-bucket/athena-results/`). When omitted, each query must specify its own result location.
- **A KMS key** (optional) for encrypting query results when using SSE-KMS or CSE-KMS encryption. Provide the ARN directly or reference an AwsKmsKey Infra Component via ValueFromRef.
- **An IAM role** (optional) for Apache Spark workloads. Required only for workgroups running PySpark notebooks or Spark SQL. Provide the ARN directly or reference an AwsIamRole Infra Component.

## Deploy

### Console

Open the deployment store, find **AWS Athena Workgroup**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Basic SQL Workgroup** preset in the [Presets](#presets) tab to pre-populate a working configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsAthenaWorkgroup
metadata:
  name: analytics
  org: acme-corp
  env: prod
spec:
  region: us-west-2
  resultConfiguration:
    outputLocation: "s3://acme-athena-results/prod/"
  enforceWorkgroupConfiguration: true
  publishCloudwatchMetricsEnabled: true
```

```shell
planton apply -f athena-workgroup.yaml
```

This creates an Athena workgroup that stores query results in the specified S3 location, enforces workgroup-level settings (individual queries cannot override), and publishes execution metrics to CloudWatch. No result encryption or cost controls are configured. An Infra Job tracks the provisioning and streams progress in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the Athena workgroup to a KMS key deployed in the same InfraPipeline:

```yaml
spec:
  resultConfiguration:
    encryptionOption: SSE_KMS
    kmsKeyArn:
      valueFrom:
        kind: AwsKmsKey
        name: analytics-key
        fieldPath: status.outputs.key_arn
  executionRole:
    valueFrom:
      kind: AwsIamRole
      name: spark-execution-role
      fieldPath: status.outputs.role_arn
```

The InfraPipeline resolves the dependency graph, deploys the KMS key and IAM role first, then provisions the Athena workgroup with the resolved values.

## Key Configuration

These are the most important decisions when configuring an Athena workgroup. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Result encryption** -- Set `resultConfiguration.encryptionOption` to `"SSE_S3"` (S3-managed keys, simplest), `"SSE_KMS"` (customer-managed KMS key with CloudTrail audit), or `"CSE_KMS"` (client-side encryption before leaving Athena). SSE-KMS and CSE-KMS require `resultConfiguration.kmsKeyArn`. Enable `enableMinimumEncryptionConfiguration` as a guardrail to ensure no results are ever written unencrypted.

**Cost controls** -- Set `bytesScannedCutoffPerQuery` to a byte limit (minimum 10 MB) to automatically cancel queries that scan too much data. This is the primary cost control mechanism for Athena. Recommended for production to prevent runaway costs from unoptimized queries.

**Configuration enforcement** -- Set `enforceWorkgroupConfiguration: true` (the default) to lock result location, encryption, and other settings at the workgroup level. Individual queries cannot override. Set to `false` for development workgroups where engineers need flexibility.

**Engine version** -- Leave `selectedEngineVersion` empty or set to `"AUTO"` to use the latest engine. Pin to a specific version (e.g., `"Athena engine version 3"`) in production to control upgrade timing. The actual version in use is reported in the `effective_engine_version` output.

**Result storage model** -- A workgroup stores results in exactly one place. `resultConfiguration.outputLocation` targets an S3 bucket you own; `managedQueryResults` (its presence is the enable switch -- an empty block is valid) stores results in AWS-owned storage with 24-hour retention, retrievable only through Athena APIs. AWS rejects a workgroup that combines both.

**Identity integration** -- `identityCenter` enables trusted identity propagation: queries run as the signed-in workforce identity instead of a shared IAM role, making per-user CloudTrail auditing and per-user data grants possible. `s3AccessGrants` (its only supported `authenticationType` is `DIRECTORY_IDENTITY`) scopes result-bucket credentials to the calling identity; `createUserLevelPrefix` isolates each user's results under their own prefix. The Identity Center pair is a create-time setting -- changing it replaces the workgroup; S3 Access Grants stays editable.

**Log delivery** -- `monitoring` carries three combinable arms, each enabled by its presence: `cloudWatchLogging` (searchable operational logs, with per-worker `logTypes` selections for Spark), `managedLogging` (zero-setup Athena-managed storage), and `s3Logging` (cheap long-term archive). Each arm accepts an optional KMS key.

**Operational state** -- `state: DISABLED` rejects new query submissions while keeping configuration, history, and saved queries intact -- the safe way to pause a team's spend. `description` documents which team or application owns the workgroup. For Spark workgroups, `customerContentEncryptionKmsKey` encrypts notebook cells and session data with a customer-managed key.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsKmsKey** (optional) | `resultConfiguration.kmsKeyArn` | `status.outputs.key_arn` |
| **AwsKmsKey** (optional) | `managedQueryResults.kmsKey` | `status.outputs.key_arn` |
| **AwsKmsKey** (optional) | `customerContentEncryptionKmsKey` | `status.outputs.key_arn` |
| **AwsKmsKey** (optional) | `monitoring.managedLogging.kmsKey` | `status.outputs.key_arn` |
| **AwsKmsKey** (optional) | `monitoring.s3Logging.kmsKey` | `status.outputs.key_arn` |
| **AwsIamRole** (optional) | `executionRole` | `status.outputs.role_arn` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `workgroup_arn` | Amazon Resource Name of the workgroup | IAM policies, cross-service permissions |
| `workgroup_name` | Workgroup name for Athena API calls | StartQueryExecution API, application configuration |
| `effective_engine_version` | Actual engine version in use | Compatibility verification, operational monitoring |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Basic SQL workgroup** -- Minimal workgroup with an S3 result location and CloudWatch metrics. No encryption or cost controls. Suitable for development and ad-hoc exploration. Start from the **Basic SQL Workgroup** preset.

**Encrypted production** -- Workgroup with SSE-KMS result encryption, enforced configuration, per-query data scan limits, and CloudWatch metrics. Designed for production environments with compliance requirements. Start from the **Encrypted Production Workgroup** preset.

**Spark workgroup** -- Workgroup configured with an IAM execution role for running PySpark notebooks and Spark SQL queries, plus a customer-managed KMS key for notebook content encryption. Start from the **Spark Workgroup** preset.

**Managed results, zero buckets** -- Workgroup whose results live in AWS-managed storage: nothing to create, secure, or lifecycle, with results retrieved through Athena APIs. Start from the **Managed Results, Zero Bucket** preset.

## Works With

- [**AWS KMS Key**](/infra-catalog/aws-kms-key) -- provides a customer-managed key for query result encryption
- [**AWS IAM Role**](/infra-catalog/aws-iam-role) -- provides an execution role for Apache Spark workloads