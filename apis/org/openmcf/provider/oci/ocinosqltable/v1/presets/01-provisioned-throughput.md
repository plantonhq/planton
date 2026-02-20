# Provisioned Throughput

This preset creates a NoSQL table with provisioned throughput capacity, a JSON-friendly schema with a timestamp-based secondary index, and 25 GB of storage. Provisioned mode gives you predictable costs and guaranteed throughput for workloads with well-understood access patterns.

## When to Use

- Applications with predictable read/write traffic patterns where capacity can be estimated
- Cost-sensitive workloads where provisioned pricing is more economical than on-demand
- Key-value or document stores backing APIs, user profiles, device registries, or session data
- Any NoSQL workload where you want to set explicit throughput ceilings

## Key Configuration Choices

- **Provisioned capacity** (`capacityMode: provisioned`) -- you specify maximum read and write units. OCI enforces these limits and charges based on provisioned capacity, which is cheaper than on-demand for sustained workloads.
- **100 read units / 100 write units** -- a moderate baseline handling approximately 100 single-row reads and 100 single-row writes per second. Adjust based on your application's access patterns.
- **25 GB storage** (`maxStorageInGbs: 25`) -- sufficient for millions of small JSON documents. Storage is allocated in advance; increase as the dataset grows.
- **JSON data column** -- the `data` column uses OCI NoSQL's native JSON type, enabling schema-flexible document storage with server-side query support (SQL-like syntax over JSON fields).
- **Secondary index on created_at** -- enables efficient range queries on the timestamp column for time-based retrieval patterns (recent items, time-window queries).
- **DDL-based schema** -- OCI NoSQL uses DDL statements as the schema definition mechanism. The template DDL creates a string primary key, a JSON data column, and a millisecond-precision timestamp.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<compartment-ocid>` | OCID of the compartment for the table | OCI Console > Identity > Compartments, or `OciCompartment` outputs |
| `<table-name>` | Name of the NoSQL table (must match in both `name` and `ddlStatement`) | Choose a name following your naming convention (e.g., `users`, `events`, `sessions`) |

## Related Presets

- **02-on-demand** -- Use instead when traffic patterns are unpredictable or bursty and you prefer automatic scaling
