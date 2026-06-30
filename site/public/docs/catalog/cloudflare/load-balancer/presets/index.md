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
    excerpt: "A monitor, a single pool with a primary and secondary origin, and a load balancer with `steeringPolicy: off`. Traffic goes to the first healthy origin; if it fails health checks, it fails over to the..."
  - slug: "02-geographic-routing"
    rank: "02"
    title: "Geographic Routing"
    excerpt: "A monitor, two regional pools (US and EU), and a load balancer with `steeringPolicy: geo` and `regionPools` mapping regions to pools. Users are routed to the nearest healthy region, falling back to..."
  - slug: "03-weighted-ab-testing"
    rank: "03"
    title: "Weighted A/B Testing"
    excerpt: "A monitor, two pools (control and variant), and a load balancer with `steeringPolicy: random` plus `randomSteering` weights. Cloudflare selects a pool at random in proportion to the configured..."
---

# Load Balancer Presets

Ready-to-deploy configuration presets for Load Balancer. Each preset is a complete manifest you can copy, customize, and deploy.
