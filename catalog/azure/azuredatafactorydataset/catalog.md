# Azure Data Factory Dataset

Deploys one dataset inside an Azure Data Factory -- a named view of data telling pipelines and data flows what the data looks like and where it sits within a system a linked service already connects to. One kind covers all 13 shapes Azure models as first-class dataset resources -- file formats on blob, Data Lake Gen2, HTTP, or SFTP locations, and table forms for the major databases -- plus a raw-JSON custom form for every other dataset type Data Factory speaks. The dataset stores metadata only: creating it never reads the data, and deleting it never touches the data.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions exactly one dataset of the shape the spec's variant block declares:

- **File formats** -- delimited text (CSV), JSON, Parquet, binary, and the flat Azure Blob form, each locating its files through a location block (blob storage, Data Lake Gen2, an HTTP server, or SFTP for binary)
- **HTTP** -- a file served by an HTTP endpoint, through a web linked service
- **Tables** -- Azure SQL Database, SQL Server, MySQL, PostgreSQL, Snowflake
- **Cosmos DB** -- a SQL API collection
- **Custom** -- any other Data Factory dataset type (Excel, XML, Avro, ORC, REST, and dozens more), as its type token plus raw type-properties JSON

All shapes share one factory-scoped name namespace (`{factory_id}/datasets/{name}`) and the same shared fields: description, annotations, parameters, additional properties, and the Studio display folder.

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription.
- **Data Factory** -- the dataset lives in a factory; reference an AzureDataFactory's `data_factory_id` output or provide the ARM ID directly.
- **Linked service** -- every dataset reads through a connection; deploy the matching AzureDataFactoryLinkedService first. Eleven variants reference it by name through `linkedServiceName`; the `azureSqlTable` variant references it by ARM ID inside its own block, and `custom` carries its own reference block.

### Azure Subscription

- **The linked service type must match the dataset shape** -- a delimited text dataset on a blob location needs a blob storage (or Data Lake Gen2) connection; a MySQL dataset needs a MySQL connection. Azure validates the pairing when a pipeline first USES the dataset, not at save time.
- **Saving does not read the data** -- Azure stores the definition without checking the file or table exists; a wrong path surfaces when a pipeline runs. Use Studio's Preview data button after deploy.
- **The Azure SQL table variant's linked service must live in the same factory** as the dataset -- Azure enforces this at deploy time.

## Deploy

### Console

Open the deployment store, find **Azure Data Factory Dataset**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **CSV on Blob Storage**, **SQL Table**, or **Parquet on Data Lake** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureDataFactoryDataset
metadata:
  name: orders-csv
  org: acme-corp
  env: prod
spec:
  dataFactoryId:
    valueFrom:
      kind: AzureDataFactory
      name: data-platform
      fieldPath: status.outputs.data_factory_id
  name: orders-csv
  linkedServiceName:
    valueFrom:
      kind: AzureDataFactoryLinkedService
      name: landing-blob
      fieldPath: status.outputs.linked_service_name
  delimitedText:
    azureBlobStorageLocation:
      container: landing
      path: raw/orders
      filename: orders.csv
    firstRowAsHeader: true
    encoding: UTF-8
```

```shell
planton apply -f data-factory-dataset.yaml
```

This creates one delimited text dataset named `orders-csv` in the factory, reading `landing/raw/orders/orders.csv` through the blob linked service with Azure's default parse settings (`,` delimiter, `"` quote, `\` escape). An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying the factory, its connections, and its datasets as one chart, ValueFromRef wires the dataset to resources deployed in the same InfraPipeline:

```yaml
spec:
  dataFactoryId:
    valueFrom:
      kind: AzureDataFactory
      name: data-platform
      fieldPath: status.outputs.data_factory_id
  linkedServiceName:
    valueFrom:
      kind: AzureDataFactoryLinkedService
      name: landing-blob
      fieldPath: status.outputs.linked_service_name
```

The InfraPipeline resolves the dependency graph -- factory first, then the linked service, then this dataset.

## Key Configuration

These are the most important decisions when configuring a dataset. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The variant block is the type decision** -- set exactly one of the 13 blocks; it selects which provider resource is created. All shapes live in one factory-scoped namespace, so changing a dataset's variant block replaces the object at the same ARM address, and every pipeline activity referencing the name picks up the new shape immediately -- there is no versioning. Rename rather than reshape when pipelines still depend on the old contract. `name`, `dataFactoryId`, and the custom form's `type` are all ForceNew: changing them destroys and recreates the dataset.

**Linked service pairing** -- Azure saves any pairing without complaint (a CSV dataset over a MySQL connection stores fine) and fails only when a pipeline first uses it. The rule is yours to keep: file formats ride storage or web connections, table shapes ride their database's own connection type, the HTTP shape rides a web connection.

**Parameters over dataset-per-partition** -- the `parameters` map plus the `dynamic*Enabled` flags turn literal paths into run-time expressions (`@{dataset().runDate}`), so one dataset serves every partition of a feed instead of one dataset per day. Pipelines override parameter values per activity, and the factory stays navigable.

**Declared columns are a contract, not a requirement** -- Data Factory infers columns happily; declare `schemaColumn` only when downstream mappings need a stable contract, and expect to maintain it when the source evolves -- a stale declared schema fails runs the inference form would have survived. Snowflake declares in its own type vocabulary with precision and scale; every other variant uses Data Factory's interim types.

**Location strictness varies by format** -- each file format takes exactly one location block, but requiredness differs: JSON requires both path and filename in every location shape, delimited text and binary require them only on HTTP locations, and Parquet's HTTP location is the one shape where the folder path may be omitted. A location the spec accepts is not a location that exists -- Preview data is the cheapest check.

**The custom form trades validation for reach** -- it carries any dataset type Data Factory speaks as raw type-properties JSON, and it is the only form whose linked service reference can pass parameter values. Azure validates the JSON at save time against the declared type, but nothing validates it before deploy: copy shapes from the Data Factory REST API exactly, and prefer a first-class variant whenever one exists.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureDataFactory** | `dataFactoryId` | `status.outputs.data_factory_id` |
| **AzureDataFactoryLinkedService** (most variants) | `linkedServiceName` | `status.outputs.linked_service_name` |
| **AzureDataFactoryLinkedService** (azureSqlTable variant) | `azureSqlTable.linkedServiceId` | `status.outputs.linked_service_id` |
| **AzureDataFactoryLinkedService** (custom variant) | `custom.linkedService.name` | `status.outputs.linked_service_name` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `dataset_name` | The dataset's name inside the factory | What pipeline activities and data flow sources/sinks resolve against |
| `dataset_id` | The ARM ID (`{factory_id}/datasets/{name}`) | ARM-level references and import tooling |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Parameterized landing-zone feed** -- a delimited text dataset on blob storage with a run-time-parameterized folder path, so one dataset serves every day's drop and the pipeline supplies the partition value per run. Start from the **CSV on Blob Storage** preset.

**Warehouse table endpoint** -- an Azure SQL Database table dataset, the shape copy activities use as a source or sink when moving data into or out of the warehouse. Leave the table undeclared when the pipeline supplies a query instead. Start from the **SQL Table** preset.

**Lakehouse curated layer** -- a Parquet dataset over Data Lake Gen2, where processed data lands in an analytical format after the raw CSV side is transformed. Start from the **Parquet on Data Lake** preset.

**Custom escape hatch** -- Excel sheets, XML, Avro, REST payloads, and every other type without a first-class block travel as the custom form's raw JSON. It trades pre-deploy validation for reach; keep the JSON small and exact.

## Works With

- [**Azure Data Factory**](/infra-catalog/azure-data-factory) -- the factory the dataset lives in, referenced by `dataFactoryId`
- [**Azure Data Factory Linked Service**](/infra-catalog/azure-data-factory-linked-service) -- the connection every dataset reads through
- [**Azure Data Factory Pipeline**](/infra-catalog/azure-data-factory-pipeline) -- copy and transform activities read from and write to datasets by name
- [**Azure Data Factory Data Flow**](/infra-catalog/azure-data-factory-data-flow) -- data flow sources and sinks bind to datasets
- [**Azure Storage Account**](/infra-catalog/azure-storage-account) -- where blob and Data Lake Gen2 dataset locations physically live
