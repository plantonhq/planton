# PostgreSQL to BigQuery (Merge)

## Use Case

Keep BigQuery tables that mirror a PostgreSQL database's current state: backfill what exists, then merge every change by primary key, so analysts query fresh operational data without touching the database.

## When to Use

- Operational reporting on an application database
- Replacing nightly batch exports with continuous replication

## What This Creates

- A stream from the `public` schema of the `orders-postgres` source into one BigQuery dataset per schema (`orders_public`), merged, readable within 15 minutes of a change, created `NOT_STARTED`

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `desiredState` | `NOT_STARTED` | Set `RUNNING` once the profiles and grants are in place; the backfill then starts and bills. |
| `destinationConfig.bigqueryDestinationConfig.dataFreshness` | `900s` | Lower for fresher reads at more BigQuery compute. |
| `sourceConfig.postgresqlSourceConfig.includeObjects` | the `public` schema | Narrow to the tables you need; fewer tables, less processed data. |
| `backfillAll` | everything included | Use `backfillNone: true` to replicate only changes from now on. |
