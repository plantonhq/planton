---
title: "HA with Read Replicas"
description: "This preset provisions a Memorystore for Redis instance with STANDARD_HA tier, three read replicas, authentication, TLS, RDB persistence, and customer-managed encryption. It is designed for..."
type: "preset"
rank: "03"
presetSlug: "03-ha-read-replicas"
componentSlug: "redis-instance"
componentTitle: "Redis Instance"
provider: "gcp"
icon: "package"
order: 3
---

# HA with Read Replicas

This preset provisions a Memorystore for Redis instance with STANDARD_HA tier, three read replicas, authentication, TLS, RDB persistence, and customer-managed encryption. It is designed for read-heavy workloads that benefit from scaling read throughput across multiple replicas.

## When to Use

- Read-heavy caching workloads (e.g., leaderboards, counters, dashboards)
- Applications that separate read and write traffic via read endpoints
- Workloads requiring CMEK for compliance or governance
- Production environments needing higher read throughput than a single primary can provide

## Key Configuration

- **STANDARD_HA tier** — primary plus replica with automatic failover
- **10 GB memory** — larger capacity for read-scaled workloads
- **readReplicasMode: READ_REPLICAS_ENABLED** — exposes a read endpoint
- **replicaCount: 3** — three read replicas (1–5 supported)
- **authEnabled** — Redis AUTH required; use `auth_string` from outputs
- **transitEncryptionMode: SERVER_AUTHENTICATION** — TLS for client connections
- **persistenceConfig** — RDB snapshots every 6 hours
- **customerManagedKey** — CMEK for encryption at rest; use full KMS key resource name or `valueFrom` to reference a GcpKmsKey

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<gcp-project-id>` | GCP project ID where the instance will be created | GCP Console or `GcpProject` outputs |
| `<instance-name>` | Name for this Redis instance (2-40 chars, lowercase, hyphens) | Choose a descriptive name (e.g., `prod-cache-read-scale`) |
| `<gcp-region>` | GCP region for the instance (e.g., `us-central1`) | [GCP regions](https://cloud.google.com/about/locations) |
| `<vpc-network-self-link>` | Full self-link of the VPC network | `GcpVpc` status outputs or GCP Console |
| `<kms-key-resource-name>` | Full KMS key resource name (e.g., `projects/my-project/locations/us-central1/keyRings/redis-keys/cryptoKeys/redis-cmek`) or reference via `valueFrom` | `GcpKmsKey` status outputs or GCP Console |

## Related Presets

- **01-basic-cache** — Minimal BASIC tier for dev/test
- **02-ha-production** — STANDARD_HA without read replicas; simpler production setup
