---
title: "Presets"
description: "Ready-to-deploy configuration presets for Request Authentication"
type: "preset-list"
componentSlug: "request-authentication"
componentTitle: "Request Authentication"
provider: "kubernetes"
icon: "package"
order: 200
presets:
  - slug: "01-namespace-jwt-validation"
    rank: "01"
    title: "Validate JWTs Across a Namespace"
    excerpt: "The canonical RequestAuthentication: accept JWTs from a single trusted issuer for every workload in a namespace. A request that carries a valid token gets an authenticated principal; a request with..."
  - slug: "02-workload-jwt-with-claim-headers"
    rank: "02"
    title: "Workload JWT Validation with a Custom Header and Claim Forwarding"
    excerpt: "Validate JWTs for a single selected workload, extract the token from a custom header, restrict it to a specific audience, and forward a verified claim to the backend as an HTTP header the application..."
---

# Request Authentication Presets

Ready-to-deploy configuration presets for Request Authentication. Each preset is a complete manifest you can copy, customize, and deploy.
