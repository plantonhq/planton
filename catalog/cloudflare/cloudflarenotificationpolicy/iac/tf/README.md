# CloudflareNotificationPolicy Terraform Module

Terraform IaC module for notification policies.

## Architecture

```
provider.tf   — Cloudflare provider configuration (~> 5.23)
variables.tf  — Input variables mirroring CloudflareNotificationPolicySpec (generated)
main.tf       — cloudflare_notification_policy
outputs.tf    — policy_id
```

## Behavior

A plain CRUD resource: real create/update/delete; only the account forces replacement.

The spec flattens each delivery mechanism to its identity (an email address or a UUID); this module rebuilds the API's `{id}` object rows for email, PagerDuty, and webhook destinations. Every filter is a list of strings and only declared filters are sent, so the payload holds exactly the fields the alert type reads. `enabled` is sent only when set -- Cloudflare's default is true.

Import as `{account_id}/{policy_id}`.

## Outputs

| Name | Description |
|------|-------------|
| `policy_id` | The Cloudflare-assigned UUID of the policy |

## Provider Version

Uses `cloudflare/cloudflare ~> 5.23`.
