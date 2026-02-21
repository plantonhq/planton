# OCI NoSQL Table Deployment Component

**Date**: February 19, 2026
**Type**: Feature
**Components**: API Definitions, Pulumi CLI Integration, Terraform Module, Protobuf Schemas

## Summary

Implemented the OciNosqlTable deployment component (CloudResourceKind 3335) -- OCI's fully managed, serverless NoSQL database with provisioned or on-demand throughput capacity and DDL-based schema definition. This is the sixth and final database resource in the OCI provider, completing Phase 4 (Databases).

## Problem Statement / Motivation

OCI NoSQL Database Cloud Service provides a serverless, key-value and document database for applications that need low-latency, high-throughput data access with flexible schema. Platform teams need to provision tables with configurable throughput capacity, define schemas via DDL, and create secondary indexes for query performance.

### Pain Points

- Teams need managed NoSQL databases without capacity planning complexity -- on-demand mode scales automatically
- Provisioned mode is needed for cost-predictable workloads with known throughput requirements
- Table schema must be defined via DDL statements, which is OCI's native schema mechanism for NoSQL
- Secondary indexes on specific columns and JSON fields are needed for efficient query patterns
- JSON field indexing requires specifying the field type and path within the JSON document

## Solution / What's New

A complete deployment component with:

1. **Proto API** -- `spec.proto` with 6 top-level fields, 3 nested messages (TableLimits, Index, IndexKey), 1 embedded enum (CapacityMode), 1 CEL conditional validation rule
2. **Validation Tests** -- 25 Ginkgo/Gomega tests (10 valid, 15 invalid scenarios), all passing
3. **Pulumi Module** -- `nosql.NewTable()` with TableLimits and conditional CapacityMode, plus `nosql.NewIndex()` loop for bundled secondary indexes across 4 Go files
4. **Terraform Module** -- `oci_nosql_table.this` with table_limits block and capacity_mode enum map, plus `oci_nosql_index.this` with `for_each` iteration and dynamic `keys` blocks
5. **Kind Registration** -- OciNosqlTable=3335 under "Databases" section, kind_map_gen.go regenerated

### Design Decision: DDL Pass-Through

Table schema is defined via a `ddl_statement` string field (e.g. `CREATE TABLE users (id INTEGER, name STRING, PRIMARY KEY(id))`). This is OCI NoSQL's native schema definition mechanism. Modeling columns as proto messages would severely limit expressiveness (JSON types, UUID columns, identity columns, nested types, TTL) and add enormous complexity without meaningful benefit. The DDL is the schema.

### Design Decision: Bundled Indexes

Secondary indexes (`oci_nosql_index`) are bundled as `repeated Index` in the spec. Indexes are tightly coupled to the table, immutable (any change forces recreation), and follow the established bundling pattern from NSG security rules, LB backend sets, and DRG sub-resources. Index names serve as natural keys for `for_each` / Pulumi resource naming.

### Design Decision: CapacityMode Enum with CEL

Two capacity modes: PROVISIONED (user-specified throughput) and ON_DEMAND (auto-scaling where read/write units are ignored). A CEL rule enforces that `max_read_units` and `max_write_units` must be greater than zero when capacity_mode is provisioned or unspecified (default).

### Design Decision: No Networking Dependency

Unlike most OCI resources, NoSQL is fully managed and serverless -- it only requires a `compartment_id`. No VCN, subnet, or NSG configuration is needed. This makes it the simplest database component in terms of infrastructure dependencies.

## Implementation Details

### Schema via DDL

Both `name` and `ddl_statement` are required fields. The `name` identifies the table for API operations; the `ddl_statement` defines the schema and contains the table name. OCI validates consistency between the two. For schema evolution, `ddl_statement` is updatable (ALTER TABLE), though column order cannot change and new columns can only be appended.

### Index Sub-Resources

Each index has a name, and one or more key columns. For JSON columns, `json_field_type` and `json_path` specify the indexed field within the JSON document. In Pulumi, indexes are created with `pulumi.DependsOn` on the table. In Terraform, `for_each` iterates over the indexes with `table_name_or_id` referencing the table OCID.

### CEL Validation

One rule enforces that provisioned throughput units are specified when not using on-demand mode. The expression checks `this.table_limits.capacity_mode == 2` (on_demand) as the exemption condition.

### Excluded from v1

- `defined_tags`, `system_tags` -- managed by platform via freeform_tags
- `freeform_tags` -- auto-populated from metadata labels (standard pattern)
- Multi-region replicas -- advanced replication feature
- Child table support (names containing ".") -- advanced pattern, separate deployments
- `is_if_not_exists` on indexes -- operational idempotency flag handled by IaC

## Benefits

- **Serverless NoSQL** -- Fully managed with zero infrastructure to provision beyond the table itself
- **Flexible capacity** -- Choose between provisioned throughput for cost predictability or on-demand for automatic scaling
- **DDL-native schema** -- Full expressiveness of OCI NoSQL DDL including JSON, UUID, identity columns, and nested types
- **Secondary indexes** -- Bundled index management including JSON field indexing for complex query patterns
- **Feature parity** -- Identical capabilities in both Pulumi and Terraform modules

## Impact

- Sixth and final Phase 4 (Databases) resource completed -- Phase 4 is now 100% complete
- Next component: R21 OciObjectStorageBucket (Phase 5: Storage)
- 20 of 37 OCI resource kinds now implemented

## Related Work

- R15 OciAutonomousDatabase (2026-02-19) -- serverless Oracle database
- R16 OciDbSystem (2026-02-19) -- traditional Oracle on VM/BM
- R17 OciMysqlDbSystem (2026-02-19) -- managed MySQL HeatWave
- R18 OciPostgresqlDbSystem (2026-02-19) -- managed PostgreSQL
- R19 OciRedisCluster (2026-02-19) -- managed Redis cache
- R21 OciObjectStorageBucket (next) -- object storage with lifecycle policies

---

**Status**: Production Ready
**Validation**: go build clean, go vet clean, 25/25 tests passed, terraform validate success
