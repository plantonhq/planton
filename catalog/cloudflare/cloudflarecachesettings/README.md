# Cloudflare Cache Settings

## Overview

`CloudflareCacheSettings` is a resource for managing a Cloudflare zone's caching and performance posture in one place: Smart Tiered Cache, generic tiered caching, Cache Reserve, Regional Tiered Cache, Argo Smart Routing, and cache variants.

Each of these settings is its own zone-scoped object in Cloudflare's API, but they describe one thing: how the zone caches and moves traffic. This kind folds them into a single manifest so the zone's cache posture is declared, reviewed, and versioned together. A field left unset is NOT MANAGED -- the module never sends it, and the zone keeps whatever value it already carries. That makes it safe to adopt this kind incrementally: manage only the settings you care about and leave the rest untouched.

## Key Features

- **One manifest, whole posture**: All six cache and performance settings for a zone live in a single resource instead of six separate toggles scattered across the dashboard
- **Unset means unmanaged**: Only fields you set are sent to Cloudflare; everything else is left exactly as it is
- **Smart Tiered Cache by default**: The recommended tiered caching mode -- Cloudflare picks the single best upper-tier data center for your origin (this is the dashboard's "Tiered Cache" toggle)
- **Honest destroy semantics**: The spec documents which settings Cloudflare actually deletes on destroy and which it silently abandons, so retiring a zone never leaves a surprise bill

## Use Cases

**Ideal for:**
- Turning on tiered caching for a production zone (Smart Tiered Cache plus Regional Tiered Cache is a free, sensible default)
- Serving WebP/AVIF variants of cached JPEG and PNG assets alongside an image optimization pipeline
- Enabling Cache Reserve for zones with long-tail assets that keep falling out of edge cache
- Enabling Argo Smart Routing for latency-sensitive, origin-heavy traffic -- with the billing lifecycle managed declaratively

**Not ideal for:**
- Per-URL cache rules (TTLs, cache keys, bypass rules) -- use `CloudflareRuleset` with the cache settings phase instead
- General zone settings like SSL mode, minification, or security level -- use `CloudflareZoneSettings`
- TLS-specific posture -- use `CloudflareZoneTlsSettings`

## API Specification

### CloudflareCacheSettingsSpec

At least one setting must be configured -- a resource that manages nothing would deploy nothing.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `zone_id` | StringValueOrRef | Yes | The zone whose cache settings are managed. Accepts a literal `value` or a `value_from` reference to a `CloudflareDnsZone` (defaults to its `status.outputs.zone_id`). |
| `smart_tiered_cache` | bool (optional) | No | Route cache misses through the single best upper-tier data center (the dashboard's Tiered Cache toggle). Real delete: destroy disables it. |
| `tiered_caching` | bool (optional) | No | Generic tiered cache topology. Use `smart_tiered_cache` unless a specific generic topology is required. No delete at Cloudflare. |
| `cache_reserve` | bool (optional) | No | Persist cached objects in Cloudflare's durable storage tier. Paid, billed by storage and operations. No delete at Cloudflare -- turn it off explicitly before retiring. |
| `regional_tiered_cache` | bool (optional) | No | Add a regional tier between lower tiers and the upper tier. No delete at Cloudflare. |
| `argo_smart_routing` | bool (optional) | No | Route origin-bound traffic over Cloudflare's measured fastest paths. PAID: monthly base fee plus per-GB usage. No delete at Cloudflare -- destroying with this true KEEPS BILLING. Apply `false` first when retiring. |
| `cache_variants` | CloudflareCacheSettingsVariants | No | Per-extension lists of content types Cloudflare may serve as variants of cached assets. Real delete: destroy resets variants to defaults. |

### CloudflareCacheSettingsVariants

Each field names the file extension of the original asset; the list carries the MIME types of acceptable variants (for example `image/webp`, `image/avif`). Only extensions with at least one value are sent.

| Field | Type | Description |
|-------|------|-------------|
| `avif` | repeated string | Variant content types for `.avif` assets |
| `bmp` | repeated string | Variant content types for `.bmp` assets |
| `gif` | repeated string | Variant content types for `.gif` assets |
| `jp2` | repeated string | Variant content types for `.jp2` assets |
| `jpeg` | repeated string | Variant content types for `.jpeg` assets |
| `jpg` | repeated string | Variant content types for `.jpg` assets |
| `jpg2` | repeated string | Variant content types for `.jpg2` assets |
| `png` | repeated string | Variant content types for `.png` assets |
| `tif` | repeated string | Variant content types for `.tif` assets |
| `tiff` | repeated string | Variant content types for `.tiff` assets |
| `webp` | repeated string | Variant content types for `.webp` assets |

### Outputs

After successful deployment, the following outputs are available:

| Field | Description |
|-------|-------------|
| `zone_id` | The zone the cache settings belong to. Cache settings are a zone-scoped singleton with no resource ID of their own -- the zone is the identity. |

## Destroy Behavior

This is the part worth reading twice. Most of these settings have NO delete operation at Cloudflare: destroying the resource drops the IaC state and abandons the last-applied live value rather than reverting it.

| Setting | On destroy |
|---------|-----------|
| `smart_tiered_cache` | Real delete: Cloudflare disables Smart Tiered Cache |
| `cache_variants` | Real delete: Cloudflare resets variants to defaults |
| `tiered_caching` | No delete: last-applied value stays on the zone |
| `cache_reserve` | No delete: last-applied value stays on the zone (storage keeps billing while on) |
| `regional_tiered_cache` | No delete: last-applied value stays on the zone |
| `argo_smart_routing` | No delete: last-applied value stays on the zone -- IF TRUE, BILLING CONTINUES |

The practical rule: before destroying a resource that manages a paid toggle (`argo_smart_routing`, `cache_reserve`), apply it as `false` first, wait for the apply to succeed, then destroy. Destroy alone does not stop the charges.

## Example Manifests

Enable tiered caching for a production zone (free, recommended posture):

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareCacheSettings
metadata:
  name: prod-cache-settings
spec:
  zone_id:
    valueFrom:
      kind: CloudflareDnsZone
      name: prod-zone
      fieldPath: status.outputs.zone_id
  smart_tiered_cache: true
  regional_tiered_cache: true
```

Serve modern image formats as variants of cached assets:

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareCacheSettings
metadata:
  name: image-cache-settings
spec:
  zone_id:
    value: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  smart_tiered_cache: true
  cache_variants:
    jpg:
      - image/webp
      - image/avif
    jpeg:
      - image/webp
    png:
      - image/webp
```

## Related Resources

- **Cloudflare DNS Zone**: The zone whose cache posture this resource manages; reference it via `zone_id.valueFrom`
- **Cloudflare Zone Settings**: General zone settings (SSL mode, minification, security level)
- **Cloudflare Zone TLS Settings**: The zone's TLS-specific posture
- **Cloudflare Ruleset**: Per-URL cache rules (TTLs, cache keys, bypass) via the cache settings phase

## Further Reading

For destroy-semantics details, cost behavior, and naming gotchas between the Terraform and Pulumi implementations, see GUIDE.md.

## References

- [Cloudflare Tiered Cache](https://developers.cloudflare.com/cache/how-to/tiered-cache/)
- [Cloudflare Cache Reserve](https://developers.cloudflare.com/cache/advanced-configuration/cache-reserve/)
- [Cloudflare Argo Smart Routing](https://developers.cloudflare.com/argo-smart-routing/)
- [Cloudflare Cache Variants](https://developers.cloudflare.com/cache/advanced-configuration/caching-variants/)

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
