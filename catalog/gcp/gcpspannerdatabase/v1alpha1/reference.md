# GcpSpannerDatabase

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

GcpSpannerDatabaseSpec defines the configuration for a Google Cloud
Spanner database.

A Spanner database is a collection of tables, views, indexes, and other
schema objects hosted on a Spanner instance (GcpSpannerInstance). Each
database belongs to exactly one instance and shares its compute capacity
and geographic configuration. Backup schedules are separate composable
resources (GcpSpannerBackupSchedule) that reference this database by
name.

Important behavioral notes:

  - database_name, database_dialect, encryption_config, and instance are
    IMMUTABLE — changing any of them recreates the database.

  - The ddl list is APPEND-ONLY after creation: new statements are
    executed via UpdateDDL; modifying or removing an existing statement
    forces recreation. For ongoing schema management use a migration
    tool (Liquibase, Flyway) — IaC-owned DDL is for the initial schema.

  - Spanner databases do not support GCP labels; labels exist only at
    the instance level.

  - THREE deletion levers exist and differ in enforcement point:
    enable_drop_protection is GCP API-side (blocks deletion through
    Console, gcloud, API, and IaC alike — and blocks deletion of the
    PARENT INSTANCE while set); deletion_protection is IaC-side (both
    engines refuse to destroy this resource while true — GCP itself
    would still allow a console delete) and defaults TRUE, matching
    the safe posture; deletion_policy decides WHAT a permitted destroy
    does (delete the database, fail, or abandon it in GCP).

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpSpannerDatabase
metadata:
  name: orders-db
spec:
  # project_id omitted: falls back to the provider's default project.
  instance:
    value: orders-spanner
  databaseName: orders
  ddl:
    - CREATE TABLE orders (order_id STRING(36) NOT NULL, created_at TIMESTAMP NOT NULL) PRIMARY KEY (order_id)
  # Allow hack-loop teardown; production manifests keep the default (true).
  deletionProtection: false
  # Explicit DELETE (the provider default) proves the round-trip.
  deletionPolicy: DELETE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.instance` | `string \| valueFrom` | yes |  | GcpSpannerInstance (`status.outputs.instance_name`) |
| `spec.databaseName` | `string` | yes |  |  |
| `spec.databaseDialect` | `string` |  |  |  |
| `spec.versionRetentionPeriod` | `string` |  |  |  |
| `spec.ddl` | `[]string` |  |  |  |
| `spec.enableDropProtection` | `bool` |  |  |  |
| `spec.encryptionConfig` | `GcpSpannerDatabaseEncryptionConfig` |  |  |  |
| `spec.encryptionConfig.kmsKeyName` | `string \| valueFrom` |  |  | GcpKmsKey (`status.outputs.key_id`), GcpKmsKeyHandle (`status.outputs.kms_key`) |
| `spec.encryptionConfig.kmsKeyNames` | `[]string \| valueFrom` |  |  | GcpKmsKey (`status.outputs.key_id`) |
| `spec.defaultTimeZone` | `string` |  |  |  |
| `spec.deletionProtection` | `bool` |  | `true` |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

The GCP project that owns the parent Spanner instance. Accepts a
literal project ID or a reference to a GcpProject resource. If
omitted, the provider's default project is used.
Immutable: changing the project destroys and recreates the database.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.instance

`string | valueFrom` · required

The Spanner instance to create the database on. This determines the
compute capacity and geographic configuration available to the
database. Immutable.

- references: GcpSpannerInstance (`status.outputs.instance_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpSpannerInstance, name: <that resource's name>, fieldPath: status.outputs.instance_name}} -- a bare string does not parse

### spec.databaseName

`string` · required

Unique name of the database within the instance. Immutable. If not
specified, defaults to metadata.name. Must be 2-30 characters: start
with a lowercase letter, contain lowercase letters, digits,
underscores, and hyphens, and end with a letter or digit.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"minLen":"2","maxLen":"30","pattern":"^[a-z][a-z0-9_\\-]*[a-z0-9]$"}}

### spec.databaseDialect

`string`

SQL dialect of the database. Immutable — this choice is permanent.

GOOGLE_STANDARD_SQL (default): Google's SQL dialect with full Spanner
feature support, including interleaved tables and STRUCT types.

POSTGRESQL: PostgreSQL-compatible interface for teams standardized on
PostgreSQL syntax and tooling. Some Spanner-specific features are not
available through this dialect.

- rule: database_dialect must be GOOGLE_STANDARD_SQL or POSTGRESQL

### spec.versionRetentionPeriod

`string`

Retention period for database versions, enabling point-in-time
recovery. Between 1 hour and 7 days; accepts duration formats such as
"1h", "24h", "3d", "86400s". GCP defaults to "1h". Mutable — a longer
window widens point-in-time recovery at the cost of extra storage.

### spec.ddl

`[]string`

DDL statements executed when creating the database (tables, indexes,
views). Statements run atomically with creation — if any fails, the
database is not created.

Lifecycle: APPEND-ONLY. New statements added later are applied via
UpdateDDL; modifying or removing an existing entry forces database
recreation. Use a migration tool for ongoing schema management.

### spec.enableDropProtection

`bool`

GCP API-side drop protection. While true, the database cannot be
deleted through ANY interface (Console, gcloud, API, Terraform,
Pulumi) and the parent Spanner instance cannot be deleted either.
Mutable. Defaults to false — prefer deletion_protection (below) for
day-to-day IaC safety and reserve this for compliance-grade locks.

### spec.encryptionConfig

`GcpSpannerDatabaseEncryptionConfig`

Customer-managed encryption (CMEK). Immutable. If omitted,
Google-managed encryption is used. Exactly one key shape inside:
kms_key_name for regional instances, kms_key_names (one per region)
for multi-region instances.

- rule: set exactly one of kms_key_name (regional) or kms_key_names (multi-region)

### spec.encryptionConfig.kmsKeyName

`string | valueFrom`

Fully qualified name of the KMS key for single-region CMEK.
Format: projects/{project}/locations/{location}/keyRings/{ring}/cryptoKeys/{key}

- references: GcpKmsKey (`status.outputs.key_id`), GcpKmsKeyHandle (`status.outputs.kms_key`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_id}} -- a bare string does not parse

### spec.encryptionConfig.kmsKeyNames

`[]string | valueFrom`

Fully qualified KMS key names for multi-region CMEK — one key per
region of the instance's multi-region configuration.

- references: GcpKmsKey (`status.outputs.key_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_id}} -- a bare string does not parse

### spec.defaultTimeZone

`string`

Default time zone for the database, affecting time-zone-dependent SQL
functions. Must be a valid IANA Time Zone Database name (e.g.
"America/New_York", "UTC"). GCP defaults to "America/Los_Angeles".

### spec.deletionProtection

`bool` · optional (explicit presence)

IaC-side deletion guard. While true (the default), BOTH engines refuse
to destroy this resource — a plan that would delete the database fails
before touching GCP. Set false explicitly before an intentional
teardown. Unlike enable_drop_protection, GCP itself does not enforce
this: a console/gcloud delete would still succeed.

- default: `true`

### spec.deletionPolicy

`string`

Deletion policy — what a PERMITTED destroy does, once the two guards
above allow one (deletion_protection false, enable_drop_protection
off):
  ""        -- same as "DELETE" (provider default)
  "DELETE"  -- the database and its data are deleted (backups
               already taken survive until their retention expires)
  "PREVENT" -- destroy FAILS — a third, explicit wall for databases
               whose teardown must never ride along with a stack's
  "ABANDON" -- the database is removed from management but left
               intact in GCP, still billing for its storage

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpSpannerDatabase, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.database_id` | `string` | Fully qualified database ID. Format: projects/{project}/instances/{instance}/databases/{database} This is the canonical identifier used for IAM bindings, API calls, and connection strings. |
| `status.outputs.database_name` | `string` | Short database name. This is the value that GcpSpannerBackupSchedule and other downstream resources use to reference this database. |
| `status.outputs.state` | `string` | Database state: CREATING or READY. CREATING indicates the database is being provisioned. READY indicates the database is available for queries. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.instance` | GcpSpannerInstance | `status.outputs.instance_name` |
| `spec.encryptionConfig.kmsKeyName` | GcpKmsKey | `status.outputs.key_id` |
| `spec.encryptionConfig.kmsKeyName` | GcpKmsKeyHandle | `status.outputs.kms_key` |
| `spec.encryptionConfig.kmsKeyNames` | GcpKmsKey | `status.outputs.key_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| GcpSpannerBackupSchedule | `spec.database` | `status.outputs.database_name` |

## See Also

- [Overview](../README.md)
