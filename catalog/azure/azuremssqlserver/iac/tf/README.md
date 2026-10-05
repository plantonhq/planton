# AzureMssqlServer - Terraform Module

Terraform implementation for the AzureMssqlServer kind.

## Resources Created

- `azurerm_mssql_server.main` -- the logical server: authentication (SQL
  and/or Entra), identity, TDE customer-managed key, connection policy,
  TLS floor, and network dials
- `azurerm_mssql_firewall_rule.main` -- one per `firewall_rules` entry
- `azurerm_mssql_virtual_network_rule.main` -- one per
  `virtual_network_rules` entry
- `azurerm_mssql_outbound_firewall_rule.main` -- one per
  `outbound_firewall_rules` FQDN
- `azurerm_mssql_server_extended_auditing_policy.main` -- when
  `extended_auditing` is present
- `azurerm_mssql_server_security_alert_policy.main` -- when
  `security_alert_policy` is present (addresses the server by name +
  resource group -- that resource's own contract)
- `data.azurerm_client_config.current` -- the tenant fallback for the
  Entra administrator

## Variable Highlights

| Variable | Notes |
| --- | --- |
| `spec.administrator_login` / `_password` | Sent only when non-empty (Entra-only servers omit them) |
| `spec.azuread_administrator.tenant_id` | Falls back to the deploying credential's tenant when unset |
| `spec.connection_policy` | Spec enum name string (`DEFAULT`/`PROXY`/`REDIRECT`); unset is not sent |
| `spec.identity.type` | Spec enum name string mapped to ARM values in `locals.tf` |
| `spec.security_alert_policy.disabled_alerts` | Spec enum name strings mapped to ARM's Snake_Pascal wire vocabulary (`Sql_Injection`, ...) |
| `spec.transparent_data_encryption_key_vault_key_id` | A VERSIONED Key Vault key ID (ARM pins the version at the server level) |

## Usage

```hcl
module "mssql_server" {
  source = "./path/to/module"

  metadata = {
    name = "my-sql"
    org  = "mycompany"
    env  = "production"
  }

  spec = {
    region                 = "eastus"
    resource_group         = "my-rg"
    server_name            = "myorg-prod-sql"
    administrator_login    = "sqladmin"
    administrator_password = "P@ssw0rd1234!"

    firewall_rules = [
      { name = "allow-azure-services", start_ip_address = "0.0.0.0", end_ip_address = "0.0.0.0" }
    ]

    security_alert_policy = {
      state = "ENABLED"
    }
  }
}
```

## Feature Parity

This Terraform module has feature parity with the Pulumi implementation:
SQL/Entra/mixed authentication, managed identity with the primary pin,
the TDE customer-managed key, connection policy, TLS floor, public and
outbound network dials, firewall/vnet/outbound rules, extended auditing,
the Defender alert policy, and user tags. The `identity_principal_id`
output is conditionally populated on both engines only when the identity
type includes SYSTEM_ASSIGNED.
