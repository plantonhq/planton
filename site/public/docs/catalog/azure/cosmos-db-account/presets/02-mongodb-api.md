---
title: "Cosmos DB with MongoDB API"
description: "This preset creates an Azure Cosmos DB account with the MongoDB wire-protocol API, MongoDB server version 6.0, Session consistency, and one MongoDB database containing a sharded collection...."
type: "preset"
rank: "02"
presetSlug: "02-mongodb-api"
componentSlug: "cosmos-db-account"
componentTitle: "Cosmos DB Account"
provider: "azure"
icon: "package"
order: 2
---

# Cosmos DB with MongoDB API

This preset creates an Azure Cosmos DB account with the MongoDB wire-protocol API, MongoDB server version 6.0, Session consistency, and one MongoDB database containing a sharded collection. Applications connect using standard MongoDB drivers and connection strings, while Cosmos DB provides global distribution, automatic scaling, and managed infrastructure. This is the recommended configuration for teams migrating from MongoDB or building new applications with MongoDB-native tooling.

## When to Use

- Migrating existing MongoDB applications to a fully managed service without code changes
- New applications where the team prefers MongoDB query syntax, aggregation pipeline, and drivers
- Workloads requiring MongoDB-compatible wire protocol with Cosmos DB's global distribution
- Applications using MongoDB ODMs (Mongoose, Motor, MongoEngine) that need a managed backend

## Key Configuration Choices

- **MongoDB API** (`kind: MongoDB`) -- Wire-protocol compatible with MongoDB. The IaC modules automatically add the `EnableMongo` capability
- **MongoDB 6.0** (`mongoServerVersion: "6.0"`) -- Latest stable MongoDB wire protocol version. Use "4.2" or "5.0" for compatibility with older drivers
- **Session consistency** (`consistencyPolicy.consistencyLevel: Session`) -- Read-your-writes within a session, matching MongoDB's default causal consistency behavior
- **Single region** (`geoLocations[0]`) -- Primary write region. Add additional geo-locations for multi-region reads or failover
- **400 RU/s database throughput** (`mongoDatabases[0].throughput: 400`) -- Minimum provisioned throughput, shared across all collections. Scale up or switch to `autoscaleMaxThroughput: 1000` for variable workloads
- **Shard key** (`collections[0].shardKey`) -- Equivalent to a partition key. Choose a high-cardinality field for even data distribution

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<azure-region>` | Azure region (e.g., "eastus", "westeurope") | Your regional deployment strategy |
| `<your-resource-group-name>` | Name of the resource group | Azure portal or `AzureResourceGroup` status outputs |
| `<your-account-name>` | Globally unique account name (3-50 chars, lowercase/numbers/hyphens) | Choose a name; connection string uses `{name}.mongo.cosmos.azure.com` |
| `<your-database-name>` | Name of the MongoDB database (1-255 chars) | Your application design |
| `<your-collection-name>` | Name of the collection (1-255 chars) | Your application design |
| `<your-shard-key>` | Shard key field name (e.g., "tenantId", "userId", "_id") | Choose based on query patterns and data cardinality |

## Related Presets

- **01-sql-api** -- Use instead for SQL-like queries over JSON documents without MongoDB protocol dependency
- **03-serverless** -- Use instead for low-traffic or spiky workloads with pay-per-request pricing (SQL API)
