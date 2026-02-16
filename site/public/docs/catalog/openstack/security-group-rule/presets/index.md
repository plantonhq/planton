---
title: "Presets"
description: "Ready-to-deploy configuration presets for Security Group Rule"
type: "preset-list"
componentSlug: "security-group-rule"
componentTitle: "Security Group Rule"
provider: "openstack"
icon: "package"
order: 200
presets:
  - slug: "01-allow-ssh"
    rank: "01"
    title: "Allow SSH Ingress Rule"
    excerpt: "This preset creates a standalone security group rule that allows inbound SSH (TCP port 22) from a trusted CIDR. Use standalone rules (instead of inline rules in `OpenStackSecurityGroup`) when..."
  - slug: "02-allow-http-https"
    rank: "02"
    title: "Allow HTTPS Ingress Rule"
    excerpt: "This preset creates a standalone security group rule that allows inbound HTTPS (TCP port 443) from any source. This is the most common standalone rule for web-facing services. For HTTP (port 80),..."
---

# Security Group Rule Presets

Ready-to-deploy configuration presets for Security Group Rule. Each preset is a complete manifest you can copy, customize, and deploy.
