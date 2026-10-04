# AzureStorageAccount - Terraform Module

Terraform implementation for the AzureStorageAccount
kind.

## Resources Created

- `azurerm_storage_account.main` -- the account: SKU trio, security
  posture, identity, customer-managed key, firewall, blob/file service
  settings, routing, custom domain, Azure Files authentication, and the
  account-level immutability policy
- `azurerm_storage_management_policy.main` (when `lifecycle_rules` are
  declared) -- the account's SINGLETON blob lifecycle policy document
  (ARM hardcodes its name to "default", which is why the rules fold into
  the account spec)
- `azurerm_storage_account_static_website.main` (when `static_website`
  is declared) -- realized via the standalone resource because azurerm
  removed the inline `static_website` block in v5

## Variable Highlights

| Variable | Notes |
| --- | --- |
| `spec.account_name` | Globally unique, 3-24 lowercase alphanumerics ONLY -- no hyphens |
| `spec.account_kind` / `account_tier` / `replication_type` | Spec enum name strings; unset materializes the documented defaults (StorageV2/Standard/LRS) in `locals.tf` -- azurerm REQUIRES tier and replication |
| `spec.access_tier` / `dns_endpoint_type` | Unset maps to null so Azure computes/defaults them itself |
| `spec.replication_type` | `RA_GRS` → `RAGRS` etc. -- the RA_ prefix collapses in the SKU suffix |
| `spec.network_rules.bypass` | Empty list maps to null so azurerm computes Azure's default (AzureServices) |
| `spec.azure_files_authentication.default_share_level_permission` | `SHARE_PERMISSION_*` enum names mapped to ARM's role-name vocabulary |
| `spec.lifecycle_rules[].actions.*` | Absent aging thresholds are simply not rendered -- the provider's -1 sentinel is never needed |

## Secret-Bearing Outputs

`primary_access_key`, `secondary_access_key`, and the four connection
strings are marked `sensitive` -- they authorize every data-plane
operation on the account. Prefer Entra data-plane roles scoped to
`storage_account_id`; reference the keys only where a consumer genuinely
requires key auth (e.g. a Function App's storage binding).

## Usage

```hcl
module "storage_account" {
  source = "./path/to/module"

  metadata = {
    name = "app-storage"
    org  = "mycompany"
    env  = "production"
  }

  spec = {
    region         = "eastus"
    resource_group = "app-rg"
    account_name   = "myappstorage001"

    replication_type = "ZRS"

    blob_properties = {
      versioning_enabled = true
      delete_retention_policy = {
        days = 14
      }
    }
  }
}
```
