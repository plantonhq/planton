# CloudflareZeroTrustAccessServiceToken Pulumi Module

Pulumi (Go) IaC module for an Access service token -- a machine credential (client-ID / client-secret) for non-human clients.

## Architecture

```
main.go                   — Entrypoint loading the IaC input
module/main.go            — Resources(): provider setup, resource, outputs
module/locals.go          — Locals initialization
module/service_token.go   — cloudflare.ZeroTrustAccessServiceToken
module/outputs.go         — output keys
```

## Behavior

Mirrors the Terraform module's contract exactly. `client_secret_version` is converted from the spec's int32 to the SDK's float64. `client_secret` is exported with `pulumi.ToSecret`.

## Outputs

| Name | Description |
|------|-------------|
| `service_token_id` | UUID used for import and policy `service_token` rules |
| `client_id` | Presented in `CF-Access-Client-ID` |
| `client_secret` | Sensitive. Returned only at create and rotation |
| `expires_at` | RFC3339 expiry |

## Provider Version

Uses `pulumi-cloudflare` SDK v6 (Cloudflare Terraform provider v5 line).
