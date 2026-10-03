# CloudflareZeroTrustMcpServer Pulumi Module

Pulumi (Go) IaC module for MCP server registrations behind Access AI Controls.

## Architecture

```
main.go                   — Entrypoint loading the IaC input
module/main.go            — Resources(): provider setup, resource, outputs
module/locals.go          — Locals initialization
module/mcp_server.go      — cloudflare.ZeroTrustAccessAiControlsMcpServer
module/outputs.go         — output keys
```

## Behavior

Mirrors the Terraform module's contract exactly. Identity is user-supplied and immutable: `server_id`, `hostname`, and `auth_type` all force replacement. The two credential fields (`auth_credentials`, `client_secret`) are WRITE-ONLY at Cloudflare (stored encrypted, never returned by any read) and are wrapped with `pulumi.ToSecret` so they never appear in plaintext state; out-of-band rotation is invisible to IaC — rotate by changing the value here.

## Outputs

| Name | Description |
|------|-------------|
| `server_id` | The server's identifier — what MCP portals reference |

## Provider Version

Uses `pulumi-cloudflare` SDK v6 (Cloudflare Terraform provider v5 line).
