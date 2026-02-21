# AliCloudRedisInstance

**Date**: 2026-02-19
**Type**: New Component
**Resource Kind**: AliCloudRedisInstance (enum 3072, id_prefix: acred)

## Summary

Added AliCloudRedisInstance component that provisions an Alibaba Cloud Redis (KVStore) instance for managed in-memory caching, session management, and real-time data processing. Supports both Redis and Memcache engines, with Redis as the default. This is a single-resource component (not composite) -- unlike RDS, accounts are not bundled because the instance-level password covers the majority of authentication use cases.

## What's Included

- **Proto API**: spec.proto with 30 fields covering instance configuration, networking, security, encryption, backup, and billing; stack_outputs.proto with 4 outputs, api.proto, stack_input.proto
- **Validations**: CEL validations for engine_version, instance_type, payment_type, ssl_enable, tde_status, vpc_auth_mode, period, db_instance_name length; range constraints on read_only_count, auto_renew_period, shard_count; password length bounds (8-32)
- **Tests**: spec_test.go with 5 valid-input and 12 invalid-input test cases covering all validation rules
- **Pulumi Module**: main.go, locals.go, outputs.go -- clean kvstore.NewInstance with optional field helpers
- **Terraform Module**: main.tf, variables.tf, outputs.tf, locals.tf, provider.tf -- single alicloud_kvstore_instance resource
- **Documentation**: catalog-page.md, examples.md, README.md, docs/README.md, Pulumi overview.md and README.md, TF README.md
- **Presets**: 3 presets (standard-single, ha-cluster, production-encrypted)
- **Registration**: Enum 3072 in cloud_resource_kind.proto, kind_map_gen.go updated

## Design Decisions

- **No account bundling**: Redis uses instance-level `password` for authentication (80% use case). `kvstore.Account` (Redis 6.0+ ACL) is niche and can be a separate component if needed.
- **Modern field names**: Uses `db_instance_name` (not deprecated `instance_name`), `payment_type` (not deprecated `instance_charge_type`), `config` (not deprecated `parameters`)
- **backup_period as string set**: Corrected from T02 guidance which had `int32 backup_period`; actual provider uses day-name strings ("Monday", "Wednesday")
- **period as string**: Follows provider's native type (string with string-based validation) to avoid unnecessary type conversion
- **No public endpoint**: Public connection management (`kvstore.Connection`) intentionally excluded; intranet `connection_domain` exported

## Dependencies

- AliCloudVswitch (vswitch_id)
- AliCloudKmsKey (optional, encryption_key for TDE encryption)
- AliCloudSecurityGroup (optional, security_group_id)
