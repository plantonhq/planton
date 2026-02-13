# AzureMysqlFlexibleServer Deployment Component (R12)

**Date**: February 13, 2026
**Type**: Feature
**Components**: Azure Provider, API Definitions, Pulumi IaC, Terraform IaC, Documentation

## Summary

Forged AzureMysqlFlexibleServer (enum 434, id_prefix `azmysql`) as the second database resource in the Azure resource expansion, following the PostgreSQL pattern from R11 with 13 corrections discovered through deep provider research against the Terraform provider source (API v2023-12-30) and Pulumi azure SDK v6.28.0. All 41 spec validation tests pass, both IaC modules build clean.

## Problem Statement / Motivation

The Azure resource expansion project requires 24 resource kinds to cover enterprise Azure workloads. AzureMysqlFlexibleServer is critical for the database-stack infra chart, serving organizations that use MySQL as their primary relational database.

### Pain Points

- T02 spec design had 13 inaccuracies that would have produced broken or suboptimal IaC modules
- MySQL Flexible Server has significant structural differences from PostgreSQL (storage, database/firewall API, auth) that required careful provider research
- Without MySQL support, the database-stack infra chart cannot serve MySQL-based workloads

## Solution / What's New

Complete AzureMysqlFlexibleServer deployment component with dual IaC (Pulumi + Terraform), following the composite bundling pattern (server + databases + firewall rules per DD03).

### 13 Corrections from T02 Spec

1. Added `resource_group` (StringValueOrRef) -- missing from T02, per DD05
2. Added `region` (string) -- missing from T02, per established pattern
3. Changed version from proto enum to string+CEL with 3 values: "5.7", "8.0.21", "8.4"
4. Changed storage from `int32 storage_size_gb` with range validation (20-16384 GB)
5. Added `auto_grow_enabled` (default **true** -- opposite of PostgreSQL's false)
6. Changed HA from bool to optional `AzureMysqlHighAvailability` message
7. Changed database from `string initial_database_name` to `repeated AzureMysqlDatabase`
8. Added `geo_redundant_backup_enabled` (ForceNew field)
9. Added `zone` (availability zone)
10. Backup retention range corrected to 1-35 (MySQL allows 1, PostgreSQL requires 7+)
11. Server name validation allows starting with digit (`^[a-z0-9]...`)
12. Administrator password as polymorphic StringValueOrRef (confirmed correct)
13. Added `database_ids` map output

## Implementation Details

### Critical MySQL vs PostgreSQL Differences in IaC

| Aspect | PostgreSQL | MySQL |
|--------|-----------|-------|
| Storage | Flat `StorageMb` + `AutoGrowEnabled` | `Storage` block with `SizeGb` + `AutoGrowEnabled` |
| Database creation | `ServerId` parameter | `ServerName` + `ResourceGroupName` |
| Firewall creation | `ServerId` parameter | `ServerName` + `ResourceGroupName` |
| Public network | `PublicNetworkAccessEnabled` (bool) | `PublicNetworkAccess` (string "Enabled"/"Disabled") |
| Authentication | Explicit `Authentication` block | None (not exposed in provider) |
| Default charset | `UTF8` / `en_US.utf8` | `utf8mb4` / `utf8mb4_0900_ai_ci` |
| FQDN | `{name}.postgres.database.azure.com` | `{name}.mysql.database.azure.com` |
| Subnet delegation | `Microsoft.DBforPostgreSQL/flexibleServers` | `Microsoft.DBforMySQL/flexibleServers` |

### Files Created

- **Proto API**: 4 proto files (spec, stack_outputs, api, stack_input) + generated .pb.go + .ts stubs
- **Spec tests**: 41 validation tests covering all fields, edge cases, MySQL-specific validations
- **Pulumi module**: main.go + module/ (main.go, locals.go, outputs.go) using `mysql` package
- **Terraform module**: main.tf, variables.tf, locals.tf, outputs.tf, provider.tf
- **Documentation**: README.md, examples.md (6 YAML examples), docs/README.md (research)
- **Supporting**: hack/manifest.yaml test manifest
- **Registry**: cloud_resource_kind.proto enum 434

## Benefits

- **Database pattern reusable**: MySQL follows the PostgreSQL pattern closely, establishing a template for R13 (MSSQL)
- **Provider-authentic**: All field names, defaults, and validations match the actual Azure API
- **Infra-chart ready**: StringValueOrRef fields enable composition in database-stack chart
- **Dual IaC**: Both Pulumi and Terraform modules with feature parity

## Impact

- **Resources completed**: 13 of 24 (R00-R12)
- **Resources remaining**: 11 (R13-R23)
- **Commit**: 33 files, 4328 insertions
- **Test coverage**: 41/41 tests pass

## Related Work

- R11 AzurePostgresqlFlexibleServer (direct template)
- DD03 Composite Bundling Rules (server + databases + firewall rules)
- DD05 AzureResourceGroup as first-class resource (resource_group field)
- T02 Resource Queue (Azure resource expansion plan)

---

**Status**: Production Ready
**Timeline**: Single session
