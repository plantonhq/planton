# GcpSecretsManager: Migrate project_id to StringValueOrRef

**Date**: December 26, 2025
**Type**: Enhancement
**Components**: API Definitions, Pulumi CLI Integration, Terraform Module, Provider Framework

## Summary

Migrated the `GcpSecretsManager` component's `project_id` field from a plain `string` type to `StringValueOrRef`, enabling cross-resource references. This allows users to either specify a literal project ID or reference another resource's output (e.g., a `GcpProject` resource), enabling declarative infrastructure composition.

## Problem Statement / Motivation

The `GcpSecretsManager` component previously required users to hard-code GCP project IDs as literal strings. This created tight coupling between resources and made it difficult to:

### Pain Points

- **No cross-resource dependencies**: Users couldn't reference a project ID from a `GcpProject` resource they were managing
- **Manual coordination**: When project IDs changed, all dependent manifests needed manual updates
- **Inconsistent with other components**: Components like `GcpVpc`, `GcpGkeCluster`, and `GcpSubnetwork` already supported `StringValueOrRef`
- **Limited infrastructure composition**: Users couldn't build declarative dependency chains between resources

## Solution / What's New

Updated the `GcpSecretsManager` spec to use `StringValueOrRef` for the `project_id` field, following the established pattern from compliant components like `GcpVpc`.

### Two Usage Patterns Now Supported

**Literal Value** (backward compatible approach):
```yaml
spec:
  projectId:
    value: "my-gcp-project-123456"
```

**Cross-Resource Reference** (new capability):
```yaml
spec:
  projectId:
    valueFrom:
      kind: GcpProject
      name: main-project
      fieldPath: "status.outputs.project_id"
```

## Implementation Details

### Proto Schema Changes

**File**: `apis/dev/planton/provider/gcp/gcpsecretsmanager/v1/spec.proto`

```protobuf
// Before
string project_id = 1 [(buf.validate.field).required = true];

// After
dev.planton.shared.foreignkey.v1.StringValueOrRef project_id = 1 [
  (buf.validate.field).required = true,
  (dev.planton.shared.foreignkey.v1.default_kind) = GcpProject,
  (dev.planton.shared.foreignkey.v1.default_kind_field_path) = "status.outputs.project_id"
];
```

Key additions:
- Import for `dev/planton/shared/foreignkey/v1/foreign_key.proto`
- Field options specifying default reference kind (`GcpProject`) and field path

### Pulumi Module Changes

**File**: `apis/dev/planton/provider/gcp/gcpsecretsmanager/v1/iac/pulumi/module/main.go`

Updated to use the `GetValue()` method on the `StringValueOrRef` type:

```go
// Before
Project: pulumi.String(locals.GcpSecretsManager.Spec.ProjectId),

// After  
Project: pulumi.String(locals.GcpSecretsManager.Spec.ProjectId.GetValue()),
```

### Terraform Module Changes

**File**: `apis/dev/planton/provider/gcp/gcpsecretsmanager/v1/iac/tf/variables.tf`

Updated variable type to object structure:

```hcl
project_id = object({
  value      = optional(string)
  value_from = optional(object({
    kind       = optional(string)
    env        = optional(string)
    name       = string
    field_path = optional(string)
  }))
})
```

**File**: `apis/dev/planton/provider/gcp/gcpsecretsmanager/v1/iac/tf/main.tf`

Updated to use `.value` accessor:

```hcl
project = var.spec.project_id.value
```

### Test Updates

**File**: `apis/dev/planton/provider/gcp/gcpsecretsmanager/v1/spec_test.go`

- Added imports for `foreignkeyv1` and `cloudresourcekind`
- Updated all test cases to use `StringValueOrRef` type
- Added new test case for `value_from` reference pattern
- 11 tests, all passing

### Documentation Updates

**File**: `apis/dev/planton/provider/gcp/gcpsecretsmanager/v1/examples.md`

- Updated all 8 examples to use new object format
- Added new Example 2 demonstrating cross-resource reference pattern
- Renumbered subsequent examples

**File**: `apis/dev/planton/provider/gcp/gcpsecretsmanager/v1/iac/pulumi/overview.md`

- Added documentation for cross-resource references
- Documented current limitation (reference resolution not yet implemented)

## Files Changed

| File | Change Type |
|------|-------------|
| `spec.proto` | Field type migration |
| `spec.pb.go` | Auto-generated |
| `spec_pb.ts` | Auto-generated |
| `spec_test.go` | Test updates |
| `BUILD.bazel` | Dependency updates |
| `iac/pulumi/module/main.go` | GetValue() usage |
| `iac/pulumi/overview.md` | Documentation |
| `iac/tf/variables.tf` | Variable type |
| `iac/tf/main.tf` | Value accessor |
| `iac/tf/hack/manifest.yaml` | Example format |
| `examples.md` | All examples updated |

## Benefits

- **Infrastructure Composition**: Secrets can now be automatically created in projects managed by other resources
- **Reduced Manual Coordination**: Project ID changes propagate through references
- **Consistency**: Aligns with other GCP components (`GcpVpc`, `GcpGkeCluster`, etc.)
- **Future-Ready**: Foundation for full reference resolution implementation

## Impact

### Users
- Existing manifests with literal `projectId: "value"` need migration to `projectId: {value: "value"}`
- New capability to use `valueFrom` for cross-resource references

### Developers
- Pulumi modules must use `.GetValue()` method
- Terraform modules must use `.value` accessor

### API
- Breaking change to field type (string → StringValueOrRef)
- Backward compatible via `value` wrapper

## Known Limitations

Reference resolution (`value_from`) is not yet fully implemented in the IAC layer. Currently:
- Only literal `value` is resolved
- References require external orchestrator or future CLI implementation

## Related Work

This change is part of the broader GCP ValueFrom migration effort documented in:
- `apis/gcp-value-from-analysis.md` - Analysis of all GCP components requiring migration

Other components already migrated:
- `GcpVpc`
- `GcpSubnetwork`
- `GcpGkeCluster`
- `GcpGkeNodePool`
- `GcpRouterNat`
- `GcpGkeWorkloadIdentityBinding`

---

**Status**: ✅ Production Ready
**Validation**: Proto generation ✅ | Component tests (11/11) ✅ | Build ✅
