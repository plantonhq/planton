# OCI NoSQL Table

Deploys an Oracle Cloud Infrastructure NoSQL table with DDL-defined schema, configurable throughput capacity (provisioned or on-demand), and optional secondary indexes. Freeform tags are auto-populated from metadata labels.

## What Gets Created

When you deploy an OciNosqlTable resource, Planton provisions:

- **NoSQL Table** — an `oci_nosql_table` resource in the specified compartment. The table schema is defined entirely through a DDL statement (`CREATE TABLE` or `ALTER TABLE`), which is OCI NoSQL's native schema mechanism. Freeform tags are automatically derived from metadata labels.
- **Secondary Indexes** — zero or more `oci_nosql_index` resources, one per entry in the `indexes` list. Each index is immutable; any change to an existing index forces its recreation. Indexes support plain columns and JSON field paths within JSON-typed columns.

## Prerequisites

- **OCI credentials** configured via environment variables or Planton provider config (API Key, Instance Principal, Security Token, Resource Principal, or OKE Workload Identity)
- **A compartment OCID** where the NoSQL table will be created — either a literal value or a reference to an OciCompartment resource
- **A valid DDL statement** — `CREATE TABLE` for new tables; `ALTER TABLE` for schema evolution. The table name in the DDL must match the `name` field.

## Quick Start

Create a file `nosql-table.yaml`:

```yaml
apiVersion: oci.planton.dev/v1
kind: OciNosqlTable
metadata:
  name: my-nosql-table
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.OciNosqlTable.my-nosql-table
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  name: users
  ddlStatement: >-
    CREATE TABLE users (
      id STRING,
      data JSON,
      PRIMARY KEY(id)
    )
  tableLimits:
    maxReadUnits: 50
    maxWriteUnits: 50
    maxStorageInGbs: 10
```

Deploy:

```shell
planton apply -f nosql-table.yaml
```

This creates a NoSQL table with provisioned throughput (50 read units, 50 write units) and 10 GB of storage. The table OCID is exported as a stack output.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `compartmentId` | `StringValueOrRef` | OCID of the compartment where the NoSQL table will be created. Can reference an OciCompartment resource via `valueFrom`. | Required |
| `name` | `string` | Table name. Must match the table name used in `ddlStatement`. Changing this forces recreation. | Minimum length 1 |
| `ddlStatement` | `string` | DDL statement defining the table schema. Use `CREATE TABLE` for new tables or `ALTER TABLE` for schema evolution. | Minimum length 1 |
| `tableLimits` | `TableLimits` | Throughput and storage limits for the table. See sub-table below. | Required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `isAutoReclaimable` | `bool` | `false` | When `true`, the table can be automatically reclaimed by OCI after an idle period. Changing this forces recreation. |
| `indexes` | `Index[]` | `[]` | Secondary indexes on the table. Each index is immutable; any change requires recreation. See sub-table below. |

### TableLimits

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `capacityMode` | `CapacityMode` | `provisioned` | How throughput is allocated. Allowed values: `provisioned`, `on_demand`. When `on_demand`, read and write units are ignored. |
| `maxReadUnits` | `int32` | — | Maximum sustained read throughput limit. Required when `capacityMode` is `provisioned` or unspecified. |
| `maxWriteUnits` | `int32` | — | Maximum sustained write throughput limit. Required when `capacityMode` is `provisioned` or unspecified. |
| `maxStorageInGbs` | `int32` | — | Maximum storage in GB that the table can use. | Must be >= 1. |

**Cross-field validation:** When `capacityMode` is `provisioned` (or unspecified), both `maxReadUnits` and `maxWriteUnits` must be greater than zero.

### Index

| Field | Type | Description |
|-------|------|-------------|
| `name` | `string` | Index name. Used as the resource key in the IaC module. Minimum length 1. |
| `keys` | `IndexKey[]` | Columns included in the index. Minimum 1 item. |

### IndexKey

| Field | Type | Description |
|-------|------|-------------|
| `columnName` | `string` | Name of the column to index. Required. |
| `jsonFieldType` | `string` | If the column is of type JSON, the scalar type of the JSON field being indexed (e.g., `"STRING"`, `"INTEGER"`, `"NUMBER"`). Optional. |
| `jsonPath` | `string` | If the column is of type JSON, the dot-separated path to the field being indexed (e.g., `"address.zipCode"`). Optional. |

## Examples

### Minimal Provisioned Table

A simple key-value table with provisioned throughput — suitable for development or low-traffic workloads:

```yaml
apiVersion: oci.planton.dev/v1
kind: OciNosqlTable
metadata:
  name: dev-kv
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.OciNosqlTable.dev-kv
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  name: kv_store
  ddlStatement: >-
    CREATE TABLE kv_store (
      key STRING,
      value JSON,
      PRIMARY KEY(key)
    )
  tableLimits:
    maxReadUnits: 50
    maxWriteUnits: 50
    maxStorageInGbs: 10
```

### On-Demand Throughput with Indexes

A table with on-demand capacity and secondary indexes for query flexibility:

```yaml
apiVersion: oci.planton.dev/v1
kind: OciNosqlTable
metadata:
  name: events-table
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: staging.OciNosqlTable.events-table
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  name: events
  ddlStatement: >-
    CREATE TABLE events (
      event_id STRING,
      source STRING,
      created_at TIMESTAMP(3),
      payload JSON,
      PRIMARY KEY(event_id)
    )
  tableLimits:
    capacityMode: on_demand
    maxStorageInGbs: 100
  indexes:
    - name: idx_source
      keys:
        - columnName: source
    - name: idx_created_at
      keys:
        - columnName: created_at
```

### JSON Field Indexing

A table that indexes specific fields within a JSON column:

```yaml
apiVersion: oci.planton.dev/v1
kind: OciNosqlTable
metadata:
  name: orders-table
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.OciNosqlTable.orders-table
  env: prod
  org: acme
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  name: orders
  ddlStatement: >-
    CREATE TABLE orders (
      order_id STRING,
      customer_id STRING,
      details JSON,
      created_at TIMESTAMP(3),
      PRIMARY KEY(SHARD(customer_id), order_id)
    )
  tableLimits:
    capacityMode: provisioned
    maxReadUnits: 200
    maxWriteUnits: 100
    maxStorageInGbs: 50
  indexes:
    - name: idx_status
      keys:
        - columnName: details
          jsonFieldType: STRING
          jsonPath: status
    - name: idx_created_at
      keys:
        - columnName: created_at

```

### Using Foreign Key References

Reference an Planton-managed compartment instead of hardcoding the OCID:

```yaml
apiVersion: oci.planton.dev/v1
kind: OciNosqlTable
metadata:
  name: ref-nosql
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.OciNosqlTable.ref-nosql
spec:
  compartmentId:
    valueFrom:
      kind: OciCompartment
      name: prod-compartment
      fieldPath: status.outputs.compartmentId
  name: sessions
  ddlStatement: >-
    CREATE TABLE sessions (
      session_id STRING,
      user_id STRING,
      expires_at TIMESTAMP(3),
      PRIMARY KEY(session_id)
    )
  tableLimits:
    maxReadUnits: 100
    maxWriteUnits: 100
    maxStorageInGbs: 25
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `tableId` | `string` | OCID of the created NoSQL table |

## Related Components

- [OciCompartment](/docs/catalog/oci/ocicompartment) — provides the compartment referenced by `compartmentId` via `valueFrom`
