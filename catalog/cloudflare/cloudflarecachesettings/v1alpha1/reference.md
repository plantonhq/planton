# CloudflareCacheSettings

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `cloudflare.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

CloudflareCacheSettingsSpec manages a zone's caching and performance posture:
tiered caching, Cache Reserve, regional tiered cache, cache variants, and Argo
Smart Routing.

A field left unset is NOT MANAGED: the module never sends it and the zone keeps
whatever value it already carries. Most of these settings also have NO DELETE at
Cloudflare -- destroying the resource abandons the last-applied values rather
than reverting them (smart tiered cache and cache variants are the exceptions
with real deletes). Argo Smart Routing deserves the loudest warning: it is a
PAID feature billed monthly plus usage, and because its delete is a no-op,
destroying this resource with argo_smart_routing still true KEEPS BILLING until
someone turns it off explicitly. Apply false first when retiring it.

Smart vs generic tiered caching are the same product family behind two API
objects: smart_tiered_cache picks the single best upper-tier data center
automatically (the recommended mode), while tiered_caching enables the generic
tiered topology. The Cloudflare dashboard's "Tiered Cache" toggle maps to
smart_tiered_cache.

## Example

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareCacheSettings
metadata:
  name: test-cache-settings
spec:
  zone_id:
    value: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  smart_tiered_cache: true
  regional_tiered_cache: false
  cache_variants:
    jpg:
      - image/webp
      - image/avif
    jpeg:
      - image/webp
    png:
      - image/webp
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.zoneId` | `string \| valueFrom` | yes |  | CloudflareDnsZone (`status.outputs.zone_id`) |
| `spec.smartTieredCache` | `bool` |  |  |  |
| `spec.tieredCaching` | `bool` |  |  |  |
| `spec.cacheReserve` | `bool` |  |  |  |
| `spec.regionalTieredCache` | `bool` |  |  |  |
| `spec.argoSmartRouting` | `bool` |  |  |  |
| `spec.cacheVariants` | `CloudflareCacheSettingsVariants` |  |  |  |
| `spec.cacheVariants.avif` | `[]string` |  |  |  |
| `spec.cacheVariants.bmp` | `[]string` |  |  |  |
| `spec.cacheVariants.gif` | `[]string` |  |  |  |
| `spec.cacheVariants.jp2` | `[]string` |  |  |  |
| `spec.cacheVariants.jpeg` | `[]string` |  |  |  |
| `spec.cacheVariants.jpg` | `[]string` |  |  |  |
| `spec.cacheVariants.jpg2` | `[]string` |  |  |  |
| `spec.cacheVariants.png` | `[]string` |  |  |  |
| `spec.cacheVariants.tif` | `[]string` |  |  |  |
| `spec.cacheVariants.tiff` | `[]string` |  |  |  |
| `spec.cacheVariants.webp` | `[]string` |  |  |  |

## Field Details

### spec.zoneId

`string | valueFrom` · required

The zone whose cache settings are managed.
When using value_from, defaults to CloudflareDnsZone kind and status.outputs.zone_id field path.

- references: CloudflareDnsZone (`status.outputs.zone_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: CloudflareDnsZone, name: <that resource's name>, fieldPath: status.outputs.zone_id}} -- a bare string does not parse

### spec.smartTieredCache

`bool` · optional (explicit presence)

Smart Tiered Cache: route cache misses through the single best upper-tier
data center for the origin (the dashboard's Tiered Cache toggle). This is the
one cache setting with a real delete at Cloudflare: destroying the resource
while this is managed disables smart tiered cache.

### spec.tieredCaching

`bool` · optional (explicit presence)

Generic tiered caching: enable the tiered cache topology where lower-tier
data centers ask upper tiers before the origin. Use smart_tiered_cache unless
a specific generic topology is required. No delete at Cloudflare -- destroy
abandons the last-applied value.

### spec.cacheReserve

`bool` · optional (explicit presence)

Cache Reserve: persist cached objects in Cloudflare's durable storage tier so
they survive cache eviction. Paid feature billed by storage and operations;
no delete at Cloudflare -- destroy abandons the last-applied value (turn it
off explicitly before retiring to stop the storage charges).

### spec.regionalTieredCache

`bool` · optional (explicit presence)

Regional Tiered Cache: add a regional tier between lower tiers and the
upper tier, cutting long-haul trips for cache misses in distant regions.
ENTERPRISE-GATED: zones below Enterprise answer 1135 "not available for
your plan type" on every read and write (measured on Free and Pro zones
at v5.23.0) -- manage this only on an Enterprise-entitled zone. No delete
at Cloudflare -- destroy abandons the last-applied value.

### spec.argoSmartRouting

`bool` · optional (explicit presence)

Argo Smart Routing: route origin-bound traffic over Cloudflare's measured
fastest paths. PAID: a monthly base fee plus per-GB usage. Cloudflare has no
delete for this setting -- destroying the resource while this is true KEEPS
THE FEATURE ON AND KEEPS BILLING. Apply false first when retiring it.

### spec.cacheVariants

`CloudflareCacheSettingsVariants`

Cache variants: which additional Accept-header content types Cloudflare may
serve for cached image extensions (for example serving image/webp variants
of .jpg URLs when Polish/WebP is active). PLAN-GATED at Free: a Free zone
answers 1135 "not available for your plan type" (measured at v5.23.0);
Pro and above accept it. Real delete at Cloudflare: destroying the
resource resets variants to defaults.

### spec.cacheVariants.avif

`[]string`

Variant content types for .avif assets.

### spec.cacheVariants.bmp

`[]string`

Variant content types for .bmp assets.

### spec.cacheVariants.gif

`[]string`

Variant content types for .gif assets.

### spec.cacheVariants.jp2

`[]string`

Variant content types for .jp2 assets.

### spec.cacheVariants.jpeg

`[]string`

Variant content types for .jpeg assets.

### spec.cacheVariants.jpg

`[]string`

Variant content types for .jpg assets.

### spec.cacheVariants.jpg2

`[]string`

Variant content types for .jpg2 assets.

### spec.cacheVariants.png

`[]string`

Variant content types for .png assets.

### spec.cacheVariants.tif

`[]string`

Variant content types for .tif assets.

### spec.cacheVariants.tiff

`[]string`

Variant content types for .tiff assets.

### spec.cacheVariants.webp

`[]string`

Variant content types for .webp assets.

## Validation Rules

- `spec.at_least_one_setting`: configure at least one cache setting -- a CloudflareCacheSettings resource that manages nothing would deploy nothing

## Outputs

Reference an output from another manifest as `valueFrom: {kind: CloudflareCacheSettings, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.zone_id` | `string` | The zone ID the cache settings belong to (the singleton's identity, and the pass-through for downstream resource references). |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.zoneId` | CloudflareDnsZone | `status.outputs.zone_id` |

## See Also

- [Overview](../README.md)
