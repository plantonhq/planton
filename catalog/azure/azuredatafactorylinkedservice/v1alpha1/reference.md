# AzureDataFactoryLinkedService

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `azure.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

**AzureDataFactoryLinkedServiceSpec** defines an Azure Data Factory
linked service -- a saved connection in the factory's address book:
where an external system lives and how pipelines, datasets, and
data flows authenticate to it.

The connection type is declared by which variant block is present:
set exactly ONE of the 23 blocks below. All types share one
factory-scoped name namespace ({factory_id}/linkedservices/{name}).
Storage connections (blob, file share, table, Data Lake Gen2),
databases (Azure SQL, SQL Server, SQL Managed Instance, Synapse,
PostgreSQL, MySQL, Cosmos DB x2, Snowflake, Kusto, ODBC), services
(Key Vault, Azure Search, Azure Function, Databricks), protocol
endpoints (web, OData, SFTP), and the raw-JSON `custom` form for
every connector type azurerm has no first-class resource for.

Because these are login recipes, most variants carry secrets. The
safest patterns, preferred wherever a variant offers them: managed
identity (no secret at all), or a Key-Vault-sourced secret -- a
reference to a KEY VAULT linked service in this same factory plus a
secret name, resolved by Data Factory at run time so no secret ever
sits in this spec.

## Example

```yaml
# Deep-shape example for docs and offline validation: a Databricks
# connection exercising the variant's full surface -- Key-Vault-held
# access token (the safe pattern), a fully-specified job cluster with
# autoscaling, Spark configuration, tags, and init scripts.
# References are literal values so the manifest validates standalone.
apiVersion: azure.planton.dev/v1alpha1
kind: AzureDataFactoryLinkedService
metadata:
  name: test-linked-service
  id: test-linked-service
  org: test-org
  env: test
spec:
  dataFactoryId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.DataFactory/factories/test-df
  name: spark-workspace
  description: Databricks over a Key-Vault-held token, running per-job clusters.
  annotations:
    - team:data
  parameters:
    env: test
  azureDatabricks:
    adbDomain: https://adb-1234567890123456.7.azuredatabricks.net
    keyVaultPassword:
      linkedServiceName:
        value: secrets-vault
      secretName: databricks-token
    newClusterConfig:
      nodeType: Standard_DS3_v2
      clusterVersion: 16.4.x-scala2.12
      minNumberOfWorkers: 2
      maxNumberOfWorkers: 8
      driverNodeType: Standard_DS4_v2
      sparkConfig:
        spark.speculation: "true"
      sparkEnvironmentVariables:
        PIPELINE_ENV: test
      customTags:
        team: data
      initScripts:
        - dbfs:/init/install-libs.sh
      logDestination: dbfs:/cluster-logs
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.dataFactoryId` | `string \| valueFrom` | yes |  | AzureDataFactory (`status.outputs.data_factory_id`) |
| `spec.name` | `string` | yes |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.annotations` | `[]string` |  |  |  |
| `spec.parameters` | `map<string, string>` |  |  |  |
| `spec.additionalProperties` | `map<string, string>` |  |  |  |
| `spec.integrationRuntimeName` | `string \| valueFrom` |  |  | AzureDataFactoryIntegrationRuntime (`status.outputs.integration_runtime_name`) |
| `spec.azureBlobStorage` | `AzureDataFactoryLinkedServiceAzureBlobStorage` |  |  |  |
| `spec.azureBlobStorage.connectionString` | `string` (sensitive) |  |  |  |
| `spec.azureBlobStorage.connectionStringInsecure` | `string` |  |  |  |
| `spec.azureBlobStorage.sasUri` | `string` (sensitive) |  |  |  |
| `spec.azureBlobStorage.serviceEndpoint` | `string \| valueFrom` (sensitive) |  |  | AzureStorageAccount (`status.outputs.primary_blob_endpoint`) |
| `spec.azureBlobStorage.sasTokenLinkedKeyVaultKey` | `AzureDataFactoryLinkedServiceKeyVaultSecretRef` |  |  |  |
| `spec.azureBlobStorage.sasTokenLinkedKeyVaultKey.linkedServiceName` | `string \| valueFrom` | yes |  | AzureDataFactoryLinkedService (`status.outputs.linked_service_name`) |
| `spec.azureBlobStorage.sasTokenLinkedKeyVaultKey.secretName` | `string` | yes |  |  |
| `spec.azureBlobStorage.servicePrincipalLinkedKeyVaultKey` | `AzureDataFactoryLinkedServiceKeyVaultSecretRef` |  |  |  |
| `spec.azureBlobStorage.servicePrincipalLinkedKeyVaultKey.linkedServiceName` | `string \| valueFrom` | yes |  | AzureDataFactoryLinkedService (`status.outputs.linked_service_name`) |
| `spec.azureBlobStorage.servicePrincipalLinkedKeyVaultKey.secretName` | `string` | yes |  |  |
| `spec.azureBlobStorage.storageKind` | `string` |  |  |  |
| `spec.azureBlobStorage.useManagedIdentity` | `bool` |  | `false` |  |
| `spec.azureBlobStorage.servicePrincipalId` | `string` |  |  |  |
| `spec.azureBlobStorage.servicePrincipalKey` | `string` (sensitive) |  |  |  |
| `spec.azureBlobStorage.tenantId` | `string` |  |  |  |
| `spec.azureDatabricks` | `AzureDataFactoryLinkedServiceAzureDatabricks` |  |  |  |
| `spec.azureDatabricks.adbDomain` | `string` | yes |  |  |
| `spec.azureDatabricks.msiWorkspaceId` | `string` |  |  |  |
| `spec.azureDatabricks.accessToken` | `string` (sensitive) |  |  |  |
| `spec.azureDatabricks.keyVaultPassword` | `AzureDataFactoryLinkedServiceKeyVaultSecretRef` |  |  |  |
| `spec.azureDatabricks.keyVaultPassword.linkedServiceName` | `string \| valueFrom` | yes |  | AzureDataFactoryLinkedService (`status.outputs.linked_service_name`) |
| `spec.azureDatabricks.keyVaultPassword.secretName` | `string` | yes |  |  |
| `spec.azureDatabricks.existingClusterId` | `string` |  |  |  |
| `spec.azureDatabricks.newClusterConfig` | `AzureDataFactoryLinkedServiceDatabricksNewCluster` |  |  |  |
| `spec.azureDatabricks.newClusterConfig.nodeType` | `string` | yes |  |  |
| `spec.azureDatabricks.newClusterConfig.clusterVersion` | `string` | yes |  |  |
| `spec.azureDatabricks.newClusterConfig.minNumberOfWorkers` | `int32` |  | `1` |  |
| `spec.azureDatabricks.newClusterConfig.maxNumberOfWorkers` | `int32` |  |  |  |
| `spec.azureDatabricks.newClusterConfig.driverNodeType` | `string` |  |  |  |
| `spec.azureDatabricks.newClusterConfig.logDestination` | `string` |  |  |  |
| `spec.azureDatabricks.newClusterConfig.sparkConfig` | `map<string, string>` |  |  |  |
| `spec.azureDatabricks.newClusterConfig.sparkEnvironmentVariables` | `map<string, string>` |  |  |  |
| `spec.azureDatabricks.newClusterConfig.customTags` | `map<string, string>` |  |  |  |
| `spec.azureDatabricks.newClusterConfig.initScripts` | `[]string` |  |  |  |
| `spec.azureDatabricks.instancePool` | `AzureDataFactoryLinkedServiceDatabricksInstancePool` |  |  |  |
| `spec.azureDatabricks.instancePool.instancePoolId` | `string` | yes |  |  |
| `spec.azureDatabricks.instancePool.clusterVersion` | `string` | yes |  |  |
| `spec.azureDatabricks.instancePool.minNumberOfWorkers` | `int32` |  | `1` |  |
| `spec.azureDatabricks.instancePool.maxNumberOfWorkers` | `int32` |  |  |  |
| `spec.azureFileStorage` | `AzureDataFactoryLinkedServiceAzureFileStorage` |  |  |  |
| `spec.azureFileStorage.connectionString` | `string` (sensitive) | yes |  |  |
| `spec.azureFileStorage.fileShare` | `string` |  |  |  |
| `spec.azureFileStorage.host` | `string` |  |  |  |
| `spec.azureFileStorage.userId` | `string` |  |  |  |
| `spec.azureFileStorage.password` | `string` (sensitive) |  |  |  |
| `spec.azureFileStorage.keyVaultPassword` | `AzureDataFactoryLinkedServiceKeyVaultSecretRef` |  |  |  |
| `spec.azureFileStorage.keyVaultPassword.linkedServiceName` | `string \| valueFrom` | yes |  | AzureDataFactoryLinkedService (`status.outputs.linked_service_name`) |
| `spec.azureFileStorage.keyVaultPassword.secretName` | `string` | yes |  |  |
| `spec.azureFunction` | `AzureDataFactoryLinkedServiceAzureFunction` |  |  |  |
| `spec.azureFunction.url` | `string` | yes |  |  |
| `spec.azureFunction.key` | `string` (sensitive) |  |  |  |
| `spec.azureFunction.keyVaultKey` | `AzureDataFactoryLinkedServiceKeyVaultSecretRef` |  |  |  |
| `spec.azureFunction.keyVaultKey.linkedServiceName` | `string \| valueFrom` | yes |  | AzureDataFactoryLinkedService (`status.outputs.linked_service_name`) |
| `spec.azureFunction.keyVaultKey.secretName` | `string` | yes |  |  |
| `spec.azureSearch` | `AzureDataFactoryLinkedServiceAzureSearch` |  |  |  |
| `spec.azureSearch.url` | `string \| valueFrom` | yes |  | AzureSearchService (`status.outputs.endpoint`) |
| `spec.azureSearch.searchServiceKey` | `string` (sensitive) | yes |  |  |
| `spec.azureSqlDatabase` | `AzureDataFactoryLinkedServiceAzureSqlDatabase` |  |  |  |
| `spec.azureSqlDatabase.connectionString` | `string` (sensitive) |  |  |  |
| `spec.azureSqlDatabase.keyVaultConnectionString` | `AzureDataFactoryLinkedServiceKeyVaultSecretRef` |  |  |  |
| `spec.azureSqlDatabase.keyVaultConnectionString.linkedServiceName` | `string \| valueFrom` | yes |  | AzureDataFactoryLinkedService (`status.outputs.linked_service_name`) |
| `spec.azureSqlDatabase.keyVaultConnectionString.secretName` | `string` | yes |  |  |
| `spec.azureSqlDatabase.keyVaultPassword` | `AzureDataFactoryLinkedServiceKeyVaultSecretRef` |  |  |  |
| `spec.azureSqlDatabase.keyVaultPassword.linkedServiceName` | `string \| valueFrom` | yes |  | AzureDataFactoryLinkedService (`status.outputs.linked_service_name`) |
| `spec.azureSqlDatabase.keyVaultPassword.secretName` | `string` | yes |  |  |
| `spec.azureSqlDatabase.useManagedIdentity` | `bool` |  | `false` |  |
| `spec.azureSqlDatabase.servicePrincipalId` | `string` |  |  |  |
| `spec.azureSqlDatabase.servicePrincipalKey` | `string` (sensitive) |  |  |  |
| `spec.azureSqlDatabase.tenantId` | `string` |  |  |  |
| `spec.azureSqlDatabase.credentialName` | `string` |  |  |  |
| `spec.azureTableStorage` | `AzureDataFactoryLinkedServiceAzureTableStorage` |  |  |  |
| `spec.azureTableStorage.connectionString` | `string` (sensitive) | yes |  |  |
| `spec.cosmosdb` | `AzureDataFactoryLinkedServiceCosmosdb` |  |  |  |
| `spec.cosmosdb.connectionString` | `string` (sensitive) |  |  |  |
| `spec.cosmosdb.accountEndpoint` | `string` |  |  |  |
| `spec.cosmosdb.accountKey` | `string` (sensitive) |  |  |  |
| `spec.cosmosdb.database` | `string` |  |  |  |
| `spec.cosmosdbMongoapi` | `AzureDataFactoryLinkedServiceCosmosdbMongoapi` |  |  |  |
| `spec.cosmosdbMongoapi.connectionString` | `string` (sensitive) | yes |  |  |
| `spec.cosmosdbMongoapi.database` | `string` |  |  |  |
| `spec.cosmosdbMongoapi.serverVersionIs32OrHigher` | `bool` |  | `false` |  |
| `spec.custom` | `AzureDataFactoryLinkedServiceCustom` |  |  |  |
| `spec.custom.type` | `string` | yes |  |  |
| `spec.custom.typePropertiesJson` | `string` | yes |  |  |
| `spec.custom.integrationRuntimeParameters` | `map<string, string>` |  |  |  |
| `spec.dataLakeStorageGen2` | `AzureDataFactoryLinkedServiceDataLakeStorageGen2` |  |  |  |
| `spec.dataLakeStorageGen2.url` | `string \| valueFrom` | yes |  | AzureStorageAccount (`status.outputs.primary_dfs_endpoint`) |
| `spec.dataLakeStorageGen2.useManagedIdentity` | `bool` |  | `false` |  |
| `spec.dataLakeStorageGen2.storageAccountKey` | `string` (sensitive) |  |  |  |
| `spec.dataLakeStorageGen2.servicePrincipalId` | `string` |  |  |  |
| `spec.dataLakeStorageGen2.servicePrincipalKey` | `string` (sensitive) |  |  |  |
| `spec.dataLakeStorageGen2.tenant` | `string` |  |  |  |
| `spec.keyVault` | `AzureDataFactoryLinkedServiceKeyVault` |  |  |  |
| `spec.keyVault.keyVaultId` | `string \| valueFrom` | yes |  | AzureKeyVault (`status.outputs.key_vault_id`) |
| `spec.kusto` | `AzureDataFactoryLinkedServiceKusto` |  |  |  |
| `spec.kusto.kustoEndpoint` | `string` | yes |  |  |
| `spec.kusto.kustoDatabaseName` | `string` | yes |  |  |
| `spec.kusto.useManagedIdentity` | `bool` |  | `false` |  |
| `spec.kusto.servicePrincipalId` | `string` |  |  |  |
| `spec.kusto.servicePrincipalKey` | `string` (sensitive) |  |  |  |
| `spec.kusto.tenant` | `string` |  |  |  |
| `spec.mysql` | `AzureDataFactoryLinkedServiceMysql` |  |  |  |
| `spec.mysql.connectionString` | `string` (sensitive) | yes |  |  |
| `spec.mysql.driverVersion` | `string` |  | `V2` |  |
| `spec.odata` | `AzureDataFactoryLinkedServiceOdata` |  |  |  |
| `spec.odata.url` | `string` | yes |  |  |
| `spec.odata.basicAuthentication` | `AzureDataFactoryLinkedServiceBasicAuth` |  |  |  |
| `spec.odata.basicAuthentication.username` | `string` | yes |  |  |
| `spec.odata.basicAuthentication.password` | `string` (sensitive) | yes |  |  |
| `spec.odbc` | `AzureDataFactoryLinkedServiceOdbc` |  |  |  |
| `spec.odbc.connectionString` | `string` (sensitive) | yes |  |  |
| `spec.odbc.basicAuthentication` | `AzureDataFactoryLinkedServiceBasicAuth` |  |  |  |
| `spec.odbc.basicAuthentication.username` | `string` | yes |  |  |
| `spec.odbc.basicAuthentication.password` | `string` (sensitive) | yes |  |  |
| `spec.postgresql` | `AzureDataFactoryLinkedServicePostgresql` |  |  |  |
| `spec.postgresql.connectionString` | `string` (sensitive) | yes |  |  |
| `spec.sftp` | `AzureDataFactoryLinkedServiceSftp` |  |  |  |
| `spec.sftp.authenticationType` | `string` | yes |  |  |
| `spec.sftp.host` | `string` | yes |  |  |
| `spec.sftp.port` | `int32` | yes |  |  |
| `spec.sftp.username` | `string` | yes |  |  |
| `spec.sftp.password` | `string` (sensitive) |  |  |  |
| `spec.sftp.keyVaultPassword` | `AzureDataFactoryLinkedServiceKeyVaultSecretRef` |  |  |  |
| `spec.sftp.keyVaultPassword.linkedServiceName` | `string \| valueFrom` | yes |  | AzureDataFactoryLinkedService (`status.outputs.linked_service_name`) |
| `spec.sftp.keyVaultPassword.secretName` | `string` | yes |  |  |
| `spec.sftp.privateKeyContentBase64` | `string` (sensitive) |  |  |  |
| `spec.sftp.keyVaultPrivateKeyContentBase64` | `AzureDataFactoryLinkedServiceKeyVaultSecretRef` |  |  |  |
| `spec.sftp.keyVaultPrivateKeyContentBase64.linkedServiceName` | `string \| valueFrom` | yes |  | AzureDataFactoryLinkedService (`status.outputs.linked_service_name`) |
| `spec.sftp.keyVaultPrivateKeyContentBase64.secretName` | `string` | yes |  |  |
| `spec.sftp.privateKeyPath` | `string` |  |  |  |
| `spec.sftp.privateKeyPassphrase` | `string` (sensitive) |  |  |  |
| `spec.sftp.keyVaultPrivateKeyPassphrase` | `AzureDataFactoryLinkedServiceKeyVaultSecretRef` |  |  |  |
| `spec.sftp.keyVaultPrivateKeyPassphrase.linkedServiceName` | `string \| valueFrom` | yes |  | AzureDataFactoryLinkedService (`status.outputs.linked_service_name`) |
| `spec.sftp.keyVaultPrivateKeyPassphrase.secretName` | `string` | yes |  |  |
| `spec.sftp.skipHostKeyValidation` | `bool` |  |  |  |
| `spec.sftp.hostKeyFingerprint` | `string` |  |  |  |
| `spec.snowflake` | `AzureDataFactoryLinkedServiceSnowflake` |  |  |  |
| `spec.snowflake.connectionString` | `string` (sensitive) | yes |  |  |
| `spec.snowflake.keyVaultPassword` | `AzureDataFactoryLinkedServiceKeyVaultSecretRef` |  |  |  |
| `spec.snowflake.keyVaultPassword.linkedServiceName` | `string \| valueFrom` | yes |  | AzureDataFactoryLinkedService (`status.outputs.linked_service_name`) |
| `spec.snowflake.keyVaultPassword.secretName` | `string` | yes |  |  |
| `spec.sqlManagedInstance` | `AzureDataFactoryLinkedServiceSqlManagedInstance` |  |  |  |
| `spec.sqlManagedInstance.connectionString` | `string` (sensitive) |  |  |  |
| `spec.sqlManagedInstance.keyVaultConnectionString` | `AzureDataFactoryLinkedServiceKeyVaultSecretRef` |  |  |  |
| `spec.sqlManagedInstance.keyVaultConnectionString.linkedServiceName` | `string \| valueFrom` | yes |  | AzureDataFactoryLinkedService (`status.outputs.linked_service_name`) |
| `spec.sqlManagedInstance.keyVaultConnectionString.secretName` | `string` | yes |  |  |
| `spec.sqlManagedInstance.keyVaultPassword` | `AzureDataFactoryLinkedServiceKeyVaultSecretRef` |  |  |  |
| `spec.sqlManagedInstance.keyVaultPassword.linkedServiceName` | `string \| valueFrom` | yes |  | AzureDataFactoryLinkedService (`status.outputs.linked_service_name`) |
| `spec.sqlManagedInstance.keyVaultPassword.secretName` | `string` | yes |  |  |
| `spec.sqlManagedInstance.servicePrincipalId` | `string` |  |  |  |
| `spec.sqlManagedInstance.servicePrincipalKey` | `string` (sensitive) |  |  |  |
| `spec.sqlManagedInstance.tenant` | `string` |  |  |  |
| `spec.sqlServer` | `AzureDataFactoryLinkedServiceSqlServer` |  |  |  |
| `spec.sqlServer.connectionString` | `string` (sensitive) |  |  |  |
| `spec.sqlServer.keyVaultConnectionString` | `AzureDataFactoryLinkedServiceKeyVaultSecretRef` |  |  |  |
| `spec.sqlServer.keyVaultConnectionString.linkedServiceName` | `string \| valueFrom` | yes |  | AzureDataFactoryLinkedService (`status.outputs.linked_service_name`) |
| `spec.sqlServer.keyVaultConnectionString.secretName` | `string` | yes |  |  |
| `spec.sqlServer.keyVaultPassword` | `AzureDataFactoryLinkedServiceKeyVaultSecretRef` |  |  |  |
| `spec.sqlServer.keyVaultPassword.linkedServiceName` | `string \| valueFrom` | yes |  | AzureDataFactoryLinkedService (`status.outputs.linked_service_name`) |
| `spec.sqlServer.keyVaultPassword.secretName` | `string` | yes |  |  |
| `spec.sqlServer.userName` | `string` |  |  |  |
| `spec.synapse` | `AzureDataFactoryLinkedServiceSynapse` |  |  |  |
| `spec.synapse.connectionString` | `string` (sensitive) | yes |  |  |
| `spec.synapse.keyVaultPassword` | `AzureDataFactoryLinkedServiceKeyVaultSecretRef` |  |  |  |
| `spec.synapse.keyVaultPassword.linkedServiceName` | `string \| valueFrom` | yes |  | AzureDataFactoryLinkedService (`status.outputs.linked_service_name`) |
| `spec.synapse.keyVaultPassword.secretName` | `string` | yes |  |  |
| `spec.web` | `AzureDataFactoryLinkedServiceWeb` |  |  |  |
| `spec.web.url` | `string` | yes |  |  |
| `spec.web.authenticationType` | `string` | yes |  |  |
| `spec.web.username` | `string` |  |  |  |
| `spec.web.password` | `string` (sensitive) |  |  |  |

## Field Details

### spec.dataFactoryId

`string | valueFrom` · required

The Data Factory the linked service lives in, by ARM ID. Can be a
literal string or a reference to an AzureDataFactory output.

**ForceNew**: changing this destroys and recreates the linked
service.

- references: AzureDataFactory (`status.outputs.data_factory_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureDataFactory, name: <that resource's name>, fieldPath: status.outputs.data_factory_id}} -- a bare string does not parse

### spec.name

`string` · required

The linked service's name -- unique within the factory across all
connection types. Azure's own rule is deliberately loose: a name
is rejected only when it consists ENTIRELY of the characters
- . + ? / < > * % & : \ (mirrored exactly; not tightened).

**ForceNew**: changing this destroys and recreates the linked
service.

- rule: Linked service names must not consist entirely of the characters - . + ? / < > * % & : \
- rule: {"required":true}

### spec.description

`string`

A human-readable description of what the connection is for.

### spec.annotations

`[]string`

Free-form annotation strings stored on the linked service.

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.parameters

`map<string, string>`

Linked-service parameters, keyed by parameter name (string values
only -- the wire grammar Data Factory accepts through this
surface). Datasets and activities can override them per use.

### spec.additionalProperties

`map<string, string>`

Additional properties passed through to Azure as-is -- Data
Factory's escape hatch for connection properties the schema does
not model. Applies to every variant EXCEPT sql_managed_instance
(the one linked service the provider models without this
argument).

### spec.integrationRuntimeName

`string | valueFrom`

The integration runtime the connection runs through, by name --
omit for the factory's default Azure runtime. Set this to a
self-hosted runtime's name to reach private networks. A literal
string or a reference to an AzureDataFactoryIntegrationRuntime's
name output.

- references: AzureDataFactoryIntegrationRuntime (`status.outputs.integration_runtime_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureDataFactoryIntegrationRuntime, name: <that resource's name>, fieldPath: status.outputs.integration_runtime_name}} -- a bare string does not parse

### spec.azureBlobStorage

`AzureDataFactoryLinkedServiceAzureBlobStorage`

Azure Blob Storage. Set exactly one variant block on this spec.

- rule: Set exactly one of connection_string, connection_string_insecure, sas_uri, or service_endpoint
- rule: use_managed_identity and service_principal_id are mutually exclusive

### spec.azureBlobStorage.connectionString

`string` · sensitive

The full connection string, account key included. Stored by Azure
as a hidden secure string. SECRET.

### spec.azureBlobStorage.connectionStringInsecure

`string`

A connection string stored as PLAIN TEXT in Azure -- readable by
anyone who can read the factory. Only for connection strings that
carry no secret material. The provider strips any AccountKey when
comparing reads, mirrored in the modules.

### spec.azureBlobStorage.sasUri

`string` · sensitive

A SAS (shared access signature) URI to the storage account.
SECRET. Pair with sas_token_linked_key_vault_key to keep the
token itself in Key Vault.

### spec.azureBlobStorage.serviceEndpoint

`string | valueFrom` · sensitive

The blob service endpoint URL (e.g.
https://account.blob.core.windows.net) -- defaults to referencing
an AzureStorageAccount's primary_blob_endpoint output. The
no-secret form: authentication comes from managed identity
(use_managed_identity) or a service principal. Azure stores it as
a secure string (mirrored: marked sensitive).

A linked service is the factory's connection definition for a store it
reads from and writes to; it never lives inside the account, so on a
diagram the reference is access, not placement -- the same rule its Key
Vault references already carry.

- references: AzureStorageAccount (`status.outputs.primary_blob_endpoint`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureStorageAccount, name: <that resource's name>, fieldPath: status.outputs.primary_blob_endpoint}} -- a bare string does not parse

### spec.azureBlobStorage.sasTokenLinkedKeyVaultKey

`AzureDataFactoryLinkedServiceKeyVaultSecretRef`

Holds the SAS token in Key Vault (used with sas_uri: the URI
carries the address, the vault carries the token).

### spec.azureBlobStorage.sasTokenLinkedKeyVaultKey.linkedServiceName

`string | valueFrom` · required

The Key Vault linked service's name inside this factory --
defaults to referencing an AzureDataFactoryLinkedService's
linked_service_name output (a linked service of the `key_vault`
variant).

- references: AzureDataFactoryLinkedService (`status.outputs.linked_service_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureDataFactoryLinkedService, name: <that resource's name>, fieldPath: status.outputs.linked_service_name}} -- a bare string does not parse

### spec.azureBlobStorage.sasTokenLinkedKeyVaultKey.secretName

`string` · required

The secret's name inside the vault.

- rule: {"required":true}

### spec.azureBlobStorage.servicePrincipalLinkedKeyVaultKey

`AzureDataFactoryLinkedServiceKeyVaultSecretRef`

Holds the service principal's key in Key Vault (used with
service_endpoint + service_principal_id instead of
service_principal_key).

### spec.azureBlobStorage.servicePrincipalLinkedKeyVaultKey.linkedServiceName

`string | valueFrom` · required

The Key Vault linked service's name inside this factory --
defaults to referencing an AzureDataFactoryLinkedService's
linked_service_name output (a linked service of the `key_vault`
variant).

- references: AzureDataFactoryLinkedService (`status.outputs.linked_service_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureDataFactoryLinkedService, name: <that resource's name>, fieldPath: status.outputs.linked_service_name}} -- a bare string does not parse

### spec.azureBlobStorage.servicePrincipalLinkedKeyVaultKey.secretName

`string` · required

The secret's name inside the vault.

- rule: {"required":true}

### spec.azureBlobStorage.storageKind

`string`

The storage account kind, when Data Factory needs it spelled out.

- rule: {"string":{"in":["","Storage","StorageV2","BlobStorage","BlockBlobStorage"]}}

### spec.azureBlobStorage.useManagedIdentity

`bool` · optional (explicit presence)

Authenticate as the factory's managed identity (with
service_endpoint). Unspecified applies false. Mutually exclusive
with service_principal_id.

- default: `false`

### spec.azureBlobStorage.servicePrincipalId

`string`

The service principal's application (client) ID, as a UUID (with
service_endpoint). Mutually exclusive with use_managed_identity.

- rule: service_principal_id must be a UUID

### spec.azureBlobStorage.servicePrincipalKey

`string` · sensitive

The service principal's client secret. SECRET. Prefer
service_principal_linked_key_vault_key.

### spec.azureBlobStorage.tenantId

`string`

The service principal's Entra tenant ID.

### spec.azureDatabricks

`AzureDataFactoryLinkedServiceAzureDatabricks`

Azure Databricks workspace. Set exactly one variant block on this
spec.

- rule: Set exactly one of msi_workspace_id, access_token, or key_vault_password
- rule: Set exactly one of existing_cluster_id, new_cluster_config, or instance_pool

### spec.azureDatabricks.adbDomain

`string` · required

The Databricks workspace domain, e.g.
https://adb-1234567890123456.7.azuredatabricks.net.

- rule: {"required":true}

### spec.azureDatabricks.msiWorkspaceId

`string`

Authenticate as the factory's managed identity, by the
workspace's ARM resource ID
(/subscriptions/.../Microsoft.Databricks/workspaces/...). One of
the three authentication methods.

### spec.azureDatabricks.accessToken

`string` · sensitive

A Databricks personal access token. SECRET -- and Azure masks it
on every read (asterisks), so it can never be read back or
imported; prefer key_vault_password or msi_workspace_id. One of
the three authentication methods.

### spec.azureDatabricks.keyVaultPassword

`AzureDataFactoryLinkedServiceKeyVaultSecretRef`

Holds the access token in Key Vault. One of the three
authentication methods.

### spec.azureDatabricks.keyVaultPassword.linkedServiceName

`string | valueFrom` · required

The Key Vault linked service's name inside this factory --
defaults to referencing an AzureDataFactoryLinkedService's
linked_service_name output (a linked service of the `key_vault`
variant).

- references: AzureDataFactoryLinkedService (`status.outputs.linked_service_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureDataFactoryLinkedService, name: <that resource's name>, fieldPath: status.outputs.linked_service_name}} -- a bare string does not parse

### spec.azureDatabricks.keyVaultPassword.secretName

`string` · required

The secret's name inside the vault.

- rule: {"required":true}

### spec.azureDatabricks.existingClusterId

`string`

Run against an existing interactive cluster, by cluster ID. One
of the three cluster choices.

### spec.azureDatabricks.newClusterConfig

`AzureDataFactoryLinkedServiceDatabricksNewCluster`

Create a new job cluster for each run. One of the three cluster
choices.

- rule: max_number_of_workers cannot be less than min_number_of_workers

### spec.azureDatabricks.newClusterConfig.nodeType

`string` · required

The worker node type, e.g. Standard_DS3_v2.

- rule: {"required":true}

### spec.azureDatabricks.newClusterConfig.clusterVersion

`string` · required

The Databricks runtime version, e.g. "16.4.x-scala2.12".

- rule: {"required":true}

### spec.azureDatabricks.newClusterConfig.minNumberOfWorkers

`int32` · optional (explicit presence)

The minimum (or fixed) number of workers, 1-25000. Unspecified
applies 1 -- the modules always send the effective value
explicitly.

- default: `1`
- rule: {"int32":{"lte":25000,"gte":1}}

### spec.azureDatabricks.newClusterConfig.maxNumberOfWorkers

`int32`

The maximum number of workers, 1-25000 -- set it above the
minimum to enable autoscaling; omit for a fixed-size cluster. May
not be below the minimum (the provider's own rule).

- rule: {"int32":{"lte":25000,"gte":0}}

### spec.azureDatabricks.newClusterConfig.driverNodeType

`string`

The driver node type -- omit to use node_type.

### spec.azureDatabricks.newClusterConfig.logDestination

`string`

dbfs:/ or abfss:/ location for cluster logs.

### spec.azureDatabricks.newClusterConfig.sparkConfig

`map<string, string>`

Spark configuration properties, keyed by property name.

### spec.azureDatabricks.newClusterConfig.sparkEnvironmentVariables

`map<string, string>`

Environment variables set on the cluster, keyed by name.

### spec.azureDatabricks.newClusterConfig.customTags

`map<string, string>`

Custom tags applied to the cluster's nodes, keyed by tag name.

### spec.azureDatabricks.newClusterConfig.initScripts

`[]string`

Init script locations run on each node at startup.

### spec.azureDatabricks.instancePool

`AzureDataFactoryLinkedServiceDatabricksInstancePool`

Draw cluster nodes from an instance pool. One of the three
cluster choices.

- rule: max_number_of_workers cannot be less than min_number_of_workers

### spec.azureDatabricks.instancePool.instancePoolId

`string` · required

The Databricks instance pool ID.

- rule: {"required":true}

### spec.azureDatabricks.instancePool.clusterVersion

`string` · required

The Databricks runtime version the pooled cluster runs.

- rule: {"required":true}

### spec.azureDatabricks.instancePool.minNumberOfWorkers

`int32` · optional (explicit presence)

The minimum (or fixed) number of workers, 1-25000. Unspecified
applies 1 -- the modules always send the effective value
explicitly.

- default: `1`
- rule: {"int32":{"lte":25000,"gte":1}}

### spec.azureDatabricks.instancePool.maxNumberOfWorkers

`int32`

The maximum number of workers, 1-25000 -- set it above the
minimum for autoscaling; omit for fixed size. May not be below
the minimum (the provider's own rule).

- rule: {"int32":{"lte":25000,"gte":0}}

### spec.azureFileStorage

`AzureDataFactoryLinkedServiceAzureFileStorage`

Azure Files share. Set exactly one variant block on this spec.

### spec.azureFileStorage.connectionString

`string` · required · sensitive

The storage account connection string. SECRET (hidden from ARM
reads).

- rule: {"required":true}

### spec.azureFileStorage.fileShare

`string`

The file share's name. Sent to Azure even when empty (the
provider's own behavior, mirrored).

### spec.azureFileStorage.host

`string`

The file server host, for host-addressed access. Azure never
returns it on reads.

### spec.azureFileStorage.userId

`string`

The user ID, for host-addressed access.

### spec.azureFileStorage.password

`string` · sensitive

The password for host-addressed access. SECRET. Prefer
key_vault_password; if both are set the Key Vault reference wins
(the provider's own behavior, mirrored).

### spec.azureFileStorage.keyVaultPassword

`AzureDataFactoryLinkedServiceKeyVaultSecretRef`

Holds the password in Key Vault instead of password.

### spec.azureFileStorage.keyVaultPassword.linkedServiceName

`string | valueFrom` · required

The Key Vault linked service's name inside this factory --
defaults to referencing an AzureDataFactoryLinkedService's
linked_service_name output (a linked service of the `key_vault`
variant).

- references: AzureDataFactoryLinkedService (`status.outputs.linked_service_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureDataFactoryLinkedService, name: <that resource's name>, fieldPath: status.outputs.linked_service_name}} -- a bare string does not parse

### spec.azureFileStorage.keyVaultPassword.secretName

`string` · required

The secret's name inside the vault.

- rule: {"required":true}

### spec.azureFunction

`AzureDataFactoryLinkedServiceAzureFunction`

Azure Function app. Set exactly one variant block on this spec.

- rule: Set exactly one of key or key_vault_key

### spec.azureFunction.url

`string` · required

The function app's URL, e.g. https://app.azurewebsites.net (the
AzureFunctionApp kind's default_hostname output carries the bare
hostname -- prepend https:// when wiring it here).

- rule: {"required":true}

### spec.azureFunction.key

`string` · sensitive

The function's host key. SECRET (hidden from ARM reads). One of
the two key forms.

### spec.azureFunction.keyVaultKey

`AzureDataFactoryLinkedServiceKeyVaultSecretRef`

Holds the function key in Key Vault. One of the two key forms.

### spec.azureFunction.keyVaultKey.linkedServiceName

`string | valueFrom` · required

The Key Vault linked service's name inside this factory --
defaults to referencing an AzureDataFactoryLinkedService's
linked_service_name output (a linked service of the `key_vault`
variant).

- references: AzureDataFactoryLinkedService (`status.outputs.linked_service_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureDataFactoryLinkedService, name: <that resource's name>, fieldPath: status.outputs.linked_service_name}} -- a bare string does not parse

### spec.azureFunction.keyVaultKey.secretName

`string` · required

The secret's name inside the vault.

- rule: {"required":true}

### spec.azureSearch

`AzureDataFactoryLinkedServiceAzureSearch`

Azure AI Search service. Set exactly one variant block on this
spec.

### spec.azureSearch.url

`string | valueFrom` · required

The search service's endpoint URL -- defaults to referencing an
AzureSearchService's endpoint output.

- references: AzureSearchService (`status.outputs.endpoint`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureSearchService, name: <that resource's name>, fieldPath: status.outputs.endpoint}} -- a bare string does not parse

### spec.azureSearch.searchServiceKey

`string` · required · sensitive

The search service's admin key. SECRET (hidden from ARM reads;
the provider leaves it unmarked -- this spec does not repeat that
oversight).

- rule: {"required":true}

### spec.azureSqlDatabase

`AzureDataFactoryLinkedServiceAzureSqlDatabase`

Azure SQL Database. Set exactly one variant block on this spec.

- rule: Set exactly one of connection_string or key_vault_connection_string
- rule: use_managed_identity and service_principal_id are mutually exclusive
- rule: service_principal_id and service_principal_key must be set together

### spec.azureSqlDatabase.connectionString

`string` · sensitive

The connection string. SECRET by content (it typically carries
credentials; Azure hides it from reads when it does). One of the
two connection forms.

### spec.azureSqlDatabase.keyVaultConnectionString

`AzureDataFactoryLinkedServiceKeyVaultSecretRef`

Holds the whole connection string in Key Vault. One of the two
connection forms.

### spec.azureSqlDatabase.keyVaultConnectionString.linkedServiceName

`string | valueFrom` · required

The Key Vault linked service's name inside this factory --
defaults to referencing an AzureDataFactoryLinkedService's
linked_service_name output (a linked service of the `key_vault`
variant).

- references: AzureDataFactoryLinkedService (`status.outputs.linked_service_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureDataFactoryLinkedService, name: <that resource's name>, fieldPath: status.outputs.linked_service_name}} -- a bare string does not parse

### spec.azureSqlDatabase.keyVaultConnectionString.secretName

`string` · required

The secret's name inside the vault.

- rule: {"required":true}

### spec.azureSqlDatabase.keyVaultPassword

`AzureDataFactoryLinkedServiceKeyVaultSecretRef`

Holds the SQL password in Key Vault (used with a connection
string that omits it).

### spec.azureSqlDatabase.keyVaultPassword.linkedServiceName

`string | valueFrom` · required

The Key Vault linked service's name inside this factory --
defaults to referencing an AzureDataFactoryLinkedService's
linked_service_name output (a linked service of the `key_vault`
variant).

- references: AzureDataFactoryLinkedService (`status.outputs.linked_service_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureDataFactoryLinkedService, name: <that resource's name>, fieldPath: status.outputs.linked_service_name}} -- a bare string does not parse

### spec.azureSqlDatabase.keyVaultPassword.secretName

`string` · required

The secret's name inside the vault.

- rule: {"required":true}

### spec.azureSqlDatabase.useManagedIdentity

`bool` · optional (explicit presence)

Authenticate as the factory's managed identity. Unspecified
applies false. Mutually exclusive with service_principal_id.

- default: `false`

### spec.azureSqlDatabase.servicePrincipalId

`string`

The service principal's application (client) ID, as a UUID.
Requires service_principal_key; mutually exclusive with
use_managed_identity.

- rule: service_principal_id must be a UUID

### spec.azureSqlDatabase.servicePrincipalKey

`string` · sensitive

The service principal's client secret. SECRET (hidden from ARM
reads; the provider leaves it unmarked -- this spec does not
repeat that oversight). Requires service_principal_id.

### spec.azureSqlDatabase.tenantId

`string`

The service principal's Entra tenant ID.

### spec.azureSqlDatabase.credentialName

`string`

The name of a Data Factory credential (a user-assigned managed
identity registered on the factory) to authenticate with.

### spec.azureTableStorage

`AzureDataFactoryLinkedServiceAzureTableStorage`

Azure Table Storage. Set exactly one variant block on this spec.

### spec.azureTableStorage.connectionString

`string` · required · sensitive

The storage account connection string. SECRET (hidden from ARM
reads).

- rule: {"required":true}

### spec.cosmosdb

`AzureDataFactoryLinkedServiceCosmosdb`

Azure Cosmos DB (SQL API). Set exactly one variant block on this
spec.

- rule: Set exactly one connection form -- connection_string, or account_endpoint + account_key (+ database)
- rule: account_endpoint, account_key, and database must all be set for account-detail authentication

### spec.cosmosdb.connectionString

`string` · sensitive

The Cosmos DB connection string. SECRET (hidden from ARM reads).
One of the two connection forms.

### spec.cosmosdb.accountEndpoint

`string`

The account endpoint URL (https://account.documents.azure.com).
With account_key and database, the other connection form.

### spec.cosmosdb.accountKey

`string` · sensitive

The account access key. SECRET (hidden from ARM reads).

### spec.cosmosdb.database

`string`

The database name. REQUIRED with account_endpoint/account_key;
optional alongside connection_string (the string may carry its
own Database= segment).

### spec.cosmosdbMongoapi

`AzureDataFactoryLinkedServiceCosmosdbMongoapi`

Azure Cosmos DB for MongoDB. Set exactly one variant block on
this spec.

### spec.cosmosdbMongoapi.connectionString

`string` · required · sensitive

The MongoDB connection string
(mongodb://account:key@account.documents.azure.com:...). SECRET
(hidden from ARM reads).

- rule: {"required":true}

### spec.cosmosdbMongoapi.database

`string`

The database name.

### spec.cosmosdbMongoapi.serverVersionIs32OrHigher

`bool` · optional (explicit presence)

Whether the Cosmos DB account's MongoDB server version is 3.2 or
higher. Unspecified applies false -- the modules always send the
effective value explicitly.

- default: `false`

### spec.custom

`AzureDataFactoryLinkedServiceCustom`

Any other connector type, as raw type-properties JSON -- the
escape hatch for the many Data Factory connectors azurerm has no
first-class resource for. Set exactly one variant block on this
spec.

### spec.custom.type

`string` · required

The connector's ARM type discriminator, e.g. "RestService",
"Salesforce", "SapTableResource" (Data Factory's REST API
vocabulary).

**ForceNew**: changing this destroys and recreates the linked
service.

- rule: {"required":true}

### spec.custom.typePropertiesJson

`string` · required

The connector's typeProperties object, as a JSON string -- the
shape Data Factory's REST API documents for the chosen type.

- rule: {"required":true}

### spec.custom.integrationRuntimeParameters

`map<string, string>`

Parameter values passed to the integration runtime named at the
spec root, keyed by parameter name -- only the custom form
carries these (the provider's integration_runtime block; its name
comes from the root integration_runtime_name).

### spec.dataLakeStorageGen2

`AzureDataFactoryLinkedServiceDataLakeStorageGen2`

Azure Data Lake Storage Gen2. Set exactly one variant block on
this spec.

- rule: Set exactly one authentication mode -- use_managed_identity, storage_account_key, or a service principal (service_principal_id + service_principal_key + tenant)
- rule: A service principal needs service_principal_id, service_principal_key, and tenant together

### spec.dataLakeStorageGen2.url

`string | valueFrom` · required

The Data Lake endpoint URL (https://account.dfs.core.windows.net)
-- defaults to referencing an AzureStorageAccount's
primary_dfs_endpoint output.

A linked service is the factory's connection definition for a store it
reads from and writes to; it never lives inside the account, so on a
diagram the reference is access, not placement.

- references: AzureStorageAccount (`status.outputs.primary_dfs_endpoint`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureStorageAccount, name: <that resource's name>, fieldPath: status.outputs.primary_dfs_endpoint}} -- a bare string does not parse

### spec.dataLakeStorageGen2.useManagedIdentity

`bool` · optional (explicit presence)

Authenticate as the factory's managed identity -- the no-secret
mode. Unspecified applies false. One of the three authentication
modes.

- default: `false`

### spec.dataLakeStorageGen2.storageAccountKey

`string` · sensitive

The storage account access key. SECRET (hidden from ARM reads).
One of the three authentication modes.

### spec.dataLakeStorageGen2.servicePrincipalId

`string`

The service principal's application (client) ID, as a UUID. With
service_principal_key and tenant, one of the three authentication
modes.

- rule: service_principal_id must be a UUID

### spec.dataLakeStorageGen2.servicePrincipalKey

`string` · sensitive

The service principal's client secret. SECRET (hidden from ARM
reads).

### spec.dataLakeStorageGen2.tenant

`string`

The service principal's Entra tenant ID.

### spec.keyVault

`AzureDataFactoryLinkedServiceKeyVault`

Azure Key Vault -- the connection OTHER linked services'
Key-Vault-sourced secrets resolve through. Set exactly one
variant block on this spec.

### spec.keyVault.keyVaultId

`string | valueFrom` · required

The Key Vault, by ARM ID -- defaults to referencing an
AzureKeyVault's key_vault_id output. The modules derive the
vault's base URI from it, exactly as the provider does. A linked
service POINTS AT the vault and lives in its factory, so the
reference is access, not placement, on a diagram.

- references: AzureKeyVault (`status.outputs.key_vault_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureKeyVault, name: <that resource's name>, fieldPath: status.outputs.key_vault_id}} -- a bare string does not parse

### spec.kusto

`AzureDataFactoryLinkedServiceKusto`

Azure Data Explorer (Kusto). Set exactly one variant block on
this spec.

- rule: Set exactly one authentication mode -- use_managed_identity or a service principal (service_principal_id + service_principal_key)
- rule: service_principal_key is required with service_principal_id, and tenant belongs to the service principal mode

### spec.kusto.kustoEndpoint

`string` · required

The cluster endpoint URI
(https://cluster.region.kusto.windows.net).

- rule: {"required":true}

### spec.kusto.kustoDatabaseName

`string` · required

The database inside the cluster.

- rule: {"required":true}

### spec.kusto.useManagedIdentity

`bool` · optional (explicit presence)

Authenticate as the factory's managed identity -- the no-secret
mode. Unspecified applies false. One of the two authentication
modes.

- default: `false`

### spec.kusto.servicePrincipalId

`string`

The service principal's application (client) ID, as a UUID. With
service_principal_key, the other authentication mode.

- rule: service_principal_id must be a UUID

### spec.kusto.servicePrincipalKey

`string` · sensitive

The service principal's client secret. SECRET (hidden from ARM
reads). Requires service_principal_id.

### spec.kusto.tenant

`string`

The service principal's Entra tenant ID. Requires
service_principal_id.

### spec.mysql

`AzureDataFactoryLinkedServiceMysql`

MySQL server. Set exactly one variant block on this spec.

### spec.mysql.connectionString

`string` · required · sensitive

The MySQL connection string. SECRET by content (Azure strips the
password from reads).

- rule: {"required":true}

### spec.mysql.driverVersion

`string` · optional (explicit presence)

The MySQL driver line: "V2" (the current driver) or "V1" (the
legacy driver -- Azure accepts it only on connections that
already use it; the provider rejects V1 on NEW connections).
Unspecified applies V2 -- the modules always send the effective
value explicitly.

- default: `V2`
- rule: {"string":{"in":["V1","V2"]}}

### spec.odata

`AzureDataFactoryLinkedServiceOdata`

OData endpoint. Set exactly one variant block on this spec.

### spec.odata.url

`string` · required

The OData service's URL.

- rule: {"required":true}

### spec.odata.basicAuthentication

`AzureDataFactoryLinkedServiceBasicAuth`

Username/password access -- omit for anonymous.

### spec.odata.basicAuthentication.username

`string` · required

The username.

- rule: {"required":true}

### spec.odata.basicAuthentication.password

`string` · required · sensitive

The password. SECRET -- prefer a pipeline secret reference over a
literal. Azure never returns it on reads.

- rule: {"required":true}

### spec.odbc

`AzureDataFactoryLinkedServiceOdbc`

ODBC data source. Set exactly one variant block on this spec.

### spec.odbc.connectionString

`string` · required · sensitive

The ODBC connection string (DSN or driver form). SECRET by
content -- credentials frequently ride inside it.

- rule: {"required":true}

### spec.odbc.basicAuthentication

`AzureDataFactoryLinkedServiceBasicAuth`

Username/password access -- omit for anonymous.

### spec.odbc.basicAuthentication.username

`string` · required

The username.

- rule: {"required":true}

### spec.odbc.basicAuthentication.password

`string` · required · sensitive

The password. SECRET -- prefer a pipeline secret reference over a
literal. Azure never returns it on reads.

- rule: {"required":true}

### spec.postgresql

`AzureDataFactoryLinkedServicePostgresql`

PostgreSQL server. Set exactly one variant block on this spec.

### spec.postgresql.connectionString

`string` · required · sensitive

The PostgreSQL connection string
(Host=...;Port=...;Database=...;UID=...;Password=...). SECRET
(hidden from ARM reads).

- rule: {"required":true}

### spec.sftp

`AzureDataFactoryLinkedServiceSftp`

SFTP server. Set exactly one variant block on this spec.

- rule: Set at most one of password or key_vault_password
- rule: Set at most one of private_key_content_base64, key_vault_private_key_content_base64, or private_key_path
- rule: Set at most one of private_key_passphrase or key_vault_private_key_passphrase
- rule: password / key_vault_password require authentication_type Basic or MultiFactor
- rule: private key fields require authentication_type SshPublicKey or MultiFactor

### spec.sftp.authenticationType

`string` · required

How Data Factory authenticates: "Basic" (password),
"SshPublicKey" (private key), or "MultiFactor".

- rule: {"required":true,"string":{"in":["Basic","MultiFactor","SshPublicKey"]}}

### spec.sftp.host

`string` · required

The SFTP server's hostname or IP.

- rule: {"required":true}

### spec.sftp.port

`int32` · required

The SFTP server's port (usually 22). The provider accepts any
integer here (mirrored; not tightened).

- rule: {"required":true}

### spec.sftp.username

`string` · required

The username.

- rule: {"required":true}

### spec.sftp.password

`string` · sensitive

The password (Basic/MultiFactor). SECRET. Prefer
key_vault_password.

### spec.sftp.keyVaultPassword

`AzureDataFactoryLinkedServiceKeyVaultSecretRef`

Holds the password in Key Vault instead of password.

### spec.sftp.keyVaultPassword.linkedServiceName

`string | valueFrom` · required

The Key Vault linked service's name inside this factory --
defaults to referencing an AzureDataFactoryLinkedService's
linked_service_name output (a linked service of the `key_vault`
variant).

- references: AzureDataFactoryLinkedService (`status.outputs.linked_service_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureDataFactoryLinkedService, name: <that resource's name>, fieldPath: status.outputs.linked_service_name}} -- a bare string does not parse

### spec.sftp.keyVaultPassword.secretName

`string` · required

The secret's name inside the vault.

- rule: {"required":true}

### spec.sftp.privateKeyContentBase64

`string` · sensitive

The SSH private key, base64-encoded
(SshPublicKey/MultiFactor). SECRET. Prefer
key_vault_private_key_content_base64. The base64 framing is taught
here rather than enforced by a validation rule, because sensitive
fields hold a managed-secret reference on consuming platforms and a
content-shape rule would reject every reference.

### spec.sftp.keyVaultPrivateKeyContentBase64

`AzureDataFactoryLinkedServiceKeyVaultSecretRef`

Holds the base64 private key in Key Vault instead of
private_key_content_base64.

### spec.sftp.keyVaultPrivateKeyContentBase64.linkedServiceName

`string | valueFrom` · required

The Key Vault linked service's name inside this factory --
defaults to referencing an AzureDataFactoryLinkedService's
linked_service_name output (a linked service of the `key_vault`
variant).

- references: AzureDataFactoryLinkedService (`status.outputs.linked_service_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureDataFactoryLinkedService, name: <that resource's name>, fieldPath: status.outputs.linked_service_name}} -- a bare string does not parse

### spec.sftp.keyVaultPrivateKeyContentBase64.secretName

`string` · required

The secret's name inside the vault.

- rule: {"required":true}

### spec.sftp.privateKeyPath

`string`

A file path to the private key on the integration runtime's own
disk -- only meaningful on a self-hosted runtime.

### spec.sftp.privateKeyPassphrase

`string` · sensitive

The private key's passphrase, if it has one. SECRET. Prefer
key_vault_private_key_passphrase.

### spec.sftp.keyVaultPrivateKeyPassphrase

`AzureDataFactoryLinkedServiceKeyVaultSecretRef`

Holds the passphrase in Key Vault instead of
private_key_passphrase.

### spec.sftp.keyVaultPrivateKeyPassphrase.linkedServiceName

`string | valueFrom` · required

The Key Vault linked service's name inside this factory --
defaults to referencing an AzureDataFactoryLinkedService's
linked_service_name output (a linked service of the `key_vault`
variant).

- references: AzureDataFactoryLinkedService (`status.outputs.linked_service_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureDataFactoryLinkedService, name: <that resource's name>, fieldPath: status.outputs.linked_service_name}} -- a bare string does not parse

### spec.sftp.keyVaultPrivateKeyPassphrase.secretName

`string` · required

The secret's name inside the vault.

- rule: {"required":true}

### spec.sftp.skipHostKeyValidation

`bool` · optional (explicit presence)

Whether to skip verifying the server's host key. Unspecified
applies false (Azure's default). When false, set
host_key_fingerprint so the runtime can verify the server.

### spec.sftp.hostKeyFingerprint

`string`

The server's host key fingerprint, verified when
skip_host_key_validation is false.

### spec.snowflake

`AzureDataFactoryLinkedServiceSnowflake`

Snowflake warehouse. Set exactly one variant block on this spec.

### spec.snowflake.connectionString

`string` · required · sensitive

The Snowflake connection string
(jdbc:snowflake://account.snowflakecomputing.com/?user=...&db=...
&warehouse=...). SECRET by content -- credentials frequently ride
inside it.

- rule: {"required":true}

### spec.snowflake.keyVaultPassword

`AzureDataFactoryLinkedServiceKeyVaultSecretRef`

Holds the Snowflake password in Key Vault (used with a connection
string that omits it).

### spec.snowflake.keyVaultPassword.linkedServiceName

`string | valueFrom` · required

The Key Vault linked service's name inside this factory --
defaults to referencing an AzureDataFactoryLinkedService's
linked_service_name output (a linked service of the `key_vault`
variant).

- references: AzureDataFactoryLinkedService (`status.outputs.linked_service_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureDataFactoryLinkedService, name: <that resource's name>, fieldPath: status.outputs.linked_service_name}} -- a bare string does not parse

### spec.snowflake.keyVaultPassword.secretName

`string` · required

The secret's name inside the vault.

- rule: {"required":true}

### spec.sqlManagedInstance

`AzureDataFactoryLinkedServiceSqlManagedInstance`

Azure SQL Managed Instance. Set exactly one variant block on this
spec.

- rule: Set exactly one of connection_string or key_vault_connection_string
- rule: service_principal_id, service_principal_key, and tenant must be set together

### spec.sqlManagedInstance.connectionString

`string` · sensitive

The connection string. SECRET by content (Azure strips the
password from reads). One of the two connection forms.

### spec.sqlManagedInstance.keyVaultConnectionString

`AzureDataFactoryLinkedServiceKeyVaultSecretRef`

Holds the whole connection string in Key Vault. One of the two
connection forms.

### spec.sqlManagedInstance.keyVaultConnectionString.linkedServiceName

`string | valueFrom` · required

The Key Vault linked service's name inside this factory --
defaults to referencing an AzureDataFactoryLinkedService's
linked_service_name output (a linked service of the `key_vault`
variant).

- references: AzureDataFactoryLinkedService (`status.outputs.linked_service_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureDataFactoryLinkedService, name: <that resource's name>, fieldPath: status.outputs.linked_service_name}} -- a bare string does not parse

### spec.sqlManagedInstance.keyVaultConnectionString.secretName

`string` · required

The secret's name inside the vault.

- rule: {"required":true}

### spec.sqlManagedInstance.keyVaultPassword

`AzureDataFactoryLinkedServiceKeyVaultSecretRef`

Holds the SQL password in Key Vault (used with a connection
string that omits it).

### spec.sqlManagedInstance.keyVaultPassword.linkedServiceName

`string | valueFrom` · required

The Key Vault linked service's name inside this factory --
defaults to referencing an AzureDataFactoryLinkedService's
linked_service_name output (a linked service of the `key_vault`
variant).

- references: AzureDataFactoryLinkedService (`status.outputs.linked_service_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureDataFactoryLinkedService, name: <that resource's name>, fieldPath: status.outputs.linked_service_name}} -- a bare string does not parse

### spec.sqlManagedInstance.keyVaultPassword.secretName

`string` · required

The secret's name inside the vault.

- rule: {"required":true}

### spec.sqlManagedInstance.servicePrincipalId

`string`

The service principal's application (client) ID, as a UUID. All
three service principal fields travel together.

- rule: service_principal_id must be a UUID

### spec.sqlManagedInstance.servicePrincipalKey

`string` · sensitive

The service principal's client secret. SECRET (hidden from ARM
reads).

### spec.sqlManagedInstance.tenant

`string`

The service principal's Entra tenant ID, as a UUID.

- rule: tenant must be a UUID

### spec.sqlServer

`AzureDataFactoryLinkedServiceSqlServer`

SQL Server (on-premises or IaaS). Set exactly one variant block
on this spec.

- rule: Set exactly one of connection_string or key_vault_connection_string

### spec.sqlServer.connectionString

`string` · sensitive

The connection string. SECRET by content (Azure strips the
password from reads). One of the two connection forms.

### spec.sqlServer.keyVaultConnectionString

`AzureDataFactoryLinkedServiceKeyVaultSecretRef`

Holds the whole connection string in Key Vault. One of the two
connection forms.

### spec.sqlServer.keyVaultConnectionString.linkedServiceName

`string | valueFrom` · required

The Key Vault linked service's name inside this factory --
defaults to referencing an AzureDataFactoryLinkedService's
linked_service_name output (a linked service of the `key_vault`
variant).

- references: AzureDataFactoryLinkedService (`status.outputs.linked_service_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureDataFactoryLinkedService, name: <that resource's name>, fieldPath: status.outputs.linked_service_name}} -- a bare string does not parse

### spec.sqlServer.keyVaultConnectionString.secretName

`string` · required

The secret's name inside the vault.

- rule: {"required":true}

### spec.sqlServer.keyVaultPassword

`AzureDataFactoryLinkedServiceKeyVaultSecretRef`

Holds the SQL password in Key Vault (used with a connection
string that omits it).

### spec.sqlServer.keyVaultPassword.linkedServiceName

`string | valueFrom` · required

The Key Vault linked service's name inside this factory --
defaults to referencing an AzureDataFactoryLinkedService's
linked_service_name output (a linked service of the `key_vault`
variant).

- references: AzureDataFactoryLinkedService (`status.outputs.linked_service_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureDataFactoryLinkedService, name: <that resource's name>, fieldPath: status.outputs.linked_service_name}} -- a bare string does not parse

### spec.sqlServer.keyVaultPassword.secretName

`string` · required

The secret's name inside the vault.

- rule: {"required":true}

### spec.sqlServer.userName

`string`

The SQL username, when carried outside the connection string.

### spec.synapse

`AzureDataFactoryLinkedServiceSynapse`

Azure Synapse Analytics (dedicated SQL pool). Set exactly one
variant block on this spec.

### spec.synapse.connectionString

`string` · required · sensitive

The Synapse SQL connection string. SECRET by content (Azure
strips the password from reads).

- rule: {"required":true}

### spec.synapse.keyVaultPassword

`AzureDataFactoryLinkedServiceKeyVaultSecretRef`

Holds the SQL password in Key Vault (used with a connection
string that omits it).

### spec.synapse.keyVaultPassword.linkedServiceName

`string | valueFrom` · required

The Key Vault linked service's name inside this factory --
defaults to referencing an AzureDataFactoryLinkedService's
linked_service_name output (a linked service of the `key_vault`
variant).

- references: AzureDataFactoryLinkedService (`status.outputs.linked_service_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureDataFactoryLinkedService, name: <that resource's name>, fieldPath: status.outputs.linked_service_name}} -- a bare string does not parse

### spec.synapse.keyVaultPassword.secretName

`string` · required

The secret's name inside the vault.

- rule: {"required":true}

### spec.web

`AzureDataFactoryLinkedServiceWeb`

Plain HTTP(S) web endpoint. Set exactly one variant block on this
spec.

- rule: authentication_type Basic requires username and password

### spec.web.url

`string` · required

The endpoint URL.

- rule: {"required":true}

### spec.web.authenticationType

`string` · required

How Data Factory authenticates: "Anonymous" or "Basic".

- rule: {"required":true,"string":{"in":["Anonymous","Basic"]}}

### spec.web.username

`string`

The username (Basic).

### spec.web.password

`string` · sensitive

The password (Basic). SECRET. Azure never returns it on reads.

## Validation Rules

- `azure_data_factory_linked_service_exactly_one_variant`: Set exactly one connection variant block -- the variant determines the linked service type

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AzureDataFactoryLinkedService, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.linked_service_id` | `string` | The linked service's Azure Resource Manager ID ({factory_id}/linkedservices/{name}) -- the same ID shape for all connection types. |
| `status.outputs.linked_service_name` | `string` | The linked service's name -- what datasets, data flows, and the Key-Vault-sourced secret references of OTHER linked services resolve against. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.dataFactoryId` | AzureDataFactory | `status.outputs.data_factory_id` |
| `spec.integrationRuntimeName` | AzureDataFactoryIntegrationRuntime | `status.outputs.integration_runtime_name` |
| `spec.azureBlobStorage.serviceEndpoint` | AzureStorageAccount | `status.outputs.primary_blob_endpoint` |
| `spec.azureBlobStorage.sasTokenLinkedKeyVaultKey.linkedServiceName` | AzureDataFactoryLinkedService | `status.outputs.linked_service_name` |
| `spec.azureBlobStorage.servicePrincipalLinkedKeyVaultKey.linkedServiceName` | AzureDataFactoryLinkedService | `status.outputs.linked_service_name` |
| `spec.azureDatabricks.keyVaultPassword.linkedServiceName` | AzureDataFactoryLinkedService | `status.outputs.linked_service_name` |
| `spec.azureFileStorage.keyVaultPassword.linkedServiceName` | AzureDataFactoryLinkedService | `status.outputs.linked_service_name` |
| `spec.azureFunction.keyVaultKey.linkedServiceName` | AzureDataFactoryLinkedService | `status.outputs.linked_service_name` |
| `spec.azureSearch.url` | AzureSearchService | `status.outputs.endpoint` |
| `spec.azureSqlDatabase.keyVaultConnectionString.linkedServiceName` | AzureDataFactoryLinkedService | `status.outputs.linked_service_name` |
| `spec.azureSqlDatabase.keyVaultPassword.linkedServiceName` | AzureDataFactoryLinkedService | `status.outputs.linked_service_name` |
| `spec.dataLakeStorageGen2.url` | AzureStorageAccount | `status.outputs.primary_dfs_endpoint` |
| `spec.keyVault.keyVaultId` | AzureKeyVault | `status.outputs.key_vault_id` |
| `spec.sftp.keyVaultPassword.linkedServiceName` | AzureDataFactoryLinkedService | `status.outputs.linked_service_name` |
| `spec.sftp.keyVaultPrivateKeyContentBase64.linkedServiceName` | AzureDataFactoryLinkedService | `status.outputs.linked_service_name` |
| `spec.sftp.keyVaultPrivateKeyPassphrase.linkedServiceName` | AzureDataFactoryLinkedService | `status.outputs.linked_service_name` |
| `spec.snowflake.keyVaultPassword.linkedServiceName` | AzureDataFactoryLinkedService | `status.outputs.linked_service_name` |
| `spec.sqlManagedInstance.keyVaultConnectionString.linkedServiceName` | AzureDataFactoryLinkedService | `status.outputs.linked_service_name` |
| `spec.sqlManagedInstance.keyVaultPassword.linkedServiceName` | AzureDataFactoryLinkedService | `status.outputs.linked_service_name` |
| `spec.sqlServer.keyVaultConnectionString.linkedServiceName` | AzureDataFactoryLinkedService | `status.outputs.linked_service_name` |
| `spec.sqlServer.keyVaultPassword.linkedServiceName` | AzureDataFactoryLinkedService | `status.outputs.linked_service_name` |
| `spec.synapse.keyVaultPassword.linkedServiceName` | AzureDataFactoryLinkedService | `status.outputs.linked_service_name` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AzureDataFactoryDataFlow | `spec.sources[].linkedService.name` | `status.outputs.linked_service_name` |
| AzureDataFactoryDataFlow | `spec.sources[].schemaLinkedService.name` | `status.outputs.linked_service_name` |
| AzureDataFactoryDataFlow | `spec.sinks[].linkedService.name` | `status.outputs.linked_service_name` |
| AzureDataFactoryDataFlow | `spec.sinks[].schemaLinkedService.name` | `status.outputs.linked_service_name` |
| AzureDataFactoryDataFlow | `spec.sinks[].rejectedLinkedService.name` | `status.outputs.linked_service_name` |
| AzureDataFactoryDataFlow | `spec.transformations[].linkedService.name` | `status.outputs.linked_service_name` |
| AzureDataFactoryDataset | `spec.linkedServiceName` | `status.outputs.linked_service_name` |
| AzureDataFactoryDataset | `spec.azureSqlTable.linkedServiceId` | `status.outputs.linked_service_id` |
| AzureDataFactoryDataset | `spec.custom.linkedService.name` | `status.outputs.linked_service_name` |
| AzureDataFactoryIntegrationRuntime | `spec.azureSsis.expressCustomSetup.commandKey[].keyVaultPassword.linkedServiceName` | `status.outputs.linked_service_name` |
| AzureDataFactoryIntegrationRuntime | `spec.azureSsis.expressCustomSetup.component[].keyVaultLicense.linkedServiceName` | `status.outputs.linked_service_name` |
| AzureDataFactoryIntegrationRuntime | `spec.azureSsis.packageStore[].linkedServiceName` | `status.outputs.linked_service_name` |
| AzureDataFactoryIntegrationRuntime | `spec.azureSsis.proxy.stagingStorageLinkedServiceName` | `status.outputs.linked_service_name` |
| AzureDataFactoryLinkedService | `spec.azureBlobStorage.sasTokenLinkedKeyVaultKey.linkedServiceName` | `status.outputs.linked_service_name` |
| AzureDataFactoryLinkedService | `spec.azureBlobStorage.servicePrincipalLinkedKeyVaultKey.linkedServiceName` | `status.outputs.linked_service_name` |
| AzureDataFactoryLinkedService | `spec.azureDatabricks.keyVaultPassword.linkedServiceName` | `status.outputs.linked_service_name` |
| AzureDataFactoryLinkedService | `spec.azureFileStorage.keyVaultPassword.linkedServiceName` | `status.outputs.linked_service_name` |
| AzureDataFactoryLinkedService | `spec.azureFunction.keyVaultKey.linkedServiceName` | `status.outputs.linked_service_name` |
| AzureDataFactoryLinkedService | `spec.azureSqlDatabase.keyVaultConnectionString.linkedServiceName` | `status.outputs.linked_service_name` |
| AzureDataFactoryLinkedService | `spec.azureSqlDatabase.keyVaultPassword.linkedServiceName` | `status.outputs.linked_service_name` |
| AzureDataFactoryLinkedService | `spec.sftp.keyVaultPassword.linkedServiceName` | `status.outputs.linked_service_name` |
| AzureDataFactoryLinkedService | `spec.sftp.keyVaultPrivateKeyContentBase64.linkedServiceName` | `status.outputs.linked_service_name` |
| AzureDataFactoryLinkedService | `spec.sftp.keyVaultPrivateKeyPassphrase.linkedServiceName` | `status.outputs.linked_service_name` |
| AzureDataFactoryLinkedService | `spec.snowflake.keyVaultPassword.linkedServiceName` | `status.outputs.linked_service_name` |
| AzureDataFactoryLinkedService | `spec.sqlManagedInstance.keyVaultConnectionString.linkedServiceName` | `status.outputs.linked_service_name` |
| AzureDataFactoryLinkedService | `spec.sqlManagedInstance.keyVaultPassword.linkedServiceName` | `status.outputs.linked_service_name` |
| AzureDataFactoryLinkedService | `spec.sqlServer.keyVaultConnectionString.linkedServiceName` | `status.outputs.linked_service_name` |
| AzureDataFactoryLinkedService | `spec.sqlServer.keyVaultPassword.linkedServiceName` | `status.outputs.linked_service_name` |
| AzureDataFactoryLinkedService | `spec.synapse.keyVaultPassword.linkedServiceName` | `status.outputs.linked_service_name` |

## See Also

- [Overview](../README.md)
