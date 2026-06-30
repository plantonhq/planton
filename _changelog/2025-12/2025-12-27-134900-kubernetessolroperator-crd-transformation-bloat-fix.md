# KubernetesSolrOperator: Remove CRD Transformation Causing 180MB Diff Bloat

**Date**: December 27, 2025
**Type**: Bug Fix
**Components**: Kubernetes Provider, Pulumi IAC Integration

## Summary

Removed the `ConfigFile` transformation from the KubernetesSolrOperator Pulumi module that was injecting background deletion annotations into CRDs. The transformation caused Pulumi to recompute diffs on every operation, resulting in 180MB+ diff sizes due to embedded OpenAPI schemas in the CRDs.

## Problem Statement

After adding a `ConfigFile` transformation to inject `pulumi.com/deletionPropagationPolicy: "background"` annotations into CRDs for improved deletion behavior, Pulumi diff sizes exploded to approximately 180MB.

### Root Cause

The `pulumiyaml.ConfigFile` with transformations modifies the in-memory state of resources loaded from the remote CRD manifest. This causes Pulumi to:

1. See every CRD resource as "changed" on every operation
2. Include the full CRD specification (including massive embedded OpenAPI schemas) in the diff
3. Generate enormous state updates for what should be no-op operations

Solr Operator CRDs contain extensive OpenAPI validation schemas for `SolrCloud`, `SolrBackup`, and `SolrPrometheusExporter` custom resources. When Pulumi thinks these need updating, it serializes the entire schema into the diff.

### Symptoms

```
# Before fix - diff size on every operation
Diff size: ~180MB

# After fix - diff size
Diff size: Normal (KB range)
```

## Solution

Removed the transformation approach entirely. The namespace background deletion policy is sufficient for clean teardown because:

1. **Namespace deletion triggers cascade**: When the namespace is deleted with background propagation, Kubernetes immediately removes the namespace object
2. **Operator stops running**: The Solr Operator pod (in the namespace) is terminated
3. **CRs become orphaned**: Without the operator reconciling, SolrCloud and related CRs can be garbage collected
4. **CRD deletion unblocks**: Once all CRs are gone, CRD deletion proceeds normally

### Before (Problematic)

```go
crds, err := pulumiyaml.NewConfigFile(ctx, locals.CrdsResourceName,
    &pulumiyaml.ConfigFileArgs{
        File: locals.CrdManifestURL,
        Transformations: []pulumiyaml.Transformation{
            func(state map[string]interface{}, opts ...pulumi.ResourceOption) {
                // This caused 180MB diffs!
                metadata := state["metadata"].(map[string]interface{})
                annotations := metadata["annotations"].(map[string]interface{})
                annotations["pulumi.com/deletionPropagationPolicy"] = "background"
            },
        },
    },
    pulumi.Provider(kubernetesProvider))
```

### After (Clean)

```go
crds, err := pulumiyaml.NewConfigFile(ctx, locals.CrdsResourceName,
    &pulumiyaml.ConfigFileArgs{
        File: locals.CrdManifestURL,
    },
    pulumi.Provider(kubernetesProvider))
```

## Implementation Details

### File Changed

**File**: `apis/dev/planton/provider/kubernetes/kubernetessolroperator/v1/iac/pulumi/module/main.go`

Removed the `Transformations` field from `ConfigFileArgs`. Added documentation explaining why we rely on namespace background deletion instead of CRD-level transformations.

### Updated Documentation

**File**: `apis/dev/planton/provider/kubernetes/kubernetessolroperator/v1/iac/pulumi/README.md`

Updated the "Deletion Behavior" section to clarify that CRDs use default deletion behavior, with a note explaining the cascade through namespace deletion.

## Benefits

1. **Normal diff sizes**: Pulumi operations complete with reasonable diff sizes (KB, not MB)
2. **Faster operations**: No unnecessary state recomputation
3. **Same deletion behavior**: Namespace background deletion provides equivalent cleanup guarantees
4. **Simpler code**: Removed complex transformation logic

## Impact

### Operations

- `pulumi preview` and `pulumi up` operations are dramatically faster
- State management is no longer bloated with unnecessary CRD updates
- Temporal workflow blob size limits are no longer at risk

### Users

- No user-facing changes
- Deletion behavior remains reliable via namespace background deletion

## Technical Notes

### Why ConfigFile Transformations Cause Bloat

Pulumi's `ConfigFile` resource loads YAML manifests and creates child resources. When transformations are applied:

1. The transformation runs on every Pulumi operation
2. Pulumi compares transformed state against stored state
3. Any difference (even annotation additions) triggers an update
4. CRDs with embedded OpenAPI schemas result in massive serialized diffs

### Why Namespace Background Deletion is Sufficient

The deletion cascade works because:

1. Namespace → Operator Pod → Operator stops reconciling
2. Operator stops → CRs no longer recreated
3. CRs deleted → CRD finalizers satisfied
4. CRDs deleted → Clean teardown

This is the same end result as having background deletion on CRDs directly, without the diff bloat side effect.

## Related Work

- **Previous**: [KubernetesSolrOperator Background Deletion Fix](./2025-12-27-103931-kubernetessolroperator-background-deletion-fix.md) - Original fix that introduced the transformation
- **Related**: [KubernetesSolr Background Deletion Fix](./2025-12-27-091641-kubernetessolr-background-deletion-fix.md) - SolrCloud CR deletion fix

---

**Status**: ✅ Production Ready
**Timeline**: Immediate fix after discovering the bloat issue

