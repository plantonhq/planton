# Kubernetes Zalando Postgres Operator

The **KubernetesZalandoPostgresOperator** is a OpenMCF component that deploys and manages the [Zalando Postgres Operator](https://github.com/zalando/postgres-operator) on Kubernetes clusters. This operator simplifies the deployment, management, and backup of PostgreSQL databases in Kubernetes, providing production-grade database automation.

## Overview

The Zalando Postgres Operator is one of the most mature and widely-used Kubernetes operators for PostgreSQL. It enables declarative database management through custom Kubernetes resources, handles automated backups, supports high availability, and provides point-in-time recovery capabilities.

This OpenMCF component abstracts the complexity of deploying and configuring the Zalando Postgres Operator, providing a simplified interface while maintaining flexibility for production deployments.

## Key Features

### Operator Management
- **Automated Deployment**: One-command deployment of the Zalando Postgres Operator via Helm
- **Namespace Management**: Choose to create a new namespace or use an existing one with the `create_namespace` flag
- **Resource Control**: Configure CPU and memory limits for the operator container
- **Label Inheritance**: Automatically propagates organization, environment, and resource labels to managed databases
- **Production-Ready Defaults**: Sensible default resource allocations (1 CPU, 1Gi memory limits; 50m CPU, 100Mi memory requests)

### Backup & Recovery
- **Cloudflare R2 Integration**: Native support for backing up to Cloudflare R2 (S3-compatible storage)
- **WAL-G Integration**: Automated WAL (Write-Ahead Log) archiving for point-in-time recovery
- **Scheduled Backups**: Cron-based backup scheduling (e.g., daily at 2 AM)
- **Automated Credentials**: Automatically creates Kubernetes Secrets for R2 access keys
- **Configurable Retention**: Control backup schedules and storage paths

### Database Features (via Operator)
- **Declarative Management**: Define PostgreSQL clusters using Kubernetes CRDs
- **High Availability**: Multi-replica PostgreSQL clusters with automated failover
- **Connection Pooling**: Built-in PgBouncer support
- **Monitoring**: Integration with Prometheus and Grafana
- **Backup/Restore**: Automated backups to S3-compatible storage (R2, AWS S3, MinIO)
- **Clone Databases**: Create database clones from backups

## Prerequisites

| Requirement | Version | Purpose |
|------------|---------|---------|
| **Kubernetes** | 1.24+ | Target cluster for operator deployment |
| **Pulumi** or **Terraform** | Latest | Infrastructure-as-Code tooling |
| **kubectl** | Latest | Kubernetes CLI for verification |
| **Cloudflare R2** (optional) | - | S3-compatible backup storage |

> **Note**: Backup configuration is optional. The operator works without backups, but production deployments should configure R2 or another S3-compatible backend.

## Installation

### Using OpenMCF CLI with Pulumi

```bash
# Create a manifest file
cat > postgres-operator.yaml <<EOF
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesZalandoPostgresOperator
metadata:
  name: postgres-operator
spec:
  container:
    resources:
      requests:
        cpu: 50m
        memory: 100Mi
      limits:
        cpu: 1000m
        memory: 1Gi
EOF

# Deploy the operator
openmcf pulumi up --manifest postgres-operator.yaml
```

### Using OpenMCF CLI with Terraform

```bash
# Create a manifest file (same as above)
cat > postgres-operator.yaml <<EOF
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesZalandoPostgresOperator
metadata:
  name: postgres-operator
spec:
  container:
    resources:
      requests:
        cpu: 50m
        memory: 100Mi
      limits:
        cpu: 1000m
        memory: 1Gi
EOF

# Deploy using Terraform backend
openmcf terraform apply --manifest postgres-operator.yaml
```

## Configuration

### Basic Configuration

The minimal configuration requires the target cluster, namespace, and container resource specifications:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesZalandoPostgresOperator
metadata:
  name: postgres-operator
spec:
  target_cluster:
    cluster_name: my-gke-cluster
  namespace:
    value: postgres-operator
  create_namespace: true  # Creates the namespace
  container:
    resources:
      requests:
        cpu: 50m
        memory: 100Mi
      limits:
        cpu: 1000m
        memory: 1Gi
```

### Using Existing Namespace

If you have a pre-existing namespace, set `create_namespace: false`:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesZalandoPostgresOperator
metadata:
  name: postgres-operator
spec:
  target_cluster:
    cluster_name: my-gke-cluster
  namespace:
    value: existing-postgres-namespace
  create_namespace: false  # Use existing namespace
  container:
    resources:
      requests:
        cpu: 50m
        memory: 100Mi
      limits:
        cpu: 1000m
        memory: 1Gi
```

**Important**: When `create_namespace: false`, the namespace must already exist in the target cluster. The deployment will fail if the namespace doesn't exist.

### Production Configuration with Backups

For production deployments, configure automated backups to Cloudflare R2:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesZalandoPostgresOperator
metadata:
  name: postgres-operator
  labels:
    org: my-company
    env: production
spec:
  target_cluster:
    cluster_name: prod-gke-cluster
  namespace:
    value: postgres-operator
  create_namespace: true
  container:
    resources:
      requests:
        cpu: 100m
        memory: 256Mi
      limits:
        cpu: 2000m
        memory: 2Gi
  backup_config:
    bucket:
      value: "postgres-backups-prod"
    object_prefix: production
    schedule: "0 2 * * *"  # 2 AM daily
    enable_wal_g_backup: true
    enable_wal_g_restore: true
    enable_clone_wal_g_restore: true
    credentials:
      cloudflare_account_id: "abc123xyz"
      access_key_id: "${R2_ACCESS_KEY_ID}"
      secret_access_key: "${R2_SECRET_ACCESS_KEY}"
```

### Configuration Fields

#### Metadata

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Unique name for the operator deployment |
| `id` | string | No | Unique identifier (defaults to name) |
| `org` | string | No | Organization label for multi-tenancy |
| `env` | string | No | Environment label (dev, staging, prod) |

#### Spec: Cluster and Namespace

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `target_cluster.cluster_name` | string | Yes | - | Name of the target Kubernetes cluster |
| `namespace.value` | string | Yes | - | Kubernetes namespace for the operator |
| `create_namespace` | bool | Yes | - | Whether to create the namespace (true) or use existing (false) |

#### Spec: Container

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `container.resources.requests.cpu` | string | Yes | `50m` | CPU request for operator pod |
| `container.resources.requests.memory` | string | Yes | `100Mi` | Memory request for operator pod |
| `container.resources.limits.cpu` | string | Yes | `1000m` | CPU limit for operator pod |
| `container.resources.limits.memory` | string | Yes | `1Gi` | Memory limit for operator pod |

#### Spec: Backup Config (Optional)

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `backup_config.bucket` | StringValueOrRef | Yes* | - | Bucket for backups (literal name or `valueFrom` reference to any S3-compatible bucket, e.g. a `CloudflareR2Bucket`) |
| `backup_config.object_prefix` | string | No | - | Base path under the bucket; the module appends the per-cluster/per-version suffix |
| `backup_config.schedule` | string | Yes* | - | Cron schedule (e.g., `0 2 * * *`) |
| `backup_config.credentials.cloudflare_account_id` | string | Yes* | - | Cloudflare account ID for R2 endpoint |
| `backup_config.credentials.access_key_id` | string | Yes* | - | R2 access key ID |
| `backup_config.credentials.secret_access_key` | string | Yes* | - | R2 secret access key |
| `backup_config.enable_wal_g_backup` | bool | No | `true` | Enable WAL-G backups |
| `backup_config.enable_wal_g_restore` | bool | No | `true` | Enable WAL-G restores |
| `backup_config.enable_clone_wal_g_restore` | bool | No | `true` | Enable WAL-G for clone operations |

\* Required when `backup_config` is specified

## How It Works

### Deployment Architecture

1. **Namespace Management**: 
   - If `create_namespace: true` - Creates namespace with OpenMCF labels
   - If `create_namespace: false` - Uses existing namespace (must already exist)
2. **Backup Resources** (if configured):
   - Creates Kubernetes Secret with R2 credentials
   - Creates ConfigMap with WAL-G environment variables
3. **Helm Deployment**: Installs Zalando Postgres Operator via Helm chart
   - Chart: `postgres-operator/postgres-operator`
   - Version: `1.12.2` (configurable)
   - Repository: https://opensource.zalando.com/postgres-operator/charts/postgres-operator
4. **Label Propagation**: Configures operator to inherit labels for all created databases

### Backup Mechanism

When backup configuration is provided:

1. **Secret Creation**: R2 credentials are stored in `r2-postgres-backup-credentials` Secret
2. **ConfigMap Creation**: WAL-G configuration stored in `postgres-pod-backup-config` ConfigMap
3. **Operator Integration**: Zalando operator references the ConfigMap via `pod_environment_configmap`
4. **Database Pods**: All PostgreSQL pods inherit the backup configuration automatically
5. **WAL-G Process**: WAL-G handles continuous WAL archiving and scheduled base backups

### What Gets Created

After deployment, the following resources exist in the cluster:

- **Namespace**: Specified namespace (created only if `create_namespace: true`)
- **Deployment**: `postgres-operator` (the operator controller)
- **Service**: `postgres-operator` (webhook and metrics endpoints)
- **ConfigMaps**: Operator configuration + backup config (if enabled)
- **Secrets**: R2 credentials (if backup configured)
- **CRDs**: PostgreSQL cluster custom resource definitions

## Using the Operator

Once the operator is deployed, you can create PostgreSQL databases using the `postgresql` CRD:

```yaml
apiVersion: "acid.zalan.do/v1"
kind: postgresql
metadata:
  name: my-database
  namespace: my-app
spec:
  teamId: "my-team"
  volume:
    size: 10Gi
  numberOfInstances: 2
  users:
    myapp:
      - superuser
      - createdb
  databases:
    myapp: myapp
  postgresql:
    version: "15"
    parameters:
      max_connections: "100"
  resources:
    requests:
      cpu: 500m
      memory: 1Gi
    limits:
      cpu: 2
      memory: 2Gi
```

The operator will automatically:
- Create PostgreSQL StatefulSet with 2 replicas
- Set up streaming replication
- Configure PgBouncer connection pooling
- Apply backup configuration (if configured at operator level)
- Create service endpoints for master and replica access

## Outputs

After deployment, the following outputs are available:

| Output | Description | Example |
|--------|-------------|---------|
| `namespace` | Operator namespace | `postgres-operator` |
| `service` | Operator service name | `postgres-operator` |
| `port_forward_command` | kubectl port-forward command | `kubectl port-forward svc/postgres-operator -n postgres-operator 8080:8080` |
| `kube_endpoint` | Internal cluster endpoint | `postgres-operator.postgres-operator.svc.cluster.local` |

## Examples

See [examples.md](./examples.md) for complete configuration examples including:
- Basic operator deployment
- Production deployment with R2 backups
- Custom resource limits
- Database creation examples

## Best Practices

### Resource Sizing

**Development/Testing:**
```yaml
container:
  resources:
    requests:
      cpu: 50m
      memory: 100Mi
    limits:
      cpu: 500m
      memory: 512Mi
```

**Production:**
```yaml
container:
  resources:
    requests:
      cpu: 100m
      memory: 256Mi
    limits:
      cpu: 2000m
      memory: 2Gi
```

### Backup Strategy

1. **Enable Backups**: Always configure backups for production databases
2. **Schedule Wisely**: Run backups during low-traffic periods (e.g., 2-4 AM)
3. **Test Restores**: Periodically verify backup restore procedures
4. **Monitor Storage**: Ensure R2 bucket has sufficient space and retention policies

### Security

1. **Secret Management**: Use external secret managers (e.g., Vault, AWS Secrets Manager) for R2 credentials
2. **Network Policies**: Restrict operator access using Kubernetes NetworkPolicies
3. **RBAC**: The operator requires cluster-wide permissions; review the Helm chart's RBAC settings
4. **Encryption**: Enable encryption at rest in R2 bucket settings

### High Availability

1. **Operator HA**: The operator itself is stateless; consider running multiple replicas (via Helm values)
2. **Database HA**: Use `numberOfInstances: 2+` in PostgreSQL CRDs for database high availability
3. **Backup Redundancy**: Consider multi-region R2 buckets or replication to multiple S3 backends

## Troubleshooting

### Operator Not Starting

```bash
# Check operator logs
kubectl logs -n postgres-operator deployment/postgres-operator

# Check events
kubectl get events -n postgres-operator --sort-by='.lastTimestamp'
```

### Backup Issues

```bash
# Verify backup ConfigMap
kubectl get configmap -n postgres-operator postgres-pod-backup-config -o yaml

# Verify R2 credentials Secret
kubectl get secret -n postgres-operator r2-postgres-backup-credentials

# Check database pod environment
kubectl exec -n <db-namespace> <pod-name> -- env | grep WAL
```

### Database Creation Fails

```bash
# Check operator logs for why it's not processing the postgresql CR
kubectl logs -n postgres-operator deployment/postgres-operator | grep <database-name>

# Verify CRD is installed
kubectl get crd postgresqls.acid.zalan.do

# Check postgresql resource status
kubectl describe postgresql <database-name> -n <namespace>
```

## Limitations

1. **Cloudflare R2 Only**: Current backup implementation supports only Cloudflare R2 (S3-compatible storage). AWS S3, Google Cloud Storage, and Azure Blob support can be added in future versions.
2. **Single Operator Per Cluster**: Deploy one operator instance per Kubernetes cluster. Multiple operators in the same cluster may conflict.
3. **Namespace Scope**: The operator is cluster-scoped and can manage PostgreSQL databases in any namespace.
4. **Version Upgrades**: Operator version upgrades should be tested in non-production environments first.

## Related Components

- **KubernetesPostgres**: Individual PostgreSQL database deployment (uses this operator)
- **KubernetesPerconaPostgresOperator**: Alternative Postgres operator with different feature set
- **CertManager**: TLS certificate management for database connections

## References

- [Zalando Postgres Operator Documentation](https://postgres-operator.readthedocs.io/)
- [Zalando Postgres Operator GitHub](https://github.com/zalando/postgres-operator)
- [WAL-G Documentation](https://github.com/wal-g/wal-g)
- [Cloudflare R2 Documentation](https://developers.cloudflare.com/r2/)
- [OpenMCF Documentation](https://docs.openmcf.io)

## Support

For issues, questions, or contributions:
- File an issue in the OpenMCF repository
- Consult the [architecture documentation](./docs/README.md) for design decisions
- Review the [Pulumi implementation](./iac/pulumi/) for customization options

