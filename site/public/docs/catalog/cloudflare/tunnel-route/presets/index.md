---
title: "Presets"
description: "Ready-to-deploy configuration presets for Tunnel Route"
type: "preset-list"
componentSlug: "tunnel-route"
componentTitle: "Tunnel Route"
provider: "cloudflare"
icon: "package"
order: 200
presets:
  - slug: "01-private-subnet"
    rank: "01"
    title: "Preset: Private subnet via a tunnel"
    excerpt: "Advertise a private subnet through a tunnel so WARP clients can reach hosts in it — the most common private-networking route."
  - slug: "02-isolated-overlap"
    rank: "02"
    title: "Preset: Overlapping CIDR isolated in a virtual network"
    excerpt: "Advertise a CIDR that overlaps another already-connected network by scoping the route to its own virtual network, so the two never collide."
---

# Tunnel Route Presets

Ready-to-deploy configuration presets for Tunnel Route. Each preset is a complete manifest you can copy, customize, and deploy.
