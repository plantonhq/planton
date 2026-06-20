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
  - slug: "01-production-dual-stack"
    rank: "01"
    title: "Production Dual-Stack VPC"
    excerpt: "This preset creates a production-ready VPC with a `/16` IPv4 CIDR (65,536 addresses) and an Amazon-provided IPv6 `/56`, with DNS support and DNS hostnames enabled. It is the standard foundation for a..."
  - slug: "02-development"
    rank: "02"
    title: "Development VPC"
    excerpt: "This preset creates a minimal IPv4-only VPC with a `/16` CIDR and DNS enabled -- a clean foundation for development and testing networks. Add `AwsSubnet` components (and, if outbound internet access..."
---

# VPC Presets

Ready-to-deploy configuration presets for VPC. Each preset is a complete manifest you can copy, customize, and deploy.
