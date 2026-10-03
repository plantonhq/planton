# CloudflareZeroTrustDeviceCustomProfile Pulumi Module

Pulumi (Go) IaC module for targeted WARP device profiles.

## Architecture

```
main.go                    — iac-input loading + module entry
module/main.go             — provider setup + resource orchestration
module/locals.go           — metadata/credential references
module/custom_profile.go   — ZeroTrustDeviceCustomProfile
                             + conditional ZeroTrustDeviceCustomProfileLocalDomainFallback
module/outputs.go          — policy_id, gateway_unique_id
```

## Behavior

A real object: create, update, and delete all do what they say (deleting the profile returns matched devices to the default profile). Unset spec fields are never sent, keeping Cloudflare's defaults. The per-profile fallback-domain companion deploys only when the spec declares rows, is wired to the created profile's id, and REPLACES this profile's whole list (the profile's own fallbackDomains attribute is read-only -- the companion is the only write path); the rows retire with the profile. Import as `{account_id}/{policy_id}` (same pair for the fallback list).

## Outputs

| Name | Description |
|------|-------------|
| `policy_id` | The Cloudflare-assigned profile identifier |
| `gateway_unique_id` | The Gateway-side identifier of the profile |

## SDK Version

Uses `pulumi-cloudflare` v6.19.0.
