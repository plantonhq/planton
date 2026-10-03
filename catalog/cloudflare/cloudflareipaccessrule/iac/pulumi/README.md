# CloudflareIpAccessRule Pulumi Module

Pulumi (Go) IaC module for one IP Access rule -- an allow, block, or challenge decision on an IP, IP range, ASN, or country, applied account-wide or to a single zone.

## Architecture

```
main.go                    — Entrypoint loading the IaC input
module/main.go             — Resources(): provider setup, resource, outputs
module/locals.go           — Locals initialization
module/ip_access_rule.go   — cloudflare.AccessRule
module/outputs.go          — output keys
```

## Behavior

Mirrors the Terraform module's contract exactly: exactly-one scope, configuration changes do not stick, `rule_id` / `zone_id` / `account_id` outputs.

## Outputs

| Name | Description |
|------|-------------|
| `rule_id` | The created rule's ID |
| `zone_id` | The zone the rule was created on (empty for account-wide rules) |
| `account_id` | The account the rule was created on (empty for zone-scoped rules) |

## Provider Version

Uses `pulumi-cloudflare` SDK v6 (Cloudflare Terraform provider v5 line).
