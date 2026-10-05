# Application User (Built-in)

This preset creates a classic username/password user for one application, with a lockout policy after repeated failed logins. One user per application — never share the instance's admin user across services.

## When to Use

- The standard credential for an application connecting to its database
- Anywhere you need password rotation without recreating the user (updating `password` rotates in place)

## Key Configuration Choices

- **Instance by reference** — resolves the `GcpCloudSql` node's `instance_name` output
- **Lockout policy** — five failed attempts lock the user (`enableFailedAttemptsCheck` turns counting on)
- **Mutable password** — the one field you can change in place; rotate by updating the manifest

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `my-gcp-project-123` | GCP project ID | GCP Console or `GcpProject` outputs |
| `my-postgres-prod` | Your instance's resource name | The instance manifest |
| `password` | The user's credential | Generate a strong secret |

## Related Presets

- **02-iam-service-account-user** — the passwordless alternative for workloads with IAM identities

## Related Kinds

- [GcpCloudSql](/docs/catalog/gcp/gcpcloudsql) — the instance this user lives on
- [GcpCloudSqlDatabase](/docs/catalog/gcp/gcpcloudsqldatabase) — pair the user with its application database
