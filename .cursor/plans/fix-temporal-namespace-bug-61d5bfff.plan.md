<!-- 61d5bfff-f44d-425a-bfc7-e8e250980cab 8d6329eb-c63a-4fe3-9505-28b10b276f85 -->
# Fix Temporal Kubernetes Namespace Initialization Bug

## Problem

The TemporalKubernetes Pulumi module fails with `"failed to create namespace: resource name argument (for URN creation) cannot be empty"` because the namespace initialization logic incorrectly uses an `else` clause that sets the namespace to an empty `stackInput.KubernetesNamespace` value when no custom namespace label is present.

## Root Cause

In `locals.go` (lines 56-64), the code uses:

```go
locals.Namespace = target.Metadata.Name
if target.Metadata.Labels != nil && target.Metadata.Labels[kuberneteslabels.NamespaceLabelKey] != "" {
    locals.Namespace = target.Metadata.Labels[kuberneteslabels.NamespaceLabelKey]
} else {
    locals.Namespace = stackInput.KubernetesNamespace  // Empty value overwrites valid default!
}
```

## Solution

Change the `else` to a separate `if` statement with a non-empty check, matching the pattern used by PostgresKubernetes, MicroserviceKubernetes, and all other workload resources.

## Files to Modify

- `/Users/suresh/scm/github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/workload/temporalkubernetes/v1/iac/pulumi/module/locals.go`

## Changes

### Update namespace initialization logic (lines 56-64)

**Before:**

```go
locals.Namespace = target.Metadata.Name
if target.Metadata.Labels != nil &&
    target.Metadata.Labels[kuberneteslabels.NamespaceLabelKey] != "" {
    locals.Namespace = target.Metadata.Labels[kuberneteslabels.NamespaceLabelKey]
} else {
    locals.Namespace = stackInput.KubernetesNamespace
}
```

**After:**

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

## Verification

After applying the fix:

1. Run `planton pulumi preview --manifest temporal-test.yaml --module-dir ${TEMPORAL_MODULE}` again
2. Verify that the namespace is now correctly set to `temporal-test` (from `metadata.name`)
3. Confirm the preview succeeds without namespace errors