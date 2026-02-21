# AliCloudEcsInstance

**Date**: 2026-02-19
**Type**: New Component
**Resource Kind**: AliCloudEcsInstance (enum 3080, id_prefix: acecs)

## Summary

Added AliCloudEcsInstance component that provisions an Alibaba Cloud Elastic Compute Service (ECS) virtual machine. Supports the full range of ECS instance families, configurable system and data disks (inline with the instance), SSH key or password authentication, public IP allocation via bandwidth setting, spot instance pricing, PrePaid subscription billing, RAM role attachment, disk encryption, and deletion protection.

## What's Included

- **Proto API**: spec.proto with 25 top-level fields + 2 nested messages (AliCloudEcsSystemDisk with 5 fields, AliCloudEcsDataDisk with 9 fields); stack_outputs.proto with 3 outputs (instance_id, private_ip, public_ip); api.proto, stack_input.proto
- **Validations**: CEL validations for instance_type prefix ("ecs."), instance_charge_type, internet_charge_type, spot_strategy, security_enhancement_strategy, period_unit, system/data disk categories, performance levels; range constraints on internet_max_bandwidth_out (0-100), data disk size (>=20), system disk size (>=20); password length bounds (8-30); instance_name (2-128) and description (2-256) length checks; repeated min_items=1 for security_group_ids, max_items=16 for data_disks
- **Tests**: spec_test.go with 5 valid-input and 13 invalid-input test cases covering all validation rules
- **Pulumi Module**: main.go (instance creation with inline data disk building), locals.go (tags, 8 helper functions), outputs.go -- uses ecs.NewInstance with ecs.InstanceDataDiskArray for data disks
- **Terraform Module**: main.tf with dynamic data_disks block, variables.tf with 7 validation blocks, outputs.tf, locals.tf, provider.tf
- **Documentation**: catalog-page.md, examples.md (4 YAML examples), README.md, docs/README.md, Pulumi overview.md and README.md, TF README.md
- **Presets**: 3 presets (basic-development, production-web-server, spot-batch-worker)
- **Registration**: Enum 3080 in cloud_resource_kind.proto, kind_map_gen.go updated

## Design Decisions

- **Inline data disks (T02 correction)**: The T02 spec listed separate alicloud_ecs_disk + alicloud_disk_attachment resources. The alicloud_instance resource's built-in data_disks block is simpler and aligns with DD07 composite bundling. Independently managed disks would be a separate component.
- **security_group_ids as repeated StringValueOrRef**: Allows cross-component references to AliCloudSecurityGroup outputs. The provider field is "security_groups" (set of IDs); the proto uses the more descriptive "security_group_ids".
- **status output dropped**: The T02 spec included a "status" output, but it's always "Running" after creation and not useful as a cross-component reference. Replaced with private_ip and public_ip which are genuinely useful.
- **System disk as nested message**: While the TF provider uses flat top-level fields (system_disk_category, system_disk_size), the proto groups them in an AliCloudEcsSystemDisk message for cleaner YAML UX.
- **Spot and billing fields included**: spot_strategy, spot_price_limit, period, period_unit go beyond the T02 minimal spec but are essential for cost optimization in real-world usage.
- **user_data and role_name added**: Not in T02 spec but critical for production -- cloud-init scripting and service-to-service authentication via RAM roles.

## Verification

- `go build ./...` -- pass
- `go vet ./...` -- pass
- `go test ./...` -- 18 test cases, all pass
- `terraform validate` -- success

## Dependencies

- AliCloudVswitch (vswitch_id)
- AliCloudSecurityGroup (security_group_ids)
- AliCloudKmsKey (optional, for system_disk.kms_key_id and data_disks.kms_key_id)
- AliCloudRamRole (optional, role_name for instance profile)
