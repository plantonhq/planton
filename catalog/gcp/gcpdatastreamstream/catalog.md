# GCP Datastream Stream

Replicates an operational database into BigQuery or Cloud Storage continuously, without pipeline code: a backfill of what exists, then every insert, update, and delete, from MySQL, PostgreSQL, Oracle, SQL Server, MongoDB, Salesforce, or Spanner.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **Stream** -- a `datastream_stream` with one source arm, one destination arm, a backfill mode, and optional per-table rules

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with Datastream admin permissions (`roles/datastream.admin`) on the project. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Required Dependencies

- **`GcpDatastreamConnectionProfile`** -- a source profile and a destination profile in the stream's region.

## Deploy

### Console

Open the deployment store, find **GCP Datastream Stream**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **PostgreSQL to BigQuery (Merge)** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpDatastreamStream
metadata:
  name: orders-cdc
  org: acme-corp
  env: prod
spec:
  location: us-central1
  sourceConfig:
    sourceConnectionProfile:
      valueFrom:
        kind: GcpDatastreamConnectionProfile
        name: orders-postgres
        fieldPath: status.outputs.name
    postgresqlSourceConfig:
      replicationSlot: datastream_slot
      publication: datastream_pub
  destinationConfig:
    destinationConnectionProfile:
      valueFrom:
        kind: GcpDatastreamConnectionProfile
        name: bigquery
        fieldPath: status.outputs.name
    bigqueryDestinationConfig:
      sourceHierarchyDatasets:
        datasetTemplate:
          location: US
  backfillAll: {}
```

```shell
planton apply -f datastream-stream.yaml
```

This creates the stream `NOT_STARTED`; set `desiredState: RUNNING` to start it. A Stack Job tracks the provisioning in real time.

### InfraChart

Reference the source and destination `GcpDatastreamConnectionProfile`s by name; a `GcpBigQueryDataset` or `GcpKmsKey` in the same chart wires in by reference.

## Key Configuration

These are the most important decisions when configuring this component. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Backfill** -- `backfillAll` copies what exists first (skipping excluded objects); `backfillNone` replicates only changes from the start.

**Write mode** -- `MERGE` mirrors the source's current state; `APPEND_ONLY` keeps every change as history. Immutable.

**Datasets** -- one dataset for everything, or one per source schema created by Datastream.

**Desired state** -- create `NOT_STARTED`, review, then `RUNNING`.

## Outputs and Dependencies

### What This Component Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** | `projectId`, `sourceHierarchyDatasets.projectId` | `status.outputs.project_id` |
| **GcpDatastreamConnectionProfile** | `sourceConfig.sourceConnectionProfile`, `destinationConfig.destinationConnectionProfile` | `status.outputs.name` |
| **GcpBigQueryDataset** | `singleTargetDataset.datasetId` | `status.outputs.self_link` |
| **GcpKmsKey** | `customerManagedEncryptionKey`, `datasetTemplate.kmsKeyName` | `status.outputs.key_id` |
| **GcpGcsBucket** | `blmtConfig.bucket` | `status.outputs.bucket_name` |
| **GcpBigQueryConnection** | `blmtConfig.connectionName` | `status.outputs.name` |

### What This Component Provides

After provisioning, `status.outputs` contains values that downstream Cloud Resources can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|-----------------------|
| `name` | The stream's full resource name | Monitoring and alerting on the stream |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Mirror to BigQuery** -- PostgreSQL merged into per-schema datasets. Start from the **PostgreSQL to BigQuery (Merge)** preset.

**Land in a lake** -- MySQL changes as Avro files in Cloud Storage. Start from the **MySQL to Cloud Storage (Avro)** preset.

**Keep history** -- Every change appended, partitioned by day, clustered by customer. Start from the **Append-Only History** preset.

## Works With

- [**GCP Datastream Connection Profile**](/cloud-catalog/gcp-datastream-connection-profile) -- source and destination
- [**GCP Datastream Private Connection**](/cloud-catalog/gcp-datastream-private-connection) -- private reachability
- [**GCP BigQuery Dataset**](/cloud-catalog/gcp-bigquery-dataset) -- where tables land
- [**GCP KMS Key**](/cloud-catalog/gcp-kms-key) -- encryption
