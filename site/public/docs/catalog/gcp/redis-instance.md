---
title: "Redis Instance"
description: "Memorystore for Redis deployment documentation"
icon: "package"
order: 100
componentName: "gcpredisinstance"
---

# Redis Instance

Deploys a Google Cloud Memorystore for Redis instance — a fully managed, in-memory data store backed by the Redis protocol. Suitable for caching, session management, real-time analytics, rate limiting, and pub/sub messaging.

## What Gets Created

When you deploy a GcpRedisInstance resource, OpenMCF provisions:

- **Redis instance** — a `google_redis_instance` in the specified project and region
- **Primary endpoint** — host and port (typically 6379) for read/write traffic
- **Read endpoint** (STANDARD_HA + read replicas) — separate host/port for read-only traffic
- **VPC connectivity** — instance attached to the specified `authorizedNetwork` via peering or Private Service Access

## Quick Start

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpRedisInstance
metadata:
  name: my-cache
spec:
  projectId:
    value: my-gcp-project-123
  instanceName: my-cache
  region: us-central1
  tier: BASIC
  memorySizeGb: 1
```

Deploy:

```shell
openmcf apply -f redis.yaml
```

## Key Features

- **Tier selection** — BASIC (standalone, no SLA) or STANDARD_HA (primary + replica, 99.9% SLA)
- **Memory sizing** — configurable from 1 GiB upward
- **Redis AUTH** — optional AUTH string for client authentication
- **TLS in transit** — optional `SERVER_AUTHENTICATION` for encrypted connections
- **Read replicas** — scale read throughput with 1–5 replicas (STANDARD_HA only)
- **RDB persistence** — optional periodic snapshots (ONE_HOUR, SIX_HOURS, TWELVE_HOURS, TWENTY_FOUR_HOURS)
- **Maintenance windows** — schedule weekly maintenance (day + hour UTC)
- **CMEK** — customer-managed encryption keys for data at rest
- **Deletion protection** — prevent accidental destruction of production instances

## Configuration Highlights

| Field | Description |
|-------|-------------|
| `tier` | `BASIC` or `STANDARD_HA` |
| `memorySizeGb` | Memory in GiB (min 1) |
| `authEnabled` | Enable Redis AUTH; AUTH string exported in outputs |
| `transitEncryptionMode` | `DISABLED` or `SERVER_AUTHENTICATION` |
| `readReplicasMode` | `READ_REPLICAS_DISABLED` or `READ_REPLICAS_ENABLED` (STANDARD_HA only) |
| `persistenceConfig` | RDB snapshots with configurable period |
| `maintenanceWindow` | Day and hour (UTC) for weekly maintenance |
| `deletionProtection` | Prevent Terraform/Pulumi from destroying the instance |

## Presets

- **[01-basic-cache](../../../../../apis/org/openmcf/provider/gcp/gcpredisinstance/v1/presets/01-basic-cache.yaml)** — BASIC tier, 1 GB, minimal config for dev/test
- **[02-ha-production](../../../../../apis/org/openmcf/provider/gcp/gcpredisinstance/v1/presets/02-ha-production.yaml)** — STANDARD_HA, auth, TLS, persistence, maintenance window, deletion protection
- **[03-ha-read-replicas](../../../../../apis/org/openmcf/provider/gcp/gcpredisinstance/v1/presets/03-ha-read-replicas.yaml)** — STANDARD_HA with read replicas, CMEK, persistence

## Related Components

- [GcpProject](/docs/catalog/gcp/project) — provides the GCP project
- [GcpVpc](/docs/catalog/gcp/vpc) — provides the VPC network for `authorizedNetwork`
- [GcpGlobalAddress](/docs/catalog/gcp/global-address) — reserve a /20 range for VPC peering with managed services
- [GcpKmsKey](/docs/catalog/gcp/kms-key) — provides a CMEK key for `customerManagedKey`
