---
title: "Presets"
description: "Ready-to-deploy configuration presets for Firewall"
type: "preset-list"
componentSlug: "firewall"
componentTitle: "Firewall"
provider: "digitalocean"
icon: "package"
order: 200
presets:
  - slug: "01-web-tier"
    rank: "01"
    title: "Web Tier Firewall"
    excerpt: "This preset creates a DigitalOcean Cloud Firewall for web-facing Droplets. It allows inbound HTTP/HTTPS from anywhere, restricts SSH to a management CIDR, and permits all outbound traffic. The..."
  - slug: "02-database-tier"
    rank: "02"
    title: "Database Tier Firewall"
    excerpt: "This preset creates a DigitalOcean Cloud Firewall for database Droplets. It restricts inbound access to PostgreSQL (port 5432) from web-tier Droplets only, and limits SSH to a management CIDR...."
---

# Firewall Presets

Ready-to-deploy configuration presets for Firewall. Each preset is a complete manifest you can copy, customize, and deploy.
