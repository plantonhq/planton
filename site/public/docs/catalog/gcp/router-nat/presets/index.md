---
title: "Presets"
description: "Ready-to-deploy configuration presets for Router NAT"
type: "preset-list"
componentSlug: "router-nat"
componentTitle: "Router NAT"
provider: "gcp"
icon: "package"
order: 200
presets:
  - slug: "01-all-subnets-auto"
    rank: "01"
    title: "All-Subnets Auto-Allocated NAT"
    excerpt: "This preset creates a Cloud Router with a NAT gateway that covers all subnets in the region using automatically allocated external IPs. This is the simplest and most common Cloud NAT configuration,..."
  - slug: "02-static-ip-specific-subnets"
    rank: "02"
    title: "Static IP NAT for Specific Subnets"
    excerpt: "This preset creates a Cloud Router with NAT restricted to specific subnets using manually assigned static external IPs. Use this when you need predictable egress IP addresses -- for example, when..."
---

# Router NAT Presets

Ready-to-deploy configuration presets for Router NAT. Each preset is a complete manifest you can copy, customize, and deploy.
