# GcpBigQueryDataset -- Terraform Module

This directory contains the Terraform implementation for the GcpBigQueryDataset component.

## Module Structure

```
provider.tf    -- Google provider configuration
variables.tf   -- Input variables matching GcpBigQueryDatasetSpec
locals.tf      -- Derived values from spec
main.tf        -- google_bigquery_dataset resource with dynamic access blocks
outputs.tf     -- Outputs matching GcpBigQueryDatasetStackOutputs
```

## What It Creates

- `google_bigquery_dataset.this` -- A BigQuery dataset with location, access control,
  encryption, and lifecycle configuration

## Inputs

| Variable | Type | Description |
|----------|------|-------------|
| `spec.project_id.value` | `string` | GCP project ID |
| `spec.dataset_id` | `string` | Dataset identifier |
| `spec.location` | `string` | Geographic location |
| `spec.friendly_name` | `string` | Display name |
| `spec.description` | `string` | Dataset description |
| `spec.default_table_expiration_ms` | `number` | Table auto-expiration (ms) |
| `spec.default_partition_expiration_ms` | `number` | Partition auto-expiration (ms) |
| `spec.max_time_travel_hours` | `number` | Time travel window (48-168) |
| `spec.is_case_insensitive` | `bool` | Case-insensitive names |
| `spec.default_collation` | `string` | Default string collation |
| `spec.storage_billing_model` | `string` | LOGICAL or PHYSICAL |
| `spec.delete_contents_on_destroy` | `bool` | Delete tables on destroy |
| `spec.kms_key_name.value` | `string` | CMEK key name |
| `spec.access` | `list(object)` | Access control entries |
| `provider_config.service_account_key_base64` | `string` | Optional GCP SA key |

## Outputs

| Output | Description |
|--------|-------------|
| `dataset_id` | Short dataset ID |
| `self_link` | Fully qualified dataset URI |
| `project` | GCP project containing the dataset |
| `creation_time` | Creation timestamp (ms since epoch) |

## Usage

```bash
terraform init
terraform plan -var-file=terraform.tfvars.json
terraform apply -var-file=terraform.tfvars.json
```

## Important Notes

- **Access is authoritative.** The `access` dynamic block controls all access entries.
  Entries not in the spec will be removed by BigQuery.
- **Dataset ID restrictions.** Only letters, numbers, and underscores are allowed.
  No hyphens, dots, or spaces.
- **Location is immutable.** Changing location triggers destroy and recreate.
- **Labels are managed by OpenMCF framework.** Framework labels are applied automatically.
