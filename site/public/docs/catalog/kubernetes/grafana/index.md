---
title: "Grafana"
description: "Grafana deployment documentation"
icon: "package"
order: 100
componentName: "kubernetesgrafana"
---

# Kubernetes Grafana

Deploys Grafana on Kubernetes using the official Grafana Helm chart (v8.7.0). Provisions a ClusterIP service with configurable container resources, optional namespace creation, and optional external/internal ingress via nginx ingress controllers.

## What Gets Created

When you deploy a KubernetesGrafana resource, OpenMCF provisions:

- **Kubernetes Namespace** — created if `createNamespace` is `true`
- **Grafana Helm Release** — the official `grafana` chart (v8.7.0) from `https://grafana.github.io/helm-charts`, which creates:
  - A Grafana pod with default admin credentials (`admin` / `admin`)
  - Kubernetes ClusterIP Service on port 80 for cluster-internal access
  - Persistence disabled by default
- **Ingress Resources** (when `ingress.enabled` is `true`):
  - External Ingress — routes traffic from the configured hostname to the Grafana service using the `nginx` ingress class
  - Internal Ingress — routes traffic from an `internal-` prefixed hostname to the same service using the `nginx-internal` ingress class

## Prerequisites

- **A Kubernetes cluster** with kubectl configured for access
- **nginx ingress controller** installed (only if using ingress)
- **nginx-internal ingress controller** installed (only if using internal ingress)

## Quick Start

Create a file `grafana.yaml`:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesGrafana
metadata:
  name: my-grafana
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.KubernetesGrafana.my-grafana
spec:
  namespace:
    value: grafana-dev
  createNamespace: true
```

Deploy:

```shell
openmcf apply -f grafana.yaml
```

This creates a Grafana instance with default resources (1 CPU / 1Gi memory limit, 50m CPU / 100Mi memory request) in the `grafana-dev` namespace. Access the dashboard with `admin` / `admin` via the port-forward command in the stack outputs.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `namespace` | `StringValueOrRef` | Kubernetes namespace for the Grafana deployment. Use `value` for a direct string or `valueFrom` to reference a KubernetesNamespace resource. | Required |
| `container` | `object` | Container specification including resource allocations. | Required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `createNamespace` | `bool` | `false` | Create the namespace if it does not exist. |
| `container.resources.limits.cpu` | `string` | `"1000m"` | CPU limit for the Grafana container. |
| `container.resources.limits.memory` | `string` | `"1Gi"` | Memory limit for the Grafana container. |
| `container.resources.requests.cpu` | `string` | `"50m"` | CPU request for the Grafana container. |
| `container.resources.requests.memory` | `string` | `"100Mi"` | Memory request for the Grafana container. |
| `ingress.enabled` | `bool` | `false` | Enable external and internal access via nginx ingress. |
| `ingress.hostname` | `string` | -- | Full hostname for external access (e.g., `grafana.example.com`). Required when `ingress.enabled` is `true`. An internal ingress is also created at `internal-{hostname}`. |

## Examples

### Grafana with Custom Resources

Increase CPU and memory for a busier monitoring environment:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesGrafana
metadata:
  name: monitoring-grafana
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.KubernetesGrafana.monitoring-grafana
spec:
  namespace:
    value: monitoring
  createNamespace: true
  container:
    resources:
      limits:
        cpu: "2000m"
        memory: "4Gi"
      requests:
        cpu: "250m"
        memory: "512Mi"
```

### Grafana in an Existing Namespace

Deploy into a pre-existing namespace without creating it, and reference a KubernetesNamespace resource via `valueFrom`:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesGrafana
metadata:
  name: team-grafana
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: staging.KubernetesGrafana.team-grafana
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      metadata:
        name: observability-ns
      fieldPath: spec.name
  container:
    resources:
      limits:
        cpu: "1000m"
        memory: "2Gi"
      requests:
        cpu: "100m"
        memory: "256Mi"
```

### Full-Featured with Ingress

External and internal access through nginx ingress controllers:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesGrafana
metadata:
  name: prod-grafana
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.KubernetesGrafana.prod-grafana
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
        cpu: "500m"
        memory: "1Gi"
  ingress:
    enabled: true
    hostname: grafana.example.com
```

This creates two ingress resources: one external at `grafana.example.com` (nginx ingress class) and one internal at `internal-grafana.example.com` (nginx-internal ingress class).

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `namespace` | `string` | Kubernetes namespace where Grafana was created |
| `service` | `string` | Name of the Kubernetes service for Grafana (e.g., `my-grafana-grafana`) |
| `portForwardCommand` | `string` | Ready-to-run `kubectl port-forward` command for local access on port 8080 |
| `kubeEndpoint` | `string` | Cluster-internal endpoint (e.g., `my-grafana-grafana.grafana-dev.svc.cluster.local`) |
| `externalHostname` | `string` | External URL when ingress is enabled (e.g., `https://grafana.example.com`) |
| `internalHostname` | `string` | Internal URL for private ingress (e.g., `https://internal-grafana.example.com`) |

## Related Components

- [KubernetesNamespace](/docs/catalog/kubernetes/namespace) — pre-create a namespace to reference via `valueFrom`
- [KubernetesIngressNginx](/docs/catalog/kubernetes/ingress-nginx) — deploy the nginx ingress controller required for Grafana ingress
- [KubernetesPostgres](/docs/catalog/kubernetes/postgres) — deploy PostgreSQL as a Grafana database backend
- [KubernetesRedis](/docs/catalog/kubernetes/redis) — deploy Redis for Grafana caching
