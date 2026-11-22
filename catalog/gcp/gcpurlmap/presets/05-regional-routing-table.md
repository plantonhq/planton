# Regional Routing Table

A REGIONAL URL map — the routing brain of a regional external Application Load Balancer or a regional internal ALB. Setting `region` switches the kind to Google's regional URL map resource; the same manifest shape serves both scopes. The preset's route rule rewrites `/v1/...` to `/v2/...` with a `pathTemplateRewrite` beside the `pathTemplateMatch` that captures the path, and the matcher's default route action bounds every other request to 30 seconds.

## When to Use

- A service that lives in one region and should be routed from that region (data residency, latency to a regional backend, no need for Google's global anycast edge)
- A regional internal Application Load Balancer inside a VPC
- Moving a path prefix between API versions at the load balancer, so clients keep calling `/v1` while the backend serves `/v2`

## Remix Notes

- Every backend the map routes to must be a regional `GcpBackendService` in the same region — a regional map never routes to a backend bucket (global only).
- A `pathTemplateRewrite` belongs in a route rule's `routeAction`, and at least one of that rule's `matchRules` needs a `pathTemplateMatch` declaring the variables the rewrite uses. Google refuses the rewrite in a path matcher's default route action, in the map's default route action, and in a path rule, on both scopes. A matcher's default can still rewrite with `pathPrefixRewrite`.
- The regional map has no route-scoped `cachePolicy` (regional ALBs have no Cloud CDN), no custom error response policies, no `maxStreamDuration` outside a path matcher's default action, and its routing `tests` name a `service` only; the spec rejects each when `region` is set.
- Reference this map from regional target proxies (`GcpTargetHttpProxy` / `GcpTargetHttpsProxy` with the same `region`).
- Reference backend services via `valueFrom` instead of literal self-links.
