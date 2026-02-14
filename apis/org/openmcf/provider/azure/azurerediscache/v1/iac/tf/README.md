# AzureRedisCache Terraform Module

This Terraform module provisions an Azure Cache for Redis instance with optional
firewall rules and patch schedules.

## Resources Created

- `azurerm_redis_cache.main` -- The Redis cache instance
- `azurerm_redis_firewall_rule.rules` -- IP-based firewall rules (via `for_each`)

## Key Implementation Details

### SKU Family Auto-Derivation

The `family` local is computed from `sku_name`: `"P"` for Premium, `"C"` for
Basic/Standard. This is in `locals.tf`.

### Patch Schedules

Patch schedules use a `dynamic` block to handle zero or more schedules.

### Firewall Rules

Firewall rules use `for_each` keyed by rule name for stable resource addresses.

## Usage

```bash
terraform init
terraform plan -var-file=terraform.tfvars
terraform apply
```

## Inputs

See `variables.tf` for the full variable specification.

## Outputs

| Output | Description |
|--------|-------------|
| `redis_id` | Azure Resource Manager ID |
| `hostname` | Cache hostname |
| `ssl_port` | SSL port (6380) |
| `primary_access_key` | Primary authentication key (sensitive) |
| `primary_connection_string` | Connection string (sensitive) |
