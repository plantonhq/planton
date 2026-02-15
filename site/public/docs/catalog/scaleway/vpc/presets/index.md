---
title: "Presets"
description: "Ready-to-deploy configuration presets for VPC"
type: "preset-list"
componentSlug: "vpc"
componentTitle: "VPC"
provider: "scaleway"
icon: "package"
order: 200
presets:
  - slug: "01-standard"
    rank: "01"
    title: "Standard VPC"
    excerpt: "This preset creates a Scaleway VPC in the Paris region with routing disabled. A VPC is a regional container that groups Private Networks. This is the simplest and most common starting configuration,..."
  - slug: "02-routing-enabled"
    rank: "02"
    title: "Routing-Enabled VPC"
    excerpt: "This preset creates a Scaleway VPC with inter-Private-Network routing enabled. When routing is on, resources in different Private Networks attached to this VPC can communicate with each other. This..."
---

# VPC Presets

Ready-to-deploy configuration presets for VPC. Each preset is a complete manifest you can copy, customize, and deploy.
