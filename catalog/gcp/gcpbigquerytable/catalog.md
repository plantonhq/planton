# GCP BigQuery Table

Deploys a BigQuery table inside an existing dataset — a native table, a logical view, a materialized view, or an external/BigLake table, all arms of the same resource. Supports time and integer-range partitioning, clustering, unenforced constraints, CMEK encryption, and deletion protection, with ValueFromRef wiring to the parent GcpBigQueryDataset, GCP projects, KMS keys, and other BigQuery tables.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **BigQuery API enablement** (`bigquery.googleapis.com`) on the target project (never disabled on destroy)
- **BigQuery Table** -- one of four shapes: a native table (BigQuery-managed storage, optionally partitioned and clustered), a logical view (a saved query evaluated at read time), a materialized view (precomputed, incrementally refreshed results), or an external table (data stays in GCS/Sheets/Bigtable and is read at query time)
- **Partitioning & Clustering** -- time-based or integer-range partitioning plus up to four clustering columns, the primary scan-cost levers for large tables
- **Table Constraints** -- when declared, unenforced primary/foreign keys the query optimizer and lineage tools consume
- **CMEK Encryption** -- created only when `kmsKeyName` is set; overrides the dataset's default key for this one table
- **GCP Labels** -- resource metadata labels (resource name, kind, organization, environment) applied automatically for tracking and governance

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with credentials for the target GCP project.
- **Planton Runner** -- required when using Runner-based credential delivery.

### GCP Project

- **A BigQuery dataset** the table will live in. Reference a GcpBigQueryDataset Infra Component via ValueFromRef or provide the dataset ID directly. The dataset pins the location and supplies encryption/expiration defaults.
- **Cloud KMS key** (if using per-table CMEK) -- the BigQuery service agent must have `roles/cloudkms.cryptoKeyEncrypterDecrypter` on the key.

## Deploy

### Console

Open the deployment store, find **GCP BigQuery Table**, and click **Deploy**. The creation wizard forks on the table kind (native / view / materialized view / external) and walks you through partitioning, clustering, and protection. Start from the **Partitioned Analytics Table** preset in the [Presets](#presets) tab for the standard fact-table shape.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpBigQueryTable
metadata:
  name: events-raw
  org: acme-corp
  env: prod
spec:
  datasetId:
    value: analytics_events_prod
  tableId: events_raw
  timePartitioning:
    type: DAY
    field: event_ts
  clustering:
    - customer_id
  requirePartitionFilter: true
```

```shell
planton apply -f bigquery-table.yaml
```

This creates a day-partitioned, clustered native table with deletion protection on (the default) — the workhorse shape for event data. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, wire the table to its dataset (and optionally a KMS key) deployed in the same InfraPipeline:

```yaml
spec:
  datasetId:
    valueFrom:
      kind: GcpBigQueryDataset
      name: analytics
      fieldPath: status.outputs.dataset_id
  kmsKeyName:
    valueFrom:
      kind: GcpKmsKey
      name: bigquery-cmek-key
      fieldPath: status.outputs.key_id
```

The InfraPipeline resolves the dependency graph, deploys the dataset (and key) first, then provisions the table.

## Key Configuration

These are the most important decisions when configuring a BigQuery table. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Table kind** -- at most one of `view`, `materializedView`, or `externalDataConfiguration` may be set; none of them is a native table. Converting between kinds recreates the table.

**Partitioning** -- `timePartitioning` (DAY/HOUR/MONTH/YEAR on a DATE/TIMESTAMP column, or ingestion time) XOR `rangePartitioning` (an INTEGER key space in fixed steps). The partitioning field is immutable. Pair with `requirePartitionFilter` to make full scans impossible.

**Clustering** -- up to four columns in precedence order; queries filtering on the leading columns scan less data. Mutable in place.

**Deletion protection** -- defaults TRUE on both engines: a destroy fails until it is set false. The guard for tables holding real data.

**Schema evolution** -- native-table columns are add-only in place; removing or retyping a column recreates the table.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpBigQueryDataset** | `datasetId` | `status.outputs.dataset_id` |
| **GcpProject** (optional) | `projectId` | `status.outputs.project_id` |
| **GcpKmsKey** (optional) | `kmsKeyName` | `status.outputs.key_id` |
| **GcpBigQueryTable** (optional, per foreign key) | `tableConstraints.foreignKeys[].referencedTable.tableId` | `status.outputs.table_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `table_id` | Short table ID (same as the spec input) | Foreign-key references from other GcpBigQueryTable resources, SQL configuration |
| `self_link` | Fully qualified table URI | IAM bindings, audit log filters |
| `project` | GCP project containing the table | Cross-project references |
| `dataset_id` | Dataset containing the table | Confirming the parent without chasing references |
| `type` | What GCP materialized: TABLE, VIEW, MATERIALIZED_VIEW, or EXTERNAL | Conditional downstream wiring |
| `location` | Geographic location (inherited from the dataset) | Co-locating downstream resources |
| `creation_time` | Creation time in milliseconds since epoch | Auditing and lifecycle tracking |
| `qualified_name` | The dotted `{project}.{dataset}.{table}` handle | Pub/Sub BigQuery delivery, query tooling — no string assembly needed |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Partitioned analytics** -- a day-partitioned, clustered native fact table with an explicit schema and the partition-filter cost guard. Start from the **Partitioned Analytics Table** preset.

**Authorized view** -- a logical view exposing a filtered slice of sensitive data; pair it with the dataset's authorized-view access entry so readers never touch the raw tables. Start from the **Authorized View** preset.

**External GCS table** -- reads hive-partitioned Parquet files in place; no load jobs, always current. Start from the **External GCS Table** preset.

## Works With

- [**GCP BigQuery Dataset**](/infra-catalog/gcp-big-query-dataset) -- the container that pins location and supplies encryption/expiration defaults
- [**GCP Project**](/infra-catalog/gcp-project) -- provides the GCP project when not inherited from the provider connection
- [**GCP KMS Key**](/infra-catalog/gcp-kms-key) -- provides the per-table CMEK key
