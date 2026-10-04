# AzureCosmosdbSqlRoleDefinition - Terraform Module

Terraform implementation for the AzureCosmosdbSqlRoleDefinition
kind.

## Resources Created

- `azurerm_cosmosdb_sql_role_definition.main` -- the Cosmos DB SQL
  data-plane role definition, addressed by the (resource group,
  account, GUID) trio azurerm requires, parsed from the parent
  account's ARM id

## Variable Highlights

| Variable | Notes |
| --- | --- |
| `spec.cosmosdb_account_id` | The parent account's resolved ARM id; the resource-group and account names azurerm wants are parsed from it in `locals.tf` with named-group regexes that fail the plan loudly on a malformed id |
| `spec.role_name` | Required; the role's display name, unique among the account's definitions; renaming is an in-place update (assignments track the GUID) |
| `spec.type` | Optional proto enum value name (CUSTOM_ROLE / BUILT_IN_ROLE) mapped to ARM's CustomRole / BuiltInRole wire vocabulary in `locals.tf`; unset sends nothing so azurerm's CustomRole default applies |
| `spec.assignable_scopes` | Required (spec-enforced min 1); resolved account/database/container paths -- WHERE assignments may be created; scopes above the account are not enforceable in Cosmos data-plane RBAC |
| `spec.permissions` | Required (spec-enforced min 1 block, min 1 data action each); additive allow-only blocks -- Cosmos has no not_data_actions carve-out |
| `spec.role_definition_id` | Optional pinned GUID; sent only when set -- unset lets the provider generate one at create time |

## Outputs

| Output | Description |
| --- | --- |
| `role_definition_id` | The fully-scoped ARM id -- exactly what an AzureCosmosdbSqlRoleAssignment binds |
| `role_definition_guid` | The definition's GUID resource name (pinned or generated) |
| `role_name` | The display name as deployed |
| `cosmosdb_account_name` | The account name, parsed from the resolved account id |

## Usage

```hcl
module "cosmosdb_sql_role_definition" {
  source = "./path/to/module"

  metadata = {
    name = "app-reader"
    org  = "mycompany"
  }

  spec = {
    cosmosdb_account_id = "/subscriptions/.../providers/Microsoft.DocumentDB/databaseAccounts/my-cosmos"
    role_name           = "app-reader"
    assignable_scopes   = ["/subscriptions/.../providers/Microsoft.DocumentDB/databaseAccounts/my-cosmos"]
    permissions = [{
      data_actions = [
        "Microsoft.DocumentDB/databaseAccounts/readMetadata",
        "Microsoft.DocumentDB/databaseAccounts/sqlDatabases/containers/items/read",
      ]
    }]
  }
}
```

The definition carries no Azure tags: ARM does not support tags on
Cosmos child resources, so the platform's identity tags live on the
parent account.
