# AzureCosmosdbAccount - Terraform Module

Terraform implementation for the AzureCosmosdbAccount
kind.

## Resources Created

- `azurerm_cosmosdb_account.main` -- the Cosmos DB account: API kind,
  regions, consistency, capabilities, network posture, managed
  identity, customer-managed-key encryption, backup, and tags.
  Databases and containers are their own kinds referencing this
  account, so this module creates exactly one resource.

## Variable Highlights

| Variable | Notes |
| --- | --- |
| `spec.account_name` | The globally unique DNS label; ForceNew |
| `spec.kind` | Spec enum name strings (GLOBAL_DOCUMENT_DB/MONGO_DB); unset materializes GlobalDocumentDB; ForceNew |
| `spec.consistency_policy` | Level enum defaulting Session; the BoundedStaleness dials carry the proto defaults (5/100) when unset; the multi-region floors are enforced by the spec before the plan |
| `spec.capabilities` | Spec enum name strings mapped to ARM's exact wire values (including MongoDBv3.4 and mongoEnableDocLevelTTL, which break the EnableX convention); never injected -- a MONGO_DB account declares ENABLE_MONGO itself |
| `spec.backup` | Type/tier/redundancy enum names; per-mode field pairings enforced by the spec; unset dials omitted so Azure's defaults apply |
| `spec.identity` + `spec.default_identity` | The default identity composes the "UserAssignedIdentity=<id>" wire string from the type enum + identity reference |
| `spec.key_vault_key_id` | The resolved versionless Key Vault key URI; sent only when set; ForceNew |
| `spec.local_authentication_enabled` | Inverted onto azurerm's `local_authentication_disabled` (the v4 surviving input) |
| `spec.restore` + `spec.create_mode` | RESTORE creates the account from a continuous-backup restore point; every restore field ForceNew |
| `spec.tags` | Merged over the platform's identity tags; user values win |

## Usage

```hcl
module "cosmosdb_account" {
  source = "./path/to/module"

  metadata = {
    name = "app-cosmos"
    org  = "mycompany"
  }

  spec = {
    region         = "eastus"
    resource_group = "data-rg"
    account_name   = "mycompany-app-cosmos"
    consistency_policy = {
      consistency_level = "SESSION"
    }
    geo_locations = [
      { location = "eastus", failover_priority = 0 }
    ]
  }
}
```

`offer_type` is hardcoded to `Standard` -- ARM accepts no other value,
so there is nothing to model.
