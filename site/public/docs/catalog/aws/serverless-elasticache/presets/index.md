---
title: "Presets"
description: "Ready-to-deploy configuration presets for Serverless ElastiCache"
type: "preset-list"
componentSlug: "serverless-elasticache"
componentTitle: "Serverless ElastiCache"
provider: "aws"
icon: "package"
order: 200
presets:
  - slug: "01-redis-minimal"
    rank: "01"
    title: "Preset: Redis Minimal"
    excerpt: "**Use case:** Development, prototyping, or low-traffic applications where you want zero configuration overhead."
  - slug: "02-memcached-with-limits"
    rank: "02"
    title: "Preset: Memcached with Scaling Limits"
    excerpt: "**Use case:** Web application response caching or session storage where you want cost control through explicit scaling bounds."
  - slug: "03-redis-production"
    rank: "03"
    title: "Preset: Redis Production"
    excerpt: "**Use case:** Production workloads requiring encryption, VPC isolation, daily snapshots, and Redis ACL access control."
---

# Serverless ElastiCache Presets

Ready-to-deploy configuration presets for Serverless ElastiCache. Each preset is a complete manifest you can copy, customize, and deploy.
