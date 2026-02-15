---
title: "Presets"
description: "Ready-to-deploy configuration presets for Route53 Zone"
type: "preset-list"
componentSlug: "route53-zone"
componentTitle: "Route53 Zone"
provider: "aws"
icon: "package"
order: 200
presets:
  - slug: "01-public-zone"
    rank: "01"
    title: "Public DNS Zone"
    excerpt: "This preset creates a public Route53 hosted zone for managing DNS records that resolve globally on the internet. Public zones are the most common type and are used for any domain that needs to be..."
  - slug: "02-private-vpc-zone"
    rank: "02"
    title: "Private VPC DNS Zone"
    excerpt: "This preset creates a private Route53 hosted zone that resolves DNS queries only within associated VPCs. Private zones enable split-horizon DNS, where internal services use private domain names..."
---

# Route53 Zone Presets

Ready-to-deploy configuration presets for Route53 Zone. Each preset is a complete manifest you can copy, customize, and deploy.
