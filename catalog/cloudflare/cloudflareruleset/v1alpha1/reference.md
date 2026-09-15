# CloudflareRuleset

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `cloudflare.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

CloudflareRulesetSpec defines the configuration for a Cloudflare Ruleset.

Rulesets are ordered collections of rules that execute during specific phases of HTTP
request processing on the Cloudflare network. They power Origin Rules, WAF Custom Rules,
Cache Rules, Redirect Rules, Transform Rules, Rate Limiting, Configuration Rules, and more.

Exactly one of zone_id or account_id must be provided. Zone-level rulesets apply to a
single domain; account-level rulesets apply across all zones in the account.

## Example

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareRuleset
metadata:
  name: test-origin-rule
spec:
  zone_id:
    value: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  ruleset_kind: zone
  phase: http_request_origin
  name: "Test origin routing"
  rules:
    - ref: "route-app-to-k8s"
      expression: 'not starts_with(http.request.uri.path, "/docs")'
      action: route
      description: "Route non-docs traffic to K8s"
      enabled: true
      action_parameters:
        host_header: "example.com"
        origin:
          host: "k8s-lb.example.com"
          port: 443
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.zoneId` | `string \| valueFrom` |  |  | CloudflareDnsZone (`status.outputs.zone_id`) |
| `spec.accountId` | `string` |  |  |  |
| `spec.rulesetKind` | `enum` |  | `zone` |  |
| `spec.phase` | `enum` | yes |  |  |
| `spec.name` | `string` | yes |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.rules` | `[]CloudflareRulesetRule` | yes |  |  |
| `spec.rules[].ref` | `string` |  |  |  |
| `spec.rules[].expression` | `string` | yes |  |  |
| `spec.rules[].action` | `enum` | yes |  |  |
| `spec.rules[].description` | `string` |  |  |  |
| `spec.rules[].enabled` | `bool` |  | `true` |  |
| `spec.rules[].actionParameters` | `CloudflareRulesetActionParameters` |  |  |  |
| `spec.rules[].actionParameters.hostHeader` | `string` |  |  |  |
| `spec.rules[].actionParameters.origin` | `CloudflareRulesetOrigin` |  |  |  |
| `spec.rules[].actionParameters.origin.host` | `string` |  |  |  |
| `spec.rules[].actionParameters.origin.port` | `int32` |  |  |  |
| `spec.rules[].actionParameters.sni` | `CloudflareRulesetSni` |  |  |  |
| `spec.rules[].actionParameters.sni.value` | `string` |  |  |  |
| `spec.rules[].actionParameters.response` | `CloudflareRulesetResponse` |  |  |  |
| `spec.rules[].actionParameters.response.statusCode` | `int32` |  |  |  |
| `spec.rules[].actionParameters.response.content` | `string` |  |  |  |
| `spec.rules[].actionParameters.response.contentType` | `string` |  |  |  |
| `spec.rules[].actionParameters.uri` | `CloudflareRulesetUri` |  |  |  |
| `spec.rules[].actionParameters.uri.path` | `CloudflareRulesetUriComponent` |  |  |  |
| `spec.rules[].actionParameters.uri.path.value` | `string` |  |  |  |
| `spec.rules[].actionParameters.uri.path.expression` | `string` |  |  |  |
| `spec.rules[].actionParameters.uri.query` | `CloudflareRulesetUriComponent` |  |  |  |
| `spec.rules[].actionParameters.uri.query.value` | `string` |  |  |  |
| `spec.rules[].actionParameters.uri.query.expression` | `string` |  |  |  |
| `spec.rules[].actionParameters.headers` | `map<string, CloudflareRulesetHeader>` |  |  |  |
| `spec.rules[].actionParameters.headers.*.operation` | `string` |  |  |  |
| `spec.rules[].actionParameters.headers.*.value` | `string` |  |  |  |
| `spec.rules[].actionParameters.headers.*.expression` | `string` |  |  |  |
| `spec.rules[].actionParameters.fromValue` | `CloudflareRulesetFromValue` |  |  |  |
| `spec.rules[].actionParameters.fromValue.targetUrl` | `CloudflareRulesetTargetUrl` |  |  |  |
| `spec.rules[].actionParameters.fromValue.targetUrl.value` | `string` |  |  |  |
| `spec.rules[].actionParameters.fromValue.targetUrl.expression` | `string` |  |  |  |
| `spec.rules[].actionParameters.fromValue.statusCode` | `int32` |  |  |  |
| `spec.rules[].actionParameters.fromValue.preserveQueryString` | `bool` |  |  |  |
| `spec.rules[].actionParameters.phases` | `[]string` |  |  |  |
| `spec.rules[].actionParameters.products` | `[]string` |  |  |  |
| `spec.rules[].actionParameters.ruleset` | `string` |  |  |  |
| `spec.rules[].actionParameters.rulesets` | `[]string` |  |  |  |
| `spec.rules[].actionParameters.rules` | `map<string, CloudflareRulesetStringList>` |  |  |  |
| `spec.rules[].actionParameters.rules.*.values` | `[]string` |  |  |  |
| `spec.rules[].actionParameters.id` | `string` |  |  |  |
| `spec.rules[].actionParameters.overrides` | `CloudflareRulesetOverrides` |  |  |  |
| `spec.rules[].actionParameters.overrides.action` | `string` |  |  |  |
| `spec.rules[].actionParameters.overrides.enabled` | `bool` |  |  |  |
| `spec.rules[].actionParameters.overrides.categories` | `[]CloudflareRulesetCategoryOverride` |  |  |  |
| `spec.rules[].actionParameters.overrides.categories[].category` | `string` | yes |  |  |
| `spec.rules[].actionParameters.overrides.categories[].action` | `string` |  |  |  |
| `spec.rules[].actionParameters.overrides.categories[].enabled` | `bool` |  |  |  |
| `spec.rules[].actionParameters.overrides.categories[].sensitivityLevel` | `string` |  |  |  |
| `spec.rules[].actionParameters.overrides.rules` | `[]CloudflareRulesetRuleOverride` |  |  |  |
| `spec.rules[].actionParameters.overrides.rules[].id` | `string` | yes |  |  |
| `spec.rules[].actionParameters.overrides.rules[].action` | `string` |  |  |  |
| `spec.rules[].actionParameters.overrides.rules[].enabled` | `bool` |  |  |  |
| `spec.rules[].actionParameters.overrides.rules[].scoreThreshold` | `int32` |  |  |  |
| `spec.rules[].actionParameters.overrides.rules[].sensitivityLevel` | `string` |  |  |  |
| `spec.rules[].actionParameters.overrides.sensitivityLevel` | `string` |  |  |  |
| `spec.rules[].actionParameters.cache` | `bool` |  |  |  |
| `spec.rules[].actionParameters.edgeTtl` | `CloudflareRulesetEdgeTtl` |  |  |  |
| `spec.rules[].actionParameters.edgeTtl.mode` | `string` |  |  |  |
| `spec.rules[].actionParameters.edgeTtl.defaultTtl` | `int32` |  |  |  |
| `spec.rules[].actionParameters.edgeTtl.statusCodeTtls` | `[]CloudflareRulesetStatusCodeTtl` |  |  |  |
| `spec.rules[].actionParameters.edgeTtl.statusCodeTtls[].value` | `int32` |  |  |  |
| `spec.rules[].actionParameters.edgeTtl.statusCodeTtls[].statusCode` | `int32` |  |  |  |
| `spec.rules[].actionParameters.edgeTtl.statusCodeTtls[].statusCodeRange` | `CloudflareRulesetStatusCodeRange` |  |  |  |
| `spec.rules[].actionParameters.edgeTtl.statusCodeTtls[].statusCodeRange.from` | `int32` |  |  |  |
| `spec.rules[].actionParameters.edgeTtl.statusCodeTtls[].statusCodeRange.to` | `int32` |  |  |  |
| `spec.rules[].actionParameters.browserTtl` | `CloudflareRulesetBrowserTtl` |  |  |  |
| `spec.rules[].actionParameters.browserTtl.mode` | `string` |  |  |  |
| `spec.rules[].actionParameters.browserTtl.defaultTtl` | `int32` |  |  |  |
| `spec.rules[].actionParameters.serveStale` | `CloudflareRulesetServeStale` |  |  |  |
| `spec.rules[].actionParameters.serveStale.disableStaleWhileUpdating` | `bool` |  |  |  |
| `spec.rules[].actionParameters.fromList` | `CloudflareRulesetFromList` |  |  |  |
| `spec.rules[].actionParameters.fromList.name` | `string \| valueFrom` | yes |  | CloudflareList (`status.outputs.name`) |
| `spec.rules[].actionParameters.fromList.key` | `string` | yes |  |  |
| `spec.rules[].actionParameters.algorithms` | `[]CloudflareRulesetAlgorithm` |  |  |  |
| `spec.rules[].actionParameters.algorithms[].name` | `string` | yes |  |  |
| `spec.rules[].actionParameters.matchedData` | `CloudflareRulesetMatchedData` |  |  |  |
| `spec.rules[].actionParameters.matchedData.publicKey` | `string` | yes |  |  |
| `spec.rules[].actionParameters.increment` | `int64` |  |  |  |
| `spec.rules[].actionParameters.assetName` | `string` |  |  |  |
| `spec.rules[].actionParameters.content` | `string` |  |  |  |
| `spec.rules[].actionParameters.contentType` | `string` |  |  |  |
| `spec.rules[].actionParameters.statusCode` | `int64` |  |  |  |
| `spec.rules[].actionParameters.automaticHttpsRewrites` | `bool` |  |  |  |
| `spec.rules[].actionParameters.autominify` | `CloudflareRulesetAutominify` |  |  |  |
| `spec.rules[].actionParameters.autominify.css` | `bool` |  |  |  |
| `spec.rules[].actionParameters.autominify.html` | `bool` |  |  |  |
| `spec.rules[].actionParameters.autominify.js` | `bool` |  |  |  |
| `spec.rules[].actionParameters.bic` | `bool` |  |  |  |
| `spec.rules[].actionParameters.contentConverter` | `bool` |  |  |  |
| `spec.rules[].actionParameters.disableApps` | `bool` |  |  |  |
| `spec.rules[].actionParameters.disableRum` | `bool` |  |  |  |
| `spec.rules[].actionParameters.disableZaraz` | `bool` |  |  |  |
| `spec.rules[].actionParameters.emailObfuscation` | `bool` |  |  |  |
| `spec.rules[].actionParameters.fonts` | `bool` |  |  |  |
| `spec.rules[].actionParameters.hotlinkProtection` | `bool` |  |  |  |
| `spec.rules[].actionParameters.mirage` | `bool` |  |  |  |
| `spec.rules[].actionParameters.opportunisticEncryption` | `bool` |  |  |  |
| `spec.rules[].actionParameters.polish` | `string` |  |  |  |
| `spec.rules[].actionParameters.redirectsForAiTraining` | `bool` |  |  |  |
| `spec.rules[].actionParameters.requestBodyBuffering` | `string` |  |  |  |
| `spec.rules[].actionParameters.responseBodyBuffering` | `string` |  |  |  |
| `spec.rules[].actionParameters.rocketLoader` | `bool` |  |  |  |
| `spec.rules[].actionParameters.securityLevel` | `string` |  |  |  |
| `spec.rules[].actionParameters.serverSideExcludes` | `bool` |  |  |  |
| `spec.rules[].actionParameters.ssl` | `string` |  |  |  |
| `spec.rules[].actionParameters.sxg` | `bool` |  |  |  |
| `spec.rules[].actionParameters.additionalCacheablePorts` | `[]int64` |  |  |  |
| `spec.rules[].actionParameters.cacheKey` | `CloudflareRulesetCacheKey` |  |  |  |
| `spec.rules[].actionParameters.cacheKey.cacheByDeviceType` | `bool` |  |  |  |
| `spec.rules[].actionParameters.cacheKey.cacheDeceptionArmor` | `bool` |  |  |  |
| `spec.rules[].actionParameters.cacheKey.customKey` | `CloudflareRulesetCacheKeyCustomKey` |  |  |  |
| `spec.rules[].actionParameters.cacheKey.customKey.cookie` | `CloudflareRulesetCacheKeyCookie` |  |  |  |
| `spec.rules[].actionParameters.cacheKey.customKey.cookie.checkPresence` | `[]string` |  |  |  |
| `spec.rules[].actionParameters.cacheKey.customKey.cookie.include` | `[]string` |  |  |  |
| `spec.rules[].actionParameters.cacheKey.customKey.header` | `CloudflareRulesetCacheKeyHeader` |  |  |  |
| `spec.rules[].actionParameters.cacheKey.customKey.header.checkPresence` | `[]string` |  |  |  |
| `spec.rules[].actionParameters.cacheKey.customKey.header.contains` | `map<string, CloudflareRulesetStringList>` |  |  |  |
| `spec.rules[].actionParameters.cacheKey.customKey.header.contains.*.values` | `[]string` |  |  |  |
| `spec.rules[].actionParameters.cacheKey.customKey.header.excludeOrigin` | `bool` |  |  |  |
| `spec.rules[].actionParameters.cacheKey.customKey.header.include` | `[]string` |  |  |  |
| `spec.rules[].actionParameters.cacheKey.customKey.host` | `CloudflareRulesetCacheKeyHost` |  |  |  |
| `spec.rules[].actionParameters.cacheKey.customKey.host.resolved` | `bool` |  |  |  |
| `spec.rules[].actionParameters.cacheKey.customKey.queryString` | `CloudflareRulesetCacheKeyQueryString` |  |  |  |
| `spec.rules[].actionParameters.cacheKey.customKey.queryString.include` | `CloudflareRulesetCacheKeyQueryStringFilter` |  |  |  |
| `spec.rules[].actionParameters.cacheKey.customKey.queryString.include.list` | `[]string` |  |  |  |
| `spec.rules[].actionParameters.cacheKey.customKey.queryString.include.all` | `bool` |  |  |  |
| `spec.rules[].actionParameters.cacheKey.customKey.queryString.exclude` | `CloudflareRulesetCacheKeyQueryStringFilter` |  |  |  |
| `spec.rules[].actionParameters.cacheKey.customKey.queryString.exclude.list` | `[]string` |  |  |  |
| `spec.rules[].actionParameters.cacheKey.customKey.queryString.exclude.all` | `bool` |  |  |  |
| `spec.rules[].actionParameters.cacheKey.customKey.user` | `CloudflareRulesetCacheKeyUser` |  |  |  |
| `spec.rules[].actionParameters.cacheKey.customKey.user.deviceType` | `bool` |  |  |  |
| `spec.rules[].actionParameters.cacheKey.customKey.user.geo` | `bool` |  |  |  |
| `spec.rules[].actionParameters.cacheKey.customKey.user.lang` | `bool` |  |  |  |
| `spec.rules[].actionParameters.cacheKey.ignoreQueryStringsOrder` | `bool` |  |  |  |
| `spec.rules[].actionParameters.cacheReserve` | `CloudflareRulesetCacheReserve` |  |  |  |
| `spec.rules[].actionParameters.cacheReserve.eligible` | `bool` |  |  |  |
| `spec.rules[].actionParameters.cacheReserve.minimumFileSize` | `int64` |  |  |  |
| `spec.rules[].actionParameters.originCacheControl` | `bool` |  |  |  |
| `spec.rules[].actionParameters.originErrorPagePassthru` | `bool` |  |  |  |
| `spec.rules[].actionParameters.readTimeout` | `int64` |  |  |  |
| `spec.rules[].actionParameters.respectStrongEtags` | `bool` |  |  |  |
| `spec.rules[].actionParameters.vary` | `CloudflareRulesetVary` |  |  |  |
| `spec.rules[].actionParameters.vary.default` | `CloudflareRulesetVaryDefault` |  |  |  |
| `spec.rules[].actionParameters.vary.default.action` | `string` | yes |  |  |
| `spec.rules[].actionParameters.vary.headers` | `map<string, CloudflareRulesetVaryHeader>` |  |  |  |
| `spec.rules[].actionParameters.vary.headers.*.action` | `string` | yes |  |  |
| `spec.rules[].actionParameters.vary.headers.*.mediaTypes` | `[]string` |  |  |  |
| `spec.rules[].actionParameters.vary.headers.*.languages` | `[]string` |  |  |  |
| `spec.rules[].actionParameters.stripEtags` | `bool` |  |  |  |
| `spec.rules[].actionParameters.stripLastModified` | `bool` |  |  |  |
| `spec.rules[].actionParameters.stripSetCookie` | `bool` |  |  |  |
| `spec.rules[].actionParameters.cookieFields` | `[]CloudflareRulesetLogField` |  |  |  |
| `spec.rules[].actionParameters.cookieFields[].name` | `string` | yes |  |  |
| `spec.rules[].actionParameters.rawResponseFields` | `[]CloudflareRulesetLogResponseField` |  |  |  |
| `spec.rules[].actionParameters.rawResponseFields[].name` | `string` | yes |  |  |
| `spec.rules[].actionParameters.rawResponseFields[].preserveDuplicates` | `bool` |  |  |  |
| `spec.rules[].actionParameters.requestFields` | `[]CloudflareRulesetLogField` |  |  |  |
| `spec.rules[].actionParameters.requestFields[].name` | `string` | yes |  |  |
| `spec.rules[].actionParameters.responseFields` | `[]CloudflareRulesetLogResponseField` |  |  |  |
| `spec.rules[].actionParameters.responseFields[].name` | `string` | yes |  |  |
| `spec.rules[].actionParameters.responseFields[].preserveDuplicates` | `bool` |  |  |  |
| `spec.rules[].actionParameters.transformedRequestFields` | `[]CloudflareRulesetLogField` |  |  |  |
| `spec.rules[].actionParameters.transformedRequestFields[].name` | `string` | yes |  |  |
| `spec.rules[].actionParameters.maxAge` | `CloudflareRulesetCacheControlValue` |  |  |  |
| `spec.rules[].actionParameters.maxAge.operation` | `string` |  |  |  |
| `spec.rules[].actionParameters.maxAge.value` | `int64` |  |  |  |
| `spec.rules[].actionParameters.maxAge.cloudflareOnly` | `bool` |  |  |  |
| `spec.rules[].actionParameters.sMaxage` | `CloudflareRulesetCacheControlValue` |  |  |  |
| `spec.rules[].actionParameters.sMaxage.operation` | `string` |  |  |  |
| `spec.rules[].actionParameters.sMaxage.value` | `int64` |  |  |  |
| `spec.rules[].actionParameters.sMaxage.cloudflareOnly` | `bool` |  |  |  |
| `spec.rules[].actionParameters.staleWhileRevalidate` | `CloudflareRulesetCacheControlValue` |  |  |  |
| `spec.rules[].actionParameters.staleWhileRevalidate.operation` | `string` |  |  |  |
| `spec.rules[].actionParameters.staleWhileRevalidate.value` | `int64` |  |  |  |
| `spec.rules[].actionParameters.staleWhileRevalidate.cloudflareOnly` | `bool` |  |  |  |
| `spec.rules[].actionParameters.staleIfError` | `CloudflareRulesetCacheControlValue` |  |  |  |
| `spec.rules[].actionParameters.staleIfError.operation` | `string` |  |  |  |
| `spec.rules[].actionParameters.staleIfError.value` | `int64` |  |  |  |
| `spec.rules[].actionParameters.staleIfError.cloudflareOnly` | `bool` |  |  |  |
| `spec.rules[].actionParameters.private` | `CloudflareRulesetCacheControlQualifiers` |  |  |  |
| `spec.rules[].actionParameters.private.operation` | `string` |  |  |  |
| `spec.rules[].actionParameters.private.qualifiers` | `[]string` |  |  |  |
| `spec.rules[].actionParameters.private.cloudflareOnly` | `bool` |  |  |  |
| `spec.rules[].actionParameters.noCache` | `CloudflareRulesetCacheControlQualifiers` |  |  |  |
| `spec.rules[].actionParameters.noCache.operation` | `string` |  |  |  |
| `spec.rules[].actionParameters.noCache.qualifiers` | `[]string` |  |  |  |
| `spec.rules[].actionParameters.noCache.cloudflareOnly` | `bool` |  |  |  |
| `spec.rules[].actionParameters.mustRevalidate` | `CloudflareRulesetCacheControlFlag` |  |  |  |
| `spec.rules[].actionParameters.mustRevalidate.operation` | `string` |  |  |  |
| `spec.rules[].actionParameters.mustRevalidate.cloudflareOnly` | `bool` |  |  |  |
| `spec.rules[].actionParameters.proxyRevalidate` | `CloudflareRulesetCacheControlFlag` |  |  |  |
| `spec.rules[].actionParameters.proxyRevalidate.operation` | `string` |  |  |  |
| `spec.rules[].actionParameters.proxyRevalidate.cloudflareOnly` | `bool` |  |  |  |
| `spec.rules[].actionParameters.mustUnderstand` | `CloudflareRulesetCacheControlFlag` |  |  |  |
| `spec.rules[].actionParameters.mustUnderstand.operation` | `string` |  |  |  |
| `spec.rules[].actionParameters.mustUnderstand.cloudflareOnly` | `bool` |  |  |  |
| `spec.rules[].actionParameters.noTransform` | `CloudflareRulesetCacheControlFlag` |  |  |  |
| `spec.rules[].actionParameters.noTransform.operation` | `string` |  |  |  |
| `spec.rules[].actionParameters.noTransform.cloudflareOnly` | `bool` |  |  |  |
| `spec.rules[].actionParameters.immutable` | `CloudflareRulesetCacheControlFlag` |  |  |  |
| `spec.rules[].actionParameters.immutable.operation` | `string` |  |  |  |
| `spec.rules[].actionParameters.immutable.cloudflareOnly` | `bool` |  |  |  |
| `spec.rules[].actionParameters.noStore` | `CloudflareRulesetCacheControlFlag` |  |  |  |
| `spec.rules[].actionParameters.noStore.operation` | `string` |  |  |  |
| `spec.rules[].actionParameters.noStore.cloudflareOnly` | `bool` |  |  |  |
| `spec.rules[].actionParameters.public` | `CloudflareRulesetCacheControlFlag` |  |  |  |
| `spec.rules[].actionParameters.public.operation` | `string` |  |  |  |
| `spec.rules[].actionParameters.public.cloudflareOnly` | `bool` |  |  |  |
| `spec.rules[].actionParameters.operation` | `string` |  |  |  |
| `spec.rules[].actionParameters.values` | `[]string` |  |  |  |
| `spec.rules[].actionParameters.expression` | `string` |  |  |  |
| `spec.rules[].ratelimit` | `CloudflareRulesetRatelimit` |  |  |  |
| `spec.rules[].ratelimit.characteristics` | `[]string` | yes |  |  |
| `spec.rules[].ratelimit.period` | `int64` |  |  |  |
| `spec.rules[].ratelimit.countingExpression` | `string` |  |  |  |
| `spec.rules[].ratelimit.mitigationTimeout` | `int64` |  |  |  |
| `spec.rules[].ratelimit.requestsPerPeriod` | `int64` |  |  |  |
| `spec.rules[].ratelimit.requestsToOrigin` | `bool` |  |  |  |
| `spec.rules[].ratelimit.scorePerPeriod` | `int64` |  |  |  |
| `spec.rules[].ratelimit.scoreResponseHeaderName` | `string` |  |  |  |
| `spec.rules[].logging` | `CloudflareRulesetLogging` |  |  |  |
| `spec.rules[].logging.enabled` | `bool` | yes |  |  |
| `spec.rules[].exposedCredentialCheck` | `CloudflareRulesetExposedCredentialCheck` |  |  |  |
| `spec.rules[].exposedCredentialCheck.usernameExpression` | `string` | yes |  |  |
| `spec.rules[].exposedCredentialCheck.passwordExpression` | `string` | yes |  |  |

## Field Details

### spec.zoneId

`string | valueFrom`

The Cloudflare Zone ID where this ruleset will be created.
Exactly one of zone_id or account_id must be set.
When using value_from, defaults to CloudflareDnsZone kind and status.outputs.zone_id field path.

- references: CloudflareDnsZone (`status.outputs.zone_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: CloudflareDnsZone, name: <that resource's name>, fieldPath: status.outputs.zone_id}} -- a bare string does not parse

### spec.accountId

`string`

The Cloudflare Account ID for account-level rulesets.
Exactly one of zone_id or account_id must be set.

### spec.rulesetKind

`enum` · optional (explicit presence)

The kind of the ruleset.
Default: zone

- default: `zone`
- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `ruleset_kind_unspecified`
- `zone` -- Zone-level ruleset (most common, applies to a single domain).
- `custom` -- Custom ruleset (reusable rule collection, invoked via "execute" action).
- `managed` -- Managed ruleset (Cloudflare-maintained, read-only rules like OWASP Core Ruleset).
- `root` -- Root ruleset (account-level entry point for a phase).

### spec.phase

`enum` · required

The phase determines when the ruleset is evaluated during request processing.

- rule: phase must be specified
- rule: {"required":true,"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `phase_unspecified`
- `ddos_l4` -- DDoS protection at L4 (network layer).
- `ddos_l7` -- DDoS protection at L7 (application layer).
- `http_config_settings` -- Configuration settings (e.g., SSL mode, security level, Rocket Loader) via set_config.
- `http_custom_errors` -- Custom error responses.
- `http_log_custom_fields` -- Custom log fields.
- `http_ratelimit` -- Rate limiting rules.
- `http_request_cache_settings` -- Cache settings rules.
- `http_request_dynamic_redirect` -- Dynamic redirect rules.
- `http_request_firewall_custom` -- Custom firewall rules (WAF custom rules).
- `http_request_firewall_managed` -- Managed firewall rules (WAF managed rulesets like OWASP).
- `http_request_late_transform` -- Late request transform (runs after other transforms).
- `http_request_origin` -- Origin Rules -- override the origin server for matching requests.
- `http_request_redirect` -- Bulk redirect rules.
- `http_request_sanitize` -- Request sanitization.
- `http_request_sbfm` -- Super Bot Fight Mode.
- `http_request_transform` -- Request transform rules (URL rewrites, header modifications).
- `http_response_cache_settings` -- Response cache settings.
- `http_response_compression` -- Response compression rules.
- `http_response_firewall_managed` -- Response-phase managed firewall rules.
- `http_response_headers_transform` -- Response header transform rules.
- `magic_transit` -- Magic Transit rules (network-layer).
- `magic_transit_ids_managed` -- Magic Transit IDS managed rules.
- `magic_transit_managed` -- Magic Transit managed rules.
- `magic_transit_ratelimit` -- Magic Transit rate limiting.

### spec.name

`string` · required

Human-readable name of the ruleset displayed in the Cloudflare dashboard.

- rule: {"string":{"minLen":"1"}}

### spec.description

`string`

Informative description of the ruleset.

### spec.rules

`[]CloudflareRulesetRule` · required

The ordered list of rules in the ruleset.
Rules are evaluated in the order they appear.

- rule: {"repeated":{"minItems":"1"}}

### spec.rules[].ref

`string`

Stable reference ID for the rule.
Prevents Terraform from destroying and recreating the rule when its position changes.

### spec.rules[].expression

`string` · required

Cloudflare wirefilter expression that determines when this rule matches.
Examples: "true" (match all), "http.request.uri.path eq \"/api\"",
"ip.src ne 1.1.1.1", "not starts_with(http.request.uri.path, \"/static\")".
Write prefix/suffix matches in the FUNCTION form (starts_with(field,
"value")): the bare starts_with/ends_with OPERATORS are a paid-plan
grammar extension -- on free-plan zones the API rejects them with 400
code 20127 "expected ComparisonOp" (measured live 2026-08-26), while the
function form parses on every plan.

- rule: {"string":{"minLen":"1"}}

### spec.rules[].action

`enum` · required

The action to perform when the expression matches.

- rule: action must be specified
- rule: {"required":true,"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `action_unspecified`
- `route` -- Override the origin server (used with http_request_origin phase).
- `block` -- Block the request and return a custom response.
- `challenge` -- Present a Cloudflare challenge page.
- `execute` -- Execute a managed ruleset.
- `js_challenge` -- Present a JavaScript challenge.
- `log` -- Log the request without taking action.
- `managed_challenge` -- Present a managed challenge (Cloudflare decides JS or interactive).
- `redirect` -- Redirect the request to a different URL.
- `rewrite` -- Rewrite the request URI or headers.
- `set_cache_settings` -- Configure caching behavior for matching requests.
- `skip` -- Skip execution of other rulesets or products.
- `compress_response` -- Compress the response body.
- `score` -- Assign a threat score to the request.
- `ddos_dynamic` -- Apply an adaptive (managed) DDoS mitigation.
- `force_connection_close` -- Force the connection to close after the response.
- `log_custom_field` -- Add custom fields to the request/response logs (http_log_custom_fields phase).
- `serve_error` -- Serve a custom error asset or inline error response.
- `set_cache_control` -- Set individual Cache-Control response directives.
- `set_cache_tags` -- Set, add, or remove cache tags on the response.
- `set_config` -- Set zone configuration settings (SSL mode, security level, etc.; http_config_settings phase).

### spec.rules[].description

`string`

Human-readable description of what this rule does.

### spec.rules[].enabled

`bool` · optional (explicit presence)

Whether the rule is active.
Default: true

- default: `true`

### spec.rules[].actionParameters

`CloudflareRulesetActionParameters`

Action-specific parameters. The applicable fields depend on the chosen action.

- rule: skip rules map is incompatible with the ruleset option

### spec.rules[].actionParameters.hostHeader

`string`

Override the Host header sent to the origin server.

### spec.rules[].actionParameters.origin

`CloudflareRulesetOrigin`

Override the origin server that Cloudflare connects to.

### spec.rules[].actionParameters.origin.host

`string`

Hostname or IP address of the origin server.

### spec.rules[].actionParameters.origin.port

`int32`

Port number on the origin server.

### spec.rules[].actionParameters.sni

`CloudflareRulesetSni`

Override the SNI value used during TLS handshake with the origin.

### spec.rules[].actionParameters.sni.value

`string`

The SNI hostname to present during the TLS handshake.

### spec.rules[].actionParameters.response

`CloudflareRulesetResponse`

Custom response returned to the client when the request is blocked.

### spec.rules[].actionParameters.response.statusCode

`int32`

HTTP status code to return (e.g., 403, 503).

### spec.rules[].actionParameters.response.content

`string`

Response body content.

### spec.rules[].actionParameters.response.contentType

`string`

MIME type of the response body (e.g., "text/plain", "application/json").

### spec.rules[].actionParameters.uri

`CloudflareRulesetUri`

URI rewrite configuration (path and/or query string).

### spec.rules[].actionParameters.uri.path

`CloudflareRulesetUriComponent`

Path rewrite configuration.

### spec.rules[].actionParameters.uri.path.value

`string`

Static replacement value.

### spec.rules[].actionParameters.uri.path.expression

`string`

Dynamic expression that evaluates to the replacement value.

### spec.rules[].actionParameters.uri.query

`CloudflareRulesetUriComponent`

Query string rewrite configuration.

### spec.rules[].actionParameters.uri.query.value

`string`

Static replacement value.

### spec.rules[].actionParameters.uri.query.expression

`string`

Dynamic expression that evaluates to the replacement value.

### spec.rules[].actionParameters.headers

`map<string, CloudflareRulesetHeader>`

Header modifications. Key is the header name, value defines the operation.

### spec.rules[].actionParameters.headers.*.operation

`string`

The operation to perform: "set", "add", or "remove".

- rule: operation must be one of "set", "add", "remove"

### spec.rules[].actionParameters.headers.*.value

`string`

Static value for the header (used with "set" and "add" operations).

### spec.rules[].actionParameters.headers.*.expression

`string`

Dynamic expression that evaluates to the header value.

### spec.rules[].actionParameters.fromValue

`CloudflareRulesetFromValue`

Redirect configuration with a static/dynamic target URL and status code.

### spec.rules[].actionParameters.fromValue.targetUrl

`CloudflareRulesetTargetUrl`

The target URL to redirect to.

### spec.rules[].actionParameters.fromValue.targetUrl.value

`string`

Static target URL.

### spec.rules[].actionParameters.fromValue.targetUrl.expression

`string`

Dynamic expression that evaluates to the target URL.

### spec.rules[].actionParameters.fromValue.statusCode

`int32`

HTTP status code for the redirect (301, 302, 303, 307, or 308).

### spec.rules[].actionParameters.fromValue.preserveQueryString

`bool`

Whether to preserve the original query string in the redirect.

### spec.rules[].actionParameters.phases

`[]string`

List of phases to skip (e.g., "http_request_firewall_managed").

### spec.rules[].actionParameters.products

`[]string`

List of Cloudflare products to skip (e.g., "zoneLockdown", "uaBlock", "waf").

### spec.rules[].actionParameters.ruleset

`string`

Single ruleset ID to skip (use "current" to skip the remainder of this ruleset).

### spec.rules[].actionParameters.rulesets

`[]string`

Multiple ruleset IDs to skip.

### spec.rules[].actionParameters.rules

`map<string, CloudflareRulesetStringList>`

Skip specific rules INSIDE other rulesets: a map of ruleset ID (32-char
hex) to the rule IDs (32-char hex) in that ruleset to skip. Incompatible
with the single `ruleset` option — use one or the other.

- rule: skip rules: keys and rule IDs must be 32-character hex ruleset/rule IDs, each with at least one rule ID

### spec.rules[].actionParameters.rules.*.values

`[]string`

The list of string values.

### spec.rules[].actionParameters.id

`string`

The ID of the managed ruleset to execute.

### spec.rules[].actionParameters.overrides

`CloudflareRulesetOverrides`

Overrides to apply when executing a managed ruleset.

### spec.rules[].actionParameters.overrides.action

`string`

Default action to apply to all rules in the managed ruleset.

### spec.rules[].actionParameters.overrides.enabled

`bool`

Whether rules in the managed ruleset are enabled by default.

### spec.rules[].actionParameters.overrides.categories

`[]CloudflareRulesetCategoryOverride`

Category-level overrides.

### spec.rules[].actionParameters.overrides.categories[].category

`string` · required

The category name (e.g., "wordpress", "xss", "sqli").

- rule: {"string":{"minLen":"1"}}

### spec.rules[].actionParameters.overrides.categories[].action

`string`

Action to apply to rules in this category.

### spec.rules[].actionParameters.overrides.categories[].enabled

`bool`

Whether rules in this category are enabled.

### spec.rules[].actionParameters.overrides.categories[].sensitivityLevel

`string`

Sensitivity level for this category: "default", "medium", "low", or "eoff".

- rule: sensitivity_level must be one of "default", "medium", "low", "eoff"

### spec.rules[].actionParameters.overrides.rules

`[]CloudflareRulesetRuleOverride`

Individual rule-level overrides.

### spec.rules[].actionParameters.overrides.rules[].id

`string` · required

The rule ID within the managed ruleset.

- rule: {"string":{"minLen":"1"}}

### spec.rules[].actionParameters.overrides.rules[].action

`string`

Action to apply for this rule.

### spec.rules[].actionParameters.overrides.rules[].enabled

`bool`

Whether this rule is enabled.

### spec.rules[].actionParameters.overrides.rules[].scoreThreshold

`int32`

Score threshold for this rule.

### spec.rules[].actionParameters.overrides.rules[].sensitivityLevel

`string`

Sensitivity level for this rule: "default", "medium", "low", or "eoff".

- rule: sensitivity_level must be one of "default", "medium", "low", "eoff"

### spec.rules[].actionParameters.overrides.sensitivityLevel

`string`

Sensitivity level for all rules: "default", "medium", "low", or "eoff".

- rule: sensitivity_level must be one of "default", "medium", "low", "eoff"

### spec.rules[].actionParameters.cache

`bool`

Whether to cache matching requests.

### spec.rules[].actionParameters.edgeTtl

`CloudflareRulesetEdgeTtl`

Edge TTL configuration (how long Cloudflare caches the response).

### spec.rules[].actionParameters.edgeTtl.mode

`string`

TTL mode: "respect_origin", "override_origin", or "bypass_by_default".

- rule: mode must be one of "respect_origin", "override_origin", "bypass_by_default"

### spec.rules[].actionParameters.edgeTtl.defaultTtl

`int32`

Default TTL in seconds when mode is "override_origin".

### spec.rules[].actionParameters.edgeTtl.statusCodeTtls

`[]CloudflareRulesetStatusCodeTtl`

Status code-specific TTL overrides.

### spec.rules[].actionParameters.edgeTtl.statusCodeTtls[].value

`int32`

TTL value in seconds.

### spec.rules[].actionParameters.edgeTtl.statusCodeTtls[].statusCode

`int32`

Specific HTTP status code (e.g., 200, 404). Use this or status_code_range, not both.

### spec.rules[].actionParameters.edgeTtl.statusCodeTtls[].statusCodeRange

`CloudflareRulesetStatusCodeRange`

Range of HTTP status codes. Use this or status_code, not both.

### spec.rules[].actionParameters.edgeTtl.statusCodeTtls[].statusCodeRange.from

`int32`

Start of the range (inclusive).

### spec.rules[].actionParameters.edgeTtl.statusCodeTtls[].statusCodeRange.to

`int32`

End of the range (inclusive).

### spec.rules[].actionParameters.browserTtl

`CloudflareRulesetBrowserTtl`

Browser TTL configuration (Cache-Control max-age sent to the client).

### spec.rules[].actionParameters.browserTtl.mode

`string`

TTL mode: "respect_origin", "override_origin", "bypass_by_default", or "bypass".

- rule: mode must be one of "respect_origin", "override_origin", "bypass_by_default", "bypass"

### spec.rules[].actionParameters.browserTtl.defaultTtl

`int32`

Default TTL in seconds when mode is "override_origin".

### spec.rules[].actionParameters.serveStale

`CloudflareRulesetServeStale`

Controls serving stale content while revalidating with the origin.

### spec.rules[].actionParameters.serveStale.disableStaleWhileUpdating

`bool`

When true, Cloudflare will not serve stale content while revalidating.

### spec.rules[].actionParameters.fromList

`CloudflareRulesetFromList`

Redirect using a Bulk Redirect list rather than an inline target.

### spec.rules[].actionParameters.fromList.name

`string | valueFrom` · required

The Bulk Redirect list to match against, identified by name. Accepts a
literal list name or a reference to a CloudflareList resource (defaulting to
that list's name output), so a redirect ruleset composes with the list that
backs it. Cloudflare resolves the list by name, so the referenced output is
the list's name (not its id).

- references: CloudflareList (`status.outputs.name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: CloudflareList, name: <that resource's name>, fieldPath: status.outputs.name}} -- a bare string does not parse

### spec.rules[].actionParameters.fromList.key

`string` · required

Expression that produces the lookup key into the list (e.g.
"http.request.full_uri").

- rule: {"string":{"minLen":"1"}}

### spec.rules[].actionParameters.algorithms

`[]CloudflareRulesetAlgorithm`

Ordered list of compression algorithms the edge may use (e.g. "brotli", "gzip",
"zstd", "auto", "default", "none").

### spec.rules[].actionParameters.algorithms[].name

`string` · required

Algorithm name: "none", "auto", "default", "gzip", "brotli", or "zstd".

- rule: {"string":{"minLen":"1"}}

### spec.rules[].actionParameters.matchedData

`CloudflareRulesetMatchedData`

Public key used to encrypt matched sensitive data for the Exposed Credentials /
Sensitive Data Detection logs.

### spec.rules[].actionParameters.matchedData.publicKey

`string` · required

Base64-encoded public key for encrypting matched sensitive data.

- rule: {"string":{"minLen":"1"}}

### spec.rules[].actionParameters.increment

`int64`

Amount to increment the threat score by.

### spec.rules[].actionParameters.assetName

`string`

Name of the custom error asset to serve.

### spec.rules[].actionParameters.content

`string`

Inline error response body to serve.

### spec.rules[].actionParameters.contentType

`string`

MIME type of the inline error response (e.g. "text/html", "application/json",
"text/plain", "text/xml").

- rule: content_type must be one of "application/json", "text/html", "text/plain", "text/xml"

### spec.rules[].actionParameters.statusCode

`int64`

HTTP status code of the served error response.

### spec.rules[].actionParameters.automaticHttpsRewrites

`bool` · optional (explicit presence)

Enable Automatic HTTPS Rewrites.

### spec.rules[].actionParameters.autominify

`CloudflareRulesetAutominify`

Auto-minify settings for CSS/HTML/JS.

### spec.rules[].actionParameters.autominify.css

`bool` · optional (explicit presence)

Minify CSS.

### spec.rules[].actionParameters.autominify.html

`bool` · optional (explicit presence)

Minify HTML.

### spec.rules[].actionParameters.autominify.js

`bool` · optional (explicit presence)

Minify JavaScript.

### spec.rules[].actionParameters.bic

`bool` · optional (explicit presence)

Enable Browser Integrity Check.

### spec.rules[].actionParameters.contentConverter

`bool` · optional (explicit presence)

Enable the content converter.

### spec.rules[].actionParameters.disableApps

`bool` · optional (explicit presence)

Disable Cloudflare Apps.

### spec.rules[].actionParameters.disableRum

`bool` · optional (explicit presence)

Disable Real User Monitoring (RUM).

### spec.rules[].actionParameters.disableZaraz

`bool` · optional (explicit presence)

Disable Zaraz.

### spec.rules[].actionParameters.emailObfuscation

`bool` · optional (explicit presence)

Enable Email Obfuscation.

### spec.rules[].actionParameters.fonts

`bool` · optional (explicit presence)

Enable Cloudflare Fonts.

### spec.rules[].actionParameters.hotlinkProtection

`bool` · optional (explicit presence)

Enable Hotlink Protection.

### spec.rules[].actionParameters.mirage

`bool` · optional (explicit presence)

Enable Mirage image optimization.

### spec.rules[].actionParameters.opportunisticEncryption

`bool` · optional (explicit presence)

Enable Opportunistic Encryption.

### spec.rules[].actionParameters.polish

`string`

Polish image optimization level: "off", "lossless", "lossy", or "webp".

- rule: polish must be one of "off", "lossless", "lossy", "webp"

### spec.rules[].actionParameters.redirectsForAiTraining

`bool` · optional (explicit presence)

Whether to allow AI training crawlers to follow redirects.

### spec.rules[].actionParameters.requestBodyBuffering

`string`

Request body buffering mode: "none", "standard", or "full".

- rule: request_body_buffering must be one of "none", "standard", "full"

### spec.rules[].actionParameters.responseBodyBuffering

`string`

Response body buffering mode: "none" or "standard".

- rule: response_body_buffering must be one of "none", "standard"

### spec.rules[].actionParameters.rocketLoader

`bool` · optional (explicit presence)

Enable Rocket Loader.

### spec.rules[].actionParameters.securityLevel

`string`

Security level: "off", "essentially_off", "low", "medium", "high", "under_attack".

- rule: security_level must be one of "off", "essentially_off", "low", "medium", "high", "under_attack"

### spec.rules[].actionParameters.serverSideExcludes

`bool` · optional (explicit presence)

Enable Server-Side Excludes.

### spec.rules[].actionParameters.ssl

`string`

SSL/TLS mode: "off", "flexible", "full", "strict", or "origin_pull".

- rule: ssl must be one of "off", "flexible", "full", "strict", "origin_pull"

### spec.rules[].actionParameters.sxg

`bool` · optional (explicit presence)

Enable Signed Exchanges (SXG).

### spec.rules[].actionParameters.additionalCacheablePorts

`[]int64`

Additional origin ports that are eligible for caching.

### spec.rules[].actionParameters.cacheKey

`CloudflareRulesetCacheKey`

Custom cache key configuration (what request attributes form the cache key).

### spec.rules[].actionParameters.cacheKey.cacheByDeviceType

`bool` · optional (explicit presence)

Include the device type (mobile/tablet/desktop) in the cache key.

### spec.rules[].actionParameters.cacheKey.cacheDeceptionArmor

`bool` · optional (explicit presence)

Protect against cache deception attacks.

### spec.rules[].actionParameters.cacheKey.customKey

`CloudflareRulesetCacheKeyCustomKey`

Fine-grained custom cache key components.

### spec.rules[].actionParameters.cacheKey.customKey.cookie

`CloudflareRulesetCacheKeyCookie`

Cookie-based cache key components.

### spec.rules[].actionParameters.cacheKey.customKey.cookie.checkPresence

`[]string`

Cookies whose presence (not value) is included in the cache key.

### spec.rules[].actionParameters.cacheKey.customKey.cookie.include

`[]string`

Cookies whose name and value are included in the cache key.

### spec.rules[].actionParameters.cacheKey.customKey.header

`CloudflareRulesetCacheKeyHeader`

Header-based cache key components.

### spec.rules[].actionParameters.cacheKey.customKey.header.checkPresence

`[]string`

Headers whose presence (not value) is included in the cache key.

### spec.rules[].actionParameters.cacheKey.customKey.header.contains

`map<string, CloudflareRulesetStringList>`

For each header name, only include the request when its value contains one of the
listed substrings.

### spec.rules[].actionParameters.cacheKey.customKey.header.contains.*.values

`[]string`

The list of string values.

### spec.rules[].actionParameters.cacheKey.customKey.header.excludeOrigin

`bool` · optional (explicit presence)

Exclude the Origin header from the cache key.

### spec.rules[].actionParameters.cacheKey.customKey.header.include

`[]string`

Headers whose name and value are included in the cache key.

### spec.rules[].actionParameters.cacheKey.customKey.host

`CloudflareRulesetCacheKeyHost`

Host-based cache key components.

### spec.rules[].actionParameters.cacheKey.customKey.host.resolved

`bool` · optional (explicit presence)

Use the resolved (origin) host rather than the request host.

### spec.rules[].actionParameters.cacheKey.customKey.queryString

`CloudflareRulesetCacheKeyQueryString`

Query-string-based cache key components.

### spec.rules[].actionParameters.cacheKey.customKey.queryString.include

`CloudflareRulesetCacheKeyQueryStringFilter`

Query parameters to include in the cache key.

### spec.rules[].actionParameters.cacheKey.customKey.queryString.include.list

`[]string`

Specific query parameter names. Used when "all" is false.

### spec.rules[].actionParameters.cacheKey.customKey.queryString.include.all

`bool` · optional (explicit presence)

Apply to all query parameters.

### spec.rules[].actionParameters.cacheKey.customKey.queryString.exclude

`CloudflareRulesetCacheKeyQueryStringFilter`

Query parameters to exclude from the cache key.

### spec.rules[].actionParameters.cacheKey.customKey.queryString.exclude.list

`[]string`

Specific query parameter names. Used when "all" is false.

### spec.rules[].actionParameters.cacheKey.customKey.queryString.exclude.all

`bool` · optional (explicit presence)

Apply to all query parameters.

### spec.rules[].actionParameters.cacheKey.customKey.user

`CloudflareRulesetCacheKeyUser`

User-attribute-based cache key components.

### spec.rules[].actionParameters.cacheKey.customKey.user.deviceType

`bool` · optional (explicit presence)

Include the user's device type.

### spec.rules[].actionParameters.cacheKey.customKey.user.geo

`bool` · optional (explicit presence)

Include the user's country/region (geo).

### spec.rules[].actionParameters.cacheKey.customKey.user.lang

`bool` · optional (explicit presence)

Include the user's language.

### spec.rules[].actionParameters.cacheKey.ignoreQueryStringsOrder

`bool` · optional (explicit presence)

Treat query string parameters as unordered when forming the cache key.

### spec.rules[].actionParameters.cacheReserve

`CloudflareRulesetCacheReserve`

Cache Reserve configuration (persist eligible objects to Cache Reserve).

### spec.rules[].actionParameters.cacheReserve.eligible

`bool`

Whether matching objects are eligible for Cache Reserve.

### spec.rules[].actionParameters.cacheReserve.minimumFileSize

`int64`

Minimum object size in bytes to be stored in Cache Reserve.

### spec.rules[].actionParameters.originCacheControl

`bool` · optional (explicit presence)

Honor origin Cache-Control headers.

### spec.rules[].actionParameters.originErrorPagePassthru

`bool` · optional (explicit presence)

Pass through origin error pages instead of Cloudflare's.

### spec.rules[].actionParameters.readTimeout

`int64`

Origin response timeout in seconds (Enterprise).

### spec.rules[].actionParameters.respectStrongEtags

`bool` · optional (explicit presence)

Respect strong ETags from the origin.

### spec.rules[].actionParameters.vary

`CloudflareRulesetVary`

Vary-based caching (cache variants keyed on response headers).

### spec.rules[].actionParameters.vary.default

`CloudflareRulesetVaryDefault`

Default vary behavior.

### spec.rules[].actionParameters.vary.default.action

`string` · required

The default vary action.

- rule: {"string":{"minLen":"1"}}

### spec.rules[].actionParameters.vary.headers

`map<string, CloudflareRulesetVaryHeader>`

Per-header vary configuration. Key is the header name.

### spec.rules[].actionParameters.vary.headers.*.action

`string` · required

The vary action for this header.

- rule: {"string":{"minLen":"1"}}

### spec.rules[].actionParameters.vary.headers.*.mediaTypes

`[]string`

Media types this header variant applies to.

### spec.rules[].actionParameters.vary.headers.*.languages

`[]string`

Languages this header variant applies to.

### spec.rules[].actionParameters.stripEtags

`bool` · optional (explicit presence)

Strip ETag headers from the response.

### spec.rules[].actionParameters.stripLastModified

`bool` · optional (explicit presence)

Strip Last-Modified headers from the response.

### spec.rules[].actionParameters.stripSetCookie

`bool` · optional (explicit presence)

Strip Set-Cookie headers from the cached response.

### spec.rules[].actionParameters.cookieFields

`[]CloudflareRulesetLogField`

Cookie fields to add to logs.

### spec.rules[].actionParameters.cookieFields[].name

`string` · required

The field name.

- rule: {"string":{"minLen":"1"}}

### spec.rules[].actionParameters.rawResponseFields

`[]CloudflareRulesetLogResponseField`

Raw (unmodified) response header fields to add to logs.

### spec.rules[].actionParameters.rawResponseFields[].name

`string` · required

The field name.

- rule: {"string":{"minLen":"1"}}

### spec.rules[].actionParameters.rawResponseFields[].preserveDuplicates

`bool`

Whether to preserve duplicate occurrences of the field.

### spec.rules[].actionParameters.requestFields

`[]CloudflareRulesetLogField`

Request header fields to add to logs.

### spec.rules[].actionParameters.requestFields[].name

`string` · required

The field name.

- rule: {"string":{"minLen":"1"}}

### spec.rules[].actionParameters.responseFields

`[]CloudflareRulesetLogResponseField`

Response header fields to add to logs.

### spec.rules[].actionParameters.responseFields[].name

`string` · required

The field name.

- rule: {"string":{"minLen":"1"}}

### spec.rules[].actionParameters.responseFields[].preserveDuplicates

`bool`

Whether to preserve duplicate occurrences of the field.

### spec.rules[].actionParameters.transformedRequestFields

`[]CloudflareRulesetLogField`

Transformed request header fields to add to logs.

### spec.rules[].actionParameters.transformedRequestFields[].name

`string` · required

The field name.

- rule: {"string":{"minLen":"1"}}

### spec.rules[].actionParameters.maxAge

`CloudflareRulesetCacheControlValue`

Cache-Control max-age directive.

### spec.rules[].actionParameters.maxAge.operation

`string`

The operation: "set" or "remove".

- rule: operation must be one of "set", "remove"

### spec.rules[].actionParameters.maxAge.value

`int64`

The numeric value in seconds (used with "set").

### spec.rules[].actionParameters.maxAge.cloudflareOnly

`bool`

Apply the directive only to Cloudflare's cache, not the client.

### spec.rules[].actionParameters.sMaxage

`CloudflareRulesetCacheControlValue`

Cache-Control s-maxage directive.

### spec.rules[].actionParameters.sMaxage.operation

`string`

The operation: "set" or "remove".

- rule: operation must be one of "set", "remove"

### spec.rules[].actionParameters.sMaxage.value

`int64`

The numeric value in seconds (used with "set").

### spec.rules[].actionParameters.sMaxage.cloudflareOnly

`bool`

Apply the directive only to Cloudflare's cache, not the client.

### spec.rules[].actionParameters.staleWhileRevalidate

`CloudflareRulesetCacheControlValue`

Cache-Control stale-while-revalidate directive.

### spec.rules[].actionParameters.staleWhileRevalidate.operation

`string`

The operation: "set" or "remove".

- rule: operation must be one of "set", "remove"

### spec.rules[].actionParameters.staleWhileRevalidate.value

`int64`

The numeric value in seconds (used with "set").

### spec.rules[].actionParameters.staleWhileRevalidate.cloudflareOnly

`bool`

Apply the directive only to Cloudflare's cache, not the client.

### spec.rules[].actionParameters.staleIfError

`CloudflareRulesetCacheControlValue`

Cache-Control stale-if-error directive.

### spec.rules[].actionParameters.staleIfError.operation

`string`

The operation: "set" or "remove".

- rule: operation must be one of "set", "remove"

### spec.rules[].actionParameters.staleIfError.value

`int64`

The numeric value in seconds (used with "set").

### spec.rules[].actionParameters.staleIfError.cloudflareOnly

`bool`

Apply the directive only to Cloudflare's cache, not the client.

### spec.rules[].actionParameters.private

`CloudflareRulesetCacheControlQualifiers`

Cache-Control private directive (with optional field qualifiers).

### spec.rules[].actionParameters.private.operation

`string`

The operation: "set" or "remove".

- rule: operation must be one of "set", "remove"

### spec.rules[].actionParameters.private.qualifiers

`[]string`

Optional header field names the directive applies to.

### spec.rules[].actionParameters.private.cloudflareOnly

`bool`

Apply the directive only to Cloudflare's cache, not the client.

### spec.rules[].actionParameters.noCache

`CloudflareRulesetCacheControlQualifiers`

Cache-Control no-cache directive (with optional field qualifiers).

### spec.rules[].actionParameters.noCache.operation

`string`

The operation: "set" or "remove".

- rule: operation must be one of "set", "remove"

### spec.rules[].actionParameters.noCache.qualifiers

`[]string`

Optional header field names the directive applies to.

### spec.rules[].actionParameters.noCache.cloudflareOnly

`bool`

Apply the directive only to Cloudflare's cache, not the client.

### spec.rules[].actionParameters.mustRevalidate

`CloudflareRulesetCacheControlFlag`

Cache-Control must-revalidate directive.

### spec.rules[].actionParameters.mustRevalidate.operation

`string`

The operation: "set" or "remove".

- rule: operation must be one of "set", "remove"

### spec.rules[].actionParameters.mustRevalidate.cloudflareOnly

`bool`

Apply the directive only to Cloudflare's cache, not the client.

### spec.rules[].actionParameters.proxyRevalidate

`CloudflareRulesetCacheControlFlag`

Cache-Control proxy-revalidate directive.

### spec.rules[].actionParameters.proxyRevalidate.operation

`string`

The operation: "set" or "remove".

- rule: operation must be one of "set", "remove"

### spec.rules[].actionParameters.proxyRevalidate.cloudflareOnly

`bool`

Apply the directive only to Cloudflare's cache, not the client.

### spec.rules[].actionParameters.mustUnderstand

`CloudflareRulesetCacheControlFlag`

Cache-Control must-understand directive.

### spec.rules[].actionParameters.mustUnderstand.operation

`string`

The operation: "set" or "remove".

- rule: operation must be one of "set", "remove"

### spec.rules[].actionParameters.mustUnderstand.cloudflareOnly

`bool`

Apply the directive only to Cloudflare's cache, not the client.

### spec.rules[].actionParameters.noTransform

`CloudflareRulesetCacheControlFlag`

Cache-Control no-transform directive.

### spec.rules[].actionParameters.noTransform.operation

`string`

The operation: "set" or "remove".

- rule: operation must be one of "set", "remove"

### spec.rules[].actionParameters.noTransform.cloudflareOnly

`bool`

Apply the directive only to Cloudflare's cache, not the client.

### spec.rules[].actionParameters.immutable

`CloudflareRulesetCacheControlFlag`

Cache-Control immutable directive.

### spec.rules[].actionParameters.immutable.operation

`string`

The operation: "set" or "remove".

- rule: operation must be one of "set", "remove"

### spec.rules[].actionParameters.immutable.cloudflareOnly

`bool`

Apply the directive only to Cloudflare's cache, not the client.

### spec.rules[].actionParameters.noStore

`CloudflareRulesetCacheControlFlag`

Cache-Control no-store directive.

### spec.rules[].actionParameters.noStore.operation

`string`

The operation: "set" or "remove".

- rule: operation must be one of "set", "remove"

### spec.rules[].actionParameters.noStore.cloudflareOnly

`bool`

Apply the directive only to Cloudflare's cache, not the client.

### spec.rules[].actionParameters.public

`CloudflareRulesetCacheControlFlag`

Cache-Control public directive.

### spec.rules[].actionParameters.public.operation

`string`

The operation: "set" or "remove".

- rule: operation must be one of "set", "remove"

### spec.rules[].actionParameters.public.cloudflareOnly

`bool`

Apply the directive only to Cloudflare's cache, not the client.

### spec.rules[].actionParameters.operation

`string`

Operation on the cache tags: "set", "add", or "remove".

- rule: operation must be one of "set", "add", "remove"

### spec.rules[].actionParameters.values

`[]string`

Cache tag values to apply.

### spec.rules[].actionParameters.expression

`string`

Expression that evaluates to additional cache tags.

### spec.rules[].ratelimit

`CloudflareRulesetRatelimit`

Rate-limiting configuration. Required for rules in the http_ratelimit phase.

### spec.rules[].ratelimit.characteristics

`[]string` · required

Characteristics (buckets) that requests are counted against, e.g.
["ip.src", "cf.colo.id"] or ["http.request.headers[\"x-api-key\"]"].

- rule: {"repeated":{"minItems":"1"}}

### spec.rules[].ratelimit.period

`int64`

The period in seconds over which requests/score are counted (10, 60, 120, 300, 600, ...).

- rule: {"int64":{"gt":"0"}}

### spec.rules[].ratelimit.countingExpression

`string`

Optional expression defining what counts toward the rate limit. Defaults to the
rule's match expression when empty.

### spec.rules[].ratelimit.mitigationTimeout

`int64`

Seconds the mitigation (the rule's action) stays in effect once triggered.
Leave 0 to use the period.

- rule: {"int64":{"gte":"0"}}

### spec.rules[].ratelimit.requestsPerPeriod

`int64`

Number of requests per period that triggers the rule. Use this OR score_per_period.

- rule: {"int64":{"gte":"0"}}

### spec.rules[].ratelimit.requestsToOrigin

`bool`

Count only requests that reach the origin (not those served from cache/edge).

### spec.rules[].ratelimit.scorePerPeriod

`int64`

Score per period that triggers the rule (for score-based rate limiting). Use this
OR requests_per_period.

- rule: {"int64":{"gte":"0"}}

### spec.rules[].ratelimit.scoreResponseHeaderName

`string`

Response header name whose value is summed into the score for score-based limiting.

### spec.rules[].logging

`CloudflareRulesetLogging`

Per-rule logging configuration (override the rule's default log behavior).

### spec.rules[].logging.enabled

`bool` · required · optional (explicit presence)

Whether request logging is enabled for this rule. Required inside logging.

- rule: {"required":true}

### spec.rules[].exposedCredentialCheck

`CloudflareRulesetExposedCredentialCheck`

Leaked-credential / exposed-credential detection for the matched request.

### spec.rules[].exposedCredentialCheck.usernameExpression

`string` · required

Expression that extracts the submitted username from the request.

- rule: {"string":{"minLen":"1"}}

### spec.rules[].exposedCredentialCheck.passwordExpression

`string` · required

Expression that extracts the submitted password from the request.

- rule: {"string":{"minLen":"1"}}

## Validation Rules

- `spec.zone_or_account`: exactly one of zone_id or account_id must be set

## Outputs

Reference an output from another manifest as `valueFrom: {kind: CloudflareRuleset, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.ruleset_id` | `string` | The Cloudflare-assigned unique identifier of the ruleset. |
| `status.outputs.version` | `string` | The current version of the ruleset (increments on each update). |
| `status.outputs.zone_id` | `string` | The zone ID the ruleset belongs to (pass-through for downstream resource references). |
| `status.outputs.phase` | `string` | The phase the ruleset executes in (pass-through for infra-chart conditionals). |
| `status.outputs.last_updated` | `string` | RFC3339 timestamp of the ruleset's last update. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.zoneId` | CloudflareDnsZone | `status.outputs.zone_id` |
| `spec.rules[].actionParameters.fromList.name` | CloudflareList | `status.outputs.name` |

## See Also

- [Overview](../README.md)
