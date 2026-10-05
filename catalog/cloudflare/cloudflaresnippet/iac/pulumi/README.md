# CloudflareSnippet Pulumi Module

Pulumi (Go) IaC module for a small JavaScript snippet at the zone's edge, invoked by snippet rules. The snippet name is the identity; create is an upsert.

## Architecture

```
main.go                    — Entrypoint loading the IaC input
module/main.go             — Resources(): provider setup, resource, outputs
module/locals.go           — Locals initialization
module/snippet.go          — cloudflare.Snippet
module/outputs.go          — output keys
```

## Behavior

Mirrors the Terraform module's contract exactly: name-as-identity upsert, `main_module` sent as metadata, byte-stable files, `snippet_name` / `zone_id` outputs.

## Outputs

| Name | Description |
|------|-------------|
| `snippet_name` | The snippet's name -- what snippet rules reference |
| `zone_id` | The zone the snippet is deployed to |

## Provider Version

Uses `pulumi-cloudflare` SDK v6 (Cloudflare Terraform provider v5 line).
