# CloudflareNotificationWebhook Terraform Module

Terraform IaC module for notification webhook destinations.

## Architecture

```
provider.tf   — Cloudflare provider configuration (~> 5.23)
variables.tf  — Input variables mirroring CloudflareNotificationWebhookSpec (generated)
main.tf       — cloudflare_notification_policy_webhooks
outputs.tf    — webhook_id, type
```

## Behavior

A plain CRUD resource: real create/update/delete; only the account forces replacement.

`secret` is write-only at the API: sent on create and update, never returned by any read, so it cannot be drift-detected and an imported destination carries no secret. `type` is a server-side echo Cloudflare infers from the URL -- an output, never an input.

Import as `{account_id}/{webhook_id}`.

## Outputs

| Name | Description |
|------|-------------|
| `webhook_id` | The Cloudflare-assigned UUID (what notification policies reference) |
| `type` | The destination kind Cloudflare inferred from the URL |

## Provider Version

Uses `cloudflare/cloudflare ~> 5.23`.
