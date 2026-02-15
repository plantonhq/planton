---
title: "Presets"
description: "Ready-to-deploy configuration presets for Network Security Group"
type: "preset-list"
componentSlug: "network-security-group"
componentTitle: "Network Security Group"
provider: "azure"
icon: "package"
order: 200
presets:
  - slug: "01-web-tier"
    rank: "01"
    title: "Web Tier NSG"
    excerpt: "This preset creates a Network Security Group for web-facing subnets, allowing inbound HTTP and HTTPS traffic from the internet. This is the standard NSG for subnets hosting load balancers,..."
  - slug: "02-database-tier"
    rank: "02"
    title: "Database Tier NSG"
    excerpt: "This preset creates a Network Security Group for database subnets, allowing only PostgreSQL and MySQL traffic from within the Virtual Network and explicitly denying all internet inbound traffic. This..."
  - slug: "03-bastion"
    rank: "03"
    title: "Bastion NSG"
    excerpt: "This preset creates a Network Security Group for bastion or jump-host subnets, allowing SSH and RDP access only from trusted IP ranges. All other internet traffic is explicitly denied. This is the..."
---

# Network Security Group Presets

Ready-to-deploy configuration presets for Network Security Group. Each preset is a complete manifest you can copy, customize, and deploy.
