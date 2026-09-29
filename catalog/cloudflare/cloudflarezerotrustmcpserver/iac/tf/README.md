# CloudflareZeroTrustMcpServer Terraform Module

Terraform IaC module for MCP server registrations behind Access AI Controls.

## Architecture

```
provider.tf   — Cloudflare provider configuration (~> 5.23)
variables.tf  — Input variables mirroring CloudflareZeroTrustMcpServerSpec (generated)
main.tf       — cloudflare_zero_trust_access_ai_controls_mcp_server
outputs.tf    — server_id
```

## Behavior

Identity is user-supplied and immutable: `server_id` (the provider's `id`), `hostname`, and `auth_type` all force replacement. The credential arguments (`auth_credentials`, `client_secret`) are WRITE-ONLY at Cloudflare — stored encrypted, never returned by any read — and are declared `sensitive()` so they never print in plans; out-of-band rotation is invisible to IaC. Real delete; honest 404 afterward. Import as `{account_id}/{server_id}` (credentials never round-trip).

## Outputs

| Name | Description |
|------|-------------|
| `server_id` | The server's identifier — what MCP portals reference |

## Provider Version

Uses `cloudflare/cloudflare ~> 5.23`.
