---
title: "Presets"
description: "Ready-to-deploy configuration presets for Load Balancer"
type: "preset-list"
componentSlug: "load-balancer"
componentTitle: "Load Balancer"
provider: "cloudflare"
icon: "package"
order: 200
presets:
  - slug: "01-active-passive-failover"
    rank: "01"
    title: "Active-Passive Failover"
    excerpt: "Two origins with steering=off: traffic goes to the first healthy origin; if it fails, traffic fails over to the second. Proxied through Cloudflare for DDoS protection and CDN. Use for..."
  - slug: "02-geographic-routing"
    rank: "02"
    title: "Geographic Routing"
    excerpt: "Multiple origins with steering=geo: Cloudflare routes clients to the geographically nearest healthy origin. Use for multi-region deployments where latency matters (e.g., US, EU, APAC)."
  - slug: "03-weighted-ab-testing"
    rank: "03"
    title: "Weighted A/B Testing"
    excerpt: "Two origins with steering=random and different weights: traffic is distributed by weight (e.g., 70% control, 30% variant). Use for A/B tests, canary deployments, or gradual rollouts."
---

# Load Balancer Presets

Ready-to-deploy configuration presets for Load Balancer. Each preset is a complete manifest you can copy, customize, and deploy.
