---
title: "Presets"
description: "Ready-to-deploy configuration presets for Security Group"
type: "preset-list"
componentSlug: "security-group"
componentTitle: "Security Group"
provider: "aws"
icon: "package"
order: 200
presets:
  - slug: "01-web-tier"
    rank: "01"
    title: "Web Tier Security Group"
    excerpt: "This preset creates a security group for internet-facing web servers or load balancers. It allows inbound HTTP (80) and HTTPS (443) traffic from any source and permits all outbound traffic. This is..."
  - slug: "02-database-tier"
    rank: "02"
    title: "Database Tier Security Group"
    excerpt: "This preset creates a security group for database instances that only accepts connections from the application tier. Ingress is restricted to PostgreSQL port 5432 from a specific CIDR block..."
  - slug: "03-bastion"
    rank: "03"
    title: "Bastion Host Security Group"
    excerpt: "This preset creates a security group for a bastion (jump) host that only accepts SSH connections from trusted IP addresses. Never use `0.0.0.0/0` for bastion SSH access -- always restrict to your..."
---

# Security Group Presets

Ready-to-deploy configuration presets for Security Group. Each preset is a complete manifest you can copy, customize, and deploy.
