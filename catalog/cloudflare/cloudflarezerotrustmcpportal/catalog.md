# Cloudflare Zero Trust MCP Portal

Publishes a Cloudflare MCP portal: the single Access-protected endpoint users and their AI clients connect to for a curated collection of MCP servers. The portal aggregates registered servers (CloudflareZeroTrustMcpServer), applies per-server and per-prompt/tool overrides for this audience, and fronts everything with Access on its hostname. The portal's `portalId` slug is immutable — it keys the portal and is part of what clients configure — while hostname and name update in place.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **MCP Portal** — a `cloudflare_zero_trust_access_ai_controls_mcp_portal` carrying the slug, serving hostname, display name, code-mode and Gateway-filtering toggles, and the published server rows with their per-portal overrides

## Before You Deploy

### Planton Setup

- **Cloudflare Provider Connection** — an active connection in the Connect module with a Cloudflare API token carrying Account → Zero Trust → Edit. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** — required when using Runner-based credential delivery. Not needed for inline API token authentication.

### Cloudflare Account

- **Zero Trust enabled** — the account must have completed Cloudflare Zero Trust onboarding (the team-name step).
- **Registered MCP servers** (unless starting empty) — the CloudflareZeroTrustMcpServer registrations this portal publishes must exist first.

## Deploy

### Console

Open the deployment store, find **Cloudflare Zero Trust MCP Portal**, and click **Deploy**. The creation wizard walks you through the owning account, the portal slug and hostname, the code-mode and Gateway toggles, and the server rows with their per-portal prompt/tool overrides. Start from the **Single-server portal** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareZeroTrustMcpPortal
metadata:
  name: eng-tools
  org: acme-corp
  env: prod
spec:
  accountId: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  portalId: eng-tools
  hostname: mcp.acme.com
  name: Engineering Tools
```

```shell
planton apply -f portal.yaml
```

This creates an empty Access-protected portal at the hostname; add server rows to publish registrations through it. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, wire the published servers to CloudflareZeroTrustMcpServer resources so the dependency graph carries the edge:

```yaml
spec:
  servers:
    - serverId:
        valueFrom:
          kind: CloudflareZeroTrustMcpServer
          name: ticketing-mcp
          fieldPath: status.outputs.server_id
      onBehalf: true
```

The InfraPipeline registers the server first, then publishes the portal against its real ID.

## Key Configuration

These are the most important decisions when configuring an MCP portal. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The slug is the URL is the identity** — `portalId` is immutable, and like the hostname it is part of what clients configure. Renaming the slug replaces the portal and breaks every AI client configured against it. Pick audience-stable slugs ("eng-tools", not "q3-pilot").

**Server-level vs portal-level curation** — disable a tool at the server registration when nobody should ever call it through Cloudflare; disable it at the portal when this audience shouldn't. A tool disabled at the server cannot be re-enabled by any portal — the server is the floor, portals only subtract.

**`onBehalf` is a trust decision** — with the default (enabled), Cloudflare authenticates to the upstream for the user: one upstream credential, held by Cloudflare, shared by every portal user. Set it false when the upstream must know which user is calling — per-user audit trails or per-user entitlements upstream.

**Rows are a set; some overrides are write-only** — server rows carry no order (the backend returns its own canonical order, so reordering never plans a change). Within a row's prompt/tool overrides, `alias` and `description` are write-only per portal: they cannot drift, and they land empty on import — re-assert them from configuration.

**Code mode is on by default** — `allowCodeMode` lets connecting clients execute code against portal tools. Cloudflare defaults it to enabled; turn it off for non-engineering audiences.

**Destroy removes only the endpoint** — deleting the portal is a real delete of the published endpoint, but the server registrations it published are untouched and stay available to other portals.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| CloudflareZeroTrustMcpServer | `spec.servers[].serverId` | `status.outputs.server_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `portal_id` | The portal's identifier (the user-chosen slug) | API automation against the portal |
| `hostname` | The hostname the portal is served on | AI client configuration; DNS records pointing at the portal |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Single-server portal** — the starter shape: one registered MCP server published on one Access-protected hostname, with the server referenced by ValueFromRef so the dependency graph carries the edge. Start from the **Single-server portal** preset.

**Curated audience portal** — the same registered servers, curated for one team: destructive tools disabled for this portal only (the registration keeps them), friendlier aliases on the rest, and code mode off for a non-engineering audience. Start from the **Curated audience portal** preset.

**Per-team portals over one server fleet** — register each upstream MCP server once, then publish different subsets to different audiences (engineering, support, leadership) through separate portals — one place to rotate upstream credentials, many curated views.

## Works With

- [**Cloudflare Zero Trust MCP Server**](/infra-catalog/cloudflare-zero-trust-mcp-server) — the registrations this portal publishes; deploy them first.
- [**Cloudflare Zero Trust Organization**](/infra-catalog/cloudflare-zero-trust-organization) — the login experience in front of the portal's Access protection.
- [**Cloudflare Zero Trust Access Application**](/infra-catalog/cloudflare-zero-trust-access-application) — the broader Access family the portal's hostname protection belongs to.
