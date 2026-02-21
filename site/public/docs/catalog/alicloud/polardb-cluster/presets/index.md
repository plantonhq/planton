---
title: "Presets"
description: "Ready-to-deploy configuration presets for PolarDB Cluster"
type: "preset-list"
componentSlug: "polardb-cluster"
componentTitle: "PolarDB Cluster"
provider: "alicloud"
icon: "package"
order: 200
presets:
  - slug: "01-mysql-dev"
    rank: "01"
    title: "MySQL Development Cluster"
    excerpt: "This preset creates a minimal MySQL 8.0 PolarDB cluster with a single node, suitable for development and testing."
  - slug: "02-mysql-production"
    rank: "02"
    title: "MySQL Production Cluster"
    excerpt: "This preset creates a production-grade MySQL 8.0 PolarDB cluster with 4 nodes, TDE encryption, audit logging, and deletion protection."
  - slug: "03-postgresql-production"
    rank: "03"
    title: "PostgreSQL Production Cluster"
    excerpt: "This preset creates a production PostgreSQL 14 PolarDB cluster with 3 nodes, proper collation settings, and deletion protection."
---

# PolarDB Cluster Presets

Ready-to-deploy configuration presets for PolarDB Cluster. Each preset is a complete manifest you can copy, customize, and deploy.
