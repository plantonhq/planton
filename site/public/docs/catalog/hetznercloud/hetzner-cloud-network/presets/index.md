---
title: "Presets"
description: "Ready-to-deploy configuration presets for Hetzner Cloud Network"
type: "preset-list"
componentSlug: "hetzner-cloud-network"
componentTitle: "Hetzner Cloud Network"
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

# Hetzner Cloud Network Presets

Ready-to-deploy configuration presets for Hetzner Cloud Network. Each preset is a complete manifest you can copy, customize, and deploy.
