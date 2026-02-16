# AWS Athena Workgroup

Deploys an Amazon Athena workgroup for managing interactive SQL analytics with
isolated query result storage, cost controls, encryption, and engine version
selection.

## When to Use

Use an Athena workgroup to:

- **Isolate query results**: Direct query output to a dedicated S3 location with
  encryption, separate from other teams or applications.
- **Control costs**: Set per-query data scan limits to prevent runaway costs from
  full-table scans on large datasets.
- **Enforce governance**: Lock workgroup configuration so individual queries
  cannot override result locations or encryption settings.
- **Pin engine versions**: Control which Athena engine version runs queries,
  avoiding surprises from automatic upgrades.
- **Run Spark workloads**: Use Apache Spark on Athena with an execution role for
  PySpark notebooks.

## Prerequisites

- An S3 bucket for storing query results (if setting `result_configuration.output_location`)
- An AWS Glue Data Catalog database with tables defined over your S3 data
- A KMS key (if using SSE_KMS or CSE_KMS encryption for query results)
- An IAM execution role (only if creating a Spark workgroup)

## Spec Fields

### Top-Level

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `result_configuration` | object | — | Query result storage and encryption settings |
| `bytes_scanned_cutoff_per_query` | int64 | 0 (no limit) | Max bytes a single query can scan. 0 or >= 10 MB |
| `enforce_workgroup_configuration` | bool | true | Lock settings so queries can't override them |
| `publish_cloudwatch_metrics_enabled` | bool | true | Publish query metrics to CloudWatch |
| `requester_pays_enabled` | bool | false | Requester pays for S3 data access |
| `enable_minimum_encryption_configuration` | bool | false | Require at least SSE_S3 for all results |
| `selected_engine_version` | string | "" (AUTO) | Athena engine version to use |
| `force_destroy` | bool | false | Delete named queries on workgroup destroy |
| `execution_role` | StringValueOrRef | — | IAM role ARN for Spark workgroups (→ AwsIamRole) |

### Result Configuration

| Field | Type | Description |
|-------|------|-------------|
| `output_location` | string | S3 URI for query results (e.g., `s3://bucket/prefix/`) |
| `encryption_option` | string | `SSE_S3`, `SSE_KMS`, or `CSE_KMS` |
| `kms_key_arn` | StringValueOrRef | KMS key for SSE_KMS/CSE_KMS (→ AwsKmsKey) |
| `expected_bucket_owner` | string | AWS account ID for cross-account S3 buckets |
| `s3_acl_option` | string | `BUCKET_OWNER_FULL_CONTROL` for cross-account |

### ForceNew Fields

- **Workgroup name** (from `metadata.name`) — Cannot be changed after creation.

## Stack Outputs

| Output | Description |
|--------|-------------|
| `workgroup_arn` | ARN of the Athena workgroup |
| `workgroup_name` | Name of the Athena workgroup |
| `effective_engine_version` | Actual engine version in use |

## Deliberately Omitted (v1)

| Feature | Reason |
|---------|--------|
| Customer content encryption | PySpark-specific, <10% usage |
| Identity Center configuration | Enterprise SSO, <5% adoption |
| Managed query results | Newer feature, changes storage model |
| Monitoring configuration (3 types) | Complex nested structure, low adoption |

These can be added in v2 based on demand.

## Related Resources

- **AwsS3Bucket** — S3 bucket for query result storage
- **AwsKmsKey** — Customer-managed encryption for query results
- **AwsIamRole** — Execution role for Spark workgroups
- **AwsGlueCatalogDatabase** — Data catalog for Athena queries (R19, upcoming)
