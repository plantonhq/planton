# ExternalDns Naming Consistency Refactoring

**Date**: November 13, 2025  
**Type**: Refactoring  
**Components**: API Definitions, Cloud Resource Registry, Documentation, Pulumi CLI Integration

## Summary

Renamed the `ExternalDnsKubernetes` resource to `ExternalDns` across all proto definitions, documentation, and implementation code to align with Planton's naming conventions for Kubernetes addon operators. This change eliminates redundant "Kubernetes" suffixes and improves consistency with other addon resources that are already scoped to Kubernetes via their provider namespace.

## Problem Statement / Motivation

The ExternalDNS addon resource was originally named `ExternalDnsKubernetes`, which included a redundant "Kubernetes" suffix. This naming pattern was inconsistent with Planton's design philosophy where:

### Pain Points

- **Redundant Context**: The resource lives under `provider/kubernetes/addon/`, making the "Kubernetes" suffix in the name redundant
- **Verbose API Surface**: Users had to write `kind: ExternalDnsKubernetes` in manifests, which is unnecessarily long
- **Naming Inconsistency**: Mixed naming patterns across addon operators—some with suffixes, some without
- **Code Verbosity**: Proto message types like `ExternalDnsKubernetesSpec` and `ExternalDnsKubernetesStackInput` were excessively long
- **Poor Developer Experience**: The redundancy made code harder to read and type

The provider namespace (`dev.planton.provider.kubernetes.addon.externaldns.v1`) already clearly indicates this is a Kubernetes component, so including "Kubernetes" in every message name adds noise without value.

## Solution / What's New

Performed a comprehensive rename from `ExternalDnsKubernetes` to `ExternalDns` across:

1. **Proto API Definitions**: Updated all message types, field references, and validation constraints
2. **Cloud Resource Registry**: Modified the enum entry in `cloud_resource_kind.proto`
3. **Documentation**: Updated all user-facing docs and implementation guides
4. **Implementation Code**: Modified Go code in Pulumi modules to use renamed types

### Naming Convention

The new naming follows Planton's established pattern:

```yaml
# Before
apiVersion: kubernetes.planton.dev/v1
kind: ExternalDnsKubernetes

# After
apiVersion: kubernetes.planton.dev/v1
kind: ExternalDns
```

The provider path remains unchanged (`provider/kubernetes/addon/externaldns/`) to avoid breaking existing directory structures and import paths.

## Implementation Details

### Proto File Changes

**File**: `apis/dev/planton/provider/kubernetes/addon/externaldns/v1/api.proto`

```protobuf
// Before
message ExternalDnsKubernetes {
  string kind = 2 [(buf.validate.field).string.const = 'ExternalDnsKubernetes'];
  ExternalDnsKubernetesSpec spec = 4;
  ExternalDnsKubernetesStatus status = 5;
}

// After
message ExternalDns {
  string kind = 2 [(buf.validate.field).string.const = 'ExternalDns'];
  ExternalDnsSpec spec = 4;
  ExternalDnsStatus status = 5;
}
```

**File**: `apis/dev/planton/provider/kubernetes/addon/externaldns/v1/spec.proto`

```protobuf
// Before
message ExternalDnsKubernetesSpec { ... }

// After
message ExternalDnsSpec { ... }
```

**File**: `apis/dev/planton/provider/kubernetes/addon/externaldns/v1/stack_input.proto`

```protobuf
// Before
message ExternalDnsKubernetesStackInput {
  ExternalDnsKubernetes target = 1;
}

// After
message ExternalDnsStackInput {
  ExternalDns target = 1;
}
```

**File**: `apis/dev/planton/provider/kubernetes/addon/externaldns/v1/stack_outputs.proto`

```protobuf
// Before
message ExternalDnsKubernetesStackOutputs { ... }

// After
message ExternalDnsStackOutputs { ... }
```

### Registry Update

**File**: `apis/dev/planton/shared/cloudresourcekind/cloud_resource_kind.proto`

```protobuf
// Before
ExternalDnsKubernetes = 823 [(kind_meta) = {
  provider: kubernetes
  version: v1
  id_prefix: "extdnsk8s"
  kubernetes_meta: {category: addon}
}];

// After
ExternalDns = 823 [(kind_meta) = {
  provider: kubernetes
  version: v1
  id_prefix: "extdnsk8s"
  kubernetes_meta: {category: addon}
}];
```

### Implementation Code Changes

**File**: `apis/dev/planton/provider/kubernetes/addon/externaldns/v1/iac/pulumi/main.go`

```go
// Before
stackInput := &externaldnsv1.ExternalDnsKubernetesStackInput{}

// After
stackInput := &externaldnsv1.ExternalDnsStackInput{}
```

**File**: `apis/dev/planton/provider/kubernetes/addon/externaldns/v1/iac/pulumi/module/main.go`

```go
// Before
func Resources(ctx *pulumi.Context, stackInput *externaldnsv1.ExternalDnsKubernetesStackInput) error

// After
func Resources(ctx *pulumi.Context, stackInput *externaldnsv1.ExternalDnsStackInput) error
```

### Documentation Updates

Updated all occurrences in:
- `docs/README.md` (main component documentation)

## Benefits

### Cleaner API Surface

```yaml
# Users write less, understand more
kind: ExternalDns  # vs. kind: ExternalDnsKubernetes
```

### Improved Code Readability

Proto message names are now more concise:
- `ExternalDnsSpec` (was `ExternalDnsKubernetesSpec`)
- `ExternalDnsStackInput` (was `ExternalDnsKubernetesStackInput`)
- `ExternalDnsStackOutputs` (was `ExternalDnsKubernetesStackOutputs`)

### Naming Consistency

Aligns with the pattern where the provider namespace provides sufficient context:
- Package: `dev.planton.provider.kubernetes.addon.externaldns.v1`
- Kind: `ExternalDns` (context is already clear from package)

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
kind: ExternalDns  # Changed from: ExternalDnsKubernetes
metadata:
  name: external-dns-cloudflare
spec:
  # ... rest unchanged
```

### Non-Breaking Aspects

- **Directory structure**: Unchanged (`provider/kubernetes/addon/externaldns/`)
- **Package namespace**: Unchanged (`dev.planton.provider.kubernetes.addon.externaldns.v1`)
- **Import paths**: Unchanged in Go code
- **Functionality**: Zero behavioral changes
- **ID prefix**: Unchanged (`extdnsk8s`)
- **Enum value**: Unchanged (823)

### Developer Impact

- Proto stubs regenerated automatically via `make protos`
- BUILD files updated via Gazelle
- No manual code migration required for internal implementation
- All existing functionality preserved

## Files Modified

**Proto Definitions** (4 files):
- `api.proto` - Main API message types
- `spec.proto` - Spec message type
- `stack_input.proto` - Stack input message type
- `stack_outputs.proto` - Stack outputs message type

**Registry** (1 file):
- `cloud_resource_kind.proto` - Enum entry

**Documentation** (1 file):
- `docs/README.md` - Component documentation

**Implementation** (2 files):
- `iac/pulumi/main.go` - Stack input type reference
- `iac/pulumi/module/main.go` - Function signature

**Generated Files**:
- `*.pb.go` files (auto-regenerated from proto definitions)
- `BUILD.bazel` files (auto-updated via Gazelle)

**Total**: 8 manually modified files + generated artifacts

## Related Work

This refactoring follows the pattern established by:
- **AltinityOperator** (2025-11-13): Already follows the cleaner naming pattern without Kubernetes suffix
- **ElasticOperator** (enum value 822): Already follows the cleaner naming pattern without Kubernetes suffix

This is part of an ongoing effort to improve naming consistency across Planton's Kubernetes addon operators. Similar patterns should be evaluated for:
- Other addon operators that may have redundant suffixes
- Future addon components to follow this cleaner naming convention
- Documentation guidelines to establish naming patterns for new resources

## Migration Notes

For users with existing manifests:

1. **Find and replace** in all manifest files:
   ```bash
   find . -name "*.yaml" -type f -exec sed -i 's/kind: ExternalDnsKubernetes/kind: ExternalDns/g' {} +
   ```

2. **No CLI changes required** - The `planton` CLI will work with the new kind name automatically after updating to the version with this change

3. **No infrastructure impact** - Existing deployed resources are unaffected; this only affects new deployments

---

**Status**: ✅ Production Ready  
**Files Changed**: 8 manual + generated artifacts  
**Build Status**: All builds passing, no linter errors

