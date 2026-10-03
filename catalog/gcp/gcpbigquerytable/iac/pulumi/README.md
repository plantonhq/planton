# GcpBigQueryTable -- Pulumi Module

This directory contains the Pulumi Go implementation for the GcpBigQueryTable component.

## Module Structure

```
module/
  main.go                          -- Entry point: creates GCP provider, orchestrates resources
  locals.go                        -- Locals struct, GCP label computation
  table.go                         -- Creates bigquery.Table with all field mappings
  external_data_configuration.go   -- Maps the external-table arm's nested options
  outputs.go                       -- Output key constants

main.go         -- Pulumi program entrypoint (loads IaC input, calls module)
Pulumi.yaml     -- Pulumi project configuration
Makefile        -- Build, preview, up, destroy targets
```

## Outputs

| Key | Description |
|-----|-------------|
| `table_id` | Short table ID (referenced by SQL and foreign keys) |
| `self_link` | Fully qualified table URI |
| `project` | GCP project containing the table (resolved under the ambient fallback) |
| `dataset_id` | Parent dataset ID |
| `type` | TABLE, VIEW, MATERIALIZED_VIEW, or EXTERNAL |
| `location` | Table location (inherited from the dataset) |
| `creation_time` | Creation timestamp (milliseconds since epoch) |
| `qualified_name` | Dotted `{project}.{dataset}.{table}` handle (what Pub/Sub BigQuery delivery consumes) |

## Local Development

```bash
make build      # Compile the Pulumi binary
make preview    # Preview changes
make up         # Apply changes
make destroy    # Destroy resources
```

## Notes

- The module enables `bigquery.googleapis.com` before creating the table so
  a fresh project works first try.
- An empty `project_id` falls back to the provider's default project (the
  ambient-project contract); outputs read the created resource's resolved
  project.
- User labels from `spec.labels` merge beneath Planton's `planton-ai_*`
  platform labels — identical order to the Terraform module.
- `deletion_protection` is sent explicitly with a default of TRUE (the
  destroy-parity contract with the Terraform module): a destroy fails until
  the spec sets it false.
- The four table arms (native / view / materialized view / external) map to
  their provider blocks only when present; the spec's CEL rules guarantee
  their mutual exclusivity before this module runs.
- The view arm always sends `use_legacy_sql` explicitly so the BigQuery
  API's legacy-SQL-by-default behavior for views never silently applies.
