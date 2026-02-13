# AzureMssqlServer Deployment Component

**Date**: February 13, 2026
**Type**: Feature
**Components**: API Definitions, Azure Provider, Pulumi CLI Integration, Provider Framework

## Summary

Forged AzureMssqlServer (R13) as the third database resource in the Azure expansion project, implementing a complete deployment component with 4 proto files, Pulumi and Terraform IaC modules, validation tests, and production-quality documentation. This resource follows the established database pattern from PostgreSQL (R11) and MySQL (R12) but adapted for Azure SQL's fundamentally different logical server + database architecture.

## Problem Statement / Motivation

The Azure resource expansion project (20260212.05.sp.azure-resource-expansion) requires 24 Azure resource kinds for enterprise Azure workloads. Azure SQL Database (MSSQL) is a critical database option for enterprises migrating from on-premises SQL Server or building new T-SQL workloads on Azure.

### Pain Points

- No Azure SQL Database resource kind existed in OpenMCF
- The database-stack infra chart needs MSSQL alongside PostgreSQL and MySQL
- Enterprise customers with existing SQL Server licenses need Azure Hybrid Benefit support
- The T02 planning spec had 14 inaccuracies discovered during deep Terraform provider research

## Solution / What's New

### Complete AzureMssqlServer Deployment Component

A production-ready deployment component with full feature parity between Pulumi and Terraform IaC modules.

### Architectural Adaptation

Azure SQL uses a **logical server + database** model that is fundamentally different from PostgreSQL/MySQL Flexible Servers:

- The server is a logical container (no compute, no storage)
- Each database carries its own compute SKU, storage, zone redundancy, and license type
- Private connectivity is exclusively via Private Endpoint (no VNet delegation)
- The `AzureMssqlDatabase` message is significantly richer than PG/MySQL counterparts

## Implementation Details

### 14 Corrections from T02 Spec

Deep Terraform provider research revealed 14 corrections to the original planning spec:

1. Added `resource_group` (StringValueOrRef) -- missing from T02
2. Added `region` (string) -- missing from T02
3. Kept `version` exposed with default "12.0" via field option
4. Kept `minimum_tls_version` exposed with default "1.2" via field option
5. Changed `initial_database` (singular) to `repeated databases` -- pattern consistency
6. Expanded `AzureMssqlDatabase` with 7 fields (sku_name, max_size_gb, collation, zone_redundant, license_type, storage_account_type)
7. Changed output `database_id` to `database_ids` map -- pattern consistency
8. Added `connection_policy` -- production performance tunable
9. Kept `public_network_access_enabled` as explicit boolean (no VNet delegation to derive from)
10. Applied server name validation (lowercase, hyphens, 3-63 chars)
11. No HA/zone at server level (HA is per-database via zone_redundant)
12. No VNet delegation fields (private access via Private Endpoint only)
13. No storage fields on server (storage is per-database max_size_gb)
14. Firewall rules use ServerId pattern (same as PostgreSQL)

### Proto API Design

- **spec.proto**: 11 server fields + AzureMssqlDatabase (7 fields) + AzureMssqlFirewallRule (3 fields)
- **stack_outputs.proto**: 5 outputs (server_id, server_name, fqdn, administrator_login, database_ids map)
- **api.proto**: KRM wiring with `azure.openmcf.org/v1` API version
- **stack_input.proto**: Azure provider config input

### Pulumi Module

- Package: `github.com/pulumi/pulumi-azure/sdk/v6/go/azure/mssql`
- Uses `mssql.NewServer`, `mssql.NewDatabase`, `mssql.NewFirewallRule`
- Database and firewall rules use `ServerId` pattern (consistent with PostgreSQL)
- MaxSizeGb uses `Float64Ptr` (Pulumi SDK uses float for fractional GB support)
- ConnectionPolicy set on server creation

### Terraform Module

- Provider: `hashicorp/azurerm ~> 4.0`
- Uses `azurerm_mssql_server`, `azurerm_mssql_database`, `azurerm_mssql_firewall_rule`
- `for_each` for databases and firewall rules
- Full feature parity with Pulumi module

### Validation Tests

36 tests covering:
- Valid inputs: minimal server, databases with all fields, firewall rules, all enum values, valueFrom references
- Invalid inputs: missing required fields, invalid versions, invalid TLS versions, invalid connection policies, invalid license types, invalid storage account types

## Benefits

- **Enterprise SQL Server support**: Azure SQL Database with T-SQL compatibility
- **Cost optimization**: Azure Hybrid Benefit (license_type: BasePrice) saves up to 55%
- **Performance tuning**: Connection policy (Redirect) for lower-latency Azure-to-Azure connections
- **Flexible compute**: Each database independently sized (DTU or vCore tiers)
- **Pattern consistency**: Same output structure (database_ids map) as PostgreSQL and MySQL
- **Composability**: server_id output enables Private Endpoint wiring in database-stack infra chart

## Impact

- **R13 completed**: 14 of 24 Azure resources now done
- **Database trifecta**: PostgreSQL, MySQL, and MSSQL all available for database-stack infra chart
- **Enum registered**: AzureMssqlServer = 433 in cloud_resource_kind.proto
- **36/36 tests green**: All validation tests pass
- **~30 files created**: Proto, IaC (Pulumi + Terraform), docs, examples, tests, manifests

## Related Work

- R11: AzurePostgresqlFlexibleServer -- PostgreSQL database pattern (reference implementation)
- R12: AzureMysqlFlexibleServer -- MySQL database pattern
- DD03: Composite Bundling Rules -- server + databases + firewall rules bundled
- DD05: AzureResourceGroup as first-class resource -- StringValueOrRef resource_group pattern

---

**Status**: Production Ready
**Timeline**: Single session (R13 forge)
