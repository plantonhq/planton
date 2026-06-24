# Kubernetes Zalando Postgres Operator

Deploys the Zalando Postgres Operator on a Kubernetes cluster using its official Helm chart (v1.12.2). The operator installs the control-plane components that watch for `postgresql` custom resources, enabling declarative PostgreSQL cluster lifecycle management including automated patroni-based failover, rolling updates, and optional WAL-G backups to Cloudflare R2-compatible object storage.

## What Gets Created

When you deploy a KubernetesZalandoPostgresOperator resource, OpenMCF provisions:

- **Namespace** — created only when `createNamespace` is `true`
- **Helm Release** — installs the `postgres-operator` Helm chart (v1.12.2) from the Zalando Helm repository, deploying the operator pod with configurable CPU and memory resources
- **Operator Pod** — runs the Zalando Postgres Operator controller that reconciles `postgresql` custom resources into running PostgreSQL clusters managed by Patroni
- **CRDs and RBAC** — Custom Resource Definitions for `postgresql` and associated ClusterRoles, ServiceAccounts, and bindings installed by the Helm chart
- **Label Inheritance** — configures the operator to propagate OpenMCF resource labels (organization, environment, resource kind, resource ID) to all managed PostgreSQL clusters
- **Backup Secret** (optional) — a Kubernetes Secret containing Cloudflare R2 credentials for WAL-G backups, created when `backupConfig` is provided
- **Backup ConfigMap** (optional) — a Kubernetes ConfigMap with WAL-G environment variables (S3 endpoint, prefix, schedule) referenced by the operator's `pod_environment_configmap` setting

## Prerequisites

- **Kubernetes credentials** configured via environment variables or OpenMCF provider config
- **A Kubernetes namespace** that already exists, or set `createNamespace` to `true`
- **Helm-capable cluster** — the cluster must support Helm chart installations (standard for all managed Kubernetes offerings)
- **Cloudflare R2 credentials** (optional) — required only if configuring WAL-G backups via `backupConfig`

## Quick Start

Create a file `zalando-postgres-operator.yaml`:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesZalandoPostgresOperator
metadata:
  name: my-pg-operator
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.KubernetesZalandoPostgresOperator.my-pg-operator
spec:
  namespace: postgres-operator
  createNamespace: true
```

Deploy:

```shell
openmcf apply -f zalando-postgres-operator.yaml
```

This installs the Zalando Postgres Operator into the `postgres-operator` namespace with default resource limits (1000m CPU / 1Gi memory) and requests (50m CPU / 100Mi memory).

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
| `container.resources.requests.cpu` | `string` | `50m` | Minimum guaranteed CPU for the operator pod. |
| `container.resources.requests.memory` | `string` | `100Mi` | Minimum guaranteed memory for the operator pod. |
| `backupConfig` | `object` | — | WAL-G backup configuration for all PostgreSQL databases managed by this operator. When provided, creates a Secret and ConfigMap for R2-based backups. |
| `backupConfig.bucket` | `StringValueOrRef` | — | Bucket that stores backups for every database on the cluster. A literal name, or a `valueFrom` reference to any S3-compatible bucket resource's output (e.g. a `CloudflareR2Bucket`'s `status.outputs.bucket_name`). Set the referenced `kind` explicitly. Required when `backupConfig` is specified. |
| `backupConfig.objectPrefix` | `string` | — | Base path under the bucket. The module appends the per-cluster/per-version suffix automatically. |
| `backupConfig.schedule` | `string` | — | Cron schedule for base backups (e.g., `0 2 * * *` for 2 AM daily). Required when `backupConfig` is specified. |
| `backupConfig.credentials.cloudflareAccountId` | `string` | — | Cloudflare account ID used to construct the R2 endpoint URL (`https://<accountId>.r2.cloudflarestorage.com`). Required when `backupConfig` is specified. |
| `backupConfig.credentials.accessKeyId` | `string` | — | R2 access key ID. Stored in a Kubernetes Secret. Required when `backupConfig` is specified. |
| `backupConfig.credentials.secretAccessKey` | `string` | — | R2 secret access key. Stored in a Kubernetes Secret. Required when `backupConfig` is specified. |
| `backupConfig.enableWalGBackup` | `bool` | `true` | Enable WAL-G for continuous archiving backups. |
| `backupConfig.enableWalGRestore` | `bool` | `true` | Enable WAL-G for point-in-time restore operations. |
| `backupConfig.enableCloneWalGRestore` | `bool` | `true` | Enable WAL-G for clone-from-backup operations. |

> **Note on `namespace`:** The `namespace` field supports `valueFrom` for referencing an OpenMCF-managed KubernetesNamespace resource. When using `valueFrom`, the namespace name is resolved at deploy time from the referenced resource's `spec.name` field.

## Examples

### Default Operator Installation

Install the Zalando Postgres Operator with default resource allocations, creating the target namespace automatically:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesZalandoPostgresOperator
metadata:
  name: pg-operator
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.KubernetesZalandoPostgresOperator.pg-operator
spec:
  namespace: postgres-operator
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

### Production Operator with Higher Resources and R2 Backups

For production clusters, increase the operator's resources and enable WAL-G backups to Cloudflare R2 with a nightly schedule:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesZalandoPostgresOperator
metadata:
  name: prod-pg-operator
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.KubernetesZalandoPostgresOperator.prod-pg-operator
spec:
  namespace: postgres-operator
  container:
    resources:
      limits:
        cpu: "2000m"
        memory: "2Gi"
      requests:
        cpu: "200m"
        memory: "256Mi"
  backupConfig:
    bucket:
      value: "prod-pg-backups"
    objectPrefix: production
    schedule: "0 2 * * *"
    enableWalGBackup: true
    enableWalGRestore: true
    enableCloneWalGRestore: true
    credentials:
      cloudflareAccountId: "abc123def456"
      accessKeyId: "R2_ACCESS_KEY"
      secretAccessKey: "R2_SECRET_KEY"
```

### Operator with Foreign Key Namespace Reference

Reference an OpenMCF-managed namespace instead of hardcoding the name. The `valueFrom` syntax resolves the namespace name from a KubernetesNamespace resource at deploy time:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesZalandoPostgresOperator
metadata:
  name: shared-pg-operator
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: staging.KubernetesZalandoPostgresOperator.shared-pg-operator
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
        cpu: "100m"
        memory: "128Mi"
```

### Backups with a Custom Object Prefix

Organize backup paths by team or environment using a custom object prefix; the module appends the per-cluster/per-version segments automatically:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesZalandoPostgresOperator
metadata:
  name: team-pg-operator
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.KubernetesZalandoPostgresOperator.team-pg-operator
spec:
  namespace: postgres-operator
  createNamespace: true
  container:
    resources:
      limits:
        cpu: "1000m"
        memory: "1Gi"
      requests:
        cpu: "50m"
        memory: "100Mi"
  backupConfig:
    bucket:
      valueFrom:
        kind: CloudflareR2Bucket
        name: shared-pg-backups
        fieldPath: status.outputs.bucket_name
    objectPrefix: team-alpha
    schedule: "0 3 * * 0"
    enableWalGBackup: true
    enableWalGRestore: true
    credentials:
      cloudflareAccountId: "abc123def456"
      accessKeyId: "R2_ACCESS_KEY"
      secretAccessKey: "R2_SECRET_KEY"
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `namespace` | `string` | Kubernetes namespace where the Zalando Postgres Operator is installed |
| `service` | `string` | Kubernetes service name for the operator |
| `portForwardCommand` | `string` | Command to set up port-forwarding to the operator from a developer laptop |
| `kubeEndpoint` | `string` | Cluster-internal endpoint for the operator service |
| `ingressEndpoint` | `string` | Public endpoint for the operator when ingress is configured |

## Related Components

- [KubernetesNamespace](/docs/catalog/kubernetes/kubernetesnamespace) — provides the target namespace via `valueFrom` reference
- [KubernetesPostgres](/docs/catalog/kubernetes/kubernetespostgres) — deploys PostgreSQL instances that can be managed by the Zalando Operator installed by this component
