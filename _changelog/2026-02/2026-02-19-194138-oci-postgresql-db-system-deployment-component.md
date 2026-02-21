# OCI PostgreSQL DB System Deployment Component

**Date**: February 19, 2026
**Type**: Feature
**Components**: API Definitions, Pulumi CLI Integration, Terraform Module, Protobuf Schemas

## Summary

Implemented the OciPostgresqlDbSystem deployment component (CloudResourceKind 3333) -- OCI's fully managed PostgreSQL service with configurable compute shapes, flexible OCPU/memory sizing, regional or AD-local storage durability, IOPS performance tiers, and discriminated backup policies (daily/weekly/monthly/none). This is the fourth database resource in the OCI provider (following OciAutonomousDatabase, OciDbSystem, and OciMysqlDbSystem).

## Problem Statement / Motivation

OCI PostgreSQL is a newer managed PostgreSQL offering that runs on dedicated compute shapes with built-in read replica support, configurable storage durability, and automatic backup scheduling. Unlike MySQL HeatWave, it has a distinct API surface with nested credentials (plain text or Vault secret), required storage details with regional durability choice, and a kind-discriminated backup policy.

### Pain Points

- Platform teams need managed PostgreSQL with configurable read replicas (instance_count > 1) for horizontal read scaling
- Storage durability must be configurable: regional (multi-AD replication) vs AD-local (single AD with higher performance)
- IOPS performance tiers are critical for latency-sensitive workloads
- Credential management must support both plain-text passwords (development) and OCI Vault secrets (production)
- Backup scheduling needs fine-grained control: daily, weekly (specific days), monthly (specific dates), or disabled

## Solution / What's New

A complete deployment component with:

1. **Proto API** -- `spec.proto` with 14 fields, 7 nested messages, 2 enums (PasswordType, BackupKind), 2 CEL conditional validation rules
2. **Validation Tests** -- 34 Ginkgo/Gomega tests (21 valid, 13 invalid scenarios), all passing
3. **Pulumi Module** -- `psql.NewDbSystem()` with 6 builder functions (buildNetworkDetails, buildStorageDetails, buildCredentials, buildManagementPolicy, buildBackupPolicy, buildInstancesDetails) across 4 Go files. Primary endpoint IP extraction via ApplyT on NetworkDetails output.
4. **Terraform Module** -- `oci_psql_db_system.this` with nested blocks for network_details, storage_details (hardcoded system_type), credentials > password_details, management_policy > backup_policy, instances_details. 2 enum maps (password_type, backup_kind). `lifecycle { ignore_changes = [credentials] }` for password drift prevention.
5. **Kind Registration** -- OciPostgresqlDbSystem=3333 under "Databases" section, kind_map_gen.go regenerated

### Design Decision: Nested Credentials (Not Flat)

PostgreSQL's API natively uses a discriminated `credentials` block with `username` + `password_details` (which itself discriminates between PLAIN_TEXT and VAULT_SECRET). Rather than flattening this to match MySQL's admin_username/admin_password pattern, the spec faithfully models PostgreSQL's richer credential structure. This preserves native Vault integration support and type safety through the PasswordType enum.

### Design Decision: storage_details.system_type Hardcoded

The `storage_details.system_type` field currently only supports one value: `"OCI_OPTIMIZED_STORAGE"`. Rather than exposing a meaningless single-option field in the spec, it is hardcoded in both IaC modules. If OCI adds more storage types, the field will be added to the spec.

## Implementation Details

### Fresh Creation Only (No source block)

The `source` block supports BACKUP restore and NONE. For v1, only fresh creation is supported -- consistent with OciMysqlDbSystem and OciDbSystem. Restore from backup represents a different operational workflow.

### Backup Policy Kind Discriminator

Unlike MySQL's simple enabled/disabled backup toggle, PostgreSQL supports four backup kinds: DAILY, WEEKLY (with days_of_the_week), MONTHLY (with days_of_the_month up to 28), and NONE. Each kind activates different fields, modeled cleanly through the BackupKind enum.

### IOPS as int64 in Proto, String in Provider

The TF provider stores IOPS as a string (validated as int64). The spec uses `int64` for type safety, and the IaC modules handle the conversion (fmt.Sprintf in Pulumi, tostring() in Terraform).

### CEL Validation Rules

Two conditional rules enforce credential consistency:
1. Plain-text credentials require a non-empty `password` when `password_type` is `plain_text`
2. Vault secret credentials require `secret_id` when `password_type` is `vault_secret`

### Credentials Lifecycle Handling

The credentials block is ForceNew (immutable after creation) and the password is not returned by the API. The Terraform module uses `lifecycle { ignore_changes = [credentials] }` to prevent plan diffs on subsequent applies.

### Excluded from v1

- `source` block (BACKUP restore)
- `patch_operations` (replica management -- operational concern)
- `apply_config` (RESTART/RELOAD on config change -- operational concern)
- `backup_policy.copy_policy` (cross-region backup copy -- advanced enterprise feature)
- `system_type` top-level (computed, no documented user-facing values)

## Benefits

- **Managed PostgreSQL** -- Fully managed PostgreSQL with automatic patching and maintenance
- **Read scaling** -- Instance count > 1 creates read replicas with optional reader endpoint
- **Storage durability choice** -- Regional (multi-AD) for HA or AD-local for performance
- **IOPS performance tiers** -- Configurable guaranteed IOPS for latency-sensitive workloads
- **Flexible credentials** -- Plain text for development, OCI Vault for production
- **Granular backups** -- Daily, weekly, monthly, or disabled with configurable retention
- **Feature parity** -- Identical capabilities in both Pulumi and Terraform modules

## Impact

- Fourth of six Phase 4 (Databases) resources completed
- Next component: R19 OciRedisCluster
- 18 of 37 OCI resource kinds now implemented

## Related Work

- R15 OciAutonomousDatabase (2026-02-19) -- serverless Oracle database
- R16 OciDbSystem (2026-02-19) -- traditional Oracle on VM/BM
- R17 OciMysqlDbSystem (2026-02-19) -- managed MySQL HeatWave
- R19 OciRedisCluster (next) -- managed Redis cache

---

**Status**: Production Ready
**Validation**: go build clean, go vet clean, 34/34 tests passed, terraform validate success
