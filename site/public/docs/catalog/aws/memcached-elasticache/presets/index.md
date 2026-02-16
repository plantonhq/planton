---
title: "Presets"
description: "Ready-to-deploy configuration presets for Memcached ElastiCache"
type: "preset-list"
componentSlug: "memcached-elasticache"
componentTitle: "Memcached ElastiCache"
provider: "aws"
icon: "package"
order: 200
presets:
  - slug: "01-single-node-dev"
    rank: "01"
    title: "Memcached Single Node (Development)"
    excerpt: "This preset creates a single-node Memcached 1.6.22 cluster. It is the fastest way to get a development or testing cache running."
  - slug: "02-multi-node-cross-az"
    rank: "02"
    title: "Memcached Multi-Node Cross-AZ"
    excerpt: "This preset creates a 3-node Memcached cluster distributed across Availability Zones for high availability. If one AZ experiences an outage, two-thirds of the cache remains available."
  - slug: "03-production-encrypted"
    rank: "03"
    title: "Memcached Production Encrypted"
    excerpt: "This preset creates a production-ready Memcached cluster with TLS encryption, cross-AZ distribution, custom parameters, and a defined maintenance window."
---

# Memcached ElastiCache Presets

Ready-to-deploy configuration presets for Memcached ElastiCache. Each preset is a complete manifest you can copy, customize, and deploy.
