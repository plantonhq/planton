---
title: "Cosmos DB with SQL API"
description: "This preset creates an Azure Cosmos DB account with the SQL (NoSQL) API, Session consistency, a single geo-location, and one SQL database containing a container with a partition key. The SQL API is..."
type: "preset"
rank: "01"
presetSlug: "01-sql-api"
componentSlug: "cosmos-db-deployment-landscape-and-design-rationale"
componentTitle: "Cosmos DB — Deployment Landscape and Design Rationale"
provider: "azure"
icon: "package"
order: 1
---

# Cosmos DB with SQL API

This preset creates an Azure Cosmos DB account with the SQL (NoSQL) API, Session consistency, a single geo-location, and one SQL database containing a container with a partition key. The SQL API is the most popular Cosmos DB interface, offering SQL-like queries over JSON documents with automatic indexing and low-latency reads. This is the recommended starting configuration for new Cosmos DB workloads.

## When to Use

- Applications needing a globally distributed NoSQL document database with SQL-like queries
- Microservices requiring low-latency reads and writes with flexible JSON schemas
- Event sourcing, user profiles, product catalogs, and IoT telemetry storage
- Teams familiar with SQL syntax who want a schemaless document store

## Key Configuration Choices

- **SQL API** (`kind: GlobalDocumentDB`) -- Query JSON documents with SQL-like syntax. The default and most widely used Cosmos DB API
- **Session consistency** (`consistencyPolicy.consistencyLevel: Session`) -- Read-your-writes within a session. The right default for most applications balancing consistency and performance
- **Single region** (`geoLocations[0]`) -- Primary write region only. Add additional geo-locations for multi-region reads or failover
- **400 RU/s database throughput** (`sqlDatabases[0].throughput: 400`) -- Minimum provisioned throughput, shared across all containers. Scale up or switch to `autoscaleMaxThroughput: 1000` for variable workloads
- **Hash partition key** (`containers[0].partitionKeyKind: Hash`) -- Single-level partitioning on the specified path. Choose a high-cardinality field (e.g., `/tenantId`, `/userId`)
- **No automatic failover** -- Enable `automaticFailoverEnabled` when adding additional geo-locations

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<azure-region>` | Azure region (e.g., "eastus", "westeurope") | Your regional deployment strategy |
| `<your-resource-group-name>` | Name of the resource group | Azure portal or `AzureResourceGroup` status outputs |
| `<your-account-name>` | Globally unique account name (3-50 chars, lowercase/numbers/hyphens) | Choose a name; becomes `https://{name}.documents.azure.com:443/` |
| `<your-database-name>` | Name of the SQL database (1-255 chars) | Your application design |
| `<your-container-name>` | Name of the container (1-255 chars) | Your application design |
| `/partitionKey` | Partition key path (e.g., `/tenantId`, `/userId`, `/region`) | Choose based on query patterns and data cardinality |

## Related Presets

- **02-mongodb-api** -- Use instead for applications requiring MongoDB wire-protocol compatibility
- **03-serverless** -- Use instead for low-traffic or spiky workloads with pay-per-request pricing
