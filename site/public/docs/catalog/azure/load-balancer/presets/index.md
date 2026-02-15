---
title: "Presets"
description: "Ready-to-deploy configuration presets for Load Balancer"
type: "preset-list"
componentSlug: "load-balancer"
componentTitle: "Load Balancer"
provider: "azure"
icon: "package"
order: 200
presets:
  - slug: "01-public"
    rank: "01"
    title: "Public Load Balancer"
    excerpt: "This preset creates a public (internet-facing) Azure Load Balancer with Standard SKU, a single backend pool, an HTTP health probe, and a TCP load balancing rule on port 80. This is the standard..."
  - slug: "02-internal"
    rank: "02"
    title: "Internal Load Balancer"
    excerpt: "This preset creates an internal (private VNet) Azure Load Balancer with Standard SKU, using a subnet frontend instead of a public IP. Traffic is distributed across backend instances using a private..."
---

# Load Balancer Presets

Ready-to-deploy configuration presets for Load Balancer. Each preset is a complete manifest you can copy, customize, and deploy.
