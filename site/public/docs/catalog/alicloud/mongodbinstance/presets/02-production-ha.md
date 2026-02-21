---
title: "Preset: Production HA MongoDB Instance"
description: "A production MongoDB replica set deployed across three availability zones with read replicas, backup policies, and operational safeguards."
type: "preset"
rank: "02"
presetSlug: "02-production-ha"
componentSlug: "mongodbinstance"
componentTitle: "MongodbInstance"
provider: "alicloud"
icon: "package"
order: 2
---

# Preset: Production HA MongoDB Instance

A production MongoDB replica set deployed across three availability zones with read replicas, backup policies, and operational safeguards.

## Use Case

- Production workloads requiring high availability
- Read-heavy applications benefiting from read replicas
- Multi-zone fault tolerance for data durability

## Configuration

- **Engine**: MongoDB 6.0 with WiredTiger
- **Instance Class**: `mongo.x8.large` (production tier)
- **Storage**: 200 GB on cloud ESSD PL2 with 3000 provisioned IOPS
- **Replication**: 5-node replica set with 2 read-only replicas
- **HA**: Three-zone deployment (primary, secondary, hidden in separate AZs)
- **Backup**: Mon/Wed/Fri at 03:00-04:00 UTC
- **Maintenance**: 02:00-06:00 UTC window
- **Protection**: Release protection enabled
- **Monitoring**: Slow query profiling at 200ms threshold

## What's Included

- Cross-AZ automatic failover
- Read scaling via 2 dedicated read replicas
- Scheduled backups with configurable retention
- IP whitelist restricted to private subnets
- Slow query profiling for performance monitoring
