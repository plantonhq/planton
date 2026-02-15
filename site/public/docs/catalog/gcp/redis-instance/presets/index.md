---
title: "Presets"
description: "Ready-to-deploy configuration presets for Redis Instance"
type: "preset-list"
componentSlug: "redis-instance"
componentTitle: "Redis Instance"
provider: "gcp"
icon: "package"
order: 200
presets:
  - slug: "01-basic-cache"
    rank: "01"
    title: "Basic Cache"
    excerpt: "This preset provisions a minimal Memorystore for Redis instance using the BASIC tier with 1 GB memory. It is ideal for development, testing, or lightweight caching workloads where high availability..."
  - slug: "02-ha-production"
    rank: "02"
    title: "HA Production"
    excerpt: "This preset provisions a production-ready Memorystore for Redis instance with STANDARD_HA tier, authentication, TLS encryption, RDB persistence, a maintenance window, and deletion protection. It is..."
  - slug: "03-ha-read-replicas"
    rank: "03"
    title: "HA with Read Replicas"
    excerpt: "This preset provisions a Memorystore for Redis instance with STANDARD_HA tier, three read replicas, authentication, TLS, RDB persistence, and customer-managed encryption. It is designed for..."
---

# Redis Instance Presets

Ready-to-deploy configuration presets for Redis Instance. Each preset is a complete manifest you can copy, customize, and deploy.
