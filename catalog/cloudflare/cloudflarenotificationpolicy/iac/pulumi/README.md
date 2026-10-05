# CloudflareNotificationPolicy Pulumi Module

Pulumi (Go) IaC module for notification policies.

## Architecture

```
main.go                        — iac-input loading + module entry
module/main.go                 — provider setup + resource orchestration
module/locals.go               — metadata/credential references
module/notification_policy.go  — NotificationPolicy (mechanisms + filters mapping)
module/outputs.go              — policy_id
```

## Behavior

A plain CRUD resource: real create/update/delete; only the account forces replacement.

The spec flattens each delivery mechanism to its identity (an email address or a UUID); this module rebuilds the SDK's `{Id}` object rows for the Emails, Pagerduties, and Webhooks arrays. Every filter is a string array and only declared filters are sent, so the payload holds exactly the fields the alert type reads. `enabled` is sent only when set -- Cloudflare's default is true.

Import as `{account_id}/{policy_id}`.

## Outputs

| Name | Description |
|------|-------------|
| `policy_id` | The Cloudflare-assigned UUID of the policy |

## SDK Version

Uses `pulumi-cloudflare` v6.19.0.
