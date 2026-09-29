# CloudflareSnippetRules Terraform Module

Terraform IaC module for a zone's snippet routing table -- the ordered list of expressions deciding which requests invoke which snippet. One manifest owns the whole table.

## Architecture

```
provider.tf   — Cloudflare provider configuration (~> 5.23)
variables.tf  — Input variables mirroring CloudflareSnippetRulesSpec
main.tf       — cloudflare_snippet_rules
outputs.tf    — zone_id
```

## Behavior

Every apply is a full-replacement PUT of the zone's snippet-rule table. Destroy wipes ALL snippet rules in the zone, including ones created outside this manifest. `enabled` defaults to true here (Cloudflare's provider default is false) -- the module coalesces a null to true so a declared rule runs.

## Outputs

| Name | Description |
|------|-------------|
| `zone_id` | The zone whose snippet routing table is managed (the singleton's identity) |

## Provider Version

Uses `cloudflare/cloudflare ~> 5.23`.
