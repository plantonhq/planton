# AliCloudNasFileSystem Pulumi Module Overview

## Module Structure

```
module/
├── main.go            # Provider setup, file system creation, mount target creation
├── locals.go          # Locals struct, tag initialization, helper functions
├── outputs.go         # Output constant names
└── access_group.go    # Access group and access rule creation functions
```

## Resource Flow

1. **Provider** -- `alicloud.NewProvider` with the spec's region
2. **File System** -- `nas.NewFileSystem` with protocol, storage type, optional encryption, and tags. For extreme NAS, VPC/VSwitch/ZoneId are set on the file system itself.
3. **Access Group** -- (conditional) `nas.NewAccessGroup` created only when `spec.AccessRules` is non-empty, with type "Vpc" and file_system_type matching the file system
4. **Access Rules** -- (conditional) `nas.NewAccessRule` for each entry, parented to the access group
5. **Mount Target** -- `nas.NewMountTarget` in the specified VPC/VSwitch, referencing the custom access group or defaulting to the built-in DEFAULT_VPC_GROUP_NAME

## Key Design Decisions

- **Conditional access group**: The access group and rules are only created when the user specifies `accessRules`. This keeps the minimal case simple (zero configuration needed for VPC-wide access).
- **Parent relationships**: Mount target and access rules are parented to the file system and access group respectively for clean Pulumi state management.
- **File system type propagation**: The `fileSystemType` value flows through to the access group and access rules, which require it as a parameter.
- **Optional field handling**: Helper functions (`fileSystemType()`, `optionalString()`) provide defaults when proto optional fields are nil.
