# KubernetesGhaRunnerScaleSetController Deployment Component

**Date**: January 7, 2026
**Type**: Feature
**Components**: API Definitions, Kubernetes Provider, Pulumi Module, Terraform Module, Documentation

## Summary

Added a new deployment component `KubernetesGhaRunnerScaleSetController` that enables declarative deployment of the GitHub Actions Runner Scale Set Controller on Kubernetes clusters. This component deploys the official GitHub controller that manages AutoScalingRunnerSets and EphemeralRunners, enabling dynamic scaling of self-hosted GitHub Actions runners based on workflow demand.

## Problem Statement / Motivation

Self-hosted GitHub Actions runners on Kubernetes require the Actions Runner Controller (ARC) to be installed and configured. The controller manages the lifecycle of runner pods, scaling them up when workflow jobs are queued and scaling down to zero when idle.

### Pain Points

- **Manual Helm Installation**: Users had to manually install the controller via Helm
- **No Validation**: No schema validation for controller configuration
- **Inconsistent Interface**: Deploying the controller didn't follow the same declarative pattern as other Planton components
- **Scattered Documentation**: Configuration options were spread across Helm chart documentation

## Solution / What's New

Created a complete deployment component following the Planton standard structure:

### Component Structure

```
kubernetesgharunnerscalesetcontroller/v1/
├── api.proto              # KRM envelope (apiVersion/kind/metadata/spec/status)
├── spec.proto             # Configuration schema with validations
├── stack_input.proto      # IaC module inputs
├── stack_outputs.proto    # Deployment outputs
├── spec_test.go           # 12 validation tests
├── README.md              # User-facing documentation
├── examples.md            # Practical deployment examples
├── docs/
│   └── README.md          # Comprehensive research documentation
└── iac/
    ├── hack/
    │   └── manifest.yaml  # Test manifest
    ├── pulumi/
    │   ├── main.go        # Entry point
    │   ├── module/        # Pulumi module implementation
    │   └── ...
    └── tf/
        ├── main.tf        # Terraform implementation
        └── ...
```

## Implementation Details

### Proto Schema (`spec.proto`)

Designed a comprehensive schema covering the 80/20 of controller configuration:

- **Namespace Configuration**: `namespace` with create flag
- **Helm Chart Version**: Defaults to `0.13.1`
- **Replica Count**: For HA deployments with leader election
- **Container Resources**: CPU/memory requests and limits
- **Controller Flags**: Log level, log format, watch single namespace, concurrent reconciles, update strategy
- **Metrics Configuration**: Prometheus-compatible metrics endpoints
- **Image Pull Secrets**: For private registries
- **Priority Class**: For critical workloads

### Enums for User-Friendly Configuration

```protobuf
enum LogLevel {
  log_level_unspecified = 0;
  debug = 1;
  info = 2;
  warn = 3;
  error = 4;
}

enum UpdateStrategy {
  update_strategy_unspecified = 0;
  immediate = 1;  // Apply changes immediately
  eventual = 2;   // Wait for jobs to complete
}
```

### Pulumi Module

Implemented Helm-based deployment:

- `module/main.go`: Resource orchestration
- `module/locals.go`: Value transformations from proto to Helm values
- `module/controller.go`: Helm release deployment with namespace creation

### Terraform Module

Feature-parity implementation:

- `variables.tf`: Mirrors spec.proto fields
- `locals.tf`: Helm values construction
- `main.tf`: Namespace and Helm release resources
- `outputs.tf`: Maps to stack_outputs.proto

### Registry Entry

```protobuf
KubernetesGhaRunnerScaleSetController = 843 [(kind_meta) = {
  provider: kubernetes
  version: v1
  id_prefix: "k8sgharsc"
}];
```

## Benefits

### For Users

- **Declarative Configuration**: Define controller in YAML, deploy with CLI
- **Validation Before Deploy**: Proto validations catch errors early
- **Consistent Interface**: Same manifest structure as other K8s components
- **Documentation**: Research docs explain landscape and design decisions

### For Platform Teams

- **Multi-IaC Support**: Choose Pulumi or Terraform
- **Auditable**: All configuration in version control
- **Extensible**: Clear patterns for adding features

## Impact

| Area | Impact |
|------|--------|
| Users | Can deploy GHA controller declaratively with validation |
| Developers | Reference implementation for controller-style components |
| Component Catalog | +1 Kubernetes addon component (843rd cloud resource kind) |

## Usage Example

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesGhaRunnerScaleSetController
metadata:
  name: arc-controller
spec:
  namespace:
    value: arc-system
  createNamespace: true
  helmChartVersion: "0.13.1"
  container:
    resources:
      requests:
        cpu: 100m
        memory: 128Mi
      limits:
        cpu: 500m
        memory: 512Mi
  flags:
    logLevel: info
    updateStrategy: eventual
```

Deploy:
```bash
planton pulumi up --manifest arc-controller.yaml --stack org/project/env
```

## Testing

- ✅ 12 validation tests covering required fields and optional configurations
- ✅ `go vet` passes
- ✅ `buf build` validates proto files
- ✅ Terraform validates successfully

## Related Work

- Similar to `KubernetesTektonOperator` - Operator/controller deployment pattern
- Uses same `StringValueOrRef` for namespace as other K8s components
- Follows established Helm-based deployment pattern

---

**Status**: ✅ Production Ready
**Timeline**: Single session implementation

