# CloudflareEmailRoutingRule Terraform Module

Terraform IaC module for provisioning a single Cloudflare Email Routing rule: match incoming mail and drop it, forward it, and/or hand it to an Email Worker.

## Architecture

```
provider.tf   — Cloudflare provider configuration
variables.tf  — Input variables mirroring CloudflareEmailRoutingRuleSpec
locals.tf     — Typed actions -> provider {type, value[]} mapping
main.tf       — cloudflare_email_routing_rule resource
outputs.tf    — outputs (rule_id, zone_id)
```

## Usage

This module is invoked by the Planton CLI, which loads variable values from the CloudflareEmailRoutingRule YAML manifest. A rule carries a LIST of actions (matching the Cloudflare API), so one rule can forward AND invoke a Worker. For standalone use:

```hcl
module "email_rule" {
  source = "./path/to/module"

  metadata = {
    name = "support-to-ops"
  }

  spec = {
    zone_id  = "your-zone-id"
    name     = "support-to-ops"
    enabled  = true
    priority = 0
    matchers = [
      {
        type  = "literal"
        field = "to"
        value = "support@example.com"
      }
    ]
    actions = [
      {
        type       = "forward"
        forward_to = ["ops@example.com"]
      },
      {
        type   = "worker"
        worker = "email-processor"
      }
    ]
  }
}
```

The API accepts rules on any zone (measured live 2026-08-26 — enablement is not a create-time requirement), but rules only take effect once the zone's Email Routing is enabled (the CloudflareEmailRoutingZone kind). Forward destinations must be VERIFIED addresses (the CloudflareEmailRoutingAddress kind).

## Outputs

| Name | Description |
|------|-------------|
| `rule_id` | Cloudflare-assigned rule ID |
| `zone_id` | Zone ID (pass-through for the compound import identity) |

## Provider Version

Uses `cloudflare/cloudflare ~> 5.23`. Actions and matchers are nested attributes (assigned with object syntax) in provider v5.
