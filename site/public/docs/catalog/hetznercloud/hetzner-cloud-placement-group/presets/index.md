---
title: "Presets"
description: "Ready-to-deploy configuration presets for Hetzner Cloud Placement Group"
type: "preset-list"
componentSlug: "hetzner-cloud-placement-group"
componentTitle: "Hetzner Cloud Placement Group"
provider: "hetznercloud"
icon: "package"
order: 200
presets:
  - slug: "01-spread"
    rank: "01"
    title: "Spread Placement Group"
    excerpt: "This preset creates a placement group with the `spread` strategy, which guarantees that servers assigned to it run on different physical hosts within a Hetzner Cloud datacenter. If the hypervisor..."
---

# Hetzner Cloud Placement Group Presets

Ready-to-deploy configuration presets for Hetzner Cloud Placement Group. Each preset is a complete manifest you can copy, customize, and deploy.
