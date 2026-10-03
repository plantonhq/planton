# CloudflareSecretsStore Pulumi Module

Pulumi (Go) IaC module for the account-level Secrets Store.

## Architecture

```
main.go                   — Entrypoint loading the IaC input
module/main.go            — Resources(): provider setup, resource, outputs
module/locals.go          — Locals initialization
module/secrets_store.go   — cloudflare.SecretsStore
module/outputs.go         — output keys
```

## Behavior

Mirrors the Terraform module's contract exactly: both arguments create-only (any change replaces the store and every secret inside it), the one-store-per-account limit, and the `store_id` output.

## Outputs

| Name | Description |
|------|-------------|
| `store_id` | The store's ID -- what secrets, Worker bindings, and AI Gateway authentication reference |

## Provider Version

Uses `pulumi-cloudflare` SDK v6 (Cloudflare Terraform provider v5 line).
