# Cloudflare Zone Settings

Manages a Cloudflare zone's behavior settings — the serving posture the dashboard shows under the Speed, Security, and Network tabs — as one typed field per Cloudflare setting ID, plus managed header transforms, URL normalization, origin cloud-region hints, and the zone-wide waiting-room crawler bypass. Only the fields you set are managed: unset fields are never sent, and because Cloudflare has no delete for zone settings, destroying the resource abandons the last-applied values rather than reverting them. One settings resource per zone — it is a zone-scoped singleton.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Zone Settings** — one zone-setting write per managed field, in Cloudflare's own vocabulary (63 settings: on/off toggles, enums, numerics, and object values like HSTS)
- **Managed Transforms** — created only when `managedRequestHeaders` or `managedResponseHeaders` has entries; one zone-wide object toggling Cloudflare-defined header transforms (both lists always travel together — a transform missing from the list reads as disabled)
- **URL Normalization Settings** — created only when `urlNormalization` is set; controls how Cloudflare rewrites URLs before rules match and before they reach the origin
- **Origin Cloud Regions** — created only when `originCloudRegions` has entries; one region hint per origin IP so Cloudflare can optimize origin-facing routing
- **Waiting Room Settings** — created only when `waitingRoomCrawlerBypass` is set; the zone-wide toggle letting verified search crawlers bypass waiting rooms

Zone settings and the crawler-bypass toggle have no delete at Cloudflare, so destroy abandons their live values. Managed transforms and URL normalization do have a real delete and reset on destroy.

## Before You Deploy

### Planton Setup

- **Cloudflare Provider Connection** — an active connection in the Connect module with an API token holding Zone → Zone Settings → Edit on the target zone; managed transforms, URL normalization, and origin cloud regions ride the same zone-scoped token. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** — required when using Runner-based credential delivery. Not needed for inline credential authentication.

### Cloudflare Account

- **A zone on the account** — `zoneId` names the zone; reference a CloudflareDnsZone Infra Component or pass the zone ID from the dashboard.
- **The right plan for gated settings** (only for those fields) — `advancedDdos`, `orangeToOrange`, `prefetchPreload`, `responseBuffering`, `sortQueryStringForCache`, `trueClientIpHeader`, and `proxyReadTimeout` need Enterprise; `polish`, `mirage`, and `imageResizing` need Pro or above. The apply fails with the API's editable=false error on a plan that lacks a setting; nothing is billed or upgraded. `ciphers` is gated by product, not plan: writes need the zone's Advanced Certificate Manager subscription (API code 1023 without it) even though the settings list reports it editable.

## Deploy

### Console

Open the deployment store, find **Cloudflare Zone Settings**, and click **Deploy**. The creation wizard walks you through environment and connection configuration, the zone reference, and the settings groups — security, TLS, caching, protocol toggles, and the companion surfaces. Start from the **Minimal** or **Production Hardened** preset in the [Presets](#presets) tab to pre-populate a working posture.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareZoneSettings
metadata:
  name: baseline-settings
  org: acme-corp
  env: prod
spec:
  zoneId:
    value: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  alwaysUseHttps: true
  minTlsVersion: "1.2"
```

```shell
planton apply -f zone-settings.yaml
```

This manages exactly two settings — every plain-http request redirects to HTTPS, and TLS below 1.2 is refused — while all other zone settings stay exactly as they were. An Infra Job tracks the provisioning in real time.

### InfraChart

When the zone is deployed in the same InfraPipeline, wire `zoneId` with ValueFromRef:

```yaml
spec:
  zoneId:
    valueFrom:
      kind: CloudflareDnsZone
      name: acme-com
      fieldPath: status.outputs.zone_id
  alwaysUseHttps: true
  ssl: strict
  minTlsVersion: "1.2"
```

The InfraPipeline resolves the dependency graph, deploys the zone first, then applies the settings against the resolved zone ID.

## Key Configuration

These are the most important decisions when configuring zone settings. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Destroy restores nothing — revert before you retire** — Cloudflare's zone settings API has no delete, so destroy drops the resource from state and walks away while the live values keep serving. If you set `securityLevel: under_attack` during an incident and later delete the resource, the zone stays in under-attack mode indefinitely with nothing in your infrastructure declaring it. Before removing a field or destroying the resource, set every managed field to the value the zone should keep, apply, then remove.

**Unset means unmanaged, by design** — the module never sends defaults, so this resource composes with dashboard-managed settings: managing `alwaysUseHttps` here does not stop anyone from flipping `rocketLoader` in the dashboard. But for a field you DO manage, every apply reasserts your value — the manifest wins any tug-of-war, one apply at a time. A spec with only `zoneId` is rejected: a resource that manages nothing would deploy nothing.

**Encryption mode to the origin** — `ssl: strict` encrypts to the origin and validates its certificate; production zones should use it. `flexible` terminates TLS at the edge and speaks plain http to the origin — a downgrade dressed as a convenience. Pair `alwaysUseHttps: true` with `minTlsVersion: "1.2"` (quote the value in YAML) for the standard floor.

**HSTS is a browser-side commitment** — `securityHeader` with `includeSubdomains: true` and a one-year `maxAge` pins browsers to HTTPS across every subdomain for that long; enable it only when every subdomain serves HTTPS. `preload: true` requests inclusion in browser preload lists, and removal from those lists is slow and manual — a one-way door in practice.

**Plan-gated settings fail at apply** — the API reports a setting the plan cannot edit as editable=false and the module surfaces it as an apply error. The fix is to remove the field or upgrade the plan (a CloudflareDnsZone `subscription` decision, and real money) — not to retry.

**Cache baseline versus cache rules** — `cacheLevel`, `browserCacheTtl`, and `edgeCacheTtl` set the zone-wide caching baseline; per-path overrides belong to CloudflareCacheSettings, and per-request setting overrides to a CloudflareRuleset in the `http_config_settings` phase. `browserCacheTtl` and the other numeric settings accept only Cloudflare's fixed value sets — out-of-set values are rejected at apply.

**The DNS boundary** — CloudflareDnsZone's `dnsSettings` owns how the zone resolves (SOA tuning, NS TTL, DNSSEC); this kind owns how Cloudflare serves HTTP. The `cnameFlattening` field here is the serving-side setting of that name, exposed through the settings API.

**Field-name quirks the module absorbs** — `zeroRtt` is Cloudflare's setting ID `0rtt` (the one name that differs, since proto identifiers cannot start with a digit); `automaticPlatformOptimization` requires every member of its object on every write; `sslRecommender` is written in the value form the provider actually accepts. You just set the fields.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **CloudflareDnsZone** | `zoneId` | `status.outputs.zone_id` |

### What This Kind Provides

This kind has no consumable outputs of its own: zone settings are a zone-scoped singleton with no resource ID, so `status.outputs` only echoes the input `zone_id` back for reference. Downstream resources that need the zone should reference the CloudflareDnsZone directly.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Minimal HTTPS floor** — manage exactly `alwaysUseHttps` and `minTlsVersion`, leaving everything else to the dashboard; the safest first adoption on a zone with hand-tuned settings, growing field by field from there. Start from the **Minimal** preset.

**Production security posture** — HTTPS everywhere, `ssl: strict`, TLS 1.2 floor, a one-year HSTS pin with subdomains, HTTP/3, and `securityLevel: medium`. Start from the **Production Hardened** preset.

**Content-zone performance** — Brotli, Early Hints, HTTP/3, aggressive cache level, a four-hour browser TTL, and lossless Polish (Pro or above) for marketing and documentation zones. Start from the **Performance** preset.

**Header hygiene companions** — toggle managed transforms (`add_true_client_ip_headers` on requests, `remove_x-powered-by_header` on responses) and pin URL normalization alongside the settings, all in one manifest.

## Works With

- [**Cloudflare DNS Zone**](/infra-catalog/cloudflare-dns-zone) — creates the zone this resource configures and owns the resolver-side settings; its `zone_id` output is this resource's foreign key
- [**Cloudflare Cache Settings**](/infra-catalog/cloudflare-cache-settings) — cache rules, tiered caching, and cache reserve layered over this kind's zone-wide caching baseline
- [**Cloudflare Zone TLS Settings**](/infra-catalog/cloudflare-zone-tls-settings) — the advanced per-zone TLS surface beyond this spec's `ssl`/`minTlsVersion`/`ciphers`
- [**Cloudflare Ruleset**](/infra-catalog/cloudflare-ruleset) — per-request overrides of these zone-wide settings via the `http_config_settings` phase
