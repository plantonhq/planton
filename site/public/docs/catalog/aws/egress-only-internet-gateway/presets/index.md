---
title: "Presets"
description: "Ready-to-deploy configuration presets for Egress-Only Internet Gateway"
type: "preset-list"
componentSlug: "egress-only-internet-gateway"
componentTitle: "Egress-Only Internet Gateway"
provider: "aws"
icon: "package"
order: 200
presets:
  - slug: "01-ipv6-egress"
    rank: "01"
    title: "IPv6 Egress (greenfield)"
    excerpt: "An egress-only internet gateway attached to a Planton-managed dual-stack `AwsVpc` by reference. This is the canonical composition path: create an `AwsVpc` with an IPv6 CIDR, create this gateway..."
  - slug: "02-attach-to-existing-vpc"
    rank: "02"
    title: "Attach to an Existing VPC (brownfield)"
    excerpt: "An egress-only internet gateway attached to a dual-stack VPC that already exists outside Planton, by literal vpc-id. Use this when the VPC is managed elsewhere (an existing landing zone, another..."
---

# Egress-Only Internet Gateway Presets

Ready-to-deploy configuration presets for Egress-Only Internet Gateway. Each preset is a complete manifest you can copy, customize, and deploy.
