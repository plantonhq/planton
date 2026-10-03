# GCP Spanner Database

Deploys a Cloud Spanner database within an existing Spanner instance, with configurable SQL dialect (GoogleSQL or PostgreSQL), version retention for point-in-time recovery, initial DDL statements, optional CMEK encryption, and API-level drop protection. The database name, dialect, encryption posture, and host instance are all immutable — the durable decisions live at creation time, while retention, drop protection, and appended DDL update in place.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Spanner Database** -- a managed database within the specified Spanner instance, configured with the chosen SQL dialect and version retention period
- **Initial Schema** -- when `ddl` statements are specified, executes them atomically during database creation (tables, indexes, views)
- **CMEK Encryption** -- created only when `encryptionConfig` is set; encrypts the database with customer-managed Cloud KMS keys. Exactly one key shape inside: `kmsKeyName` (one key, for databases on regional instance configurations) or `kmsKeyNames` (one key per region of a multi-region configuration)
- **Drop Protection** -- created only when `enableDropProtection` is set to `true`; prevents deletion of the database and its parent instance through any interface. A second, IaC-side guard — `deletionProtection`, default `true` — makes both engines refuse to destroy the database until it is explicitly set `false`
- **Spanner API enablement** -- `spanner.googleapis.com` enabled in the target project so a fresh project can host databases (never disabled on destroy)

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with credentials for the target GCP project. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### GCP Project

- **A GCP project** matching the project of the parent Spanner instance. Provide the project ID directly or reference a GcpProject Infra Component via ValueFromRef.
- **An existing Spanner instance** to host the database. Provide the instance name directly or reference a GcpSpannerInstance Infra Component via ValueFromRef.
- **Cloud KMS key** (if using CMEK) -- the key must be in the same location as the Spanner instance. The Spanner service account must have `roles/cloudkms.cryptoKeyEncrypterDecrypter` on the key.

## Deploy

### Console

Open the deployment store, find **GCP Spanner Database**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Basic Database** preset in the [Presets](#presets) tab to pre-populate a working configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpSpannerDatabase
metadata:
  name: app-database
  org: acme-corp
  env: prod
spec:
  projectId:
    value: "acme-prod-12345"
  instance:
    value: "app-spanner-prod"
  databaseName: app-db
```

```shell
planton apply -f spanner-database.yaml
```

This creates a database with the GoogleSQL dialect, 1-hour default version retention, Google-managed encryption, and the IaC-side deletion guard on by default. Schema must be managed separately via migration tools or subsequent DDL updates. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the database to a GCP project, Spanner instance, and KMS key deployed in the same InfraPipeline:

```yaml
spec:
  projectId:
    valueFrom:
      kind: GcpProject
      name: production-project
      fieldPath: status.outputs.project_id
  instance:
    valueFrom:
      kind: GcpSpannerInstance
      name: app-spanner
      fieldPath: status.outputs.instance_name
  encryptionConfig:
    kmsKeyName:
      valueFrom:
        kind: GcpKmsKey
        name: spanner-cmek-key
        fieldPath: status.outputs.key_id
```

The InfraPipeline resolves the dependency graph, deploys the project and Spanner instance first, then provisions the database with the resolved instance reference and CMEK encryption.

## Key Configuration

These are the most important decisions when configuring a Spanner database. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**SQL dialect** -- `databaseDialect` chooses between `GOOGLE_STANDARD_SQL` (default, full Spanner feature support including interleaved tables and STRUCT types) and `POSTGRESQL` (PostgreSQL-compatible syntax and wire protocol for teams migrating from PostgreSQL). The dialect is immutable after creation.

**Version retention** -- `versionRetentionPeriod` controls the point-in-time recovery window (1 hour to 7 days, e.g., `"168h"`). Longer retention provides a wider recovery window but consumes more storage. Default is `"1h"`.

**Initial DDL** -- The `ddl` field executes DDL statements atomically with database creation. New statements can be appended after creation, but modifying or removing existing statements forces database recreation. For ongoing schema management, use a migration tool (Liquibase, Flyway).

**Three deletion levers** -- `enableDropProtection` is GCP API-side: while `true`, neither the database nor its parent instance can be deleted through any interface (Console, gcloud, API, or IaC) — reserve it for compliance-grade locks. `deletionProtection` is IaC-side and defaults to `true`: both engines refuse to destroy the database until it is explicitly set `false` — GCP itself would still allow a console delete. `deletionPolicy` decides what a PERMITTED destroy does once both guards allow one: `DELETE` removes the database (backups already taken survive until their retention expires), `PREVENT` fails the destroy, `ABANDON` drops it from management while it keeps billing for storage in GCP.

**CMEK encryption** -- `encryptionConfig` is immutable (the encryption posture is fixed at creation). The key shape follows the parent instance's configuration: `kmsKeyName` for a regional configuration (the key must live in the same location), `kmsKeyNames` for a multi-region configuration (one key per region, each in its region). Omit the message entirely for Google-managed encryption.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** (optional) | `projectId` | `status.outputs.project_id` |
| **GcpSpannerInstance** | `instance` | `status.outputs.instance_name` |
| **GcpKmsKey** (optional) | `encryptionConfig.kmsKeyName` / `encryptionConfig.kmsKeyNames` | `status.outputs.key_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `database_id` | Fully qualified database ID (`projects/{p}/instances/{i}/databases/{d}`) | Application connection strings, IAM bindings |
| `database_name` | Short database name | GcpSpannerBackupSchedule `database` field, display, application configuration |
| `state` | Database state (`CREATING` or `READY`) | Deployment validation, health checks |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Basic database** -- GoogleSQL dialect with a 24-hour point-in-time recovery window and no CMEK or drop protection. The simplest starting point for any Spanner workload. Start from the **Basic Database** preset.

**PostgreSQL database** -- PostgreSQL-compatible dialect with 7-day version retention. Designed for teams migrating from PostgreSQL or using PostgreSQL-compatible client libraries and tooling. Start from the **PostgreSQL-Dialect Database** preset.

**CMEK encrypted** -- GoogleSQL dialect with customer-managed encryption, API-level drop protection, 3-day version retention, and explicit UTC time zone. Designed for regulated environments requiring encryption key control and deletion safeguards. Start from the **CMEK-Encrypted Database** preset.

## Works With

- [**GCP Project**](/infra-catalog/gcp-project) -- provides the GCP project where the Spanner database is created
- [**GCP Spanner Instance**](/infra-catalog/gcp-spanner-instance) -- provides the Spanner instance that hosts the database
- [**GCP KMS Key**](/infra-catalog/gcp-kms-key) -- provides the Cloud KMS key for database-level CMEK encryption