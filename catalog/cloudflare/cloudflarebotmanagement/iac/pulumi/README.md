# CloudflareBotManagement Pulumi Module

Pulumi (Go) IaC module for a zone's Bot Management configuration -- the singleton switchboard for automated traffic, from free Bot Fight Mode through Super Bot Fight Mode and Enterprise knobs.

## Architecture

```
main.go                    — Entrypoint loading the IaC input
module/main.go             — Resources(): provider setup, resource, outputs
module/locals.go           — Locals initialization
module/bot_management.go   — cloudflare.BotManagement
module/outputs.go          — output keys
```

## Behavior

Mirrors the Terraform module's contract exactly: unset means unmanaged, destroy is a no-op, `zone_id` output.

## Outputs

| Name | Description |
|------|-------------|
| `zone_id` | The zone whose Bot Management configuration is managed (the singleton's identity) |

## Provider Version

Uses `pulumi-cloudflare` SDK v6 (Cloudflare Terraform provider v5 line).
