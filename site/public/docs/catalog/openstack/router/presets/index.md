---
title: "Presets"
description: "Ready-to-deploy configuration presets for Router"
type: "preset-list"
componentSlug: "router"
componentTitle: "Router"
provider: "openstack"
icon: "package"
order: 200
presets:
  - slug: "01-edge-with-snat"
    rank: "01"
    title: "Edge Router with SNAT"
    excerpt: "This preset creates a router with an external gateway and Source NAT enabled. Tenant instances on connected subnets can reach the internet through this router without needing individual floating IPs...."
  - slug: "02-internal-only"
    rank: "02"
    title: "Internal-Only Router"
    excerpt: "This preset creates a router with no external gateway. It provides Layer 3 routing between connected subnets within the tenant but has no path to external networks. Use this for isolated environments..."
---

# Router Presets

Ready-to-deploy configuration presets for Router. Each preset is a complete manifest you can copy, customize, and deploy.
