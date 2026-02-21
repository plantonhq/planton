---
title: "Presets"
description: "Ready-to-deploy configuration presets for Network Security Group"
type: "preset-list"
componentSlug: "network-security-group"
componentTitle: "Network Security Group"
provider: "oci"
icon: "package"
order: 200
presets:
  - slug: "01-web-tier"
    rank: "01"
    title: "Web Tier NSG"
    excerpt: "This preset creates a Network Security Group for internet-facing resources such as load balancers, web servers, and API gateways. Inbound traffic is restricted to HTTP (80) and HTTPS (443) from..."
  - slug: "02-private-backend"
    rank: "02"
    title: "Private Backend NSG"
    excerpt: "This preset creates a Network Security Group for resources that should only be reachable from within the VCN. All protocols and ports are allowed from the VCN CIDR block, while traffic from outside..."
  - slug: "03-development"
    rank: "03"
    title: "Development NSG"
    excerpt: "This preset creates a fully permissive Network Security Group that allows all inbound and outbound traffic on all protocols and ports. This is the simplest NSG configuration, suitable for..."
---

# Network Security Group Presets

Ready-to-deploy configuration presets for Network Security Group. Each preset is a complete manifest you can copy, customize, and deploy.
