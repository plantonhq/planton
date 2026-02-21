# AliCloudVswitch Pulumi Module - Architecture Overview

## Purpose

This document explains the internal architecture of the Pulumi module for deploying Alibaba Cloud VSwitches. It is intended for contributors and maintainers who want to understand how the module works.

## Design Philosophy

The Pulumi module follows OpenMCF's core principles:

1. **Declarative API**: Users specify VSwitch configuration via protobuf messages
2. **Foreign Key Resolution**: VPC ID is resolved from `StringValueOrRef` before IaC execution
3. **Idempotent Operations**: Safe to run multiple times -- Pulumi handles state reconciliation
4. **Minimal Boilerplate**: Simple Go code that platform engineers can understand immediately

## Module Structure

```
iac/pulumi/
├── main.go               # Entry point: loads stack input and invokes module
├── Pulumi.yaml           # Project metadata (name, runtime)
├── Makefile              # Build automation
└── module/
    ├── main.go           # VSwitch creation, provider setup, output exports
    ├── locals.go         # VPC ID resolution, tag computation
    └── outputs.go        # Output key constants
```

### Separation of Concerns

- **`main.go`** (entrypoint): Minimal -- loads input and delegates to module
- **`module/main.go`**: Creates the provider and the VSwitch resource
- **`module/locals.go`**: Pure data transformation (resolves VPC ID, builds tags)
- **`module/outputs.go`**: Constants for output keys (avoids magic strings)

## Data Flow

```
User YAML manifest
    |
    v
AliCloudVswitchStackInput (protobuf)
    |
    v
main.go (entrypoint)
    |
    v
module/locals.go: initializeLocals()
    - Resolves vpc_id from StringValueOrRef via GetValue()
    - Builds tag map from metadata + spec.tags
    |
    v
module/main.go: Resources()
    - Creates alicloud.NewProvider() with spec.Region
    - Creates vpc.NewSwitch() with resolved VPC ID, zone, CIDR, name
    - Exports stack outputs
    |
    v
Pulumi Stack Outputs (vswitch_id, vswitch_name, cidr_block, zone_id, ipv6_cidr_block)
```

## Key Implementation Details

### StringValueOrRef Resolution

```go
locals.VpcId = target.Spec.VpcId.GetValue()
```

The `vpc_id` field is the first `StringValueOrRef` in the Alibaba Cloud provider. The platform resolves `value_from` references before IaC modules execute, so `GetValue()` always returns a literal string at runtime.

### IPv6 Handling

IPv6 CIDR block mask is only set when the user provides a non-zero value:

```go
if spec.Ipv6CidrBlockMask != 0 {
    switchArgs.Ipv6CidrBlockMask = pulumi.Int(int(spec.Ipv6CidrBlockMask))
}
```

The parent VPC must have IPv6 enabled for the VSwitch IPv6 allocation to succeed.

### Tag Management

Tags follow the same pattern as AliCloudVpc:
- Base tags from metadata: `resource`, `resource_name`, `resource_kind`
- Optional metadata tags: `resource_id`, `organization`, `environment`
- User-provided tags merged last (can override base tags)

### Immutable Fields

Three VSwitch fields are ForceNew in the provider:
- `vpc_id` -- cannot move between VPCs
- `zone_id` -- cannot move between AZs
- `cidr_block` -- cannot resize

Changing any of these causes Pulumi to destroy and recreate the VSwitch.
