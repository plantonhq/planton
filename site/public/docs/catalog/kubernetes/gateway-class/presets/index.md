---
title: "Presets"
description: "Ready-to-deploy configuration presets for Gateway Class"
type: "preset-list"
componentSlug: "gateway-class"
componentTitle: "Gateway Class"
provider: "kubernetes"
icon: "package"
order: 200
presets:
  - slug: "01-istio-gateway-class"
    rank: "01"
    title: "Istio GatewayClass"
    excerpt: "This preset creates a GatewayClass for the Istio gateway controller. Istio is one of the most widely deployed Gateway API implementations and a common choice for production ingress and service mesh..."
  - slug: "02-envoy-gateway-class"
    rank: "02"
    title: "Envoy Gateway GatewayClass"
    excerpt: "This preset creates a GatewayClass for the Envoy Gateway controller. Envoy Gateway is a standalone Gateway API implementation built on Envoy proxy, popular for teams that want Envoy's data plane..."
---

# Gateway Class Presets

Ready-to-deploy configuration presets for Gateway Class. Each preset is a complete manifest you can copy, customize, and deploy.
