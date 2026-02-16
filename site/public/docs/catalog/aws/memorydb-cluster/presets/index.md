---
title: "Presets"
description: "Ready-to-deploy configuration presets for MemoryDB Cluster"
type: "preset-list"
componentSlug: "memorydb-cluster"
componentTitle: "MemoryDB Cluster"
provider: "aws"
icon: "package"
order: 200
presets:
  - slug: "01-dev-single-shard"
    rank: "01"
    title: "Dev Single Shard"
    excerpt: "This preset creates a single-shard, single-node MemoryDB cluster with TLS encryption and the open-access ACL. It is the simplest way to get a durable Redis-compatible database running for development."
  - slug: "02-production-ha"
    rank: "02"
    title: "Production HA"
    excerpt: "This preset creates a 2-shard MemoryDB cluster with 2 replicas per shard (6 total nodes), daily snapshots, a tuned eviction policy, and active defragmentation. Production-ready for session stores,..."
  - slug: "03-high-throughput"
    rank: "03"
    title: "High-Throughput with Data Tiering"
    excerpt: "This preset creates a 4-shard MemoryDB cluster with 2 replicas per shard (12 total nodes), data tiering for cost-efficient cold data management, customer-managed KMS encryption, SNS notifications,..."
---

# MemoryDB Cluster Presets

Ready-to-deploy configuration presets for MemoryDB Cluster. Each preset is a complete manifest you can copy, customize, and deploy.
