# CloudflareSecretsStore Terraform Module

Terraform IaC module for the account-level Secrets Store.

## Architecture

```
provider.tf   — Cloudflare provider configuration (~> 5.23)
variables.tf  — Input variables mirroring CloudflareSecretsStoreSpec (generated)
main.tf       — cloudflare_secrets_store
outputs.tf    — store_id
```

## Behavior

Both arguments are create-only at the API (the provider's Update is an empty stub; every field forces replacement) -- a name change replaces the store AND every secret inside it. Cloudflare allows one store per account: adopt an existing store by import (`{account_id}/{store_id}`) instead of creating a second. Destroy is a real delete that takes every secret with it.

## Outputs

| Name | Description |
|------|-------------|
| `store_id` | The store's ID -- what secrets, Worker bindings, and AI Gateway authentication reference |

## Provider Version

Uses `cloudflare/cloudflare ~> 5.23`.
