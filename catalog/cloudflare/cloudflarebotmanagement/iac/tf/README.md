# CloudflareBotManagement Terraform Module

Terraform IaC module for a zone's Bot Management configuration -- the singleton switchboard for automated traffic, from free Bot Fight Mode through Super Bot Fight Mode and Enterprise knobs.

## Architecture

```
provider.tf   — Cloudflare provider configuration (~> 5.23)
variables.tf  — Input variables mirroring CloudflareBotManagementSpec
main.tf       — cloudflare_bot_management
outputs.tf    — zone_id
```

## Behavior

Unset fields are never sent (the zone keeps its current values). Destroy is a NO-OP at Cloudflare -- the state entry disappears but the live configuration stays as last applied. Plan-gated fields fail at the API; non-entitled zones omit those fields from responses and refresh as drift.

## Outputs

| Name | Description |
|------|-------------|
| `zone_id` | The zone whose Bot Management configuration is managed (the singleton's identity) |

## Provider Version

Uses `cloudflare/cloudflare ~> 5.23`.
