---
title: "Presets"
description: "Ready-to-deploy configuration presets for VPN Gateway"
type: "preset-list"
componentSlug: "vpn-gateway"
componentTitle: "VPN Gateway"
provider: "alicloud"
icon: "package"
order: 200
presets:
  - slug: "01-basic-site-to-site"
    rank: "01"
    title: "Basic Site-to-Site VPN"
    excerpt: "This preset creates a VPN Gateway with a single IPsec connection to one remote site. This is the most common pattern for connecting an Alibaba Cloud VPC to an on-premises data center or branch office."
  - slug: "02-production-multi-site"
    rank: "02"
    title: "Production Multi-Site VPN"
    excerpt: "This preset creates a VPN Gateway connecting to two remote sites with production-grade encryption (AES-256 + SHA-256 + DH group14) and health check monitoring on both tunnels."
  - slug: "03-ssl-enabled"
    rank: "03"
    title: "SSL VPN Enabled Gateway"
    excerpt: "This preset creates a VPN Gateway with SSL VPN enabled for remote client access, alongside a site-to-site IPsec connection. SSL VPN allows individual users (developers, admins) to connect to the VPC..."
---

# VPN Gateway Presets

Ready-to-deploy configuration presets for VPN Gateway. Each preset is a complete manifest you can copy, customize, and deploy.
