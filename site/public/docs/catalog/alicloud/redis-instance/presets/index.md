---
title: "Presets"
description: "Ready-to-deploy configuration presets for Redis Instance"
type: "preset-list"
componentSlug: "redis-instance"
componentTitle: "Redis Instance"
provider: "alicloud"
icon: "package"
order: 200
presets:
  - slug: "01-standard-single"
    rank: "01"
    title: "Standard Single-Zone Redis"
    excerpt: "This preset creates a minimal Redis 7.0 instance for development and testing, suitable for single-zone deployments with basic configuration."
  - slug: "02-ha-cluster"
    rank: "02"
    title: "HA Cluster Redis"
    excerpt: "This preset creates a production-grade Redis cluster with multi-zone high availability, horizontal sharding for throughput, and read replicas for read scaling."
  - slug: "03-production-encrypted"
    rank: "03"
    title: "Production Encrypted Redis"
    excerpt: "This preset creates a security-hardened Redis instance with TDE encryption at rest, SSL encryption in transit, subscription billing, and daily backups -- designed for compliance-sensitive..."
---

# Redis Instance Presets

Ready-to-deploy configuration presets for Redis Instance. Each preset is a complete manifest you can copy, customize, and deploy.
