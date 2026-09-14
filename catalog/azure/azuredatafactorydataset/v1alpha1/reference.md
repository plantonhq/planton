# AzureDataFactoryDataset

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `azure.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

**AzureDataFactoryDatasetSpec** defines an Azure Data Factory
dataset -- a named view of data inside a system a linked service
already connects to: which container and path, which table, which
file format. Pipelines and data flows read from and write to
datasets; the dataset itself stores only metadata and carries no
secret material.

The data shape is declared by which variant block is present: set
exactly ONE of the 13 blocks below. All shapes share one
factory-scoped name namespace ({factory_id}/datasets/{name}).
File formats (delimited text/CSV, JSON, Parquet, binary) locate
their files through a location block (blob storage, Data Lake
Gen2, an HTTP server, or SFTP for binary); table forms (Azure SQL,
SQL Server, MySQL, PostgreSQL, Snowflake, Cosmos DB) name a table
or collection; the raw-JSON `custom` form covers every dataset
type Azure Data Factory speaks that has no first-class block here.

Every dataset points at the linked service that carries its
connection: 11 variants reference it by NAME through the shared
`linked_service_name` field; the `azure_sql_table` variant
references it by ARM ID inside its own block, and `custom` carries
its own reference block (the only form that can pass
reference-level parameter values).

## Example

```yaml
# Deep-shape example for docs and offline validation: a delimited
# text (CSV) dataset exercising the variant's full surface -- a
# dynamic-path blob location, every parse setting, compression, and
# declared columns. References are literal values so the manifest
# validates standalone.
apiVersion: azure.planton.dev/v1alpha1
kind: AzureDataFactoryDataset
metadata:
  name: test-dataset
  id: test-dataset
  org: test-org
  env: test
spec:
  dataFactoryId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.DataFactory/factories/test-df
  name: orders-csv
  description: Daily order exports -- CSV files landing in blob storage.
  linkedServiceName:
    value: lakehouse-blob
  annotations:
    - team:data
  parameters:
    runDate: ""
  folder: ingest/orders
  delimitedText:
    azureBlobStorageLocation:
      container: landing
      path: raw/orders/@{dataset().runDate}
      dynamicPathEnabled: true
      filename: orders.csv
    columnDelimiter: ";"
    rowDelimiter: "\n"
    quoteCharacter: "'"
    escapeCharacter: "/"
    encoding: UTF-8
    firstRowAsHeader: true
    nullValue: "NULL"
    compressionCodec: gzip
    compressionLevel: Fastest
    schemaColumn:
      - name: order_id
        type: Int64
      - name: placed_at
        type: DateTime
        description: Order timestamp
      - name: total
        type: Decimal
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.dataFactoryId` | `string \| valueFrom` | yes |  | AzureDataFactory (`status.outputs.data_factory_id`) |
| `spec.name` | `string` | yes |  |  |
| `spec.linkedServiceName` | `string \| valueFrom` |  |  | AzureDataFactoryLinkedService (`status.outputs.linked_service_name`) |
| `spec.description` | `string` |  |  |  |
| `spec.annotations` | `[]string` |  |  |  |
| `spec.parameters` | `map<string, string>` |  |  |  |
| `spec.additionalProperties` | `map<string, string>` |  |  |  |
| `spec.folder` | `string` |  |  |  |
| `spec.azureBlob` | `AzureDataFactoryDatasetAzureBlob` |  |  |  |
| `spec.azureBlob.path` | `string` |  |  |  |
| `spec.azureBlob.filename` | `string` |  |  |  |
| `spec.azureBlob.dynamicPathEnabled` | `bool` |  |  |  |
| `spec.azureBlob.dynamicFilenameEnabled` | `bool` |  |  |  |
| `spec.azureBlob.schemaColumn` | `[]AzureDataFactoryDatasetSchemaColumn` |  |  |  |
| `spec.azureBlob.schemaColumn[].name` | `string` | yes |  |  |
| `spec.azureBlob.schemaColumn[].type` | `string` |  |  |  |
| `spec.azureBlob.schemaColumn[].description` | `string` |  |  |  |
| `spec.azureSqlTable` | `AzureDataFactoryDatasetAzureSqlTable` |  |  |  |
| `spec.azureSqlTable.linkedServiceId` | `string \| valueFrom` | yes |  | AzureDataFactoryLinkedService (`status.outputs.linked_service_id`) |
| `spec.azureSqlTable.schema` | `string` |  |  |  |
| `spec.azureSqlTable.table` | `string` |  |  |  |
| `spec.azureSqlTable.schemaColumn` | `[]AzureDataFactoryDatasetSchemaColumn` |  |  |  |
| `spec.azureSqlTable.schemaColumn[].name` | `string` | yes |  |  |
| `spec.azureSqlTable.schemaColumn[].type` | `string` |  |  |  |
| `spec.azureSqlTable.schemaColumn[].description` | `string` |  |  |  |
| `spec.binary` | `AzureDataFactoryDatasetBinary` |  |  |  |
| `spec.binary.httpServerLocation` | `AzureDataFactoryDatasetHttpServerLocation` |  |  |  |
| `spec.binary.httpServerLocation.relativeUrl` | `string` | yes |  |  |
| `spec.binary.httpServerLocation.path` | `string` |  |  |  |
| `spec.binary.httpServerLocation.dynamicPathEnabled` | `bool` |  |  |  |
| `spec.binary.httpServerLocation.filename` | `string` |  |  |  |
| `spec.binary.httpServerLocation.dynamicFilenameEnabled` | `bool` |  |  |  |
| `spec.binary.azureBlobStorageLocation` | `AzureDataFactoryDatasetBlobStorageLocation` |  |  |  |
| `spec.binary.azureBlobStorageLocation.container` | `string` | yes |  |  |
| `spec.binary.azureBlobStorageLocation.dynamicContainerEnabled` | `bool` |  |  |  |
| `spec.binary.azureBlobStorageLocation.path` | `string` |  |  |  |
| `spec.binary.azureBlobStorageLocation.dynamicPathEnabled` | `bool` |  |  |  |
| `spec.binary.azureBlobStorageLocation.filename` | `string` |  |  |  |
| `spec.binary.azureBlobStorageLocation.dynamicFilenameEnabled` | `bool` |  |  |  |
| `spec.binary.sftpServerLocation` | `AzureDataFactoryDatasetSftpServerLocation` |  |  |  |
| `spec.binary.sftpServerLocation.path` | `string` | yes |  |  |
| `spec.binary.sftpServerLocation.dynamicPathEnabled` | `bool` |  |  |  |
| `spec.binary.sftpServerLocation.filename` | `string` | yes |  |  |
| `spec.binary.sftpServerLocation.dynamicFilenameEnabled` | `bool` |  |  |  |
| `spec.binary.compression` | `AzureDataFactoryDatasetBinaryCompression` |  |  |  |
| `spec.binary.compression.type` | `string` | yes |  |  |
| `spec.binary.compression.level` | `string` |  |  |  |
| `spec.cosmosdbSqlapi` | `AzureDataFactoryDatasetCosmosdbSqlapi` |  |  |  |
| `spec.cosmosdbSqlapi.collectionName` | `string` |  |  |  |
| `spec.cosmosdbSqlapi.schemaColumn` | `[]AzureDataFactoryDatasetSchemaColumn` |  |  |  |
| `spec.cosmosdbSqlapi.schemaColumn[].name` | `string` | yes |  |  |
| `spec.cosmosdbSqlapi.schemaColumn[].type` | `string` |  |  |  |
| `spec.cosmosdbSqlapi.schemaColumn[].description` | `string` |  |  |  |
| `spec.custom` | `AzureDataFactoryDatasetCustom` |  |  |  |
| `spec.custom.linkedService` | `AzureDataFactoryDatasetCustomLinkedService` | yes |  |  |
| `spec.custom.linkedService.name` | `string \| valueFrom` | yes |  | AzureDataFactoryLinkedService (`status.outputs.linked_service_name`) |
| `spec.custom.linkedService.parameters` | `map<string, string>` |  |  |  |
| `spec.custom.type` | `string` | yes |  |  |
| `spec.custom.typePropertiesJson` | `string` | yes |  |  |
| `spec.custom.schemaJson` | `string` |  |  |  |
| `spec.delimitedText` | `AzureDataFactoryDatasetDelimitedText` |  |  |  |
| `spec.delimitedText.httpServerLocation` | `AzureDataFactoryDatasetHttpServerLocation` |  |  |  |
| `spec.delimitedText.httpServerLocation.relativeUrl` | `string` | yes |  |  |
| `spec.delimitedText.httpServerLocation.path` | `string` |  |  |  |
| `spec.delimitedText.httpServerLocation.dynamicPathEnabled` | `bool` |  |  |  |
| `spec.delimitedText.httpServerLocation.filename` | `string` |  |  |  |
| `spec.delimitedText.httpServerLocation.dynamicFilenameEnabled` | `bool` |  |  |  |
| `spec.delimitedText.azureBlobStorageLocation` | `AzureDataFactoryDatasetBlobStorageLocation` |  |  |  |
| `spec.delimitedText.azureBlobStorageLocation.container` | `string` | yes |  |  |
| `spec.delimitedText.azureBlobStorageLocation.dynamicContainerEnabled` | `bool` |  |  |  |
| `spec.delimitedText.azureBlobStorageLocation.path` | `string` |  |  |  |
| `spec.delimitedText.azureBlobStorageLocation.dynamicPathEnabled` | `bool` |  |  |  |
| `spec.delimitedText.azureBlobStorageLocation.filename` | `string` |  |  |  |
| `spec.delimitedText.azureBlobStorageLocation.dynamicFilenameEnabled` | `bool` |  |  |  |
| `spec.delimitedText.azureBlobFsLocation` | `AzureDataFactoryDatasetBlobFsLocation` |  |  |  |
| `spec.delimitedText.azureBlobFsLocation.fileSystem` | `string` |  |  |  |
| `spec.delimitedText.azureBlobFsLocation.dynamicFileSystemEnabled` | `bool` |  |  |  |
| `spec.delimitedText.azureBlobFsLocation.path` | `string` |  |  |  |
| `spec.delimitedText.azureBlobFsLocation.dynamicPathEnabled` | `bool` |  |  |  |
| `spec.delimitedText.azureBlobFsLocation.filename` | `string` |  |  |  |
| `spec.delimitedText.azureBlobFsLocation.dynamicFilenameEnabled` | `bool` |  |  |  |
| `spec.delimitedText.columnDelimiter` | `string` |  |  |  |
| `spec.delimitedText.rowDelimiter` | `string` |  |  |  |
| `spec.delimitedText.quoteCharacter` | `string` |  |  |  |
| `spec.delimitedText.escapeCharacter` | `string` |  |  |  |
| `spec.delimitedText.encoding` | `string` |  |  |  |
| `spec.delimitedText.firstRowAsHeader` | `bool` |  | `false` |  |
| `spec.delimitedText.nullValue` | `string` |  |  |  |
| `spec.delimitedText.compressionCodec` | `string` |  |  |  |
| `spec.delimitedText.compressionLevel` | `string` |  |  |  |
| `spec.delimitedText.schemaColumn` | `[]AzureDataFactoryDatasetSchemaColumn` |  |  |  |
| `spec.delimitedText.schemaColumn[].name` | `string` | yes |  |  |
| `spec.delimitedText.schemaColumn[].type` | `string` |  |  |  |
| `spec.delimitedText.schemaColumn[].description` | `string` |  |  |  |
| `spec.http` | `AzureDataFactoryDatasetHttp` |  |  |  |
| `spec.http.relativeUrl` | `string` |  |  |  |
| `spec.http.requestBody` | `string` |  |  |  |
| `spec.http.requestMethod` | `string` |  |  |  |
| `spec.http.schemaColumn` | `[]AzureDataFactoryDatasetSchemaColumn` |  |  |  |
| `spec.http.schemaColumn[].name` | `string` | yes |  |  |
| `spec.http.schemaColumn[].type` | `string` |  |  |  |
| `spec.http.schemaColumn[].description` | `string` |  |  |  |
| `spec.json` | `AzureDataFactoryDatasetJson` |  |  |  |
| `spec.json.httpServerLocation` | `AzureDataFactoryDatasetHttpServerLocation` |  |  |  |
| `spec.json.httpServerLocation.relativeUrl` | `string` | yes |  |  |
| `spec.json.httpServerLocation.path` | `string` |  |  |  |
| `spec.json.httpServerLocation.dynamicPathEnabled` | `bool` |  |  |  |
| `spec.json.httpServerLocation.filename` | `string` |  |  |  |
| `spec.json.httpServerLocation.dynamicFilenameEnabled` | `bool` |  |  |  |
| `spec.json.azureBlobStorageLocation` | `AzureDataFactoryDatasetBlobStorageLocation` |  |  |  |
| `spec.json.azureBlobStorageLocation.container` | `string` | yes |  |  |
| `spec.json.azureBlobStorageLocation.dynamicContainerEnabled` | `bool` |  |  |  |
| `spec.json.azureBlobStorageLocation.path` | `string` |  |  |  |
| `spec.json.azureBlobStorageLocation.dynamicPathEnabled` | `bool` |  |  |  |
| `spec.json.azureBlobStorageLocation.filename` | `string` |  |  |  |
| `spec.json.azureBlobStorageLocation.dynamicFilenameEnabled` | `bool` |  |  |  |
| `spec.json.encoding` | `string` |  |  |  |
| `spec.json.schemaColumn` | `[]AzureDataFactoryDatasetSchemaColumn` |  |  |  |
| `spec.json.schemaColumn[].name` | `string` | yes |  |  |
| `spec.json.schemaColumn[].type` | `string` |  |  |  |
| `spec.json.schemaColumn[].description` | `string` |  |  |  |
| `spec.mysql` | `AzureDataFactoryDatasetMysql` |  |  |  |
| `spec.mysql.tableName` | `string` |  |  |  |
| `spec.mysql.schemaColumn` | `[]AzureDataFactoryDatasetSchemaColumn` |  |  |  |
| `spec.mysql.schemaColumn[].name` | `string` | yes |  |  |
| `spec.mysql.schemaColumn[].type` | `string` |  |  |  |
| `spec.mysql.schemaColumn[].description` | `string` |  |  |  |
| `spec.parquet` | `AzureDataFactoryDatasetParquet` |  |  |  |
| `spec.parquet.httpServerLocation` | `AzureDataFactoryDatasetHttpServerLocation` |  |  |  |
| `spec.parquet.httpServerLocation.relativeUrl` | `string` | yes |  |  |
| `spec.parquet.httpServerLocation.path` | `string` |  |  |  |
| `spec.parquet.httpServerLocation.dynamicPathEnabled` | `bool` |  |  |  |
| `spec.parquet.httpServerLocation.filename` | `string` |  |  |  |
| `spec.parquet.httpServerLocation.dynamicFilenameEnabled` | `bool` |  |  |  |
| `spec.parquet.azureBlobStorageLocation` | `AzureDataFactoryDatasetBlobStorageLocation` |  |  |  |
| `spec.parquet.azureBlobStorageLocation.container` | `string` | yes |  |  |
| `spec.parquet.azureBlobStorageLocation.dynamicContainerEnabled` | `bool` |  |  |  |
| `spec.parquet.azureBlobStorageLocation.path` | `string` |  |  |  |
| `spec.parquet.azureBlobStorageLocation.dynamicPathEnabled` | `bool` |  |  |  |
| `spec.parquet.azureBlobStorageLocation.filename` | `string` |  |  |  |
| `spec.parquet.azureBlobStorageLocation.dynamicFilenameEnabled` | `bool` |  |  |  |
| `spec.parquet.azureBlobFsLocation` | `AzureDataFactoryDatasetBlobFsLocation` |  |  |  |
| `spec.parquet.azureBlobFsLocation.fileSystem` | `string` |  |  |  |
| `spec.parquet.azureBlobFsLocation.dynamicFileSystemEnabled` | `bool` |  |  |  |
| `spec.parquet.azureBlobFsLocation.path` | `string` |  |  |  |
| `spec.parquet.azureBlobFsLocation.dynamicPathEnabled` | `bool` |  |  |  |
| `spec.parquet.azureBlobFsLocation.filename` | `string` |  |  |  |
| `spec.parquet.azureBlobFsLocation.dynamicFilenameEnabled` | `bool` |  |  |  |
| `spec.parquet.compressionCodec` | `string` |  |  |  |
| `spec.parquet.schemaColumn` | `[]AzureDataFactoryDatasetSchemaColumn` |  |  |  |
| `spec.parquet.schemaColumn[].name` | `string` | yes |  |  |
| `spec.parquet.schemaColumn[].type` | `string` |  |  |  |
| `spec.parquet.schemaColumn[].description` | `string` |  |  |  |
| `spec.postgresql` | `AzureDataFactoryDatasetPostgresql` |  |  |  |
| `spec.postgresql.tableName` | `string` |  |  |  |
| `spec.postgresql.schemaColumn` | `[]AzureDataFactoryDatasetSchemaColumn` |  |  |  |
| `spec.postgresql.schemaColumn[].name` | `string` | yes |  |  |
| `spec.postgresql.schemaColumn[].type` | `string` |  |  |  |
| `spec.postgresql.schemaColumn[].description` | `string` |  |  |  |
| `spec.snowflake` | `AzureDataFactoryDatasetSnowflake` |  |  |  |
| `spec.snowflake.tableName` | `string` |  |  |  |
| `spec.snowflake.schemaName` | `string` |  |  |  |
| `spec.snowflake.schemaColumn` | `[]AzureDataFactoryDatasetSnowflakeSchemaColumn` |  |  |  |
| `spec.snowflake.schemaColumn[].name` | `string` | yes |  |  |
| `spec.snowflake.schemaColumn[].type` | `string` |  |  |  |
| `spec.snowflake.schemaColumn[].precision` | `int32` |  |  |  |
| `spec.snowflake.schemaColumn[].scale` | `int32` |  |  |  |
| `spec.sqlServerTable` | `AzureDataFactoryDatasetSqlServerTable` |  |  |  |
| `spec.sqlServerTable.tableName` | `string` |  |  |  |
| `spec.sqlServerTable.schemaColumn` | `[]AzureDataFactoryDatasetSchemaColumn` |  |  |  |
| `spec.sqlServerTable.schemaColumn[].name` | `string` | yes |  |  |
| `spec.sqlServerTable.schemaColumn[].type` | `string` |  |  |  |
| `spec.sqlServerTable.schemaColumn[].description` | `string` |  |  |  |

## Field Details

### spec.dataFactoryId

`string | valueFrom` · required

The Data Factory the dataset lives in, by ARM ID. Can be a
literal string or a reference to an AzureDataFactory output.

**ForceNew**: changing this destroys and recreates the dataset.

- references: AzureDataFactory (`status.outputs.data_factory_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureDataFactory, name: <that resource's name>, fieldPath: status.outputs.data_factory_id}} -- a bare string does not parse

### spec.name

`string` · required

The dataset's name -- unique within the factory across all
dataset shapes. Azure's own rule is deliberately loose: a name
is rejected only when it consists ENTIRELY of the characters
- . + ? / < > * % & : \ (mirrored exactly; not tightened).

**ForceNew**: changing this destroys and recreates the dataset.

- rule: Dataset names must not consist entirely of the characters - . + ? / < > * % & : \
- rule: {"required":true}

### spec.linkedServiceName

`string | valueFrom`

The linked service the dataset reads through, by name -- defaults
to referencing an AzureDataFactoryLinkedService's
linked_service_name output. Required for every variant EXCEPT
azure_sql_table (which references its linked service by ARM ID
inside its own block) and custom (which carries its own reference
block with parameter values).

- references: AzureDataFactoryLinkedService (`status.outputs.linked_service_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureDataFactoryLinkedService, name: <that resource's name>, fieldPath: status.outputs.linked_service_name}} -- a bare string does not parse

### spec.description

`string`

A human-readable description of what the dataset is for.

### spec.annotations

`[]string`

Free-form annotation strings stored on the dataset.

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.parameters

`map<string, string>`

Dataset parameters, keyed by parameter name (string values only
-- the wire grammar Data Factory accepts through this surface).
Pipelines and data flows can override them per use.

### spec.additionalProperties

`map<string, string>`

Additional top-level properties passed through to Azure as-is --
Data Factory's escape hatch for dataset properties the schema
does not model. Applies to every variant.

### spec.folder

`string`

The Data Factory Studio folder the dataset appears under -- a
display path only ("/" separated), with no effect on the wire
behavior. Omit for the factory root.

### spec.azureBlob

`AzureDataFactoryDatasetAzureBlob`

Azure Blob Storage files addressed by a flat path + filename
pair.

### spec.azureBlob.path

`string`

The folder path inside the container -- omit to address the
container root.

### spec.azureBlob.filename

`string`

The file's name -- omit to address a folder (all files).

### spec.azureBlob.dynamicPathEnabled

`bool`

Treat path as a Data Factory expression (evaluated at run time,
e.g. from dataset parameters) instead of a literal.

### spec.azureBlob.dynamicFilenameEnabled

`bool`

Treat filename as a Data Factory expression (evaluated at run
time, e.g. from dataset parameters) instead of a literal.

### spec.azureBlob.schemaColumn

`[]AzureDataFactoryDatasetSchemaColumn`

Declared columns -- omit to let Data Factory infer or map at run
time.

### spec.azureBlob.schemaColumn[].name

`string` · required

The column's name.

- rule: {"required":true}

### spec.azureBlob.schemaColumn[].type

`string`

The column's Data Factory interim data type -- omit to leave the
type undeclared.

- rule: {"string":{"in":["","Byte","Byte[]","Boolean","Date","DateTime","DateTimeOffset","Decimal","Double","Guid","Int16","Int32","Int64","Single","String","TimeSpan"]}}

### spec.azureBlob.schemaColumn[].description

`string`

A human-readable description of the column.

### spec.azureSqlTable

`AzureDataFactoryDatasetAzureSqlTable`

An Azure SQL Database table. The one variant that references its
linked service by ARM ID (same factory enforced).

### spec.azureSqlTable.linkedServiceId

`string | valueFrom` · required

The Azure SQL Database linked service, by ARM ID
({factory_id}/linkedservices/{name}) -- defaults to referencing
an AzureDataFactoryLinkedService's linked_service_id output.
Must belong to the same factory as this dataset.

- references: AzureDataFactoryLinkedService (`status.outputs.linked_service_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureDataFactoryLinkedService, name: <that resource's name>, fieldPath: status.outputs.linked_service_id}} -- a bare string does not parse

### spec.azureSqlTable.schema

`string`

The table's schema name (e.g. dbo) -- omit to leave it
undeclared.

### spec.azureSqlTable.table

`string`

The table's name -- omit to leave it undeclared (e.g. when the
pipeline supplies a query instead).

### spec.azureSqlTable.schemaColumn

`[]AzureDataFactoryDatasetSchemaColumn`

Declared columns -- omit to let Data Factory infer or map at run
time.

### spec.azureSqlTable.schemaColumn[].name

`string` · required

The column's name.

- rule: {"required":true}

### spec.azureSqlTable.schemaColumn[].type

`string`

The column's Data Factory interim data type -- omit to leave the
type undeclared.

- rule: {"string":{"in":["","Byte","Byte[]","Boolean","Date","DateTime","DateTimeOffset","Decimal","Double","Guid","Int16","Int32","Int64","Single","String","TimeSpan"]}}

### spec.azureSqlTable.schemaColumn[].description

`string`

A human-readable description of the column.

### spec.binary

`AzureDataFactoryDatasetBinary`

Opaque binary files (no column structure).

- rule: Set exactly one of http_server_location, azure_blob_storage_location, or sftp_server_location
- rule: http_server_location requires both path and filename for the binary format

### spec.binary.httpServerLocation

`AzureDataFactoryDatasetHttpServerLocation`

Files on an HTTP server. Set exactly one location block. path and
filename are both required for this format (see the message
rule).

### spec.binary.httpServerLocation.relativeUrl

`string` · required

The URL path below the linked service's base URL.

- rule: {"required":true}

### spec.binary.httpServerLocation.path

`string`

The folder path on the server. Required by the delimited text,
JSON, and binary formats; optional for Parquet.

### spec.binary.httpServerLocation.dynamicPathEnabled

`bool`

Treat path as a Data Factory expression instead of a literal.

### spec.binary.httpServerLocation.filename

`string`

The file's name. Required by every format that uses this
location.

### spec.binary.httpServerLocation.dynamicFilenameEnabled

`bool`

Treat filename as a Data Factory expression instead of a literal.

### spec.binary.azureBlobStorageLocation

`AzureDataFactoryDatasetBlobStorageLocation`

Files in Azure Blob Storage. Set exactly one location block.

### spec.binary.azureBlobStorageLocation.container

`string` · required

The blob container's name.

- rule: {"required":true}

### spec.binary.azureBlobStorageLocation.dynamicContainerEnabled

`bool`

Treat container as a Data Factory expression instead of a
literal.

### spec.binary.azureBlobStorageLocation.path

`string`

The folder path inside the container -- omit to address the
container root. Required by the JSON format.

### spec.binary.azureBlobStorageLocation.dynamicPathEnabled

`bool`

Treat path as a Data Factory expression instead of a literal.

### spec.binary.azureBlobStorageLocation.filename

`string`

The file's name -- omit to address a folder (all files). Required
by the JSON format.

### spec.binary.azureBlobStorageLocation.dynamicFilenameEnabled

`bool`

Treat filename as a Data Factory expression instead of a literal.

### spec.binary.sftpServerLocation

`AzureDataFactoryDatasetSftpServerLocation`

Files on an SFTP server. Set exactly one location block.

### spec.binary.sftpServerLocation.path

`string` · required

The folder path on the server.

- rule: {"required":true}

### spec.binary.sftpServerLocation.dynamicPathEnabled

`bool`

Treat path as a Data Factory expression instead of a literal.

### spec.binary.sftpServerLocation.filename

`string` · required

The file's name.

- rule: {"required":true}

### spec.binary.sftpServerLocation.dynamicFilenameEnabled

`bool`

Treat filename as a Data Factory expression instead of a literal.

### spec.binary.compression

`AzureDataFactoryDatasetBinaryCompression`

How the files are compressed -- omit for uncompressed data.

### spec.binary.compression.type

`string` · required

The compression codec.

- rule: {"required":true,"string":{"in":["BZip2","Deflate","GZip","Tar","TarGZip","ZipDeflate"]}}

### spec.binary.compression.level

`string`

The compression level -- applies to the TarGZip, GZip, and
ZipDeflate codecs. Omit for the service default.

- rule: {"string":{"in":["","Optimal","Fastest"]}}

### spec.cosmosdbSqlapi

`AzureDataFactoryDatasetCosmosdbSqlapi`

An Azure Cosmos DB (SQL API) collection.

### spec.cosmosdbSqlapi.collectionName

`string`

The collection's name -- omit to leave it undeclared (e.g. when
the pipeline supplies a query instead).

### spec.cosmosdbSqlapi.schemaColumn

`[]AzureDataFactoryDatasetSchemaColumn`

Declared columns -- omit to let Data Factory infer or map at run
time.

### spec.cosmosdbSqlapi.schemaColumn[].name

`string` · required

The column's name.

- rule: {"required":true}

### spec.cosmosdbSqlapi.schemaColumn[].type

`string`

The column's Data Factory interim data type -- omit to leave the
type undeclared.

- rule: {"string":{"in":["","Byte","Byte[]","Boolean","Date","DateTime","DateTimeOffset","Decimal","Double","Guid","Int16","Int32","Int64","Single","String","TimeSpan"]}}

### spec.cosmosdbSqlapi.schemaColumn[].description

`string`

A human-readable description of the column.

### spec.custom

`AzureDataFactoryDatasetCustom`

Any other dataset type, as raw type-properties JSON -- the escape
hatch for the many Data Factory dataset types azurerm has no
first-class resource for.

### spec.custom.linkedService

`AzureDataFactoryDatasetCustomLinkedService` · required

The linked service the dataset reads through -- the only variant
form that can pass reference-level parameter values.

- rule: {"required":true}

### spec.custom.linkedService.name

`string | valueFrom` · required

The linked service's name inside this factory -- defaults to
referencing an AzureDataFactoryLinkedService's
linked_service_name output.

- references: AzureDataFactoryLinkedService (`status.outputs.linked_service_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureDataFactoryLinkedService, name: <that resource's name>, fieldPath: status.outputs.linked_service_name}} -- a bare string does not parse

### spec.custom.linkedService.parameters

`map<string, string>`

Values for the linked service's parameters, keyed by parameter
name.

### spec.custom.type

`string` · required

The Data Factory dataset type token, e.g. Excel, Xml, Avro --
exactly as the Data Factory REST API names it.

**ForceNew**: changing this destroys and recreates the dataset.

- rule: {"required":true}

### spec.custom.typePropertiesJson

`string` · required

The dataset's typeProperties object, as raw JSON -- the payload
is passed to Azure as-is, exactly as the Data Factory REST API
documents it for the chosen type. Azure validates it at save
time.

- rule: {"required":true}

### spec.custom.schemaJson

`string`

The dataset's schema object, as raw JSON -- omit to leave the
schema undeclared.

### spec.delimitedText

`AzureDataFactoryDatasetDelimitedText`

Delimited text (CSV) files.

- rule: Set exactly one of http_server_location, azure_blob_storage_location, or azure_blob_fs_location
- rule: http_server_location requires both path and filename for the delimited text format

### spec.delimitedText.httpServerLocation

`AzureDataFactoryDatasetHttpServerLocation`

Files on an HTTP server. Set exactly one location block. path and
filename are both required for this format (see the message
rule).

### spec.delimitedText.httpServerLocation.relativeUrl

`string` · required

The URL path below the linked service's base URL.

- rule: {"required":true}

### spec.delimitedText.httpServerLocation.path

`string`

The folder path on the server. Required by the delimited text,
JSON, and binary formats; optional for Parquet.

### spec.delimitedText.httpServerLocation.dynamicPathEnabled

`bool`

Treat path as a Data Factory expression instead of a literal.

### spec.delimitedText.httpServerLocation.filename

`string`

The file's name. Required by every format that uses this
location.

### spec.delimitedText.httpServerLocation.dynamicFilenameEnabled

`bool`

Treat filename as a Data Factory expression instead of a literal.

### spec.delimitedText.azureBlobStorageLocation

`AzureDataFactoryDatasetBlobStorageLocation`

Files in Azure Blob Storage. Set exactly one location block.

### spec.delimitedText.azureBlobStorageLocation.container

`string` · required

The blob container's name.

- rule: {"required":true}

### spec.delimitedText.azureBlobStorageLocation.dynamicContainerEnabled

`bool`

Treat container as a Data Factory expression instead of a
literal.

### spec.delimitedText.azureBlobStorageLocation.path

`string`

The folder path inside the container -- omit to address the
container root. Required by the JSON format.

### spec.delimitedText.azureBlobStorageLocation.dynamicPathEnabled

`bool`

Treat path as a Data Factory expression instead of a literal.

### spec.delimitedText.azureBlobStorageLocation.filename

`string`

The file's name -- omit to address a folder (all files). Required
by the JSON format.

### spec.delimitedText.azureBlobStorageLocation.dynamicFilenameEnabled

`bool`

Treat filename as a Data Factory expression instead of a literal.

### spec.delimitedText.azureBlobFsLocation

`AzureDataFactoryDatasetBlobFsLocation`

Files in Azure Data Lake Storage Gen2. Set exactly one location
block.

### spec.delimitedText.azureBlobFsLocation.fileSystem

`string`

The Data Lake Gen2 file system's name.

### spec.delimitedText.azureBlobFsLocation.dynamicFileSystemEnabled

`bool`

Treat file_system as a Data Factory expression instead of a
literal.

### spec.delimitedText.azureBlobFsLocation.path

`string`

The folder path inside the file system.

### spec.delimitedText.azureBlobFsLocation.dynamicPathEnabled

`bool`

Treat path as a Data Factory expression instead of a literal.

### spec.delimitedText.azureBlobFsLocation.filename

`string`

The file's name.

### spec.delimitedText.azureBlobFsLocation.dynamicFilenameEnabled

`bool`

Treat filename as a Data Factory expression instead of a literal.

### spec.delimitedText.columnDelimiter

`string`

The column separator -- omit for the default ",".

### spec.delimitedText.rowDelimiter

`string`

The row separator -- omit for the service's default (auto-detect
\r, \n, or \r\n).

### spec.delimitedText.quoteCharacter

`string`

The quote character wrapping values -- omit for the default '"'.

### spec.delimitedText.escapeCharacter

`string`

The escape character -- omit for the default "\".

### spec.delimitedText.encoding

`string`

The file encoding, e.g. UTF-8 -- omit for the service default.

### spec.delimitedText.firstRowAsHeader

`bool` · optional (explicit presence)

Whether the first row carries column names. Unspecified applies
false.

- default: `false`

### spec.delimitedText.nullValue

`string`

The string that represents NULL values -- omit for the default
(empty string).

### spec.delimitedText.compressionCodec

`string`

The compression codec the files use -- omit for uncompressed
data.

- rule: {"string":{"in":["","None","bzip2","gzip","deflate","ZipDeflate","TarGzip","Tar","snappy","lz4"]}}

### spec.delimitedText.compressionLevel

`string`

The compression level -- omit for the service default.

- rule: {"string":{"in":["","Optimal","Fastest"]}}

### spec.delimitedText.schemaColumn

`[]AzureDataFactoryDatasetSchemaColumn`

Declared columns -- omit to let Data Factory infer or map at run
time.

### spec.delimitedText.schemaColumn[].name

`string` · required

The column's name.

- rule: {"required":true}

### spec.delimitedText.schemaColumn[].type

`string`

The column's Data Factory interim data type -- omit to leave the
type undeclared.

- rule: {"string":{"in":["","Byte","Byte[]","Boolean","Date","DateTime","DateTimeOffset","Decimal","Double","Guid","Int16","Int32","Int64","Single","String","TimeSpan"]}}

### spec.delimitedText.schemaColumn[].description

`string`

A human-readable description of the column.

### spec.http

`AzureDataFactoryDatasetHttp`

A file served by an HTTP endpoint (through a web linked service).

### spec.http.relativeUrl

`string`

The URL path below the linked service's base URL.

### spec.http.requestBody

`string`

The body sent with the HTTP request.

### spec.http.requestMethod

`string`

The HTTP method, e.g. GET or POST.

### spec.http.schemaColumn

`[]AzureDataFactoryDatasetSchemaColumn`

Declared columns -- omit to let Data Factory infer or map at run
time.

### spec.http.schemaColumn[].name

`string` · required

The column's name.

- rule: {"required":true}

### spec.http.schemaColumn[].type

`string`

The column's Data Factory interim data type -- omit to leave the
type undeclared.

- rule: {"string":{"in":["","Byte","Byte[]","Boolean","Date","DateTime","DateTimeOffset","Decimal","Double","Guid","Int16","Int32","Int64","Single","String","TimeSpan"]}}

### spec.http.schemaColumn[].description

`string`

A human-readable description of the column.

### spec.json

`AzureDataFactoryDatasetJson`

JSON files.

- rule: Set exactly one of http_server_location or azure_blob_storage_location
- rule: http_server_location requires both path and filename for the JSON format
- rule: azure_blob_storage_location requires both path and filename for the JSON format

### spec.json.httpServerLocation

`AzureDataFactoryDatasetHttpServerLocation`

Files on an HTTP server. Set exactly one location block. path and
filename are both required for this format (see the message
rule).

### spec.json.httpServerLocation.relativeUrl

`string` · required

The URL path below the linked service's base URL.

- rule: {"required":true}

### spec.json.httpServerLocation.path

`string`

The folder path on the server. Required by the delimited text,
JSON, and binary formats; optional for Parquet.

### spec.json.httpServerLocation.dynamicPathEnabled

`bool`

Treat path as a Data Factory expression instead of a literal.

### spec.json.httpServerLocation.filename

`string`

The file's name. Required by every format that uses this
location.

### spec.json.httpServerLocation.dynamicFilenameEnabled

`bool`

Treat filename as a Data Factory expression instead of a literal.

### spec.json.azureBlobStorageLocation

`AzureDataFactoryDatasetBlobStorageLocation`

Files in Azure Blob Storage. Set exactly one location block. path
and filename are both required for this format (see the message
rule).

### spec.json.azureBlobStorageLocation.container

`string` · required

The blob container's name.

- rule: {"required":true}

### spec.json.azureBlobStorageLocation.dynamicContainerEnabled

`bool`

Treat container as a Data Factory expression instead of a
literal.

### spec.json.azureBlobStorageLocation.path

`string`

The folder path inside the container -- omit to address the
container root. Required by the JSON format.

### spec.json.azureBlobStorageLocation.dynamicPathEnabled

`bool`

Treat path as a Data Factory expression instead of a literal.

### spec.json.azureBlobStorageLocation.filename

`string`

The file's name -- omit to address a folder (all files). Required
by the JSON format.

### spec.json.azureBlobStorageLocation.dynamicFilenameEnabled

`bool`

Treat filename as a Data Factory expression instead of a literal.

### spec.json.encoding

`string`

The file encoding, e.g. UTF-8 -- omit for the service default.

### spec.json.schemaColumn

`[]AzureDataFactoryDatasetSchemaColumn`

Declared columns -- omit to let Data Factory infer or map at run
time.

### spec.json.schemaColumn[].name

`string` · required

The column's name.

- rule: {"required":true}

### spec.json.schemaColumn[].type

`string`

The column's Data Factory interim data type -- omit to leave the
type undeclared.

- rule: {"string":{"in":["","Byte","Byte[]","Boolean","Date","DateTime","DateTimeOffset","Decimal","Double","Guid","Int16","Int32","Int64","Single","String","TimeSpan"]}}

### spec.json.schemaColumn[].description

`string`

A human-readable description of the column.

### spec.mysql

`AzureDataFactoryDatasetMysql`

A MySQL table.

### spec.mysql.tableName

`string`

The table's name -- omit to leave it undeclared (e.g. when the
pipeline supplies a query instead).

### spec.mysql.schemaColumn

`[]AzureDataFactoryDatasetSchemaColumn`

Declared columns -- omit to let Data Factory infer or map at run
time.

### spec.mysql.schemaColumn[].name

`string` · required

The column's name.

- rule: {"required":true}

### spec.mysql.schemaColumn[].type

`string`

The column's Data Factory interim data type -- omit to leave the
type undeclared.

- rule: {"string":{"in":["","Byte","Byte[]","Boolean","Date","DateTime","DateTimeOffset","Decimal","Double","Guid","Int16","Int32","Int64","Single","String","TimeSpan"]}}

### spec.mysql.schemaColumn[].description

`string`

A human-readable description of the column.

### spec.parquet

`AzureDataFactoryDatasetParquet`

Parquet files.

- rule: Set exactly one of http_server_location, azure_blob_storage_location, or azure_blob_fs_location
- rule: http_server_location requires filename for the Parquet format

### spec.parquet.httpServerLocation

`AzureDataFactoryDatasetHttpServerLocation`

Files on an HTTP server. Set exactly one location block. filename
is required for this format (see the message rule); path is
optional.

### spec.parquet.httpServerLocation.relativeUrl

`string` · required

The URL path below the linked service's base URL.

- rule: {"required":true}

### spec.parquet.httpServerLocation.path

`string`

The folder path on the server. Required by the delimited text,
JSON, and binary formats; optional for Parquet.

### spec.parquet.httpServerLocation.dynamicPathEnabled

`bool`

Treat path as a Data Factory expression instead of a literal.

### spec.parquet.httpServerLocation.filename

`string`

The file's name. Required by every format that uses this
location.

### spec.parquet.httpServerLocation.dynamicFilenameEnabled

`bool`

Treat filename as a Data Factory expression instead of a literal.

### spec.parquet.azureBlobStorageLocation

`AzureDataFactoryDatasetBlobStorageLocation`

Files in Azure Blob Storage. Set exactly one location block.

### spec.parquet.azureBlobStorageLocation.container

`string` · required

The blob container's name.

- rule: {"required":true}

### spec.parquet.azureBlobStorageLocation.dynamicContainerEnabled

`bool`

Treat container as a Data Factory expression instead of a
literal.

### spec.parquet.azureBlobStorageLocation.path

`string`

The folder path inside the container -- omit to address the
container root. Required by the JSON format.

### spec.parquet.azureBlobStorageLocation.dynamicPathEnabled

`bool`

Treat path as a Data Factory expression instead of a literal.

### spec.parquet.azureBlobStorageLocation.filename

`string`

The file's name -- omit to address a folder (all files). Required
by the JSON format.

### spec.parquet.azureBlobStorageLocation.dynamicFilenameEnabled

`bool`

Treat filename as a Data Factory expression instead of a literal.

### spec.parquet.azureBlobFsLocation

`AzureDataFactoryDatasetBlobFsLocation`

Files in Azure Data Lake Storage Gen2. Set exactly one location
block.

### spec.parquet.azureBlobFsLocation.fileSystem

`string`

The Data Lake Gen2 file system's name.

### spec.parquet.azureBlobFsLocation.dynamicFileSystemEnabled

`bool`

Treat file_system as a Data Factory expression instead of a
literal.

### spec.parquet.azureBlobFsLocation.path

`string`

The folder path inside the file system.

### spec.parquet.azureBlobFsLocation.dynamicPathEnabled

`bool`

Treat path as a Data Factory expression instead of a literal.

### spec.parquet.azureBlobFsLocation.filename

`string`

The file's name.

### spec.parquet.azureBlobFsLocation.dynamicFilenameEnabled

`bool`

Treat filename as a Data Factory expression instead of a literal.

### spec.parquet.compressionCodec

`string`

The compression codec the files use -- omit for uncompressed
data. Parquet natively favors snappy.

- rule: {"string":{"in":["","bzip2","gzip","deflate","ZipDeflate","TarGzip","Tar","snappy","lz4"]}}

### spec.parquet.schemaColumn

`[]AzureDataFactoryDatasetSchemaColumn`

Declared columns -- omit to let Data Factory infer or map at run
time.

### spec.parquet.schemaColumn[].name

`string` · required

The column's name.

- rule: {"required":true}

### spec.parquet.schemaColumn[].type

`string`

The column's Data Factory interim data type -- omit to leave the
type undeclared.

- rule: {"string":{"in":["","Byte","Byte[]","Boolean","Date","DateTime","DateTimeOffset","Decimal","Double","Guid","Int16","Int32","Int64","Single","String","TimeSpan"]}}

### spec.parquet.schemaColumn[].description

`string`

A human-readable description of the column.

### spec.postgresql

`AzureDataFactoryDatasetPostgresql`

A PostgreSQL table.

### spec.postgresql.tableName

`string`

The table's name -- omit to leave it undeclared (e.g. when the
pipeline supplies a query instead).

### spec.postgresql.schemaColumn

`[]AzureDataFactoryDatasetSchemaColumn`

Declared columns -- omit to let Data Factory infer or map at run
time.

### spec.postgresql.schemaColumn[].name

`string` · required

The column's name.

- rule: {"required":true}

### spec.postgresql.schemaColumn[].type

`string`

The column's Data Factory interim data type -- omit to leave the
type undeclared.

- rule: {"string":{"in":["","Byte","Byte[]","Boolean","Date","DateTime","DateTimeOffset","Decimal","Double","Guid","Int16","Int32","Int64","Single","String","TimeSpan"]}}

### spec.postgresql.schemaColumn[].description

`string`

A human-readable description of the column.

### spec.snowflake

`AzureDataFactoryDatasetSnowflake`

A Snowflake table.

### spec.snowflake.tableName

`string`

The table's name -- omit to leave it undeclared (e.g. when the
pipeline supplies a query instead).

### spec.snowflake.schemaName

`string`

The Snowflake schema the table lives in -- omit to leave it
undeclared.

### spec.snowflake.schemaColumn

`[]AzureDataFactoryDatasetSnowflakeSchemaColumn`

Declared columns, in Snowflake's own type vocabulary -- omit to
let Data Factory infer or map at run time.

### spec.snowflake.schemaColumn[].name

`string` · required

The column's name.

- rule: {"required":true}

### spec.snowflake.schemaColumn[].type

`string`

The column's Snowflake data type -- omit to leave the type
undeclared.

- rule: {"string":{"in":["","NUMBER","DECIMAL","NUMERIC","INT","INTEGER","BIGINT","SMALLINT","FLOAT","FLOAT4","FLOAT8","DOUBLE","DOUBLE PRECISION","REAL","VARCHAR","CHAR","CHARACTER","STRING","TEXT","BINARY","VARBINARY","BOOLEAN","DATE","DATETIME","TIME","TIMESTAMP","TIMESTAMP_LTZ","TIMESTAMP_NTZ","TIMESTAMP_TZ","VARIANT","OBJECT","ARRAY","GEOGRAPHY"]}}

### spec.snowflake.schemaColumn[].precision

`int32`

The total number of digits, for numeric types.

- rule: {"int32":{"gte":0}}

### spec.snowflake.schemaColumn[].scale

`int32`

The number of digits after the decimal point, for numeric types.

- rule: {"int32":{"gte":0}}

### spec.sqlServerTable

`AzureDataFactoryDatasetSqlServerTable`

A SQL Server table.

### spec.sqlServerTable.tableName

`string`

The table's name -- omit to leave it undeclared (e.g. when the
pipeline supplies a query instead).

### spec.sqlServerTable.schemaColumn

`[]AzureDataFactoryDatasetSchemaColumn`

Declared columns -- omit to let Data Factory infer or map at run
time.

### spec.sqlServerTable.schemaColumn[].name

`string` · required

The column's name.

- rule: {"required":true}

### spec.sqlServerTable.schemaColumn[].type

`string`

The column's Data Factory interim data type -- omit to leave the
type undeclared.

- rule: {"string":{"in":["","Byte","Byte[]","Boolean","Date","DateTime","DateTimeOffset","Decimal","Double","Guid","Int16","Int32","Int64","Single","String","TimeSpan"]}}

### spec.sqlServerTable.schemaColumn[].description

`string`

A human-readable description of the column.

## Validation Rules

- `azure_data_factory_dataset_linked_service_name_required`: linked_service_name is required for every variant except azure_sql_table and custom
- `azure_data_factory_dataset_linked_service_name_conflicts`: azure_sql_table and custom carry their own linked service reference -- do not also set linked_service_name

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AzureDataFactoryDataset, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.dataset_id` | `string` | The dataset's Azure Resource Manager ID ({factory_id}/datasets/{name}) -- the same ID shape for all dataset types. |
| `status.outputs.dataset_name` | `string` | The dataset's name -- what pipelines and data flows resolve against. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.dataFactoryId` | AzureDataFactory | `status.outputs.data_factory_id` |
| `spec.linkedServiceName` | AzureDataFactoryLinkedService | `status.outputs.linked_service_name` |
| `spec.azureSqlTable.linkedServiceId` | AzureDataFactoryLinkedService | `status.outputs.linked_service_id` |
| `spec.custom.linkedService.name` | AzureDataFactoryLinkedService | `status.outputs.linked_service_name` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AzureDataFactoryDataFlow | `spec.sources[].dataset.name` | `status.outputs.dataset_name` |
| AzureDataFactoryDataFlow | `spec.sinks[].dataset.name` | `status.outputs.dataset_name` |
| AzureDataFactoryDataFlow | `spec.transformations[].dataset.name` | `status.outputs.dataset_name` |

## See Also

- [Overview](../README.md)
