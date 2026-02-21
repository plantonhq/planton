# AliCloudEcsInstance Pulumi Module Overview

## Module Structure

```
module/
├── main.go            # Provider setup, instance creation, data disk building, exports
├── locals.go          # Locals struct, tag initialization, default helpers
└── outputs.go         # Output constant names
```

## Resource Flow

1. **Provider** -- `alicloud.NewProvider` with the spec's region
2. **ECS Instance** -- `ecs.NewInstance` with instance type, image, networking, disk config, and optional spot/billing settings

## Key Design Decisions

- **Inline data disks**: Data disks use the instance resource's built-in `data_disks` block rather than separate `ecs.Disk` + `ecs.DiskAttachment` resources. This keeps disk lifecycle tied to the instance (DD07 composite bundling).
- **Security groups as array**: Each `StringValueOrRef` in `security_group_ids` is resolved to a string and passed as a `pulumi.StringArray` to the `SecurityGroups` argument.
- **System disk as flat fields**: The Pulumi SDK models system disk properties as top-level `SystemDisk*` fields on `ecs.InstanceArgs`, not as a nested struct.
- **Optional field handling**: Helper functions (`instanceChargeType()`, `systemDiskCategory()`, `systemDiskSize()`, `dataDiskCategory()`, `dataDiskDeleteWithInstance()`, `optionalString()`, `optionalBool()`, `optionalInt()`, `optionalStringPtr()`, `optionalFloat64Ptr()`) provide defaults when proto optional fields are nil.
