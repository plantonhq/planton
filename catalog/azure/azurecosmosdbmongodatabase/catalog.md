# Azure Cosmos DB Mongo Database

Deploys a MongoDB API database inside an Azure Cosmos DB account — the namespace collections live in and the boundary for SHARED throughput. Databases are many-per-account with independent lifecycles, which is why they are a first-class kind referencing the account rather than a list folded into the account's spec. Collections are their own kind (AzureCosmosdbMongoCollection) referencing this database's `mongo_database_id` output.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Cosmos DB Mongo Database** -- a named database inside the referenced Cosmos DB account (which must be a MONGO_DB-kind account with the ENABLE_MONGO capability)
- **Shared Throughput** (optional) -- fixed RU/s or an autoscale ceiling that every collection in the database shares, when either is declared; omit both to let each collection bring its own dedicated throughput (or on serverless accounts, where provisioned throughput is rejected)

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Azure Subscription

- **A Cosmos DB account** speaking the MongoDB API (kind MONGO_DB with the ENABLE_MONGO capability). Reference an AzureCosmosdbAccount Infra Component via ValueFromRef, or provide the account's ARM ID directly.

## Deploy

### Console

Open the deployment store, find **Azure Cosmos DB Mongo Database**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Dedicated Collection Throughput** preset in the [Presets](#presets) tab for the common production shape.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureCosmosdbMongoDatabase
metadata:
  name: catalog-database
  org: acme-corp
  env: prod
spec:
  cosmosdbAccountId:
    valueFrom:
      kind: AzureCosmosdbAccount
      name: catalog-db
      fieldPath: status.outputs.cosmosdb_account_id
  databaseName: catalog
```

```shell
planton apply -f cosmosdb-mongo-database.yaml
```

This creates a database with no shared throughput — each collection brings its own. Add `throughput` (fixed RU/s) or `autoscaleMaxThroughput` (a ceiling) for the shared model. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, the account, database, and collections compose in one InfraPipeline: the pipeline deploys the account first, resolves `cosmosdb_account_id` into this database, then resolves this database's `mongo_database_id` into its collections.

## Key Configuration

These are the most important decisions when configuring a Mongo database. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Throughput model** -- The single real decision. Leave both fields unset (the production norm) and each collection provisions its own dedicated throughput; set `throughput` for a fixed shared budget (minimum 400 RU/s, increments of 100); or set `autoscaleMaxThroughput` (minimum 1000 RU/s, increments of 1000) and Azure floats the shared budget between 10% and 100% of it. The two are mutually exclusive. On serverless accounts (the ENABLE_SERVERLESS capability) neither may be set — Azure rejects provisioned throughput at apply.

**Database name** -- Unique within the account, 1-255 characters (Azure's only constraint), and what Mongo drivers address after connecting to the account. Changing the name replaces the database AND everything in it.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| AzureCosmosdbAccount | `cosmosdbAccountId` | `status.outputs.cosmosdb_account_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `mongo_database_id` | The ARM ID of the database | AzureCosmosdbMongoCollection `mongoDatabaseId` |
| `mongo_database_name` | The name Mongo drivers reference inside the account's connection | Application configuration |
| `cosmosdb_account_name` | The parent account's name | Connection string composition |

There are deliberately no endpoint or credential outputs here: connectivity and the MongoDB connection strings live on the account (AzureCosmosdbAccount's outputs); the database is addressed inside that connection by name.

## Common Patterns

**Pure namespace, dedicated collection throughput** — leave both throughput fields unset so the database is only a namespace and each collection provisions its own RU/s. Workloads stay isolated: a hot collection throttles itself, never its siblings. The production default. Start from the **Dedicated Collection Throughput** preset.

**Shared autoscale budget for many small collections** — `autoscaleMaxThroughput` gives every collection inside a shared budget that Azure floats between 10% and 100% of the ceiling. Economical for fleets of small collections that would each waste a dedicated 400 RU/s minimum — but shared means coupled: one hot collection can starve its siblings, so give genuinely hot collections their own throughput on the collection instead. The 10% floor always bills, so size the ceiling to real peaks. Start from the **Shared Autoscale Database** preset.

**Serverless namespace** — on an account with the ENABLE_SERVERLESS capability, the database (and its collections) must leave both throughput fields unset; Azure rejects provisioned throughput at apply. Right for intermittent traffic and workloads whose profile is still unknown. Start from the **Serverless Database** preset.

## Works With

- [**Azure Cosmos DB Account**](/infra-catalog/azure-cosmosdb-account) — the MONGO_DB account (with the ENABLE_MONGO capability) this database lives in, referenced via `cosmosdb_account_id`
- [**Azure Cosmos DB Mongo Collection**](/infra-catalog/azure-cosmosdb-mongo-collection) — the collections inside, referencing this database's `mongo_database_id` output
