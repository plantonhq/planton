---
title: "Presets"
description: "Ready-to-deploy configuration presets for NAT Gateway"
type: "preset-list"
componentSlug: "nat-gateway"
componentTitle: "NAT Gateway"
provider: "alicloud"
icon: "package"
order: 200
presets:
  - slug: "01-single-vswitch"
    rank: "01"
    title: "Single VSwitch NAT Gateway"
    excerpt: "This preset creates a NAT Gateway that provides outbound internet access for a single VSwitch. This is the most common pattern for development or simple production environments."
  - slug: "02-multi-az-production"
    rank: "02"
    title: "Multi-AZ Production NAT Gateway"
    excerpt: "This preset creates a production-grade NAT Gateway serving multiple VSwitches across availability zones. Deletion protection is enabled to prevent accidental removal. Resource tags support cost..."
  - slug: "03-cidr-based-snat"
    rank: "03"
    title: "CIDR-Based SNAT NAT Gateway"
    excerpt: "This preset creates a NAT Gateway with SNAT entries specified by CIDR blocks instead of VSwitch IDs. This provides fine-grained control over which IP ranges get outbound internet access."
---

# NAT Gateway Presets

Ready-to-deploy configuration presets for NAT Gateway. Each preset is a complete manifest you can copy, customize, and deploy.
