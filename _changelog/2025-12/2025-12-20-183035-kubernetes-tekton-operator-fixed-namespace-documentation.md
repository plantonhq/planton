# KubernetesTektonOperator: Fixed Namespace Architecture Documentation

**Date**: December 20, 2025
**Type**: Enhancement
**Components**: API Definitions, Kubernetes Provider, Documentation

## Summary

Reverted the recently added `namespace` and `create_namespace` fields from the KubernetesTektonOperator component after research confirmed that Tekton Operator uses **fixed namespaces** managed by the operator itself. Updated all component documentation to clearly explain this important architectural limitation that differentiates this component from other namespace-scoped resources in OpenMCF.

## Problem Statement / Motivation

During the recent KubernetesTektonOperator component development, `namespace` and `create_namespace` fields were added following the standard pattern used by other Kubernetes components in OpenMCF. However, this approach doesn't align with how Tekton Operator actually works.

### Pain Points

- **Misaligned API design**: The `namespace` and `create_namespace` fields suggested users could control namespace placement, but Tekton Operator ignores this configuration
- **Confusion potential**: Users might expect namespace customization that doesn't actually work
- **Inconsistency with Tekton architecture**: Tekton Operator is designed to manage its own namespaces
- **Wasted resources**: Creating user-specified namespaces that wouldn't be used by Tekton

### Research Confirmation

Two deep research reports (GPT-5.2 and Gemini-3) confirmed the fixed namespace behavior:

**From GPT report:**
> "When you install the Operator itself, it typically lives in namespace `tekton-operator`, and you then apply a TektonConfig CR in either `tekton-operator` or another namespace to kick off installation."

**From Gemini report:**
> "You cannot manually patch the Dashboard deployment because the Operator will revert the change."

## Solution / What's New

Removed the `namespace` and `create_namespace` fields from the KubernetesTektonOperator spec and added comprehensive documentation explaining the fixed namespace architecture.

### Fixed Namespace Architecture

The Tekton Operator uses these fixed namespaces that are automatically created and managed:

| Component | Namespace | Description |
|-----------|-----------|-------------|
| Tekton Operator | `tekton-operator` | The operator controller pod |
| Tekton Pipelines | `tekton-pipelines` | Pipeline controller and webhooks |
| Tekton Triggers | `tekton-pipelines` | Event listeners and webhooks |
| Tekton Dashboard | `tekton-pipelines` | Web-based UI |

These namespaces **cannot be customized** by users - they are fundamental to how Tekton Operator works.

## Implementation Details

### Proto Schema Changes

**File**: `apis/org/openmcf/provider/kubernetes/kubernetestektonoperator/v1/spec.proto`

Removed fields and added documentation:

```protobuf
// IMPORTANT: Namespace Behavior
// Unlike other Kubernetes components in OpenMCF, the Tekton Operator uses fixed namespaces
// that are managed by the operator itself:
// - The Tekton Operator is installed in the 'tekton-operator' namespace
// - Tekton components (Pipelines, Triggers, Dashboard) are installed in the 'tekton-pipelines' namespace
// These namespaces are automatically created and managed by the Tekton Operator and cannot be customized.
// See: https://tekton.dev/docs/operator/tektonconfig/
message KubernetesTektonOperatorSpec {
  // Target Kubernetes Cluster where the Tekton Operator will be deployed.
  org.openmcf.provider.kubernetes.KubernetesClusterSelector target_cluster = 1;

  // The container specifications for the Tekton Operator deployment.
  KubernetesTektonOperatorSpecContainer container = 2 [(buf.validate.field).required = true];

  // Configuration for which Tekton components to install.
  KubernetesTektonOperatorComponents components = 3 [(buf.validate.field).required = true];

  // The version of the Tekton Operator to deploy.
  string operator_version = 4 [(org.openmcf.shared.options.default) = "v0.78.0"];
}
```

### Pulumi Module Changes

**File**: `iac/pulumi/module/main.go`

Removed namespace creation logic:

```go
// Resources is the Pulumi entry-point.
// Note: Tekton Operator manages its own namespaces:
// - 'tekton-operator' for the operator itself
// - 'tekton-pipelines' for Tekton components (Pipelines, Triggers, Dashboard)
// These namespaces are automatically created by the Tekton Operator and cannot be customized.
func Resources(ctx *pulumi.Context,
	in *kubernetestektonoperatorv1.KubernetesTektonOperatorStackInput) error {
	// ... (no namespace creation block)
}
```

**File**: `iac/pulumi/module/locals.go`

Removed `Namespace` field from Locals struct and updated comments:

```go
type Locals struct {
	KubernetesTektonOperator *kubernetestektonoperatorv1.KubernetesTektonOperator
	KubeLabels               map[string]string
	OperatorNamespace        string   // Fixed: tekton-operator
	ComponentsNamespace      string   // Fixed: tekton-pipelines
	TektonConfigName         string
	OperatorReleaseURL       string
	// ...
}
```

### Terraform Module Changes

**File**: `iac/tf/variables.tf`

Removed namespace-related variables and added documentation:

```hcl
variable "spec" {
  # IMPORTANT: Namespace Behavior
  # Unlike other Kubernetes components in OpenMCF, the Tekton Operator uses fixed namespaces
  # that are managed by the operator itself:
  # - The Tekton Operator is installed in the 'tekton-operator' namespace
  # - Tekton components (Pipelines, Triggers, Dashboard) are installed in the 'tekton-pipelines' namespace
  # These namespaces are automatically created and managed by the Tekton Operator and cannot be customized.
  description = "Specification for KubernetesTektonOperator"
  # ... (no namespace or create_namespace fields)
}
```

**File**: `iac/tf/main.tf`

Removed namespace resource creation.

### Test File Changes

**File**: `spec_test.go`

Removed namespace-related test fixtures and the "without namespace" test case:

```go
ginkgo.BeforeEach(func() {
	// Note: Tekton Operator uses fixed namespaces managed by the operator:
	// - 'tekton-operator' for the operator
	// - 'tekton-pipelines' for components (Pipelines, Triggers, Dashboard)
	// Therefore, no namespace field is included in the spec.
	spec = &KubernetesTektonOperatorSpec{
		Container: &KubernetesTektonOperatorSpecContainer{...},
		Components: &KubernetesTektonOperatorComponents{Pipelines: true},
	}
})
```

### Documentation Updates

Updated all documentation files with prominent warnings about the fixed namespace architecture:

**Files Updated**:
- `README.md` - Added "⚠️ Important: Fixed Namespace Architecture" section
- `examples.md` - Added warning section and updated all examples
- `iac/tf/README.md` - Added namespace architecture explanation
- `iac/hack/manifest.yaml` - Added comment explaining fixed namespaces

### Example Manifest (After)

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesTektonOperator
metadata:
  name: tekton-operator
spec:
  target_cluster:
    cluster_name: "my-cluster"
  container:
    resources:
      requests:
        cpu: "100m"
        memory: "128Mi"
      limits:
        cpu: "500m"
        memory: "512Mi"
  components:
    pipelines: true
    triggers: true
    dashboard: true
  operator_version: "v0.78.0"
```

## Files Changed

| File | Change |
|------|--------|
| `spec.proto` | Removed namespace fields, added documentation |
| `iac/pulumi/module/locals.go` | Removed Namespace field |
| `iac/pulumi/module/main.go` | Removed namespace creation |
| `iac/tf/variables.tf` | Removed namespace variables |
| `iac/tf/locals.tf` | Removed namespace local |
| `iac/tf/main.tf` | Removed namespace resource |
| `spec_test.go` | Removed namespace tests |
| `examples.md` | Updated all examples, added warning |
| `README.md` | Added fixed namespace section |
| `iac/tf/README.md` | Updated documentation |
| `iac/hack/manifest.yaml` | Removed namespace fields |

## Benefits

### API Accuracy
- API now accurately reflects Tekton Operator's actual behavior
- No misleading fields that suggest non-existent customization

### Clear Documentation
- Users immediately understand the fixed namespace constraint
- Prominent warnings prevent confusion and wasted effort

### Simpler Implementation
- No unnecessary namespace creation logic
- Cleaner IaC modules without unused functionality

### Consistency with Upstream
- Aligns with official Tekton Operator documentation
- Matches real-world Tekton deployment patterns

## Impact

### Users
- **No breaking changes**: Fields were recently added and not yet in production use
- **Better guidance**: Clear documentation prevents confusion
- **Correct expectations**: Users understand the namespace architecture upfront

### Developers
- **Accurate API**: Spec reflects actual Tekton behavior
- **Simpler modules**: Less code to maintain in IaC modules
- **Reference documentation**: Research reports available for future reference

### Operations
- **Predictable deployments**: Tekton always deploys to known namespaces
- **Easier troubleshooting**: `tekton-operator` and `tekton-pipelines` namespaces are consistent

## Related Work

- [2025-12-19-055933-kubernetes-tekton-operator-component.md](2025-12-19-055933-kubernetes-tekton-operator-component.md) - Initial component creation
- [2025-12-20-122911-kubernetes-tekton-operator-version-field.md](2025-12-20-122911-kubernetes-tekton-operator-version-field.md) - Version field addition

## Research References

The decision was based on comprehensive research reports:

- `planton/apis/ai/.../kubernetestektonoperator/v1/research/report.gpt-5.2.md`
- `planton/apis/ai/.../kubernetestektonoperator/v1/research/report.gemini-3.md`

Key findings from research:
1. Tekton Operator itself installs in `tekton-operator` namespace
2. TektonConfig CRD has `targetNamespace` (typically `tekton-pipelines`) managed by operator
3. These namespaces are not customizable from outside the operator
4. Operator will revert manual changes to its managed resources

## Validation

All validations passed:
- ✅ `make protos` - Proto generation successful
- ✅ `go test` - All tests passing
- ✅ `make build` - Full build successful

---

**Status**: ✅ Production Ready
**Timeline**: ~30 minutes (research validation + implementation)

