# GcpBigQueryDataset -- Pulumi Module

This directory contains the Pulumi Go implementation for the GcpBigQueryDataset component.

## Module Structure

```
module/
  main.go       -- Entry point: creates GCP provider, orchestrates resources
  locals.go     -- Locals struct, GCP label computation
  dataset.go    -- Creates bigquery.Dataset with all field mappings
  outputs.go    -- Output key constants

main.go         -- Pulumi program entrypoint (loads IaC input, calls module)
Pulumi.yaml     -- Pulumi project configuration
Makefile        -- Build, preview, up, destroy targets
```

## Outputs

| Key | Description |
|-----|-------------|
| `dataset_id` | Short dataset ID (used in SQL queries and API calls) |
| `self_link` | Fully qualified dataset URI |
| `project` | GCP project containing the dataset (resolved under the ambient fallback) |
| `creation_time` | Creation timestamp (milliseconds since epoch) |
| `location` | Dataset location |
| `etag` | Entity tag (changes on every metadata modification) |

## Local Development

```bash
make build      # Compile the Pulumi binary
make preview    # Preview changes
make up         # Apply changes
make destroy    # Destroy resources
```

## Notes

- The module enables `bigquery.googleapis.com` before creating the dataset so
  a fresh project works first try.
- An empty `project_id` falls back to the provider's default project (the
  ambient-project contract); outputs read the created resource's resolved
  project.
- User labels from `spec.labels` merge beneath Planton's `planton-ai_*`
  platform labels — identical order to the Terraform module.
- Access entries are mapped from the proto `repeated GcpBigQueryDatasetAccessEntry` to
  Pulumi's `DatasetAccessTypeArray`, preserving all identity types (user, group, domain,
  special group, IAM member) and resource authorizations (view, routine,
  dataset) plus IAM conditions.
- CMEK encryption is configured through the `DefaultEncryptionConfiguration` when
  `kms_key_name` is set. All new tables inherit this encryption.
- The `max_time_travel_hours` field is converted from `int32` to `string` for the
  Pulumi SDK (which expects a string representation).
