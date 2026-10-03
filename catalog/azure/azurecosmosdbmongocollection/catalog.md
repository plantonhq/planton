# Azure Cosmos DB Mongo Collection

Deploys a MongoDB API collection inside a Cosmos DB Mongo database — the unit of storage and scale-out where documents actually live. The shard key is the MongoDB face of the partition key and carries the same design weight: it is fixed at creation, and resharding means a new collection plus a data migration. Account → database → collection compose in one manifest set.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Cosmos DB Mongo Collection** -- a named collection inside the referenced Mongo database, with its shard key and declared indexes (including the `_id` index Azure requires)
- **Dedicated Throughput** (optional) -- fixed RU/s or an autoscale ceiling owned by this collection alone; omit both to share the database's provisioned throughput (or on serverless accounts)
- **TTL Policies** (optional) -- a document TTL (an expireAfter index on `_ts`) and an analytical-store TTL

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Azure Subscription

- **A Cosmos DB Mongo database**. Reference an AzureCosmosdbMongoDatabase Infra Component via ValueFromRef, or provide the database's ARM ID directly.
- **For the analytical TTL**: analytical storage enabled on the account.

## Deploy

### Console

Open the deployment store, find **Azure Cosmos DB Mongo Collection**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Tenant-sharded Mongo collection** preset in the [Presets](#presets) tab for the multi-tenant production shape.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureCosmosdbMongoCollection
metadata:
  name: tenant-events
  org: acme-corp
  env: prod
spec:
  mongoDatabaseId:
    valueFrom:
      kind: AzureCosmosdbMongoDatabase
      name: catalog-database
      fieldPath: status.outputs.mongo_database_id
  collectionName: events
  shardKey: tenantId
  autoscaleMaxThroughput: 4000
  indexes:
    - keys:
        - _id
      unique: true
    - keys:
        - tenantId
        - createdAt
```

```shell
planton apply -f cosmosdb-mongo-collection.yaml
```

This creates a tenant-sharded collection with dedicated autoscale throughput and a compound index for per-tenant recency queries. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, the account, database, and collections compose in one InfraPipeline: the pipeline resolves `cosmosdb_account_id` into the database, then this collection joins via `mongo_database_id`.

## Key Configuration

These are the most important decisions when configuring a Mongo collection. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Shard key** -- The one decision that cannot change. Pick a document property with high cardinality and even request distribution (tenantId, userId, deviceId). Unset creates an UNSHARDED collection confined to a single physical partition — acceptable only for small, bounded collections.

**Indexes** -- Azure REQUIRES an index on `_id` (declare it with `unique: true`); a collection cannot be created without it. Every other index is a query-performance decision: compound indexes serve queries that filter on their keys in order, and each index taxes every write with extra RU/s. Indexes update in place.

**Throughput model** -- Leave both fields unset to share the database's budget; set `throughput` (minimum 400 RU/s, increments of 100) or `autoscaleMaxThroughput` (minimum 1000 RU/s, increments of 1000) for a dedicated budget. Mutually exclusive; neither may be set on serverless accounts.

**TTLs** -- `defaultTtlSeconds`: -1 turns TTL on with per-document opt-in expiry; a positive value expires documents that many seconds after their last write; 0 is invalid; unset leaves TTL off. `analyticalStorageTtl` governs the analytical-store copy and requires analytical storage on the account.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| AzureCosmosdbMongoDatabase | `mongoDatabaseId` | `status.outputs.mongo_database_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `mongo_collection_id` | The ARM ID of the collection | ARM reads, policy targets |
| `mongo_collection_name` | The name Mongo drivers reference inside the database | Application configuration |
| `mongo_database_name` | The parent database's name | Application configuration |
| `cosmosdb_account_name` | The account's name | Connection string composition |

There are deliberately no endpoint or credential outputs here: connectivity and the MongoDB connection strings live on the account (AzureCosmosdbAccount's outputs); the collection is addressed inside that connection by database and collection name.

## Common Patterns

**Tenant-sharded event stream** — shard by `tenantId` with dedicated autoscale throughput, so each tenant's writes land on distinct physical partitions and the RU budget follows the spiky multi-tenant load. The production-default shape for event and audit streams. Start from the **Tenant-sharded Mongo collection** preset.

**Fleet of small collections on a shared budget** — leave both throughput fields unset so the collection draws from the database's shared throughput, while still sharding (e.g. by `userId`) so it scales out across partitions. More economical than per-collection dedication when collections are small and similarly sized; the trade is noisy-neighbor contention inside the shared budget. Start from the **Shared-throughput Mongo collection** preset.

**TTL session store** — `defaultTtlSeconds: 86400` expires every document 24 hours after its last write, so storage stays flat without a cleanup job. Fixed dedicated throughput suits the predictable steady-state load. Start from the **TTL session-store Mongo collection** preset.

## Works With

- [**Azure Cosmos DB Mongo Database**](/infra-catalog/azure-cosmosdb-mongo-database) — the parent database this collection lives in, referenced via `mongo_database_id`
- [**Azure Cosmos DB Account**](/infra-catalog/azure-cosmosdb-account) — the MONGO_DB account that owns connectivity, connection strings, and network posture for everything inside
