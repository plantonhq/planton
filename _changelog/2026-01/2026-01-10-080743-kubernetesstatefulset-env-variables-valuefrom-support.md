# KubernetesStatefulSet Environment Variables ValueFrom Support

**Date**: January 10, 2026
**Type**: Enhancement
**Components**: API Definitions, Kubernetes Provider, Pulumi IaC Module, Terraform IaC Module

## Summary

Added support for environment variables in `KubernetesStatefulSet` to be provided either as direct string values or as references to other Planton resource fields using the `StringValueOrRef` type. This enables dynamic configuration where environment variables can be derived from outputs of other deployed resources.

## Problem Statement / Motivation

When deploying stateful applications to Kubernetes via `KubernetesStatefulSet`, users previously had to provide environment variable values as plaintext strings directly in their manifest files via `spec.container.app.env.variables`. This created several limitations:

### Pain Points

- **Static configuration**: Environment variables had to be hardcoded, requiring manual updates when dependent resources changed
- **No cross-resource references**: Could not reference outputs from other Planton resources (e.g., database host from another cluster)
- **Manual coordination**: When deploying interconnected stateful services, users had to manually copy outputs and update configurations
- **Error-prone**: Typos or stale values could lead to runtime failures
- **Inconsistent with KubernetesDeployment**: The pattern didn't match the enhanced `valueFrom` pattern added to KubernetesDeployment

## Solution / What's New

Changed the `variables` field in `KubernetesStatefulSetContainerAppEnv` to use `StringValueOrRef` type from `foreign_key.proto`, which supports:

1. **Direct string value** (for static configuration)
2. **ValueFrom reference** (for dynamic cross-resource references)

### Before (Old Format)

```yaml
spec:
  container:
    app:
      env:
        variables:
          PGDATA: "/var/lib/postgresql/data"
          POSTGRES_PORT: "5432"
```

### After (New Format)

**Option 1: Direct String Value**
```yaml
spec:
  container:
    app:
      env:
        variables:
          PGDATA:
            value: "/var/lib/postgresql/data"
```

**Option 2: ValueFrom Reference**
```yaml
spec:
  container:
    app:
      env:
        variables:
          POSTGRES_HOST:
            valueFrom:
              kind: PostgresCluster
              name: my-postgres
              fieldPath: "status.outputs.host"
```

**Option 3: Mixed (Both Types)**
```yaml
spec:
  container:
    app:
      env:
        variables:
          # Static value
          PGDATA:
            value: "/var/lib/postgresql/data"
          # Dynamic reference - resolved by orchestrator
          POSTGRES_HOST:
            valueFrom:
              kind: PostgresCluster
              name: my-postgres
              fieldPath: "status.outputs.host"
```

## Implementation Details

### 1. Proto Schema Changes

**File**: `apis/dev/planton/provider/kubernetes/kubernetesstatefulset/v1/spec.proto`

**Before**:
```protobuf
message KubernetesStatefulSetContainerAppEnv {
  map<string, string> variables = 1;
  map<string, dev.planton.provider.kubernetes.KubernetesSensitiveValue> secrets = 2;
}
```

**After**:
```protobuf
message KubernetesStatefulSetContainerAppEnv {
  map<string, dev.planton.shared.foreignkey.v1.StringValueOrRef> variables = 1;
  map<string, dev.planton.provider.kubernetes.KubernetesSensitiveValue> secrets = 2;
}
```

### 2. Pulumi Module Updates

**File**: `apis/dev/planton/provider/kubernetes/kubernetesstatefulset/v1/iac/pulumi/module/statefulset.go`

**Key Logic**:
- The orchestrator resolves all `valueFrom` references and populates the `.value` field before invoking the Pulumi module
- The Pulumi module only reads from `.GetValue()` - it never handles `valueFrom` resolution

### 3. Terraform Module Updates

**File**: `apis/dev/planton/provider/kubernetes/kubernetesstatefulset/v1/iac/tf/variables.tf`

Updated the `variables` type definition to match `StringValueOrRef`:
```hcl
variables = optional(map(object({
  value = optional(string)
  value_from = optional(object({
    kind       = optional(string)
    env        = optional(string)
    name       = string
    field_path = optional(string)
  }))
})))
```

**File**: `apis/dev/planton/provider/kubernetes/kubernetesstatefulset/v1/iac/tf/statefulset.tf`

Updated the dynamic env block to read from `.value`.

### 4. Test Updates

**File**: `apis/dev/planton/provider/kubernetes/kubernetesstatefulset/v1/spec_test.go`

Added test cases for:
- Variables with direct string values - should pass
- Variables with valueFrom references - should pass validation
- Variables with mixed types (both value and valueFrom) - should pass
- ValueFrom ref missing required `name` field - should fail validation

## Files Changed

| File | Change |
|------|--------|
| `kubernetesstatefulset/v1/spec.proto` | Changed `variables` from `map<string, string>` to `map<string, StringValueOrRef>` |
| `kubernetesstatefulset/v1/iac/pulumi/module/statefulset.go` | Updated to read `.GetValue()` from `StringValueOrRef` |
| `kubernetesstatefulset/v1/iac/tf/variables.tf` | Updated variables type definition |
| `kubernetesstatefulset/v1/iac/tf/statefulset.tf` | Updated dynamic env block to read `.value` |
| `kubernetesstatefulset/v1/spec_test.go` | Added 4 new test cases for environment variables validation |

## Key Implementation Notes

- **Orchestrator resolves `valueFrom`**: The Planton CLI/orchestrator resolves all `valueFrom` references and places the derived values into the `.value` field before invoking IaC modules
- **IaC modules only read `.value`**: Pulumi and Terraform modules always expect `.value` to be populated - they never handle `valueFrom` resolution
- **Consistent with KubernetesDeployment**: This change maintains feature parity with `KubernetesDeployment` which received the same enhancement

## Build & Validation Commands

After making changes, run:

```bash
# 1. Regenerate proto stubs
make protos

# 2. Run component-specific tests
go test ./apis/dev/planton/provider/kubernetes/kubernetesstatefulset/v1/...

# 3. Full build
make build

# 4. Full test suite
make test
```

## Benefits

- **Dynamic configuration**: Environment variables can reference outputs from other Planton resources
- **Feature parity**: Consistent with `KubernetesDeployment` which has the same capability
- **Reduced manual coordination**: No need to copy-paste values between resource configurations
- **Type-safe references**: Protobuf validation ensures `valueFrom` references have required fields

## Impact

### Users
- Users can now create interconnected stateful deployments where environment variables are derived from other resources
- Particularly useful for databases that need connection info from other services
- No breaking changes to existing deployments

### Developers
- Pattern consistent with `KubernetesDeployment` and AWS provider components
- All tests updated and passing

## Related Work

- **KubernetesDeployment**: Same change applied (2026-01-10)
- **Shared type**: Uses `StringValueOrRef` from `apis/dev/planton/shared/foreignkey/v1/foreign_key.proto`
- **AWS components**: Pattern already used in `AwsAlb`, `AwsEcsService`, etc.

---

**Status**: Production Ready
**Test Results**: All tests passing
