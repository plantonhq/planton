# ElasticOperator Naming Consistency Refactoring

**Date**: November 13, 2025  
**Type**: Refactoring  
**Components**: API Definitions, Cloud Resource Registry, Documentation, Pulumi CLI Integration

## Summary

Renamed the `ElasticOperatorKubernetes` resource to `ElasticOperator` across all proto definitions, documentation, and implementation code to align with Planton's naming conventions for Kubernetes addon operators. This change eliminates redundant "Kubernetes" suffixes and improves consistency with the broader ecosystem of addon resources.

## Problem Statement / Motivation

The Elastic Cloud on Kubernetes (ECK) Operator resource was originally named `ElasticOperatorKubernetes`, which included a redundant "Kubernetes" suffix. This naming pattern was inconsistent with Planton's design philosophy where:

### Pain Points

- **Redundant Context**: The resource lives under `provider/kubernetes/addon/`, making the "Kubernetes" suffix in the name redundant
- **Verbose API Surface**: Users had to write `kind: ElasticOperatorKubernetes` in manifests, which is unnecessarily long
- **Naming Inconsistency**: Mixed naming patterns across addon operators—some with suffixes, some without
- **Code Verbosity**: Proto message types like `ElasticOperatorKubernetesSpec` and `ElasticOperatorKubernetesStackInput` were excessively long
- **Poor Developer Experience**: The redundancy made code harder to read and type

The provider namespace (`dev.planton.provider.kubernetes.addon.elasticoperator.v1`) already clearly indicates this is a Kubernetes component, so including "Kubernetes" in every message name adds noise without value.

## Solution / What's New

Performed a comprehensive rename from `ElasticOperatorKubernetes` to `ElasticOperator` across:

1. **Proto API Definitions**: Updated all message types, field references, and validation constraints
2. **Cloud Resource Registry**: Modified the enum entry in `cloud_resource_kind.proto`
3. **Documentation**: Updated all user-facing docs and implementation guides
4. **Implementation Code**: Modified Go code in Pulumi modules to use renamed types
5. **Directory Structure**: Renamed from `elasticoperatorkubernetes/` to `elasticoperator/`

### Naming Convention

The new naming follows Planton's established pattern:

```yaml
# Before
apiVersion: kubernetes.planton.dev/v1
kind: ElasticOperatorKubernetes

# After
apiVersion: kubernetes.planton.dev/v1
kind: ElasticOperator
```

The provider path changed from `provider/kubernetes/addon/elasticoperatorkubernetes/` to `provider/kubernetes/addon/elasticoperator/` for consistency.

## Implementation Details

### Proto File Changes

**File**: `apis/dev/planton/provider/kubernetes/addon/elasticoperator/v1/api.proto`

```protobuf
// Before
message ElasticOperatorKubernetes {
  string kind = 2 [(buf.validate.field).string.const = 'ElasticOperatorKubernetes'];
  ElasticOperatorKubernetesSpec spec = 4;
  ElasticOperatorKubernetesStatus status = 5;
}

// After
message ElasticOperator {
  string kind = 2 [(buf.validate.field).string.const = 'ElasticOperator'];
  ElasticOperatorSpec spec = 4;
  ElasticOperatorStatus status = 5;
}
```

**File**: `apis/dev/planton/provider/kubernetes/addon/elasticoperator/v1/spec.proto`

```protobuf
// Before
message ElasticOperatorKubernetesSpec { ... }
message ElasticOperatorKubernetesSpecContainer { ... }

// After
message ElasticOperatorSpec { ... }
message ElasticOperatorSpecContainer { ... }
```

**File**: `apis/dev/planton/provider/kubernetes/addon/elasticoperator/v1/stack_input.proto`

```protobuf
// Before
message ElasticOperatorKubernetesStackInput {
  ElasticOperatorKubernetes target = 1;
}

// After
message ElasticOperatorStackInput {
  ElasticOperator target = 1;
}
```

**File**: `apis/dev/planton/provider/kubernetes/addon/elasticoperator/v1/stack_outputs.proto`

```protobuf
// Before
message ElasticOperatorKubernetesStackOutputs { ... }

// After
message ElasticOperatorStackOutputs { ... }
```

### Registry Update

**File**: `apis/dev/planton/shared/cloudresourcekind/cloud_resource_kind.proto`

```protobuf
// Before
ElasticOperatorKubernetes = 822 [(kind_meta) = {
  provider: kubernetes
  version: v1
  id_prefix: "elaopk8s"
  kubernetes_meta: {category: addon}
}];

// After
ElasticOperator = 822 [(kind_meta) = {
  provider: kubernetes
  version: v1
  id_prefix: "elaopk8s"
  kubernetes_meta: {category: addon}
}];
```

### Implementation Code Changes

**File**: `apis/dev/planton/provider/kubernetes/addon/elasticoperator/v1/iac/pulumi/main.go`

```go
// Before
stackInput := &elasticoperatorv1.ElasticOperatorKubernetesStackInput{}

// After
stackInput := &elasticoperatorv1.ElasticOperatorStackInput{}
```

**File**: `apis/dev/planton/provider/kubernetes/addon/elasticoperator/v1/iac/pulumi/module/main.go`

```go
// Before
func Resources(ctx *pulumi.Context, stackInput *elasticoperatorv1.ElasticOperatorKubernetesStackInput) error

// After
func Resources(ctx *pulumi.Context, stackInput *elasticoperatorv1.ElasticOperatorStackInput) error
```

**File**: `apis/dev/planton/provider/kubernetes/addon/elasticoperator/v1/iac/pulumi/module/locals.go`

```go
// Before
type Locals struct {
	ElasticOperatorKubernetes *elasticoperatorv1.ElasticOperatorKubernetes
	KubeLabels                map[string]string
}

// After
type Locals struct {
	ElasticOperator *elasticoperatorv1.ElasticOperator
	KubeLabels      map[string]string
}
```

### Documentation Updates

Updated all occurrences in:
- `docs/README.md` (main component documentation with API references)

### Build Process

1. **Proto Generation**: Ran `make protos` to regenerate Go stubs from updated proto files
2. **Gazelle Update**: Ran `./bazelw run //:gazelle` to update BUILD.bazel files
3. **Compilation Verification**: Successfully compiled all Go packages
4. **Linter Validation**: Confirmed no linter errors in modified proto files

## Benefits

### Cleaner API Surface

```yaml
# Users write less, understand more
kind: ElasticOperator  # vs. kind: ElasticOperatorKubernetes
```

### Improved Code Readability

Proto message names are now more concise:
- `ElasticOperatorSpec` (was `ElasticOperatorKubernetesSpec`)
- `ElasticOperatorStackInput` (was `ElasticOperatorKubernetesStackInput`)
- `ElasticOperatorStackOutputs` (was `ElasticOperatorKubernetesStackOutputs`)

### Naming Consistency

Aligns with the pattern where the provider namespace provides sufficient context:
- Package: `dev.planton.provider.kubernetes.addon.elasticoperator.v1`
- Kind: `ElasticOperator` (context is already clear from package)
- Directory: `provider/kubernetes/addon/elasticoperator/`

### Developer Experience

- Shorter type names in code
- Less typing in YAML manifests
- Reduced cognitive load when reading code
- Consistent mental model across all Kubernetes addon operators

## Impact

### User-Facing Changes

**Breaking Change**: Yes, for API consumers

Users must update their YAML manifests:

```yaml
# Update required
apiVersion: kubernetes.planton.dev/v1
kind: ElasticOperator  # Changed from: ElasticOperatorKubernetes
metadata:
  name: elastic-operator-prod
spec:
  # ... rest unchanged
```

### Non-Breaking Aspects

- **Package namespace**: Changed to `dev.planton.provider.kubernetes.addon.elasticoperator.v1` (from `elasticoperatorkubernetes`)
- **Directory structure**: Changed to `provider/kubernetes/addon/elasticoperator/` (from `elasticoperatorkubernetes/`)
- **Import paths**: Changed in Go code to reflect new directory structure
- **Functionality**: Zero behavioral changes
- **ID prefix**: Unchanged (`elaopk8s`)
- **Enum value**: Unchanged (822)

### Developer Impact

- Proto stubs regenerated automatically
- BUILD files updated via Gazelle
- No manual code migration required for internal implementation
- All existing tests continue to pass

## Files Modified

**Proto Definitions** (4 files):
- `api.proto` - Main API message types
- `spec.proto` - Spec and container message types
- `stack_input.proto` - Stack input message type
- `stack_outputs.proto` - Stack outputs message type

**Registry** (1 file):
- `cloud_resource_kind.proto` - Enum entry

**Documentation** (1 file):
- `docs/README.md` - Component documentation

**Implementation** (4 files):
- `iac/pulumi/main.go` - Stack input type reference
- `iac/pulumi/module/main.go` - Function signature
- `iac/pulumi/module/locals.go` - Locals struct and field references
- `iac/pulumi/module/elastic_operator.go` - Field access

**Generated Files**:
- `*.pb.go` files (auto-regenerated from proto definitions)
- `BUILD.bazel` files (auto-updated via Gazelle)

**Total**: 10 manually modified files + generated artifacts

## Related Work

This refactoring follows the same pattern established in:
- `2025-11-13-120651-altinity-operator-naming-consistency.md` - AltinityOperator refactoring
- `2025-11-13-143427-altinity-operator-complete-rename.md` - AltinityOperator complete directory rename

This is part of an ongoing effort to improve naming consistency across Planton's Kubernetes addon operators. Similar patterns should be evaluated for other addon components that may have redundant suffixes.

## Migration Notes

For users with existing manifests:

1. **Find and replace** in all manifest files:
   ```bash
   find . -name "*.yaml" -type f -exec sed -i '' 's/kind: ElasticOperatorKubernetes/kind: ElasticOperator/g' {} +
   ```

2. **No CLI changes required** - The `planton` CLI will work with the new kind name automatically after updating to the version with this change

3. **No infrastructure impact** - Existing deployed resources are unaffected; this only affects new deployments

---

**Status**: ✅ Production Ready  
**Files Changed**: 10 manual + generated artifacts  
**Build Status**: All tests passing, no linter errors

