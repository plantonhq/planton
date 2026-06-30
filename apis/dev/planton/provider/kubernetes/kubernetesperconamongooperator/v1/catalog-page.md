# Kubernetes Percona Mongo Operator

Deploys the Percona Operator for MongoDB on a Kubernetes cluster using its official Helm chart. The operator runs in cluster-wide mode, watching all namespaces for PerconaServerMongoDB custom resources, enabling declarative MongoDB lifecycle management across the cluster.

## What Gets Created

When you deploy a KubernetesPerconaMongoOperator resource, Planton provisions:

- **Namespace** — created only when `createNamespace` is `true`
- **Helm Release** — installs the `psmdb-operator` Helm chart (v1.20.1) from the Percona Helm repository, deploying the operator pod with configurable CPU and memory resources
- **Operator Pod** — runs in cluster-wide mode (`watchAllNamespaces: true`), monitoring all namespaces for PerconaServerMongoDB custom resources
- **CRDs and RBAC** — Custom Resource Definitions for PerconaServerMongoDB and associated ClusterRoles, ServiceAccounts, and bindings installed by the Helm chart

## Prerequisites

- **Kubernetes credentials** configured via environment variables or Planton provider config
- **A Kubernetes namespace** that already exists, or set `createNamespace` to `true`
- **Helm-capable cluster** — the cluster must support Helm chart installations (standard for all managed Kubernetes offerings)

## Quick Start

Create a file `percona-mongo-operator.yaml`:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesPerconaMongoOperator
metadata:
  name: my-percona-operator
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.KubernetesPerconaMongoOperator.my-percona-operator
spec:
  namespace: percona-system
  createNamespace: true
```

Deploy:

```shell
planton apply -f percona-mongo-operator.yaml
```

This installs the Percona MongoDB Operator into the `percona-system` namespace with default resource limits (1000m CPU / 1Gi memory) and requests (100m CPU / 256Mi memory), watching all namespaces for PerconaServerMongoDB custom resources.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `namespace` | `string` | Kubernetes namespace where the operator is installed. Can reference a KubernetesNamespace resource via `valueFrom`. | Required |
| `container` | `object` | Container specification for the operator pod, including resource limits and requests. | Required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `targetCluster.clusterKind` | `enum` | — | Kubernetes cluster kind. Valid values: `AwsEksCluster`, `GcpGkeCluster`, `AzureAksCluster`, `DigitalOceanKubernetesCluster`, `CivoKubernetesCluster`. |
| `targetCluster.clusterName` | `string` | — | Name of the target Kubernetes cluster in the same environment. |
| `createNamespace` | `bool` | `false` | When `true`, creates the namespace before deploying the operator. |
| `container.resources.limits.cpu` | `string` | `1000m` | Maximum CPU allocation for the operator pod. |
| `container.resources.limits.memory` | `string` | `1Gi` | Maximum memory allocation for the operator pod. |
| `container.resources.requests.cpu` | `string` | `100m` | Minimum guaranteed CPU for the operator pod. |
| `container.resources.requests.memory` | `string` | `256Mi` | Minimum guaranteed memory for the operator pod. |

## Examples

### Default Operator Installation

Install the Percona MongoDB Operator with default resource allocations, creating the target namespace automatically:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesPerconaMongoOperator
metadata:
  name: percona-operator
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.KubernetesPerconaMongoOperator.percona-operator
spec:
  namespace: percona-system
  createNamespace: true
  container:
    resources:
      limits:
        cpu: "1000m"
        memory: "1Gi"
      requests:
        cpu: "100m"
        memory: "256Mi"
```

### Production Operator with Higher Resource Limits

For production clusters managing many MongoDB instances, increase the operator's resource allocation to handle the additional reconciliation workload:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesPerconaMongoOperator
metadata:
  name: prod-percona-operator
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.KubernetesPerconaMongoOperator.prod-percona-operator
spec:
  namespace: percona-system
  container:
    resources:
      limits:
        cpu: "2000m"
        memory: "2Gi"
      requests:
        cpu: "500m"
        memory: "512Mi"
```

### Operator with Foreign Key Namespace Reference

Reference an Planton-managed namespace instead of hardcoding the name. The `valueFrom` syntax resolves the namespace name from a KubernetesNamespace resource at deploy time:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesPerconaMongoOperator
metadata:
  name: shared-percona-operator
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: staging.KubernetesPerconaMongoOperator.shared-percona-operator
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: operators-namespace
      field: spec.name
  container:
    resources:
      limits:
        cpu: "1000m"
        memory: "1Gi"
      requests:
        cpu: "200m"
        memory: "256Mi"
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `namespace` | `string` | Kubernetes namespace where the Percona MongoDB Operator is installed |

## Related Components

- [KubernetesNamespace](/docs/catalog/kubernetes/kubernetesnamespace) — provides the target namespace via `valueFrom` reference
- [KubernetesMongodb](/docs/catalog/kubernetes/kubernetesmongodb) — deploys MongoDB instances managed by the Percona Operator installed by this component
