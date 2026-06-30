# On-Demand

This preset creates a NoSQL table with on-demand throughput capacity and 50 GB of storage. On-demand mode automatically scales read and write throughput in response to application traffic, eliminating the need to estimate capacity upfront. You pay per request rather than for provisioned units.

## When to Use

- Applications with unpredictable or bursty traffic patterns (viral content, flash sales, batch imports)
- New workloads where access patterns are not yet understood and capacity cannot be estimated
- Event-driven architectures where write volume depends on external triggers
- Tables that experience long periods of low activity punctuated by traffic spikes

## Key Configuration Choices

- **On-demand capacity** (`capacityMode: on_demand`) -- throughput scales automatically. No `maxReadUnits` or `maxWriteUnits` are needed; OCI handles scaling transparently. Pricing is per-request, which is more expensive at steady high throughput but cheaper for variable or low-traffic workloads.
- **50 GB storage** (`maxStorageInGbs: 50`) -- provides room for moderate dataset growth. Storage limits can be increased after creation.
- **Minimal schema** -- a simple string primary key and JSON data column. On-demand tables are often used for schema-flexible workloads where the data model evolves frequently.
- **No secondary indexes** -- kept minimal for the on-demand use case. Add indexes as query patterns become clear. Each index consumes additional storage and write throughput.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<compartment-ocid>` | OCID of the compartment for the table | OCI Console > Identity > Compartments, or `OciCompartment` outputs |
| `<table-name>` | Name of the NoSQL table (must match in both `name` and `ddlStatement`) | Choose a name following your naming convention |

## Related Presets

- **01-provisioned-throughput** -- Use instead when traffic patterns are predictable and provisioned pricing offers better cost efficiency
