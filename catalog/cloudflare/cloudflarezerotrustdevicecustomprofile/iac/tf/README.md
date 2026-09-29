# CloudflareZeroTrustDeviceCustomProfile Terraform Module

Terraform IaC module for targeted WARP device profiles.

## Architecture

```
provider.tf   — Cloudflare provider configuration (~> 5.23)
variables.tf  — Input variables mirroring CloudflareZeroTrustDeviceCustomProfileSpec (generated)
main.tf       — cloudflare_zero_trust_device_custom_profile
                + conditional cloudflare_zero_trust_device_custom_profile_local_domain_fallback
outputs.tf    — policy_id, gateway_unique_id
```

## Behavior

A real object: create, update, and delete all do what they say (deleting the profile returns matched devices to the default profile). Unset spec fields are never sent, keeping Cloudflare's defaults. The per-profile fallback-domain companion deploys only when the spec declares rows, is wired to the created profile's id, and REPLACES this profile's whole list (the profile's own fallback_domains attribute is read-only -- the companion is the only write path); the rows retire with the profile. Import as `{account_id}/{policy_id}` (same pair for the fallback list).

## Outputs

| Name | Description |
|------|-------------|
| `policy_id` | The Cloudflare-assigned profile identifier |
| `gateway_unique_id` | The Gateway-side identifier of the profile |

## Provider Version

Uses `cloudflare/cloudflare ~> 5.23`.
