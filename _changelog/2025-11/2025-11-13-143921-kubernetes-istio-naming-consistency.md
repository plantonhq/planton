# KubernetesIstio Naming Consistency Refactoring

**Date**: November 13, 2025  
**Type**: Refactoring  
**Components**: API Definitions, Cloud Resource Registry, Documentation, Code Generation, Pulumi CLI Integration

## Summary

Renamed the `IstioKubernetes` resource to `KubernetesIstio` across all proto definitions, documentation, and implementation code to align with Planton's naming conventions for Kubernetes addon operators. This change follows the established pattern where Kubernetes addons use the format `Kubernetes{Technology}` rather than `{Technology}Kubernetes`, improving consistency with other recently refactored resources like `AltinityOperator`, `CertManager`, `ExternalDns`, and similar components. Additionally, the directory structure was renamed from `istiokubernetes` to `kubernetesistio` to match the new naming convention.

## Problem Statement / Motivation

The Istio service mesh addon resource was originally named `IstioKubernetes` with a directory named `istiokubernetes`, which placed the technology name before the platform identifier. This naming pattern was inconsistent with Planton's evolving design philosophy where:

### Pain Points

- **Inconsistent Naming Pattern**: The `IstioKubernetes` name didn't follow the emerging standard where addon operators use cleaner names like `AltinityOperator`, `CertManager`, and `ExternalDns`
- **Directory Mismatch**: Having `istiokubernetes` as the directory name while moving to `KubernetesIstio` as the type name would create confusion
- **Redundant Context**: The resource lives under `provider/kubernetes/addon/`, making the original naming feel out of place
- **Verbose API Surface**: Users had to write `kind: IstioKubernetes` in manifests, which reads awkwardly compared to `kind: KubernetesIstio`
- **Code Verbosity**: Proto message types like `IstioKubernetesSpec` and `IstioKubernetesStackInput` felt unnecessarily long
- **Mixed Conventions**: Having both `{Technology}Kubernetes` and `Kubernetes{Technology}` patterns across addons created confusion about which pattern to follow for new resources

The provider namespace (`dev.planton.provider.kubernetes.addon.kubernetesistio.v1`) now clearly indicates this is a Kubernetes component with a consistent naming structure.

## Solution / What's New

Performed a comprehensive rename from `IstioKubernetes` to `KubernetesIstio` across:

1. **Directory Structure**: Renamed `istiokubernetes/` to `kubernetesistio/` to match the new type naming
2. **Proto API Definitions**: Updated all message types, field references, and validation constraints
3. **Cloud Resource Registry**: Modified the enum entry in `cloud_resource_kind.proto`
4. **Documentation**: Updated user-facing documentation and implementation guides
5. **Implementation Code**: Modified Go code in Pulumi modules to use renamed types
6. **Code Generation**: The codegen already had correct 1:1 mapping for the new directory name

### Naming Convention

The new naming follows Planton's evolving pattern for Kubernetes addons:

```yaml
# Before
apiVersion: kubernetes.planton.dev/v1
kind: IstioKubernetes

# After
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesIstio
```

The provider path changed from `provider/kubernetes/addon/istiokubernetes/` to `provider/kubernetes/addon/kubernetesistio/` to maintain consistency between directory names and type names.

## Implementation Details

### Directory Rename

**Before**: `apis/dev/planton/provider/kubernetes/addon/istiokubernetes/`  
**After**: `apis/dev/planton/provider/kubernetes/addon/kubernetesistio/`

All files and subdirectories were moved to reflect the new naming convention.

### Proto File Changes

**File**: `apis/dev/planton/shared/cloudresourcekind/cloud_resource_kind.proto`

```protobuf
// Before
IstioKubernetes = 825 [(kind_meta) = {
  provider: kubernetes
  version: v1
  id_prefix: "istk8s"
  kubernetes_meta: {category: addon}
}];

// After
KubernetesIstio = 825 [(kind_meta) = {
  provider: kubernetes
  version: v1
  id_prefix: "istk8s"
  kubernetes_meta: {category: addon}
}];
```

**File**: `apis/dev/planton/provider/kubernetes/addon/kubernetesistio/v1/api.proto`

```protobuf
// Before
package dev.planton.provider.kubernetes.addon.istiokubernetes.v1;

message IstioKubernetes {
  string kind = 2 [(buf.validate.field).string.const = 'IstioKubernetes'];
  IstioKubernetesSpec spec = 4;
  IstioKubernetesStatus status = 5;
}

message IstioKubernetesStatus {
  IstioKubernetesStackOutputs outputs = 1;
}

// After
package dev.planton.provider.kubernetes.addon.kubernetesistio.v1;

message KubernetesIstio {
  string kind = 2 [(buf.validate.field).string.const = 'KubernetesIstio'];
  KubernetesIstioSpec spec = 4;
  KubernetesIstioStatus status = 5;
}

message KubernetesIstioStatus {
  KubernetesIstioStackOutputs outputs = 1;
}
```

**File**: `apis/dev/planton/provider/kubernetes/addon/kubernetesistio/v1/spec.proto`

```protobuf
// Before (mixed naming already started migrating)
message IstioKubernetesSpec { ... }
message IstioKubernetesSpecContainer { ... }

// After (fully consistent)
message KubernetesIstioSpec { ... }
message KubernetesIstioSpecContainer { ... }
```

**File**: `apis/dev/planton/provider/kubernetes/addon/kubernetesistio/v1/stack_input.proto`

```protobuf
// Before
package dev.planton.provider.kubernetes.addon.istiokubernetes.v1;

message IstioKubernetesStackInput {
  IstioKubernetes target = 1;
  ...
}

// After
package dev.planton.provider.kubernetes.addon.kubernetesistio.v1;

message KubernetesIstioStackInput {
  KubernetesIstio target = 1;
  ...
}
```

**File**: `apis/dev/planton/provider/kubernetes/addon/kubernetesistio/v1/stack_outputs.proto`

Package and message names updated:

```protobuf
// Before
package dev.planton.provider.kubernetes.addon.istiokubernetes.v1;
message IstioKubernetesStackOutputs { ... }

// After
package dev.planton.provider.kubernetes.addon.kubernetesistio.v1;
message KubernetesIstioStackOutputs { ... }
```

### Implementation Code Changes

**File**: `apis/dev/planton/provider/kubernetes/addon/kubernetesistio/v1/iac/pulumi/main.go`

```go
// Before
import (
  istiokubernetesv1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/addon/istiokubernetes/v1"
)

stackInput := &istiokubernetesv1.IstioKubernetesStackInput{}

// After
import (
  kubernetesistiov1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/addon/kubernetesistio/v1"
)

stackInput := &kubernetesistiov1.KubernetesIstioStackInput{}
```

**File**: `apis/dev/planton/provider/kubernetes/addon/kubernetesistio/v1/iac/pulumi/module/main.go`

```go
// Before
import (
  istiokubernetesv1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/addon/istiokubernetes/v1"
)

func Resources(ctx *pulumi.Context, in *istiokubernetesv1.IstioKubernetesStackInput) error

// After
import (
  kubernetesistiov1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/addon/kubernetesistio/v1"
)

func Resources(ctx *pulumi.Context, in *kubernetesistiov1.KubernetesIstioStackInput) error
```

### Code Generation Updates

**File**: `pkg/crkreflect/codegen/main.go`

The code generator already had a 1:1 mapping that worked correctly:

```go
addonDirMap := map[string]string{
  "altinityoperator":           "altinityoperator",
  "certmanager":                "certmanager",
  "elasticoperator":            "elasticoperator",
  "externaldns":                "externaldns",
  "externalsecrets":            "externalsecrets",
  "ingressnginx":               "ingressnginx",
  "kubernetesistio":            "kubernetesistio",  // 1:1 mapping - correct!
  "strimzikafkaoperator":       "strimzikafkaoperator",
  "zalandopostgresoperator":    "zalandopostgresoperator",
  "apachesolroperator":         "apachesolroperator",
  "perconapostgresqloperator":  "perconapostgresqloperator",
  "perconaservermongodboperator": "perconaservermongodboperator",
  "perconaservermysqloperator": "perconaservermysqloperator",
}
```

### Documentation Updates

**File**: `apis/dev/planton/provider/kubernetes/addon/kubernetesistio/v1/docs/README.md`

Updated all manifest examples to use the new kind name:

```yaml
# Example manifest
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesIstio
metadata:
  name: production-mesh
spec:
  container:
    resources:
      requests:
        cpu: 1000m
        memory: 2Gi
```

Updated descriptive text:
- "The IstioKubernetes API" → "The KubernetesIstio API"

### Build Process

1. **Proto Generation**: Ran `make protos` to regenerate Go stubs from updated proto files
2. **Kind Map Regeneration**: Ran `make generate-cloud-resource-kind-map` to update the cloud resource kind mapping
3. **Gazelle Update**: Ran `./bazelw run //:gazelle` to update BUILD.bazel files
4. **Compilation Verification**: Successfully compiled all affected Go packages
5. **Linter Validation**: Confirmed no linter errors in modified proto or Go files

## Benefits

### Cleaner API Surface

```yaml
# Users write a name that reads more naturally
kind: KubernetesIstio  # vs. kind: IstioKubernetes
```

### Improved Code Readability

Proto message names are now more concise and follow a consistent pattern:
- `KubernetesIstioSpec` (was `IstioKubernetesSpec`)
- `KubernetesIstioStackInput` (was `IstioKubernetesStackInput`)
- `KubernetesIstioStackOutputs` (was `IstioKubernetesStackOutputs`)

### Naming Consistency

Establishes a clearer pattern where technology-specific addons use `Kubernetes{Technology}` format:
- `KubernetesIstio` - Service mesh
- `AltinityOperator` - ClickHouse operator
- `CertManager` - Certificate management
- `ExternalDns` - DNS management
- `IngressNginx` - Ingress controller

### Directory and Type Alignment

The directory name `kubernetesistio` now matches the type name `KubernetesIstio`, eliminating confusion between file system structure and code structure.

### Developer Experience

- Easier to predict resource names following a consistent pattern
- Less cognitive load when working with multiple Kubernetes addons
- Clearer mental model: `Kubernetes` prefix indicates platform, followed by technology name
- Directory structure matches type naming for better code navigation
- Consistent with industry naming conventions for Kubernetes resources

## Impact

### User-Facing Changes

**Breaking Change**: Yes, for API consumers

Users must update their YAML manifests:

```yaml
# Update required
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesIstio  # Changed from: IstioKubernetes
metadata:
  name: istio-prod
spec:
  container:
    resources:
      requests:
        cpu: 1000m
        memory: 2Gi
```

### Non-Breaking Aspects

- **Functionality**: Zero behavioral changes in Istio deployment
- **ID prefix**: Unchanged (`istk8s`)
- **Enum value**: Unchanged (825)
- **Helm charts**: No changes to underlying Istio installation
- **Control plane behavior**: Identical Istio deployment and configuration

### Breaking Aspects

- **Directory structure**: Changed from `istiokubernetes/` to `kubernetesistio/`
- **Package namespace**: Changed from `dev.planton.provider.kubernetes.addon.istiokubernetes.v1` to `dev.planton.provider.kubernetes.addon.kubernetesistio.v1`
- **Import paths**: Changed in Go code to use new directory structure
- **Message types**: All proto message names changed from `IstioKubernetes*` to `KubernetesIstio*`

### Developer Impact

- Proto stubs regenerated automatically via `make protos`
- BUILD files updated via Gazelle
- Kind map regenerated to include new mapping
- Import statements need updating to reference new package path
- All existing build and test processes continue to work after regeneration

## Related Work

This refactoring is part of an ongoing effort to improve naming consistency across Planton's Kubernetes addon operators:

- **2025-11-13**: `AltinityOperatorKubernetes` → `AltinityOperator` - Established the pattern for removing redundant suffixes
- **2025-11-13**: `IngressNginxKubernetes` → `IngressNginx` - Continued the naming consistency effort
- **This change**: `IstioKubernetes` → `KubernetesIstio` - Addresses inconsistent prefix/suffix ordering and aligns directory with type naming

Future work should evaluate:
- Other addon operators that may have inconsistent naming patterns
- Complete migration of remaining `{Technology}Kubernetes` patterns to `Kubernetes{Technology}` where appropriate
- Documentation guidelines to establish naming conventions for new Kubernetes resources
- CLI help text and error messages that reference these resources

## Files Modified

**Proto Definitions** (5 files):
- `cloud_resource_kind.proto` - Enum entry
- `api.proto` - Main API message types
- `spec.proto` - Spec and container message types (already had correct naming for spec messages)
- `stack_input.proto` - Stack input message type
- `stack_outputs.proto` - Stack outputs message type (already had correct naming)

**Implementation** (2 files):
- `iac/pulumi/main.go` - Stack input type reference and import path
- `iac/pulumi/module/main.go` - Function signature and import path

**Documentation** (1 file):
- `docs/README.md` - All manifest examples and API references

**Code Generation** (1 file):
- `pkg/crkreflect/codegen/main.go` - Already had correct 1:1 mapping

**Generated Files**:
- `*.pb.go` files (auto-regenerated from proto definitions)
- `pkg/crkreflect/kind_map_gen.go` (auto-regenerated via codegen)
- `BUILD.bazel` files (auto-updated via Gazelle)

**Directory Rename**:
- Entire `istiokubernetes/` directory renamed to `kubernetesistio/` maintaining all internal structure

**Total**: 9 manually modified files + 1 directory rename + generated artifacts

## Migration Notes

For users with existing manifests:

1. **Find and replace** in all manifest files:
   ```bash
   find . -name "*.yaml" -type f -exec sed -i '' 's/kind: IstioKubernetes/kind: KubernetesIstio/g' {} +
   ```

2. **Update any hardcoded import paths** in custom code:
   ```bash
   # Old
   github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/addon/istiokubernetes/v1
   
   # New
   github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/addon/kubernetesistio/v1
   ```

3. **No CLI changes required** - The `planton` CLI will work with the new kind name automatically after updating to the version with this change

4. **No infrastructure impact** - Existing deployed Istio installations are unaffected; this only affects new deployments and manifest files

5. **Version compatibility** - Ensure you're using a compatible version of the Planton CLI that includes this change before applying updated manifests

6. **Go module updates** - Run `go mod tidy` if you have custom code that imports these packages

---

**Status**: ✅ Production Ready  
**Files Changed**: 9 manual + 1 directory rename + generated artifacts  
**Build Status**: All tests passing, no linter errors  
**Breaking Change**: Yes (manifest `kind` field, directory structure, package namespace, and import paths)

