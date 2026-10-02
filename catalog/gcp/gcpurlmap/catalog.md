# GCP URL Map

Deploys a Compute Engine URL map — the L7 routing brain of an Application Load Balancer, global (the default) or regional when `region` is set. The URL map matches each request's Host and path and decides what happens: send it to a backend service or backend bucket, split it across weighted backends (canary/blue-green), redirect, or rewrite. Target proxies bind this map's self link; the forwarding rule and its IP sit in front of the proxy.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **Compute Engine URL Map** -- global, or regional when `region` is set; host rules, path matchers with path or route rules, default targets at every level, header policies, routing self-tests, and per-route traffic management (timeouts, retries, mirroring, CORS, fault injection); on the global map also custom error pages, stream-duration limits, and route-scoped CDN caching
- **Compute Engine API enablement** -- `compute.googleapis.com` is enabled in the target project; tearing down the URL map never disables the API

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with credentials for the target GCP project. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### GCP Project

- **A GCP project** where the URL map will be created. The module enables the Compute Engine API itself, so the connection's principal needs permission to enable services on a fresh project.
- **At least one backend** (GcpBackendService or GcpBackendBucket) — the default target is required, and it is the natural creation order: backends first, then the routing that references them.

## Deploy

### Console

Open the deployment store, find **GCP URL Map**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Host and Path Fan-Out** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpUrlMap
metadata:
  name: prod-web-map
  org: acme-corp
  env: prod
spec:
  projectId:
    value: "acme-prod-12345"
  defaultService:
    value: "https://www.googleapis.com/compute/v1/projects/acme-prod-12345/global/backendServices/web-api"
  hostRules:
    - hosts: [acme.com, www.acme.com]
      pathMatcher: site
  pathMatchers:
    - name: site
      defaultService:
        value: "https://www.googleapis.com/compute/v1/projects/acme-prod-12345/global/backendServices/web-api"
      pathRules:
        - paths: ["/assets/*"]
          service:
            value: "https://www.googleapis.com/compute/v1/projects/acme-prod-12345/global/backendBuckets/assets-backend"
```

```shell
planton apply -f url-map.yaml
```

This creates the classic fan-out: dynamic traffic to a backend service, /assets/* to a CDN-backed bucket. A Stack Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the backends:

```yaml
spec:
  defaultService:
    valueFrom:
      kind: GcpBackendService
      name: web-api
      fieldPath: status.outputs.self_link
```

The InfraPipeline resolves the dependency graph — backends first, then this URL map — and a downstream GcpTargetHttpsProxy references this map's `self_link`.

## Key Configuration

These are the most important decisions when configuring a URL map. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Scope** -- `region` empty builds the global map (global external ALB, cross-region internal ALB, Traffic Director); a region name builds the regional map (regional external and internal ALBs), which routes only to regional backend services in that region and is referenced only by regional proxies. Immutable.

**Default target** -- Exactly one of: a backend service/bucket, a URL redirect (the http→https front half), or a weighted split across backend services. All of it is MUTABLE — repointing a live map is an in-place, zero-downtime update, which is the canary lever. On the backend arm, a host/path rewrite may accompany the service.

**Path matchers and rules** -- Named routing tables host rules point at. Each matcher uses path rules (longest prefix — the simple choice) OR route rules (priority-ordered with header/query matching), never both. Every rule target is the same three-arm choice, so a weighted canary can be scoped to exactly one path.

**Host rules** -- Fan hosts out to matchers by name ("*.acme.com" covers subdomains, not the apex). Every host routed here must also be covered by a certificate on the target HTTPS proxy.

**Routing tests** -- Self-tests GCP evaluates on every create/update; a failing test BLOCKS the change. One test per critical path is the cheapest routing-regression guard on GCP.

**Advanced traffic management** -- Every route action carries the full per-route surface: timeout and retry policies, request mirroring, CORS, fault injection, stream-duration limits, and a route-scoped Cloud CDN cache policy — alongside the weighted splits and rewrites. Route-scoped policies override the backend service's own resilience and CDN settings for matching traffic only, so one backend can serve aggressive-retry API routes and long-timeout upload routes at once. A route action is a routing *target* only via weighted backends; sub-policies may accompany a plain service target.

**Destroy stance** -- `deletionPolicy` decides what a teardown may do: DELETE (default), PREVENT (the destroy fails at this resource), or ABANDON (removed from state, keeps serving unmanaged) — the guard rail for routing that other teams' DNS points at.

## Outputs and Dependencies

### What This Component Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** | `projectId` | `status.outputs.project_id` |
| **GcpBackendService** | `defaultService`, path/route rule `service`, `weightedBackendServices[].backendService`, `tests[].service` | `status.outputs.self_link` |
| **GcpBackendBucket** | `defaultService`, path rule `service`, error policy `errorService` | `status.outputs.self_link` |

### What This Component Provides

After provisioning, `status.outputs` contains values that downstream Cloud Resources can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `self_link` | Self-link URI of the URL map | GcpTargetHttpProxy / GcpTargetHttpsProxy `url_map` |
| `url_map_name` | Name as it exists in GCP | Audit, fleet inventory |
| `map_id` | Server-assigned numeric ID | Diagnostics |
| `fingerprint` | Optimistic-concurrency token | Out-of-band gcloud updates |
| `region` | Region of a regional URL map; empty for global | Scope checks on downstream blocks |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Host + path fan-out** -- One domain, paths split across API and static backends. Start from the **Host and Path Fan-Out** preset.

**Weighted canary** -- A 90/10 split walked up as confidence grows; weight 0 drains a backend. Start from the **Weighted Canary Split** preset.

**Apex redirect** -- The redirect-only map a port-80 HTTP proxy serves (http→https 301), or the apex-to-www bounce. Start from the **Apex-to-WWW HTTPS Redirect** preset.

**Traffic-managed canary** -- A weighted split hardened with bounded timeouts, deliberate retries, load-balancer CORS, and an attribution response header. Start from the **Traffic-Managed Canary** preset.

**Regional routing table** -- A regional URL map fronting a regional backend service, with a path template rewrite in the path matcher's default action (the regional map's own knob). Start from the **Regional Routing Table** preset.

## Works With

- [**GCP Project**](/cloud-catalog/gcp-project) -- provides the GCP project where the URL map is created
- [**GCP Backend Service**](/cloud-catalog/gcp-backend-service) -- the dynamic targets this map routes to
- [**GCP Backend Bucket**](/cloud-catalog/gcp-backend-bucket) -- the static targets and custom-error-page origins
- [**GCP Target HTTPS Proxy**](/cloud-catalog/gcp-target-https-proxy) -- consumes this map's `self_link` behind TLS
- [**GCP Target HTTP Proxy**](/cloud-catalog/gcp-target-http-proxy) -- consumes a redirect-only map for the http→https upgrade
