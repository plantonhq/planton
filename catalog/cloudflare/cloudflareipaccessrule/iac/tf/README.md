# CloudflareIpAccessRule Terraform Module

Terraform IaC module for one IP Access rule -- an allow, block, or challenge decision on an IP, IP range, ASN, or country, applied account-wide or to a single zone.

## Architecture

```
provider.tf   — Cloudflare provider configuration (~> 5.23)
variables.tf  — Input variables mirroring CloudflareIpAccessRuleSpec
main.tf       — cloudflare_access_rule
outputs.tf    — rule_id, zone_id, account_id
```

## Behavior

Exactly one of `account_id` / `zone_id` is sent (the spec's CEL guarantees it). Only `mode` and `notes` update in place; a `configuration` (target/value) change plans as an in-place update that does not stick at the API. IPv6 values must be fully-expanded long form. Destroy is a real delete.

## Outputs

| Name | Description |
|------|-------------|
| `rule_id` | The created rule's ID |
| `zone_id` | The zone the rule was created on (empty for account-wide rules) |
| `account_id` | The account the rule was created on (empty for zone-scoped rules) |

## Provider Version

Uses `cloudflare/cloudflare ~> 5.23`.
