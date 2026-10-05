# Azure Front Door Route

Deploys a Front Door route -- the rule that connects client traffic arriving at an endpoint to an origin group, by URL pattern, with protocol posture, HTTPS redirect, custom-domain attachment, rule-set attachment, and optional edge caching. A route is the traffic-serving edge of the Front Door graph (endpoint, then route, then origin group, then origins), and routes are many-per-endpoint with independent lifecycles: one endpoint commonly serves `/api/*` and `/static/*` from different backends, which is why the route is a first-class kind referencing the endpoint rather than a list folded into it.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Front Door Route** -- a named child of an endpoint matching URL patterns and forwarding to an origin group
- **Protocol posture** -- supported client protocols (HTTP/HTTPS), the origin-leg forwarding protocol, and the edge HTTPS redirect
- **Domain binding** -- the endpoint's generated `*.azurefd.net` hostname and/or attached custom domains (the route side owns the domain attachment)
- **Rule-set attachment** -- references to rule sets whose delivery policies apply on this route
- **Edge cache** (created only when `cache` is set) -- query-string cache-key behavior and edge compression; Azure treats absent cache settings as caching OFF, so the block is a real switch

`originIds` is deploy sequencing only: Azure never receives the list -- the provider uses it to create the route after its backends exist, because ARM rejects a route whose origin group has no origins yet. ARM does not support tags on routes.

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription.
- **Planton Runner** -- required when using Runner-based credential delivery.

### Azure Subscription

- **An endpoint and an origin group** under the same Front Door profile (`endpointId`, `originGroupId`), with at least one origin in the group before the route applies -- or list the origins in `originIds` for sequencing.
- **DNS-validated custom domains** (only for `customDomainIds`) -- every referenced domain must belong to the route's profile and be validated before traffic can flow.

## Deploy

### Console

Open the deployment store, find **Azure Front Door Route**, and click **Deploy**. The wizard walks you through attachment (endpoint, origin group, sequencing origins), matching and protocols, domains, and optional edge caching. Start from the **Catch-All HTTPS Route** preset in the [Presets](#presets) tab for the default route shape.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureFrontDoorRoute
metadata:
  name: my-default-route
  org: acme-corp
  env: prod
spec:
  endpointId:
    valueFrom:
      kind: AzureFrontDoorEndpoint
      name: my-web-endpoint
      fieldPath: status.outputs.endpoint_id
  routeName: default-route
  originGroupId:
    valueFrom:
      kind: AzureFrontDoorOriginGroup
      name: my-app-backends
      fieldPath: status.outputs.origin_group_id
  originIds:
    - valueFrom:
        kind: AzureFrontDoorOrigin
        name: my-app-origin
        fieldPath: status.outputs.origin_id
  patternsToMatch:
    - /*
  supportedProtocols:
    - HTTP
    - HTTPS
```

```shell
planton apply -f front-door-route.yaml
```

This creates a catch-all route accepting both protocols with the default HTTPS redirect on -- HTTP arrives only to be 301'd to HTTPS -- and caching off until a `cache` block is declared. An Infra Job tracks the provisioning in real time.

### InfraChart

A path-scoped cached route beside the catch-all, wired to resources in the same InfraPipeline:

```yaml
spec:
  endpointId:
    valueFrom:
      kind: AzureFrontDoorEndpoint
      name: my-web-endpoint
      fieldPath: status.outputs.endpoint_id
  routeName: static-assets
  originGroupId:
    valueFrom:
      kind: AzureFrontDoorOriginGroup
      name: my-app-backends
      fieldPath: status.outputs.origin_group_id
  patternsToMatch:
    - /assets/*
  supportedProtocols:
    - HTTPS
  httpsRedirectEnabled: false
  cache:
    queryStringCachingBehavior: IGNORE_QUERY_STRING
    compressionEnabled: true
```

The InfraPipeline resolves the dependency graph, deploys the endpoint, group, and origins first, then provisions the route with the resolved ARM IDs.

## Key Configuration

These are the most important decisions when configuring a route. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Patterns** -- at least one, each starting with `/`; `/*` is the catch-all. Route matching picks the MOST SPECIFIC pattern across the endpoint's routes, so `/api/*` on one route and `/*` on another cleanly split API and catch-all traffic -- ordering is automatic, not configured.

**HTTPS redirect** -- `httpsRedirectEnabled` defaults to true and requires BOTH protocols in `supportedProtocols` (the redirect needs HTTP to arrive and HTTPS to land on); the spec rejects redirect-with-one-protocol before Azure would. For APIs, prefer HTTPS-only with the redirect disabled: clients mishandle redirected POST bodies, and failing loudly surfaces misconfigured base URLs.

**Forwarding protocol** -- the origin-leg protocol, independent of what the client used. Unspecified deploys MATCH_REQUEST (mirror the client); HTTPS_ONLY keeps the origin leg encrypted even for HTTP clients; HTTP_ONLY is for origins without TLS -- pair it with Private Link rather than sending plaintext over the internet.

**Domain binding** -- `linkToDefaultDomain` can only be false when at least one custom domain is attached, otherwise the route would answer on no hostname (spec-enforced). Disable it for production routes that should only answer on their branded domains.

**Origin path** -- `originPath` is PREPENDED on the origin side: a request for `/api/users` on a route with `originPath: /v1` fetches `/v1/api/users` from the backend. It namespaces one origin under a directory; it does not strip the public prefix.

**Cache** -- omitting the block means every request hits the origin. When present, IGNORE_QUERY_STRING (the default) shares one cache entry across query variants -- right for path-versioned assets like `app.3f9c.js`; USE_QUERY_STRING keys every variant separately for query-driven content; the *_SPECIFIED behaviors include or exclude only the names in `queryStrings`. Compression applies only to responses between 1 KiB and 8 MiB whose MIME type is in Azure's supported list (text, JSON, XML, JavaScript, SVG, fonts) -- binary media is already compressed and not eligible.

**originIds** -- deploy sequencing only. List the group's origins when deploying the whole chain in one manifest set; omit when the origins already exist. Azure never sees this field.

**ForceNew fields** -- `endpointId` and `routeName` fix the route's ARM identity at creation. `originGroupId` is deliberately updatable in place: repointing a route is how traffic moves between backend pools.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureFrontDoorEndpoint** | `endpointId` | `status.outputs.endpoint_id` |
| **AzureFrontDoorOriginGroup** | `originGroupId` | `status.outputs.origin_group_id` |
| **AzureFrontDoorOrigin** | `originIds[]` | `status.outputs.origin_id` (sequencing only) |
| **AzureFrontDoorCustomDomain** | `customDomainIds[]` | `status.outputs.custom_domain_id` |
| **AzureFrontDoorRuleSet** | `ruleSetIds[]` | `status.outputs.rule_set_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `route_id` | ARM resource ID of the route | Operator tooling, diagnostics |
| `route_name` | The route's name within its endpoint | Operator tooling, portal cross-reference |

There is deliberately no hostname output: the client-facing hostname lives on the ENDPOINT (AzureFrontDoorEndpoint's `host_name` output) -- the route is policy attached to that hostname.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Catch-all entry route** -- every path, both protocols, HTTP redirected at the edge, no caching: the standard production entry rule for a dynamic application, and the fallback beside more specific routes since Front Door picks the most specific pattern. Start from the **Catch-All HTTPS Route** preset.

**Cached static assets** -- a route dedicated to `/static/*` and `/assets/*` with edge caching and text-type compression, so the origin sees each asset roughly once per edge instead of once per user. The uncached catch-all keeps handling everything else. Start from the **Cached Static Assets Route** preset.

**Strict HTTPS-only API** -- HTTPS the only accepted protocol (plain HTTP fails instead of silently redirecting), the origin leg pinned to HTTPS_ONLY, and the public prefix rewritten with `originPath`. Start from the **HTTPS-Only API Route** preset.

**Blue/green by repointing** -- because `originGroupId` updates in place, cutting traffic to a new backend pool is a one-field route update; rollback is the same field changed back.

## Works With

- [**Azure Front Door Endpoint**](/infra-catalog/azure-front-door-endpoint) -- the public entry hostname the route attaches to (ForceNew)
- [**Azure Front Door Origin Group**](/infra-catalog/azure-front-door-origin-group) -- the backend pool matched traffic forwards to
- [**Azure Front Door Origin**](/infra-catalog/azure-front-door-origin) -- listed in `originIds` so the route waits for backends to exist
- [**Azure Front Door Custom Domain**](/infra-catalog/azure-front-door-custom-domain) -- branded hostnames this route serves
- [**Azure Front Door Rule Set**](/infra-catalog/azure-front-door-rule-set) -- request/response transforms applied to traffic on this route
