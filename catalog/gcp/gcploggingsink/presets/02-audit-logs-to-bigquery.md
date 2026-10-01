# Audit Logs to BigQuery

Audit-log forensics in SQL: every Cloud Audit Logs entry lands in
partitioned BigQuery tables within seconds, queryable by principal,
method, and resource.

## What it configures

- A `bigqueryDataset` destination with `usePartitionedTables: true` —
  date-partitioned tables (pruning, expiration) instead of the legacy
  date-sharded names. BigQuery destinations require the unique writer
  identity, which is the default.
- A `logName` filter scoped to Cloud Audit Logs.

## The deploy's second half

Grant the sink's `writer_identity` output `roles/bigquery.dataEditor`
on the dataset. The dataset's access list is authoritative on
`GcpBigQueryDataset.access`, and a reference there would make the dataset
depend on the sink that depends on it, so grant it outside the catalog or
add a literal `access` entry (`iamMember` set to the identity) once the
sink has minted it.

## Adjust before deploying

- **bigqueryDataset** — reference a GcpBigQueryDataset resource via
  valueFrom (its `self_link` output normalizes automatically).
- Set a partition expiration on the dataset side to cap storage cost.

## When to choose something else

For long-term cold retention, GCS is an order of magnitude cheaper —
many teams run BOTH: this preset for 90-day forensics, the **Error
Archive to GCS** preset for the multi-year trail.
