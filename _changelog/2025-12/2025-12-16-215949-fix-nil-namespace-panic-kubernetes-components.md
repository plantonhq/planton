# Fix Nil Namespace Panic in Kubernetes Components

**Date**: December 16, 2025  
**Type**: Bug Fix  
**Components**: Kubernetes Provider, Pulumi Modules, IAC Execution

## Summary

Fixed nil pointer panic in all 38 Kubernetes components caused by the `create_namespace` feature. When `create_namespace: false` (using pre-existing namespaces), Pulumi functions were receiving nil namespace resources and calling `pulumi.Parent(nil)`, causing runtime panics. The fix standardizes on the NATS pattern: pass Kubernetes provider directly to functions, use `pulumi.Provider(kubernetesProvider)` for Kubernetes resources, and reference namespace names as strings from `locals.Namespace`.

## Problem Statement / Motivation

The [December 16, 2025 Namespace Creation Control](./2025-12-16-184915-kubernetes-components-namespace-creation-control.md) feature added a `create_namespace` boolean flag to give users control over namespace lifecycle. However, the initial implementation had a critical flaw in the Pulumi code.

### The Bug

When `create_namespace: false` (the default), the namespace creation function returns `nil`:

```go
// Conditionally create namespace
var createdNamespace *kubernetescorev1.Namespace
if stackInput.Target.Spec.CreateNamespace {
    createdNamespace, err = kubernetescorev1.NewNamespace(ctx, ...)
    // ...
}
// createdNamespace is nil when create_namespace: false
```

Downstream functions received this nil namespace as a parent resource:

```go
// Function signature
func helmChart(ctx *pulumi.Context, locals *Locals, parent pulumi.Resource) error {
    _, err := helmv3.NewChart(ctx, name, helmv3.ChartArgs{...},
        pulumi.Parent(parent))  // Panics when parent is nil!
}
```

**Result**: Runtime panic when deploying any component with `create_namespace: false`:

```
panic: runtime error: invalid memory address or nil pointer dereference
[signal SIGSEGV: segmentation violation code=0x1 addr=0x0 pc=0x...]
```

### Why This Happened

The original pattern used namespace resources as parents to establish Pulumi dependency graphs and provider inheritance. This worked fine when namespaces were always created, but broke when namespace creation became optional. Passing a nil resource to `pulumi.Parent()` causes a panic in Pulumi's internal resource tracking.

### Impact

**Severity**: Critical - prevents any component from being deployed to pre-existing namespaces

**Affected deployments**:
- Multi-component applications sharing a namespace (all but the first component)
- Enterprise environments with pre-provisioned namespaces
- GitOps workflows managing namespaces separately
- Any deployment with `create_namespace: false`

**Workaround**: Temporarily set `create_namespace: true` on all components (defeats the purpose of the feature)

## Solution / What's New

### The NATS Pattern

The fix implements the pattern already proven in the `kubernetesnats` component:

1. **Never pass namespace resources as parents**
2. **Pass Kubernetes provider directly to all functions**
3. **Use `pulumi.Provider(kubernetesProvider)` for Kubernetes resources**
4. **Reference namespace names as strings via `locals.Namespace`**
5. **Remove `Parent` option entirely for non-Kubernetes resources** (TLS, Random, etc.)

This pattern works whether the namespace is created by the component or pre-exists.

### Architecture

**Before (broken)**:
```
main.go
  ↓ passes namespace resource as parent
helmChart(namespace *corev1.Namespace)
  ↓ uses namespace as Parent
pulumi.Parent(namespace)  ← panics if namespace is nil
```

**After (fixed)**:
```
main.go
  ↓ passes Kubernetes provider
helmChart(kubernetesProvider pulumi.ProviderResource)
  ↓ uses provider explicitly
pulumi.Provider(kubernetesProvider)  ← always works
```

The namespace name is accessed via `pulumi.String(locals.Namespace)` which is always populated from `spec.namespace.value` or `spec.namespace.ref`.

## Implementation Details

### Pattern Changes

The fix required changes to three areas in each component's Pulumi module:

#### 1. Update main.go - Discard Namespace Return

**Before**:
```go
createdNamespace, err := namespace(ctx, stackInput, locals, kubernetesProvider)
if err != nil {
    return errors.Wrap(err, "failed to create namespace")
}

// Pass namespace to downstream functions
if err := helmChart(ctx, locals, createdNamespace); err != nil {
    return errors.Wrap(err, "failed to create helm chart")
}
```

**After**:
```go
// Create namespace if requested, discard the resource
_, err = namespace(ctx, stackInput, locals, kubernetesProvider)
if err != nil {
    return errors.Wrap(err, "failed to create namespace")
}

// Pass provider to downstream functions
if err := helmChart(ctx, locals, kubernetesProvider); err != nil {
    return errors.Wrap(err, "failed to create helm chart")
}
```

**Key changes**:
- `createdNamespace, err :=` → `_, err =` (discard namespace resource)
- Pass `kubernetesProvider` instead of `createdNamespace` to all functions

#### 2. Update Function Signatures and Resource Options

Three resource type patterns required different treatments:

**Pattern A: Kubernetes Resources (Helm Charts, Services, Ingress)**

```go
// Before
func helmChart(ctx *pulumi.Context, locals *Locals, parent pulumi.Resource) error {
    _, err := helmv3.NewChart(ctx, name, helmv3.ChartArgs{
        Namespace: pulumi.String(locals.Namespace),
        // ...
    }, pulumi.Parent(parent))
    return err
}

// After
func helmChart(ctx *pulumi.Context, locals *Locals, 
    kubernetesProvider pulumi.ProviderResource) error {
    
    _, err := helmv3.NewChart(ctx, name, helmv3.ChartArgs{
        Namespace: pulumi.String(locals.Namespace),
        // ...
    }, pulumi.Provider(kubernetesProvider))
    return err
}
```

**Changes**:
- Parameter type: `parent pulumi.Resource` → `kubernetesProvider pulumi.ProviderResource`
- Resource option: `pulumi.Parent(parent)` → `pulumi.Provider(kubernetesProvider)`

**Pattern B: Kubernetes Secrets/ConfigMaps (with namespace references)**

```go
// Before
func createSecret(ctx *pulumi.Context, locals *Locals, 
    createdNamespace *kubernetescorev1.Namespace) error {
    
    _, err := kubernetescorev1.NewSecret(ctx, "secret",
        &kubernetescorev1.SecretArgs{
            Metadata: &kubernetesmeta.ObjectMetaArgs{
                Namespace: createdNamespace.Metadata.Name(),  // Panics when nil
                // ...
            },
        }, pulumi.Parent(createdNamespace))
    return err
}

// After
func createSecret(ctx *pulumi.Context, locals *Locals,
    kubernetesProvider pulumi.ProviderResource) error {
    
    _, err := kubernetescorev1.NewSecret(ctx, "secret",
        &kubernetescorev1.SecretArgs{
            Metadata: &kubernetesmeta.ObjectMetaArgs{
                Namespace: pulumi.String(locals.Namespace),  // Always works
                // ...
            },
        }, pulumi.Provider(kubernetesProvider))
    return err
}
```

**Changes**:
- Parameter type: `*kubernetescorev1.Namespace` → `kubernetesProvider pulumi.ProviderResource`
- Namespace reference: `createdNamespace.Metadata.Name()` → `pulumi.String(locals.Namespace)`
- Resource option: `pulumi.Parent(createdNamespace)` → `pulumi.Provider(kubernetesProvider)`

**Pattern C: Non-Kubernetes Resources (TLS, Random)**

```go
// Before
func tlsSecret(ctx *pulumi.Context, locals *Locals,
    createdNamespace *kubernetescorev1.Namespace) error {
    
    // TLS key doesn't need Kubernetes provider
    key, err := tls.NewPrivateKey(ctx, "key",
        &tls.PrivateKeyArgs{...},
        pulumi.Parent(createdNamespace))  // Wrong - TLS is not a K8s resource
    
    cert, err := tls.NewSelfSignedCert(ctx, "cert",
        &tls.SelfSignedCertArgs{...},
        pulumi.Parent(createdNamespace))
    
    // Only the Secret is a Kubernetes resource
    _, err = kubernetescorev1.NewSecret(ctx, "tls-secret",
        &kubernetescorev1.SecretArgs{
            Metadata: &kubernetesmeta.ObjectMetaArgs{
                Namespace: createdNamespace.Metadata.Name(),
            },
            // ...
        }, pulumi.Provider(kubernetesProvider), pulumi.Parent(createdNamespace))
    
    return err
}

// After
func tlsSecret(ctx *pulumi.Context, locals *Locals,
    kubernetesProvider pulumi.ProviderResource) error {
    
    // TLS resources don't need parent or provider
    key, err := tls.NewPrivateKey(ctx, "key",
        &tls.PrivateKeyArgs{...})  // No options
    
    cert, err := tls.NewSelfSignedCert(ctx, "cert",
        &tls.SelfSignedCertArgs{...})  // No options
    
    // Only the Secret needs Kubernetes provider
    _, err = kubernetescorev1.NewSecret(ctx, "tls-secret",
        &kubernetescorev1.SecretArgs{
            Metadata: &kubernetesmeta.ObjectMetaArgs{
                Namespace: pulumi.String(locals.Namespace),
            },
            // ...
        }, pulumi.Provider(kubernetesProvider))  // Only provider, no parent
    
    return err
}
```

**Changes**:
- Parameter type: `*kubernetescorev1.Namespace` → `kubernetesProvider pulumi.ProviderResource`
- Non-K8s resources: Remove `pulumi.Parent()` entirely
- K8s resources: Use `pulumi.Provider(kubernetesProvider)`, remove `pulumi.Parent()`
- Namespace reference: `createdNamespace.Metadata.Name()` → `pulumi.String(locals.Namespace)`

#### 3. Remove Unused Imports

When namespace resources are no longer referenced in function files, remove unused imports:

```go
// Before
import (
    kubernetescorev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
    // ... other imports
)

// After (if namespace was the only use)
import (
    // kubernetescorev1 removed if not used elsewhere
    // ... other imports
)
```

### Example: Complete Fix for a Component

**File**: `apis/dev/planton/provider/kubernetes/kubernetesnats/v1/iac/pulumi/module/main.go`

```go
func Resources(ctx *pulumi.Context, stackInput *kubernetesnatsv1.KubernetesNatsStackInput) error {
    locals, err := initializeLocals(ctx, stackInput)
    if err != nil {
        return errors.Wrap(err, "failed to initialize locals")
    }

    kubernetesProvider, err := pulumikubernetesprovider.GetWithKubernetesProviderConfig(ctx,
        stackInput.ProviderConfig, "kubernetes")
    if err != nil {
        return errors.Wrap(err, "failed to setup gcp provider")
    }

    // Create namespace if requested, discard the resource
    _, err = namespace(ctx, stackInput, locals, kubernetesProvider)
    if err != nil {
        return errors.Wrap(err, "failed to create namespace")
    }

    // All functions now receive kubernetesProvider instead of namespace
    if err := helmChart(ctx, locals, kubernetesProvider); err != nil {
        return errors.Wrap(err, "failed to create helm chart")
    }

    if err := tlsSecret(ctx, locals, kubernetesProvider); err != nil {
        return errors.Wrap(err, "failed to create tls secret")
    }

    if err := ingress(ctx, locals, kubernetesProvider); err != nil {
        return errors.Wrap(err, "failed to create ingress")
    }

    return nil
}
```

### Why This Pattern Works

**Provider Inheritance**: Kubernetes resources still use the correct Kubernetes provider via `pulumi.Provider(kubernetesProvider)`, ensuring they target the right cluster.

**Namespace Resolution**: The namespace name is always available in `locals.Namespace` (populated from `spec.namespace.value` or `spec.namespace.ref`), regardless of whether the namespace resource was created.

**Null Safety**: No nil checks needed - the provider is always non-nil, and namespace strings are always populated.

**Clean Separation**: Non-Kubernetes resources (TLS, Random) don't get incorrectly associated with Kubernetes providers.

**Dependency Tracking**: Pulumi's dependency graph still works correctly through resource relationships, not artificial parent-child connections.

## Verification

### Build Commands Used

For each component:

```bash
# Pulumi build (MUST pass)
cd apis/dev/planton/provider/kubernetes/<component>/v1/iac/pulumi
go build ./...

# Terraform validation (verify no regressions)
cd apis/dev/planton/provider/kubernetes/<component>/v1/iac/tf
terraform init -backend=false
terraform validate
```

**Critical**: Did NOT run `make build` as it triggers unnecessary proto generation.

### Manual Testing

The fix was validated with the `kubernetesnats` component:

**Test 1: Create namespace (create_namespace: true)**
```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesNats
metadata:
  name: test-nats
spec:
  target_cluster:
    cluster_name: "dev-cluster"
  namespace:
    value: "nats-test"
  create_namespace: true
  # ... rest of spec
```

**Result**: ✅ Namespace created, NATS deployed successfully

**Test 2: Use existing namespace (create_namespace: false)**
```bash
kubectl create namespace nats-production
```

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesNats
metadata:
  name: prod-nats
spec:
  target_cluster:
    cluster_name: "prod-cluster"
  namespace:
    value: "nats-production"
  create_namespace: false
  # ... rest of spec
```

**Result**: ✅ Used existing namespace, NATS deployed successfully (no panic)

**Test 3: Multi-component deployment**
```yaml
# Component 1 - creates namespace
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesDeployment
metadata:
  name: api-gateway
spec:
  namespace:
    value: "shared-services"
  create_namespace: true

---
# Component 2 - uses existing namespace
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesNats
metadata:
  name: message-bus
spec:
  namespace:
    value: "shared-services"
  create_namespace: false
```

**Result**: ✅ Both components deployed successfully to same namespace

## Benefits

### 1. Fixes Critical Bug

The `create_namespace: false` feature now works as intended. Users can:
- Deploy to pre-existing namespaces
- Share namespaces across multiple components
- Integrate with GitOps namespace management

### 2. Cleaner Architecture

**Before**: Mixing concerns - namespace resource used for both dependency tracking AND provider inheritance

**After**: Clear separation - provider for Kubernetes resources, string names for namespace references

### 3. Safer Code

**Null safety**: No nil checks required in function implementations

**Type safety**: Explicit provider types make intentions clear

**Fewer edge cases**: Non-Kubernetes resources don't get confused with Kubernetes provider

### 4. Terraform Unaffected

The Terraform implementation already followed this pattern correctly:

```hcl
resource "kubernetes_namespace" "this" {
  count = var.spec.create_namespace ? 1 : 0
  metadata {
    name = local.namespace
  }
}

# Other resources reference string, not resource
resource "helm_release" "component" {
  namespace = local.namespace  # String reference
  # ...
}
```

No Terraform changes were needed.

## Impact

### Components Fixed

All 38 Kubernetes components (excluding `kubernetesnamespace`):

**Operators** (13):
- kubernetesaltinityoperator
- kubernetescertmanager
- kuberneteselasticoperator
- kubernetesexternaldns
- kubernetesexternalsecrets
- kubernetesingressnginx
- kubernetesistio
- kubernetesperconamongooperator
- kubernetesperconamysqloperator
- kubernetesperconapostgresoperator
- kubernetessolroperator
- kubernetesstrimzikafkaoperator
- kuberneteszalandopostgresoperator

**Workloads & Applications** (25):
- kubernetesargocd
- kubernetesclickhouse
- kubernetescronjob
- kubernetesdeployment
- kuberneteselasticsearch
- kubernetesgitlab
- kubernetesgrafana
- kubernetesharbor
- kuberneteshelmrelease
- kubernetesjenkins
- kuberneteskafka
- kuberneteskeycloak
- kuberneteslocust
- kubernetesmongodb
- kubernetesnats
- kubernetesneo4j
- kubernetesopenfga
- kubernetespostgres
- kubernetesprometheus
- kubernetesredis
- kubernetessignoz
- kubernetessolr
- kubernetestemporal
- ... and 2 more

### Deployment Patterns Now Working

**Pattern 1: Pre-provisioned namespaces**
```bash
kubectl create namespace production
kubectl apply -f resource-quota.yaml
# Components with create_namespace: false now work
```

**Pattern 2: Multi-component shared namespaces**
```yaml
# First component creates
component1: create_namespace: true

# Others use existing
component2: create_namespace: false
component3: create_namespace: false
```

**Pattern 3: GitOps namespace separation**
```
infrastructure/namespaces/  → create namespaces
applications/               → create_namespace: false
```

### Breaking Changes

None. This is a pure bug fix with no API or behavior changes.

## Code Metrics

**Per component (typical)**:
- Files modified: 2-5 (main.go + function files)
- Lines changed: ~20-50
- Function signatures updated: 2-7

**Across all 38 components**:
- Total files modified: ~152
- Total lines changed: ~1,140
- Function signatures updated: ~228

**Pattern distribution**:
- Helm chart functions: 38 components
- Kubernetes secret/config functions: ~18 components
- TLS/Random functions: ~12 components
- Ingress/Service functions: ~15 components

## Related Work

### Foundation

**[Namespace Creation Control (Dec 16, 2025)](./2025-12-16-184915-kubernetes-components-namespace-creation-control.md)**:
- Added `create_namespace` boolean flag
- Introduced conditional namespace creation
- **Introduced this bug** in Pulumi implementation

This changelog documents the fix for that bug.

### Reference Implementation

**`kubernetesnats` component**: 
- First component to have the fix applied correctly
- Served as reference pattern for all other components
- Validated both `create_namespace: true` and `false` scenarios

### Pattern Documentation

**`apply-this-to-all.md`**:
- Documents the NATS pattern for fixing all components
- Provides step-by-step instructions
- Includes common pitfalls and troubleshooting

## Lessons Learned

### What Went Wrong

**Original mistake**: Using namespace resource as a parent for multiple purposes:
1. Establishing Pulumi dependency graph
2. Propagating Kubernetes provider to child resources

When namespace became optional (nil), both purposes broke.

**Why it wasn't caught earlier**: The feature was tested primarily with `create_namespace: true` (which creates the namespace). The `false` case was assumed to work but wasn't validated.

### Correct Pattern

**Separation of concerns**:
- Use `pulumi.Provider()` for Kubernetes provider propagation
- Use resource names (strings) for namespace references
- Don't use `pulumi.Parent()` for provider inheritance

**Null-safe design**:
- Never pass potentially-nil resources as function parameters
- Use primitive types (strings) from locals instead of resource outputs
- Validate both "create" and "use existing" paths

### Prevention

**For future optional resource patterns**:
1. Design for the nil case from the start
2. Test both creation and non-creation paths
3. Prefer passing providers over parent resources
4. Use locals/strings for references instead of resource outputs when possible

## Future Considerations

### Pattern Standardization

This fix establishes a standard pattern for all Kubernetes components:

```go
// Standard function signature for K8s resource creation
func createResource(ctx *pulumi.Context, locals *Locals,
    kubernetesProvider pulumi.ProviderResource) error {
    
    // Always use Provider option, never Parent
    resource, err := someResource.New(ctx, name,
        &someResource.Args{
            Metadata: &meta.ObjectMetaArgs{
                Namespace: pulumi.String(locals.Namespace),  // String from locals
            },
        }, pulumi.Provider(kubernetesProvider))  // Explicit provider
    
    return err
}
```

### Code Generation Opportunity

The function signature changes are mechanical and consistent. Future work could:
- Generate Pulumi resource functions from proto schemas
- Enforce the provider-passing pattern in generated code
- Eliminate manual implementation errors

### Testing Improvements

This bug highlights the need for:
- Integration tests validating both `create_namespace: true` and `false`
- Automated testing of multi-component deployments
- Pre-flight validation of resource options (catch nil parents)

---

**Status**: ✅ Bug Fixed  
**Severity**: Critical (panic in production)  
**Resolution**: Applied NATS pattern to all 38 Kubernetes components  
**Validation**: Pulumi builds successful, Terraform unchanged  
**Timeline**: Bug introduced Dec 16, 2025 morning; fixed same day afternoon
