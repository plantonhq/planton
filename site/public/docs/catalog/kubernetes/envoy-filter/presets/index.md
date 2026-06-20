---
title: "Presets"
description: "Ready-to-deploy configuration presets for Envoy Filter"
type: "preset-list"
componentSlug: "envoy-filter"
componentTitle: "Envoy Filter"
provider: "kubernetes"
icon: "package"
order: 200
presets:
  - slug: "01-grpc-web-cors-gateway"
    rank: "01"
    title: "Add a CORS HTTP filter to a gateway (gRPC-Web)"
    excerpt: "The canonical \"escape hatch\" use: insert Envoy's native CORS HTTP filter into a gateway's HTTP connection manager so a browser can call a gRPC-Web (or any cross-origin) backend through the gateway...."
  - slug: "02-outbound-cluster-merge"
    rank: "02"
    title: "Tune an outbound cluster (MERGE)"
    excerpt: "Merge low-level Envoy cluster settings -- connection timeout, TCP keepalive, circuit-breaker internals, and other knobs not exposed by `DestinationRule` -- onto the CDS cluster a sidecar generates..."
---

# Envoy Filter Presets

Ready-to-deploy configuration presets for Envoy Filter. Each preset is a complete manifest you can copy, customize, and deploy.
