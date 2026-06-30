# KubernetesManifest Deployment Component

**Date**: December 21, 2025  
**Type**: Feature  
**Components**: API Definitions, Kubernetes Provider, Pulumi CLI Integration, Provider Framework

## Summary

Implemented a new `KubernetesManifest` deployment component that enables deploying arbitrary Kubernetes manifests (single or multi-document YAML) to any Kubernetes cluster. This component serves as a flexible "escape hatch" for deploying custom resources, operator CRDs, RBAC configurations, and any Kubernetes resources that don't fit into more specialized deployment components.

## Problem Statement / Motivation

While Planton provides specialized deployment components (KubernetesDeployment, KubernetesStatefulSet, KubernetesDaemonSet, KubernetesHelmRelease), there's a gap for scenarios requiring raw manifest deployment:

### Pain Points

- **Operator Custom Resources**: Operators require deploying CRs alongside CRDs—these don't warrant dedicated components
- **Vendor-Provided Manifests**: Third-party vendors provide raw YAML that shouldn't need Helm chart wrapping
- **Infrastructure Resources**: NetworkPolicies, ResourceQuotas, LimitRanges don't fit application-focused components
- **Migration Paths**: Teams moving from kubectl workflows need a bridge to IaC
- **Rapid Prototyping**: Engineers need to deploy test resources quickly during development
- **CRD Timing Issues**: Using `kubectl apply` or older Pulumi yaml module causes race conditions when CRDs and CRs are in the same manifest

## Solution / What's New

Created a complete deployment component following the forge pattern with:

1. **Proto API Definitions** - All 4 proto files with validations
2. **Pulumi Module** - Using yamlv2 for smart CRD ordering
3. **Documentation** - User-facing, research, and technical docs
4. **Validation Tests** - 14 test specs covering all validation rules
5. **Supporting Files** - Test manifests, debug scripts

### API Design

```protobuf
message KubernetesManifestSpec {
  // Target Kubernetes Cluster (optional)
  KubernetesClusterSelector target_cluster = 1;

  // Kubernetes Namespace (required)
  StringValueOrRef namespace = 2;

  // Flag to create namespace
  bool create_namespace = 3;

  // Raw Kubernetes manifest YAML (required)
  // Supports single or multi-document (--- separated)
  string manifest_yaml = 4;
}
```

### Key Implementation: yamlv2 for CRD Ordering

The critical design decision was using Pulumi's `yaml/v2` module instead of the older `yaml.ConfigFile`:

```go
// module/main.go
_, err := yamlv2.NewConfigGroup(ctx, "manifest", &yamlv2.ConfigGroupArgs{
    Yaml: pulumi.StringPtr(locals.ManifestYAML),
}, opts...)
```

This provides automatic CRD ordering—when a manifest contains both a CRD and Custom Resources:

```yaml
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: myresources.example.com
...
---
apiVersion: example.com/v1
kind: MyResource
metadata:
  name: my-instance
```

The yamlv2 module ensures the CRD is fully registered before creating the Custom Resource, preventing the common "no matches for kind" timing error.

## Implementation Details

### Files Created

```
apis/dev/planton/provider/kubernetes/kubernetesmanifest/v1/
├── Proto Files
│   ├── api.proto              # KRM wiring (metadata/spec/status)
│   ├── spec.proto             # Spec with manifest_yaml field
│   ├── stack_input.proto      # IaC module inputs
│   └── stack_outputs.proto    # Deployment outputs (namespace)
│
├── Generated Stubs
│   ├── api.pb.go
│   ├── spec.pb.go
│   ├── stack_input.pb.go
│   └── stack_outputs.pb.go
│
├── Tests
│   └── spec_test.go           # 14 validation test specs
│
├── User Documentation
│   ├── README.md              # Overview and use cases
│   ├── examples.md            # 7 usage examples
│   └── docs/README.md         # Comprehensive research (~400 lines)
│
└── IaC Implementation
    └── iac/
        ├── hack/manifest.yaml # Test manifest
        └── pulumi/
            ├── main.go        # Entrypoint
            ├── Makefile       # Build helpers
            ├── Pulumi.yaml    # Project config
            ├── README.md      # Module documentation
            ├── overview.md    # Architecture overview
            ├── debug.sh       # Debug helper
            └── module/
                ├── main.go    # Resources using yamlv2
                ├── locals.go  # Input transformation
                └── outputs.go # Output constants
```

### Cloud Resource Kind Registration

Added to `cloud_resource_kind.proto`:

```protobuf
KubernetesManifest = 842 [(kind_meta) = {
  provider: kubernetes
  version: v1
  id_prefix: "k8smfst"
}];
```

### Pulumi Module Architecture

```
StackInput (spec + credentials)
    ↓
initializeLocals()
    ├── Extract namespace from spec
    ├── Build resource labels
    └── Store manifest YAML
    ↓
Resources()
    ├── Create Kubernetes Provider
    ├── (Optional) Create Namespace with DependsOn
    └── Apply Manifest via yamlv2.ConfigGroup
    ↓
Outputs (namespace)
```

### Validation Tests

14 test specs covering:
- Valid input with all fields
- Multi-document manifest support
- Optional target_cluster handling
- Missing namespace validation
- Empty manifest_yaml validation
- Missing metadata validation
- Missing spec validation
- Namespace creation flag (true/false)
- API version validation
- Kind validation

## Benefits

### For Platform Engineers

- **Zero Transformation**: Manifest YAML is applied exactly as written
- **State Management**: Full Pulumi state tracking and drift detection
- **CRD Safety**: Automatic ordering prevents timing issues
- **Unified Experience**: Same patterns as all other Planton components

### For Developers

- **Escape Hatch**: Deploy anything that doesn't fit specialized components
- **Migration Path**: Bridge from kubectl workflows to IaC
- **Rapid Prototyping**: Test resources without creating new components

### For Operations

- **Audit Trail**: Full IaC history for compliance
- **Rollback Support**: Built-in through Pulumi
- **Multi-Cluster**: Consistent deployment across GKE, EKS, AKS, etc.

## Usage Examples

### Basic ConfigMap

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesManifest
metadata:
  name: my-config
spec:
  namespace: my-namespace
  create_namespace: true
  manifest_yaml: |
    apiVersion: v1
    kind: ConfigMap
    metadata:
      name: app-config
    data:
      environment: production
```

### Multi-Resource Application

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesManifest
metadata:
  name: complete-app
spec:
  namespace: production
  create_namespace: true
  manifest_yaml: |
    apiVersion: v1
    kind: ConfigMap
    metadata:
      name: app-config
    data:
      environment: production
    ---
    apiVersion: apps/v1
    kind: Deployment
    metadata:
      name: my-app
    spec:
      replicas: 3
      ...
    ---
    apiVersion: v1
    kind: Service
    metadata:
      name: my-app
    spec:
      selector:
        app: my-app
      ports:
      - port: 80
```

### Applying with CLI

```bash
planton apply -f manifest.yaml
planton pulumi up --stack-input manifest.yaml
```

## Impact

### Component Ecosystem

- Fills the gap for "raw" deployments in Planton
- Complements specialized components (Deployment, StatefulSet, DaemonSet, Helm)
- Provides consistent patterns with other Kubernetes components

### Use Cases Enabled

| Use Case | Previously | Now |
|----------|------------|-----|
| Operator CRs | Custom scripts or Helm | Single KubernetesManifest |
| RBAC configs | Manual kubectl | Tracked in IaC |
| Vendor manifests | Copy-paste apply | Version-controlled |
| Infrastructure resources | Scattered YAMLs | Unified management |

## Design Decisions

### Why manifest_yaml as a String Field?

Chose a single string field over structured YAML for:
- **Preservation**: Exact formatting and comments preserved
- **Flexibility**: Supports any Kubernetes resource type without schema updates
- **Simplicity**: Easy copy-paste from existing manifests
- **No Maintenance**: No need to update schemas for new resource types

### Why Required Namespace?

Namespace is required even though manifests can specify their own:
- Provides default for resources without explicit namespaces
- Enables namespace creation when needed
- Follows pattern of other Kubernetes components

### Why Optional target_cluster?

For consistency with other components:
- Defaults to current cluster context
- Enables multi-cluster deployments when specified
- Supports all Kubernetes cluster types

## Related Work

- **KubernetesDeployment**: For microservice deployments (more structured)
- **KubernetesHelmRelease**: For Helm chart deployments (templating support)
- **KubernetesStatefulSet**: For stateful workloads
- **KubernetesDaemonSet**: For node-level agents

KubernetesManifest is the "raw" option when these specialized components don't fit.

## Metrics

| Metric | Value |
|--------|-------|
| Proto files | 4 |
| Generated Go stubs | 4 |
| Pulumi module files | 3 |
| Documentation files | 6 |
| Test specifications | 14 |
| Examples provided | 7 |
| Research doc lines | ~400 |

---

**Status**: ✅ Production Ready  
**Build**: All Bazel targets pass  
**Tests**: 14/14 specs passing

