# TemporalKubernetes Namespace Initialization Bug Fix

**Date**: October 17, 2025  
**Type**: Bug Fix  
**Component**: TemporalKubernetes

## Summary

Fixed a critical bug in the TemporalKubernetes Pulumi module where namespace initialization logic incorrectly used an `else` clause that overwrote the valid default namespace with an empty value, causing Pulumi deployments to fail with "resource name argument (for URN creation) cannot be empty" error.

## Problem

The TemporalKubernetes Pulumi module failed during deployment with the following error:

```
error: an unhandled error occurred: program failed: 
1 error occurred:
    * failed to create namespace: failed to create  namespace: resource name argument (for URN creation) cannot be empty
```

### Root Cause

The namespace initialization logic in `locals.go` used an `else` clause that would unconditionally set the namespace to `stackInput.KubernetesNamespace` when no custom namespace label was present:

```go
locals.Namespace = target.Metadata.Name
if target.Metadata.Labels != nil &&
    target.Metadata.Labels[kuberneteslabels.NamespaceLabelKey] != "" {
    locals.Namespace = target.Metadata.Labels[kuberneteslabels.NamespaceLabelKey]
} else {
    locals.Namespace = stackInput.KubernetesNamespace  // ❌ Empty value overwrites valid default!
}
```

When `stackInput.KubernetesNamespace` was empty (the common case), it would overwrite the valid default namespace (`target.Metadata.Name`) with an empty string, causing the namespace resource creation to fail.

### Symptoms

- Pulumi preview/update failed immediately during namespace creation
- Port-forward commands showed empty namespace: `kubectl port-forward -n  service/temporal-test-frontend`
- Outputs showed empty namespace in service endpoints: `temporal-test-frontend..svc.cluster.local:7233`

## Solution

Changed the `else` clause to a separate `if` statement with a non-empty check, matching the pattern used by all other Kubernetes workload resources (PostgresKubernetes, MicroserviceKubernetes, ClickHouseKubernetes, etc.):

```go
// Priority order:
// 1. Default: metadata.name
// 2. Override with custom label if provided
// 3. Override with stackInput if provided

locals.Namespace = target.Metadata.Name

if target.Metadata.Labels != nil &&
    target.Metadata.Labels[kuberneteslabels.NamespaceLabelKey] != "" {
    locals.Namespace = target.Metadata.Labels[kuberneteslabels.NamespaceLabelKey]
}

if stackInput.KubernetesNamespace != "" {
    locals.Namespace = stackInput.KubernetesNamespace
}
```

### Key Improvements

1. **Preserves default**: The namespace now correctly defaults to `metadata.name` when no overrides are provided
2. **Conditional override**: `stackInput.KubernetesNamespace` only overrides when it has a non-empty value
3. **Consistent pattern**: Matches the namespace initialization logic used across all other Kubernetes workload resources
4. **Clear priority**: Added comments documenting the three-level priority order

## Files Changed

### Modified

- `apis/project/planton/provider/kubernetes/workload/temporalkubernetes/v1/iac/pulumi/module/locals.go`
  - Lines 56-73: Updated namespace initialization logic
  - Added priority order comments
  - Changed `else` to separate conditional check

## Testing

### Verification Steps

1. Run Pulumi preview with a basic Temporal manifest:
   ```bash
   planton pulumi preview --manifest temporal-test.yaml --module-dir ${TEMPORAL_MODULE}
   ```

2. Verify namespace is correctly set:
   - Namespace should be `temporal-test` (from `metadata.name`)
   - Outputs should show correct service FQDNs: `temporal-test-frontend.temporal-test.svc.cluster.local:7233`
   - Port-forward commands should include namespace: `kubectl port-forward -n temporal-test service/temporal-test-frontend 7233:7233`

3. Preview should succeed without namespace errors

### Test Scenarios

- ✅ **Default behavior**: Namespace defaults to `metadata.name` when no overrides provided
- ✅ **Label override**: Custom namespace label (`kubernetes.planton.dev/namespace`) correctly overrides default
- ✅ **StackInput override**: Non-empty `stackInput.KubernetesNamespace` takes precedence over both defaults and labels
- ✅ **Empty StackInput**: Empty `stackInput.KubernetesNamespace` does not overwrite valid defaults

## Impact

### Before Fix
- ❌ All TemporalKubernetes deployments failed at namespace creation
- ❌ Impossible to deploy Temporal via Pulumi module
- ❌ No workaround available without code changes

### After Fix
- ✅ TemporalKubernetes deployments succeed with correct namespace
- ✅ Namespace correctly derived from `metadata.name`
- ✅ Optional namespace overrides work as expected
- ✅ Consistent behavior with other Kubernetes workload resources

## Related Resources

### Similar Patterns in Codebase

The corrected pattern is consistent with namespace initialization in:
- `postgreskubernetes/v1/iac/pulumi/module/locals.go` (lines 50-64)
- `kubernetesmicroservice/v1/iac/pulumi/module/locals.go` (lines 68-82)
- `clickhousekubernetes/v1/iac/pulumi/module/locals.go` (lines 51-65)
- `kafkakubernetes/v1/iac/pulumi/module/locals.go` (lines 72-86)
- `mongodbkubernetes/v1/iac/pulumi/module/locals.go` (lines 51-65)
- `helmrelease/v1/iac/pulumi/module/locals.go` (lines 44-58)

### Documentation

The namespace priority order is now explicitly documented in comments:
1. **Default**: Uses `metadata.name` as the namespace
2. **Label override**: Uses `kubernetes.planton.dev/namespace` label if present and non-empty
3. **StackInput override**: Uses `stackInput.KubernetesNamespace` if present and non-empty (highest precedence)

## Notes

This bug likely existed since the initial implementation of TemporalKubernetes and went undetected because most testing may have used different namespace override mechanisms. The fix ensures the module follows the established pattern used across all Kubernetes workload resources in the project.

