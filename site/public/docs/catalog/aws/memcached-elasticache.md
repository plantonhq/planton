---
title: Memcached ElastiCache
description: Provision and manage AWS ElastiCache Memcached clusters for high-performance distributed caching.
---

# AWS Memcached ElastiCache

Deploy fully managed Memcached clusters on AWS ElastiCache for high-throughput, low-latency distributed caching.

## Overview

AwsMemcachedElasticache provisions ElastiCache clusters running the Memcached engine. Memcached is a simple, multi-threaded, in-memory key-value store optimized for caching workloads that require sub-millisecond response times and horizontal scalability across 1–40 nodes.

## Key Features

- **Horizontal scaling** — distribute cache across 1–40 nodes with consistent hashing
- **Cross-AZ deployment** — spread nodes across Availability Zones for resilience
- **Transit encryption** — TLS support on engine version 1.6.12+
- **Auto-discovery** — clients automatically discover cluster topology via configuration endpoint
- **Custom parameters** — tune engine behavior with managed parameter groups
- **VPC integration** — deploy in private subnets with security group access control
- **SNS notifications** — receive alerts for cluster events

## When to Use

Choose Memcached when you need a simple, high-throughput distributed cache and do not require data persistence, replication, or authentication. Common use cases include session caching, database query result caching, and API response caching.

For use cases requiring persistence, replication, failover, complex data structures, or authentication, use [Redis ElastiCache](/docs/catalog/aws/redis-elasticache) instead.

## Quick Start

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsMemcachedElasticache
metadata:
  name: my-cache
spec:
  engineVersion: "1.6.22"
  nodeType: cache.t3.micro
  numCacheNodes: 1
```

## Available Presets

| Preset | Description |
|--------|-------------|
| `01-single-node-dev` | Single-node development cache |
| `02-multi-node-cross-az` | 3-node cluster across Availability Zones |
| `03-production-encrypted` | Production setup with TLS, custom parameters, maintenance window |

## Related Components

- [Redis ElastiCache](/docs/catalog/aws/redis-elasticache) — Redis/Valkey caching with replication and persistence
- [VPC](/docs/catalog/aws/vpc) — Network isolation for cache deployments
- [Security Group](/docs/catalog/aws/security-group) — Access control for Memcached endpoints
- [SNS Topic](/docs/catalog/aws/sns-topic) — Event notifications for cluster operations
