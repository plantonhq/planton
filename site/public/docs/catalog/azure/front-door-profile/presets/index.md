---
title: "Presets"
description: "Ready-to-deploy configuration presets for Front Door Profile"
type: "preset-list"
componentSlug: "front-door-profile"
componentTitle: "Front Door Profile"
provider: "azure"
icon: "package"
order: 200
presets:
  - slug: "01-standard-web-acceleration"
    rank: "01"
    title: "Standard Web Acceleration"
    excerpt: "A Standard-tier Front Door profile optimized for accelerating a web application with edge caching and compression."
  - slug: "02-multi-region-api-gateway"
    rank: "02"
    title: "Multi-Region API Gateway"
    excerpt: "A Standard-tier Front Door profile configured as an API gateway with multi-origin health-based failover and path-based routing to separate API and static asset backends."
  - slug: "03-premium-enterprise-cdn"
    rank: "03"
    title: "Premium Enterprise CDN"
    excerpt: "A Premium-tier Front Door profile with Private Link connectivity to an Azure App Service backend. The origin is reached exclusively through Azure's backbone network without any public internet..."
---

# Front Door Profile Presets

Ready-to-deploy configuration presets for Front Door Profile. Each preset is a complete manifest you can copy, customize, and deploy.
