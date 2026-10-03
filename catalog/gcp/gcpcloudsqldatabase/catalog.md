# GCP Cloud SQL Database

Creates a database inside an existing Google Cloud SQL instance. Databases are composable satellites of the instance — one instance hosts many databases, each owned by its own application, each created, reviewed, and deleted as a first-class Infra Component. Pairs naturally with GCP Cloud SQL User for per-application credentials.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Cloud SQL Database** -- a `google_sql_database` on the referenced instance, with the specified name and (optionally) an engine-specific character set and collation

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with credentials for the target GCP project. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.

### GCP Project

- **A Cloud SQL instance** -- the [GcpCloudSql](/infra-catalog/gcp-cloud-sql) instance that hosts the database. Reference it via ValueFromRef so the pipeline deploys the instance first.

## Deploy

### Console

Open the deployment store, find **GCP Cloud SQL Database**, and click **Deploy**. The creation wizard walks two decisions: which instance hosts the database, then the database's name and collation. Start from the **PostgreSQL Application Database** or **MySQL Database (utf8mb4)** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpCloudSqlDatabase
metadata:
  name: orders
  org: acme-corp
  env: prod
spec:
  projectId:
    value: "acme-prod-12345"
  instance:
    valueFrom:
      kind: GcpCloudSql
      name: orders-db-prod
      fieldPath: status.outputs.instance_name
  databaseName: orders
  charset: UTF8
  collation: en_US.UTF8
```

```shell
planton apply -f database.yaml
```

This creates a UTF8 `orders` database on the referenced instance, ready for a GcpCloudSqlUser to connect to. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, wire the hosting instance by reference so the pipeline orders the graph:

```yaml
spec:
  instance:
    valueFrom:
      kind: GcpCloudSql
      name: orders-db-prod
      fieldPath: status.outputs.instance_name
  databaseName: orders
```

The InfraPipeline deploys the Cloud SQL instance first, then creates the database on it.

## Key Configuration

These are the most important decisions when configuring a Cloud SQL database. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Instance** -- Immutable. A database never moves between instances; relocating data is an export/import. Reference the GcpCloudSql resource rather than typing the name.

**Database name** -- Immutable, max 128 characters. It is what applications put in their connection strings — name by the owning application.

**Charset and collation** -- Engine-interpreted: MySQL accepts any supported pair (`utf8mb4` + `utf8mb4_0900_ai_ci` is the modern default); PostgreSQL wants `UTF8` with an OS locale collation (`en_US.UTF8`); SQL Server ignores charset and uses its own collation names. Empty keeps the engine default.

**Deletion policy** -- `DELETE` (the default) drops the database on destroy; `PREVENT` fails any plan that would drop it. `ABANDON` removes it from IaC management while leaving it in the instance — the documented answer for PostgreSQL databases that cannot be dropped while clients hold connections.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** | `projectId` | `status.outputs.project_id` |
| **GcpCloudSql** | `instance` | `status.outputs.instance_name` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `database_name` | The created database's name | Application connection strings, service configuration |
| `self_link` | GCP resource self link | Audit log filters |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**One database per application** -- the composition this kind exists for: a shared production instance hosts `orders`, `billing`, and `analytics` as separate resources, each owned, reviewed, and torn down with its application — never with the instance. Start from the **PostgreSQL Application Database** preset.

**MySQL with modern Unicode** -- MySQL's legacy `utf8` charset is three-byte and silently rejects emoji and supplementary characters; declare `utf8mb4` + `utf8mb4_0900_ai_ci` explicitly at creation, because converting a populated database later means rewriting every table. Start from the **MySQL Database (utf8mb4)** preset.

**Instance-plus-satellites chart** -- one GcpCloudSql node with a database and a user node per application, all wired by ValueFromRef; adding the next application to the environment is two small manifests, not an instance change.

## Works With

- [**GCP Cloud SQL**](/infra-catalog/gcp-cloud-sql) -- the instance that hosts this database
- [**GCP Cloud SQL User**](/infra-catalog/gcp-cloud-sql-user) -- per-application credentials for connecting to this database
