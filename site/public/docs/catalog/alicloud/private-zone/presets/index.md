---
title: "Presets"
description: "Ready-to-deploy configuration presets for Private Zone"
type: "preset-list"
componentSlug: "private-zone"
componentTitle: "Private Zone"
provider: "alicloud"
icon: "package"
order: 200
presets:
  - slug: "01-internal-service-discovery"
    rank: "01"
    title: "Internal Service Discovery"
    excerpt: "This preset creates a private DNS zone for internal service discovery within a single VPC. Services register A records pointing to their private IP addresses, enabling hostname-based discovery..."
  - slug: "02-multi-vpc-database-zone"
    rank: "02"
    title: "Multi-VPC Database Zone"
    excerpt: "This preset creates a private DNS zone for database endpoint discovery, shared across multiple VPCs including cross-region. Applications in any attached VPC can resolve database hostnames without..."
---

# Private Zone Presets

Ready-to-deploy configuration presets for Private Zone. Each preset is a complete manifest you can copy, customize, and deploy.
