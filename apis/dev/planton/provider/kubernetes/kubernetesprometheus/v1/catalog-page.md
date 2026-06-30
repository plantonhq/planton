# Kubernetes Prometheus

Deploys a Prometheus monitoring instance on Kubernetes with configurable resource limits, optional persistent storage for metric data, and optional ingress for external access via a hostname.

## What Gets Created

When you deploy a KubernetesPrometheus resource, Planton provisions:

- **Namespace** — created only when `createNamespace` is `true`
- **Prometheus Deployment** — deploys Prometheus with configurable CPU, memory, replica count, and persistence settings
- **Kubernetes Service** — exposes Prometheus on port 9090 within the cluster (format: `{name}-prometheus`)
- **PersistentVolumeClaim** — created only when `container.persistenceEnabled` is `true`, sized according to `container.diskSize`, used to retain metric data across pod restarts
- **Ingress** — created only when `ingress.enabled` is `true`, routes external traffic to Prometheus using the configured hostname

## Prerequisites

- **Kubernetes credentials** configured via environment variables or Planton provider config
- **A Kubernetes namespace** that already exists, or set `createNamespace` to `true`
- **A StorageClass** available in the cluster if enabling persistence (most managed Kubernetes clusters provide a default)
- **A DNS-managed domain** if enabling ingress with a hostname

## Quick Start

Create a file `prometheus.yaml`:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesPrometheus
metadata:
  name: my-prometheus
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.KubernetesPrometheus.my-prometheus
spec:
  namespace: monitoring
  createNamespace: true
  container:
    replicas: 1
```

Deploy:

```shell
planton apply -f prometheus.yaml
```

This creates a single-replica Prometheus instance in the `monitoring` namespace with default resource limits (1000m CPU, 1Gi memory) and persistence disabled.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `namespace` | `string` | Kubernetes namespace for the Prometheus deployment. Can reference a KubernetesNamespace resource via `valueFrom`. | Required |
| `container` | `object` | Container specification for the Prometheus deployment. | Required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `targetCluster.clusterKind` | `enum` | — | Kubernetes cluster kind. Valid values: `AwsEksCluster`, `GcpGkeCluster`, `AzureAksCluster`, `DigitalOceanKubernetesCluster`, `CivoKubernetesCluster`. |
| `targetCluster.clusterName` | `string` | — | Name of the target Kubernetes cluster in the same environment. |
| `createNamespace` | `bool` | `false` | When `true`, creates the namespace before deploying resources. |
| `container.replicas` | `int32` | `1` | Number of Prometheus pods to deploy. |
| `container.resources.limits.cpu` | `string` | `1000m` | Maximum CPU allocation for each Prometheus pod. |
| `container.resources.limits.memory` | `string` | `1Gi` | Maximum memory allocation for each Prometheus pod. |
| `container.resources.requests.cpu` | `string` | `50m` | Minimum guaranteed CPU for each Prometheus pod. |
| `container.resources.requests.memory` | `string` | `100Mi` | Minimum guaranteed memory for each Prometheus pod. |
| `container.persistenceEnabled` | `bool` | `false` | Enables persistent storage for Prometheus metric data. When enabled, data is persisted to a PersistentVolumeClaim and restored on pod restart. |
| `container.diskSize` | `string` | — | Size of the PersistentVolumeClaim attached to each Prometheus pod. Required when `persistenceEnabled` is `true`. Must be a valid Kubernetes quantity (e.g., `10Gi`, `50Gi`). Cannot be modified after creation. |
| `ingress.enabled` | `bool` | `false` | Enables external access to the Prometheus web UI. |
| `ingress.hostname` | `string` | — | Full hostname for external access (e.g., `prometheus.example.com`). Required when `ingress.enabled` is `true`. |

## Examples

### Development Prometheus without Persistence

A lightweight Prometheus instance for development with reduced resources and no persistent storage:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesPrometheus
metadata:
  name: dev-prometheus
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.KubernetesPrometheus.dev-prometheus
spec:
  namespace: dev-monitoring
  createNamespace: true
  container:
    replicas: 1
    resources:
      limits:
        cpu: "500m"
        memory: "512Mi"
      requests:
        cpu: "100m"
        memory: "128Mi"
```

### Production Prometheus with Persistent Storage

A production Prometheus instance with larger disk allocation, higher resource limits, and data persistence enabled to retain metrics across pod restarts:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesPrometheus
metadata:
  name: prod-prometheus
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.KubernetesPrometheus.prod-prometheus
spec:
  namespace: monitoring
  container:
    replicas: 2
    resources:
      limits:
        cpu: "2000m"
        memory: "4Gi"
      requests:
        cpu: "500m"
        memory: "1Gi"
    persistenceEnabled: true
    diskSize: "50Gi"
```

### Prometheus with External Access

Prometheus exposed outside the cluster via ingress for access from a web browser:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesPrometheus
metadata:
  name: shared-prometheus
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.KubernetesPrometheus.shared-prometheus
spec:
  namespace: monitoring
  container:
    replicas: 2
    resources:
      limits:
        cpu: "2000m"
        memory: "4Gi"
      requests:
        cpu: "500m"
        memory: "1Gi"
    persistenceEnabled: true
    diskSize: "100Gi"
  ingress:
    enabled: true
    hostname: prometheus.example.com
```

### Using Foreign Key References

Reference an Planton-managed namespace instead of hardcoding the name:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesPrometheus
metadata:
  name: metrics
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.KubernetesPrometheus.metrics
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: monitoring-namespace
      field: spec.name
  container:
    replicas: 1
    persistenceEnabled: true
    diskSize: "20Gi"
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `namespace` | `string` | Kubernetes namespace where Prometheus is deployed |
| `service` | `string` | Kubernetes Service name for Prometheus (format: `{name}-prometheus`) |
| `port_forward_command` | `string` | kubectl port-forward command for local access on port 9090 |
| `kube_endpoint` | `string` | Cluster-internal FQDN (e.g., `my-prometheus-prometheus.monitoring.svc.cluster.local`) |
| `external_hostname` | `string` | Public hostname for external access, only set when ingress is enabled |
| `internal_hostname` | `string` | Internal hostname for VPC-internal access |

## Related Components

- [KubernetesNamespace](/docs/catalog/kubernetes/kubernetesnamespace) — provides the target namespace via `valueFrom` reference
- [KubernetesDeployment](/docs/catalog/kubernetes/kubernetesdeployment) — application deployments that send metrics to Prometheus
- [KubernetesIngressNginx](/docs/catalog/kubernetes/kubernetesingressnginx) — ingress controller for routing external traffic to the Prometheus web UI
- [KubernetesExternalDns](/docs/catalog/kubernetes/kubernetesexternaldns) — manages DNS records for the ingress hostname
