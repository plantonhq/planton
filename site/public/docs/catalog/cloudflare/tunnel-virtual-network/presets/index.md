---
title: "Presets"
description: "Ready-to-deploy configuration presets for Tunnel Virtual Network"
type: "preset-list"
componentSlug: "tunnel-virtual-network"
componentTitle: "Tunnel Virtual Network"
provider: "cloudflare"
icon: "package"
order: 200
presets:
  - slug: "01-isolated-segment"
    rank: "01"
    title: "Preset: Isolated routing segment"
    excerpt: "A standard, non-default virtual network used to isolate a set of routes (and the overlapping private CIDRs behind them) from other segments in the account."
  - slug: "02-default-network"
    rank: "02"
    title: "Preset: Account default virtual network"
    excerpt: "The single virtual network that routes and WARP clients fall back to when they do not name one explicitly."
---

# Tunnel Virtual Network Presets

Ready-to-deploy configuration presets for Tunnel Virtual Network. Each preset is a complete manifest you can copy, customize, and deploy.
