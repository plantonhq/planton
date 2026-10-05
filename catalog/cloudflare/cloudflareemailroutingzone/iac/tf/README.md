# CloudflareEmailRoutingZone Terraform Module

Terraform IaC module for enabling Cloudflare Email Routing on a zone, with the folded per-zone catch-all rule and optional managed-DNS locking.

## Architecture

```
provider.tf   — Cloudflare provider configuration
variables.tf  — Input variables mirroring CloudflareEmailRoutingZoneSpec
locals.tf     — Catch-all typed-action -> provider {type, value[]} mapping
main.tf       — email_routing_settings + conditional catch_all and dns resources
outputs.tf    — outputs (zone_id, enabled, status, name)
```

## Resource semantics worth knowing

- `cloudflare_email_routing_settings` is a create/destroy toggle: creating it ENABLES Email Routing (Cloudflare provisions the MX/SPF/DKIM records), destroying it DISABLES routing. There is no update.
- `cloudflare_email_routing_catch_all` is a zone singleton whose provider Delete is a genuine no-op — destroying it drops it from state and the zone keeps its last catch-all configuration. The zone-level disable is what retires the behavior.
- `cloudflare_email_routing_dns` (created when `lock_dns_records` is true) manages the routing DNS records explicitly; `dns_name` targets a subdomain, empty routes the zone apex.

## Usage

This module is invoked by the Planton CLI, which loads variable values from the CloudflareEmailRoutingZone YAML manifest. For standalone use:

```hcl
module "email_routing" {
  source = "./path/to/module"

  metadata = {
    name = "example-email-routing"
  }

  spec = {
    zone_id = "your-zone-id"
    catch_all = {
      enabled = true
      name    = "fallback-forward"
      actions = [
        {
          type       = "forward"
          forward_to = ["ops@example.com"]
        }
      ]
    }
    lock_dns_records = false
  }
}
```

## Outputs

| Name | Description |
|------|-------------|
| `zone_id` | Zone ID (pass-through; the import identity of all three folded resources) |
| `enabled` | Whether Email Routing is enabled |
| `status` | Routing status reported by Cloudflare |
| `name` | The routing settings' domain name |

## Provider Version

Uses `cloudflare/cloudflare ~> 5.23`. Catch-all actions and matchers are nested attributes (assigned with object syntax) in provider v5.
