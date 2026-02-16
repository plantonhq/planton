# GCP Spanner Database Examples

This document provides YAML examples for deploying Cloud Spanner databases via OpenMCF. Each example includes a use-case description and the manifest.

---

## Example 1: Minimal GoogleSQL Database

**When to use:** Simplest starting point. Creates an empty database with the default GoogleSQL dialect and 1-hour version retention.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpSpannerDatabase
metadata:
  name: app-db
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.GcpSpannerDatabase.app-db
spec:
  projectId:
    value: my-gcp-project-123
  instance:
    value: my-spanner-instance
  databaseName: app-db
```

---

## Example 2: PostgreSQL Dialect with Custom Retention

**When to use:** Teams working with PostgreSQL-compatible syntax who need a longer point-in-time recovery window.

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

---

## Example 3: Database with Initial Schema (DDL)

**When to use:** Create the database and its initial tables in a single atomic operation. Good for greenfield applications where the schema is known upfront.

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

---

## Example 4: CMEK-Encrypted Database

**When to use:** Compliance requirements demand customer-managed encryption keys. The KMS key must exist in the same location as the Spanner instance.

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
```

---

## Example 5: Infra Chart Composition (valueFrom References)

**When to use:** When deploying as part of an infra chart where the instance and KMS key are created by other resources in the same chart.

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

---

## Example 6: Full Production Configuration

**When to use:** Maximum production configuration with all features: PostgreSQL dialect, DDL, CMEK encryption, drop protection, custom retention, and explicit time zone.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpSpannerDatabase
metadata:
  name: prod-db-full
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.GcpSpannerDatabase.prod-db-full
spec:
  projectId:
    value: my-gcp-project-123
  instance:
    value: prod-spanner
  databaseName: prod-db-full
  databaseDialect: GOOGLE_STANDARD_SQL
  versionRetentionPeriod: "7d"
  defaultTimeZone: UTC
  enableDropProtection: true
  kmsKeyName:
    value: projects/my-gcp-project-123/locations/us-central1/keyRings/spanner-ring/cryptoKeys/spanner-key
  ddl:
    - |
      CREATE TABLE Accounts (
        AccountId STRING(36) NOT NULL,
        Name STRING(255) NOT NULL,
        Status STRING(20) NOT NULL,
        CreatedAt TIMESTAMP NOT NULL OPTIONS (allow_commit_timestamp=true),
        UpdatedAt TIMESTAMP OPTIONS (allow_commit_timestamp=true)
      ) PRIMARY KEY (AccountId)
    - |
      CREATE INDEX AccountsByStatus ON Accounts(Status)
```

---

## Deployment

```shell
openmcf apply -f <manifest>.yaml
```

For more details, see the [main README](README.md).
