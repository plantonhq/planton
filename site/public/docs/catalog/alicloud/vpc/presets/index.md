---
title: "Presets"
description: "Ready-to-deploy configuration presets for VPC"
type: "preset-list"
componentSlug: "vpc"
componentTitle: "VPC"
provider: "alicloud"
icon: "package"
order: 200
presets:
  - slug: "01-standard-production"
    rank: "01"
    title: "Standard Production VPC"
    excerpt: "This preset creates a production-ready Alibaba Cloud VPC with a /16 CIDR block providing 65,536 IP addresses. The address space accommodates dozens of VSwitches across multiple availability zones..."
  - slug: "02-development"
    rank: "02"
    title: "Development VPC"
    excerpt: "This preset creates a minimal VPC for development and testing environments. It uses a 192.168.x.x CIDR range (different from the 10.x production range) so development and production VPCs can coexist..."
  - slug: "03-dual-stack-ipv6"
    rank: "03"
    title: "Dual-Stack IPv6 VPC"
    excerpt: "This preset creates a VPC with dual-stack networking enabled. When `enableIpv6` is set to `true`, Alibaba Cloud allocates a /56 IPv6 CIDR block to the VPC in addition to the IPv4 CIDR. VSwitches..."
---

# VPC Presets

Ready-to-deploy configuration presets for VPC. Each preset is a complete manifest you can copy, customize, and deploy.
