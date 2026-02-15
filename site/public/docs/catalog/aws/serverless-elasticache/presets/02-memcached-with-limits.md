---
title: "Preset: Memcached with Scaling Limits"
description: "**Use case:** Web application response caching or session storage where you want cost control through explicit scaling bounds."
type: "preset"
rank: "02"
presetSlug: "02-memcached-with-limits"
componentSlug: "serverless-elasticache"
componentTitle: "Serverless ElastiCache"
provider: "aws"
icon: "package"
order: 2
---

# Preset: Memcached with Scaling Limits

**Use case:** Web application response caching or session storage where you want
cost control through explicit scaling bounds.

**What it creates:**
- A Memcached serverless cache with bounded auto-scaling
- Data: 1–50 GB range
- Compute: 1,000–25,000 ECPU/sec range
- No VPC placement — uses default networking
- AWS-managed encryption (always on for serverless)

**Cost profile:** Predictable — scaling bounds prevent cost surprises from traffic spikes.

**Trade-offs:**
- Memcached has no persistence — data is lost on eviction or failure
- No authentication — access control relies on network isolation
- No snapshots — volatile cache only

**When to upgrade:** Add VPC placement and security groups for production. Consider
Redis if you need persistence or authentication.
