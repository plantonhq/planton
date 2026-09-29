# CloudflareZeroTrustMcpPortal Terraform Module

Terraform IaC module for MCP portals (curated, Access-protected MCP server collections).

## Architecture

```
provider.tf   — Cloudflare provider configuration (~> 5.23)
variables.tf  — Input variables mirroring CloudflareZeroTrustMcpPortalSpec (generated)
main.tf       — cloudflare_zero_trust_access_ai_controls_mcp_portal
outputs.tf    — portal_id, hostname
```

## Behavior

The portal's `portal_id` (the provider's `id`) is user-supplied and forces replacement; `hostname` and `name` update in place. The `servers` rows are a SET at the provider — reordering spec rows never plans a change. Per-portal prompt/tool overrides carry write-only `alias`/`description` fields (the API returns them under different names, so the provider never refreshes them). Real delete; honest 404 afterward. Import as `{account_id}/{portal_id}`.

## Outputs

| Name | Description |
|------|-------------|
| `portal_id` | The portal's identifier (the user-chosen slug) |
| `hostname` | The hostname the portal is served on |

## Provider Version

Uses `cloudflare/cloudflare ~> 5.23`.
