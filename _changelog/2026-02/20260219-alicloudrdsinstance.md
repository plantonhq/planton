# AlicloudRdsInstance

**Date**: 2026-02-19
**Type**: New Component
**Resource Kind**: AlicloudRdsInstance (enum 3070, id_prefix: acrds)

## Summary

Added AlicloudRdsInstance component that provisions an Alibaba Cloud RDS (Relational Database Service) instance with bundled databases, accounts, and account privileges. This is a composite component (per DD07) supporting all RDS engines (MySQL, PostgreSQL, SQL Server, MariaDB, PPAS) through a single component type (per DD02).

## What's Included

- **Proto API**: spec.proto with 28 fields and 4 nested message types (AlicloudRdsParameter, AlicloudRdsDatabase, AlicloudRdsAccount, AlicloudRdsAccountPrivilege), stack_outputs.proto with 4 outputs, api.proto, stack_input.proto
- **Validations**: CEL validations for engine, instance_charge_type, category, db_instance_storage_type, monitoring_period, ssl_action, tde_status, account_type, privilege; length and range constraints on names, passwords, periods
- **Tests**: spec_test.go with 6 valid-input and 12 invalid-input test cases
- **Pulumi Module**: main.go, locals.go, outputs.go, databases.go, accounts.go -- engine-specific character set defaults, parent-child resource hierarchy
- **Terraform Module**: main.tf, databases.tf, accounts.tf, variables.tf, outputs.tf, locals.tf, provider.tf -- for_each iteration with flattened privilege maps
- **Documentation**: catalog-page.md, examples.md, README.md, docs/README.md, Pulumi overview.md
- **Presets**: 3 presets (mysql-basic, postgresql-ha, mysql-production)
- **Registration**: Enum 3070 in cloud_resource_kind.proto

## Design Decisions

- **Single component for all engines (DD02)**: `engine` field selects MySQL/PostgreSQL/SQLServer/MariaDB/PPAS
- **`category` replaces `bool high_availability`**: Provider-authentic field with richer semantics (Basic, HighAvailability, AlwaysOn, Finance, cluster)
- **Structured account privileges**: `AlicloudRdsAccountPrivilege` with explicit database_names + privilege instead of unstructured string
- **No public connection endpoint**: `alicloud_db_connection` intentionally excluded as a security best practice; intranet connection string exported
- **Engine-specific charset defaults**: Automatic character set selection based on engine (utf8mb4 for MySQL, UTF8 for PostgreSQL, etc.)

## Dependencies

- AlicloudVswitch (vswitch_id)
- AlicloudKmsKey (optional, encryption_key for disk/TDE encryption)
- AlicloudSecurityGroup (optional, security_group_ids)
