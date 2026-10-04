# AzureCosmosdbMongoCollection - Pulumi Module

Pulumi implementation for the AzureCosmosdbMongoCollection
kind.

## Resources Created

- `cosmosdb.MongoCollection` -- the MongoDB API collection, addressed by
  the (resource group, account, database, name) quartet the Azure
  provider requires, parsed from the parent database's ARM id

## Key Behaviors

- Parent addressing: `mongo_database_id` is parsed into resource group,
  account, and database names on both engines identically
- Shard key is ForceNew; indexes render as nested index blocks
- Throughput XOR autoscale enforced by spec CELs
- Shared Azure provider via `pulumiazureprovider.Get`

## Outputs

| Output | Description |
| --- | --- |
| `mongo_collection_id` | ARM id for management-plane references |
| `mongo_collection_name` | Collection name inside the database |
| `mongo_database_name` | Parent database name |
| `cosmosdb_account_name` | Account name completing the triple |
