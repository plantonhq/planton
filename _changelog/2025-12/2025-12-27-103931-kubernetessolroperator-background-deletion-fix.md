# KubernetesSolrOperator Background Deletion Fix

**Date**: December 27, 2025
**Type**: Bug Fix
**Components**: Kubernetes Provider, Pulumi CLI Integration, IAC Stack Runner

## Summary

Fixed the `pulumi destroy` timeout issue for the KubernetesSolrOperator module by adding background deletion propagation policy annotations. This ensures namespace and CRD resources are deleted immediately rather than waiting for child resources, preventing the 10-minute timeout caused by foreground deletion race conditions.

## Problem Statement / Motivation

When destroying a KubernetesSolrOperator deployment using `pulumi destroy`, the operation would consistently timeout after 10 minutes. The namespace and CRDs would get stuck in a "deleting" state.

### Pain Points

- `pulumi destroy` for KubernetesSolrOperator takes 10+ minutes and fails
- Namespace stuck in "Terminating" state indefinitely
- CRDs stuck in "deleting" state during teardown
- Clean `create → destroy → create` cycles impossible without manual intervention
- Same issue as KubernetesSolr (SolrCloud CR) but affecting the operator module

## Root Cause

The issue is a **Foreground Deletion Race Condition** between Kubernetes and the resources managed by this module:

### Namespace Deletion Issue

1. Pulumi issues DELETE with `propagationPolicy: Foreground` (default)
2. Kubernetes adds `foregroundDeletion` finalizer to namespace
3. Kubernetes waits for all resources inside the namespace to be deleted first
4. If child resources have their own finalizers or are being recreated by operators, deletion stalls
5. 10-minute timeout occurs

### CRD Deletion Issue

1. CRDs have built-in protection - they cannot be deleted while CustomResources of that type exist
2. If SolrCloud instances exist in other namespaces (not managed by this stack), CRD deletion blocks
3. The operator may recreate CRs during the deletion window, causing an infinite loop

## Solution / What's New

Added `pulumi.com/deletionPropagationPolicy: "background"` annotation to both the Namespace and CRD resources:

### Background Deletion Behavior

With background deletion:
1. Pulumi issues DELETE with `propagationPolicy: Background`
2. Namespace and CRDs are removed from the API server immediately
3. Kubernetes garbage collector cleans up child resources asynchronously
4. Destroy completes in seconds instead of timing out

## Implementation Details

### 1. Namespace Resource Annotation

**File**: `apis/dev/planton/provider/kubernetes/kubernetessolroperator/v1/iac/pulumi/module/main.go`

Added detailed inline documentation and the annotation to the namespace metadata:

```go
Metadata: &metav1.ObjectMetaArgs{
    Name:   pulumi.String(locals.Namespace),
    Labels: pulumi.ToStringMap(locals.Labels),
    // CRITICAL: Background Deletion Propagation Policy
    //
    // This annotation prevents namespace deletion from timing out during `pulumi destroy`.
    // [detailed explanation in code]
    Annotations: pulumi.StringMap{
        "pulumi.com/deletionPropagationPolicy": pulumi.String("background"),
    },
},
```

### 2. CRD Transformation for Remote Manifest

The CRDs are loaded from a remote URL using `yaml.NewConfigFile`. Added a transformation function to inject the annotation into all CRD resources:

```go
crds, err := pulumiyaml.NewConfigFile(ctx, locals.CrdsResourceName,
    &pulumiyaml.ConfigFileArgs{
        File: locals.CrdManifestURL,
        Transformations: []pulumiyaml.Transformation{
            // Inject background deletion policy annotation into all resources
            func(state map[string]interface{}, opts ...pulumi.ResourceOption) {
                // ... annotation injection logic
                annotations["pulumi.com/deletionPropagationPolicy"] = "background"
            },
        },
    },
    pulumi.Provider(kubernetesProvider))
```

### 3. README Documentation

Added a comprehensive "Deletion Behavior" section to the README explaining:
- Background deletion propagation and why it matters
- The specific issues with namespace and CRD deletion
- How background deletion solves the race condition
- A table of affected resources
- Testing instructions for verifying destroy operations

## Benefits

### Operational Improvements

- **Destroy completes in seconds** instead of 10+ minute timeout
- **Clean lifecycle management**: `create → destroy → create` works reliably
- **No manual intervention**: No need to manually patch finalizers or delete stuck resources
- **Consistent with KubernetesSolr module**: Same fix pattern applied for consistency

### Developer Experience

- **Detailed inline documentation**: Future developers understand why the annotation exists
- **README documentation**: Users know what to expect during destroy operations
- **Troubleshooting guidance**: README includes debugging tips for edge cases

## Impact

### Affected Resources

| Resource | Annotation Location | Purpose |
|----------|-------------------|---------|
| Namespace | `Metadata.Annotations` | Prevents blocking on child resource finalizers |
| CRDs (via ConfigFile) | Transformation function | Allows deletion even if CRs exist in other namespaces |

### Testing

Successfully verified:
1. `pulumi destroy` completes without timeout
2. Resources are cleaned up from the cluster
3. Build passes with no errors

## Related Work

- **KubernetesSolr Background Deletion Fix** (2025-12-27-091641): Same fix pattern applied to SolrCloud CR in the KubernetesSolr module
- Both fixes stem from the same root cause: foreground deletion race condition with operator-managed resources

## Files Changed

| File | Change |
|------|--------|
| `apis/.../kubernetessolroperator/v1/iac/pulumi/module/main.go` | Added background deletion annotations with documentation |
| `apis/.../kubernetessolroperator/v1/iac/pulumi/README.md` | Added Deletion Behavior documentation section |

---

**Status**: ✅ Production Ready
**Timeline**: ~1 hour (including investigation and documentation)

