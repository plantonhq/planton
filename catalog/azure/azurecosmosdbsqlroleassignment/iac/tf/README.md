# AzureCosmosdbSqlRoleAssignment - Terraform Module

Terraform implementation for the AzureCosmosdbSqlRoleAssignment
kind.

## Resources Created

- `azurerm_cosmosdb_sql_role_assignment.main` -- the Cosmos DB SQL
  data-plane grant record, addressed by the (resource group, account,
  GUID) trio azurerm requires, parsed from the parent account's ARM id

## Variable Highlights

| Variable | Notes |
| --- | --- |
| `spec.cosmosdb_account_id` | The parent account's resolved ARM id; the resource-group and account names azurerm wants are parsed from it in `locals.tf` with named-group regexes that fail the plan loudly on a malformed id |
| `spec.role_definition_id` | Required; a built-in's well-known id or a custom definition's resolved output. The provider validates the id's shape at plan time (`ValidateSqlRoleDefinitionID`). Rebinding is the assignment's one in-place update |
| `spec.principal_id` | Required; the Entra OBJECT id (a client id deploys but grants nothing -- no directory object carries it) |
| `spec.scope` | Required; the account's ARM id or a `dbs/{db}[/colls/{container}]` path under it; must sit at or below one of the definition's assignable scopes (Azure enforces the pairing at apply) |
| `spec.name` | Optional pinned GUID; sent only when set -- unset lets the provider generate one at create time |

## Outputs

| Output | Description |
| --- | --- |
| `role_assignment_id` | The fully-scoped ARM id of the grant record |
| `role_assignment_guid` | The assignment's GUID resource name (pinned or generated) |
| `cosmosdb_account_name` | The account name, parsed from the resolved account id |

## Usage

```hcl
module "cosmosdb_sql_role_assignment" {
  source = "./path/to/module"

  metadata = {
    name = "app-data-access"
    org  = "mycompany"
  }

  spec = {
    cosmosdb_account_id = "/subscriptions/.../providers/Microsoft.DocumentDB/databaseAccounts/my-cosmos"
    role_definition_id  = "/subscriptions/.../databaseAccounts/my-cosmos/sqlRoleDefinitions/00000000-0000-0000-0000-000000000002"
    principal_id        = "c3b2a190-8f7e-4d6c-b5a4-93d2c1b0a987"
    scope               = "/subscriptions/.../providers/Microsoft.DocumentDB/databaseAccounts/my-cosmos"
  }
}
```

The assignment carries no Azure tags: ARM does not support tags on
Cosmos child resources, so the platform's identity tags live on the
parent account.
