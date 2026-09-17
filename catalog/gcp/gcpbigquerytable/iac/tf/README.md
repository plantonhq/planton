# GcpBigQueryTable — Terraform Module

Provisions a `google_bigquery_table` in the referenced dataset. Enables
`bigquery.googleapis.com` on the target project when needed.

## Inputs

The module receives a `spec` object matching `GcpBigQueryTableSpec` (see
`variables.tf`). References arrive as plain strings via the converter contract.

## Outputs

See `outputs.tf` — `table_id`, `self_link`, `project`, `dataset_id`, `type`,
`location`, `creation_time`, and the pre-assembled dotted `qualified_name`
(`{project}.{dataset}.{table}` — what Pub/Sub BigQuery delivery consumes).

## Provider

Uses `hashicorp/google ~> 8.3` (GA line; dataset and table schemas are
byte-identical on google and google-beta).

## Destroy Parity

`deletion_protection` is set explicitly from the spec (default `true`). Set
`deletionProtection: false` on disposable tables before destroy.
