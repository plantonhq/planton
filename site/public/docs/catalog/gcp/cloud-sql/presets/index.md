---
title: "Presets"
description: "Ready-to-deploy configuration presets for Cloud SQL"
type: "preset-list"
componentSlug: "cloud-sql"
componentTitle: "Cloud SQL"
provider: "gcp"
icon: "package"
order: 200
presets:
  - slug: "01-postgresql-production"
    rank: "01"
    title: "PostgreSQL Production Instance"
    excerpt: "This preset creates a production-grade Cloud SQL PostgreSQL instance with high availability, private IP networking, automated backups with point-in-time recovery, query insights, and deletion..."
  - slug: "02-mysql-production"
    rank: "02"
    title: "MySQL Production Instance"
    excerpt: "This preset creates a production-grade Cloud SQL MySQL 8.0 instance with the same security and reliability posture as the PostgreSQL production preset: high availability, private IP, automated..."
  - slug: "03-postgresql-development"
    rank: "03"
    title: "PostgreSQL Development Instance"
    excerpt: "This preset creates a minimal Cloud SQL PostgreSQL instance for development and testing. It uses the smallest available tier, public IP access (via Cloud SQL Proxy), and basic backups. No high..."
---

# Cloud SQL Presets

Ready-to-deploy configuration presets for Cloud SQL. Each preset is a complete manifest you can copy, customize, and deploy.
