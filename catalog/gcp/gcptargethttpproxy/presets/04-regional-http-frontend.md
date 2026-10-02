# Regional HTTP Frontend

A REGIONAL target HTTP proxy — the frontend adapter of a regional external Application Load Balancer (`EXTERNAL_MANAGED` with a region) or a regional internal ALB (`INTERNAL_MANAGED`). Setting `region` switches the kind to Google's regional proxy resource; the same manifest shape serves both scopes.

## When to Use

- A service that lives in one region and should be fronted from that region (data residency, latency to a regional backend, no need for Google's global anycast edge)
- A regional internal Application Load Balancer for services inside a VPC
- A regional HTTPS frontend's port-80 half, pointing at a redirect-only regional URL map

## Remix Notes

- Every link in the chain must be regional in the same region: the URL map (`GcpUrlMap` with `region`), the forwarding rule in front (`GcpGlobalForwardingRule` with `region`), and the backend services and health check behind the map.
- A regional external ALB needs a proxy-only subnet (`GcpSubnetwork` with `purpose: REGIONAL_MANAGED_PROXY`) in the region before the forwarding rule can be created.
- `proxyBind` does not exist on the regional proxy (Traffic Director has no regional form); the spec rejects it when `region` is set.
- `httpKeepAliveTimeoutSec` is immutable on the regional proxy — changing it recreates the proxy.
- Reference the `GcpUrlMap` via `valueFrom` instead of a literal self-link.
