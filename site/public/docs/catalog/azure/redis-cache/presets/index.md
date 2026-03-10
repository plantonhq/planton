---
title: "Presets"
description: "Ready-to-deploy configuration presets for Redis Cache"
type: "preset-list"
componentSlug: "redis-cache"
componentTitle: "Redis Cache"
provider: "azure"
icon: "package"
order: 200
presets:
  - slug: "01-standard"
    rank: "01"
    title: "Standard Redis Cache"
    excerpt: "This preset creates an Azure Cache for Redis with Standard tier (primary + replica), 1 GB cache size (C1), Redis 6, and TLS 1.2 enforcement. The Standard tier provides a replicated two-node cache..."
  - slug: "02-premium-vnet"
    rank: "02"
    title: "Premium Redis Cache with VNet Injection"
    excerpt: "This preset creates an Azure Cache for Redis with Premium tier injected into a virtual network subnet. VNet injection provides private IP addressing and network isolation -- the cache is not..."
  - slug: "03-development"
    rank: "03"
    title: "Development Redis Cache"
    excerpt: "This preset creates an Azure Cache for Redis with Basic tier, the smallest cache size (C0, 250 MB), and `allkeys-lru` eviction for pure caching workloads. Basic tier provides a single-node cache with..."
---

# Redis Cache Presets

Ready-to-deploy configuration presets for Redis Cache. Each preset is a complete manifest you can copy, customize, and deploy.
