# AzureCosmosdbSqlContainer - Terraform Module

Terraform implementation for the AzureCosmosdbSqlContainer
kind.

## Resources Created

- `azurerm_cosmosdb_sql_container.main` -- the SQL (NoSQL) API
  container, addressed by the (resource group, account, database, name)
  tuple the provider requires, parsed from the parent database's ARM id

## Variable Highlights

| Variable | Notes |
| --- | --- |
| `spec.sql_database_id` | The parent database's resolved ARM id; the database, account, and resource-group NAMES are derived from it in `locals.tf` (loud plan failure on a malformed id) |
| `spec.partition_key_paths` | 1-3 paths starting with "/"; HASH takes exactly one, MULTI_HASH several with version 2 (spec-enforced pairings) |
| `spec.partition_key_kind` | Spec enum name strings (HASH/MULTI_HASH); unset materializes Hash |
| `spec.throughput` / `spec.autoscale_max_throughput` | Mutually exclusive (spec-enforced); both unset shares the database's throughput -- and is required on serverless accounts |
| `spec.indexing_policy` | Enum name strings mapped to the provider's lowercase modes; a declared policy replaces Azure's default wholesale |
| `spec.conflict_resolution_policy` | LAST_WRITER_WINS/CUSTOM with per-mode fields sent only when set |

## Usage

```hcl
module "sql_container" {
  source = "./path/to/module"

  metadata = {
    name = "orders"
    org  = "mycompany"
  }

  spec = {
    sql_database_id     = "/subscriptions/.../databaseAccounts/app-cosmos/sqlDatabases/app-data"
    container_name      = "orders"
    partition_key_paths = ["/tenantId"]
    throughput          = 400
  }
}
```

Containers carry no Azure tags: ARM does not support tags on Cosmos
child resources, so the platform's identity tags live on the account.
