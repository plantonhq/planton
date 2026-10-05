# AzureMssqlElasticPool - Terraform Module

Terraform implementation for the AzureMssqlElasticPool
kind.

## Resources Created

- `azurerm_mssql_elasticpool.main` -- the pool: derived sku
  (tier + family computed from the sku name), capacity, per-database
  bounds, storage cap, and availability posture

## Variable Highlights

| Variable | Notes |
| --- | --- |
| `spec.server_id` | The parent server's resolved ARM id; the server NAME and RESOURCE GROUP the resource wants are derived from it in `locals.tf` |
| `spec.region` | Must match the parent server's region (ARM rejects a mismatch) |
| `spec.sku_name` | The tier and hardware family are derived through exhaustive lookups -- a name/tier/family mismatch is unrepresentable |
| `spec.max_size_gb` / `max_size_bytes` | Mutually exclusive (spec-validated); both null lets ARM apply the SKU default |
| `spec.enclave_type` / `license_type` | Spec enum name strings mapped to ARM values in `locals.tf` |

## Usage

```hcl
module "mssql_elastic_pool" {
  source = "./path/to/module"

  metadata = {
    name = "tenant-pool"
    org  = "mycompany"
    env  = "production"
  }

  spec = {
    server_id = "/subscriptions/.../providers/Microsoft.Sql/servers/myorg-prod-sql"
    region    = "eastus"
    pool_name = "tenant-pool"
    sku_name  = "StandardPool"
    capacity  = 100

    per_database_settings = {
      min_capacity = 0
      max_capacity = 50
    }
  }
}
```

## Feature Parity

This Terraform module has feature parity with the Pulumi implementation:
DTU and vCore SKUs with derived tier/family, per-database bounds, both
storage-cap shapes, zone redundancy, enclaves, Hybrid Benefit licensing,
Hyperscale HA replicas, maintenance windows, and user tags.
