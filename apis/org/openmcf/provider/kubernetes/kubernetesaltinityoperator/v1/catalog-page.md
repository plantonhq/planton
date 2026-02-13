# Kubernetes Altinity Operator

Deploys the Altinity ClickHouse Operator on Kubernetes using the official `altinity-clickhouse-operator` Helm chart (v0.25.4). The operator installs ClickHouse-related CRDs and watches all namespaces, enabling declarative management of ClickHouse clusters across the entire Kubernetes cluster. Supports configurable CPU and memory resources for the operator pod and optional namespace creation.

## What Gets Created

When you deploy a KubernetesAltinityOperator resource, OpenMCF provisions:

- **Kubernetes Namespace** — created only when `createNamespace` is `true`
- **Helm Release (Altinity ClickHouse Operator)** — deploys the `altinity-clickhouse-operator` chart (v0.25.4) from `https://docs.altinity.com/clickhouse-operator/`, with atomic rollback enabled, cleanup-on-fail, wait-for-jobs, and a 5-minute timeout; installs ClickHouse CRDs automatically (`createCRD: true`)
- **Cluster-Wide Namespace Watch** — the operator is configured to watch all namespaces using the `".*"` regex pattern, so ClickHouseInstallation resources in any namespace are reconciled
- **Standard Labels** — all created resources are labeled with the resource kind, name, organization, and environment for consistent discovery

## Prerequisites

- **Kubernetes credentials** configured via environment variables or OpenMCF provider config
- **A Kubernetes namespace** that already exists, or set `createNamespace` to `true`
- **Helm 3** support in the target cluster (provided by the Pulumi Kubernetes provider)

## Quick Start

Create a file `altinity-operator.yaml`:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesAltinityOperator
metadata:
  name: my-altinity-operator
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.KubernetesAltinityOperator.my-altinity-operator
spec:
  namespace:
    value: altinity-operator
  createNamespace: true
  container: {}
```

Deploy:

```shell
openmcf apply -f altinity-operator.yaml
```

This creates the Altinity ClickHouse Operator in the `altinity-operator` namespace with default resources (1000m CPU / 1Gi memory limit, 100m CPU / 256Mi memory request). The operator watches all namespaces for ClickHouseInstallation custom resources.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `namespace` | `StringValueOrRef` | Kubernetes namespace for the operator deployment. Use `value` for a direct string or `valueFrom` to reference a KubernetesNamespace resource. | Required |
| `container` | `object` | Container specification for the operator pod. Pass `{}` to accept all defaults. | Required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `targetCluster.clusterKind` | `enum` | — | Kubernetes cluster kind. Valid values: `AwsEksCluster`, `GcpGkeCluster`, `AzureAksCluster`, `DigitalOceanKubernetesCluster`, `CivoKubernetesCluster`. |
| `targetCluster.clusterName` | `string` | — | Name of the target Kubernetes cluster in the same environment. |
| `createNamespace` | `bool` | `false` | When `true`, creates the namespace before deploying resources. |
| `container.resources.limits.cpu` | `string` | `"1000m"` | Maximum CPU allocation for the operator pod. |
| `container.resources.limits.memory` | `string` | `"1Gi"` | Maximum memory allocation for the operator pod. |
| `container.resources.requests.cpu` | `string` | `"100m"` | Minimum guaranteed CPU for the operator pod. |
| `container.resources.requests.memory` | `string` | `"256Mi"` | Minimum guaranteed memory for the operator pod. |

## Examples

### Development Instance with Default Resources

A minimal operator deployment for development or testing:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesAltinityOperator
metadata:
  name: dev-altinity
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.KubernetesAltinityOperator.dev-altinity
spec:
  namespace:
    value: altinity-dev
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

### Production Instance with Higher Resources

An operator deployment sized for production workloads managing multiple ClickHouse clusters:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesAltinityOperator
metadata:
  name: prod-altinity
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.KubernetesAltinityOperator.prod-altinity
spec:
  namespace:
    value: altinity-system
  createNamespace: true
  container:
    resources:
      limits:
        cpu: "2000m"
        memory: "2Gi"
      requests:
        cpu: "250m"
        memory: "512Mi"
```

### Using Foreign Key References

Reference an OpenMCF-managed namespace instead of hardcoding the name:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesAltinityOperator
metadata:
  name: platform-altinity
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.KubernetesAltinityOperator.platform-altinity
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: platform-operators
      field: spec.name
  container:
    resources:
      limits:
        cpu: "1000m"
        memory: "1Gi"
      requests:
        cpu: "100m"
        memory: "256Mi"
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `namespace` | `string` | Kubernetes namespace where the Altinity ClickHouse Operator is installed |

## Related Components

- [KubernetesNamespace](/docs/catalog/kubernetes/kubernetesnamespace) — provides the target namespace via `valueFrom` reference
- [KubernetesClickhouse](/docs/catalog/kubernetes/kubernetesclickhouse) — deploy ClickHouse clusters managed by the Altinity operator
- [KubernetesHelmRelease](/docs/catalog/kubernetes/kuberneteshelmrelease) — alternative for deploying Helm charts directly when a dedicated component is not available
