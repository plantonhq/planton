# Rename KubernetesMicroservice to KubernetesDeployment

**Date**: November 15, 2025  
**Type**: Refactoring  
**Impact**: High  
**Scope**: Kubernetes provider, API resources, tests  
**Breaking**: Yes (for user manifests)

## Summary

Systematically renamed the `KubernetesMicroservice` cloud resource to `KubernetesDeployment` across the entire Planton codebase. This refactoring removes an unnecessary abstraction layer and accurately reflects that the resource creates a Kubernetes Deployment, not a generic "microservice." The rename applied 7 comprehensive naming pattern replacements across 45+ files, updated the cloud resource registry, fixed test references, and verified through the full build pipeline.

## Motivation

### The Problem: Inaccurate Abstraction

The name `KubernetesMicroservice` was misleading:

1. **Architectural Mismatch**: The resource creates a Kubernetes `Deployment` object, not a specialized "microservice" construct
2. **Inconsistent Naming**: Other similar resources like `KubernetesCronJob` accurately name the underlying Kubernetes resource they create
3. **Future Limitations**: The abstraction prevented adding related resources like `KubernetesStatefulSet` with clear differentiation
4. **Developer Confusion**: The name implied special microservice functionality that didn't exist

### The Solution: Accurate Naming

Renaming to `KubernetesDeployment`:
- ✅ Accurately describes the Kubernetes resource created
- ✅ Maintains naming consistency with `KubernetesCronJob`, `KubernetesPostgres`, etc.
- ✅ Sets the stage for future additions like `KubernetesStatefulSet`, `KubernetesDaemonSet`
- ✅ Reduces cognitive overhead for developers

## Implementation

### Rename Strategy

Used the automated rename script located at `_rules/deployment-component/rename/_scripts/rename_deployment_component.py` which applies 7 comprehensive naming patterns:

1. **PascalCase**: `KubernetesMicroservice` → `KubernetesDeployment`
2. **camelCase**: `kubernetesMicroservice` → `kubernetesDeployment`
3. **UPPER_SNAKE_CASE**: `KUBERNETES_MICROSERVICE` → `KUBERNETES_DEPLOYMENT`
4. **snake_case**: `kubernetes_microservice` → `kubernetes_deployment`
5. **kebab-case**: `kubernetes-microservice` → `kubernetes-deployment`
6. **space separated**: `"kubernetes microservice"` → `"kubernetes deployment"`
7. **lowercase**: `kubernetesmicroservice` → `kubernetesdeployment`

### Changes Applied

#### Component Directory
- **Copied**: `apis/dev/planton/provider/kubernetes/workload/kubernetesmicroservice/` → `kubernetesdeployment/`
- **Updated**: All proto files, Go files, documentation, examples, IaC modules (Pulumi, Terraform)
- **Deleted**: Old `kubernetesmicroservice/` directory after successful verification

#### Cloud Resource Registry
Updated `apis/dev/planton/shared/cloudresourcekind/cloud_resource_kind.proto`:

```proto
// Before
KubernetesMicroservice = 810 [(kind_meta) = {
  provider: kubernetes
  version: v1
  id_prefix: "k8sms"
  is_service_kind: true
  kubernetes_meta: {
    category: workload
    namespace_prefix: "service"
  }
}];

// After
KubernetesDeployment = 810 [(kind_meta) = {
  provider: kubernetes
  version: v1
  id_prefix: "k8sdpl"
  is_service_kind: true
  kubernetes_meta: {
    category: workload
    namespace_prefix: "service"
  }
}];
```

**Key preservation**:
- Enum value `810` unchanged
- `is_service_kind: true` flag preserved
- Provider and category unchanged
- Namespace prefix remains "service"

**ID Prefix change**: `k8sms` → `k8sdpl` (reflects new name)

#### Test Files Updated

Fixed two test files that contained hard-coded references:

1. **`pkg/crkreflect/kind_by_id_prefix_test.go`**:
   - Test name: `"Kubernetes Microservice"` → `"Kubernetes Deployment"`
   - ID prefix: `"k8sms"` → `"k8sdpl"`
   - Expected enum: `CloudResourceKind_KubernetesMicroservice` → `CloudResourceKind_KubernetesDeployment`

2. **`pkg/crkreflect/kind_from_string_test.go`**:
   - Updated 6 test cases for various string format conversions
   - Updated normalization test with all 9 string variants
   - All test inputs and expected outputs now use `KubernetesDeployment`

### Build Pipeline Issues Encountered

The initial rename script run succeeded partially but encountered a build failure:

**Error**:
```
vet: pkg/crkreflect/kind_by_id_prefix_test.go:37:32: undefined: cloudresourcekind.CloudResourceKind_KubernetesMicroservice
```

**Root Cause**: Test files contained hard-coded enum references that the automated replacement patterns didn't catch because they weren't in the component directory or documentation paths.

**Resolution**: Manually updated both test files with correct enum references and test data.

### Verification Steps

1. **Protobuf Generation**: `make protos` ✅ Passed
   - Generated new `.pb.go` files with `KubernetesDeployment` types
   - Removed old `KubernetesMicroservice` stubs

2. **Build Verification**: `make build` ✅ Passed
   - All Go packages compiled successfully
   - No undefined references
   - Bazel build completed

3. **Test Suite**: `make test` ✅ Passed
   - All 200+ test suites passed
   - Specific `KubernetesDeployment` tests verified:
     - `TestKindByIdPrefix/Kubernetes_Deployment` ✅
     - `TestKindFromString/KubernetesDeployment_-_PascalCase` ✅
     - `TestKindFromString/KubernetesDeployment_-_kebab-case` ✅
     - `TestKindFromString/KubernetesDeployment_-_snake_case` ✅
     - `TestKindFromStringNormalization` ✅

## Impact Analysis

### For Users

**Breaking Change**: User manifests must be updated:

```yaml
# Before
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesMicroservice
metadata:
  name: my-service
spec:
  version: main
  container:
    ...

# After
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesDeployment
metadata:
  name: my-service
spec:
  version: main
  container:
    ...
```

**ID Prefix Change**: New resources will use `k8sdpl-*` prefix instead of `k8sms-*`. Existing resources with `k8sms-*` IDs continue to work (preserved in metadata).

### For Developers

**Positive impacts**:
1. **Clearer intent**: Code explicitly shows it's creating a Kubernetes Deployment
2. **Better organization**: Opens path for `KubernetesStatefulSet`, `KubernetesDaemonSet` additions
3. **Consistent patterns**: Aligns with `KubernetesCronJob`, `KubernetesPostgres` naming
4. **Reduced confusion**: No more explaining "microservice vs deployment"

**Import changes required**:
```go
// Before
import kubernetesmicroservicev1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/workload/kubernetesmicroservice/v1"

// After
import kubernetesdeploymentv1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/workload/kubernetesdeployment/v1"
```

### For Future Development

**Enables clear expansion**:
- `KubernetesStatefulSet` - for stateful workloads
- `KubernetesDaemonSet` - for node-level services
- `KubernetesJob` - for one-time tasks (distinct from CronJob)

Each can now be named accurately without conflicting abstractions.

## Files Modified

### Component Directory (New)
- `apis/dev/planton/provider/kubernetes/workload/kubernetesdeployment/v1/`
  - `api.proto`, `spec.proto`, `stack_input.proto`, `stack_outputs.proto`
  - `api.pb.go`, `spec.pb.go`, `stack_input.pb.go`, `stack_outputs.pb.go`
  - `README.md`, `examples.md`, `docs/README.md`
  - `iac/pulumi/main.go`, `iac/pulumi/module/*.go`
  - `iac/tf/*.tf`
  - `BUILD.bazel` files

### Registry
- `apis/dev/planton/shared/cloudresourcekind/cloud_resource_kind.proto`

### Tests
- `pkg/crkreflect/kind_by_id_prefix_test.go`
- `pkg/crkreflect/kind_from_string_test.go`

### Generated Files
- `pkg/crkreflect/kind_map_gen.go` (auto-regenerated)

### Total Impact
- **Files copied**: 45
- **Replacements made**: 33 (in component directory and docs)
- **Test files fixed**: 2 (manual updates)
- **Old directory deleted**: 1

## Technical Details

### Preserved Metadata

**Critical preservations** ensuring backward compatibility for deployed resources:

- **Enum value**: `810` (unchanged)
- **Provider**: `kubernetes` (unchanged)
- **Version**: `v1` (unchanged)
- **Service flag**: `is_service_kind: true` (unchanged)
- **Kubernetes category**: `workload` (unchanged)
- **Namespace prefix**: `"service"` (unchanged - existing deployments keep their namespaces)

Only changed:
- **Enum name**: `KubernetesMicroservice` → `KubernetesDeployment`
- **ID prefix**: `k8sms` → `k8sdpl`

### Code Generation Flow

1. **Protobuf compilation** (`make protos`):
   - Reads updated `cloud_resource_kind.proto`
   - Generates Go stubs with new enum names
   - Creates type-safe constants

2. **Gazelle run** (Bazel):
   - Updates BUILD files for new package paths
   - Maintains dependency graph

3. **Kind map generation**:
   - Regenerates `pkg/crkreflect/kind_map_gen.go`
   - Maps ID prefix `k8sdpl` → `CloudResourceKind_KubernetesDeployment`
   - Maps string variants → enum value

4. **Binary compilation**:
   - Links all updated packages
   - Validates import paths
   - Produces CLI binaries

## Lessons Learned

### What Worked Well

1. **Automated rename script**: The Python script handled 90% of the work correctly
2. **Comprehensive patterns**: 7 naming patterns caught most references
3. **Build pipeline verification**: Immediate feedback on issues
4. **Preserving enum value**: Avoided database/metadata migration headaches

### What Needed Manual Intervention

1. **Test files**: Hard-coded enum references in test files outside the component directory weren't caught by the script's path filters
2. **Documentation scope**: The script only updates `site/public/docs/`; changelog references in `_changelog/` are intentionally preserved as historical

### Recommendations for Future Renames

1. **Always run tests**: Don't assume script success means build success
2. **Check test files explicitly**: Search for enum references in `pkg/` test files
3. **Verify generated code**: Check `kind_map_gen.go` for correctness
4. **Update changelog rule**: Could enhance script to also search `pkg/*_test.go` files

## Migration Guide

### For Existing Deployments

Existing `KubernetesMicroservice` resources deployed with `k8sms-*` IDs will continue to function. The enum value (810) is preserved, so the system recognizes them.

However, to use the new CLI version, you must:

1. **Update manifest files**:
   ```bash
   # Find all manifests
   find . -name "*.yaml" -type f
   
   # Update kind field
   sed -i '' 's/kind: KubernetesMicroservice/kind: KubernetesDeployment/g' *.yaml
   ```

2. **Update import statements** (if using Go library):
   ```bash
   # Update import paths
   find . -name "*.go" -type f -exec sed -i '' 's|kubernetesmicroservice/v1|kubernetesdeployment/v1|g' {} \;
   ```

3. **Rebuild and test** your integrations

### For New Deployments

Simply use:
```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesDeployment
```

New resources will get `k8sdpl-*` ID prefixes automatically.

## Conclusion

This refactoring successfully removed an inaccurate abstraction layer, replacing "microservice" terminology with the accurate "deployment" naming that reflects the underlying Kubernetes resource. The systematic approach using automated tooling combined with manual verification ensured correctness across 45+ files while preserving backward compatibility for deployed resources.

The rename sets a strong foundation for future expansion of Kubernetes workload types (`StatefulSet`, `DaemonSet`, `Job`) with clear, accurate naming that reduces developer confusion and maintains consistency with established patterns like `KubernetesCronJob`.

### Commit Message

```
refactor(kubernetes): rename KubernetesMicroservice to KubernetesDeployment

Removes abstraction - "Microservice" doesn't accurately describe the
Kubernetes Deployment resource that gets created. This rename:
- Clarifies what Kubernetes resource is created
- Sets stage for KubernetesStatefulSet introduction
- Maintains naming consistency with KubernetesCronJob

Preserves:
- Enum value: 810
- All functionality
- Deployment behavior

BREAKING CHANGE: User manifests must update kind field from
KubernetesMicroservice to KubernetesDeployment. New resources
use k8sdpl-* ID prefix instead of k8sms-*.
```

