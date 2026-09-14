# CloudflareZoneSettings

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `cloudflare.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

CloudflareZoneSettingsSpec manages a zone's behavior settings: the HTTP, SSL/TLS,
cache-adjacent, and network posture the Cloudflare dashboard shows under the
Speed, Security, and Network tabs.

Each setting field maps to one Cloudflare zone setting (the API's
PATCH /zones/{zone_id}/settings/{setting_id}). A field left unset is NOT MANAGED:
the module never sends it, and whatever value the zone already carries stays
untouched. Setting a field manages that setting; clearing it back to unset stops
managing it but does NOT revert the live value -- Cloudflare has no delete for
zone settings, so destroy abandons the last-applied configuration. To return a
setting to a specific value, set that value explicitly before removing the field.

Boundary: DNS plumbing (SOA, NS TTL, CNAME flattening at the DNS level, DNSSEC)
belongs to CloudflareDnsZone's dns_settings. This kind owns the serving posture:
how Cloudflare terminates, caches, transforms, and protects HTTP traffic for the
zone. Managed transforms, URL normalization, origin cloud-region hints, and the
waiting-room crawler bypass are zone-wide serving configuration by other API
names, so they live here too.

Plan gating: several settings are only editable on paid plans (for example
advanced_ddos, orange_to_orange, image_resizing). The API answers with
editable=false when the zone's plan does not include a setting; the module
surfaces that as an apply error rather than silently skipping.

## Example

```yaml
# Offline shape proof across every value class. NOTE: ciphers stays here as
# the list value class's shape proof, but it is NOT free-plan-deployable --
# writing it requires the zone's Advanced Certificate Manager subscription
# (400 code 1023 measured live, 2026-08-27), so the live scenarios omit it.
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareZoneSettings
metadata:
  name: test-zone-settings
spec:
  zone_id:
    value: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  always_use_https: true
  min_tls_version: "1.2"
  ssl: strict
  browser_cache_ttl: 14400
  security_level: medium
  ciphers:
    - "ECDHE-ECDSA-AES128-GCM-SHA256"
    - "ECDHE-RSA-AES128-GCM-SHA256"
  security_header:
    enabled: true
    include_subdomains: true
    max_age: 31536000
    nosniff: true
    preload: false
  managed_request_headers:
    - id: add_true_client_ip_headers
      enabled: true
  managed_response_headers:
    - id: remove_x-powered-by_header
      enabled: true
  url_normalization:
    scope: incoming
    type: cloudflare
  origin_cloud_regions:
    - origin_ip: "203.0.113.10"
      region: us-east-1
      vendor: aws
  waiting_room_crawler_bypass: false
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.zoneId` | `string \| valueFrom` | yes |  | CloudflareDnsZone (`status.outputs.zone_id`) |
| `spec.managedRequestHeaders` | `[]CloudflareZoneSettingsManagedHeader` |  |  |  |
| `spec.managedRequestHeaders[].id` | `string` | yes |  |  |
| `spec.managedRequestHeaders[].enabled` | `bool` |  |  |  |
| `spec.managedResponseHeaders` | `[]CloudflareZoneSettingsManagedHeader` |  |  |  |
| `spec.managedResponseHeaders[].id` | `string` | yes |  |  |
| `spec.managedResponseHeaders[].enabled` | `bool` |  |  |  |
| `spec.urlNormalization` | `CloudflareZoneSettingsUrlNormalization` |  |  |  |
| `spec.urlNormalization.scope` | `string` | yes |  |  |
| `spec.urlNormalization.type` | `string` | yes |  |  |
| `spec.originCloudRegions` | `[]CloudflareZoneSettingsOriginCloudRegion` |  |  |  |
| `spec.originCloudRegions[].originIp` | `string` | yes |  |  |
| `spec.originCloudRegions[].region` | `string` | yes |  |  |
| `spec.originCloudRegions[].vendor` | `string` | yes |  |  |
| `spec.waitingRoomCrawlerBypass` | `bool` |  |  |  |
| `spec.zeroRtt` | `bool` |  |  |  |
| `spec.advancedDdos` | `bool` |  |  |  |
| `spec.aegis` | `CloudflareZoneSettingsAegis` |  |  |  |
| `spec.aegis.enabled` | `bool` |  |  |  |
| `spec.aegis.poolId` | `string` |  |  |  |
| `spec.alwaysOnline` | `bool` |  |  |  |
| `spec.alwaysUseHttps` | `bool` |  |  |  |
| `spec.automaticHttpsRewrites` | `bool` |  |  |  |
| `spec.automaticPlatformOptimization` | `CloudflareZoneSettingsAutomaticPlatformOptimization` |  |  |  |
| `spec.automaticPlatformOptimization.enabled` | `bool` | yes |  |  |
| `spec.automaticPlatformOptimization.cacheByDeviceType` | `bool` |  |  |  |
| `spec.automaticPlatformOptimization.cf` | `bool` |  |  |  |
| `spec.automaticPlatformOptimization.hostnames` | `[]string` | yes |  |  |
| `spec.automaticPlatformOptimization.wordpress` | `bool` |  |  |  |
| `spec.automaticPlatformOptimization.wpPlugin` | `bool` |  |  |  |
| `spec.brotli` | `bool` |  |  |  |
| `spec.browserCacheTtl` | `int64` |  |  |  |
| `spec.browserCheck` | `bool` |  |  |  |
| `spec.cacheLevel` | `string` |  |  |  |
| `spec.challengeTtl` | `int64` |  |  |  |
| `spec.ciphers` | `[]string` |  |  |  |
| `spec.cnameFlattening` | `string` |  |  |  |
| `spec.contentConverter` | `bool` |  |  |  |
| `spec.developmentMode` | `bool` |  |  |  |
| `spec.earlyHints` | `bool` |  |  |  |
| `spec.edgeCacheTtl` | `int64` |  |  |  |
| `spec.emailObfuscation` | `bool` |  |  |  |
| `spec.h2Prioritization` | `string` |  |  |  |
| `spec.hotlinkProtection` | `bool` |  |  |  |
| `spec.http2` | `bool` |  |  |  |
| `spec.http3` | `bool` |  |  |  |
| `spec.imageResizing` | `string` |  |  |  |
| `spec.ipGeolocation` | `bool` |  |  |  |
| `spec.ipv6` | `bool` |  |  |  |
| `spec.maxUpload` | `int64` |  |  |  |
| `spec.minTlsVersion` | `string` |  |  |  |
| `spec.mirage` | `bool` |  |  |  |
| `spec.nel` | `CloudflareZoneSettingsNel` |  |  |  |
| `spec.nel.enabled` | `bool` | yes |  |  |
| `spec.opportunisticEncryption` | `bool` |  |  |  |
| `spec.opportunisticOnion` | `bool` |  |  |  |
| `spec.orangeToOrange` | `bool` |  |  |  |
| `spec.originErrorPagePassThru` | `bool` |  |  |  |
| `spec.originH2MaxStreams` | `int64` |  |  |  |
| `spec.originMaxHttpVersion` | `string` |  |  |  |
| `spec.polish` | `string` |  |  |  |
| `spec.prefetchPreload` | `bool` |  |  |  |
| `spec.privacyPass` | `bool` |  |  |  |
| `spec.proxyReadTimeout` | `int64` |  |  |  |
| `spec.pseudoIpv4` | `string` |  |  |  |
| `spec.redirectsForAiTraining` | `bool` |  |  |  |
| `spec.replaceInsecureJs` | `bool` |  |  |  |
| `spec.responseBuffering` | `bool` |  |  |  |
| `spec.rocketLoader` | `bool` |  |  |  |
| `spec.searchForAgents` | `bool` |  |  |  |
| `spec.securityHeader` | `CloudflareZoneSettingsSecurityHeader` |  |  |  |
| `spec.securityHeader.enabled` | `bool` | yes |  |  |
| `spec.securityHeader.includeSubdomains` | `bool` |  |  |  |
| `spec.securityHeader.maxAge` | `int64` |  |  |  |
| `spec.securityHeader.nosniff` | `bool` |  |  |  |
| `spec.securityHeader.preload` | `bool` |  |  |  |
| `spec.securityLevel` | `string` |  |  |  |
| `spec.serverSideExclude` | `bool` |  |  |  |
| `spec.sha1Support` | `bool` |  |  |  |
| `spec.sortQueryStringForCache` | `bool` |  |  |  |
| `spec.ssl` | `string` |  |  |  |
| `spec.sslRecommender` | `bool` |  |  |  |
| `spec.tls12Only` | `bool` |  |  |  |
| `spec.tls13` | `string` |  |  |  |
| `spec.tlsClientAuth` | `bool` |  |  |  |
| `spec.transformations` | `string` |  |  |  |
| `spec.transformationsAllowedOrigins` | `string` |  |  |  |
| `spec.trueClientIpHeader` | `bool` |  |  |  |
| `spec.waf` | `bool` |  |  |  |
| `spec.webp` | `bool` |  |  |  |
| `spec.websockets` | `bool` |  |  |  |
| `spec.longLivedGrpc` | `bool` |  |  |  |

## Field Details

### spec.zoneId

`string | valueFrom` · required

The zone whose settings are managed.
When using value_from, defaults to CloudflareDnsZone kind and status.outputs.zone_id field path.

- references: CloudflareDnsZone (`status.outputs.zone_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: CloudflareDnsZone, name: <that resource's name>, fieldPath: status.outputs.zone_id}} -- a bare string does not parse

### spec.managedRequestHeaders

`[]CloudflareZoneSettingsManagedHeader`

Managed request headers to toggle (Cloudflare-defined transforms that add headers
to requests before they reach the origin, for example add_true_client_ip_headers
or add_visitor_location_headers). Each row names a Cloudflare-managed transform id
and whether it is enabled. Removing a row stops managing that transform; Cloudflare
keeps its last state.

### spec.managedRequestHeaders[].id

`string` · required

The Cloudflare-defined transform id, for example add_true_client_ip_headers,
add_visitor_location_headers, remove_x-powered-by_header. The API rejects
unknown ids; Cloudflare's documentation lists the available transforms.

- rule: {"required":true}

### spec.managedRequestHeaders[].enabled

`bool`

Whether the managed transform is enabled. Both fields are required by the
API on every row.

### spec.managedResponseHeaders

`[]CloudflareZoneSettingsManagedHeader`

Managed response headers to toggle (Cloudflare-defined transforms that add headers
to responses sent to clients, for example remove_x-powered-by_header or
add_security_headers). Same semantics as managed_request_headers.

### spec.managedResponseHeaders[].id

`string` · required

The Cloudflare-defined transform id, for example add_true_client_ip_headers,
add_visitor_location_headers, remove_x-powered-by_header. The API rejects
unknown ids; Cloudflare's documentation lists the available transforms.

- rule: {"required":true}

### spec.managedResponseHeaders[].enabled

`bool`

Whether the managed transform is enabled. Both fields are required by the
API on every row.

### spec.urlNormalization

`CloudflareZoneSettingsUrlNormalization`

URL normalization: how Cloudflare rewrites URLs before rules match them and
before they reach the origin. Unset means not managed.

### spec.urlNormalization.scope

`string` · required

Which URLs are normalized: incoming normalizes URLs before rules match them,
both also normalizes the URL sent to the origin, none disables normalization.

- rule: {"required":true,"string":{"in":["incoming","both","none"]}}

### spec.urlNormalization.type

`string` · required

The normalization standard: cloudflare applies Cloudflare's normalization,
rfc3986 applies RFC 3986 normalization only.

- rule: {"required":true,"string":{"in":["cloudflare","rfc3986"]}}

### spec.originCloudRegions

`[]CloudflareZoneSettingsOriginCloudRegion`

Origin cloud-region hints: per-origin-IP declarations of which cloud vendor and
region an origin lives in, used by Cloudflare to optimize origin-facing routing.
Each row is keyed by origin_ip (changing the IP replaces the row's API object).

### spec.originCloudRegions[].originIp

`string` · required

The origin server's IP address this hint describes. Changing the IP replaces
the API object (the IP is the row's identity).

- rule: {"required":true}

### spec.originCloudRegions[].region

`string` · required

The vendor-specific region name the origin lives in (for example us-east-1
for aws). Free-form; Cloudflare validates it against the vendor's region set.

- rule: {"required":true}

### spec.originCloudRegions[].vendor

`string` · required

The cloud vendor hosting the origin.

- rule: {"required":true,"string":{"in":["aws","azure","gcp","oci"]}}

### spec.waitingRoomCrawlerBypass

`bool` · optional (explicit presence)

Allow verified search-engine crawlers to bypass the waiting room on this zone.
This is the zone-wide waiting-room setting (one boolean); per-room configuration
lives on the waiting room itself. Cloudflare's default is false.

### spec.zeroRtt

`bool` · optional (explicit presence)

0-RTT session resumption (setting_id "0rtt"): lets returning TLS 1.3 visitors
send application data in the first round trip. Slightly faster repeat visits;
replay-sensitive endpoints should keep it off.

### spec.advancedDdos

`bool` · optional (explicit presence)

Advanced DDoS protection (setting_id "advanced_ddos"). Enterprise-plan zones
only -- the API reports this setting as not editable on lower plans.

### spec.aegis

`CloudflareZoneSettingsAegis`

Aegis dedicated egress IPs: pin the Cloudflare-to-origin traffic for this zone
to a dedicated egress IP pool, so origin firewalls can allowlist stable
addresses. Requires an Aegis entitlement and a provisioned pool.

### spec.aegis.enabled

`bool` · optional (explicit presence)

Whether Aegis egress is enabled for the zone.

### spec.aegis.poolId

`string`

The Aegis IP pool to egress from. Pools are provisioned by Cloudflare under
the Aegis entitlement; the id comes from the Aegis dashboard or API.

### spec.alwaysOnline

`bool` · optional (explicit presence)

Always Online: serve limited cached copies of pages from the Internet Archive
when the origin is unreachable.

### spec.alwaysUseHttps

`bool` · optional (explicit presence)

Always Use HTTPS: answer every plain-http request with a 301 redirect to the
https equivalent.

### spec.automaticHttpsRewrites

`bool` · optional (explicit presence)

Automatic HTTPS Rewrites: rewrite http:// references inside served HTML to
https:// when the target is known to support it, fixing mixed content.

### spec.automaticPlatformOptimization

`CloudflareZoneSettingsAutomaticPlatformOptimization`

Automatic Platform Optimization for WordPress: serve the whole site from
Cloudflare's edge cache with the APO WordPress plugin invalidating on change.
The API requires every field of this object on writes.

### spec.automaticPlatformOptimization.enabled

`bool` · required · optional (explicit presence)

Whether APO serves this zone from the edge cache. Required inside
automatic_platform_optimization, as every member of this object is on
writes.

- rule: {"required":true}

### spec.automaticPlatformOptimization.cacheByDeviceType

`bool`

Cache separately by device type (mobile, tablet, desktop).

### spec.automaticPlatformOptimization.cf

`bool`

Whether the zone is proxied through Cloudflare (the API expects the current
proxy state; true for orange-clouded zones).

### spec.automaticPlatformOptimization.hostnames

`[]string` · required

The WordPress hostnames APO serves, for example example.com and
www.example.com.

- rule: {"repeated":{"minItems":"1"}}

### spec.automaticPlatformOptimization.wordpress

`bool`

Whether the origin runs WordPress.

### spec.automaticPlatformOptimization.wpPlugin

`bool`

Whether the Cloudflare APO WordPress plugin is installed (the plugin purges
the edge cache on content changes).

### spec.brotli

`bool` · optional (explicit presence)

Brotli compression for responses to clients that accept it.

### spec.browserCacheTtl

`int64` · optional (explicit presence)

Browser Cache TTL in seconds: how long browsers may cache Cloudflare-served
resources. The API accepts a fixed value set (0 and specific durations from 30
seconds to one year); values outside the set are rejected by Cloudflare, and 0
means "respect existing headers".

- rule: {"int64":{"gte":"0"}}

### spec.browserCheck

`bool` · optional (explicit presence)

Browser Integrity Check: block requests whose HTTP headers are commonly abused
by spammers and bots.

### spec.cacheLevel

`string` · optional (explicit presence)

Cache Level: basic caches static content ignoring query strings, simplified
ignores query strings entirely, aggressive caches static content with all
query-string variants.

- rule: {"string":{"in":["aggressive","basic","simplified"]}}

### spec.challengeTtl

`int64` · optional (explicit presence)

Challenge TTL in seconds: how long a visitor who passed a challenge is allowed
in before being re-challenged. The API accepts a fixed value set (300 through
31536000); out-of-set values are rejected by Cloudflare.

- rule: {"int64":{"gte":"0"}}

### spec.ciphers

`[]string`

TLS cipher allowlist for this zone's HTTPS termination, in BoringSSL format
(for example ECDHE-ECDSA-AES128-GCM-SHA256). An empty list is "not managed";
to reset the zone to Cloudflare defaults, apply the API's documented reset
(an empty value) manually -- the module never sends an empty list.
ENTITLEMENT-GATED WRITE: the zone must carry an Advanced Certificate
Manager subscription or the API rejects the write with 400 code 1023
(measured at v5.23.0) -- even though the settings API reports the id
editable=true on every plan.

### spec.cnameFlattening

`string` · optional (explicit presence)

CNAME flattening behavior at the zone apex and beyond: flatten_at_root
flattens only the root record, flatten_all flattens every CNAME.

- rule: {"string":{"in":["flatten_at_root","flatten_all"]}}

### spec.contentConverter

`bool` · optional (explicit presence)

Content converter (setting_id "content_converter"): when a client sends an
Accept header requesting text/markdown for an HTML page, convert the response
to Markdown at the edge.

### spec.developmentMode

`bool` · optional (explicit presence)

Development Mode: temporarily bypass Cloudflare's edge cache so origin changes
are visible immediately. Cloudflare auto-disables it after three hours; a
managed "on" re-enables it on the next apply.

### spec.earlyHints

`bool` · optional (explicit presence)

Early Hints: send HTTP 103 responses with Link headers from cached 200s while
the origin is still thinking.

### spec.edgeCacheTtl

`int64` · optional (explicit presence)

Edge Cache TTL in seconds: how long Cloudflare's edge keeps a resource before
revalidating. The API accepts a fixed value set (30 seconds through 31 days on
Enterprise; shorter maxima on lower plans).

- rule: {"int64":{"gte":"0"}}

### spec.emailObfuscation

`bool` · optional (explicit presence)

Email Obfuscation: hide email addresses on pages from bots while keeping them
visible to humans.

### spec.h2Prioritization

`string` · optional (explicit presence)

HTTP/2 Edge Prioritization: custom enables the enhanced scheduler that
reprioritizes resources based on browser signals.

- rule: {"string":{"in":["on","off","custom"]}}

### spec.hotlinkProtection

`bool` · optional (explicit presence)

Hotlink Protection: block other sites from embedding this zone's images.

### spec.http2

`bool` · optional (explicit presence)

HTTP/2 support for client connections.

### spec.http3

`bool` · optional (explicit presence)

HTTP/3 (QUIC) support for client connections.

### spec.imageResizing

`string` · optional (explicit presence)

Image Transformations for this zone: on restricts transforms to this zone's
own URLs, open also allows transforming remote images fetched by URL.

- rule: {"string":{"in":["on","off","open"]}}

### spec.ipGeolocation

`bool` · optional (explicit presence)

IP Geolocation: add the CF-IPCountry header with the visitor's country to
origin-bound requests.

### spec.ipv6

`bool` · optional (explicit presence)

IPv6 termination on all Cloudflare-proxied hostnames of the zone.

### spec.maxUpload

`int64` · optional (explicit presence)

Maximum upload size in megabytes for requests through this zone. The API
accepts a fixed value set (100 through 1000 depending on plan).

- rule: {"int64":{"gte":"0"}}

### spec.minTlsVersion

`string` · optional (explicit presence)

Minimum TLS version accepted for HTTPS connections to this zone.

- rule: {"string":{"in":["1.0","1.1","1.2","1.3"]}}

### spec.mirage

`bool` · optional (explicit presence)

Mirage: optimize image delivery for slow mobile connections (lazy loading,
low-resolution placeholders).

### spec.nel

`CloudflareZoneSettingsNel`

Network Error Logging: have browsers report connectivity failures for this
zone to Cloudflare.

### spec.nel.enabled

`bool` · required · optional (explicit presence)

Whether browsers report network errors for this zone to Cloudflare.
Required inside nel.

- rule: {"required":true}

### spec.opportunisticEncryption

`bool` · optional (explicit presence)

Opportunistic Encryption: advertise TLS capability to browsers for
plain-http resources via Alt-Svc.

### spec.opportunisticOnion

`bool` · optional (explicit presence)

Opportunistic Onion: serve Tor visitors over Cloudflare's onion services via
an Alt-Svc header, avoiding Tor exit-node challenges.

### spec.orangeToOrange

`bool` · optional (explicit presence)

Orange-to-Orange (O2O): allow this zone to CNAME to another Cloudflare-proxied
zone and layer both zones' features. Enterprise feature; the API reports it as
not editable on lower plans.

### spec.originErrorPagePassThru

`bool` · optional (explicit presence)

Origin error page pass-through: hand origin 502/504 error pages to visitors
untouched instead of showing Cloudflare's error page.

### spec.originH2MaxStreams

`int64` · optional (explicit presence)

Origin H2 Max Streams: the maximum concurrent requests Cloudflare multiplexes
onto one HTTP/2 origin connection (1 through 200; API-validated).

- rule: {"int64":{"gte":"0"}}

### spec.originMaxHttpVersion

`string` · optional (explicit presence)

Origin Max HTTP Version: the highest HTTP version Cloudflare speaks to the
origin ("2" or "1").

- rule: {"string":{"in":["1","2"]}}

### spec.polish

`string` · optional (explicit presence)

Polish: strip image metadata and recompress. lossless keeps pixel data intact;
lossy also recompresses JPEGs for further savings.

- rule: {"string":{"in":["off","lossless","lossy"]}}

### spec.prefetchPreload

`bool` · optional (explicit presence)

Prefetch URLs from response headers: fetch URLs named in Link prefetch headers
into cache before visitors ask for them. Enterprise feature.

### spec.privacyPass

`bool` · optional (explicit presence)

Privacy Pass v1 compatibility (legacy browser extension support). Cloudflare
has deprecated the product; the setting remains writable.

### spec.proxyReadTimeout

`int64` · optional (explicit presence)

Proxy Read Timeout in seconds: how long Cloudflare waits between reads from
the origin before serving a 524 (1 through 6000; API-validated; Enterprise).

- rule: {"int64":{"gte":"0"}}

### spec.pseudoIpv4

`string` · optional (explicit presence)

Pseudo IPv4: how to represent IPv6 visitors to IPv4-only origins --
add_header adds Cf-Pseudo-IPv4, overwrite_header overwrites the visitor IP
headers with the pseudo address.

- rule: {"string":{"in":["off","add_header","overwrite_header"]}}

### spec.redirectsForAiTraining

`bool` · optional (explicit presence)

Redirect verified AI training crawlers to canonical content locations.

### spec.replaceInsecureJs

`bool` · optional (explicit presence)

Replace insecure JavaScript libraries with safer, faster alternatives at the
edge where a compatible replacement exists.

### spec.responseBuffering

`bool` · optional (explicit presence)

Response Buffering: buffer the full origin response at the edge before sending
it to the visitor, instead of streaming as it arrives. Enterprise feature.

### spec.rocketLoader

`bool` · optional (explicit presence)

Rocket Loader: defer JavaScript loading until after render for faster paint
times. Test carefully -- it rewrites how scripts execute.

### spec.searchForAgents

`bool` · optional (explicit presence)

Search for agents (setting_id "search_for_agents"): provision an AI Search
instance for the zone and expose site search endpoints to AI agents.

### spec.securityHeader

`CloudflareZoneSettingsSecurityHeader`

HTTP Strict Transport Security and related security headers for the zone.

### spec.securityHeader.enabled

`bool` · required · optional (explicit presence)

Whether Strict-Transport-Security is emitted. Required inside
security_header: the block is written to Cloudflare whole, so it must say
whether HSTS is on before the other dials mean anything.

- rule: {"required":true}

### spec.securityHeader.includeSubdomains

`bool`

Apply HSTS to all subdomains. Only enable when every subdomain serves HTTPS --
browsers will refuse plain http on all of them for max_age seconds.

### spec.securityHeader.maxAge

`int64`

The Strict-Transport-Security max-age in seconds. Common production value is
31536000 (one year). 0 tells browsers to drop the pin.

- rule: {"int64":{"gte":"0"}}

### spec.securityHeader.nosniff

`bool`

Emit X-Content-Type-Options: nosniff alongside HSTS.

### spec.securityHeader.preload

`bool`

Add the preload directive, requesting inclusion in browser preload lists.
Preload-list removal is slow and manual -- enable deliberately.

### spec.securityLevel

`string` · optional (explicit presence)

Security Level: how aggressively Cloudflare challenges suspicious visitors.
under_attack shows an interstitial challenge to everyone.

- rule: {"string":{"in":["off","essentially_off","low","medium","high","under_attack"]}}

### spec.serverSideExclude

`bool` · optional (explicit presence)

Server Side Excludes: strip content between <!--sse--> tags from pages served
to suspicious visitors.

### spec.sha1Support

`bool` · optional (explicit presence)

SHA1 certificate support for legacy clients.

### spec.sortQueryStringForCache

`bool` · optional (explicit presence)

Sort query strings before cache lookup, so ?a=1&b=2 and ?b=2&a=1 share one
cache entry. Enterprise feature.

### spec.ssl

`string` · optional (explicit presence)

SSL/TLS encryption mode between Cloudflare and the origin: off disables HTTPS
entirely, flexible terminates TLS at the edge and speaks plain http to the
origin, full encrypts to the origin without validating its certificate, strict
also validates the origin certificate. Production zones should use strict.

- rule: {"string":{"in":["off","flexible","full","strict"]}}

### spec.sslRecommender

`bool` · optional (explicit presence)

SSL/TLS Recommender: have Cloudflare probe the origin and email
recommendations for the strongest workable ssl mode. The provider's schema
requires the on/off value form on writes (its documented enabled-attribute
form fails validation at v5.23.0), so the module sends value on/off.

### spec.tls12Only

`bool` · optional (explicit presence)

Accept only TLS 1.2 connections (legacy zone setting; prefer min_tls_version).

### spec.tls13

`string` · optional (explicit presence)

TLS 1.3 support: zrt enables TLS 1.3 with Cloudflare's zero round-trip
resumption.

- rule: {"string":{"in":["on","off","zrt"]}}

### spec.tlsClientAuth

`bool` · optional (explicit presence)

TLS Client Auth: require client certificates on connections to the origin
(Cloudflare presents a certificate the origin validates).

### spec.transformations

`string` · optional (explicit presence)

Media Transformations for this zone: on restricts transforms to this zone's
own media URLs, open also allows transforming remote media fetched by URL.

- rule: {"string":{"in":["on","off","open"]}}

### spec.transformationsAllowedOrigins

`string` · optional (explicit presence)

Media Transformations allowed origins: a space-separated list of origins
permitted to request transformations, restricting who may use this zone's
transformation endpoints.

### spec.trueClientIpHeader

`bool` · optional (explicit presence)

True-Client-IP header: send the visitor's IP in the True-Client-IP header to
the origin (Akamai-compatible). Enterprise feature.

### spec.waf

`bool` · optional (explicit presence)

Legacy WAF (setting_id "waf"): the previous-generation WAF toggle. New WAF
configuration belongs to CloudflareRuleset phases; this setting remains for
zones still on the legacy WAF.

### spec.webp

`bool` · optional (explicit presence)

WebP conversion: serve WebP variants of cached images to clients that accept
them (effective when Polish is enabled).

### spec.websockets

`bool` · optional (explicit presence)

WebSockets support for proxied connections on this zone.

### spec.longLivedGrpc

`bool` · optional (explicit presence)

Long-lived gRPC connections (setting_id "long_lived_grpc"): keep gRPC streams
on this zone open beyond standard proxy read timeouts. Undocumented in the
API's setting table at provider v5.23.0 but accepted by the settings endpoint
(the provider's own regression tests exercise it).

## Validation Rules

- `spec.at_least_one_setting`: configure at least one zone setting -- a CloudflareZoneSettings resource that manages nothing would deploy nothing

## Outputs

Reference an output from another manifest as `valueFrom: {kind: CloudflareZoneSettings, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.zone_id` | `string` | The zone ID the settings belong to (the singleton's identity, and the pass-through for downstream resource references). |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.zoneId` | CloudflareDnsZone | `status.outputs.zone_id` |

## See Also

- [Overview](../README.md)
