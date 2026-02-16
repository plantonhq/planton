---
title: "Presets"
description: "Ready-to-deploy configuration presets for Redis ElastiCache"
type: "preset-list"
componentSlug: "redis-elasticache"
componentTitle: "Redis ElastiCache"
provider: "aws"
icon: "package"
order: 200
presets:
  - slug: "01-redis-single-node"
    rank: "01"
    title: "Redis Single Node"
    excerpt: "This preset creates a single-node Redis 7.1 cluster with encryption enabled. It is the fastest way to get a development or testing cache running with secure defaults."
  - slug: "02-redis-ha-cluster"
    rank: "02"
    title: "Redis HA Cluster"
    excerpt: "This preset creates a 3-node Redis 7.1 cluster (1 primary + 2 read replicas) with automatic failover, multi-AZ deployment, encryption, daily snapshots, and a tuned eviction policy. Production-ready..."
  - slug: "03-redis-clustered-production"
    rank: "03"
    title: "Redis Clustered Production"
    excerpt: "This preset creates a Cluster Mode Enabled Redis 7.1 deployment with 3 shards, 2 replicas per shard (9 total nodes), customer-managed KMS encryption, slow-log delivery to CloudWatch, SNS event..."
---

# Redis ElastiCache Presets

Ready-to-deploy configuration presets for Redis ElastiCache. Each preset is a complete manifest you can copy, customize, and deploy.
