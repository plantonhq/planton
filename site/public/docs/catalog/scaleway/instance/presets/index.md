---
title: "Presets"
description: "Ready-to-deploy configuration presets for Instance"
type: "preset-list"
componentSlug: "instance"
componentTitle: "Instance"
provider: "scaleway"
icon: "package"
order: 200
presets:
  - slug: "01-dev-instance"
    rank: "01"
    title: "Development Instance"
    excerpt: "This preset creates a small Scaleway Instance with a public IP for quick development and testing. It uses the DEV1-S type (2 vCPU, 2 GB RAM) with Ubuntu 22.04 -- the most affordable and commonly used..."
  - slug: "02-production-private"
    rank: "02"
    title: "Production Private Instance"
    excerpt: "This preset creates a production-grade Scaleway Instance on a Private Network with no public IP, an explicit security group, and deletion protection enabled. The instance is only reachable through..."
---

# Instance Presets

Ready-to-deploy configuration presets for Instance. Each preset is a complete manifest you can copy, customize, and deploy.
