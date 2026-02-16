---
title: "Presets"
description: "Ready-to-deploy configuration presets for Redis"
type: "preset-list"
componentSlug: "redis"
componentTitle: "Redis"
provider: "kubernetes"
icon: "package"
order: 200
presets:
  - slug: "01-single-instance"
    rank: "01"
    title: "Single Instance Redis"
    excerpt: "This preset deploys a single-replica Redis instance with persistence enabled and 1Gi of disk storage. The most common Redis configuration for caching and session storage."
  - slug: "02-persistent-with-replicas"
    rank: "02"
    title: "Production Redis with Replicas"
    excerpt: "This preset deploys a 3-replica Redis deployment with persistence and production-grade resources. Provides read scaling and data durability through replication and persistent storage."
---

# Redis Presets

Ready-to-deploy configuration presets for Redis. Each preset is a complete manifest you can copy, customize, and deploy.
