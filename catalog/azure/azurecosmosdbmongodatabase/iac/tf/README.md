# AzureCosmosdbMongoDatabase - Terraform Module

Terraform implementation for the AzureCosmosdbMongoDatabase
kind.

## Resources Created

- `azurerm_cosmosdb_mongo_database.main` -- the MongoDB API database,
  addressed by the (resource group, account, name) trio azurerm
  requires, parsed from the parent account's ARM id

## Variable Highlights

| Variable | Notes |
| --- | --- |
| `spec.cosmosdb_account_id` | The parent account's resolved ARM id; the resource-group and account names azurerm wants are parsed from it in `locals.tf` |
| `spec.database_name` | Required; unique within the account; changing it replaces the database and everything in it |
| `spec.throughput` | Optional fixed shared RU/s (min 400, increments of 100); XOR with autoscale |
| `spec.autoscale_max_throughput` | Optional autoscale ceiling (min 1000, increments of 1000) |

## Outputs

| Output | Description |
| --- | --- |
| `mongo_database_id` | The ARM id of the database -- what collections reference |
| `mongo_database_name` | The database's name inside the account |
| `cosmosdb_account_name` | The account name, parsed from the resolved account id |
