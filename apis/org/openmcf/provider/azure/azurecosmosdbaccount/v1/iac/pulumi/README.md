# AzureCosmosdbAccount - Pulumi Module

## Overview

This Pulumi module provisions an Azure Cosmos DB account with optional SQL API
databases/containers or MongoDB API databases/collections using the Azure provider.

## Components Created

| Resource | Pulumi Constructor | Identifier |
|----------|---------------------|------------|
| Cosmos DB Account | `cosmosdb.Account` | Account name |
| SQL Database | `cosmosdb.SqlDatabase` | `{account}-{db}` |
| SQL Container | `cosmosdb.SqlContainer` | `{account}-{db}-{container}` |
| Mongo Database | `cosmosdb.MongoDatabase` | `{account}-{db}` |
| Mongo Collection | `cosmosdb.MongoCollection` | `{account}-{db}-{collection}` |

## Architecture

The module creates a Cosmos DB account and, based on `kind`, provisions either
SQL API (GlobalDocumentDB) or MongoDB API sub-resources. SQL databases contain
containers with partition keys; MongoDB databases contain collections with
shard keys and optional indexes.

## How to Run

```bash
make deps    # go mod tidy
make build   # compile module and entrypoint
make test    # run tests
make run     # run the Pulumi program
```

## Outputs

- `account_id` - ARM resource ID
- `account_name` - Account name
- `endpoint` - Document endpoint URI
- `primary_key` - Primary access key
- `primary_connection_string` - SQL API connection string
- `primary_mongodb_connection_string` - MongoDB API connection string
- `database_ids` - Map of database names to ARM IDs
