# GcpBigQueryDataset -- Terraform Module

This directory contains the Terraform/OpenTofu implementation for the
GcpBigQueryDataset component.

## Module Structure

```
provider.tf    -- Google provider on the ~> 8.3 line
variables.tf   -- Input variables matching GcpBigQueryDatasetSpec
locals.tf      -- Ambient-project fallback, null-mapping, label merge
main.tf        -- BigQuery API enablement + google_bigquery_dataset resource
outputs.tf     -- Outputs matching GcpBigQueryDatasetStackOutputs
```

## What It Creates

- `google_project_service.bigquery_api` -- enables `bigquery.googleapis.com`
  so a fresh project works first try (`disable_on_destroy = false`)
- `google_bigquery_dataset.this` -- the dataset with location, authoritative
  access control, CMEK, lifecycle defaults, federation, and catalog interop

## Behavioral Notes

- `StringValueOrRef` fields (`project_id`, `kms_key_name`) arrive from the
  proto→tfvars converter as plain resolved strings.
- An empty `project_id` falls back to the provider's default project
  (`GOOGLE_PROJECT` / `GOOGLE_CLOUD_PROJECT` / ADC chain).
- User labels from `spec.labels` merge beneath Planton's `planton-ai_*`
  platform labels — identical order to the Pulumi module.
- Zero/empty optional scalars map to `null` so the API applies its own
  server-side defaults (168h time travel, LOGICAL billing).
- The spec's CEL rules guarantee every access entry's shape (exactly one
  identity; role only on principal grants) before this module runs.

## Outputs

| Output | Description |
|--------|-------------|
| `dataset_id` | Short dataset ID referenced by SQL and downstream tables |
| `self_link` | Fully qualified dataset URI |
| `project` | Containing project (resolved under the ambient fallback) |
| `creation_time` | Creation time in ms since epoch |
| `location` | Dataset location |
| `etag` | Entity tag (changes on every metadata modification) |

## Local Development

```shell
tofu init
planton tofu plan --manifest ../../e2e/manifest.yaml
```
