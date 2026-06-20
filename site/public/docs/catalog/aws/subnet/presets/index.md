---
title: "Presets"
description: "Ready-to-deploy configuration presets for Subnet"
type: "preset-list"
componentSlug: "subnet"
componentTitle: "Subnet"
provider: "aws"
icon: "package"
order: 200
presets:
  - slug: "01-private"
    rank: "01"
    title: "Private Subnet"
    excerpt: "A private subnet whose default route points at a NAT gateway: instances can reach the internet outbound (patching, image pulls, external APIs) but cannot be reached from it. This is the standard..."
  - slug: "02-public"
    rank: "02"
    title: "Public Subnet"
    excerpt: "A public subnet whose default route points at an internet gateway, with launch-time public IP assignment enabled. This is where internet-facing resources live: application load balancers, bastion..."
  - slug: "03-isolated"
    rank: "03"
    title: "Isolated Subnet"
    excerpt: "A subnet with no route rules of its own: it stays on the VPC main route table and has no path to the internet. This is the right placement for data stores and other resources that should only ever be..."
---

# Subnet Presets

Ready-to-deploy configuration presets for Subnet. Each preset is a complete manifest you can copy, customize, and deploy.
