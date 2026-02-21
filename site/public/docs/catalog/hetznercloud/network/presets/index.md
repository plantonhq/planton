---
title: "Presets"
description: "Ready-to-deploy configuration presets for Network"
type: "preset-list"
componentSlug: "network"
componentTitle: "Network"
provider: "hetznercloud"
icon: "package"
order: 200
presets:
  - slug: "01-single-zone"
    rank: "01"
    title: "Single-Zone Cloud Network"
    excerpt: "This preset creates a Hetzner Cloud private network with a single cloud subnet in the eu-central zone. It is the simplest usable network configuration -- one subnet providing private IPv4..."
  - slug: "02-multi-zone"
    rank: "02"
    title: "Multi-Zone Cloud Network"
    excerpt: "This preset creates a Hetzner Cloud private network spanning two network zones, enabling private IPv4 connectivity between resources in geographically separate locations. Servers in eu-central can..."
---

# Network Presets

Ready-to-deploy configuration presets for Network. Each preset is a complete manifest you can copy, customize, and deploy.
