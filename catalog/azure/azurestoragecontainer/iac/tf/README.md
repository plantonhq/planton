# AzureStorageContainer - Terraform Module

Terraform implementation for the AzureStorageContainer
kind.

## Resources Created

- `azurerm_storage_container.main` -- the container, addressed by the
  parent account's ARM id (the control-plane path; the account-name form
  is the provider's legacy data-plane path, removed in azurerm v5)

## Variable Highlights

| Variable | Notes |
| --- | --- |
| `spec.storage_account_id` | The parent account's resolved ARM id; the account NAME (exported as an output) is derived from it in `locals.tf` |
| `spec.container_access_type` | Spec enum name strings (PRIVATE/BLOB/CONTAINER) mapped to azurerm's lowercase wire values; unset materializes `private` |
| `spec.default_encryption_scope` | Sent only when non-empty (empty vs omitted differ on ForceNew fields); the override flag rides with it |

## Usage

```hcl
module "storage_container" {
  source = "./path/to/module"

  metadata = {
    name = "app-uploads"
    org  = "mycompany"
  }

  spec = {
    storage_account_id = "/subscriptions/.../providers/Microsoft.Storage/storageAccounts/myappstorage001"
    container_name     = "uploads"
  }
}
```

Containers carry no Azure tags: ARM does not support tags on
`blobServices/containers`, so the platform's identity tags live on the
parent account.
