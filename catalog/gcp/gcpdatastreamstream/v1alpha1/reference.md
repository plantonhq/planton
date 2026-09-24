# GcpDatastreamStream

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

GcpDatastreamStreamSpec defines a Datastream stream
(`google_datastream_stream`) -- continuous change data capture from one
source database into BigQuery or Cloud Storage. A stream reads through a
source connection profile and writes through a destination profile, both
in its project and location.

A stream first BACKFILLS (copies what exists) and then streams every
change. Choose exactly one of backfill_all (optionally skipping objects)
or backfill_none (changes only). Datastream bills the GiB it processes,
backfill and changes at different rates.

desired_state controls whether it runs: create it NOT_STARTED (the
default) to review before any data moves, then set RUNNING; PAUSED stops
it without losing its position.

Immutable: project_id, location, stream_id, both profiles,
customer_managed_encryption_key, create_without_validation, the BigQuery
write_mode, the dataset template's key. Object lists, backfill, rule
sets, freshness, and file settings update in place.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpDatastreamStream
metadata:
  name: orders-cdc
spec:
  projectId:
    value: my-gcp-project
  location: us-central1
  displayName: Orders CDC to BigQuery
  sourceConfig:
    sourceConnectionProfile:
      value: projects/my-gcp-project/locations/us-central1/connectionProfiles/orders-postgres
    postgresqlSourceConfig:
      replicationSlot: datastream_slot
      publication: datastream_pub
      includeObjects:
        postgresqlSchemas:
          - schema: public
            postgresqlTables:
              - table: orders
              - table: order_items
  destinationConfig:
    destinationConnectionProfile:
      value: projects/my-gcp-project/locations/us-central1/connectionProfiles/bigquery
    bigqueryDestinationConfig:
      sourceHierarchyDatasets:
        datasetTemplate:
          location: US
          datasetIdPrefix: orders
      dataFreshness: 900s
      writeMode: MERGE
  backfillAll: {}
  desiredState: RUNNING
  labels:
    team: data
  deletionPolicy: DELETE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.location` | `string` | yes |  |  |
| `spec.streamId` | `string` |  |  |  |
| `spec.displayName` | `string` |  |  |  |
| `spec.labels` | `map<string, string>` |  |  |  |
| `spec.sourceConfig` | `GcpDatastreamStreamSourceConfig` | yes |  |  |
| `spec.sourceConfig.sourceConnectionProfile` | `string \| valueFrom` | yes |  | GcpDatastreamConnectionProfile (`status.outputs.name`) |
| `spec.sourceConfig.mysqlSourceConfig` | `GcpDatastreamStreamMysqlSourceConfig` |  |  |  |
| `spec.sourceConfig.mysqlSourceConfig.includeObjects` | `GcpDatastreamStreamMysqlRdbms` |  |  |  |
| `spec.sourceConfig.mysqlSourceConfig.includeObjects.mysqlDatabases` | `[]GcpDatastreamStreamMysqlDatabase` | yes |  |  |
| `spec.sourceConfig.mysqlSourceConfig.includeObjects.mysqlDatabases[].database` | `string` | yes |  |  |
| `spec.sourceConfig.mysqlSourceConfig.includeObjects.mysqlDatabases[].mysqlTables` | `[]GcpDatastreamStreamMysqlTable` |  |  |  |
| `spec.sourceConfig.mysqlSourceConfig.includeObjects.mysqlDatabases[].mysqlTables[].table` | `string` | yes |  |  |
| `spec.sourceConfig.mysqlSourceConfig.includeObjects.mysqlDatabases[].mysqlTables[].mysqlColumns` | `[]GcpDatastreamStreamMysqlColumn` |  |  |  |
| `spec.sourceConfig.mysqlSourceConfig.includeObjects.mysqlDatabases[].mysqlTables[].mysqlColumns[].column` | `string` |  |  |  |
| `spec.sourceConfig.mysqlSourceConfig.includeObjects.mysqlDatabases[].mysqlTables[].mysqlColumns[].collation` | `string` |  |  |  |
| `spec.sourceConfig.mysqlSourceConfig.includeObjects.mysqlDatabases[].mysqlTables[].mysqlColumns[].dataType` | `string` |  |  |  |
| `spec.sourceConfig.mysqlSourceConfig.includeObjects.mysqlDatabases[].mysqlTables[].mysqlColumns[].nullable` | `bool` |  |  |  |
| `spec.sourceConfig.mysqlSourceConfig.includeObjects.mysqlDatabases[].mysqlTables[].mysqlColumns[].ordinalPosition` | `int32` |  |  |  |
| `spec.sourceConfig.mysqlSourceConfig.includeObjects.mysqlDatabases[].mysqlTables[].mysqlColumns[].primaryKey` | `bool` |  |  |  |
| `spec.sourceConfig.mysqlSourceConfig.excludeObjects` | `GcpDatastreamStreamMysqlRdbms` |  |  |  |
| `spec.sourceConfig.mysqlSourceConfig.excludeObjects.mysqlDatabases` | `[]GcpDatastreamStreamMysqlDatabase` | yes |  |  |
| `spec.sourceConfig.mysqlSourceConfig.excludeObjects.mysqlDatabases[].database` | `string` | yes |  |  |
| `spec.sourceConfig.mysqlSourceConfig.excludeObjects.mysqlDatabases[].mysqlTables` | `[]GcpDatastreamStreamMysqlTable` |  |  |  |
| `spec.sourceConfig.mysqlSourceConfig.excludeObjects.mysqlDatabases[].mysqlTables[].table` | `string` | yes |  |  |
| `spec.sourceConfig.mysqlSourceConfig.excludeObjects.mysqlDatabases[].mysqlTables[].mysqlColumns` | `[]GcpDatastreamStreamMysqlColumn` |  |  |  |
| `spec.sourceConfig.mysqlSourceConfig.excludeObjects.mysqlDatabases[].mysqlTables[].mysqlColumns[].column` | `string` |  |  |  |
| `spec.sourceConfig.mysqlSourceConfig.excludeObjects.mysqlDatabases[].mysqlTables[].mysqlColumns[].collation` | `string` |  |  |  |
| `spec.sourceConfig.mysqlSourceConfig.excludeObjects.mysqlDatabases[].mysqlTables[].mysqlColumns[].dataType` | `string` |  |  |  |
| `spec.sourceConfig.mysqlSourceConfig.excludeObjects.mysqlDatabases[].mysqlTables[].mysqlColumns[].nullable` | `bool` |  |  |  |
| `spec.sourceConfig.mysqlSourceConfig.excludeObjects.mysqlDatabases[].mysqlTables[].mysqlColumns[].ordinalPosition` | `int32` |  |  |  |
| `spec.sourceConfig.mysqlSourceConfig.excludeObjects.mysqlDatabases[].mysqlTables[].mysqlColumns[].primaryKey` | `bool` |  |  |  |
| `spec.sourceConfig.mysqlSourceConfig.maxConcurrentBackfillTasks` | `int32` |  |  |  |
| `spec.sourceConfig.mysqlSourceConfig.maxConcurrentCdcTasks` | `int32` |  |  |  |
| `spec.sourceConfig.mysqlSourceConfig.cdcMethod` | `string` |  |  |  |
| `spec.sourceConfig.postgresqlSourceConfig` | `GcpDatastreamStreamPostgresqlSourceConfig` |  |  |  |
| `spec.sourceConfig.postgresqlSourceConfig.includeObjects` | `GcpDatastreamStreamPostgresqlRdbms` |  |  |  |
| `spec.sourceConfig.postgresqlSourceConfig.includeObjects.postgresqlSchemas` | `[]GcpDatastreamStreamPostgresqlSchema` | yes |  |  |
| `spec.sourceConfig.postgresqlSourceConfig.includeObjects.postgresqlSchemas[].schema` | `string` | yes |  |  |
| `spec.sourceConfig.postgresqlSourceConfig.includeObjects.postgresqlSchemas[].postgresqlTables` | `[]GcpDatastreamStreamPostgresqlTable` |  |  |  |
| `spec.sourceConfig.postgresqlSourceConfig.includeObjects.postgresqlSchemas[].postgresqlTables[].table` | `string` | yes |  |  |
| `spec.sourceConfig.postgresqlSourceConfig.includeObjects.postgresqlSchemas[].postgresqlTables[].postgresqlColumns` | `[]GcpDatastreamStreamPostgresqlColumn` |  |  |  |
| `spec.sourceConfig.postgresqlSourceConfig.includeObjects.postgresqlSchemas[].postgresqlTables[].postgresqlColumns[].column` | `string` |  |  |  |
| `spec.sourceConfig.postgresqlSourceConfig.includeObjects.postgresqlSchemas[].postgresqlTables[].postgresqlColumns[].dataType` | `string` |  |  |  |
| `spec.sourceConfig.postgresqlSourceConfig.includeObjects.postgresqlSchemas[].postgresqlTables[].postgresqlColumns[].nullable` | `bool` |  |  |  |
| `spec.sourceConfig.postgresqlSourceConfig.includeObjects.postgresqlSchemas[].postgresqlTables[].postgresqlColumns[].ordinalPosition` | `int32` |  |  |  |
| `spec.sourceConfig.postgresqlSourceConfig.includeObjects.postgresqlSchemas[].postgresqlTables[].postgresqlColumns[].primaryKey` | `bool` |  |  |  |
| `spec.sourceConfig.postgresqlSourceConfig.excludeObjects` | `GcpDatastreamStreamPostgresqlRdbms` |  |  |  |
| `spec.sourceConfig.postgresqlSourceConfig.excludeObjects.postgresqlSchemas` | `[]GcpDatastreamStreamPostgresqlSchema` | yes |  |  |
| `spec.sourceConfig.postgresqlSourceConfig.excludeObjects.postgresqlSchemas[].schema` | `string` | yes |  |  |
| `spec.sourceConfig.postgresqlSourceConfig.excludeObjects.postgresqlSchemas[].postgresqlTables` | `[]GcpDatastreamStreamPostgresqlTable` |  |  |  |
| `spec.sourceConfig.postgresqlSourceConfig.excludeObjects.postgresqlSchemas[].postgresqlTables[].table` | `string` | yes |  |  |
| `spec.sourceConfig.postgresqlSourceConfig.excludeObjects.postgresqlSchemas[].postgresqlTables[].postgresqlColumns` | `[]GcpDatastreamStreamPostgresqlColumn` |  |  |  |
| `spec.sourceConfig.postgresqlSourceConfig.excludeObjects.postgresqlSchemas[].postgresqlTables[].postgresqlColumns[].column` | `string` |  |  |  |
| `spec.sourceConfig.postgresqlSourceConfig.excludeObjects.postgresqlSchemas[].postgresqlTables[].postgresqlColumns[].dataType` | `string` |  |  |  |
| `spec.sourceConfig.postgresqlSourceConfig.excludeObjects.postgresqlSchemas[].postgresqlTables[].postgresqlColumns[].nullable` | `bool` |  |  |  |
| `spec.sourceConfig.postgresqlSourceConfig.excludeObjects.postgresqlSchemas[].postgresqlTables[].postgresqlColumns[].ordinalPosition` | `int32` |  |  |  |
| `spec.sourceConfig.postgresqlSourceConfig.excludeObjects.postgresqlSchemas[].postgresqlTables[].postgresqlColumns[].primaryKey` | `bool` |  |  |  |
| `spec.sourceConfig.postgresqlSourceConfig.replicationSlot` | `string` | yes |  |  |
| `spec.sourceConfig.postgresqlSourceConfig.publication` | `string` | yes |  |  |
| `spec.sourceConfig.postgresqlSourceConfig.maxConcurrentBackfillTasks` | `int32` |  |  |  |
| `spec.sourceConfig.oracleSourceConfig` | `GcpDatastreamStreamOracleSourceConfig` |  |  |  |
| `spec.sourceConfig.oracleSourceConfig.includeObjects` | `GcpDatastreamStreamOracleRdbms` |  |  |  |
| `spec.sourceConfig.oracleSourceConfig.includeObjects.oracleSchemas` | `[]GcpDatastreamStreamOracleSchema` | yes |  |  |
| `spec.sourceConfig.oracleSourceConfig.includeObjects.oracleSchemas[].schema` | `string` | yes |  |  |
| `spec.sourceConfig.oracleSourceConfig.includeObjects.oracleSchemas[].oracleTables` | `[]GcpDatastreamStreamOracleTable` |  |  |  |
| `spec.sourceConfig.oracleSourceConfig.includeObjects.oracleSchemas[].oracleTables[].table` | `string` | yes |  |  |
| `spec.sourceConfig.oracleSourceConfig.includeObjects.oracleSchemas[].oracleTables[].oracleColumns` | `[]GcpDatastreamStreamOracleColumn` |  |  |  |
| `spec.sourceConfig.oracleSourceConfig.includeObjects.oracleSchemas[].oracleTables[].oracleColumns[].column` | `string` |  |  |  |
| `spec.sourceConfig.oracleSourceConfig.includeObjects.oracleSchemas[].oracleTables[].oracleColumns[].dataType` | `string` |  |  |  |
| `spec.sourceConfig.oracleSourceConfig.excludeObjects` | `GcpDatastreamStreamOracleRdbms` |  |  |  |
| `spec.sourceConfig.oracleSourceConfig.excludeObjects.oracleSchemas` | `[]GcpDatastreamStreamOracleSchema` | yes |  |  |
| `spec.sourceConfig.oracleSourceConfig.excludeObjects.oracleSchemas[].schema` | `string` | yes |  |  |
| `spec.sourceConfig.oracleSourceConfig.excludeObjects.oracleSchemas[].oracleTables` | `[]GcpDatastreamStreamOracleTable` |  |  |  |
| `spec.sourceConfig.oracleSourceConfig.excludeObjects.oracleSchemas[].oracleTables[].table` | `string` | yes |  |  |
| `spec.sourceConfig.oracleSourceConfig.excludeObjects.oracleSchemas[].oracleTables[].oracleColumns` | `[]GcpDatastreamStreamOracleColumn` |  |  |  |
| `spec.sourceConfig.oracleSourceConfig.excludeObjects.oracleSchemas[].oracleTables[].oracleColumns[].column` | `string` |  |  |  |
| `spec.sourceConfig.oracleSourceConfig.excludeObjects.oracleSchemas[].oracleTables[].oracleColumns[].dataType` | `string` |  |  |  |
| `spec.sourceConfig.oracleSourceConfig.maxConcurrentBackfillTasks` | `int32` |  |  |  |
| `spec.sourceConfig.oracleSourceConfig.maxConcurrentCdcTasks` | `int32` |  |  |  |
| `spec.sourceConfig.oracleSourceConfig.largeObjectsHandling` | `string` |  |  |  |
| `spec.sourceConfig.sqlServerSourceConfig` | `GcpDatastreamStreamSqlServerSourceConfig` |  |  |  |
| `spec.sourceConfig.sqlServerSourceConfig.includeObjects` | `GcpDatastreamStreamSqlServerRdbms` |  |  |  |
| `spec.sourceConfig.sqlServerSourceConfig.includeObjects.schemas` | `[]GcpDatastreamStreamSqlServerSchema` | yes |  |  |
| `spec.sourceConfig.sqlServerSourceConfig.includeObjects.schemas[].schema` | `string` | yes |  |  |
| `spec.sourceConfig.sqlServerSourceConfig.includeObjects.schemas[].tables` | `[]GcpDatastreamStreamSqlServerTable` |  |  |  |
| `spec.sourceConfig.sqlServerSourceConfig.includeObjects.schemas[].tables[].table` | `string` | yes |  |  |
| `spec.sourceConfig.sqlServerSourceConfig.includeObjects.schemas[].tables[].columns` | `[]GcpDatastreamStreamSqlServerColumn` |  |  |  |
| `spec.sourceConfig.sqlServerSourceConfig.includeObjects.schemas[].tables[].columns[].column` | `string` |  |  |  |
| `spec.sourceConfig.sqlServerSourceConfig.includeObjects.schemas[].tables[].columns[].dataType` | `string` |  |  |  |
| `spec.sourceConfig.sqlServerSourceConfig.excludeObjects` | `GcpDatastreamStreamSqlServerRdbms` |  |  |  |
| `spec.sourceConfig.sqlServerSourceConfig.excludeObjects.schemas` | `[]GcpDatastreamStreamSqlServerSchema` | yes |  |  |
| `spec.sourceConfig.sqlServerSourceConfig.excludeObjects.schemas[].schema` | `string` | yes |  |  |
| `spec.sourceConfig.sqlServerSourceConfig.excludeObjects.schemas[].tables` | `[]GcpDatastreamStreamSqlServerTable` |  |  |  |
| `spec.sourceConfig.sqlServerSourceConfig.excludeObjects.schemas[].tables[].table` | `string` | yes |  |  |
| `spec.sourceConfig.sqlServerSourceConfig.excludeObjects.schemas[].tables[].columns` | `[]GcpDatastreamStreamSqlServerColumn` |  |  |  |
| `spec.sourceConfig.sqlServerSourceConfig.excludeObjects.schemas[].tables[].columns[].column` | `string` |  |  |  |
| `spec.sourceConfig.sqlServerSourceConfig.excludeObjects.schemas[].tables[].columns[].dataType` | `string` |  |  |  |
| `spec.sourceConfig.sqlServerSourceConfig.maxConcurrentBackfillTasks` | `int32` |  |  |  |
| `spec.sourceConfig.sqlServerSourceConfig.maxConcurrentCdcTasks` | `int32` |  |  |  |
| `spec.sourceConfig.sqlServerSourceConfig.cdcMethod` | `string` |  |  |  |
| `spec.sourceConfig.mongodbSourceConfig` | `GcpDatastreamStreamMongodbSourceConfig` |  |  |  |
| `spec.sourceConfig.mongodbSourceConfig.includeObjects` | `GcpDatastreamStreamMongodbCluster` |  |  |  |
| `spec.sourceConfig.mongodbSourceConfig.includeObjects.databases` | `[]GcpDatastreamStreamMongodbDatabase` |  |  |  |
| `spec.sourceConfig.mongodbSourceConfig.includeObjects.databases[].database` | `string` |  |  |  |
| `spec.sourceConfig.mongodbSourceConfig.includeObjects.databases[].collections` | `[]GcpDatastreamStreamMongodbCollection` |  |  |  |
| `spec.sourceConfig.mongodbSourceConfig.includeObjects.databases[].collections[].collection` | `string` |  |  |  |
| `spec.sourceConfig.mongodbSourceConfig.includeObjects.databases[].collections[].fields` | `[]GcpDatastreamStreamMongodbField` |  |  |  |
| `spec.sourceConfig.mongodbSourceConfig.includeObjects.databases[].collections[].fields[].field` | `string` |  |  |  |
| `spec.sourceConfig.mongodbSourceConfig.excludeObjects` | `GcpDatastreamStreamMongodbCluster` |  |  |  |
| `spec.sourceConfig.mongodbSourceConfig.excludeObjects.databases` | `[]GcpDatastreamStreamMongodbDatabase` |  |  |  |
| `spec.sourceConfig.mongodbSourceConfig.excludeObjects.databases[].database` | `string` |  |  |  |
| `spec.sourceConfig.mongodbSourceConfig.excludeObjects.databases[].collections` | `[]GcpDatastreamStreamMongodbCollection` |  |  |  |
| `spec.sourceConfig.mongodbSourceConfig.excludeObjects.databases[].collections[].collection` | `string` |  |  |  |
| `spec.sourceConfig.mongodbSourceConfig.excludeObjects.databases[].collections[].fields` | `[]GcpDatastreamStreamMongodbField` |  |  |  |
| `spec.sourceConfig.mongodbSourceConfig.excludeObjects.databases[].collections[].fields[].field` | `string` |  |  |  |
| `spec.sourceConfig.mongodbSourceConfig.maxConcurrentBackfillTasks` | `int32` |  |  |  |
| `spec.sourceConfig.salesforceSourceConfig` | `GcpDatastreamStreamSalesforceSourceConfig` |  |  |  |
| `spec.sourceConfig.salesforceSourceConfig.includeObjects` | `GcpDatastreamStreamSalesforceOrg` |  |  |  |
| `spec.sourceConfig.salesforceSourceConfig.includeObjects.objects` | `[]GcpDatastreamStreamSalesforceObject` | yes |  |  |
| `spec.sourceConfig.salesforceSourceConfig.includeObjects.objects[].objectName` | `string` |  |  |  |
| `spec.sourceConfig.salesforceSourceConfig.includeObjects.objects[].fields` | `[]GcpDatastreamStreamSalesforceField` |  |  |  |
| `spec.sourceConfig.salesforceSourceConfig.includeObjects.objects[].fields[].name` | `string` |  |  |  |
| `spec.sourceConfig.salesforceSourceConfig.excludeObjects` | `GcpDatastreamStreamSalesforceOrg` |  |  |  |
| `spec.sourceConfig.salesforceSourceConfig.excludeObjects.objects` | `[]GcpDatastreamStreamSalesforceObject` | yes |  |  |
| `spec.sourceConfig.salesforceSourceConfig.excludeObjects.objects[].objectName` | `string` |  |  |  |
| `spec.sourceConfig.salesforceSourceConfig.excludeObjects.objects[].fields` | `[]GcpDatastreamStreamSalesforceField` |  |  |  |
| `spec.sourceConfig.salesforceSourceConfig.excludeObjects.objects[].fields[].name` | `string` |  |  |  |
| `spec.sourceConfig.salesforceSourceConfig.pollingInterval` | `string` | yes |  |  |
| `spec.sourceConfig.spannerSourceConfig` | `GcpDatastreamStreamSpannerSourceConfig` |  |  |  |
| `spec.sourceConfig.spannerSourceConfig.includeObjects` | `GcpDatastreamStreamSpannerDatabase` |  |  |  |
| `spec.sourceConfig.spannerSourceConfig.includeObjects.schemas` | `[]GcpDatastreamStreamSpannerSchema` | yes |  |  |
| `spec.sourceConfig.spannerSourceConfig.includeObjects.schemas[].schema` | `string` | yes |  |  |
| `spec.sourceConfig.spannerSourceConfig.includeObjects.schemas[].tables` | `[]GcpDatastreamStreamSpannerTable` |  |  |  |
| `spec.sourceConfig.spannerSourceConfig.includeObjects.schemas[].tables[].table` | `string` | yes |  |  |
| `spec.sourceConfig.spannerSourceConfig.includeObjects.schemas[].tables[].columns` | `[]GcpDatastreamStreamSpannerColumn` |  |  |  |
| `spec.sourceConfig.spannerSourceConfig.includeObjects.schemas[].tables[].columns[].column` | `string` |  |  |  |
| `spec.sourceConfig.spannerSourceConfig.excludeObjects` | `GcpDatastreamStreamSpannerDatabase` |  |  |  |
| `spec.sourceConfig.spannerSourceConfig.excludeObjects.schemas` | `[]GcpDatastreamStreamSpannerSchema` | yes |  |  |
| `spec.sourceConfig.spannerSourceConfig.excludeObjects.schemas[].schema` | `string` | yes |  |  |
| `spec.sourceConfig.spannerSourceConfig.excludeObjects.schemas[].tables` | `[]GcpDatastreamStreamSpannerTable` |  |  |  |
| `spec.sourceConfig.spannerSourceConfig.excludeObjects.schemas[].tables[].table` | `string` | yes |  |  |
| `spec.sourceConfig.spannerSourceConfig.excludeObjects.schemas[].tables[].columns` | `[]GcpDatastreamStreamSpannerColumn` |  |  |  |
| `spec.sourceConfig.spannerSourceConfig.excludeObjects.schemas[].tables[].columns[].column` | `string` |  |  |  |
| `spec.sourceConfig.spannerSourceConfig.changeStreamName` | `string` | yes |  |  |
| `spec.sourceConfig.spannerSourceConfig.backfillDataBoostEnabled` | `bool` |  |  |  |
| `spec.sourceConfig.spannerSourceConfig.fgacRole` | `string` |  |  |  |
| `spec.sourceConfig.spannerSourceConfig.maxConcurrentBackfillTasks` | `int32` |  |  |  |
| `spec.sourceConfig.spannerSourceConfig.maxConcurrentCdcTasks` | `int32` |  |  |  |
| `spec.sourceConfig.spannerSourceConfig.spannerRpcPriority` | `string` |  |  |  |
| `spec.destinationConfig` | `GcpDatastreamStreamDestinationConfig` | yes |  |  |
| `spec.destinationConfig.destinationConnectionProfile` | `string \| valueFrom` | yes |  | GcpDatastreamConnectionProfile (`status.outputs.name`) |
| `spec.destinationConfig.bigqueryDestinationConfig` | `GcpDatastreamStreamBigqueryDestinationConfig` |  |  |  |
| `spec.destinationConfig.bigqueryDestinationConfig.singleTargetDataset` | `GcpDatastreamStreamSingleTargetDataset` |  |  |  |
| `spec.destinationConfig.bigqueryDestinationConfig.singleTargetDataset.datasetId` | `string \| valueFrom` | yes |  | GcpBigQueryDataset (`status.outputs.self_link`) |
| `spec.destinationConfig.bigqueryDestinationConfig.sourceHierarchyDatasets` | `GcpDatastreamStreamSourceHierarchyDatasets` |  |  |  |
| `spec.destinationConfig.bigqueryDestinationConfig.sourceHierarchyDatasets.datasetTemplate` | `GcpDatastreamStreamDatasetTemplate` | yes |  |  |
| `spec.destinationConfig.bigqueryDestinationConfig.sourceHierarchyDatasets.datasetTemplate.location` | `string` | yes |  |  |
| `spec.destinationConfig.bigqueryDestinationConfig.sourceHierarchyDatasets.datasetTemplate.datasetIdPrefix` | `string` |  |  |  |
| `spec.destinationConfig.bigqueryDestinationConfig.sourceHierarchyDatasets.datasetTemplate.kmsKeyName` | `string \| valueFrom` |  |  | GcpKmsKey (`status.outputs.key_id`) |
| `spec.destinationConfig.bigqueryDestinationConfig.sourceHierarchyDatasets.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.destinationConfig.bigqueryDestinationConfig.dataFreshness` | `string` |  |  |  |
| `spec.destinationConfig.bigqueryDestinationConfig.writeMode` | `string` |  |  |  |
| `spec.destinationConfig.bigqueryDestinationConfig.blmtConfig` | `GcpDatastreamStreamBlmtConfig` |  |  |  |
| `spec.destinationConfig.bigqueryDestinationConfig.blmtConfig.bucket` | `string \| valueFrom` | yes |  | GcpGcsBucket (`status.outputs.bucket_name`) |
| `spec.destinationConfig.bigqueryDestinationConfig.blmtConfig.rootPath` | `string` |  |  |  |
| `spec.destinationConfig.bigqueryDestinationConfig.blmtConfig.connectionName` | `string \| valueFrom` | yes |  | GcpBigQueryConnection (`status.outputs.name`) |
| `spec.destinationConfig.bigqueryDestinationConfig.blmtConfig.fileFormat` | `string` | yes |  |  |
| `spec.destinationConfig.bigqueryDestinationConfig.blmtConfig.tableFormat` | `string` | yes |  |  |
| `spec.destinationConfig.gcsDestinationConfig` | `GcpDatastreamStreamGcsDestinationConfig` |  |  |  |
| `spec.destinationConfig.gcsDestinationConfig.path` | `string` |  |  |  |
| `spec.destinationConfig.gcsDestinationConfig.fileRotationInterval` | `string` |  |  |  |
| `spec.destinationConfig.gcsDestinationConfig.fileRotationMb` | `int32` |  |  |  |
| `spec.destinationConfig.gcsDestinationConfig.avroFileFormat` | `bool` |  |  |  |
| `spec.destinationConfig.gcsDestinationConfig.jsonFileFormat` | `GcpDatastreamStreamGcsJsonFileFormat` |  |  |  |
| `spec.destinationConfig.gcsDestinationConfig.jsonFileFormat.compression` | `string` |  |  |  |
| `spec.destinationConfig.gcsDestinationConfig.jsonFileFormat.schemaFileFormat` | `string` |  |  |  |
| `spec.backfillAll` | `GcpDatastreamStreamBackfillAll` |  |  |  |
| `spec.backfillAll.mysqlExcludedObjects` | `GcpDatastreamStreamMysqlRdbms` |  |  |  |
| `spec.backfillAll.mysqlExcludedObjects.mysqlDatabases` | `[]GcpDatastreamStreamMysqlDatabase` | yes |  |  |
| `spec.backfillAll.mysqlExcludedObjects.mysqlDatabases[].database` | `string` | yes |  |  |
| `spec.backfillAll.mysqlExcludedObjects.mysqlDatabases[].mysqlTables` | `[]GcpDatastreamStreamMysqlTable` |  |  |  |
| `spec.backfillAll.mysqlExcludedObjects.mysqlDatabases[].mysqlTables[].table` | `string` | yes |  |  |
| `spec.backfillAll.mysqlExcludedObjects.mysqlDatabases[].mysqlTables[].mysqlColumns` | `[]GcpDatastreamStreamMysqlColumn` |  |  |  |
| `spec.backfillAll.mysqlExcludedObjects.mysqlDatabases[].mysqlTables[].mysqlColumns[].column` | `string` |  |  |  |
| `spec.backfillAll.mysqlExcludedObjects.mysqlDatabases[].mysqlTables[].mysqlColumns[].collation` | `string` |  |  |  |
| `spec.backfillAll.mysqlExcludedObjects.mysqlDatabases[].mysqlTables[].mysqlColumns[].dataType` | `string` |  |  |  |
| `spec.backfillAll.mysqlExcludedObjects.mysqlDatabases[].mysqlTables[].mysqlColumns[].nullable` | `bool` |  |  |  |
| `spec.backfillAll.mysqlExcludedObjects.mysqlDatabases[].mysqlTables[].mysqlColumns[].ordinalPosition` | `int32` |  |  |  |
| `spec.backfillAll.mysqlExcludedObjects.mysqlDatabases[].mysqlTables[].mysqlColumns[].primaryKey` | `bool` |  |  |  |
| `spec.backfillAll.postgresqlExcludedObjects` | `GcpDatastreamStreamPostgresqlRdbms` |  |  |  |
| `spec.backfillAll.postgresqlExcludedObjects.postgresqlSchemas` | `[]GcpDatastreamStreamPostgresqlSchema` | yes |  |  |
| `spec.backfillAll.postgresqlExcludedObjects.postgresqlSchemas[].schema` | `string` | yes |  |  |
| `spec.backfillAll.postgresqlExcludedObjects.postgresqlSchemas[].postgresqlTables` | `[]GcpDatastreamStreamPostgresqlTable` |  |  |  |
| `spec.backfillAll.postgresqlExcludedObjects.postgresqlSchemas[].postgresqlTables[].table` | `string` | yes |  |  |
| `spec.backfillAll.postgresqlExcludedObjects.postgresqlSchemas[].postgresqlTables[].postgresqlColumns` | `[]GcpDatastreamStreamPostgresqlColumn` |  |  |  |
| `spec.backfillAll.postgresqlExcludedObjects.postgresqlSchemas[].postgresqlTables[].postgresqlColumns[].column` | `string` |  |  |  |
| `spec.backfillAll.postgresqlExcludedObjects.postgresqlSchemas[].postgresqlTables[].postgresqlColumns[].dataType` | `string` |  |  |  |
| `spec.backfillAll.postgresqlExcludedObjects.postgresqlSchemas[].postgresqlTables[].postgresqlColumns[].nullable` | `bool` |  |  |  |
| `spec.backfillAll.postgresqlExcludedObjects.postgresqlSchemas[].postgresqlTables[].postgresqlColumns[].ordinalPosition` | `int32` |  |  |  |
| `spec.backfillAll.postgresqlExcludedObjects.postgresqlSchemas[].postgresqlTables[].postgresqlColumns[].primaryKey` | `bool` |  |  |  |
| `spec.backfillAll.oracleExcludedObjects` | `GcpDatastreamStreamOracleRdbms` |  |  |  |
| `spec.backfillAll.oracleExcludedObjects.oracleSchemas` | `[]GcpDatastreamStreamOracleSchema` | yes |  |  |
| `spec.backfillAll.oracleExcludedObjects.oracleSchemas[].schema` | `string` | yes |  |  |
| `spec.backfillAll.oracleExcludedObjects.oracleSchemas[].oracleTables` | `[]GcpDatastreamStreamOracleTable` |  |  |  |
| `spec.backfillAll.oracleExcludedObjects.oracleSchemas[].oracleTables[].table` | `string` | yes |  |  |
| `spec.backfillAll.oracleExcludedObjects.oracleSchemas[].oracleTables[].oracleColumns` | `[]GcpDatastreamStreamOracleColumn` |  |  |  |
| `spec.backfillAll.oracleExcludedObjects.oracleSchemas[].oracleTables[].oracleColumns[].column` | `string` |  |  |  |
| `spec.backfillAll.oracleExcludedObjects.oracleSchemas[].oracleTables[].oracleColumns[].dataType` | `string` |  |  |  |
| `spec.backfillAll.sqlServerExcludedObjects` | `GcpDatastreamStreamSqlServerRdbms` |  |  |  |
| `spec.backfillAll.sqlServerExcludedObjects.schemas` | `[]GcpDatastreamStreamSqlServerSchema` | yes |  |  |
| `spec.backfillAll.sqlServerExcludedObjects.schemas[].schema` | `string` | yes |  |  |
| `spec.backfillAll.sqlServerExcludedObjects.schemas[].tables` | `[]GcpDatastreamStreamSqlServerTable` |  |  |  |
| `spec.backfillAll.sqlServerExcludedObjects.schemas[].tables[].table` | `string` | yes |  |  |
| `spec.backfillAll.sqlServerExcludedObjects.schemas[].tables[].columns` | `[]GcpDatastreamStreamSqlServerColumn` |  |  |  |
| `spec.backfillAll.sqlServerExcludedObjects.schemas[].tables[].columns[].column` | `string` |  |  |  |
| `spec.backfillAll.sqlServerExcludedObjects.schemas[].tables[].columns[].dataType` | `string` |  |  |  |
| `spec.backfillAll.mongodbExcludedObjects` | `GcpDatastreamStreamMongodbCluster` |  |  |  |
| `spec.backfillAll.mongodbExcludedObjects.databases` | `[]GcpDatastreamStreamMongodbDatabase` |  |  |  |
| `spec.backfillAll.mongodbExcludedObjects.databases[].database` | `string` |  |  |  |
| `spec.backfillAll.mongodbExcludedObjects.databases[].collections` | `[]GcpDatastreamStreamMongodbCollection` |  |  |  |
| `spec.backfillAll.mongodbExcludedObjects.databases[].collections[].collection` | `string` |  |  |  |
| `spec.backfillAll.mongodbExcludedObjects.databases[].collections[].fields` | `[]GcpDatastreamStreamMongodbField` |  |  |  |
| `spec.backfillAll.mongodbExcludedObjects.databases[].collections[].fields[].field` | `string` |  |  |  |
| `spec.backfillAll.salesforceExcludedObjects` | `GcpDatastreamStreamSalesforceOrg` |  |  |  |
| `spec.backfillAll.salesforceExcludedObjects.objects` | `[]GcpDatastreamStreamSalesforceObject` | yes |  |  |
| `spec.backfillAll.salesforceExcludedObjects.objects[].objectName` | `string` |  |  |  |
| `spec.backfillAll.salesforceExcludedObjects.objects[].fields` | `[]GcpDatastreamStreamSalesforceField` |  |  |  |
| `spec.backfillAll.salesforceExcludedObjects.objects[].fields[].name` | `string` |  |  |  |
| `spec.backfillAll.spannerExcludedObjects` | `GcpDatastreamStreamSpannerDatabase` |  |  |  |
| `spec.backfillAll.spannerExcludedObjects.schemas` | `[]GcpDatastreamStreamSpannerSchema` | yes |  |  |
| `spec.backfillAll.spannerExcludedObjects.schemas[].schema` | `string` | yes |  |  |
| `spec.backfillAll.spannerExcludedObjects.schemas[].tables` | `[]GcpDatastreamStreamSpannerTable` |  |  |  |
| `spec.backfillAll.spannerExcludedObjects.schemas[].tables[].table` | `string` | yes |  |  |
| `spec.backfillAll.spannerExcludedObjects.schemas[].tables[].columns` | `[]GcpDatastreamStreamSpannerColumn` |  |  |  |
| `spec.backfillAll.spannerExcludedObjects.schemas[].tables[].columns[].column` | `string` |  |  |  |
| `spec.backfillNone` | `bool` |  |  |  |
| `spec.ruleSets` | `[]GcpDatastreamStreamRuleSet` |  |  |  |
| `spec.ruleSets[].objectFilter` | `GcpDatastreamStreamObjectFilter` | yes |  |  |
| `spec.ruleSets[].objectFilter.sourceObjectIdentifier` | `GcpDatastreamStreamSourceObjectIdentifier` |  |  |  |
| `spec.ruleSets[].objectFilter.sourceObjectIdentifier.mysqlIdentifier` | `GcpDatastreamStreamMysqlIdentifier` |  |  |  |
| `spec.ruleSets[].objectFilter.sourceObjectIdentifier.mysqlIdentifier.database` | `string` | yes |  |  |
| `spec.ruleSets[].objectFilter.sourceObjectIdentifier.mysqlIdentifier.table` | `string` | yes |  |  |
| `spec.ruleSets[].objectFilter.sourceObjectIdentifier.postgresqlIdentifier` | `GcpDatastreamStreamSchemaTableIdentifier` |  |  |  |
| `spec.ruleSets[].objectFilter.sourceObjectIdentifier.postgresqlIdentifier.schema` | `string` | yes |  |  |
| `spec.ruleSets[].objectFilter.sourceObjectIdentifier.postgresqlIdentifier.table` | `string` | yes |  |  |
| `spec.ruleSets[].objectFilter.sourceObjectIdentifier.oracleIdentifier` | `GcpDatastreamStreamSchemaTableIdentifier` |  |  |  |
| `spec.ruleSets[].objectFilter.sourceObjectIdentifier.oracleIdentifier.schema` | `string` | yes |  |  |
| `spec.ruleSets[].objectFilter.sourceObjectIdentifier.oracleIdentifier.table` | `string` | yes |  |  |
| `spec.ruleSets[].objectFilter.sourceObjectIdentifier.sqlServerIdentifier` | `GcpDatastreamStreamSchemaTableIdentifier` |  |  |  |
| `spec.ruleSets[].objectFilter.sourceObjectIdentifier.sqlServerIdentifier.schema` | `string` | yes |  |  |
| `spec.ruleSets[].objectFilter.sourceObjectIdentifier.sqlServerIdentifier.table` | `string` | yes |  |  |
| `spec.ruleSets[].objectFilter.sourceObjectIdentifier.mongodbIdentifier` | `GcpDatastreamStreamMongodbIdentifier` |  |  |  |
| `spec.ruleSets[].objectFilter.sourceObjectIdentifier.mongodbIdentifier.database` | `string` | yes |  |  |
| `spec.ruleSets[].objectFilter.sourceObjectIdentifier.mongodbIdentifier.collection` | `string` | yes |  |  |
| `spec.ruleSets[].objectFilter.sourceObjectIdentifier.salesforceIdentifier` | `GcpDatastreamStreamSalesforceIdentifier` |  |  |  |
| `spec.ruleSets[].objectFilter.sourceObjectIdentifier.salesforceIdentifier.objectName` | `string` | yes |  |  |
| `spec.ruleSets[].objectFilter.sourceObjectIdentifier.spannerIdentifier` | `GcpDatastreamStreamSpannerIdentifier` |  |  |  |
| `spec.ruleSets[].objectFilter.sourceObjectIdentifier.spannerIdentifier.schema` | `string` |  |  |  |
| `spec.ruleSets[].objectFilter.sourceObjectIdentifier.spannerIdentifier.table` | `string` | yes |  |  |
| `spec.ruleSets[].customizationRules` | `[]GcpDatastreamStreamCustomizationRule` | yes |  |  |
| `spec.ruleSets[].customizationRules[].bigqueryPartitioning` | `GcpDatastreamStreamBigqueryPartitioning` |  |  |  |
| `spec.ruleSets[].customizationRules[].bigqueryPartitioning.ingestionTimePartition` | `GcpDatastreamStreamIngestionTimePartition` |  |  |  |
| `spec.ruleSets[].customizationRules[].bigqueryPartitioning.ingestionTimePartition.partitioningTimeGranularity` | `string` |  |  |  |
| `spec.ruleSets[].customizationRules[].bigqueryPartitioning.timeUnitPartition` | `GcpDatastreamStreamTimeUnitPartition` |  |  |  |
| `spec.ruleSets[].customizationRules[].bigqueryPartitioning.timeUnitPartition.column` | `string` | yes |  |  |
| `spec.ruleSets[].customizationRules[].bigqueryPartitioning.timeUnitPartition.partitioningTimeGranularity` | `string` |  |  |  |
| `spec.ruleSets[].customizationRules[].bigqueryPartitioning.integerRangePartition` | `GcpDatastreamStreamIntegerRangePartition` |  |  |  |
| `spec.ruleSets[].customizationRules[].bigqueryPartitioning.integerRangePartition.column` | `string` | yes |  |  |
| `spec.ruleSets[].customizationRules[].bigqueryPartitioning.integerRangePartition.start` | `int64` |  |  |  |
| `spec.ruleSets[].customizationRules[].bigqueryPartitioning.integerRangePartition.end` | `int64` |  |  |  |
| `spec.ruleSets[].customizationRules[].bigqueryPartitioning.integerRangePartition.interval` | `int64` |  |  |  |
| `spec.ruleSets[].customizationRules[].bigqueryPartitioning.requirePartitionFilter` | `bool` |  |  |  |
| `spec.ruleSets[].customizationRules[].bigqueryClustering` | `GcpDatastreamStreamBigqueryClustering` |  |  |  |
| `spec.ruleSets[].customizationRules[].bigqueryClustering.columns` | `[]string` | yes |  |  |
| `spec.customerManagedEncryptionKey` | `string \| valueFrom` |  |  | GcpKmsKey (`status.outputs.key_id`) |
| `spec.desiredState` | `string` |  |  |  |
| `spec.createWithoutValidation` | `bool` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

The GCP project the stream lives in: a literal project ID or a
GcpProject reference. If omitted, the provider's default project is
used. Immutable.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.location

`string` · required

The Datastream region, e.g. "us-central1" -- the profiles' region.
Immutable.

- rule: {"required":true,"string":{"pattern":"^[a-z]+-[a-z]+[0-9]+$"}}

### spec.streamId

`string`

The stream's ID. Defaults to metadata.name. Immutable.

### spec.displayName

`string`

The name shown in the console. Defaults to metadata.name.

### spec.labels

`map<string, string>`

Labels on the stream. The platform attribution labels are added on top
and win on a key conflict.

### spec.sourceConfig

`GcpDatastreamStreamSourceConfig` · required

The source and how it is read.

- rule: {"required":true}
- rule: set exactly one of mysql_source_config, postgresql_source_config, oracle_source_config, sql_server_source_config, mongodb_source_config, salesforce_source_config, spanner_source_config

### spec.sourceConfig.sourceConnectionProfile

`string | valueFrom` · required

The source profile -- a GcpDatastreamConnectionProfile reference (its
name output) or a literal
projects/{project}/locations/{location}/connectionProfiles/{id} in the
stream's location. Salesforce and Spanner profiles are made outside
the catalog and named literally. Immutable.

- references: GcpDatastreamConnectionProfile (`status.outputs.name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpDatastreamConnectionProfile, name: <that resource's name>, fieldPath: status.outputs.name}} -- a bare string does not parse

### spec.sourceConfig.mysqlSourceConfig

`GcpDatastreamStreamMysqlSourceConfig`

A MySQL source.

### spec.sourceConfig.mysqlSourceConfig.includeObjects

`GcpDatastreamStreamMysqlRdbms`

What to replicate; omitted means every database the user can read.

### spec.sourceConfig.mysqlSourceConfig.includeObjects.mysqlDatabases

`[]GcpDatastreamStreamMysqlDatabase` · required

The databases in the set -- at least one.

- rule: {"repeated":{"minItems":"1"}}

### spec.sourceConfig.mysqlSourceConfig.includeObjects.mysqlDatabases[].database

`string` · required

The database's name.

- rule: {"required":true}

### spec.sourceConfig.mysqlSourceConfig.includeObjects.mysqlDatabases[].mysqlTables

`[]GcpDatastreamStreamMysqlTable`

The tables to narrow to; empty selects every table.

### spec.sourceConfig.mysqlSourceConfig.includeObjects.mysqlDatabases[].mysqlTables[].table

`string` · required

The table's name.

- rule: {"required":true}

### spec.sourceConfig.mysqlSourceConfig.includeObjects.mysqlDatabases[].mysqlTables[].mysqlColumns

`[]GcpDatastreamStreamMysqlColumn`

The columns to narrow to; empty selects every column.

### spec.sourceConfig.mysqlSourceConfig.includeObjects.mysqlDatabases[].mysqlTables[].mysqlColumns[].column

`string`

The column's name.

### spec.sourceConfig.mysqlSourceConfig.includeObjects.mysqlDatabases[].mysqlTables[].mysqlColumns[].collation

`string`

The column's collation, e.g. "utf8mb4_general_ci".

### spec.sourceConfig.mysqlSourceConfig.includeObjects.mysqlDatabases[].mysqlTables[].mysqlColumns[].dataType

`string`

The column's MySQL data type, e.g. "VARCHAR".

### spec.sourceConfig.mysqlSourceConfig.includeObjects.mysqlDatabases[].mysqlTables[].mysqlColumns[].nullable

`bool`

Whether the column is nullable.

### spec.sourceConfig.mysqlSourceConfig.includeObjects.mysqlDatabases[].mysqlTables[].mysqlColumns[].ordinalPosition

`int32`

The column's position in the table.

### spec.sourceConfig.mysqlSourceConfig.includeObjects.mysqlDatabases[].mysqlTables[].mysqlColumns[].primaryKey

`bool`

Whether the column is part of the primary key.

### spec.sourceConfig.mysqlSourceConfig.excludeObjects

`GcpDatastreamStreamMysqlRdbms`

What to skip.

### spec.sourceConfig.mysqlSourceConfig.excludeObjects.mysqlDatabases

`[]GcpDatastreamStreamMysqlDatabase` · required

The databases in the set -- at least one.

- rule: {"repeated":{"minItems":"1"}}

### spec.sourceConfig.mysqlSourceConfig.excludeObjects.mysqlDatabases[].database

`string` · required

The database's name.

- rule: {"required":true}

### spec.sourceConfig.mysqlSourceConfig.excludeObjects.mysqlDatabases[].mysqlTables

`[]GcpDatastreamStreamMysqlTable`

The tables to narrow to; empty selects every table.

### spec.sourceConfig.mysqlSourceConfig.excludeObjects.mysqlDatabases[].mysqlTables[].table

`string` · required

The table's name.

- rule: {"required":true}

### spec.sourceConfig.mysqlSourceConfig.excludeObjects.mysqlDatabases[].mysqlTables[].mysqlColumns

`[]GcpDatastreamStreamMysqlColumn`

The columns to narrow to; empty selects every column.

### spec.sourceConfig.mysqlSourceConfig.excludeObjects.mysqlDatabases[].mysqlTables[].mysqlColumns[].column

`string`

The column's name.

### spec.sourceConfig.mysqlSourceConfig.excludeObjects.mysqlDatabases[].mysqlTables[].mysqlColumns[].collation

`string`

The column's collation, e.g. "utf8mb4_general_ci".

### spec.sourceConfig.mysqlSourceConfig.excludeObjects.mysqlDatabases[].mysqlTables[].mysqlColumns[].dataType

`string`

The column's MySQL data type, e.g. "VARCHAR".

### spec.sourceConfig.mysqlSourceConfig.excludeObjects.mysqlDatabases[].mysqlTables[].mysqlColumns[].nullable

`bool`

Whether the column is nullable.

### spec.sourceConfig.mysqlSourceConfig.excludeObjects.mysqlDatabases[].mysqlTables[].mysqlColumns[].ordinalPosition

`int32`

The column's position in the table.

### spec.sourceConfig.mysqlSourceConfig.excludeObjects.mysqlDatabases[].mysqlTables[].mysqlColumns[].primaryKey

`bool`

Whether the column is part of the primary key.

### spec.sourceConfig.mysqlSourceConfig.maxConcurrentBackfillTasks

`int32`

How many tables backfill at once. Empty (or 0) uses Google's default.

- rule: {"int32":{"gte":0}}

### spec.sourceConfig.mysqlSourceConfig.maxConcurrentCdcTasks

`int32`

How many change-capture tasks run at once. Empty (or 0) uses Google's
default.

- rule: {"int32":{"gte":0}}

### spec.sourceConfig.mysqlSourceConfig.cdcMethod

`string`

How changes are read (Google's cdc_method union):
  "GTID"                -- global transaction IDs; survives a failover
                           to a replica (needs gtid_mode=ON)
  "BINARY_LOG_POSITION" -- binary log file and position
Empty leaves Google's choice (binary log position).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["GTID","BINARY_LOG_POSITION"]}}

### spec.sourceConfig.postgresqlSourceConfig

`GcpDatastreamStreamPostgresqlSourceConfig`

A PostgreSQL source.

### spec.sourceConfig.postgresqlSourceConfig.includeObjects

`GcpDatastreamStreamPostgresqlRdbms`

What to replicate; omitted means every table in the publication.

### spec.sourceConfig.postgresqlSourceConfig.includeObjects.postgresqlSchemas

`[]GcpDatastreamStreamPostgresqlSchema` · required

The schemas in the set -- at least one.

- rule: {"repeated":{"minItems":"1"}}

### spec.sourceConfig.postgresqlSourceConfig.includeObjects.postgresqlSchemas[].schema

`string` · required

The schema's name, e.g. "public".

- rule: {"required":true}

### spec.sourceConfig.postgresqlSourceConfig.includeObjects.postgresqlSchemas[].postgresqlTables

`[]GcpDatastreamStreamPostgresqlTable`

The tables to narrow to; empty selects every table.

### spec.sourceConfig.postgresqlSourceConfig.includeObjects.postgresqlSchemas[].postgresqlTables[].table

`string` · required

The table's name.

- rule: {"required":true}

### spec.sourceConfig.postgresqlSourceConfig.includeObjects.postgresqlSchemas[].postgresqlTables[].postgresqlColumns

`[]GcpDatastreamStreamPostgresqlColumn`

The columns to narrow to; empty selects every column.

### spec.sourceConfig.postgresqlSourceConfig.includeObjects.postgresqlSchemas[].postgresqlTables[].postgresqlColumns[].column

`string`

The column's name.

### spec.sourceConfig.postgresqlSourceConfig.includeObjects.postgresqlSchemas[].postgresqlTables[].postgresqlColumns[].dataType

`string`

The column's PostgreSQL data type, e.g. "TEXT".

### spec.sourceConfig.postgresqlSourceConfig.includeObjects.postgresqlSchemas[].postgresqlTables[].postgresqlColumns[].nullable

`bool`

Whether the column is nullable.

### spec.sourceConfig.postgresqlSourceConfig.includeObjects.postgresqlSchemas[].postgresqlTables[].postgresqlColumns[].ordinalPosition

`int32`

The column's position in the table.

### spec.sourceConfig.postgresqlSourceConfig.includeObjects.postgresqlSchemas[].postgresqlTables[].postgresqlColumns[].primaryKey

`bool`

Whether the column is part of the primary key.

### spec.sourceConfig.postgresqlSourceConfig.excludeObjects

`GcpDatastreamStreamPostgresqlRdbms`

What to skip.

### spec.sourceConfig.postgresqlSourceConfig.excludeObjects.postgresqlSchemas

`[]GcpDatastreamStreamPostgresqlSchema` · required

The schemas in the set -- at least one.

- rule: {"repeated":{"minItems":"1"}}

### spec.sourceConfig.postgresqlSourceConfig.excludeObjects.postgresqlSchemas[].schema

`string` · required

The schema's name, e.g. "public".

- rule: {"required":true}

### spec.sourceConfig.postgresqlSourceConfig.excludeObjects.postgresqlSchemas[].postgresqlTables

`[]GcpDatastreamStreamPostgresqlTable`

The tables to narrow to; empty selects every table.

### spec.sourceConfig.postgresqlSourceConfig.excludeObjects.postgresqlSchemas[].postgresqlTables[].table

`string` · required

The table's name.

- rule: {"required":true}

### spec.sourceConfig.postgresqlSourceConfig.excludeObjects.postgresqlSchemas[].postgresqlTables[].postgresqlColumns

`[]GcpDatastreamStreamPostgresqlColumn`

The columns to narrow to; empty selects every column.

### spec.sourceConfig.postgresqlSourceConfig.excludeObjects.postgresqlSchemas[].postgresqlTables[].postgresqlColumns[].column

`string`

The column's name.

### spec.sourceConfig.postgresqlSourceConfig.excludeObjects.postgresqlSchemas[].postgresqlTables[].postgresqlColumns[].dataType

`string`

The column's PostgreSQL data type, e.g. "TEXT".

### spec.sourceConfig.postgresqlSourceConfig.excludeObjects.postgresqlSchemas[].postgresqlTables[].postgresqlColumns[].nullable

`bool`

Whether the column is nullable.

### spec.sourceConfig.postgresqlSourceConfig.excludeObjects.postgresqlSchemas[].postgresqlTables[].postgresqlColumns[].ordinalPosition

`int32`

The column's position in the table.

### spec.sourceConfig.postgresqlSourceConfig.excludeObjects.postgresqlSchemas[].postgresqlTables[].postgresqlColumns[].primaryKey

`bool`

Whether the column is part of the primary key.

### spec.sourceConfig.postgresqlSourceConfig.replicationSlot

`string` · required

The logical replication slot Datastream consumes, created beforehand
with the pgoutput plugin. One slot per stream; an idle slot makes the
server retain WAL, so delete a stream's slot with the stream.

- rule: {"required":true}

### spec.sourceConfig.postgresqlSourceConfig.publication

`string` · required

The publication naming the tables the stream may read, created
beforehand (CREATE PUBLICATION ... FOR ALL TABLES or FOR TABLE ...).

- rule: {"required":true}

### spec.sourceConfig.postgresqlSourceConfig.maxConcurrentBackfillTasks

`int32`

How many tables backfill at once. Empty (or 0) uses Google's default.

- rule: {"int32":{"gte":0}}

### spec.sourceConfig.oracleSourceConfig

`GcpDatastreamStreamOracleSourceConfig`

An Oracle source.

### spec.sourceConfig.oracleSourceConfig.includeObjects

`GcpDatastreamStreamOracleRdbms`

What to replicate; omitted means every schema the user can read.

### spec.sourceConfig.oracleSourceConfig.includeObjects.oracleSchemas

`[]GcpDatastreamStreamOracleSchema` · required

The schemas in the set -- at least one.

- rule: {"repeated":{"minItems":"1"}}

### spec.sourceConfig.oracleSourceConfig.includeObjects.oracleSchemas[].schema

`string` · required

The schema's name.

- rule: {"required":true}

### spec.sourceConfig.oracleSourceConfig.includeObjects.oracleSchemas[].oracleTables

`[]GcpDatastreamStreamOracleTable`

The tables to narrow to; empty selects every table.

### spec.sourceConfig.oracleSourceConfig.includeObjects.oracleSchemas[].oracleTables[].table

`string` · required

The table's name.

- rule: {"required":true}

### spec.sourceConfig.oracleSourceConfig.includeObjects.oracleSchemas[].oracleTables[].oracleColumns

`[]GcpDatastreamStreamOracleColumn`

The columns to narrow to; empty selects every column.

### spec.sourceConfig.oracleSourceConfig.includeObjects.oracleSchemas[].oracleTables[].oracleColumns[].column

`string`

The column's name.

### spec.sourceConfig.oracleSourceConfig.includeObjects.oracleSchemas[].oracleTables[].oracleColumns[].dataType

`string`

The column's Oracle data type, e.g. "VARCHAR2".

### spec.sourceConfig.oracleSourceConfig.excludeObjects

`GcpDatastreamStreamOracleRdbms`

What to skip.

### spec.sourceConfig.oracleSourceConfig.excludeObjects.oracleSchemas

`[]GcpDatastreamStreamOracleSchema` · required

The schemas in the set -- at least one.

- rule: {"repeated":{"minItems":"1"}}

### spec.sourceConfig.oracleSourceConfig.excludeObjects.oracleSchemas[].schema

`string` · required

The schema's name.

- rule: {"required":true}

### spec.sourceConfig.oracleSourceConfig.excludeObjects.oracleSchemas[].oracleTables

`[]GcpDatastreamStreamOracleTable`

The tables to narrow to; empty selects every table.

### spec.sourceConfig.oracleSourceConfig.excludeObjects.oracleSchemas[].oracleTables[].table

`string` · required

The table's name.

- rule: {"required":true}

### spec.sourceConfig.oracleSourceConfig.excludeObjects.oracleSchemas[].oracleTables[].oracleColumns

`[]GcpDatastreamStreamOracleColumn`

The columns to narrow to; empty selects every column.

### spec.sourceConfig.oracleSourceConfig.excludeObjects.oracleSchemas[].oracleTables[].oracleColumns[].column

`string`

The column's name.

### spec.sourceConfig.oracleSourceConfig.excludeObjects.oracleSchemas[].oracleTables[].oracleColumns[].dataType

`string`

The column's Oracle data type, e.g. "VARCHAR2".

### spec.sourceConfig.oracleSourceConfig.maxConcurrentBackfillTasks

`int32`

How many tables backfill at once. Empty (or 0) uses Google's default.

- rule: {"int32":{"gte":0}}

### spec.sourceConfig.oracleSourceConfig.maxConcurrentCdcTasks

`int32`

How many change-capture tasks run at once. Empty (or 0) uses Google's
default.

- rule: {"int32":{"gte":0}}

### spec.sourceConfig.oracleSourceConfig.largeObjectsHandling

`string`

What happens to LOB values -- CLOB, BLOB, NCLOB (Google's
large_objects_handling union):
  "DROP"   -- the values are dropped; rows replicate without them
  "STREAM" -- the values are streamed with the row
Empty leaves Google's default.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["DROP","STREAM"]}}

### spec.sourceConfig.sqlServerSourceConfig

`GcpDatastreamStreamSqlServerSourceConfig`

A SQL Server source.

### spec.sourceConfig.sqlServerSourceConfig.includeObjects

`GcpDatastreamStreamSqlServerRdbms`

What to replicate; omitted means every schema the user can read.

### spec.sourceConfig.sqlServerSourceConfig.includeObjects.schemas

`[]GcpDatastreamStreamSqlServerSchema` · required

The schemas in the set -- at least one.

- rule: {"repeated":{"minItems":"1"}}

### spec.sourceConfig.sqlServerSourceConfig.includeObjects.schemas[].schema

`string` · required

The schema's name, e.g. "dbo".

- rule: {"required":true}

### spec.sourceConfig.sqlServerSourceConfig.includeObjects.schemas[].tables

`[]GcpDatastreamStreamSqlServerTable`

The tables to narrow to; empty selects every table.

### spec.sourceConfig.sqlServerSourceConfig.includeObjects.schemas[].tables[].table

`string` · required

The table's name.

- rule: {"required":true}

### spec.sourceConfig.sqlServerSourceConfig.includeObjects.schemas[].tables[].columns

`[]GcpDatastreamStreamSqlServerColumn`

The columns to narrow to; empty selects every column.

### spec.sourceConfig.sqlServerSourceConfig.includeObjects.schemas[].tables[].columns[].column

`string`

The column's name.

### spec.sourceConfig.sqlServerSourceConfig.includeObjects.schemas[].tables[].columns[].dataType

`string`

The column's SQL Server data type, e.g. "NVARCHAR".

### spec.sourceConfig.sqlServerSourceConfig.excludeObjects

`GcpDatastreamStreamSqlServerRdbms`

What to skip.

### spec.sourceConfig.sqlServerSourceConfig.excludeObjects.schemas

`[]GcpDatastreamStreamSqlServerSchema` · required

The schemas in the set -- at least one.

- rule: {"repeated":{"minItems":"1"}}

### spec.sourceConfig.sqlServerSourceConfig.excludeObjects.schemas[].schema

`string` · required

The schema's name, e.g. "dbo".

- rule: {"required":true}

### spec.sourceConfig.sqlServerSourceConfig.excludeObjects.schemas[].tables

`[]GcpDatastreamStreamSqlServerTable`

The tables to narrow to; empty selects every table.

### spec.sourceConfig.sqlServerSourceConfig.excludeObjects.schemas[].tables[].table

`string` · required

The table's name.

- rule: {"required":true}

### spec.sourceConfig.sqlServerSourceConfig.excludeObjects.schemas[].tables[].columns

`[]GcpDatastreamStreamSqlServerColumn`

The columns to narrow to; empty selects every column.

### spec.sourceConfig.sqlServerSourceConfig.excludeObjects.schemas[].tables[].columns[].column

`string`

The column's name.

### spec.sourceConfig.sqlServerSourceConfig.excludeObjects.schemas[].tables[].columns[].dataType

`string`

The column's SQL Server data type, e.g. "NVARCHAR".

### spec.sourceConfig.sqlServerSourceConfig.maxConcurrentBackfillTasks

`int32`

How many tables backfill at once. Empty (or 0) uses Google's default.

- rule: {"int32":{"gte":0}}

### spec.sourceConfig.sqlServerSourceConfig.maxConcurrentCdcTasks

`int32`

How many change-capture tasks run at once. Empty (or 0) uses Google's
default.

- rule: {"int32":{"gte":0}}

### spec.sourceConfig.sqlServerSourceConfig.cdcMethod

`string`

How changes are read (Google's cdc_method union):
  "CHANGE_TABLES"    -- SQL Server's CDC change tables (CDC enabled on
                        the database and each table)
  "TRANSACTION_LOGS" -- the transaction log directly; lighter on the
                        source, needs log access and retention
Empty leaves Google's default.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["CHANGE_TABLES","TRANSACTION_LOGS"]}}

### spec.sourceConfig.mongodbSourceConfig

`GcpDatastreamStreamMongodbSourceConfig`

A MongoDB source.

### spec.sourceConfig.mongodbSourceConfig.includeObjects

`GcpDatastreamStreamMongodbCluster`

What to replicate; omitted means every database the user can read.

### spec.sourceConfig.mongodbSourceConfig.includeObjects.databases

`[]GcpDatastreamStreamMongodbDatabase`

The databases in the set (at least one inside a backfill exclusion).

### spec.sourceConfig.mongodbSourceConfig.includeObjects.databases[].database

`string`

The database's name (required inside a backfill exclusion).

### spec.sourceConfig.mongodbSourceConfig.includeObjects.databases[].collections

`[]GcpDatastreamStreamMongodbCollection`

The collections to narrow to; empty selects every collection.

### spec.sourceConfig.mongodbSourceConfig.includeObjects.databases[].collections[].collection

`string`

The collection's name (required inside a backfill exclusion).

### spec.sourceConfig.mongodbSourceConfig.includeObjects.databases[].collections[].fields

`[]GcpDatastreamStreamMongodbField`

The fields to narrow to; empty selects every field.

### spec.sourceConfig.mongodbSourceConfig.includeObjects.databases[].collections[].fields[].field

`string`

The field's name.

### spec.sourceConfig.mongodbSourceConfig.excludeObjects

`GcpDatastreamStreamMongodbCluster`

What to skip.

### spec.sourceConfig.mongodbSourceConfig.excludeObjects.databases

`[]GcpDatastreamStreamMongodbDatabase`

The databases in the set (at least one inside a backfill exclusion).

### spec.sourceConfig.mongodbSourceConfig.excludeObjects.databases[].database

`string`

The database's name (required inside a backfill exclusion).

### spec.sourceConfig.mongodbSourceConfig.excludeObjects.databases[].collections

`[]GcpDatastreamStreamMongodbCollection`

The collections to narrow to; empty selects every collection.

### spec.sourceConfig.mongodbSourceConfig.excludeObjects.databases[].collections[].collection

`string`

The collection's name (required inside a backfill exclusion).

### spec.sourceConfig.mongodbSourceConfig.excludeObjects.databases[].collections[].fields

`[]GcpDatastreamStreamMongodbField`

The fields to narrow to; empty selects every field.

### spec.sourceConfig.mongodbSourceConfig.excludeObjects.databases[].collections[].fields[].field

`string`

The field's name.

### spec.sourceConfig.mongodbSourceConfig.maxConcurrentBackfillTasks

`int32`

How many collections backfill at once, 0-50. Empty (or 0) uses
Google's default.

- rule: {"int32":{"lte":50,"gte":0}}

### spec.sourceConfig.salesforceSourceConfig

`GcpDatastreamStreamSalesforceSourceConfig`

A Salesforce source.

### spec.sourceConfig.salesforceSourceConfig.includeObjects

`GcpDatastreamStreamSalesforceOrg`

What to replicate; omitted means every object the integration user can
read.

### spec.sourceConfig.salesforceSourceConfig.includeObjects.objects

`[]GcpDatastreamStreamSalesforceObject` · required

The objects in the set -- at least one.

- rule: {"repeated":{"minItems":"1"}}

### spec.sourceConfig.salesforceSourceConfig.includeObjects.objects[].objectName

`string`

The object's API name, e.g. "Account".

### spec.sourceConfig.salesforceSourceConfig.includeObjects.objects[].fields

`[]GcpDatastreamStreamSalesforceField`

The fields to narrow to; empty selects every field.

### spec.sourceConfig.salesforceSourceConfig.includeObjects.objects[].fields[].name

`string`

The field's API name, e.g. "AnnualRevenue".

### spec.sourceConfig.salesforceSourceConfig.excludeObjects

`GcpDatastreamStreamSalesforceOrg`

What to skip.

### spec.sourceConfig.salesforceSourceConfig.excludeObjects.objects

`[]GcpDatastreamStreamSalesforceObject` · required

The objects in the set -- at least one.

- rule: {"repeated":{"minItems":"1"}}

### spec.sourceConfig.salesforceSourceConfig.excludeObjects.objects[].objectName

`string`

The object's API name, e.g. "Account".

### spec.sourceConfig.salesforceSourceConfig.excludeObjects.objects[].fields

`[]GcpDatastreamStreamSalesforceField`

The fields to narrow to; empty selects every field.

### spec.sourceConfig.salesforceSourceConfig.excludeObjects.objects[].fields[].name

`string`

The field's API name, e.g. "AnnualRevenue".

### spec.sourceConfig.salesforceSourceConfig.pollingInterval

`string` · required

How often each object is polled for changes, as a duration in seconds
-- Google allows 5 minutes to 24 hours ("300s" to "86400s"). Shorter
means fresher data and more Salesforce API calls against the org's
daily limit.

- rule: {"required":true,"string":{"pattern":"^[0-9]+(\\.[0-9]+)?s$"}}

### spec.sourceConfig.spannerSourceConfig

`GcpDatastreamStreamSpannerSourceConfig`

A Spanner source.

### spec.sourceConfig.spannerSourceConfig.includeObjects

`GcpDatastreamStreamSpannerDatabase`

What to replicate; omitted means every table the change stream
watches.

### spec.sourceConfig.spannerSourceConfig.includeObjects.schemas

`[]GcpDatastreamStreamSpannerSchema` · required

The schemas in the set -- at least one.

- rule: {"repeated":{"minItems":"1"}}

### spec.sourceConfig.spannerSourceConfig.includeObjects.schemas[].schema

`string` · required

The schema's name.

- rule: {"required":true}

### spec.sourceConfig.spannerSourceConfig.includeObjects.schemas[].tables

`[]GcpDatastreamStreamSpannerTable`

The tables to narrow to; empty selects every table.

### spec.sourceConfig.spannerSourceConfig.includeObjects.schemas[].tables[].table

`string` · required

The table's name.

- rule: {"required":true}

### spec.sourceConfig.spannerSourceConfig.includeObjects.schemas[].tables[].columns

`[]GcpDatastreamStreamSpannerColumn`

The columns to narrow to; empty selects every column.

### spec.sourceConfig.spannerSourceConfig.includeObjects.schemas[].tables[].columns[].column

`string`

The column's name (required inside a backfill exclusion).

### spec.sourceConfig.spannerSourceConfig.excludeObjects

`GcpDatastreamStreamSpannerDatabase`

What to skip.

### spec.sourceConfig.spannerSourceConfig.excludeObjects.schemas

`[]GcpDatastreamStreamSpannerSchema` · required

The schemas in the set -- at least one.

- rule: {"repeated":{"minItems":"1"}}

### spec.sourceConfig.spannerSourceConfig.excludeObjects.schemas[].schema

`string` · required

The schema's name.

- rule: {"required":true}

### spec.sourceConfig.spannerSourceConfig.excludeObjects.schemas[].tables

`[]GcpDatastreamStreamSpannerTable`

The tables to narrow to; empty selects every table.

### spec.sourceConfig.spannerSourceConfig.excludeObjects.schemas[].tables[].table

`string` · required

The table's name.

- rule: {"required":true}

### spec.sourceConfig.spannerSourceConfig.excludeObjects.schemas[].tables[].columns

`[]GcpDatastreamStreamSpannerColumn`

The columns to narrow to; empty selects every column.

### spec.sourceConfig.spannerSourceConfig.excludeObjects.schemas[].tables[].columns[].column

`string`

The column's name (required inside a backfill exclusion).

### spec.sourceConfig.spannerSourceConfig.changeStreamName

`string` · required

The Spanner change stream Datastream reads, created beforehand (CREATE
CHANGE STREAM ... FOR ALL). Google requires it. Immutable.

- rule: {"required":true}

### spec.sourceConfig.spannerSourceConfig.backfillDataBoostEnabled

`bool`

Run the backfill's reads on Spanner Data Boost -- independent compute
that leaves the instance's serving capacity untouched (billed by
Spanner). Off by default.

### spec.sourceConfig.spannerSourceConfig.fgacRole

`string`

A fine-grained access control role the reads run as. Empty reads with
the service agent's database-level permissions.

### spec.sourceConfig.spannerSourceConfig.maxConcurrentBackfillTasks

`int32`

How many tables backfill at once. Empty (or 0) uses Google's default.

- rule: {"int32":{"gte":0}}

### spec.sourceConfig.spannerSourceConfig.maxConcurrentCdcTasks

`int32`

How many change-capture tasks run at once. Empty (or 0) uses Google's
default.

- rule: {"int32":{"gte":0}}

### spec.sourceConfig.spannerSourceConfig.spannerRpcPriority

`string`

The priority of Datastream's Spanner reads: "LOW", "MEDIUM", or
"HIGH". Lower yields to the application's traffic. Empty leaves
Google's default.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["LOW","MEDIUM","HIGH"]}}

### spec.destinationConfig

`GcpDatastreamStreamDestinationConfig` · required

The destination and how data lands.

- rule: {"required":true}
- rule: set exactly one of bigquery_destination_config or gcs_destination_config

### spec.destinationConfig.destinationConnectionProfile

`string | valueFrom` · required

The destination profile -- a GcpDatastreamConnectionProfile reference
(its name output) or a literal full profile name in the stream's
location. Immutable.

- references: GcpDatastreamConnectionProfile (`status.outputs.name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpDatastreamConnectionProfile, name: <that resource's name>, fieldPath: status.outputs.name}} -- a bare string does not parse

### spec.destinationConfig.bigqueryDestinationConfig

`GcpDatastreamStreamBigqueryDestinationConfig`

Write into BigQuery (needs a bigquery_profile).

- rule: set exactly one of single_target_dataset or source_hierarchy_datasets

### spec.destinationConfig.bigqueryDestinationConfig.singleTargetDataset

`GcpDatastreamStreamSingleTargetDataset`

Every table into one dataset.

### spec.destinationConfig.bigqueryDestinationConfig.singleTargetDataset.datasetId

`string | valueFrom` · required

The dataset -- a GcpBigQueryDataset reference (its self_link output,
which the modules trim to Google's projects/{project}/datasets/{dataset}
form) or a literal in that form or {project}:{dataset}. Tables are
named {schema}_{table}.

- references: GcpBigQueryDataset (`status.outputs.self_link`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpBigQueryDataset, name: <that resource's name>, fieldPath: status.outputs.self_link}} -- a bare string does not parse

### spec.destinationConfig.bigqueryDestinationConfig.sourceHierarchyDatasets

`GcpDatastreamStreamSourceHierarchyDatasets`

One dataset per source schema or database, created by Datastream.

### spec.destinationConfig.bigqueryDestinationConfig.sourceHierarchyDatasets.datasetTemplate

`GcpDatastreamStreamDatasetTemplate` · required

The template for the created datasets.

- rule: {"required":true}

### spec.destinationConfig.bigqueryDestinationConfig.sourceHierarchyDatasets.datasetTemplate.location

`string` · required

Where the datasets live, e.g. "US" or "us-central1".

- rule: {"required":true}

### spec.destinationConfig.bigqueryDestinationConfig.sourceHierarchyDatasets.datasetTemplate.datasetIdPrefix

`string`

A prefix for every created dataset's name, joined with an underscore:
prefix "crm" and schema "public" make "crm_public".

### spec.destinationConfig.bigqueryDestinationConfig.sourceHierarchyDatasets.datasetTemplate.kmsKeyName

`string | valueFrom`

A Cloud KMS key the created datasets use by default (CMEK) -- a
GcpKmsKey reference or a literal key path, in the datasets' location.
Datastream's service agent needs cryptoKeyEncrypterDecrypter on it.
Immutable.

- references: GcpKmsKey (`status.outputs.key_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_id}} -- a bare string does not parse

### spec.destinationConfig.bigqueryDestinationConfig.sourceHierarchyDatasets.projectId

`string | valueFrom`

The project the datasets are created in -- a literal project ID or a
GcpProject reference. Empty uses the stream's project.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.destinationConfig.bigqueryDestinationConfig.dataFreshness

`string`

How stale a query may read (merge mode), as a duration, e.g. "900s".
Lower is fresher and costs more BigQuery compute. Changing it affects
only tables created afterwards. Empty leaves Google's default.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^[0-9]+(\\.[0-9]+)?s$"}}

### spec.destinationConfig.bigqueryDestinationConfig.writeMode

`string`

How changes land (Google's write_mode union):
  "MERGE"       -- tables mirror the source's current state; changes
                   are merged by primary key (tables need one)
  "APPEND_ONLY" -- every change is appended as a row with its change
                   type; the full history, no merging
Empty leaves Google's default (merge). Immutable.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["MERGE","APPEND_ONLY"]}}

### spec.destinationConfig.bigqueryDestinationConfig.blmtConfig

`GcpDatastreamStreamBlmtConfig`

Write BigLake managed (Iceberg) tables instead of native tables.

### spec.destinationConfig.bigqueryDestinationConfig.blmtConfig.bucket

`string | valueFrom` · required

The bucket holding the table data -- a GcpGcsBucket reference (its
bucket_name output) or a literal.

- references: GcpGcsBucket (`status.outputs.bucket_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpGcsBucket, name: <that resource's name>, fieldPath: status.outputs.bucket_name}} -- a bare string does not parse

### spec.destinationConfig.bigqueryDestinationConfig.blmtConfig.rootPath

`string`

The folder inside the bucket.

### spec.destinationConfig.bigqueryDestinationConfig.blmtConfig.connectionName

`string | valueFrom` · required

The BigQuery connection whose service account writes the bucket -- a
GcpBigQueryConnection reference (its name output, which the modules
convert to Google's {project}.{location}.{connection_id} form) or a
literal in the dotted form. Use a cloud_resource connection and grant
its service account storage access on the bucket.

- references: GcpBigQueryConnection (`status.outputs.name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpBigQueryConnection, name: <that resource's name>, fieldPath: status.outputs.name}} -- a bare string does not parse

### spec.destinationConfig.bigqueryDestinationConfig.blmtConfig.fileFormat

`string` · required

The data file format. Google offers "PARQUET".

- rule: {"required":true,"string":{"in":["PARQUET"]}}

### spec.destinationConfig.bigqueryDestinationConfig.blmtConfig.tableFormat

`string` · required

The table format. Google offers "ICEBERG".

- rule: {"required":true,"string":{"in":["ICEBERG"]}}

### spec.destinationConfig.gcsDestinationConfig

`GcpDatastreamStreamGcsDestinationConfig`

Write files into Cloud Storage (needs a gcs_profile).

- rule: set exactly one of avro_file_format or json_file_format

### spec.destinationConfig.gcsDestinationConfig.path

`string`

The folder under the profile's root path, e.g. "/orders".

### spec.destinationConfig.gcsDestinationConfig.fileRotationInterval

`string`

The longest a file stays open before a new one starts -- Google allows
"15s" to "60s". Empty leaves Google's default.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^(1[5-9]|[2-5][0-9]|60)s$"}}

### spec.destinationConfig.gcsDestinationConfig.fileRotationMb

`int32`

The largest a file grows, in MB, before a new one starts. Empty (or 0)
leaves Google's default.

- rule: {"int32":{"gte":0}}

### spec.destinationConfig.gcsDestinationConfig.avroFileFormat

`bool`

Write Avro files. Google's Avro block has no settings: true declares
it.

### spec.destinationConfig.gcsDestinationConfig.jsonFileFormat

`GcpDatastreamStreamGcsJsonFileFormat`

Write JSON files.

### spec.destinationConfig.gcsDestinationConfig.jsonFileFormat.compression

`string`

"NO_COMPRESSION" or "GZIP". Empty leaves Google's default.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["NO_COMPRESSION","GZIP"]}}

### spec.destinationConfig.gcsDestinationConfig.jsonFileFormat.schemaFileFormat

`string`

Whether an Avro schema file is written beside the data files:
"NO_SCHEMA_FILE" or "AVRO_SCHEMA_FILE". Empty leaves Google's default.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["NO_SCHEMA_FILE","AVRO_SCHEMA_FILE"]}}

### spec.backfillAll

`GcpDatastreamStreamBackfillAll`

Backfill every included object, except what is listed.

### spec.backfillAll.mysqlExcludedObjects

`GcpDatastreamStreamMysqlRdbms`

MySQL objects to skip.

### spec.backfillAll.mysqlExcludedObjects.mysqlDatabases

`[]GcpDatastreamStreamMysqlDatabase` · required

The databases in the set -- at least one.

- rule: {"repeated":{"minItems":"1"}}

### spec.backfillAll.mysqlExcludedObjects.mysqlDatabases[].database

`string` · required

The database's name.

- rule: {"required":true}

### spec.backfillAll.mysqlExcludedObjects.mysqlDatabases[].mysqlTables

`[]GcpDatastreamStreamMysqlTable`

The tables to narrow to; empty selects every table.

### spec.backfillAll.mysqlExcludedObjects.mysqlDatabases[].mysqlTables[].table

`string` · required

The table's name.

- rule: {"required":true}

### spec.backfillAll.mysqlExcludedObjects.mysqlDatabases[].mysqlTables[].mysqlColumns

`[]GcpDatastreamStreamMysqlColumn`

The columns to narrow to; empty selects every column.

### spec.backfillAll.mysqlExcludedObjects.mysqlDatabases[].mysqlTables[].mysqlColumns[].column

`string`

The column's name.

### spec.backfillAll.mysqlExcludedObjects.mysqlDatabases[].mysqlTables[].mysqlColumns[].collation

`string`

The column's collation, e.g. "utf8mb4_general_ci".

### spec.backfillAll.mysqlExcludedObjects.mysqlDatabases[].mysqlTables[].mysqlColumns[].dataType

`string`

The column's MySQL data type, e.g. "VARCHAR".

### spec.backfillAll.mysqlExcludedObjects.mysqlDatabases[].mysqlTables[].mysqlColumns[].nullable

`bool`

Whether the column is nullable.

### spec.backfillAll.mysqlExcludedObjects.mysqlDatabases[].mysqlTables[].mysqlColumns[].ordinalPosition

`int32`

The column's position in the table.

### spec.backfillAll.mysqlExcludedObjects.mysqlDatabases[].mysqlTables[].mysqlColumns[].primaryKey

`bool`

Whether the column is part of the primary key.

### spec.backfillAll.postgresqlExcludedObjects

`GcpDatastreamStreamPostgresqlRdbms`

PostgreSQL objects to skip.

### spec.backfillAll.postgresqlExcludedObjects.postgresqlSchemas

`[]GcpDatastreamStreamPostgresqlSchema` · required

The schemas in the set -- at least one.

- rule: {"repeated":{"minItems":"1"}}

### spec.backfillAll.postgresqlExcludedObjects.postgresqlSchemas[].schema

`string` · required

The schema's name, e.g. "public".

- rule: {"required":true}

### spec.backfillAll.postgresqlExcludedObjects.postgresqlSchemas[].postgresqlTables

`[]GcpDatastreamStreamPostgresqlTable`

The tables to narrow to; empty selects every table.

### spec.backfillAll.postgresqlExcludedObjects.postgresqlSchemas[].postgresqlTables[].table

`string` · required

The table's name.

- rule: {"required":true}

### spec.backfillAll.postgresqlExcludedObjects.postgresqlSchemas[].postgresqlTables[].postgresqlColumns

`[]GcpDatastreamStreamPostgresqlColumn`

The columns to narrow to; empty selects every column.

### spec.backfillAll.postgresqlExcludedObjects.postgresqlSchemas[].postgresqlTables[].postgresqlColumns[].column

`string`

The column's name.

### spec.backfillAll.postgresqlExcludedObjects.postgresqlSchemas[].postgresqlTables[].postgresqlColumns[].dataType

`string`

The column's PostgreSQL data type, e.g. "TEXT".

### spec.backfillAll.postgresqlExcludedObjects.postgresqlSchemas[].postgresqlTables[].postgresqlColumns[].nullable

`bool`

Whether the column is nullable.

### spec.backfillAll.postgresqlExcludedObjects.postgresqlSchemas[].postgresqlTables[].postgresqlColumns[].ordinalPosition

`int32`

The column's position in the table.

### spec.backfillAll.postgresqlExcludedObjects.postgresqlSchemas[].postgresqlTables[].postgresqlColumns[].primaryKey

`bool`

Whether the column is part of the primary key.

### spec.backfillAll.oracleExcludedObjects

`GcpDatastreamStreamOracleRdbms`

Oracle objects to skip.

### spec.backfillAll.oracleExcludedObjects.oracleSchemas

`[]GcpDatastreamStreamOracleSchema` · required

The schemas in the set -- at least one.

- rule: {"repeated":{"minItems":"1"}}

### spec.backfillAll.oracleExcludedObjects.oracleSchemas[].schema

`string` · required

The schema's name.

- rule: {"required":true}

### spec.backfillAll.oracleExcludedObjects.oracleSchemas[].oracleTables

`[]GcpDatastreamStreamOracleTable`

The tables to narrow to; empty selects every table.

### spec.backfillAll.oracleExcludedObjects.oracleSchemas[].oracleTables[].table

`string` · required

The table's name.

- rule: {"required":true}

### spec.backfillAll.oracleExcludedObjects.oracleSchemas[].oracleTables[].oracleColumns

`[]GcpDatastreamStreamOracleColumn`

The columns to narrow to; empty selects every column.

### spec.backfillAll.oracleExcludedObjects.oracleSchemas[].oracleTables[].oracleColumns[].column

`string`

The column's name.

### spec.backfillAll.oracleExcludedObjects.oracleSchemas[].oracleTables[].oracleColumns[].dataType

`string`

The column's Oracle data type, e.g. "VARCHAR2".

### spec.backfillAll.sqlServerExcludedObjects

`GcpDatastreamStreamSqlServerRdbms`

SQL Server objects to skip.

### spec.backfillAll.sqlServerExcludedObjects.schemas

`[]GcpDatastreamStreamSqlServerSchema` · required

The schemas in the set -- at least one.

- rule: {"repeated":{"minItems":"1"}}

### spec.backfillAll.sqlServerExcludedObjects.schemas[].schema

`string` · required

The schema's name, e.g. "dbo".

- rule: {"required":true}

### spec.backfillAll.sqlServerExcludedObjects.schemas[].tables

`[]GcpDatastreamStreamSqlServerTable`

The tables to narrow to; empty selects every table.

### spec.backfillAll.sqlServerExcludedObjects.schemas[].tables[].table

`string` · required

The table's name.

- rule: {"required":true}

### spec.backfillAll.sqlServerExcludedObjects.schemas[].tables[].columns

`[]GcpDatastreamStreamSqlServerColumn`

The columns to narrow to; empty selects every column.

### spec.backfillAll.sqlServerExcludedObjects.schemas[].tables[].columns[].column

`string`

The column's name.

### spec.backfillAll.sqlServerExcludedObjects.schemas[].tables[].columns[].dataType

`string`

The column's SQL Server data type, e.g. "NVARCHAR".

### spec.backfillAll.mongodbExcludedObjects

`GcpDatastreamStreamMongodbCluster`

MongoDB objects to skip. Here every database and collection needs its
name, and at least one database is listed.

- rule: mongodb_excluded_objects needs at least one database, and every database and collection named

### spec.backfillAll.mongodbExcludedObjects.databases

`[]GcpDatastreamStreamMongodbDatabase`

The databases in the set (at least one inside a backfill exclusion).

### spec.backfillAll.mongodbExcludedObjects.databases[].database

`string`

The database's name (required inside a backfill exclusion).

### spec.backfillAll.mongodbExcludedObjects.databases[].collections

`[]GcpDatastreamStreamMongodbCollection`

The collections to narrow to; empty selects every collection.

### spec.backfillAll.mongodbExcludedObjects.databases[].collections[].collection

`string`

The collection's name (required inside a backfill exclusion).

### spec.backfillAll.mongodbExcludedObjects.databases[].collections[].fields

`[]GcpDatastreamStreamMongodbField`

The fields to narrow to; empty selects every field.

### spec.backfillAll.mongodbExcludedObjects.databases[].collections[].fields[].field

`string`

The field's name.

### spec.backfillAll.salesforceExcludedObjects

`GcpDatastreamStreamSalesforceOrg`

Salesforce objects to skip.

### spec.backfillAll.salesforceExcludedObjects.objects

`[]GcpDatastreamStreamSalesforceObject` · required

The objects in the set -- at least one.

- rule: {"repeated":{"minItems":"1"}}

### spec.backfillAll.salesforceExcludedObjects.objects[].objectName

`string`

The object's API name, e.g. "Account".

### spec.backfillAll.salesforceExcludedObjects.objects[].fields

`[]GcpDatastreamStreamSalesforceField`

The fields to narrow to; empty selects every field.

### spec.backfillAll.salesforceExcludedObjects.objects[].fields[].name

`string`

The field's API name, e.g. "AnnualRevenue".

### spec.backfillAll.spannerExcludedObjects

`GcpDatastreamStreamSpannerDatabase`

Spanner objects to skip. Here every column needs its name.

- rule: spanner_excluded_objects needs every column named

### spec.backfillAll.spannerExcludedObjects.schemas

`[]GcpDatastreamStreamSpannerSchema` · required

The schemas in the set -- at least one.

- rule: {"repeated":{"minItems":"1"}}

### spec.backfillAll.spannerExcludedObjects.schemas[].schema

`string` · required

The schema's name.

- rule: {"required":true}

### spec.backfillAll.spannerExcludedObjects.schemas[].tables

`[]GcpDatastreamStreamSpannerTable`

The tables to narrow to; empty selects every table.

### spec.backfillAll.spannerExcludedObjects.schemas[].tables[].table

`string` · required

The table's name.

- rule: {"required":true}

### spec.backfillAll.spannerExcludedObjects.schemas[].tables[].columns

`[]GcpDatastreamStreamSpannerColumn`

The columns to narrow to; empty selects every column.

### spec.backfillAll.spannerExcludedObjects.schemas[].tables[].columns[].column

`string`

The column's name (required inside a backfill exclusion).

### spec.backfillNone

`bool`

Backfill nothing: only changes from the stream's start replicate.
Google's block has no settings: true declares it.

### spec.ruleSets

`[]GcpDatastreamStreamRuleSet`

Per-object BigQuery table customizations.

### spec.ruleSets[].objectFilter

`GcpDatastreamStreamObjectFilter` · required

The source object the rules apply to.

- rule: {"required":true}

### spec.ruleSets[].objectFilter.sourceObjectIdentifier

`GcpDatastreamStreamSourceObjectIdentifier`

The object.

- rule: set exactly one source object identifier

### spec.ruleSets[].objectFilter.sourceObjectIdentifier.mysqlIdentifier

`GcpDatastreamStreamMysqlIdentifier`

A MySQL table.

### spec.ruleSets[].objectFilter.sourceObjectIdentifier.mysqlIdentifier.database

`string` · required

The database.

- rule: {"required":true}

### spec.ruleSets[].objectFilter.sourceObjectIdentifier.mysqlIdentifier.table

`string` · required

The table.

- rule: {"required":true}

### spec.ruleSets[].objectFilter.sourceObjectIdentifier.postgresqlIdentifier

`GcpDatastreamStreamSchemaTableIdentifier`

A PostgreSQL table.

### spec.ruleSets[].objectFilter.sourceObjectIdentifier.postgresqlIdentifier.schema

`string` · required

The schema.

- rule: {"required":true}

### spec.ruleSets[].objectFilter.sourceObjectIdentifier.postgresqlIdentifier.table

`string` · required

The table.

- rule: {"required":true}

### spec.ruleSets[].objectFilter.sourceObjectIdentifier.oracleIdentifier

`GcpDatastreamStreamSchemaTableIdentifier`

An Oracle table.

### spec.ruleSets[].objectFilter.sourceObjectIdentifier.oracleIdentifier.schema

`string` · required

The schema.

- rule: {"required":true}

### spec.ruleSets[].objectFilter.sourceObjectIdentifier.oracleIdentifier.table

`string` · required

The table.

- rule: {"required":true}

### spec.ruleSets[].objectFilter.sourceObjectIdentifier.sqlServerIdentifier

`GcpDatastreamStreamSchemaTableIdentifier`

A SQL Server table.

### spec.ruleSets[].objectFilter.sourceObjectIdentifier.sqlServerIdentifier.schema

`string` · required

The schema.

- rule: {"required":true}

### spec.ruleSets[].objectFilter.sourceObjectIdentifier.sqlServerIdentifier.table

`string` · required

The table.

- rule: {"required":true}

### spec.ruleSets[].objectFilter.sourceObjectIdentifier.mongodbIdentifier

`GcpDatastreamStreamMongodbIdentifier`

A MongoDB collection.

### spec.ruleSets[].objectFilter.sourceObjectIdentifier.mongodbIdentifier.database

`string` · required

The database.

- rule: {"required":true}

### spec.ruleSets[].objectFilter.sourceObjectIdentifier.mongodbIdentifier.collection

`string` · required

The collection.

- rule: {"required":true}

### spec.ruleSets[].objectFilter.sourceObjectIdentifier.salesforceIdentifier

`GcpDatastreamStreamSalesforceIdentifier`

A Salesforce object.

### spec.ruleSets[].objectFilter.sourceObjectIdentifier.salesforceIdentifier.objectName

`string` · required

The object's API name.

- rule: {"required":true}

### spec.ruleSets[].objectFilter.sourceObjectIdentifier.spannerIdentifier

`GcpDatastreamStreamSpannerIdentifier`

A Spanner table.

### spec.ruleSets[].objectFilter.sourceObjectIdentifier.spannerIdentifier.schema

`string`

The schema; empty is the database's default schema.

### spec.ruleSets[].objectFilter.sourceObjectIdentifier.spannerIdentifier.table

`string` · required

The table.

- rule: {"required":true}

### spec.ruleSets[].customizationRules

`[]GcpDatastreamStreamCustomizationRule` · required

The rules -- at least one.

- rule: {"repeated":{"minItems":"1"}}
- rule: set exactly one of bigquery_partitioning or bigquery_clustering

### spec.ruleSets[].customizationRules[].bigqueryPartitioning

`GcpDatastreamStreamBigqueryPartitioning`

Partitioning for the matched tables.

- rule: set exactly one of ingestion_time_partition, time_unit_partition, integer_range_partition

### spec.ruleSets[].customizationRules[].bigqueryPartitioning.ingestionTimePartition

`GcpDatastreamStreamIngestionTimePartition`

Partition by arrival time.

### spec.ruleSets[].customizationRules[].bigqueryPartitioning.ingestionTimePartition.partitioningTimeGranularity

`string`

"PARTITIONING_TIME_GRANULARITY_HOUR", "..._DAY", "..._MONTH", or
"..._YEAR". Empty leaves Google's default.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["PARTITIONING_TIME_GRANULARITY_HOUR","PARTITIONING_TIME_GRANULARITY_DAY","PARTITIONING_TIME_GRANULARITY_MONTH","PARTITIONING_TIME_GRANULARITY_YEAR"]}}

### spec.ruleSets[].customizationRules[].bigqueryPartitioning.timeUnitPartition

`GcpDatastreamStreamTimeUnitPartition`

Partition by a date or timestamp column.

### spec.ruleSets[].customizationRules[].bigqueryPartitioning.timeUnitPartition.column

`string` · required

The partitioning column.

- rule: {"required":true}

### spec.ruleSets[].customizationRules[].bigqueryPartitioning.timeUnitPartition.partitioningTimeGranularity

`string`

"PARTITIONING_TIME_GRANULARITY_HOUR", "..._DAY", "..._MONTH", or
"..._YEAR". Empty leaves Google's default.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["PARTITIONING_TIME_GRANULARITY_HOUR","PARTITIONING_TIME_GRANULARITY_DAY","PARTITIONING_TIME_GRANULARITY_MONTH","PARTITIONING_TIME_GRANULARITY_YEAR"]}}

### spec.ruleSets[].customizationRules[].bigqueryPartitioning.integerRangePartition

`GcpDatastreamStreamIntegerRangePartition`

Partition by integer ranges.

### spec.ruleSets[].customizationRules[].bigqueryPartitioning.integerRangePartition.column

`string` · required

The partitioning column.

- rule: {"required":true}

### spec.ruleSets[].customizationRules[].bigqueryPartitioning.integerRangePartition.start

`int64`

The first range's start (inclusive).

### spec.ruleSets[].customizationRules[].bigqueryPartitioning.integerRangePartition.end

`int64`

The last range's end (exclusive).

### spec.ruleSets[].customizationRules[].bigqueryPartitioning.integerRangePartition.interval

`int64`

Each range's width.

- rule: {"int64":{"gt":"0"}}

### spec.ruleSets[].customizationRules[].bigqueryPartitioning.requirePartitionFilter

`bool`

Refuse queries over the tables that do not filter on the partition --
protects against full-table scans.

### spec.ruleSets[].customizationRules[].bigqueryClustering

`GcpDatastreamStreamBigqueryClustering`

Clustering for the matched tables.

### spec.ruleSets[].customizationRules[].bigqueryClustering.columns

`[]string` · required

The clustering columns, most selective first (up to four in
BigQuery).

- rule: {"repeated":{"minItems":"1"}}

### spec.customerManagedEncryptionKey

`string | valueFrom`

A Cloud KMS key encrypting the data Datastream holds in flight (CMEK)
-- a GcpKmsKey reference or a literal key path in the stream's region.
Datastream's service agent needs cryptoKeyEncrypterDecrypter on it.
Empty uses a Google-managed key. Immutable.

- references: GcpKmsKey (`status.outputs.key_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_id}} -- a bare string does not parse

### spec.desiredState

`string`

Whether the stream runs:
  "" / "NOT_STARTED" -- created but not started; nothing moves or bills
  "RUNNING"          -- backfill (if chosen) and change capture run
  "PAUSED"           -- stopped, keeping its position; RUNNING resumes
Google accepts NOT_STARTED or RUNNING at create, then RUNNING or PAUSED;
a started stream never returns to NOT_STARTED.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["NOT_STARTED","RUNNING","PAUSED"]}}

### spec.createWithoutValidation

`bool`

Create the stream without Google's validation (source reachable,
objects exist, destination writable). Problems then surface when it
starts. Immutable.

### spec.deletionPolicy

`string`

What happens to the stream when this resource is destroyed:
  "" / "DELETE" -- deleted (data already written stays in the
                   destination)
  "PREVENT"     -- destroy fails
  "ABANDON"     -- it leaves management and keeps running in GCP

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Validation Rules

- `spec.exactly_one_backfill`: set exactly one of backfill_all or backfill_none

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpDatastreamStream, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.name` | `string` | Full resource name -- projects/{project}/locations/{location}/streams/{stream_id}. |
| `status.outputs.stream_id` | `string` | The stream's ID. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.sourceConfig.sourceConnectionProfile` | GcpDatastreamConnectionProfile | `status.outputs.name` |
| `spec.destinationConfig.destinationConnectionProfile` | GcpDatastreamConnectionProfile | `status.outputs.name` |
| `spec.destinationConfig.bigqueryDestinationConfig.singleTargetDataset.datasetId` | GcpBigQueryDataset | `status.outputs.self_link` |
| `spec.destinationConfig.bigqueryDestinationConfig.sourceHierarchyDatasets.datasetTemplate.kmsKeyName` | GcpKmsKey | `status.outputs.key_id` |
| `spec.destinationConfig.bigqueryDestinationConfig.sourceHierarchyDatasets.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.destinationConfig.bigqueryDestinationConfig.blmtConfig.bucket` | GcpGcsBucket | `status.outputs.bucket_name` |
| `spec.destinationConfig.bigqueryDestinationConfig.blmtConfig.connectionName` | GcpBigQueryConnection | `status.outputs.name` |
| `spec.customerManagedEncryptionKey` | GcpKmsKey | `status.outputs.key_id` |

## See Also

- [Overview](../README.md)
