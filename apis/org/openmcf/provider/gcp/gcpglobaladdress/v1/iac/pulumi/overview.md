# GCP Global Address Pulumi Module - Architecture Overview

## Purpose

This document explains the internal architecture of the Pulumi module for deploying GCP global addresses. It is intended for contributors, maintainers, and advanced users who want to understand how the module works under the hood.

## Design Philosophy

1. **Single Resource**: The module creates exactly one `compute.GlobalAddress` resource (backed by `google_compute_global_address`). No supporting resources are needed.
2. **Label Management**: Labels are derived from metadata (org, env, id) and resource kind, and applied to the global address for consistency with other GCP components.
3. **Conditional Fields**: Optional spec fields (address, description, network, purpose, prefix_length) are only passed to the resource when non-empty, letting GCP apply its own defaults otherwise.
4. **Address Types**: Supports both EXTERNAL (public IP) and INTERNAL (private IP or CIDR range for VPC peering / Private Service Connect).

## Module Structure

```
iac/pulumi/
├── main.go               # Entry point: loads stack input and invokes module
├── Pulumi.yaml           # Project metadata (name, runtime, description)
├── Makefile              # Build automation
├── debug.sh              # Local debugging helper
└── module/
    ├── main.go           # Module coordinator: initializes locals, gets provider, calls globalAddress()
    ├── global_address.go # Global address resource creation
    ├── locals.go         # Compute labels and extract config from stack input
    └── outputs.go        # Output key constants
```

### Separation of Concerns

- **`main.go`** (entry point): Minimal — loads `GcpGlobalAddressStackInput` from Pulumi config and delegates to `module.Resources()`.
- **`module/main.go`**: Orchestrator — calls `initializeLocals()`, obtains a GCP provider, then calls `globalAddress()`.
- **`module/locals.go`**: Pure data transformation — extracts the `GcpGlobalAddress` target and computes GCP labels from metadata (org, env, id) and resource kind.
- **`module/global_address.go`**: Infrastructure logic — builds `compute.GlobalAddressArgs`, creates the resource, exports outputs.
- **`module/outputs.go`**: Constants for output keys (`address`, `self_link`, `creation_timestamp`).

## Data Flow

```
User YAML manifest
    ↓
GcpGlobalAddressStackInput (protobuf)
    ↓
main.go → module.Resources()
    ↓
initializeLocals() → Locals struct
    ↓
pulumigoogleprovider.Get() → GCP Provider
    ↓
globalAddress() → compute.NewGlobalAddress()
    ↓
ctx.Export() → Stack Outputs
```

## Field Mapping

| Spec Field | Pulumi Argument | Notes |
|------------|-----------------|-------|
| `address_name` | `Name` | Required |
| `project_id` | `Project` | Required |
| `address_type` | `AddressType` | EXTERNAL (default) or INTERNAL |
| `ip_version` | `IpVersion` | IPV4 (default) or IPV6 |
| `address` | `Address` | Optional; omit to let GCP assign |
| `description` | `Description` | Optional |
| `network` | `Network` | Required for INTERNAL addresses |
| `purpose` | `Purpose` | VPC_PEERING or PRIVATE_SERVICE_CONNECT |
| `prefix_length` | `PrefixLength` | CIDR range for VPC peering |
| (from locals) | `Labels` | Derived from metadata and resource kind |

## Output Mapping

Three outputs are exported after the global address is created:

| Output Constant | Pulumi Attribute | Description |
|-----------------|------------------|-------------|
| `OpAddress` | `createdAddress.Address` | Reserved IP or start of CIDR range |
| `OpSelfLink` | `createdAddress.SelfLink` | Full self-link URI for referencing in other resources |
| `OpCreationTimestamp` | `createdAddress.CreationTimestamp` | RFC 3339 creation timestamp |

## Design Decisions

### Why Single Resource

A global address is a standalone GCP resource. It does not require VPCs, subnets, or other supporting resources to be created first (except for INTERNAL addresses, which require a network). Keeping the module focused on one resource simplifies maintenance and aligns with the OpenMCF component-per-resource pattern.

### Why Labels from Metadata

Labels (org, env, resource_id, resource_kind) enable cost allocation, filtering, and governance. They are computed in `locals.go` from the `GcpGlobalAddress` metadata and applied to the resource. Empty metadata fields are omitted from the label map.

### Why Conditional Optional Fields

The `global_address.go` function only sets optional Pulumi arguments when the corresponding spec field is non-empty:

```go
if spec.Address != "" {
    args.Address = pulumi.StringPtr(spec.Address)
}
if spec.Description != "" {
    args.Description = pulumi.StringPtr(spec.Description)
}
```

This avoids sending empty strings or zero values that could override GCP defaults or cause validation issues.

### Why address_type and ip_version as Strings

GCP's API uses string enums for `address_type` and `ip_version`. The spec mirrors this with defaults (EXTERNAL, IPV4) so manifests stay concise. The module passes these through directly via `GetAddressType()` and `GetIpVersion()`.

## Customization Guide

### Adding New Spec Fields

1. Add the field to the `GcpGlobalAddress` spec in the proto definition.
2. Regenerate Go stubs: `make protos`.
3. Map the new field in `global_address.go` within the `compute.GlobalAddressArgs` construction.
4. Test: `pulumi preview`.

### Adding New Outputs

1. Add the field to `stack_outputs.proto`.
2. Add a constant in `outputs.go`.
3. Add a `ctx.Export()` call in `global_address.go`.

### Changing Label Strategy

Modify `initializeLocals()` in `locals.go` to add, remove, or rename label keys. Ensure label keys comply with GCP label constraints (lowercase, numbers, hyphens, underscores; max 63 chars).

## Error Handling

Errors are wrapped with context using `github.com/pkg/errors`:

```go
return errors.Wrap(err, "failed to create global address")
```

The error chain flows as:
1. `globalAddress()` returns a wrapped error
2. `module.Resources()` wraps again: "failed to create global address"
3. Pulumi CLI displays the full chain with stack trace

## Related

- [Pulumi README](README.md) — deployment instructions
- [Terraform Module](../tf/README.md) — Terraform implementation
