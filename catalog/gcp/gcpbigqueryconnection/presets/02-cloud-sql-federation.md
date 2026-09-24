# Cloud SQL Federation

## Use Case

Query a Cloud SQL for PostgreSQL database from BigQuery in place with `EXTERNAL_QUERY`, joining live operational data with warehouse tables without copying it.

## When to Use

- Joining operational rows with analytics
- Occasional reporting without a replication pipeline

## What This Creates

- A Cloud SQL connection in `US` to the `orders` database as a read-only user

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `cloudSql.instanceId` | `...:us-central1:orders` | A `GcpCloudSql` reference (`connection_name`); us-central1 pairs with `US`. |
| `credential` | read-only user | Use a database user with SELECT only; the password is stored as a managed secret. |
