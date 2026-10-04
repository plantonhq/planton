# AzureStorageShare - Terraform Module

Terraform implementation for the AzureStorageShare kind.

## Resources Created

- `azurerm_storage_share.main` -- the Azure Files share, addressed by
  the parent account's ARM id (the control-plane path; the account-name
  form is the provider's legacy data-plane path, removed in azurerm v5)

## Variable Highlights

| Variable | Notes |
| --- | --- |
| `spec.storage_account_id` | The parent account's resolved ARM id; the account NAME (exported as an output) is derived from it in `locals.tf` |
| `spec.quota_gb` | Required; standard accounts cap at 5120 GB without the account's `large_file_share_enabled`, premium FileStorage floors at 100 GB (Azure enforces both at apply) |
| `spec.enabled_protocol` | Spec enum name strings (SMB/NFS) -- unset materializes `SMB`; NFS requires a FileStorage account |
| `spec.access_tier` | Spec enum name strings mapped to azurerm's wire values; sent only when chosen so Azure's per-account-kind default applies |
| `spec.acls` | Stored access policies; share policies may leave the validity window to the SAS token |

## Usage

```hcl
module "storage_share" {
  source = "./path/to/module"

  metadata = {
    name = "team-files"
    org  = "mycompany"
  }

  spec = {
    storage_account_id = "/subscriptions/.../providers/Microsoft.Storage/storageAccounts/myappstorage001"
    share_name         = "team-files"
    quota_gb           = 500
  }
}
```

Shares carry no Azure tags: ARM does not support tags on
`fileServices/shares`, so the platform's identity tags live on the
parent account.
