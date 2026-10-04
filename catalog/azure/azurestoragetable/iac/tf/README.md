# AzureStorageTable - Terraform Module

Terraform implementation for the AzureStorageTable kind.

## Resources Created

- `azurerm_storage_table.main` -- the table, addressed by the parent
  account's ARM id (the control-plane path; the account-name form is
  the provider's legacy data-plane path, removed in azurerm v5)

## Variable Highlights

| Variable | Notes |
| --- | --- |
| `spec.storage_account_id` | The parent account's resolved ARM id; the account NAME (exported as an output) is derived from it in `locals.tf` |
| `spec.table_name` | Letter-start 3-63 alphanumerics; never the literal word "table" |
| `spec.acls` | Stored access policies; table policies REQUIRE the full validity window (start + expiry) |

## Engine Parity Note

This module uses the resource-manager addressing (`storage_account_id`)
while the Pulumi module passes the account name -- the bridge has not
yet picked up the table's RM input (see the `PARITY-EXCEPTION` comment
in `main.tf`). The `table_id` output carries `resource_manager_id` on
BOTH engines, so outputs are byte-identical regardless.

## Usage

```hcl
module "storage_table" {
  source = "./path/to/module"

  metadata = {
    name = "app-entities"
    org  = "mycompany"
  }

  spec = {
    storage_account_id = "/subscriptions/.../providers/Microsoft.Storage/storageAccounts/myappstorage001"
    table_name         = "AppEntities"
  }
}
```

The provider drives table creation and ACLs through the data plane with
shared-key authorization -- the parent account must keep
`shared_access_key_enabled` true (Azure's default). Tables carry no
Azure tags: the platform's identity tags live on the parent account.
