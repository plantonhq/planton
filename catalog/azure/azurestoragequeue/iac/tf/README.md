# AzureStorageQueue - Terraform Module

Terraform implementation for the AzureStorageQueue kind.

## Resources Created

- `azurerm_storage_queue.main` -- the queue, addressed by the parent
  account's ARM id (the control-plane path; the account-name form is
  the provider's legacy data-plane path, removed in azurerm v5)

## Variable Highlights

| Variable | Notes |
| --- | --- |
| `spec.storage_account_id` | The parent account's resolved ARM id; the account NAME (exported as an output) is derived from it in `locals.tf` |
| `spec.queue_name` | 3-63 lowercase letters/digits/hyphens; becomes the URL path segment |
| `spec.metadata` | Free-form key/value pairs on the queue -- NOT Azure tags |

## Usage

```hcl
module "storage_queue" {
  source = "./path/to/module"

  metadata = {
    name = "work-items"
    org  = "mycompany"
  }

  spec = {
    storage_account_id = "/subscriptions/.../providers/Microsoft.Storage/storageAccounts/myappstorage001"
    queue_name         = "work-items"
  }
}
```

Queues carry no Azure tags: ARM does not support tags on
`queueServices/queues`, so the platform's identity tags live on the
parent account.
