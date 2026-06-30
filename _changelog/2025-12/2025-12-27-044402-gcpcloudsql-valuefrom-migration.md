# GcpCloudSql StringValueOrRef Migration

**Date**: December 27, 2025  
**Type**: Enhancement  
**Components**: API Definitions, GCP Provider, Pulumi IaC, Terraform IaC, Resource References

## Summary

Migrated the GcpCloudSql component to use `StringValueOrRef` type for `project_id` and `network.vpc_id` fields, enabling dynamic cross-resource references. This allows Cloud SQL instances to reference GcpProject and GcpVpc resources dynamically instead of requiring hardcoded values, improving infrastructure composability.

## Problem Statement / Motivation

The GcpCloudSql component previously used plain `string` types for resource identifiers like `project_id` and `vpc_id`. This created limitations:

### Pain Points

- **Hard-coded dependencies**: Users had to manually copy project IDs and VPC network IDs into their manifests
- **Maintenance burden**: When referenced resources changed, all dependent manifests needed manual updates
- **No dynamic composition**: Couldn't create infrastructure stacks where Cloud SQL automatically gets its project from a GcpProject resource
- **Inconsistency**: Other GCP components (GcpGkeCluster, GcpVpc, GcpSubnetwork) already supported `StringValueOrRef`, making GcpCloudSql an outlier

## Solution / What's New

Implemented the `StringValueOrRef` pattern for the GcpCloudSql component, allowing both literal values and dynamic references:

### Before: Hard-coded Values Only

```yaml
spec:
  projectId: my-gcp-project
  network:
    vpcId: projects/my-gcp-project/global/networks/my-vpc
```

### After: Flexible References

```yaml
# Option 1: Literal values (backward compatible)
spec:
  projectId:
    value: my-gcp-project
  network:
    vpcId:
      value: projects/my-gcp-project/global/networks/my-vpc

# Option 2: Cross-resource references
spec:
  projectId:
    valueFrom:
      kind: GcpProject
      name: main-project
      fieldPath: status.outputs.project_id
  network:
    vpcId:
      valueFrom:
        kind: GcpVpc
        name: main-vpc
        fieldPath: status.outputs.network_id
```

## Implementation Details

### Proto Schema Changes

Updated `spec.proto` with new field types and validation:

```protobuf
// Before
string project_id = 1 [
  (buf.validate.field).required = true,
  (buf.validate.field).string = {pattern: "^[a-z][a-z0-9-]{4,28}[a-z0-9]$"}
];

// After
dev.planton.shared.foreignkey.v1.StringValueOrRef project_id = 1 [
  (buf.validate.field).required = true,
  (dev.planton.shared.foreignkey.v1.default_kind) = GcpProject,
  (dev.planton.shared.foreignkey.v1.default_kind_field_path) = "status.outputs.project_id"
];
```

Also updated the CEL validation for the network message:

```protobuf
// Before
expression: "!this.private_ip_enabled || size(this.vpc_id) > 0"

// After
expression: "!this.private_ip_enabled || (has(this.vpc_id) && (has(this.vpc_id.value) || has(this.vpc_id.value_from)))"
```

### Test Updates

Updated `spec_test.go` with comprehensive tests for both patterns:

- Added helper function for creating `StringValueOrRef` with literal values
- Updated all existing tests to use new field types
- Added new tests for `value_from` reference validation
- Added production configuration examples with both patterns

### Pulumi IaC Changes

Updated `database.go` to extract values from `StringValueOrRef`:

```go
// Before
Project: pulumi.String(spec.ProjectId),
PrivateNetwork: pulumi.String(spec.Network.VpcId),

// After
Project: pulumi.String(spec.ProjectId.GetValue()),
PrivateNetwork: pulumi.String(spec.Network.VpcId.GetValue()),
```

### Terraform IaC Changes

Updated `variables.tf` with new object structure:

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

Updated `locals.tf` to extract values:

```hcl
project_id = (
  var.spec.project_id != null
  ? coalesce(var.spec.project_id.value, "")
  : ""
)
```

### Files Changed

| File | Changes |
|------|---------|
| `spec.proto` | Added foreignkey import, changed field types, updated CEL validation |
| `spec_test.go` | Added helper, updated 20+ test cases, added new reference tests |
| `database.go` | Updated to use `GetValue()` for StringValueOrRef fields |
| `examples.md` | Added examples with both literal and value_from patterns |
| `iac/pulumi/examples.md` | Updated all examples with new format |
| `iac/tf/variables.tf` | Changed to object type for StringValueOrRef |
| `iac/tf/locals.tf` | Added value extraction logic |
| `iac/tf/main.tf` | Updated to use local values |
| `iac/tf/examples.md` | Updated all examples with new format |
| `iac/hack/manifest.yaml` | Updated test manifest |

## Benefits

- **Dynamic Infrastructure**: Cloud SQL can now automatically derive its project and VPC from other resources
- **Reduced Maintenance**: Changes to parent resources (GcpProject, GcpVpc) automatically propagate
- **Composable Stacks**: Build infrastructure stacks where resources reference each other
- **Consistency**: Aligns with other GCP components that already use `StringValueOrRef`
- **Backward Compatible**: Literal values still work via the `{value: "..."}` syntax

## Impact

### API Changes

- **Breaking**: Field types changed from `string` to `StringValueOrRef`
- **Migration**: Existing manifests need to wrap values in `{value: "..."}` structure
- **YAML Format**: New structure required for both Pulumi and Terraform manifests

### Who is Affected

- Users deploying GcpCloudSql via Pulumi or Terraform
- Existing manifests need migration to new format
- Documentation and examples updated accordingly

## Related Work

This change is part of the GCP ValueFrom Migration initiative documented in `apis/gcp-value-from-anaylasis.md`. The following components were identified for similar migration:

- ✅ GcpCloudSql (this change)
- ⏳ GcpCloudRun
- ⏳ GcpCloudFunction
- ⏳ GcpServiceAccount
- ⏳ GcpSecretsManager
- ⏳ GcpDnsZone
- ⏳ GcpArtifactRegistryRepo
- ⏳ GcpCloudCdn
- ⏳ GcpGcsBucket

---

**Status**: ✅ Production Ready  
**Validation**: All tests pass, build successful

