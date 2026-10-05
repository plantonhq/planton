# CloudflareZeroTrustMcpPortal Pulumi Module

Pulumi (Go) IaC module for MCP portals (curated, Access-protected MCP server collections).

## Architecture

```
main.go                   — Entrypoint loading the IaC input
module/main.go            — Resources(): provider setup, resource, outputs
module/locals.go          — Locals initialization
module/mcp_portal.go      — cloudflare.ZeroTrustAccessAiControlsMcpPortal
module/outputs.go         — output keys
```

## Behavior

Mirrors the Terraform module's contract exactly. The portal's `portal_id` (the provider's `id`) is user-supplied and forces replacement; `hostname` and `name` update in place. The `servers` rows are a SET at the provider — the backend ignores declaration order and returns its own canonical order, so reordering spec rows never plans a change.

## Outputs

| Name | Description |
|------|-------------|
| `portal_id` | The portal's identifier (the user-chosen slug) |
| `hostname` | The hostname the portal is served on |

## Provider Version

Uses `pulumi-cloudflare` SDK v6 (Cloudflare Terraform provider v5 line).
