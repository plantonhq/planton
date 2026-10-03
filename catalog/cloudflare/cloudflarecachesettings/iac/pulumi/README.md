# CloudflareCacheSettings Pulumi Module

Pulumi (Go) IaC module for managing a Cloudflare zone's caching and performance posture — tiered caching (smart and generic), Cache Reserve, regional tiered cache, cache variants, and Argo Smart Routing.

## Architecture

```
main.go              — Entrypoint loading the IaC input
module/main.go       — Resources(): provider setup, settings, outputs
module/locals.go     — Locals initialization
module/cache_settings.go — One conditionally-emitted resource per managed setting
module/outputs.go    — output keys (zone_id)
```

## Behavior

An unset spec field is NOT MANAGED: the module never sends it. Most of these settings have no delete at Cloudflare — destroy drops state and abandons the live values (smart tiered cache and cache variants are the real-delete exceptions). Argo Smart Routing is paid and KEEPS BILLING after destroy: apply `argo_smart_routing: false` first when retiring it.

The module mirrors the Terraform module's contract exactly: same resource set, same on/off mapping, same managed-extensions-only variants object (the Pulumi SDK pluralizes the per-extension field names where Terraform keeps the API's singular names), same `zone_id` output.

## Outputs

| Name | Description |
|------|-------------|
| `zone_id` | The zone the cache settings belong to (the singleton's identity) |

## Provider Version

Uses `pulumi-cloudflare` SDK v6 (Cloudflare Terraform provider v5 line).
