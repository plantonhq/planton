# KubernetesTektonOperator Deployment Component

**Date**: December 19, 2025
**Type**: Feature
**Components**: Kubernetes Provider, API Definitions, Pulumi CLI Integration, Provider Framework, IAC Stack Runner

## Summary

Created a complete new deployment component `KubernetesTektonOperator` for deploying the Tekton CI/CD framework operator on Kubernetes clusters. The component follows the Planton forge workflow, implementing all proto API definitions, validation rules, unit tests, Pulumi and Terraform IaC modules, and comprehensive documentation.

## Problem Statement / Motivation

Organizations adopting Kubernetes-native CI/CD need a standardized, declarative way to deploy and manage Tekton components. Tekton is a powerful Kubernetes-native CI/CD framework, but its installation involves multiple components (Pipelines, Triggers, Dashboard) that need to be managed together.

### Pain Points

- Manual Tekton installation requires applying multiple YAML manifests with version coordination
- No declarative way to specify which Tekton components to enable
- Component lifecycle management (upgrades, configuration changes) is manual
- Inconsistent deployment patterns across different environments
- No integration with Planton's unified infrastructure management

## Solution / What's New

Created the `KubernetesTektonOperator` deployment component using the Planton forge workflow. The component deploys the Tekton Operator, which then manages Tekton components via the TektonConfig CRD.

### Component Architecture

```
KubernetesTektonOperator (Planton)
         │
         ▼
    Tekton Operator Deployment
         │
         ▼
    TektonConfig CRD
         │
         ▼
    Tekton Components (managed by operator)
        ├── tekton-pipelines-controller
        ├── tekton-pipelines-webhook
        ├── tekton-triggers-controller (optional)
        ├── tekton-triggers-webhook (optional)
        └── tekton-dashboard (optional)
```

### Key Features

1. **Component Selection**: Users can enable/disable individual Tekton components:
   - Pipelines (core CI/CD execution)
   - Triggers (event-driven automation)
   - Dashboard (web UI)

2. **Profile-Based Installation**: Maps component selection to Tekton profiles:
   - `all`: Pipelines + Triggers + Dashboard
   - `basic`: Pipelines + Triggers
   - `lite`: Pipelines only

3. **Validation Rules**: CEL validation ensures at least one component is enabled

4. **Container Resources**: Configurable operator pod resource allocation

## Implementation Details

### Proto API Definitions

Created four proto files following KRM conventions:

**spec.proto** - Configuration schema:
```protobuf
message KubernetesTektonOperatorSpec {
  KubernetesClusterSelector target_cluster = 1;
  KubernetesTektonOperatorSpecContainer container = 2;
  KubernetesTektonOperatorComponents components = 3;
}

message KubernetesTektonOperatorComponents {
  bool pipelines = 1;
  bool triggers = 2;
  bool dashboard = 3;

  option (buf.validate.message).cel = {
    id: "components.at_least_one"
    expression: "this.pipelines || this.triggers || this.dashboard"
    message: "at least one Tekton component must be enabled"
  };
}
```

**api.proto** - KRM wiring:
```protobuf
message KubernetesTektonOperator {
  string api_version = 1 [(buf.validate.field).string.const = 'kubernetes.planton.dev/v1'];
  string kind = 2 [(buf.validate.field).string.const = 'KubernetesTektonOperator'];
  CloudResourceMetadata metadata = 3;
  KubernetesTektonOperatorSpec spec = 4;
  KubernetesTektonOperatorStatus status = 5;
}
```

**stack_outputs.proto** - Deployment outputs:
- Namespace
- TektonConfig name
- Service names for enabled components
- Dashboard port-forward command

**stack_input.proto** - IaC module inputs

### Registry Entry

Added to `cloud_resource_kind.proto`:
```protobuf
KubernetesTektonOperator = 838 [(kind_meta) = {
  provider: kubernetes
  version: v1
  id_prefix: "k8stktnop"
}];
```

### Pulumi Module

The Pulumi implementation deploys Tekton Operator using YAML manifests:

```go
// Install Tekton Operator from release manifests
operatorManifests, err := yaml.NewConfigFile(ctx, "tekton-operator", &yaml.ConfigFileArgs{
    File: vars.OperatorReleaseURL,
}, pulumi.Provider(k8sProvider))

// Create TektonConfig to configure components
tektonConfigYAML := buildTektonConfigYAML(locals, profile)
_, err = yaml.NewConfigGroup(ctx, "tekton-config", &yaml.ConfigGroupArgs{
    YAML: []string{tektonConfigYAML},
}, pulumi.Provider(k8sProvider), pulumi.DependsOn([]pulumi.Resource{operatorManifests}))
```

### Terraform Module

The Terraform implementation uses the `kubectl` provider for manifest application:

```hcl
resource "kubectl_manifest" "tekton_operator" {
  for_each = {
    for idx, doc in split("---", data.http.tekton_operator_manifest.response_body) : idx => doc
    if trimspace(doc) != "" && !startswith(trimspace(doc), "#")
  }
  yaml_body = each.value
  wait = true
}

resource "kubectl_manifest" "tekton_config" {
  yaml_body = <<-YAML
    apiVersion: operator.tekton.dev/v1alpha1
    kind: TektonConfig
    metadata:
      name: ${local.tekton_config_name}
    spec:
      profile: ${local.tekton_profile}
      targetNamespace: ${local.components_namespace}
  YAML
  depends_on = [kubectl_manifest.tekton_operator]
}
```

### Validation Tests

Created `spec_test.go` with Ginkgo/Gomega tests covering:
- Valid configurations with different component combinations
- Required field validations (container, components)
- CEL validation for at least one component enabled

```go
ginkgo.Context("with no components enabled", func() {
    ginkgo.It("should return a validation error", func() {
        spec.Components = &KubernetesTektonOperatorComponents{
            Pipelines: false,
            Triggers:  false,
            Dashboard: false,
        }
        err := protovalidate.Validate(spec)
        gomega.Expect(err).NotTo(gomega.BeNil())
    })
})
```

## Files Created

```
apis/dev/planton/provider/kubernetes/kubernetestektonoperator/v1/
├── Proto Files
│   ├── spec.proto              # Configuration schema
│   ├── api.proto               # KRM wiring
│   ├── stack_input.proto       # IaC inputs
│   ├── stack_outputs.proto     # Deployment outputs
│   └── spec_test.go            # Validation tests
│
├── Documentation
│   ├── README.md               # User guide (~200 lines)
│   ├── examples.md             # Usage examples (~400 lines)
│   └── docs/README.md          # Research docs (~450 lines)
│
├── Pulumi IaC
│   ├── main.go                 # Entrypoint
│   ├── Pulumi.yaml             # Project config
│   ├── Makefile                # Build automation
│   ├── debug.sh                # Debug helper
│   ├── README.md               # Pulumi guide
│   ├── overview.md             # Architecture docs
│   └── module/
│       ├── main.go             # Module orchestration
│       ├── locals.go           # Local variables
│       ├── outputs.go          # Output constants
│       ├── vars.go             # Configuration constants
│       └── tekton_operator.go  # Implementation
│
├── Terraform IaC
│   ├── main.tf                 # Resource definitions
│   ├── variables.tf            # Input variables
│   ├── locals.tf               # Local computations
│   ├── outputs.tf              # Output definitions
│   ├── provider.tf             # Provider config
│   └── README.md               # Terraform guide
│
└── Supporting Files
    └── iac/hack/manifest.yaml  # Test manifest
```

## Usage Example

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesTektonOperator
metadata:
  name: tekton-operator
spec:
  target_cluster:
    kubernetes_credential_id: "my-cluster"
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
```

Deploy using:
```bash
planton apply -f tekton-operator.yaml
```

## Benefits

### For Platform Engineers

- **Declarative Management**: Define Tekton deployments as code
- **Consistent Deployments**: Same configuration across environments
- **Component Flexibility**: Enable only needed components
- **Unified Tooling**: Manage Tekton alongside other Planton resources

### For Developers

- **Quick Setup**: Deploy complete CI/CD infrastructure with one manifest
- **Dashboard Access**: Easy port-forward command in outputs
- **Documentation**: Comprehensive examples for different scenarios

### For Operations

- **Resource Control**: Configurable operator resource allocation
- **Visibility**: Stack outputs provide service endpoints
- **Dual IaC Support**: Choose Pulumi or Terraform based on team preferences

## Impact

### Users Affected

- Platform engineers deploying CI/CD infrastructure
- DevOps teams standardizing on Tekton
- Organizations using Planton for infrastructure management

### Ecosystem Integration

- Extends Kubernetes provider with CI/CD capability
- Follows established patterns from other operator components (Elastic, Strimzi, Zalando)
- Compatible with Planton credential management

## Validation Results

| Check | Status |
|-------|--------|
| Proto build (`buf lint/generate`) | ✅ Passed |
| Go stubs generated | ✅ Passed |
| Unit tests (`go test`) | ✅ Passed |
| Build validation (`make build`) | ✅ Passed |
| Registry entry added | ✅ Enum 838 |

## Related Work

- **KubernetesElasticOperator**: Reference implementation for operator pattern
- **KubernetesStrimziKafkaOperator**: Similar operator deployment pattern
- **Forge Workflow**: Used `_rules/deployment-component/forge/` rules

## Design Decisions

### Why Tekton Operator vs Direct Installation?

**Decision**: Deploy via Tekton Operator rather than applying component manifests directly.

**Rationale**:
- Operator handles version compatibility between components
- TektonConfig CRD provides unified configuration
- Automatic reconciliation and self-healing
- Simpler upgrade path through operator

### Why Profile-Based Component Selection?

**Decision**: Map boolean component flags to Tekton profiles (all/basic/lite).

**Rationale**:
- Aligns with Tekton Operator's native design
- Reduces complexity in IaC modules
- Ensures correct component dependencies
- Single configuration point via TektonConfig

### Why YAML Manifests vs Helm?

**Decision**: Use official YAML release manifests instead of Helm charts.

**Rationale**:
- No official Helm chart from Tekton project
- Release manifests are the primary distribution method
- Simpler dependency management
- Direct alignment with Tekton release process

---

**Status**: ✅ Production Ready
**Timeline**: Single session implementation following forge workflow
