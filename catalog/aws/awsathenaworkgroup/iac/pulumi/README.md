# Pulumi Module: AWS Athena Workgroup

Provisions an Amazon Athena workgroup using Pulumi (Go).

## Resources Created

- `athena.Workgroup` — The Athena workgroup carrying the full
  configuration surface: result storage (customer S3 or AWS-managed),
  per-query spend limits, engine selection (SQL or Spark), Identity
  Center integration, S3 Access Grants, and the three monitoring/logging
  arms (CloudWatch, managed, S3).

## How It Works

The module receives an `AwsAthenaWorkgroupIacInput` (the manifest plus
provider credentials), builds the AWS provider through the shared
builder, and renders the workgroup from the spec. Send conditions match
the Terraform module argument-for-argument — the spec's presence
semantics translate to block presence, and the two Optional+Computed
compliance dials (`enforce`/`publish` tri-states and the
minimum-encryption guardrail) are always sent explicitly so they can be
turned off once on.

Result storage is exactly one of customer S3 (`result_configuration`
with `output_location`) or AWS-managed (`managed_query_results`); the
spec validation enforces the exclusivity.

## Outputs

| Name | Description |
|------|-------------|
| `workgroup_arn` | ARN of the workgroup |
| `workgroup_name` | Name of the workgroup |
| `effective_engine_version` | The engine version Athena actually runs (resolved from AUTO) |
