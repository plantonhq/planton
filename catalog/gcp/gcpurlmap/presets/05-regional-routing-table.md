# Regional Routing Table

A REGIONAL URL map — the routing brain of a regional external Application Load Balancer or a regional internal ALB. Setting `region` switches the kind to Google's regional URL map resource; the same manifest shape serves both scopes, and this preset uses the one knob only the regional map honors: a `pathTemplateRewrite` in a path matcher's default route action.

## When to Use

- A service that lives in one region and should be routed from that region (data residency, latency to a regional backend, no need for Google's global anycast edge)
- A regional internal Application Load Balancer inside a VPC
- Path-template rewriting of every request a matcher handles, without spelling out route rules

## Remix Notes

- Every backend the map routes to must be a regional `GcpBackendService` in the same region — a regional map never routes to a backend bucket (global only).
- A `pathTemplateRewrite` in the matcher's default route action needs a route rule in the same matcher whose `pathTemplateMatch` declares the variables the rewrite uses — Google rejects the rewrite without one.
- The regional map has no route-scoped `cachePolicy` (regional ALBs have no Cloud CDN), no custom error response policies, no `maxStreamDuration` outside a path matcher's default action, and its routing `tests` name a `service` only; the spec rejects each when `region` is set.
- Reference this map from regional target proxies (`GcpTargetHttpProxy` / `GcpTargetHttpsProxy` with the same `region`).
- Reference backend services via `valueFrom` instead of literal self-links.
