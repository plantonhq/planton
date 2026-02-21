# Overview

The AlicloudVpc Pulumi module creates a single Alibaba Cloud VPC from an OpenMCF manifest. The module is intentionally minimal — three files, one cloud resource — because the VPC is a composable building block. VSwitches, security groups, NAT gateways, and other networking resources are managed by separate components.

## Module Architecture

```
iac/pulumi/
├── main.go              Entry point: loads stack input, calls module.Resources()
└── module/
    ├── locals.go         Transforms stack input into computed values (tag merging)
    ├── main.go           Controller: creates the Alicloud provider and the VPC resource
    └── outputs.go        Defines output constant names (vpc_id, vpc_name, etc.)
```

**Entry point** (`iac/pulumi/main.go`): Deserializes the Pulumi config into `AlicloudVpcStackInput` via `stackinput.LoadStackInput()`, then delegates to `module.Resources()`.

**Controller** (`module/main.go`): Initializes locals, creates the Alicloud provider scoped to `spec.region`, provisions the VPC via `vpc.NewNetwork`, and exports five outputs.

**Locals** (`module/locals.go`): Builds the merged tag map. System tags (`resource`, `resource_name`, `resource_kind`) are set first. Metadata fields (`resource_id`, `organization`, `environment`) are added conditionally. User-defined `spec.tags` are merged last, so user values override system tags on key conflict.

**Outputs** (`module/outputs.go`): String constants for the five export names, ensuring consistency between the Pulumi exports and `stack_outputs.proto`.

## Data Flow

```
AlicloudVpcStackInput
  │
  ├─ target.Metadata  ──► initializeLocals() ──► Locals.Tags (merged map)
  ├─ target.Spec.Tags ─┘
  │
  └─ target.Spec ──► Resources()
                        │
                        ├─ alicloud.NewProvider (region)
                        │
                        └─ vpc.NewNetwork
                              │
                              ├─► Export: vpc_id
                              ├─► Export: vpc_name
                              ├─► Export: cidr_block
                              ├─► Export: router_id
                              └─► Export: route_table_id
```

## Design Decisions

**Single resource scope**: The VPC component creates only the VPC. Alibaba Cloud auto-creates a VRouter and system route table as part of VPC creation — these are not separate Pulumi resources but are exposed as computed attributes on `vpc.Network`.

**`optionalString()` helper**: Empty proto strings are converted to `nil` before passing to the Pulumi SDK. This prevents Alibaba Cloud API errors that occur when empty strings are sent for optional fields like `description` and `resource_group_id`.

**Tag merge order**: System tags are written first, then user tags overwrite. This gives users the final say on all tag values while ensuring that every VPC gets baseline metadata tags for governance and tracking.

## Customization

| Goal | File to Change |
|------|---------------|
| Add spec fields to the VPC resource | `module/main.go` — add args to `vpc.NetworkArgs` |
| Change tag logic or add new system tags | `module/locals.go` — modify `initializeLocals()` |
| Add new stack outputs | `module/outputs.go` (constant) + `module/main.go` (export call) |
| Add a new cloud resource (e.g., secondary CIDR) | Create a new file in `module/` and call it from `Resources()` |
