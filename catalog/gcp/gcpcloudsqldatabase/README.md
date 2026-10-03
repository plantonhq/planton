# GCP Cloud SQL Database

Deploys a logical database (`google_sql_database`) inside a Cloud SQL instance. Databases are first-class nodes with their own lifecycle: create and drop application databases freely without ever touching the instance they live on.

## What Gets Created

When you deploy a GcpCloudSqlDatabase resource, Planton provisions:

- **Cloud SQL database** — a `google_sql_database` inside the referenced instance

No API enablement is needed: the instance the database lives on cannot exist without `sqladmin.googleapis.com` already enabled.

## Prerequisites

- **An existing Cloud SQL instance** — referenced via `instance` (a [GcpCloudSql](/docs/catalog/gcp/gcpcloudsql) resource or a literal instance name)
- **GCP credentials** — [`iac/permissions.yaml`](iac/permissions.yaml) lists the exact least-privilege permissions

## Quick Start

Create a file `database.yaml`:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpCloudSqlDatabase
metadata:
  name: orders-database
spec:
  instance:
    valueFrom:
      kind: GcpCloudSql
      name: my-postgres
      fieldPath: status.outputs.instance_name
  databaseName: orders
```

Deploy:

```shell
planton apply -f database.yaml
```

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `instance` | `StringValueOrRef` | The hosting Cloud SQL instance. Immutable. | Ref → GcpCloudSql `instance_name` |
| `databaseName` | `string` | Database name inside the instance. Immutable. | 1–128 chars |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `projectId` | `StringValueOrRef` | provider default project | GCP project. Can reference a GcpProject. |
| `charset` | `string` | engine default | MySQL: e.g. `utf8mb4`. PostgreSQL: must be `UTF8`. Ignored by SQL Server. |
| `collation` | `string` | engine default | MySQL: e.g. `utf8mb4_0900_ai_ci`. PostgreSQL: an OS locale (`en_US.UTF8`). SQL Server: a SQL Server collation. |

## Outputs

| Output | Type | Description |
|--------|------|-------------|
| `database_name` | `string` | Name of the database inside the instance |
| `self_link` | `string` | GCP resource self link |

## Deployment Methods

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md).

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md).

## Important Notes

- **Engine-specific charset/collation** — the API validates the combination at deploy time; PostgreSQL only accepts UTF8 at creation.
- **One database per application** — the standard pattern on a shared instance; pair each database with its own [GcpCloudSqlUser](/docs/catalog/gcp/gcpcloudsqluser).
- **Dropping the resource drops the database and its data** — the instance's backups (and PITR) are the recovery path.

### Deliberately not modeled (recorded reasons)

Nothing. Every configurable argument of `google_sql_database` at the pinned provider is representable through this spec — including `deletionPolicy`, whose `ABANDON` mode is the documented answer for PostgreSQL databases that cannot be dropped while clients hold connections.

## Related Kinds

- [GcpCloudSql](/docs/catalog/gcp/gcpcloudsql) — the instance this database lives on
- [GcpCloudSqlUser](/docs/catalog/gcp/gcpcloudsqluser) — per-application users

## Additional Resources

- [Creating and managing databases](https://cloud.google.com/sql/docs/postgres/create-manage-databases)

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
