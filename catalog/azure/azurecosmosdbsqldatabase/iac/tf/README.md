# AzureCosmosdbSqlDatabase - Terraform Module

Terraform implementation for the AzureCosmosdbSqlDatabase
kind.

## Resources Created

- `azurerm_cosmosdb_sql_database.main` -- the SQL (NoSQL) API database,
  addressed by the (resource group, account, name) trio azurerm
  requires, parsed from the parent account's ARM id

## Variable Highlights

| Variable | Notes |
| --- | --- |
| `spec.cosmosdb_account_id` | The parent account's resolved ARM id; the resource-group and account names azurerm wants are parsed from it in `locals.tf` with named-group regexes that fail the plan loudly on a malformed id |
| `spec.database_name` | Required; unique within the account; changing it replaces the database and everything in it |
| `spec.throughput` | Optional fixed shared RU/s (min 400, increments of 100 -- spec-enforced); sent only when set, because serverless accounts reject provisioned throughput and unset means containers bring their own |
| `spec.autoscale_max_throughput` | Optional autoscale ceiling (min 1000, increments of 1000 -- spec-enforced); rendered as an `autoscale_settings` block only when set; mutually exclusive with `spec.throughput` (spec-enforced XOR) |

## Outputs

| Output | Description |
| --- | --- |
| `sql_database_id` | The ARM id of the database -- what containers reference |
| `sql_database_name` | The database's name inside the account |
| `cosmosdb_account_name` | The account name, parsed from the resolved account id |

## Usage

```hcl
module "cosmosdb_sql_database" {
  source = "./path/to/module"

  metadata = {
    name = "app-data"
    org  = "mycompany"
  }

  spec = {
    cosmosdb_account_id = "/subscriptions/.../providers/Microsoft.DocumentDB/databaseAccounts/my-cosmos"
    database_name       = "app-data"
  }
}
```

The database carries no Azure tags: ARM does not support tags on
Cosmos child resources, so the platform's identity tags live on the
parent account.
