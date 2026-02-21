# AlicloudMongodbInstance

**Date**: 2026-02-19
**Type**: New Component
**Resource Kind**: AlicloudMongodbInstance (enum 3073, id_prefix: acmdb)

## Summary

Added AlicloudMongodbInstance component that provisions an Alibaba Cloud ApsaraDB for MongoDB replica-set instance. Supports configurable replication factors (1, 3, 5, 7 nodes), multi-zone HA across three availability zones, read-only replicas for read scaling, and both TDE and cloud disk encryption at rest. This is a single-resource component wrapping `alicloud_mongodb_instance` (replica-set mode only; sharding is a separate TF resource).

## What's Included

- **Proto API**: spec.proto with 34 fields covering instance configuration, multi-zone HA, storage engine, encryption, backup, maintenance, and billing; stack_outputs.proto with 2 outputs (instance_id, replica_set_name); api.proto, stack_input.proto
- **Validations**: CEL validations for engine_version, storage_engine, storage_type, instance_charge_type, ssl_action, tde_status, period, replication_factor, db_instance_name length; range constraints on readonly_replicas (0-5), auto_renew_duration (1-12); password length bounds (8-32)
- **Tests**: spec_test.go with 7 valid-input and 14 invalid-input test cases covering all validation rules
- **Pulumi Module**: main.go, locals.go, outputs.go -- clean mongodb.NewInstance with optional field helpers and parameter array mapping
- **Terraform Module**: main.tf, variables.tf, outputs.tf, locals.tf, provider.tf -- single alicloud_mongodb_instance resource with dynamic parameters block
- **Documentation**: catalog-page.md, examples.md, README.md, docs/README.md, Pulumi overview.md and README.md, TF README.md
- **Presets**: 3 presets (development, production-ha, encrypted-compliance)
- **Registration**: Enum 3073 in cloud_resource_kind.proto, kind_map_gen.go updated

## Design Decisions

- **engine_version is required (no default)**: Unlike Redis which defaults to "7.0", MongoDB version is a critical architectural choice -- different versions have different feature sets and wire protocols. The TF provider also marks it Required.
- **connection_string output dropped**: The TF provider does not expose a simple `connection_string` attribute. Connection details are nested in the `replica_sets` computed array (per-node domains/ports). We output `instance_id` and `replica_set_name` instead -- clean and sufficient for downstream lookups.
- **Dual encryption paths**: TDE (tde_status + encryption_key) and cloud disk encryption (encrypted + cloud_disk_encryption_key) are both supported as mutually exclusive options, matching the provider's ConflictsWith constraints.
- **Multi-zone HA**: MongoDB's 3-node architecture (primary, secondary, hidden) maps naturally to three AZ fields (zone_id, secondary_zone_id, hidden_zone_id).
- **Parameters as map**: Spec uses `map<string, string>` for MongoDB engine parameters; Pulumi maps this to `InstanceParameterArray`, Terraform uses a `dynamic` block.

## Dependencies

- AlicloudVswitch (vswitch_id)
- AlicloudKmsKey (optional, encryption_key for TDE or cloud_disk_encryption_key)
- AlicloudSecurityGroup (optional, security_group_id)
