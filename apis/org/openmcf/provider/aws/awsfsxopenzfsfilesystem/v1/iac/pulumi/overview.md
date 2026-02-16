# Pulumi Module: AWS FSx OpenZFS File System

## Architecture

This module creates a single FSx for OpenZFS file system resource with its root volume configured inline.

```
module/
├── main.go           # Entry point: provider setup, resource orchestration, output exports
├── locals.go         # Tag computation and naming
├── outputs.go        # Stack output key constants
└── file_system.go    # fsx.OpenZfsFileSystem resource creation with all field mappings
```

## Resource Flow

1. `main.go` loads stack input, initializes locals, creates the AWS provider
2. `file_system.go` maps all spec fields to `fsx.OpenZfsFileSystemArgs`, including nested blocks for disk IOPS, root volume configuration (NFS exports, quotas, compression)
3. `main.go` exports 8 stack outputs from the created resource

## Key Design Decisions

- **Single resource**: Only `fsx.OpenZfsFileSystem` is created. The root volume is configured inline via `RootVolumeConfiguration`.
- **No child volumes**: Child volumes have independent lifecycle and are not managed here. The `root_volume_id` output enables external child volume creation.
- **StringValueOrRef**: All cross-resource reference fields (subnets, security groups, KMS key) use `.GetValue()` for resolution.
- **Conditional fields**: Optional fields are only set when non-zero/non-empty, letting the AWS provider use its defaults for unset fields.
