# OciNosqlTable

## Overview

OciNosqlTable is an Planton component that deploys and manages Oracle Cloud Infrastructure NoSQL tables. It provides a declarative YAML interface over the OCI NoSQL Database service, handling table creation, throughput configuration, and secondary index management through a single resource manifest.

## Purpose

OCI NoSQL Database is a fully managed, serverless database service. This component abstracts the infrastructure provisioning so that teams can define NoSQL tables — including schema, capacity, and indexes — in version-controlled YAML manifests. The Pulumi IaC module translates the manifest into the underlying `oci_nosql_table` and `oci_nosql_index` resources.

## Key Features

- **DDL-based schema definition** — Table schemas are defined using OCI NoSQL's native DDL syntax (`CREATE TABLE`, `ALTER TABLE`), giving full access to all supported column types: INTEGER, LONG, FLOAT, DOUBLE, NUMBER, STRING, BOOLEAN, BINARY, TIMESTAMP, JSON, ENUM, ARRAY, MAP, RECORD, and UUID.
- **Provisioned or on-demand throughput** — Choose between `provisioned` mode with explicit read/write unit limits, or `on_demand` mode where OCI scales throughput automatically.
- **Secondary index management** — Declare indexes as sub-resources within the same manifest. Each index supports plain column keys and JSON field path keys for indexing specific fields within JSON-typed columns.
- **Foreign key references** — The `compartmentId` field accepts either a literal OCID or a `valueFrom` reference to an Planton-managed OciCompartment, enabling cross-resource composition.
- **Automatic freeform tagging** — Metadata labels, organization, and environment are propagated as OCI freeform tags on the table, providing consistent resource attribution without manual tag configuration.

## Critical Constraints

- **DDL/name consistency** — The `name` field in the spec must exactly match the table name in the `ddlStatement`. The IaC module passes both independently to OCI.
- **Index immutability** — Secondary indexes cannot be updated in place. Any change to an index definition (columns, JSON paths) forces OCI to drop and recreate the index.
- **Table name change forces recreation** — Changing the `name` field destroys the existing table and creates a new one. Data is not migrated.
- **Provisioned throughput validation** — When `capacityMode` is `provisioned` (or unspecified), both `maxReadUnits` and `maxWriteUnits` must be greater than zero. This is enforced by a CEL validation rule on the proto.
- **No child table support** — Table names containing dots (`.`) are not supported in v1. Child tables must be deployed as separate resources.
- **No multi-region replicas** — Cross-region replication is excluded from v1.
- **No defined_tags or system_tags** — OCI defined tags and system tags are not exposed. Freeform tags are the only tagging mechanism, auto-populated from metadata.

## Use Cases

- **Session stores** — Low-latency key-value tables with TTL-based cleanup via `isAutoReclaimable`.
- **Event logs** — Append-heavy tables with on-demand throughput and timestamp-based secondary indexes.
- **IoT telemetry** — High-ingestion tables with shard keys for write distribution and JSON payload indexing.
- **Catalog or configuration data** — Provisioned-throughput tables with predictable read patterns and JSON field indexes for flexible querying.

## Production Considerations

- **Capacity planning** — Provisioned mode is cost-effective for predictable workloads. On-demand mode avoids throttling for variable traffic but costs more per operation at sustained high throughput.
- **Storage limits** — `maxStorageInGbs` is a hard limit. Monitor table size and increase the limit before it is reached.
- **Schema evolution** — Use `ALTER TABLE` DDL statements to add columns. Column removal and type changes are restricted by OCI NoSQL. Always append new columns; do not reorder existing ones.
- **Index planning** — Plan indexes at table design time. Adding an index to a large table is an online operation but can take time. Removing or changing an index requires recreation.
- **Tag propagation** — All metadata labels, plus `resource`, `resource_kind`, `resource_id`, `organization`, and `environment` tags, are applied as freeform tags. OCI limits freeform tags to 10 per resource by default; ensure label count stays within this limit.
