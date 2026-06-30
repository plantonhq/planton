---
title: "Presets"
description: "Ready-to-deploy configuration presets for Load Balancer Pool"
type: "preset-list"
componentSlug: "load-balancer-pool"
componentTitle: "Load Balancer Pool"
provider: "cloudflare"
icon: "package"
order: 200
presets:
  - slug: "01-web-pool"
    rank: "01"
    title: "Preset: Web pool with two origins"
    excerpt: "A pool of two web origins health-checked by an HTTPS monitor, ready to attach to a load balancer's `default_pools`."
  - slug: "02-geo-located-pool"
    rank: "02"
    title: "Preset: Geo-located pool (proximity steering)"
    excerpt: "A regional pool tagged with latitude/longitude and a least-connections origin policy, for use with a load balancer's `proximity` or `geo` steering."
---

# Load Balancer Pool Presets

Ready-to-deploy configuration presets for Load Balancer Pool. Each preset is a complete manifest you can copy, customize, and deploy.
