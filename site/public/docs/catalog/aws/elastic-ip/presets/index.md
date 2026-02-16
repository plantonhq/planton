---
title: "Presets"
description: "Ready-to-deploy configuration presets for Elastic IP"
type: "preset-list"
componentSlug: "elastic-ip"
componentTitle: "Elastic IP"
provider: "aws"
icon: "package"
order: 200
presets:
  - slug: "01-standard-eip"
    rank: "01"
    title: "Preset: Standard Elastic IP"
    excerpt: "**Use case:** Allocate a static public IPv4 address from Amazon's default pool."
  - slug: "02-byoip-pool"
    rank: "02"
    title: "Preset: BYOIP Pool Elastic IP"
    excerpt: "**Use case:** Allocate a static public IPv4 address from your own registered IP address range."
---

# Elastic IP Presets

Ready-to-deploy configuration presets for Elastic IP. Each preset is a complete manifest you can copy, customize, and deploy.
