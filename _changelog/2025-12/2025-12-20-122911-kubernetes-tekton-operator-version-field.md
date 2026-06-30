# Add operator_version Field to KubernetesTektonOperator

**Date**: December 20, 2025
**Type**: Enhancement
**Components**: API Definitions, Kubernetes Provider, Pulumi Module, Terraform Module

## Summary

Added `operator_version` field to the KubernetesTektonOperator component spec, allowing users to specify which version of the Tekton Operator to deploy. The default value is set to `v0.78.0` (the latest release from OperatorHub) via proto field options. Also fixed a Server-Side Apply field conflict that prevented successful deployment.

## Problem Statement / Motivation

The KubernetesTektonOperator component had hardcoded operator versions in both the Terraform and Pulumi modules, making it difficult for users to:

### Pain Points

- Deploy specific versions of the Tekton Operator for compatibility or testing
- Upgrade to newer operator releases without modifying IaC code
- Pin to stable versions in production while testing newer versions elsewhere
- Track which operator version is deployed as part of the manifest specification

Additionally, the TektonConfig resource was setting `pipeline.enable-api-fields: stable`, which the Tekton Operator automatically manages, causing Server-Side Apply field conflicts during deployment.

## Solution / What's New

Added a new `operator_version` field to `KubernetesTektonOperatorSpec` with:

1. **Proto schema update** - New field with default value via `options.default` annotation
2. **Terraform support** - Release URL computed from version input
3. **Pulumi support** - Release URL computed from spec version
4. **TektonConfig fix** - Removed operator-managed fields to prevent conflicts

### Version Flow

```
spec.proto (default: v0.78.0)
         │
         ▼
┌─────────────────────────────────────┐
│  User Manifest (optional override)  │
│  operator_version: "v0.75.0"        │
└─────────────────────────────────────┘
         │
         ▼
┌─────────────────┬─────────────────┐
│   Terraform     │     Pulumi      │
│   locals.tf     │   locals.go     │
│                 │                 │
│  Release URL    │  Release URL    │
│  computed from  │  computed from  │
│  version input  │  version input  │
└─────────────────┴─────────────────┘
```

## Implementation Details

### Proto Schema

**File**: `apis/dev/planton/provider/kubernetes/kubernetestektonoperator/v1/spec.proto`

```protobuf
import "dev/planton/shared/options/options.proto";

message KubernetesTektonOperatorSpec {
  // ... existing fields ...

  // The version of the Tekton Operator to deploy.
  // https://github.com/tektoncd/operator/releases
  // https://operatorhub.io/operator/tektoncd-operator
  string operator_version = 4 [(dev.planton.shared.options.default) = "v0.78.0"];
}
```

### Pulumi Module

**File**: `iac/pulumi/module/vars.go`
```go
var vars = struct {
    OperatorNamespace        string
    ComponentsNamespace      string
    OperatorReleaseURLFormat string  // Format string for versioned releases
    TektonConfigName         string
}{
    OperatorReleaseURLFormat: "https://storage.googleapis.com/tekton-releases/operator/previous/%s/release.yaml",
    // ...
}
```

**File**: `iac/pulumi/module/locals.go`
```go
// Compute operator release URL from version (default comes from proto options)
l.OperatorReleaseURL = fmt.Sprintf(vars.OperatorReleaseURLFormat, in.Target.Spec.OperatorVersion)
```

### Terraform Module

**File**: `iac/tf/variables.tf`
```hcl
variable "spec" {
  type = object({
    # ...
    # Default value (v0.78.0) is set in spec.proto via options.default
    operator_version = optional(string)
  })
}
```

**File**: `iac/tf/locals.tf`
```hcl
locals {
  # Operator release URL (uses version from spec)
  operator_release_url = "https://storage.googleapis.com/tekton-releases/operator/previous/${var.spec.operator_version}/release.yaml"
}
```

### TektonConfig Fix

Removed the `pipeline.enable-api-fields: stable` field from TektonConfig in both Pulumi and Terraform modules to prevent Server-Side Apply conflicts with the operator:

```yaml
# Before (caused conflict)
spec:
  profile: all
  targetNamespace: tekton-pipelines
  pipeline:
    enable-api-fields: stable  # ❌ Managed by operator

# After (works correctly)
spec:
  profile: all
  targetNamespace: tekton-pipelines
  # ✅ Let operator manage pipeline settings
```

### Key Design Decisions

1. **Default only in proto**: No hardcoded fallbacks in IaC modules; defaults set via `options.default` annotation
2. **Version format**: Stored with `v` prefix (e.g., `v0.78.0`) to match GitHub releases naming
3. **Release URL pattern**: Uses `previous/v0.78.0/release.yaml` path for versioned releases
4. **Minimal TektonConfig**: Only set fields users need to control (profile, targetNamespace)

## Files Changed

| File | Change |
|------|--------|
| `spec.proto` | Added `operator_version` field with default annotation |
| `iac/pulumi/module/vars.go` | Changed to URL format template |
| `iac/pulumi/module/locals.go` | Added release URL computation |
| `iac/pulumi/module/tekton_operator.go` | Use computed URL, fix TektonConfig |
| `iac/tf/variables.tf` | Added `operator_version` variable |
| `iac/tf/locals.tf` | Compute release URL from version |
| `iac/tf/main.tf` | Fixed TektonConfig (removed conflicting field) |
| `examples.md` | Added version usage examples |
| `v1/README.md` | Added operator_version to component structure |
| `iac/pulumi/README.md` | Updated vars.go example, fixed TektonConfig docs |
| `iac/pulumi/overview.md` | Updated upgrade path documentation |
| `iac/tf/README.md` | Added operator_version to variables table |
| `iac/hack/manifest.yaml` | Added commented operator_version field |

## Usage Examples

### Default Version (v0.78.0)

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesTektonOperator
metadata:
  name: tekton-operator
spec:
  targetCluster:
    clusterName: "my-cluster"
  container: {}
  components:
    pipelines: true
    triggers: true
    dashboard: true
  # operator_version defaults to v0.78.0
```

### Specific Version

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesTektonOperator
metadata:
  name: tekton-operator
spec:
  targetCluster:
    clusterName: "my-cluster"
  container: {}
  components:
    pipelines: true
    triggers: true
    dashboard: false
  operator_version: "v0.75.0"  # Pin to specific version
```

## Benefits

- **Version flexibility**: Users can deploy any Tekton Operator version
- **Explicit configuration**: Version is visible in the manifest specification
- **Upgrade path**: Easy to test new versions before rolling out
- **Consistency**: Same version pattern applied to other operator components
- **Clean IaC**: No hardcoded versions in Terraform/Pulumi code
- **Successful deployments**: Fixed Server-Side Apply conflict that blocked deployments

## Impact

### Users Affected

- Platform engineers deploying Tekton CI/CD infrastructure
- Organizations requiring version pinning for compliance
- Teams testing new operator releases

### Validation

| Check | Status |
|-------|--------|
| Proto build (`make protos`) | ✅ Passed |
| Go stubs generated | ✅ Passed |
| Unit tests (`go test`) | ✅ Passed |
| Build validation (`make build`) | ✅ Passed |
| Deployment test | ✅ Fixed (TektonConfig conflict resolved) |

## Related Work

- [KubernetesSolrOperator version field](2025-12-20-052215-kubernetes-solr-operator-version-field.md) - Similar pattern for Solr Operator
- [KubernetesTektonOperator component](2025-12-19-055933-kubernetes-tekton-operator-component.md) - Original component creation
- `dev/planton/shared/options/options.proto` - Provides the `default` field option

---

**Status**: ✅ Production Ready
**Validation**: `make protos`, `go test`, `make build` all passing, deployment conflict resolved
