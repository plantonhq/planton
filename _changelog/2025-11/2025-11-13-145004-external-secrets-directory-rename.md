# ExternalSecrets Directory Naming Consistency

**Date**: November 13, 2025  
**Type**: Refactoring  
**Components**: Directory Structure, Build Configuration

## Summary

Renamed the directory from `externalsecretskubernetes` to `externalsecrets` to align with Planton's naming conventions for Kubernetes addon operators. Unlike other addons, ExternalSecrets proto definitions were already correctly named without the redundant "Kubernetes" suffix—this change simply brings the directory structure into alignment with the API naming.

## Problem Statement / Motivation

The External Secrets Operator addon directory was named `externalsecretskubernetes`, which included a redundant "Kubernetes" suffix. However, the proto API definitions were already correctly named:

### Pain Points

- **Directory Name Inconsistency**: Directory was `externalsecretskubernetes/` but API kind was already `ExternalSecrets`
- **Mixed Signals**: Directory name suggested "Kubernetes" suffix but API was clean
- **Pattern Mismatch**: Other correctly-named addons (like `altinityoperator/`, `elasticoperator/`) use clean directory names

The proto message types were already correctly defined as `ExternalSecrets`, `ExternalSecretsSpec`, `ExternalSecretsStackInput`, etc.—only the directory name needed updating.

## Solution / What's New

Renamed the directory from `externalsecretskubernetes/` to `externalsecrets/` to match:

1. **Directory Structure**: Updated from `provider/kubernetes/addon/externalsecretskubernetes/` to `provider/kubernetes/addon/externalsecrets/`
2. **Build Configuration**: BUILD.bazel files automatically updated via Gazelle
3. **Package Paths**: Updated Go import paths from `...externalsecretskubernetes/v1` to `...externalsecrets/v1`

### What Was Already Correct

The proto definitions were already using the correct naming:

```protobuf
// Already correct - no changes needed
message ExternalSecrets {
  string kind = 2 [(buf.validate.field).string.const = 'ExternalSecrets'];
  ExternalSecretsSpec spec = 4;
  ExternalSecretsStatus status = 5;
}
```

## Implementation Details

### Directory Structure Change

```
# Before
apis/dev/planton/provider/kubernetes/addon/externalsecretskubernetes/v1/
├── api.proto
├── spec.proto
├── stack_input.proto
├── stack_outputs.proto
└── iac/

# After
apis/dev/planton/provider/kubernetes/addon/externalsecrets/v1/
├── api.proto
├── spec.proto
├── stack_input.proto
├── stack_outputs.proto
└── iac/
```

### Package Namespace

The package namespace was updated to match the new directory:

```protobuf
// Before
package dev.planton.provider.kubernetes.addon.externalsecretskubernetes.v1;

// After
package dev.planton.provider.kubernetes.addon.externalsecrets.v1;
```

### Go Import Paths

```go
// Before
externalsecretsv1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/addon/externalsecretskubernetes/v1"

// After
externalsecretsv1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/addon/externalsecrets/v1"
```

### Registry Entry

The cloud resource registry entry was already correct:

```protobuf
// Already correct - no suffix
ExternalSecrets = 829 [(kind_meta) = {
  provider: kubernetes
  version: v1
  id_prefix: "extseck8s"
  kubernetes_meta: {category: addon}
}];
```

## Benefits

### Naming Consistency

- Directory name now matches the API kind: `externalsecrets/` ↔ `ExternalSecrets`
- Eliminates confusion between directory name and API surface
- Aligns with other well-named addons (AltinityOperator, ElasticOperator)

### Developer Experience

- Intuitive directory structure matches proto definitions
- No mental translation needed between filesystem and API
- Consistent pattern across all addon operators

## Impact

### User-Facing Changes

**Breaking Change**: No

The API kind was already `ExternalSecrets`, so users' YAML manifests require no changes:

```yaml
# Users continue to use the same kind name
apiVersion: kubernetes.planton.dev/v1
kind: ExternalSecrets  # Already correct
metadata:
  name: external-secrets-prod
spec:
  # ... no changes needed
```

### Developer Impact

- **Proto stubs**: Regenerated automatically via `make protos`
- **Import paths**: Updated in Go code
- **BUILD files**: Updated via Gazelle
- **Functionality**: Zero behavioral changes

### Non-Breaking Aspects

- **API surface**: Unchanged (already correct)
- **Message types**: Unchanged (already correct)
- **Kind validation**: Unchanged (already correct)
- **Functionality**: Zero behavioral changes
- **ID prefix**: Unchanged (`extseck8s`)
- **Enum value**: Unchanged (829)

## Files Modified

**Directory Structure**:
- Renamed `externalsecretskubernetes/` → `externalsecrets/`

**Proto Files** (package declarations updated):
- `api.proto`
- `spec.proto`
- `stack_input.proto`
- `stack_outputs.proto`

**Implementation** (import paths updated):
- `iac/pulumi/main.go`
- `iac/pulumi/module/main.go`

**Generated Files**:
- `*.pb.go` files (auto-regenerated with updated package paths)
- `BUILD.bazel` files (auto-updated via Gazelle)

**Total**: 6 manually updated files (primarily package declarations) + generated artifacts

## Related Work

This directory rename completes the naming consistency effort:
- **AltinityOperator** (2025-11-13): Already had clean directory and API naming
- **ElasticOperator**: Already had clean directory and API naming
- **ExternalDns** (2025-11-13): Renamed from ExternalDnsKubernetes to ExternalDns
- **CertManager** (2025-11-13): Renamed from CertManagerKubernetes to CertManager
- **ExternalSecrets**: Directory renamed to match already-correct API naming

All Kubernetes addon operators now follow consistent naming patterns.

## Migration Notes

For users:

1. **No manifest changes required** - The API kind was already `ExternalSecrets`
2. **No CLI changes required** - Everything continues to work as before
3. **No infrastructure impact** - Purely a development-side refactoring

For developers updating code:

1. **Update import paths** in Go code if directly importing the package
2. **Rebuild** to regenerate proto stubs with updated paths
3. **Run Gazelle** to update BUILD files

---

**Status**: ✅ Production Ready  
**Files Changed**: 6 manual (package paths) + generated artifacts  
**Build Status**: All builds passing, no linter errors  
**User Impact**: None - API was already correctly named

