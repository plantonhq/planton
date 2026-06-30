---
title: "Gitlab"
description: "Gitlab deployment documentation"
icon: "package"
order: 100
componentName: "kubernetesgitlab"
---

# Kubernetes Gitlab

Deploys a GitLab instance on Kubernetes with a ClusterIP Service, optional namespace creation, configurable container resources, and optional Ingress with TLS termination via cert-manager and Istio.

## What Gets Created

When you deploy a KubernetesGitlab resource, Planton provisions:

- **Namespace** — created only when `createNamespace` is `true`
- **ClusterIP Service** — exposes GitLab on port 80 (targeting container port 8080) with app-level selectors derived from the resource metadata
- **Ingress** — created only when `ingress.enabled` is `true`, routes HTTPS traffic to the Service using the Istio ingress class with automatic TLS certificates from cert-manager (letsencrypt-prod issuer)

## Prerequisites

- **Kubernetes credentials** configured via environment variables or Planton provider config
- **A Kubernetes namespace** that already exists, or set `createNamespace` to `true`
- **Istio** installed in the cluster if enabling ingress (the Ingress uses `ingressClassName: istio`)
- **cert-manager** with a `letsencrypt-prod` ClusterIssuer if enabling ingress with TLS

## Quick Start

Create a file `gitlab.yaml`:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesGitlab
metadata:
  name: my-gitlab
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.KubernetesGitlab.my-gitlab
spec:
  namespace: gitlab
  createNamespace: true
  container:
    resources:
      limits:
        cpu: "1000m"
        memory: "1Gi"
      requests:
        cpu: "50m"
        memory: "100Mi"
```

Deploy:

```shell
planton apply -f gitlab.yaml
```

This creates a GitLab instance in the `gitlab` namespace with a ClusterIP Service on port 80, using the default resource limits.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `namespace` | `string` | Kubernetes namespace for the GitLab deployment. Can reference a KubernetesNamespace resource via `valueFrom`. | Required |
| `container` | `object` | Container specification for the GitLab deployment. | Required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `targetCluster.clusterKind` | `enum` | — | Kubernetes cluster kind. Valid values: `AwsEksCluster`, `GcpGkeCluster`, `AzureAksCluster`, `DigitalOceanKubernetesCluster`, `CivoKubernetesCluster`. |
| `targetCluster.clusterName` | `string` | — | Name of the target Kubernetes cluster in the same environment. |
| `createNamespace` | `bool` | `false` | When `true`, creates the namespace before deploying resources. |
| `container.resources.limits.cpu` | `string` | `1000m` | Maximum CPU allocation for the GitLab container. |
| `container.resources.limits.memory` | `string` | `1Gi` | Maximum memory allocation for the GitLab container. |
| `container.resources.requests.cpu` | `string` | `50m` | Minimum guaranteed CPU for the GitLab container. |
| `container.resources.requests.memory` | `string` | `100Mi` | Minimum guaranteed memory for the GitLab container. |
| `ingress.enabled` | `bool` | `false` | Creates a Kubernetes Ingress resource with Istio ingress class and cert-manager TLS. |
| `ingress.hostname` | `string` | — | Full hostname for external access (e.g., `gitlab.example.com`). Required when `ingress.enabled` is `true`. |

## Examples

### Development GitLab with Minimal Resources

A lightweight GitLab instance for development with reduced CPU and memory:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesGitlab
metadata:
  name: dev-gitlab
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.KubernetesGitlab.dev-gitlab
spec:
  namespace: dev
  createNamespace: true
  container:
    resources:
      limits:
        cpu: "500m"
        memory: "512Mi"
      requests:
        cpu: "50m"
        memory: "128Mi"
```

### Production GitLab with Ingress

A production GitLab instance with higher resource limits and HTTPS ingress:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesGitlab
metadata:
  name: prod-gitlab
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.KubernetesGitlab.prod-gitlab
spec:
  namespace: production
  container:
    resources:
      limits:
        cpu: "4000m"
        memory: "8Gi"
      requests:
        cpu: "1000m"
        memory: "2Gi"
  ingress:
    enabled: true
    hostname: gitlab.example.com
```

### Using Foreign Key References

Reference an Planton-managed namespace instead of hardcoding the name:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesGitlab
metadata:
  name: team-gitlab
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: staging.KubernetesGitlab.team-gitlab
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: platform-ns
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
    hostname: gitlab.staging.example.com
```

> **Note:** The `namespace` field accepts either a plain string value or a `valueFrom` reference to another resource. When using `valueFrom`, the value is resolved at deployment time from the referenced resource's field.

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `namespace` | `string` | Kubernetes namespace where GitLab is deployed |
| `service` | `string` | Kubernetes Service name for the GitLab instance |
| `portForwardCommand` | `string` | kubectl port-forward command for local access (e.g., `kubectl port-forward -n gitlab service/my-gitlab 8080:80`) |
| `kubeEndpoint` | `string` | Cluster-internal FQDN (e.g., `my-gitlab.gitlab.svc.cluster.local`) |
| `ingressEndpoint` | `string` | Public HTTPS endpoint for external access, only set when ingress is enabled |

## Related Components

- [KubernetesNamespace](/docs/catalog/kubernetes/namespace) — provides the target namespace via `valueFrom` reference
- [KubernetesDeployment](/docs/catalog/kubernetes/deployment) — application deployments that integrate with GitLab
- [KubernetesRedis](/docs/catalog/kubernetes/redis) — Redis cache commonly used alongside GitLab for session storage and caching
- [KubernetesPostgres](/docs/catalog/kubernetes/postgres) — PostgreSQL database used by GitLab as its primary data store
