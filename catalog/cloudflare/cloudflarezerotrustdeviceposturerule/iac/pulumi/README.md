# CloudflareZeroTrustDevicePostureRule Pulumi Module

Pulumi (Go) IaC module for device posture rules.

## Architecture

```
main.go                  — iac-input loading + module entry
module/main.go           — provider setup + resource orchestration
module/locals.go         — metadata/credential references
module/posture_rule.go   — ZeroTrustDevicePostureRule
module/outputs.go        — rule_id
```

## Behavior

A plain CRUD resource: real create/update/delete; only the account forces replacement. The input tree carries every check family's parameters; unset fields are never sent, so each rule's payload holds exactly the fields its type reads. Import as `{account_id}/{rule_id}`.

## Outputs

| Name | Description |
|------|-------------|
| `rule_id` | The Cloudflare-assigned UUID of the rule (what policies reference) |

## SDK Version

Uses `pulumi-cloudflare` v6.19.0.
