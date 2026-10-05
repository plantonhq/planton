# Cloudflare Zero Trust MCP Portal

## Overview

`CloudflareZeroTrustMcpPortal` publishes an MCP portal: the single Access-protected endpoint users (and their AI clients) connect to for a curated collection of MCP servers. The portal aggregates registered servers (`CloudflareZeroTrustMcpServer`), applies per-server and per-prompt/tool overrides, and fronts everything with Access on its hostname.

The portal's id is user-supplied and IMMUTABLE; hostname and name update in place. Server rows are an UNORDERED set -- the backend returns its own canonical order, so reordering rows never plans a change.

## Key Features

- **One endpoint, many servers** -- AI clients configure a single URL for the whole curated collection
- **Per-portal curation** -- disable servers or individual tools for this audience without touching the registration
- **On-behalf authentication** -- Cloudflare signs in to upstreams for the user (toggleable per server row)
- **Code mode control** -- allow or forbid client code execution against portal tools

## Use Cases

**Ideal for:**

- The engineering portal: docs, CI, ticketing MCP servers behind one Access-protected URL
- Audience-specific curation: the same servers, different tool sets per portal

**Not ideal for:**

- Registering the servers themselves -- that is `CloudflareZeroTrustMcpServer`
- Non-MCP applications -- ordinary Access applications cover those

## API Specification

### Required Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `account_id` | string | Yes | The Cloudflare account (32-hex). |
| `portal_id` | string | Yes | Your slug for the portal. IMMUTABLE. |
| `hostname` | string | Yes | The hostname the portal is served on. Mutable. |
| `name` | string | Yes | Display name. |

### The servers rows

| Field | Type | Description |
|-------|------|-------------|
| `server_id` | string/ref | The registered server (references CloudflareZeroTrustMcpServer). |
| `default_disabled` | bool | Listed but not usable until enabled. |
| `on_behalf` | bool | Cloudflare authenticates upstream for the user (default true). |
| `updated_prompts` / `updated_tools` | list | Per-portal overrides on top of the server's own. |

### Outputs

| Field | Description |
|-------|-------------|
| `portal_id` | The portal's identifier (the user-chosen slug) |
| `hostname` | The hostname users and AI clients connect to |

## Example Manifest

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareZeroTrustMcpPortal
metadata:
  name: eng-tools
spec:
  account_id: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  portal_id: eng-tools
  hostname: mcp.example.com
  name: Engineering Tools
  servers:
    - server_id:
        value: docs-search
```

## Destroy Semantics

Destroy is a real delete: the portal endpoint disappears; the registered servers it published are untouched.

## Related Resources

- **CloudflareZeroTrustMcpServer** -- the registrations this portal publishes
- **CloudflareZeroTrustAccessApplication** -- the broader Access application family

## Further Reading

For operational judgment -- the portal/server curation split, the immutable slug, write-only override fields -- see GUIDE.md.

## References

- [Cloudflare Access AI Controls](https://developers.cloudflare.com/cloudflare-one/applications/ai-controls/)

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
