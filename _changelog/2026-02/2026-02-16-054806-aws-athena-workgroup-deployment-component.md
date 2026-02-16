# AWS Athena Workgroup Deployment Component

**Date**: February 16, 2026
**Type**: Feature
**Components**: API Definitions, Pulumi CLI Integration, Provider Framework

## Summary

Added `AwsAthenaWorkgroup` as a new deployment component in OpenMCF, enabling
declarative management of Amazon Athena workgroups with query result storage
isolation, cost controls, encryption, engine version pinning, and Apache Spark
execution support. This is the 22nd new AWS resource kind in the cloud provider
expansion project.

## Problem Statement / Motivation

Teams running SQL analytics on S3 data need isolated workgroup environments with
governed result storage, cost protection against runaway queries, and consistent
encryption. Without declarative workgroup management, teams manually configure
these settings through the AWS console or write ad-hoc IaC, leading to
inconsistent governance across environments.

### Pain Points

- No OpenMCF abstraction for Athena workgroup lifecycle management
- Cost runaway risk from uncontrolled per-query data scanning
- Inconsistent encryption policies across analytics teams
- Manual engine version management across dev/staging/prod workgroups
- No declarative support for Spark-on-Athena workloads

## Solution / What's New

A complete deployment component following the OpenMCF standard:

### Proto API

- `spec.proto` with 9 top-level fields and 1 nested message (`AwsAthenaWorkgroupResultConfig` with 5 fields)
- 3 CEL validations: bytes cutoff range, encryption option values, ACL option values
- `StringValueOrRef` for `kms_key_arn` (→ AwsKmsKey) and `execution_role` (→ AwsIamRole)
- Two `optional` fields with `(org.openmcf.shared.options.default)` annotations for `enforce_workgroup_configuration` (true) and `publish_cloudwatch_metrics_enabled` (true)

### IaC Modules

- **Pulumi**: 4 files (main.go, locals.go, outputs.go, workgroup.go) — single `athena.NewWorkgroup` with conditional configuration blocks
- **Terraform**: 5 files with dynamic blocks for result_configuration, encryption, ACL, and engine_version
- Feature parity between both modules (except `enable_minimum_encryption_configuration` deferred in Pulumi due to SDK v7 limitation)

### Stack Outputs

- `workgroup_arn` — for IAM policies and cross-service references
- `workgroup_name` — for Athena API calls
- `effective_engine_version` — computed by AWS, useful for audit/verification

## Implementation Details

### Spec Design

The Terraform provider wraps all configuration in a `configuration {}` block.
We flatten this wrapper entirely (no semantic value in proto) and keep one
nested message for result configuration since it groups related result storage
settings.

Key design choice: `output_location` is a plain `string`, not `StringValueOrRef`.
S3 URIs include user-defined path prefixes (e.g., `s3://bucket/team-a/queries/`)
that cannot be derived from another resource's outputs.

### CEL Validation Scope

StringValueOrRef fields cannot be validated for presence in CEL expressions
(CEL operates on proto messages where `oneof` variants like `.getValue()` are
Go-only methods). KMS key presence for SSE_KMS/CSE_KMS is documented as a
requirement and enforced at the IaC/AWS API level.

### Deferred Features

| Feature | Reason | Adoption |
|---------|--------|----------|
| Customer content encryption | PySpark-specific | <10% |
| Identity Center configuration | Enterprise SSO | <5% |
| Managed query results | Newer AWS feature | <10% |
| Monitoring configuration (3 types) | Complex schema | <15% |
| `enable_minimum_encryption_configuration` in Pulumi | SDK v7 limitation | Spec-ready |

## Benefits

- **Cost governance**: Per-query scan limits prevent runaway Athena costs
- **Security compliance**: SSE_S3/SSE_KMS/CSE_KMS encryption with KMS key references
- **Team isolation**: Separate result locations and governance per workgroup
- **Engine control**: Pin workgroup to specific Athena or Spark engine versions
- **Cross-resource composition**: `StringValueOrRef` enables infra chart DAG wiring

## Impact

- 32 validation tests passing (21 happy path, 7 failure, 4 envelope)
- Enum 263 registered in `cloud_resource_kind.proto` (Analytics category)
- Catalog page published to site docs
- 3 presets: basic SQL, encrypted production, Spark workgroup
- R19 AwsGlueCatalogDatabase renumbered to 264 (was 263, collision)
- R20 AwsRedshiftCluster renumbered to 265

## Related Work

- Part of the AWS resource expansion project (20260215.02.sp.aws-resource-expansion)
- R17 AwsKinesisFirehose was the previous component (2026-02-15)
- R19 AwsGlueCatalogDatabase is next — foundation for Athena queries
- Data analytics infra chart (T03) will compose Athena + Glue + S3

---

**Status**: Production Ready
**Timeline**: Single session (~1 hour)
