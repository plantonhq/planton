# AzureMssqlDatabase - Terraform Module

Terraform implementation for the AzureMssqlDatabase kind.

## Resources Created

- `azurerm_mssql_database.main` -- the database: SKU or pool membership,
  storage, availability, lifecycle mode, encryption, backups, and threat
  detection. The azurerm resource internally orchestrates ARM's TDE,
  retention, and threat-detection sub-APIs, so those are blocks here
  rather than separate resources.

## Variable Highlights

| Variable | Notes |
| --- | --- |
| `spec.server_id` | The parent server's resolved ARM id (the resource's native addressing) |
| `spec.sku_name` | Empty maps to null so Azure computes its serverless default; "ElasticPool" pairs with `elastic_pool_id` (spec-validated) |
| `spec.create_mode` + sources | Spec enum name strings mapped to ARM values in `locals.tf`; the mode↔source pairings are spec-validated before the plan runs |
| `spec.enclave_type` / `storage_account_type` / `license_type` / `secondary_type` | Spec enum name strings mapped through exhaustive lookups (a missing entry fails the plan loudly) |
| `spec.threat_detection_policy.email_account_admins` | A bool in the spec, mapped to this resource's Enabled/Disabled wire STRING (unlike the server-scope policy's bool) |
| `spec.max_size_gb` | A number -- fractional ARM sizes (0.5, 0.1) survive |

## Usage

```hcl
module "mssql_database" {
  source = "./path/to/module"

  metadata = {
    name = "orders-db"
    org  = "mycompany"
    env  = "production"
  }

  spec = {
    server_id     = "/subscriptions/.../providers/Microsoft.Sql/servers/myorg-prod-sql"
    database_name = "orders"
    sku_name      = "GP_Gen5_2"
    max_size_gb   = 128

    short_term_retention_policy = {
      retention_days = 14
    }

    long_term_retention_policy = {
      weekly_retention  = "P12W"
      monthly_retention = "P12M"
    }
  }
}
```

## Feature Parity

This Terraform module has feature parity with the Pulumi implementation:
every SKU family (DTU/vCore/serverless/Hyperscale/DW/pooled), all eight
create modes with their sources, serverless dials, read replicas and
scale-out, zone redundancy, ledger, enclaves, database-scoped CMK with
rotation, bacpac import, short/long-term retention, threat detection,
and user tags.
