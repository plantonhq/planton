# CloudflareZoneTlsSettings Terraform Module

Terraform IaC module for managing a Cloudflare zone's edge TLS posture — Universal SSL issuance, Total TLS, automatic origin TLS key exchange, origin TLS compliance modes, per-hostname TLS overrides, and certificate-authority hostname associations.

## Architecture

```
provider.tf   — Cloudflare provider configuration
variables.tf  — Input variables mirroring CloudflareZoneTlsSettingsSpec
locals.tf     — Per-hostname override maps + CA association keying
main.tf       — Count-gated singletons + per-(setting, hostname) fan-out
outputs.tf    — outputs (zone_id)
```

## Usage

This module is invoked by the Planton CLI, which loads variable values from the CloudflareZoneTlsSettings YAML manifest. For standalone use:

```hcl
module "zone_tls_settings" {
  source = "./path/to/module"

  metadata = {
    name = "prod-zone-tls"
  }

  spec = {
    zone_id = "your-zone-id"
    total_tls = {
      enabled               = true
      certificate_authority = "google"
    }
  }
}
```

An unset field is NOT MANAGED: the module never sends it. Universal SSL, Total TLS, auto origin key exchange, and CA hostname associations have no delete at Cloudflare — destroy drops state and abandons the live values. Per-hostname overrides and origin TLS compliance modes have real deletes. Disabling Universal SSL can make proxied hostnames unreachable over HTTPS unless other certificates cover them — only disable it deliberately.

Each hostname row's set attributes become separate `cloudflare_hostname_tls_setting` resources keyed by (setting id, hostname): editing one row never churns another row's resources.

## Outputs

| Name | Description |
|------|-------------|
| `zone_id` | The zone the TLS settings belong to (the singleton's identity) |

## Provider Version

Uses `cloudflare/cloudflare ~> 5.23`.
