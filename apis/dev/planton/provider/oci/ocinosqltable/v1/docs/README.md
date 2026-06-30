# OciNosqlTable — Design Notes

## Overview

This document covers the design rationale behind the OciNosqlTable component, including why DDL-based schema definition was chosen over field-based schema, the trade-offs of bundling indexes as sub-resources, and what was deferred from v1.

## DDL vs. Field-Based Schema

### Decision

Table schema is defined via a raw DDL statement (`ddlStatement` field) rather than a structured proto message that enumerates columns, types, and keys.

### Rationale

1. **OCI NoSQL's native interface is DDL** — The OCI NoSQL API (`CreateTable`, `UpdateTable`) accepts a DDL string. The Terraform and Pulumi providers pass this string through directly. Wrapping DDL in a structured proto would require building a DDL generator that handles all column types (INTEGER, LONG, FLOAT, DOUBLE, NUMBER, STRING, BOOLEAN, BINARY, TIMESTAMP, JSON, ENUM, ARRAY, MAP, RECORD, UUID), primary key declarations, shard key declarations, TTL defaults, and identity columns.

2. **Schema evolution uses DDL** — Adding or modifying columns is done via `ALTER TABLE` DDL. A field-based model would need to diff the current column list against the desired state and emit the correct `ALTER TABLE` statements. This diffing logic is non-trivial and error-prone, especially for type changes that OCI NoSQL does not allow.

3. **Full DDL expressiveness** — DDL supports features like `USING TTL`, `DEFAULT`, `COMMENT`, `IDENTITY`, and nested type constructors (`ARRAY(RECORD(...))`). A proto-based schema would either need to model all of these or lose expressiveness.

4. **Existing ecosystem familiarity** — Teams already working with OCI NoSQL write DDL. Requiring them to translate DDL into a proto-derived YAML structure adds friction without reducing errors.

### Trade-offs

- **No structural validation at the manifest level** — The proto validates that `ddlStatement` is a non-empty string, but it does not parse or validate the DDL syntax. Invalid DDL fails at deploy time, not at manifest validation time.
- **Name consistency is the user's responsibility** — The `name` field and the table name inside the DDL must match. The IaC module passes both to OCI independently and does not cross-validate them.
- **Schema drift detection is limited** — If someone modifies the table schema outside of Planton (e.g., through the OCI Console), the manifest's DDL may not reflect the actual table state. Pulumi's state tracking handles the table resource itself, but DDL changes are opaque to the state model.

## Index Sub-Resource Design

### Decision

Secondary indexes are declared as a repeated `Index` message within `OciNosqlTableSpec`, not as separate top-level Planton resources.

### Rationale

1. **Indexes are tightly coupled to their table** — An index references columns defined in the table's DDL. It cannot exist independently. Modeling them as sub-resources reflects this lifecycle dependency.

2. **Single manifest for the full table definition** — Teams can see the table schema (DDL) and all its indexes in one file. This avoids the coordination problem of keeping separate index manifests in sync with the table schema.

3. **OCI API dependency** — Creating an index requires the table's OCID. The Pulumi module uses `pulumi.DependsOn` to ensure the table is created before any index. Bundling them in one module makes this dependency explicit and avoids cross-stack references.

### Trade-offs

- **Index changes redeploy the full stack** — Modifying any index in the list triggers a Pulumi update for the entire OciNosqlTable stack, even if the table itself is unchanged. For tables with many indexes, this increases the blast radius of index-only changes.
- **No independent index lifecycle** — You cannot deploy or destroy a single index without deploying the full table stack. If independent index management becomes necessary, a separate OciNosqlIndex resource type could be introduced in a future version.

## Freeform Tag Propagation

The IaC module automatically populates OCI freeform tags from the manifest metadata:

| Tag Key | Source |
|---------|--------|
| `resource` | Always `"true"` |
| `resource_kind` | `OciNosqlTable` (from the CloudResourceKind enum) |
| `resource_id` | `metadata.id` |
| `organization` | `metadata.org` (if non-empty) |
| `environment` | `metadata.env` (if non-empty) |
| All label keys | `metadata.labels` (merged) |

OCI limits freeform tags to 10 per resource by default. The fixed tags consume 3–5 slots, leaving 5–7 for user-defined labels.

## Capacity Mode Mapping

The proto enum `CapacityMode` maps to OCI API values as follows:

| Proto Value | OCI API Value | Behavior |
|-------------|---------------|----------|
| `capacity_mode_unspecified` (0) | *(not sent)* | OCI defaults to provisioned. `maxReadUnits` and `maxWriteUnits` are required. |
| `provisioned` (1) | `PROVISIONED` | User specifies read/write units. |
| `on_demand` (2) | `ON_DEMAND` | OCI auto-scales throughput. Read/write units are ignored. |

The Go module converts the enum to its uppercase string form via `strings.ToUpper(spec.TableLimits.CapacityMode.String())` and only sets the `CapacityMode` field on the Pulumi args when it is not `capacity_mode_unspecified`.

## What's Deferred from v1

| Feature | Reason for Deferral |
|---------|---------------------|
| `defined_tags` / `system_tags` | Managed at the platform level. Freeform tags cover the common use case. |
| `freeform_tags` (user-specified) | Auto-populated from metadata labels. Direct specification would conflict. |
| Multi-region replicas | Advanced replication feature with significant operational complexity. Requires cross-region provider setup and replica table coordination. |
| Child tables (names containing `.`) | OCI NoSQL child tables inherit the parent's shard key. Managing parent-child relationships across separate Planton resources requires a dependency model not yet in place. Workaround: deploy each table as a standalone resource with its own shard key. |
| Table-level TTL defaults | Can be specified within the DDL (`USING TTL`). A dedicated proto field is unnecessary until DDL parsing is implemented. |
| Read-only replicas | Not available in the OCI NoSQL API as a table-level configuration. |

## Module Structure

```
v1/
├── api.proto              # Top-level OciNosqlTable message
├── spec.proto             # OciNosqlTableSpec, TableLimits, Index, IndexKey
├── stack_outputs.proto    # OciNosqlTableStackOutputs (tableId)
└── iac/pulumi/module/
    ├── main.go            # Entry point: provider setup, orchestration
    ├── locals.go          # Locals struct: table name, freeform tags
    ├── nosql_table.go     # Table + index creation logic
    └── outputs.go         # Stack output key constants
```

The module follows the standard Planton Pulumi module pattern: `main.go` initializes locals and the OCI provider, then delegates to resource-specific functions. `nosql_table.go` creates the table, exports the table OCID, and iterates over the index list to create each secondary index with an explicit dependency on the table resource.
