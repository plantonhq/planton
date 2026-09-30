# CloudflareCacheSettings Terraform Module

Terraform IaC module for managing a Cloudflare zone's caching and performance posture — tiered caching (smart and generic), Cache Reserve, regional tiered cache, cache variants, and Argo Smart Routing.

## Architecture

```
provider.tf   — Cloudflare provider configuration
variables.tf  — Input variables mirroring CloudflareCacheSettingsSpec
locals.tf     — The cache-variants value (managed extensions only)
main.tf       — One count-gated resource per managed setting
outputs.tf    — Stack outputs (zone_id)
```

## Usage

This module is invoked by the Planton CLI, which loads variable values from the CloudflareCacheSettings YAML manifest. For standalone use:

```hcl
module "cache_settings" {
  source = "./path/to/module"

  metadata = {
    name = "prod-cache-settings"
  }

  spec = {
    zone_id            = "your-zone-id"
    smart_tiered_cache = true
  }
}
```

An unset field is NOT MANAGED: the module never sends it. Most of these settings have no delete at Cloudflare — destroy drops state and abandons the live values (smart tiered cache and cache variants are the real-delete exceptions). Argo Smart Routing is paid and KEEPS BILLING after destroy: apply `argo_smart_routing = false` first when retiring it.

## Outputs

| Name | Description |
|------|-------------|
| `zone_id` | The zone the cache settings belong to (the singleton's identity) |

## Provider Version

Uses `cloudflare/cloudflare ~> 5.23`.
