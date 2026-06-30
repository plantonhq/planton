---
title: "Argo CD"
description: "Argo CD deployment documentation"
icon: "package"
order: 100
componentName: "kubernetesargocd"
---

# Kubernetes Argo CD

Deploys Argo CD on Kubernetes using the official Argo Helm chart (argo-cd v7.7.12), with configurable resource limits for the server, controller, repo-server, and Redis components, optional namespace creation, and optional ingress for external browser access with automatic TLS certificate provisioning.

## What Gets Created

When you deploy a KubernetesArgocd resource, Planton provisions:

- **Namespace** — created only when `createNamespace` is `true`
- **Helm Release (Argo CD)** — deploys Argo CD from the `argo-cd` chart at `https://argoproj.github.io/argo-helm`, pinned to version 7.7.12, with atomic rollback enabled and a 10-minute timeout; configures resource requests/limits for the server, application controller, repo-server, and embedded Redis
- **Ingress** — when enabled, exposes the Argo CD server externally at the specified hostname with TLS termination via a cert-manager ClusterIssuer derived from the hostname's domain

## Prerequisites

- **Kubernetes credentials** configured via environment variables or Planton provider config
- **A Kubernetes namespace** that already exists, or set `createNamespace` to `true`
- **cert-manager** with a ClusterIssuer named after the parent domain (e.g., `example.com`) if enabling ingress with TLS
- **An ingress controller** running in the cluster if enabling external access

## Quick Start

Create a file `argocd.yaml`:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesArgocd
metadata:
  name: my-argocd
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.KubernetesArgocd.my-argocd
spec:
  namespace: argocd
  createNamespace: true
  container: {}
```

Deploy:

```shell
planton apply -f argocd.yaml
```

This creates an Argo CD instance in the `argocd` namespace with default resource limits (1000m CPU / 1Gi memory limits, 50m CPU / 100Mi memory requests) and no external ingress. Access the UI locally with the port-forward command from stack outputs.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `namespace` | `string` | Kubernetes namespace for the Argo CD deployment. Can reference a KubernetesNamespace resource via `valueFrom`. | Required |
| `container` | `object` | Container specification for the Argo CD components. Pass `{}` to accept all defaults. | Required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `targetCluster.clusterKind` | `enum` | — | Kubernetes cluster kind. Valid values: `AwsEksCluster`, `GcpGkeCluster`, `AzureAksCluster`, `DigitalOceanKubernetesCluster`, `CivoKubernetesCluster`. |
| `targetCluster.clusterName` | `string` | — | Name of the target Kubernetes cluster in the same environment. |
| `createNamespace` | `bool` | `false` | When `true`, creates the namespace before deploying resources. |
| `container.resources.limits.cpu` | `string` | `1000m` | Maximum CPU allocation for the Argo CD server, controller, and repo-server pods. |
| `container.resources.limits.memory` | `string` | `1Gi` | Maximum memory allocation for the Argo CD server, controller, and repo-server pods. |
| `container.resources.requests.cpu` | `string` | `50m` | Minimum guaranteed CPU for the Argo CD server, controller, and repo-server pods. |
| `container.resources.requests.memory` | `string` | `100Mi` | Minimum guaranteed memory for the Argo CD server, controller, and repo-server pods. |
| `ingress.enabled` | `bool` | `false` | Enables external access to the Argo CD web UI via ingress. |
| `ingress.hostname` | `string` | — | Full hostname for external access (e.g., `argocd.example.com`). Required when `ingress.enabled` is `true`. |

## Examples

### Development Instance with Minimal Resources

A lightweight Argo CD instance for development or testing with reduced resource allocations:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesArgocd
metadata:
  name: dev-argocd
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.KubernetesArgocd.dev-argocd
spec:
  namespace: argocd-dev
  createNamespace: true
  container:
    resources:
      limits:
        cpu: "500m"
        memory: "512Mi"
      requests:
        cpu: "50m"
        memory: "64Mi"
```

### Production Instance with Ingress

A production Argo CD deployment exposed externally with higher resource limits:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesArgocd
metadata:
  name: prod-argocd
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.KubernetesArgocd.prod-argocd
spec:
  namespace: argocd
  createNamespace: true
  container:
    resources:
      limits:
        cpu: "2000m"
        memory: "2Gi"
      requests:
        cpu: "250m"
        memory: "512Mi"
  ingress:
    enabled: true
    hostname: argocd.example.com
```

### Using Foreign Key References

Reference an Planton-managed namespace instead of hardcoding the name:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesArgocd
metadata:
  name: platform-argocd
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.KubernetesArgocd.platform-argocd
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: platform-namespace
      field: spec.name
  container:
    resources:
      limits:
        cpu: "2000m"
        memory: "4Gi"
      requests:
        cpu: "500m"
        memory: "1Gi"
  ingress:
    enabled: true
    hostname: argocd.platform.example.com
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `namespace` | `string` | Kubernetes namespace where Argo CD is deployed |
| `service` | `string` | Kubernetes Service name for the Argo CD server (format: `{name}-argocd-server`) |
| `port_forward_command` | `string` | kubectl port-forward command for local UI access on `http://localhost:8080` |
| `kube_endpoint` | `string` | Cluster-internal FQDN (e.g., `my-argocd-argocd-server.argocd.svc.cluster.local`) |
| `external_hostname` | `string` | Public hostname for browser access, only set when ingress is enabled |
| `internal_hostname` | `string` | Internal hostname for VPC-internal access (format: `internal-{hostname}`), only set when ingress is enabled |

## Related Components

- [KubernetesNamespace](/docs/catalog/kubernetes/namespace) — provides the target namespace via `valueFrom` reference
- [KubernetesDeployment](/docs/catalog/kubernetes/deployment) — application deployments managed by Argo CD
- [KubernetesHelmRelease](/docs/catalog/kubernetes/helm-release) — alternative for deploying Helm charts when full GitOps is not needed
