# Terraform Module to Deploy AwsDynamodb

This module provisions an Amazon DynamoDB table -- key schema and
indexes, capacity in any of AWS's three shapes, streams, Global Tables
v2 replication, encryption, recovery, and the table-scoped governance
satellites (resource policy, Kinesis change-data destination,
contributor insights) -- aligned with the Planton API. On provisioned
tables, Application Auto Scaling owns live read/write capacity (user
bounds with target tracking, or pinned min = max targets from the
declared throughput). The KMS key, the Kinesis stream, and the S3
import bucket all attach by reference; the module never creates a
resource that deserves to be its own node.

## CLI (local backend)

```shell
planton tofu init --manifest ../../e2e/manifest.yaml
planton tofu plan --manifest ../../e2e/manifest.yaml
planton tofu apply --manifest ../../e2e/manifest.yaml --auto-approve
planton tofu destroy --manifest ../../e2e/manifest.yaml --auto-approve
```

Credentials are passed via the IaC input through the CLI, not in `spec`.

## Files

- `variables.tf` (generated; do not edit)
- `provider.tf` — provider setup (`hashicorp/aws ~> 6.58`, the catalog's
  shared pessimistic pin)
- `locals.tf` — naming basis, identity tags, key-schema lowering
- `table.tf` — the `aws_dynamodb_table` resource (capacity is
  autoscaler-owned; see autoscaling.tf)
- `satellites.tf` — resource policy, Kinesis streaming destination, and
  contributor insights (table + per-GSI)
- `autoscaling.tf` — Application Auto Scaling targets (both modes),
  target-tracking policies, and scheduled adjustments
- `outputs.tf` — outputs matching `AwsDynamodbOutputs`

## Outputs

`table_name`, `table_arn`, `table_id`, `stream_arn`, `stream_label` --
the stream pair is empty when streams are disabled, matching the Pulumi
module exactly.
