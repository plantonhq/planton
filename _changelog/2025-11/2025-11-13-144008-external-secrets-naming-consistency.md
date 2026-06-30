# ExternalSecrets Naming Consistency Refactoring

**Date**: November 13, 2025  
**Type**: Refactoring  
**Components**: API Definitions, Cloud Resource Registry, Documentation, Pulumi CLI Integration, Code Generation

## Summary

Renamed the `ExternalSecretsKubernetes` resource to `ExternalSecrets` across all proto definitions, documentation, and implementation code to align with Planton's naming conventions for Kubernetes addon operators. This change eliminates redundant "Kubernetes" suffixes and improves consistency with other recently refactored addon resources like `AltinityOperator`, `ElasticOperator`, and `ExternalDns`.

## Problem Statement / Motivation

The External Secrets Operator resource was originally named `ExternalSecretsKubernetes`, which included a redundant "Kubernetes" suffix. This naming pattern was inconsistent with Planton's design philosophy where:

### Pain Points

- **Redundant Context**: The resource lives under `provider/kubernetes/addon/`, making the "Kubernetes" suffix in the name redundant
- **Verbose API Surface**: Users had to write `kind: ExternalSecretsKubernetes` in manifests, which is unnecessarily long
- **Naming Inconsistency**: After recent refactorings (AltinityOperator, ElasticOperator, ExternalDns), this remained one of the few addons with a "Kubernetes" suffix
- **Code Verbosity**: Proto message types like `ExternalSecretsKubernetesSpec` and `ExternalSecretsKubernetesStackInput` were excessively long
- **Poor Developer Experience**: The redundancy made code harder to read and type

The provider namespace (`dev.planton.provider.kubernetes.addon.externalsecrets.v1`) already clearly indicates this is a Kubernetes component, so including "Kubernetes" in every message name adds noise without value.

## Solution / What's New

Performed a comprehensive rename from `ExternalSecretsKubernetes` to `ExternalSecrets` across:

1. **Proto API Definitions**: Updated all message types, field references, and validation constraints
2. **Cloud Resource Registry**: Modified the enum entry in `cloud_resource_kind.proto`
3. **Documentation**: Updated user-facing documentation with the new naming
4. **Implementation Code**: Modified Go code in Pulumi modules to use renamed types
5. **Code Generation**: Enhanced the kind map generator to properly handle directory mappings for Kubernetes addons
6. **Directory Structure**: Directory renamed from `externalsecretskubernetes/` to `externalsecrets/`

### Naming Convention

The new naming follows Planton's established pattern:

```yaml
# Before
apiVersion: kubernetes.planton.dev/v1
kind: ExternalSecretsKubernetes

# After
apiVersion: kubernetes.planton.dev/v1
kind: ExternalSecrets
```

The directory path is now `provider/kubernetes/addon/externalsecrets/` (renamed from `externalsecretskubernetes/`), aligning the directory name with the resource kind.

## Implementation Details

### Proto File Changes

**File**: `apis/dev/planton/provider/kubernetes/addon/externalsecrets/v1/api.proto`

```protobuf
// Before
message ExternalSecretsKubernetes {
  string kind = 2 [(buf.validate.field).string.const = 'ExternalSecretsKubernetes'];
  ExternalSecretsKubernetesSpec spec = 4;
  ExternalSecretsKubernetesStatus status = 5;
}

// After
message ExternalSecrets {
  string kind = 2 [(buf.validate.field).string.const = 'ExternalSecrets'];
  ExternalSecretsSpec spec = 4;
  ExternalSecretsStatus status = 5;
}
```

**File**: `apis/dev/planton/provider/kubernetes/addon/externalsecrets/v1/spec.proto`

```protobuf
// Before
message ExternalSecretsKubernetesSpec { ... }
message ExternalSecretsKubernetesSpecContainer { ... }

// After
message ExternalSecretsSpec { ... }
message ExternalSecretsSpecContainer { ... }
```

**File**: `apis/dev/planton/provider/kubernetes/addon/externalsecrets/v1/stack_input.proto`

```protobuf
// Before
message ExternalSecretsKubernetesStackInput {
  ExternalSecretsKubernetes target = 1;
}

// After
message ExternalSecretsStackInput {
  ExternalSecrets target = 1;
}
```

**File**: `apis/dev/planton/provider/kubernetes/addon/externalsecrets/v1/stack_outputs.proto`

```protobuf
// Before
message ExternalSecretsKubernetesStackOutputs { ... }

// After
message ExternalSecretsStackOutputs { ... }
```

### Registry Update

**File**: `apis/dev/planton/shared/cloudresourcekind/cloud_resource_kind.proto`

```protobuf
// Before
ExternalSecretsKubernetes = 829 [(kind_meta) = {
  provider: kubernetes
  version: v1
  id_prefix: "extseck8s"
  kubernetes_meta: {category: addon}
}];

// After
ExternalSecrets = 829 [(kind_meta) = {
  provider: kubernetes
  version: v1
  id_prefix: "extseck8s"
  kubernetes_meta: {category: addon}
}];
```

### Implementation Code Changes

**File**: `apis/dev/planton/provider/kubernetes/addon/externalsecrets/v1/iac/pulumi/main.go`

```go
// Before
stackInput := &externalsecretskubernetesv1.ExternalSecretsKubernetesStackInput{}

// After
stackInput := &externalsecretsv1.ExternalSecretsStackInput{}
```

**File**: `apis/dev/planton/provider/kubernetes/addon/externalsecrets/v1/iac/pulumi/module/main.go`

```go
// Before
func Resources(ctx *pulumi.Context, in *externalsecretsv1.ExternalSecretsKubernetesStackInput) error

// After
func Resources(ctx *pulumi.Context, in *externalsecretsv1.ExternalSecretsStackInput) error
```

### Code Generation Enhancement

**File**: `pkg/crkreflect/codegen/main.go`

Enhanced the kind map generator to properly handle directory name mappings for Kubernetes addons where the enum name differs from the directory name:

```go
// Mapping for kubernetes addon directories where enum name differs from directory name
dirName := lowerKind
if kubernetesResourceType == cloudresourcekind.KubernetesCloudResourceCategory_addon {
  addonDirMap := map[string]string{
    "altinityoperator":                "altinityoperator",
    "certmanagerkubernetes":           "certmanager",
    "externaldnskubernetes":           "externaldns",
    "externalsecrets":                 "externalsecrets",  // New mapping
    "ingressnginxkubernetes":          "ingressnginx",
    "kubernetesistio":                 "kubernetesistio",
    "strimzikafkaoperator":            "strimzikafkaoperator",
    "postgresoperatorkubernetes":      "zalandopostgresoperator",
    "zalandopostgresoperator":         "zalandopostgresoperator",
    // ... more mappings
  }
  if mapped, ok := addonDirMap[lowerKind]; ok {
    dirName = mapped
  }
}
```

This enhancement ensures that:
- The code generator correctly maps enum names to actual directory paths
- Import paths are generated correctly for all Kubernetes addons
- The kind map compiles without errors
- The mapping handles both old and new directory structures during transition

### Documentation Updates

Updated all occurrences in:
- `docs/README.md` - Updated protobuf message type references in examples from `ExternalSecretsKubernetesSpec` to `ExternalSecretsSpec`

### Build Process

1. **Proto Generation**: Ran `make protos` to regenerate Go stubs from updated proto files
2. **Kind Map Regeneration**: Ran `make generate-cloud-resource-kind-map` to update reflection mappings
3. **Compilation Verification**: Successfully compiled all Go packages
4. **Import Verification**: Confirmed no broken imports or references

## Benefits

### Cleaner API Surface

```yaml
# Users write less, understand more
kind: ExternalSecrets  # vs. kind: ExternalSecretsKubernetes
```

### Improved Code Readability

Proto message names are now more concise:
- `ExternalSecretsSpec` (was `ExternalSecretsKubernetesSpec`)
- `ExternalSecretsStackInput` (was `ExternalSecretsKubernetesStackInput`)
- `ExternalSecretsStackOutputs` (was `ExternalSecretsKubernetesStackOutputs`)

### Naming Consistency

Aligns with the pattern where the provider namespace provides sufficient context:
- Package: `dev.planton.provider.kubernetes.addon.externalsecrets.v1` (updated from `externalsecretskubernetes`)
- Directory: `provider/kubernetes/addon/externalsecrets/` (renamed from `externalsecretskubernetes/`)
- Kind: `ExternalSecrets` (context is already clear from package and directory)

Joins the family of consistently named addons:
- `AltinityOperator` (not AltinityOperatorKubernetes)
- `ElasticOperator` (not ElasticOperatorKubernetes)
- `ExternalDns` (not ExternalDnsKubernetes)
- `ExternalSecrets` (not ExternalSecretsKubernetes) ✅

### Developer Experience

- Shorter type names in code
- Less typing in YAML manifests
- Reduced cognitive load when reading code
- Consistent mental model across all Kubernetes addon operators
- Aligned directory structure with resource naming

## Impact

### User-Facing Changes

**Breaking Change**: Yes, for API consumers

Users must update their YAML manifests:

```yaml
# Update required
apiVersion: kubernetes.planton.dev/v1
kind: ExternalSecrets  # Changed from: ExternalSecretsKubernetes
metadata:
  name: external-secrets-prod
spec:
  # ... rest unchanged
```

### Non-Breaking Aspects

- **Import paths**: Import path base changed from `externalsecretskubernetes` to `externalsecrets` (follows directory rename)
- **Package namespace**: Updated to `dev.planton.provider.kubernetes.addon.externalsecrets.v1`
- **Functionality**: Zero behavioral changes
- **ID prefix**: Unchanged (`extseck8s`)
- **Enum value**: Unchanged (829)

### Developer Impact

- Proto stubs regenerated automatically
- Kind map regenerated with enhanced codegen logic
- Directory structure updated to match resource name
- No manual code migration required for internal implementation
- All existing tests continue to pass
- Enhanced code generator benefits future addon refactorings

## Related Work

This refactoring continues the naming consistency initiative across Planton's Kubernetes addon operators:

- **2025-11-13**: AltinityOperator naming consistency (completed)
- **2025-11-13**: ExternalSecrets naming consistency (this change)
- **Previous**: ElasticOperator, ExternalDns, CertManager, IngressNginx refactorings

Similar patterns should be evaluated for remaining addons to establish uniform naming conventions across the platform.

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

**Implementation** (2 files):
- `iac/pulumi/main.go` - Stack input type reference
- `iac/pulumi/module/main.go` - Function signature

**Code Generation** (1 file):
- `pkg/crkreflect/codegen/main.go` - Directory mapping logic

**Generated Files**:
- `*.pb.go` files (auto-regenerated from proto definitions)
- `pkg/crkreflect/kind_map_gen.go` (auto-generated from codegen)

**Total**: 9 manually modified files + generated artifacts

## Code Generation Enhancement

A key improvement in this refactoring was enhancing the kind map code generator to handle cases where Kubernetes addon enum names differ from their directory names. This enhancement:

- **Prevents future errors**: Automatically handles directory mappings for all addons
- **Improves maintainability**: Centralized mapping logic instead of scattered imports
- **Scales well**: New addon refactorings automatically benefit from the mapping
- **Self-documenting**: The map clearly shows which addons have been renamed
- **Flexible**: Supports both old and new directory structures during transition periods

This investment pays dividends for future naming consistency work across other Kubernetes addons and makes the codebase more resilient to directory structure changes.

## Migration Notes

For users with existing manifests:

1. **Find and replace** in all manifest files:
   ```bash
   find . -name "*.yaml" -type f -exec sed -i '' 's/kind: ExternalSecretsKubernetes/kind: ExternalSecrets/g' {} +
   ```

2. **No CLI changes required** - The `planton` CLI will work with the new kind name automatically after updating to the version with this change

3. **No infrastructure impact** - Existing deployed resources are unaffected; this only affects new deployments

4. **Directory structure** - The directory has been renamed from `externalsecretskubernetes/` to `externalsecrets/`, improving discoverability and consistency

---

**Status**: ✅ Production Ready  
**Files Changed**: 9 manual + generated artifacts  
**Build Status**: All tests passing, no linter errors, all packages compile successfully

