# CloudflareZoneSettings Terraform Module

Terraform IaC module for managing a Cloudflare zone's behavior settings — the HTTP, SSL/TLS, cache-adjacent, and network posture shown under the dashboard's Speed, Security, and Network tabs, plus managed header transforms, URL normalization, origin cloud-region hints, and the zone-wide waiting-room crawler bypass.

## Architecture

```
provider.tf   — Cloudflare provider configuration
variables.tf  — Input variables mirroring CloudflareZoneSettingsSpec
locals.tf     — The managed-settings map (unset spec fields never enter it)
main.tf       — cloudflare_zone_setting fan-out + companion resources
outputs.tf    — outputs (zone_id)
```

## Usage

This module is invoked by the Planton CLI, which loads variable values from the CloudflareZoneSettings YAML manifest. For standalone use:

```hcl
module "zone_settings" {
  source = "./path/to/module"

  metadata = {
    name = "prod-zone-settings"
  }

  spec = {
    zone_id          = "your-zone-id"
    always_use_https = true
    min_tls_version  = "1.2"
    ssl              = "strict"
  }
}
```

Each managed setting emits one `cloudflare_zone_setting` resource keyed by its setting id. An unset field is NOT MANAGED: the module never sends it, and the zone keeps its current value. Zone settings have no delete at Cloudflare — destroy drops state and abandons the live values (the provider warns about this on every plan), so revert settings explicitly before retiring the resource.

## Outputs

| Name | Description |
|------|-------------|
| `zone_id` | The zone the settings belong to (the singleton's identity) |

## Provider Version

Uses `cloudflare/cloudflare ~> 5.23`.
