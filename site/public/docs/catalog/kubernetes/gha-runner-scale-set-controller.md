---
title: "GHA Runner Scale Set Controller"
description: "GHA Runner Scale Set Controller deployment documentation"
icon: "package"
order: 100
componentName: "kubernetesgharunnerscalesetcontroller"
---

# Kubernetes GHA Runner Scale Set Controller

Deploys the GitHub Actions Runner Scale Set Controller on Kubernetes using the official OCI Helm chart from `ghcr.io/actions/actions-runner-controller-charts`. The controller manages `AutoScalingRunnerSet` and `EphemeralRunner` custom resources, enabling self-hosted GitHub Actions runners that scale dynamically based on workflow demand. This component installs only the controller; runner scale sets (the actual runner pods) are deployed separately.

## What Gets Created

When you deploy a KubernetesGhaRunnerScaleSetController resource, OpenMCF provisions:

- **Kubernetes Namespace** — created if `createNamespace` is `true`
- **Helm Release** — the `gha-runner-scale-set-controller` OCI chart (default version 0.13.1) from `ghcr.io/actions/actions-runner-controller-charts`, which creates:
  - A controller Deployment that watches for `AutoScalingRunnerSet` resources and manages `EphemeralRunner` pods
  - A ServiceAccount for the controller
  - CRD management for `AutoScalingRunnerSet` and `EphemeralRunner` resource types
  - Leader election support when running multiple replicas
  - Metrics endpoints (when metrics configuration is provided)

## Prerequisites

- **A Kubernetes cluster** with kubectl configured for access
- **Helm 3** available (the module uses an OCI-based Helm chart)
- **A GitHub App or PAT** configured for runner registration (required when deploying runner scale sets, not for the controller itself)

## Quick Start

Create a file `gha-controller.yaml`:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesGhaRunnerScaleSetController
metadata:
  name: arc-controller
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.KubernetesGhaRunnerScaleSetController.arc-controller
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

Deploy:

```shell
openmcf apply -f gha-controller.yaml
```

This installs the controller in the `arc-system` namespace with default resource limits. Once running, you can deploy `AutoScalingRunnerSet` resources in any namespace to register self-hosted GitHub Actions runners.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `namespace` | `StringValueOrRef` | Kubernetes namespace where the controller will be installed. Use `value` for a direct string or `valueFrom` to reference a KubernetesNamespace resource. | Required |
| `container` | `object` | Container specifications for the controller pod, including CPU/memory resources. | Required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `createNamespace` | `bool` | `false` | Create the namespace if it does not exist. |
| `helmChartVersion` | `string` | `"0.13.1"` | Version of the Helm chart to deploy. Chart versions match controller image versions. See [releases](https://github.com/actions/actions-runner-controller/releases). |
| `replicaCount` | `int32` | `1` | Number of controller replicas. Leader election is automatically enabled when greater than 1. |
| `container.resources.requests.cpu` | `string` | `"100m"` | CPU request for the controller container. |
| `container.resources.requests.memory` | `string` | `"128Mi"` | Memory request for the controller container. |
| `container.resources.limits.cpu` | `string` | `"500m"` | CPU limit for the controller container. |
| `container.resources.limits.memory` | `string` | `"512Mi"` | Memory limit for the controller container. |
| `container.image.repository` | `string` | `"ghcr.io/actions/gha-runner-scale-set-controller"` | Custom container image repository. |
| `container.image.tag` | `string` | chart appVersion | Custom image tag. |
| `container.image.pullPolicy` | `string` | -- | Image pull policy: `Always`, `IfNotPresent`, or `Never`. |
| `flags.logLevel` | `enum` | `"debug"` | Log level for the controller. Valid values: `debug`, `info`, `warn`, `error`. |
| `flags.logFormat` | `enum` | `"text"` | Log format. Valid values: `text`, `json`. |
| `flags.watchSingleNamespace` | `string` | -- | Restrict the controller to watch only the specified namespace. By default, watches all namespaces. |
| `flags.runnerMaxConcurrentReconciles` | `int32` | `2` | Maximum concurrent reconciles for the EphemeralRunner controller. |
| `flags.updateStrategy` | `enum` | `"immediate"` | How upgrades are handled while jobs are running. `immediate` applies changes right away (may cause overprovisioning). `eventual` waits for running jobs to complete. |
| `flags.excludeLabelPropagationPrefixes` | `string[]` | `[]` | Label prefixes to exclude from propagation to internal resources (e.g., ArgoCD labels). |
| `flags.k8sClientRateLimiterQps` | `int32` | `0` | Kubernetes API client rate limiter QPS. |
| `flags.k8sClientRateLimiterBurst` | `int32` | `0` | Kubernetes API client rate limiter burst. |
| `metrics.controllerManagerAddr` | `string` | -- | Metrics address for the controller manager (e.g., `":8080"`). Providing this value enables metrics. |
| `metrics.listenerAddr` | `string` | -- | Metrics address for the listener (e.g., `":8080"`). |
| `metrics.listenerEndpoint` | `string` | -- | Metrics endpoint path for the listener (e.g., `"/metrics"`). |
| `imagePullSecrets` | `string[]` | `[]` | Image pull secrets for private container registries. Also passed to the auto-scaler for pulling listener images. |
| `priorityClassName` | `string` | -- | Priority class name for controller pods. Use `"system-cluster-critical"` to ensure the controller survives resource pressure. |

## Examples

### Minimal Controller with Defaults

Deploy the controller with default settings in a dedicated namespace:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesGhaRunnerScaleSetController
metadata:
  name: arc-controller
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.KubernetesGhaRunnerScaleSetController.arc-controller
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

### High-Availability with Structured Logging

Run multiple replicas for high availability and configure JSON logging for structured log aggregation:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesGhaRunnerScaleSetController
metadata:
  name: arc-controller-ha
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: staging.KubernetesGhaRunnerScaleSetController.arc-controller-ha
spec:
  namespace:
    value: arc-system
  createNamespace: true
  helmChartVersion: "0.13.1"
  replicaCount: 3
  container:
    resources:
      requests:
        cpu: "250m"
        memory: "256Mi"
      limits:
        cpu: "1000m"
        memory: "1Gi"
  flags:
    logLevel: info
    logFormat: json
    updateStrategy: eventual
    runnerMaxConcurrentReconciles: 5
  priorityClassName: system-cluster-critical
```

### Production with Metrics, Private Registry, and Namespace Scoping

Full production configuration with metrics enabled, a private container registry, and the controller scoped to a single namespace:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesGhaRunnerScaleSetController
metadata:
  name: arc-controller-prod
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.KubernetesGhaRunnerScaleSetController.arc-controller-prod
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: arc-system-prod
      fieldPath: spec.name
  container:
    resources:
      requests:
        cpu: "500m"
        memory: "512Mi"
      limits:
        cpu: "2000m"
        memory: "2Gi"
    image:
      repository: registry.internal.example.com/actions/gha-runner-scale-set-controller
      tag: "0.13.1"
      pullPolicy: IfNotPresent
  replicaCount: 3
  flags:
    logLevel: info
    logFormat: json
    watchSingleNamespace: runners-prod
    updateStrategy: eventual
    runnerMaxConcurrentReconciles: 10
    excludeLabelPropagationPrefixes:
      - argocd.argoproj.io/
      - app.kubernetes.io/managed-by
    k8sClientRateLimiterQps: 50
    k8sClientRateLimiterBurst: 100
  metrics:
    controllerManagerAddr: ":8080"
    listenerAddr: ":8080"
    listenerEndpoint: "/metrics"
  imagePullSecrets:
    - registry-credentials
  priorityClassName: system-cluster-critical
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `namespace` | `string` | Namespace where the controller is deployed |
| `releaseName` | `string` | Name of the Helm release |
| `chartVersion` | `string` | Version of the deployed Helm chart |
| `deploymentName` | `string` | Name of the controller Deployment |
| `serviceAccountName` | `string` | Name of the controller ServiceAccount |
| `metricsEndpoint` | `string` | Controller metrics endpoint in `<service>.<namespace>.svc.cluster.local:<port>` format (only present when metrics are enabled) |

## Related Components

- [KubernetesNamespace](/docs/catalog/kubernetes/kubernetesnamespace) — pre-create a namespace to reference via `valueFrom`
- [KubernetesHelmRelease](/docs/catalog/kubernetes/kuberneteshelmrelease) — deploy additional Helm charts, such as the `gha-runner-scale-set` chart for runner pods
- [KubernetesDeployment](/docs/catalog/kubernetes/kubernetesdeployment) — deploy custom workloads alongside the controller
- [KubernetesSecret](/docs/catalog/kubernetes/kubernetessecret) — manage secrets for GitHub App credentials or PAT tokens used by runner scale sets
