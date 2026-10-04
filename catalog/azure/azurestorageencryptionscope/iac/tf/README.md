# AzureStorageEncryptionScope - Terraform Module

Terraform implementation for the AzureStorageEncryptionScope
kind.

## Resources Created

- `azurerm_storage_encryption_scope.main` -- the scope, addressed by the
  parent account's ARM id (a pure management-plane resource)

## Variable Highlights

| Variable | Notes |
| --- | --- |
| `spec.storage_account_id` | The parent account's resolved ARM id; the account NAME (exported as an output) is derived from it in `locals.tf` |
| `spec.scope_name` | 4-63 alphanumerics, no hyphens -- stricter than other storage names |
| `spec.source` | Spec enum name strings (MICROSOFT_STORAGE/MICROSOFT_KEY_VAULT) mapped to ARM's dotted wire values |
| `spec.key_vault_key_id` | The resolved key URI; sent only when set (the spec enforces required-when-KeyVault) |

## Usage

```hcl
module "storage_encryption_scope" {
  source = "./path/to/module"

  metadata = {
    name = "tenant42-scope"
    org  = "mycompany"
  }

  spec = {
    storage_account_id = "/subscriptions/.../providers/Microsoft.Storage/storageAccounts/myappstorage001"
    scope_name         = "tenant42scope"
    source             = "MICROSOFT_STORAGE"
  }
}
```

Deletion is a SOFT-DISABLE: ARM has no true delete for scopes, so
destroy flips the state to Disabled and the name stays reserved within
the account. Scopes carry no Azure tags: the platform's identity tags
live on the parent account.
