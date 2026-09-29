# CloudflareSnippet Terraform Module

Terraform IaC module for a small JavaScript snippet at the zone's edge, invoked by snippet rules. The snippet name is the identity; create is an upsert.

## Architecture

```
provider.tf   — Cloudflare provider configuration (~> 5.23)
variables.tf  — Input variables mirroring CloudflareSnippetSpec
main.tf       — cloudflare_snippet
outputs.tf    — snippet_name, zone_id
```

## Behavior

`snippet_name` is the identity -- a name collision silently overwrites the existing snippet. `main_module` is sent as `metadata.main_module` and must name one of `files`. Content must be byte-stable (the provider rebuilds stored source from a multipart read). Destroy is a real delete.

## Outputs

| Name | Description |
|------|-------------|
| `snippet_name` | The snippet's name -- what snippet rules reference |
| `zone_id` | The zone the snippet is deployed to |

## Provider Version

Uses `cloudflare/cloudflare ~> 5.23`.
