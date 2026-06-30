# GcpFirestoreDatabase - Terraform Module

Terraform implementation for the GcpFirestoreDatabase Planton component.

## Resources Created

- `google_firestore_database` -- Cloud Firestore database

## Usage

```hcl
module "firestore_database" {
  source = "."

  metadata = {
    name = "my-database"
  }

  spec = {
    project_id = {
      value = "my-gcp-project"
    }
    location_id   = "nam5"
    database_name = "(default)"
    type          = "FIRESTORE_NATIVE"
  }
}
```

## Notes

- Firestore databases do not support GCP labels.
- The `database_name`, `location_id`, `database_edition`, and `kms_key_name` fields are immutable after creation.
- The `deletion_policy` is set to `"DELETE"` so that `terraform destroy` actually deletes the database. GCP's default is `"ABANDON"` which only removes from Terraform state.
- Use `delete_protection_state` for GCP API-level deletion protection that works across all interfaces.
- Provider version `~> 6.0` is required for `cmek_config` and `database_edition` support.
- ENTERPRISE edition requires `type = "FIRESTORE_NATIVE"`.
