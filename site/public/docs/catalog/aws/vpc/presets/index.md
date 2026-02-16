---
title: "Presets"
description: "Ready-to-deploy configuration presets for VPC"
type: "preset-list"
componentSlug: "vpc"
componentTitle: "VPC"
provider: "aws"
icon: "package"
order: 200
presets:
  - slug: "01-production-multi-az"
    rank: "01"
    title: "Production Multi-AZ VPC"
    excerpt: "This preset creates a production-ready VPC spanning two Availability Zones with NAT gateway, DNS support, and DNS hostnames enabled. The `/16` CIDR provides 65,536 IP addresses, enough for most..."
  - slug: "02-development"
    rank: "02"
    title: "Development VPC"
    excerpt: "This preset creates a minimal VPC in a single Availability Zone without a NAT gateway. This reduces costs significantly (NAT gateway charges ~$32/month + data transfer) while still providing a..."
---

# VPC Presets

Ready-to-deploy configuration presets for VPC. Each preset is a complete manifest you can copy, customize, and deploy.
