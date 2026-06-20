---
title: "Presets"
description: "Ready-to-deploy configuration presets for gRPC Route"
type: "preset-list"
componentSlug: "grpc-route"
componentTitle: "gRPC Route"
provider: "kubernetes"
icon: "package"
order: 200
presets:
  - slug: "01-grpc-service-routing"
    rank: "01"
    title: "gRPC Service Routing"
    excerpt: "The most common GRPCRoute: match a public hostname and a gRPC service (and optionally a method), then forward to a backend gRPC Service. This is the standard pattern for exposing a gRPC API behind a..."
  - slug: "02-grpc-weighted-canary"
    rank: "02"
    title: "gRPC Weighted Canary"
    excerpt: "Split gRPC traffic for a service across two backends by weight -- the standard progressive-delivery pattern. Here 90% of calls go to the stable backend and 10% to the canary; adjust the weights to..."
---

# gRPC Route Presets

Ready-to-deploy configuration presets for gRPC Route. Each preset is a complete manifest you can copy, customize, and deploy.
