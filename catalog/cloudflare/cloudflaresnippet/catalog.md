# Cloudflare Snippet

Deploys a Cloudflare Snippet: a small JavaScript module at the zone's edge, invoked on requests that match snippet rules managed by the separate Cloudflare Snippet Rules kind. Snippets are the lightweight sibling of Workers — same runtime, no bindings, sized for header rewrites, redirects, and request/response touch-ups, with code capped at 32 KB per snippet. The snippet name is the identity, and Cloudflare's create call is an upsert: deploying a name that already exists in the zone silently adopts and overwrites it, with no "already exists" error to catch the collision.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Snippet** -- one `cloudflare_snippet` on the zone, with the inline source files and a `metadata.main_module` entry point. The snippet does nothing until a snippet rule invokes it by name

## Before You Deploy

### Planton Setup

- **Cloudflare Provider Connection** -- an active connection in the Connect module with a Cloudflare API token that has Snippets Edit on the target zone. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline API token authentication.

### Cloudflare Account

- **A Cloudflare zone** -- provide the zone ID directly or wire `zoneId` to a Cloudflare DNS Zone resource by reference.
- **A unique `snippetName`** -- because create is an upsert, a name that already exists in the zone is silently overwritten. Coordinate names across teams the way you coordinate Worker script names.
- **A paid-plan zone (Pro+)** -- free zones refuse every snippet upload with "snippets are not allowed" (measured live); Pro and above include Snippets at no extra cost, with per-plan count limits the API enforces at create.

## Deploy

### Console

Open the deployment store, find **Cloudflare Snippet**, and click **Deploy**. The creation wizard walks you through the zone, the snippet name, the inline source files, and the entry module. Start from the **Redirect snippet** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareSnippet
metadata:
  name: redirect-legacy
  org: acme-corp
  env: prod
spec:
  zoneId:
    value: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  snippetName: redirect_legacy
  files:
    - name: main.js
      content: "export default { async fetch(request) { return Response.redirect(\"https://www.acme.com\", 302); } };"
  mainModule: main.js
```

```shell
planton apply -f snippet.yaml
```

This deploys a single-file snippet that returns a 302 — inert until a Cloudflare Snippet Rules entry invokes `redirect_legacy` by name. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, wire the zone to a DNS zone managed in the same InfraPipeline:

```yaml
spec:
  zoneId:
    valueFrom:
      kind: CloudflareDnsZone
      name: acme-com
      fieldPath: status.outputs.zone_id
  snippetName: redirect_legacy
```

The InfraPipeline resolves the dependency graph, deploys the zone first, then provisions the snippet against it.

## Key Configuration

These are the most important decisions when configuring a snippet. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Name is identity, and create is an upsert.** `snippetName` is how Cloudflare addresses the snippet — the create call adopts whatever already lives under that name and overwrites its files. Two teams, two manifests, one name: the last apply wins and the first team's code is gone. Letters, numbers, and underscores only (hyphens are rejected); pick names unique in the zone, stable, and never recycled.

**Renaming does not carry rules along.** Changing `snippetName` replaces the resource under a new identity, and any snippet rule referencing the old name keeps pointing at whatever the old name holds — the rules do not follow the rename.

**`mainModule` must name a file.** It is the entry module whose default export handles the request, and validation rejects a dangling entry point. Most snippets are a single `main.js`; multi-file snippets import siblings by name via ES module imports.

**Content must be byte-stable.** The provider refetches stored content on refresh, rebuilt from the API's multipart response — server-side normalization, or a trailing newline your editor added, reads back as configuration drift. Consistent line endings, no trailing-whitespace churn, no pretty-print-on-save the API will not echo.

**Snippet or Worker?** Snippets have no bindings, no cron, no custom domain, and a 32 KB code cap. The moment the logic needs state, secrets, or scheduling, reach for the Worker kind instead.

**Destroy is a real delete.** Rules that referenced `snippetName` keep pointing at that name and start invoking nothing. Update or delete those rules first, or accept a window where matching requests hit a missing snippet.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **CloudflareDnsZone** | `zoneId` | `status.outputs.zone_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `snippet_name` | The snippet's name — its identity in the zone | Referenced by Cloudflare Snippet Rules entries, wired instead of repeating the literal |
| `zone_id` | The zone the snippet is deployed to | Keeping the snippet-rules manifest on the same zone without a second literal |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Legacy-path redirect** -- A single-file 302 snippet too small to justify a Worker, invoked by one or more rules that match the legacy paths. Start from the **Redirect snippet** preset.

**Shared snippet, many rules** -- One snippet invoked by several rules with different match expressions — the rules carry the routing judgment while the snippet stays a single, testable module.

**Header touch-up at the edge** -- A snippet that adds, strips, or rewrites headers on matched traffic; the classic case for a snippet over a Worker because there is no state and no binding, just a transform.

## Works With

- [**Cloudflare Snippet Rules**](/infra-catalog/cloudflare-snippet-rules) -- the zone's routing table that invokes this snippet by name; nothing runs without it.
- [**Cloudflare Worker**](/infra-catalog/cloudflare-worker) -- the full Worker for logic that needs bindings, cron, or a custom domain.
- [**Cloudflare DNS Zone**](/infra-catalog/cloudflare-dns-zone) -- the zone scope; wire `zoneId` via ValueFromRef.
