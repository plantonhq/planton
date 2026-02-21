---
title: "Presets"
description: "Ready-to-deploy configuration presets for Redis Cluster"
type: "preset-list"
componentSlug: "redis-cluster"
componentTitle: "Redis Cluster"
provider: "oci"
icon: "package"
order: 200
presets:
  - slug: "01-non-sharded-cluster"
    rank: "01"
    title: "Non-Sharded Cluster"
    excerpt: "This preset creates a non-sharded OCI Cache (Redis) cluster with 3 nodes: one primary and two replicas. The primary handles all writes while replicas serve read traffic and provide automatic..."
  - slug: "02-sharded-cluster"
    rank: "02"
    title: "Sharded Cluster"
    excerpt: "This preset creates a sharded OCI Cache (Redis) cluster with 3 shards and 3 nodes per shard (9 nodes total). Data is automatically distributed across shards using Redis Cluster's hash slot mechanism,..."
---

# Redis Cluster Presets

Ready-to-deploy configuration presets for Redis Cluster. Each preset is a complete manifest you can copy, customize, and deploy.
