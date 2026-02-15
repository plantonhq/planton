---
title: "Presets"
description: "Ready-to-deploy configuration presets for Postgres"
type: "preset-list"
componentSlug: "postgres"
componentTitle: "Postgres"
provider: "kubernetes"
icon: "package"
order: 200
presets:
  - slug: "01-single-instance"
    rank: "01"
    title: "Single Instance PostgreSQL"
    excerpt: "This preset deploys a single-replica PostgreSQL instance with 10Gi of persistent storage and no backup configuration. Suitable for development, testing, or low-criticality workloads."
  - slug: "02-production-with-backup"
    rank: "02"
    title: "Production PostgreSQL with Backup"
    excerpt: "This preset deploys a 2-replica PostgreSQL cluster with streaming replication, daily backups, pre-configured database and user, and production-grade resources."
---

# Postgres Presets

Ready-to-deploy configuration presets for Postgres. Each preset is a complete manifest you can copy, customize, and deploy.
