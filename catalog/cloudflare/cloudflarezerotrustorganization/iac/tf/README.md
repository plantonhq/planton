# CloudflareZeroTrustOrganization Terraform Module

Terraform IaC module for the Zero Trust organization (Access login experience + service-key rotation cadence).

## Architecture

```
provider.tf   — Cloudflare provider configuration (~> 5.23)
variables.tf  — Input variables mirroring CloudflareZeroTrustOrganizationSpec (generated)
main.tf       — cloudflare_zero_trust_organization + the folded
                cloudflare_zero_trust_access_key_configuration (count-gated)
outputs.tf    — auth_domain, account_id
```

## Behavior

Both resources are singleton UPSERTS with NO-OP destroys: Cloudflare has no create call for an organization (create and update are the same PUT mutating whatever the account or zone already carries), and destroy abandons the live configuration exactly as last applied. Unset spec fields are never sent, leaving live values untouched. The key configuration deploys only when the spec declares `key_rotation_interval_days` (account scope only — spec validation enforces it), ordered after the organization.

The organization imports account-scoped only (`{account_id}`); a zone-scoped organization cannot be imported (the provider's importer is account-only).

## Outputs

| Name | Description |
|------|-------------|
| `auth_domain` | The full team domain (including `.cloudflareaccess.com`) |
| `account_id` | The account the organization was applied to (empty for zone scope) |

## Provider Version

Uses `cloudflare/cloudflare ~> 5.23`.
