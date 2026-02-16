---
title: "Presets"
description: "Ready-to-deploy configuration presets for Public IP"
type: "preset-list"
componentSlug: "public-ip"
componentTitle: "Public IP"
provider: "azure"
icon: "package"
order: 200
presets:
  - slug: "01-standard-static"
    rank: "01"
    title: "Standard Static Public IP"
    excerpt: "This preset creates a zone-redundant Azure Public IP with Standard SKU and static allocation. Standard SKU with static allocation is the only supported configuration (Azure retired Basic SKU in..."
---

# Public IP Presets

Ready-to-deploy configuration presets for Public IP. Each preset is a complete manifest you can copy, customize, and deploy.
