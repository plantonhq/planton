# GCP Redis Instance (Memorystore for Redis)

Deploys a Google Cloud Memorystore for Redis instance via `google_redis_instance` (Terraform) or Pulumi `redis.Instance`. Memorystore for Redis is a fully managed, in-memory data store backed by the Redis protocol, suitable for caching, session management, real-time analytics, rate limiting, and pub/sub messaging.

## Overview

Memorystore for Redis provides a managed Redis service on GCP with automatic patching, monitoring, and high availability options. It eliminates the operational burden of running Redis yourself while delivering sub-millisecond latency for in-memory workloads. The component provisions a Redis instance in your chosen region and VPC, with configurable tiers, memory sizes, and security controls.

## Purpose

This component exists to give platform engineers a declarative, infrastructure-as-code interface for provisioning Redis on GCP. It abstracts the underlying Terraform/Pulumi resources behind a consistent spec, supports cross-resource references (project, VPC, KMS key), and exports connection details and secrets as outputs for downstream consumers.

## Key Features

- **Tier selection**: BASIC (standalone, no SLA) or STANDARD_HA (primary + replica, 99.9% SLA)
- **Memory sizing**: Configurable from 1 GiB upward
- **Redis AUTH**: Optional AUTH string for client authentication (GCP-managed, auto-rotated)
- **TLS in transit**: Optional `SERVER_AUTHENTICATION` for encrypted client connections
- **Read replicas**: Scale read throughput with 1–5 replicas (STANDARD_HA only), with `secondary_ip_range` to grow an existing instance's address space in place
- **Zone pinning**: `location_id` for the primary and `alternative_location_id` for the STANDARD_HA replica — co-locate the cache with zonal workloads
- **RDB persistence**: Optional periodic snapshots, with a schedule anchor (`rdb_snapshot_start_time`) to place snapshot I/O in a low-traffic window
- **Maintenance control**: Weekly window to the exact minute (day + hour + minute UTC), plus `maintenance_version` for applying a security patch on your schedule
- **CMEK**: Customer-managed encryption keys for data at rest
- **VPC integration**: Connect via DIRECT_PEERING or PRIVATE_SERVICE_ACCESS (composing GcpGlobalAddress + GcpServiceNetworkingConnection)

## Use Cases

- **Application caching**: Offload database reads, reduce latency for frequently accessed data
- **Session storage**: Store user sessions for stateless web applications
- **Rate limiting**: Track request counts and enforce limits
- **Real-time analytics**: Leaderboards, counters, and live dashboards
- **Pub/sub messaging**: Decouple services with Redis pub/sub
- **Development and testing**: Quick spin-up of Redis for local or CI environments

## Architecture

When you deploy a GcpRedisInstance, Planton provisions:

- **Redis instance**: A `google_redis_instance` resource in the specified project and region
- **Primary endpoint**: Host and port (typically 6379) for read/write traffic
- **Read endpoint** (STANDARD_HA + read replicas): Separate host/port for read-only traffic
- **VPC connectivity**: Instance attached to the specified `authorized_network` via peering or Private Service Access

For STANDARD_HA, GCP automatically places the primary and replica in different zones within the region. Failover is automatic.

## Configuration Options

| Category | Options |
|----------|---------|
| **Tier** | `BASIC` (single node) or `STANDARD_HA` (primary + replica) |
| **Memory** | `memory_size_gb` (min 1) |
| **Auth** | `auth_enabled: true` — AUTH string exported in outputs |
| **TLS** | `transit_encryption_mode: SERVER_AUTHENTICATION` or `DISABLED` |
| **Persistence** | `persistence_config` with `RDB` mode, `rdb_snapshot_period` (ONE_HOUR, SIX_HOURS, TWELVE_HOURS, TWENTY_FOUR_HOURS), and `rdb_snapshot_start_time` (RFC3339 schedule anchor) |
| **Read replicas** | `read_replicas_mode: READ_REPLICAS_ENABLED`, `replica_count` 1–5 (STANDARD_HA only); `secondary_ip_range` when enabling replicas on an existing instance |
| **Maintenance** | `maintenance_window.day` (MONDAY–SUNDAY), `.hour` (0–23) + `.minute` (0–59) UTC, `.description` (what the window is for, ≤512 chars); `maintenance_version` for self-service patching |
| **CMEK** | `customer_managed_key` — full KMS key resource name or reference to GcpKmsKey |
| **Networking** | `authorized_network`, `connect_mode`, `reserved_ip_range`, `secondary_ip_range` |
| **Placement** | `location_id` (primary zone), `alternative_location_id` (STANDARD_HA replica zone) |
| **Labels** | `labels` — user labels for cost attribution and fleet queries, merged beneath the platform attribution labels |
| **Lifecycle** | `deletion_protection` (default `true` — destroy fails until explicitly set `false`); `deletion_policy` — `DELETE` (default), `PREVENT`, or `ABANDON` |
| **Other** | `redis_version`, `display_name`, `redis_configs` |

**Immutable fields** (require instance replacement if changed): `instance_name`, `tier`, `connect_mode`, `transit_encryption_mode`, `authorized_network`, `reserved_ip_range`, `location_id`, `alternative_location_id`, `customer_managed_key`, `region`.

## Security

- **Encryption at rest**: Google-managed keys by default; use `customer_managed_key` for CMEK
- **Encryption in transit**: Enable `transit_encryption_mode: SERVER_AUTHENTICATION` for TLS — clients install the `server_ca_certs` output as trust anchors
- **AUTH**: Enable `auth_enabled` and use the `auth_string` output — treat as a secret
- **Network isolation**: Attach to a private VPC via `authorized_network`; avoid public exposure

## Outputs

| Output | Type | Description |
|--------|------|-------------|
| `host` | string | Primary Redis endpoint hostname |
| `port` | int32 | Primary port (typically 6379) |
| `read_endpoint` | string | Read replica hostname (STANDARD_HA + read replicas only) |
| `read_endpoint_port` | int32 | Read replica port |
| `current_location_id` | string | Zone where the primary is running |
| `auth_string` | string | Redis AUTH string (when `auth_enabled` is true) |
| `server_ca_certs` | string[] | PEM CA certificates clients trust for TLS (when transit encryption is on) |
| `persistence_iam_identity` | string | Service identity (`serviceAccount:<email>`) to grant on GCS buckets for RDB import/export |
| `effective_reserved_ip_range` | string | The CIDR actually in use — explicit or GCP-selected |
| `instance_name` | string | Instance name in GCP — the handle API callers address the instance by |
| `region` | string | Plain region name hosting the instance |

## Deliberately not modeled (recorded reasons)

- **Maintenance window `seconds`/`nanos`** — sub-minute precision on a fixed
  1-hour maintenance window is meaningless; the spec models the start to the
  minute (recorded in `iac/provider-parity.yaml`).
- **Redis Cluster mode (sharded, PSC-based)** — a structurally different
  resource (`google_redis_cluster`); a separate kind if demand appears.
  Valkey-based clusters are covered by GcpMemorystoreInstance.

Every other configurable argument of `google_redis_instance` at the pinned
provider version is representable through this spec; the recorded judgment
lives in `iac/provider-parity.yaml`, checked by
`planton provider-parity --check`.

## Related Kinds

- **GcpProject** — provides the GCP project
- **GcpVpcNetwork** — provides the VPC network for `authorized_network`
- **GcpGlobalAddress** — reserves the VPC_PEERING range named by `reserved_ip_range` in PRIVATE_SERVICE_ACCESS mode
- **GcpServiceNetworkingConnection** — the private services access peering PRIVATE_SERVICE_ACCESS instances ride on
- **GcpKmsKey** — provides a CMEK key for `customer_managed_key`

## Additional Resources

- [Memorystore for Redis Documentation](https://cloud.google.com/memorystore/docs/redis)
- [Redis Instance REST API](https://cloud.google.com/memorystore/docs/redis/reference/rest/v1/projects.locations.instances)

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
