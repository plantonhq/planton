# CloudflareAuthenticatedOriginPulls Pulumi Module

Pulumi (Go) IaC module for a zone's Authenticated Origin Pulls enablement: the zone-wide toggle plus per-hostname certificate associations.

## Architecture

```
main.go                                 — Entrypoint loading the IaC input
module/main.go                          — Resources(): provider setup, resources, outputs
module/locals.go                        — Locals initialization
module/authenticated_origin_pulls.go    — cloudflare.AuthenticatedOriginPullsSettings
                                          + cloudflare.AuthenticatedOriginPulls per row
module/outputs.go                       — output keys
```

## Behavior

Mirrors the Terraform module's contract exactly: the toggle managed only when `zone_enabled` is present, one association resource per hostname row (single-element config -- the provider hard-fails otherwise), omitted row `enabled` sent as true, and the `zone_id` output. Destroy deletes nothing: the toggle is abandoned and associations revert by a null write.

## Outputs

| Name | Description |
|------|-------------|
| `zone_id` | The zone whose Authenticated Origin Pulls surface is managed |

## Provider Version

Uses `pulumi-cloudflare` SDK v6 (Cloudflare Terraform provider v5 line).
