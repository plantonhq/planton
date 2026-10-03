# CloudflareCustomHostnameFallbackOrigin Terraform Module

Terraform IaC module for setting a zone's fallback origin — the default backend every custom hostname on that zone routes to when no more-specific origin is configured.

## Architecture

```
provider.tf   — Cloudflare provider configuration
variables.tf  — Input variables mirroring CloudflareCustomHostnameFallbackOriginSpec
locals.tf     — Zone id and origin flattening (StringValueOrRef → string)
main.tf       — cloudflare_custom_hostname_fallback_origin resource
outputs.tf    — outputs (created_at, updated_at, errors, zone_id)
```

## Usage

This module is invoked by the Planton CLI, which loads variable values from the CloudflareCustomHostnameFallbackOrigin YAML manifest. For standalone use:

```hcl
module "fallback_origin" {
  source = "./path/to/module"

  metadata = {
    name = "saas-fallback"
  }

  spec = {
    zone_id = "your-zone-id"
    origin  = "origin.example.com"
  }
}
```

This is a zone singleton: one fallback origin per zone, and its API identity IS the zone. The write path is PUT (create equals update). `zone_id` is exported because there is no resource id of its own — verification, import, and chart blocks all key on the zone.

## Outputs

| Name | Description |
|------|-------------|
| `created_at` | RFC3339 creation timestamp |
| `updated_at` | RFC3339 last-updated timestamp |
| `zone_id` | The zone this singleton belongs to |

There is no `status` output: deployment is asynchronous (`pending_deployment` → `active`), and a point-in-time phase is never a stable output — it flips on the first refresh after the transition and re-plans forever.

## Provider Version

Uses `cloudflare/cloudflare ~> 5.23`.
