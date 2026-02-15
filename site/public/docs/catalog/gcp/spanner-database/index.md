---
title: "Spanner Database"
description: "Spanner Database deployment documentation"
icon: "package"
order: 100
componentName: "gcpspannerdatabase"
---

# GCP Spanner Database

Deploys a Cloud Spanner database within an existing Spanner instance, with support for GoogleSQL and PostgreSQL dialects, initial DDL schema creation, CMEK encryption, and configurable point-in-time recovery.

## What Gets Created

When you deploy a GcpSpannerDatabase resource, OpenMCF provisions:

- **Spanner Database** — a `google_spanner_database` resource in the specified instance with the chosen SQL dialect and version retention period
- **Initial Schema** — created only when `ddl` is provided, DDL statements execute atomically with database creation
- **CMEK Encryption** — configured only when `kmsKeyName` is provided, encrypts the database with a customer-managed KMS key

## Prerequisites

- **GCP credentials** configured via environment variables or OpenMCF provider config
- **An existing Spanner instance** (deploy via GcpSpannerInstance first)
- **A KMS key** in the same location as the Spanner instance if enabling CMEK encryption

## Quick Start

Create a file `spanner-database.yaml`:

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpSpannerDatabase
metadata:
  name: my-database
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.GcpSpannerDatabase.my-database
spec:
  projectId:
    value: my-gcp-project-123
  instance:
    value: my-spanner-instance
  databaseName: my-database
```

Deploy:

```shell
openmcf apply -f spanner-database.yaml
```

This creates an empty Spanner database with the default GoogleSQL dialect and 1-hour version retention in the specified instance.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `projectId` | `StringValueOrRef` | GCP project ID. Must match the instance's project. Can reference a GcpProject resource via `valueFrom`. | Required |
| `instance` | `StringValueOrRef` | Spanner instance name to create the database in. Can reference a GcpSpannerInstance resource via `valueFrom`. | Required |
| `databaseName` | `string` | Unique database name within the instance. Immutable after creation. | Pattern: `^[a-z][a-z0-9_\-]*[a-z0-9]$`, 2-30 characters |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `databaseDialect` | `string` | `GOOGLE_STANDARD_SQL` | SQL dialect. `GOOGLE_STANDARD_SQL` for full Spanner features or `POSTGRESQL` for PostgreSQL-compatible syntax. Immutable after creation. |
| `versionRetentionPeriod` | `string` | `1h` | Point-in-time recovery window. Range: 1 hour to 7 days. Accepts formats: `1h`, `24h`, `1d`, `1440m`, `86400s`. |
| `ddl` | `string[]` | `[]` | DDL statements executed atomically with creation. New statements can be appended after creation; modifying or removing existing statements forces recreation. |
| `enableDropProtection` | `bool` | `false` | GCP API-level deletion protection. When `true`, prevents deletion of the database and its parent instance through any interface. |
| `kmsKeyName` | `StringValueOrRef` | — | Fully qualified KMS key name for CMEK encryption. Must be in the same location as the instance. Immutable. Can reference a GcpKmsKey resource via `valueFrom`. |
| `defaultTimeZone` | `string` | `America/Los_Angeles` | Default time zone for SQL functions like `CURRENT_TIMESTAMP()` and `FORMAT_TIMESTAMP()`. Must be a valid IANA time zone name. |

## Examples

### PostgreSQL Dialect with Extended Retention

A PostgreSQL-compatible database with a 7-day point-in-time recovery window:

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpSpannerDatabase
metadata:
  name: pg-analytics
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.GcpSpannerDatabase.pg-analytics
spec:
  projectId:
    value: my-gcp-project-123
  instance:
    value: prod-spanner
  databaseName: pg-analytics
  databaseDialect: POSTGRESQL
  versionRetentionPeriod: "7d"
```

### Database with Initial Schema

Create the database and its initial tables in a single atomic operation:

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpSpannerDatabase
metadata:
  name: users-db
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.GcpSpannerDatabase.users-db
spec:
  projectId:
    value: my-gcp-project-123
  instance:
    value: prod-spanner
  databaseName: users-db
  ddl:
    - |
      CREATE TABLE Users (
        UserId STRING(36) NOT NULL,
        Email STRING(255) NOT NULL,
        DisplayName STRING(100),
        CreatedAt TIMESTAMP NOT NULL OPTIONS (allow_commit_timestamp=true)
      ) PRIMARY KEY (UserId)
    - |
      CREATE UNIQUE INDEX UsersByEmail ON Users(Email)
```

### CMEK-Encrypted Production Database

A protected database with customer-managed encryption and drop protection:

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpSpannerDatabase
metadata:
  name: secure-db
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.GcpSpannerDatabase.secure-db
spec:
  projectId:
    value: my-gcp-project-123
  instance:
    value: prod-spanner
  databaseName: secure-db
  kmsKeyName:
    value: projects/my-gcp-project-123/locations/us-central1/keyRings/spanner-ring/cryptoKeys/spanner-key
  enableDropProtection: true
  versionRetentionPeriod: "3d"
  defaultTimeZone: UTC
```

### Using Foreign Key References

Reference other OpenMCF-managed resources instead of hardcoding values:

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpSpannerDatabase
metadata:
  name: composed-db
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.GcpSpannerDatabase.composed-db
spec:
  projectId:
    valueFrom:
      kind: GcpProject
      name: my-project
      fieldPath: status.outputs.project_id
  instance:
    valueFrom:
      kind: GcpSpannerInstance
      name: prod-spanner
      fieldPath: status.outputs.instance_name
  databaseName: composed-db
  kmsKeyName:
    valueFrom:
      kind: GcpKmsKey
      name: spanner-key
      fieldPath: status.outputs.key_id
  databaseDialect: GOOGLE_STANDARD_SQL
  versionRetentionPeriod: "7d"
  enableDropProtection: true
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `database_id` | `string` | Fully qualified database path (`projects/{project}/instances/{instance}/databases/{database}`) |
| `database_name` | `string` | Short database name as specified in `databaseName` |
| `state` | `string` | Database state: `CREATING` during provisioning, `READY` when available for queries |

## Related Components

- [GcpSpannerInstance](/docs/catalog/gcp/spanner-instance) — provides the compute instance that hosts this database
- [GcpKmsKey](/docs/catalog/gcp/kms-key) — provides the encryption key for CMEK
- [GcpKmsKeyRing](/docs/catalog/gcp/kms-key-ring) — provides the key ring containing the encryption key
- [GcpProject](/docs/catalog/gcp/project) — provides the GCP project
