---
title: "Presets"
description: "Ready-to-deploy configuration presets for Private Network"
type: "preset-list"
componentSlug: "private-network"
componentTitle: "Private Network"
provider: "scaleway"
icon: "package"
order: 200
presets:
  - slug: "01-auto-subnet"
    rank: "01"
    title: "Auto-Subnet Private Network"
    excerpt: "This preset creates a Scaleway Private Network with IPAM-managed automatic subnet allocation. Scaleway assigns an IPv4 CIDR from its default range, so you do not need to plan address space upfront...."
  - slug: "02-explicit-subnet"
    rank: "02"
    title: "Explicit-Subnet Private Network"
    excerpt: "This preset creates a Scaleway Private Network with a user-defined IPv4 CIDR block. Specifying the subnet gives you full control over address space, which is essential when multiple Private Networks..."
---

# Private Network Presets

Ready-to-deploy configuration presets for Private Network. Each preset is a complete manifest you can copy, customize, and deploy.
