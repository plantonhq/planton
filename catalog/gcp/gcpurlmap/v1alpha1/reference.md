# GcpUrlMap

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

GcpUrlMapSpec defines a Compute Engine URL map — the L7 routing brain of an
Application Load Balancer (and of Traffic Director meshes). A URL map
matches each request's host and path and decides what happens: send it to
a backend service or backend bucket, split it across weighted backends,
rewrite or redirect it, inject faults, or return a custom error page.

Routing is evaluated in this order:
  1. host_rules match the request Host header to a named path_matcher.
  2. Within that path_matcher, route_rules (priority-ordered, with rich
     header/query/path matching) are tried first, then path_rules (longest
     prefix), then the path_matcher's own default.
  3. If no host_rule matches, the URL map's top-level default applies.

At every level the "what to do" is exactly one of: a service (backend
service or bucket), a url_redirect, or a route_action (which can weight
across backends and rewrite/retry/mirror). Target proxies reference this URL
map; forwarding rules and addresses sit in front of the proxy.

One kind, two scopes. With region empty the map is GLOBAL (the global
external ALB, the cross-region internal ALB, Traffic Director); with region
set it is REGIONAL (the regional external ALB and the regional internal
ALB), routing only to regional backend services in that region. The two
scopes share the whole routing surface except Cloud CDN route caching,
custom error pages, stream-duration limits, and header-driven routing
tests, which exist only on the global map (and are rejected when region is
set); the regional map alone honors a path_template_rewrite in a path
matcher's default route action. A URL map cannot move between scopes.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpUrlMap
metadata:
  name: my-sample-url-map
spec:
  # GCP project that owns the URL map.
  # Omit to use the provider's default project.
  projectId:
    value: my-gcp-project-123

  # Cloud-side name; omit to default to metadata.name.
  urlMapName: web-routing

  description: Routes www.example.com traffic to the web backend

  # What a destroy may do: DELETE (default), PREVENT, or ABANDON.
  deletionPolicy: DELETE

  # Exactly one top-level default target. Here: send unmatched traffic to
  # the web backend service (reference a GcpBackendService or self-link).
  defaultService:
    value: https://www.googleapis.com/compute/v1/projects/my-gcp-project-123/global/backendServices/web-backend

  # A route action may accompany a plain service target when it carries
  # only sub-policies (no weighted split): here the default route gets an
  # edge-enforced time budget and deliberate retries.
  defaultRouteAction:
    timeout:
      seconds: 15
    retryPolicy:
      numRetries: 2
      retryConditions:
        - 5xx
        - connect-failure
      perTryTimeout:
        seconds: 5

  # Host-based fan-out: map hosts to named path matchers.
  hostRules:
    - hosts:
        - www.example.com
        - example.com
      pathMatcher: main
      description: Primary site hosts

  pathMatchers:
    - name: main
      defaultService:
        value: https://www.googleapis.com/compute/v1/projects/my-gcp-project-123/global/backendServices/web-backend
      pathRules:
        - paths:
            - /api/*
          service:
            value: https://www.googleapis.com/compute/v1/projects/my-gcp-project-123/global/backendServices/api-backend
        - paths:
            - /static/*
          service:
            value: https://www.googleapis.com/compute/v1/projects/my-gcp-project-123/global/backendBuckets/static-assets

  # Routing self-test: GCP evaluates these at create/update time.
  tests:
    - host: www.example.com
      path: /api/health
      service:
        value: https://www.googleapis.com/compute/v1/projects/my-gcp-project-123/global/backendServices/api-backend
      description: API prefix resolves to the API backend
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.urlMapName` | `string` |  |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.region` | `string` |  |  |  |
| `spec.defaultService` | `string \| valueFrom` |  |  |  |
| `spec.defaultUrlRedirect` | `GcpUrlMapUrlRedirect` |  |  |  |
| `spec.defaultUrlRedirect.hostRedirect` | `string` |  |  |  |
| `spec.defaultUrlRedirect.httpsRedirect` | `bool` |  |  |  |
| `spec.defaultUrlRedirect.pathRedirect` | `string` |  |  |  |
| `spec.defaultUrlRedirect.prefixRedirect` | `string` |  |  |  |
| `spec.defaultUrlRedirect.redirectResponseCode` | `string` |  |  |  |
| `spec.defaultUrlRedirect.stripQuery` | `bool` |  |  |  |
| `spec.defaultRouteAction` | `GcpUrlMapRouteAction` |  |  |  |
| `spec.defaultRouteAction.weightedBackendServices` | `[]GcpUrlMapWeightedBackendService` |  |  |  |
| `spec.defaultRouteAction.weightedBackendServices[].backendService` | `string \| valueFrom` | yes |  | GcpBackendService (`status.outputs.self_link`) |
| `spec.defaultRouteAction.weightedBackendServices[].weight` | `int32` |  |  |  |
| `spec.defaultRouteAction.weightedBackendServices[].headerAction` | `GcpUrlMapHeaderAction` |  |  |  |
| `spec.defaultRouteAction.weightedBackendServices[].headerAction.requestHeadersToAdd` | `[]GcpUrlMapHeaderValue` |  |  |  |
| `spec.defaultRouteAction.weightedBackendServices[].headerAction.requestHeadersToAdd[].headerName` | `string` | yes |  |  |
| `spec.defaultRouteAction.weightedBackendServices[].headerAction.requestHeadersToAdd[].headerValue` | `string` | yes |  |  |
| `spec.defaultRouteAction.weightedBackendServices[].headerAction.requestHeadersToAdd[].replace` | `bool` |  |  |  |
| `spec.defaultRouteAction.weightedBackendServices[].headerAction.requestHeadersToRemove` | `[]string` |  |  |  |
| `spec.defaultRouteAction.weightedBackendServices[].headerAction.responseHeadersToAdd` | `[]GcpUrlMapHeaderValue` |  |  |  |
| `spec.defaultRouteAction.weightedBackendServices[].headerAction.responseHeadersToAdd[].headerName` | `string` | yes |  |  |
| `spec.defaultRouteAction.weightedBackendServices[].headerAction.responseHeadersToAdd[].headerValue` | `string` | yes |  |  |
| `spec.defaultRouteAction.weightedBackendServices[].headerAction.responseHeadersToAdd[].replace` | `bool` |  |  |  |
| `spec.defaultRouteAction.weightedBackendServices[].headerAction.responseHeadersToRemove` | `[]string` |  |  |  |
| `spec.defaultRouteAction.urlRewrite` | `GcpUrlMapUrlRewrite` |  |  |  |
| `spec.defaultRouteAction.urlRewrite.hostRewrite` | `string` |  |  |  |
| `spec.defaultRouteAction.urlRewrite.pathPrefixRewrite` | `string` |  |  |  |
| `spec.defaultRouteAction.urlRewrite.pathTemplateRewrite` | `string` |  |  |  |
| `spec.defaultRouteAction.timeout` | `GcpUrlMapDuration` |  |  |  |
| `spec.defaultRouteAction.timeout.seconds` | `int64` |  |  |  |
| `spec.defaultRouteAction.timeout.nanos` | `int32` |  |  |  |
| `spec.defaultRouteAction.retryPolicy` | `GcpUrlMapRetryPolicy` |  |  |  |
| `spec.defaultRouteAction.retryPolicy.numRetries` | `int32` |  |  |  |
| `spec.defaultRouteAction.retryPolicy.retryConditions` | `[]string` |  |  |  |
| `spec.defaultRouteAction.retryPolicy.perTryTimeout` | `GcpUrlMapDuration` |  |  |  |
| `spec.defaultRouteAction.retryPolicy.perTryTimeout.seconds` | `int64` |  |  |  |
| `spec.defaultRouteAction.retryPolicy.perTryTimeout.nanos` | `int32` |  |  |  |
| `spec.defaultRouteAction.requestMirrorPolicy` | `GcpUrlMapRequestMirrorPolicy` |  |  |  |
| `spec.defaultRouteAction.requestMirrorPolicy.backendService` | `string \| valueFrom` | yes |  | GcpBackendService (`status.outputs.self_link`) |
| `spec.defaultRouteAction.corsPolicy` | `GcpUrlMapCorsPolicy` |  |  |  |
| `spec.defaultRouteAction.corsPolicy.allowCredentials` | `bool` |  |  |  |
| `spec.defaultRouteAction.corsPolicy.allowHeaders` | `[]string` |  |  |  |
| `spec.defaultRouteAction.corsPolicy.allowMethods` | `[]string` |  |  |  |
| `spec.defaultRouteAction.corsPolicy.allowOriginRegexes` | `[]string` |  |  |  |
| `spec.defaultRouteAction.corsPolicy.allowOrigins` | `[]string` |  |  |  |
| `spec.defaultRouteAction.corsPolicy.disabled` | `bool` |  |  |  |
| `spec.defaultRouteAction.corsPolicy.exposeHeaders` | `[]string` |  |  |  |
| `spec.defaultRouteAction.corsPolicy.maxAge` | `int32` |  |  |  |
| `spec.defaultRouteAction.faultInjectionPolicy` | `GcpUrlMapFaultInjectionPolicy` |  |  |  |
| `spec.defaultRouteAction.faultInjectionPolicy.abort` | `GcpUrlMapFaultAbort` |  |  |  |
| `spec.defaultRouteAction.faultInjectionPolicy.abort.httpStatus` | `int32` |  |  |  |
| `spec.defaultRouteAction.faultInjectionPolicy.abort.percentage` | `double` |  |  |  |
| `spec.defaultRouteAction.faultInjectionPolicy.delay` | `GcpUrlMapFaultDelay` |  |  |  |
| `spec.defaultRouteAction.faultInjectionPolicy.delay.fixedDelay` | `GcpUrlMapDuration` |  |  |  |
| `spec.defaultRouteAction.faultInjectionPolicy.delay.fixedDelay.seconds` | `int64` |  |  |  |
| `spec.defaultRouteAction.faultInjectionPolicy.delay.fixedDelay.nanos` | `int32` |  |  |  |
| `spec.defaultRouteAction.faultInjectionPolicy.delay.percentage` | `double` |  |  |  |
| `spec.defaultRouteAction.maxStreamDuration` | `GcpUrlMapDuration` |  |  |  |
| `spec.defaultRouteAction.maxStreamDuration.seconds` | `int64` |  |  |  |
| `spec.defaultRouteAction.maxStreamDuration.nanos` | `int32` |  |  |  |
| `spec.defaultRouteAction.cachePolicy` | `GcpUrlMapCachePolicy` |  |  |  |
| `spec.defaultRouteAction.cachePolicy.cacheMode` | `string` |  |  |  |
| `spec.defaultRouteAction.cachePolicy.cacheBypassRequestHeaderNames` | `[]string` |  |  |  |
| `spec.defaultRouteAction.cachePolicy.negativeCaching` | `bool` |  |  |  |
| `spec.defaultRouteAction.cachePolicy.requestCoalescing` | `bool` |  |  |  |
| `spec.defaultRouteAction.cachePolicy.cacheKeyPolicy` | `GcpUrlMapCacheKeyPolicy` |  |  |  |
| `spec.defaultRouteAction.cachePolicy.cacheKeyPolicy.excludedQueryParameters` | `[]string` |  |  |  |
| `spec.defaultRouteAction.cachePolicy.cacheKeyPolicy.includeHost` | `bool` |  |  |  |
| `spec.defaultRouteAction.cachePolicy.cacheKeyPolicy.includeProtocol` | `bool` |  |  |  |
| `spec.defaultRouteAction.cachePolicy.cacheKeyPolicy.includeQueryString` | `bool` |  |  |  |
| `spec.defaultRouteAction.cachePolicy.cacheKeyPolicy.includedCookieNames` | `[]string` |  |  |  |
| `spec.defaultRouteAction.cachePolicy.cacheKeyPolicy.includedHeaderNames` | `[]string` |  |  |  |
| `spec.defaultRouteAction.cachePolicy.cacheKeyPolicy.includedQueryParameters` | `[]string` |  |  |  |
| `spec.defaultRouteAction.cachePolicy.clientTtl` | `GcpUrlMapDuration` |  |  |  |
| `spec.defaultRouteAction.cachePolicy.clientTtl.seconds` | `int64` |  |  |  |
| `spec.defaultRouteAction.cachePolicy.clientTtl.nanos` | `int32` |  |  |  |
| `spec.defaultRouteAction.cachePolicy.defaultTtl` | `GcpUrlMapDuration` |  |  |  |
| `spec.defaultRouteAction.cachePolicy.defaultTtl.seconds` | `int64` |  |  |  |
| `spec.defaultRouteAction.cachePolicy.defaultTtl.nanos` | `int32` |  |  |  |
| `spec.defaultRouteAction.cachePolicy.maxTtl` | `GcpUrlMapDuration` |  |  |  |
| `spec.defaultRouteAction.cachePolicy.maxTtl.seconds` | `int64` |  |  |  |
| `spec.defaultRouteAction.cachePolicy.maxTtl.nanos` | `int32` |  |  |  |
| `spec.defaultRouteAction.cachePolicy.serveWhileStale` | `GcpUrlMapDuration` |  |  |  |
| `spec.defaultRouteAction.cachePolicy.serveWhileStale.seconds` | `int64` |  |  |  |
| `spec.defaultRouteAction.cachePolicy.serveWhileStale.nanos` | `int32` |  |  |  |
| `spec.defaultRouteAction.cachePolicy.negativeCachingPolicy` | `[]GcpUrlMapNegativeCachingPolicy` |  |  |  |
| `spec.defaultRouteAction.cachePolicy.negativeCachingPolicy[].code` | `int32` |  |  |  |
| `spec.defaultRouteAction.cachePolicy.negativeCachingPolicy[].ttl` | `GcpUrlMapDuration` |  |  |  |
| `spec.defaultRouteAction.cachePolicy.negativeCachingPolicy[].ttl.seconds` | `int64` |  |  |  |
| `spec.defaultRouteAction.cachePolicy.negativeCachingPolicy[].ttl.nanos` | `int32` |  |  |  |
| `spec.defaultCustomErrorResponsePolicy` | `GcpUrlMapCustomErrorResponsePolicy` |  |  |  |
| `spec.defaultCustomErrorResponsePolicy.errorService` | `string \| valueFrom` |  |  | GcpBackendBucket (`status.outputs.self_link`) |
| `spec.defaultCustomErrorResponsePolicy.errorResponseRules` | `[]GcpUrlMapCustomErrorResponseRule` |  |  |  |
| `spec.defaultCustomErrorResponsePolicy.errorResponseRules[].matchResponseCodes` | `[]string` | yes |  |  |
| `spec.defaultCustomErrorResponsePolicy.errorResponseRules[].overrideResponseCode` | `int32` |  |  |  |
| `spec.defaultCustomErrorResponsePolicy.errorResponseRules[].path` | `string` | yes |  |  |
| `spec.headerAction` | `GcpUrlMapHeaderAction` |  |  |  |
| `spec.headerAction.requestHeadersToAdd` | `[]GcpUrlMapHeaderValue` |  |  |  |
| `spec.headerAction.requestHeadersToAdd[].headerName` | `string` | yes |  |  |
| `spec.headerAction.requestHeadersToAdd[].headerValue` | `string` | yes |  |  |
| `spec.headerAction.requestHeadersToAdd[].replace` | `bool` |  |  |  |
| `spec.headerAction.requestHeadersToRemove` | `[]string` |  |  |  |
| `spec.headerAction.responseHeadersToAdd` | `[]GcpUrlMapHeaderValue` |  |  |  |
| `spec.headerAction.responseHeadersToAdd[].headerName` | `string` | yes |  |  |
| `spec.headerAction.responseHeadersToAdd[].headerValue` | `string` | yes |  |  |
| `spec.headerAction.responseHeadersToAdd[].replace` | `bool` |  |  |  |
| `spec.headerAction.responseHeadersToRemove` | `[]string` |  |  |  |
| `spec.hostRules` | `[]GcpUrlMapHostRule` |  |  |  |
| `spec.hostRules[].hosts` | `[]string` | yes |  |  |
| `spec.hostRules[].pathMatcher` | `string` | yes |  |  |
| `spec.hostRules[].description` | `string` |  |  |  |
| `spec.pathMatchers` | `[]GcpUrlMapPathMatcher` |  |  |  |
| `spec.pathMatchers[].name` | `string` | yes |  |  |
| `spec.pathMatchers[].defaultService` | `string \| valueFrom` |  |  |  |
| `spec.pathMatchers[].defaultUrlRedirect` | `GcpUrlMapUrlRedirect` |  |  |  |
| `spec.pathMatchers[].defaultUrlRedirect.hostRedirect` | `string` |  |  |  |
| `spec.pathMatchers[].defaultUrlRedirect.httpsRedirect` | `bool` |  |  |  |
| `spec.pathMatchers[].defaultUrlRedirect.pathRedirect` | `string` |  |  |  |
| `spec.pathMatchers[].defaultUrlRedirect.prefixRedirect` | `string` |  |  |  |
| `spec.pathMatchers[].defaultUrlRedirect.redirectResponseCode` | `string` |  |  |  |
| `spec.pathMatchers[].defaultUrlRedirect.stripQuery` | `bool` |  |  |  |
| `spec.pathMatchers[].defaultRouteAction` | `GcpUrlMapRouteAction` |  |  |  |
| `spec.pathMatchers[].defaultRouteAction.weightedBackendServices` | `[]GcpUrlMapWeightedBackendService` |  |  |  |
| `spec.pathMatchers[].defaultRouteAction.weightedBackendServices[].backendService` | `string \| valueFrom` | yes |  | GcpBackendService (`status.outputs.self_link`) |
| `spec.pathMatchers[].defaultRouteAction.weightedBackendServices[].weight` | `int32` |  |  |  |
| `spec.pathMatchers[].defaultRouteAction.weightedBackendServices[].headerAction` | `GcpUrlMapHeaderAction` |  |  |  |
| `spec.pathMatchers[].defaultRouteAction.weightedBackendServices[].headerAction.requestHeadersToAdd` | `[]GcpUrlMapHeaderValue` |  |  |  |
| `spec.pathMatchers[].defaultRouteAction.weightedBackendServices[].headerAction.requestHeadersToAdd[].headerName` | `string` | yes |  |  |
| `spec.pathMatchers[].defaultRouteAction.weightedBackendServices[].headerAction.requestHeadersToAdd[].headerValue` | `string` | yes |  |  |
| `spec.pathMatchers[].defaultRouteAction.weightedBackendServices[].headerAction.requestHeadersToAdd[].replace` | `bool` |  |  |  |
| `spec.pathMatchers[].defaultRouteAction.weightedBackendServices[].headerAction.requestHeadersToRemove` | `[]string` |  |  |  |
| `spec.pathMatchers[].defaultRouteAction.weightedBackendServices[].headerAction.responseHeadersToAdd` | `[]GcpUrlMapHeaderValue` |  |  |  |
| `spec.pathMatchers[].defaultRouteAction.weightedBackendServices[].headerAction.responseHeadersToAdd[].headerName` | `string` | yes |  |  |
| `spec.pathMatchers[].defaultRouteAction.weightedBackendServices[].headerAction.responseHeadersToAdd[].headerValue` | `string` | yes |  |  |
| `spec.pathMatchers[].defaultRouteAction.weightedBackendServices[].headerAction.responseHeadersToAdd[].replace` | `bool` |  |  |  |
| `spec.pathMatchers[].defaultRouteAction.weightedBackendServices[].headerAction.responseHeadersToRemove` | `[]string` |  |  |  |
| `spec.pathMatchers[].defaultRouteAction.urlRewrite` | `GcpUrlMapUrlRewrite` |  |  |  |
| `spec.pathMatchers[].defaultRouteAction.urlRewrite.hostRewrite` | `string` |  |  |  |
| `spec.pathMatchers[].defaultRouteAction.urlRewrite.pathPrefixRewrite` | `string` |  |  |  |
| `spec.pathMatchers[].defaultRouteAction.urlRewrite.pathTemplateRewrite` | `string` |  |  |  |
| `spec.pathMatchers[].defaultRouteAction.timeout` | `GcpUrlMapDuration` |  |  |  |
| `spec.pathMatchers[].defaultRouteAction.timeout.seconds` | `int64` |  |  |  |
| `spec.pathMatchers[].defaultRouteAction.timeout.nanos` | `int32` |  |  |  |
| `spec.pathMatchers[].defaultRouteAction.retryPolicy` | `GcpUrlMapRetryPolicy` |  |  |  |
| `spec.pathMatchers[].defaultRouteAction.retryPolicy.numRetries` | `int32` |  |  |  |
| `spec.pathMatchers[].defaultRouteAction.retryPolicy.retryConditions` | `[]string` |  |  |  |
| `spec.pathMatchers[].defaultRouteAction.retryPolicy.perTryTimeout` | `GcpUrlMapDuration` |  |  |  |
| `spec.pathMatchers[].defaultRouteAction.retryPolicy.perTryTimeout.seconds` | `int64` |  |  |  |
| `spec.pathMatchers[].defaultRouteAction.retryPolicy.perTryTimeout.nanos` | `int32` |  |  |  |
| `spec.pathMatchers[].defaultRouteAction.requestMirrorPolicy` | `GcpUrlMapRequestMirrorPolicy` |  |  |  |
| `spec.pathMatchers[].defaultRouteAction.requestMirrorPolicy.backendService` | `string \| valueFrom` | yes |  | GcpBackendService (`status.outputs.self_link`) |
| `spec.pathMatchers[].defaultRouteAction.corsPolicy` | `GcpUrlMapCorsPolicy` |  |  |  |
| `spec.pathMatchers[].defaultRouteAction.corsPolicy.allowCredentials` | `bool` |  |  |  |
| `spec.pathMatchers[].defaultRouteAction.corsPolicy.allowHeaders` | `[]string` |  |  |  |
| `spec.pathMatchers[].defaultRouteAction.corsPolicy.allowMethods` | `[]string` |  |  |  |
| `spec.pathMatchers[].defaultRouteAction.corsPolicy.allowOriginRegexes` | `[]string` |  |  |  |
| `spec.pathMatchers[].defaultRouteAction.corsPolicy.allowOrigins` | `[]string` |  |  |  |
| `spec.pathMatchers[].defaultRouteAction.corsPolicy.disabled` | `bool` |  |  |  |
| `spec.pathMatchers[].defaultRouteAction.corsPolicy.exposeHeaders` | `[]string` |  |  |  |
| `spec.pathMatchers[].defaultRouteAction.corsPolicy.maxAge` | `int32` |  |  |  |
| `spec.pathMatchers[].defaultRouteAction.faultInjectionPolicy` | `GcpUrlMapFaultInjectionPolicy` |  |  |  |
| `spec.pathMatchers[].defaultRouteAction.faultInjectionPolicy.abort` | `GcpUrlMapFaultAbort` |  |  |  |
| `spec.pathMatchers[].defaultRouteAction.faultInjectionPolicy.abort.httpStatus` | `int32` |  |  |  |
| `spec.pathMatchers[].defaultRouteAction.faultInjectionPolicy.abort.percentage` | `double` |  |  |  |
| `spec.pathMatchers[].defaultRouteAction.faultInjectionPolicy.delay` | `GcpUrlMapFaultDelay` |  |  |  |
| `spec.pathMatchers[].defaultRouteAction.faultInjectionPolicy.delay.fixedDelay` | `GcpUrlMapDuration` |  |  |  |
| `spec.pathMatchers[].defaultRouteAction.faultInjectionPolicy.delay.fixedDelay.seconds` | `int64` |  |  |  |
| `spec.pathMatchers[].defaultRouteAction.faultInjectionPolicy.delay.fixedDelay.nanos` | `int32` |  |  |  |
| `spec.pathMatchers[].defaultRouteAction.faultInjectionPolicy.delay.percentage` | `double` |  |  |  |
| `spec.pathMatchers[].defaultRouteAction.maxStreamDuration` | `GcpUrlMapDuration` |  |  |  |
| `spec.pathMatchers[].defaultRouteAction.maxStreamDuration.seconds` | `int64` |  |  |  |
| `spec.pathMatchers[].defaultRouteAction.maxStreamDuration.nanos` | `int32` |  |  |  |
| `spec.pathMatchers[].defaultRouteAction.cachePolicy` | `GcpUrlMapCachePolicy` |  |  |  |
| `spec.pathMatchers[].defaultRouteAction.cachePolicy.cacheMode` | `string` |  |  |  |
| `spec.pathMatchers[].defaultRouteAction.cachePolicy.cacheBypassRequestHeaderNames` | `[]string` |  |  |  |
| `spec.pathMatchers[].defaultRouteAction.cachePolicy.negativeCaching` | `bool` |  |  |  |
| `spec.pathMatchers[].defaultRouteAction.cachePolicy.requestCoalescing` | `bool` |  |  |  |
| `spec.pathMatchers[].defaultRouteAction.cachePolicy.cacheKeyPolicy` | `GcpUrlMapCacheKeyPolicy` |  |  |  |
| `spec.pathMatchers[].defaultRouteAction.cachePolicy.cacheKeyPolicy.excludedQueryParameters` | `[]string` |  |  |  |
| `spec.pathMatchers[].defaultRouteAction.cachePolicy.cacheKeyPolicy.includeHost` | `bool` |  |  |  |
| `spec.pathMatchers[].defaultRouteAction.cachePolicy.cacheKeyPolicy.includeProtocol` | `bool` |  |  |  |
| `spec.pathMatchers[].defaultRouteAction.cachePolicy.cacheKeyPolicy.includeQueryString` | `bool` |  |  |  |
| `spec.pathMatchers[].defaultRouteAction.cachePolicy.cacheKeyPolicy.includedCookieNames` | `[]string` |  |  |  |
| `spec.pathMatchers[].defaultRouteAction.cachePolicy.cacheKeyPolicy.includedHeaderNames` | `[]string` |  |  |  |
| `spec.pathMatchers[].defaultRouteAction.cachePolicy.cacheKeyPolicy.includedQueryParameters` | `[]string` |  |  |  |
| `spec.pathMatchers[].defaultRouteAction.cachePolicy.clientTtl` | `GcpUrlMapDuration` |  |  |  |
| `spec.pathMatchers[].defaultRouteAction.cachePolicy.clientTtl.seconds` | `int64` |  |  |  |
| `spec.pathMatchers[].defaultRouteAction.cachePolicy.clientTtl.nanos` | `int32` |  |  |  |
| `spec.pathMatchers[].defaultRouteAction.cachePolicy.defaultTtl` | `GcpUrlMapDuration` |  |  |  |
| `spec.pathMatchers[].defaultRouteAction.cachePolicy.defaultTtl.seconds` | `int64` |  |  |  |
| `spec.pathMatchers[].defaultRouteAction.cachePolicy.defaultTtl.nanos` | `int32` |  |  |  |
| `spec.pathMatchers[].defaultRouteAction.cachePolicy.maxTtl` | `GcpUrlMapDuration` |  |  |  |
| `spec.pathMatchers[].defaultRouteAction.cachePolicy.maxTtl.seconds` | `int64` |  |  |  |
| `spec.pathMatchers[].defaultRouteAction.cachePolicy.maxTtl.nanos` | `int32` |  |  |  |
| `spec.pathMatchers[].defaultRouteAction.cachePolicy.serveWhileStale` | `GcpUrlMapDuration` |  |  |  |
| `spec.pathMatchers[].defaultRouteAction.cachePolicy.serveWhileStale.seconds` | `int64` |  |  |  |
| `spec.pathMatchers[].defaultRouteAction.cachePolicy.serveWhileStale.nanos` | `int32` |  |  |  |
| `spec.pathMatchers[].defaultRouteAction.cachePolicy.negativeCachingPolicy` | `[]GcpUrlMapNegativeCachingPolicy` |  |  |  |
| `spec.pathMatchers[].defaultRouteAction.cachePolicy.negativeCachingPolicy[].code` | `int32` |  |  |  |
| `spec.pathMatchers[].defaultRouteAction.cachePolicy.negativeCachingPolicy[].ttl` | `GcpUrlMapDuration` |  |  |  |
| `spec.pathMatchers[].defaultRouteAction.cachePolicy.negativeCachingPolicy[].ttl.seconds` | `int64` |  |  |  |
| `spec.pathMatchers[].defaultRouteAction.cachePolicy.negativeCachingPolicy[].ttl.nanos` | `int32` |  |  |  |
| `spec.pathMatchers[].defaultCustomErrorResponsePolicy` | `GcpUrlMapCustomErrorResponsePolicy` |  |  |  |
| `spec.pathMatchers[].defaultCustomErrorResponsePolicy.errorService` | `string \| valueFrom` |  |  | GcpBackendBucket (`status.outputs.self_link`) |
| `spec.pathMatchers[].defaultCustomErrorResponsePolicy.errorResponseRules` | `[]GcpUrlMapCustomErrorResponseRule` |  |  |  |
| `spec.pathMatchers[].defaultCustomErrorResponsePolicy.errorResponseRules[].matchResponseCodes` | `[]string` | yes |  |  |
| `spec.pathMatchers[].defaultCustomErrorResponsePolicy.errorResponseRules[].overrideResponseCode` | `int32` |  |  |  |
| `spec.pathMatchers[].defaultCustomErrorResponsePolicy.errorResponseRules[].path` | `string` | yes |  |  |
| `spec.pathMatchers[].description` | `string` |  |  |  |
| `spec.pathMatchers[].headerAction` | `GcpUrlMapHeaderAction` |  |  |  |
| `spec.pathMatchers[].headerAction.requestHeadersToAdd` | `[]GcpUrlMapHeaderValue` |  |  |  |
| `spec.pathMatchers[].headerAction.requestHeadersToAdd[].headerName` | `string` | yes |  |  |
| `spec.pathMatchers[].headerAction.requestHeadersToAdd[].headerValue` | `string` | yes |  |  |
| `spec.pathMatchers[].headerAction.requestHeadersToAdd[].replace` | `bool` |  |  |  |
| `spec.pathMatchers[].headerAction.requestHeadersToRemove` | `[]string` |  |  |  |
| `spec.pathMatchers[].headerAction.responseHeadersToAdd` | `[]GcpUrlMapHeaderValue` |  |  |  |
| `spec.pathMatchers[].headerAction.responseHeadersToAdd[].headerName` | `string` | yes |  |  |
| `spec.pathMatchers[].headerAction.responseHeadersToAdd[].headerValue` | `string` | yes |  |  |
| `spec.pathMatchers[].headerAction.responseHeadersToAdd[].replace` | `bool` |  |  |  |
| `spec.pathMatchers[].headerAction.responseHeadersToRemove` | `[]string` |  |  |  |
| `spec.pathMatchers[].pathRules` | `[]GcpUrlMapPathRule` |  |  |  |
| `spec.pathMatchers[].pathRules[].paths` | `[]string` | yes |  |  |
| `spec.pathMatchers[].pathRules[].service` | `string \| valueFrom` |  |  |  |
| `spec.pathMatchers[].pathRules[].routeAction` | `GcpUrlMapRouteAction` |  |  |  |
| `spec.pathMatchers[].pathRules[].routeAction.weightedBackendServices` | `[]GcpUrlMapWeightedBackendService` |  |  |  |
| `spec.pathMatchers[].pathRules[].routeAction.weightedBackendServices[].backendService` | `string \| valueFrom` | yes |  | GcpBackendService (`status.outputs.self_link`) |
| `spec.pathMatchers[].pathRules[].routeAction.weightedBackendServices[].weight` | `int32` |  |  |  |
| `spec.pathMatchers[].pathRules[].routeAction.weightedBackendServices[].headerAction` | `GcpUrlMapHeaderAction` |  |  |  |
| `spec.pathMatchers[].pathRules[].routeAction.weightedBackendServices[].headerAction.requestHeadersToAdd` | `[]GcpUrlMapHeaderValue` |  |  |  |
| `spec.pathMatchers[].pathRules[].routeAction.weightedBackendServices[].headerAction.requestHeadersToAdd[].headerName` | `string` | yes |  |  |
| `spec.pathMatchers[].pathRules[].routeAction.weightedBackendServices[].headerAction.requestHeadersToAdd[].headerValue` | `string` | yes |  |  |
| `spec.pathMatchers[].pathRules[].routeAction.weightedBackendServices[].headerAction.requestHeadersToAdd[].replace` | `bool` |  |  |  |
| `spec.pathMatchers[].pathRules[].routeAction.weightedBackendServices[].headerAction.requestHeadersToRemove` | `[]string` |  |  |  |
| `spec.pathMatchers[].pathRules[].routeAction.weightedBackendServices[].headerAction.responseHeadersToAdd` | `[]GcpUrlMapHeaderValue` |  |  |  |
| `spec.pathMatchers[].pathRules[].routeAction.weightedBackendServices[].headerAction.responseHeadersToAdd[].headerName` | `string` | yes |  |  |
| `spec.pathMatchers[].pathRules[].routeAction.weightedBackendServices[].headerAction.responseHeadersToAdd[].headerValue` | `string` | yes |  |  |
| `spec.pathMatchers[].pathRules[].routeAction.weightedBackendServices[].headerAction.responseHeadersToAdd[].replace` | `bool` |  |  |  |
| `spec.pathMatchers[].pathRules[].routeAction.weightedBackendServices[].headerAction.responseHeadersToRemove` | `[]string` |  |  |  |
| `spec.pathMatchers[].pathRules[].routeAction.urlRewrite` | `GcpUrlMapUrlRewrite` |  |  |  |
| `spec.pathMatchers[].pathRules[].routeAction.urlRewrite.hostRewrite` | `string` |  |  |  |
| `spec.pathMatchers[].pathRules[].routeAction.urlRewrite.pathPrefixRewrite` | `string` |  |  |  |
| `spec.pathMatchers[].pathRules[].routeAction.urlRewrite.pathTemplateRewrite` | `string` |  |  |  |
| `spec.pathMatchers[].pathRules[].routeAction.timeout` | `GcpUrlMapDuration` |  |  |  |
| `spec.pathMatchers[].pathRules[].routeAction.timeout.seconds` | `int64` |  |  |  |
| `spec.pathMatchers[].pathRules[].routeAction.timeout.nanos` | `int32` |  |  |  |
| `spec.pathMatchers[].pathRules[].routeAction.retryPolicy` | `GcpUrlMapRetryPolicy` |  |  |  |
| `spec.pathMatchers[].pathRules[].routeAction.retryPolicy.numRetries` | `int32` |  |  |  |
| `spec.pathMatchers[].pathRules[].routeAction.retryPolicy.retryConditions` | `[]string` |  |  |  |
| `spec.pathMatchers[].pathRules[].routeAction.retryPolicy.perTryTimeout` | `GcpUrlMapDuration` |  |  |  |
| `spec.pathMatchers[].pathRules[].routeAction.retryPolicy.perTryTimeout.seconds` | `int64` |  |  |  |
| `spec.pathMatchers[].pathRules[].routeAction.retryPolicy.perTryTimeout.nanos` | `int32` |  |  |  |
| `spec.pathMatchers[].pathRules[].routeAction.requestMirrorPolicy` | `GcpUrlMapRequestMirrorPolicy` |  |  |  |
| `spec.pathMatchers[].pathRules[].routeAction.requestMirrorPolicy.backendService` | `string \| valueFrom` | yes |  | GcpBackendService (`status.outputs.self_link`) |
| `spec.pathMatchers[].pathRules[].routeAction.corsPolicy` | `GcpUrlMapCorsPolicy` |  |  |  |
| `spec.pathMatchers[].pathRules[].routeAction.corsPolicy.allowCredentials` | `bool` |  |  |  |
| `spec.pathMatchers[].pathRules[].routeAction.corsPolicy.allowHeaders` | `[]string` |  |  |  |
| `spec.pathMatchers[].pathRules[].routeAction.corsPolicy.allowMethods` | `[]string` |  |  |  |
| `spec.pathMatchers[].pathRules[].routeAction.corsPolicy.allowOriginRegexes` | `[]string` |  |  |  |
| `spec.pathMatchers[].pathRules[].routeAction.corsPolicy.allowOrigins` | `[]string` |  |  |  |
| `spec.pathMatchers[].pathRules[].routeAction.corsPolicy.disabled` | `bool` |  |  |  |
| `spec.pathMatchers[].pathRules[].routeAction.corsPolicy.exposeHeaders` | `[]string` |  |  |  |
| `spec.pathMatchers[].pathRules[].routeAction.corsPolicy.maxAge` | `int32` |  |  |  |
| `spec.pathMatchers[].pathRules[].routeAction.faultInjectionPolicy` | `GcpUrlMapFaultInjectionPolicy` |  |  |  |
| `spec.pathMatchers[].pathRules[].routeAction.faultInjectionPolicy.abort` | `GcpUrlMapFaultAbort` |  |  |  |
| `spec.pathMatchers[].pathRules[].routeAction.faultInjectionPolicy.abort.httpStatus` | `int32` |  |  |  |
| `spec.pathMatchers[].pathRules[].routeAction.faultInjectionPolicy.abort.percentage` | `double` |  |  |  |
| `spec.pathMatchers[].pathRules[].routeAction.faultInjectionPolicy.delay` | `GcpUrlMapFaultDelay` |  |  |  |
| `spec.pathMatchers[].pathRules[].routeAction.faultInjectionPolicy.delay.fixedDelay` | `GcpUrlMapDuration` |  |  |  |
| `spec.pathMatchers[].pathRules[].routeAction.faultInjectionPolicy.delay.fixedDelay.seconds` | `int64` |  |  |  |
| `spec.pathMatchers[].pathRules[].routeAction.faultInjectionPolicy.delay.fixedDelay.nanos` | `int32` |  |  |  |
| `spec.pathMatchers[].pathRules[].routeAction.faultInjectionPolicy.delay.percentage` | `double` |  |  |  |
| `spec.pathMatchers[].pathRules[].routeAction.maxStreamDuration` | `GcpUrlMapDuration` |  |  |  |
| `spec.pathMatchers[].pathRules[].routeAction.maxStreamDuration.seconds` | `int64` |  |  |  |
| `spec.pathMatchers[].pathRules[].routeAction.maxStreamDuration.nanos` | `int32` |  |  |  |
| `spec.pathMatchers[].pathRules[].routeAction.cachePolicy` | `GcpUrlMapCachePolicy` |  |  |  |
| `spec.pathMatchers[].pathRules[].routeAction.cachePolicy.cacheMode` | `string` |  |  |  |
| `spec.pathMatchers[].pathRules[].routeAction.cachePolicy.cacheBypassRequestHeaderNames` | `[]string` |  |  |  |
| `spec.pathMatchers[].pathRules[].routeAction.cachePolicy.negativeCaching` | `bool` |  |  |  |
| `spec.pathMatchers[].pathRules[].routeAction.cachePolicy.requestCoalescing` | `bool` |  |  |  |
| `spec.pathMatchers[].pathRules[].routeAction.cachePolicy.cacheKeyPolicy` | `GcpUrlMapCacheKeyPolicy` |  |  |  |
| `spec.pathMatchers[].pathRules[].routeAction.cachePolicy.cacheKeyPolicy.excludedQueryParameters` | `[]string` |  |  |  |
| `spec.pathMatchers[].pathRules[].routeAction.cachePolicy.cacheKeyPolicy.includeHost` | `bool` |  |  |  |
| `spec.pathMatchers[].pathRules[].routeAction.cachePolicy.cacheKeyPolicy.includeProtocol` | `bool` |  |  |  |
| `spec.pathMatchers[].pathRules[].routeAction.cachePolicy.cacheKeyPolicy.includeQueryString` | `bool` |  |  |  |
| `spec.pathMatchers[].pathRules[].routeAction.cachePolicy.cacheKeyPolicy.includedCookieNames` | `[]string` |  |  |  |
| `spec.pathMatchers[].pathRules[].routeAction.cachePolicy.cacheKeyPolicy.includedHeaderNames` | `[]string` |  |  |  |
| `spec.pathMatchers[].pathRules[].routeAction.cachePolicy.cacheKeyPolicy.includedQueryParameters` | `[]string` |  |  |  |
| `spec.pathMatchers[].pathRules[].routeAction.cachePolicy.clientTtl` | `GcpUrlMapDuration` |  |  |  |
| `spec.pathMatchers[].pathRules[].routeAction.cachePolicy.clientTtl.seconds` | `int64` |  |  |  |
| `spec.pathMatchers[].pathRules[].routeAction.cachePolicy.clientTtl.nanos` | `int32` |  |  |  |
| `spec.pathMatchers[].pathRules[].routeAction.cachePolicy.defaultTtl` | `GcpUrlMapDuration` |  |  |  |
| `spec.pathMatchers[].pathRules[].routeAction.cachePolicy.defaultTtl.seconds` | `int64` |  |  |  |
| `spec.pathMatchers[].pathRules[].routeAction.cachePolicy.defaultTtl.nanos` | `int32` |  |  |  |
| `spec.pathMatchers[].pathRules[].routeAction.cachePolicy.maxTtl` | `GcpUrlMapDuration` |  |  |  |
| `spec.pathMatchers[].pathRules[].routeAction.cachePolicy.maxTtl.seconds` | `int64` |  |  |  |
| `spec.pathMatchers[].pathRules[].routeAction.cachePolicy.maxTtl.nanos` | `int32` |  |  |  |
| `spec.pathMatchers[].pathRules[].routeAction.cachePolicy.serveWhileStale` | `GcpUrlMapDuration` |  |  |  |
| `spec.pathMatchers[].pathRules[].routeAction.cachePolicy.serveWhileStale.seconds` | `int64` |  |  |  |
| `spec.pathMatchers[].pathRules[].routeAction.cachePolicy.serveWhileStale.nanos` | `int32` |  |  |  |
| `spec.pathMatchers[].pathRules[].routeAction.cachePolicy.negativeCachingPolicy` | `[]GcpUrlMapNegativeCachingPolicy` |  |  |  |
| `spec.pathMatchers[].pathRules[].routeAction.cachePolicy.negativeCachingPolicy[].code` | `int32` |  |  |  |
| `spec.pathMatchers[].pathRules[].routeAction.cachePolicy.negativeCachingPolicy[].ttl` | `GcpUrlMapDuration` |  |  |  |
| `spec.pathMatchers[].pathRules[].routeAction.cachePolicy.negativeCachingPolicy[].ttl.seconds` | `int64` |  |  |  |
| `spec.pathMatchers[].pathRules[].routeAction.cachePolicy.negativeCachingPolicy[].ttl.nanos` | `int32` |  |  |  |
| `spec.pathMatchers[].pathRules[].urlRedirect` | `GcpUrlMapUrlRedirect` |  |  |  |
| `spec.pathMatchers[].pathRules[].urlRedirect.hostRedirect` | `string` |  |  |  |
| `spec.pathMatchers[].pathRules[].urlRedirect.httpsRedirect` | `bool` |  |  |  |
| `spec.pathMatchers[].pathRules[].urlRedirect.pathRedirect` | `string` |  |  |  |
| `spec.pathMatchers[].pathRules[].urlRedirect.prefixRedirect` | `string` |  |  |  |
| `spec.pathMatchers[].pathRules[].urlRedirect.redirectResponseCode` | `string` |  |  |  |
| `spec.pathMatchers[].pathRules[].urlRedirect.stripQuery` | `bool` |  |  |  |
| `spec.pathMatchers[].pathRules[].customErrorResponsePolicy` | `GcpUrlMapCustomErrorResponsePolicy` |  |  |  |
| `spec.pathMatchers[].pathRules[].customErrorResponsePolicy.errorService` | `string \| valueFrom` |  |  | GcpBackendBucket (`status.outputs.self_link`) |
| `spec.pathMatchers[].pathRules[].customErrorResponsePolicy.errorResponseRules` | `[]GcpUrlMapCustomErrorResponseRule` |  |  |  |
| `spec.pathMatchers[].pathRules[].customErrorResponsePolicy.errorResponseRules[].matchResponseCodes` | `[]string` | yes |  |  |
| `spec.pathMatchers[].pathRules[].customErrorResponsePolicy.errorResponseRules[].overrideResponseCode` | `int32` |  |  |  |
| `spec.pathMatchers[].pathRules[].customErrorResponsePolicy.errorResponseRules[].path` | `string` | yes |  |  |
| `spec.pathMatchers[].routeRules` | `[]GcpUrlMapRouteRule` |  |  |  |
| `spec.pathMatchers[].routeRules[].priority` | `int32` |  |  |  |
| `spec.pathMatchers[].routeRules[].service` | `string \| valueFrom` |  |  | GcpBackendService (`status.outputs.self_link`) |
| `spec.pathMatchers[].routeRules[].matchRules` | `[]GcpUrlMapRouteRuleMatchRule` | yes |  |  |
| `spec.pathMatchers[].routeRules[].matchRules[].prefixMatch` | `string` |  |  |  |
| `spec.pathMatchers[].routeRules[].matchRules[].fullPathMatch` | `string` |  |  |  |
| `spec.pathMatchers[].routeRules[].matchRules[].regexMatch` | `string` |  |  |  |
| `spec.pathMatchers[].routeRules[].matchRules[].pathTemplateMatch` | `string` |  |  |  |
| `spec.pathMatchers[].routeRules[].matchRules[].ignoreCase` | `bool` |  |  |  |
| `spec.pathMatchers[].routeRules[].matchRules[].headerMatches` | `[]GcpUrlMapHeaderMatch` |  |  |  |
| `spec.pathMatchers[].routeRules[].matchRules[].headerMatches[].headerName` | `string` | yes |  |  |
| `spec.pathMatchers[].routeRules[].matchRules[].headerMatches[].exactMatch` | `string` |  |  |  |
| `spec.pathMatchers[].routeRules[].matchRules[].headerMatches[].prefixMatch` | `string` |  |  |  |
| `spec.pathMatchers[].routeRules[].matchRules[].headerMatches[].suffixMatch` | `string` |  |  |  |
| `spec.pathMatchers[].routeRules[].matchRules[].headerMatches[].regexMatch` | `string` |  |  |  |
| `spec.pathMatchers[].routeRules[].matchRules[].headerMatches[].presentMatch` | `bool` |  |  |  |
| `spec.pathMatchers[].routeRules[].matchRules[].headerMatches[].rangeMatch` | `GcpUrlMapHeaderMatchRange` |  |  |  |
| `spec.pathMatchers[].routeRules[].matchRules[].headerMatches[].rangeMatch.rangeStart` | `int64` |  |  |  |
| `spec.pathMatchers[].routeRules[].matchRules[].headerMatches[].rangeMatch.rangeEnd` | `int64` |  |  |  |
| `spec.pathMatchers[].routeRules[].matchRules[].headerMatches[].invertMatch` | `bool` |  |  |  |
| `spec.pathMatchers[].routeRules[].matchRules[].queryParameterMatches` | `[]GcpUrlMapQueryParameterMatch` |  |  |  |
| `spec.pathMatchers[].routeRules[].matchRules[].queryParameterMatches[].name` | `string` | yes |  |  |
| `spec.pathMatchers[].routeRules[].matchRules[].queryParameterMatches[].exactMatch` | `string` |  |  |  |
| `spec.pathMatchers[].routeRules[].matchRules[].queryParameterMatches[].presentMatch` | `bool` |  |  |  |
| `spec.pathMatchers[].routeRules[].matchRules[].queryParameterMatches[].regexMatch` | `string` |  |  |  |
| `spec.pathMatchers[].routeRules[].matchRules[].metadataFilters` | `[]GcpUrlMapMetadataFilter` |  |  |  |
| `spec.pathMatchers[].routeRules[].matchRules[].metadataFilters[].filterMatchCriteria` | `string` | yes |  |  |
| `spec.pathMatchers[].routeRules[].matchRules[].metadataFilters[].filterLabels` | `[]GcpUrlMapMetadataFilterLabel` | yes |  |  |
| `spec.pathMatchers[].routeRules[].matchRules[].metadataFilters[].filterLabels[].name` | `string` | yes |  |  |
| `spec.pathMatchers[].routeRules[].matchRules[].metadataFilters[].filterLabels[].value` | `string` | yes |  |  |
| `spec.pathMatchers[].routeRules[].routeAction` | `GcpUrlMapRouteAction` |  |  |  |
| `spec.pathMatchers[].routeRules[].routeAction.weightedBackendServices` | `[]GcpUrlMapWeightedBackendService` |  |  |  |
| `spec.pathMatchers[].routeRules[].routeAction.weightedBackendServices[].backendService` | `string \| valueFrom` | yes |  | GcpBackendService (`status.outputs.self_link`) |
| `spec.pathMatchers[].routeRules[].routeAction.weightedBackendServices[].weight` | `int32` |  |  |  |
| `spec.pathMatchers[].routeRules[].routeAction.weightedBackendServices[].headerAction` | `GcpUrlMapHeaderAction` |  |  |  |
| `spec.pathMatchers[].routeRules[].routeAction.weightedBackendServices[].headerAction.requestHeadersToAdd` | `[]GcpUrlMapHeaderValue` |  |  |  |
| `spec.pathMatchers[].routeRules[].routeAction.weightedBackendServices[].headerAction.requestHeadersToAdd[].headerName` | `string` | yes |  |  |
| `spec.pathMatchers[].routeRules[].routeAction.weightedBackendServices[].headerAction.requestHeadersToAdd[].headerValue` | `string` | yes |  |  |
| `spec.pathMatchers[].routeRules[].routeAction.weightedBackendServices[].headerAction.requestHeadersToAdd[].replace` | `bool` |  |  |  |
| `spec.pathMatchers[].routeRules[].routeAction.weightedBackendServices[].headerAction.requestHeadersToRemove` | `[]string` |  |  |  |
| `spec.pathMatchers[].routeRules[].routeAction.weightedBackendServices[].headerAction.responseHeadersToAdd` | `[]GcpUrlMapHeaderValue` |  |  |  |
| `spec.pathMatchers[].routeRules[].routeAction.weightedBackendServices[].headerAction.responseHeadersToAdd[].headerName` | `string` | yes |  |  |
| `spec.pathMatchers[].routeRules[].routeAction.weightedBackendServices[].headerAction.responseHeadersToAdd[].headerValue` | `string` | yes |  |  |
| `spec.pathMatchers[].routeRules[].routeAction.weightedBackendServices[].headerAction.responseHeadersToAdd[].replace` | `bool` |  |  |  |
| `spec.pathMatchers[].routeRules[].routeAction.weightedBackendServices[].headerAction.responseHeadersToRemove` | `[]string` |  |  |  |
| `spec.pathMatchers[].routeRules[].routeAction.urlRewrite` | `GcpUrlMapUrlRewrite` |  |  |  |
| `spec.pathMatchers[].routeRules[].routeAction.urlRewrite.hostRewrite` | `string` |  |  |  |
| `spec.pathMatchers[].routeRules[].routeAction.urlRewrite.pathPrefixRewrite` | `string` |  |  |  |
| `spec.pathMatchers[].routeRules[].routeAction.urlRewrite.pathTemplateRewrite` | `string` |  |  |  |
| `spec.pathMatchers[].routeRules[].routeAction.timeout` | `GcpUrlMapDuration` |  |  |  |
| `spec.pathMatchers[].routeRules[].routeAction.timeout.seconds` | `int64` |  |  |  |
| `spec.pathMatchers[].routeRules[].routeAction.timeout.nanos` | `int32` |  |  |  |
| `spec.pathMatchers[].routeRules[].routeAction.retryPolicy` | `GcpUrlMapRetryPolicy` |  |  |  |
| `spec.pathMatchers[].routeRules[].routeAction.retryPolicy.numRetries` | `int32` |  |  |  |
| `spec.pathMatchers[].routeRules[].routeAction.retryPolicy.retryConditions` | `[]string` |  |  |  |
| `spec.pathMatchers[].routeRules[].routeAction.retryPolicy.perTryTimeout` | `GcpUrlMapDuration` |  |  |  |
| `spec.pathMatchers[].routeRules[].routeAction.retryPolicy.perTryTimeout.seconds` | `int64` |  |  |  |
| `spec.pathMatchers[].routeRules[].routeAction.retryPolicy.perTryTimeout.nanos` | `int32` |  |  |  |
| `spec.pathMatchers[].routeRules[].routeAction.requestMirrorPolicy` | `GcpUrlMapRequestMirrorPolicy` |  |  |  |
| `spec.pathMatchers[].routeRules[].routeAction.requestMirrorPolicy.backendService` | `string \| valueFrom` | yes |  | GcpBackendService (`status.outputs.self_link`) |
| `spec.pathMatchers[].routeRules[].routeAction.corsPolicy` | `GcpUrlMapCorsPolicy` |  |  |  |
| `spec.pathMatchers[].routeRules[].routeAction.corsPolicy.allowCredentials` | `bool` |  |  |  |
| `spec.pathMatchers[].routeRules[].routeAction.corsPolicy.allowHeaders` | `[]string` |  |  |  |
| `spec.pathMatchers[].routeRules[].routeAction.corsPolicy.allowMethods` | `[]string` |  |  |  |
| `spec.pathMatchers[].routeRules[].routeAction.corsPolicy.allowOriginRegexes` | `[]string` |  |  |  |
| `spec.pathMatchers[].routeRules[].routeAction.corsPolicy.allowOrigins` | `[]string` |  |  |  |
| `spec.pathMatchers[].routeRules[].routeAction.corsPolicy.disabled` | `bool` |  |  |  |
| `spec.pathMatchers[].routeRules[].routeAction.corsPolicy.exposeHeaders` | `[]string` |  |  |  |
| `spec.pathMatchers[].routeRules[].routeAction.corsPolicy.maxAge` | `int32` |  |  |  |
| `spec.pathMatchers[].routeRules[].routeAction.faultInjectionPolicy` | `GcpUrlMapFaultInjectionPolicy` |  |  |  |
| `spec.pathMatchers[].routeRules[].routeAction.faultInjectionPolicy.abort` | `GcpUrlMapFaultAbort` |  |  |  |
| `spec.pathMatchers[].routeRules[].routeAction.faultInjectionPolicy.abort.httpStatus` | `int32` |  |  |  |
| `spec.pathMatchers[].routeRules[].routeAction.faultInjectionPolicy.abort.percentage` | `double` |  |  |  |
| `spec.pathMatchers[].routeRules[].routeAction.faultInjectionPolicy.delay` | `GcpUrlMapFaultDelay` |  |  |  |
| `spec.pathMatchers[].routeRules[].routeAction.faultInjectionPolicy.delay.fixedDelay` | `GcpUrlMapDuration` |  |  |  |
| `spec.pathMatchers[].routeRules[].routeAction.faultInjectionPolicy.delay.fixedDelay.seconds` | `int64` |  |  |  |
| `spec.pathMatchers[].routeRules[].routeAction.faultInjectionPolicy.delay.fixedDelay.nanos` | `int32` |  |  |  |
| `spec.pathMatchers[].routeRules[].routeAction.faultInjectionPolicy.delay.percentage` | `double` |  |  |  |
| `spec.pathMatchers[].routeRules[].routeAction.maxStreamDuration` | `GcpUrlMapDuration` |  |  |  |
| `spec.pathMatchers[].routeRules[].routeAction.maxStreamDuration.seconds` | `int64` |  |  |  |
| `spec.pathMatchers[].routeRules[].routeAction.maxStreamDuration.nanos` | `int32` |  |  |  |
| `spec.pathMatchers[].routeRules[].routeAction.cachePolicy` | `GcpUrlMapCachePolicy` |  |  |  |
| `spec.pathMatchers[].routeRules[].routeAction.cachePolicy.cacheMode` | `string` |  |  |  |
| `spec.pathMatchers[].routeRules[].routeAction.cachePolicy.cacheBypassRequestHeaderNames` | `[]string` |  |  |  |
| `spec.pathMatchers[].routeRules[].routeAction.cachePolicy.negativeCaching` | `bool` |  |  |  |
| `spec.pathMatchers[].routeRules[].routeAction.cachePolicy.requestCoalescing` | `bool` |  |  |  |
| `spec.pathMatchers[].routeRules[].routeAction.cachePolicy.cacheKeyPolicy` | `GcpUrlMapCacheKeyPolicy` |  |  |  |
| `spec.pathMatchers[].routeRules[].routeAction.cachePolicy.cacheKeyPolicy.excludedQueryParameters` | `[]string` |  |  |  |
| `spec.pathMatchers[].routeRules[].routeAction.cachePolicy.cacheKeyPolicy.includeHost` | `bool` |  |  |  |
| `spec.pathMatchers[].routeRules[].routeAction.cachePolicy.cacheKeyPolicy.includeProtocol` | `bool` |  |  |  |
| `spec.pathMatchers[].routeRules[].routeAction.cachePolicy.cacheKeyPolicy.includeQueryString` | `bool` |  |  |  |
| `spec.pathMatchers[].routeRules[].routeAction.cachePolicy.cacheKeyPolicy.includedCookieNames` | `[]string` |  |  |  |
| `spec.pathMatchers[].routeRules[].routeAction.cachePolicy.cacheKeyPolicy.includedHeaderNames` | `[]string` |  |  |  |
| `spec.pathMatchers[].routeRules[].routeAction.cachePolicy.cacheKeyPolicy.includedQueryParameters` | `[]string` |  |  |  |
| `spec.pathMatchers[].routeRules[].routeAction.cachePolicy.clientTtl` | `GcpUrlMapDuration` |  |  |  |
| `spec.pathMatchers[].routeRules[].routeAction.cachePolicy.clientTtl.seconds` | `int64` |  |  |  |
| `spec.pathMatchers[].routeRules[].routeAction.cachePolicy.clientTtl.nanos` | `int32` |  |  |  |
| `spec.pathMatchers[].routeRules[].routeAction.cachePolicy.defaultTtl` | `GcpUrlMapDuration` |  |  |  |
| `spec.pathMatchers[].routeRules[].routeAction.cachePolicy.defaultTtl.seconds` | `int64` |  |  |  |
| `spec.pathMatchers[].routeRules[].routeAction.cachePolicy.defaultTtl.nanos` | `int32` |  |  |  |
| `spec.pathMatchers[].routeRules[].routeAction.cachePolicy.maxTtl` | `GcpUrlMapDuration` |  |  |  |
| `spec.pathMatchers[].routeRules[].routeAction.cachePolicy.maxTtl.seconds` | `int64` |  |  |  |
| `spec.pathMatchers[].routeRules[].routeAction.cachePolicy.maxTtl.nanos` | `int32` |  |  |  |
| `spec.pathMatchers[].routeRules[].routeAction.cachePolicy.serveWhileStale` | `GcpUrlMapDuration` |  |  |  |
| `spec.pathMatchers[].routeRules[].routeAction.cachePolicy.serveWhileStale.seconds` | `int64` |  |  |  |
| `spec.pathMatchers[].routeRules[].routeAction.cachePolicy.serveWhileStale.nanos` | `int32` |  |  |  |
| `spec.pathMatchers[].routeRules[].routeAction.cachePolicy.negativeCachingPolicy` | `[]GcpUrlMapNegativeCachingPolicy` |  |  |  |
| `spec.pathMatchers[].routeRules[].routeAction.cachePolicy.negativeCachingPolicy[].code` | `int32` |  |  |  |
| `spec.pathMatchers[].routeRules[].routeAction.cachePolicy.negativeCachingPolicy[].ttl` | `GcpUrlMapDuration` |  |  |  |
| `spec.pathMatchers[].routeRules[].routeAction.cachePolicy.negativeCachingPolicy[].ttl.seconds` | `int64` |  |  |  |
| `spec.pathMatchers[].routeRules[].routeAction.cachePolicy.negativeCachingPolicy[].ttl.nanos` | `int32` |  |  |  |
| `spec.pathMatchers[].routeRules[].urlRedirect` | `GcpUrlMapUrlRedirect` |  |  |  |
| `spec.pathMatchers[].routeRules[].urlRedirect.hostRedirect` | `string` |  |  |  |
| `spec.pathMatchers[].routeRules[].urlRedirect.httpsRedirect` | `bool` |  |  |  |
| `spec.pathMatchers[].routeRules[].urlRedirect.pathRedirect` | `string` |  |  |  |
| `spec.pathMatchers[].routeRules[].urlRedirect.prefixRedirect` | `string` |  |  |  |
| `spec.pathMatchers[].routeRules[].urlRedirect.redirectResponseCode` | `string` |  |  |  |
| `spec.pathMatchers[].routeRules[].urlRedirect.stripQuery` | `bool` |  |  |  |
| `spec.pathMatchers[].routeRules[].headerAction` | `GcpUrlMapHeaderAction` |  |  |  |
| `spec.pathMatchers[].routeRules[].headerAction.requestHeadersToAdd` | `[]GcpUrlMapHeaderValue` |  |  |  |
| `spec.pathMatchers[].routeRules[].headerAction.requestHeadersToAdd[].headerName` | `string` | yes |  |  |
| `spec.pathMatchers[].routeRules[].headerAction.requestHeadersToAdd[].headerValue` | `string` | yes |  |  |
| `spec.pathMatchers[].routeRules[].headerAction.requestHeadersToAdd[].replace` | `bool` |  |  |  |
| `spec.pathMatchers[].routeRules[].headerAction.requestHeadersToRemove` | `[]string` |  |  |  |
| `spec.pathMatchers[].routeRules[].headerAction.responseHeadersToAdd` | `[]GcpUrlMapHeaderValue` |  |  |  |
| `spec.pathMatchers[].routeRules[].headerAction.responseHeadersToAdd[].headerName` | `string` | yes |  |  |
| `spec.pathMatchers[].routeRules[].headerAction.responseHeadersToAdd[].headerValue` | `string` | yes |  |  |
| `spec.pathMatchers[].routeRules[].headerAction.responseHeadersToAdd[].replace` | `bool` |  |  |  |
| `spec.pathMatchers[].routeRules[].headerAction.responseHeadersToRemove` | `[]string` |  |  |  |
| `spec.pathMatchers[].routeRules[].customErrorResponsePolicy` | `GcpUrlMapCustomErrorResponsePolicy` |  |  |  |
| `spec.pathMatchers[].routeRules[].customErrorResponsePolicy.errorService` | `string \| valueFrom` |  |  | GcpBackendBucket (`status.outputs.self_link`) |
| `spec.pathMatchers[].routeRules[].customErrorResponsePolicy.errorResponseRules` | `[]GcpUrlMapCustomErrorResponseRule` |  |  |  |
| `spec.pathMatchers[].routeRules[].customErrorResponsePolicy.errorResponseRules[].matchResponseCodes` | `[]string` | yes |  |  |
| `spec.pathMatchers[].routeRules[].customErrorResponsePolicy.errorResponseRules[].overrideResponseCode` | `int32` |  |  |  |
| `spec.pathMatchers[].routeRules[].customErrorResponsePolicy.errorResponseRules[].path` | `string` | yes |  |  |
| `spec.tests` | `[]GcpUrlMapTest` |  |  |  |
| `spec.tests[].host` | `string` | yes |  |  |
| `spec.tests[].path` | `string` | yes |  |  |
| `spec.tests[].service` | `string \| valueFrom` |  |  |  |
| `spec.tests[].description` | `string` |  |  |  |
| `spec.tests[].expectedOutputUrl` | `string` |  |  |  |
| `spec.tests[].expectedRedirectResponseCode` | `int32` |  |  |  |
| `spec.tests[].headers` | `[]GcpUrlMapTestHeader` |  |  |  |
| `spec.tests[].headers[].name` | `string` | yes |  |  |
| `spec.tests[].headers[].value` | `string` | yes |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

The GCP project that owns the URL map.
Can be a literal project ID or a reference to a GcpProject resource.
If omitted, the provider's default project is used.
Immutable: changing it destroys and recreates the URL map.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.urlMapName

`string`

Name of the URL map in GCP. Must be 1-63 characters: lowercase letters,
digits, and hyphens; must start with a letter and end with a letter or
digit. If not specified, defaults to metadata.name.
Immutable: changing it destroys and recreates the URL map, briefly
breaking every target proxy that references the old self_link.

- rule: url_map_name must be RFC1035-compliant: 1-63 lowercase letters, digits, or hyphens; must start with a letter and end with a letter or digit

### spec.description

`string`

What this URL map fronts and how it routes — write it for the operator
reading a routing incident later. Mutable.

- rule: {"string":{"maxLen":"2048"}}

### spec.region

`string`

The scope selector. Empty builds a GLOBAL URL map (the global external
ALB, the cross-region internal ALB, Traffic Director); a region name
such as us-central1 builds a REGIONAL one (the regional external ALB and
the regional internal ALB). A regional map routes only to regional
backend services in its own region — never to a backend bucket, which
is a global-only resource — and is referenced only by regional target
proxies. Cloud CDN route caching (cache_policy), custom error pages
(custom_error_response_policy), max_stream_duration, and the
header-driven routing-test fields exist only on the global map and are
rejected when region is set. Immutable: a URL map cannot move between
scopes or regions.

- rule: region must be a valid GCP region name such as us-central1, or empty for a global URL map

### spec.defaultService

`string | valueFrom`

The default target when no host/path rule matches — a backend service or
backend bucket. Reference a GcpBackendService or GcpBackendBucket, or
provide a self-link directly (a regional map takes only a regional
GcpBackendService in its region). Exactly one of default_service,
default_url_redirect, or default_route_action must be set. Mutable.

- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.defaultUrlRedirect

`GcpUrlMapUrlRedirect`

Redirect unmatched requests instead of serving them (e.g. an
apex-to-www or http-to-https redirect as the catch-all). Mutually
exclusive with default_service and default_route_action.

- rule: path_redirect and prefix_redirect are mutually exclusive — set at most one

### spec.defaultUrlRedirect.hostRedirect

`string`

Replace the host in the redirect Location. Empty keeps the request host.

- rule: {"string":{"maxLen":"255"}}

### spec.defaultUrlRedirect.httpsRedirect

`bool`

Redirect to HTTPS (scheme becomes https). The standard http→https
upgrade. Default false.

### spec.defaultUrlRedirect.pathRedirect

`string`

Replace the entire path with this value. Mutually exclusive with
prefix_redirect. Empty keeps the request path.

- rule: {"string":{"maxLen":"1024"}}

### spec.defaultUrlRedirect.prefixRedirect

`string`

Replace the matched path prefix with this value, keeping the remainder.
Mutually exclusive with path_redirect.

- rule: {"string":{"maxLen":"1024"}}

### spec.defaultUrlRedirect.redirectResponseCode

`string`

The HTTP redirect status code: FOUND (302), MOVED_PERMANENTLY_DEFAULT
(301), PERMANENT_REDIRECT (308), SEE_OTHER (303), or TEMPORARY_REDIRECT
(307). Empty uses the GCP default (MOVED_PERMANENTLY_DEFAULT).

- rule: redirect_response_code must be one of FOUND, MOVED_PERMANENTLY_DEFAULT, PERMANENT_REDIRECT, SEE_OTHER, or TEMPORARY_REDIRECT

### spec.defaultUrlRedirect.stripQuery

`bool`

Drop the query string from the redirect Location. Default false (the query
string is preserved).

### spec.defaultRouteAction

`GcpUrlMapRouteAction`

Advanced default handling: weight traffic across several backends, rewrite
URLs, set timeouts/retries. Mutually exclusive with default_service and
default_url_redirect (its weighted_backend_services is the third arm of
the default-target choice).

### spec.defaultRouteAction.weightedBackendServices

`[]GcpUrlMapWeightedBackendService`

Split traffic across multiple backend services by weight — the mechanism
for weighted canary and blue/green rollouts. The weights are relative; a
backend's share is its weight over the sum of weights.

### spec.defaultRouteAction.weightedBackendServices[].backendService

`string | valueFrom` · required

The backend service receiving this share of traffic. Reference a
GcpBackendService or provide a self-link directly.

- references: GcpBackendService (`status.outputs.self_link`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpBackendService, name: <that resource's name>, fieldPath: status.outputs.self_link}} -- a bare string does not parse

### spec.defaultRouteAction.weightedBackendServices[].weight

`int32`

Relative weight of this backend (0-1000). Its share is weight over the sum
of all weights in the split; 0 drains this backend from the split.

- rule: {"int32":{"lte":1000,"gte":0}}

### spec.defaultRouteAction.weightedBackendServices[].headerAction

`GcpUrlMapHeaderAction`

Header mutations applied ONLY to the share of traffic sent to this
backend — e.g. tag canary responses with an identifying header so
clients and dashboards can tell which arm served them. Applied after
the URL-map-level and route-level header actions.

### spec.defaultRouteAction.weightedBackendServices[].headerAction.requestHeadersToAdd

`[]GcpUrlMapHeaderValue`

Headers to add to the request before it reaches the backend.

### spec.defaultRouteAction.weightedBackendServices[].headerAction.requestHeadersToAdd[].headerName

`string` · required

The header name (e.g. "X-Client-Geo").

- rule: {"required":true,"string":{"maxLen":"1024"}}

### spec.defaultRouteAction.weightedBackendServices[].headerAction.requestHeadersToAdd[].headerValue

`string` · required

The header value. May use load-balancer variables (e.g. "{client_region}").

- rule: {"required":true,"string":{"maxLen":"1024"}}

### spec.defaultRouteAction.weightedBackendServices[].headerAction.requestHeadersToAdd[].replace

`bool`

Replace an existing header of the same name (true) or append to it
(false).

### spec.defaultRouteAction.weightedBackendServices[].headerAction.requestHeadersToRemove

`[]string`

Header names to strip from the request before it reaches the backend.

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.defaultRouteAction.weightedBackendServices[].headerAction.responseHeadersToAdd

`[]GcpUrlMapHeaderValue`

Headers to add to the response before it returns to the client.

### spec.defaultRouteAction.weightedBackendServices[].headerAction.responseHeadersToAdd[].headerName

`string` · required

The header name (e.g. "X-Client-Geo").

- rule: {"required":true,"string":{"maxLen":"1024"}}

### spec.defaultRouteAction.weightedBackendServices[].headerAction.responseHeadersToAdd[].headerValue

`string` · required

The header value. May use load-balancer variables (e.g. "{client_region}").

- rule: {"required":true,"string":{"maxLen":"1024"}}

### spec.defaultRouteAction.weightedBackendServices[].headerAction.responseHeadersToAdd[].replace

`bool`

Replace an existing header of the same name (true) or append to it
(false).

### spec.defaultRouteAction.weightedBackendServices[].headerAction.responseHeadersToRemove

`[]string`

Header names to strip from the response before it returns to the client.

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.defaultRouteAction.urlRewrite

`GcpUrlMapUrlRewrite`

Rewrite the host and/or path before forwarding to the backend.

- rule: path_prefix_rewrite and path_template_rewrite are mutually exclusive — set at most one

### spec.defaultRouteAction.urlRewrite.hostRewrite

`string`

Replace the request Host header with this value before forwarding.

- rule: {"string":{"maxLen":"255"}}

### spec.defaultRouteAction.urlRewrite.pathPrefixRewrite

`string`

Replace the matched path prefix with this value. Mutually exclusive with
path_template_rewrite.

- rule: {"string":{"maxLen":"1024"}}

### spec.defaultRouteAction.urlRewrite.pathTemplateRewrite

`string`

Rewrite the path using a template that references named path variables
captured by a route rule's path_template_match (e.g. "/v2/{country}").
Honored inside a route_rule's route_action on both scopes and, on a
REGIONAL map only, in a path matcher's default route action — GCP
rejects it in the map's default route action and in path-rule route
actions everywhere. Mutually exclusive with path_prefix_rewrite.

- rule: {"string":{"maxLen":"1024"}}

### spec.defaultRouteAction.timeout

`GcpUrlMapDuration`

Total time budget for the request, INCLUDING all retries (from the
first byte of the request to the last byte of the response). Pair with
retry_policy.per_try_timeout: per-try bounds one attempt, this bounds
the whole exchange. Unset uses the backend service's own timeout.
Not permitted when the route targets a redirect.

### spec.defaultRouteAction.timeout.seconds

`int64` · optional (explicit presence)

Whole seconds (0 to 315,576,000,000 — GCP's int64 Duration bound).

- rule: {"int64":{"lte":"315576000000","gte":"0"}}

### spec.defaultRouteAction.timeout.nanos

`int32` · optional (explicit presence)

Fraction of a second at nanosecond resolution (0 to 999,999,999).
Durations under one second use seconds = 0 and a positive nanos.

- rule: {"int32":{"lte":999999999,"gte":0}}

### spec.defaultRouteAction.retryPolicy

`GcpUrlMapRetryPolicy`

Retry failed requests to the backend. Which failures count is chosen
by retry_conditions; how long each attempt may run by per_try_timeout.
Retries consume the overall timeout's budget — they never extend it.

### spec.defaultRouteAction.retryPolicy.numRetries

`int32`

Number of allowed retries (GCP defaults to 1 when unset/0). Each retry
still spends the route's overall timeout budget.

- rule: {"int32":{"gte":0}}

### spec.defaultRouteAction.retryPolicy.retryConditions

`[]string`

Which failures trigger a retry. 5xx (any 5xx or no response at all),
gateway-error (502/503/504 only), connect-failure, retriable-4xx
(currently only 409), refused-stream, and the gRPC status conditions
cancelled, deadline-exceeded, resource-exhausted, unavailable.

- rule: each retry condition must be one of: 5xx, gateway-error, connect-failure, retriable-4xx, refused-stream, cancelled, deadline-exceeded, resource-exhausted, unavailable

### spec.defaultRouteAction.retryPolicy.perTryTimeout

`GcpUrlMapDuration`

Time budget for EACH retry attempt (the route's timeout bounds the
whole exchange across attempts). Unset uses the overall timeout for
every attempt.

### spec.defaultRouteAction.retryPolicy.perTryTimeout.seconds

`int64` · optional (explicit presence)

Whole seconds (0 to 315,576,000,000 — GCP's int64 Duration bound).

- rule: {"int64":{"lte":"315576000000","gte":"0"}}

### spec.defaultRouteAction.retryPolicy.perTryTimeout.nanos

`int32` · optional (explicit presence)

Fraction of a second at nanosecond resolution (0 to 999,999,999).
Durations under one second use seconds = 0 and a positive nanos.

- rule: {"int32":{"lte":999999999,"gte":0}}

### spec.defaultRouteAction.requestMirrorPolicy

`GcpUrlMapRequestMirrorPolicy`

Mirror every matched request to a second backend service, fire-and-
forget (responses from the mirror are discarded; the client sees only
the primary's response). Useful for shadow-testing a new stack with
production traffic — the mirror backend must be sized for the full
mirrored load.

### spec.defaultRouteAction.requestMirrorPolicy.backendService

`string | valueFrom` · required

The backend service receiving the mirrored copy of every matched
request. Reference a GcpBackendService or provide a self-link
directly. Responses from this backend are discarded.

- references: GcpBackendService (`status.outputs.self_link`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpBackendService, name: <that resource's name>, fieldPath: status.outputs.self_link}} -- a bare string does not parse

### spec.defaultRouteAction.corsPolicy

`GcpUrlMapCorsPolicy`

Answer cross-origin (CORS) preflights and stamp CORS headers at the
load balancer, before requests reach the backend.

### spec.defaultRouteAction.corsPolicy.allowCredentials

`bool`

Sets Access-Control-Allow-Credentials: allow requests carrying
credentials (cookies, authorization headers). Default false.

### spec.defaultRouteAction.corsPolicy.allowHeaders

`[]string`

Headers the client may send (Access-Control-Allow-Headers).

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.defaultRouteAction.corsPolicy.allowMethods

`[]string`

Methods the client may use (Access-Control-Allow-Methods), e.g.
["GET", "POST", "OPTIONS"].

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.defaultRouteAction.corsPolicy.allowOriginRegexes

`[]string`

Regular expressions matching allowed origins; a request origin is
allowed when it matches any listed regex or exact allow_origins entry.

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.defaultRouteAction.corsPolicy.allowOrigins

`[]string`

Exact origins allowed, e.g. "https://app.example.com".

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.defaultRouteAction.corsPolicy.disabled

`bool`

Disable this CORS policy without deleting its configuration (true
turns the policy off). Default false — the policy is in effect.

### spec.defaultRouteAction.corsPolicy.exposeHeaders

`[]string`

Headers the browser may read from the response
(Access-Control-Expose-Headers).

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.defaultRouteAction.corsPolicy.maxAge

`int32`

Seconds a browser may cache the preflight response
(Access-Control-Max-Age).

- rule: {"int32":{"gte":0}}

### spec.defaultRouteAction.faultInjectionPolicy

`GcpUrlMapFaultInjectionPolicy`

Deliberately inject failures (aborts with a chosen status, fixed
delays) into a percentage of matched requests — chaos/resilience
testing of the CLIENTS of this route. Never leave enabled on a
production default route: the injected failures are real to callers.

### spec.defaultRouteAction.faultInjectionPolicy.abort

`GcpUrlMapFaultAbort`

Abort a percentage of requests with a fixed HTTP status, before they
reach the backend.

### spec.defaultRouteAction.faultInjectionPolicy.abort.httpStatus

`int32`

HTTP status returned to aborted requests (200-599).

- rule: http_status must be between 200 and 599

### spec.defaultRouteAction.faultInjectionPolicy.abort.percentage

`double`

Percentage of requests aborted (0.0-100.0).

- rule: {"double":{"lte":100,"gte":0}}

### spec.defaultRouteAction.faultInjectionPolicy.delay

`GcpUrlMapFaultDelay`

Delay a percentage of requests by a fixed duration before forwarding.

### spec.defaultRouteAction.faultInjectionPolicy.delay.fixedDelay

`GcpUrlMapDuration`

How long delayed requests are held before forwarding.

### spec.defaultRouteAction.faultInjectionPolicy.delay.fixedDelay.seconds

`int64` · optional (explicit presence)

Whole seconds (0 to 315,576,000,000 — GCP's int64 Duration bound).

- rule: {"int64":{"lte":"315576000000","gte":"0"}}

### spec.defaultRouteAction.faultInjectionPolicy.delay.fixedDelay.nanos

`int32` · optional (explicit presence)

Fraction of a second at nanosecond resolution (0 to 999,999,999).
Durations under one second use seconds = 0 and a positive nanos.

- rule: {"int32":{"lte":999999999,"gte":0}}

### spec.defaultRouteAction.faultInjectionPolicy.delay.percentage

`double`

Percentage of requests delayed (0.0-100.0).

- rule: {"double":{"lte":100,"gte":0}}

### spec.defaultRouteAction.maxStreamDuration

`GcpUrlMapDuration`

Upper bound on how long a STREAM on this route may stay open
(gRPC/long-poll streams — distinct from timeout, which bounds a
request/response exchange). Live API truth: GCP rejects this field
unless the URL map's backend service uses the INTERNAL_SELF_MANAGED
(Traffic Director) load-balancing scheme ("Max stream duration is
only supported when UrlMap is used with BackendService whose Load
Balancing Scheme is INTERNAL_SELF_MANAGED") — leave it unset on
external application load balancers. Global maps only: the regional
map has no such argument (except in a path matcher's default route
action, which carries it on both scopes).

### spec.defaultRouteAction.maxStreamDuration.seconds

`int64` · optional (explicit presence)

Whole seconds (0 to 315,576,000,000 — GCP's int64 Duration bound).

- rule: {"int64":{"lte":"315576000000","gte":"0"}}

### spec.defaultRouteAction.maxStreamDuration.nanos

`int32` · optional (explicit presence)

Fraction of a second at nanosecond resolution (0 to 999,999,999).
Durations under one second use seconds = 0 and a positive nanos.

- rule: {"int32":{"lte":999999999,"gte":0}}

### spec.defaultRouteAction.cachePolicy

`GcpUrlMapCachePolicy`

Cloud CDN caching for the routes using this action — overrides the
backend service's cdn_policy for matching traffic only. Takes effect
only when the target backend service (or bucket) has CDN enabled;
GCP ignores it otherwise. Global maps only — regional Application Load
Balancers have no Cloud CDN, and the regional map has no such argument.

- rule: with cache_mode USE_ORIGIN_HEADERS the origin's headers control lifetimes — remove client_ttl, default_ttl, and max_ttl (GCP would silently ignore them)

### spec.defaultRouteAction.cachePolicy.cacheMode

`string`

What gets cached. CACHE_ALL_STATIC (the GCP default) caches static
content types and honors origin cache headers for the rest;
USE_ORIGIN_HEADERS caches only what the backends explicitly mark
cacheable (TTL fields must be unset — the origin controls lifetimes);
FORCE_CACHE_ALL caches everything, ignoring origin headers (never
combine with private or per-user content).

- rule: cache_mode must be CACHE_ALL_STATIC, USE_ORIGIN_HEADERS, or FORCE_CACHE_ALL

### spec.defaultRouteAction.cachePolicy.cacheBypassRequestHeaderNames

`[]string`

Requests carrying any of these headers bypass the cache and go
straight to the origin (e.g. a debug or authorization header).

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.defaultRouteAction.cachePolicy.negativeCaching

`bool`

Cache negative responses (404, 410, ...) so repeated misses do not
hammer the backends. Pair with negative_caching_policy to set
per-status TTLs.

### spec.defaultRouteAction.cachePolicy.requestCoalescing

`bool`

Collapse concurrent cache-fill requests for the same key into one
origin fetch. Default true on GCP's side.

### spec.defaultRouteAction.cachePolicy.cacheKeyPolicy

`GcpUrlMapCacheKeyPolicy`

What forms the cache key — trim protocol, host, query string, or
select specific query parameters, headers, and cookies.

- rule: included_query_parameters and excluded_query_parameters are mutually exclusive — set at most one list

### spec.defaultRouteAction.cachePolicy.cacheKeyPolicy.excludedQueryParameters

`[]string`

Query parameters EXCLUDED from the cache key (everything else is
included). Mutually exclusive with included_query_parameters.

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.defaultRouteAction.cachePolicy.cacheKeyPolicy.includeHost

`bool`

Include the request host in the cache key (distinct hosts cache
separately).

### spec.defaultRouteAction.cachePolicy.cacheKeyPolicy.includeProtocol

`bool`

Include the protocol (http/https) in the cache key.

### spec.defaultRouteAction.cachePolicy.cacheKeyPolicy.includeQueryString

`bool`

Include the entire query string in the cache key. When false, the
included/excluded parameter lists refine what participates.

### spec.defaultRouteAction.cachePolicy.cacheKeyPolicy.includedCookieNames

`[]string`

Cookie names whose values join the cache key.

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.defaultRouteAction.cachePolicy.cacheKeyPolicy.includedHeaderNames

`[]string`

Header names whose values join the cache key.

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.defaultRouteAction.cachePolicy.cacheKeyPolicy.includedQueryParameters

`[]string`

Query parameters INCLUDED in the cache key (everything else is
ignored). Mutually exclusive with excluded_query_parameters.

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.defaultRouteAction.cachePolicy.clientTtl

`GcpUrlMapDuration`

Max lifetime in browsers and downstream caches (sets the max-age
clients see). Not respected with cache_mode USE_ORIGIN_HEADERS.

### spec.defaultRouteAction.cachePolicy.clientTtl.seconds

`int64` · optional (explicit presence)

Whole seconds (0 to 315,576,000,000 — GCP's int64 Duration bound).

- rule: {"int64":{"lte":"315576000000","gte":"0"}}

### spec.defaultRouteAction.cachePolicy.clientTtl.nanos

`int32` · optional (explicit presence)

Fraction of a second at nanosecond resolution (0 to 999,999,999).
Durations under one second use seconds = 0 and a positive nanos.

- rule: {"int32":{"lte":999999999,"gte":0}}

### spec.defaultRouteAction.cachePolicy.defaultTtl

`GcpUrlMapDuration`

Edge-cache lifetime for responses without their own cache headers.
Not respected with cache_mode USE_ORIGIN_HEADERS.

### spec.defaultRouteAction.cachePolicy.defaultTtl.seconds

`int64` · optional (explicit presence)

Whole seconds (0 to 315,576,000,000 — GCP's int64 Duration bound).

- rule: {"int64":{"lte":"315576000000","gte":"0"}}

### spec.defaultRouteAction.cachePolicy.defaultTtl.nanos

`int32` · optional (explicit presence)

Fraction of a second at nanosecond resolution (0 to 999,999,999).
Durations under one second use seconds = 0 and a positive nanos.

- rule: {"int32":{"lte":999999999,"gte":0}}

### spec.defaultRouteAction.cachePolicy.maxTtl

`GcpUrlMapDuration`

Upper bound on any edge-cache lifetime, capping even origin-supplied
max-age. Not respected with cache_mode USE_ORIGIN_HEADERS or
FORCE_CACHE_ALL.

### spec.defaultRouteAction.cachePolicy.maxTtl.seconds

`int64` · optional (explicit presence)

Whole seconds (0 to 315,576,000,000 — GCP's int64 Duration bound).

- rule: {"int64":{"lte":"315576000000","gte":"0"}}

### spec.defaultRouteAction.cachePolicy.maxTtl.nanos

`int32` · optional (explicit presence)

Fraction of a second at nanosecond resolution (0 to 999,999,999).
Durations under one second use seconds = 0 and a positive nanos.

- rule: {"int32":{"lte":999999999,"gte":0}}

### spec.defaultRouteAction.cachePolicy.serveWhileStale

`GcpUrlMapDuration`

How long stale content may still be served while revalidating in the
background (up to 1 day).

### spec.defaultRouteAction.cachePolicy.serveWhileStale.seconds

`int64` · optional (explicit presence)

Whole seconds (0 to 315,576,000,000 — GCP's int64 Duration bound).

- rule: {"int64":{"lte":"315576000000","gte":"0"}}

### spec.defaultRouteAction.cachePolicy.serveWhileStale.nanos

`int32` · optional (explicit presence)

Fraction of a second at nanosecond resolution (0 to 999,999,999).
Durations under one second use seconds = 0 and a positive nanos.

- rule: {"int32":{"lte":999999999,"gte":0}}

### spec.defaultRouteAction.cachePolicy.negativeCachingPolicy

`[]GcpUrlMapNegativeCachingPolicy`

Per-status TTLs for cached negative responses; only meaningful with
negative_caching enabled.

### spec.defaultRouteAction.cachePolicy.negativeCachingPolicy[].code

`int32`

The HTTP status code this TTL applies to. GCP accepts only 300, 301,
302, 307, 308, 404, 405, 410, 421, 451, and 501, each at most once.

- rule: code must be one of 300, 301, 302, 307, 308, 404, 405, 410, 421, 451, 501

### spec.defaultRouteAction.cachePolicy.negativeCachingPolicy[].ttl

`GcpUrlMapDuration`

How long responses with this status are cached.

### spec.defaultRouteAction.cachePolicy.negativeCachingPolicy[].ttl.seconds

`int64` · optional (explicit presence)

Whole seconds (0 to 315,576,000,000 — GCP's int64 Duration bound).

- rule: {"int64":{"lte":"315576000000","gte":"0"}}

### spec.defaultRouteAction.cachePolicy.negativeCachingPolicy[].ttl.nanos

`int32` · optional (explicit presence)

Fraction of a second at nanosecond resolution (0 to 999,999,999).
Durations under one second use seconds = 0 and a positive nanos.

- rule: {"int32":{"lte":999999999,"gte":0}}

### spec.defaultCustomErrorResponsePolicy

`GcpUrlMapCustomErrorResponsePolicy`

Return a custom error page (from a backend bucket) for chosen response
codes at the top level. Global external Application Load Balancers only
— rejected when region is set.

### spec.defaultCustomErrorResponsePolicy.errorService

`string | valueFrom`

The backend bucket serving the error pages. Reference a GcpBackendBucket
or provide a self-link directly.

- references: GcpBackendBucket (`status.outputs.self_link`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpBackendBucket, name: <that resource's name>, fieldPath: status.outputs.self_link}} -- a bare string does not parse

### spec.defaultCustomErrorResponsePolicy.errorResponseRules

`[]GcpUrlMapCustomErrorResponseRule`

Rules mapping response codes to the error page and (optionally) an
overridden status code.

### spec.defaultCustomErrorResponsePolicy.errorResponseRules[].matchResponseCodes

`[]string` · required

Response codes this rule matches: exact ("404", "503") or a class ("4xx",
"5xx"). At least one is required.

- rule: {"repeated":{"minItems":"1","items":{"string":{"minLen":"1"}}}}

### spec.defaultCustomErrorResponsePolicy.errorResponseRules[].overrideResponseCode

`int32`

Override the response code returned to the client (e.g. serve a 200 with a
maintenance page). Empty keeps the original code.

- rule: override_response_code must be between 200 and 599

### spec.defaultCustomErrorResponsePolicy.errorResponseRules[].path

`string` · required

Path within the error backend bucket to serve for matched codes (e.g.
"/errors/404.html").

- rule: {"required":true,"string":{"minLen":"1","maxLen":"1024"}}

### spec.headerAction

`GcpUrlMapHeaderAction`

Headers added to or removed from every request/response at the URL-map
level, before any per-route header action. Mutable.

### spec.headerAction.requestHeadersToAdd

`[]GcpUrlMapHeaderValue`

Headers to add to the request before it reaches the backend.

### spec.headerAction.requestHeadersToAdd[].headerName

`string` · required

The header name (e.g. "X-Client-Geo").

- rule: {"required":true,"string":{"maxLen":"1024"}}

### spec.headerAction.requestHeadersToAdd[].headerValue

`string` · required

The header value. May use load-balancer variables (e.g. "{client_region}").

- rule: {"required":true,"string":{"maxLen":"1024"}}

### spec.headerAction.requestHeadersToAdd[].replace

`bool`

Replace an existing header of the same name (true) or append to it
(false).

### spec.headerAction.requestHeadersToRemove

`[]string`

Header names to strip from the request before it reaches the backend.

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.headerAction.responseHeadersToAdd

`[]GcpUrlMapHeaderValue`

Headers to add to the response before it returns to the client.

### spec.headerAction.responseHeadersToAdd[].headerName

`string` · required

The header name (e.g. "X-Client-Geo").

- rule: {"required":true,"string":{"maxLen":"1024"}}

### spec.headerAction.responseHeadersToAdd[].headerValue

`string` · required

The header value. May use load-balancer variables (e.g. "{client_region}").

- rule: {"required":true,"string":{"maxLen":"1024"}}

### spec.headerAction.responseHeadersToAdd[].replace

`bool`

Replace an existing header of the same name (true) or append to it
(false).

### spec.headerAction.responseHeadersToRemove

`[]string`

Header names to strip from the response before it returns to the client.

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.hostRules

`[]GcpUrlMapHostRule`

Map request Host headers to named path matchers. Each host rule points a
set of hosts (with optional wildcards) at one path_matcher by name.
Mutable.

### spec.hostRules[].hosts

`[]string` · required

Hosts this rule matches. A "*" wildcard may lead a domain
(e.g. "*.example.com") or match everything ("*"). At least one required.

- rule: {"repeated":{"minItems":"1","items":{"string":{"minLen":"1"}}}}

### spec.hostRules[].pathMatcher

`string` · required

The name of the path_matcher (in path_matchers) that handles these hosts.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.hostRules[].description

`string`

What this host rule covers — write it for the operator reading the routing
table later.

- rule: {"string":{"maxLen":"1024"}}

### spec.pathMatchers

`[]GcpUrlMapPathMatcher`

The named path matchers host rules point at — each owns the path-level
routing (path_rules, route_rules, and its own default). Mutable.

- rule: a path matcher may set at most one default target: default_service, default_url_redirect, or default_route_action with weighted_backend_services (a route action carrying only sub-policies may accompany default_service)
- rule: default_route_action and default_url_redirect are mutually exclusive — a redirect never reaches a backend
- rule: a path matcher uses either path_rules or route_rules, not both

### spec.pathMatchers[].name

`string` · required

The path matcher's name, referenced by host_rules.path_matcher.

- rule: {"required":true,"string":{"minLen":"1","maxLen":"63"}}

### spec.pathMatchers[].defaultService

`string | valueFrom`

The default target when no path_rule or route_rule matches — a backend
service or backend bucket. Reference a GcpBackendService or
GcpBackendBucket, or provide a self-link. Set exactly one of
default_service, default_url_redirect, or default_route_action.

- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.pathMatchers[].defaultUrlRedirect

`GcpUrlMapUrlRedirect`

Redirect as the path matcher's default instead of serving.

- rule: path_redirect and prefix_redirect are mutually exclusive — set at most one

### spec.pathMatchers[].defaultUrlRedirect.hostRedirect

`string`

Replace the host in the redirect Location. Empty keeps the request host.

- rule: {"string":{"maxLen":"255"}}

### spec.pathMatchers[].defaultUrlRedirect.httpsRedirect

`bool`

Redirect to HTTPS (scheme becomes https). The standard http→https
upgrade. Default false.

### spec.pathMatchers[].defaultUrlRedirect.pathRedirect

`string`

Replace the entire path with this value. Mutually exclusive with
prefix_redirect. Empty keeps the request path.

- rule: {"string":{"maxLen":"1024"}}

### spec.pathMatchers[].defaultUrlRedirect.prefixRedirect

`string`

Replace the matched path prefix with this value, keeping the remainder.
Mutually exclusive with path_redirect.

- rule: {"string":{"maxLen":"1024"}}

### spec.pathMatchers[].defaultUrlRedirect.redirectResponseCode

`string`

The HTTP redirect status code: FOUND (302), MOVED_PERMANENTLY_DEFAULT
(301), PERMANENT_REDIRECT (308), SEE_OTHER (303), or TEMPORARY_REDIRECT
(307). Empty uses the GCP default (MOVED_PERMANENTLY_DEFAULT).

- rule: redirect_response_code must be one of FOUND, MOVED_PERMANENTLY_DEFAULT, PERMANENT_REDIRECT, SEE_OTHER, or TEMPORARY_REDIRECT

### spec.pathMatchers[].defaultUrlRedirect.stripQuery

`bool`

Drop the query string from the redirect Location. Default false (the query
string is preserved).

### spec.pathMatchers[].defaultRouteAction

`GcpUrlMapRouteAction`

Advanced default handling (weighted split / rewrite) for the path matcher.
On a regional map this is the one place besides route rules where a
url_rewrite may carry path_template_rewrite.

### spec.pathMatchers[].defaultRouteAction.weightedBackendServices

`[]GcpUrlMapWeightedBackendService`

Split traffic across multiple backend services by weight — the mechanism
for weighted canary and blue/green rollouts. The weights are relative; a
backend's share is its weight over the sum of weights.

### spec.pathMatchers[].defaultRouteAction.weightedBackendServices[].backendService

`string | valueFrom` · required

The backend service receiving this share of traffic. Reference a
GcpBackendService or provide a self-link directly.

- references: GcpBackendService (`status.outputs.self_link`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpBackendService, name: <that resource's name>, fieldPath: status.outputs.self_link}} -- a bare string does not parse

### spec.pathMatchers[].defaultRouteAction.weightedBackendServices[].weight

`int32`

Relative weight of this backend (0-1000). Its share is weight over the sum
of all weights in the split; 0 drains this backend from the split.

- rule: {"int32":{"lte":1000,"gte":0}}

### spec.pathMatchers[].defaultRouteAction.weightedBackendServices[].headerAction

`GcpUrlMapHeaderAction`

Header mutations applied ONLY to the share of traffic sent to this
backend — e.g. tag canary responses with an identifying header so
clients and dashboards can tell which arm served them. Applied after
the URL-map-level and route-level header actions.

### spec.pathMatchers[].defaultRouteAction.weightedBackendServices[].headerAction.requestHeadersToAdd

`[]GcpUrlMapHeaderValue`

Headers to add to the request before it reaches the backend.

### spec.pathMatchers[].defaultRouteAction.weightedBackendServices[].headerAction.requestHeadersToAdd[].headerName

`string` · required

The header name (e.g. "X-Client-Geo").

- rule: {"required":true,"string":{"maxLen":"1024"}}

### spec.pathMatchers[].defaultRouteAction.weightedBackendServices[].headerAction.requestHeadersToAdd[].headerValue

`string` · required

The header value. May use load-balancer variables (e.g. "{client_region}").

- rule: {"required":true,"string":{"maxLen":"1024"}}

### spec.pathMatchers[].defaultRouteAction.weightedBackendServices[].headerAction.requestHeadersToAdd[].replace

`bool`

Replace an existing header of the same name (true) or append to it
(false).

### spec.pathMatchers[].defaultRouteAction.weightedBackendServices[].headerAction.requestHeadersToRemove

`[]string`

Header names to strip from the request before it reaches the backend.

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.pathMatchers[].defaultRouteAction.weightedBackendServices[].headerAction.responseHeadersToAdd

`[]GcpUrlMapHeaderValue`

Headers to add to the response before it returns to the client.

### spec.pathMatchers[].defaultRouteAction.weightedBackendServices[].headerAction.responseHeadersToAdd[].headerName

`string` · required

The header name (e.g. "X-Client-Geo").

- rule: {"required":true,"string":{"maxLen":"1024"}}

### spec.pathMatchers[].defaultRouteAction.weightedBackendServices[].headerAction.responseHeadersToAdd[].headerValue

`string` · required

The header value. May use load-balancer variables (e.g. "{client_region}").

- rule: {"required":true,"string":{"maxLen":"1024"}}

### spec.pathMatchers[].defaultRouteAction.weightedBackendServices[].headerAction.responseHeadersToAdd[].replace

`bool`

Replace an existing header of the same name (true) or append to it
(false).

### spec.pathMatchers[].defaultRouteAction.weightedBackendServices[].headerAction.responseHeadersToRemove

`[]string`

Header names to strip from the response before it returns to the client.

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.pathMatchers[].defaultRouteAction.urlRewrite

`GcpUrlMapUrlRewrite`

Rewrite the host and/or path before forwarding to the backend.

- rule: path_prefix_rewrite and path_template_rewrite are mutually exclusive — set at most one

### spec.pathMatchers[].defaultRouteAction.urlRewrite.hostRewrite

`string`

Replace the request Host header with this value before forwarding.

- rule: {"string":{"maxLen":"255"}}

### spec.pathMatchers[].defaultRouteAction.urlRewrite.pathPrefixRewrite

`string`

Replace the matched path prefix with this value. Mutually exclusive with
path_template_rewrite.

- rule: {"string":{"maxLen":"1024"}}

### spec.pathMatchers[].defaultRouteAction.urlRewrite.pathTemplateRewrite

`string`

Rewrite the path using a template that references named path variables
captured by a route rule's path_template_match (e.g. "/v2/{country}").
Honored inside a route_rule's route_action on both scopes and, on a
REGIONAL map only, in a path matcher's default route action — GCP
rejects it in the map's default route action and in path-rule route
actions everywhere. Mutually exclusive with path_prefix_rewrite.

- rule: {"string":{"maxLen":"1024"}}

### spec.pathMatchers[].defaultRouteAction.timeout

`GcpUrlMapDuration`

Total time budget for the request, INCLUDING all retries (from the
first byte of the request to the last byte of the response). Pair with
retry_policy.per_try_timeout: per-try bounds one attempt, this bounds
the whole exchange. Unset uses the backend service's own timeout.
Not permitted when the route targets a redirect.

### spec.pathMatchers[].defaultRouteAction.timeout.seconds

`int64` · optional (explicit presence)

Whole seconds (0 to 315,576,000,000 — GCP's int64 Duration bound).

- rule: {"int64":{"lte":"315576000000","gte":"0"}}

### spec.pathMatchers[].defaultRouteAction.timeout.nanos

`int32` · optional (explicit presence)

Fraction of a second at nanosecond resolution (0 to 999,999,999).
Durations under one second use seconds = 0 and a positive nanos.

- rule: {"int32":{"lte":999999999,"gte":0}}

### spec.pathMatchers[].defaultRouteAction.retryPolicy

`GcpUrlMapRetryPolicy`

Retry failed requests to the backend. Which failures count is chosen
by retry_conditions; how long each attempt may run by per_try_timeout.
Retries consume the overall timeout's budget — they never extend it.

### spec.pathMatchers[].defaultRouteAction.retryPolicy.numRetries

`int32`

Number of allowed retries (GCP defaults to 1 when unset/0). Each retry
still spends the route's overall timeout budget.

- rule: {"int32":{"gte":0}}

### spec.pathMatchers[].defaultRouteAction.retryPolicy.retryConditions

`[]string`

Which failures trigger a retry. 5xx (any 5xx or no response at all),
gateway-error (502/503/504 only), connect-failure, retriable-4xx
(currently only 409), refused-stream, and the gRPC status conditions
cancelled, deadline-exceeded, resource-exhausted, unavailable.

- rule: each retry condition must be one of: 5xx, gateway-error, connect-failure, retriable-4xx, refused-stream, cancelled, deadline-exceeded, resource-exhausted, unavailable

### spec.pathMatchers[].defaultRouteAction.retryPolicy.perTryTimeout

`GcpUrlMapDuration`

Time budget for EACH retry attempt (the route's timeout bounds the
whole exchange across attempts). Unset uses the overall timeout for
every attempt.

### spec.pathMatchers[].defaultRouteAction.retryPolicy.perTryTimeout.seconds

`int64` · optional (explicit presence)

Whole seconds (0 to 315,576,000,000 — GCP's int64 Duration bound).

- rule: {"int64":{"lte":"315576000000","gte":"0"}}

### spec.pathMatchers[].defaultRouteAction.retryPolicy.perTryTimeout.nanos

`int32` · optional (explicit presence)

Fraction of a second at nanosecond resolution (0 to 999,999,999).
Durations under one second use seconds = 0 and a positive nanos.

- rule: {"int32":{"lte":999999999,"gte":0}}

### spec.pathMatchers[].defaultRouteAction.requestMirrorPolicy

`GcpUrlMapRequestMirrorPolicy`

Mirror every matched request to a second backend service, fire-and-
forget (responses from the mirror are discarded; the client sees only
the primary's response). Useful for shadow-testing a new stack with
production traffic — the mirror backend must be sized for the full
mirrored load.

### spec.pathMatchers[].defaultRouteAction.requestMirrorPolicy.backendService

`string | valueFrom` · required

The backend service receiving the mirrored copy of every matched
request. Reference a GcpBackendService or provide a self-link
directly. Responses from this backend are discarded.

- references: GcpBackendService (`status.outputs.self_link`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpBackendService, name: <that resource's name>, fieldPath: status.outputs.self_link}} -- a bare string does not parse

### spec.pathMatchers[].defaultRouteAction.corsPolicy

`GcpUrlMapCorsPolicy`

Answer cross-origin (CORS) preflights and stamp CORS headers at the
load balancer, before requests reach the backend.

### spec.pathMatchers[].defaultRouteAction.corsPolicy.allowCredentials

`bool`

Sets Access-Control-Allow-Credentials: allow requests carrying
credentials (cookies, authorization headers). Default false.

### spec.pathMatchers[].defaultRouteAction.corsPolicy.allowHeaders

`[]string`

Headers the client may send (Access-Control-Allow-Headers).

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.pathMatchers[].defaultRouteAction.corsPolicy.allowMethods

`[]string`

Methods the client may use (Access-Control-Allow-Methods), e.g.
["GET", "POST", "OPTIONS"].

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.pathMatchers[].defaultRouteAction.corsPolicy.allowOriginRegexes

`[]string`

Regular expressions matching allowed origins; a request origin is
allowed when it matches any listed regex or exact allow_origins entry.

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.pathMatchers[].defaultRouteAction.corsPolicy.allowOrigins

`[]string`

Exact origins allowed, e.g. "https://app.example.com".

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.pathMatchers[].defaultRouteAction.corsPolicy.disabled

`bool`

Disable this CORS policy without deleting its configuration (true
turns the policy off). Default false — the policy is in effect.

### spec.pathMatchers[].defaultRouteAction.corsPolicy.exposeHeaders

`[]string`

Headers the browser may read from the response
(Access-Control-Expose-Headers).

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.pathMatchers[].defaultRouteAction.corsPolicy.maxAge

`int32`

Seconds a browser may cache the preflight response
(Access-Control-Max-Age).

- rule: {"int32":{"gte":0}}

### spec.pathMatchers[].defaultRouteAction.faultInjectionPolicy

`GcpUrlMapFaultInjectionPolicy`

Deliberately inject failures (aborts with a chosen status, fixed
delays) into a percentage of matched requests — chaos/resilience
testing of the CLIENTS of this route. Never leave enabled on a
production default route: the injected failures are real to callers.

### spec.pathMatchers[].defaultRouteAction.faultInjectionPolicy.abort

`GcpUrlMapFaultAbort`

Abort a percentage of requests with a fixed HTTP status, before they
reach the backend.

### spec.pathMatchers[].defaultRouteAction.faultInjectionPolicy.abort.httpStatus

`int32`

HTTP status returned to aborted requests (200-599).

- rule: http_status must be between 200 and 599

### spec.pathMatchers[].defaultRouteAction.faultInjectionPolicy.abort.percentage

`double`

Percentage of requests aborted (0.0-100.0).

- rule: {"double":{"lte":100,"gte":0}}

### spec.pathMatchers[].defaultRouteAction.faultInjectionPolicy.delay

`GcpUrlMapFaultDelay`

Delay a percentage of requests by a fixed duration before forwarding.

### spec.pathMatchers[].defaultRouteAction.faultInjectionPolicy.delay.fixedDelay

`GcpUrlMapDuration`

How long delayed requests are held before forwarding.

### spec.pathMatchers[].defaultRouteAction.faultInjectionPolicy.delay.fixedDelay.seconds

`int64` · optional (explicit presence)

Whole seconds (0 to 315,576,000,000 — GCP's int64 Duration bound).

- rule: {"int64":{"lte":"315576000000","gte":"0"}}

### spec.pathMatchers[].defaultRouteAction.faultInjectionPolicy.delay.fixedDelay.nanos

`int32` · optional (explicit presence)

Fraction of a second at nanosecond resolution (0 to 999,999,999).
Durations under one second use seconds = 0 and a positive nanos.

- rule: {"int32":{"lte":999999999,"gte":0}}

### spec.pathMatchers[].defaultRouteAction.faultInjectionPolicy.delay.percentage

`double`

Percentage of requests delayed (0.0-100.0).

- rule: {"double":{"lte":100,"gte":0}}

### spec.pathMatchers[].defaultRouteAction.maxStreamDuration

`GcpUrlMapDuration`

Upper bound on how long a STREAM on this route may stay open
(gRPC/long-poll streams — distinct from timeout, which bounds a
request/response exchange). Live API truth: GCP rejects this field
unless the URL map's backend service uses the INTERNAL_SELF_MANAGED
(Traffic Director) load-balancing scheme ("Max stream duration is
only supported when UrlMap is used with BackendService whose Load
Balancing Scheme is INTERNAL_SELF_MANAGED") — leave it unset on
external application load balancers. Global maps only: the regional
map has no such argument (except in a path matcher's default route
action, which carries it on both scopes).

### spec.pathMatchers[].defaultRouteAction.maxStreamDuration.seconds

`int64` · optional (explicit presence)

Whole seconds (0 to 315,576,000,000 — GCP's int64 Duration bound).

- rule: {"int64":{"lte":"315576000000","gte":"0"}}

### spec.pathMatchers[].defaultRouteAction.maxStreamDuration.nanos

`int32` · optional (explicit presence)

Fraction of a second at nanosecond resolution (0 to 999,999,999).
Durations under one second use seconds = 0 and a positive nanos.

- rule: {"int32":{"lte":999999999,"gte":0}}

### spec.pathMatchers[].defaultRouteAction.cachePolicy

`GcpUrlMapCachePolicy`

Cloud CDN caching for the routes using this action — overrides the
backend service's cdn_policy for matching traffic only. Takes effect
only when the target backend service (or bucket) has CDN enabled;
GCP ignores it otherwise. Global maps only — regional Application Load
Balancers have no Cloud CDN, and the regional map has no such argument.

- rule: with cache_mode USE_ORIGIN_HEADERS the origin's headers control lifetimes — remove client_ttl, default_ttl, and max_ttl (GCP would silently ignore them)

### spec.pathMatchers[].defaultRouteAction.cachePolicy.cacheMode

`string`

What gets cached. CACHE_ALL_STATIC (the GCP default) caches static
content types and honors origin cache headers for the rest;
USE_ORIGIN_HEADERS caches only what the backends explicitly mark
cacheable (TTL fields must be unset — the origin controls lifetimes);
FORCE_CACHE_ALL caches everything, ignoring origin headers (never
combine with private or per-user content).

- rule: cache_mode must be CACHE_ALL_STATIC, USE_ORIGIN_HEADERS, or FORCE_CACHE_ALL

### spec.pathMatchers[].defaultRouteAction.cachePolicy.cacheBypassRequestHeaderNames

`[]string`

Requests carrying any of these headers bypass the cache and go
straight to the origin (e.g. a debug or authorization header).

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.pathMatchers[].defaultRouteAction.cachePolicy.negativeCaching

`bool`

Cache negative responses (404, 410, ...) so repeated misses do not
hammer the backends. Pair with negative_caching_policy to set
per-status TTLs.

### spec.pathMatchers[].defaultRouteAction.cachePolicy.requestCoalescing

`bool`

Collapse concurrent cache-fill requests for the same key into one
origin fetch. Default true on GCP's side.

### spec.pathMatchers[].defaultRouteAction.cachePolicy.cacheKeyPolicy

`GcpUrlMapCacheKeyPolicy`

What forms the cache key — trim protocol, host, query string, or
select specific query parameters, headers, and cookies.

- rule: included_query_parameters and excluded_query_parameters are mutually exclusive — set at most one list

### spec.pathMatchers[].defaultRouteAction.cachePolicy.cacheKeyPolicy.excludedQueryParameters

`[]string`

Query parameters EXCLUDED from the cache key (everything else is
included). Mutually exclusive with included_query_parameters.

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.pathMatchers[].defaultRouteAction.cachePolicy.cacheKeyPolicy.includeHost

`bool`

Include the request host in the cache key (distinct hosts cache
separately).

### spec.pathMatchers[].defaultRouteAction.cachePolicy.cacheKeyPolicy.includeProtocol

`bool`

Include the protocol (http/https) in the cache key.

### spec.pathMatchers[].defaultRouteAction.cachePolicy.cacheKeyPolicy.includeQueryString

`bool`

Include the entire query string in the cache key. When false, the
included/excluded parameter lists refine what participates.

### spec.pathMatchers[].defaultRouteAction.cachePolicy.cacheKeyPolicy.includedCookieNames

`[]string`

Cookie names whose values join the cache key.

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.pathMatchers[].defaultRouteAction.cachePolicy.cacheKeyPolicy.includedHeaderNames

`[]string`

Header names whose values join the cache key.

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.pathMatchers[].defaultRouteAction.cachePolicy.cacheKeyPolicy.includedQueryParameters

`[]string`

Query parameters INCLUDED in the cache key (everything else is
ignored). Mutually exclusive with excluded_query_parameters.

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.pathMatchers[].defaultRouteAction.cachePolicy.clientTtl

`GcpUrlMapDuration`

Max lifetime in browsers and downstream caches (sets the max-age
clients see). Not respected with cache_mode USE_ORIGIN_HEADERS.

### spec.pathMatchers[].defaultRouteAction.cachePolicy.clientTtl.seconds

`int64` · optional (explicit presence)

Whole seconds (0 to 315,576,000,000 — GCP's int64 Duration bound).

- rule: {"int64":{"lte":"315576000000","gte":"0"}}

### spec.pathMatchers[].defaultRouteAction.cachePolicy.clientTtl.nanos

`int32` · optional (explicit presence)

Fraction of a second at nanosecond resolution (0 to 999,999,999).
Durations under one second use seconds = 0 and a positive nanos.

- rule: {"int32":{"lte":999999999,"gte":0}}

### spec.pathMatchers[].defaultRouteAction.cachePolicy.defaultTtl

`GcpUrlMapDuration`

Edge-cache lifetime for responses without their own cache headers.
Not respected with cache_mode USE_ORIGIN_HEADERS.

### spec.pathMatchers[].defaultRouteAction.cachePolicy.defaultTtl.seconds

`int64` · optional (explicit presence)

Whole seconds (0 to 315,576,000,000 — GCP's int64 Duration bound).

- rule: {"int64":{"lte":"315576000000","gte":"0"}}

### spec.pathMatchers[].defaultRouteAction.cachePolicy.defaultTtl.nanos

`int32` · optional (explicit presence)

Fraction of a second at nanosecond resolution (0 to 999,999,999).
Durations under one second use seconds = 0 and a positive nanos.

- rule: {"int32":{"lte":999999999,"gte":0}}

### spec.pathMatchers[].defaultRouteAction.cachePolicy.maxTtl

`GcpUrlMapDuration`

Upper bound on any edge-cache lifetime, capping even origin-supplied
max-age. Not respected with cache_mode USE_ORIGIN_HEADERS or
FORCE_CACHE_ALL.

### spec.pathMatchers[].defaultRouteAction.cachePolicy.maxTtl.seconds

`int64` · optional (explicit presence)

Whole seconds (0 to 315,576,000,000 — GCP's int64 Duration bound).

- rule: {"int64":{"lte":"315576000000","gte":"0"}}

### spec.pathMatchers[].defaultRouteAction.cachePolicy.maxTtl.nanos

`int32` · optional (explicit presence)

Fraction of a second at nanosecond resolution (0 to 999,999,999).
Durations under one second use seconds = 0 and a positive nanos.

- rule: {"int32":{"lte":999999999,"gte":0}}

### spec.pathMatchers[].defaultRouteAction.cachePolicy.serveWhileStale

`GcpUrlMapDuration`

How long stale content may still be served while revalidating in the
background (up to 1 day).

### spec.pathMatchers[].defaultRouteAction.cachePolicy.serveWhileStale.seconds

`int64` · optional (explicit presence)

Whole seconds (0 to 315,576,000,000 — GCP's int64 Duration bound).

- rule: {"int64":{"lte":"315576000000","gte":"0"}}

### spec.pathMatchers[].defaultRouteAction.cachePolicy.serveWhileStale.nanos

`int32` · optional (explicit presence)

Fraction of a second at nanosecond resolution (0 to 999,999,999).
Durations under one second use seconds = 0 and a positive nanos.

- rule: {"int32":{"lte":999999999,"gte":0}}

### spec.pathMatchers[].defaultRouteAction.cachePolicy.negativeCachingPolicy

`[]GcpUrlMapNegativeCachingPolicy`

Per-status TTLs for cached negative responses; only meaningful with
negative_caching enabled.

### spec.pathMatchers[].defaultRouteAction.cachePolicy.negativeCachingPolicy[].code

`int32`

The HTTP status code this TTL applies to. GCP accepts only 300, 301,
302, 307, 308, 404, 405, 410, 421, 451, and 501, each at most once.

- rule: code must be one of 300, 301, 302, 307, 308, 404, 405, 410, 421, 451, 501

### spec.pathMatchers[].defaultRouteAction.cachePolicy.negativeCachingPolicy[].ttl

`GcpUrlMapDuration`

How long responses with this status are cached.

### spec.pathMatchers[].defaultRouteAction.cachePolicy.negativeCachingPolicy[].ttl.seconds

`int64` · optional (explicit presence)

Whole seconds (0 to 315,576,000,000 — GCP's int64 Duration bound).

- rule: {"int64":{"lte":"315576000000","gte":"0"}}

### spec.pathMatchers[].defaultRouteAction.cachePolicy.negativeCachingPolicy[].ttl.nanos

`int32` · optional (explicit presence)

Fraction of a second at nanosecond resolution (0 to 999,999,999).
Durations under one second use seconds = 0 and a positive nanos.

- rule: {"int32":{"lte":999999999,"gte":0}}

### spec.pathMatchers[].defaultCustomErrorResponsePolicy

`GcpUrlMapCustomErrorResponsePolicy`

Custom error pages for this path matcher's default. Global external ALBs
only — rejected when the map's region is set.

### spec.pathMatchers[].defaultCustomErrorResponsePolicy.errorService

`string | valueFrom`

The backend bucket serving the error pages. Reference a GcpBackendBucket
or provide a self-link directly.

- references: GcpBackendBucket (`status.outputs.self_link`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpBackendBucket, name: <that resource's name>, fieldPath: status.outputs.self_link}} -- a bare string does not parse

### spec.pathMatchers[].defaultCustomErrorResponsePolicy.errorResponseRules

`[]GcpUrlMapCustomErrorResponseRule`

Rules mapping response codes to the error page and (optionally) an
overridden status code.

### spec.pathMatchers[].defaultCustomErrorResponsePolicy.errorResponseRules[].matchResponseCodes

`[]string` · required

Response codes this rule matches: exact ("404", "503") or a class ("4xx",
"5xx"). At least one is required.

- rule: {"repeated":{"minItems":"1","items":{"string":{"minLen":"1"}}}}

### spec.pathMatchers[].defaultCustomErrorResponsePolicy.errorResponseRules[].overrideResponseCode

`int32`

Override the response code returned to the client (e.g. serve a 200 with a
maintenance page). Empty keeps the original code.

- rule: override_response_code must be between 200 and 599

### spec.pathMatchers[].defaultCustomErrorResponsePolicy.errorResponseRules[].path

`string` · required

Path within the error backend bucket to serve for matched codes (e.g.
"/errors/404.html").

- rule: {"required":true,"string":{"minLen":"1","maxLen":"1024"}}

### spec.pathMatchers[].description

`string`

What this path matcher covers.

- rule: {"string":{"maxLen":"1024"}}

### spec.pathMatchers[].headerAction

`GcpUrlMapHeaderAction`

Header mutations applied to all traffic through this path matcher.

### spec.pathMatchers[].headerAction.requestHeadersToAdd

`[]GcpUrlMapHeaderValue`

Headers to add to the request before it reaches the backend.

### spec.pathMatchers[].headerAction.requestHeadersToAdd[].headerName

`string` · required

The header name (e.g. "X-Client-Geo").

- rule: {"required":true,"string":{"maxLen":"1024"}}

### spec.pathMatchers[].headerAction.requestHeadersToAdd[].headerValue

`string` · required

The header value. May use load-balancer variables (e.g. "{client_region}").

- rule: {"required":true,"string":{"maxLen":"1024"}}

### spec.pathMatchers[].headerAction.requestHeadersToAdd[].replace

`bool`

Replace an existing header of the same name (true) or append to it
(false).

### spec.pathMatchers[].headerAction.requestHeadersToRemove

`[]string`

Header names to strip from the request before it reaches the backend.

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.pathMatchers[].headerAction.responseHeadersToAdd

`[]GcpUrlMapHeaderValue`

Headers to add to the response before it returns to the client.

### spec.pathMatchers[].headerAction.responseHeadersToAdd[].headerName

`string` · required

The header name (e.g. "X-Client-Geo").

- rule: {"required":true,"string":{"maxLen":"1024"}}

### spec.pathMatchers[].headerAction.responseHeadersToAdd[].headerValue

`string` · required

The header value. May use load-balancer variables (e.g. "{client_region}").

- rule: {"required":true,"string":{"maxLen":"1024"}}

### spec.pathMatchers[].headerAction.responseHeadersToAdd[].replace

`bool`

Replace an existing header of the same name (true) or append to it
(false).

### spec.pathMatchers[].headerAction.responseHeadersToRemove

`[]string`

Header names to strip from the response before it returns to the client.

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.pathMatchers[].pathRules

`[]GcpUrlMapPathRule`

Longest-prefix path rules: each maps a set of path patterns to a service,
redirect, or route action. Evaluated after route_rules.

- rule: a path rule must have exactly one target: service, url_redirect, or route_action with weighted_backend_services (a route action carrying only sub-policies may accompany service)
- rule: route_action and url_redirect are mutually exclusive — a redirect never reaches a backend
- rule: path_template_rewrite is honored only inside a route rule's route_action — GCP rejects it in a path rule's route action

### spec.pathMatchers[].pathRules[].paths

`[]string` · required

Path patterns this rule matches. Each must start with "/" and may end with
a single "*" wildcard (e.g. "/api/*"). At least one required.

- rule: {"repeated":{"minItems":"1","items":{"string":{"minLen":"1"}}}}

### spec.pathMatchers[].pathRules[].service

`string | valueFrom`

The target when a path matches — a backend service or backend bucket.
Reference a GcpBackendService or GcpBackendBucket, or provide a self-link.
Set exactly one of service, url_redirect, or route_action.

- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.pathMatchers[].pathRules[].routeAction

`GcpUrlMapRouteAction`

Advanced handling (weighted split / rewrite) for matched paths.

### spec.pathMatchers[].pathRules[].routeAction.weightedBackendServices

`[]GcpUrlMapWeightedBackendService`

Split traffic across multiple backend services by weight — the mechanism
for weighted canary and blue/green rollouts. The weights are relative; a
backend's share is its weight over the sum of weights.

### spec.pathMatchers[].pathRules[].routeAction.weightedBackendServices[].backendService

`string | valueFrom` · required

The backend service receiving this share of traffic. Reference a
GcpBackendService or provide a self-link directly.

- references: GcpBackendService (`status.outputs.self_link`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpBackendService, name: <that resource's name>, fieldPath: status.outputs.self_link}} -- a bare string does not parse

### spec.pathMatchers[].pathRules[].routeAction.weightedBackendServices[].weight

`int32`

Relative weight of this backend (0-1000). Its share is weight over the sum
of all weights in the split; 0 drains this backend from the split.

- rule: {"int32":{"lte":1000,"gte":0}}

### spec.pathMatchers[].pathRules[].routeAction.weightedBackendServices[].headerAction

`GcpUrlMapHeaderAction`

Header mutations applied ONLY to the share of traffic sent to this
backend — e.g. tag canary responses with an identifying header so
clients and dashboards can tell which arm served them. Applied after
the URL-map-level and route-level header actions.

### spec.pathMatchers[].pathRules[].routeAction.weightedBackendServices[].headerAction.requestHeadersToAdd

`[]GcpUrlMapHeaderValue`

Headers to add to the request before it reaches the backend.

### spec.pathMatchers[].pathRules[].routeAction.weightedBackendServices[].headerAction.requestHeadersToAdd[].headerName

`string` · required

The header name (e.g. "X-Client-Geo").

- rule: {"required":true,"string":{"maxLen":"1024"}}

### spec.pathMatchers[].pathRules[].routeAction.weightedBackendServices[].headerAction.requestHeadersToAdd[].headerValue

`string` · required

The header value. May use load-balancer variables (e.g. "{client_region}").

- rule: {"required":true,"string":{"maxLen":"1024"}}

### spec.pathMatchers[].pathRules[].routeAction.weightedBackendServices[].headerAction.requestHeadersToAdd[].replace

`bool`

Replace an existing header of the same name (true) or append to it
(false).

### spec.pathMatchers[].pathRules[].routeAction.weightedBackendServices[].headerAction.requestHeadersToRemove

`[]string`

Header names to strip from the request before it reaches the backend.

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.pathMatchers[].pathRules[].routeAction.weightedBackendServices[].headerAction.responseHeadersToAdd

`[]GcpUrlMapHeaderValue`

Headers to add to the response before it returns to the client.

### spec.pathMatchers[].pathRules[].routeAction.weightedBackendServices[].headerAction.responseHeadersToAdd[].headerName

`string` · required

The header name (e.g. "X-Client-Geo").

- rule: {"required":true,"string":{"maxLen":"1024"}}

### spec.pathMatchers[].pathRules[].routeAction.weightedBackendServices[].headerAction.responseHeadersToAdd[].headerValue

`string` · required

The header value. May use load-balancer variables (e.g. "{client_region}").

- rule: {"required":true,"string":{"maxLen":"1024"}}

### spec.pathMatchers[].pathRules[].routeAction.weightedBackendServices[].headerAction.responseHeadersToAdd[].replace

`bool`

Replace an existing header of the same name (true) or append to it
(false).

### spec.pathMatchers[].pathRules[].routeAction.weightedBackendServices[].headerAction.responseHeadersToRemove

`[]string`

Header names to strip from the response before it returns to the client.

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.pathMatchers[].pathRules[].routeAction.urlRewrite

`GcpUrlMapUrlRewrite`

Rewrite the host and/or path before forwarding to the backend.

- rule: path_prefix_rewrite and path_template_rewrite are mutually exclusive — set at most one

### spec.pathMatchers[].pathRules[].routeAction.urlRewrite.hostRewrite

`string`

Replace the request Host header with this value before forwarding.

- rule: {"string":{"maxLen":"255"}}

### spec.pathMatchers[].pathRules[].routeAction.urlRewrite.pathPrefixRewrite

`string`

Replace the matched path prefix with this value. Mutually exclusive with
path_template_rewrite.

- rule: {"string":{"maxLen":"1024"}}

### spec.pathMatchers[].pathRules[].routeAction.urlRewrite.pathTemplateRewrite

`string`

Rewrite the path using a template that references named path variables
captured by a route rule's path_template_match (e.g. "/v2/{country}").
Honored inside a route_rule's route_action on both scopes and, on a
REGIONAL map only, in a path matcher's default route action — GCP
rejects it in the map's default route action and in path-rule route
actions everywhere. Mutually exclusive with path_prefix_rewrite.

- rule: {"string":{"maxLen":"1024"}}

### spec.pathMatchers[].pathRules[].routeAction.timeout

`GcpUrlMapDuration`

Total time budget for the request, INCLUDING all retries (from the
first byte of the request to the last byte of the response). Pair with
retry_policy.per_try_timeout: per-try bounds one attempt, this bounds
the whole exchange. Unset uses the backend service's own timeout.
Not permitted when the route targets a redirect.

### spec.pathMatchers[].pathRules[].routeAction.timeout.seconds

`int64` · optional (explicit presence)

Whole seconds (0 to 315,576,000,000 — GCP's int64 Duration bound).

- rule: {"int64":{"lte":"315576000000","gte":"0"}}

### spec.pathMatchers[].pathRules[].routeAction.timeout.nanos

`int32` · optional (explicit presence)

Fraction of a second at nanosecond resolution (0 to 999,999,999).
Durations under one second use seconds = 0 and a positive nanos.

- rule: {"int32":{"lte":999999999,"gte":0}}

### spec.pathMatchers[].pathRules[].routeAction.retryPolicy

`GcpUrlMapRetryPolicy`

Retry failed requests to the backend. Which failures count is chosen
by retry_conditions; how long each attempt may run by per_try_timeout.
Retries consume the overall timeout's budget — they never extend it.

### spec.pathMatchers[].pathRules[].routeAction.retryPolicy.numRetries

`int32`

Number of allowed retries (GCP defaults to 1 when unset/0). Each retry
still spends the route's overall timeout budget.

- rule: {"int32":{"gte":0}}

### spec.pathMatchers[].pathRules[].routeAction.retryPolicy.retryConditions

`[]string`

Which failures trigger a retry. 5xx (any 5xx or no response at all),
gateway-error (502/503/504 only), connect-failure, retriable-4xx
(currently only 409), refused-stream, and the gRPC status conditions
cancelled, deadline-exceeded, resource-exhausted, unavailable.

- rule: each retry condition must be one of: 5xx, gateway-error, connect-failure, retriable-4xx, refused-stream, cancelled, deadline-exceeded, resource-exhausted, unavailable

### spec.pathMatchers[].pathRules[].routeAction.retryPolicy.perTryTimeout

`GcpUrlMapDuration`

Time budget for EACH retry attempt (the route's timeout bounds the
whole exchange across attempts). Unset uses the overall timeout for
every attempt.

### spec.pathMatchers[].pathRules[].routeAction.retryPolicy.perTryTimeout.seconds

`int64` · optional (explicit presence)

Whole seconds (0 to 315,576,000,000 — GCP's int64 Duration bound).

- rule: {"int64":{"lte":"315576000000","gte":"0"}}

### spec.pathMatchers[].pathRules[].routeAction.retryPolicy.perTryTimeout.nanos

`int32` · optional (explicit presence)

Fraction of a second at nanosecond resolution (0 to 999,999,999).
Durations under one second use seconds = 0 and a positive nanos.

- rule: {"int32":{"lte":999999999,"gte":0}}

### spec.pathMatchers[].pathRules[].routeAction.requestMirrorPolicy

`GcpUrlMapRequestMirrorPolicy`

Mirror every matched request to a second backend service, fire-and-
forget (responses from the mirror are discarded; the client sees only
the primary's response). Useful for shadow-testing a new stack with
production traffic — the mirror backend must be sized for the full
mirrored load.

### spec.pathMatchers[].pathRules[].routeAction.requestMirrorPolicy.backendService

`string | valueFrom` · required

The backend service receiving the mirrored copy of every matched
request. Reference a GcpBackendService or provide a self-link
directly. Responses from this backend are discarded.

- references: GcpBackendService (`status.outputs.self_link`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpBackendService, name: <that resource's name>, fieldPath: status.outputs.self_link}} -- a bare string does not parse

### spec.pathMatchers[].pathRules[].routeAction.corsPolicy

`GcpUrlMapCorsPolicy`

Answer cross-origin (CORS) preflights and stamp CORS headers at the
load balancer, before requests reach the backend.

### spec.pathMatchers[].pathRules[].routeAction.corsPolicy.allowCredentials

`bool`

Sets Access-Control-Allow-Credentials: allow requests carrying
credentials (cookies, authorization headers). Default false.

### spec.pathMatchers[].pathRules[].routeAction.corsPolicy.allowHeaders

`[]string`

Headers the client may send (Access-Control-Allow-Headers).

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.pathMatchers[].pathRules[].routeAction.corsPolicy.allowMethods

`[]string`

Methods the client may use (Access-Control-Allow-Methods), e.g.
["GET", "POST", "OPTIONS"].

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.pathMatchers[].pathRules[].routeAction.corsPolicy.allowOriginRegexes

`[]string`

Regular expressions matching allowed origins; a request origin is
allowed when it matches any listed regex or exact allow_origins entry.

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.pathMatchers[].pathRules[].routeAction.corsPolicy.allowOrigins

`[]string`

Exact origins allowed, e.g. "https://app.example.com".

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.pathMatchers[].pathRules[].routeAction.corsPolicy.disabled

`bool`

Disable this CORS policy without deleting its configuration (true
turns the policy off). Default false — the policy is in effect.

### spec.pathMatchers[].pathRules[].routeAction.corsPolicy.exposeHeaders

`[]string`

Headers the browser may read from the response
(Access-Control-Expose-Headers).

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.pathMatchers[].pathRules[].routeAction.corsPolicy.maxAge

`int32`

Seconds a browser may cache the preflight response
(Access-Control-Max-Age).

- rule: {"int32":{"gte":0}}

### spec.pathMatchers[].pathRules[].routeAction.faultInjectionPolicy

`GcpUrlMapFaultInjectionPolicy`

Deliberately inject failures (aborts with a chosen status, fixed
delays) into a percentage of matched requests — chaos/resilience
testing of the CLIENTS of this route. Never leave enabled on a
production default route: the injected failures are real to callers.

### spec.pathMatchers[].pathRules[].routeAction.faultInjectionPolicy.abort

`GcpUrlMapFaultAbort`

Abort a percentage of requests with a fixed HTTP status, before they
reach the backend.

### spec.pathMatchers[].pathRules[].routeAction.faultInjectionPolicy.abort.httpStatus

`int32`

HTTP status returned to aborted requests (200-599).

- rule: http_status must be between 200 and 599

### spec.pathMatchers[].pathRules[].routeAction.faultInjectionPolicy.abort.percentage

`double`

Percentage of requests aborted (0.0-100.0).

- rule: {"double":{"lte":100,"gte":0}}

### spec.pathMatchers[].pathRules[].routeAction.faultInjectionPolicy.delay

`GcpUrlMapFaultDelay`

Delay a percentage of requests by a fixed duration before forwarding.

### spec.pathMatchers[].pathRules[].routeAction.faultInjectionPolicy.delay.fixedDelay

`GcpUrlMapDuration`

How long delayed requests are held before forwarding.

### spec.pathMatchers[].pathRules[].routeAction.faultInjectionPolicy.delay.fixedDelay.seconds

`int64` · optional (explicit presence)

Whole seconds (0 to 315,576,000,000 — GCP's int64 Duration bound).

- rule: {"int64":{"lte":"315576000000","gte":"0"}}

### spec.pathMatchers[].pathRules[].routeAction.faultInjectionPolicy.delay.fixedDelay.nanos

`int32` · optional (explicit presence)

Fraction of a second at nanosecond resolution (0 to 999,999,999).
Durations under one second use seconds = 0 and a positive nanos.

- rule: {"int32":{"lte":999999999,"gte":0}}

### spec.pathMatchers[].pathRules[].routeAction.faultInjectionPolicy.delay.percentage

`double`

Percentage of requests delayed (0.0-100.0).

- rule: {"double":{"lte":100,"gte":0}}

### spec.pathMatchers[].pathRules[].routeAction.maxStreamDuration

`GcpUrlMapDuration`

Upper bound on how long a STREAM on this route may stay open
(gRPC/long-poll streams — distinct from timeout, which bounds a
request/response exchange). Live API truth: GCP rejects this field
unless the URL map's backend service uses the INTERNAL_SELF_MANAGED
(Traffic Director) load-balancing scheme ("Max stream duration is
only supported when UrlMap is used with BackendService whose Load
Balancing Scheme is INTERNAL_SELF_MANAGED") — leave it unset on
external application load balancers. Global maps only: the regional
map has no such argument (except in a path matcher's default route
action, which carries it on both scopes).

### spec.pathMatchers[].pathRules[].routeAction.maxStreamDuration.seconds

`int64` · optional (explicit presence)

Whole seconds (0 to 315,576,000,000 — GCP's int64 Duration bound).

- rule: {"int64":{"lte":"315576000000","gte":"0"}}

### spec.pathMatchers[].pathRules[].routeAction.maxStreamDuration.nanos

`int32` · optional (explicit presence)

Fraction of a second at nanosecond resolution (0 to 999,999,999).
Durations under one second use seconds = 0 and a positive nanos.

- rule: {"int32":{"lte":999999999,"gte":0}}

### spec.pathMatchers[].pathRules[].routeAction.cachePolicy

`GcpUrlMapCachePolicy`

Cloud CDN caching for the routes using this action — overrides the
backend service's cdn_policy for matching traffic only. Takes effect
only when the target backend service (or bucket) has CDN enabled;
GCP ignores it otherwise. Global maps only — regional Application Load
Balancers have no Cloud CDN, and the regional map has no such argument.

- rule: with cache_mode USE_ORIGIN_HEADERS the origin's headers control lifetimes — remove client_ttl, default_ttl, and max_ttl (GCP would silently ignore them)

### spec.pathMatchers[].pathRules[].routeAction.cachePolicy.cacheMode

`string`

What gets cached. CACHE_ALL_STATIC (the GCP default) caches static
content types and honors origin cache headers for the rest;
USE_ORIGIN_HEADERS caches only what the backends explicitly mark
cacheable (TTL fields must be unset — the origin controls lifetimes);
FORCE_CACHE_ALL caches everything, ignoring origin headers (never
combine with private or per-user content).

- rule: cache_mode must be CACHE_ALL_STATIC, USE_ORIGIN_HEADERS, or FORCE_CACHE_ALL

### spec.pathMatchers[].pathRules[].routeAction.cachePolicy.cacheBypassRequestHeaderNames

`[]string`

Requests carrying any of these headers bypass the cache and go
straight to the origin (e.g. a debug or authorization header).

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.pathMatchers[].pathRules[].routeAction.cachePolicy.negativeCaching

`bool`

Cache negative responses (404, 410, ...) so repeated misses do not
hammer the backends. Pair with negative_caching_policy to set
per-status TTLs.

### spec.pathMatchers[].pathRules[].routeAction.cachePolicy.requestCoalescing

`bool`

Collapse concurrent cache-fill requests for the same key into one
origin fetch. Default true on GCP's side.

### spec.pathMatchers[].pathRules[].routeAction.cachePolicy.cacheKeyPolicy

`GcpUrlMapCacheKeyPolicy`

What forms the cache key — trim protocol, host, query string, or
select specific query parameters, headers, and cookies.

- rule: included_query_parameters and excluded_query_parameters are mutually exclusive — set at most one list

### spec.pathMatchers[].pathRules[].routeAction.cachePolicy.cacheKeyPolicy.excludedQueryParameters

`[]string`

Query parameters EXCLUDED from the cache key (everything else is
included). Mutually exclusive with included_query_parameters.

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.pathMatchers[].pathRules[].routeAction.cachePolicy.cacheKeyPolicy.includeHost

`bool`

Include the request host in the cache key (distinct hosts cache
separately).

### spec.pathMatchers[].pathRules[].routeAction.cachePolicy.cacheKeyPolicy.includeProtocol

`bool`

Include the protocol (http/https) in the cache key.

### spec.pathMatchers[].pathRules[].routeAction.cachePolicy.cacheKeyPolicy.includeQueryString

`bool`

Include the entire query string in the cache key. When false, the
included/excluded parameter lists refine what participates.

### spec.pathMatchers[].pathRules[].routeAction.cachePolicy.cacheKeyPolicy.includedCookieNames

`[]string`

Cookie names whose values join the cache key.

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.pathMatchers[].pathRules[].routeAction.cachePolicy.cacheKeyPolicy.includedHeaderNames

`[]string`

Header names whose values join the cache key.

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.pathMatchers[].pathRules[].routeAction.cachePolicy.cacheKeyPolicy.includedQueryParameters

`[]string`

Query parameters INCLUDED in the cache key (everything else is
ignored). Mutually exclusive with excluded_query_parameters.

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.pathMatchers[].pathRules[].routeAction.cachePolicy.clientTtl

`GcpUrlMapDuration`

Max lifetime in browsers and downstream caches (sets the max-age
clients see). Not respected with cache_mode USE_ORIGIN_HEADERS.

### spec.pathMatchers[].pathRules[].routeAction.cachePolicy.clientTtl.seconds

`int64` · optional (explicit presence)

Whole seconds (0 to 315,576,000,000 — GCP's int64 Duration bound).

- rule: {"int64":{"lte":"315576000000","gte":"0"}}

### spec.pathMatchers[].pathRules[].routeAction.cachePolicy.clientTtl.nanos

`int32` · optional (explicit presence)

Fraction of a second at nanosecond resolution (0 to 999,999,999).
Durations under one second use seconds = 0 and a positive nanos.

- rule: {"int32":{"lte":999999999,"gte":0}}

### spec.pathMatchers[].pathRules[].routeAction.cachePolicy.defaultTtl

`GcpUrlMapDuration`

Edge-cache lifetime for responses without their own cache headers.
Not respected with cache_mode USE_ORIGIN_HEADERS.

### spec.pathMatchers[].pathRules[].routeAction.cachePolicy.defaultTtl.seconds

`int64` · optional (explicit presence)

Whole seconds (0 to 315,576,000,000 — GCP's int64 Duration bound).

- rule: {"int64":{"lte":"315576000000","gte":"0"}}

### spec.pathMatchers[].pathRules[].routeAction.cachePolicy.defaultTtl.nanos

`int32` · optional (explicit presence)

Fraction of a second at nanosecond resolution (0 to 999,999,999).
Durations under one second use seconds = 0 and a positive nanos.

- rule: {"int32":{"lte":999999999,"gte":0}}

### spec.pathMatchers[].pathRules[].routeAction.cachePolicy.maxTtl

`GcpUrlMapDuration`

Upper bound on any edge-cache lifetime, capping even origin-supplied
max-age. Not respected with cache_mode USE_ORIGIN_HEADERS or
FORCE_CACHE_ALL.

### spec.pathMatchers[].pathRules[].routeAction.cachePolicy.maxTtl.seconds

`int64` · optional (explicit presence)

Whole seconds (0 to 315,576,000,000 — GCP's int64 Duration bound).

- rule: {"int64":{"lte":"315576000000","gte":"0"}}

### spec.pathMatchers[].pathRules[].routeAction.cachePolicy.maxTtl.nanos

`int32` · optional (explicit presence)

Fraction of a second at nanosecond resolution (0 to 999,999,999).
Durations under one second use seconds = 0 and a positive nanos.

- rule: {"int32":{"lte":999999999,"gte":0}}

### spec.pathMatchers[].pathRules[].routeAction.cachePolicy.serveWhileStale

`GcpUrlMapDuration`

How long stale content may still be served while revalidating in the
background (up to 1 day).

### spec.pathMatchers[].pathRules[].routeAction.cachePolicy.serveWhileStale.seconds

`int64` · optional (explicit presence)

Whole seconds (0 to 315,576,000,000 — GCP's int64 Duration bound).

- rule: {"int64":{"lte":"315576000000","gte":"0"}}

### spec.pathMatchers[].pathRules[].routeAction.cachePolicy.serveWhileStale.nanos

`int32` · optional (explicit presence)

Fraction of a second at nanosecond resolution (0 to 999,999,999).
Durations under one second use seconds = 0 and a positive nanos.

- rule: {"int32":{"lte":999999999,"gte":0}}

### spec.pathMatchers[].pathRules[].routeAction.cachePolicy.negativeCachingPolicy

`[]GcpUrlMapNegativeCachingPolicy`

Per-status TTLs for cached negative responses; only meaningful with
negative_caching enabled.

### spec.pathMatchers[].pathRules[].routeAction.cachePolicy.negativeCachingPolicy[].code

`int32`

The HTTP status code this TTL applies to. GCP accepts only 300, 301,
302, 307, 308, 404, 405, 410, 421, 451, and 501, each at most once.

- rule: code must be one of 300, 301, 302, 307, 308, 404, 405, 410, 421, 451, 501

### spec.pathMatchers[].pathRules[].routeAction.cachePolicy.negativeCachingPolicy[].ttl

`GcpUrlMapDuration`

How long responses with this status are cached.

### spec.pathMatchers[].pathRules[].routeAction.cachePolicy.negativeCachingPolicy[].ttl.seconds

`int64` · optional (explicit presence)

Whole seconds (0 to 315,576,000,000 — GCP's int64 Duration bound).

- rule: {"int64":{"lte":"315576000000","gte":"0"}}

### spec.pathMatchers[].pathRules[].routeAction.cachePolicy.negativeCachingPolicy[].ttl.nanos

`int32` · optional (explicit presence)

Fraction of a second at nanosecond resolution (0 to 999,999,999).
Durations under one second use seconds = 0 and a positive nanos.

- rule: {"int32":{"lte":999999999,"gte":0}}

### spec.pathMatchers[].pathRules[].urlRedirect

`GcpUrlMapUrlRedirect`

Redirect matched paths instead of serving them.

- rule: path_redirect and prefix_redirect are mutually exclusive — set at most one

### spec.pathMatchers[].pathRules[].urlRedirect.hostRedirect

`string`

Replace the host in the redirect Location. Empty keeps the request host.

- rule: {"string":{"maxLen":"255"}}

### spec.pathMatchers[].pathRules[].urlRedirect.httpsRedirect

`bool`

Redirect to HTTPS (scheme becomes https). The standard http→https
upgrade. Default false.

### spec.pathMatchers[].pathRules[].urlRedirect.pathRedirect

`string`

Replace the entire path with this value. Mutually exclusive with
prefix_redirect. Empty keeps the request path.

- rule: {"string":{"maxLen":"1024"}}

### spec.pathMatchers[].pathRules[].urlRedirect.prefixRedirect

`string`

Replace the matched path prefix with this value, keeping the remainder.
Mutually exclusive with path_redirect.

- rule: {"string":{"maxLen":"1024"}}

### spec.pathMatchers[].pathRules[].urlRedirect.redirectResponseCode

`string`

The HTTP redirect status code: FOUND (302), MOVED_PERMANENTLY_DEFAULT
(301), PERMANENT_REDIRECT (308), SEE_OTHER (303), or TEMPORARY_REDIRECT
(307). Empty uses the GCP default (MOVED_PERMANENTLY_DEFAULT).

- rule: redirect_response_code must be one of FOUND, MOVED_PERMANENTLY_DEFAULT, PERMANENT_REDIRECT, SEE_OTHER, or TEMPORARY_REDIRECT

### spec.pathMatchers[].pathRules[].urlRedirect.stripQuery

`bool`

Drop the query string from the redirect Location. Default false (the query
string is preserved).

### spec.pathMatchers[].pathRules[].customErrorResponsePolicy

`GcpUrlMapCustomErrorResponsePolicy`

Custom error pages for matched paths. Global external ALBs only.

### spec.pathMatchers[].pathRules[].customErrorResponsePolicy.errorService

`string | valueFrom`

The backend bucket serving the error pages. Reference a GcpBackendBucket
or provide a self-link directly.

- references: GcpBackendBucket (`status.outputs.self_link`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpBackendBucket, name: <that resource's name>, fieldPath: status.outputs.self_link}} -- a bare string does not parse

### spec.pathMatchers[].pathRules[].customErrorResponsePolicy.errorResponseRules

`[]GcpUrlMapCustomErrorResponseRule`

Rules mapping response codes to the error page and (optionally) an
overridden status code.

### spec.pathMatchers[].pathRules[].customErrorResponsePolicy.errorResponseRules[].matchResponseCodes

`[]string` · required

Response codes this rule matches: exact ("404", "503") or a class ("4xx",
"5xx"). At least one is required.

- rule: {"repeated":{"minItems":"1","items":{"string":{"minLen":"1"}}}}

### spec.pathMatchers[].pathRules[].customErrorResponsePolicy.errorResponseRules[].overrideResponseCode

`int32`

Override the response code returned to the client (e.g. serve a 200 with a
maintenance page). Empty keeps the original code.

- rule: override_response_code must be between 200 and 599

### spec.pathMatchers[].pathRules[].customErrorResponsePolicy.errorResponseRules[].path

`string` · required

Path within the error backend bucket to serve for matched codes (e.g.
"/errors/404.html").

- rule: {"required":true,"string":{"minLen":"1","maxLen":"1024"}}

### spec.pathMatchers[].routeRules

`[]GcpUrlMapRouteRule`

Priority-ordered route rules with rich header/query/path matching.
Evaluated before path_rules. A path matcher uses either path_rules or
route_rules, not both.

- rule: a route rule must have exactly one target: service, url_redirect, or route_action with weighted_backend_services (a route action carrying only sub-policies may accompany service)
- rule: route_action and url_redirect are mutually exclusive — a redirect never reaches a backend

### spec.pathMatchers[].routeRules[].priority

`int32`

Evaluation priority (0 to 2147483647); lower numbers are evaluated first
and must be unique within the path matcher. Proto3 int32 has no presence,
so required is omitted — priority 0 is valid and must not be rejected as
"unset".

- rule: {"int32":{"gte":0}}

### spec.pathMatchers[].routeRules[].service

`string | valueFrom`

The target backend service when this rule matches. Reference a
GcpBackendService or provide a self-link. Route rules target backend
services only (not buckets). Set exactly one of service, url_redirect, or
route_action.

- references: GcpBackendService (`status.outputs.self_link`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpBackendService, name: <that resource's name>, fieldPath: status.outputs.self_link}} -- a bare string does not parse

### spec.pathMatchers[].routeRules[].matchRules

`[]GcpUrlMapRouteRuleMatchRule` · required

Conditions a request must meet for this rule to fire (path/header/query
matching). At least one match rule is required.

- rule: {"repeated":{"minItems":"1"}}
- rule: set at most one path matcher: prefix_match, full_path_match, regex_match, or path_template_match

### spec.pathMatchers[].routeRules[].matchRules[].prefixMatch

`string`

Match when the path starts with this prefix (e.g. "/api"). Mutually
exclusive with full_path_match, regex_match, and path_template_match.

- rule: {"string":{"maxLen":"1024"}}

### spec.pathMatchers[].routeRules[].matchRules[].fullPathMatch

`string`

Match when the path equals this exactly. Mutually exclusive with the other
path matchers.

- rule: {"string":{"maxLen":"1024"}}

### spec.pathMatchers[].routeRules[].matchRules[].regexMatch

`string`

Match the path against this regular expression. Mutually exclusive with
the other path matchers.

- rule: {"string":{"maxLen":"1024"}}

### spec.pathMatchers[].routeRules[].matchRules[].pathTemplateMatch

`string`

Match the path against a wildcard template capturing named variables
(e.g. "/v1/{country}/**"), usable by a route_action's
path_template_rewrite. Mutually exclusive with the other path matchers.

- rule: {"string":{"maxLen":"1024"}}

### spec.pathMatchers[].routeRules[].matchRules[].ignoreCase

`bool`

Case-insensitive path matching. Default false.

### spec.pathMatchers[].routeRules[].matchRules[].headerMatches

`[]GcpUrlMapHeaderMatch`

Match on request headers — all listed header matches must hold.

### spec.pathMatchers[].routeRules[].matchRules[].headerMatches[].headerName

`string` · required

The header name to match.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.pathMatchers[].routeRules[].matchRules[].headerMatches[].exactMatch

`string`

Match when the header equals this exactly.

### spec.pathMatchers[].routeRules[].matchRules[].headerMatches[].prefixMatch

`string`

Match when the header starts with this.

### spec.pathMatchers[].routeRules[].matchRules[].headerMatches[].suffixMatch

`string`

Match when the header ends with this.

### spec.pathMatchers[].routeRules[].matchRules[].headerMatches[].regexMatch

`string`

Match the header against this regular expression.

### spec.pathMatchers[].routeRules[].matchRules[].headerMatches[].presentMatch

`bool`

Match when the header is present (any value).

### spec.pathMatchers[].routeRules[].matchRules[].headerMatches[].rangeMatch

`GcpUrlMapHeaderMatchRange`

Match when the header's integer value falls in a range.

### spec.pathMatchers[].routeRules[].matchRules[].headerMatches[].rangeMatch.rangeStart

`int64`

Inclusive lower bound.

### spec.pathMatchers[].routeRules[].matchRules[].headerMatches[].rangeMatch.rangeEnd

`int64`

Exclusive upper bound.

### spec.pathMatchers[].routeRules[].matchRules[].headerMatches[].invertMatch

`bool`

Invert the whole match — fire when the condition does NOT hold. Default
false.

### spec.pathMatchers[].routeRules[].matchRules[].queryParameterMatches

`[]GcpUrlMapQueryParameterMatch`

Match on query parameters — all listed matches must hold.

### spec.pathMatchers[].routeRules[].matchRules[].queryParameterMatches[].name

`string` · required

The query parameter name to match.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.pathMatchers[].routeRules[].matchRules[].queryParameterMatches[].exactMatch

`string`

Match when the parameter equals this exactly.

### spec.pathMatchers[].routeRules[].matchRules[].queryParameterMatches[].presentMatch

`bool`

Match when the parameter is present (any value).

### spec.pathMatchers[].routeRules[].matchRules[].queryParameterMatches[].regexMatch

`string`

Match the parameter against this regular expression.

### spec.pathMatchers[].routeRules[].matchRules[].metadataFilters

`[]GcpUrlMapMetadataFilter`

Traffic Director metadata filters (xDS node metadata matching).

### spec.pathMatchers[].routeRules[].matchRules[].metadataFilters[].filterMatchCriteria

`string` · required

How the labels combine: MATCH_ALL (every label must match) or MATCH_ANY
(at least one).

- rule: filter_match_criteria must be MATCH_ALL or MATCH_ANY
- rule: {"required":true}

### spec.pathMatchers[].routeRules[].matchRules[].metadataFilters[].filterLabels

`[]GcpUrlMapMetadataFilterLabel` · required

The xDS node metadata labels to match against (1-64 entries).

- rule: {"repeated":{"minItems":"1","maxItems":"64"}}

### spec.pathMatchers[].routeRules[].matchRules[].metadataFilters[].filterLabels[].name

`string` · required

Label name.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.pathMatchers[].routeRules[].matchRules[].metadataFilters[].filterLabels[].value

`string` · required

Label value.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.pathMatchers[].routeRules[].routeAction

`GcpUrlMapRouteAction`

Advanced handling (weighted split / rewrite / retry) for matched requests.

### spec.pathMatchers[].routeRules[].routeAction.weightedBackendServices

`[]GcpUrlMapWeightedBackendService`

Split traffic across multiple backend services by weight — the mechanism
for weighted canary and blue/green rollouts. The weights are relative; a
backend's share is its weight over the sum of weights.

### spec.pathMatchers[].routeRules[].routeAction.weightedBackendServices[].backendService

`string | valueFrom` · required

The backend service receiving this share of traffic. Reference a
GcpBackendService or provide a self-link directly.

- references: GcpBackendService (`status.outputs.self_link`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpBackendService, name: <that resource's name>, fieldPath: status.outputs.self_link}} -- a bare string does not parse

### spec.pathMatchers[].routeRules[].routeAction.weightedBackendServices[].weight

`int32`

Relative weight of this backend (0-1000). Its share is weight over the sum
of all weights in the split; 0 drains this backend from the split.

- rule: {"int32":{"lte":1000,"gte":0}}

### spec.pathMatchers[].routeRules[].routeAction.weightedBackendServices[].headerAction

`GcpUrlMapHeaderAction`

Header mutations applied ONLY to the share of traffic sent to this
backend — e.g. tag canary responses with an identifying header so
clients and dashboards can tell which arm served them. Applied after
the URL-map-level and route-level header actions.

### spec.pathMatchers[].routeRules[].routeAction.weightedBackendServices[].headerAction.requestHeadersToAdd

`[]GcpUrlMapHeaderValue`

Headers to add to the request before it reaches the backend.

### spec.pathMatchers[].routeRules[].routeAction.weightedBackendServices[].headerAction.requestHeadersToAdd[].headerName

`string` · required

The header name (e.g. "X-Client-Geo").

- rule: {"required":true,"string":{"maxLen":"1024"}}

### spec.pathMatchers[].routeRules[].routeAction.weightedBackendServices[].headerAction.requestHeadersToAdd[].headerValue

`string` · required

The header value. May use load-balancer variables (e.g. "{client_region}").

- rule: {"required":true,"string":{"maxLen":"1024"}}

### spec.pathMatchers[].routeRules[].routeAction.weightedBackendServices[].headerAction.requestHeadersToAdd[].replace

`bool`

Replace an existing header of the same name (true) or append to it
(false).

### spec.pathMatchers[].routeRules[].routeAction.weightedBackendServices[].headerAction.requestHeadersToRemove

`[]string`

Header names to strip from the request before it reaches the backend.

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.pathMatchers[].routeRules[].routeAction.weightedBackendServices[].headerAction.responseHeadersToAdd

`[]GcpUrlMapHeaderValue`

Headers to add to the response before it returns to the client.

### spec.pathMatchers[].routeRules[].routeAction.weightedBackendServices[].headerAction.responseHeadersToAdd[].headerName

`string` · required

The header name (e.g. "X-Client-Geo").

- rule: {"required":true,"string":{"maxLen":"1024"}}

### spec.pathMatchers[].routeRules[].routeAction.weightedBackendServices[].headerAction.responseHeadersToAdd[].headerValue

`string` · required

The header value. May use load-balancer variables (e.g. "{client_region}").

- rule: {"required":true,"string":{"maxLen":"1024"}}

### spec.pathMatchers[].routeRules[].routeAction.weightedBackendServices[].headerAction.responseHeadersToAdd[].replace

`bool`

Replace an existing header of the same name (true) or append to it
(false).

### spec.pathMatchers[].routeRules[].routeAction.weightedBackendServices[].headerAction.responseHeadersToRemove

`[]string`

Header names to strip from the response before it returns to the client.

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.pathMatchers[].routeRules[].routeAction.urlRewrite

`GcpUrlMapUrlRewrite`

Rewrite the host and/or path before forwarding to the backend.

- rule: path_prefix_rewrite and path_template_rewrite are mutually exclusive — set at most one

### spec.pathMatchers[].routeRules[].routeAction.urlRewrite.hostRewrite

`string`

Replace the request Host header with this value before forwarding.

- rule: {"string":{"maxLen":"255"}}

### spec.pathMatchers[].routeRules[].routeAction.urlRewrite.pathPrefixRewrite

`string`

Replace the matched path prefix with this value. Mutually exclusive with
path_template_rewrite.

- rule: {"string":{"maxLen":"1024"}}

### spec.pathMatchers[].routeRules[].routeAction.urlRewrite.pathTemplateRewrite

`string`

Rewrite the path using a template that references named path variables
captured by a route rule's path_template_match (e.g. "/v2/{country}").
Honored inside a route_rule's route_action on both scopes and, on a
REGIONAL map only, in a path matcher's default route action — GCP
rejects it in the map's default route action and in path-rule route
actions everywhere. Mutually exclusive with path_prefix_rewrite.

- rule: {"string":{"maxLen":"1024"}}

### spec.pathMatchers[].routeRules[].routeAction.timeout

`GcpUrlMapDuration`

Total time budget for the request, INCLUDING all retries (from the
first byte of the request to the last byte of the response). Pair with
retry_policy.per_try_timeout: per-try bounds one attempt, this bounds
the whole exchange. Unset uses the backend service's own timeout.
Not permitted when the route targets a redirect.

### spec.pathMatchers[].routeRules[].routeAction.timeout.seconds

`int64` · optional (explicit presence)

Whole seconds (0 to 315,576,000,000 — GCP's int64 Duration bound).

- rule: {"int64":{"lte":"315576000000","gte":"0"}}

### spec.pathMatchers[].routeRules[].routeAction.timeout.nanos

`int32` · optional (explicit presence)

Fraction of a second at nanosecond resolution (0 to 999,999,999).
Durations under one second use seconds = 0 and a positive nanos.

- rule: {"int32":{"lte":999999999,"gte":0}}

### spec.pathMatchers[].routeRules[].routeAction.retryPolicy

`GcpUrlMapRetryPolicy`

Retry failed requests to the backend. Which failures count is chosen
by retry_conditions; how long each attempt may run by per_try_timeout.
Retries consume the overall timeout's budget — they never extend it.

### spec.pathMatchers[].routeRules[].routeAction.retryPolicy.numRetries

`int32`

Number of allowed retries (GCP defaults to 1 when unset/0). Each retry
still spends the route's overall timeout budget.

- rule: {"int32":{"gte":0}}

### spec.pathMatchers[].routeRules[].routeAction.retryPolicy.retryConditions

`[]string`

Which failures trigger a retry. 5xx (any 5xx or no response at all),
gateway-error (502/503/504 only), connect-failure, retriable-4xx
(currently only 409), refused-stream, and the gRPC status conditions
cancelled, deadline-exceeded, resource-exhausted, unavailable.

- rule: each retry condition must be one of: 5xx, gateway-error, connect-failure, retriable-4xx, refused-stream, cancelled, deadline-exceeded, resource-exhausted, unavailable

### spec.pathMatchers[].routeRules[].routeAction.retryPolicy.perTryTimeout

`GcpUrlMapDuration`

Time budget for EACH retry attempt (the route's timeout bounds the
whole exchange across attempts). Unset uses the overall timeout for
every attempt.

### spec.pathMatchers[].routeRules[].routeAction.retryPolicy.perTryTimeout.seconds

`int64` · optional (explicit presence)

Whole seconds (0 to 315,576,000,000 — GCP's int64 Duration bound).

- rule: {"int64":{"lte":"315576000000","gte":"0"}}

### spec.pathMatchers[].routeRules[].routeAction.retryPolicy.perTryTimeout.nanos

`int32` · optional (explicit presence)

Fraction of a second at nanosecond resolution (0 to 999,999,999).
Durations under one second use seconds = 0 and a positive nanos.

- rule: {"int32":{"lte":999999999,"gte":0}}

### spec.pathMatchers[].routeRules[].routeAction.requestMirrorPolicy

`GcpUrlMapRequestMirrorPolicy`

Mirror every matched request to a second backend service, fire-and-
forget (responses from the mirror are discarded; the client sees only
the primary's response). Useful for shadow-testing a new stack with
production traffic — the mirror backend must be sized for the full
mirrored load.

### spec.pathMatchers[].routeRules[].routeAction.requestMirrorPolicy.backendService

`string | valueFrom` · required

The backend service receiving the mirrored copy of every matched
request. Reference a GcpBackendService or provide a self-link
directly. Responses from this backend are discarded.

- references: GcpBackendService (`status.outputs.self_link`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpBackendService, name: <that resource's name>, fieldPath: status.outputs.self_link}} -- a bare string does not parse

### spec.pathMatchers[].routeRules[].routeAction.corsPolicy

`GcpUrlMapCorsPolicy`

Answer cross-origin (CORS) preflights and stamp CORS headers at the
load balancer, before requests reach the backend.

### spec.pathMatchers[].routeRules[].routeAction.corsPolicy.allowCredentials

`bool`

Sets Access-Control-Allow-Credentials: allow requests carrying
credentials (cookies, authorization headers). Default false.

### spec.pathMatchers[].routeRules[].routeAction.corsPolicy.allowHeaders

`[]string`

Headers the client may send (Access-Control-Allow-Headers).

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.pathMatchers[].routeRules[].routeAction.corsPolicy.allowMethods

`[]string`

Methods the client may use (Access-Control-Allow-Methods), e.g.
["GET", "POST", "OPTIONS"].

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.pathMatchers[].routeRules[].routeAction.corsPolicy.allowOriginRegexes

`[]string`

Regular expressions matching allowed origins; a request origin is
allowed when it matches any listed regex or exact allow_origins entry.

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.pathMatchers[].routeRules[].routeAction.corsPolicy.allowOrigins

`[]string`

Exact origins allowed, e.g. "https://app.example.com".

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.pathMatchers[].routeRules[].routeAction.corsPolicy.disabled

`bool`

Disable this CORS policy without deleting its configuration (true
turns the policy off). Default false — the policy is in effect.

### spec.pathMatchers[].routeRules[].routeAction.corsPolicy.exposeHeaders

`[]string`

Headers the browser may read from the response
(Access-Control-Expose-Headers).

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.pathMatchers[].routeRules[].routeAction.corsPolicy.maxAge

`int32`

Seconds a browser may cache the preflight response
(Access-Control-Max-Age).

- rule: {"int32":{"gte":0}}

### spec.pathMatchers[].routeRules[].routeAction.faultInjectionPolicy

`GcpUrlMapFaultInjectionPolicy`

Deliberately inject failures (aborts with a chosen status, fixed
delays) into a percentage of matched requests — chaos/resilience
testing of the CLIENTS of this route. Never leave enabled on a
production default route: the injected failures are real to callers.

### spec.pathMatchers[].routeRules[].routeAction.faultInjectionPolicy.abort

`GcpUrlMapFaultAbort`

Abort a percentage of requests with a fixed HTTP status, before they
reach the backend.

### spec.pathMatchers[].routeRules[].routeAction.faultInjectionPolicy.abort.httpStatus

`int32`

HTTP status returned to aborted requests (200-599).

- rule: http_status must be between 200 and 599

### spec.pathMatchers[].routeRules[].routeAction.faultInjectionPolicy.abort.percentage

`double`

Percentage of requests aborted (0.0-100.0).

- rule: {"double":{"lte":100,"gte":0}}

### spec.pathMatchers[].routeRules[].routeAction.faultInjectionPolicy.delay

`GcpUrlMapFaultDelay`

Delay a percentage of requests by a fixed duration before forwarding.

### spec.pathMatchers[].routeRules[].routeAction.faultInjectionPolicy.delay.fixedDelay

`GcpUrlMapDuration`

How long delayed requests are held before forwarding.

### spec.pathMatchers[].routeRules[].routeAction.faultInjectionPolicy.delay.fixedDelay.seconds

`int64` · optional (explicit presence)

Whole seconds (0 to 315,576,000,000 — GCP's int64 Duration bound).

- rule: {"int64":{"lte":"315576000000","gte":"0"}}

### spec.pathMatchers[].routeRules[].routeAction.faultInjectionPolicy.delay.fixedDelay.nanos

`int32` · optional (explicit presence)

Fraction of a second at nanosecond resolution (0 to 999,999,999).
Durations under one second use seconds = 0 and a positive nanos.

- rule: {"int32":{"lte":999999999,"gte":0}}

### spec.pathMatchers[].routeRules[].routeAction.faultInjectionPolicy.delay.percentage

`double`

Percentage of requests delayed (0.0-100.0).

- rule: {"double":{"lte":100,"gte":0}}

### spec.pathMatchers[].routeRules[].routeAction.maxStreamDuration

`GcpUrlMapDuration`

Upper bound on how long a STREAM on this route may stay open
(gRPC/long-poll streams — distinct from timeout, which bounds a
request/response exchange). Live API truth: GCP rejects this field
unless the URL map's backend service uses the INTERNAL_SELF_MANAGED
(Traffic Director) load-balancing scheme ("Max stream duration is
only supported when UrlMap is used with BackendService whose Load
Balancing Scheme is INTERNAL_SELF_MANAGED") — leave it unset on
external application load balancers. Global maps only: the regional
map has no such argument (except in a path matcher's default route
action, which carries it on both scopes).

### spec.pathMatchers[].routeRules[].routeAction.maxStreamDuration.seconds

`int64` · optional (explicit presence)

Whole seconds (0 to 315,576,000,000 — GCP's int64 Duration bound).

- rule: {"int64":{"lte":"315576000000","gte":"0"}}

### spec.pathMatchers[].routeRules[].routeAction.maxStreamDuration.nanos

`int32` · optional (explicit presence)

Fraction of a second at nanosecond resolution (0 to 999,999,999).
Durations under one second use seconds = 0 and a positive nanos.

- rule: {"int32":{"lte":999999999,"gte":0}}

### spec.pathMatchers[].routeRules[].routeAction.cachePolicy

`GcpUrlMapCachePolicy`

Cloud CDN caching for the routes using this action — overrides the
backend service's cdn_policy for matching traffic only. Takes effect
only when the target backend service (or bucket) has CDN enabled;
GCP ignores it otherwise. Global maps only — regional Application Load
Balancers have no Cloud CDN, and the regional map has no such argument.

- rule: with cache_mode USE_ORIGIN_HEADERS the origin's headers control lifetimes — remove client_ttl, default_ttl, and max_ttl (GCP would silently ignore them)

### spec.pathMatchers[].routeRules[].routeAction.cachePolicy.cacheMode

`string`

What gets cached. CACHE_ALL_STATIC (the GCP default) caches static
content types and honors origin cache headers for the rest;
USE_ORIGIN_HEADERS caches only what the backends explicitly mark
cacheable (TTL fields must be unset — the origin controls lifetimes);
FORCE_CACHE_ALL caches everything, ignoring origin headers (never
combine with private or per-user content).

- rule: cache_mode must be CACHE_ALL_STATIC, USE_ORIGIN_HEADERS, or FORCE_CACHE_ALL

### spec.pathMatchers[].routeRules[].routeAction.cachePolicy.cacheBypassRequestHeaderNames

`[]string`

Requests carrying any of these headers bypass the cache and go
straight to the origin (e.g. a debug or authorization header).

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.pathMatchers[].routeRules[].routeAction.cachePolicy.negativeCaching

`bool`

Cache negative responses (404, 410, ...) so repeated misses do not
hammer the backends. Pair with negative_caching_policy to set
per-status TTLs.

### spec.pathMatchers[].routeRules[].routeAction.cachePolicy.requestCoalescing

`bool`

Collapse concurrent cache-fill requests for the same key into one
origin fetch. Default true on GCP's side.

### spec.pathMatchers[].routeRules[].routeAction.cachePolicy.cacheKeyPolicy

`GcpUrlMapCacheKeyPolicy`

What forms the cache key — trim protocol, host, query string, or
select specific query parameters, headers, and cookies.

- rule: included_query_parameters and excluded_query_parameters are mutually exclusive — set at most one list

### spec.pathMatchers[].routeRules[].routeAction.cachePolicy.cacheKeyPolicy.excludedQueryParameters

`[]string`

Query parameters EXCLUDED from the cache key (everything else is
included). Mutually exclusive with included_query_parameters.

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.pathMatchers[].routeRules[].routeAction.cachePolicy.cacheKeyPolicy.includeHost

`bool`

Include the request host in the cache key (distinct hosts cache
separately).

### spec.pathMatchers[].routeRules[].routeAction.cachePolicy.cacheKeyPolicy.includeProtocol

`bool`

Include the protocol (http/https) in the cache key.

### spec.pathMatchers[].routeRules[].routeAction.cachePolicy.cacheKeyPolicy.includeQueryString

`bool`

Include the entire query string in the cache key. When false, the
included/excluded parameter lists refine what participates.

### spec.pathMatchers[].routeRules[].routeAction.cachePolicy.cacheKeyPolicy.includedCookieNames

`[]string`

Cookie names whose values join the cache key.

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.pathMatchers[].routeRules[].routeAction.cachePolicy.cacheKeyPolicy.includedHeaderNames

`[]string`

Header names whose values join the cache key.

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.pathMatchers[].routeRules[].routeAction.cachePolicy.cacheKeyPolicy.includedQueryParameters

`[]string`

Query parameters INCLUDED in the cache key (everything else is
ignored). Mutually exclusive with excluded_query_parameters.

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.pathMatchers[].routeRules[].routeAction.cachePolicy.clientTtl

`GcpUrlMapDuration`

Max lifetime in browsers and downstream caches (sets the max-age
clients see). Not respected with cache_mode USE_ORIGIN_HEADERS.

### spec.pathMatchers[].routeRules[].routeAction.cachePolicy.clientTtl.seconds

`int64` · optional (explicit presence)

Whole seconds (0 to 315,576,000,000 — GCP's int64 Duration bound).

- rule: {"int64":{"lte":"315576000000","gte":"0"}}

### spec.pathMatchers[].routeRules[].routeAction.cachePolicy.clientTtl.nanos

`int32` · optional (explicit presence)

Fraction of a second at nanosecond resolution (0 to 999,999,999).
Durations under one second use seconds = 0 and a positive nanos.

- rule: {"int32":{"lte":999999999,"gte":0}}

### spec.pathMatchers[].routeRules[].routeAction.cachePolicy.defaultTtl

`GcpUrlMapDuration`

Edge-cache lifetime for responses without their own cache headers.
Not respected with cache_mode USE_ORIGIN_HEADERS.

### spec.pathMatchers[].routeRules[].routeAction.cachePolicy.defaultTtl.seconds

`int64` · optional (explicit presence)

Whole seconds (0 to 315,576,000,000 — GCP's int64 Duration bound).

- rule: {"int64":{"lte":"315576000000","gte":"0"}}

### spec.pathMatchers[].routeRules[].routeAction.cachePolicy.defaultTtl.nanos

`int32` · optional (explicit presence)

Fraction of a second at nanosecond resolution (0 to 999,999,999).
Durations under one second use seconds = 0 and a positive nanos.

- rule: {"int32":{"lte":999999999,"gte":0}}

### spec.pathMatchers[].routeRules[].routeAction.cachePolicy.maxTtl

`GcpUrlMapDuration`

Upper bound on any edge-cache lifetime, capping even origin-supplied
max-age. Not respected with cache_mode USE_ORIGIN_HEADERS or
FORCE_CACHE_ALL.

### spec.pathMatchers[].routeRules[].routeAction.cachePolicy.maxTtl.seconds

`int64` · optional (explicit presence)

Whole seconds (0 to 315,576,000,000 — GCP's int64 Duration bound).

- rule: {"int64":{"lte":"315576000000","gte":"0"}}

### spec.pathMatchers[].routeRules[].routeAction.cachePolicy.maxTtl.nanos

`int32` · optional (explicit presence)

Fraction of a second at nanosecond resolution (0 to 999,999,999).
Durations under one second use seconds = 0 and a positive nanos.

- rule: {"int32":{"lte":999999999,"gte":0}}

### spec.pathMatchers[].routeRules[].routeAction.cachePolicy.serveWhileStale

`GcpUrlMapDuration`

How long stale content may still be served while revalidating in the
background (up to 1 day).

### spec.pathMatchers[].routeRules[].routeAction.cachePolicy.serveWhileStale.seconds

`int64` · optional (explicit presence)

Whole seconds (0 to 315,576,000,000 — GCP's int64 Duration bound).

- rule: {"int64":{"lte":"315576000000","gte":"0"}}

### spec.pathMatchers[].routeRules[].routeAction.cachePolicy.serveWhileStale.nanos

`int32` · optional (explicit presence)

Fraction of a second at nanosecond resolution (0 to 999,999,999).
Durations under one second use seconds = 0 and a positive nanos.

- rule: {"int32":{"lte":999999999,"gte":0}}

### spec.pathMatchers[].routeRules[].routeAction.cachePolicy.negativeCachingPolicy

`[]GcpUrlMapNegativeCachingPolicy`

Per-status TTLs for cached negative responses; only meaningful with
negative_caching enabled.

### spec.pathMatchers[].routeRules[].routeAction.cachePolicy.negativeCachingPolicy[].code

`int32`

The HTTP status code this TTL applies to. GCP accepts only 300, 301,
302, 307, 308, 404, 405, 410, 421, 451, and 501, each at most once.

- rule: code must be one of 300, 301, 302, 307, 308, 404, 405, 410, 421, 451, 501

### spec.pathMatchers[].routeRules[].routeAction.cachePolicy.negativeCachingPolicy[].ttl

`GcpUrlMapDuration`

How long responses with this status are cached.

### spec.pathMatchers[].routeRules[].routeAction.cachePolicy.negativeCachingPolicy[].ttl.seconds

`int64` · optional (explicit presence)

Whole seconds (0 to 315,576,000,000 — GCP's int64 Duration bound).

- rule: {"int64":{"lte":"315576000000","gte":"0"}}

### spec.pathMatchers[].routeRules[].routeAction.cachePolicy.negativeCachingPolicy[].ttl.nanos

`int32` · optional (explicit presence)

Fraction of a second at nanosecond resolution (0 to 999,999,999).
Durations under one second use seconds = 0 and a positive nanos.

- rule: {"int32":{"lte":999999999,"gte":0}}

### spec.pathMatchers[].routeRules[].urlRedirect

`GcpUrlMapUrlRedirect`

Redirect matched requests instead of serving them.

- rule: path_redirect and prefix_redirect are mutually exclusive — set at most one

### spec.pathMatchers[].routeRules[].urlRedirect.hostRedirect

`string`

Replace the host in the redirect Location. Empty keeps the request host.

- rule: {"string":{"maxLen":"255"}}

### spec.pathMatchers[].routeRules[].urlRedirect.httpsRedirect

`bool`

Redirect to HTTPS (scheme becomes https). The standard http→https
upgrade. Default false.

### spec.pathMatchers[].routeRules[].urlRedirect.pathRedirect

`string`

Replace the entire path with this value. Mutually exclusive with
prefix_redirect. Empty keeps the request path.

- rule: {"string":{"maxLen":"1024"}}

### spec.pathMatchers[].routeRules[].urlRedirect.prefixRedirect

`string`

Replace the matched path prefix with this value, keeping the remainder.
Mutually exclusive with path_redirect.

- rule: {"string":{"maxLen":"1024"}}

### spec.pathMatchers[].routeRules[].urlRedirect.redirectResponseCode

`string`

The HTTP redirect status code: FOUND (302), MOVED_PERMANENTLY_DEFAULT
(301), PERMANENT_REDIRECT (308), SEE_OTHER (303), or TEMPORARY_REDIRECT
(307). Empty uses the GCP default (MOVED_PERMANENTLY_DEFAULT).

- rule: redirect_response_code must be one of FOUND, MOVED_PERMANENTLY_DEFAULT, PERMANENT_REDIRECT, SEE_OTHER, or TEMPORARY_REDIRECT

### spec.pathMatchers[].routeRules[].urlRedirect.stripQuery

`bool`

Drop the query string from the redirect Location. Default false (the query
string is preserved).

### spec.pathMatchers[].routeRules[].headerAction

`GcpUrlMapHeaderAction`

Header mutations applied to matched requests.

### spec.pathMatchers[].routeRules[].headerAction.requestHeadersToAdd

`[]GcpUrlMapHeaderValue`

Headers to add to the request before it reaches the backend.

### spec.pathMatchers[].routeRules[].headerAction.requestHeadersToAdd[].headerName

`string` · required

The header name (e.g. "X-Client-Geo").

- rule: {"required":true,"string":{"maxLen":"1024"}}

### spec.pathMatchers[].routeRules[].headerAction.requestHeadersToAdd[].headerValue

`string` · required

The header value. May use load-balancer variables (e.g. "{client_region}").

- rule: {"required":true,"string":{"maxLen":"1024"}}

### spec.pathMatchers[].routeRules[].headerAction.requestHeadersToAdd[].replace

`bool`

Replace an existing header of the same name (true) or append to it
(false).

### spec.pathMatchers[].routeRules[].headerAction.requestHeadersToRemove

`[]string`

Header names to strip from the request before it reaches the backend.

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.pathMatchers[].routeRules[].headerAction.responseHeadersToAdd

`[]GcpUrlMapHeaderValue`

Headers to add to the response before it returns to the client.

### spec.pathMatchers[].routeRules[].headerAction.responseHeadersToAdd[].headerName

`string` · required

The header name (e.g. "X-Client-Geo").

- rule: {"required":true,"string":{"maxLen":"1024"}}

### spec.pathMatchers[].routeRules[].headerAction.responseHeadersToAdd[].headerValue

`string` · required

The header value. May use load-balancer variables (e.g. "{client_region}").

- rule: {"required":true,"string":{"maxLen":"1024"}}

### spec.pathMatchers[].routeRules[].headerAction.responseHeadersToAdd[].replace

`bool`

Replace an existing header of the same name (true) or append to it
(false).

### spec.pathMatchers[].routeRules[].headerAction.responseHeadersToRemove

`[]string`

Header names to strip from the response before it returns to the client.

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.pathMatchers[].routeRules[].customErrorResponsePolicy

`GcpUrlMapCustomErrorResponsePolicy`

Custom error pages for matched requests. Global external ALBs only.

### spec.pathMatchers[].routeRules[].customErrorResponsePolicy.errorService

`string | valueFrom`

The backend bucket serving the error pages. Reference a GcpBackendBucket
or provide a self-link directly.

- references: GcpBackendBucket (`status.outputs.self_link`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpBackendBucket, name: <that resource's name>, fieldPath: status.outputs.self_link}} -- a bare string does not parse

### spec.pathMatchers[].routeRules[].customErrorResponsePolicy.errorResponseRules

`[]GcpUrlMapCustomErrorResponseRule`

Rules mapping response codes to the error page and (optionally) an
overridden status code.

### spec.pathMatchers[].routeRules[].customErrorResponsePolicy.errorResponseRules[].matchResponseCodes

`[]string` · required

Response codes this rule matches: exact ("404", "503") or a class ("4xx",
"5xx"). At least one is required.

- rule: {"repeated":{"minItems":"1","items":{"string":{"minLen":"1"}}}}

### spec.pathMatchers[].routeRules[].customErrorResponsePolicy.errorResponseRules[].overrideResponseCode

`int32`

Override the response code returned to the client (e.g. serve a 200 with a
maintenance page). Empty keeps the original code.

- rule: override_response_code must be between 200 and 599

### spec.pathMatchers[].routeRules[].customErrorResponsePolicy.errorResponseRules[].path

`string` · required

Path within the error backend bucket to serve for matched codes (e.g.
"/errors/404.html").

- rule: {"required":true,"string":{"minLen":"1","maxLen":"1024"}}

### spec.tests

`[]GcpUrlMapTest`

Routing self-tests evaluated by GCP at create/update time: each asserts
that a given host+path resolves to an expected service or redirect. A
failing test blocks the update — a guard against a routing change that
silently breaks a path. On a regional map every test names its expected
service and carries no headers or redirect expectations (those forms
exist only on the global map). Mutable.

### spec.tests[].host

`string` · required

The request Host header the test sends.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.tests[].path

`string` · required

The request path the test sends.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.tests[].service

`string | valueFrom`

The backend service or backend bucket the request is expected to resolve
to. Reference a GcpBackendService or GcpBackendBucket, or provide a
self-link. Leave empty when asserting a redirect via
expected_redirect_response_code. Required on a regional map, whose
tests can only assert a service.

- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.tests[].description

`string`

What this test guards — write it for whoever reads a failed-test error.

- rule: {"string":{"maxLen":"1024"}}

### spec.tests[].expectedOutputUrl

`string`

The URL the request is expected to be redirected/rewritten to. Optional
when service is set. Global maps only.

- rule: {"string":{"maxLen":"1024"}}

### spec.tests[].expectedRedirectResponseCode

`int32`

The redirect status code the request is expected to produce. Cannot be set
together with service. Global maps only.

- rule: {"int32":{"gte":0}}

### spec.tests[].headers

`[]GcpUrlMapTestHeader`

Request headers the test sends. Global maps only.

### spec.tests[].headers[].name

`string` · required

Header name.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.tests[].headers[].value

`string` · required

Header value.

- rule: {"required":true}

### spec.deletionPolicy

`string`

What `terraform destroy` (or a stack teardown) may do to the URL map.
DELETE (the default when empty) allows deletion; PREVENT fails the
destroy outright — the URL map and anything orchestrating its teardown
stop there; ABANDON removes it from state without deleting it in GCP
(the map keeps serving, unmanaged). Client-side only — never sent to
the GCP API.

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Validation Rules

- `exactly_one_default_target`: set exactly one default target: default_service, default_url_redirect, or default_route_action (with weighted_backend_services)
- `default_route_action_conflicts_redirect`: default_route_action and default_url_redirect are mutually exclusive
- `default_no_path_template_rewrite`: path_template_rewrite is honored only inside a route rule's route_action (and, on a regional map, a path matcher's default route action) — GCP rejects it in the URL map's default route action
- `path_matcher_path_template_rewrite_regional_only`: path_template_rewrite in a path matcher's default route action is honored only by a regional URL map — set region, move the rewrite into a route rule, or use path_prefix_rewrite
- `cache_policy_global_only`: cache_policy (Cloud CDN route caching) exists only on a global URL map — a regional Application Load Balancer has no Cloud CDN; clear region or remove every cache_policy
- `custom_error_response_policy_global_only`: custom error response policies exist only on a global URL map — the regional map has no error-page surface; clear region or remove every custom_error_response_policy and default_custom_error_response_policy
- `max_stream_duration_global_only`: max_stream_duration exists only on a global URL map (Traffic Director) at the map's default route action and inside path rules and route rules — a regional map carries none; clear region or remove it
- `regional_tests_are_service_only`: on a regional URL map every routing test names its expected service and carries no headers, expected_output_url, or expected_redirect_response_code — those test forms exist only on the global map

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpUrlMap, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.self_link` | `string` | Self-link URI of the URL map. This is the value a target HTTP(S) proxy references as its url_map — the composition handle that puts this routing brain behind a load-balancer frontend. Format: https://www.googleapis.com/compute/v1/projects/{project}/global/urlMaps/{name} |
| `status.outputs.url_map_name` | `string` | Name of the URL map as it exists in GCP. |
| `status.outputs.map_id` | `string` | Server-assigned numeric ID of the URL map. |
| `status.outputs.fingerprint` | `string` | Server-computed fingerprint of the URL map. Used for optimistic concurrency control when updating the map outside of IaC. |
| `status.outputs.region` | `string` | Region of a regional URL map; empty for a global one. Downstream blocks read it to confirm scope compatibility (a regional link must point at a regional target in the same region), and the E2E verifier picks the regional or global API by it. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.defaultRouteAction.weightedBackendServices[].backendService` | GcpBackendService | `status.outputs.self_link` |
| `spec.defaultRouteAction.requestMirrorPolicy.backendService` | GcpBackendService | `status.outputs.self_link` |
| `spec.defaultCustomErrorResponsePolicy.errorService` | GcpBackendBucket | `status.outputs.self_link` |
| `spec.pathMatchers[].defaultRouteAction.weightedBackendServices[].backendService` | GcpBackendService | `status.outputs.self_link` |
| `spec.pathMatchers[].defaultRouteAction.requestMirrorPolicy.backendService` | GcpBackendService | `status.outputs.self_link` |
| `spec.pathMatchers[].defaultCustomErrorResponsePolicy.errorService` | GcpBackendBucket | `status.outputs.self_link` |
| `spec.pathMatchers[].pathRules[].routeAction.weightedBackendServices[].backendService` | GcpBackendService | `status.outputs.self_link` |
| `spec.pathMatchers[].pathRules[].routeAction.requestMirrorPolicy.backendService` | GcpBackendService | `status.outputs.self_link` |
| `spec.pathMatchers[].pathRules[].customErrorResponsePolicy.errorService` | GcpBackendBucket | `status.outputs.self_link` |
| `spec.pathMatchers[].routeRules[].service` | GcpBackendService | `status.outputs.self_link` |
| `spec.pathMatchers[].routeRules[].routeAction.weightedBackendServices[].backendService` | GcpBackendService | `status.outputs.self_link` |
| `spec.pathMatchers[].routeRules[].routeAction.requestMirrorPolicy.backendService` | GcpBackendService | `status.outputs.self_link` |
| `spec.pathMatchers[].routeRules[].customErrorResponsePolicy.errorService` | GcpBackendBucket | `status.outputs.self_link` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| GcpTargetHttpProxy | `spec.urlMap` | `status.outputs.self_link` |
| GcpTargetHttpsProxy | `spec.urlMap` | `status.outputs.self_link` |

## See Also

- [Overview](../README.md)
