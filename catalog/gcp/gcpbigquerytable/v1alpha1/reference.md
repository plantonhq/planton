# GcpBigQueryTable

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

GcpBigQueryTableSpec defines a BigQuery table — a native table, a logical
view, a materialized view, or an external/BigLake table, all arms of the
same resource.

The table lives inside a GcpBigQueryDataset (which pins the location and
supplies encryption/expiration defaults) and is the unit application SQL
actually reads. Model the tables infrastructure should own: partitioned
fact tables, authorized views over sensitive data, external tables over
data-lake files. Application-owned schemas (dbt models, migrations) can
coexist in the same dataset without conflict — the dataset never
enumerates its tables.

Important behavioral notes:

  - table_id, dataset_id, and project are IMMUTABLE, as are the
    encryption key, BigLake configuration, replication info, foreign
    type info, a materialized view's query, and each partitioning
    field. Everything else updates in place.

  - deletion_protection defaults TRUE on both engines: a destroy fails
    until it is set false — the guard for tables holding real data.

  - The four table arms are mutually exclusive: at most one of view,
    materialized_view, or external_data_configuration (none = native
    table). Partitioning/clustering apply to native tables.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpBigQueryTable
metadata:
  name: test-bq-table
spec:
  projectId:
    value: "test-project"
  datasetId:
    value: test_analytics
  tableId: events_raw
  friendlyName: Raw Events
  description: Hack manifest exercising the table surface offline
  labels:
    team: platform
  schema: '[{"name":"event_time","type":"TIMESTAMP","mode":"REQUIRED"},{"name":"customer_id","type":"INT64"},{"name":"payload","type":"JSON"}]'
  timePartitioning:
    type: DAY
    field: event_time
    expirationMs: 7776000000
  clustering:
    - customer_id
  requirePartitionFilter: true
  deletionProtection: false
  # Destroy really destroys in E2E: the live lanes prove the full lifecycle.
  deletionPolicy: DELETE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.datasetId` | `string \| valueFrom` | yes |  | GcpBigQueryDataset (`status.outputs.dataset_id`) |
| `spec.tableId` | `string` | yes |  |  |
| `spec.friendlyName` | `string` |  |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.labels` | `map<string, string>` |  |  |  |
| `spec.resourceTags` | `map<string, string>` |  |  |  |
| `spec.schema` | `string` |  |  |  |
| `spec.timePartitioning` | `GcpBigQueryTableTimePartitioning` |  |  |  |
| `spec.timePartitioning.type` | `string` | yes |  |  |
| `spec.timePartitioning.field` | `string` |  |  |  |
| `spec.timePartitioning.expirationMs` | `int64` |  |  |  |
| `spec.rangePartitioning` | `GcpBigQueryTableRangePartitioning` |  |  |  |
| `spec.rangePartitioning.field` | `string` | yes |  |  |
| `spec.rangePartitioning.range` | `GcpBigQueryTableRangePartitioningRange` | yes |  |  |
| `spec.rangePartitioning.range.start` | `int64` |  |  |  |
| `spec.rangePartitioning.range.end` | `int64` |  |  |  |
| `spec.rangePartitioning.range.interval` | `int64` |  |  |  |
| `spec.clustering` | `[]string` |  |  |  |
| `spec.requirePartitionFilter` | `bool` |  |  |  |
| `spec.expirationTime` | `int64` |  |  |  |
| `spec.maxStaleness` | `string` |  |  |  |
| `spec.kmsKeyName` | `string \| valueFrom` |  |  | GcpKmsKey (`status.outputs.key_id`) |
| `spec.view` | `GcpBigQueryTableView` |  |  |  |
| `spec.view.query` | `string` | yes |  |  |
| `spec.view.useLegacySql` | `bool` |  |  |  |
| `spec.materializedView` | `GcpBigQueryTableMaterializedView` |  |  |  |
| `spec.materializedView.query` | `string` | yes |  |  |
| `spec.materializedView.enableRefresh` | `bool` |  |  |  |
| `spec.materializedView.refreshIntervalMs` | `int64` |  |  |  |
| `spec.materializedView.allowNonIncrementalDefinition` | `bool` |  |  |  |
| `spec.externalDataConfiguration` | `GcpBigQueryTableExternalDataConfiguration` |  |  |  |
| `spec.externalDataConfiguration.autodetect` | `bool` |  |  |  |
| `spec.externalDataConfiguration.sourceUris` | `[]string` | yes |  |  |
| `spec.externalDataConfiguration.sourceFormat` | `string` |  |  |  |
| `spec.externalDataConfiguration.objectMetadata` | `string` |  |  |  |
| `spec.externalDataConfiguration.compression` | `string` |  |  |  |
| `spec.externalDataConfiguration.schema` | `string` |  |  |  |
| `spec.externalDataConfiguration.ignoreUnknownValues` | `bool` |  |  |  |
| `spec.externalDataConfiguration.maxBadRecords` | `int32` |  |  |  |
| `spec.externalDataConfiguration.connectionId` | `string` |  |  |  |
| `spec.externalDataConfiguration.referenceFileSchemaUri` | `string` |  |  |  |
| `spec.externalDataConfiguration.metadataCacheMode` | `string` |  |  |  |
| `spec.externalDataConfiguration.fileSetSpecType` | `string` |  |  |  |
| `spec.externalDataConfiguration.jsonExtension` | `string` |  |  |  |
| `spec.externalDataConfiguration.csvOptions` | `GcpBigQueryTableCsvOptions` |  |  |  |
| `spec.externalDataConfiguration.csvOptions.quote` | `string` |  |  |  |
| `spec.externalDataConfiguration.csvOptions.allowJaggedRows` | `bool` |  |  |  |
| `spec.externalDataConfiguration.csvOptions.allowQuotedNewlines` | `bool` |  |  |  |
| `spec.externalDataConfiguration.csvOptions.encoding` | `string` |  |  |  |
| `spec.externalDataConfiguration.csvOptions.fieldDelimiter` | `string` |  |  |  |
| `spec.externalDataConfiguration.csvOptions.skipLeadingRows` | `int32` |  |  |  |
| `spec.externalDataConfiguration.csvOptions.sourceColumnMatch` | `string` |  |  |  |
| `spec.externalDataConfiguration.jsonOptions` | `GcpBigQueryTableJsonOptions` |  |  |  |
| `spec.externalDataConfiguration.jsonOptions.encoding` | `string` |  |  |  |
| `spec.externalDataConfiguration.googleSheetsOptions` | `GcpBigQueryTableGoogleSheetsOptions` |  |  |  |
| `spec.externalDataConfiguration.googleSheetsOptions.range` | `string` |  |  |  |
| `spec.externalDataConfiguration.googleSheetsOptions.skipLeadingRows` | `int32` |  |  |  |
| `spec.externalDataConfiguration.hivePartitioningOptions` | `GcpBigQueryTableHivePartitioningOptions` |  |  |  |
| `spec.externalDataConfiguration.hivePartitioningOptions.mode` | `string` |  |  |  |
| `spec.externalDataConfiguration.hivePartitioningOptions.requirePartitionFilter` | `bool` |  |  |  |
| `spec.externalDataConfiguration.hivePartitioningOptions.sourceUriPrefix` | `string` |  |  |  |
| `spec.externalDataConfiguration.avroOptions` | `GcpBigQueryTableAvroOptions` |  |  |  |
| `spec.externalDataConfiguration.avroOptions.useAvroLogicalTypes` | `bool` |  |  |  |
| `spec.externalDataConfiguration.parquetOptions` | `GcpBigQueryTableParquetOptions` |  |  |  |
| `spec.externalDataConfiguration.parquetOptions.enumAsString` | `bool` |  |  |  |
| `spec.externalDataConfiguration.parquetOptions.enableListInference` | `bool` |  |  |  |
| `spec.externalDataConfiguration.bigtableOptions` | `GcpBigQueryTableBigtableOptions` |  |  |  |
| `spec.externalDataConfiguration.bigtableOptions.ignoreUnspecifiedColumnFamilies` | `bool` |  |  |  |
| `spec.externalDataConfiguration.bigtableOptions.readRowkeyAsString` | `bool` |  |  |  |
| `spec.externalDataConfiguration.bigtableOptions.outputColumnFamiliesAsJson` | `bool` |  |  |  |
| `spec.externalDataConfiguration.bigtableOptions.columnFamilies` | `[]GcpBigQueryTableBigtableColumnFamily` |  |  |  |
| `spec.externalDataConfiguration.bigtableOptions.columnFamilies[].familyId` | `string` |  |  |  |
| `spec.externalDataConfiguration.bigtableOptions.columnFamilies[].type` | `string` |  |  |  |
| `spec.externalDataConfiguration.bigtableOptions.columnFamilies[].encoding` | `string` |  |  |  |
| `spec.externalDataConfiguration.bigtableOptions.columnFamilies[].onlyReadLatest` | `bool` |  |  |  |
| `spec.externalDataConfiguration.bigtableOptions.columnFamilies[].columns` | `[]GcpBigQueryTableBigtableColumn` |  |  |  |
| `spec.externalDataConfiguration.bigtableOptions.columnFamilies[].columns[].qualifierEncoded` | `string` |  |  |  |
| `spec.externalDataConfiguration.bigtableOptions.columnFamilies[].columns[].qualifierString` | `string` |  |  |  |
| `spec.externalDataConfiguration.bigtableOptions.columnFamilies[].columns[].fieldName` | `string` |  |  |  |
| `spec.externalDataConfiguration.bigtableOptions.columnFamilies[].columns[].type` | `string` |  |  |  |
| `spec.externalDataConfiguration.bigtableOptions.columnFamilies[].columns[].encoding` | `string` |  |  |  |
| `spec.externalDataConfiguration.bigtableOptions.columnFamilies[].columns[].onlyReadLatest` | `bool` |  |  |  |
| `spec.externalDataConfiguration.decimalTargetTypes` | `[]string` |  |  |  |
| `spec.tableConstraints` | `GcpBigQueryTableConstraints` |  |  |  |
| `spec.tableConstraints.primaryKey` | `GcpBigQueryTablePrimaryKey` |  |  |  |
| `spec.tableConstraints.primaryKey.columns` | `[]string` | yes |  |  |
| `spec.tableConstraints.foreignKeys` | `[]GcpBigQueryTableForeignKey` |  |  |  |
| `spec.tableConstraints.foreignKeys[].name` | `string` |  |  |  |
| `spec.tableConstraints.foreignKeys[].referencedTable` | `GcpBigQueryTableForeignKeyReferencedTable` | yes |  |  |
| `spec.tableConstraints.foreignKeys[].referencedTable.projectId` | `string` | yes |  |  |
| `spec.tableConstraints.foreignKeys[].referencedTable.datasetId` | `string` | yes |  |  |
| `spec.tableConstraints.foreignKeys[].referencedTable.tableId` | `string \| valueFrom` | yes |  | GcpBigQueryTable (`status.outputs.table_id`) |
| `spec.tableConstraints.foreignKeys[].columnReferences` | `GcpBigQueryTableForeignKeyColumnReferences` | yes |  |  |
| `spec.tableConstraints.foreignKeys[].columnReferences.referencingColumn` | `string` | yes |  |  |
| `spec.tableConstraints.foreignKeys[].columnReferences.referencedColumn` | `string` | yes |  |  |
| `spec.tableReplicationInfo` | `GcpBigQueryTableReplicationInfo` |  |  |  |
| `spec.tableReplicationInfo.sourceProjectId` | `string` | yes |  |  |
| `spec.tableReplicationInfo.sourceDatasetId` | `string` | yes |  |  |
| `spec.tableReplicationInfo.sourceTableId` | `string` | yes |  |  |
| `spec.tableReplicationInfo.replicationIntervalMs` | `int64` |  |  |  |
| `spec.biglakeConfiguration` | `GcpBigQueryTableBiglakeConfiguration` |  |  |  |
| `spec.biglakeConfiguration.connectionId` | `string` | yes |  |  |
| `spec.biglakeConfiguration.storageUri` | `string` | yes |  |  |
| `spec.biglakeConfiguration.fileFormat` | `string` | yes |  |  |
| `spec.biglakeConfiguration.tableFormat` | `string` | yes |  |  |
| `spec.schemaForeignTypeInfo` | `GcpBigQueryTableSchemaForeignTypeInfo` |  |  |  |
| `spec.schemaForeignTypeInfo.typeSystem` | `string` | yes |  |  |
| `spec.externalCatalogTableOptions` | `GcpBigQueryTableExternalCatalogTableOptions` |  |  |  |
| `spec.externalCatalogTableOptions.parameters` | `map<string, string>` |  |  |  |
| `spec.externalCatalogTableOptions.connectionId` | `string` |  |  |  |
| `spec.externalCatalogTableOptions.storageDescriptor` | `GcpBigQueryTableStorageDescriptor` |  |  |  |
| `spec.externalCatalogTableOptions.storageDescriptor.locationUri` | `string` |  |  |  |
| `spec.externalCatalogTableOptions.storageDescriptor.inputFormat` | `string` |  |  |  |
| `spec.externalCatalogTableOptions.storageDescriptor.outputFormat` | `string` |  |  |  |
| `spec.externalCatalogTableOptions.storageDescriptor.serdeInfo` | `GcpBigQueryTableSerDeInfo` |  |  |  |
| `spec.externalCatalogTableOptions.storageDescriptor.serdeInfo.name` | `string` |  |  |  |
| `spec.externalCatalogTableOptions.storageDescriptor.serdeInfo.serializationLibrary` | `string` | yes |  |  |
| `spec.externalCatalogTableOptions.storageDescriptor.serdeInfo.parameters` | `map<string, string>` |  |  |  |
| `spec.deletionProtection` | `bool` |  | `true` |  |
| `spec.deletionPolicy` | `string` |  |  |  |
| `spec.ignoreAutoGeneratedSchema` | `bool` |  |  |  |
| `spec.ignoreSchemaChanges` | `[]string` |  |  |  |
| `spec.tableMetadataView` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

The GCP project the table is created in. Accepts a literal project ID
or a reference to a GcpProject resource. If omitted, the provider's
default project is used.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.datasetId

`string | valueFrom` · required

The dataset containing the table. Accepts a literal dataset ID or a
reference to a GcpBigQueryDataset resource. Immutable.

- references: GcpBigQueryDataset (`status.outputs.dataset_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpBigQueryDataset, name: <that resource's name>, fieldPath: status.outputs.dataset_id}} -- a bare string does not parse

### spec.tableId

`string` · required

Unique identifier for the table within the dataset. Immutable.
Letters, numbers, and underscores; maximum 1024 characters.
Example: "events_raw", "revenue_summary"

- rule: {"required":true,"string":{"maxLen":"1024","pattern":"^[0-9A-Za-z_]+$"}}

### spec.friendlyName

`string`

User-friendly display name for the table.

### spec.description

`string`

Description of the table's contents or purpose.

### spec.labels

`map<string, string>`

Labels applied to the table for cost attribution and organization.
Merged with Planton's platform labels (which win on key conflicts).

### spec.resourceTags

`map<string, string>`

Resource Manager tags bound to the table, as
"tagKeys-namespaced-name" -> "tagValue short name" pairs. Unlike
labels, tags participate in IAM conditions and organization policy.

### spec.schema

`string`

Table schema as a JSON array of field definitions, e.g.
[{"name":"id","type":"INT64","mode":"REQUIRED"},
 {"name":"payload","type":"JSON"}]. Columns are add-only in place:
removing or retyping a column recreates the table. Omit for views
(defined by their query) and autodetected external tables.

### spec.timePartitioning

`GcpBigQueryTableTimePartitioning`

Time-based partitioning. Mutually exclusive with range_partitioning.
The partitioning field is immutable.

### spec.timePartitioning.type

`string` · required

Partition granularity: DAY, HOUR, MONTH, or YEAR. DAY is the common
default for event data; HOUR suits very high-volume streams.

- rule: {"required":true,"string":{"in":["DAY","HOUR","MONTH","YEAR"]}}

### spec.timePartitioning.field

`string`

The DATE, TIMESTAMP, or DATETIME column to partition on. Immutable.
If omitted, the table is ingestion-time partitioned (the pseudo-columns
_PARTITIONTIME / _PARTITIONDATE carry the partition key).

### spec.timePartitioning.expirationMs

`int64`

Number of milliseconds to keep each partition before it is dropped.
If not set (0), partitions never expire (the dataset's
default_partition_expiration_ms still applies to new tables).

### spec.rangePartitioning

`GcpBigQueryTableRangePartitioning`

Integer-range partitioning. Mutually exclusive with time_partitioning.
The partitioning field is immutable.

### spec.rangePartitioning.field

`string` · required

The INTEGER column to partition on. Immutable.

- rule: {"required":true}

### spec.rangePartitioning.range

`GcpBigQueryTableRangePartitioningRange` · required

The partition key space (start / end / interval).

- rule: {"required":true}
- rule: range partitioning requires end > start and interval >= 1

### spec.rangePartitioning.range.start

`int64`

Start of range partitioning, inclusive.

### spec.rangePartitioning.range.end

`int64`

End of range partitioning, exclusive. Must be greater than start.

### spec.rangePartitioning.range.interval

`int64`

Width of each partition interval. Must be at least 1.

### spec.clustering

`[]string`

Up to four columns to cluster by, in precedence order. Queries
filtering on the leading clustering columns scan less data.

- rule: {"repeated":{"maxItems":"4"}}

### spec.requirePartitionFilter

`bool`

Require every query against the table to carry a partition filter
predicate — the cost guard for large partitioned tables.

### spec.expirationTime

`int64`

Time when the table expires and is deleted, in milliseconds since
epoch. If not set (0), the table never expires (the dataset's
default_table_expiration_ms applied at creation still governs).

### spec.maxStaleness

`string`

Maximum staleness tolerated when reading a BigLake table with metadata
caching (SQL interval string, e.g. "0-0 0 4:0:0" for 4 hours).

### spec.kmsKeyName

`string | valueFrom`

Cloud KMS key encrypting the table (CMEK). Immutable — changing the
key recreates the table. Overrides the dataset's default key. The
BigQuery service agent must hold cryptoKeyEncrypterDecrypter on it.

- references: GcpKmsKey (`status.outputs.key_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_id}} -- a bare string does not parse

### spec.view

`GcpBigQueryTableView`

Makes the table a logical view. Mutually exclusive with
materialized_view and external_data_configuration.

### spec.view.query

`string` · required

The SQL query defining the view. Referenced tables must live in the
same location as the view's dataset.

- rule: {"required":true}

### spec.view.useLegacySql

`bool`

Whether the query uses BigQuery's legacy SQL dialect. Defaults to
false (GoogleSQL) — both engines always send the value explicitly, so
the BigQuery API's own legacy-SQL-by-default behavior for views never
silently applies.

### spec.materializedView

`GcpBigQueryTableMaterializedView`

Makes the table a materialized view. Mutually exclusive with view and
external_data_configuration.

### spec.materializedView.query

`string` · required

The SQL query defining the materialized view. Immutable — changing the
query recreates the table.

- rule: {"required":true}

### spec.materializedView.enableRefresh

`bool` · optional (explicit presence)

Whether BigQuery automatically refreshes the results when base tables
change. When omitted, GCP's default (enabled) applies; set false to
refresh only manually.

### spec.materializedView.refreshIntervalMs

`int64`

Maximum staleness between refreshes, in milliseconds. If not set (0),
GCP's default of 1800000 (30 minutes) applies.

### spec.materializedView.allowNonIncrementalDefinition

`bool`

Allow a query shape BigQuery cannot refresh incrementally (full
re-computation on refresh). Immutable. Default false.

### spec.externalDataConfiguration

`GcpBigQueryTableExternalDataConfiguration`

Makes the table external (data stays in GCS/Sheets/Bigtable/...).
Mutually exclusive with view and materialized_view.

- rule: source_format and object_metadata are mutually exclusive

### spec.externalDataConfiguration.autodetect

`bool`

Let BigQuery infer the schema and options from the source. Both
engines always send this value explicitly (the API requires the field).

### spec.externalDataConfiguration.sourceUris

`[]string` · required

Fully qualified source URIs, e.g. gs://bucket/path/*. Wildcards are
allowed after the bucket name (not for Bigtable or Google Drive).

- rule: {"repeated":{"minItems":"1"}}

### spec.externalDataConfiguration.sourceFormat

`string`

Format of the data: CSV, GOOGLE_SHEETS, NEWLINE_DELIMITED_JSON, AVRO,
ICEBERG, DATASTORE_BACKUP, PARQUET, ORC, BIGTABLE, or DELTA_LAKE.
Mutually exclusive with object_metadata (object tables).

- rule: source_format must be one of CSV, GOOGLE_SHEETS, NEWLINE_DELIMITED_JSON, AVRO, ICEBERG, DATASTORE_BACKUP, PARQUET, ORC, BIGTABLE, DELTA_LAKE

### spec.externalDataConfiguration.objectMetadata

`string`

Set to SIMPLE to create an OBJECT table over unstructured GCS objects
(requires connection_id; mutually exclusive with source_format).

- rule: object_metadata must be SIMPLE when set

### spec.externalDataConfiguration.compression

`string`

Compression of the source data: NONE (default) or GZIP.

- rule: compression must be NONE or GZIP

### spec.externalDataConfiguration.schema

`string`

Explicit schema for the external data as a JSON array of field
definitions. Immutable. Mutually exclusive with autodetect in
practice; most formats are self-describing.

### spec.externalDataConfiguration.ignoreUnknownValues

`bool`

Tolerate rows with values that do not match the schema (extra columns
ignored, missing values become NULL).

### spec.externalDataConfiguration.maxBadRecords

`int32`

Maximum number of bad records to tolerate before failing the query.

### spec.externalDataConfiguration.connectionId

`string`

Connection (projects/{p}/locations/{l}/connections/{c}) whose
credential reads the source — this is what upgrades a plain external
table to a BigLake table. Plain string until a connection kind exists.

### spec.externalDataConfiguration.referenceFileSchemaUri

`string`

Reference file providing the table schema for AVRO/PARQUET/ORC.

### spec.externalDataConfiguration.metadataCacheMode

`string`

Metadata caching for BigLake tables: AUTOMATIC (refresh within
max_staleness) or MANUAL (on-demand refresh). Requires connection_id.

- rule: metadata_cache_mode must be AUTOMATIC or MANUAL

### spec.externalDataConfiguration.fileSetSpecType

`string`

How source URIs are interpreted: FILE_SYSTEM_MATCH (default; expand
via object listing) or NEW_LINE_DELIMITED_MANIFEST (URIs point to
manifest files, one data URI per line).

- rule: file_set_spec_type must be FILE_SET_SPEC_TYPE_FILE_SYSTEM_MATCH or FILE_SET_SPEC_TYPE_NEW_LINE_DELIMITED_MANIFEST

### spec.externalDataConfiguration.jsonExtension

`string`

Parse extension for JSON data: GEOJSON (newline-delimited GeoJSON).

- rule: json_extension must be GEOJSON when set

### spec.externalDataConfiguration.csvOptions

`GcpBigQueryTableCsvOptions`

Format-specific parsing options — set the block matching source_format.

### spec.externalDataConfiguration.csvOptions.quote

`string` · optional (explicit presence)

The value used to quote data sections. When omitted, the API default
double-quote (") applies; set an explicit empty string for unquoted
data. Presence-tracked because the empty string is meaningful.

### spec.externalDataConfiguration.csvOptions.allowJaggedRows

`bool`

Accept rows missing trailing optional columns (missing values become
NULL).

### spec.externalDataConfiguration.csvOptions.allowQuotedNewlines

`bool`

Allow quoted data sections containing newlines.

### spec.externalDataConfiguration.csvOptions.encoding

`string`

Character encoding of the data: UTF-8 (default) or ISO-8859-1.

- rule: encoding must be UTF-8 or ISO-8859-1

### spec.externalDataConfiguration.csvOptions.fieldDelimiter

`string`

Field separator. Defaults to comma.

### spec.externalDataConfiguration.csvOptions.skipLeadingRows

`int32`

Number of header rows to skip.

### spec.externalDataConfiguration.csvOptions.sourceColumnMatch

`string`

How source columns map onto the table schema: POSITION (by ordering,
the safe choice for stable extracts) or NAME (reads the header row
and reorders to match schema field names — tolerates column
reshuffles, requires a header).

- rule: source_column_match must be POSITION or NAME

### spec.externalDataConfiguration.jsonOptions

`GcpBigQueryTableJsonOptions`

### spec.externalDataConfiguration.jsonOptions.encoding

`string`

Character encoding of the data. Defaults to UTF-8.

- rule: encoding must be one of UTF-8, UTF-16BE, UTF-16LE, UTF-32BE, UTF-32LE

### spec.externalDataConfiguration.googleSheetsOptions

`GcpBigQueryTableGoogleSheetsOptions`

### spec.externalDataConfiguration.googleSheetsOptions.range

`string`

Sheet range to query, e.g. "sheet1!A1:B20". When omitted the first
sheet is used.

### spec.externalDataConfiguration.googleSheetsOptions.skipLeadingRows

`int32`

Number of header rows to skip.

### spec.externalDataConfiguration.hivePartitioningOptions

`GcpBigQueryTableHivePartitioningOptions`

### spec.externalDataConfiguration.hivePartitioningOptions.mode

`string`

Partition-key inference mode: AUTO (infer types), STRINGS (all keys as
strings), or CUSTOM (types encoded in source_uri_prefix).

- rule: mode must be AUTO, STRINGS, or CUSTOM

### spec.externalDataConfiguration.hivePartitioningOptions.requirePartitionFilter

`bool`

Require a partition filter predicate in every query against the table.

### spec.externalDataConfiguration.hivePartitioningOptions.sourceUriPrefix

`string`

Common prefix of all source URIs before partition-key encoding begins,
e.g. gs://bucket/path (or gs://bucket/path/{dt:DATE} for CUSTOM mode).

### spec.externalDataConfiguration.avroOptions

`GcpBigQueryTableAvroOptions`

### spec.externalDataConfiguration.avroOptions.useAvroLogicalTypes

`bool`

Interpret Avro logical types (timestamp-micros, decimal, ...) as their
corresponding BigQuery types instead of raw primitives. Both engines
always send this value explicitly.

### spec.externalDataConfiguration.parquetOptions

`GcpBigQueryTableParquetOptions`

### spec.externalDataConfiguration.parquetOptions.enumAsString

`bool`

Infer Parquet ENUM logical type as STRING instead of BYTES.

### spec.externalDataConfiguration.parquetOptions.enableListInference

`bool`

Infer Parquet LIST logical type as the list's element type instead of
a repeated record wrapper.

### spec.externalDataConfiguration.bigtableOptions

`GcpBigQueryTableBigtableOptions`

### spec.externalDataConfiguration.bigtableOptions.ignoreUnspecifiedColumnFamilies

`bool`

Skip unspecified column families instead of failing.

### spec.externalDataConfiguration.bigtableOptions.readRowkeyAsString

`bool`

Read the row key as a STRING instead of BYTES.

### spec.externalDataConfiguration.bigtableOptions.outputColumnFamiliesAsJson

`bool`

Expose unlisted column families as a JSON-typed column each.

### spec.externalDataConfiguration.bigtableOptions.columnFamilies

`[]GcpBigQueryTableBigtableColumnFamily`

Column families to expose, with per-family and per-column typing.

### spec.externalDataConfiguration.bigtableOptions.columnFamilies[].familyId

`string`

Identifier of the column family.

### spec.externalDataConfiguration.bigtableOptions.columnFamilies[].type

`string`

Default type for values in this family (overridable per column).

### spec.externalDataConfiguration.bigtableOptions.columnFamilies[].encoding

`string`

Default encoding for values in this family (overridable per column).

### spec.externalDataConfiguration.bigtableOptions.columnFamilies[].onlyReadLatest

`bool`

Default only-latest-version posture for this family.

### spec.externalDataConfiguration.bigtableOptions.columnFamilies[].columns

`[]GcpBigQueryTableBigtableColumn`

Columns of the family to expose individually.

- rule: exactly one of qualifier_encoded or qualifier_string must be set

### spec.externalDataConfiguration.bigtableOptions.columnFamilies[].columns[].qualifierEncoded

`string`

Qualifier of the column, base64-encoded (for qualifiers that are not
valid UTF-8). Exactly one of qualifier_encoded or qualifier_string.

### spec.externalDataConfiguration.bigtableOptions.columnFamilies[].columns[].qualifierString

`string`

Qualifier of the column as a UTF-8 string.

### spec.externalDataConfiguration.bigtableOptions.columnFamilies[].columns[].fieldName

`string`

Field name to use in the table instead of the qualifier.

### spec.externalDataConfiguration.bigtableOptions.columnFamilies[].columns[].type

`string`

Type to convert the value to: BYTES (default), STRING, INTEGER, FLOAT,
BOOLEAN, JSON.

### spec.externalDataConfiguration.bigtableOptions.columnFamilies[].columns[].encoding

`string`

Encoding of the values: TEXT or BINARY (default).

### spec.externalDataConfiguration.bigtableOptions.columnFamilies[].columns[].onlyReadLatest

`bool`

Expose only the latest cell version instead of all versions.

### spec.externalDataConfiguration.decimalTargetTypes

`[]string`

Types decimal source values may convert to. The API picks the FIRST
type (in its fixed NUMERIC → BIGNUMERIC → STRING precedence, not this
list's order) that is listed here and fits the value's precision and
scale.

- rule: {"repeated":{"unique":true,"items":{"string":{"in":["NUMERIC","BIGNUMERIC","STRING"]}}}}

### spec.tableConstraints

`GcpBigQueryTableConstraints`

Unenforced primary/foreign keys for the optimizer and lineage tools.

### spec.tableConstraints.primaryKey

`GcpBigQueryTablePrimaryKey`

Primary key of the table.

### spec.tableConstraints.primaryKey.columns

`[]string` · required

Columns composing the primary key.

- rule: {"repeated":{"minItems":"1"}}

### spec.tableConstraints.foreignKeys

`[]GcpBigQueryTableForeignKey`

Foreign keys of the table.

### spec.tableConstraints.foreignKeys[].name

`string`

Optional name of the constraint.

### spec.tableConstraints.foreignKeys[].referencedTable

`GcpBigQueryTableForeignKeyReferencedTable` · required

The table this key references.

- rule: {"required":true}

### spec.tableConstraints.foreignKeys[].referencedTable.projectId

`string` · required

Project of the referenced table.

- rule: {"required":true}

### spec.tableConstraints.foreignKeys[].referencedTable.datasetId

`string` · required

Dataset of the referenced table.

- rule: {"required":true}

### spec.tableConstraints.foreignKeys[].referencedTable.tableId

`string | valueFrom` · required

The referenced table. Accepts a literal table ID or a reference to a
GcpBigQueryTable resource.

- references: GcpBigQueryTable (`status.outputs.table_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpBigQueryTable, name: <that resource's name>, fieldPath: status.outputs.table_id}} -- a bare string does not parse

### spec.tableConstraints.foreignKeys[].columnReferences

`GcpBigQueryTableForeignKeyColumnReferences` · required

The referencing/referenced column pair.

- rule: {"required":true}

### spec.tableConstraints.foreignKeys[].columnReferences.referencingColumn

`string` · required

The column in this table.

- rule: {"required":true}

### spec.tableConstraints.foreignKeys[].columnReferences.referencedColumn

`string` · required

The column in the referenced table.

- rule: {"required":true}

### spec.tableReplicationInfo

`GcpBigQueryTableReplicationInfo`

Makes the table a replica of a source materialized view. Immutable.

### spec.tableReplicationInfo.sourceProjectId

`string` · required

Project of the source materialized view.

- rule: {"required":true}

### spec.tableReplicationInfo.sourceDatasetId

`string` · required

Dataset of the source materialized view.

- rule: {"required":true}

### spec.tableReplicationInfo.sourceTableId

`string` · required

The source materialized view.

- rule: {"required":true}

### spec.tableReplicationInfo.replicationIntervalMs

`int64`

Replication interval in milliseconds. If not set (0), GCP's default of
300000 (5 minutes) applies.

### spec.biglakeConfiguration

`GcpBigQueryTableBiglakeConfiguration`

BigLake managed-table storage (Iceberg files in your GCS bucket).
Immutable.

### spec.biglakeConfiguration.connectionId

`string` · required

Connection (projects/{p}/locations/{l}/connections/{c}) whose
credential accesses the storage. Plain string until a connection kind
exists.

- rule: {"required":true}

### spec.biglakeConfiguration.storageUri

`string` · required

Fully qualified storage location prefix, e.g. gs://bucket/path.

- rule: {"required":true}

### spec.biglakeConfiguration.fileFormat

`string` · required

Open-source file format of the data. Currently PARQUET.

- rule: {"required":true}

### spec.biglakeConfiguration.tableFormat

`string` · required

Open-source table format managing the metadata. Currently ICEBERG.

- rule: {"required":true}

### spec.schemaForeignTypeInfo

`GcpBigQueryTableSchemaForeignTypeInfo`

Marks the schema as expressed in a foreign type system (e.g. HIVE).
Immutable.

### spec.schemaForeignTypeInfo.typeSystem

`string` · required

The type system of the schema, e.g. HIVE. Immutable.

- rule: {"required":true}

### spec.externalCatalogTableOptions

`GcpBigQueryTableExternalCatalogTableOptions`

Hive-metastore-compatible metadata for open-source engines.

### spec.externalCatalogTableOptions.parameters

`map<string, string>`

Hive-table-style key/value parameters (the whole map is limited to
30 KB by the API).

### spec.externalCatalogTableOptions.connectionId

`string`

Connection whose credential accesses the storage.

### spec.externalCatalogTableOptions.storageDescriptor

`GcpBigQueryTableStorageDescriptor`

Physical storage description of the table.

### spec.externalCatalogTableOptions.storageDescriptor.locationUri

`string`

Physical location of the table, e.g. gs://bucket/path.

### spec.externalCatalogTableOptions.storageDescriptor.inputFormat

`string`

Fully qualified Hive input format class name.

### spec.externalCatalogTableOptions.storageDescriptor.outputFormat

`string`

Fully qualified Hive output format class name.

### spec.externalCatalogTableOptions.storageDescriptor.serdeInfo

`GcpBigQueryTableSerDeInfo`

Serializer/deserializer information.

### spec.externalCatalogTableOptions.storageDescriptor.serdeInfo.name

`string`

Optional name of the SerDe.

### spec.externalCatalogTableOptions.storageDescriptor.serdeInfo.serializationLibrary

`string` · required

Fully qualified SerDe class name, e.g.
org.apache.hadoop.hive.ql.io.orc.OrcSerde.

- rule: {"required":true}

### spec.externalCatalogTableOptions.storageDescriptor.serdeInfo.parameters

`map<string, string>`

SerDe key/value parameters.

### spec.deletionProtection

`bool` · optional (explicit presence)

Prevents the table from being destroyed while true. Defaults to true:
a destroy fails until this is set to false — the guard for tables
holding real data. Set false for disposable/dev tables.

- default: `true`

### spec.deletionPolicy

`string`

What destroying this resource does to the table. Two levers guard a
table: this one decides whether removal is attempted at all, and
deletion_protection (checked second) blocks an attempted delete.
ABANDON therefore bypasses deletion_protection — the table leaves
management untouched — while DELETE still requires
deletion_protection=false to actually destroy data:
  ""        -- same as "DELETE" (provider default)
  "DELETE"  -- the table (and its data) is deleted, if
               deletion_protection allows it
  "PREVENT" -- destroy FAILS before deletion_protection is even
               consulted
  "ABANDON" -- the table is removed from management and keeps serving
               queries in GCP

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

### spec.ignoreAutoGeneratedSchema

`bool`

Hide diffs from columns BigQuery adds on its own (e.g. columns
materialized by flexible-schema features) so they are not fought over
on every apply. Terraform-plan-level behavior; no API field.

### spec.ignoreSchemaChanges

`[]string`

Schema sub-fields treated as non-authoritative per column — the
provider stops reconciling them. The provider currently supports
exactly "dataPolicies" (column-level data-policy attachments managed
outside this spec); other strings are accepted by the provider but
are inert.

- rule: {"repeated":{"maxItems":"10","unique":true,"items":{"string":{"in":["dataPolicies"]}}}}

### spec.tableMetadataView

`string`

How much table metadata the provider requests when reading the table
back: BASIC (cheapest), STORAGE_STATS (the API default), or FULL.
A read-tuning knob for very large fleets; no effect on the table
itself.

- rule: table_metadata_view must be one of: TABLE_METADATA_VIEW_UNSPECIFIED, BASIC, STORAGE_STATS, FULL

## Validation Rules

- `one_partitioning_method`: time_partitioning and range_partitioning are mutually exclusive
- `one_table_arm`: at most one of view, materialized_view, or external_data_configuration may be set
- `logical_view_has_no_native_shape`: a logical view must not set schema, time_partitioning, range_partitioning, or clustering
- `materialized_view_has_no_schema`: a materialized view must not set schema (the query defines it)

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpBigQueryTable, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.table_id` | `string` | The short table ID (same as the spec's table_id input) — the value SQL queries and foreign keys reference. |
| `status.outputs.self_link` | `string` | The fully qualified URI of the table. Format: https://bigquery.googleapis.com/bigquery/v2/projects/{project}/datasets/{dataset}/tables/{table} |
| `status.outputs.project` | `string` | The GCP project that contains the table (resolved even under the ambient-project fallback). |
| `status.outputs.dataset_id` | `string` | The dataset that contains the table. |
| `status.outputs.type` | `string` | The kind of table GCP materialized: TABLE, VIEW, MATERIALIZED_VIEW, or EXTERNAL. |
| `status.outputs.location` | `string` | Geographic location of the table (inherited from the dataset). |
| `status.outputs.creation_time` | `int64` | The creation time of the table in milliseconds since epoch. |
| `status.outputs.qualified_name` | `string` | The dotted fully qualified table name: {project}.{dataset}.{table}. Pre-assembled so consumers that address tables in SQL-style dotted form (Pub/Sub BigQuery delivery, query tooling) can reference the table without string assembly. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.datasetId` | GcpBigQueryDataset | `status.outputs.dataset_id` |
| `spec.kmsKeyName` | GcpKmsKey | `status.outputs.key_id` |
| `spec.tableConstraints.foreignKeys[].referencedTable.tableId` | GcpBigQueryTable | `status.outputs.table_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| GcpBigQueryTable | `spec.tableConstraints.foreignKeys[].referencedTable.tableId` | `status.outputs.table_id` |
| GcpPubSubSubscription | `spec.bigqueryConfig.table` | `status.outputs.qualified_name` |
| GcpVertexAiFeatureGroup | `spec.bigQuery.inputUri` | `status.outputs.qualified_name` |
| GcpVertexAiFeatureOnlineStore | `spec.featureViews[].bigQuerySource.uri` | `status.outputs.qualified_name` |

## See Also

- [Overview](../README.md)
