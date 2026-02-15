---
title: "Memcached Multi-Node Cross-AZ"
description: "This preset creates a 3-node Memcached cluster distributed across Availability Zones for high availability. If one AZ experiences an outage, two-thirds of the cache remains available."
type: "preset"
rank: "02"
presetSlug: "02-multi-node-cross-az"
componentSlug: "memcached-elasticache"
componentTitle: "Memcached ElastiCache"
provider: "aws"
icon: "package"
order: 2
---

# Memcached Multi-Node Cross-AZ

This preset creates a 3-node Memcached cluster distributed across Availability Zones for high availability. If one AZ experiences an outage, two-thirds of the cache remains available.

## When to Use

- Staging and production environments requiring cache resilience
- Applications where partial cache loss is acceptable but total loss is not
- Multi-AZ VPC deployments

## Key Configuration Choices

- **3 nodes** (`numCacheNodes: 3`) — distributes keys across three nodes for ~3x cache capacity
- **cross-az** — nodes placed in different AZs for AZ-failure resilience
- **cache.t3.medium** — moderate instance size for production-like workloads
- **Explicit AZ placement** — nodes pinned to `us-east-1a`, `us-east-1b`, `us-east-1c`

## Placeholders to Replace

- `metadata.name` — your cache name
- `preferredAvailabilityZones` — adjust to your region's AZs

## Common Additions

- Add `subnetIds` and `securityGroupIds` for VPC-based deployments
- Enable `transitEncryptionEnabled: true` for TLS encryption
- Add `notificationTopicArn` for cluster event alerts
- Increase `numCacheNodes` for higher capacity (up to 40)

## Related Presets

- **01-single-node-dev** — minimal single-node for development
- **03-production-encrypted** — full production setup with encryption and notifications
