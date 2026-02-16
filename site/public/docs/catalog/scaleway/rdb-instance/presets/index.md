---
title: "Presets"
description: "Ready-to-deploy configuration presets for RDB Instance"
type: "preset-list"
componentSlug: "rdb-instance"
componentTitle: "RDB Instance"
provider: "scaleway"
icon: "package"
order: 200
presets:
  - slug: "01-dev-postgres"
    rank: "01"
    title: "Development PostgreSQL Instance"
    excerpt: "This preset creates a minimal Scaleway RDB instance running PostgreSQL 16 on the smallest available node type. It uses local SSD storage and has no Private Network attachment or HA -- the simplest..."
  - slug: "02-production-postgres-ha"
    rank: "02"
    title: "Production PostgreSQL HA Instance"
    excerpt: "This preset creates a high-availability Scaleway RDB instance running PostgreSQL 16 with a standby replica, Private Network connectivity, encryption at rest, and frequent backups. This is the..."
  - slug: "03-mysql-web-app"
    rank: "03"
    title: "MySQL Web Application Database"
    excerpt: "This preset creates a Scaleway RDB instance running MySQL 8 on a general-purpose node with Private Network connectivity. It is sized for typical web application databases (CMS, e-commerce, SaaS)..."
---

# RDB Instance Presets

Ready-to-deploy configuration presets for RDB Instance. Each preset is a complete manifest you can copy, customize, and deploy.
