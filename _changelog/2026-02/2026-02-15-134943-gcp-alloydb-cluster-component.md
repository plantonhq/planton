# GCP AlloyDB Cluster Component

**Date**: February 15, 2026
**Type**: Feature
**Components**: GCP Provider, API Definitions, Pulumi Module, Terraform Module

## Summary

Added GcpAlloydbCluster as a new deployment component, covering Google Cloud's fully managed PostgreSQL-compatible database. The component bundles a cluster with its primary instance (since a cluster without one cannot serve queries), supports three-level CMEK encryption, automated and continuous backup policies, and VPC-based private networking. This is the 11th resource in the GCP expansion project (R11 of 22).

## Problem Statement / Motivation

AlloyDB is Google Cloud's enterprise-grade PostgreSQL-compatible database, designed for demanding transactional and analytical workloads. It fills a gap between Cloud SQL (simpler managed PostgreSQL) and Cloud Spanner (globally distributed). Prior to this change, OpenMCF had no way to provision AlloyDB infrastructure, leaving users to manage it manually or through custom IaC.

### Pain Points

- No declarative way to provision AlloyDB clusters through OpenMCF
- AlloyDB has a cluster+instance model that requires careful bundling to be useful
- Three separate CMEK encryption scopes (data, automated backups, continuous backups) need clear modeling
- Backup policies have two retention strategies (quantity-based vs time-based) that are mutually exclusive

## Solution / What's New

### Proto API (4 proto files, 10 message types)

- `spec.proto` with 14 top-level fields, 8 sub-messages covering all backup, encryption, maintenance, and primary instance configuration
- 5 `StringValueOrRef` fields for infra-chart composability: `project_id`, `network`, `kms_key_name`, and 2 backup encryption keys
- CEL validations: `cpu_count`/`machine_type` mutual exclusion, retention policy mutual exclusion, duration format validation, recovery window range, enum value validation
- `stack_outputs.proto` with 6 outputs including `primary_instance_ip` for direct application connectivity

### Pulumi Module (5 Go files)

- `cluster.go`: creates `alloydb.NewCluster` with network config, backup policies (automated + continuous), CMEK, maintenance window
- `instance.go`: creates `alloydb.NewInstance` of type PRIMARY with machine config, query insights, client connection security
- Labels applied to both cluster and instance for consistent resource tracking

### Terraform Module (6 files)

- Feature parity with Pulumi: dynamic blocks for all optional configurations
- Google provider `~> 6.0` (required for AlloyDB features)
- Discovery: `deletion_protection` is not a valid argument on `google_alloydb_cluster` in TF (unlike Pulumi)

### Validation Tests (58 tests)

- 30 positive cases covering all field combinations, boundary values, and a full-featured spec
- 28 negative cases covering missing required fields, invalid patterns, mutual exclusions, range violations, and invalid enum values

### Documentation

- User-facing README with key configuration overview
- 6 YAML examples from minimal to full-featured
- Research document covering AlloyDB architecture, comparison with Cloud SQL/Spanner, and 80/20 scoping analysis
- Catalog page following the exemplar structure
- 3 presets: dev-basic, ha-production, enterprise-encrypted

## Implementation Details

### Design Decisions (Corrections to Plan)

| Decision | Rationale |
|----------|-----------|
| Added `cluster_name` field | GCP requires explicit cluster ID. Consistent with R01-R10 naming pattern. |
| `initial_user` as sub-message, not StringValueOrRef | Passwords are not cross-resource references. |
| Added `continuous_backup_config` | Critical for PITR. Not in original plan. |
| Added `query_insights_config` on primary instance | Essential for database operations monitoring. |
| Added `cpu_count` as alternative to `machine_type` | Most AlloyDB docs use cpu_count. Simpler for users. |
| Three separate CMEK fields | Authentic to GCP's encryption model. Each scope can have its own key. |
| Excluded `cluster_type` (SECONDARY) | Cross-region replication is too advanced for v1. |
| Excluded PSC networking | VPC peering covers 90%+ of deployments. PSC deferred to v2. |
| Excluded `deletion_protection` from TF | Not a valid argument on `google_alloydb_cluster`. Handled by Pulumi SDK only. |

### File Inventory

- 4 proto files + 3 generated `.pb.go` stubs
- 1 enum registration in `cloud_resource_kind.proto` (GcpAlloydbCluster = 630)
- 5 Pulumi Go files + entrypoint + Pulumi.yaml + debug.sh
- 6 Terraform files
- 1 test file (58 tests)
- 1 hack manifest
- 6 preset files (3 YAML + 3 MD)
- 7 documentation files (README, examples, research, catalog page, TF README, Pulumi README, Pulumi overview)

## Benefits

- Complete AlloyDB provisioning in a single YAML manifest
- Infra-chart composable via StringValueOrRef (project, VPC, KMS keys)
- Production-ready with REGIONAL HA, CMEK, automated + continuous backups
- 58 validation tests catch configuration errors before deployment
- Feature parity between Pulumi and Terraform implementations

## Impact

- **GCP users**: Can now provision AlloyDB clusters through OpenMCF with a single manifest
- **Infra charts**: GcpAlloydbCluster can be composed into `gcp-alloydb-environment` and `gcp-spanner-application` charts
- **GCP expansion**: 11 of 22 resources now complete (50% milestone)

## Related Work

- Part of the GCP Resource Expansion project (20260215.01.sp.gcp-resource-expansion)
- Follows R10 GcpSpannerDatabase. Next: R12 GcpBigtableInstance.
- References GcpVpc, GcpKmsKey, GcpProject via StringValueOrRef

---

**Status**: Production Ready
**Timeline**: Single session
