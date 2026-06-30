# Namespace Creation Control for Kubernetes Components

**Date**: December 16, 2025
**Type**: Feature | Enhancement
**Components**: API Definitions, Kubernetes Provider, Pulumi Modules, Terraform Modules, Documentation

## Summary

Added `create_namespace` boolean flag to all 38 Kubernetes components, giving users explicit control over namespace creation. Components now support two namespace management modes: automatic namespace creation or using pre-existing namespaces. This enhancement builds on the recent namespace standardization work and provides critical flexibility for multi-component deployments and enterprise environments with strict namespace lifecycle management.

## Problem Statement / Motivation

Following the [November 23, 2025 namespace standardization](../2025-11/2025-11-23-220641-standardize-kubernetes-components-target-cluster-namespace.md), all Kubernetes components began requiring a `namespace` field and automatically creating that namespace during deployment. While this simplified single-component deployments, it created challenges in real-world scenarios:

### Pain Points

- **Multi-component deployments**: When deploying multiple components to the same namespace (e.g., multiple workloads in `backend-services`), all components would attempt to create the same namespace, causing conflicts and deployment failures
- **Pre-existing namespace management**: Organizations with existing namespace management practices (GitOps, separate namespace provisioning pipelines) couldn't leverage Planton components without workarounds
- **Namespace ownership ambiguity**: Unclear which component "owns" a shared namespace, complicating lifecycle management and deletion
- **Resource policy enforcement**: Namespaces with pre-configured ResourceQuotas, NetworkPolicies, or LimitRanges couldn't be used by Planton components
- **Compliance and security**: Enterprises with namespace creation restricted to platform teams needed a way to deploy components without namespace creation privileges
- **Dependency ordering**: No way to ensure namespace is created first with proper labels/annotations before deploying components

### Real-World Example

Consider deploying three microservices to a single `ecommerce-backend` namespace:

```yaml
# cart-service - first deployment succeeds
spec:
  namespace:
    value: "ecommerce-backend"
  # implicitly creates namespace

# order-service - fails or conflicts
spec:
  namespace:
    value: "ecommerce-backend"
  # tries to create same namespace!

# payment-service - fails or conflicts  
spec:
  namespace:
    value: "ecommerce-backend"
  # tries to create same namespace!
```

**Result**: Deployment failures, race conditions, or namespace ownership confusion.

## Solution / What's New

Introduced a `create_namespace` boolean field in all Kubernetes component specs, enabling two distinct namespace management modes:

### Namespace Management Modes

**Mode 1: Automatic Creation (`create_namespace: true`)**
- Component creates the namespace if it doesn't exist
- Namespace is labeled with component metadata (resource_id, resource_kind, organization, environment)
- Namespace becomes a child resource of the component in the Pulumi/Terraform dependency graph
- Ideal for single-component namespaces or first component in a shared namespace

**Mode 2: Use Existing (`create_namespace: false`)**
- Component assumes namespace already exists
- No namespace creation attempted
- Component resources reference the namespace name directly
- Ideal for shared namespaces, pre-provisioned namespaces, or multi-component deployments

### Proto Schema Addition

Added to all 38 Kubernetes component specs (example shown for `KubernetesElasticOperator`):

```protobuf
message KubernetesElasticOperatorSpec {
  // Target Kubernetes Cluster
  dev.planton.provider.kubernetes.KubernetesClusterSelector target_cluster = 1;

  // Kubernetes Namespace
  dev.planton.shared.foreignkey.v1.StringValueOrRef namespace = 2 [
    (buf.validate.field).required = true,
    (dev.planton.shared.foreignkey.v1.default_kind) = KubernetesNamespace,
    (dev.planton.shared.foreignkey.v1.default_kind_field_path) = "spec.name"
  ];

  // flag to indicate if the namespace should be created
  bool create_namespace = 3;

  // ... rest of component-specific fields
}
```

**Key characteristics**:
- Field position: Always field #3 (after `target_cluster` and `namespace`)
- Type: `bool` (defaults to `false` in proto3)
- No validation annotations (optional flag)
- Consistent naming across all components

### Behavior Matrix

| create_namespace | Namespace Exists? | Behavior |
|-----------------|-------------------|----------|
| `true` | No | ✅ Creates namespace with component labels |
| `true` | Yes | ✅ Uses existing (or overwrites - provider dependent) |
| `false` | Yes | ✅ Uses existing namespace |
| `false` | No | ❌ Deployment fails (namespace not found) |

## Implementation Details

### Phase 1: Proto Schema Changes

**Date**: December 16, 2025 (commit `82e92d3b`)  
**Files changed**: 108 files (spec.proto, generated Go stubs, generated TypeScript stubs)

Added `bool create_namespace = 3;` field to all Kubernetes component specs:

**Components updated** (38 total):
- kubernetesaltinityoperator
- kubernetesargocd
- kubernetescertmanager
- kubernetesclickhouse
- kubernetescronjob
- kubernetesdeployment
- kuberneteselasticoperator
- kuberneteselasticsearch
- kubernetesexternaldns
- kubernetesexternalsecrets
- kubernetesgitlab
- kubernetesgrafana
- kubernetesharbor
- kuberneteshelmrelease
- kubernetesingressnginx
- kubernetesistio
- kubernetesjenkins
- kuberneteskafka
- kuberneteskeycloak
- kuberneteslocust
- kubernetesmongodb
- kubernetesnats
- kubernetesneo4j
- kubernetesopenfga
- kubernetesperconamongooperator
- kubernetesperconamysqloperator
- kubernetesperconapostgresoperator
- kubernetespostgres
- kubernetesprometheus
- kubernetesredis
- kubernetessignoz
- kubernetessolr
- kubernetessolroperator
- kubernetesstrimzikafkaoperator
- kubernetestemporal
- kuberneteszalandopostgresoperator
- ... and 2 more

**Note**: `kubernetesnamespace` component itself was NOT modified (it creates namespaces by definition).

**Code generation**: Protobuf stubs regenerated for both Go and TypeScript.

### Phase 2: Pulumi Module Implementation

Updated all Pulumi modules to implement conditional namespace creation logic.

#### Pattern 1: Component-Specific Resources (e.g., KubernetesDeployment)

Used for components that create multiple Kubernetes resources that depend on the namespace.

**File**: `iac/pulumi/module/main.go`

```go
func Resources(ctx *pulumi.Context, stackInput *kubernetesdeploymentv1.KubernetesDeploymentStackInput) error {
	locals, err := initializeLocals(ctx, stackInput)
	if err != nil {
		return errors.Wrap(err, "failed to initialize locals")
	}

	// Create kubernetes provider
	kubernetesProvider, err := pulumikubernetesprovider.GetWithKubernetesProviderConfig(ctx,
		stackInput.ProviderConfig, "kubernetes")
	if err != nil {
		return errors.Wrap(err, "failed to create kubernetes provider")
	}

	// Conditionally create namespace resource based on create_namespace flag
	var createdNamespace *kubernetescorev1.Namespace
	if stackInput.Target.Spec.CreateNamespace {
		createdNamespace, err = kubernetescorev1.NewNamespace(ctx,
			locals.Namespace,
			&kubernetescorev1.NamespaceArgs{
				Metadata: kubernetesmetav1.ObjectMetaPtrInput(
					&kubernetesmetav1.ObjectMetaArgs{
						Name:   pulumi.String(locals.Namespace),
						Labels: pulumi.ToStringMap(locals.Labels),
					}),
			}, pulumi.Provider(kubernetesProvider))
		if err != nil {
			return errors.Wrapf(err, "failed to create %s namespace", locals.Namespace)
		}
	}

	// Create deployment - passes namespace as dependency
	createdDeployment, err := deployment(ctx, locals, createdNamespace)
	if err != nil {
		return errors.Wrap(err, "failed to create deployment")
	}

	// Create service - passes namespace as dependency
	if err := service(ctx, locals, createdNamespace, createdDeployment); err != nil {
		return errors.Wrap(err, "failed to create service")
	}

	// Additional resources...
	return nil
}
```

**Key implementation details**:
- `createdNamespace` is `nil` when `create_namespace: false`
- Resources accept `*kubernetescorev1.Namespace` parameter (can be nil)
- Dependency tracking: When namespace is created, resources depend on it; when nil, no dependency

**Example resource function signature**:

```go
func deployment(ctx *pulumi.Context, locals *locals, 
	namespace *kubernetescorev1.Namespace) (*appsv1.Deployment, error) {
	
	opts := []pulumi.ResourceOption{}
	
	// Add namespace as parent if it was created
	if namespace != nil {
		opts = append(opts, pulumi.Parent(namespace))
	}
	
	// Create deployment with options
	deployment, err := appsv1.NewDeployment(ctx, "deployment", 
		&appsv1.DeploymentArgs{
			Metadata: &metav1.ObjectMetaArgs{
				Name:      pulumi.String(locals.Name),
				Namespace: pulumi.String(locals.Namespace), // Just the name
				Labels:    pulumi.ToStringMap(locals.Labels),
			},
			// ... spec
		}, opts...)
	
	return deployment, err
}
```

#### Pattern 2: Helm-Based Components (e.g., KubernetesElasticOperator)

Used for components deployed via Helm charts.

**File**: `iac/pulumi/module/kubernetes_elastic_operator.go`

```go
func kubernetesElasticOperator(ctx *pulumi.Context, locals *Locals,
	k8sProvider *pulumikubernetes.Provider) error {

	// Get namespace from spec
	namespace := locals.KubernetesElasticOperator.Spec.Namespace.GetValue()
	if namespace == "" {
		namespace = vars.Namespace // fallback to default
	}

	// --------------------------------------------------------------------
	// 1. Namespace - conditionally create based on create_namespace flag
	// --------------------------------------------------------------------
	var ns *corev1.Namespace
	var err error
	var namespaceOutput pulumi.StringInput

	if locals.KubernetesElasticOperator.Spec.CreateNamespace {
		// Create the namespace
		ns, err = corev1.NewNamespace(ctx, namespace, &corev1.NamespaceArgs{
			Metadata: metav1.ObjectMetaPtrInput(&metav1.ObjectMetaArgs{
				Name:   pulumi.String(namespace),
				Labels: pulumi.ToStringMap(locals.KubeLabels),
			}),
		}, pulumi.Provider(k8sProvider))
		if err != nil {
			return errors.Wrap(err, "create namespace")
		}
		namespaceOutput = ns.Metadata.Name().Elem()
	} else {
		// Use existing namespace - just reference the name
		namespaceOutput = pulumi.String(namespace)
	}

	// --------------------------------------------------------------------
	// 2. Helm Release
	// --------------------------------------------------------------------
	helmReleaseOpts := []pulumi.ResourceOption{
		pulumi.IgnoreChanges([]string{"status", "description", "resourceNames"}),
	}

	// Only set parent if we created the namespace
	if ns != nil {
		helmReleaseOpts = append(helmReleaseOpts, pulumi.Parent(ns))
	}

	_, err = helm.NewRelease(ctx, "kubernetes-elastic-operator", &helm.ReleaseArgs{
		Name:            pulumi.String(vars.HelmChartName),
		Namespace:       namespaceOutput, // String or namespace.Name output
		Chart:           pulumi.String(vars.HelmChartName),
		Version:         pulumi.String(vars.HelmChartVersion),
		RepositoryOpts:  helm.RepositoryOptsArgs{Repo: pulumi.String(vars.HelmChartRepo)},
		CreateNamespace: pulumi.Bool(false), // We handle namespace creation
		// ... rest of helm args
	}, helmReleaseOpts...)
	
	return err
}
```

**Key differences from Pattern 1**:
- Uses `pulumi.StringInput` type to handle both string literals and output values
- Sets `CreateNamespace: pulumi.Bool(false)` in Helm release (we manage namespace separately)
- Exports namespace output for stack consumers

**Files updated per component**:
- `iac/pulumi/module/main.go` - Entry point with conditional namespace creation
- `iac/pulumi/module/<component>.go` - Resource functions accepting namespace parameter (where applicable)
- `iac/pulumi/module/locals.go` - No changes (namespace extraction already handled via `.GetValue()`)

### Phase 3: Terraform Module Implementation

Updated all Terraform modules with conditional resource creation using `count` parameter.

**File**: `iac/tf/main.tf`

```hcl
##############################################
# Namespace Resource
#
# Creates a dedicated Kubernetes namespace for the
# component deployment when create_namespace is true.
#
# When create_namespace is false, assumes the namespace
# already exists and will be referenced by name.
##############################################
resource "kubernetes_namespace" "this" {
  count = var.spec.create_namespace ? 1 : 0

  metadata {
    name   = local.namespace
    labels = local.final_labels
  }
}
```

**Key implementation details**:
- `count = var.spec.create_namespace ? 1 : 0` - Creates 0 or 1 namespace resources
- When `count = 0`, the resource doesn't exist in the Terraform state
- Other resources reference `local.namespace` (the string name), not the resource itself
- No explicit dependency needed since Terraform tracks namespace by name in Kubernetes API

**Example usage in dependent resources**:

```hcl
# Deployment references namespace by name, not resource
resource "kubernetes_deployment" "app" {
  metadata {
    name      = local.deployment_name
    namespace = local.namespace  # String reference
    labels    = local.final_labels
  }
  # ... spec
}
```

**Terraform behavior**:
- When namespace resource is created (`count = 1`), Terraform ensures it exists before creating dependent resources
- When namespace resource is not created (`count = 0`), Terraform assumes namespace exists and proceeds
- If namespace doesn't exist and `count = 0`, deployment fails with Kubernetes API error

**Variables updated**: No new variables needed - `create_namespace` field accessed via `var.spec.create_namespace` from existing spec variable.

### Phase 4: Documentation Updates

Updated documentation across all 38 components to explain namespace management clearly.

#### Documentation Files Updated (per component)

1. **Pulumi Examples** (`iac/pulumi/examples.md`)
   - Added `create_namespace` field to all YAML examples
   - Documented both `true` and `false` scenarios
   - Explained when to use each mode

2. **Terraform Examples** (`iac/tf/examples.md`)
   - Updated Terraform configurations to include `create_namespace`
   - Showed variable declaration and usage patterns

3. **Component README** (`v1/README.md` or `v1/examples.md`)
   - Updated basic examples with new field
   - Added namespace management guidance

#### Documentation Content Pattern

All documentation now includes:

**Basic Example (standalone component)**:
```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesDeployment
metadata:
  name: minimal-example
spec:
  target_cluster:
    cluster_name: "my-gke-cluster"
  namespace:
    value: "minimal-example"
  create_namespace: true  # Create namespace for this component
  # ... rest of spec
```

**Multi-Component Example** (shown in relevant docs):
```yaml
# First component - creates namespace
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesDeployment
metadata:
  name: cart-service
spec:
  target_cluster:
    cluster_name: "prod-cluster"
  namespace:
    value: "ecommerce-backend"
  create_namespace: true  # First component creates namespace
  # ... spec

---
# Additional components - use existing namespace
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesDeployment
metadata:
  name: order-service
spec:
  target_cluster:
    cluster_name: "prod-cluster"
  namespace:
    value: "ecommerce-backend"
  create_namespace: false  # Use existing namespace
  # ... spec
```

**Namespace Management Guidance** (added to component docs):

> **Namespace Management**
>
> The `create_namespace` field controls whether this component should create the Kubernetes namespace:
>
> - **`create_namespace: true`** (default for new deployments): The component creates the namespace if it doesn't exist. Use this for:
>   - Standalone component deployments
>   - First component in a shared namespace
>   - Development/testing environments
>
> - **`create_namespace: false`**: The component assumes the namespace already exists. Use this for:
>   - Additional components sharing a namespace
>   - Pre-provisioned namespaces with custom ResourceQuotas or NetworkPolicies
>   - Environments where namespace creation requires elevated privileges
>   - GitOps workflows with separate namespace management
>
> **Multi-Component Deployments**: When deploying multiple components to the same namespace, set `create_namespace: true` only for the first component or create the namespace separately.

### Phase 5: Testing and Validation

Updated test fixtures across all components to include the new field.

**File**: `v1/spec_test.go` (example pattern)

```go
func TestKubernetesElasticOperatorSpec_Validate(t *testing.T) {
	tests := []struct {
		name    string
		spec    *KubernetesElasticOperatorSpec
		wantErr bool
	}{
		{
			name: "valid spec with namespace creation",
			spec: &KubernetesElasticOperatorSpec{
				TargetCluster: &kubernetes.KubernetesClusterSelector{
					ClusterName: "test-cluster",
				},
				Namespace: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
						Value: "elastic-system",
					},
				},
				CreateNamespace: true,  // Test with creation enabled
				Container: &KubernetesElasticOperatorSpecContainer{
					Resources: &kubernetes.ContainerResources{
						Requests: &kubernetes.ContainerResourceQuantity{
							Cpu:    "50m",
							Memory: "100Mi",
						},
						Limits: &kubernetes.ContainerResourceQuantity{
							Cpu:    "1000m",
							Memory: "1Gi",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "valid spec without namespace creation",
			spec: &KubernetesElasticOperatorSpec{
				TargetCluster: &kubernetes.KubernetesClusterSelector{
					ClusterName: "test-cluster",
				},
				Namespace: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
						Value: "elastic-system",
					},
				},
				CreateNamespace: false,  // Test with existing namespace
				Container: &KubernetesElasticOperatorSpecContainer{
					Resources: &kubernetes.ContainerResourceQuantity{
						// ... resources
					},
				},
			},
			wantErr: false,
		},
	}
	// ... test execution
}
```

**Test coverage**:
- ✅ Validation with `create_namespace: true`
- ✅ Validation with `create_namespace: false`
- ✅ Proto validation (field presence, types)
- ✅ No additional validation errors (field is optional)

**Build validation**:
```bash
# For each component (example)
cd apis/dev/planton/provider/kubernetes/kubernetesdeployment/v1
go test ./...
go build ./...
```

**Results**: All 38 components pass tests and build successfully.

## Benefits

### 1. Multi-Component Deployment Support

**Before**: Deployment failures when multiple components share a namespace

```yaml
# Component 1
spec:
  namespace:
    value: "shared-backend"
# Attempts to create namespace

# Component 2  
spec:
  namespace:
    value: "shared-backend"
# Conflicts trying to create same namespace!
```

**After**: Clean multi-component deployments

```yaml
# Component 1 - creates namespace
spec:
  namespace:
    value: "shared-backend"
  create_namespace: true

# Component 2 - uses existing
spec:
  namespace:
    value: "shared-backend"
  create_namespace: false

# Component 3 - uses existing
spec:
  namespace:
    value: "shared-backend"
  create_namespace: false
```

**Impact**: Enables true microservices architectures with multiple components per namespace.

### 2. Pre-Existing Namespace Integration

Organizations can now pre-provision namespaces with:
- **ResourceQuotas**: CPU/memory limits per namespace
- **NetworkPolicies**: Default deny/allow rules
- **LimitRanges**: Default resource requests/limits for pods
- **Custom annotations**: Cost center, team ownership, compliance tags
- **ServiceAccounts**: Pre-configured service accounts with RBAC

**Example workflow**:
```bash
# Platform team creates namespace with policies
kubectl create namespace production-apps
kubectl apply -f resource-quota.yaml
kubectl apply -f network-policy.yaml
kubectl apply -f limit-range.yaml

# Development team deploys components
planton apply -f cart-service.yaml     # create_namespace: false
planton apply -f order-service.yaml    # create_namespace: false
planton apply -f payment-service.yaml  # create_namespace: false
```

### 3. Enterprise Compliance and Security

**RBAC-restricted environments**:
- Developers can deploy components without `create namespace` permissions
- Platform teams control namespace lifecycle separately
- Clear separation of concerns between infrastructure and application teams

**Audit and compliance**:
- Namespace creation events separate from component deployments
- Easier to track who created namespaces vs. who deployed components
- Supports least-privilege access models

### 4. GitOps and IaC Workflow Compatibility

**Scenario**: Namespaces managed by one Git repository, components by another

```
infrastructure-repo/
  namespaces/
    - production.yaml
    - staging.yaml
    - dev.yaml

applications-repo/
  services/
    - cart-service.yaml      # create_namespace: false
    - order-service.yaml     # create_namespace: false
    - payment-service.yaml   # create_namespace: false
```

**Workflow**:
1. Infrastructure team merges namespace changes → namespaces created/updated
2. Application team merges service changes → services deployed to existing namespaces
3. No conflicts, clear ownership boundaries

### 5. Dependency Ordering Control

Users can now explicitly control resource creation order:

```yaml
# Step 1: Create namespace with custom configuration
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesNamespace
metadata:
  name: backend
spec:
  # ... namespace config with quotas, policies

# Step 2: Deploy components to pre-configured namespace
apiVersion: kubernetes.planton.dev/v1  
kind: KubernetesPostgres
metadata:
  name: backend-db
spec:
  namespace:
    ref: backend  # Reference to KubernetesNamespace resource
  create_namespace: false
```

### 6. Backward Compatibility

**Default behavior**: `create_namespace` defaults to `false` in proto3
- Existing deployments continue working (assuming namespaces exist)
- Users must explicitly opt in to namespace creation
- Prevents unexpected namespace creation in production

**Migration path**: Users can gradually adopt the new flag as needed.

## Impact

### Who Benefits

**Development Teams**:
- Simplified multi-component deployments
- No more namespace conflict errors
- Clear control over namespace lifecycle

**Platform Engineers**:
- Can enforce namespace standards via pre-provisioning
- Maintain control over namespace creation
- Better separation of concerns

**Enterprise Organizations**:
- Compliance with namespace governance policies
- Support for least-privilege RBAC models
- Integration with existing namespace management tools

**GitOps Practitioners**:
- Clean separation of infrastructure and application manifests
- Easier to implement progressive delivery patterns
- Better change tracking and auditing

### Deployment Patterns Enabled

**Pattern 1: Single Namespace, Multiple Workloads**
```yaml
# API Gateway creates namespace
kind: KubernetesDeployment (api-gateway)
spec:
  namespace: { value: "api-services" }
  create_namespace: true

# Other services use existing namespace
kind: KubernetesDeployment (auth-service)
spec:
  namespace: { value: "api-services" }
  create_namespace: false

kind: KubernetesDeployment (user-service)
spec:
  namespace: { value: "api-services" }
  create_namespace: false
```

**Pattern 2: Namespace Component + Workloads**
```yaml
# Dedicated namespace resource
kind: KubernetesNamespace
metadata:
  name: data-platform
spec:
  resource_quota:
    hard:
      cpu: "100"
      memory: "200Gi"
  network_policies:
    - default-deny-all

# Workloads reference namespace
kind: KubernetesPostgres
spec:
  namespace: { ref: data-platform }
  create_namespace: false

kind: KubernetesRedis  
spec:
  namespace: { ref: data-platform }
  create_namespace: false
```

**Pattern 3: Environment-Specific Strategies**

Development:
```yaml
# Create namespaces automatically for fast iteration
spec:
  create_namespace: true
```

Production:
```yaml
# Use pre-provisioned namespaces with policies
spec:
  create_namespace: false
```

### Components Affected

**All 38 Kubernetes components** (excluding `kubernetesnamespace`):

**Operators & Add-ons** (13):
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

### Breaking Changes

**⚠️ Behavioral Change**

**Previous behavior** (after Nov 23 namespace standardization):
- All components **always created** the namespace specified in `spec.namespace`
- No way to disable namespace creation

**New behavior**:
- Components create namespace **only if** `create_namespace: true`
- Default value is `false` (proto3 default for boolean)
- Existing deployments may fail if namespace doesn't exist and flag not set

### Migration Guide

**For existing deployments** (manifests created after Nov 23, 2025):

**Option 1: Explicitly enable namespace creation (recommended)**
```yaml
spec:
  namespace:
    value: "my-namespace"
  create_namespace: true  # Add this line
```

**Option 2: Pre-create namespaces**
```bash
kubectl create namespace my-namespace
# Keep create_namespace: false (or omit the field)
```

**For new deployments**:

**Single component per namespace**:
```yaml
spec:
  create_namespace: true  # Component creates its own namespace
```

**Multiple components per namespace**:
```yaml
# First component
spec:
  create_namespace: true

# Subsequent components
spec:
  create_namespace: false
```

**Pre-provisioned namespaces**:
```yaml
# All components
spec:
  create_namespace: false  # Assume namespace exists
```

## Implementation Timeline

This feature was implemented across multiple work sessions using the `complete-planton-component` rule:

### Initial Development
- **Date**: December 16, 2025
- **Proto changes**: Manual addition of `create_namespace` field to 38 component specs
- **Commit**: `82e92d3b` - "added create_namespace field to all kubernetes components"
- **Files changed**: 108 (proto definitions, Go stubs, TypeScript stubs)

### Component Updates (38 separate sessions)
- **Approach**: One conversation per component using `@complete-planton-component` rule
- **Scope per session**:
  - Updated Pulumi modules with conditional namespace creation
  - Updated Terraform modules with `count` parameter
  - Updated all documentation (examples.md, README.md)
  - Updated test fixtures (spec_test.go)
  - Compiled and validated (Go, Pulumi, Terraform)
  
**Components processed**: All 38 Kubernetes components (excluding namespace)

### Verification
- ✅ All proto validations pass
- ✅ All Go code compiles
- ✅ All Pulumi modules build successfully
- ✅ All Terraform modules validate
- ✅ Documentation updated consistently
- ✅ Test fixtures include new field

## Code Metrics

- **Proto files changed**: 38 (one per component spec.proto)
- **Generated Go files changed**: 38 (spec.pb.go)
- **Generated TypeScript files changed**: 38 (spec_pb.ts)
- **Pulumi module files updated**: ~76 (main.go and component files)
- **Terraform module files updated**: ~38 (main.tf)
- **Documentation files updated**: ~114 (examples.md, README.md per component)
- **Test files updated**: 38 (spec_test.go per component)
- **Total files modified**: ~380
- **Lines added**: ~1,499
- **Lines removed**: ~776
- **Net change**: ~+723 lines

**Proto change pattern** (consistent across all 38 components):
```diff
+ // flag to indicate if the namespace should be created
+ bool create_namespace = 3;
```

## Related Work

### Foundation

**[Namespace Standardization (Nov 23, 2025)](../2025-11/2025-11-23-220641-standardize-kubernetes-components-target-cluster-namespace.md)**:
- Added required `target_cluster` and `namespace` fields to all components
- Standardized namespace as `StringValueOrRef` type
- Established consistent field ordering

**This enhancement builds directly on that work**:
- Previous change: Made namespace field mandatory
- This change: Added control over namespace creation
- Combined result: Consistent namespace specification + flexible namespace management

### Complementary Features

**Foreign Key Support**:
- `namespace` field supports both literal values and references
- Enables referencing `KubernetesNamespace` resource
- Future: Could support cross-component namespace dependencies

**Target Cluster Selection**:
- All components have standardized cluster targeting
- Combined with namespace control, enables precise deployment control
- Foundation for multi-cluster deployments

### Future Enhancements

This feature enables:

**1. Namespace Lifecycle Policies**
- Components could respect namespace deletion policies
- Cascade deletion or orphan resources based on policy
- Integration with Kubernetes finalizers

**2. Namespace Templates**
- Predefined namespace configurations (dev, staging, prod)
- Component references template, gets consistent policies
- Reduces configuration duplication

**3. Multi-Tenancy Support**
- Tenant-specific namespace provisioning
- Automatic isolation via NetworkPolicies
- Resource quota enforcement per tenant

**4. GitOps Integration**
- Dedicated namespace sync waves (ArgoCD, Flux)
- Namespace creation in wave 0, components in wave 1+
- Better progressive delivery support

**5. Cost Allocation**
- Namespace-level cost tracking
- Automatic labeling for billing integration
- Per-team or per-project resource usage

## Usage Examples

### Example 1: Microservices Application (Shared Namespace)

Deploying a full microservices application with multiple components to a single `ecommerce` namespace:

```yaml
# 1. API Gateway (creates namespace)
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesDeployment
metadata:
  name: api-gateway
spec:
  target_cluster:
    cluster_name: "prod-gke-us-central1"
  namespace:
    value: "ecommerce"
  create_namespace: true  # First component creates namespace
  version: main
  container:
    app:
      image:
        repo: myorg/api-gateway
        tag: "2.1.0"
      # ... container config

---
# 2. Cart Service (uses existing)
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesDeployment
metadata:
  name: cart-service
spec:
  target_cluster:
    cluster_name: "prod-gke-us-central1"
  namespace:
    value: "ecommerce"
  create_namespace: false  # Uses namespace created by api-gateway
  version: main
  container:
    app:
      image:
        repo: myorg/cart-service
        tag: "1.5.2"
      # ... container config

---
# 3. Redis Cache (uses existing)
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesRedis
metadata:
  name: cart-cache
spec:
  target_cluster:
    cluster_name: "prod-gke-us-central1"
  namespace:
    value: "ecommerce"
  create_namespace: false  # Uses namespace created by api-gateway
  container:
    resources:
      requests:
        memory: "256Mi"
      limits:
        memory: "1Gi"

---
# 4. PostgreSQL Database (uses existing)
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesPostgres
metadata:
  name: ecommerce-db
spec:
  target_cluster:
    cluster_name: "prod-gke-us-central1"
  namespace:
    value: "ecommerce"
  create_namespace: false  # Uses namespace created by api-gateway
  database_name: "ecommerce"
  container:
    resources:
      requests:
        cpu: "500m"
        memory: "2Gi"
      limits:
        cpu: "2000m"
        memory: "8Gi"
```

**Deployment order**: Can be applied simultaneously - Pulumi/Terraform handles dependencies.

### Example 2: Operator + Workload (Pre-Provisioned Namespace)

Deploying Elastic Operator and Elasticsearch cluster to a pre-configured namespace:

```bash
# Platform team creates namespace with policies
kubectl apply -f - <<EOF
apiVersion: v1
kind: Namespace
metadata:
  name: elastic-stack
  labels:
    istio-injection: enabled
    cost-center: "engineering"
---
apiVersion: v1
kind: ResourceQuota
metadata:
  name: elastic-quota
  namespace: elastic-stack
spec:
  hard:
    requests.cpu: "50"
    requests.memory: "100Gi"
    limits.cpu: "100"
    limits.memory: "200Gi"
EOF
```

```yaml
# Development team deploys Elastic components
---
# Elastic Operator
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesElasticOperator
metadata:
  name: elastic-operator
spec:
  target_cluster:
    cluster_name: "prod-gke-us-central1"
  namespace:
    value: "elastic-stack"
  create_namespace: false  # Uses pre-provisioned namespace
  container:
    resources:
      requests:
        cpu: "100m"
        memory: "256Mi"
      limits:
        cpu: "1000m"
        memory: "1Gi"

---
# Elasticsearch Cluster
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesElasticsearch
metadata:
  name: logs-cluster
spec:
  target_cluster:
    cluster_name: "prod-gke-us-central1"
  namespace:
    value: "elastic-stack"
  create_namespace: false  # Uses pre-provisioned namespace
  version: "8.11.0"
  cluster_name: "production-logs"
  # ... elasticsearch config
```

### Example 3: Development vs. Production Strategy

Different namespace strategies per environment:

```yaml
# Development - auto-create namespaces for speed
---
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesDeployment
metadata:
  name: feature-x-api
spec:
  target_cluster:
    cluster_name: "dev-cluster"
  namespace:
    value: "feature-x-dev"
  create_namespace: true  # Dev environments auto-create
  version: feature-x
  # ... container config

---
# Production - use pre-provisioned namespaces with governance
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesDeployment
metadata:
  name: api-production
spec:
  target_cluster:
    cluster_name: "prod-cluster"
  namespace:
    value: "production-apis"
  create_namespace: false  # Production requires pre-provisioning
  version: main
  # ... container config
```

### Example 4: GitOps Repository Structure

Organizing manifests for namespace separation:

```
k8s-manifests/
├── infrastructure/
│   ├── namespaces/
│   │   ├── production.yaml       # Namespace with quotas, policies
│   │   ├── staging.yaml
│   │   └── development.yaml
│   └── operators/
│       ├── cert-manager.yaml     # create_namespace: true (own namespace)
│       └── external-dns.yaml     # create_namespace: true (own namespace)
│
└── applications/
    ├── backend/
    │   ├── cart-service.yaml     # create_namespace: false (uses production)
    │   ├── order-service.yaml    # create_namespace: false (uses production)
    │   └── payment-service.yaml  # create_namespace: false (uses production)
    │
    └── databases/
        ├── postgres.yaml         # create_namespace: false (uses production)
        └── redis.yaml            # create_namespace: false (uses production)
```

**ArgoCD Application Setup**:
```yaml
# Sync Wave 0: Infrastructure (namespaces)
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: infrastructure
spec:
  syncPolicy:
    syncOptions:
      - CreateNamespace=false
  source:
    path: infrastructure/namespaces
  # ...

---
# Sync Wave 1: Operators (create their own namespaces)
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: operators
spec:
  source:
    path: infrastructure/operators
  # ...

---
# Sync Wave 2: Applications (use existing namespaces)
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: applications
spec:
  source:
    path: applications
  # ...
```

## Troubleshooting

### Common Issues

**Issue 1: Namespace not found error**

```
Error: namespaces "my-namespace" not found
```

**Cause**: `create_namespace: false` but namespace doesn't exist

**Solutions**:
- Set `create_namespace: true` to auto-create
- Pre-create namespace: `kubectl create namespace my-namespace`
- Deploy a component with `create_namespace: true` first

---

**Issue 2: Multiple components trying to create same namespace**

```
Error: namespace "shared-namespace" already exists
```

**Cause**: Multiple components with `create_namespace: true` targeting same namespace

**Solution**: Set `create_namespace: true` only on the first component:
```yaml
# First component
spec:
  create_namespace: true

# Other components  
spec:
  create_namespace: false
```

---

**Issue 3: Permission denied creating namespace**

```
Error: namespaces is forbidden: User "developer" cannot create resource "namespaces"
```

**Cause**: Service account lacks `create namespace` permissions

**Solutions**:
- Set `create_namespace: false` and pre-create namespace
- Grant namespace creation permissions (if appropriate)
- Use platform team to provision namespaces

---

**Issue 4: Namespace deletion removes all components**

**Cause**: When namespace is created by a component, deleting that component may delete the namespace (and all resources in it)

**Solutions**:
- Use dedicated `KubernetesNamespace` resource for shared namespaces
- Set `create_namespace: false` on all components using shared namespace
- Document namespace ownership clearly

## Best Practices

### 1. Namespace Ownership

**✅ Do**: Clearly document which component creates the namespace
```yaml
# Namespace creator (designated "anchor" component)
metadata:
  name: api-gateway
  annotations:
    description: "Creates 'api-services' namespace for all API components"
spec:
  create_namespace: true
```

**❌ Don't**: Have multiple components creating the same namespace
```yaml
# Multiple components with create_namespace: true → conflicts
```

### 2. Production Environments

**✅ Do**: Pre-provision namespaces with policies
```bash
# Platform team provisions
kubectl apply -f production-namespaces.yaml
```

```yaml
# Components use existing
spec:
  create_namespace: false
```

**❌ Don't**: Auto-create production namespaces from components
```yaml
# Risky in production - bypasses governance
spec:
  create_namespace: true
```

### 3. Development Environments

**✅ Do**: Use auto-creation for developer productivity
```yaml
# Fast iteration in dev
spec:
  create_namespace: true
```

### 4. Multi-Component Applications

**✅ Do**: Use consistent namespace strategy
```yaml
# Option A: First component creates
component1: create_namespace: true
component2: create_namespace: false
component3: create_namespace: false

# Option B: Dedicated namespace resource
namespace: KubernetesNamespace
component1: create_namespace: false
component2: create_namespace: false
component3: create_namespace: false
```

**❌ Don't**: Inconsistent namespace creation
```yaml
# Confusing - who owns the namespace?
component1: create_namespace: true
component2: create_namespace: true
component3: create_namespace: false
```

### 5. GitOps Workflows

**✅ Do**: Separate namespace and component manifests
```
infrastructure/namespaces/  → sync wave 0
applications/               → sync wave 1+
```

**✅ Do**: Use `create_namespace: false` with GitOps
```yaml
# Clear dependency: namespace provisioned first
spec:
  create_namespace: false
```

## Lessons Learned

### What Worked Well

1. **Consistent proto pattern**: Adding field #3 in the same position across all 38 components made implementation predictable
2. **Minimal comment**: Simple proto comment "flag to indicate if the namespace should be created" was clear without over-documentation
3. **Two implementation patterns**: Having clear patterns for resource-based (Deployment) vs. Helm-based (operators) components simplified coding
4. **Parallel component processing**: Using separate conversations per component allowed parallel work and isolated issues
5. **Comprehensive documentation**: Updating examples to show both `true` and `false` scenarios helped clarify usage

### Challenges Overcome

1. **Dependency tracking**: Ensuring resources properly depend on namespace when created required careful nil-checking in Pulumi code
2. **Terraform count semantics**: Understanding Terraform's `count = 0` behavior (resource doesn't exist vs. creates nothing) was critical
3. **Documentation consistency**: Ensuring all 38 × 3 = 114 documentation files had consistent explanations required systematic approach
4. **Test fixture updates**: Remembering to add field to test fixtures in all 38 components required checklist discipline

### Recommendations for Similar Changes

1. **Proto defaults matter**: Boolean fields default to `false` in proto3 - consider if this matches desired default behavior
2. **Document migration**: Clear migration guide prevents production issues when changing component behavior
3. **Test both modes**: Validate components work with both `create_namespace: true` and `false`
4. **Use automation rules**: Leveraging `@complete-planton-component` rule ensured consistent updates across components
5. **Separate concerns**: Keeping proto changes in one commit, implementation in separate conversations simplified tracking

## Future Considerations

### Potential Enhancements

**1. Default Value Configuration**
- Allow system-wide or component-level defaults for `create_namespace`
- Environment-specific defaults (auto-create in dev, require existing in prod)

**2. Namespace Configuration**
- When creating namespace, apply component-specific labels/annotations
- Support namespace resource quotas from component spec
- Enable namespace network policies from component spec

**3. Validation Improvements**
- Warn if multiple components in same stack have `create_namespace: true` for same namespace
- Validate namespace exists when `create_namespace: false` (pre-flight check)
- Suggest `create_namespace: false` when detecting shared namespaces

**4. Documentation Generation**
- Auto-generate namespace ownership documentation from manifests
- Visualize namespace-to-component relationships
- Detect orphaned namespaces (created but no longer referenced)

**5. Lifecycle Management**
- Support namespace retention policies (delete with component or orphan)
- Integrate with Kubernetes finalizers for safer deletion
- Provide namespace migration commands (move components between namespaces)

---

**Status**: ✅ Production Ready  
**Impact**: All 38 Kubernetes components support flexible namespace management  
**Testing**: Validated across all components - Pulumi, Terraform, and documentation  
**Timeline**: Proto changes December 16, 2025; Implementation across 38 separate component update sessions  
**Migration Required**: Existing deployments should add `create_namespace: true` or pre-create namespaces
