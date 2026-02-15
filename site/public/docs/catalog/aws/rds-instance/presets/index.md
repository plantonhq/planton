---
title: "Presets"
description: "Ready-to-deploy configuration presets for RDS Instance"
type: "preset-list"
componentSlug: "rds-instance"
componentTitle: "RDS Instance"
provider: "aws"
icon: "package"
order: 200
presets:
  - slug: "01-postgresql-production"
    rank: "01"
    title: "PostgreSQL Production Instance"
    excerpt: "This preset creates a Multi-AZ RDS PostgreSQL instance with encrypted storage and private network access. Multi-AZ deploys a synchronous standby replica in a different Availability Zone for automatic..."
  - slug: "02-mysql-production"
    rank: "02"
    title: "MySQL Production Instance"
    excerpt: "This preset creates a Multi-AZ RDS MySQL instance with encrypted storage and private network access. Same production-grade defaults as the PostgreSQL preset, but configured for MySQL 8.0. Suitable..."
---

# RDS Instance Presets

Ready-to-deploy configuration presets for RDS Instance. Each preset is a complete manifest you can copy, customize, and deploy.
