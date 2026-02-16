---
title: "Presets"
description: "Ready-to-deploy configuration presets for VPC"
type: "preset-list"
componentSlug: "vpc"
componentTitle: "VPC"
provider: "gcp"
icon: "package"
order: 200
presets:
  - slug: "01-custom-mode-regional"
    rank: "01"
    title: "Custom Mode VPC with Regional Routing"
    excerpt: "This preset creates a VPC in custom subnet mode with regional routing and Private Services Access enabled. Custom mode gives full control over subnet CIDR ranges and regions. Private Services Access..."
  - slug: "02-custom-mode-global"
    rank: "02"
    title: "Custom Mode VPC with Global Routing"
    excerpt: "This preset creates a VPC in custom subnet mode with global dynamic routing. Global routing enables Cloud Routers to advertise routes across all regions in the VPC, which is required for multi-region..."
---

# VPC Presets

Ready-to-deploy configuration presets for VPC. Each preset is a complete manifest you can copy, customize, and deploy.
