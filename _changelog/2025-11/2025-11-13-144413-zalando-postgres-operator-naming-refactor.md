# ZalandoPostgresOperator Naming Consistency Refactoring

**Date**: November 13, 2025  
**Type**: Refactoring  
**Components**: API Definitions, Cloud Resource Registry, Documentation, Pulumi CLI Integration

## Summary

Renamed the `PostgresOperatorKubernetes` resource to `ZalandoPostgresOperator` across all proto definitions, documentation, and implementation code to align with Planton's naming conventions for Kubernetes addon operators. This change eliminates redundant "Kubernetes" suffixes, clarifies that this is specifically the Zalando implementation (not Percona or other operators), and improves consistency with other addon resources like `CertManager`, `ExternalDns`, and `IngressNginx` that are already scoped to Kubernetes via their provider namespace. The directory structure was also updated from `postgresoperatorkubernetes` to `zalandopostgresoperator`.

## Problem Statement / Motivation

The Postgres operator resource was originally named `PostgresOperatorKubernetes` with a directory path of `postgresoperatorkubernetes`, which had two issues: a redundant "Kubernetes" suffix and an ambiguous operator vendor identity. This naming pattern was inconsistent with Planton's design philosophy and created confusion about which specific Postgres operator implementation was being deployed.

### Pain Points

- **Redundant Context**: The resource lives under `provider/kubernetes/addon/`, making the "Kubernetes" suffix in the name redundant
- **Ambiguous Identity**: "PostgresOperator" doesn't indicate that this is specifically the **Zalando Postgres Operator**, not Percona PostgreSQL Operator, CloudNativePG, or Crunchy Postgres Operator
- **Verbose API Surface**: Users had to write `kind: PostgresOperatorKubernetes` in manifests, which is unnecessarily long
- **Naming Inconsistency**: Mixed naming patterns across addon operators—some with suffixes, some without
- **Code Verbosity**: Proto message types like `PostgresOperatorKubernetesSpec` and `PostgresOperatorKubernetesStackInput` were excessively long
- **Poor Developer Experience**: The redundancy made code harder to read and type
- **Vendor Clarity**: In an ecosystem with multiple PostgreSQL operators (Zalando, Percona, CloudNativePG, Crunchy), the name should clearly indicate which one is being deployed

The provider namespace (`dev.planton.provider.kubernetes.addon.zalandopostgresoperator.v1`) now clearly indicates this is a Kubernetes component, and the resource should explicitly name the vendor (Zalando) to distinguish it from other PostgreSQL operator implementations in the Planton ecosystem.

## Solution / What's New

Performed a comprehensive rename from `PostgresOperatorKubernetes` to `ZalandoPostgresOperator` across:

1. **Directory Structure**: Renamed from `postgresoperatorkubernetes` to `zalandopostgresoperator`
2. **Package Namespace**: Updated to `dev.planton.provider.kubernetes.addon.zalandopostgresoperator.v1`
3. **Proto API Definitions**: Updated all message types, field references, and validation constraints
4. **Cloud Resource Registry**: Modified the enum entry in `cloud_resource_kind.proto`
5. **Documentation**: Updated all user-facing docs and implementation guides
6. **Implementation Code**: Modified Go code in Pulumi modules to use renamed types
7. **Test Suite**: Created comprehensive validation tests with the new naming

### Naming Convention

The new naming follows Planton's established pattern while adding vendor specificity:

```yaml
# Before
apiVersion: kubernetes.planton.dev/v1
kind: PostgresOperatorKubernetes

# After
apiVersion: kubernetes.planton.dev/v1
kind: ZalandoPostgresOperator
```

The directory path changed from `provider/kubernetes/addon/postgresoperatorkubernetes/` to `provider/kubernetes/addon/zalandopostgresoperator/`, and the package namespace changed accordingly. The resource name now explicitly identifies the Zalando implementation, making it clear this deploys the [Zalando Postgres Operator](https://opensource.zalando.com/postgres-operator/).

## Implementation Details

### Proto File Changes

**File**: `apis/dev/planton/provider/kubernetes/addon/zalandopostgresoperator/v1/api.proto`

```protobuf
// Before
message PostgresOperatorKubernetes {
  string kind = 2 [(buf.validate.field).string.const = 'PostgresOperatorKubernetes'];
  PostgresOperatorKubernetesSpec spec = 4;
  PostgresOperatorKubernetesStatus status = 5;
}

// After
message ZalandoPostgresOperator {
  string kind = 2 [(buf.validate.field).string.const = 'ZalandoPostgresOperator'];
  ZalandoPostgresOperatorSpec spec = 4;
  ZalandoPostgresOperatorStatus status = 5;
}
```

**File**: `apis/dev/planton/provider/kubernetes/addon/zalandopostgresoperator/v1/spec.proto`

```protobuf
// Before
message PostgresOperatorKubernetesSpec { ... }
message PostgresOperatorKubernetesSpecContainer { ... }
message PostgresOperatorKubernetesBackupConfig { ... }
message PostgresOperatorKubernetesBackupR2Config { ... }

// After
message ZalandoPostgresOperatorSpec { ... }
message ZalandoPostgresOperatorSpecContainer { ... }
message ZalandoPostgresOperatorBackupConfig { ... }
message ZalandoPostgresOperatorBackupR2Config { ... }
```

**File**: `apis/dev/planton/provider/kubernetes/addon/zalandopostgresoperator/v1/stack_input.proto`

```protobuf
// Before
message PostgresOperatorKubernetesStackInput {
  PostgresOperatorKubernetes target = 1;
}

// After
message ZalandoPostgresOperatorStackInput {
  ZalandoPostgresOperator target = 1;
}
```

**File**: `apis/dev/planton/provider/kubernetes/addon/zalandopostgresoperator/v1/stack_outputs.proto`

```protobuf
// Before
message PostgresOperatorKubernetesStackOutputs { ... }

// After
message ZalandoPostgresOperatorStackOutputs { ... }
```

### Registry Update

**File**: `apis/dev/planton/shared/cloudresourcekind/cloud_resource_kind.proto`

```protobuf
// Before
PostgresOperatorKubernetes = 827 [(kind_meta) = {
  provider: kubernetes
  version: v1
  id_prefix: "pgopk8s"
  kubernetes_meta: {category: addon}
}];

// After
ZalandoPostgresOperator = 827 [(kind_meta) = {
  provider: kubernetes
  version: v1
  id_prefix: "pgopk8s"
  kubernetes_meta: {category: addon}
}];
```

### Implementation Code Changes

**File**: `apis/dev/planton/provider/kubernetes/addon/zalandopostgresoperator/v1/iac/pulumi/main.go`

```go
// Before
stackInput := &postgresoperatorkubernetesv1.PostgresOperatorKubernetesStackInput{}

// After
stackInput := &zalandopostgresoperatorv1.ZalandoPostgresOperatorStackInput{}
```

**File**: `apis/dev/planton/provider/kubernetes/addon/zalandopostgresoperator/v1/iac/pulumi/module/main.go`

```go
// Before
func Resources(ctx *pulumi.Context, stackInput *postgresoperatorkubernetesv1.PostgresOperatorKubernetesStackInput) error

// After
func Resources(ctx *pulumi.Context, stackInput *zalandopostgresoperatorv1.ZalandoPostgresOperatorStackInput) error
```

**File**: `apis/dev/planton/provider/kubernetes/addon/zalandopostgresoperator/v1/iac/pulumi/module/locals.go`

```go
// Before
type Locals struct {
  PostgresOperatorKubernetes *postgresoperatorkubernetesv1.PostgresOperatorKubernetes
  KubernetesLabels           map[string]string
}
kubeLabels[kuberneteslabelkeys.ResourceKind] = "PostgresOperatorKubernetes"

// After
type Locals struct {
  ZalandoPostgresOperator *zalandopostgresoperatorv1.ZalandoPostgresOperator
  KubernetesLabels        map[string]string
}
kubeLabels[kuberneteslabelkeys.ResourceKind] = "ZalandoPostgresOperator"
```

**File**: `apis/dev/planton/provider/kubernetes/addon/zalandopostgresoperator/v1/iac/pulumi/module/backup_config.go`

```go
// Before
func createBackupResources(
  ctx *pulumi.Context,
  backupConfig *postgresoperatorkubernetesv1.PostgresOperatorKubernetesBackupConfig,
  // ...
) (pulumi.StringOutput, error)

// After
func createBackupResources(
  ctx *pulumi.Context,
  backupConfig *zalandopostgresoperatorv1.ZalandoPostgresOperatorBackupConfig,
  // ...
) (pulumi.StringOutput, error)
```

### Documentation Updates

Updated all occurrences in:
- `docs/README.md` - Updated references to `ZalandoPostgresOperator` API and added introductory note about the resource

### Test Suite Creation

Created comprehensive validation tests to ensure the refactoring works correctly:

**File**: `apis/dev/planton/provider/kubernetes/addon/zalandopostgresoperator/v1/api_test.go`

```go
func TestZalandoPostgresOperator(t *testing.T) {
  gomega.RegisterFailHandler(ginkgo.Fail)
  ginkgo.RunSpecs(t, "ZalandoPostgresOperator Suite")
}

var _ = ginkgo.Describe("ZalandoPostgresOperator Custom Validation Tests", func() {
  var input *ZalandoPostgresOperator
  
  ginkgo.BeforeEach(func() {
    input = &ZalandoPostgresOperator{
      ApiVersion: "kubernetes.planton.dev/v1",
      Kind:       "ZalandoPostgresOperator",
      Metadata: &shared.CloudResourceMetadata{
        Name: "test-zalando-postgres-operator",
      },
      Spec: &ZalandoPostgresOperatorSpec{
        Container: &ZalandoPostgresOperatorSpecContainer{},
      },
    }
  })
  
  // Test cases for basic validation and backup config
})
```

### Build Process

1. **Proto Generation**: Ran `make protos` to regenerate Go stubs from updated proto files
2. **Gazelle Update**: Ran `./bazelw run //:gazelle` to update BUILD.bazel files
3. **Compilation Verification**: Successfully compiled all Go packages
4. **Linter Validation**: Confirmed no linter errors in modified proto files

## Benefits

### Cleaner API Surface

```yaml
# Users write less, understand more
kind: ZalandoPostgresOperator  # vs. kind: PostgresOperatorKubernetes
```

### Improved Code Readability

Proto message names are now more concise:
- `ZalandoPostgresOperatorSpec` (was `PostgresOperatorKubernetesSpec`)
- `ZalandoPostgresOperatorStackInput` (was `PostgresOperatorKubernetesStackInput`)
- `ZalandoPostgresOperatorStackOutputs` (was `PostgresOperatorKubernetesStackOutputs`)
- `ZalandoPostgresOperatorBackupConfig` (was `PostgresOperatorKubernetesBackupConfig`)

### Clear Vendor Identity

The name now explicitly identifies the Zalando implementation, making it clear this is not:
- Percona PostgreSQL Operator
- CloudNativePG
- Crunchy Postgres Operator
- Other PostgreSQL operator implementations

This clarity is crucial as Planton may support multiple PostgreSQL operators in the future.

### Naming Consistency

Aligns with the pattern where the provider namespace provides sufficient context:
- Package: `dev.planton.provider.kubernetes.addon.zalandopostgresoperator.v1`
- Directory: `provider/kubernetes/addon/zalandopostgresoperator/`
- Kind: `ZalandoPostgresOperator` (context is already clear from package)

### Developer Experience

- Shorter type names in code
- Less typing in YAML manifests
- Reduced cognitive load when reading code
- Vendor clarity when choosing between operators
- Consistent mental model across all Kubernetes addon operators
- Cleaner import paths reflecting the vendor name

## Impact

### User-Facing Changes

**Breaking Change**: Yes, for API consumers

Users must update their YAML manifests:

```yaml
# Update required
apiVersion: kubernetes.planton.dev/v1
kind: ZalandoPostgresOperator  # Changed from: PostgresOperatorKubernetes
metadata:
  name: zalando-postgres-operator-prod
spec:
  # ... rest unchanged
```

### Non-Breaking Aspects

- **Functionality**: Zero behavioral changes
- **ID prefix**: Unchanged (`pgopk8s`)
- **Enum value**: Unchanged (827)
- **Backup configuration**: All features remain identical
- **Helm chart integration**: No changes to how the operator is deployed

### Directory and Namespace Changes

- **Directory structure**: Changed from `postgresoperatorkubernetes` to `zalandopostgresoperator`
- **Package namespace**: Changed from `dev.planton.provider.kubernetes.addon.postgresoperatorkubernetes.v1` to `dev.planton.provider.kubernetes.addon.zalandopostgresoperator.v1`
- **Import paths**: Changed in Go code to reflect new package namespace

### Developer Impact

- Proto stubs regenerated automatically via `make protos`
- BUILD files updated via Gazelle
- No manual code migration required for internal implementation
- New validation test suite ensures correctness

## Files Modified

**Proto Definitions** (4 files):
- `api.proto` - Main API message types
- `spec.proto` - Spec, container, and backup message types
- `stack_input.proto` - Stack input message type
- `stack_outputs.proto` - Stack outputs message type

**Registry** (1 file):
- `cloud_resource_kind.proto` - Enum entry

**Documentation** (1 file):
- `docs/README.md` - Component documentation with vendor clarity

**Implementation** (5 files):
- `iac/pulumi/main.go` - Stack input type reference
- `iac/pulumi/module/main.go` - Function signature
- `iac/pulumi/module/locals.go` - Struct fields and labels
- `iac/pulumi/module/postgres_operator.go` - Field access
- `iac/pulumi/module/backup_config.go` - Function parameter type

**Test Files** (2 files):
- `api_test.go` - Comprehensive validation test suite
- `BUILD.bazel` - Test target configuration

**Generated Files**:
- `*.pb.go` files (auto-regenerated from proto definitions)
- BUILD.bazel files (auto-updated via Gazelle)

**Total**: 13 manually modified files + generated artifacts

## Related Work

This refactoring is part of an ongoing effort to improve naming consistency across Planton's Kubernetes addon operators. Similar refactoring was completed for:

- **AltinityOperator** (November 13, 2025): Renamed `AltinityOperatorKubernetes` to `AltinityOperator` to remove redundant suffix
- **ExternalDns** (November 13, 2025): Renamed `ExternalDnsKubernetes` to `ExternalDns`
- **IngressNginx** (November 13, 2025): Renamed `IngressNginxKubernetes` to `IngressNginx`

Future considerations:
- Potential support for additional PostgreSQL operators (Percona, CloudNativePG) with clear naming
- Documentation guidelines to establish naming patterns for new resources
- Migration guide for users updating from previous versions

## Migration Notes

For users with existing manifests:

1. **Find and replace** in all manifest files:
   ```bash
   find . -name "*.yaml" -type f -exec sed -i '' 's/kind: PostgresOperatorKubernetes/kind: ZalandoPostgresOperator/g' {} +
   ```

2. **No CLI changes required** - The `planton` CLI will work with the new kind name automatically after updating to the version with this change

3. **No infrastructure impact** - Existing deployed resources are unaffected; this only affects new deployments

4. **Verify manifests** - After updating, validate manifests still pass proto validation:
   ```bash
   planton pulumi preview --manifest zalando-postgres-operator.yaml --module-dir ${MODULE}
   ```

---

**Status**: ✅ Production Ready  
**Files Changed**: 13 manual + generated artifacts  
**Build Status**: All protos regenerated, no linter errors  
**Test Coverage**: Comprehensive validation tests created

