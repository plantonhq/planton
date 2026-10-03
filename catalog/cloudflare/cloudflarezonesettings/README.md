# Cloudflare Zone Settings

## Overview

`CloudflareZoneSettings` manages a Cloudflare zone's behavior settings -- the HTTP, SSL/TLS, cache-adjacent, and network posture the Cloudflare dashboard shows under the Speed, Security, and Network tabs. One typed spec field maps to one Cloudflare zone setting (the API's `PATCH /zones/{zone_id}/settings/{setting_id}`), and the same resource also carries four companion surfaces that are zone-wide serving configuration by other API names: managed header transforms, URL normalization, origin cloud-region hints, and the waiting-room crawler bypass.

The central contract is presence-based management. A field left unset is NOT MANAGED: the module never sends it, and whatever value the zone already carries stays untouched. Setting a field manages that setting. Clearing it back to unset stops managing it but does NOT revert the live value -- Cloudflare has no delete for zone settings, so destroying the resource abandons the last-applied configuration on the zone. To return a setting to a specific value, set that value explicitly and apply before removing the field or destroying the resource.

## Key Features

- **One field per setting, in the API's own vocabulary**: 63 zone settings as typed spec fields whose names match Cloudflare's setting ids exactly (the single exception is `zero_rtt` for setting id `0rtt`, because proto identifiers cannot start with a digit)
- **Presence-based management**: only the fields you set are ever sent; the module never pushes defaults onto settings you did not ask it to own
- **Companion surfaces included**: managed request/response header transforms, URL normalization, per-origin-IP cloud-region hints, and the zone-wide waiting-room crawler bypass live in the same spec
- **Plan gating surfaced honestly**: settings the zone's plan cannot edit (for example `advanced_ddos` and `orange_to_orange` on Enterprise; `polish`, `mirage`, `image_resizing` on Pro and above) fail the apply with the API's editable=false error instead of silently skipping or upgrading the plan
- **Zone singleton**: one resource per zone, identified by the zone itself; the output is the zone id, passed through for downstream references

## Use Cases

**Ideal for:**

- Declaring a zone's production security posture in one place: HTTPS enforcement, minimum TLS version, SSL mode, security level, and HSTS headers
- Turning on performance features (Brotli, Early Hints, HTTP/3, Polish, cache level, browser cache TTL) as reviewable configuration instead of dashboard clicks
- Managing Cloudflare's managed header transforms (True-Client-IP, visitor location headers, removing `X-Powered-By`) alongside the settings they complement
- Pinning URL normalization and origin routing hints so rules and origins see consistent traffic

**Not ideal for:**

- DNS plumbing -- SOA tuning, NS TTL, DNS-level CNAME flattening, and DNSSEC belong to `CloudflareDnsZone`'s `dns_settings`; this kind owns how Cloudflare serves HTTP traffic, not how the zone resolves
- Per-path or per-request overrides -- use `CloudflareRuleset` (`http_config_settings` phase) to change settings for matching requests only
- Cache rules and cache reserve -- use `CloudflareCacheSettings`
- Per-hostname TLS material -- certificates and custom hostnames have their own kinds

## API Specification

### Required Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `zone_id` | StringValueOrRef | Yes | The zone whose settings are managed. Accepts a literal value or a reference to a `CloudflareDnsZone` (defaults to its `status.outputs.zone_id`). |

At least one setting must be configured -- a `CloudflareZoneSettings` resource that manages nothing would deploy nothing, and validation rejects it.

### Optional Fields

Every other field is optional, and unset means not managed. Rather than tabling all 63 settings, here is the shape of the fan-out:

**On/off toggles (39 settings)** -- booleans the module sends as the API's `"on"`/`"off"` strings. Highlights: `always_use_https`, `automatic_https_rewrites`, `brotli`, `early_hints`, `http2`, `http3`, `ipv6`, `websockets`, `zero_rtt` (setting id `0rtt`), `browser_check`, `email_obfuscation`, `hotlink_protection`, `ip_geolocation`, `rocket_loader`, `mirage`, `webp`, `development_mode`, `ssl_recommender`, `tls_client_auth`, and the plan-gated `advanced_ddos`, `orange_to_orange`, `prefetch_preload`, `response_buffering`, `sort_query_string_for_cache`, `true_client_ip_header`, plus `long_lived_grpc`.

**Enum settings (13 settings)** -- strings validated against Cloudflare's accepted values. Highlights: `ssl` (`off`/`flexible`/`full`/`strict`), `min_tls_version` (`1.0`-`1.3`), `security_level` (`off` through `under_attack`), `cache_level` (`basic`/`simplified`/`aggressive`), `polish` (`off`/`lossless`/`lossy`), `tls_1_3` (`on`/`off`/`zrt`), `h2_prioritization`, `image_resizing`, `pseudo_ipv4`, `cname_flattening`, `origin_max_http_version`, `transformations`, plus the free-string `transformations_allowed_origins`.

**Numeric settings (6 settings)** -- `browser_cache_ttl`, `challenge_ttl`, `edge_cache_ttl`, `max_upload`, `origin_h2_max_streams`, `proxy_read_timeout`. Cloudflare validates each against a fixed accepted value set (for example browser cache TTL accepts 0 and specific durations from 30 seconds to one year).

**Object settings** -- `security_header` (HSTS: enabled, include_subdomains, max_age, nosniff, preload), `nel` (network error logging), `aegis` (dedicated egress IPs, requires an Aegis entitlement), `automatic_platform_optimization` (APO for WordPress; the API requires every field of this object on writes), and the list-valued `ciphers` (BoringSSL-format TLS cipher allowlist; writing it requires the zone's Advanced Certificate Manager subscription -- the API rejects it with code 1023 otherwise).

**Companion surfaces** -- `managed_request_headers` and `managed_response_headers` (Cloudflare-defined transform toggles by id), `url_normalization` (scope + type), `origin_cloud_regions` (origin_ip + vendor + region, keyed by IP), and `waiting_room_crawler_bypass` (the zone-wide boolean; per-room configuration lives on the waiting room itself).

### Outputs

| Field | Description |
|-------|-------------|
| `zone_id` | The zone ID the settings belong to. Zone settings are a zone-scoped singleton with no resource id of their own -- the zone is the identity. |

## Example Manifests

Minimal -- manage exactly two settings and touch nothing else:

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareZoneSettings
metadata:
  name: baseline-settings
spec:
  zone_id:
    value: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  always_use_https: true
  min_tls_version: "1.2"
```

Production posture with companions:

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareZoneSettings
metadata:
  name: prod-zone-settings
spec:
  zone_id:
    value_from:
      kind: CloudflareDnsZone
      name: example-zone
  always_use_https: true
  ssl: strict
  min_tls_version: "1.2"
  security_level: medium
  http3: true
  security_header:
    enabled: true
    include_subdomains: true
    max_age: 31536000
    nosniff: true
    preload: false
  managed_response_headers:
    - id: remove_x-powered-by_header
      enabled: true
  url_normalization:
    scope: incoming
    type: cloudflare
```

## Destroy Semantics

Read this twice: destroying a `CloudflareZoneSettings` resource does not restore anything. Cloudflare's zone settings API has no delete -- the provider's destroy is a no-op for settings, and the URL normalization and managed transforms companions are the only pieces with a real delete. Whatever values were last applied keep serving. To retire the resource cleanly, first set every managed field to the value you want the zone to keep (or Cloudflare's default), apply, and only then destroy.

## Related Resources

- **CloudflareDnsZone**: creates the zone and owns DNS-level settings (SOA, NS TTL, DNSSEC); its `zone_id` output is this resource's foreign key
- **CloudflareCacheSettings**: cache rules, tiered caching, and cache reserve for the same zone
- **CloudflareZoneTlsSettings**: advanced per-zone TLS configuration beyond the settings fan-out
- **CloudflareRuleset**: per-request setting overrides via the `http_config_settings` phase

## Further Reading

For operational judgment -- the destroy contract, plan gating, and the boundary with DNS settings -- see GUIDE.md.

## References

- [Cloudflare Zone Settings API](https://developers.cloudflare.com/api/resources/zones/subresources/settings/)
- [Managed Transforms](https://developers.cloudflare.com/rules/transform/managed-transforms/)
- [URL Normalization](https://developers.cloudflare.com/rules/normalization/)

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
