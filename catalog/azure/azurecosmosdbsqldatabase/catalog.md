# Azure Cosmos DB SQL Database

Deploys a SQL (NoSQL) API database inside an Azure Cosmos DB account — the namespace containers live in and the boundary for SHARED throughput. Databases are many-per-account with independent lifecycles, which is why they are a first-class kind referencing the account rather than a list folded into the account's spec. Containers are their own kind (AzureCosmosdbSqlContainer) referencing this database's `sql_database_id` output.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Cosmos DB SQL Database** -- a named database inside the referenced Cosmos DB account (which must speak the SQL/NoSQL API)
- **Shared Throughput** (optional) -- fixed RU/s or an autoscale ceiling that every container in the database shares, when either is declared; omit both to let each container bring its own dedicated throughput (or on serverless accounts, where provisioned throughput is rejected)

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Azure Subscription

- **A Cosmos DB account** speaking the SQL (NoSQL) API. Reference an AzureCosmosdbAccount Infra Component via ValueFromRef, or provide the account's ARM ID directly.

## Deploy

### Console

Open the deployment store, find **Azure Cosmos DB SQL Database**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Dedicated Container Throughput** preset in the [Presets](#presets) tab for the common production shape.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureCosmosdbSqlDatabase
metadata:
  name: orders-database
  org: acme-corp
  env: prod
spec:
  cosmosdbAccountId:
    valueFrom:
      kind: AzureCosmosdbAccount
      name: orders-db
      fieldPath: status.outputs.cosmosdb_account_id
  databaseName: orders
```

```shell
planton apply -f cosmosdb-sql-database.yaml
```

This creates a database with no shared throughput — each container brings its own. Add `throughput` (fixed RU/s) or `autoscaleMaxThroughput` (a ceiling) for the shared model. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, the account, database, and containers compose in one InfraPipeline: the pipeline deploys the account first, resolves `cosmosdb_account_id` into this database, then resolves this database's `sql_database_id` into its containers.

## Key Configuration

These are the most important decisions when configuring a SQL database. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Throughput model** -- The single real decision. Leave both fields unset (the production norm) and each container provisions its own dedicated throughput; set `throughput` for a fixed shared budget (minimum 400 RU/s, increments of 100); or set `autoscaleMaxThroughput` (minimum 1000 RU/s, increments of 1000) and Azure floats the shared budget between 10% and 100% of it. The two are mutually exclusive. On serverless accounts (the ENABLE_SERVERLESS capability) neither may be set — Azure rejects provisioned throughput at apply.

**Database name** -- Unique within the account, 1-255 characters (Azure's only constraint). Changing the name replaces the database AND everything in it.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| AzureCosmosdbAccount | `cosmosdbAccountId` | `status.outputs.cosmosdb_account_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `sql_database_id` | The ARM ID of the database | AzureCosmosdbSqlContainer `sqlDatabaseId` |
| `sql_database_name` | The database's name | Application configuration |
| `cosmosdb_account_name` | The parent account's name | Connection string composition |

There are deliberately no endpoint or credential outputs here: connectivity and keys live on the account (AzureCosmosdbAccount's outputs); the database is addressed inside that connection by name.

## Common Patterns

**Pure namespace, dedicated container throughput** — leave both throughput fields unset so the database is only a namespace and each container provisions its own RU/s. Workloads stay isolated: a hot container throttles itself, never its siblings. The production default. Start from the **Dedicated Container Throughput** preset.

**Shared autoscale budget for many small containers** — `autoscaleMaxThroughput` gives every container inside a shared budget Azure floats between 10% and 100% of the ceiling. Economical when each container would waste a dedicated 400 RU/s minimum — but shared means coupled: one hot container can starve its siblings, so keep genuinely hot containers out of the pool by giving them their own throughput. The 10% floor always bills, so size the ceiling to real peaks. Start from the **Shared Autoscale Database** preset.

**Serverless namespace** — on an account with the ENABLE_SERVERLESS capability, the database (and its containers) must leave both throughput fields unset; Azure rejects provisioned throughput at apply. Right for intermittent traffic and workloads whose profile is still unknown. Start from the **Serverless Database** preset.

## Works With

- [**Azure Cosmos DB Account**](/infra-catalog/azure-cosmosdb-account) — the SQL-API account this database lives in, referenced via `cosmosdb_account_id`
- [**Azure Cosmos DB SQL Container**](/infra-catalog/azure-cosmosdb-sql-container) — the containers inside, referencing this database's `sql_database_id` output
- [**Azure Cosmos DB SQL Role Assignment**](/infra-catalog/azure-cosmosdb-sql-role-assignment) — data-plane grants scoped to this database (`{account-id}/dbs/{database-name}`)
