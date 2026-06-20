---
title: "Presets"
description: "Ready-to-deploy configuration presets for Reference Grant"
type: "preset-list"
componentSlug: "reference-grant"
componentTitle: "Reference Grant"
provider: "kubernetes"
icon: "package"
order: 200
presets:
  - slug: "01-allow-gateway-secret-ref"
    rank: "01"
    title: "Allow a Gateway to Reference TLS Secrets in Another Namespace"
    excerpt: "The most common ReferenceGrant: a Gateway terminates TLS using a certificate Secret that lives in a different namespace (typically the cert-manager namespace). By default that cross-namespace..."
  - slug: "02-allow-route-backend-ref"
    rank: "02"
    title: "Allow Routes to Reference Backend Services in Another Namespace"
    excerpt: "Authorize HTTP and gRPC routes in an application/frontend namespace to forward traffic to backend Services that live in a different namespace. By default a route's cross-namespace `backendRefs`..."
---

# Reference Grant Presets

Ready-to-deploy configuration presets for Reference Grant. Each preset is a complete manifest you can copy, customize, and deploy.
