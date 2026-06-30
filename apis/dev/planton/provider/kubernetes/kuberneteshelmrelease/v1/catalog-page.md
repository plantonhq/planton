# Kubernetes Helm Release

Deploys any Helm chart to a Kubernetes cluster through Planton's lifecycle management, acting as a generic escape hatch for workloads that are already packaged as Helm charts but do not have a dedicated Planton component. The module handles chart fetching, namespace creation, value overrides, and the full apply/update/destroy lifecycle automatically.

## What Gets Created

When you deploy a KubernetesHelmRelease resource, Planton provisions:

- **Namespace** — created only when `createNamespace` is `true`
- **Helm Chart Resources** — all Kubernetes resources defined by the Helm chart are rendered and applied via Pulumi's `helm/v3.Chart`, using the specified chart name, version, repository, and custom value overrides

## Prerequisites

- **Kubernetes credentials** configured via environment variables or Planton provider config
- **A Kubernetes namespace** that already exists, or set `createNamespace` to `true`
- **A reachable Helm chart repository** hosting the chart at the specified version
- **CRD definitions available on the cluster** if the Helm chart deploys or references custom resource types

## Quick Start

Create a file `helm-release.yaml`:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesHelmRelease
metadata:
  name: my-nginx
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.KubernetesHelmRelease.my-nginx
spec:
  namespace:
    value: ingress
  createNamespace: true
  repo: https://charts.bitnami.com/bitnami
  name: nginx
  version: "18.1.11"
```

Deploy:

```shell
planton apply -f helm-release.yaml
```

This deploys the Bitnami nginx chart at version 18.1.11 into the `ingress` namespace, creating the namespace if it does not already exist.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `namespace` | `StringValueOrRef` | Kubernetes namespace for the Helm release. Can reference a KubernetesNamespace resource via `valueFrom`. | Required |
| `repo` | `string` | URL of the Helm chart repository (e.g., `https://charts.bitnami.com/bitnami`). | Required |
| `name` | `string` | Name of the Helm chart within the repository (e.g., `nginx`, `redis`). | Required |
| `version` | `string` | Semantic version of the Helm chart to deploy (e.g., `18.1.11`). | Required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `targetCluster.clusterKind` | `enum` | — | Kubernetes cluster kind. Valid values: `AwsEksCluster`, `GcpGkeCluster`, `AzureAksCluster`, `DigitalOceanKubernetesCluster`, `CivoKubernetesCluster`. |
| `targetCluster.clusterName` | `string` | — | Name of the target Kubernetes cluster in the same environment. |
| `createNamespace` | `bool` | `false` | When `true`, creates the namespace before deploying the Helm release. When `false`, the namespace must already exist. |
| `values` | `map<string, string>` | — | Key-value pairs that override defaults in the chart's `values.yaml`. Each key uses Helm dot-notation flattened to a single string key. |

## Examples

### Minimal Nginx Ingress Controller

Deploys the ingress-nginx controller with default settings:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesHelmRelease
metadata:
  name: ingress-nginx
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.KubernetesHelmRelease.ingress-nginx
spec:
  namespace:
    value: ingress-nginx
  createNamespace: true
  repo: https://kubernetes.github.io/ingress-nginx
  name: ingress-nginx
  version: "4.11.3"
```

### Prometheus Stack with Custom Values

Deploys kube-prometheus-stack with custom retention, resource limits, and Grafana disabled:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesHelmRelease
metadata:
  name: prometheus
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: staging.KubernetesHelmRelease.prometheus
spec:
  namespace:
    value: monitoring
  createNamespace: true
  repo: https://prometheus-community.github.io/helm-charts
  name: kube-prometheus-stack
  version: "65.1.0"
  values:
    grafana.enabled: "false"
    prometheus.prometheusSpec.retention: "30d"
    prometheus.prometheusSpec.resources.limits.cpu: "2000m"
    prometheus.prometheusSpec.resources.limits.memory: "4Gi"
    prometheus.prometheusSpec.resources.requests.cpu: "500m"
    prometheus.prometheusSpec.resources.requests.memory: "1Gi"
    prometheus.prometheusSpec.storageSpec.volumeClaimTemplate.spec.resources.requests.storage: "50Gi"
```

### Cert-Manager with Target Cluster and Foreign Key Namespace

Deploys cert-manager on a specific GKE cluster, referencing an Planton-managed namespace:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesHelmRelease
metadata:
  name: cert-manager
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.KubernetesHelmRelease.cert-manager
spec:
  targetCluster:
    clusterKind: GcpGkeCluster
    clusterName: prod-cluster
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: cert-manager-ns
      field: spec.name
  repo: https://charts.jetstack.io
  name: cert-manager
  version: "1.16.2"
  values:
    crds.enabled: "true"
    replicaCount: "2"
    resources.limits.cpu: "500m"
    resources.limits.memory: "512Mi"
    resources.requests.cpu: "100m"
    resources.requests.memory: "128Mi"
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `namespace` | `string` | Kubernetes namespace where the Helm release was deployed |

## Related Components

- [KubernetesNamespace](/docs/catalog/kubernetes/kubernetesnamespace) — provides the target namespace via `valueFrom` reference
- [KubernetesManifest](/docs/catalog/kubernetes/kubernetesmanifest) — preferred when deploying raw Kubernetes YAML rather than a packaged Helm chart
- [KubernetesDeployment](/docs/catalog/kubernetes/kubernetesdeployment) — preferred for containerized applications that need built-in Service, ingress, and autoscaling support
