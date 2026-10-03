# Cloudflare Zero Trust MCP Server

## Overview

`CloudflareZeroTrustMcpServer` registers an MCP (Model Context Protocol) server with Access AI Controls: an upstream tool server whose prompts and tools Cloudflare proxies, audits, and gates behind Access. Registered servers are published to users through MCP portals (`CloudflareZeroTrustMcpPortal`) that reference them.

Identity is user-supplied and immutable: `server_id`, `hostname`, and `auth_type` all force replacement. Cloudflare syncs the server's prompt/tool inventory in the background after registration.

## Key Features

- **Upstream auth handled once** -- oauth, bearer, or unauthenticated; users never hold the upstream credential
- **Write-only credentials** -- `auth_credentials` and `client_secret` are stored encrypted and never returned by any read
- **Per-item overrides** -- rename, re-describe, or disable individual prompts/tools from the synced inventory
- **Secure Web Gateway option** -- route upstream traffic through Gateway policies

## Use Cases

**Ideal for:**

- Registering internal tool servers (docs search, ticketing, CI control) for safe publication to AI clients
- Fronting a third-party MCP server with your own audit and access layer

**Not ideal for:**

- The user-facing endpoint -- that is the portal (`CloudflareZeroTrustMcpPortal`)
- Non-MCP HTTP APIs -- those are ordinary Access applications

## API Specification

### Required Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `account_id` | string | Yes | The Cloudflare account (32-hex). |
| `server_id` | string | Yes | Your identifier for the registration. IMMUTABLE. |
| `name` | string | Yes | Display name. |
| `hostname` | string | Yes | The upstream MCP server URL. IMMUTABLE. |
| `auth_type` | string | Yes | oauth, bearer, or unauthenticated. IMMUTABLE. |

### Sensitive Fields

| Field | Type | Description |
|-------|------|-------------|
| `auth_credentials` | secret ref | The bearer token for auth_type bearer. WRITE-ONLY at Cloudflare. |
| `client_secret` | secret ref | The OAuth client secret for manually-configured OAuth. WRITE-ONLY. |

### Outputs

| Field | Description |
|-------|-------------|
| `server_id` | The server's identifier -- what portals reference |

## Example Manifest

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareZeroTrustMcpServer
metadata:
  name: docs-search
spec:
  account_id: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  server_id: docs-search
  name: Docs Search
  hostname: https://mcp.example.com
  auth_type: unauthenticated
  updated_tools:
    - name: delete_page
      enabled: false
```

## Destroy Semantics

Destroy is a real delete; portals referencing the server must drop their rows first (or in the same change).

## Related Resources

- **CloudflareZeroTrustMcpPortal** -- publishes this server to users
- **CloudflareZeroTrustGatewaySettings** -- the Gateway posture behind secure_web_gateway

## Further Reading

For operational judgment -- credential rotation with write-only fields, the immutable identity trio, sync-status semantics -- see GUIDE.md.

## References

- [Cloudflare Access AI Controls](https://developers.cloudflare.com/cloudflare-one/applications/ai-controls/)

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
