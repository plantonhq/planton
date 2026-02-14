# Cosmos DB Serverless (SQL API)

This preset creates an Azure Cosmos DB account in serverless mode with the SQL (NoSQL) API. Serverless mode uses pay-per-request pricing with no provisioned throughput -- you only pay for the Request Units consumed by your operations and the storage used. No `throughput` or `autoscaleMaxThroughput` configuration is needed. This is the most cost-effective option for low-traffic, spiky, or development workloads.

## When to Use

- Development and testing environments with unpredictable or low traffic patterns
- Event-driven workloads with long idle periods punctuated by bursts of activity
- Proof-of-concept and prototyping where cost minimization is the priority
- Applications consuming fewer than ~5000 RU/s on average that benefit from pay-per-request billing

## Key Configuration Choices

- **Serverless mode** (`capabilities: ["EnableServerless"]`) -- Pay-per-request with no provisioned throughput. Maximum burst of 5000 RU/s per container. This capability is ForceNew -- cannot be changed after creation
- **SQL API** (`kind: GlobalDocumentDB`) -- Query JSON documents with SQL-like syntax
- **No throughput configuration** -- Throughput fields (`throughput`, `autoscaleMaxThroughput`) are omitted because serverless mode handles scaling automatically
- **Session consistency** (`consistencyPolicy.consistencyLevel: Session`) -- Read-your-writes within a session
- **Single region only** -- Serverless accounts support only a single write region. Multi-region writes and additional geo-locations are not available
- **No automatic failover** -- Not applicable in serverless single-region mode
- **50 GB storage limit per container** -- Serverless containers have a maximum of 50 GB. Use provisioned throughput for larger data volumes

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

- **01-sql-api** -- Use instead for predictable production workloads where provisioned or autoscale throughput provides better cost efficiency
- **02-mongodb-api** -- Use instead for MongoDB wire-protocol compatibility with provisioned throughput
