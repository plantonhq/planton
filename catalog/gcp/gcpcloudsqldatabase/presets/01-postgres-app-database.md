# PostgreSQL Application Database

This preset creates one application database on an existing PostgreSQL instance, referencing the instance by name. PostgreSQL databases are created as UTF8; the engine defaults handle collation, so the minimal form is usually the right form.

## When to Use

- One database per application/service on a shared instance — the standard multi-tenant-instance pattern
- Ephemeral databases (previews, staging datasets) that come and go without touching the instance

## Key Configuration Choices

- **Instance by reference** — resolves the `GcpCloudSql` node's `instance_name` output, so the database deploys after (and composes with) its instance
- **No charset/collation** — PostgreSQL requires UTF8 at creation; the engine default collation applies

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `my-gcp-project-123` | GCP project ID | GCP Console or `GcpProject` outputs |
| `my-postgres-prod` | Your instance's resource name | The instance manifest |
| `orders` | The database name | Your application's configuration |

## Related Presets

- **02-mysql-utf8mb4-database** — the MySQL form with an explicit modern charset

## Related Kinds

- [GcpCloudSql](/docs/catalog/gcp/gcpcloudsql) — the instance this database lives on
- [GcpCloudSqlUser](/docs/catalog/gcp/gcpcloudsqluser) — pair each application database with its own user
