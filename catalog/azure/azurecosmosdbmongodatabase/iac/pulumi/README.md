# AzureCosmosdbMongoDatabase - Pulumi Module

Pulumi implementation for the AzureCosmosdbMongoDatabase
kind.

## Resources Created

- `cosmosdb.MongoDatabase` -- the MongoDB API database, addressed by
  the (resource group, account, name) trio the Azure provider requires,
  parsed from the parent account's ARM id via `parseCosmosdbAccountId`

## Key Behaviors

- Parent addressing: `cosmosdb_account_id` is parsed into resource
  group and account names on both engines identically
- Throughput: fixed RU/s and autoscale are mutually exclusive
  (spec-enforced); both omitted on serverless accounts
- Shared Azure provider via `pulumiazureprovider.Get`

## Outputs

| Output | Description |
| --- | --- |
| `mongo_database_id` | ARM id for collection FK references |
| `mongo_database_name` | Database name inside the account |
| `cosmosdb_account_name` | Account name completing the triple |
