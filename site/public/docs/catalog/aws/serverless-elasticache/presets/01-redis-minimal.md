---
title: "Preset: Redis Minimal"
description: "**Use case:** Development, prototyping, or low-traffic applications where you want zero configuration overhead."
type: "preset"
rank: "01"
presetSlug: "01-redis-minimal"
componentSlug: "serverless-elasticache"
componentTitle: "Serverless ElastiCache"
provider: "aws"
icon: "package"
order: 1
---

# Preset: Redis Minimal

**Use case:** Development, prototyping, or low-traffic applications where you want
zero configuration overhead.

**What it creates:**
- A Redis 7.x serverless cache with all AWS defaults
- No scaling limits — AWS manages capacity automatically
- No VPC placement — uses default networking
- AWS-managed encryption (always on for serverless)

**Cost profile:** Minimal — you pay only for actual ECPU and storage consumption.

**When to upgrade:** When you need explicit scaling bounds, VPC isolation, customer-managed encryption, or snapshots.
