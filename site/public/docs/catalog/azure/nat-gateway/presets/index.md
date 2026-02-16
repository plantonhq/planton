---
title: "Presets"
description: "Ready-to-deploy configuration presets for NAT Gateway"
type: "preset-list"
componentSlug: "nat-gateway"
componentTitle: "NAT Gateway"
provider: "azure"
icon: "package"
order: 200
presets:
  - slug: "01-standard"
    rank: "01"
    title: "Standard NAT Gateway"
    excerpt: "This preset creates an Azure NAT Gateway attached to a subnet, providing reliable SNAT for outbound internet connectivity. The NAT Gateway automatically provisions a public IP and associates it with..."
---

# NAT Gateway Presets

Ready-to-deploy configuration presets for NAT Gateway. Each preset is a complete manifest you can copy, customize, and deploy.
