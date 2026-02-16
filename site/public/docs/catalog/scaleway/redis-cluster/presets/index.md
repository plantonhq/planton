---
title: "Presets"
description: "Ready-to-deploy configuration presets for Redis Cluster"
type: "preset-list"
componentSlug: "redis-cluster"
componentTitle: "Redis Cluster"
provider: "scaleway"
icon: "package"
order: 200
presets:
  - slug: "01-dev-standalone"
    rank: "01"
    title: "Development Standalone Redis"
    excerpt: "This preset creates a single-node Scaleway Redis cluster using the smallest available node type with public ACL access. It is the fastest path to a working Redis instance for development, caching,..."
  - slug: "02-production-ha"
    rank: "02"
    title: "Production HA Redis Cluster"
    excerpt: "This preset creates a 3-node Scaleway Redis cluster with TLS encryption and Private Network connectivity. The cluster provides automatic failover -- if the primary fails, a replica is promoted within..."
---

# Redis Cluster Presets

Ready-to-deploy configuration presets for Redis Cluster. Each preset is a complete manifest you can copy, customize, and deploy.
