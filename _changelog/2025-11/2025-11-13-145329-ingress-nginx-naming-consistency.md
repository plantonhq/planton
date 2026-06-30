# IngressNginx Naming Consistency: Removing Redundant Kubernetes Suffix

**Date**: November 13, 2025  
**Type**: Breaking Change / Refactoring  
**Components**: API Definitions, Cloud Resource Registry, Provider Framework, Code Generation

## Summary

Completed a comprehensive rename of the Ingress NGINX controller component, removing the redundant "Kubernetes" suffix from all layers: package namespace, proto message types (`IngressNginxKubernetes` → `IngressNginx`), and API kind name. This refactoring improves naming consistency across Planton's Kubernetes addon operators and reduces unnecessary verbosity in user manifests and code.

## Problem Statement / Motivation

The Ingress NGINX controller component had "kubernetes" appearing redundantly in multiple places, creating unnecessary verbosity and inconsistency with Planton's naming conventions.

### Pain Points

- **Proto Messages**: `IngressNginxKubernetes`, `IngressNginxKubernetesSpec`, `IngressNginxKubernetesStackInput` - verbose suffixes
- **API Kind**: `kind: IngressNginxKubernetes` - unnecessarily long in user YAML manifests
- **Code References**: Every Go reference included the redundant suffix
- **Package Context**: The component's location under `provider/kubernetes/addon/` already indicates it's Kubernetes-specific
- **Inconsistency**: Other addon operators like `CertManager`, `ExternalDns`, `ExternalSecrets` don't have the suffix

The "kubernetes" suffix added no semantic value since the component's namespace path (`dev.planton.provider.kubernetes.addon.ingressnginx.v1`) already provides complete context.

## Solution / What's New

Performed a systematic refactoring across all component layers to remove the redundant suffix while preserving backward compatibility in non-breaking elements.

### Name Changes

```protobuf
// Before
message IngressNginxKubernetes {
  string kind = 2 [(buf.validate.field).string.const = 'IngressNginxKubernetes'];
  IngressNginxKubernetesSpec spec = 4;
  IngressNginxKubernetesStatus status = 5;
}

// After
message IngressNginx {
  string kind = 2 [(buf.validate.field).string.const = 'IngressNginx'];
  IngressNginxSpec spec = 4;
  IngressNginxStatus status = 5;
}
```

### Cloud Resource Registry Update

```protobuf
// Before
IngressNginxKubernetes = 824 [(kind_meta) = {
  provider: kubernetes
  version: v1
  id_prefix: "ngxk8s"
  kubernetes_meta: {category: addon}
}];

// After
IngressNginx = 824 [(kind_meta) = {
  provider: kubernetes
  version: v1
  id_prefix: "ngxk8s"
  kubernetes_meta: {category: addon}
}];
```

## Implementation Details

### Proto Files Updated

**File**: `apis/dev/planton/provider/kubernetes/addon/ingressnginx/v1/api.proto`

```protobuf
syntax = "proto3";

package dev.planton.provider.kubernetes.addon.ingressnginx.v1;

//ingress-nginx
message IngressNginx {
  string api_version = 1 [(buf.validate.field).string.const = 'kubernetes.planton.dev/v1'];
  string kind = 2 [(buf.validate.field).string.const = 'IngressNginx'];
  dev.planton.shared.CloudResourceMetadata metadata = 3;
  IngressNginxSpec spec = 4;
  IngressNginxStatus status = 5;
}

//ingress-nginx status.
message IngressNginxStatus {
  IngressNginxStackOutputs outputs = 1;
}
```

**File**: `apis/dev/planton/provider/kubernetes/addon/ingressnginx/v1/spec.proto`

```protobuf
// IngressNginxSpec defines configuration for ingress‑nginx on any cluster.
message IngressNginxSpec {
  dev.planton.shared.kubernetes.KubernetesAddonTargetCluster target_cluster = 1;
  string chart_version = 2;
  bool internal = 3;
  oneof provider_config {
    IngressNginxGkeConfig gke = 100;
    IngressNginxEksConfig eks = 101;
    IngressNginxAksConfig aks = 102;
  }
}
```

**File**: `apis/dev/planton/provider/kubernetes/addon/ingressnginx/v1/stack_input.proto`

```protobuf
//input for ingress-nginx stack
message IngressNginxStackInput {
  IngressNginx target = 1;
  dev.planton.provider.kubernetes.KubernetesProviderConfig provider_config = 2;
}
```

**File**: `apis/dev/planton/provider/kubernetes/addon/ingressnginx/v1/stack_outputs.proto`

```protobuf
// IngressNginxStackOutputs defines the outputs for the Ingress Nginx stack.
message IngressNginxStackOutputs {
  string namespace = 1;
  string release_name = 2;
  string service_name = 3;
  string service_type = 4;
}
```

### Go Implementation Changes

**File**: `apis/dev/planton/provider/kubernetes/addon/ingressnginx/v1/iac/pulumi/main.go`

```go
package main

import (
	"github.com/pkg/errors"
	ingressnginxv1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/addon/ingressnginx/v1"
	"github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/addon/ingressnginx/v1/iac/pulumi/module"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/stackinput"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		stackInput := &ingressnginxv1.IngressNginxStackInput{}
		if err := stackinput.LoadStackInput(ctx, stackInput); err != nil {
			return errors.Wrap(err, "failed to load stack-input")
		}
		return module.Resources(ctx, stackInput)
	})
}
```

**File**: `apis/dev/planton/provider/kubernetes/addon/ingressnginx/v1/iac/pulumi/module/main.go`

```go
// Resources creates all Pulumi resources for the Ingress‑Nginx add‑on.
func Resources(ctx *pulumi.Context,
	stackInput *ingressnginxv1.IngressNginxStackInput) error {
	
	kubeProvider, err := pulumikubernetesprovider.GetWithKubernetesProviderConfig(
		ctx, stackInput.ProviderConfig, "kubernetes")
	if err != nil {
		return errors.Wrap(err, "failed to set up kubernetes provider")
	}
	
	spec := stackInput.Target.Spec
	// ... implementation continues
}
```

### Code Generation and Build

1. **Proto Stub Regeneration**: 
   ```bash
   make protos
   ```
   Regenerated all `*.pb.go` files with updated type names

2. **Cloud Resource Kind Map**:
   ```bash
   make generate-cloud-resource-kind-map
   ```
   Updated `pkg/crkreflect/kind_map_gen.go` to map `CloudResourceKind_IngressNginx` to `&ingressnginxv1.IngressNginx{}`

3. **Gazelle Update**: 
   ```bash
   ./bazelw run //:gazelle
   ```
   Updated all `BUILD.bazel` files with new import paths

4. **Compilation Verification**: Go code builds successfully without errors

## Benefits

### Reduced Verbosity

**Proto Message Names**:
- `IngressNginxKubernetes` → `IngressNginx` (10 characters shorter)
- `IngressNginxKubernetesSpec` → `IngressNginxSpec` (10 characters shorter)
- `IngressNginxKubernetesStackInput` → `IngressNginxStackInput` (10 characters shorter)

**User Manifests**:
```yaml
# Before: 22 characters
kind: IngressNginxKubernetes

# After: 12 characters
kind: IngressNginx
```

### Improved Naming Consistency

Now aligns with Planton's pattern where provider namespace provides context:

```
✅ dev.planton.provider.kubernetes.addon.certmanager.v1     → CertManager
✅ dev.planton.provider.kubernetes.addon.externaldns.v1     → ExternalDns
✅ dev.planton.provider.kubernetes.addon.ingressnginx.v1    → IngressNginx
✅ dev.planton.provider.kubernetes.addon.externalsecrets.v1 → ExternalSecrets
```

The "kubernetes" context is clear from the provider path, not from redundant suffixes.

### Better Code Readability

**Go Type References**:
```go
// Before
stackInput := &ingressnginxv1.IngressNginxKubernetesStackInput{}

// After
stackInput := &ingressnginxv1.IngressNginxStackInput{}
```

Shorter type names make code more readable and reduce line length.

## Impact

### Breaking Changes

This is a **breaking change** affecting multiple layers:

#### 1. User Manifests

**Required Change**:
```yaml
# Before
apiVersion: kubernetes.planton.dev/v1
kind: IngressNginxKubernetes
metadata:
  name: my-ingress
spec:
  targetCluster:
    credentialId: prod-cluster
  internal: false
  gke:
    staticIpName: ingress-ip

# After
apiVersion: kubernetes.planton.dev/v1
kind: IngressNginx
metadata:
  name: my-ingress
spec:
  targetCluster:
    credentialId: prod-cluster
  internal: false
  gke:
    staticIpName: ingress-ip
```

Users must update the `kind` field in all IngressNginx manifest files.

#### 2. Import Paths (for SDK users)

**Before**:
```go
import ingressnginxv1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/addon/ingressnginx/v1"

stackInput := &ingressnginxv1.IngressNginxKubernetesStackInput{}
```

**After**:
```go
import ingressnginxv1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/addon/ingressnginx/v1"

stackInput := &ingressnginxv1.IngressNginxStackInput{}
```

#### 3. Proto Import Paths

Any custom proto files importing these definitions must update type names (though import paths remain the same):

```protobuf
// Message type references change
// Before: IngressNginxKubernetes
// After: IngressNginx
```

### Non-Breaking Aspects

- **Enum Value**: Still `824` in cloud_resource_kind.proto
- **ID Prefix**: Still `ngxk8s` for resource ID generation
- **API Version**: Still `kubernetes.planton.dev/v1`
- **Provider**: Still `kubernetes`
- **Package Path**: Still `dev.planton.provider.kubernetes.addon.ingressnginx.v1`
- **Functionality**: Zero behavioral changes to Ingress NGINX deployment

### Scope of Changes

**Proto Definitions**: 4 files
- api.proto, spec.proto, stack_input.proto, stack_outputs.proto

**Generated Code**: 4 files
- *.pb.go files (auto-regenerated)

**Implementation**: 2 files
- iac/pulumi/main.go
- iac/pulumi/module/main.go

**Registry**: 1 file
- cloud_resource_kind.proto

**Code Generation**: 1 file
- pkg/crkreflect/kind_map_gen.go (auto-regenerated)

**Build Files**: Multiple
- BUILD.bazel files (auto-updated via Gazelle)

**Total**: ~8 manually updated files + generated artifacts

## Migration Guide

### For CLI Users (Manifest Updates)

**Step 1**: Update kind in all manifest files

```bash
# Find all manifests with the old kind
find . -name "*.yaml" -type f -exec grep -l "kind: IngressNginxKubernetes" {} \;

# Update in place (macOS)
find . -name "*.yaml" -type f -exec sed -i '' 's/kind: IngressNginxKubernetes/kind: IngressNginx/g' {} +

# Update in place (Linux)
find . -name "*.yaml" -type f -exec sed -i 's/kind: IngressNginxKubernetes/kind: IngressNginx/g' {} +
```

**Step 2**: Deploy with new kind

```bash
planton pulumi up --manifest ingress-nginx.yaml
```

### For SDK Users (Go Code Updates)

**Step 1**: Update type references throughout your codebase

```go
// Replace type names
- stackInput := &ingressnginxv1.IngressNginxKubernetesStackInput{}
+ stackInput := &ingressnginxv1.IngressNginxStackInput{}

- var ingress *ingressnginxv1.IngressNginxKubernetes
+ var ingress *ingressnginxv1.IngressNginx

- spec *ingressnginxv1.IngressNginxKubernetesSpec
+ spec *ingressnginxv1.IngressNginxSpec
```

**Step 2**: Recompile and test

```bash
go mod tidy
go build ./...
go test ./...
```

## Related Work

This refactoring is part of a broader initiative to improve naming consistency across Planton's Kubernetes addon operators:

### Recent Similar Refactorings

- **AltinityOperator** (2025-11-13): Removed "Kubernetes" suffix
- **StrimziKafkaOperator** (2025-11-13): Naming consistency improvements
- **ApacheSolrOperator** (2025-11-13): Naming consistency improvements
- **KubernetesIstio** (2025-11-13): Naming consistency improvements
- **ExternalSecrets** (2025-11-13): Naming consistency improvements
- **ElasticOperator** (2025-11-13): Naming consistency improvements
- **ZalandoPostgresOperator** (2025-11-13): Naming consistency improvements

### Established Pattern

This change reinforces the pattern for addon operators:
- ✅ **Package**: `dev.planton.provider.kubernetes.addon.{operatorname}.v1`
- ✅ **Kind**: `{OperatorName}` (no "Kubernetes" suffix)
- ✅ **Message**: `{OperatorName}`, `{OperatorName}Spec`, `{OperatorName}StackInput`, etc.

### Future Consistency

With this change, all Kubernetes addon operators now follow consistent naming:
- `CertManager`, `ExternalDns`, `ExternalSecrets`
- `ElasticOperator`, `AltinityOperator`, `StrimziKafkaOperator`
- `ApacheSolrOperator`, `ZalandoPostgresOperator`
- `IngressNginx`, `KubernetesIstio`

## Technical Notes

### Provider-Specific Configuration Preserved

The IngressNginx component supports cloud-specific configuration for GKE, EKS, and AKS. This refactoring only changed type names—all provider-specific fields remain unchanged:

- **GKE**: `IngressNginxGkeConfig` with static IP and subnetwork configuration
- **EKS**: `IngressNginxEksConfig` with security groups, subnets, and IRSA role
- **AKS**: `IngressNginxAksConfig` with managed identity and public IP

### Import Path Resolution

Go module resolution automatically handles the type name changes because:
1. The import path (`github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/addon/ingressnginx/v1`) remains unchanged
2. Only the exported type names within the package changed
3. Proto generation updates all cross-references automatically
4. Gazelle updates BUILD.bazel files to reflect new type names

No manual import resolution was required beyond updating type references.

---

**Status**: ✅ Production Ready  
**Breaking Change**: Yes - requires manifest updates from `IngressNginxKubernetes` to `IngressNginx`  
**Timeline**: Completed November 13, 2025  
**Files Changed**: ~8 manual files + generated artifacts  
**Build Status**: All code compiles successfully, no linter errors

