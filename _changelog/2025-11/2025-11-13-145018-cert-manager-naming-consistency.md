# CertManager Naming Consistency: Removing Redundant Kubernetes Suffix

**Date**: November 13, 2025  
**Type**: Breaking Change / Refactoring  
**Components**: API Definitions, Kubernetes Provider, Build System, Documentation

## Summary

Completed a comprehensive rename of the cert-manager Kubernetes addon component, removing the redundant "Kubernetes" suffix from `CertManagerKubernetes` to `CertManager` across all layers: proto message types, cloud resource registry enum, implementation code, tests, and documentation. This refactoring aligns with Planton's naming conventions where the `provider/kubernetes/addon/` path context already makes the Kubernetes association clear, eliminating unnecessary verbosity throughout the codebase and in user-facing manifests.

## Problem Statement / Motivation

The cert-manager addon was structured with "Kubernetes" appearing redundantly in multiple places, creating unnecessary verbosity without adding semantic value.

### Pain Points

- **Proto Message Names**: `CertManagerKubernetes`, `CertManagerKubernetesSpec`, `CertManagerKubernetesStatus`, `CertManagerKubernetesStackInput`, `CertManagerKubernetesStackOutputs` - all included redundant suffix
- **API Kind**: `kind: CertManagerKubernetes` - verbose in user manifests
- **Cloud Resource Enum**: `CertManagerKubernetes = 821` - inconsistent with other addons
- **Code References**: Every Go import and type reference included the redundant suffix
- **Path Context Ignored**: The component lives under `provider/kubernetes/addon/certmanager/v1/` - the "kubernetes" suffix in the type name added no information
- **Inconsistency**: Other recent addons (AltinityOperator, ElasticOperator) already followed the simpler naming pattern

The component's location under `provider/kubernetes/addon/` already establishes it as a Kubernetes component. The "Kubernetes" suffix in type names was pure redundancy.

## Solution / What's New

Performed a systematic, multi-layer refactoring following the established pattern from the AltinityOperator rename:

### Scope of Changes

**Proto Definitions** (4 files):
- `api.proto` - Main resource definition
- `spec.proto` - Configuration specification  
- `stack_outputs.proto` - Output definition
- `stack_input.proto` - Stack input definition

**Cloud Resource Registry** (1 file):
- `cloud_resource_kind.proto` - Enum value update

**Implementation Code** (2 files):
- `iac/pulumi/main.go` - Pulumi entry point
- `iac/pulumi/module/main.go` - Pulumi resource module

**Tests** (1 file):
- `api_test.go` - Validation tests

**Documentation** (1 file):
- `README.md` - User documentation with YAML examples

**Generated Artifacts**:
- All `.pb.go` files regenerated via `make protos`
- All `BUILD.bazel` files updated via Gazelle

### Key Naming Changes

```protobuf
// Before
message CertManagerKubernetes {
  string kind = 2 [(buf.validate.field).string.const = 'CertManagerKubernetes'];
  CertManagerKubernetesSpec spec = 4;
  CertManagerKubernetesStatus status = 5;
}

message CertManagerKubernetesSpec { ... }
message CertManagerKubernetesStatus { ... }
message CertManagerKubernetesStackInput { ... }
message CertManagerKubernetesStackOutputs { ... }

// After
message CertManager {
  string kind = 2 [(buf.validate.field).string.const = 'CertManager'];
  CertManagerSpec spec = 4;
  CertManagerStatus status = 5;
}

message CertManagerSpec { ... }
message CertManagerStatus { ... }
message CertManagerStackInput { ... }
message CertManagerStackOutputs { ... }
```

## Implementation Details

### Proto File Changes

**File**: `apis/dev/planton/provider/kubernetes/addon/certmanager/v1/api.proto`

```protobuf
//cert-manager
message CertManager {
  //api-version
  string api_version = 1 [(buf.validate.field).string.const = 'kubernetes.planton.dev/v1'];

  //resource-kind
  string kind = 2 [(buf.validate.field).string.const = 'CertManager'];

  //metadata
  dev.planton.shared.CloudResourceMetadata metadata = 3 [(buf.validate.field).required = true];

  //spec
  CertManagerSpec spec = 4 [(buf.validate.field).required = true];

  //status
  CertManagerStatus status = 5;
}

//cert-manager status.
message CertManagerStatus {
  //stack-outputs
  CertManagerStackOutputs outputs = 1;
}
```

**File**: `apis/dev/planton/provider/kubernetes/addon/certmanager/v1/spec.proto`

```protobuf
// CertManagerSpec defines configuration for cert-manager on any cluster.
message CertManagerSpec {
  dev.planton.shared.kubernetes.KubernetesAddonTargetCluster target_cluster = 1;
  optional string namespace = 2;
  optional string cert_manager_version = 3;
  optional string helm_chart_version = 4;
  bool skip_install_self_signed_issuer = 5;
  AcmeConfig acme = 6;
  repeated DnsProviderConfig dns_providers = 7;
}
```

**File**: `apis/dev/planton/provider/kubernetes/addon/certmanager/v1/stack_input.proto`

```protobuf
//input for cert-manager stack
message CertManagerStackInput {
  //target cloud-resource
  CertManager target = 1;
  //provider-config
  dev.planton.provider.kubernetes.KubernetesProviderConfig provider_config = 2;
}
```

**File**: `apis/dev/planton/provider/kubernetes/addon/certmanager/v1/stack_outputs.proto`

```protobuf
// Outputs emitted after cert‑manager installation.
message CertManagerStackOutputs {
  string namespace = 1;
  string release_name = 2;
  string solver_identity = 3;
  string cloudflare_secret_name = 4;
}
```

### Cloud Resource Registry Update

**File**: `apis/dev/planton/shared/cloudresourcekind/cloud_resource_kind.proto`

```protobuf
// Before
CertManagerKubernetes = 821 [(kind_meta) = {
  provider: kubernetes
  version: v1
  id_prefix: "cmk8s"
  kubernetes_meta: {category: addon}
}];

// After
CertManager = 821 [(kind_meta) = {
  provider: kubernetes
  version: v1
  id_prefix: "cmk8s"
  kubernetes_meta: {category: addon}
}];
```

### Go Implementation Updates

**File**: `apis/dev/planton/provider/kubernetes/addon/certmanager/v1/iac/pulumi/main.go`

```go
func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		stackInput := &certmanagerv1.CertManagerStackInput{}

		if err := stackinput.LoadStackInput(ctx, stackInput); err != nil {
			return errors.Wrap(err, "failed to load stack-input")
		}

		return module.Resources(ctx, stackInput)
	})
}
```

**File**: `apis/dev/planton/provider/kubernetes/addon/certmanager/v1/iac/pulumi/module/main.go`

```go
// Resources create all Pulumi resources for the Cert‑Manager Kubernetes add‑on.
func Resources(ctx *pulumi.Context, stackInput *certmanagerv1.CertManagerStackInput) error {
	// ... implementation using stackInput.Target.Spec
}

func createClusterIssuerForDomain(
	ctx *pulumi.Context,
	kubeProvider *kubernetes.Provider,
	helmRelease *helm.Release,
	spec *certmanagerv1.CertManagerSpec,
	cloudflareSecrets map[string]pulumi.StringOutput,
	dnsProvider *certmanagerv1.DnsProviderConfig,
	domain string,
) error {
	// ... implementation
}
```

### Test Updates

**File**: `apis/dev/planton/provider/kubernetes/addon/certmanager/v1/api_test.go`

```go
func TestCertManager(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "CertManager Suite")
}

var _ = ginkgo.Describe("CertManager Custom Validation Tests", func() {
	var input *CertManager

	ginkgo.BeforeEach(func() {
		input = &CertManager{
			ApiVersion: "kubernetes.planton.dev/v1",
			Kind:       "CertManager",
			Metadata: &shared.CloudResourceMetadata{
				Name: "test-cert-manager",
			},
			Spec: &CertManagerSpec{
				Acme: &AcmeConfig{
					Email: "admin@example.com",
				},
				DnsProviders: []*DnsProviderConfig{
					{
						Name:     "cloudflare-test",
						DnsZones: []string{"example.com"},
						Provider: &DnsProviderConfig_Cloudflare{
							Cloudflare: &CloudflareProvider{
								ApiToken: "test-token",
							},
						},
					},
				},
			},
		}
	})

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("cert_manager", func() {
			ginkgo.It("should not return a validation error", func() {
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})
})
```

### Documentation Updates

**File**: `apis/dev/planton/provider/kubernetes/addon/certmanager/v1/README.md`

Updated all references from `CertManagerKubernetes` to `CertManager`, including:

- Main heading: `# CertManager` (removed "Kubernetes" suffix)
- YAML examples throughout the document
- Configuration references in text
- FAQ section references

**Example YAML** (appears multiple times in README):

```yaml
# Before
apiVersion: kubernetes.planton.dev/v1
kind: CertManagerKubernetes
metadata:
  name: cert-manager
spec:
  targetCluster:
    kubernetesProviderConfigId: my-cluster
  acme:
    email: "admin@example.com"
  dnsProviders:
    - name: cloudflare-prod
      dnsZones:
        - example.com
      cloudflare:
        apiToken: "your-token"

# After
apiVersion: kubernetes.planton.dev/v1
kind: CertManager
metadata:
  name: cert-manager
spec:
  targetCluster:
    kubernetesProviderConfigId: my-cluster
  acme:
    email: "admin@example.com"
  dnsProviders:
    - name: cloudflare-prod
      dnsZones:
        - example.com
      cloudflare:
        apiToken: "your-token"
```

### Build and Verification

**Proto Generation**:
```bash
cd ~/scm/github.com/plantonhq/planton
make protos
```

Output:
- All `.pb.go` files regenerated with updated package imports
- Gazelle updated all `BUILD.bazel` files with new paths
- Go module resolution handled import path changes automatically

**Test Execution**:
```bash
go test ./apis/dev/planton/provider/kubernetes/addon/certmanager/v1/...
```

Result: ✅ All tests passed

**Build Verification**:
```bash
go build ./apis/dev/planton/provider/kubernetes/addon/certmanager/v1/...
```

Result: ✅ Build successful with no errors

## Benefits

### Reduced Verbosity

**Proto Message Names** (10 characters shorter):
- `CertManagerKubernetes` → `CertManager`
- `CertManagerKubernetesSpec` → `CertManagerSpec`
- `CertManagerKubernetesStatus` → `CertManagerStatus`
- `CertManagerKubernetesStackInput` → `CertManagerStackInput`
- `CertManagerKubernetesStackOutputs` → `CertManagerStackOutputs`

**User Manifests**:
```yaml
kind: CertManager  # vs. kind: CertManagerKubernetes
```
10 characters shorter in every manifest file.

**Go Code Readability**:
```go
// Before
stackInput := &certmanagerv1.CertManagerKubernetesStackInput{}
spec := stackInput.Target.Spec  // type: *CertManagerKubernetesSpec

// After
stackInput := &certmanagerv1.CertManagerStackInput{}
spec := stackInput.Target.Spec  // type: *CertManagerSpec
```

### Naming Consistency

Now aligns with Planton's established pattern where provider namespace provides context:

```
✅ dev.planton.provider.kubernetes.addon.certmanager.v1      → CertManager
✅ dev.planton.provider.kubernetes.addon.externaldns.v1      → ExternalDns
✅ dev.planton.provider.kubernetes.addon.altinityoperator.v1 → AltinityOperator
✅ dev.planton.provider.kubernetes.addon.elasticoperator.v1  → ElasticOperator
```

The "kubernetes" context is clear from the provider path, not from redundant suffixes.

### Developer Experience

- **Less typing** in code and manifests
- **Easier to read** import statements and type references
- **Clearer mental model** when working with addon operators
- **Faster code navigation** with shorter paths
- **Consistent patterns** across all Kubernetes addons

## Impact

### Breaking Changes

This is a **major breaking change** affecting multiple layers:

#### 1. User Manifests

**Required Change**:
```yaml
# Before
apiVersion: kubernetes.planton.dev/v1
kind: CertManagerKubernetes
metadata:
  name: my-cert-manager
spec:
  # ... configuration

# After
apiVersion: kubernetes.planton.dev/v1
kind: CertManager
metadata:
  name: my-cert-manager
spec:
  # ... configuration
```

**Migration Steps**:
1. Find all manifests: `find . -name "*.yaml" -exec grep -l "kind: CertManagerKubernetes" {} \;`
2. Update kind field: `sed -i 's/kind: CertManagerKubernetes/kind: CertManager/g' *.yaml`
3. Validate: `planton validate --manifest cert-manager.yaml`
4. Deploy: `planton pulumi up --manifest cert-manager.yaml`

#### 2. SDK Users (Go)

**Import Path Changes**:
```go
// No change needed - package path remains the same
import (
  certmanagerv1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/addon/certmanager/v1"
)
```

**Type Reference Changes**:
```go
// Before
var cm *certmanagerv1.CertManagerKubernetes
var spec *certmanagerv1.CertManagerKubernetesSpec
var input *certmanagerv1.CertManagerKubernetesStackInput
var outputs *certmanagerv1.CertManagerKubernetesStackOutputs

// After
var cm *certmanagerv1.CertManager
var spec *certmanagerv1.CertManagerSpec
var input *certmanagerv1.CertManagerStackInput
var outputs *certmanagerv1.CertManagerStackOutputs
```

#### 3. Proto Consumers

**Proto Import Paths** (unchanged):
```protobuf
// No change needed - package path remains the same
import "dev/planton/provider/kubernetes/addon/certmanager/v1/api.proto";
```

**Message Type References**:
```protobuf
// Before
CertManagerKubernetes cert_manager = 1;
CertManagerKubernetesSpec spec = 2;

// After
CertManager cert_manager = 1;
CertManagerSpec spec = 2;
```

### Non-Breaking Aspects

- **Enum Value**: Still `821` in `cloud_resource_kind.proto`
- **ID Prefix**: Still `cmk8s` for resource ID generation
- **API Version**: Still `kubernetes.planton.dev/v1`
- **Provider**: Still `kubernetes`
- **Package Path**: Still `dev.planton.provider.kubernetes.addon.certmanager.v1`
- **Functionality**: Zero behavioral changes to cert-manager deployment or operation

### Scope of Changes

**Proto Definitions**: 4 files  
**Generated Code**: 4 files (`*.pb.go`, auto-regenerated)  
**Implementation**: 2 files (Pulumi main.go and module)  
**Tests**: 1 file  
**Documentation**: 1 file  
**Registry**: 1 file (cloud_resource_kind.proto)  
**Build Files**: Multiple `BUILD.bazel` files (auto-updated via Gazelle)

**Total**: ~15 files manually updated + generated artifacts

## Related Work

### Established Pattern

This refactoring follows the same pattern as the AltinityOperator rename (completed earlier on the same branch):

**Reference**: `_changelog/2025-11/2025-11-13-143427-altinity-operator-complete-rename.md`

Both refactorings are part of a broader initiative to improve naming consistency across Planton's Kubernetes addon operators:

### Pattern for Addon Operators

- ✅ **Directory**: `provider/kubernetes/addon/{operatorname}/`
- ✅ **Package**: `dev.planton.provider.kubernetes.addon.{operatorname}.v1`
- ✅ **Kind**: `{OperatorName}` (no "Kubernetes" suffix)
- ✅ **Message**: `{OperatorName}`, `{OperatorName}Spec`, etc.

### Future Work

This pattern can be applied to other resources that may have similar naming redundancies:
- Evaluate all Kubernetes addons for naming consistency
- Apply the same pattern where appropriate
- Update documentation guidelines for new resources

### Branch Context

This work was completed on branch: `refactor/rename-all-kubernetes-addons-to-remove-kubernetes-suffix`

This indicates a comprehensive effort to apply consistent naming across multiple addon operators.

## Migration Guide

### For End Users (CLI/Manifest Updates)

**Step 1**: Identify affected manifests

```bash
# Find all cert-manager manifests
find . -name "*.yaml" -exec grep -l "kind: CertManagerKubernetes" {} \;
```

**Step 2**: Update the kind field

```bash
# Option 1: Interactive replacement (recommended)
# Manually edit each file to change:
# kind: CertManagerKubernetes → kind: CertManager

# Option 2: Automated replacement (verify before using)
find . -name "*.yaml" -exec sed -i '' 's/kind: CertManagerKubernetes/kind: CertManager/g' {} +
```

**Step 3**: Validate manifests

```bash
planton validate --manifest cert-manager.yaml
```

**Step 4**: Deploy with new kind

```bash
# Preview first
planton pulumi preview --manifest cert-manager.yaml

# Apply
planton pulumi up --manifest cert-manager.yaml
```

### For SDK Users (Go Code Updates)

**Step 1**: Update type references in your code

```go
// Replace in all Go files
import (
  certmanagerv1 "github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/addon/certmanager/v1"
)

// Update type references
- var cm *certmanagerv1.CertManagerKubernetes
+ var cm *certmanagerv1.CertManager

- var spec *certmanagerv1.CertManagerKubernetesSpec
+ var spec *certmanagerv1.CertManagerSpec

- var input *certmanagerv1.CertManagerKubernetesStackInput
+ var input *certmanagerv1.CertManagerStackInput

- var outputs *certmanagerv1.CertManagerKubernetesStackOutputs
+ var outputs *certmanagerv1.CertManagerStackOutputs
```

**Step 2**: Update go.mod

```bash
go mod tidy
```

**Step 3**: Verify compilation

```bash
go build ./...
go test ./...
```

### For Proto Consumers

**Step 1**: Update message type references

```protobuf
// Replace in all proto files
- CertManagerKubernetes cert_manager = 1;
+ CertManager cert_manager = 1;

- CertManagerKubernetesSpec spec = 2;
+ CertManagerSpec spec = 2;
```

**Step 2**: Regenerate proto stubs

```bash
buf generate
```

## Technical Notes

### Why This Pattern?

The redundant "Kubernetes" suffix emerged from early design decisions before naming patterns were fully established. As the project matured, a clearer pattern emerged:

**Provider path provides context** → The namespace `provider/kubernetes/addon/` unambiguously indicates this is a Kubernetes addon. The type name doesn't need to repeat this information.

**Consistency matters** → As more addons were added (AltinityOperator, ElasticOperator, etc.), the simpler naming pattern became standard. CertManagerKubernetes was an outlier.

**Verbosity has costs** → Every extra character makes code harder to read, manifests longer, and developer experience worse. When that verbosity provides no semantic value, it should be removed.

### Import Path Stability

One advantage of this refactoring: **Go import paths remain unchanged**. The package path `dev.planton.provider.kubernetes.addon.certmanager.v1` stays the same, only the exported type names change. This minimizes disruption for SDK users.

### Git History Preservation

Proto and Go file changes maintain Git history since they're in-place edits, not file renames. History tracking remains intact for:
- `git log --follow api.proto`
- `git blame api.proto`
- IDE history navigation

---

**Status**: ✅ Production Ready  
**Breaking Change**: Yes - requires manifest and code updates  
**Timeline**: Completed November 13, 2025  
**Branch**: `refactor/rename-all-kubernetes-addons-to-remove-kubernetes-suffix`  
**Files Changed**: ~15 manual files + generated artifacts  
**Build Status**: All tests passing, proto generation successful, Go build verified

