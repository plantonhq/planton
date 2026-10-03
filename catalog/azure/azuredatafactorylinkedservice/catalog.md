# Azure Data Factory Linked Service

Deploys one linked service inside an Azure Data Factory -- a saved connection in the factory's address book telling pipelines, datasets, and data flows where an external system lives and how to authenticate to it. One kind covers all 23 connection types Azure models as first-class linked service resources -- storage, databases, services, and protocol endpoints -- including a raw-JSON custom form for every other connector Data Factory speaks. Because these are login recipes, most variants carry secrets; the safest patterns, preferred wherever a variant offers them, are managed identity (no secret at all) and Key-Vault-sourced secrets resolved by Data Factory at run time.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions exactly one linked service of the type the spec's variant block declares:

- **Storage** -- Azure Blob Storage, Azure Files, Table Storage, Data Lake Storage Gen2
- **Databases** -- Azure SQL Database, SQL Server, SQL Managed Instance, Synapse, PostgreSQL, MySQL, Cosmos DB (SQL API), Cosmos DB for MongoDB, Snowflake, Kusto, ODBC
- **Services** -- Key Vault, Azure AI Search, Azure Function, Databricks
- **Protocol endpoints** -- web (HTTP), OData, SFTP
- **Custom** -- any other Data Factory connector (Salesforce, SAP, Oracle, REST, and dozens more), as its ARM type name plus type-properties JSON

All types share one factory-scoped name namespace (`{factory_id}/linkedservices/{name}`).

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription.
- **Data Factory** -- the linked service lives in a factory; reference an AzureDataFactory's `data_factory_id` output or provide the ARM ID directly.
- **For Key-Vault-sourced secrets** -- deploy the KEY VAULT variant of this same kind first; other connections' secret blocks reference it by name.

### Azure Subscription

- **Creating a connection does not test it** -- Azure saves the definition without dialing the target; a wrong password or unreachable host surfaces when a pipeline first USES the connection. Use Studio's Test connection button after deploy.
- **Managed identity needs data-plane permission** -- grant the factory's identity the right role on the target (Storage Blob Data Contributor, Key Vault get/list on secrets, a database AAD user); the connection saves fine without it and fails at run time.
- **Private networks need a self-hosted integration runtime** -- set `integrationRuntimeName` to reach systems the shared Azure runtime cannot; an ODBC DSN's driver must be installed on that runtime's machine.

## Deploy

### Console

Open the deployment store, find **Azure Data Factory Linked Service**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Key Vault Connection**, **Blob Storage via Managed Identity**, or **SQL Database with Key Vault Secrets** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureDataFactoryLinkedService
metadata:
  name: lakehouse-blob
  org: acme-corp
  env: prod
spec:
  dataFactoryId:
    valueFrom:
      kind: AzureDataFactory
      name: data-platform
      fieldPath: status.outputs.data_factory_id
  name: lakehouse-blob
  azureBlobStorage:
    serviceEndpoint:
      value: https://acmelakehouse.blob.core.windows.net
    useManagedIdentity: true
```

```shell
planton apply -f data-factory-linked-service.yaml
```

This creates a blob storage connection authenticated as the factory's managed identity -- no connection string, no SAS token, no key to rotate or leak. An Infra Job tracks the provisioning in real time.

For variants that do carry inline secrets (connection strings, passwords, access tokens, service principal keys), reference managed secrets as `$secret/<slug>` instead of pasting values -- or better, use the variant's Key Vault reference block so no secret sits in the manifest at all.

### InfraChart

When deploying the factory, its vault, and its connections as one chart, ValueFromRef wires the linked service to resources deployed in the same InfraPipeline:

```yaml
spec:
  dataFactoryId:
    valueFrom:
      kind: AzureDataFactory
      name: data-platform
      fieldPath: status.outputs.data_factory_id
  azureBlobStorage:
    serviceEndpoint:
      valueFrom:
        kind: AzureStorageAccount
        name: lakehouse
        fieldPath: status.outputs.primary_blob_endpoint
    useManagedIdentity: true
```

The InfraPipeline resolves the dependency graph -- factory and storage account first, then this connection.

## Key Configuration

These are the most important decisions when configuring a linked service. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Deploy the Key Vault connection first, then stop pasting secrets** -- the single highest-leverage move: one `keyVault` connection per factory, then every other connection's password, connection string, or token lives in the vault and is referenced by name (`keyVaultPassword`, `keyVaultConnectionString`, and friends). Rotation becomes a vault operation -- no manifest changes, no redeploys -- and the Key Vault connection itself carries no credential: Data Factory authenticates to the vault as its managed identity.

**Choose the connection form deliberately** -- most variants offer a spectrum from paste-the-key (connection strings, SAS URIs, access tokens) to no-secret (managed identity against a service endpoint). Blob storage, Data Lake Gen2, Azure SQL, and Kusto all support managed identity; prefer it, and remember it needs a data-plane role grant on the target. The blob variant's `connectionStringInsecure` is stored as PLAIN TEXT readable by anyone who can read the factory -- it exists only for strings that carry no secret material.

**Secrets never come back** -- inline secrets are stored as hidden secure strings: ARM reads return them masked or not at all, and an import carries the connection's address but never its credential. Keep the source of truth in Key Vault (or in a managed secret the manifest references), because the deployed resource cannot give it back. Databricks access tokens are masked on every read and can never be imported.

**One name, one connection -- switching types replaces it** -- changing the variant block redeploys the same ARM address as a different type, and every dataset and pipeline referencing the name follows instantly. That is the upgrade lever (swap a paste-the-key blob connection for a managed-identity one with no downstream edits) and the foot-gun (a wrong variant swap breaks every consumer at once). `name` and `dataFactoryId` are ForceNew.

**Private targets ride the integration runtime** -- the default Azure runtime reaches public endpoints only. On-premises SQL Server, ODBC DSNs, and VNet-isolated systems need a self-hosted integration runtime named in `integrationRuntimeName`. The runtime is part of the connection's failure domain: when it is down, every connection through it is down.

**The custom form trades validation for reach** -- it carries any connector Data Factory speaks as raw JSON, with no schema checking until Azure parses it at save time. Keep secrets inside it as Key Vault reference objects, never literals; typed fields, validation, and secret marking are exactly what the custom form gives up, so prefer a first-class variant the moment one exists.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureDataFactory** | `dataFactoryId` | `status.outputs.data_factory_id` |
| **AzureDataFactoryIntegrationRuntime** (optional) | `integrationRuntimeName` | `status.outputs.integration_runtime_name` |
| **AzureStorageAccount** (blob variant) | `azureBlobStorage.serviceEndpoint` | `status.outputs.primary_blob_endpoint` |
| **AzureStorageAccount** (Data Lake Gen2 variant) | `dataLakeStorageGen2.url` | `status.outputs.primary_dfs_endpoint` |
| **AzureKeyVault** (key vault variant) | `keyVault.keyVaultId` | `status.outputs.key_vault_id` |
| **AzureSearchService** (search variant) | `azureSearch.url` | `status.outputs.endpoint` |
| **AzureDataFactoryLinkedService** (every Key Vault secret reference) | `*.linkedServiceName` | `status.outputs.linked_service_name` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `linked_service_name` | The connection's name inside the factory | AzureDataFactoryDataset's `linkedServiceName`, other connections' Key Vault secret references, SSIS package stores and proxy staging |
| `linked_service_id` | The ARM ID (`{factory_id}/linkedservices/{name}`) | The Azure SQL table dataset variant's `linkedServiceId` |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Secrets backbone first** -- one Key Vault connection per factory, deployed before everything else; every later connection references the vault by name instead of pasting credentials. Start from the **Key Vault Connection** preset.

**Managed-identity storage** -- the storage account's blob endpoint plus the factory's own identity: no connection string, no SAS token, nothing to rotate. Grant the identity Storage Blob Data Contributor (or Reader) on the account. Start from the **Blob Storage via Managed Identity** preset.

**Vault-held database credentials** -- an Azure SQL Database connection whose entire connection string lives in Key Vault; the manifest carries an address into the vault, never a credential. Start from the **SQL Database with Key Vault Secrets** preset.

**On-premises bridge** -- a SQL Server or ODBC connection pinned to a self-hosted integration runtime by `integrationRuntimeName`, reaching systems the shared Azure runtime cannot.

## Works With

- [**Azure Data Factory**](/infra-catalog/azure-data-factory) -- the factory the connection lives in, referenced by `dataFactoryId`
- [**Azure Data Factory Dataset**](/infra-catalog/azure-data-factory-dataset) -- every dataset reads through a linked service
- [**Azure Data Factory Integration Runtime**](/infra-catalog/azure-data-factory-integration-runtime) -- the compute the connection runs through, for private targets
- [**Azure Key Vault**](/infra-catalog/azure-key-vault) -- the vault behind the key vault variant and every Key-Vault-sourced secret
- [**Azure Storage Account**](/infra-catalog/azure-storage-account) -- blob and Data Lake Gen2 connections point at its endpoints
- [**Azure AI Search Service**](/infra-catalog/azure-search-service) -- the search variant references its endpoint output
