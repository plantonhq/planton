# AWS Athena Workgroup

Deploys an Amazon Athena workgroup: the resource that isolates query
execution, enforces cost controls, and manages result storage for teams and
applications running interactive SQL (or Apache Spark) analytics over data in
S3, the Glue Data Catalog, and federated sources.

## When to Use

Use an Athena workgroup to:

- **Isolate query results**: Direct query output to a dedicated S3 location
  with its own encryption — or to AWS-managed storage with no bucket at all.
- **Control costs**: Set per-query data scan limits to prevent runaway costs
  from full-table scans on large datasets.
- **Enforce governance**: Lock workgroup configuration so individual queries
  cannot override result locations or encryption settings.
- **Pin engine versions**: Control which Athena engine version runs queries,
  avoiding surprises from automatic upgrades.
- **Run Spark workloads**: Use Apache Spark on Athena with an execution role
  for PySpark notebooks, with session content encrypted under your KMS key.
- **Propagate workforce identity**: Integrate IAM Identity Center so queries
  run and audit as the signed-in user, with S3 Access Grants scoping result
  access per user.
- **Deliver logs**: Publish query/Spark logs to CloudWatch Logs, S3, or
  Athena-managed storage.

## Prerequisites

- An S3 bucket for query results (only if using
  `result_configuration.output_location`; `managed_query_results` needs none)
- An AWS Glue Data Catalog database with tables defined over your S3 data
- A KMS key (if using SSE_KMS/CSE_KMS results, managed-results encryption, or
  Spark content encryption)
- An IAM execution role (Spark and Identity Center workgroups)
- An IAM Identity Center instance (only for `identity_center`)

## Spec Fields

### Top-Level

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `description` | string | — | Console description (max 1024 chars) |
| `state` | string | ENABLED | `ENABLED` or `DISABLED` (pause switch; config and history survive) |
| `result_configuration` | object | — | Customer-managed S3 result storage (see below) |
| `managed_query_results` | object | — | AWS-managed result storage; presence enables it. Mutually exclusive with `output_location` |
| `bytes_scanned_cutoff_per_query` | int64 | 0 (no limit) | Max bytes a single query can scan. 0 or >= 10 MB |
| `enforce_workgroup_configuration` | bool | true | Lock settings so queries can't override them |
| `publish_cloudwatch_metrics_enabled` | bool | true | Publish query metrics to CloudWatch |
| `requester_pays_enabled` | bool | false | Requester pays for S3 data access |
| `enable_minimum_encryption_configuration` | bool | false | Require at least SSE_S3 for all results |
| `selected_engine_version` | string | "" (AUTO) | Athena engine version to use |
| `execution_role` | StringValueOrRef | — | IAM role for Spark / Identity Center workgroups (→ AwsIamRole) |
| `customer_content_encryption_kms_key` | StringValueOrRef | — | KMS key for Spark notebook/session content (→ AwsKmsKey) |
| `identity_center` | object | — | IAM Identity Center trusted identity propagation (create-time) |
| `s3_access_grants` | object | — | S3 Access Grants for query results (pairs with Identity Center) |
| `monitoring` | object | — | Log delivery: CloudWatch / managed / S3 arms, each enabled by presence |
| `force_destroy` | bool | false | Delete named queries on workgroup destroy |

### Result Configuration

| Field | Type | Description |
|-------|------|-------------|
| `output_location` | string | S3 URI for query results (e.g., `s3://bucket/prefix/`) |
| `encryption_option` | string | `SSE_S3`, `SSE_KMS`, or `CSE_KMS` |
| `kms_key_arn` | StringValueOrRef | KMS key for SSE_KMS/CSE_KMS (→ AwsKmsKey) |
| `expected_bucket_owner` | string | AWS account ID for cross-account S3 buckets |
| `s3_acl_option` | string | `BUCKET_OWNER_FULL_CONTROL` for cross-account |

### Monitoring

| Arm | Fields | Notes |
|-----|--------|-------|
| `cloud_watch_logging` | `log_group`, `log_stream_name_prefix`, `log_types` (worker type → streams, e.g. `SPARK_DRIVER` → `STDOUT`) | Searchable, alarmable logs |
| `managed_logging` | `kms_key` | Zero-setup Athena-managed storage |
| `s3_logging` | `log_location` (s3:// URI), `kms_key` | Cheap long-term archive |

### ForceNew Fields

- **Workgroup name** (from `metadata.name`) — cannot be changed after creation.
- **`identity_center`** — both values are fixed at creation; changing them
  replaces the workgroup.

## Outputs

| Output | Description |
|--------|-------------|
| `workgroup_arn` | ARN of the Athena workgroup |
| `workgroup_name` | Name of the Athena workgroup |
| `effective_engine_version` | Actual engine version in use |

## Validation Highlights

- `managed_query_results` and `result_configuration.output_location` are
  mutually exclusive — AWS stores results in exactly one place (enforced at
  manifest validation, mirroring AWS's own plan-time rule).
- `bytes_scanned_cutoff_per_query` must be 0 or at least 10 MB.
- `s3_access_grants.authentication_type` only accepts `DIRECTORY_IDENTITY`
  (the only mode AWS supports today).

## Related Resources

- **AwsS3Bucket** — S3 bucket for query result storage
- **AwsKmsKey** — Customer-managed encryption for results, logs, and Spark content
- **AwsIamRole** — Execution role for Spark workgroups
- **AwsGlueCatalogDatabase** — Data catalog namespace Athena queries against

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
