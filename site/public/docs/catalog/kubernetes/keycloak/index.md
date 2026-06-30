---
title: "Keycloak"
description: "Keycloak deployment documentation"
icon: "package"
order: 100
componentName: "kuberneteskeycloak"
---

# Kubernetes Keycloak

Deploys Keycloak on Kubernetes as an identity and access management solution. Provisions a Keycloak instance with configurable container resources, optional namespace creation, and optional external access through ingress with TLS. Keycloak provides single sign-on, identity brokering, user federation, and fine-grained authorization for applications and services.

## What Gets Created

When you deploy a KubernetesKeycloak resource, Planton provisions:

- **Kubernetes Namespace** — created if `createNamespace` is `true`
- **Keycloak Deployment** — a Keycloak application instance with:
  - A Keycloak pod running with the configured CPU and memory resources
  - Kubernetes Service for cluster-internal access on port 8080
  - Admin password stored in a Kubernetes Secret (`{name}-password`)
  - PostgreSQL database password stored in a Kubernetes Secret (`{name}-db-password`)
- **Ingress Resources** (when `ingress.enabled` is `true`):
  - External LoadBalancer service (`{name}-external-lb`) for routing traffic to Keycloak
  - TLS-terminated external access at the configured hostname

## Prerequisites

- **A Kubernetes cluster** with kubectl configured for access
- **Ingress controller** installed in the cluster (only if using ingress)
- **cert-manager** or equivalent TLS provider (only if using ingress with HTTPS)

## Quick Start

Create a file `keycloak.yaml`:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesKeycloak
metadata:
  name: my-keycloak
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.KubernetesKeycloak.my-keycloak
spec:
  namespace:
    value: keycloak-dev
  createNamespace: true
```

Deploy:

```shell
planton apply -f keycloak.yaml
```

This creates a Keycloak instance with default resources (1 CPU / 1Gi memory limit, 50m CPU / 100Mi memory request) in the `keycloak-dev` namespace.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `namespace` | `StringValueOrRef` | Kubernetes namespace for the Keycloak deployment. Use `value` for a direct string or `valueFrom` to reference a KubernetesNamespace resource. | Required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `createNamespace` | `bool` | `false` | Create the namespace if it does not exist. |
| `container.resources.limits.cpu` | `string` | `"1000m"` | CPU limit for the Keycloak container. |
| `container.resources.limits.memory` | `string` | `"1Gi"` | Memory limit for the Keycloak container. |
| `container.resources.requests.cpu` | `string` | `"50m"` | CPU request for the Keycloak container. |
| `container.resources.requests.memory` | `string` | `"100Mi"` | Memory request for the Keycloak container. |
| `ingress.enabled` | `bool` | `false` | Enable external access to Keycloak via ingress. |
| `ingress.hostname` | `string` | — | Full hostname for external access (e.g., `keycloak.example.com`). Required when `ingress.enabled` is `true`. |

## Examples

### Keycloak with Custom Resources

Increase CPU and memory for a Keycloak instance handling a larger user base:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesKeycloak
metadata:
  name: auth-keycloak
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.KubernetesKeycloak.auth-keycloak
spec:
  namespace:
    value: auth-services
  createNamespace: true
  container:
    resources:
      limits:
        cpu: "2000m"
        memory: "4Gi"
      requests:
        cpu: "500m"
        memory: "1Gi"
```

### Keycloak with Namespace Reference

Use `valueFrom` to reference a namespace managed by a separate KubernetesNamespace resource:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesKeycloak
metadata:
  name: shared-keycloak
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: staging.KubernetesKeycloak.shared-keycloak
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      metadata:
        name: platform-ns
      fieldPath: spec.name
  container:
    resources:
      limits:
        cpu: "2000m"
        memory: "2Gi"
      requests:
        cpu: "250m"
        memory: "512Mi"
```

### Full-Featured with Ingress

External access over HTTPS for production use:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesKeycloak
metadata:
  name: prod-keycloak
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.KubernetesKeycloak.prod-keycloak
spec:
  namespace:
    value: production
  createNamespace: true
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
    hostname: keycloak.example.com
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `namespace` | `string` | Kubernetes namespace where Keycloak was created |
| `service` | `string` | Name of the Kubernetes service for Keycloak |
| `port_forward_command` | `string` | Ready-to-run `kubectl port-forward` command for local access on port 8080 |
| `kube_endpoint` | `string` | Cluster-internal endpoint (e.g., `my-keycloak.keycloak-dev.svc.cluster.local:8080`) |
| `external_hostname` | `string` | External HTTPS hostname when ingress is enabled (e.g., `https://keycloak.example.com`) |
| `internal_hostname` | `string` | Internal HTTPS hostname for private access (e.g., `https://internal-keycloak.example.com`) |

## Related Components

- [KubernetesNamespace](/docs/catalog/kubernetes/namespace) — pre-create a namespace to reference via `valueFrom`
- [KubernetesPostgres](/docs/catalog/kubernetes/postgres) — deploy PostgreSQL as the backing database for Keycloak
- [KubernetesJenkins](/docs/catalog/kubernetes/jenkins) — integrate Jenkins CI/CD with Keycloak for authentication
