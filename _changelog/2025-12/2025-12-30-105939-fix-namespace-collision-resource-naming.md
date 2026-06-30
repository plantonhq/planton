# Fix KubernetesDeployment Resource Naming for Multi-Deployment Namespaces

**Date**: December 30, 2025
**Type**: Bug Fix
**Components**: Kubernetes Provider, Pulumi CLI Integration, IAC Stack Runner

## Summary

Updated the KubernetesDeployment Pulumi and Terraform modules to use `metadata.name` for resource naming instead of hardcoded values like `spec.version`. This prevents resource name conflicts when multiple KubernetesDeployment resources are deployed to the same namespace.

## Problem Statement / Motivation

When deploying multiple `KubernetesDeployment` resources to the same Kubernetes namespace, several resources were using hardcoded or `spec.version`-based names that caused collisions:

### Pain Points

- **Service name collisions**: Services used `spec.version` (e.g., "main") as the name, causing conflicts when multiple deployments had the same version
- **Traffic routing to wrong pods**: Selector labels were hardcoded to `app: microservice`, meaning all deployments in a namespace would match the same selector
- **ConfigMap name collisions**: ConfigMaps used user-provided keys directly without namespacing
- **Pulumi state conflicts**: Resource IDs used hardcoded values, causing state management issues

The root cause was documented in the code itself:

```go
// Static selector labels that never change
// Since there's always only one deployment per namespace, we use a constant label
locals.SelectorLabels = map[string]string{
    "app": "microservice",
}
```

This assumption no longer holds when users need multiple deployments per namespace.

## Solution / What's New

Updated all resource names to be prefixed or derived from `metadata.name`, ensuring uniqueness across deployments in the same namespace.

### Changes Summary

| Resource | Before | After |
|----------|--------|-------|
| Selector Labels | `app: microservice` | `app: {metadata.name}` |
| Service Name | `spec.version` | `metadata.name` |
| ConfigMap Names | `{key}` | `{metadata.name}-{key}` |
| Pulumi Resource IDs | Various hardcoded | `metadata.name` based |

## Implementation Details

### Pulumi Module Changes (Go)

**`locals.go`** - Updated selector labels and service name:

```go
// Before
locals.SelectorLabels = map[string]string{
    "app": "microservice",
}
locals.KubeServiceName = target.Spec.Version

// After
locals.SelectorLabels = map[string]string{
    "app":                            target.Metadata.Name,
    kuberneteslabelkeys.ResourceName: target.Metadata.Name,
}
locals.KubeServiceName = target.Metadata.Name
```

**`service.go`** - Updated Service to use computed name:

```go
// Before
Name: pulumi.String(locals.KubernetesDeployment.Spec.Version)

// After
Name: pulumi.String(locals.KubeServiceName)
```

**`deployment.go`** - Updated Pulumi resource ID:

```go
// Before
createdDeployment, err := appsv1.NewDeployment(ctx,
    locals.KubernetesDeployment.Spec.Version, ...)

// After
createdDeployment, err := appsv1.NewDeployment(ctx,
    locals.KubernetesDeployment.Metadata.Name, ...)
```

**`configmap.go`** - Added name prefixing:

```go
configMapName := fmt.Sprintf("%s-%s", locals.KubernetesDeployment.Metadata.Name, name)
```

**`pdb.go`** - Updated PDB resource ID:

```go
pdbResourceName := fmt.Sprintf("%s-pdb", locals.KubernetesDeployment.Metadata.Name)
```

### Terraform Module Changes

**`locals.tf`** - Added selector labels and updated service name:

```hcl
selector_labels = {
  "app"           = var.metadata.name
  "resource_name" = var.metadata.name
}

kube_service_name = var.metadata.name
```

**`service.tf`** - Updated to use computed values:

```hcl
name     = local.kube_service_name
selector = local.selector_labels
```

**`deployment.tf`** - Updated selector to use new labels:

```hcl
selector {
  match_labels = local.selector_labels
}
```

**`configmap.tf`** - Added name prefixing:

```hcl
name = "${var.metadata.name}-${each.key}"
```

## Benefits

- **Multi-deployment support**: Users can now deploy multiple KubernetesDeployment resources to the same namespace without conflicts
- **Correct traffic routing**: Each deployment's pods are uniquely identified by their selector labels
- **Predictable naming**: Resource names are derived from `metadata.name`, making them predictable and debuggable
- **Pulumi state isolation**: Each deployment has unique resource IDs in Pulumi state

## Impact

### Who is affected

- Users deploying multiple microservices to the same namespace
- Existing deployments will see resource name changes on next apply

### Breaking Changes

This is a **breaking change** for existing deployments:

1. **Service names will change**: DNS references using the old service name will break
2. **Selector labels change**: Pods will be re-selected under new labels
3. **ConfigMap names change**: References to old ConfigMap names will fail

### Migration

For existing deployments, users should:

1. Update any hardcoded service DNS references to use the new naming pattern
2. Plan for a brief traffic disruption during the selector label transition
3. Update volume mount references if ConfigMaps are mounted by name

## Files Changed

**Pulumi Module (Go):**
- `apis/dev/planton/provider/kubernetes/kubernetesdeployment/v1/iac/pulumi/module/locals.go`
- `apis/dev/planton/provider/kubernetes/kubernetesdeployment/v1/iac/pulumi/module/service.go`
- `apis/dev/planton/provider/kubernetes/kubernetesdeployment/v1/iac/pulumi/module/deployment.go`
- `apis/dev/planton/provider/kubernetes/kubernetesdeployment/v1/iac/pulumi/module/configmap.go`
- `apis/dev/planton/provider/kubernetes/kubernetesdeployment/v1/iac/pulumi/module/pdb.go`

**Terraform Module:**
- `apis/dev/planton/provider/kubernetes/kubernetesdeployment/v1/iac/tf/locals.tf`
- `apis/dev/planton/provider/kubernetes/kubernetesdeployment/v1/iac/tf/service.tf`
- `apis/dev/planton/provider/kubernetes/kubernetesdeployment/v1/iac/tf/deployment.tf`
- `apis/dev/planton/provider/kubernetes/kubernetesdeployment/v1/iac/tf/configmap.tf`

## Related Work

This change enables the multi-microservice deployment pattern where teams deploy several related services to a shared namespace for easier service discovery and network policies.

---

**Status**: ✅ Production Ready
**Timeline**: Single session implementation

