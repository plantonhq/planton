# CloudflareRuleset Terraform Module

Terraform IaC module for provisioning Cloudflare Rulesets.

## Architecture

```
provider.tf   — Cloudflare provider configuration
variables.tf  — Input variables mirroring CloudflareRulesetSpec
locals.tf     — Derived values (zone_id extraction, kind, phase)
main.tf       — cloudflare_ruleset resource with dynamic rule blocks
outputs.tf    — Stack outputs (ruleset_id, version, zone_id, phase)
```

## Usage

This module is invoked by the OpenMCF CLI, which generates `variables.tf` values from the CloudflareRuleset YAML manifest. For standalone use:

```hcl
module "origin_rule" {
  source = "./path/to/module"

  metadata = {
    name = "planton-origin-routing"
  }

  spec = {
    zone_id = {
      value = "your-zone-id"
    }
    ruleset_kind = "zone"
    phase        = "http_request_origin"
    name         = "Route app traffic to K8s"
    rules = [
      {
        ref        = "route-app-to-k8s"
        expression = "not http.request.uri.path starts_with \"/docs\""
        action     = "route"
        action_parameters = {
          host_header = "planton.ai"
          origin = {
            host = "k8s-lb.example.com"
            port = 443
          }
        }
      }
    ]
  }
}
```

## Outputs

| Name | Description |
|------|-------------|
| `ruleset_id` | Cloudflare-assigned ruleset ID |
| `version` | Current ruleset version |
| `zone_id` | Zone ID (pass-through) |
| `phase` | Phase (pass-through) |

## Dynamic Blocks

The `main.tf` uses Terraform dynamic blocks extensively to handle the optional nature of action parameters. Each action type's parameters are wrapped in `dynamic` blocks that only render when the corresponding fields are non-null.

## Provider Version

Uses `cloudflare/cloudflare ~> 4.0`. The `cloudflare_ruleset` resource is available in all 4.x versions.
