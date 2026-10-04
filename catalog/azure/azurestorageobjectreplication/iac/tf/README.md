# AzureStorageObjectReplication - Terraform Module

Terraform implementation for the AzureStorageObjectReplication
kind.

## Resources Created

- `azurerm_storage_object_replication.main` -- ONE resource managing the
  policy pair Azure materializes on BOTH accounts (destination first --
  which assigns rule IDs -- then the source mirror)

## Variable Highlights

| Variable | Notes |
| --- | --- |
| `spec.source_storage_account_id` / `spec.destination_storage_account_id` | The account pair's resolved ARM ids; both ForceNew |
| `spec.rules` | Container names arrive resolved; `copy_blobs_created_after` passes through (unset lets the provider default OnlyNewObjects apply) |
| `spec.rules[].prefix_match` | The spec's ARM-authentic name for the provider's `filter_out_blobs_with_prefix` -- INCLUDE filters despite that attribute name |

## Usage

```hcl
module "storage_object_replication" {
  source = "./path/to/module"

  metadata = {
    name = "invoices-dr"
    org  = "mycompany"
  }

  spec = {
    source_storage_account_id      = "/subscriptions/.../storageAccounts/primarystorage"
    destination_storage_account_id = "/subscriptions/.../storageAccounts/drstorage"
    rules = [{
      source_container_name      = "invoices"
      destination_container_name = "invoices-replica"
      copy_blobs_created_after   = "Everything"
    }]
  }
}
```

Apply-time prerequisites live on the ACCOUNTS: blob versioning + change
feed on the source, blob versioning on the destination. The policy
carries no Azure tags; the platform's identity tags live on the
accounts.
