# Regional External ALB VIP

The front door of a REGIONAL external Application Load Balancer: a regional forwarding rule on `EXTERNAL_MANAGED` pointing at a regional target HTTP proxy, with the `STANDARD` network tier only regional rules offer. Setting `region` switches the kind to Google's regional forwarding rule resource.

## When to Use

- A service that lives in one region and should be fronted from that region (data residency, lower cost on the STANDARD tier, no need for Google's global anycast edge)
- The port-80 half of a regional frontend pair; clone it on port 443 pointing at a regional `GcpTargetHttpsProxy`

## Remix Notes

- Every link in the chain must be regional in the same region: the proxy, its URL map, the backend services, the health check, and the reserved address (`GcpAddress` attached with an explicit `valueFrom.kind: GcpAddress`, since `ipAddress` defaults to the global address kind).
- A regional external ALB needs a proxy-only subnet in the region before this rule can be created: a `GcpSubnetwork` with `purpose: REGIONAL_MANAGED_PROXY` and `role: ACTIVE` in the network named here — the regional external ALB is the one external scheme that takes `network`.
- `networkTier: STANDARD` routes egress through regional ISP transit instead of Google's backbone; drop it (or set `PREMIUM`) for the global-quality path. A reserved address must carry the same tier.
- Reference the proxy, network, and address via `valueFrom` instead of literal self-links.
