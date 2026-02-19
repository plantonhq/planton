# AlicloudPolardbCluster

**Date**: 2026-02-19
**Type**: New Component
**Resource Kind**: AlicloudPolardbCluster (enum 3071, id_prefix: acpdb)

## Summary

Added AlicloudPolardbCluster component that provisions an Alibaba Cloud PolarDB cluster with bundled databases, accounts, and account privileges. PolarDB is Alibaba Cloud's cloud-native relational database with a shared-storage, compute-storage-separated architecture. This is a first-class component separate from RDS (per DD06) supporting MySQL, PostgreSQL, and Oracle compatibility modes.

## What's Included

- **Proto API**: spec.proto with 29 fields and 4 nested message types (AlicloudPolardbParameter, AlicloudPolardbDatabase, AlicloudPolardbAccount, AlicloudPolardbAccountPrivilege), stack_outputs.proto with 4 outputs, api.proto, stack_input.proto
- **Validations**: CEL validations for db_type, pay_type, creation_category, sub_category, storage_type, tde_status, collector_status, backup_retention_policy, deletion_lock, account_type, account_privilege; length and range constraints on description, db_node_count, storage_space, passwords, periods
- **Tests**: spec_test.go with 10 valid-input and 18 invalid-input test cases (28 total)
- **Pulumi Module**: main.go, locals.go, outputs.go, databases.go, accounts.go -- engine-specific character set defaults, parent-child resource hierarchy, PolarDB-specific fields
- **Terraform Module**: main.tf, databases.tf, accounts.tf, variables.tf, outputs.tf, locals.tf, provider.tf -- for_each iteration with flattened privilege maps, collate/ctype support for PostgreSQL/Oracle
- **Documentation**: catalog-page.md, examples.md, README.md, docs/README.md, Pulumi overview.md and README.md, TF README.md
- **Presets**: 3 presets (mysql-dev, mysql-production, postgresql-production)
- **Registration**: Enum 3071 in cloud_resource_kind.proto

## Design Decisions

- **First-class component (DD06)**: PolarDB is separate from RDS due to fundamentally different architecture (cluster-based, compute-storage separated, different TF/Pulumi resources)
- **Composite bundling (DD07)**: cluster + databases + accounts + privileges bundled as a single deployable unit
- **db_node_class + db_node_count**: Replaces RDS instance_type and category; node-based sizing is PolarDB's native model
- **Edition support**: creation_category controls Enterprise (Normal), Standard (SENormal), and Basic editions
- **Storage flexibility**: storage_type + storage_space for Standard Edition; Enterprise Edition auto-scales
- **PostgreSQL/Oracle collation**: collate and ctype fields on databases for locale-aware character handling
- **No custom endpoints**: Primary and cluster endpoints are auto-created; custom endpoints excluded (can be managed separately)
- **PolarDB privilege values**: ReadOnly, ReadWrite, DDLOnly, DMLOnly (no DBOwner -- PolarDB does not support it)
- **Default charset**: utf8 for MySQL (PolarDB default), UTF8 for PostgreSQL/Oracle

## Dependencies

- AlicloudVswitch (vswitch_id)
- AlicloudKmsKey (optional, encryption_key for TDE)
- AlicloudSecurityGroup (optional, security_group_ids)
