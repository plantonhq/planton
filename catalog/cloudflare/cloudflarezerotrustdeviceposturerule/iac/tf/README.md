# CloudflareZeroTrustDevicePostureRule Terraform Module

Terraform IaC module for device posture rules.

## Architecture

```
provider.tf   — Cloudflare provider configuration (~> 5.23)
variables.tf  — Input variables mirroring CloudflareZeroTrustDevicePostureRuleSpec (generated)
main.tf       — cloudflare_zero_trust_device_posture_rule
outputs.tf    — rule_id
```

## Behavior

A plain CRUD resource: real create/update/delete; only the account forces replacement. The input tree carries every check family's parameters; unset fields are never sent, so each rule's payload holds exactly the fields its type reads. Import as `{account_id}/{rule_id}`.

## Outputs

| Name | Description |
|------|-------------|
| `rule_id` | The Cloudflare-assigned UUID of the rule (what policies reference) |

## Provider Version

Uses `cloudflare/cloudflare ~> 5.23`.
