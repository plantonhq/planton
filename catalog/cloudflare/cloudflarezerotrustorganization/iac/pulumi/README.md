# CloudflareZeroTrustOrganization Pulumi Module

Pulumi (Go) IaC module for the Zero Trust organization (Access login experience + service-key rotation cadence).

## Architecture

```
main.go                   — Entrypoint loading the IaC input
module/main.go            — Resources(): provider setup, resources, outputs
module/locals.go          — Locals initialization
module/organization.go    — cloudflare.ZeroTrustOrganization + the folded
                            cloudflare.ZeroTrustAccessKeyConfiguration
module/outputs.go         — output keys
```

## Behavior

Mirrors the Terraform module's contract exactly. Both folded resources are singleton UPSERTS with NO-OP destroys: applying mutates whatever organization (and key-rotation cadence) the account already carries, and destroying abandons the live configuration exactly as last applied. Unset spec fields are never sent. The key configuration deploys only when the spec declares `key_rotation_interval_days` (account scope only), ordered after the organization.

## Outputs

| Name | Description |
|------|-------------|
| `auth_domain` | The full team domain (including `.cloudflareaccess.com`) |
| `account_id` | The account the organization was applied to (empty for zone scope) |

## Provider Version

Uses `pulumi-cloudflare` SDK v6 (Cloudflare Terraform provider v5 line).
