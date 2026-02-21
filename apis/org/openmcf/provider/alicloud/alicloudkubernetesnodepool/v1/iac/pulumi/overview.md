# Overview

This Pulumi module deploys an ACK Kubernetes node pool from a single YAML manifest conforming to OpenMCF's resource model. The module provisions a single `cs.NodePool` resource with support for multiple instance types, ESSD disk configuration, auto-scaling, managed lifecycle (auto-repair, auto-upgrade), spot instances, Kubernetes labels/taints, and billing configuration.

---

## Module Architecture

```
iac/pulumi/
├── main.go           # Entrypoint: loads stack input, calls module.Resources
└── module/
    ├── main.go       # Controller: creates provider, builds NodePoolArgs, provisions pool
    ├── locals.go     # Locals: foreign key resolution, tag merging, default helpers
    └── outputs.go    # Output constants: node_pool_id, scaling_group_id
```

**Control flow**: `main.go` deserializes the protobuf stack input → `module.Resources` initializes locals (resolving StringValueOrRef fields for cluster ID, VSwitch IDs, security group IDs) → builds the `NodePoolArgs` struct → conditionally sets optional fields → creates the `cs.NodePool` resource → exports outputs.

## Key Components

**Controller (`module/main.go`)**: The `Resources` function orchestrates node pool creation. It handles complex nested configurations: data disks are converted via `dataDisks()`, labels via `nodeLabels()`, taints via `nodeTaints()`, scaling config via `scalingConfig()`, management via `managementConfig()`, and spot price limits via `spotPriceLimits()`. Each optional field is set only when the user provides a value.

**Locals (`module/locals.go`)**: Resolves `StringValueOrRef` fields (cluster ID, VSwitch IDs, security group IDs) into plain strings. Provides default helpers: `imageType` (AliyunLinux3), `systemDiskCategory` (cloud_essd), `systemDiskSize` (120), `instanceChargeType` (PostPaid), `installCloudMonitor` (true). Merges resource metadata tags with user tags.

**Outputs (`module/outputs.go`)**: Two outputs: `node_pool_id` (ACK node pool ID) and `scaling_group_id` (Auto Scaling group ID).

## Design Decisions

- **Single resource**: Wraps exactly one `cs.NodePool` resource.
- **Foreign key resolution in locals**: Cluster ID, VSwitch IDs, and security group IDs are resolved from `StringValueOrRef` in `initializeLocals`, keeping the controller clean.
- **desired_size as string**: The Pulumi provider expects `desired_size` as a string despite it being logically an integer. The module formats it with `fmt.Sprintf("%d", ...)`.
- **Conditional field setting**: Optional fields are only set when non-zero/non-nil to avoid overriding provider defaults.

## Customization Guide

| Goal | File to Modify | What to Change |
|------|---------------|----------------|
| Add a new output | `module/outputs.go` + `module/main.go` | Add a constant and `ctx.Export` call |
| Change a default value | `module/locals.go` | Update the relevant helper function |
| Add a new spec field | `module/main.go` | Add conditional field mapping in `Resources` |

---

## Next Steps

- Refer to the [README.md](./README.md) for CLI flows and debugging instructions.
- Review the [examples.md](./examples.md) for runnable manifests.
