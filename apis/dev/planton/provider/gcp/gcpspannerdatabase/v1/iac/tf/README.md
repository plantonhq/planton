# GcpSpannerDatabase - Terraform Module

Terraform implementation for the GcpSpannerDatabase Planton component.

## Resources Created

- `google_spanner_database` -- Cloud Spanner database

## Usage

```hcl
module "spanner_database" {
  source = "."

  metadata = {
    name = "my-database"
  }

  spec = {
    project_id = {
      value = "my-gcp-project"
    }
    instance = {
      value = "my-spanner-instance"
    }
    database_name = "my-database"
  }
}
```

## Notes

- Spanner databases do not support GCP labels.
- The `database_dialect`, `kms_key_name`, and `instance` fields are immutable after creation.
- DDL statements execute atomically with database creation. New statements can be appended, but modifying or removing existing statements forces recreation.
- Terraform's `deletion_protection` is set to `false` by default to allow Planton to manage the lifecycle. Use `enable_drop_protection` for GCP API-level protection.
- Provider version `~> 6.0` is required for consistency with GcpSpannerInstance.
