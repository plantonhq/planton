# Append-Only History

## Use Case

Keep every change to a table as its own BigQuery row -- inserts, updates, and deletes with their change type -- for auditing, point-in-time analysis, and slowly changing dimensions, partitioned by day and clustered by customer so the history stays cheap to query.

## When to Use

- Audit trails and change history
- Rebuilding a table's state as of any past moment

## What This Creates

- A stream from `public.orders` into the `orders-history` dataset, append-only, changes only (no backfill), the table partitioned by ingestion day with a required partition filter and clustered by `customer_id`, created `NOT_STARTED`

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `destinationConfig.bigqueryDestinationConfig.writeMode` | `APPEND_ONLY` | Immutable; `MERGE` mirrors current state instead. |
| `ruleSets[].customizationRules` | day partitions, clustered by `customer_id` | Partition on a timestamp column (`timeUnitPartition`) or integer ranges instead. |
| `backfillNone` | `true` | Use `backfillAll: {}` to include today's rows as the first history. |
