# Kubernetes Percona Postgres Operator

Deploys the Percona Operator for PostgreSQL on a Kubernetes cluster using its official Helm chart. The operator installs the control-plane components that watch for PerconaPGCluster custom resources, enabling declarative PostgreSQL cluster lifecycle management including automated failover, backups, and connection pooling.

## What Gets Created

When you deploy a KubernetesPerconaPostgresOperator resource, OpenMCF provisions:

- **Namespace** — created only when `createNamespace` is `true`
- **Helm Release** — installs the `pg-operator` Helm chart (v2.7.0) from the Percona Helm repository, deploying the operator pod with configurable CPU and memory resources
- **Operator Pod** — runs the Percona PG Operator controller that reconciles PerconaPGCluster custom resources into running PostgreSQL clusters
- **CRDs and RBAC** — Custom Resource Definitions for PerconaPGCluster and associated ClusterRoles, ServiceAccounts, and bindings installed by the Helm chart

## Prerequisites

- **Kubernetes credentials** configured via environment variables or OpenMCF provider config
- **A Kubernetes namespace** that already exists, or set `createNamespace` to `true`
- **Helm-capable cluster** — the cluster must support Helm chart installations (standard for all managed Kubernetes offerings)

## Quick Start

Create a file `percona-postgres-operator.yaml`:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesPerconaPostgresOperator
metadata:
  name: my-pg-operator
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.KubernetesPerconaPostgresOperator.my-pg-operator
spec:
  namespace: percona-system
  createNamespace: true
```

Deploy:

```shell
openmcf apply -f percona-postgres-operator.yaml
```

This installs the Percona PostgreSQL Operator into the `percona-system` namespace with default resource limits (1000m CPU / 1Gi memory) and requests (100m CPU / 256Mi memory).

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

Install the Percona PostgreSQL Operator with default resource allocations, creating the target namespace automatically:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesPerconaPostgresOperator
metadata:
  name: pg-operator
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.KubernetesPerconaPostgresOperator.pg-operator
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

For production clusters managing many PostgreSQL instances, increase the operator's resource allocation to handle the additional reconciliation workload:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesPerconaPostgresOperator
metadata:
  name: prod-pg-operator
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.KubernetesPerconaPostgresOperator.prod-pg-operator
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

Reference an OpenMCF-managed namespace instead of hardcoding the name. The `valueFrom` syntax resolves the namespace name from a KubernetesNamespace resource at deploy time:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesPerconaPostgresOperator
metadata:
  name: shared-pg-operator
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: staging.KubernetesPerconaPostgresOperator.shared-pg-operator
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
| `namespace` | `string` | Kubernetes namespace where the Percona PostgreSQL Operator is installed |

## Related Components

- [KubernetesNamespace](/docs/catalog/kubernetes/kubernetesnamespace) — provides the target namespace via `valueFrom` reference
- [KubernetesPostgres](/docs/catalog/kubernetes/kubernetespostgres) — deploys PostgreSQL instances that can be managed by the Percona Operator installed by this component
