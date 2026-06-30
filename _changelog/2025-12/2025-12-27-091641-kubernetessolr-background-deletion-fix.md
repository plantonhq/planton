# KubernetesSolr: Fix Pulumi Destroy Timeout with Background Deletion

**Date**: December 27, 2025
**Type**: Bug Fix
**Components**: Pulumi CLI Integration, Kubernetes Provider, IAC Stack Runner

## Summary

Fixed a critical issue where `pulumi destroy` operations on KubernetesSolr resources would timeout after 10 minutes due to a race condition between the Kubernetes Garbage Collector and the Apache Solr Operator's reconciliation loop. The fix adds the `pulumi.com/deletionPropagationPolicy: "background"` annotation to the SolrCloud custom resource.

## Problem Statement / Motivation

When deploying Apache SolrCloud on Kubernetes via Pulumi and subsequently running `pulumi destroy`, the operation would consistently hang for 10 minutes before timing out with the error:

```
kubernetes:solr.apache.org/v1beta1:SolrCloud solr-cloud deleting (600s) 
  warning: finalizers might be preventing deletion (foregroundDeletion)
  error: timed out waiting for the condition
```

This made reliable `create → destroy → create → destroy` testing cycles impossible without manual intervention.

### Pain Points

- `pulumi destroy` would hang for 10 minutes before failing
- The `foregroundDeletion` finalizer was never removed
- Required manual `kubectl patch` to remove finalizers
- Blocked CI/CD pipelines and automated testing
- Made iterative development extremely slow

### Root Cause Analysis

The issue stemmed from a **Foreground Deletion Race Condition**:

1. Pulumi issues DELETE with default `propagationPolicy: Foreground`
2. Kubernetes adds `foregroundDeletion` finalizer to SolrCloud CR
3. Garbage Collector starts deleting child resources (StatefulSets, Services)
4. **Solr Operator sees SolrCloud still exists and recreates deleted children**
5. GC deletes them again, operator recreates them again
6. This infinite loop continues until the 10-minute timeout

Evidence from operator logs showed continuous recreation:

```
Creating ConfigMap solr-dev-solrcloud-configmap
ERROR configmaps "solr-dev-solrcloud-configmap" already exists
Creating StatefulSet solr-dev-solrcloud
Creating Common Service solr-dev-solrcloud-common
ERROR services "solr-dev-solrcloud-common" already exists
... (repeats continuously)
```

## Solution / What's New

Added the `pulumi.com/deletionPropagationPolicy: "background"` annotation to the SolrCloud resource metadata. This tells Pulumi to use background cascading deletion instead of foreground.

### How Background Deletion Works

With background deletion:
1. Pulumi issues DELETE with `propagationPolicy: Background`
2. SolrCloud CR is removed from the API server **immediately**
3. Solr Operator stops reconciling (the CR it watches is gone)
4. Kubernetes GC cleans up child resources asynchronously
5. Destroy completes in seconds, not minutes

## Implementation Details

### Code Change

**File**: `apis/dev/planton/provider/kubernetes/kubernetessolr/v1/iac/pulumi/module/solr_cloud.go`

Added comprehensive documentation and the annotation:

```go
Metadata: metav1.ObjectMetaArgs{
    Name:      pulumi.String(locals.KubernetesSolr.Metadata.Name),
    Namespace: pulumi.String(locals.Namespace),
    Labels:    pulumi.ToStringMap(locals.Labels),
    // CRITICAL: Background Deletion Propagation Policy
    //
    // This annotation is required to prevent a 10-minute timeout during `pulumi destroy`.
    //
    // Problem: By default, Pulumi uses "Foreground" cascading deletion...
    // Solution: Using "background" propagation policy causes Pulumi to delete the SolrCloud
    // CR immediately...
    //
    // Reference: https://www.pulumi.com/registry/packages/kubernetes/installation-configuration/
    Annotations: pulumi.StringMap{
        "pulumi.com/deletionPropagationPolicy": pulumi.String("background"),
    },
},
```

### Documentation Update

**File**: `apis/dev/planton/provider/kubernetes/kubernetessolr/v1/iac/pulumi/README.md`

Added a new "Deletion Behavior" section explaining:
- Why background deletion is used
- The race condition that occurs without it
- How to test destroy operations
- What to look for if issues occur

## Benefits

- **Reliable Destroys**: `pulumi destroy` now completes in seconds instead of timing out
- **Automated Testing**: Enables reliable `create → destroy → create` cycles for CI/CD
- **No Manual Intervention**: No more `kubectl patch` commands to remove stuck finalizers
- **Documented Knowledge**: The fix and rationale are preserved in code comments and README

## Impact

### Users
- KubernetesSolr deployments can now be cleanly destroyed via Pulumi
- Faster iteration cycles during development and testing

### Developers
- Detailed comments explain the issue for future maintainers
- README documents the deletion behavior pattern

### Operations
- CI/CD pipelines with KubernetesSolr no longer timeout on cleanup
- Reliable infrastructure lifecycle management

## Related Work

This pattern (using `pulumi.com/deletionPropagationPolicy: "background"`) may be applicable to other operator-managed CRDs that exhibit similar race conditions, including:
- Other Kubernetes operators that aggressively reconcile child resources
- Any CRD where the operator doesn't check `deletionTimestamp` before reconciling

Research was conducted using deep analysis of:
- Apache Solr Operator source code (finalizer and reconciliation logic)
- Pulumi Kubernetes Provider documentation
- Kubernetes garbage collection and finalizer mechanics
- Similar patterns in Strimzi Kafka, ECK Elasticsearch, and Argo CD operators

---

**Status**: ✅ Production Ready
**Timeline**: ~2 hours (investigation, research, implementation, testing)

