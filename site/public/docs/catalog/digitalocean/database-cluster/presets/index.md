---
title: "Presets"
description: "Ready-to-deploy configuration presets for Database Cluster"
type: "preset-list"
componentSlug: "database-cluster"
componentTitle: "Database Cluster"
provider: "digitalocean"
icon: "package"
order: 200
presets:
  - slug: "01-postgresql-ha"
    rank: "01"
    title: "Production PostgreSQL HA"
    excerpt: "This preset creates a production-grade PostgreSQL database cluster with three nodes for high availability, VPC isolation for secure private access, and PostgreSQL 16 on a 2 vCPU / 4 GB node size...."
  - slug: "02-postgresql-dev"
    rank: "02"
    title: "Development PostgreSQL"
    excerpt: "This preset creates a single-node PostgreSQL database for development and testing. No VPC is required, and the smallest node size keeps costs minimal. Ideal for local development, CI/CD test..."
  - slug: "03-redis"
    rank: "03"
    title: "Redis Cache Cluster"
    excerpt: "This preset creates a managed Redis cache cluster with one node, suitable for session storage, caching, and pub/sub workloads. VPC placement keeps cache traffic within your private network. Redis 7..."
---

# Database Cluster Presets

Ready-to-deploy configuration presets for Database Cluster. Each preset is a complete manifest you can copy, customize, and deploy.
