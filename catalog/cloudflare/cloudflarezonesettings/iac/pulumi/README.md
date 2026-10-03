# CloudflareZoneSettings Pulumi Module

Pulumi (Go) IaC module for managing a Cloudflare zone's behavior settings — the HTTP, SSL/TLS, cache-adjacent, and network posture shown under the dashboard's Speed, Security, and Network tabs, plus managed header transforms, URL normalization, origin cloud-region hints, and the zone-wide waiting-room crawler bypass.

## Architecture

```
main.go             — Entrypoint loading the IaC input
module/main.go      — Resources(): provider setup, fan-out, companions, outputs
module/locals.go    — Locals initialization
module/zone_settings.go — collectSettings() + the cloudflare_zone_setting fan-out
module/companions.go — Managed transforms, URL normalization, origin regions, waiting-room bypass
module/outputs.go   — output keys (zone_id)
```

## Behavior

Each managed setting emits one `ZoneSetting` resource keyed by its setting id. An unset spec field is NOT MANAGED: the module never sends it, and the zone keeps its current value. Zone settings have no delete at Cloudflare — destroy drops state and abandons the live values, so revert settings explicitly before retiring the resource.

The module mirrors the Terraform module's contract exactly: same resource set, same on/off mapping for boolean settings, same `zone_id` output.

## Outputs

| Name | Description |
|------|-------------|
| `zone_id` | The zone the settings belong to (the singleton's identity) |

## Provider Version

Uses `pulumi-cloudflare` SDK v6 (Cloudflare Terraform provider v5 line).
