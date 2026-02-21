---
title: "Presets"
description: "Ready-to-deploy configuration presets for API Gateway"
type: "preset-list"
componentSlug: "api-gateway"
componentTitle: "API Gateway"
provider: "oci"
icon: "package"
order: 200
presets:
  - slug: "01-public-http-proxy"
    rank: "01"
    title: "Public HTTP Proxy"
    excerpt: "This preset creates a public-facing OCI API Gateway with a single deployment that proxies HTTP requests to a backend service. It includes CORS configuration for browser-based clients, a health check..."
  - slug: "02-jwt-authenticated-api"
    rank: "02"
    title: "JWT Authenticated API"
    excerpt: "This preset creates a public OCI API Gateway with JWT token validation, per-route authorization, rate limiting, CORS, and logging. The gateway validates Bearer tokens against a remote JWKS endpoint..."
  - slug: "03-private-functions-backend"
    rank: "03"
    title: "Private Functions Backend"
    excerpt: "This preset creates a private (VCN-internal) OCI API Gateway that routes requests to OCI Functions backends. The gateway is accessible only from within the VCN (or via peered VCNs, VPN, and..."
---

# API Gateway Presets

Ready-to-deploy configuration presets for API Gateway. Each preset is a complete manifest you can copy, customize, and deploy.
