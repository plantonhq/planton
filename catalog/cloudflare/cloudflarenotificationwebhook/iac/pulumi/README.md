# CloudflareNotificationWebhook Pulumi Module

Pulumi (Go) IaC module for notification webhook destinations.

## Architecture

```
main.go                         — iac-input loading + module entry
module/main.go                  — provider setup + resource orchestration
module/locals.go                — metadata/credential references
module/notification_webhook.go  — NotificationPolicyWebhooks
module/outputs.go               — webhook_id, type
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

## SDK Version

Uses `pulumi-cloudflare` v6.19.0.
