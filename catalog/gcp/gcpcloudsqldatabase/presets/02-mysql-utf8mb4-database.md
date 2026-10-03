# MySQL Database (utf8mb4)

This preset creates a MySQL application database with the modern `utf8mb4` character set — full 4-byte UTF-8 that stores emoji and astral-plane characters correctly, unlike MySQL's legacy 3-byte `utf8`.

## When to Use

- Any new MySQL application database — utf8mb4 should be the default choice in 2026
- Databases storing user-generated content (names, comments, anything with emoji)

## Key Configuration Choices

- **`charset: utf8mb4`** — MySQL's real UTF-8; the legacy `utf8` alias silently truncates 4-byte characters
- **`collation: utf8mb4_0900_ai_ci`** — MySQL 8.0's default modern collation (accent-insensitive, case-insensitive, Unicode 9.0 rules)

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `my-gcp-project-123` | GCP project ID | GCP Console or `GcpProject` outputs |
| `my-mysql-ha` | Your instance's resource name | The instance manifest |
| `app` | The database name | Your application's configuration |

## Related Presets

- **01-postgres-app-database** — the PostgreSQL form (UTF8 is implicit there)

## Related Kinds

- [GcpCloudSql](/docs/catalog/gcp/gcpcloudsql) — the instance this database lives on
- [GcpCloudSqlUser](/docs/catalog/gcp/gcpcloudsqluser) — pair each application database with its own user
