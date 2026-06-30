# GcpAlloydbCluster — Research & Design Documentation

Comprehensive research document covering AlloyDB deployment, architecture, networking, backup strategies, CMEK encryption, machine configuration, best practices, and design decisions.

---

## 1. AlloyDB Deployment Landscape

### 1.1 Comparison with Cloud SQL and Spanner

| Aspect | AlloyDB | Cloud SQL for PostgreSQL | Cloud Spanner |
|--------|---------|--------------------------|---------------|
| **API** | PostgreSQL wire protocol | PostgreSQL wire protocol | Spanner SQL (GoogleSQL or PostgreSQL) |
| **Consistency** | Strong (PostgreSQL) | Strong (PostgreSQL) | Strong (distributed) |
| **Scale** | Vertical (instance size) + read pools | Vertical (instance size) + read replicas | Horizontal (sharding) |
| **Use case** | OLTP, analytics, mixed workloads | OLTP, simple workloads | Global, horizontally scale-out |
| **Storage** | Colossus (compute-storage separation) | Attached disk | Distributed |
| **Latency** | Sub-ms for reads | ms-level | ms-level (global) |
| **Networking** | Private Service Access (VPC peering) | Private IP, PSC | Private Service Access |

### 1.2 When to Choose AlloyDB

- PostgreSQL compatibility required
- High throughput and low latency for transactional workloads
- Need for read pools (independent scaling) or columnar engine
- Preference for compute-storage separation over attached-disk model

### 1.3 When to Choose Cloud SQL

- Simpler workloads, smaller scale
- Cost sensitivity over raw performance
- Preference for shared-core or smaller instances

### 1.4 When to Choose Spanner

- Global distribution, horizontal scale-out
- Multi-region writes
- Schema flexibility (e.g., interleaved tables)

---

## 2. Architecture

### 2.1 Cluster + Instance Model

```
AlloyDB Cluster (logical container)
├── Backup policy
├── Continuous backup config
├── Encryption (KMS)
├── Network config (VPC + allocated IP range)
├── Maintenance window
└── Primary Instance (compute node)
    ├── Machine config (cpu_count or machine_type)
    ├── Availability (ZONAL / REGIONAL)
    ├── Database flags
    ├── Query insights
    └── SSL / Auth Proxy config
```

A cluster without a primary instance cannot serve queries. This is why the Planton component bundles cluster and primary instance together. Read pool instances are separate resources with independent scaling lifecycles and are not included in this component.

### 2.2 Storage Engine: Colossus

AlloyDB uses Google's Colossus distributed storage system. Compute and storage are separated:

- **Compute** — Primary instance (and optional read pools) run PostgreSQL-compatible engine
- **Storage** — Colossus provides durable, replicated storage with automatic scaling
- **Benefit** — Storage can grow independently; compute can be scaled without storage migration

This separation enables faster storage I/O and eliminates the need for storage migration when resizing instances.

### 2.3 Compute Separation

The primary instance runs the PostgreSQL engine in memory and communicates with Colossus for persistent storage. WAL (write-ahead log) streaming to Colossus enables continuous backup and point-in-time recovery without impacting primary performance.

---

## 3. Networking

### 3.1 VPC Peering via Private Service Access

AlloyDB uses Private Service Access for connectivity:

1. **VPC peering** — A private connection is established between the consumer VPC and the AlloyDB service producer network
2. **Allocated IP range** — The cluster uses an IP range from the consumer VPC (or a pre-allocated range via `allocated_ip_range`)
3. **Private IP** — The primary instance receives a private IP in the consumer VPC; applications connect via this IP on port 5432

**Prerequisites:**

- The VPC must have Private Service Access configured (typically via `private_services_access` on a GcpVpc component)
- The allocated IP range must not overlap with existing subnets

### 3.2 Why PSC Is Deferred

Private Service Connect (PSC) is an alternative connectivity model used by newer GCP services (e.g., Memorystore new-gen). AlloyDB currently uses Private Service Access. PSC support for AlloyDB may be added in future GCP releases; we defer PSC support until the AlloyDB API supports it.

---

## 4. Backup Strategies

### 4.1 Automated Backups vs. Continuous Backup

| Feature | Automated | Continuous |
|---------|-----------|------------|
| **Mechanism** | Periodic snapshots | WAL streaming |
| **Recovery** | Restore to point of backup | Point-in-time recovery (PITR) |
| **Retention** | Quantity-based or time-based | Recovery window (1–35 days) |
| **Schedule** | Optional weekly schedule | Always-on (when enabled) |
| **CMEK** | Separate key per backup policy | Separate key per config |

Both can be enabled simultaneously. Automated backups provide full restores; continuous backup enables recovery to any second within the recovery window.

### 4.2 Automated Backup Retention

- **Quantity-based** — Keep N backups (e.g., 7, 14)
- **Time-based** — Keep backups for N seconds (e.g., `1209600s` = 14 days)
- Mutually exclusive: only one retention policy per automated backup policy

### 4.3 Weekly Schedule

- `daysOfWeek`: MONDAY, TUESDAY, WEDNESDAY, THURSDAY, FRIDAY, SATURDAY, SUNDAY
- `startHour`: 0–23 (UTC)
- If not specified, GCP defaults to daily backups

### 4.4 Backup Window Format

Duration fields use seconds with an `s` suffix:

- `backup_window`: e.g., `"3600s"` (1 hour)
- `time_based_retention_period`: e.g., `"1209600s"` (14 days)

Invalid formats (e.g., `"1h"`, `"14d"`) are rejected by validation.

---

## 5. CMEK Encryption

AlloyDB supports customer-managed encryption keys (CMEK) at three levels:

| Level | Field | Purpose |
|-------|-------|---------|
| Cluster | `kms_key_name` | Encrypt cluster data at rest |
| Automated backup | `automated_backup_policy.encryption_kms_key_name` | Encrypt backup snapshots |
| Continuous backup | `continuous_backup_config.encryption_kms_key_name` | Encrypt PITR data |

Each can use a different key, enabling independent key lifecycle management (e.g., separate keys for data vs. backups for compliance).

**Requirements:**

- KMS key must exist in the same location as the cluster
- AlloyDB service account must have `cloudkms.cryptoKeyEncrypterDecrypter` on the key
- Encryption config is immutable after creation

---

## 6. Machine Configuration

### 6.1 cpu_count vs. machine_type

| Option | Use Case | Notes |
|--------|----------|-------|
| `cpu_count` | Simple sizing; let GCP choose machine family | Valid: 2, 4, 8, 16, 32, 64, 96, 128 |
| `machine_type` | Explicit machine family (e.g., n2-highmem-4, c4a-highmem-4-lssd) | Advanced tuning; specific SKU |

Only one of `cpu_count` or `machine_type` may be set. `cpu_count` is recommended for most workloads.

### 6.2 Availability Types

| Type | Behavior | Use Case |
|------|----------|----------|
| ZONAL | Single-zone deployment | Dev/test; lower cost; single zone of failure |
| REGIONAL | Multi-zone deployment with automatic failover | Production; recommended for HA |

---

## 7. Best Practices for Production Deployments

1. **Use REGIONAL availability** — Multi-zone for automatic failover; ZONAL only for dev/test
2. **Enable deletion protection** — `deletion_protection: true` by default; keep it for production
3. **CMEK for compliance** — Use `kms_key_name` when HIPAA, PCI-DSS, or FedRAMP requires customer-managed keys
4. **Configure maintenance window** — Avoid peak hours; use `maintenance_window` with day and start hour
5. **Use SSL** — `sslMode: ENCRYPTED_ONLY` for production
6. **Require connectors** — `requireConnectors: true` enforces AlloyDB Auth Proxy or Language Connectors for IAM-based auth
7. **Continuous backup** — Enable with appropriate `recovery_window_days` (e.g., 14–21) for PITR
8. **Query insights** — Enable for performance tuning; set `query_plans_per_minute` and `query_string_length` as needed
9. **Initial user** — Use for bootstrap; otherwise configure access via Auth Proxy + IAM
10. **allocated_ip_range** — Use when IP ranges are pre-planned in enterprise networks

---

## 8. 80/20 Analysis

### 8.1 What We Cover

| Feature | Included | Rationale |
|---------|----------|-----------|
| Cluster + primary instance | Yes | Core lifecycle; no cluster without primary |
| CMEK (cluster, backup, continuous) | Yes | Enterprise compliance |
| Automated backup (quantity/time, weekly) | Yes | Common backup patterns |
| Continuous backup (PITR) | Yes | Disaster recovery essential |
| cpu_count / machine_type | Yes | Core sizing options |
| ZONAL / REGIONAL | Yes | HA choice |
| Query insights | Yes | Performance monitoring |
| SSL mode | Yes | Security |
| require_connectors (Auth Proxy) | Yes | IAM-based auth enforcement |
| Maintenance window | Yes | Operational control |
| Initial user | Yes | Bootstrap superuser |
| allocated_ip_range | Yes | Enterprise IP planning |

### 8.2 What We Exclude

| Feature | Excluded | Rationale |
|---------|----------|-----------|
| Read pool instances | Yes | Separate lifecycle; scale independently |
| IAM bindings | Yes | Managed via GCP IAM, not database config |
| Database flags (advanced) | Partially | Supported via `database_flags`; full list is GCP-documented |
| PSC networking | Yes | AlloyDB uses Private Service Access; PSC not yet supported |
| Labels | Yes | GCP AlloyDB supports labels; can add if needed |
| Backup location | Partially | `automated_backup_policy.location` supported for backup region |
| Read replica instances | Yes | Separate component; different scaling model |

### 8.3 Deliberate Design Choices

**Bundled primary instance:** A cluster without a primary cannot serve traffic. Bundling ensures a usable deployment in one component. Read pools are excluded because they have independent scaling and lifecycle.

**Three CMEK levels:** Cluster, automated backup, and continuous backup each support their own key. This matches GCP's API and allows compliance scenarios where separate keys are required.

**No PSC:** AlloyDB uses Private Service Access. PSC support for AlloyDB is deferred until it is available in the GCP API.

---

## 9. Immutable Fields

The following fields cannot be changed after cluster creation; changing them requires recreating the cluster:

| Field | Notes |
|-------|-------|
| `cluster_name` | GCP resource identifier |
| `location` | Region placement |
| `network` | VPC connectivity |
| `kms_key_name` | Encryption at rest |
| `primary_instance.instance_id` | Instance ID |

---

## 10. Database Version Support

| Version | Status | Notes |
|---------|--------|-------|
| POSTGRES_14 | Supported | Legacy; consider upgrading |
| POSTGRES_15 | Supported | Stable |
| POSTGRES_16 | Supported | Latest |

If `database_version` is omitted, GCP selects the latest stable version. Version upgrades are managed separately from cluster creation.

---

## 11. Connection Patterns

### 11.1 Direct IP (Private)

Applications in the same VPC (or connected via VPC peering) connect to `primary_instance_ip` on port 5432. Use `initial_user` or configure PostgreSQL users manually.

### 11.2 AlloyDB Auth Proxy

AlloyDB Auth Proxy provides IAM-based authentication without storing passwords. When `require_connectors: true` is set, direct IP connections are rejected; all access must go through the Auth Proxy or AlloyDB Language Connectors.

### 11.3 AlloyDB Language Connectors

AlloyDB offers connectors for Java, Python, Go, Node.js, and others that integrate with IAM. These work with `require_connectors: true` and avoid the need for a separate Auth Proxy process.

---

## 12. Maintenance Window

- **Day**: MONDAY, TUESDAY, WEDNESDAY, THURSDAY, FRIDAY, SATURDAY, SUNDAY
- **Start hour**: 0–23 (UTC)

GCP applies system updates and patches during this window. Maintenance is typically scheduled weekly; avoid peak hours for your workload.

---

## 13. Query Insights Configuration

| Field | Range | Default | Notes |
|-------|-------|---------|-------|
| `query_plans_per_minute` | 0–20 | 5 | Set to 0 to disable plan capture |
| `query_string_length` | 256–4500 | 1024 | Longer strings for complex queries |
| `record_application_tags` | bool | — | Tag queries by application |
| `record_client_address` | bool | — | Record client IP per query |

Query insights help diagnose slow queries and performance issues. Enable in production.

---

## 14. SSL Mode

| Value | Behavior |
|-------|----------|
| `ENCRYPTED_ONLY` | All connections must use TLS (recommended for production) |
| `ALLOW_UNENCRYPTED_AND_ENCRYPTED` | Both TLS and plaintext allowed (dev only) |

---

## 15. Deletion Protection

`deletion_protection` defaults to `true`. Before destroying a cluster:

1. Set `deletion_protection: false` in the spec
2. Apply the change
3. Then destroy the cluster

This prevents accidental deletion of production data.

---

## 16. Infra Chart Composition Patterns

### Pattern 1: VPC + Cluster (Minimal)

```
GcpVpc (with private_services_access)
└── GcpAlloydbCluster (references network via valueFrom)
```

### Pattern 2: VPC + KMS + Cluster (Enterprise)

```
GcpVpc (with private_services_access)
├── GcpKmsKeyRing
│   └── GcpKmsKey (cluster key)
│   └── GcpKmsKey (backup key)
│   └── GcpKmsKey (PITR key)
└── GcpAlloydbCluster (references network via valueFrom, kms_key_name via valueFrom)
```

### Pattern 3: Multi-Cluster (Multi-Tenant)

```
GcpVpc (with private_services_access)
├── GcpAlloydbCluster "tenant-a"
├── GcpAlloydbCluster "tenant-b"
└── GcpAlloydbCluster "tenant-c"
```

---

## 17. References

- [AlloyDB Documentation](https://cloud.google.com/alloydb/docs)
- [AlloyDB REST API](https://cloud.google.com/alloydb/docs/reference/rest)
- [Private Service Access](https://cloud.google.com/vpc/docs/configure-private-service-access)
- [AlloyDB Auth Proxy](https://cloud.google.com/alloydb/docs/auth-proxy/overview)
