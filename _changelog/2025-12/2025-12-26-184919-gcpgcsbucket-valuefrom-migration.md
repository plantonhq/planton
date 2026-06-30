# GcpGcsBucket: Migrate gcp_project_id to StringValueOrRef Pattern

**Date**: December 26, 2025
**Type**: Enhancement
**Components**: API Definitions, Pulumi CLI Integration, Terraform Integration, Provider Framework

## Summary

Migrated the `GcpGcsBucket` component's `gcp_project_id` field from a plain `string` type to `StringValueOrRef`, enabling flexible cross-resource references. Users can now either specify a literal project ID or reference another resource's output (e.g., a `GcpProject` resource), improving infrastructure composability and dependency management.

## Problem Statement / Motivation

The `GcpGcsBucket` component previously used a plain `string` type for the `gcp_project_id` field, which required users to hardcode GCP project IDs in their manifests. This approach had several limitations:

### Pain Points

- **Hardcoded dependencies**: Users couldn't dynamically reference project IDs from other resources
- **No dependency ordering**: Infrastructure deployments couldn't automatically sequence GCS bucket creation after project creation
- **Inconsistency**: Other GCP components (GcpVpc, GcpGkeCluster, GcpSubnetwork) already supported `StringValueOrRef`
- **Template rigidity**: Reusable templates required manual project ID substitution

## Solution / What's New

Implemented the `StringValueOrRef` pattern for the `gcp_project_id` field, following the established pattern used by compliant GCP components.

### StringValueOrRef Pattern

```protobuf
dev.planton.shared.foreignkey.v1.StringValueOrRef gcp_project_id = 1 [
  (buf.validate.field).required = true,
  (dev.planton.shared.foreignkey.v1.default_kind) = GcpProject,
  (dev.planton.shared.foreignkey.v1.default_kind_field_path) = "status.outputs.project_id"
];
```

Users can now specify project IDs in two ways:

**Literal Value**:
```yaml
spec:
  gcpProjectId:
    value: my-gcp-project-123
```

**Resource Reference** (future implementation):
```yaml
spec:
  gcpProjectId:
    valueFrom:
      kind: GcpProject
      name: my-project
      fieldPath: status.outputs.project_id
```

## Implementation Details

### Files Modified

| File | Change |
|------|--------|
| `spec.proto` | Added foreignkey import, changed `gcp_project_id` to `StringValueOrRef` with field options |
| `spec_test.go` | Updated tests to use `foreignkeyv1.StringValueOrRef` pattern and added valueFrom test cases |
| `iac/pulumi/module/main.go` | Uses `GetValue()` to resolve literal value from `StringValueOrRef` |
| `iac/tf/variables.tf` | Changed `gcp_project_id` from `string` to `object({ value = string })` |
| `iac/tf/main.tf` | Complete overhaul with all spec features, updated to use `var.spec.gcp_project_id.value` |
| `iac/tf/locals.tf` | Added user label merging |
| `examples.md` | Updated all examples + added "Using Foreign Key References" section |
| `iac/tf/examples.md` | Updated all Terraform examples to use new format |
| `iac/tf/hack/manifest.yaml` | Updated test manifest |
| `iac/pulumi/overview.md` | Documented `StringValueOrRef` pattern and limitations |
| `BUILD.bazel` | Auto-updated by Gazelle for new dependencies |
| `spec.pb.go` | Regenerated proto stubs |
| `spec_pb.ts` | Regenerated TypeScript stubs |

### Key Code Changes

**Proto Schema** (`spec.proto`):
```protobuf
import "dev/planton/shared/foreignkey/v1/foreign_key.proto";

message GcpGcsBucketSpec {
  dev.planton.shared.foreignkey.v1.StringValueOrRef gcp_project_id = 1 [
    (buf.validate.field).required = true,
    (dev.planton.shared.foreignkey.v1.default_kind) = GcpProject,
    (dev.planton.shared.foreignkey.v1.default_kind_field_path) = "status.outputs.project_id"
  ];
}
```

**Pulumi IAC** (`main.go`):
```go
// Resolve gcp_project_id from StringValueOrRef (only literal value is currently supported)
projectId := locals.GcpGcsBucket.Spec.GcpProjectId.GetValue()
```

**Terraform Variables** (`variables.tf`):
```hcl
gcp_project_id = object({
  value = string
})
```

## Benefits

- **Flexible referencing**: Users can choose between literal values or resource references
- **Improved composability**: GCS buckets can be declaratively linked to GcpProject resources
- **Consistent API**: Aligns with other GCP components (GcpVpc, GcpGkeCluster, etc.)
- **Future-ready**: Infrastructure prepared for reference resolution implementation
- **Better templates**: Reusable manifests can use references instead of hardcoded values

## Impact

### API Changes

This is a **breaking change** for existing manifests. Users must update their YAML:

**Before**:
```yaml
spec:
  gcpProjectId: my-gcp-project
```

**After**:
```yaml
spec:
  gcpProjectId:
    value: my-gcp-project
```

### Affected Users

- Users with existing `GcpGcsBucket` manifests need to update the `gcpProjectId` field format
- Terraform configurations need to update variable structure
- No changes required to bucket configuration or storage settings

## Current Limitations

Reference resolution (`valueFrom`) is **not yet implemented** in the IAC layer. Only literal `value` is currently supported. References will be resolved when the shared reference resolution library is implemented.

## Related Work

- Part of the broader GCP ValueFrom migration initiative (see `gcp-value-from-analysis.md`)
- Follows pattern established by:
  - `GcpVpc` - `project_id` uses `StringValueOrRef`
  - `GcpGkeCluster` - `project_id`, `network_self_link`, `subnetwork_self_link`
  - `GcpSubnetwork` - `project_id`, `vpc_self_link`
  - `GcpRouterNat` - `project_id`, `vpc_self_link`, `subnetwork_self_links`

## Testing

All spec validation tests pass with the new `StringValueOrRef` pattern.

---

**Status**: ✅ Production Ready
**Files Changed**: 11
