# Kubernetes Postgres

Deploys a production-grade PostgreSQL cluster on Kubernetes using the Zalando Postgres Operator. Supports custom databases and users, resource tuning, persistent storage, external access via ingress, and disaster recovery with backup and restore from S3/R2-compatible storage.

## What Gets Created

When you deploy a KubernetesPostgres resource, Planton provisions:

- **Kubernetes Namespace** — created if `createNamespace` is `true`
- **Zalando Postgresql Custom Resource** — a `postgresql` CR managed by the Zalando operator, which in turn creates:
  - StatefulSet with the configured number of replicas
  - Persistent Volume Claims sized to `container.diskSize`
  - Kubernetes Service for cluster-internal access (port 5432)
  - PostgreSQL users and databases as specified
- **LoadBalancer Service** — created only when ingress is enabled, with external-dns annotations for automatic DNS record creation

## Prerequisites

- **A Kubernetes cluster** with the Zalando Postgres Operator installed
- **kubectl** configured to access the target cluster
- **Storage class** available in the cluster for persistent volumes
- **external-dns** running in the cluster (only if using ingress with hostname)

## Quick Start

Create a file `postgres.yaml`:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesPostgres
metadata:
  name: my-postgres
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.KubernetesPostgres.my-postgres
spec:
  namespace:
    value: postgres-dev
  createNamespace: true
```

Deploy:

```shell
planton apply -f postgres.yaml
```

This creates a single-replica PostgreSQL instance with 1Gi disk, 1 CPU, and 1Gi memory in the `postgres-dev` namespace.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `namespace` | `StringValueOrRef` | Kubernetes namespace for the PostgreSQL deployment. Use `value` for a direct string or `valueFrom` to reference a KubernetesNamespace resource. | Required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `createNamespace` | `bool` | `false` | Create the namespace if it does not exist. |
| `container.replicas` | `int` | `1` | Number of PostgreSQL pod replicas. Set to 2+ for high availability with streaming replication. |
| `container.resources.limits.cpu` | `string` | `"1000m"` | CPU limit per pod. |
| `container.resources.limits.memory` | `string` | `"1Gi"` | Memory limit per pod. |
| `container.resources.requests.cpu` | `string` | `"50m"` | CPU request per pod. |
| `container.resources.requests.memory` | `string` | `"100Mi"` | Memory request per pod. |
| `container.diskSize` | `string` | `"1Gi"` | Persistent volume size per replica. Must match pattern like `1Gi`, `10Gi`, `500Mi`. |
| `databases` | `Database[]` | `[]` | Databases to create. Each has a `name` (required) and optional `ownerRole` referencing a user declared in `users`. |
| `users` | `User[]` | `[]` | PostgreSQL users/roles to create. Each has a `name` (required) and optional `flags` (e.g., `["createdb"]`, `["superuser"]`). |
| `ingress.enabled` | `bool` | `false` | Expose PostgreSQL externally via a LoadBalancer service with external-dns. |
| `ingress.hostname` | `string` | — | Hostname for external access (e.g., `postgres.example.com`). Required when `ingress.enabled` is `true`. |
| `backupConfig.enabled` | `bool` | — | Enable continuous WAL-G backups for this database. |
| `backupConfig.bucket` | `StringValueOrRef` | — | Bucket that stores backups. A literal name, or a `valueFrom` reference to any S3-compatible bucket resource's output (e.g. a `CloudflareR2Bucket`'s `status.outputs.bucket_name`). Set the referenced `kind` explicitly. |
| `backupConfig.objectPrefix` | `string` | — | Base path under the bucket (e.g. the environment). The module appends the per-cluster/per-version suffix automatically. |
| `backupConfig.schedule` | `string` | — | Cron expression for base backups (e.g., `"0 2 * * *"`). |
| `backupConfig.retainCount` | `int32` | — | Number of base backups to retain before the oldest is pruned. |
| `backupConfig.credentials` | `object` | — | R2 credentials (`cloudflareAccountId`, `accessKeyId`, `secretAccessKey`) WAL-G uses to write backups. |
| `backupConfig.restore.enabled` | `bool` | `false` | Enable restore mode to bootstrap from a backup. Set to `false` after restore to promote to primary. |
| `backupConfig.restore.bucket` | `StringValueOrRef` | — | Bucket holding the source backup. A literal name or a `valueFrom` reference (set `kind` explicitly). |
| `backupConfig.restore.objectPrefix` | `string` | — | Path under the bucket locating the source backup. |
| `backupConfig.restore.credentials` | `object` | — | R2 credentials (`cloudflareAccountId`, `accessKeyId`, `secretAccessKey`) used to read the source backup. |

## Examples

### PostgreSQL with Custom Databases and Users

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesPostgres
metadata:
  name: app-postgres
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.KubernetesPostgres.app-postgres
spec:
  namespace:
    value: app-backend
  createNamespace: true
  users:
    - name: app_user
      flags: []
    - name: analytics_role
      flags:
        - createdb
  databases:
    - name: app_database
      ownerRole: app_user
    - name: analytics_db
      ownerRole: analytics_role
```

### Production HA with Resource Tuning

A multi-replica setup with tuned resources and larger storage:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesPostgres
metadata:
  name: prod-postgres
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.KubernetesPostgres.prod-postgres
spec:
  namespace:
    value: production
  createNamespace: true
  container:
    replicas: 3
    resources:
      limits:
        cpu: "4000m"
        memory: "8Gi"
      requests:
        cpu: "1000m"
        memory: "2Gi"
    diskSize: "50Gi"
  users:
    - name: app_user
    - name: migration_user
      flags:
        - createdb
  databases:
    - name: app_production
      ownerRole: app_user
```

### Full-Featured with Ingress and Backup

External access, automated backups, and disaster recovery configuration:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesPostgres
metadata:
  name: full-postgres
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.KubernetesPostgres.full-postgres
spec:
  namespace:
    value: databases
  createNamespace: true
  container:
    replicas: 2
    resources:
      limits:
        cpu: "2000m"
        memory: "4Gi"
      requests:
        cpu: "500m"
        memory: "1Gi"
    diskSize: "20Gi"
  ingress:
    enabled: true
    hostname: postgres.example.com
  users:
    - name: app_user
  databases:
    - name: app_db
      ownerRole: app_user
  backupConfig:
    enabled: true
    schedule: "0 3 * * *"
    retainCount: 14
    bucket:
      valueFrom:
        kind: CloudflareR2Bucket
        name: full-postgres-backups
        fieldPath: status.outputs.bucket_name
    objectPrefix: production
    credentials:
      cloudflareAccountId: a1b2c3d4
      accessKeyId: r2-access-key-id
      secretAccessKey: $secret/postgres-r2-secret-access-key
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `namespace` | `string` | Kubernetes namespace where PostgreSQL was created |
| `service` | `string` | Name of the Kubernetes service for PostgreSQL |
| `port_forward_command` | `string` | Ready-to-run `kubectl port-forward` command for local access on port 5432 |
| `kube_endpoint` | `string` | Cluster-internal endpoint (e.g., `teamid-my-postgres.namespace.svc.cluster.local:5432`) |
| `external_hostname` | `string` | External hostname when ingress is enabled (port 5432) |
| `username_secret` | `KubernetesSecretKey` | Reference to the Kubernetes secret containing the PostgreSQL username |
| `password_secret` | `KubernetesSecretKey` | Reference to the Kubernetes secret containing the PostgreSQL password |

## Related Components

- [KubernetesNamespace](/docs/catalog/kubernetes/kubernetesnamespace) — pre-create a namespace to reference via `valueFrom`
- [KubernetesRedis](/docs/catalog/kubernetes/kubernetesredis) — deploy Redis alongside PostgreSQL for caching
- [KubernetesIngress](/docs/catalog/kubernetes/kubernetesingressnginx) — ingress controller for HTTP traffic (PostgreSQL uses LoadBalancer directly)
