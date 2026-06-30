# KubernetesTektonOperator

## Overview

**KubernetesTektonOperator** is a Planton component that deploys the Tekton Operator on Kubernetes clusters to manage the lifecycle of Tekton components. Tekton is a Kubernetes-native CI/CD framework that provides cloud-native building blocks for creating continuous integration and delivery (CI/CD) pipelines.

The Tekton Operator simplifies the installation and management of Tekton components including:

- **Tekton Pipelines**: Core component for building and running CI/CD pipelines
- **Tekton Triggers**: Event-driven execution of pipelines from webhooks and other sources
- **Tekton Dashboard**: Web-based UI for viewing and managing pipelines

## Key Features

### Component-Based Installation

The operator allows selective installation of Tekton components:

- **Pipelines**: Core pipeline execution engine (typically always enabled)
- **Triggers**: Event listeners and webhook handlers for automated pipeline execution
- **Dashboard**: Visual interface for monitoring and managing pipelines

### Kubernetes-Native CI/CD

Tekton provides truly cloud-native CI/CD:

- **CRD-Based**: Pipelines defined as Kubernetes Custom Resources
- **Portable**: Run the same pipelines on any Kubernetes cluster
- **Scalable**: Kubernetes-native scaling and resource management
- **Declarative**: Infrastructure-as-code approach to CI/CD

### Operator Pattern Benefits

By using the Kubernetes Operator pattern:

- **Automated Upgrades**: Seamless upgrades of Tekton components
- **Self-Healing**: Automatic recovery from component failures
- **Consistent Deployments**: Same deployment process across environments
- **Configuration Management**: Centralized configuration through TektonConfig CRD

### Resource Management

The operator deployment can be configured with custom resource allocations:

- **CPU and Memory Limits**: Define maximum resources the operator can consume
- **Resource Requests**: Guaranteed baseline resources for operator pod
- **Default Configuration**: Pre-configured with production-ready defaults

## ⚠️ Important: Fixed Namespace Architecture

**Unlike other Kubernetes components in Planton, the Tekton Operator uses fixed namespaces** that are managed by the operator itself:

| Component | Namespace | Description |
|-----------|-----------|-------------|
| Tekton Operator | `tekton-operator` | The operator controller pod |
| Tekton Pipelines | `tekton-pipelines` | Pipeline controller and webhooks |
| Tekton Triggers | `tekton-pipelines` | Event listeners and webhooks |
| Tekton Dashboard | `tekton-pipelines` | Web-based UI |

**These namespaces are automatically created and managed by the Tekton Operator and cannot be customized.**

This is a fundamental design decision of the Tekton Operator project. For more information, see:
- https://tekton.dev/docs/operator/tektonconfig/

## Component Structure

### API Definition

The KubernetesTektonOperator API follows Planton's resource structure:

```
kubernetes-tekton-operator
├── api_version: "kubernetes.planton.dev/v1"
├── kind: "KubernetesTektonOperator"
├── metadata: CloudResourceMetadata
└── spec: KubernetesTektonOperatorSpec
    ├── target_cluster: KubernetesClusterSelector
    ├── container: KubernetesTektonOperatorSpecContainer
    │   └── resources: ContainerResources
    │       ├── limits: {cpu, memory}
    │       └── requests: {cpu, memory}
    ├── components: KubernetesTektonOperatorComponents
    │   ├── pipelines: bool
    │   ├── triggers: bool
    │   └── dashboard: bool
    └── operator_version: string (default: "v0.78.0")
```

### Deployment Model

**Installation Method**: Tekton Operator release manifests  
**Operator Pod**: Single pod deployment  
**CRDs Installed**: TektonConfig, TektonPipeline, TektonTrigger, TektonDashboard

## Configuration

### Components Configuration

At least one component must be enabled:

```yaml
spec:
  components:
    pipelines: true   # Core pipeline execution
    triggers: true    # Event-driven automation
    dashboard: true   # Web UI for management
```

### Resource Specification

The `spec.container.resources` field controls the operator pod's resource allocation:

```yaml
spec:
  container:
    resources:
      requests:
        cpu: "100m"      # Minimum guaranteed CPU
        memory: "128Mi"  # Minimum guaranteed memory
      limits:
        cpu: "500m"      # Maximum CPU
        memory: "512Mi"  # Maximum memory
```

### Default Resources

If not specified, the operator uses production-ready defaults:

- **Requests**: 100m CPU, 128Mi memory
- **Limits**: 500m CPU, 512Mi memory

## Usage Patterns

### Basic Installation with All Components

```yaml
apiVersion: kubernetes.planton.dev/v1
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
```

### Minimal Installation (Pipelines Only)

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesTektonOperator
metadata:
  name: tekton-operator
spec:
  target_cluster:
    cluster_name: "my-cluster"
  container: {}
  components:
    pipelines: true
```

## Post-Installation

After deploying the Tekton Operator, you can create CI/CD resources:

### Example: Simple Pipeline

```yaml
apiVersion: tekton.dev/v1beta1
kind: Pipeline
metadata:
  name: build-and-deploy
spec:
  tasks:
    - name: build
      taskRef:
        name: build-task
    - name: deploy
      taskRef:
        name: deploy-task
      runAfter:
        - build
```

### Example: Event Listener (Triggers)

```yaml
apiVersion: triggers.tekton.dev/v1beta1
kind: EventListener
metadata:
  name: github-listener
spec:
  triggers:
    - name: github-push
      bindings:
        - ref: github-push-binding
      template:
        ref: pipeline-template
```

## Benefits

### Operational Efficiency

- **Simplified Installation**: One-click deployment of Tekton components
- **Version Management**: Operator handles component version compatibility
- **Consistent Configuration**: TektonConfig CRD manages all settings

### Reliability

- **Self-Healing**: Automatic recovery from component failures
- **Rolling Updates**: Zero-downtime upgrades of Tekton components
- **Health Monitoring**: Operator continuously monitors component health

### Scalability

- **Event-Driven**: Triggers enable event-based pipeline execution
- **Kubernetes-Native**: Leverages Kubernetes scheduling and scaling
- **Multi-Pipeline**: Support for concurrent pipeline executions

## Documentation

For detailed information, see:

- **Research Documentation**: [docs/README.md](docs/README.md) - Comprehensive guide on Tekton deployment patterns
- **Examples**: [examples.md](examples.md) - Practical usage examples
- **Pulumi Module**: [iac/pulumi/README.md](iac/pulumi/README.md) - Pulumi-specific details
- **Terraform Module**: [iac/tf/README.md](iac/tf/README.md) - Terraform-specific details

### Operator Version

You can specify the Tekton Operator version to deploy:

```yaml
spec:
  operator_version: "v0.78.0"  # Default version
```

Available versions: https://github.com/tektoncd/operator/releases

## Version Information

- **Tekton Operator**: v0.78.0 (default, configurable via `operator_version`)
- **Namespace**: tekton-operator (operator), tekton-pipelines (components)
- **Supported Kubernetes**: 1.28+

## Support

For issues, questions, or contributions:

- **GitHub**: [plantonhq/planton](https://github.com/plantonhq/planton)
- **Tekton Documentation**: [Tekton Docs](https://tekton.dev/docs/)
- **Tekton Operator**: [Tekton Operator GitHub](https://github.com/tektoncd/operator)
