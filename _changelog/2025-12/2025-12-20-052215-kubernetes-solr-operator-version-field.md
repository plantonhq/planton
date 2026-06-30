# Add operator_version Field to KubernetesSolrOperator

**Date**: December 20, 2025
**Type**: Enhancement
**Components**: API Definitions, Kubernetes Provider, Terraform Module, Pulumi Module

## Summary

Added `operator_version` field to the KubernetesSolrOperator component spec, allowing users to specify which version of the Apache Solr Operator to deploy. The default value is set to `v0.9.1` (the latest release) via proto field options, with the version flowing through to both Terraform and Pulumi IaC modules.

## Problem Statement / Motivation

The KubernetesSolrOperator component had hardcoded operator versions in both the Terraform and Pulumi modules, making it difficult for users to:

### Pain Points

- Deploy specific versions of the Solr Operator for compatibility or testing
- Upgrade to newer operator releases without modifying IaC code
- Pin to stable versions in production while testing newer versions elsewhere
- Track which operator version is deployed as part of the manifest specification

## Solution / What's New

Added a new `operator_version` field to `KubernetesSolrOperatorSpec` with:

1. **Proto schema update** - New field with default value via `options.default` annotation
2. **Terraform support** - Variable with default, used for Helm chart version and CRD URL
3. **Pulumi support** - Reads version from spec, computes chart version and CRD manifest URL
4. **Input-driven configuration** - Removed hardcoded defaults from IaC modules; values always come from input

### Version Flow

```
spec.proto (default: v0.9.1)
         │
         ▼
┌─────────────────────────────────────┐
│  User Manifest (optional override)  │
│  operator_version: "v0.8.0"         │
└─────────────────────────────────────┘
         │
         ▼
┌─────────────────┬─────────────────┐
│   Terraform     │     Pulumi      │
│   locals.tf     │   locals.go     │
│                 │                 │
│  Strip 'v' →    │  Strip 'v' →    │
│  helm: "0.9.1"  │  helm: "0.9.1"  │
│                 │                 │
│  CRD URL with   │  CRD URL with   │
│  full version   │  full version   │
└─────────────────┴─────────────────┘
```

## Implementation Details

### Proto Schema

**File**: `apis/dev/planton/provider/kubernetes/kubernetessolroperator/v1/spec.proto`

```protobuf
import "dev/planton/shared/options/options.proto";

message KubernetesSolrOperatorSpec {
  // ... existing fields ...

  // The version of the Apache Solr Operator to deploy.
  // https://github.com/apache/solr-operator/releases
  string operator_version = 4 [(dev.planton.shared.options.default) = "v0.9.1"];

  // Container field moved to field number 5
  KubernetesSolrOperatorSpecContainer container = 5 [(buf.validate.field).required = true];
}
```

### Terraform Module

**File**: `iac/tf/variables.tf`
```hcl
variable "spec" {
  type = object({
    # ... other fields ...
    operator_version = optional(string, "v0.9.1")
    # ...
  })
}
```

**File**: `iac/tf/locals.tf`
```hcl
locals {
  # Namespace comes from input
  namespace = var.spec.namespace

  # Helm chart version (strip 'v' prefix)
  helm_chart_version = trimprefix(var.spec.operator_version, "v")

  # CRD manifest URL (uses full version with 'v' prefix)
  crd_manifest_url = "https://solr.apache.org/operator/downloads/crds/${var.spec.operator_version}/all-with-dependencies.yaml"
}
```

### Pulumi Module

**File**: `iac/pulumi/module/vars.go`
```go
var vars = struct {
    HelmChartName        string
    HelmChartRepo        string
    CrdManifestURLFormat string
}{
    HelmChartName:        "solr-operator",
    HelmChartRepo:        "https://solr.apache.org/charts",
    CrdManifestURLFormat: "https://solr.apache.org/operator/downloads/crds/v%s/all-with-dependencies.yaml",
}
```

**File**: `iac/pulumi/module/locals.go`
```go
// Helm chart version without 'v' prefix
locals.ChartVersion = strings.TrimPrefix(target.Spec.OperatorVersion, "v")

// CRD manifest URL
locals.CrdManifestURL = fmt.Sprintf(vars.CrdManifestURLFormat, locals.ChartVersion)
```

### Key Design Decisions

1. **Version format**: Stored with `v` prefix (e.g., `v0.9.1`) to match GitHub releases naming
2. **Helm chart version**: Strips `v` prefix since Helm charts use `0.9.1` format
3. **No hardcoded fallbacks**: IaC modules rely entirely on input; defaults are set at the proto/variable level
4. **Single source of truth**: Default version defined once in spec.proto, mirrored in Terraform variable

## Benefits

- **Version flexibility**: Users can deploy any Solr Operator version
- **Explicit configuration**: Version is visible in the manifest specification
- **Upgrade path**: Easy to test new versions before rolling out
- **Consistency**: Same version mechanism can be applied to other operator components
- **Clean IaC**: No hardcoded versions scattered across Terraform/Pulumi code

## Usage Examples

### Default Version (v0.9.1)

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesSolrOperator
metadata:
  name: solr-operator
spec:
  targetCluster:
    clusterName: "my-cluster"
  namespace:
    value: "solr-operator-system"
  create_namespace: true
  container: {}
```

### Specific Version

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesSolrOperator
metadata:
  name: solr-operator
spec:
  targetCluster:
    clusterName: "my-cluster"
  namespace:
    value: "solr-operator-system"
  create_namespace: true
  operator_version: "v0.8.0"  # Pin to specific version
  container: {}
```

## Files Changed

| File | Change |
|------|--------|
| `spec.proto` | Added `operator_version` field with default annotation |
| `iac/tf/variables.tf` | Added `operator_version` variable |
| `iac/tf/locals.tf` | Compute helm version and CRD URL from input |
| `iac/pulumi/module/vars.go` | Removed hardcoded defaults |
| `iac/pulumi/module/locals.go` | Added version and CRD URL computation |
| `iac/pulumi/module/main.go` | Use computed CRD URL from locals |
| `examples.md` | Updated examples to show operator_version field |

## Impact

- **Users**: Can now specify operator version in manifests
- **Operations**: Easier version management and upgrades
- **Consistency**: Pattern can be replicated to other Kubernetes operator components

## Related Work

- Apache Solr Operator releases: https://github.com/apache/solr-operator/releases
- `dev/planton/shared/options/options.proto` - Provides the `default` field option

---

**Status**: ✅ Production Ready
**Validation**: `make protos`, `go test`, `make build` all passing
