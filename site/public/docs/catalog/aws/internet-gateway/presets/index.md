---
title: "Presets"
description: "Ready-to-deploy configuration presets for Internet Gateway"
type: "preset-list"
componentSlug: "internet-gateway"
componentTitle: "Internet Gateway"
provider: "aws"
icon: "package"
order: 200
presets:
  - slug: "01-public-internet-gateway"
    rank: "01"
    title: "Public Internet Gateway (greenfield)"
    excerpt: "An internet gateway attached to a Planton-managed `AwsVpc` by reference. This is the canonical composition path: create an `AwsVpc`, create this gateway pointing at it, then give your public subnets..."
  - slug: "02-attach-to-existing-vpc"
    rank: "02"
    title: "Attach to an Existing VPC (brownfield)"
    excerpt: "An internet gateway attached to a VPC that already exists outside Planton, by literal vpc-id. Use this when the VPC is managed elsewhere (an existing landing zone, another tool, or a hand-created..."
---

# Internet Gateway Presets

Ready-to-deploy configuration presets for Internet Gateway. Each preset is a complete manifest you can copy, customize, and deploy.
