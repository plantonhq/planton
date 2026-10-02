# Spanner with Data Boost

## Use Case

Federated analytics over a Spanner database on Data Boost -- independent compute that leaves the instance's serving capacity to the application -- under a fine-grained read role.

## When to Use

- Analytics over a production Spanner database
- Workloads that must not affect Spanner latency

## What This Creates

- A Spanner connection in `us-central1` reading as the `analyst` role on Data Boost with up to 8 parallel reads

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `cloudSpanner.database` | `project/instance/database` | Slashes, no `projects/` prefix. |
| `useDataBoost` | `true` | Billed by Spanner as Data Boost compute; needs parallelism. |
