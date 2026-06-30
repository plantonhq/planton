# KubernetesGhaRunnerScaleSetController

## Overview

**KubernetesGhaRunnerScaleSetController** is a Planton component that deploys the GitHub Actions Runner Scale Set Controller on Kubernetes clusters. The controller enables dynamic scaling of self-hosted GitHub Actions runners based on workflow demand using the Kubernetes-native autoscaling pattern.

The controller manages the following Custom Resources:

- **AutoScalingRunnerSet**: Defines a set of runners that scale automatically
- **AutoScalingListener**: Listens to GitHub for workflow job events
- **EphemeralRunner**: Individual runner pods that process single jobs
- **EphemeralRunnerSet**: Manages the lifecycle of EphemeralRunner pods

## Key Features

### Dynamic Runner Scaling

The controller enables true autoscaling of GitHub Actions runners:

- **Scale to Zero**: No runners running when there are no jobs
- **Scale on Demand**: New runners spin up when workflow jobs are queued
- **Ephemeral Runners**: Each runner processes a single job and terminates
- **Cost Efficient**: Pay only for resources when actively running workflows

### Kubernetes-Native Architecture

Built specifically for Kubernetes environments:

- **CRD-Based**: Runners defined as Kubernetes Custom Resources
- **Pod-Based Runners**: Runners are Kubernetes pods with full K8s capabilities
- **Service Account Integration**: Runners can use K8s RBAC and secrets
- **Node Affinity**: Control which nodes run your workflow jobs

### High Availability

Designed for production workloads:

- **Leader Election**: Multiple controller replicas for HA
- **Graceful Upgrades**: Update strategies to avoid job interruption
- **Metrics Support**: Prometheus-compatible metrics for monitoring

### Resource Management

Fine-grained control over controller resources:

- **CPU and Memory Limits**: Define controller resource allocation
- **Rate Limiting**: Configure K8s API client rate limits
- **Concurrent Reconciles**: Tune throughput vs API load

## Component Structure

### API Definition

The KubernetesGhaRunnerScaleSetController API follows Planton's resource structure:

```
kubernetes-gha-runner-scale-set-controller
├── api_version: "kubernetes.planton.dev/v1"
├── kind: "KubernetesGhaRunnerScaleSetController"
├── metadata: CloudResourceMetadata
└── spec: KubernetesGhaRunnerScaleSetControllerSpec
    ├── target_cluster: KubernetesClusterSelector
    ├── namespace: StringValueOrRef (required)
    ├── create_namespace: bool
    ├── helm_chart_version: string (default: "0.13.1")
    ├── replica_count: int32 (default: 1)
    ├── container: KubernetesGhaRunnerScaleSetControllerContainer
    │   ├── resources: ContainerResources
    │   └── image: KubernetesGhaRunnerScaleSetControllerImage
    ├── flags: KubernetesGhaRunnerScaleSetControllerFlags
    │   ├── log_level: LogLevel
    │   ├── log_format: LogFormat
    │   ├── watch_single_namespace: string
    │   ├── runner_max_concurrent_reconciles: int32
    │   ├── update_strategy: UpdateStrategy
    │   └── exclude_label_propagation_prefixes: []string
    ├── metrics: KubernetesGhaRunnerScaleSetControllerMetrics
    ├── image_pull_secrets: []string
    └── priority_class_name: string
```

### Deployment Model

**Installation Method**: Helm chart (gha-runner-scale-set-controller)  
**Controller Pod**: Single or multi-replica deployment  
**CRDs Installed**: AutoScalingRunnerSet, AutoScalingListener, EphemeralRunner, EphemeralRunnerSet

## Configuration

### Namespace Configuration

The controller can be installed in any namespace:

```yaml
spec:
  namespace:
    value: arc-system
  createNamespace: true
```

### Resource Specification

The `spec.container.resources` field controls the controller pod's resource allocation:

```yaml
spec:
  container:
    resources:
      requests:
        cpu: "100m"
        memory: "128Mi"
      limits:
        cpu: "500m"
        memory: "512Mi"
```

### Controller Flags

Configure controller behavior:

```yaml
spec:
  flags:
    logLevel: info           # debug, info, warn, error
    logFormat: json          # text, json
    runnerMaxConcurrentReconciles: 5
    updateStrategy: eventual # immediate, eventual
```

### Metrics Configuration

Enable Prometheus metrics:

```yaml
spec:
  metrics:
    controllerManagerAddr: ":8080"
    listenerAddr: ":8080"
    listenerEndpoint: "/metrics"
```

## Usage Patterns

### Basic Installation

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesGhaRunnerScaleSetController
metadata:
  name: arc-controller
spec:
  namespace:
    value: arc-system
  createNamespace: true
  container:
    resources:
      requests:
        cpu: "100m"
        memory: "128Mi"
      limits:
        cpu: "500m"
        memory: "512Mi"
```

### High Availability Setup

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesGhaRunnerScaleSetController
metadata:
  name: arc-controller-ha
spec:
  namespace:
    value: arc-system
  createNamespace: true
  replicaCount: 3
  priorityClassName: system-cluster-critical
  container:
    resources:
      requests:
        cpu: "200m"
        memory: "256Mi"
      limits:
        cpu: "1000m"
        memory: "1Gi"
```

## Post-Installation

After deploying the controller, create runner scale sets:

### Example: AutoScalingRunnerSet

```yaml
apiVersion: actions.github.com/v1alpha1
kind: AutoScalingRunnerSet
metadata:
  name: arc-runner-set
  namespace: arc-runners
spec:
  githubConfigUrl: "https://github.com/myorg"
  githubConfigSecret: github-auth
  maxRunners: 10
  minRunners: 0
  template:
    spec:
      containers:
        - name: runner
          image: ghcr.io/actions/actions-runner:latest
          resources:
            requests:
              cpu: "500m"
              memory: "1Gi"
```

## Benefits

### Cost Optimization

- **Scale to Zero**: No idle runners consuming resources
- **Right-Sized Runners**: Configure runner resources per workflow needs
- **Ephemeral**: No persistent storage costs for runners

### Security

- **Ephemeral Runners**: Fresh environment for each job
- **Network Isolation**: Runners in their own namespace
- **RBAC Integration**: Use K8s service accounts for fine-grained access

### Operational Efficiency

- **Self-Healing**: Kubernetes manages runner pod lifecycle
- **Declarative**: GitOps-friendly configuration
- **Observable**: Prometheus metrics and logging

## Documentation

For detailed information, see:

- **Research Documentation**: [docs/README.md](docs/README.md) - Comprehensive guide on runner deployment patterns
- **Examples**: [examples.md](examples.md) - Practical usage examples
- **Pulumi Module**: [iac/pulumi/README.md](iac/pulumi/README.md) - Pulumi-specific details
- **Terraform Module**: [iac/tf/README.md](iac/tf/README.md) - Terraform-specific details

## Version Information

- **Helm Chart**: 0.13.1 (default, configurable via `helm_chart_version`)
- **Controller Image**: ghcr.io/actions/gha-runner-scale-set-controller
- **Supported Kubernetes**: 1.23+

## Support

For issues, questions, or contributions:

- **GitHub**: [plantonhq/planton](https://github.com/plantonhq/planton)
- **Actions Runner Controller**: [actions/actions-runner-controller](https://github.com/actions/actions-runner-controller)
- **Documentation**: [ARC Documentation](https://docs.github.com/en/actions/hosting-your-own-runners/managing-self-hosted-runners-with-actions-runner-controller/about-actions-runner-controller)

