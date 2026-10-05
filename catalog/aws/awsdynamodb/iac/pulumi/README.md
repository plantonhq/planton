# Pulumi Module to Deploy AwsDynamodb

This module provisions an Amazon DynamoDB table -- key schema and
indexes, capacity in any of AWS's three shapes, streams, Global Tables
v2 replication, encryption, recovery, and the table-scoped governance
satellites (resource policy, Kinesis change-data destination,
contributor insights) -- aligned with the Planton API. On provisioned
tables, Application Auto Scaling owns live read/write capacity (user
bounds with target tracking, or pinned min = max targets from the
declared throughput; see `module/autoscaling.go`). The KMS key, the
Kinesis stream, and the S3 import bucket all attach by reference; the
module never creates a resource that deserves to be its own node.

## CLI

```bash
# Preview
planton pulumi preview \
  --manifest ../../e2e/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir .

# Update (apply)
planton pulumi update \
  --manifest ../../e2e/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir . \
  --yes

# Destroy
planton pulumi destroy \
  --manifest ../../e2e/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir .
```

Credentials are passed via the IaC input through the CLI, not in `spec`.

## Files

- `main.go` — entrypoint loading the IaC input
- `module/main.go` — orchestration and output exports
- `module/locals.go` — naming basis and identity tags
- `module/table.go` — the `dynamodb.Table` resource
- `module/satellites.go` — resource policy, Kinesis streaming
  destination, and contributor insights (table + per-GSI)
- `module/outputs.go` — output-key constants matching
  `AwsDynamodbOutputs`

## Outputs

`table_name`, `table_arn`, `table_id`, `stream_arn`, `stream_label` --
the stream pair is empty when streams are disabled, matching the
Terraform module exactly.
