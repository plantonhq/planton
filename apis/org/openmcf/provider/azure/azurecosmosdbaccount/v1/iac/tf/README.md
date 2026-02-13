# AzureCosmosdbAccount - Terraform Module

## Overview

This Terraform module provisions an Azure Cosmos DB account with optional SQL API
databases/containers or MongoDB API databases/collections using the `azurerm`
provider (~> 4.0).

## Resources Created

| Resource | Type | For Each |
|----------|------|----------|
| Cosmos DB Account | `azurerm_cosmosdb_account` | 1 |
| SQL Database | `azurerm_cosmosdb_sql_database` | Per `spec.sql_databases` (when kind = GlobalDocumentDB) |
| SQL Container | `azurerm_cosmosdb_sql_container` | Per container in sql_databases |
| Mongo Database | `azurerm_cosmosdb_mongo_database` | Per `spec.mongo_databases` (when kind = MongoDB) |
| Mongo Collection | `azurerm_cosmosdb_mongo_collection` | Per collection in mongo_databases |

## Required Variables

- `metadata` - Resource metadata (name, org, env, etc.)
- `spec.region` - Azure region
- `spec.resource_group` - Resource group name
- `spec.name` - Cosmos DB account name (globally unique)
- `spec.geo_locations` - At least one geo location with failover_priority

## Outputs

| Output | Description |
|--------|-------------|
| `account_id` | ARM resource ID of the Cosmos DB account |
| `account_name` | Account name |
| `endpoint` | Document endpoint URI |
| `primary_key` | Primary access key (sensitive) |
| `primary_connection_string` | SQL API connection string (sensitive) |
| `primary_mongodb_connection_string` | MongoDB API connection string (sensitive) |
| `database_ids` | Map of database names to ARM IDs |

## Usage

```hcl
module "cosmos" {
  source = "./path/to/module"

  metadata = {
    name = "my-cosmos"
    org  = "mycompany"
    env  = "production"
  }

  spec = {
    region        = "eastus"
    resource_group = "my-rg"
    name          = "my-cosmos-account"
    kind          = "GlobalDocumentDB"
    geo_locations = [{
      location          = "eastus"
      failover_priority = 0
    }]
    sql_databases = [{
      name       = "myapp"
      throughput = 400
      containers = [{
        name                = "items"
        partition_key_paths = ["/tenantId"]
        throughput          = 400
      }]
    }]
  }
}
```
