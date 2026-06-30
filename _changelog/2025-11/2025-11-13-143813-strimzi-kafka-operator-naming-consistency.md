# StrimziKafkaOperator Naming Consistency Refactoring

**Date**: November 13, 2025  
**Type**: Refactoring  
**Components**: API Definitions, Cloud Resource Registry, Documentation, Pulumi Implementation

## Summary

Renamed the `KafkaOperatorKubernetes` resource to `StrimziKafkaOperator` across all proto definitions, documentation, and implementation code to align with Planton's naming conventions for Kubernetes addon operators. This change eliminates redundant "Kubernetes" suffixes, explicitly identifies the Strimzi operator, and improves consistency with other addon resources that are already scoped to Kubernetes via their provider namespace.

## Problem Statement / Motivation

The Kafka operator resource was originally named `KafkaOperatorKubernetes`, which included a redundant "Kubernetes" suffix and lacked specificity about which Kafka operator was being deployed. This naming pattern was inconsistent with Planton's design philosophy where:

### Pain Points

- **Redundant Context**: The resource lives under `provider/kubernetes/addon/`, making the "Kubernetes" suffix in the name redundant
- **Lack of Specificity**: The name "Kafka Operator" is generic—it doesn't indicate we're specifically deploying **Strimzi**, the CNCF-backed, production-ready Kafka operator
- **Verbose API Surface**: Users had to write `kind: KafkaOperatorKubernetes` in manifests, which is unnecessarily long and doesn't communicate the actual operator being deployed
- **Naming Inconsistency**: Mixed naming patterns across addon operators—some with suffixes, some without, and none clearly identifying the specific operator implementation
- **Code Verbosity**: Proto message types like `KafkaOperatorKubernetesSpec` and `KafkaOperatorKubernetesStackInput` were excessively long
- **Poor Developer Experience**: The redundancy and lack of clarity made code harder to read and type

The provider namespace (`dev.planton.provider.kubernetes.addon.strimzikafkaoperator.v1`) already clearly indicates this is a Kubernetes component, so including "Kubernetes" in every message name adds noise without value. More importantly, the generic "Kafka Operator" name doesn't convey that this is **Strimzi specifically**—a critical detail for users choosing between Strimzi, Confluent for Kubernetes, or Banzai Cloud Koperator.

## Solution / What's New

Performed a comprehensive rename from `KafkaOperatorKubernetes` to `StrimziKafkaOperator` across:

1. **Proto API Definitions**: Updated all message types, field references, and validation constraints
2. **Cloud Resource Registry**: Modified the enum entry in `cloud_resource_kind.proto`
3. **Documentation**: Updated all user-facing docs to reflect the Strimzi-specific naming
4. **Implementation Code**: Modified Go code in Pulumi modules to use renamed types

### Naming Convention

The new naming follows Planton's established pattern while adding specificity:

```yaml
# Before
apiVersion: kubernetes.planton.dev/v1
kind: KafkaOperatorKubernetes

# After
apiVersion: kubernetes.planton.dev/v1
kind: StrimziKafkaOperator
```

The provider path remains unchanged (`provider/kubernetes/addon/strimzikafkaoperator/`) to maintain consistency with the directory structure that was already renamed. The new name clearly identifies:
- **Strimzi**: The specific operator implementation (not Confluent, not Koperator)
- **Kafka**: The technology being operated
- **Operator**: The resource type (operator deployment, not a Kafka cluster)

## Implementation Details

### Proto File Changes

**File**: `apis/dev/planton/provider/kubernetes/addon/strimzikafkaoperator/v1/api.proto`

```protobuf
// Before
message KafkaOperatorKubernetes {
  string kind = 2 [(buf.validate.field).string.const = 'KafkaOperatorKubernetes'];
  KafkaOperatorKubernetesSpec spec = 4;
  KafkaOperatorKubernetesStatus status = 5;
}

// After
message StrimziKafkaOperator {
  string kind = 2 [(buf.validate.field).string.const = 'StrimziKafkaOperator'];
  StrimziKafkaOperatorSpec spec = 4;
  StrimziKafkaOperatorStatus status = 5;
}
```

**File**: `apis/dev/planton/provider/kubernetes/addon/strimzikafkaoperator/v1/spec.proto`

```protobuf
// Before
message KafkaOperatorKubernetesSpec { ... }
message KafkaOperatorKubernetesSpecContainer { ... }

// After
message StrimziKafkaOperatorSpec { ... }
message StrimziKafkaOperatorSpecContainer { ... }
```

Updated documentation in spec messages to clarify this is the **Strimzi Kafka Operator** specifically.

**File**: `apis/dev/planton/provider/kubernetes/addon/strimzikafkaoperator/v1/stack_input.proto`

```protobuf
// Before
message KafkaOperatorKubernetesStackInput {
  KafkaOperatorKubernetes target = 1;
}

// After
message StrimziKafkaOperatorStackInput {
  StrimziKafkaOperator target = 1;
}
```

**File**: `apis/dev/planton/provider/kubernetes/addon/strimzikafkaoperator/v1/stack_outputs.proto`

```protobuf
// Before
message KafkaOperatorKubernetesStackOutputs { ... }

// After
message StrimziKafkaOperatorStackOutputs { ... }
```

### Registry Update

**File**: `apis/dev/planton/shared/cloudresourcekind/cloud_resource_kind.proto`

```protobuf
// Before
KafkaOperatorKubernetes = 826 [(kind_meta) = {
  provider: kubernetes
  version: v1
  id_prefix: "kfkopk8s"
  kubernetes_meta: {category: addon}
}];

// After
StrimziKafkaOperator = 826 [(kind_meta) = {
  provider: kubernetes
  version: v1
  id_prefix: "kfkopk8s"
  kubernetes_meta: {category: addon}
}];
```

### Implementation Code Changes

**File**: `apis/dev/planton/provider/kubernetes/addon/strimzikafkaoperator/v1/iac/pulumi/main.go`

```go
// Before
stackInput := &strimzikafkaoperatorv1.KafkaOperatorKubernetesStackInput{}

// After
stackInput := &strimzikafkaoperatorv1.StrimziKafkaOperatorStackInput{}
```

**File**: `apis/dev/planton/provider/kubernetes/addon/strimzikafkaoperator/v1/iac/pulumi/module/main.go`

```go
// Before
func Resources(ctx *pulumi.Context, stackInput *strimzikafkaoperatorv1.KafkaOperatorKubernetesStackInput) error

// After
func Resources(ctx *pulumi.Context, stackInput *strimzikafkaoperatorv1.StrimziKafkaOperatorStackInput) error
```

**File**: `apis/dev/planton/provider/kubernetes/addon/strimzikafkaoperator/v1/iac/pulumi/module/kafka_operator.go`

```go
// Before
func kafkaOperator(ctx *pulumi.Context, target *strimzikafkaoperatorv1.KafkaOperatorKubernetes, ...)

// After
func kafkaOperator(ctx *pulumi.Context, target *strimzikafkaoperatorv1.StrimziKafkaOperator, ...)
```

### Documentation Updates

Updated all occurrences in:
- `docs/README.md` (main component documentation)
  - Updated all references from `KafkaOperatorKubernetes` to `StrimziKafkaOperator`
  - Documentation already contained extensive Strimzi-specific content, now the API name matches the documented implementation

### Build Process

1. **Proto Generation**: Ran `make build` in apis directory to regenerate Go stubs from updated proto files
2. **Gazelle Update**: Ran `./bazelw run //:gazelle` to update BUILD.bazel files
3. **Compilation Verification**: Successfully compiled all Go packages
4. **Linter Validation**: Confirmed no linter errors in modified proto files

## Benefits

### Clearer Identity

```yaml
# Users immediately know which operator is being deployed
kind: StrimziKafkaOperator  # vs. kind: KafkaOperatorKubernetes
```

### Improved Code Readability

Proto message names are now more concise and specific:
- `StrimziKafkaOperatorSpec` (was `KafkaOperatorKubernetesSpec`)
- `StrimziKafkaOperatorStackInput` (was `KafkaOperatorKubernetesStackInput`)
- `StrimziKafkaOperatorStackOutputs` (was `KafkaOperatorKubernetesStackOutputs`)

### Naming Consistency

Aligns with the pattern where the provider namespace provides sufficient context:
- Package: `dev.planton.provider.kubernetes.addon.strimzikafkaoperator.v1`
- Kind: `StrimziKafkaOperator` (context is already clear from package, name identifies the specific operator)

### Developer Experience

- Shorter, clearer type names in code
- Explicitly identifies Strimzi as the operator implementation
- Reduced cognitive load when reading code
- Consistent mental model across all Kubernetes addon operators
- No ambiguity about which Kafka operator is being deployed

### Production Clarity

Users evaluating Kafka operators (Strimzi vs. Confluent vs. Koperator) can immediately see from the manifest that this deploys **Strimzi specifically**—the open-source, CNCF-backed, production-ready operator.

## Impact

### User-Facing Changes

**Breaking Change**: Yes, for API consumers

Users must update their YAML manifests:

```yaml
# Update required
apiVersion: kubernetes.planton.dev/v1
kind: StrimziKafkaOperator  # Changed from: KafkaOperatorKubernetes
metadata:
  name: kafka-operator-prod
spec:
  # ... rest unchanged
```

### Non-Breaking Aspects

- **Directory structure**: Already renamed to `strimzikafkaoperator`
- **Package namespace**: Unchanged (`dev.planton.provider.kubernetes.addon.strimzikafkaoperator.v1`)
- **Import paths**: Unchanged in Go code
- **Functionality**: Zero behavioral changes—still deploys Strimzi via Helm
- **ID prefix**: Unchanged (`kfkopk8s`)
- **Enum value**: Unchanged (826)

### Developer Impact

- Proto stubs regenerated automatically
- BUILD files updated via Gazelle
- No manual code migration required for internal implementation
- All existing functionality continues to work

## Files Modified

**Proto Definitions** (4 files):
- `api.proto` - Main API message types and comments
- `spec.proto` - Spec and container message types with updated descriptions
- `stack_input.proto` - Stack input message type
- `stack_outputs.proto` - Stack outputs message type with updated comments

**Registry** (1 file):
- `cloud_resource_kind.proto` - Enum entry

**Documentation** (1 file):
- `docs/README.md` - Component documentation (5 occurrences updated)

**Implementation** (3 files):
- `iac/pulumi/main.go` - Stack input type reference
- `iac/pulumi/module/main.go` - Function signature
- `iac/pulumi/module/kafka_operator.go` - Function signature

**Generated Files**:
- `*.pb.go` files (auto-regenerated from proto definitions)
- `BUILD.bazel` files (auto-updated via Gazelle)

**Total**: 9 manually modified files + generated artifacts

## Related Work

This refactoring is part of an ongoing effort to improve naming consistency across Planton's Kubernetes addon operators. Recent similar work includes:

- `AltinityOperatorKubernetes` → `AltinityOperator` (2025-11-13)
- `CertManagerKubernetes` → `CertManager` (previous)
- `ExternalDnsKubernetes` → `ExternalDns` (previous)
- `ExternalSecretsKubernetes` → `ExternalSecrets` (previous)

Future candidates for similar operator-specific naming:
- `PostgresOperatorKubernetes` → `ZalandoPostgresOperator` (likely rename)
- Other operators that should identify their specific implementation

## Migration Notes

For users with existing manifests:

1. **Find and replace** in all manifest files:
   ```bash
   find . -name "*.yaml" -type f -exec sed -i '' 's/kind: KafkaOperatorKubernetes/kind: StrimziKafkaOperator/g' {} +
   ```

2. **No CLI changes required** - The `planton` CLI will work with the new kind name automatically after updating to the version with this change

3. **No infrastructure impact** - Existing deployed operators are unaffected; this only affects new deployments

4. **Documentation benefit** - The new name makes it immediately clear that this deploys the Strimzi operator specifically, aligning with the extensive Strimzi-focused documentation already present

---

**Status**: ✅ Complete  
**Files Changed**: 9 manual + generated artifacts  
**Build Status**: All builds passing, no linter errors  
**Operator**: Strimzi Kafka Operator (unchanged functionality, clearer naming)

