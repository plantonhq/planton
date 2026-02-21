---
title: "Presets"
description: "Ready-to-deploy configuration presets for Security Group"
type: "preset-list"
componentSlug: "security-group"
componentTitle: "Security Group"
provider: "alicloud"
icon: "package"
order: 200
presets:
  - slug: "01-web-tier"
    rank: "01"
    title: "Web Tier Security Group"
    excerpt: "This preset creates a security group suitable for public-facing web servers or load balancers. It allows HTTP (port 80) and HTTPS (port 443) inbound from any source, with unrestricted outbound access."
  - slug: "02-database-tier"
    rank: "02"
    title: "Database Tier Security Group"
    excerpt: "This preset creates a locked-down security group for database instances (RDS, PolarDB, Redis, MongoDB). It only allows connections on standard database ports from the VPC CIDR range, with no public..."
  - slug: "03-bastion-host"
    rank: "03"
    title: "Bastion Host Security Group"
    excerpt: "This preset creates a security group for bastion (jump) hosts that serve as the single entry point into a VPC. It allows SSH inbound from a trusted network and restricts outbound to SSH and database..."
---

# Security Group Presets

Ready-to-deploy configuration presets for Security Group. Each preset is a complete manifest you can copy, customize, and deploy.
