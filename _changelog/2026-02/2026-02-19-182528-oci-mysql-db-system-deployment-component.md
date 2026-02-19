# OCI MySQL DB System Deployment Component

**Date**: February 19, 2026
**Type**: Feature
**Components**: API Definitions, Pulumi CLI Integration, Terraform Module, Protobuf Schemas

## Summary

Implemented the OciMysqlDbSystem deployment component (CloudResourceKind 3332) -- OCI's MySQL HeatWave managed database service with integrated in-memory analytics acceleration. This is the third database resource in the OCI provider (following OciAutonomousDatabase and OciDbSystem). The component manages a single resource with 10 configurable nested blocks covering backup, storage, maintenance, encryption, TLS, read endpoints, database console, and REST API.

## Problem Statement / Motivation

OCI MySQL HeatWave is a fully managed MySQL database that uniquely combines OLTP processing with in-memory analytics through HeatWave acceleration. Unlike self-managed MySQL, HeatWave DB Systems run on dedicated compute shapes with configurable high availability (3-instance automatic failover), PITR, and OCI-native backup policies.

### Pain Points

- Platform teams need managed MySQL with enterprise HA (3 fault domain failover) and automated backup policies
- Data-at-rest encryption must support both Oracle-managed and customer-managed (BYOK) keys
- Maintenance windows, version track preferences, and upgrade scheduling are critical for production operations
- Read scaling via dedicated read endpoints reduces load on the primary without application changes
- Modern MySQL features (REST API, database console, PITR) need first-class configuration support

## Solution / What's New

A complete deployment component with:

1. **Proto API** -- `spec.proto` with 29 fields, 10 nested messages, 6 enums, 2 CEL conditional validation rules
2. **Validation Tests** -- 39 Ginkgo/Gomega tests (26 valid, 13 invalid scenarios), all passing
3. **Pulumi Module** -- `mysql.NewMysqlDbSystem()` with 9 builder functions (buildDataStorage, buildBackupPolicy, buildMaintenance, buildDeletionPolicy, buildEncryptData, buildSecureConnections, buildReadEndpoint, buildDatabaseConsole, buildRest) across 4 Go files. Endpoint extraction via ApplyT.
4. **Terraform Module** -- `oci_mysql_mysql_db_system.this` with 10 dynamic blocks (data_storage, backup_policy with nested pitr_policy, maintenance, deletion_policy, encrypt_data, secure_connections, customer_contacts, read_endpoint, database_console, rest), 6 enum maps, `lifecycle { ignore_changes }` for admin_password
5. **Kind Registration** -- OciMysqlDbSystem=3332 under "Databases" section, kind_map_gen.go regenerated

### Design Decision: Single Resource, Not a Bundle

The original plan stub described bundling channels and HeatWave cluster with the DB System. Provider analysis revealed that both are separate Terraform/Pulumi resources (`oci_mysql_heat_wave_cluster`, `oci_mysql_channel`) with independent lifecycles. On the DB System itself, they appear only as computed read-only outputs. The component was implemented as a single-resource component, consistent with OciAutonomousDatabase and the DD03 precedent (independently-managed resources stay as separate components).

## Implementation Details

### Fresh Creation Only (source_type=NONE)

The `source` block supports 4 modes (NONE, BACKUP, PITR, IMPORTURL). For v1, only fresh creation is supported -- consistent with OciDbSystem's approach. Clone/restore require different field sets and represent different operational workflows.

### Data Storage Block (Not Deprecated Top-Level)

The top-level `data_storage_size_in_gb` is deprecated by Oracle in favor of the `data_storage` nested block which adds auto-expansion support. The spec uses the modern `DataStorage` message exclusively.

### CEL Validation Rules

Two conditional rules enforce key-material consistency:
1. BYOK encryption requires `key_id` when `key_generation_type` is `byok`
2. BYOC secure connections requires `certificate_id` when `certificate_generation_type` is `byoc`

### Secure Connections Enum Handling

The `CertificateGenerationType` enum uses `system_cert` (not `system`) to avoid collision with the `KeyGenerationType.system` enum value in the same message scope. The Pulumi module maps `system_cert` back to the API value `"SYSTEM"`.

### Admin Password Sensitivity

Admin password is not returned by the OCI API after creation. The Terraform module uses `lifecycle { ignore_changes = [admin_password] }` to prevent plan diffs on subsequent applies.

### Excluded from v1

- `source` block (clone/restore scenarios)
- `shutdown_type`, `state` (operational lifecycle controls)
- `access_mode`, `database_mode` (runtime toggles)
- `backup_policy.copy_policies` (cross-region backup copy)
- `maintenance.maintenance_disabled_windows` (scheduling exceptions)

## Benefits

- **Production-ready MySQL** -- Managed MySQL with 3-instance HA, PITR, automated backups
- **Modern storage management** -- Auto-expanding storage with configurable limits
- **Enterprise encryption** -- BYOK and BYOC support for keys and TLS certificates
- **Read scaling** -- Dedicated read endpoint for distributing read queries
- **Flexible maintenance** -- Version track preferences (LTS/Innovation), schedule types (Early/Regular)
- **Feature parity** -- Identical capabilities in both Pulumi and Terraform modules

## Impact

- Third of six Phase 4 (Databases) resources completed
- Next component: R18 OciPostgresqlDbSystem
- 17 of 37 OCI resource kinds now implemented

## Related Work

- R15 OciAutonomousDatabase (2026-02-19) -- first database component (serverless Oracle)
- R16 OciDbSystem (2026-02-19) -- second database component (traditional Oracle on VM/BM)
- R18 OciPostgresqlDbSystem (next) -- managed PostgreSQL on OCI

---

**Status**: Production Ready
**Validation**: go build clean, go vet clean, 39/39 tests passed, terraform validate success
