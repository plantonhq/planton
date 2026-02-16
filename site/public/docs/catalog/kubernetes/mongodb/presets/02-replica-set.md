---
title: "MongoDB Replica Set"
description: "This preset deploys a 3-node MongoDB replica set with persistence. Provides automatic failover and read scaling for production workloads."
type: "preset"
rank: "02"
presetSlug: "02-replica-set"
componentSlug: "mongodb"
componentTitle: "MongoDB"
provider: "kubernetes"
icon: "package"
order: 2
---

# MongoDB Replica Set

This preset deploys a 3-node MongoDB replica set with persistence. Provides automatic failover and read scaling for production workloads.

## When to Use

- Production MongoDB databases requiring high availability
- Applications needing automatic failover and read replica support
- Workloads where data durability across node failures is critical

## Key Configuration Choices

- **3 replicas** -- one primary with two secondaries; provides automatic failover via MongoDB's Raft-based election
- **20Gi disk per replica** -- production-appropriate; increase based on data growth
- **Higher resources** -- production-grade memory for WiredTiger cache

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<your-namespace>` | Target namespace | Your namespace management or `KubernetesNamespace` resource |

## Related Presets

- **01-single-instance** -- Minimal standalone MongoDB for development
