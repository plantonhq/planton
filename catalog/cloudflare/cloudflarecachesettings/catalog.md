# Cloudflare Cache Settings

Manages a Cloudflare zone's caching and performance posture as one resource: Smart Tiered Cache, generic tiered caching, Cache Reserve, Regional Tiered Cache, Argo Smart Routing, and cache variants. A field left unset is not managed — the module never sends it, so the zone keeps whatever value it already carries, which makes this safe to adopt on a zone a human has been configuring for years. Destroy is not symmetric with create: only Smart Tiered Cache and cache variants have a real delete at Cloudflare, and Argo Smart Routing — a paid feature — keeps billing after destroy until someone applies `false` explicitly.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions one zone-scoped API object per managed setting:

- **Smart Tiered Cache** — one `cloudflare_tiered_cache`, created only when `smartTieredCache` is set; the dashboard's Tiered Cache toggle. Real delete: destroy disables it.
- **Generic Tiered Caching** — one `cloudflare_argo_tiered_caching`, created only when `tieredCaching` is set. No delete at Cloudflare.
- **Cache Reserve** — one `cloudflare_zone_cache_reserve`, created only when `cacheReserve` is set. Paid feature; no delete at Cloudflare.
- **Regional Tiered Cache** — one `cloudflare_regional_tiered_cache`, created only when `regionalTieredCache` is set. No delete at Cloudflare.
- **Argo Smart Routing** — one `cloudflare_argo_smart_routing`, created only when `argoSmartRouting` is set. Paid feature; no delete at Cloudflare — destroy keeps it on and keeps billing.
- **Cache Variants** — one `cloudflare_zone_cache_variants`, created only when `cacheVariants` is set; only extensions with at least one MIME type are sent. Real delete: destroy resets variants to Cloudflare defaults.

## Before You Deploy

### Planton Setup

- **Cloudflare Provider Connection** — an active connection in the Connect module whose API token has Cache Settings edit access on the target zone; Argo Smart Routing additionally requires the Argo edit permission. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** — required when using Runner-based credential delivery. Not needed for inline credential authentication.

### Cloudflare Account

- **An active zone** for `zoneId` — manage it with a CloudflareDnsZone and reference its `zone_id` output, or pass an externally managed zone ID directly.
- **Plan entitlements for the paid features** (only for `cacheReserve` / `argoSmartRouting`) — the account must have Cache Reserve and Argo available before the apply can enable them.
- **A variant-producing pipeline** (only for `cacheVariants`) — variants tell the cache which alternate content types are acceptable; something like Polish with WebP must actually produce them.

## Deploy

### Console

Open the deployment store, find **Cloudflare Cache Settings**, and click **Deploy**. The creation wizard walks you through environment and connection configuration, the target zone, the six cache toggles, and the per-extension variant lists. Start from the **Tiered Caching** preset in the [Presets](#presets) tab to pre-populate a working configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareCacheSettings
metadata:
  name: prod-cache
  org: acme-corp
  env: prod
spec:
  zoneId:
    value: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  smartTieredCache: true
```

```shell
planton apply -f cache-settings.yaml
```

This enables Smart Tiered Cache on the zone and manages nothing else — every other setting stays exactly as it is. An Infra Job tracks the provisioning in real time.

### InfraChart

When the zone is deployed in the same InfraPipeline, wire the reference with ValueFromRef:

```yaml
spec:
  zoneId:
    valueFrom:
      kind: CloudflareDnsZone
      name: prod-zone
      fieldPath: status.outputs.zone_id
  smartTieredCache: true
  regionalTieredCache: true
```

The InfraPipeline resolves the dependency graph, deploys the zone first, then applies the cache settings to the resolved zone ID.

## Key Configuration

These are the most important decisions when configuring cache settings. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Smart over generic** — `smartTieredCache` and `tieredCaching` are the same product family behind two API objects: smart picks the single best upper-tier data center for your origin automatically and is what the dashboard's Tiered Cache toggle controls; generic exists for the rare case where a specific topology is required. Use smart unless you know why you need generic — managing both is legal but pointless.

**Unset means unmanaged** — an unset field is never sent, so the zone keeps whatever a dashboard click or another tool set. The flip side: removing a field from the spec does not turn the setting off, it abandons the last-applied value. To turn a setting off, apply it as `false`; to stop managing it, remove it after that apply. A manifest with only `zoneId` is rejected — a resource that manages nothing would deploy nothing.

**Argo Smart Routing bills past destroy** — it costs a monthly base fee plus per-GB usage on origin-bound traffic, and its delete is a no-op at Cloudflare: destroying this resource with `argoSmartRouting: true` keeps the feature on and keeps the charges running. The retirement sequence for paid toggles is always: apply `false`, confirm, then destroy.

**Cache Reserve accumulates storage cost** — it persists cached objects in Cloudflare's durable storage tier, billed by storage volume and operations, and the charges continue as long as it is on regardless of traffic. It also has no delete — turn it off explicitly before retiring.

**The two real deletes are the opposite trap** — destroying a resource that manages `smartTieredCache` disables it, and one managing `cacheVariants` resets variants to defaults — even if something downstream depended on them.

**Variants describe, they don't convert** — `cacheVariants` maps original-asset extensions (`jpg`, `png`, `webp`, …) to the MIME types Cloudflare may serve as variants. Extensions with empty lists are omitted, never sent as empty arrays, so you cannot accidentally clear one extension's variants by listing it empty. Manifests always use the singular extension names, even though the Pulumi SDK pluralizes them internally.

**Per-URL behavior lives elsewhere** — TTL overrides, custom cache keys, and bypass rules belong to a Cloudflare Ruleset in the cache-settings phase. If you find yourself wanting "cache this path differently", that is a ruleset, not this resource.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **CloudflareDnsZone** | `zoneId` | `status.outputs.zone_id` |

### What This Kind Provides

This component's `status.outputs` only echoes the managed zone's ID back (`zone_id`) — cache settings are a zone singleton with no resource ID of their own, so the zone is the identity and there is nothing new for downstream resources to consume.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Tiered caching for production** — Smart Tiered Cache plus Regional Tiered Cache: the free, recommended posture — fewer origin requests, shorter long-haul misses, no cost. Start from the **Tiered Caching** preset.

**Image variants for a media zone** — Smart Tiered Cache plus WebP/AVIF variants for cached JPEG and PNG assets, paired with a pipeline that actually produces them. Start from the **Image Variants** preset.

**Full performance posture** — tiered caching, Cache Reserve for long-tail assets, and Argo Smart Routing for origin-bound traffic. Both paid toggles are managed in one manifest, so retiring the zone means applying them as `false` before any destroy.

## Works With

- [**Cloudflare DNS Zone**](/infra-catalog/cloudflare-dns-zone) — the anchor; wire `zoneId` via ValueFromRef so the dependency is explicit in the graph
- [**Cloudflare Ruleset**](/infra-catalog/cloudflare-ruleset) — per-URL TTLs, cache keys, and bypass rules layered on top of the zone-wide posture set here
- [**Cloudflare Zone Settings**](/infra-catalog/cloudflare-zone-settings) — the sibling settings kind for general zone toggles (SSL mode, minification, security level)
- [**Cloudflare Zone TLS Settings**](/infra-catalog/cloudflare-zone-tls-settings) — the sibling settings kind for the zone's TLS posture
