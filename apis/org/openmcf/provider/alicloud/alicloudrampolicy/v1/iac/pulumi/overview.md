# Pulumi Module Overview

## Module Architecture

The AlicloudRamPolicy Pulumi module is organized into three files under `iac/pulumi/module/`:

| File | Responsibility |
|------|---------------|
| `main.go` | Controller — creates the provider and RAM policy, exports outputs |
| `locals.go` | Transformations — tag computation, default resolution for optional fields |
| `outputs.go` | Constants — defines output key names exported to the stack |

The entry point binary at `iac/pulumi/main.go` loads the stack input (manifest + credentials) and delegates to `module.Resources()`.

## Control Flow

```
LoadStackInput (manifest YAML → AlicloudRamPolicyStackInput)
    ↓
initializeLocals() → Locals{Tags, AlicloudRamPolicy}
    ↓
alicloud.NewProvider (region-scoped)
    ↓
ram.NewPolicy (name, document, description, rotateStrategy, force, tags)
    ↓
ctx.Export (policy_name, policy_type)
```

## Key Components

### Controller (`main.go`)

The `Resources` function orchestrates the entire deployment:

1. Initializes locals from the stack input (tag computation, reference to spec)
2. Creates a region-scoped Alibaba Cloud provider
3. Creates a single `ram.NewPolicy` resource with all spec fields
4. Exports two outputs: `policy_name`, `policy_type`

Two helper functions handle optional proto fields:

- `optionalString(s string)` — returns `nil` for empty strings, allowing the Pulumi SDK to skip optional API parameters rather than sending empty values
- `optionalStringPtr(s *string)` — unwraps optional proto pointer fields (`*string` from proto3 `optional`), returning `nil` if the pointer is nil

A third helper resolves the `force` default:

- `forceDelete(spec)` — returns `*spec.Force` if set, otherwise `false`

### Locals (`locals.go`)

The `initializeLocals` function computes:

- **Tags**: Merges standard OpenMCF tags (`resource`, `resource_name`, `resource_kind`, `resource_id`, `organization`, `environment`) with user-provided `spec.Tags`. User tags take precedence on key conflict.
- **Reference**: Stores the full `AlicloudRamPolicy` resource for convenient access throughout the module.

### Outputs (`outputs.go`)

Defines two output key constants:

- `OpPolicyName = "policy_name"` — the policy name as created
- `OpPolicyType = "policy_type"` — always `"Custom"` for user-created policies

These constants are referenced by `ctx.Export()` in the controller and must match the field names in `stack_outputs.proto`.

## Resource Relationships

```
Provider (alicloud, region from spec.region)
  └── ram.Policy (spec.policyName)
```

This is a single-resource module — no sub-resources, no iteration, no parent chaining. The `ram.Policy` resource maps directly to the `alicloud_ram_policy` cloud resource.

## Design Decisions

**Single Resource**: Unlike AlicloudRamRole (which bundles a role + N policy attachments), this module creates exactly one resource. The simplicity is intentional — a policy is a standalone artifact that is referenced by name from other components.

**Tag Merging**: Standard OpenMCF tags are always applied. User-provided `spec.Tags` are merged in, with user tags overriding standard tags on key conflict. This ensures every policy is discoverable via standard tag queries while allowing custom metadata.

**Optional Field Handling**: Proto3 `optional` fields are pointer types in Go (`*string`, `*bool`). The `optionalString`, `optionalStringPtr`, and `forceDelete` helpers dereference pointers safely, returning the documented default when the pointer is nil. This pattern avoids sending empty or zero-value arguments to the Alibaba Cloud API.

**Region as Provider Config**: RAM is a global Alibaba Cloud service, but the provider SDK requires a region for API endpoint routing. The `region` field configures the provider, not the policy's scope.

## Customization Guide

| Customization | File to Modify | Notes |
|--------------|---------------|-------|
| Add a new spec field | `locals.go` (default resolver) + `main.go` (pass to resource) | Add a helper if the field is optional with a default |
| Change tag logic | `locals.go` (`initializeLocals`) | Modify the tag map construction |
| Add a new output | `outputs.go` (constant) + `main.go` (`ctx.Export`) | Must also update `stack_outputs.proto` |
| Change force-delete default | `locals.go` (`forceDelete`) | Currently returns `false` when unset |
