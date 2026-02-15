---
title: "Postgres"
description: "Postgres deployment documentation"
icon: "package"
order: 100
componentName: "kubernetespostgres"
---

# Kubernetes Postgres

Deploys a production-grade PostgreSQL cluster on Kubernetes using the Zalando Postgres Operator. Supports custom databases and users, resource tuning, persistent storage, external access via ingress, and disaster recovery with backup and restore from S3/R2-compatible storage.

## What Gets Created

When you deploy a KubernetesPostgres resource, OpenMCF provisions:

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
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesPostgres
metadata:
  name: my-postgres
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.KubernetesPostgres.my-postgres
spec:
  namespace:
    value: postgres-dev
  createNamespace: true
```

Deploy:

```shell
openmcf apply -f postgres.yaml
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
| `backupConfig.s3Prefix` | `string` | — | Custom S3/R2 prefix path for backups. |
| `backupConfig.backupSchedule` | `string` | — | Cron expression for backup schedule (e.g., `"0 3 * * *"`). |
| `backupConfig.enableBackup` | `bool` | — | Explicitly enable or disable backups for this database. |
| `backupConfig.restore.enabled` | `bool` | `false` | Enable restore mode to bootstrap from a backup. Set to `false` after restore to promote to primary. |
| `backupConfig.restore.bucketName` | `string` | — | S3/R2 bucket name for the backup source. |
| `backupConfig.restore.s3Path` | `string` | — | Path to backup directory within the bucket. |
| `backupConfig.restore.r2Config` | `object` | — | R2-specific credentials (`cloudflareAccountId`, `accessKeyId`, `secretAccessKey`) for cross-cluster restore. |

## Examples

### PostgreSQL with Custom Databases and Users

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesPostgres
metadata:
  name: app-postgres
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.KubernetesPostgres.app-postgres
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
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesPostgres
metadata:
  name: prod-postgres
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.KubernetesPostgres.prod-postgres
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
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesPostgres
metadata:
  name: full-postgres
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.KubernetesPostgres.full-postgres
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
    enableBackup: true
    backupSchedule: "0 3 * * *"
    s3Prefix: "backups/full-postgres/$(PGVERSION)"
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

- [KubernetesNamespace](/docs/catalog/kubernetes/namespace) — pre-create a namespace to reference via `valueFrom`
- [KubernetesRedis](/docs/catalog/kubernetes/redis) — deploy Redis alongside PostgreSQL for caching
- [KubernetesIngress](/docs/catalog/kubernetes/ingress-nginx) — ingress controller for HTTP traffic (PostgreSQL uses LoadBalancer directly)
