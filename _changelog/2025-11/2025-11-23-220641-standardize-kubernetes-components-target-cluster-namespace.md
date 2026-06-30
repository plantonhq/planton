# Standardize Kubernetes Components with Target Cluster and Namespace Fields

**Date**: November 23, 2025  
**Type**: Refactoring | Enhancement  
**Components**: API Definitions, Kubernetes Provider, Protobuf Schemas, Pulumi Modules, Documentation

## Summary

Standardized all 37 Kubernetes component specifications by adding required `target_cluster` and `namespace` fields as the first two fields in every spec. This brings consistency across all Kubernetes components (both add-ons and workloads), enabling unified cluster targeting and namespace management. Updated all implementation code, documentation, and tests to work with the new namespace type (`StringValueOrRef` instead of direct string) and simplified target_cluster format.

## Problem Statement / Motivation

The Kubernetes provider in Planton had inconsistent field structures across components:

### Pain Points

- **Inconsistent cluster targeting**: Only add-ons had `target_cluster` field, while other components lacked it entirely
- **Varied namespace handling**: Some components used string namespace, some had no namespace field, creating confusion
- **Documentation inconsistency**: Examples across components showed different patterns for cluster and namespace specification
- **Poor developer experience**: No standardized way to specify where a Kubernetes resource should be deployed
- **Migration complexity**: After simplifying `target_cluster` structure (removing `oneof` wrapper), documentation wasn't updated consistently

## Solution / What's New

Added two required fields at the beginning of every Kubernetes component spec:

### New Standard Structure

```protobuf
message Kubernetes<ComponentName>Spec {
  // The Kubernetes cluster to install this component on.
  dev.planton.provider.kubernetes.KubernetesClusterSelector target_cluster = 1;

  // Kubernetes namespace to install the operator.
  dev.planton.shared.foreignkey.v1.StringValueOrRef namespace = 2 [
    (buf.validate.field).required = true,
    (dev.planton.shared.foreignkey.v1.default_kind) = KubernetesNamespace,
    (dev.planton.shared.foreignkey.v1.default_kind_field_path) = "spec.name"
  ];

  // ... rest of component-specific fields
}
```

### Key Changes

1. **Target Cluster Field (field 1)**: Direct cluster selector without wrapper
2. **Namespace Field (field 2)**: `StringValueOrRef` supporting both literal values and references
3. **Required Imports**: Added `target_cluster.proto` and `foreign_key.proto` imports where missing
4. **Field Sequence Updates**: Incremented all existing field numbers by 2 to accommodate new fields
5. **Unified Format**: All 37 components now follow identical structure

### Namespace Type Change

**From direct string:**
```protobuf
string namespace = 1;
```

**To foreign key reference:**
```protobuf
dev.planton.shared.foreignkey.v1.StringValueOrRef namespace = 2 [
  (buf.validate.field).required = true,
  (dev.planton.shared.foreignkey.v1.default_kind) = KubernetesNamespace,
  (dev.planton.shared.foreignkey.v1.default_kind_field_path) = "spec.name"
];
```

This enables both literal values and references to `KubernetesNamespace` resources.

## Implementation Details

### Phase 1: Proto Schema Updates

**Scope**: 37 Kubernetes component `spec.proto` files

**Process**:
1. Added required imports (if not present)
2. Added `target_cluster` and `namespace` fields as fields 1 and 2
3. Updated all existing field sequence numbers (+2 offset)
4. Maintained only the main Spec message (no changes to nested messages/enums)

**Example transformation**:

```diff
 message KubernetesKeycloakSpec {
+  dev.planton.provider.kubernetes.KubernetesClusterSelector target_cluster = 1;
+
+  dev.planton.shared.foreignkey.v1.StringValueOrRef namespace = 2 [
+    (buf.validate.field).required = true,
+    (dev.planton.shared.foreignkey.v1.default_kind) = KubernetesNamespace,
+    (dev.planton.shared.foreignkey.v1.default_kind_field_path) = "spec.name"
+  ];
+
-  string admin_username = 1;
+  string admin_username = 3;
-  string admin_password = 2;
+  string admin_password = 4;
-  KubernetesKeycloakContainer container = 3;
+  KubernetesKeycloakContainer container = 5;
 }
```

### Phase 2: Pulumi Module Updates

**Critical change**: Namespace field access pattern changed due to type change.

**Old pattern**:
```go
namespace := stackInput.Target.Spec.Namespace
```

**New pattern**:
```go
namespace := stackInput.Target.Spec.Namespace.GetValue()
```

**Files updated** (pattern applied across all 37 components):
- `iac/pulumi/module/locals.go` - Updated namespace extraction
- `iac/pulumi/module/main.go` - Updated namespace usage
- `iac/pulumi/module/<component>.go` - Updated component-specific namespace handling

**Example from `kubernetesaltinityoperator`**:

```go
// newLocals creates and initializes local values from the stack input
func newLocals(stackInput *kubernetesaltinityoperatorv1.KubernetesAltinityOperatorStackInput) *locals {
	l := &locals{}

	// Determine namespace - use from spec or default
	l.Namespace = stackInput.Target.Spec.Namespace.GetValue()  // Added .GetValue()
	if l.Namespace == "" {
		l.Namespace = vars.DefaultNamespace
	}

	// ... rest of locals initialization
}
```

### Phase 3: Test Fixture Updates

**Updated** `v1/spec_test.go` in all 37 components to include new required fields.

**Example test fixture**:

```go
import (
    foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
    "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes"
)

spec = &KubernetesKeycloakSpec{
    TargetCluster: &kubernetes.KubernetesClusterSelector{
        ClusterName: "my-gke-cluster",
    },
    Namespace: &foreignkeyv1.StringValueOrRef{
        LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
            Value: "keycloak",
        },
    },
    // ... rest of spec fields
}
```

### Phase 4: Documentation Updates

**Updated** documentation across all components to reflect both changes:

#### 4.1 Target Cluster Format Simplification

Removed deprecated formats from all documentation:

**Old Format 1 (removed)**:
```yaml
spec:
  target_cluster:
    kubernetes_credential_id: "my-cluster"
```

**Old Format 2 (removed)**:
```yaml
spec:
  target_cluster:
    kubernetes_cluster_selector:
      cluster_kind: 607
      cluster_name: "my-cluster"
```

**New Format (current)**:
```yaml
spec:
  target_cluster:
    cluster_name: "my-gke-cluster"
  namespace:
    value: keycloak
```

#### 4.2 Updated Documentation Files

For each of the 37 components:
- ✅ `iac/pulumi/examples.md` - All examples updated with new fields
- ✅ `v1/README.md` - Examples updated where present
- ✅ `iac/tf/examples.md` - Terraform examples updated (where applicable)

**Example from updated documentation**:

```yaml
kind: KubernetesKeycloak
version: v1
metadata:
  name: production-keycloak
spec:
  target_cluster:
    cluster_name: "prod-gke-us-central1"
  namespace:
    value: "auth-services"
  admin_username: "admin"
  admin_password: "change-me-in-production"
  container:
    replicas: 3
    resources:
      requests:
        cpu: "500m"
        memory: "1Gi"
      limits:
        cpu: "2000m"
        memory: "4Gi"
```

### Phase 5: Terraform Module Updates

For components with Terraform modules:
- Added `spec_target_cluster_cluster_name` variable
- Added `spec_namespace` variable
- Updated variable references in `main.tf`

**Example `variables.tf` additions**:

```hcl
variable "spec_target_cluster_cluster_name" {
  description = "Name of the target Kubernetes cluster"
  type        = string
}

variable "spec_namespace" {
  description = "Kubernetes namespace for the component"
  type        = string
}
```

## Benefits

### Consistency and Predictability

- **Uniform API**: All 37 Kubernetes components now follow identical structure
- **Developer experience**: Clear, consistent pattern for cluster and namespace specification
- **IDE support**: Autocomplete and type checking work identically across components
- **Reduced cognitive load**: Learn once, apply everywhere

### Foreign Key Support

The `StringValueOrRef` type enables:
- **Literal values**: Direct namespace name specification
- **Resource references**: Reference `KubernetesNamespace` resource for the value
- **Future flexibility**: Foundation for cross-resource dependencies

### Simplified Target Cluster Format

- **No wrapper overhead**: Direct `cluster_name` field instead of nested structures
- **Clear intent**: Obvious what the field does
- **Backward compatible proto**: Type system unchanged, only documentation updated

### Improved Documentation Quality

- **Consistent examples**: Every component shows same pattern
- **Copy-paste friendly**: Examples work immediately with new schema
- **Up-to-date**: No deprecated formats lingering in docs

## Impact

### Components Affected

**37 total Kubernetes components updated**:

**Add-ons** (had `target_cluster`, updated namespace type):
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

**Workloads and other components** (added both fields):
- kubernetesargocd
- kubernetesclickhouse
- kubernetesdeployment
- kubernetesharbor
- kuberneteshelmrelease
- kubernetesjenkins
- kuberneteskafka
- kuberneteskeycloak
- kubernetesmongodb
- kubernetesnamespace
- kubernetesnats
- kubernetesneo4j
- kubernetesopenfga
- kubernetespostgres
- kubernetesredis
- kubernetessignoz
- kubernetessolr
- kubernetestemporal
- ... and 7 more

### Breaking Changes

This is a **breaking API change** requiring:

1. **Manifest updates**: All existing Kubernetes component manifests must add `target_cluster` and `namespace`
2. **Code updates**: Any direct access to `Spec.Namespace` must change to `Spec.Namespace.GetValue()`
3. **Test updates**: Test fixtures must include new required fields
4. **Documentation refresh**: All examples must use new format

### Migration Required

**For YAML manifests**:
```diff
 kind: KubernetesKeycloak
 version: v1
 metadata:
   name: my-keycloak
 spec:
+  target_cluster:
+    cluster_name: "my-gke-cluster"
+  namespace:
+    value: "keycloak"
   admin_username: "admin"
   # ... rest of spec
```

**For Go code accessing namespace**:
```diff
-namespace := stackInput.Target.Spec.Namespace
+namespace := stackInput.Target.Spec.Namespace.GetValue()
```

**For test fixtures**:
```diff
 spec := &KubernetesKeycloakSpec{
+    TargetCluster: &kubernetes.KubernetesClusterSelector{
+        ClusterName: "test-cluster",
+    },
+    Namespace: &foreignkeyv1.StringValueOrRef{
+        LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
+            Value: "test-namespace",
+        },
+    },
     AdminUsername: "admin",
     // ... rest
 }
```

## Code Metrics

- **Proto files changed**: 37 (all Kubernetes component specs)
- **Pulumi module files updated**: ~74 (locals.go and main.go for most components)
- **Test files updated**: 37 (spec_test.go in each component)
- **Documentation files updated**: ~111 (examples.md, README.md for all components)
- **Total files modified**: ~259
- **Field sequence numbers updated**: ~500+ (across all specs)
- **Lines of code changed**: ~1,500+ (across proto, Go, and documentation)

## Automation Strategy

Created two reusable prompts for coding agents:

### 1. Proto Update Prompt
**File**: `apis/dev/planton/provider/kubernetes/_cursor/add-fields.prompt.md`

Guides agents to:
- Add required imports
- Insert exact field definitions at top of spec
- Update field sequence numbers by +2
- Validate changes

### 2. Implementation Update Prompt  
**File**: `apis/dev/planton/provider/kubernetes/_cursor/update-all-other-aspects.md`

Guides agents to:
- Update Pulumi Go code with `.GetValue()` pattern
- Add Terraform variables
- Update all documentation with new format
- Remove deprecated target_cluster formats
- Run validation commands
- Verify tests pass

Combined with `@update-planton-component` rule for orchestration.

## Validation

All 37 components were validated:

```bash
# For each component
cd apis/dev/planton/provider/kubernetes/<component>/v1
go test ./...
go build ./...
```

Results:
- ✅ All proto validations pass
- ✅ All Go compilation successful
- ✅ All tests passing with new fixtures
- ✅ No breaking errors in Pulumi modules
- ✅ Documentation examples verified

## Related Work

### Prior Changes Referenced

- **Target Cluster Simplification** (2025-11-23): Removed `KubernetesAddonTargetCluster` wrapper with `oneof` abstraction, this change updated all documentation to reflect the simplified format
- **Foreign Key Support**: Leveraged existing `StringValueOrRef` infrastructure for namespace references

### Foundation for Future Work

This standardization enables:
- **Unified CLI commands**: Generic operations across all Kubernetes components
- **Better validation**: Consistent validation rules for cluster and namespace
- **Resource dependencies**: Namespace can reference `KubernetesNamespace` resource
- **Multi-cluster workflows**: Clear cluster targeting in all components
- **Automated tooling**: Scripts and tools can rely on consistent structure

## Developer Experience Improvements

### Before (Inconsistent)

```yaml
# Add-on (had target_cluster)
kind: KubernetesCertManager
spec:
  target_cluster:
    kubernetes_credential_id: "cluster-1"
  # no namespace field

# Workload (no target_cluster)
kind: KubernetesKeycloak
spec:
  # no way to specify cluster
  admin_username: "admin"
```

### After (Consistent)

```yaml
# Add-on
kind: KubernetesCertManager
spec:
  target_cluster:
    cluster_name: "prod-gke"
  namespace:
    value: "cert-manager"

# Workload  
kind: KubernetesKeycloak
spec:
  target_cluster:
    cluster_name: "prod-gke"
  namespace:
    value: "auth-services"
  admin_username: "admin"
```

**Same pattern, every component, every time.**

## Lessons Learned

### What Worked Well

- **Coding agent prompts**: Reusable prompts enabled consistent updates across 37 components
- **Phased approach**: Proto → Code → Tests → Documentation sequence prevented rework
- **Validation at each step**: Running tests after each component caught issues early

### Challenges Overcome

- **Namespace type confusion**: Clear documentation of `.GetValue()` pattern prevented errors
- **Large scope**: 37 components × ~7 files each = 259 files, automation was critical
- **Documentation debt**: Cleaned up deprecated formats while adding new fields

### Recommendations for Similar Changes

1. **Create reusable prompts first**: Invest time upfront in clear, detailed prompts
2. **Test one component fully**: Validate entire flow with one component before scaling
3. **Combine with orchestration rules**: Use existing rule framework for consistency
4. **Document the change**: Create changelog capturing context for future reference

---

**Status**: ✅ Production Ready  
**Impact**: Breaking API change across all 37 Kubernetes components  
**Testing**: All component tests passing, documentation verified  
**Timeline**: Completed in one session using automated coding agents

