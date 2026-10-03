# Cloudflare Zero Trust MCP Server

Registers an MCP (Model Context Protocol) server with Cloudflare Access AI Controls: an upstream tool server whose prompts and tools Cloudflare proxies, audits, and gates behind Access. Registered servers are published to users through one or more MCP portals that reference them. Identity is user-supplied and immutable — `serverId`, `hostname`, and `authType` all force replacement — and the upstream credentials are write-only at Cloudflare, which makes IaC the rotation path.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **MCP Server Registration** — a `cloudflare_zero_trust_access_ai_controls_mcp_server` carrying the upstream URL, the auth posture (oauth, bearer, or unauthenticated) with its write-only credential, per-prompt/tool overrides, and the Gateway-filtering toggle. After registration, Cloudflare syncs the server's prompt/tool inventory in the background

## Before You Deploy

### Planton Setup

- **Cloudflare Provider Connection** — an active connection in the Connect module with a Cloudflare API token carrying Account → Zero Trust → Edit. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** — required when using Runner-based credential delivery. Not needed for inline API token authentication.

### Cloudflare Account

- **Zero Trust enabled** — the account must have completed Cloudflare Zero Trust onboarding (the team-name step).
- **The upstream MCP server's URL** — reachable from Cloudflare's egress; servers behind ACLs must admit Cloudflare or the background inventory sync reports an error until they do.
- **The upstream credential as a managed secret** (only for `bearer` and manual `oauth`) — the bearer token or OAuth client secret, stored as a managed secret the platform resolves just-in-time at deploy.

## Deploy

### Console

Open the deployment store, find **Cloudflare Zero Trust MCP Server**, and click **Deploy**. The creation wizard walks you through the owning account, the server identity (ID, upstream URL, auth type), the credential reference, and the prompt/tool overrides. Start from the **Bearer-authenticated third-party server** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareZeroTrustMcpServer
metadata:
  name: docs-search
  org: acme-corp
  env: prod
spec:
  accountId: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  serverId: docs-search
  name: Docs Search
  hostname: https://mcp.acme.com
  authType: bearer
  authCredentials:
    value: $secret/mcp-docs-search-token
```

```shell
planton apply -f mcp-server.yaml
```

This registers the upstream server with a bearer token resolved from a managed secret; Cloudflare then syncs its prompt/tool inventory in the background. An Infra Job tracks the provisioning in real time.

## Key Configuration

These are the most important decisions when configuring an MCP server registration. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Write-only credentials: IaC is the rotation path** — `authCredentials` and `clientSecret` are stored encrypted at Cloudflare and never returned by any read. Out-of-band rotation (dashboard, API) is invisible to IaC — no drift ever shows — and an import lands with empty credentials that must be re-asserted from configuration. Rotate by changing the managed-secret value and re-applying; the modules send the full body on update, so the new value lands.

**The immutable trio** — `serverId`, `hostname`, and `authType` all force replacement, and a replacement is a new registration: portals referencing the old ID must be re-pointed, and users lose in-flight OAuth grants. Upgrading a server from bearer to oauth is therefore a migration (new registration, portal re-point, old registration retired), not an edit.

**Auth type is the trust posture** — `unauthenticated` suits internal servers that trust their network path (Cloudflare authenticates the user at the portal; the upstream takes Cloudflare's word for it). `bearer` has Cloudflare hold a vendor token and present it on every call, so users never touch the credential. `oauth` has Cloudflare complete an OAuth flow against the server; `isSharedOauthCallbackEnabled` uses Cloudflare's shared callback URL for providers that allow only one registered redirect URI.

**Sync status is information, not health** — a sync status of "error" means the upstream did not answer, common for servers behind ACLs before network setup completes. The registration is real either way; fix reachability and the next sync recovers. Never gate deployment success on sync state.

**Overrides are allow-list thinking** — `updatedTools`/`updatedPrompts` rows with `enabled: false` remove sharp tools (deletes, writes) from what anyone can invoke through Cloudflare, regardless of portal. Disable at the server for "nobody should ever call this"; disable at the portal for "this audience shouldn't". Portals can only subtract from what the server allows.

**Destroy breaks publishing portals** — destroy is a real delete; portals referencing the server must drop their rows first.

## Outputs and Dependencies

### What This Kind Consumes

This kind has no foreign key dependencies. `authCredentials` and `clientSecret` are managed-secret references (resolved just-in-time at deploy), not references to other Infra Components.

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `server_id` | The server's identifier | MCP portal `servers[].serverId` rows publishing this registration |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Unauthenticated internal server** — an internal MCP server that trusts its network path, with its sharp tools disabled at the server level so nobody can invoke them through Cloudflare regardless of portal. Start from the **Unauthenticated internal server** preset.

**Bearer-authenticated third-party server** — Cloudflare holds the vendor's bearer token and presents it on every upstream call, with upstream traffic additionally filtered through Secure Web Gateway policies. Start from the **Bearer-authenticated third-party server** preset.

**Register once, publish many** — one registration per upstream server, published to multiple audiences through separate MCP portals — the credential rotates in one place while each portal curates its own view.

## Works With

- [**Cloudflare Zero Trust MCP Portal**](/infra-catalog/cloudflare-zero-trust-mcp-portal) — the user-facing endpoint that publishes this registration via `servers[].serverId`.
- [**Cloudflare Zero Trust Gateway Settings**](/infra-catalog/cloudflare-zero-trust-gateway-settings) — the Gateway posture that filters upstream traffic when `secureWebGateway` is on.
- [**Cloudflare Zero Trust Organization**](/infra-catalog/cloudflare-zero-trust-organization) — the login experience users authenticate through before reaching any published server.
