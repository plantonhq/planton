# Regional HTTPS Frontend

A REGIONAL target HTTPS proxy — the TLS front door of a regional external Application Load Balancer (`EXTERNAL_MANAGED` with a region) or a regional internal ALB (`INTERNAL_MANAGED`). Setting `region` switches the kind to Google's regional proxy resource; the same manifest shape serves both scopes.

## When to Use

- A service that lives in one region and should terminate TLS there (data residency, latency to a regional backend, no need for Google's global anycast edge)
- A regional internal Application Load Balancer that serves HTTPS inside a VPC
- The port-443 half of a regional frontend pair, beside a regional `GcpTargetHttpProxy` serving the redirect

## Remix Notes

- Every link in the chain must be regional in the same region: the URL map (`GcpUrlMap` with `region`), the certificates, the SSL policy (`GcpSslPolicy` with `region`), the forwarding rule in front (`GcpGlobalForwardingRule` with `region`), and the backend services and health check behind the map.
- Certificates are regional self-managed `GcpSslCertificate` resources attached with an explicit `valueFrom.kind: GcpSslCertificate` (the list defaults to the global managed-certificate kind), or regional Certificate Manager certificates through `certificateManagerCertificates`. Google-managed compute certificates are global only.
- `certificateMap`, `quicOverride`, `tlsEarlyData`, and `proxyBind` exist only on the global proxy; the spec rejects them when `region` is set.
- A regional external ALB needs a proxy-only subnet (`GcpSubnetwork` with `purpose: REGIONAL_MANAGED_PROXY`) in the region before the forwarding rule can be created.
- Reference the URL map, certificate, and SSL policy via `valueFrom` instead of literal self-links.
