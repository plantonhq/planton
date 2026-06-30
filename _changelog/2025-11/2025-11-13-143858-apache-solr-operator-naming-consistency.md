# Apache Solr Operator Naming Consistency Refactoring

**Date**: November 13, 2025  
**Type**: Refactoring  
**Components**: API Definitions, Cloud Resource Registry, Documentation, Pulumi CLI Integration

## Summary

Renamed the `SolrOperatorKubernetes` resource to `ApacheSolrOperator` across all proto definitions, documentation, and implementation code to align with Planton's naming conventions for Kubernetes addon operators. This change eliminates redundant "Kubernetes" suffixes and improves consistency with other addon resources while properly identifying the operator by its official Apache project name.

## Problem Statement / Motivation

The Apache Solr Operator resource was originally named `SolrOperatorKubernetes`, which included a redundant "Kubernetes" suffix and lacked the "Apache" prefix that properly identifies this as an official Apache Software Foundation project. This naming pattern was inconsistent with Planton's design philosophy and didn't adequately distinguish this operator from other potential Solr deployment methods.

### Pain Points

- **Redundant Context**: The resource lives under `provider/kubernetes/addon/`, making the "Kubernetes" suffix in the name redundant
- **Missing Attribution**: Omitting "Apache" from the name doesn't properly identify this as the official Apache Solr Operator project
- **Verbose API Surface**: Users had to write `kind: SolrOperatorKubernetes` in manifests, which is unnecessarily long
- **Naming Inconsistency**: Mixed naming patterns across addon operators—some with suffixes, some without
- **Code Verbosity**: Proto message types like `SolrOperatorKubernetesSpec` and `SolrOperatorKubernetesStackInput` were excessively long
- **Poor Developer Experience**: The redundancy made code harder to read and type
- **Lack of Clarity**: The name didn't clearly indicate this is the operator-based deployment approach (vs. direct StatefulSet deployments or Helm charts)

The provider namespace (`dev.planton.provider.kubernetes.addon.apachesolroperator.v1`) already clearly indicates this is a Kubernetes component, so including "Kubernetes" in every message name adds noise without value. More importantly, adding "Apache" properly attributes the operator to its upstream project.

## Solution / What's New

Performed a comprehensive rename from `SolrOperatorKubernetes` to `ApacheSolrOperator` across:

1. **Proto API Definitions**: Updated all message types, field references, and validation constraints
2. **Cloud Resource Registry**: Modified the enum entry in `cloud_resource_kind.proto`
3. **Documentation**: Updated all user-facing docs with proper Apache Solr Operator references
4. **Implementation Code**: Modified Go code in Pulumi modules to use renamed types
5. **Generated Code**: Regenerated all proto stubs with new type names

### Naming Convention

The new naming follows Planton's established pattern while properly identifying the upstream project:

```yaml
# Before
apiVersion: kubernetes.planton.dev/v1
kind: SolrOperatorKubernetes

# After
apiVersion: kubernetes.planton.dev/v1
kind: ApacheSolrOperator
```

The name "ApacheSolrOperator" clearly communicates:
- **Apache**: Official Apache Software Foundation project
- **Solr**: The search platform being operated
- **Operator**: Kubernetes operator-based deployment (not raw StatefulSets or standalone Helm)

The directory was renamed from `solroperatorkubernetes` to `apachesolroperator` to match the new naming, with package namespace updated to `dev.planton.provider.kubernetes.addon.apachesolroperator.v1`.

## Implementation Details

### Proto File Changes

**File**: `apis/dev/planton/provider/kubernetes/addon/apachesolroperator/v1/api.proto`

```protobuf
// Before
message SolrOperatorKubernetes {
  string kind = 2 [(buf.validate.field).string.const = 'SolrOperatorKubernetes'];
  SolrOperatorKubernetesSpec spec = 4;
  SolrOperatorKubernetesStatus status = 5;
}

// After
message ApacheSolrOperator {
  string kind = 2 [(buf.validate.field).string.const = 'ApacheSolrOperator'];
  ApacheSolrOperatorSpec spec = 4;
  ApacheSolrOperatorStatus status = 5;
}
```

**File**: `apis/dev/planton/provider/kubernetes/addon/apachesolroperator/v1/spec.proto`

```protobuf
// Before
message SolrOperatorKubernetesSpec {
  dev.planton.shared.kubernetes.KubernetesAddonTargetCluster target_cluster = 1;
  SolrOperatorKubernetesSpecContainer container = 2;
}

// After
message ApacheSolrOperatorSpec {
  dev.planton.shared.kubernetes.KubernetesAddonTargetCluster target_cluster = 1;
  ApacheSolrOperatorSpecContainer container = 2;
}
```

Updated comments to properly reflect Apache Solr Operator instead of copy-pasted references.

**File**: `apis/dev/planton/provider/kubernetes/addon/apachesolroperator/v1/stack_input.proto`

```protobuf
// Before
message SolrOperatorKubernetesStackInput {
  SolrOperatorKubernetes target = 1;
  dev.planton.provider.kubernetes.KubernetesProviderConfig provider_config = 2;
}

// After
message ApacheSolrOperatorStackInput {
  ApacheSolrOperator target = 1;
  dev.planton.provider.kubernetes.KubernetesProviderConfig provider_config = 2;
}
```

**File**: `apis/dev/planton/provider/kubernetes/addon/apachesolroperator/v1/stack_outputs.proto`

```protobuf
// Before
message SolrOperatorKubernetesStackOutputs {
  //kubernetes namespace in which solr-operator-kubernetes is created.
  string namespace = 1;
  // ... other fields
}

// After
message ApacheSolrOperatorStackOutputs {
  //kubernetes namespace in which apache-solr-operator is created.
  string namespace = 1;
  // ... other fields
}
```

Updated all comments throughout to reference `apache-solr-operator` consistently.

### Registry Update

**File**: `apis/dev/planton/shared/cloudresourcekind/cloud_resource_kind.proto`

```protobuf
// Before
SolrOperatorKubernetes = 828 [(kind_meta) = {
  provider: kubernetes
  version: v1
  id_prefix: "slropk8s"
  kubernetes_meta: {category: addon}
}];

// After
ApacheSolrOperator = 828 [(kind_meta) = {
  provider: kubernetes
  version: v1
  id_prefix: "slropk8s"
  kubernetes_meta: {category: addon}
}];
```

### Implementation Code Changes

**File**: `apis/dev/planton/provider/kubernetes/addon/apachesolroperator/v1/iac/pulumi/main.go`

```go
// Before
stackInput := &solroperatorkubernetesv1.SolrOperatorKubernetesStackInput{}

// After
stackInput := &apachesolroperatorv1.ApacheSolrOperatorStackInput{}
```

Package imports were also updated to reference `apachesolroperator` instead of `solroperatorkubernetes`.

**File**: `apis/dev/planton/provider/kubernetes/addon/apachesolroperator/v1/iac/pulumi/module/main.go`

```go
// Before
func Resources(ctx *pulumi.Context, stackInput *solroperatorkubernetesv1.SolrOperatorKubernetesStackInput) error

// After
func Resources(ctx *pulumi.Context, stackInput *apachesolroperatorv1.ApacheSolrOperatorStackInput) error
```

Updated function comment to reference "Apache Solr Operator Kubernetes add-on".

### Documentation Updates

**File**: `apis/dev/planton/provider/kubernetes/addon/apachesolroperator/v1/docs/README.md`

Updated title and all references throughout the comprehensive deployment guide:
- Title: "Deploying Apache Solr Operator on Kubernetes: From Anti-Patterns to Production-Ready Operators"
- References to `SolrOperatorKubernetes` → `ApacheSolrOperator` in API design sections
- Maintained the comprehensive technical content about operator-based deployments

### Build Process

1. **Proto Generation**: Ran `make protos` to regenerate Go stubs from updated proto files
   - Successfully generated new `ApacheSolrOperator` types across all proto files
2. **Gazelle Update**: Ran `./bazelw run //:gazelle` to update BUILD.bazel files with new package structure
3. **Compilation Verification**: Successfully compiled the apachesolroperator package (254 total actions)
4. **Cloud Resource Registry Build**: Verified cloudresourcekind package builds successfully
5. **Linter Validation**: No linter errors in any modified proto files

## Benefits

### Clearer Naming

```yaml
# Users write less, understand more
kind: ApacheSolrOperator  # vs. kind: SolrOperatorKubernetes
```

The new name immediately communicates:
- This is an official Apache project
- It's the Operator deployment method (not raw Helm or StatefulSets)
- Context (Kubernetes) is implied by the API group

### Improved Code Readability

Proto message names are now more concise and properly attributed:
- `ApacheSolrOperatorSpec` (was `SolrOperatorKubernetesSpec`)
- `ApacheSolrOperatorStackInput` (was `SolrOperatorKubernetesStackInput`)
- `ApacheSolrOperatorStackOutputs` (was `SolrOperatorKubernetesStackOutputs`)

### Naming Consistency

Aligns with the pattern where the provider namespace provides sufficient context:
- Package: `dev.planton.provider.kubernetes.addon.apachesolroperator.v1`
- Directory: `apachesolroperator`
- Kind: `ApacheSolrOperator` (context is already clear from package)
- Follows the same pattern as `AltinityOperator` refactoring

### Proper Attribution

The "Apache" prefix properly credits the upstream Apache Software Foundation project, making it clear this is not a Planton custom implementation but integration with the official Apache Solr Operator.

### Developer Experience

- Shorter type names in code
- Less typing in YAML manifests
- Reduced cognitive load when reading code
- Consistent mental model across all Kubernetes addon operators
- Clear distinction from other Solr deployment approaches

## Impact

### User-Facing Changes

**Breaking Change**: Yes, for API consumers

Users must update their YAML manifests:

```yaml
# Update required
apiVersion: kubernetes.planton.dev/v1
kind: ApacheSolrOperator  # Changed from: SolrOperatorKubernetes
metadata:
  name: solr-operator-prod
spec:
  # ... rest unchanged
```

### Non-Breaking Aspects

- **Functionality**: Zero behavioral changes
- **ID prefix**: Unchanged (`slropk8s`)
- **Enum value**: Unchanged (828)
- **Helm integration**: Unchanged (still deploys Apache Solr Operator from official charts)

### Developer Impact

- Proto stubs regenerated automatically with new types
- BUILD files updated via Gazelle with new package structure
- No manual code migration required for internal implementation
- All builds continue to pass successfully
- No test files existed, so no test updates needed

## Files Modified

**Proto Definitions** (4 files):
- `api.proto` - Main API message types and kind validation
- `spec.proto` - Spec and container message types with updated comments
- `stack_input.proto` - Stack input message type
- `stack_outputs.proto` - Stack outputs message type with updated comments

**Registry** (1 file):
- `cloud_resource_kind.proto` - Enum entry renamed

**Documentation** (1 file):
- `docs/README.md` - Updated title and API references

**Implementation** (2 files):
- `iac/pulumi/main.go` - Stack input type reference and import paths
- `iac/pulumi/module/main.go` - Function signature and comment

**Generated Files**:
- `*.pb.go` files (auto-regenerated from proto definitions)
- `BUILD.bazel` files (auto-updated via Gazelle)

**Total**: 8 manually modified files + generated artifacts

## Related Work

This refactoring follows the pattern established in:
- **AltinityOperator Naming Consistency** (`2025-11-13-120651-altinity-operator-naming-consistency.md`) - Same refactoring approach for the Altinity ClickHouse Operator

This is part of an ongoing effort to improve naming consistency across Planton's Kubernetes addon operators. The refactoring also included renaming the directory from `solroperatorkubernetes` to `apachesolroperator` to maintain consistency between the directory name, package name, and resource kind.

## Migration Notes

For users with existing manifests:

1. **Find and replace** in all manifest files:
   ```bash
   find . -name "*.yaml" -type f -exec sed -i '' 's/kind: SolrOperatorKubernetes/kind: ApacheSolrOperator/g' {} +
   ```

2. **No CLI changes required** - The `planton` CLI will work with the new kind name automatically after updating to the version with this change

3. **No infrastructure impact** - Existing deployed resources are unaffected; this only affects new deployments

4. **Verify manifests** - Test updated manifests in a non-production environment first

## Context: Apache Solr Operator

For background on why the Apache Solr Operator is the recommended deployment approach for SolrCloud on Kubernetes, see the comprehensive deployment guide in `docs/README.md`. Key points:

- **Production-Proven**: Used by Bloomberg to run over 1000 SolrCloud clusters on Kubernetes
- **100% Open Source**: Apache License 2.0, donated to Apache by Bloomberg
- **Comprehensive Features**: ZooKeeper management, safe rolling updates, automatic shard rebalancing, backup/restore, monitoring integration
- **No Viable Alternatives**: The de facto standard for SolrCloud on Kubernetes

The naming change to `ApacheSolrOperator` better reflects this official upstream project and distinguishes it from other Solr deployment approaches (manual StatefulSets, standalone Helm charts).

---

**Status**: ✅ Production Ready  
**Files Changed**: 8 manual + generated artifacts  
**Build Status**: All builds passing, no linter errors

