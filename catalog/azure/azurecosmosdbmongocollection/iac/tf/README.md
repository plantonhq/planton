# AzureCosmosdbMongoCollection - Terraform Module

Terraform implementation for the AzureCosmosdbMongoCollection
kind.

## Resources Created

- `azurerm_cosmosdb_mongo_collection.main` -- the MongoDB API collection,
  addressed by the (resource group, account, database, name) quartet
  azurerm requires, parsed from the parent database's ARM id

## Variable Highlights

| Variable | Notes |
| --- | --- |
| `spec.mongo_database_id` | The parent database's resolved ARM id; RG, account, and database names are parsed from it in `locals.tf` |
| `spec.collection_name` | Required; unique within the database; ForceNew |
| `spec.shard_key` | Optional document property for sharding; ForceNew; unset creates an unsharded collection |
| `spec.throughput` / `spec.autoscale_max_throughput` | Mutually exclusive dedicated throughput dials |
| `spec.indexes` | Mongo-style index blocks (keys + unique) |
| `spec.default_ttl_seconds` | Document TTL; -1 enables TTL without default expiry; 0 rejected |

## Outputs

| Output | Description |
| --- | --- |
| `mongo_collection_id` | The ARM id of the collection |
| `mongo_collection_name` | The collection's name inside the database |
| `mongo_database_name` | The parent database name |
| `cosmosdb_account_name` | The account name completing the addressing triple |
