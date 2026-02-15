---
title: "Presets"
description: "Ready-to-deploy configuration presets for Droplet"
type: "preset-list"
componentSlug: "droplet"
componentTitle: "Droplet"
provider: "digitalocean"
icon: "package"
order: 200
presets:
  - slug: "01-production"
    rank: "01"
    title: "Production Droplet"
    excerpt: "This preset creates a production-ready DigitalOcean Droplet with automated backups enabled, VPC isolation, and resource tags for firewall targeting. It uses a general-purpose 2 vCPU / 4 GB instance..."
  - slug: "02-development"
    rank: "02"
    title: "Development Droplet"
    excerpt: "This preset creates a minimal DigitalOcean Droplet for development and testing. It uses the smallest general-purpose instance with no backups, keeping costs low while still providing VPC isolation."
---

# Droplet Presets

Ready-to-deploy configuration presets for Droplet. Each preset is a complete manifest you can copy, customize, and deploy.
