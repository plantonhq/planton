# GcpSpannerDatabase

Provisions a [Google Cloud Spanner](https://cloud.google.com/spanner) database within an existing Spanner instance.

## What It Does

A Spanner database is a collection of tables, views, indexes, and other schema objects that live inside a Spanner instance. This component creates and manages the database itself, including its SQL dialect, initial schema (via DDL), encryption configuration, and point-in-time recovery settings.

A Spanner database is **not** an instance. The instance provides compute capacity and regional configuration; the database stores your data and schema. Multiple databases can share a single instance.

## When to Use

- You have an existing GcpSpannerInstance and need to create a database within it
- You want to manage the database's lifecycle (creation, schema initialization, encryption) through OpenMCF
- You need to enforce CMEK encryption or drop protection for compliance

## Key Configuration

### SQL Dialect (`database_dialect`)

Choose the SQL dialect at creation time. This is permanent and cannot be changed:

| Dialect | When to Use |
|---|---|
| `GOOGLE_STANDARD_SQL` | Full Spanner feature set including interleaved tables, STRUCT types, parameterized queries. Default. |
| `POSTGRESQL` | Teams familiar with PostgreSQL syntax. Some Spanner-specific features may not be available. |

### Initial DDL (`ddl`)

Provide DDL statements to create tables, indexes, and other schema objects atomically with the database. If any statement fails, the database is not created.

**Important:** After creation, new DDL statements can be appended. However, modifying or removing existing statements forces database recreation. For ongoing schema management, use a migration tool (Liquibase, Flyway, etc.).

### Point-in-Time Recovery (`version_retention_period`)

Controls how far back you can read data at a previous timestamp. Range: 1 hour to 7 days. Default: 1 hour. Higher values consume more storage but enable longer recovery windows.

### Labels

Spanner databases do **not** support GCP labels. Labels are managed at the instance level via GcpSpannerInstance.

## Outputs

| Output | Description |
|---|---|
| `database_id` | Fully qualified path (`projects/{project}/instances/{instance}/databases/{name}`) |
| `database_name` | Short name (the value passed in `database_name`) |
| `state` | CREATING or READY |

## Relationships

- **Depends on**: GcpSpannerInstance (instance), GcpProject (project_id), optionally GcpKmsKey (kms_key_name)
- **Referenced by**: Application connection strings, IAM bindings

## Deployment

```shell
openmcf apply -f spanner-database.yaml
```

For copy-paste ready manifests, see [examples.md](examples.md).
