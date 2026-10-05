# CloudflareCacheSettings Guide

The judgment this guide protects you from: assuming destroy reverts what apply enabled. Four of the six settings here have no delete operation at Cloudflare, and one of them (Argo Smart Routing) bills money monthly plus per-GB. Destroying this resource with `argoSmartRouting: true` does not stop the charges -- nothing does, except explicitly applying `false`.

## When to use it (and when not)

This kind owns the zone-wide caching and performance toggles. It does not own per-URL cache behavior: TTL overrides, custom cache keys, and bypass rules belong to `CloudflareRuleset` in the cache settings phase. If you find yourself wanting "cache this path differently," that is a ruleset, not this resource. General zone toggles (SSL mode, minification, security level) live in `CloudflareZoneSettings`, and TLS posture in `CloudflareZoneTlsSettings` -- the three settings kinds partition the zone's configuration surface so each manifest stays reviewable.

Between the two tiered caching fields, use `smartTieredCache`. Smart and generic tiered caching are the same product family behind two API objects: smart picks the single best upper-tier data center for your origin automatically and is what the dashboard's "Tiered Cache" toggle controls; `tieredCaching` enables the generic topology and exists for the rare case where a specific topology is required. Managing both is legal but pointless for almost everyone.

## Unset means unmanaged

Every settings field is optional, and an unset field is never sent -- the zone keeps whatever value it carries, whether that came from the dashboard, another tool, or a Cloudflare default. This is the adoption path: you can put this resource on a zone that a human has been clicking at for years and manage only `smartTieredCache` without disturbing anything else. The flip side: removing a field from the spec does not turn the setting off. It stops managing it, which abandons the last-applied value. To turn a setting off, apply it as `false`; to stop managing it, remove it after that apply.

Validation enforces at least one managed setting -- a manifest with only `zoneId` is rejected, because a resource that manages nothing would deploy nothing.

## Destroy classes: two real deletes, four no-ops

| Setting | Cloudflare delete | What destroy actually does |
|---------|-------------------|----------------------------|
| `smartTieredCache` | real | Disables Smart Tiered Cache |
| `cacheVariants` | real | Resets variants to Cloudflare defaults |
| `tieredCaching` | none | Abandons the last-applied value |
| `regionalTieredCache` | none | Abandons the last-applied value |
| `cacheReserve` | none | Abandons the last-applied value; storage keeps billing while on |
| `argoSmartRouting` | none | Abandons the last-applied value; KEEPS BILLING if true |

The operational consequence: destroying this resource is not symmetric with creating it. A zone that had `argoSmartRouting: true` and `cacheReserve: true` applied keeps both features on -- and both bills running -- after the resource is gone. The retirement sequence for paid toggles is always: apply `false`, confirm the apply succeeded, then destroy. The two real-delete settings are the opposite trap in miniature: destroying a resource that manages `cacheVariants` resets variants to defaults even if someone downstream depended on them.

## Cost behavior

Two fields cost money. `argoSmartRouting` is a monthly base fee plus per-GB usage on origin-bound traffic; it is the loudest warning in this kind because its no-op delete means the bill outlives the resource. `cacheReserve` bills by storage volume and operations -- cheaper to reason about, but it accumulates stored objects, so the charges continue as long as it is on regardless of traffic. Everything else here (`smartTieredCache`, `tieredCaching`, `regionalTieredCache`, `cacheVariants`) is free on any plan that has the features.

## Conventions and gotchas

- **Variant field names differ between the IaC backends.** The spec and the Terraform module use the API's singular extension names (`jpg`, `png`, `webp`); the Pulumi SDK pluralizes them (`Jpgs`, `Pngs`, `Webps`). The modules map between them, so manifests always use the singular spelling -- but if you are reading provider-level diffs or SDK code, expect the pluralized names on the Pulumi side.
- **Only managed extensions are sent.** An extension with an empty list is omitted, never sent as an empty array -- both modules enforce this, so you cannot accidentally clear one extension's variants by listing it empty. To reset variants, destroy the resource (real delete) or apply the desired end state.
- **The Terraform module pins `cloudflare/cloudflare ~> 5.23`.** The cache settings resources (`cloudflare_tiered_cache`, `cloudflare_zone_cache_reserve`, `cloudflare_zone_cache_variants`, and friends) took their current shape in the v5 provider; the pin is deliberate, not incidental.
- **Outputs are just `zone_id`.** Cache settings are a zone singleton with no resource ID of their own -- the zone is the identity. Downstream references pass the zone through, nothing more.

## On the diagram

One CloudflareCacheSettings node per zone renders the zone's whole cache posture as a single reviewable spec next to its `CloudflareDnsZone`. That is the point of the folding: six dashboard toggles become one node whose diff history answers "who turned on Argo and when." Splitting the settings across multiple resources of this kind for one zone is possible but buys nothing and risks two manifests fighting over the same toggle.

## Pairs well with

- **CloudflareDnsZone** -- the anchor; reference it via `zoneId` with `valueFrom` so the dependency is explicit in the graph.
- **CloudflareRuleset** (cache settings phase) -- per-URL TTLs, cache keys, and bypass rules layered on top of the zone-wide posture set here.
- **CloudflareZoneSettings / CloudflareZoneTlsSettings** -- the sibling settings kinds covering the rest of the zone's configuration surface.
