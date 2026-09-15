# AzureFrontDoorRoute

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `azure.planton.dev/v1alpha1`

**AzureFrontDoorRouteSpec** defines the configuration for creating a
route inside an Azure Front Door endpoint -- the rule that connects
client traffic arriving at the endpoint to an origin group, by URL
pattern, with protocol and edge-caching behavior.

A route is the traffic-serving edge of the Front Door graph: endpoint
(entry hostname) -> route (match + policy) -> origin group (backend
pool) -> origins (backends). Routes are many-per-endpoint with
independent lifecycles (one endpoint commonly serves "/api/*" and
"/static/*" from different backends), which is why the route is a
first-class kind referencing the endpoint rather than a list folded
into it.

**ForceNew fields**: `endpoint_id`, `route_name` -- both fix the
route's ARM identity at creation.

## Example

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureFrontDoorRoute
metadata:
  name: test-front-door-route
spec:
  endpointId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.Cdn/profiles/test-frontdoor/afdEndpoints/test-web
  routeName: api-route
  originGroupId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.Cdn/profiles/test-frontdoor/originGroups/api-backends
  # Exercises the deploy-ordering seam (never sent to Azure).
  originIds:
    - value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.Cdn/profiles/test-frontdoor/originGroups/api-backends/origins/primary-app
  # Exercises the attached delivery policy.
  ruleSetIds:
    - value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.Cdn/profiles/test-frontdoor/ruleSets/deliverypolicy
  # Exercises the custom-domain attachment with the default domain
  # disabled (legal only because a domain is attached).
  customDomainIds:
    - value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.Cdn/profiles/test-frontdoor/customDomains/www-example-com
  linkToDefaultDomain: false
  patternsToMatch:
    - /api/*
  supportedProtocols:
    - HTTP
    - HTTPS
  # Exercises the origin-leg protocol pin and the origin-path prefix.
  forwardingProtocol: HTTPS_ONLY
  originPath: /v1
  # Exercises the cache block with query-string keying and compression.
  cache:
    queryStringCachingBehavior: INCLUDE_SPECIFIED_QUERY_STRINGS
    queryStrings:
      - page
      - sort
    compressionEnabled: true
    contentTypesToCompress:
      - application/json
      - text/html
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.endpointId` | `string \| valueFrom` | yes |  | AzureFrontDoorEndpoint (`status.outputs.endpoint_id`) |
| `spec.routeName` | `string` | yes |  |  |
| `spec.originGroupId` | `string \| valueFrom` | yes |  | AzureFrontDoorOriginGroup (`status.outputs.origin_group_id`) |
| `spec.originIds` | `[]string \| valueFrom` |  |  | AzureFrontDoorOrigin (`status.outputs.origin_id`) |
| `spec.ruleSetIds` | `[]string \| valueFrom` |  |  | AzureFrontDoorRuleSet (`status.outputs.rule_set_id`) |
| `spec.patternsToMatch` | `[]string` | yes |  |  |
| `spec.supportedProtocols` | `[]enum` | yes |  |  |
| `spec.forwardingProtocol` | `enum` |  |  |  |
| `spec.httpsRedirectEnabled` | `bool` |  | `true` |  |
| `spec.customDomainIds` | `[]string \| valueFrom` |  |  | AzureFrontDoorCustomDomain (`status.outputs.custom_domain_id`) |
| `spec.linkToDefaultDomain` | `bool` |  | `true` |  |
| `spec.enabled` | `bool` |  | `true` |  |
| `spec.originPath` | `string` | yes |  |  |
| `spec.cache` | `AzureFrontDoorRouteCache` |  |  |  |
| `spec.cache.queryStringCachingBehavior` | `enum` |  |  |  |
| `spec.cache.queryStrings` | `[]string` |  |  |  |
| `spec.cache.compressionEnabled` | `bool` |  | `false` |  |
| `spec.cache.contentTypesToCompress` | `[]string` |  |  |  |

## Field Details

### spec.endpointId

`string | valueFrom` · required

The Front Door endpoint the route attaches to, by ARM ID. References
an AzureFrontDoorEndpoint's endpoint_id output so the endpoint and
its routes compose in one manifest set. Fixed at creation.

- references: AzureFrontDoorEndpoint (`status.outputs.endpoint_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureFrontDoorEndpoint, name: <that resource's name>, fieldPath: status.outputs.endpoint_id}} -- a bare string does not parse

### spec.routeName

`string` · required

The route's name -- unique within the endpoint.

2-90 characters; letters, digits, and hyphens; must start and end
with a letter or digit.

**ForceNew**: changing the name replaces the route.

- rule: route_name must be 2-90 characters, start and end with a letter or digit, and contain only letters, digits, and hyphens
- rule: {"required":true,"string":{"minLen":"2","maxLen":"90"}}

### spec.originGroupId

`string | valueFrom` · required

The origin group that answers requests matched by this route, by ARM
ID. References an AzureFrontDoorOriginGroup's origin_group_id
output. Updatable in place (repointing a route is how traffic moves
between backend pools). The route lives in its endpoint and forwards
TO the origin group, so the reference is access, not placement, on a
diagram.

- references: AzureFrontDoorOriginGroup (`status.outputs.origin_group_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureFrontDoorOriginGroup, name: <that resource's name>, fieldPath: status.outputs.origin_group_id}} -- a bare string does not parse

### spec.originIds

`[]string | valueFrom`

The origins this route depends on, by ARM ID -- each references an
AzureFrontDoorOrigin's origin_id output. Azure never receives these
(membership is defined by the origin group); they exist purely to
sequence deployment and teardown, because ARM rejects a route whose
origin group has no origins yet. List the group's origins here when
deploying the whole chain in one manifest set; omit when the origins
already exist.

- references: AzureFrontDoorOrigin (`status.outputs.origin_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureFrontDoorOrigin, name: <that resource's name>, fieldPath: status.outputs.origin_id}} -- a bare string does not parse

### spec.ruleSetIds

`[]string | valueFrom`

The rule sets whose delivery policies apply to traffic on this
route, by ARM ID -- each references an AzureFrontDoorRuleSet's
rule_set_id output. Rule sets and the route must live in the same
profile. Order here does not matter: WITHIN a set, rules run by
their own order; ACROSS sets, Azure evaluates all attached sets'
rules together. One rule set is commonly attached to many routes --
that sharing is why the policy is its own kind.

- references: AzureFrontDoorRuleSet (`status.outputs.rule_set_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureFrontDoorRuleSet, name: <that resource's name>, fieldPath: status.outputs.rule_set_id}} -- a bare string does not parse

### spec.patternsToMatch

`[]string` · required

The URL path patterns this route matches, e.g. "/*" (everything),
"/api/*", "/images/*". At least one; every pattern starts with "/".
Route matching picks the most specific pattern across the
endpoint's routes, so "/api/*" on one route and "/*" on another
cleanly split API and catch-all traffic.

- rule: {"repeated":{"minItems":"1","items":{"cel":[{"id":"front_door_route_pattern_format","message":"each pattern must start with '/'","expression":"this.startsWith('/')"}]}}}

### spec.supportedProtocols

`[]enum` · required

The client-facing protocols the route accepts: HTTP, HTTPS, or both.
At least one. Serving both with https_redirect_enabled (the
default) is the standard production posture -- HTTP arrives only to
be redirected.

- rule: {"repeated":{"minItems":"1","maxItems":"2","unique":true,"items":{"enum":{"definedOnly":true,"notIn":[0]}}}}

Allowed values (use exactly as shown):

- `azure_front_door_route_protocol_unspecified` -- Not specified -- invalid; list HTTP and/or HTTPS explicitly.
- `HTTP` -- Plain HTTP (port 80 at the edge).
- `HTTPS` -- TLS (port 443 at the edge, Front Door's managed edge certificate on the default domain).

### spec.forwardingProtocol

`enum`

The protocol Front Door uses toward the ORIGIN (independent of what
the client used). Unspecified deploys MATCH_REQUEST -- mirror the
client's protocol. HTTPS_ONLY keeps the origin leg encrypted even
for HTTP clients; HTTP_ONLY is for origins without TLS (pair it
with Private Link rather than sending plaintext over the internet).

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `azure_front_door_route_forwarding_protocol_unspecified` -- Not specified -- deploys MATCH_REQUEST, Azure's default.
- `MATCH_REQUEST` -- Mirror the client's protocol on the origin leg.
- `HTTP_ONLY` -- Always connect to the origin over plain HTTP.
- `HTTPS_ONLY` -- Always connect to the origin over TLS -- the right choice when the origin has a certificate, regardless of the client protocol.

### spec.httpsRedirectEnabled

`bool` · optional (explicit presence)

Redirect HTTP requests to HTTPS at the edge (301). Default true.
Requires the route to accept BOTH protocols -- the redirect needs
HTTP to arrive and HTTPS to land on.

- default: `true`

### spec.customDomainIds

`[]string | valueFrom`

The custom domains this route serves, by ARM ID -- each references
an AzureFrontDoorCustomDomain's custom_domain_id output. The route
side owns the domain attachment (a domain lists no routes); every
referenced domain must belong to the route's profile and should be
DNS-validated before traffic can flow. Empty means the route serves
only the endpoint's generated *.azurefd.net hostname.

- references: AzureFrontDoorCustomDomain (`status.outputs.custom_domain_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureFrontDoorCustomDomain, name: <that resource's name>, fieldPath: status.outputs.custom_domain_id}} -- a bare string does not parse

### spec.linkToDefaultDomain

`bool` · optional (explicit presence)

Serve this route on the endpoint's generated *.azurefd.net
hostname. Default true. Disabling requires at least one custom
domain in custom_domain_ids -- otherwise the route would serve no
hostname at all. Disable it for production routes that should only
answer on their custom domains.

- default: `true`

### spec.enabled

`bool` · optional (explicit presence)

Whether the route matches traffic. Disabling stops matching without
deleting the route's configuration. Default true.

- default: `true`

### spec.originPath

`string` · required · optional (explicit presence)

A path prepended on the ORIGIN side before forwarding, e.g.
"/site1". A request for "/page.html" on a route with origin_path
"/site1" fetches "/site1/page.html" from the origin -- how one
backend hosts many routes' content in subdirectories. Unset
forwards the client path unchanged.

- rule: {"string":{"minLen":"1"}}

### spec.cache

`AzureFrontDoorRouteCache`

Cache matched responses at Front Door's edge locations. Omit the
block to disable caching entirely (every request hits the origin) --
Azure treats the ABSENCE of cache settings as caching off, so this
is a real switch, not a defaults bundle.

### spec.cache.queryStringCachingBehavior

`enum`

How query strings participate in the cache key. Unspecified deploys
IGNORE_QUERY_STRING (all variants share one entry -- right for
static assets). USE_QUERY_STRING keys every variant separately;
the *_SPECIFIED behaviors include or exclude only the names listed
in query_strings.

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `azure_front_door_route_query_string_caching_behavior_unspecified` -- Not specified -- deploys IGNORE_QUERY_STRING, Azure's default.
- `IGNORE_QUERY_STRING` -- Strip query strings from the cache key: every variant of a URL shares one cached entry. Right for static assets.
- `USE_QUERY_STRING` -- Key the cache on the full query string: each variant caches separately. Right for query-driven dynamic content.
- `IGNORE_SPECIFIED_QUERY_STRINGS` -- Ignore ONLY the parameters named in query_strings when building the cache key (e.g. drop tracking parameters like utm_source).
- `INCLUDE_SPECIFIED_QUERY_STRINGS` -- Include ONLY the parameters named in query_strings in the cache key (e.g. key on page= but nothing else).

### spec.cache.queryStrings

`[]string`

The query-string parameter NAMES the *_SPECIFIED behaviors operate
on (Azure ignores this list for the other two behaviors). Names
must not contain commas -- Azure transports the list as a
comma-separated string.

- rule: {"repeated":{"items":{"cel":[{"id":"front_door_route_cache_query_string_no_comma","message":"query string names must not contain commas","expression":"!this.contains(',')"}]}}}

### spec.cache.compressionEnabled

`bool` · optional (explicit presence)

Compress eligible responses (gzip/brotli) at the edge before
delivering to clients. Default false. Azure only compresses
responses between 1 KiB and 8 MiB whose content type is listed in
content_types_to_compress.

- default: `false`

### spec.cache.contentTypesToCompress

`[]string`

The MIME types eligible for edge compression (only meaningful with
compression_enabled). Values must come from Azure's supported list
-- text, JSON, XML, JavaScript, SVG, and font types; binary media
(images, video, archives) is already compressed and not eligible.

- rule: {"repeated":{"items":{"cel":[{"id":"front_door_route_cache_content_type_supported","message":"content type is not in Azure Front Door's supported compression list (text/*, application JSON/XML/JavaScript, fonts, and SVG variants)","expression":"this in ['application/eot', 'application/font', 'application/font-sfnt', 'application/javascript', 'application/json', 'application/opentype', 'application/otf', 'application/pkcs7-mime', 'application/truetype', 'application/ttf', 'application/vnd.ms-fontobject', 'application/xhtml+xml', 'application/xml', 'application/xml+rss', 'application/x-font-opentype', 'application/x-font-truetype', 'application/x-font-ttf', 'application/x-httpd-cgi', 'application/x-mpegurl', 'application/x-opentype', 'application/x-otf', 'application/x-perl', 'application/x-ttf', 'application/x-javascript', 'font/eot', 'font/ttf', 'font/otf', 'font/opentype', 'image/svg+xml', 'text/css', 'text/csv', 'text/html', 'text/javascript', 'text/js', 'text/plain', 'text/richtext', 'text/tab-separated-values', 'text/xml', 'text/x-script', 'text/x-component', 'text/x-java-source']"}]}}}

## Validation Rules

- `front_door_route_https_redirect_needs_both_protocols`: https_redirect_enabled requires supported_protocols to include both HTTP and HTTPS (the redirect needs HTTP to arrive and HTTPS to land on); either list both or disable the redirect
- `front_door_route_default_domain_or_custom_domains`: link_to_default_domain can only be disabled when the route serves at least one custom domain -- otherwise it would answer on no hostname at all

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AzureFrontDoorRoute, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.route_id` | `string` | The Azure Resource Manager ID of the route. Format: /subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.Cdn/profiles/{profile}/afdEndpoints/{endpoint}/routes/{name} |
| `status.outputs.route_name` | `string` | The route's name -- unique within its endpoint. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.endpointId` | AzureFrontDoorEndpoint | `status.outputs.endpoint_id` |
| `spec.originGroupId` | AzureFrontDoorOriginGroup | `status.outputs.origin_group_id` |
| `spec.originIds` | AzureFrontDoorOrigin | `status.outputs.origin_id` |
| `spec.ruleSetIds` | AzureFrontDoorRuleSet | `status.outputs.rule_set_id` |
| `spec.customDomainIds` | AzureFrontDoorCustomDomain | `status.outputs.custom_domain_id` |

## See Also

- [Overview](../README.md)
