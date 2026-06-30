# GCP Cloud Composer Environment Deployment Component

**Date**: February 15, 2026
**Type**: Feature
**Components**: API Definitions, GCP Provider, Pulumi Module, Terraform Module

## Summary

Added GcpCloudComposerEnvironment (R15) as a new deployment component for provisioning managed Apache Airflow environments on Google Cloud. The component targets Composer 2.x and 3, with full support for workload sizing, private networking (VPC peering and PSC), CMEK encryption, scheduled recovery snapshots, maintenance windows, and web server access control. All Composer 1.x fields are deliberately excluded as that version is deprecated by Google.

## Problem Statement / Motivation

Cloud Composer is the standard orchestration layer for GCP data pipelines, ML workflows, and ETL processes. Before this component, Planton users had no way to provision Composer environments declaratively -- they needed to fall back to raw Terraform or the GCP console.

### Pain Points

- No Planton support for managed Airflow environments
- Cloud Composer is one of the most complex GCP resources with deep nesting and version-conditional fields
- Composer 1.x, 2.x, and 3 have fundamentally different configuration paradigms (especially networking)

## Solution / What's New

A complete deployment component following the forge workflow (20 phases), targeting Composer 2.x and 3 exclusively.

### Key Design Decisions

- **Flattened `config` wrapper** -- The Terraform resource nests everything under `config`, but since the Planton component IS the environment, the wrapper was removed to reduce YAML nesting depth
- **Composer 2.x and 3 only** -- All Composer 1.x-only fields excluded (node_count, machine_type, disk_size_gb, python_version, database_config, web_server_config, etc.) since Composer 1.x is deprecated
- **5 StringValueOrRef fields** for infra-chart composability: project_id (GcpProject), network (GcpVpc), subnetwork (GcpSubnetwork), service_account (GcpServiceAccount), kms_key_name (GcpKmsKey)
- **Triggerer support** -- Critical for Airflow 2.x deferrable operators, not in the original plan
- **DAG processor support** -- Composer 3 feature for independent DAG parsing
- **Recovery config** -- Scheduled snapshots for disaster recovery, not in the original plan
- **Web server access control** -- IP allowlisting for Airflow UI security

## Implementation Details

### Proto API (4 files, 14 message types)

- `spec.proto`: 15 top-level fields, 13 sub-messages covering node config, software config, private environment config, workloads config, maintenance window, recovery config, and web server access control
- 5 `StringValueOrRef` fields with `default_kind` annotations
- 7 CEL/buf validations: environment_name regex, environment_size in-list, resilience_mode in-list, connection_type in-list, web_server_plugins_mode in-list, worker max_count >= min_count
- `stack_outputs.proto`: environment_id, environment_name, airflow_uri, dag_gcs_prefix, gke_cluster

### Pulumi Module (4 Go files)

- `composer_environment.go`: Conditional block creation for all 10+ optional nested configs
- Triggerer uses non-pointer types (Pulumi SDK requirement: `Count int`, `Cpu float64`, `MemoryGb float64`)
- Outputs extracted from `createdEnv.Config.AirflowUri()`, `Config.DagGcsPrefix()`, `Config.GkeCluster()`
- Framework GCP labels applied to environment

### Terraform Module (6 files)

- Provider `~> 6.0` required for Composer 3 fields
- Dynamic blocks for all optional nested configs
- Nested dynamic blocks for workloads_config (5 workload types) and web_server_network_access_control
- Feature parity with Pulumi implementation

### Validation Tests (43 passing)

- 25 positive cases: all field values, all enum variants, worker boundary testing, full-featured spec
- 18 negative cases: missing required fields, invalid names, invalid enum values, worker max < min, maintenance window missing fields, access control missing CIDR

### Documentation

- User-facing README, 5 YAML examples, comprehensive research docs (600+ lines)
- Catalog page following exemplar structure
- Pulumi overview and Terraform README

### Presets (3)

- `01-dev-small`: Minimal development environment
- `02-production-private`: Medium with VPC peering, private endpoint, HA, maintenance window
- `03-enterprise-encrypted`: Large with CMEK, recovery snapshots, IP access control

## Benefits

- Planton users can now provision managed Airflow environments declaratively
- Full Composer 2.x and 3 coverage without Composer 1.x legacy complexity
- 5 StringValueOrRef fields enable composing Composer environments with VPCs, KMS keys, and service accounts in infra charts
- Production-ready configuration with private networking, CMEK, and disaster recovery out of the box

## Impact

- **New GCP resource kind**: GcpCloudComposerEnvironment (enum 680, id_prefix: gcpcce)
- **Files created**: ~50 files across proto, Go, Terraform, documentation, and presets
- **Test coverage**: 43 validation tests covering all proto validations

## Related Work

- Part of the GCP Resource Expansion project (20260215.01.sp.gcp-resource-expansion)
- 16th resource forged in a series of 23 planned GCP resource kinds
- Builds on patterns established in R01-R14b (GcpFirewallRule through GcpDataprocVirtualCluster)
- Dependencies: GcpVpc, GcpSubnetwork, GcpServiceAccount, GcpKmsKey

---

**Status**: Production Ready
**Timeline**: Single session
