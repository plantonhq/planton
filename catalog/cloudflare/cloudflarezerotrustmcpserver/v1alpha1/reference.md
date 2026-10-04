# CloudflareZeroTrustMcpServer

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `cloudflare.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

CloudflareZeroTrustMcpServerSpec registers an MCP (Model Context Protocol)
server with Access AI Controls: an upstream tool server whose prompts and
tools Cloudflare proxies, audits, and gates behind Access. Registered
servers are then published to users through one or more MCP portals
(CloudflareZeroTrustMcpPortal) that reference them.

Identity is user-supplied and immutable: server_id, hostname, and
auth_type all force replacement when changed. Cloudflare syncs the
server's prompt/tool inventory in the background after registration --
the status of that sync lives on the API object, not in this spec.

## Example

```yaml
# Complete example manifest for CloudflareZeroTrustMcpServer. Registers an
# upstream MCP tool server with Access AI Controls; portals then publish it
# to users. server_id, hostname, and auth_type are immutable (changing any
# replaces the registration).
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareZeroTrustMcpServer
metadata:
  name: docs-search
spec:
  account_id: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  server_id: docs-search
  name: Docs Search
  hostname: https://mcp.example.com
  auth_type: bearer
  auth_credentials:
    value: REPLACE_WITH_UPSTREAM_BEARER_TOKEN
  description: Company documentation search tools
  updated_tools:
    - name: delete_page
      enabled: false
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.accountId` | `string` | yes |  |  |
| `spec.serverId` | `string` | yes |  |  |
| `spec.name` | `string` | yes |  |  |
| `spec.hostname` | `string` | yes |  |  |
| `spec.authType` | `string` | yes |  |  |
| `spec.authCredentials` | `string \| valueFrom` (sensitive) |  |  |  |
| `spec.clientSecret` | `string \| valueFrom` (sensitive) |  |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.updatedPrompts` | `[]CloudflareZeroTrustMcpServerItemOverride` |  |  |  |
| `spec.updatedPrompts[].name` | `string` | yes |  |  |
| `spec.updatedPrompts[].alias` | `string` |  |  |  |
| `spec.updatedPrompts[].description` | `string` |  |  |  |
| `spec.updatedPrompts[].enabled` | `bool` |  | `true` |  |
| `spec.updatedTools` | `[]CloudflareZeroTrustMcpServerItemOverride` |  |  |  |
| `spec.updatedTools[].name` | `string` | yes |  |  |
| `spec.updatedTools[].alias` | `string` |  |  |  |
| `spec.updatedTools[].description` | `string` |  |  |  |
| `spec.updatedTools[].enabled` | `bool` |  | `true` |  |
| `spec.isSharedOauthCallbackEnabled` | `bool` |  |  |  |
| `spec.secureWebGateway` | `bool` |  |  |  |

## Field Details

### spec.accountId

`string` · required

The Cloudflare account the server is registered in.

- rule: account_id must be a 32-character hex string
- rule: {"required":true}

### spec.serverId

`string` · required

The server's identifier, chosen by you. IMMUTABLE: changing it replaces
the registration, and portals referencing the old id must be re-pointed.

- rule: {"required":true}

### spec.name

`string` · required

The server's display name, shown in portals and the dashboard.

- rule: {"required":true}

### spec.hostname

`string` · required

The URL of the upstream MCP server Cloudflare proxies to. IMMUTABLE:
changing it replaces the registration.

- rule: {"required":true}

### spec.authType

`string` · required

How Cloudflare authenticates to the upstream server: oauth (Cloudflare
completes an OAuth flow against the server), bearer (a static token sent
on every request), or unauthenticated. IMMUTABLE: changing it replaces
the registration.

- rule: auth_type must be one of oauth, bearer, unauthenticated
- rule: {"required":true}

### spec.authCredentials

`string | valueFrom` · sensitive

The upstream credential for the bearer auth type (the token value).
WRITE-ONLY: Cloudflare stores it encrypted and never returns it -- it
cannot be read back, imported, or drift-detected. Provide a
managed-secret reference; the platform resolves it just-in-time at
deploy.

- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.clientSecret

`string | valueFrom` · sensitive

The OAuth client secret for manually-configured OAuth (Cloudflare's
auth_config auth_mode "manual" -- a server-reported mode; with dynamic
client registration the secret is minted by the flow instead).
WRITE-ONLY: stored encrypted, never returned by any read. Provide a
managed-secret reference.

- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.description

`string`

A free-form description shown in portals and the dashboard.

### spec.updatedPrompts

`[]CloudflareZeroTrustMcpServerItemOverride`

Per-prompt overrides applied on top of the synced inventory: rename,
re-describe, or disable individual prompts the upstream server exposes.

### spec.updatedPrompts[].name

`string` · required

The prompt/tool name as the upstream server exposes it -- the key the
override matches on.

- rule: {"required":true}

### spec.updatedPrompts[].alias

`string`

The name shown to users instead of the upstream name.

### spec.updatedPrompts[].description

`string`

The description shown to users instead of the upstream description.

### spec.updatedPrompts[].enabled

`bool` · optional (explicit presence)

Whether the prompt/tool is available to users. Cloudflare's default is
enabled; set false to hide it.

- default: `true`

### spec.updatedTools

`[]CloudflareZeroTrustMcpServerItemOverride`

Per-tool overrides applied on top of the synced inventory: rename,
re-describe, or disable individual tools the upstream server exposes.

### spec.updatedTools[].name

`string` · required

The prompt/tool name as the upstream server exposes it -- the key the
override matches on.

- rule: {"required":true}

### spec.updatedTools[].alias

`string`

The name shown to users instead of the upstream name.

### spec.updatedTools[].description

`string`

The description shown to users instead of the upstream description.

### spec.updatedTools[].enabled

`bool` · optional (explicit presence)

Whether the prompt/tool is available to users. Cloudflare's default is
enabled; set false to hide it.

- default: `true`

### spec.isSharedOauthCallbackEnabled

`bool` · optional (explicit presence)

When true, OAuth flows use Cloudflare's shared callback URL instead of a
per-server one -- some upstream providers allow only one registered
redirect URI.

### spec.secureWebGateway

`bool` · optional (explicit presence)

When true, traffic to the upstream server is additionally filtered by
Secure Web Gateway policies.

## Outputs

Reference an output from another manifest as `valueFrom: {kind: CloudflareZeroTrustMcpServer, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.server_id` | `string` | The server's identifier -- what MCP portals reference in their servers[] rows. |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| CloudflareZeroTrustMcpPortal | `spec.servers[].serverId` | `status.outputs.server_id` |

## See Also

- [Overview](../README.md)
