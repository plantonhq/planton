---
title: "Presets"
description: "Ready-to-deploy configuration presets for Firewall"
type: "preset-list"
componentSlug: "firewall"
componentTitle: "Firewall"
provider: "civo"
icon: "package"
order: 200
presets:
  - slug: "01-web-tier"
    rank: "01"
    title: "Web Tier Firewall"
    excerpt: "This preset creates a firewall allowing inbound HTTP (80), HTTPS (443), and restricted SSH (22) access. It is the most common configuration for internet-facing web servers, API gateways, and reverse..."
  - slug: "02-database-tier"
    rank: "02"
    title: "Database Tier Firewall"
    excerpt: "This preset creates a firewall that restricts inbound access to standard database ports (PostgreSQL 5432, MySQL 3306) from the application tier CIDR only. No public internet access is permitted,..."
---

# Firewall Presets

Ready-to-deploy configuration presets for Firewall. Each preset is a complete manifest you can copy, customize, and deploy.
