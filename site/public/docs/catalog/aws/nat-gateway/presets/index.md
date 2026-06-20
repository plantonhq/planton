---
title: "Presets"
description: "Ready-to-deploy configuration presets for NAT Gateway"
type: "preset-list"
componentSlug: "nat-gateway"
componentTitle: "NAT Gateway"
provider: "aws"
icon: "package"
order: 200
presets:
  - slug: "01-public-nat-gateway"
    rank: "01"
    title: "Public NAT Gateway (greenfield)"
    excerpt: "A public NAT gateway placed in a Planton-managed public `AwsSubnet` and fronted by a Planton-managed `AwsElasticIp`, both by reference. This is the canonical egress path: create a public subnet..."
  - slug: "02-private-nat-gateway"
    rank: "02"
    title: "Private NAT Gateway"
    excerpt: "A private NAT gateway (no Elastic IP) placed in an `AwsSubnet` by reference. A private gateway provides outbound access to other private networks — peered VPCs, a transit gateway, or an on-premises..."
---

# NAT Gateway Presets

Ready-to-deploy configuration presets for NAT Gateway. Each preset is a complete manifest you can copy, customize, and deploy.
