---
title: "Production MongoDB Replica Set"
description: "This preset creates a 3-node Scaleway MongoDB replica set with Private Network connectivity and automated snapshot scheduling. The replica set provides automatic failover -- if the primary node..."
type: "preset"
rank: "02"
presetSlug: "02-production-replica-set"
componentSlug: "mongodb-instance"
componentTitle: "MongoDB Instance"
provider: "scaleway"
icon: "package"
order: 2
---

# Production MongoDB Replica Set

This preset creates a 3-node Scaleway MongoDB replica set with Private Network connectivity and automated snapshot scheduling. The replica set provides automatic failover -- if the primary node fails, one of the secondaries is elected as the new primary within seconds.

## When to Use

- Production applications requiring high availability for document data
- Workloads that need automatic failover and read replicas
- Applications with sensitive data requiring network isolation and regular backups

## Key Configuration Choices

- **MongoDB 7.0** (`version: 7.0.12`) -- latest stable version
- **MGDB-POP2-2C-8G nodes** (`nodeType: MGDB-POP2-2C-8G`) -- 2 vCPU, 8 GB RAM; production-grade nodes for moderate workloads
- **3-node replica set** (`nodeNumber: 3`) -- provides automatic failover and majority write concern
- **Private Network** (`privateNetworkId`) -- database is reachable only via private IPs
- **50 GB volume** (`volumeSizeInGb: 50`) -- starting size; increase as data grows
- **Snapshot schedule** -- snapshots every 12 hours, retained for 14 days, enabling point-in-time recovery
- **SBS 5k storage** (`volumeType: sbs_5k`) -- network-attached with 5,000 IOPS baseline; upgrade to `sbs_15k` for write-intensive workloads

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| ----------- | ----------- | ------------- |
| `<your-private-network-id>` | UUID of the Private Network for database connectivity | Scaleway console or `ScalewayPrivateNetwork` status outputs |
| `<your-admin-user>` | Database admin username (max 63 characters) | Choose a username |
| `<your-admin-password>` | Database admin password (min 8 characters) | Generate a strong password |

## Related Presets

- **01-dev-standalone** -- Use instead for development with a single node and no HA
