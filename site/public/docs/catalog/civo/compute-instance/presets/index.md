---
title: "Presets"
description: "Ready-to-deploy configuration presets for Compute Instance"
type: "preset-list"
componentSlug: "compute-instance"
componentTitle: "Compute Instance"
provider: "civo"
icon: "package"
order: 200
presets:
  - slug: "01-production-web"
    rank: "01"
    title: "Production Web Server"
    excerpt: "This preset creates a production-grade compute instance on a medium-sized node with Ubuntu 22.04 LTS, VPC networking, firewall protection, and a cloud-init script that applies security updates on..."
  - slug: "02-development"
    rank: "02"
    title: "Development Instance"
    excerpt: "This preset creates a minimal, cost-effective compute instance for development and testing. Uses the smallest instance size with VPC networking but no explicit firewall or cloud-init, keeping..."
---

# Compute Instance Presets

Ready-to-deploy configuration presets for Compute Instance. Each preset is a complete manifest you can copy, customize, and deploy.
