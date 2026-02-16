---
title: "Presets"
description: "Ready-to-deploy configuration presets for Memorystore Instance"
type: "preset-list"
componentSlug: "memorystore-instance"
componentTitle: "Memorystore Instance"
provider: "gcp"
icon: "package"
order: 200
presets:
  - slug: "01-dev-single-shard"
    rank: "01"
    title: "Dev Single Shard"
    excerpt: "This preset provisions a minimal Memorystore instance in standalone (CLUSTER_DISABLED) mode with a single shard and the smallest available node type. It is ideal for development, testing, or..."
  - slug: "02-ha-production"
    rank: "02"
    title: "HA Production"
    excerpt: "This preset provisions a production-ready Memorystore instance in CLUSTER mode with 3 shards, 1 replica per shard, TLS encryption, RDB persistence, multi-zone distribution, a maintenance window, and..."
  - slug: "03-enterprise-cluster"
    rank: "03"
    title: "Enterprise Cluster"
    excerpt: "This preset provisions a fully-featured Memorystore instance in CLUSTER mode with 5 shards, 2 replicas per shard, IAM authentication, TLS encryption, customer-managed encryption keys (CMEK), AOF..."
---

# Memorystore Instance Presets

Ready-to-deploy configuration presets for Memorystore Instance. Each preset is a complete manifest you can copy, customize, and deploy.
