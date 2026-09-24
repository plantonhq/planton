# GCP Datastream Stream

Continuous change data capture from one source database into BigQuery or Cloud Storage. A stream reads through a source `GcpDatastreamConnectionProfile` (MySQL, PostgreSQL, Oracle, SQL Server, MongoDB, Salesforce, or Spanner), writes through a destination profile, backfills what exists (or not), and then carries every insert, update, and delete -- merged into BigQuery tables that mirror the source, appended as history, or written as Avro or JSON files.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **Stream** -- a `datastream_stream` with one source arm, one destination arm, a backfill mode, and optional per-table rules

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with Datastream admin permissions (`roles/datastream.admin`) on the project.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Required Dependencies

- **`GcpDatastreamConnectionProfile`** -- a source profile and a destination profile in the stream's region.

### Optional Dependencies

- **`GcpBigQueryDataset`** -- a single target dataset (`singleTargetDataset.datasetId`, its `self_link`).
- **`GcpKmsKey`** -- CMEK for the stream and the datasets it creates.
- **`GcpBigQueryConnection`** / **`GcpGcsBucket`** -- BigLake managed (Iceberg) tables.

## Deploy

### CLI

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpDatastreamStream
metadata:
  name: orders-cdc
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

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `location` | `string` | The Datastream region, e.g. `us-central1`. Immutable. |
| `sourceConfig` | `object` | `sourceConnectionProfile` and exactly one source arm. |
| `destinationConfig` | `object` | `destinationConnectionProfile` and exactly one of `bigqueryDestinationConfig` or `gcsDestinationConfig`. |
| backfill mode | -- | Exactly one of `backfillAll` (optionally with exclusions) or `backfillNone: true`. |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `projectId` | `StringValueOrRef` | provider project | The project. |
| `streamId` | `string` | `metadata.name` | Immutable. |
| `displayName` | `string` | `metadata.name` | Console name. |
| `sourceConfig.*SourceConfig` | `object` | -- | Include and exclude object lists; concurrency; MySQL and SQL Server `cdcMethod`; Oracle `largeObjectsHandling`; PostgreSQL publication and slot; Salesforce `pollingInterval`; Spanner change stream, Data Boost, role, priority. |
| `destinationConfig.bigqueryDestinationConfig` | `object` | -- | `singleTargetDataset` or `sourceHierarchyDatasets`, `dataFreshness`, `writeMode` (`MERGE`/`APPEND_ONLY`), `blmtConfig`. |
| `destinationConfig.gcsDestinationConfig` | `object` | -- | `path`, rotation interval and size, `avroFileFormat` or `jsonFileFormat`. |
| `ruleSets` | `list` | none | Per-object BigQuery partitioning or clustering. |
| `customerManagedEncryptionKey` | `StringValueOrRef` | Google-managed | CMEK for data in flight. Immutable. |
| `desiredState` | `string` | `NOT_STARTED` | `NOT_STARTED`, `RUNNING`, or `PAUSED`. |
| `createWithoutValidation` | `bool` | `false` | Skip Google's validation. Immutable. |
| `labels` | `map<string,string>` | none | Merged with the platform attribution labels. |
| `deletionPolicy` | `string` | `DELETE` | `DELETE`, `PREVENT`, or `ABANDON`. |

### Validation Rules

- Exactly one source arm, one destination arm, one BigQuery dataset config, one GCS file format, and one backfill mode.
- Object lists name their first level; MongoDB and Spanner backfill exclusions name every level.
- Rule sets: exactly one object identifier, at least one rule, each rule exactly one of partitioning or clustering, each partitioning exactly one kind.
- Values follow Google's lists and ranges: MongoDB backfill concurrency 0-50, GCS rotation 15-60 seconds, durations in seconds.

## Stack Outputs

| Output | Type | Description |
|--------|------|-------------|
| `name` | `string` | `projects/{project}/locations/{location}/streams/{id}` |
| `stream_id` | `string` | The stream's id |

## Deployment Methods

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md) for Pulumi-specific deployment instructions.

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md) for Terraform-specific deployment instructions.

## Important Notes

- **Streams start `NOT_STARTED`.** Nothing moves or bills until `desiredState: RUNNING`; `PAUSED` stops a running stream without losing its position, and a started stream never returns to `NOT_STARTED`.
- **Datastream bills the GiB it processes** -- the backfill and the changes at different rates -- so narrow `includeObjects` to what you need.
- **Write mode is immutable.** `MERGE` mirrors current state (tables need a primary key); `APPEND_ONLY` keeps every change.
- **Grants live on the destination.** Datastream's service agent writes BigQuery and Cloud Storage and uses KMS keys; grant it there.

## Examples

For a complete example, see `e2e/manifest.yaml`. Scenario variants live under `e2e/scenarios/`.

## Related Components

- **GcpDatastreamConnectionProfile** -- the source and destination
- **GcpDatastreamPrivateConnection** -- private reachability for sources
- **GcpBigQueryDataset** -- where tables land

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
