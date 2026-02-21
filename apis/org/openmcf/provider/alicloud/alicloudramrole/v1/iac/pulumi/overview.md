# Pulumi Module Overview

## Module Architecture

The AlicloudRamRole Pulumi module is organized into three files under `iac/pulumi/module/`:

| File | Responsibility |
|------|---------------|
| `main.go` | Controller — creates the provider, RAM role, and policy attachments |
| `locals.go` | Transformations — tag computation, default resolution for optional fields |
| `outputs.go` | Constants — defines output key names exported to the stack |

The entry point binary at `iac/pulumi/main.go` loads the stack input (manifest + credentials) and delegates to `module.Resources()`.

## Control Flow

```
LoadStackInput (manifest YAML → AlicloudRamRoleStackInput)
    ↓
initializeLocals() → Locals{Tags, AlicloudRamRole}
    ↓
alicloud.NewProvider (region-scoped)
    ↓
ram.NewRole (trust policy, session duration, tags, force)
    ↓
ram.NewRolePolicyAttachment × N (parented to role)
    ↓
ctx.Export (role_id, role_name, arn)
```

## Key Components

### Controller (`main.go`)

The `Resources` function orchestrates the entire deployment:

1. Initializes locals from the stack input (tag computation, reference to spec)
2. Creates a region-scoped Alibaba Cloud provider
3. Creates the RAM role with all spec fields
4. Iterates over `spec.PolicyAttachments`, creating one `ram.RolePolicyAttachment` per entry
5. Exports three outputs: `role_id`, `role_name`, `arn`

The `policyAttachment` helper function handles individual attachments. Each attachment is named `{roleName}-{policyName}-{policyType}` to ensure uniqueness within the Pulumi state.

The `optionalString` helper returns `nil` for empty strings, allowing Pulumi to skip optional API parameters rather than sending empty values.

### Locals (`locals.go`)

The `initializeLocals` function computes:

- **Tags**: Merges standard OpenMCF tags (`resource`, `resource_name`, `resource_kind`, `resource_id`, `organization`, `environment`) with user-provided `spec.Tags`. User tags take precedence on key conflict.
- **Reference**: Stores the full `AlicloudRamRole` resource for convenient access throughout the module.

Three default-resolution helpers handle optional proto fields with pointer semantics:

- `maxSessionDuration(spec)` — returns the configured value or `3600`
- `forceDelete(spec)` — returns the configured value or `false`
- `policyType(pa)` — returns the configured value or `"System"`

These helpers exist because Go proto3 `optional` fields are pointer types (`*int32`, `*bool`, `*string`). The helpers dereference the pointer or return the documented default.

### Outputs (`outputs.go`)

Defines three output key constants:

- `OpRoleId = "role_id"` — the RAM role ID
- `OpRoleName = "role_name"` — the role name as created
- `OpArn = "arn"` — the full ARN (`acs:ram::<account-id>:role/<role-name>`)

These constants are referenced by `ctx.Export()` in the controller and must match the field names in `stack_outputs.proto`.

## Resource Relationships

```
Provider (alicloud, region from spec.region)
  └── ram.Role (spec.roleName)
        ├── ram.RolePolicyAttachment ("roleName-policyName-System")
        ├── ram.RolePolicyAttachment ("roleName-policyName-Custom")
        └── ... (one per spec.policyAttachments entry)
```

All resources use `pulumi.Provider(alicloudProvider)` to ensure consistent region configuration. Policy attachments use `pulumi.Parent(role)` to create an explicit dependency hierarchy — deleting the role cascades to attachments, and the Pulumi dependency graph reflects the RAM API constraint that attachments require the role to exist.

## Design Decisions

**DD07 Bundling**: The role and its policy attachments are managed as a single Pulumi program. This bundles what would otherwise be separate lifecycle operations (create role, then attach policies) into one atomic deployment. A role without policies can authenticate via STS but has zero permissions — the bundling ensures roles are provisioned with their intended permissions.

**Parent Chaining**: Each `RolePolicyAttachment` is parented to the `Role` resource. This ensures correct ordering (role first, then attachments) and clean teardown (detach before delete) without explicit `dependsOn` declarations.

**Naming Convention**: Attachment resources are named `{roleName}-{policyName}-{policyType}` to guarantee uniqueness. Two attachments with the same policy name but different types (`System` vs `Custom`) will not collide.

**Region as Provider Config**: RAM is a global Alibaba Cloud service, but the provider SDK requires a region for API endpoint routing. The `region` field configures the provider, not the role's scope.

## Customization Guide

| Customization | File to Modify | Notes |
|--------------|---------------|-------|
| Add a new spec field | `locals.go` (default resolver) + `main.go` (pass to resource) | Add a helper for optional fields with defaults |
| Change tag logic | `locals.go` (`initializeLocals`) | Modify the tag map construction |
| Add a new output | `outputs.go` (constant) + `main.go` (`ctx.Export`) | Must also update `stack_outputs.proto` |
| Add a new resource type | `main.go` (new resource creation) | Use `pulumi.Provider(alicloudProvider)` and consider parent chaining |
