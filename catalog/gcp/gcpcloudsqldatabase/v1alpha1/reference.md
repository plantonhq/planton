# GcpCloudSqlDatabase

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

GcpCloudSqlDatabaseSpec defines a logical database (`google_sql_database`)
inside a Cloud SQL instance.

An instance hosts many databases, each with its own lifecycle — create and
drop application databases without ever touching the instance. The
instance is referenced by name (the GcpCloudSql instance_name output), so
one instance node composes with any number of database nodes.

Character set and collation semantics are engine-specific:
  - MySQL accepts any supported charset/collation pair (e.g. utf8mb4 +
    utf8mb4_0900_ai_ci).
  - PostgreSQL databases are created with UTF8; collation is an operating
    system locale name (e.g. en_US.UTF8).
  - SQL Server ignores charset; collation is a SQL Server collation name.

## Example

```yaml
# Exercises the database surface offline: instance by literal name plus
# explicit MySQL charset/collation.
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpCloudSqlDatabase
metadata:
  name: hack-orders-db
spec:
  # project_id omitted — falls back to the provider's default project.
  instance:
    value: hack-mysql
  databaseName: orders
  charset: utf8mb4
  collation: utf8mb4_0900_ai_ci
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.instance` | `string \| valueFrom` | yes |  | GcpCloudSql (`status.outputs.instance_name`) |
| `spec.databaseName` | `string` | yes |  |  |
| `spec.charset` | `string` |  |  |  |
| `spec.collation` | `string` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

The GCP project that owns the Cloud SQL instance.
Can be a literal project ID or a reference to a GcpProject resource.
If omitted, the provider's default project is used.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.instance

`string | valueFrom` · required

The Cloud SQL instance hosting this database. Accepts the instance name
or a reference to a GcpCloudSql resource. Immutable — a database cannot
move between instances.

- references: GcpCloudSql (`status.outputs.instance_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpCloudSql, name: <that resource's name>, fieldPath: status.outputs.instance_name}} -- a bare string does not parse

### spec.databaseName

`string` · required

Name of the database inside the instance. Immutable.
Example: "orders", "analytics_staging"

- rule: {"required":true,"string":{"minLen":"1","maxLen":"128"}}

### spec.charset

`string`

Character set. MySQL: e.g. "utf8mb4" (recommended). PostgreSQL: must be
"UTF8" at creation. Ignored by SQL Server. If empty, the engine default
applies.

### spec.collation

`string`

Collation. MySQL: e.g. "utf8mb4_0900_ai_ci". PostgreSQL: an OS locale
such as "en_US.UTF8". SQL Server: a SQL Server collation name. If
empty, the engine default applies.

### spec.deletionPolicy

`string`

Engine-side teardown behavior. "DELETE" (default) drops the database;
"PREVENT" fails any plan that would drop it; "ABANDON" removes it
from IaC management while leaving it in the instance. ABANDON is the
documented answer for PostgreSQL databases that cannot be dropped
while clients hold connections.

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpCloudSqlDatabase, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.database_name` | `string` | Name of the database inside the instance. |
| `status.outputs.self_link` | `string` | GCP resource self link for the database. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.instance` | GcpCloudSql | `status.outputs.instance_name` |

## See Also

- [Overview](../README.md)
