---
title: "Presets"
description: "Ready-to-deploy configuration presets for Instance Security Group"
type: "preset-list"
componentSlug: "instance-security-group"
componentTitle: "Instance Security Group"
provider: "scaleway"
icon: "package"
order: 200
presets:
  - slug: "01-web-server"
    rank: "01"
    title: "Web Server Security Group"
    excerpt: "This preset creates a security group for web-facing instances using an allowlist model. Inbound traffic is dropped by default, with explicit rules accepting SSH (from a restricted CIDR), HTTP, and..."
  - slug: "02-deny-all-allowlist"
    rank: "02"
    title: "Deny-All Allowlist Security Group"
    excerpt: "This preset creates a strict security group for internal services. All inbound traffic is dropped by default, with rules accepting only TCP traffic from the private network range (10.0.0.0/8) and SSH..."
---

# Instance Security Group Presets

Ready-to-deploy configuration presets for Instance Security Group. Each preset is a complete manifest you can copy, customize, and deploy.
