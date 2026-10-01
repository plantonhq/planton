# Cloudflare Web Analytics Site

Deploys a Cloudflare Web Analytics (RUM) site: privacy-first real-user monitoring for one website, measured by a JavaScript beacon, with include/exclude measurement rules folded into the same manifest. The site is identified by either a hostname (any site, on Cloudflare or not — you embed the snippet yourself) or a Cloudflare zone (where `autoInstall` can inject the beacon at the edge). Web Analytics is free on every plan.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **Web Analytics Site** — one RUM site on the account, identified by `host` or `zoneTag`
- **Measurement Rules** — created only when `rules` is non-empty; one include/exclude rule object per declared row, in order, under the site's ruleset

## Before You Deploy

### Planton Setup

- **Cloudflare Provider Connection** — an active connection in the Connect module with an API token holding Account → Account Settings → Write. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.
- **Planton Runner** — required when using Runner-based credential delivery. Not needed for inline credential authentication.

### Cloudflare Account

- **The account ID** — `accountId` is the 32-character hex account identifier the site belongs to.
- **A proxied zone** (only for `autoInstall`) — edge injection needs an orange-clouded zone; on a DNS-only zone Cloudflare has no response to inject into, so the setting is accepted and does nothing.

## Deploy

### Console

Open the deployment store, find **Cloudflare Web Analytics Site**, and click **Deploy**. The creation wizard walks you through environment and connection configuration, the site identity (hostname or zone), the beacon options, and the measurement rules. Start from the **Zone with automatic installation** preset in the [Presets](#presets) tab to pre-populate the zero-code shape.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareWebAnalyticsSite
metadata:
  name: www-acme-rum
  org: acme-corp
  env: prod
spec:
  accountId: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  host: www.acme.com
```

```shell
planton apply -f web-analytics-site.yaml
```

This creates a hostname-identified site measuring `www.acme.com` with the full beacon and no rules — embed the `snippet` output in your pages to start collecting. A Stack Job tracks the provisioning in real time.

### InfraChart

When measuring a zone deployed in the same InfraPipeline, wire `zoneTag` with ValueFromRef:

```yaml
spec:
  accountId: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  zoneTag:
    valueFrom:
      kind: CloudflareDnsZone
      name: acme-com
      fieldPath: status.outputs.zone_id
  autoInstall: true
```

The InfraPipeline resolves the dependency graph, deploys the zone first, then provisions the analytics site against the resolved zone ID.

## Key Configuration

These are the most important decisions when configuring a Web Analytics site. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Host or zone, never both** — `host` measures any site by hostname and you embed the snippet yourself; `zoneTag` measures a Cloudflare zone and unlocks `autoInstall`, which puts the beacon on every proxied page with no code change. The spec enforces exactly one. Choose `zoneTag` when the site is proxied through Cloudflare — the zero-code path is the whole point. Measurement rules are zone-identified-site territory: a host-identified site has no ruleset to attach rules to, and validation enforces it.

**Rules are write-only at the provider** — the provider never reads rules back after writing them, so rule edits made in the Cloudflare dashboard are invisible to IaC (no drift appears in any plan) until the next apply quietly overwrites them with the declared rows. Pick one owner for the rule set: if it is this manifest, dashboard edits are temporary by definition.

**Rule order is positional** — the module manages one rule object per declared row, keyed by position. Inserting a row at the top of a long list rewrites the objects below it; the end state is still exactly what the list says, but expect plan churn.

**Several site fields never refresh** — `enabled`, `host`, `lite`, and `zoneTag` are sent on write and never populated from a read, so drift on them is undetectable by design. State carries whatever you configured.

**Lite beacon** — `lite: true` serves a smaller script with a reduced metric set. Choose it when page weight matters more than metric completeness; it is a per-site, not per-page, decision.

**Destroy deletes the history** — destroying the site is not just "stop measuring": the site's collected analytics stop being reachable, and every folded rule is deleted with it. Export what you need before retiring a site.

## Outputs and Dependencies

### What This Component Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **CloudflareDnsZone** (zone-measured sites) | `zoneTag` | `status.outputs.zone_id` |

### What This Component Provides

After provisioning, `status.outputs` contains values that downstream Cloud Resources can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `site_tag` | The Cloudflare-assigned site tag — the site's identity in every RUM API path | RUM API queries, import recipes |
| `site_token` | The beacon's measurement token | Manual beacon embeds that build their own script tag |
| `snippet` | The ready-to-embed script tag, carrying the token | Pasting into page templates for hostname-identified sites |
| `ruleset_id` | The parent object the include/exclude rules live under (empty for host-identified sites — they have no ruleset) | Rule management via the RUM API |

The token ships inside public pages once deployed, so it is not a secret the way an API key is — the secret marking keeps it out of plan logs and CI output.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Zone with automatic installation** — measure a proxied zone and let the edge inject the beacon into every page: no snippet to paste, no deploy to coordinate. Start from the **Zone with automatic installation** preset.

**Host with path exclusions** — a hostname-identified site running the lite beacon, with admin and checkout paths excluded and everything else included; works whether or not the site is behind Cloudflare. Start from the **Host with path exclusions** preset.

## Works With

- [**Cloudflare DNS Zone**](/cloud-catalog/cloudflare-dns-zone) — the zone a zone-measured site references via `zoneTag`
- [**Cloudflare Logpush Job**](/cloud-catalog/cloudflare-logpush-job) — the server-side view of the same traffic, including non-browser requests
- [**Cloudflare Notification Policy**](/cloud-catalog/cloudflare-notification-policy) — alerting on the web-analytics metrics family
