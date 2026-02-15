# GcpSpannerDatabase - Research & Design Documentation

## Cloud Spanner Database Deployment Landscape

Cloud Spanner databases sit at the heart of the Spanner ecosystem. While the Spanner instance defines compute capacity and regional placement, the database is where data lives -- tables, indexes, views, and schema definitions are all database-level objects.

### Architecture: Instance vs. Database

```
Project
└── Spanner Instance (compute + region)
    ├── Database A (schema + data)
    │   ├── Tables
    │   ├── Indexes
    │   └── Views
    ├── Database B (schema + data)
    └── Database C (schema + data)
```

A single instance can host many databases, all sharing the instance's compute capacity. Databases have independent schemas, access controls, and lifecycle settings but inherit the instance's regional configuration and capacity allocation.

### Why Split Instance and Database?

The instance-database split reflects how Spanner works in production:

1. **Different lifecycles** -- Instances are long-lived infrastructure; databases may be created/destroyed more frequently (e.g., per-tenant databases, test databases).
2. **Different personas** -- Platform teams manage instances (capacity, regions, encryption at rest); application teams manage databases (schema, retention, DDL).
3. **Different composition patterns** -- An infra chart might create one instance with three databases, or reference an existing instance to add a new database.

## 80/20 Scoping Rationale

### What We Include

The OpenMCF GcpSpannerDatabase component covers the ~80% of Spanner database configuration that matters for infrastructure provisioning:

| Feature | Included | Rationale |
|---|---|---|
| Database creation with dialect | Yes | Core lifecycle event |
| Initial DDL (schema) | Yes | Common pattern for greenfield deployments |
| CMEK encryption | Yes | Compliance requirement for enterprise users |
| Version retention (PITR) | Yes | Disaster recovery essential |
| Drop protection | Yes | Production safety guard |
| Default time zone | Yes | Affects SQL function behavior |

### What We Exclude

| Feature | Excluded | Rationale |
|---|---|---|
| IAM bindings | Yes | Managed separately via GCP IAM, not database configuration |
| Backup schedules | Yes | Managed at instance level via `default_backup_schedule_type` |
| Multi-key CMEK (`kms_key_names`) | Yes | Advanced feature for multi-region; can add in v2 |
| Change streams | Yes | Application-level concern, typically managed by app code |
| Table/index creation | Partially | Supported via `ddl` for initial setup; ongoing management via migration tools |
| Database roles | Yes | Fine-grained access control, managed via DDL or application |

### Deliberate Design Choices

**Single KMS key, not multi-key:** The Terraform provider supports both `kms_key_name` (single key) and `kms_key_names` (multiple keys for multi-region). We chose single-key for v1 because:
- Single-key covers 90%+ of CMEK use cases (regional instances)
- Multi-key adds complexity (one key per region in the instance config)
- Multi-key can be added as a backward-compatible enhancement in v2

**No Terraform `deletion_protection`:** We include GCP's `enable_drop_protection` (API-level) but not Terraform's `deletion_protection` (IaC-tool-level virtual field). Rationale:
- `enable_drop_protection` works across all interfaces (Console, gcloud, API, IaC)
- `deletion_protection` is Terraform/Pulumi-specific, not a GCP concept
- GcpSpannerInstance (the parent component) follows the same pattern
- Users who want IaC-level protection can use Terraform's `lifecycle { prevent_destroy }` or Pulumi's `protect` option

**No labels support:** Spanner databases do not support GCP labels. This is a GCP platform limitation, not a design choice. Labels are available at the instance level (GcpSpannerInstance).

## SQL Dialect Deep Dive

The `database_dialect` choice is permanent and affects the entire database experience:

### Google Standard SQL
- Full Spanner feature set
- Interleaved tables (parent-child co-location for low-latency joins)
- STRUCT and ARRAY types
- Parameterized queries with `@param` syntax
- DDL syntax: `CREATE TABLE`, `ALTER TABLE`, etc.

### PostgreSQL
- PostgreSQL-compatible wire protocol and SQL syntax
- Teams can use PostgreSQL tools (psycopg2, JDBC PostgreSQL driver)
- Some Spanner-specific features unavailable (interleaved tables, STRUCT types)
- DDL syntax follows PostgreSQL conventions
- Information schema follows PostgreSQL conventions

Most new Spanner projects choose Google Standard SQL for full feature access. PostgreSQL mode is valuable when migrating existing PostgreSQL workloads or when team expertise is PostgreSQL-centric.

## DDL Lifecycle Behavior

The `ddl` field has nuanced lifecycle behavior that users must understand:

### On Creation
- All DDL statements execute atomically with database creation
- If any statement fails, the entire database creation fails
- This is the safest way to ensure a database starts with a known schema

### After Creation (Updates)
- New DDL statements can be appended to the list
- Appended statements execute via the GCP UpdateDDL API
- Modifying or removing existing statements triggers database recreation
- This behavior is enforced by the Terraform provider's custom diff logic

### Recommendations
- Use `ddl` for initial schema only (tables, indexes, foreign keys)
- For ongoing schema evolution, use a migration tool (Liquibase, Flyway, Atlas)
- Do not manage `version_retention_period` via DDL if also setting it in the spec

## Encryption: Google-Managed vs. CMEK

| Feature | Google-Managed (default) | CMEK |
|---|---|---|
| Key management | Automatic, no user action | User manages key lifecycle in KMS |
| Compliance | Meets most requirements | Required for HIPAA, PCI-DSS, FedRAMP |
| Key rotation | Automatic | User-controlled rotation schedule |
| Key location | Same region as data | Must be in same location as instance |
| Cost | Free | KMS key usage charges apply |

When using CMEK:
- The KMS key must exist in the same GCP location as the Spanner instance
- The Spanner service account must have `cloudkms.cryptoKeyEncrypterDecrypter` role on the key
- Encryption config is immutable -- changing the key requires recreating the database

## Infra Chart Composition Patterns

### Pattern 1: Instance + Database (most common)
```
GcpSpannerInstance (Layer 0)
└── GcpSpannerDatabase (Layer 1, references instance via valueFrom)
```

### Pattern 2: Instance + Database + KMS (enterprise)
```
GcpKmsKeyRing (Layer 0)
├── GcpKmsKey (Layer 1)
│   └── GcpSpannerDatabase (Layer 2, references key via valueFrom)
└── GcpSpannerInstance (Layer 1)
    └── GcpSpannerDatabase (Layer 2, references instance via valueFrom)
```

### Pattern 3: Multi-database instance
```
GcpSpannerInstance (Layer 0)
├── GcpSpannerDatabase "users" (Layer 1)
├── GcpSpannerDatabase "orders" (Layer 1)
└── GcpSpannerDatabase "analytics" (Layer 1)
```

## Provider Version Requirements

- **Terraform**: `~> 6.0` (consistent with GcpSpannerInstance)
- **Pulumi**: `v9` (pulumi-gcp SDK)

All fields used in this component are available in both Terraform 5.x and 6.x. We use `~> 6.0` for consistency with GcpSpannerInstance, which requires v6 for `instance_type`, `edition`, and `default_backup_schedule_type`.

## References

- [Cloud Spanner Documentation](https://cloud.google.com/spanner/docs)
- [Terraform google_spanner_database](https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/spanner_database)
- [Pulumi gcp.spanner.Database](https://www.pulumi.com/registry/packages/gcp/api-docs/spanner/database/)
- [Spanner DDL Reference](https://cloud.google.com/spanner/docs/reference/standard-sql/data-definition-language)
- [Spanner CMEK](https://cloud.google.com/spanner/docs/cmek)
