---
title: "Presets"
description: "Ready-to-deploy configuration presets for Authorization Policy"
type: "preset-list"
componentSlug: "authorization-policy"
componentTitle: "Authorization Policy"
provider: "kubernetes"
icon: "package"
order: 200
presets:
  - slug: "01-require-jwt-principal"
    rank: "01"
    title: "Require an Authenticated JWT Principal"
    excerpt: "The canonical AuthorizationPolicy: allow a request only if it carries a valid JWT identity (`request_principals: [\"*\"]` matches any issuer/subject). Anonymous requests -- those with no token, or a..."
  - slug: "02-custom-ext-authz-ingress"
    rank: "02"
    title: "Delegate to an External Authorizer (CUSTOM) on the Ingress Gateway"
    excerpt: "Delegate the authorization decision for sensitive ingress paths to an external authorization service (an OPA sidecar, a custom authz server, an OAuth2 proxy, ...) registered as an extension provider..."
---

# Authorization Policy Presets

Ready-to-deploy configuration presets for Authorization Policy. Each preset is a complete manifest you can copy, customize, and deploy.
