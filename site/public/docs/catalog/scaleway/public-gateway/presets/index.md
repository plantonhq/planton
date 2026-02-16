---
title: "Presets"
description: "Ready-to-deploy configuration presets for Public Gateway"
type: "preset-list"
componentSlug: "public-gateway"
componentTitle: "Public Gateway"
provider: "scaleway"
icon: "package"
order: 200
presets:
  - slug: "01-nat-gateway"
    rank: "01"
    title: "NAT Gateway"
    excerpt: "This preset creates a Scaleway Public Gateway that provides NAT masquerade for a Private Network. Resources in the attached network can reach the internet through the gateway's public IP without..."
  - slug: "02-bastion-enabled"
    rank: "02"
    title: "Bastion-Enabled Gateway"
    excerpt: "This preset creates a Scaleway Public Gateway with both NAT masquerade and an SSH bastion. In addition to providing outbound internet access for the Private Network, the gateway acts as a secure SSH..."
---

# Public Gateway Presets

Ready-to-deploy configuration presets for Public Gateway. Each preset is a complete manifest you can copy, customize, and deploy.
