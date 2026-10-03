# CloudflareSecretsStoreSecret Pulumi Module

Pulumi (Go) IaC module for one secret inside the account Secrets Store.

## Architecture

```
main.go             — Entrypoint loading the IaC input
module/main.go      — Resources(): provider setup, resource, outputs
module/locals.go    — Locals initialization
module/secret.go    — cloudflare.SecretsStoreSecret
module/outputs.go   — output keys
```

## Behavior

Mirrors the Terraform module's contract exactly: the write-only value kept secret in Pulumi state (`pulumi.ToSecret`), scopes passed through in the spec-enforced canonical order, and the `secret_id` / `store_id` outputs. account_id, store_id, and name are create-only; value, scopes, and comment update in place.

## Outputs

| Name | Description |
|------|-------------|
| `secret_id` | The secret's ID within its store |
| `store_id` | The store holding the secret |

## Provider Version

Uses `pulumi-cloudflare` SDK v6 (Cloudflare Terraform provider v5 line).
