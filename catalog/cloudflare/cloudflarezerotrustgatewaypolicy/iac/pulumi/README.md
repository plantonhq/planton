# CloudflareZeroTrustGatewayPolicy Pulumi Module

Pulumi (Go) IaC module for one Gateway policy -- a filter over employee DNS/HTTP/network traffic plus the action taken on a match.

## Architecture

```
main.go                   — Entrypoint loading the IaC input
module/main.go            — Resources(): provider setup, resource, outputs
module/locals.go          — Locals initialization
module/gateway_policy.go  — cloudflare.ZeroTrustGatewayPolicy + rule_settings builder
module/outputs.go         — output keys
```

## Behavior

Mirrors the Terraform module's contract exactly: singular `filter` wrapped to a one-element list, `rule_settings` always sent (empty object when unset), `enabled` passed through explicitly.

## Outputs

| Name | Description |
|------|-------------|
| `policy_id` | UUID of the created policy |
| `precedence` | Evaluation order |

## Provider Version

Uses `pulumi-cloudflare` SDK v6 (Cloudflare Terraform provider v5 line).
