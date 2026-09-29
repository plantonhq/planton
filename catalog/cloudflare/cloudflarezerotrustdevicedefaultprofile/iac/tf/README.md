# CloudflareZeroTrustDeviceDefaultProfile Terraform Module

Terraform IaC module for the account's default WARP device profile.

## Architecture

```
provider.tf   — Cloudflare provider configuration (~> 5.23)
variables.tf  — Input variables mirroring CloudflareZeroTrustDeviceDefaultProfileSpec (generated)
main.tf       — cloudflare_zero_trust_device_default_profile
                + conditional cloudflare_zero_trust_device_default_profile_local_domain_fallback
                + conditional cloudflare_zero_trust_device_default_profile_certificates
outputs.tf    — account_id, gateway_unique_id, policy_id
```

## Behavior

The profile is an account singleton: create and update are the same PATCH, and DESTROY IS A NO-OP on all three surfaces -- the last-applied values stand. Unset spec fields are never sent, keeping Cloudflare's defaults, with two live-measured exceptions: `dns_search_suffixes` is ALWAYS sent (an empty declaration clears the account's list -- the provider's attribute re-plans forever on an omitted send), and `tunnel_protocol` is blanked by the provider's own schema default when unset (declare it explicitly on accounts running a non-default protocol). The fallback-domain companion deploys only when the spec declares rows and REPLACES the account's whole list (the profile's own fallback_domains attribute is read-only -- this companion is the only write path). The zone-scoped certificate toggle deploys only when the spec declares the fold; it has no delete and no import, and its API endpoint refuses account-owned tokens (401 code 1039 -- a user-actor credential is required). Import the profile and fallback list as the bare `{account_id}`.

## Outputs

| Name | Description |
|------|-------------|
| `account_id` | The account the profile was applied to (the singleton's identity) |
| `gateway_unique_id` | The Gateway-side identifier of the profile |
| `policy_id` | The profile's policy identifier |

## Provider Version

Uses `cloudflare/cloudflare ~> 5.23`.
