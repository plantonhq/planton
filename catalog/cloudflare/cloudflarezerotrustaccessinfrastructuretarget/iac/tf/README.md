# CloudflareZeroTrustAccessInfrastructureTarget Terraform Module

Terraform IaC module for Zero Trust infrastructure targets.

## Architecture

```
provider.tf   — Cloudflare provider configuration (~> 5.23)
variables.tf  — Input variables mirroring CloudflareZeroTrustAccessInfrastructureTargetSpec (generated)
main.tf       — cloudflare_zero_trust_access_infrastructure_target
outputs.tf    — target_id
```

## Behavior

A plain CRUD resource: real create, in-place updates (hostname, addresses, virtual networks), real delete. Only the account forces replacement. An omitted `virtual_network_id` is not sent — Cloudflare assigns the account's default virtual network without drift. Import as `{account_id}/{target_id}`.

## Outputs

| Name | Description |
|------|-------------|
| `target_id` | The Cloudflare-assigned UUID of the target |

## Provider Version

Uses `cloudflare/cloudflare ~> 5.23`.
