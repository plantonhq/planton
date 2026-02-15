---
title: "Presets"
description: "Ready-to-deploy configuration presets for VPC (Virtual Network)"
type: "preset-list"
componentSlug: "vpc-virtual-network"
componentTitle: "VPC (Virtual Network)"
provider: "azure"
icon: "package"
order: 200
presets:
  - slug: "01-production-nat"
    rank: "01"
    title: "Production VNet with NAT Gateway"
    excerpt: "This preset creates an Azure Virtual Network with a /16 address space, a /18 nodes subnet, and a NAT Gateway for outbound internet connectivity. This is the standard production configuration for..."
  - slug: "02-development"
    rank: "02"
    title: "Development VNet"
    excerpt: "This preset creates an Azure Virtual Network with a /16 address space and a smaller /20 nodes subnet without a NAT Gateway. This is a cost-effective configuration for development and testing..."
---

# VPC (Virtual Network) Presets

Ready-to-deploy configuration presets for VPC (Virtual Network). Each preset is a complete manifest you can copy, customize, and deploy.
