# CloudflareZeroTrustMcpPortal

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `cloudflare.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

CloudflareZeroTrustMcpPortalSpec publishes an MCP portal: the single
Access-protected endpoint users (and their AI clients) connect to for a
curated collection of MCP servers. The portal aggregates registered
servers (CloudflareZeroTrustMcpServer), applies per-server and
per-prompt/tool overrides, and fronts everything with Access policies on
its hostname.

The portal's id is user-supplied and IMMUTABLE (it keys the portal);
hostname and name update in place. Cloudflare returns the servers
collection in its own canonical order, so rows are declared as an
unordered set -- reordering them is never a change.

## Example

```yaml
# Complete example manifest for CloudflareZeroTrustMcpPortal. Publishes a
# curated, Access-protected collection of registered MCP servers on one
# hostname. portal_id is immutable; server rows are an unordered set.
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareZeroTrustMcpPortal
metadata:
  name: eng-tools
spec:
  account_id: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  portal_id: eng-tools
  hostname: mcp.example.com
  name: Engineering Tools
  description: MCP servers for the engineering org
  servers:
    - server_id:
        value: docs-search
      on_behalf: true
      updated_tools:
        - name: delete_page
          enabled: false
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.accountId` | `string` | yes |  |  |
| `spec.portalId` | `string` | yes |  |  |
| `spec.hostname` | `string` | yes |  |  |
| `spec.name` | `string` | yes |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.allowCodeMode` | `bool` |  | `true` |  |
| `spec.secureWebGateway` | `bool` |  |  |  |
| `spec.servers` | `[]CloudflareZeroTrustMcpPortalServer` |  |  |  |
| `spec.servers[].serverId` | `string \| valueFrom` | yes |  | CloudflareZeroTrustMcpServer (`status.outputs.server_id`) |
| `spec.servers[].defaultDisabled` | `bool` |  |  |  |
| `spec.servers[].onBehalf` | `bool` |  | `true` |  |
| `spec.servers[].updatedPrompts` | `[]CloudflareZeroTrustMcpPortalItemOverride` |  |  |  |
| `spec.servers[].updatedPrompts[].name` | `string` | yes |  |  |
| `spec.servers[].updatedPrompts[].alias` | `string` |  |  |  |
| `spec.servers[].updatedPrompts[].description` | `string` |  |  |  |
| `spec.servers[].updatedPrompts[].enabled` | `bool` |  | `true` |  |
| `spec.servers[].updatedTools` | `[]CloudflareZeroTrustMcpPortalItemOverride` |  |  |  |
| `spec.servers[].updatedTools[].name` | `string` | yes |  |  |
| `spec.servers[].updatedTools[].alias` | `string` |  |  |  |
| `spec.servers[].updatedTools[].description` | `string` |  |  |  |
| `spec.servers[].updatedTools[].enabled` | `bool` |  | `true` |  |

## Field Details

### spec.accountId

`string` · required

The Cloudflare account the portal belongs to.

- rule: account_id must be a 32-character hex string
- rule: {"required":true}

### spec.portalId

`string` · required

The portal's identifier, chosen by you (a URL-friendly slug, e.g.
"eng-tools"). IMMUTABLE: changing it replaces the portal.

- rule: {"required":true}

### spec.hostname

`string` · required

The hostname the portal is served on. Mutable, unlike the MCP server's
hostname -- moving a portal keeps its id and server rows.

The hostname MUST belong to a domain the account serves (the portal is
hosted by Cloudflare on your zone): a hostname on an unowned domain is
rejected at create with API error 7012 "invalid_domain" (live-measured).

- rule: {"required":true}

### spec.name

`string` · required

The portal's display name, shown to users.

- rule: {"required":true}

### spec.description

`string`

A free-form description of the portal.

### spec.allowCodeMode

`bool` · optional (explicit presence)

Whether users may connect coding agents in code mode (the portal
feature that lets clients execute code against portal tools).
Cloudflare's default is enabled; set false to disable.

- default: `true`

### spec.secureWebGateway

`bool` · optional (explicit presence)

When true, portal traffic is additionally filtered by Secure Web
Gateway policies.

### spec.servers

`[]CloudflareZeroTrustMcpPortalServer`

The MCP servers published through this portal. An UNORDERED set: the
backend ignores declaration order and returns its own canonical order.

### spec.servers[].serverId

`string | valueFrom` · required

The registered server to publish: a literal server id, or a reference
to a CloudflareZeroTrustMcpServer resource.

- references: CloudflareZeroTrustMcpServer (`status.outputs.server_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: CloudflareZeroTrustMcpServer, name: <that resource's name>, fieldPath: status.outputs.server_id}} -- a bare string does not parse

### spec.servers[].defaultDisabled

`bool` · optional (explicit presence)

When true, the server starts disabled for portal users (it is listed
but not usable until enabled).

### spec.servers[].onBehalf

`bool` · optional (explicit presence)

Whether Cloudflare authenticates to the upstream server on behalf of
the connecting user. Cloudflare's default is enabled; set false to make
users bring their own upstream session.

- default: `true`

### spec.servers[].updatedPrompts

`[]CloudflareZeroTrustMcpPortalItemOverride`

Per-prompt overrides applied for THIS portal, on top of the server's
own overrides. Cloudflare refreshes only name and enabled from the API
for these rows -- alias and description are write-only per portal.

Cloudflare VALIDATES every override name against the server's
actually-synced inventory: an override naming a prompt the server has
not synced is rejected at write with API error 7001 (live-measured on
tools; e.g. "Tool 'x' does not exist on server"). Overrides can only
be declared after the server's background sync has succeeded.

### spec.servers[].updatedPrompts[].name

`string` · required

The prompt/tool name as the upstream server exposes it -- the key the
override matches on.

- rule: {"required":true}

### spec.servers[].updatedPrompts[].alias

`string`

The name shown to portal users instead of the upstream name.

### spec.servers[].updatedPrompts[].description

`string`

The description shown to portal users instead of the upstream
description.

### spec.servers[].updatedPrompts[].enabled

`bool` · optional (explicit presence)

Whether the prompt/tool is available in this portal. Cloudflare's
default is enabled; set false to hide it.

- default: `true`

### spec.servers[].updatedTools

`[]CloudflareZeroTrustMcpPortalItemOverride`

Per-tool overrides applied for THIS portal, on top of the server's own
overrides. Same write-only behavior AND the same synced-inventory
validation as updated_prompts.

### spec.servers[].updatedTools[].name

`string` · required

The prompt/tool name as the upstream server exposes it -- the key the
override matches on.

- rule: {"required":true}

### spec.servers[].updatedTools[].alias

`string`

The name shown to portal users instead of the upstream name.

### spec.servers[].updatedTools[].description

`string`

The description shown to portal users instead of the upstream
description.

### spec.servers[].updatedTools[].enabled

`bool` · optional (explicit presence)

Whether the prompt/tool is available in this portal. Cloudflare's
default is enabled; set false to hide it.

- default: `true`

## Outputs

Reference an output from another manifest as `valueFrom: {kind: CloudflareZeroTrustMcpPortal, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.portal_id` | `string` | The portal's identifier (the user-chosen slug). |
| `status.outputs.hostname` | `string` | The hostname the portal is served on -- what users and AI clients connect to. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.servers[].serverId` | CloudflareZeroTrustMcpServer | `status.outputs.server_id` |

## See Also

- [Overview](../README.md)
