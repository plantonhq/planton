# OciNosqlTable Examples

## Table of Contents

- [Minimal Key-Value Table](#minimal-key-value-table)
- [On-Demand Throughput Table](#on-demand-throughput-table)
- [Table with Secondary Indexes](#table-with-secondary-indexes)
- [JSON Field Indexing](#json-field-indexing)
- [Foreign Key Compartment Reference](#foreign-key-compartment-reference)
- [Common Operations](#common-operations)
- [Best Practices](#best-practices)

---

## Minimal Key-Value Table

A provisioned-throughput table with a single string primary key and a JSON value column. Suitable for development or caching use cases.

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciNosqlTable
metadata:
  name: dev-cache
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.OciNosqlTable.dev-cache
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  name: cache
  ddlStatement: >-
    CREATE TABLE cache (
      key STRING,
      value JSON,
      PRIMARY KEY(key)
    )
  tableLimits:
    maxReadUnits: 25
    maxWriteUnits: 25
    maxStorageInGbs: 5
```

## On-Demand Throughput Table

An on-demand table where OCI manages throughput scaling automatically. Read and write units are not specified. Useful for workloads with unpredictable traffic patterns.

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciNosqlTable
metadata:
  name: audit-log
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.OciNosqlTable.audit-log
  env: prod
  org: acme
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  name: audit_log
  ddlStatement: >-
    CREATE TABLE audit_log (
      log_id STRING,
      actor STRING,
      action STRING,
      timestamp TIMESTAMP(3),
      details JSON,
      PRIMARY KEY(log_id)
    )
  tableLimits:
    capacityMode: on_demand
    maxStorageInGbs: 200
```

## Table with Secondary Indexes

A table with multiple secondary indexes for different query patterns. Each index is an independent OCI resource and is immutable after creation.

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciNosqlTable
metadata:
  name: user-profiles
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: staging.OciNosqlTable.user-profiles
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  name: user_profiles
  ddlStatement: >-
    CREATE TABLE user_profiles (
      user_id STRING,
      email STRING,
      display_name STRING,
      created_at TIMESTAMP(3),
      profile JSON,
      PRIMARY KEY(user_id)
    )
  tableLimits:
    maxReadUnits: 100
    maxWriteUnits: 50
    maxStorageInGbs: 25
  indexes:
    - name: idx_email
      keys:
        - columnName: email
    - name: idx_created_at
      keys:
        - columnName: created_at
```

## JSON Field Indexing

A table that indexes specific scalar fields within a JSON-typed column. The `jsonFieldType` and `jsonPath` fields on the index key identify which nested field to index and its type.

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciNosqlTable
metadata:
  name: orders
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.OciNosqlTable.orders
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
    - name: idx_order_status
      keys:
        - columnName: details
          jsonFieldType: STRING
          jsonPath: status
    - name: idx_order_total
      keys:
        - columnName: details
          jsonFieldType: NUMBER
          jsonPath: total
    - name: idx_created_at
      keys:
        - columnName: created_at
```

## Foreign Key Compartment Reference

Reference an OpenMCF-managed OciCompartment instead of hardcoding the OCID. The `valueFrom` field resolves the compartment OCID at deploy time.

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciNosqlTable
metadata:
  name: sessions
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.OciNosqlTable.sessions
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
      data JSON,
      PRIMARY KEY(session_id)
    )
  tableLimits:
    maxReadUnits: 100
    maxWriteUnits: 100
    maxStorageInGbs: 25
  isAutoReclaimable: true
```

---

## Common Operations

### Evolving a Table Schema

To add a column to an existing table, update the `ddlStatement` to an `ALTER TABLE` statement:

```yaml
  ddlStatement: >-
    ALTER TABLE user_profiles (ADD last_login TIMESTAMP(3))
```

After the schema change is applied, revert the `ddlStatement` back to the full `CREATE TABLE` form with the new column included. This ensures the manifest remains the authoritative schema definition.

### Switching from Provisioned to On-Demand

Update the `tableLimits` section. When switching to `on_demand`, the `maxReadUnits` and `maxWriteUnits` fields are ignored:

```yaml
  tableLimits:
    capacityMode: on_demand
    maxStorageInGbs: 50
```

### Adding a New Index

Append a new entry to the `indexes` list. Existing indexes are not affected:

```yaml
  indexes:
    - name: idx_email
      keys:
        - columnName: email
    - name: idx_display_name    # new index
      keys:
        - columnName: display_name
```

---

## Best Practices

1. **Keep `name` and DDL in sync** — The `name` field and the table name in `ddlStatement` must match exactly. A mismatch can cause deployment errors.
2. **Use `CREATE TABLE` as the source of truth** — After applying `ALTER TABLE` for schema evolution, update the manifest back to a `CREATE TABLE` statement that reflects the full current schema. This keeps the manifest self-documenting.
3. **Plan indexes early** — Adding indexes to large tables is an online operation but can take time. Define indexes when the table is first created when possible.
4. **Set storage limits with headroom** — `maxStorageInGbs` is a hard cap. Set it above current needs and monitor usage to avoid hitting the limit unexpectedly.
5. **Use on-demand for variable workloads** — If traffic is unpredictable, `on_demand` capacity avoids throttling. For steady-state workloads, `provisioned` mode is more cost-effective.
6. **Leverage shard keys for write distribution** — Use `PRIMARY KEY(SHARD(partition_key), sort_key)` in the DDL to distribute writes across storage nodes for high-ingestion tables.
