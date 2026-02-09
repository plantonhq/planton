# Fix: Kubernetes Namespace Dependency Propagation

**Date**: February 9, 2026
**Type**: Bug Fix
**Components**: Pulumi Modules, Terraform Modules (48 Kubernetes components)

## Summary

Fixed a race condition across all 48 Kubernetes components where child resources (Deployments, Services, Secrets, Helm Releases, etc.) could be created before the namespace existed when `create_namespace: true`. The fix adds explicit `pulumi.DependsOn` dependencies from all child resources to the conditionally-created namespace resource.

Additionally, standardized all components on a separate `namespace.go` file pattern for namespace creation, replacing the inconsistent mix of inline creation in `main.go` (~42 components) and helper function patterns (~6 components).

## Problem

When `spec.create_namespace = true`, the namespace resource was created but its return value was **discarded** in every Kubernetes component's Pulumi module:

```go
// Pattern A (~42 components): inline in main.go
_, err = kubernetescorev1.NewNamespace(ctx, locals.Namespace, ...)

// Pattern B (~6 components): namespace() helper
_, err = namespace(ctx, stackInput, locals, kubernetesProvider)
```

Child resources referenced the namespace as a raw string (`pulumi.String(locals.Namespace)`), not as an output of the namespace resource. Pulumi cannot infer dependencies from string literals, so resources could be created in parallel with (or before) the namespace, causing deployment failures.

## Solution

### Pulumi: `pulumi.DependsOn` (Terraform `depends_on` equivalent)

Every component now follows this standardized pattern:

**1. `namespace.go`** -- Dedicated file for conditional namespace creation:

```go
func namespace(ctx *pulumi.Context,
    stackInput *componentv1.ComponentStackInput,
    locals *Locals,
    kubernetesProvider pulumi.ProviderResource,
) (*kubernetescorev1.Namespace, error) {
    if !stackInput.Target.Spec.CreateNamespace {
        return nil, nil
    }
    createdNamespace, err := kubernetescorev1.NewNamespace(ctx, ...)
    return createdNamespace, err
}
```

**2. `main.go`** -- Captures namespace and builds conditional dependency:

```go
createdNamespace, err := namespace(ctx, stackInput, locals, kubernetesProvider)

var namespaceDeps []pulumi.ResourceOption
if createdNamespace != nil {
    namespaceDeps = append(namespaceDeps, pulumi.DependsOn([]pulumi.Resource{createdNamespace}))
}

// Passed to every child resource function
deployment(ctx, locals, kubernetesProvider, namespaceDeps)
```

**3. Child resource functions** -- Accept and apply dependency:

```go
func deployment(ctx *pulumi.Context, locals *Locals,
    kubernetesProvider pulumi.ProviderResource, namespaceDeps []pulumi.ResourceOption) error {
    opts := append([]pulumi.ResourceOption{pulumi.Provider(kubernetesProvider)}, namespaceDeps...)
    _, err := appsv1.NewDeployment(ctx, name, &args, opts...)
}
```

### Terraform: `depends_on` audit

Audited all `.tf` files and added `depends_on = [kubernetes_namespace.this]` to resources that were missing it. Terraform safely treats `depends_on` on a `count = 0` resource as a no-op.

### Why `pulumi.DependsOn` over `pulumi.Parent`

- Direct Terraform analog -- ops engineers immediately understand it
- No URN changes -- `pulumi.Parent` would change resource URNs, forcing replacement of existing resources in live stacks
- Safe for existing deployments -- only affects ordering, not resource identity

## Components Updated (48)

kubernetesaltinityoperator, kubernetesargocd, kubernetescertmanager, kubernetesclickhouse, kubernetescronjob, kubernetesdaemonset, kubernetesdeployment, kuberneteselasticoperator, kuberneteselasticsearch, kubernetesexternaldns, kubernetesexternalsecrets, kubernetesgharunnerscaleset, kubernetesgharunnerscalesetcontroller, kubernetesgitlab, kubernetesgrafana, kubernetesharbor, kuberneteshelmrelease, kubernetesingressnginx, kubernetesistio, kubernetesjenkins, kubernetesjob, kuberneteskafka, kuberneteskeycloak, kuberneteslocust, kubernetesmanifest, kubernetesmongodb, kubernetesnats, kubernetesneo4j, kubernetesopenbao, kubernetesopenfga, kubernetesperconamongooperator, kubernetesperconamysqloperator, kubernetesperconapostgresoperator, kubernetespostgres, kubernetesprometheus, kubernetesredis, kubernetesrookcephcluster, kubernetesrookcephoperator, kubernetessignoz, kubernetessolr, kubernetessolroperator, kubernetesstatefulset, kubernetesstrimzikafkaoperator, kubernetestekton, kubernetestektonoperator, kubernetestemporal, kuberneteszalandopostgresoperator

**Excluded**: `kubernetesnamespace` (creates namespaces by definition), `kubernetesgatewayapicrds` (cluster-scoped CRDs, no namespace field)

## Files Changed

Per component (typical):
- **New**: `iac/pulumi/module/namespace.go` (for ~42 components that had inline creation)
- **Modified**: `iac/pulumi/module/main.go` (all 48 components)
- **Modified**: `iac/pulumi/module/<resource>.go` (child resource files, varies per component)
- **Modified**: `iac/tf/*.tf` (Terraform files missing `depends_on`)

**Estimated total**: ~200+ files across 48 components

## Validation

- All 49 Kubernetes component packages build: `go build ./apis/org/openmcf/provider/kubernetes/...`
- All 49 Kubernetes component packages pass tests: `go test ./apis/org/openmcf/provider/kubernetes/...`
- Zero compilation errors, zero test failures

## Breaking Changes

None. Adding `pulumi.DependsOn` only affects resource creation ordering, not resource identity or URNs. Existing Pulumi stacks will see no state changes or resource replacements.

## Related Work

- [Namespace Creation Control (Dec 16, 2025)](../2025-12/2025-12-16-184915-kubernetes-components-namespace-creation-control.md) -- introduced `create_namespace` field
- [Namespace Standardization (Nov 23, 2025)](../2025-11/2025-11-23-220641-standardize-kubernetes-components-target-cluster-namespace.md) -- standardized namespace field across components
