# GCP Redis Instance (Memorystore for Redis)

Deploys a Google Cloud Memorystore for Redis instance via `google_redis_instance` (Terraform) or Pulumi `redis.Instance`. Memorystore for Redis is a fully managed, in-memory data store backed by the Redis protocol, suitable for caching, session management, real-time analytics, rate limiting, and pub/sub messaging.

## Overview

Memorystore for Redis provides a managed Redis service on GCP with automatic patching, monitoring, and high availability options. It eliminates the operational burden of running Redis yourself while delivering sub-millisecond latency for in-memory workloads. The component provisions a Redis instance in your chosen region and VPC, with configurable tiers, memory sizes, and security controls.

## Purpose

This component exists to give platform engineers a declarative, infrastructure-as-code interface for provisioning Redis on GCP. It abstracts the underlying Terraform/Pulumi resources behind a consistent spec, supports cross-resource references (project, VPC, KMS key), and exports connection details and secrets as stack outputs for downstream consumers.

## Key Features

- **Tier selection**: BASIC (standalone, no SLA) or STANDARD_HA (primary + replica, 99.9% SLA)
- **Memory sizing**: Configurable from 1 GiB upward
- **Redis AUTH**: Optional AUTH string for client authentication (GCP-managed, auto-rotated)
- **TLS in transit**: Optional `SERVER_AUTHENTICATION` for encrypted client connections
- **Read replicas**: Scale read throughput with 1–5 replicas (STANDARD_HA only)
- **RDB persistence**: Optional periodic snapshots for durability
- **Maintenance windows**: Schedule weekly maintenance (day + hour UTC)
- **CMEK**: Customer-managed encryption keys for data at rest
- **Deletion protection**: Prevent accidental destruction of production instances
- **VPC integration**: Connect via DIRECT_PEERING or PRIVATE_SERVICE_ACCESS

## Use Cases

- **Application caching**: Offload database reads, reduce latency for frequently accessed data
- **Session storage**: Store user sessions for stateless web applications
- **Rate limiting**: Track request counts and enforce limits
- **Real-time analytics**: Leaderboards, counters, and live dashboards
- **Pub/sub messaging**: Decouple services with Redis pub/sub
- **Development and testing**: Quick spin-up of Redis for local or CI environments

## Architecture

When you deploy a GcpRedisInstance, OpenMCF provisions:

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
| **Persistence** | `persistence_config` with `RDB` mode and `rdb_snapshot_period` (ONE_HOUR, SIX_HOURS, TWELVE_HOURS, TWENTY_FOUR_HOURS) |
| **Read replicas** | `read_replicas_mode: READ_REPLICAS_ENABLED`, `replica_count` 1–5 (STANDARD_HA only) |
| **Maintenance** | `maintenance_window.day` (MONDAY–SUNDAY), `maintenance_window.hour` (0–23 UTC) |
| **CMEK** | `customer_managed_key` — full KMS key resource name or reference to GcpKmsKey |
| **Networking** | `authorized_network`, `connect_mode`, `reserved_ip_range` |
| **Other** | `redis_version`, `display_name`, `location_id`, `redis_configs`, `deletion_protection` |

**Immutable fields** (require instance replacement if changed): `instance_name`, `tier`, `connect_mode`, `transit_encryption_mode`, `authorized_network`, `reserved_ip_range`, `customer_managed_key`.

## Security

- **Encryption at rest**: Google-managed keys by default; use `customer_managed_key` for CMEK
- **Encryption in transit**: Enable `transit_encryption_mode: SERVER_AUTHENTICATION` for TLS
- **AUTH**: Enable `auth_enabled` and use the `auth_string` output — treat as a secret
- **Network isolation**: Attach to a private VPC via `authorized_network`; avoid public exposure
- **Deletion protection**: Enable for production to prevent accidental deletion

## Stack Outputs

| Output | Type | Description |
|--------|------|-------------|
| `host` | string | Primary Redis endpoint hostname |
| `port` | int32 | Primary port (typically 6379) |
| `read_endpoint` | string | Read replica hostname (STANDARD_HA + read replicas only) |
| `read_endpoint_port` | int32 | Read replica port |
| `current_location_id` | string | Zone where the primary is running |
| `auth_string` | string | Redis AUTH string (when `auth_enabled` is true) |

## Future Enhancements

- Support for Redis Cluster mode (sharding)
- Automated backup and restore workflows
- Integration with Secret Manager for AUTH string injection
- Metrics and alerting presets
- Cross-region replication

## Related Components

- **GcpProject** — provides the GCP project
- **GcpVpc** — provides the VPC network for `authorized_network`
- **GcpGlobalAddress** — reserve a /20 range for VPC peering with managed services
- **GcpKmsKey** — provides a CMEK key for `customer_managed_key`

## Additional Resources

- [Memorystore for Redis Documentation](https://cloud.google.com/memorystore/docs/redis)
- [Redis Instance REST API](https://cloud.google.com/memorystore/docs/redis/reference/rest/v1/projects.locations.instances)
