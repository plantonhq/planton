# AWS Athena Workgroup — Architecture and Design

## Overview

Amazon Athena is an interactive query service that analyzes data directly in
Amazon S3 using standard SQL. Athena is serverless — there is no infrastructure
to manage, and you pay only for the queries you run.

A **workgroup** is the administrative boundary for Athena. It controls where
query results are stored, how they're encrypted, who pays for data access, and
what engine version runs the queries. Workgroups also enforce cost controls by
setting per-query data scan limits.

## Architecture

### Query Execution Model

```
User/Application
     │
     ▼
┌─────────────┐
│  Athena API  │ ← StartQueryExecution(workgroup=...)
└─────┬───────┘
      │
      ▼
┌──────────────────┐
│  Athena Engine   │ ← Engine version from workgroup config
│  (Trino-based)   │
└─────┬────────────┘
      │
      ├──── reads from ──→ S3 (data), Glue Catalog (schema)
      │
      └──── writes to ──→ S3 (results at output_location)
```

Athena queries read data from S3 via the Glue Data Catalog (table definitions,
partitions, column types). Query results are written to the S3 location defined
by the workgroup's `output_location`.

### Workgroup Isolation Model

Each workgroup provides:

1. **Result isolation** — Separate S3 locations for query results per team or
   application.
2. **Cost isolation** — Per-query scan limits and CloudWatch metrics scoped to
   the workgroup.
3. **Security isolation** — Independent encryption settings and IAM policy
   controls.
4. **Engine isolation** — Pin different workgroups to different engine versions
   for testing upgrades.

### Configuration Enforcement

When `enforce_workgroup_configuration` is `true` (the default), the workgroup
settings override any client-side settings. This means:

- Queries cannot change the result location (even if they specify one).
- Queries cannot change the encryption settings.
- Queries cannot disable CloudWatch metrics.

This is critical for production governance — it prevents individual queries from
bypassing security or cost policies.

## Encryption

### Three Encryption Options

| Option | Description | KMS Key Required | Cost |
|--------|-------------|------------------|------|
| `SSE_S3` | S3-managed encryption keys | No | Free |
| `SSE_KMS` | AWS KMS-managed key | Yes | KMS API costs |
| `CSE_KMS` | Client-side encryption with KMS | Yes | KMS API costs |

**SSE_S3** is the simplest option — Amazon S3 manages the encryption keys with
no additional cost. Suitable for most workloads.

**SSE_KMS** provides key rotation control, CloudTrail audit trail, and
cross-account access via key policy. Required for regulated industries.

**CSE_KMS** encrypts data before it leaves the Athena service. Provides the
strongest protection but is less commonly used.

### Minimum Encryption Configuration

The `enable_minimum_encryption_configuration` flag acts as a compliance guardrail:
when enabled, all queries in the workgroup must produce encrypted results (at
least SSE_S3). Queries that don't specify encryption default to SSE_S3 instead
of writing unencrypted results.

## Cost Model

Athena charges per query based on the amount of data scanned:

- **$5 per TB** of data scanned (standard pricing).
- Cancelled queries are charged for the amount of data scanned before
  cancellation.
- DDL statements (CREATE TABLE, ALTER TABLE, DROP TABLE) are free.
- Failed queries are free.

### Cost Control via `bytes_scanned_cutoff_per_query`

This is the primary cost control mechanism. When a query's planned scan exceeds
the limit, Athena cancels it before execution. The minimum value is 10 MB
(10,485,760 bytes) — this prevents trivially small limits that would break most
useful queries.

**Cost control strategies:**

| Workgroup Type | Recommended Limit | Rationale |
|---------------|-------------------|-----------|
| Development | No limit or 100 GB | Flexibility for exploration |
| Production ETL | 500 GB - 1 TB | Known dataset sizes |
| Ad-hoc analytics | 10 - 50 GB | Prevent accidental full scans |
| Cost-sensitive | 1 - 10 GB | Strict budget control |

### Cost Optimization Tips

1. **Partition data** — Athena only scans partitions that match the WHERE clause.
2. **Use columnar formats** — Parquet and ORC allow Athena to read only needed
   columns.
3. **Compress data** — Gzip, Snappy, or ZSTD reduce data scanned.
4. **Use LIMIT** — For exploration queries, LIMIT reduces output but not scan
   (use WHERE instead).

## Engine Versions

Athena has two major engine generations:

| Engine | Based On | Key Features |
|--------|----------|-------------|
| Athena engine version 2 | Presto 0.217 | Federated queries, UDFs |
| Athena engine version 3 | Trino 410+ | Better performance, MERGE, new functions |
| PySpark engine version 3 | Apache Spark | PySpark notebooks, DataFrames |

When `selected_engine_version` is empty or `"AUTO"`, Athena uses the latest
engine version. Pinning to a specific version is useful for:

- Testing engine upgrades before production rollout.
- Ensuring query behavior consistency across deployments.
- Running Spark workloads (requires explicit engine selection).

## Spark Workgroups

Athena supports Apache Spark workloads via the `execution_role` field. Spark
workgroups differ from SQL workgroups:

- Require an IAM execution role with S3, Glue, and CloudWatch permissions.
- Use PySpark engine version (not Athena engine version).
- Support PySpark notebooks and Spark SQL.
- Have different pricing (DPU-hours instead of per-TB scanned).

## Security

### IAM Policy Patterns

Athena workgroup access is controlled via IAM policies using the workgroup ARN:

```json
{
  "Effect": "Allow",
  "Action": [
    "athena:StartQueryExecution",
    "athena:GetQueryExecution",
    "athena:GetQueryResults"
  ],
  "Resource": "arn:aws:athena:*:*:workgroup/my-workgroup"
}
```

### Cross-Account Access

For cross-account result storage:

1. Set `expected_bucket_owner` to the bucket owner's account ID.
2. Set `s3_acl_option` to `BUCKET_OWNER_FULL_CONTROL`.
3. The bucket policy must allow the Athena workgroup's account to write objects.

## Limits

| Limit | Value |
|-------|-------|
| Workgroups per account per region | 1,000 |
| Concurrent queries per account | 25 (default, can be increased) |
| Query timeout | 30 minutes |
| Query result size | No limit |
| Workgroup name length | 1-128 characters |
| Workgroup name pattern | `[0-9A-Za-z_.-]+` |

## What's Not Included (v1)

| Feature | Reason | Adoption |
|---------|--------|----------|
| Customer content encryption | PySpark user data stores only | <10% |
| Identity Center configuration | Enterprise SSO integration | <5% |
| Managed query results | AWS-managed result storage, newer feature | <10% |
| Monitoring configuration | 3 logging sub-types, complex schema | <15% |

These features may be added in v2 based on demand.
