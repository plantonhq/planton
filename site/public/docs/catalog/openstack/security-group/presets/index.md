---
title: "Presets"
description: "Ready-to-deploy configuration presets for Security Group"
type: "preset-list"
componentSlug: "security-group"
componentTitle: "Security Group"
provider: "openstack"
icon: "package"
order: 200
presets:
  - slug: "01-web-server"
    rank: "01"
    title: "Web Server Security Group"
    excerpt: "This preset creates a security group for a typical web server: SSH from a trusted network, HTTP and HTTPS from anywhere, and unrestricted egress (OpenStack's default). The default egress rules are..."
  - slug: "02-restrictive"
    rank: "02"
    title: "Restrictive Security Group (Zero-Trust Baseline)"
    excerpt: "This preset creates a security group with OpenStack's default egress rules deleted, providing a zero-trust starting point. Only explicitly defined rules are active: SSH from a trusted CIDR and all..."
---

# Security Group Presets

Ready-to-deploy configuration presets for Security Group. Each preset is a complete manifest you can copy, customize, and deploy.
