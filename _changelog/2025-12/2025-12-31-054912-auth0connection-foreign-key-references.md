# Auth0Connection Foreign Key References for Client Linking

**Date**: December 31, 2025
**Type**: Breaking Change
**Components**: API Definitions, Auth0 Provider, Pulumi CLI Integration, Terraform Module

## Summary

Added foreign key reference support to Auth0Connection's `enabled_clients` field, enabling declarative linking between Auth0Connection and Auth0Client deployment components. Users can now reference managed Auth0Client resources by name instead of hardcoding client IDs, with automatic resolution of the client ID from the referenced component's outputs.

## Problem Statement / Motivation

Auth0 connections require a list of client IDs to specify which applications can use the connection. Previously, users had to hardcode these client IDs directly in the Auth0Connection manifest, creating several issues.

### Pain Points

- **Hardcoded Dependencies**: Client IDs had to be copied manually from Auth0Client outputs
- **Fragile Configuration**: Changes to Auth0Client required manual updates to Auth0Connection
- **No Declarative Relationships**: Infrastructure dependencies were implicit rather than explicit
- **Error-Prone**: Typos in client IDs resulted in runtime failures instead of validation errors

## Solution / What's New

Changed the `enabled_clients` field from `repeated string` to `repeated StringValueOrRef`, leveraging Planton's foreign key reference system. This allows users to either:

1. **Direct Value**: Specify client IDs directly using `{value: "client-id"}`
2. **Foreign Key Reference**: Reference an Auth0Client component using `{value_from: {kind: Auth0Client, name: "my-app"}}`

When using foreign key references, the client ID is automatically resolved from the Auth0Client's `status.outputs.client_id` field at deployment time.

### Reference Pattern

```yaml
# Before (hardcoded)
enabled_clients:
  - "abc123def456"

# After - Direct value
enabled_clients:
  - value: "abc123def456"

# After - Foreign key reference
enabled_clients:
  - value_from:
      kind: Auth0Client
      name: my-web-app
```

## Implementation Details

### Proto Schema Changes

**File**: `apis/dev/planton/provider/auth0/auth0connection/v1/spec.proto`

```protobuf
import "dev/planton/shared/foreignkey/v1/foreign_key.proto";

// enabled_clients changed from repeated string to:
repeated dev.planton.shared.foreignkey.v1.StringValueOrRef enabled_clients = 3 [
  (dev.planton.shared.foreignkey.v1.default_kind) = Auth0Client,
  (dev.planton.shared.foreignkey.v1.default_kind_field_path) = "status.outputs.client_id"
];
```

The field options specify:
- `default_kind`: References resolve to Auth0Client by default
- `default_kind_field_path`: The client_id is extracted from `status.outputs.client_id`

### Pulumi Module Updates

**File**: `apis/dev/planton/provider/auth0/auth0connection/v1/iac/pulumi/module/locals.go`

```go
// Extract values from StringValueOrRef
for _, client := range spec.EnabledClients {
    if client != nil && client.GetValue() != "" {
        locals.EnabledClients = append(locals.EnabledClients, client.GetValue())
    }
}
```

### Terraform Module Updates

**File**: `apis/dev/planton/provider/auth0/auth0connection/v1/iac/tf/variables.tf`

```hcl
enabled_clients = optional(list(object({
  value = string
})))
```

**File**: `apis/dev/planton/provider/auth0/auth0connection/v1/iac/tf/locals.tf`

```hcl
enabled_clients = var.spec.enabled_clients != null ? [
  for client in var.spec.enabled_clients : client.value
] : []
```

### Test Updates

Updated `spec_test.go` to use the new `StringValueOrRef` type:

```go
EnabledClients: []*foreignkeyv1.StringValueOrRef{
    {LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "client-id-1"}},
    {LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "client-id-2"}},
},
```

## Breaking Changes

This is a **breaking change** to the Auth0Connection manifest format.

### Migration Required

All existing Auth0Connection manifests must be updated:

**Before:**
```yaml
spec:
  enabled_clients:
    - "client-id-1"
    - "client-id-2"
```

**After:**
```yaml
spec:
  enabled_clients:
    - value: "client-id-1"
    - value: "client-id-2"
```

The migration is straightforward - wrap each client ID in a `value:` object.

## Benefits

### For Platform Engineers
- **Declarative Dependencies**: Infrastructure relationships are explicit in the manifest
- **Automatic Resolution**: Client IDs are resolved from Auth0Client outputs
- **Reduced Errors**: Type-safe references catch issues at validation time

### For DevOps Workflows
- **Single Source of Truth**: No need to copy client IDs between resources
- **Change Propagation**: Auth0Client changes automatically reflect in connections
- **Clearer Infrastructure Graph**: Dependencies are visible and traceable

### For Multi-Environment Deployments
- **Environment Agnostic**: Reference by name, not environment-specific IDs
- **Consistent Patterns**: Same manifest works across dev/staging/prod

## Impact

| Area | Impact |
|------|--------|
| Auth0Connection manifests | **Breaking** - Must update format |
| Auth0Client | No changes required |
| Pulumi deployment | Compatible - extracts values seamlessly |
| Terraform deployment | Compatible - updated variable types |
| Tests | All 30 tests passing |

## Files Changed

| File | Change |
|------|--------|
| `spec.proto` | Field type changed to StringValueOrRef with options |
| `api.proto` | Example comments updated |
| `spec_test.go` | Test cases updated for new type |
| `iac/pulumi/module/locals.go` | Value extraction logic |
| `iac/tf/variables.tf` | Variable type updated |
| `iac/tf/locals.tf` | Value extraction logic |
| `examples.md` | All examples updated + new FK section |
| `README.md` | Added FK feature, updated examples |
| `iac/hack/manifest.yaml` | Test manifest updated |

## Usage Examples

### Complete Auth0Client + Auth0Connection Setup

```yaml
# 1. Create the Auth0Client
apiVersion: auth0.planton.dev/v1
kind: Auth0Client
metadata:
  name: my-web-app
  org: my-organization
  env: production
spec:
  application_type: spa
  name: My Web Application
  callbacks:
    - "https://app.example.com/callback"

---
# 2. Create Auth0Connection referencing the client
apiVersion: auth0.planton.dev/v1
kind: Auth0Connection
metadata:
  name: user-database
  org: my-organization
  env: production
spec:
  strategy: auth0
  display_name: Sign Up with Email
  enabled_clients:
    - value_from:
        kind: Auth0Client
        name: my-web-app
  database_options:
    password_policy: good
    brute_force_protection: true
```

### Mixed Direct and Reference Values

```yaml
enabled_clients:
  # Managed Auth0Client
  - value_from:
      kind: Auth0Client
      name: internal-portal
  # External/legacy client ID
  - value: "legacy-app-client-id-abc123"
```

## Related Work

- **Auth0Client Component**: Created in previous session (2025-12-30-070305)
- **Foreign Key System**: Uses `StringValueOrRef` from `shared/foreignkey/v1`
- **Similar Patterns**: GCP VPC `project_id`, Civo Database `network_id`

---

**Status**: ✅ Production Ready
**Validation**: All tests passing (Auth0Connection: 30, Auth0Client: 28)

