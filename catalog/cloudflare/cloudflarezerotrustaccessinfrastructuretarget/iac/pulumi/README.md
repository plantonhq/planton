# CloudflareZeroTrustAccessInfrastructureTarget Pulumi Module

Pulumi (Go) IaC module for Zero Trust infrastructure targets.

## Architecture

```
main.go                   — Entrypoint loading the IaC input
module/main.go            — Resources(): provider setup, resource, outputs
module/locals.go          — Locals initialization
module/target.go          — cloudflare.ZeroTrustAccessInfrastructureTarget
module/outputs.go         — output keys
```

## Behavior

Mirrors the Terraform module's contract exactly: a plain CRUD resource (real create, in-place updates, real delete; only the account forces replacement). An omitted `virtual_network_id` is not sent — Cloudflare assigns the account's default virtual network, and because the attribute is computed on the provider side the assigned value never drifts.

## Outputs

| Name | Description |
|------|-------------|
| `target_id` | The Cloudflare-assigned UUID of the target |

## Provider Version

Uses `pulumi-cloudflare` SDK v6 (Cloudflare Terraform provider v5 line).
