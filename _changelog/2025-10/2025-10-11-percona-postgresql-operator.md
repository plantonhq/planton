# Percona PostgreSQL Operator Support

**Date**: October 11, 2025  
**Type**: Feature  
**Component**: PerconaPostgresqlOperator

## Summary

Added support for deploying and managing the Percona Distribution for PostgreSQL Operator on Kubernetes clusters. This operator enables automated deployment, management, and operations of production-ready PostgreSQL databases with built-in backup, monitoring, high availability, and disaster recovery features.

## Motivation

Organizations running PostgreSQL on Kubernetes needed enterprise-grade database management capabilities including:
- Automated PostgreSQL cluster deployment and scaling
- Built-in backup and point-in-time recovery
- High availability with streaming replication and automatic failover
- Monitoring and observability integration
- Security hardening and encryption
- Simplified database operations and upgrades
- Production-tested reliability for mission-critical workloads

The Percona Distribution for PostgreSQL Operator (based on Crunchy Data's PGO) provides these capabilities through Kubernetes-native declarative configuration, reducing operational complexity and improving database reliability.

## What's New

### 1. PerconaPostgresqlOperator API Resource

New Kubernetes cloud resource kind for deploying the Percona PostgreSQL operator:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: PerconaPostgresqlOperator
metadata:
  name: percona-postgresql-operator
spec:
  targetCluster:
    kubernetesProviderConfigId: k8s-cluster-01
  namespace: percona-postgresql-operator
  container:
    resources:
      requests:
        cpu: 100m
        memory: 256Mi
      limits:
        cpu: 1000m
        memory: 1Gi
```

**Features**:
- Declarative operator deployment via manifest
- Configurable resource allocation
- Target cluster selection via credential ID
- Namespace isolation
- Helm chart-based installation

### 2. CloudResourceKind Registration

Added `PerconaPostgresqlOperator` to the cloud resource kind enum:

```protobuf
PerconaPostgresqlOperator = 838 [(kind_meta) = {
  provider: kubernetes, 
  version: v1, 
  id_prefix: "percpgop", 
  kubernetes_meta: {category: addon}
}];
```

**Properties**:
- Enum value: 838
- ID prefix: `percpgop` (Percona PostgreSQL Operator)
- Provider: Kubernetes
- Category: Addon (operator-level infrastructure)

### 3. Operator Capabilities

The Percona Distribution for PostgreSQL Operator manages:

**PostgreSQL Cluster Types**:
- High Availability clusters with streaming replication
- Standby clusters for disaster recovery
- Multiple replicas with automatic failover
- Connection pooling with PgBouncer

**Operational Features**:
- Automated backups to S3-compatible storage
- Point-in-Time Recovery (PITR)
- Scheduled and on-demand backups
- Automated failover with Patroni
- Rolling updates with zero downtime
- TLS/SSL encryption
- RBAC and authentication management
- Connection pooling for high-performance applications

**Monitoring Integration**:
- Prometheus metrics export
- PostgreSQL Exporter included
- Custom metric collection
- Query performance monitoring

## Implementation Details

### Pulumi Module

**Location**: `apis/project/planton/provider/kubernetes/addon/perconapostgresqloperator/v1/iac/pulumi`

**Key Files**:
- `main.go` - Main Pulumi program
- `module/percona_operator.go` - Helm release and namespace creation
- `module/outputs.go` - Stack outputs
- `module/vars.go` - Configuration variables

**Helm Chart**:
- Chart: `percona/pg-operator`
- Repository: `https://percona.github.io/percona-helm-charts`
- Version: 2.7.0 (configurable)
- Deployment: Single operator pod per cluster

### Terraform Module

**Location**: `apis/project/planton/provider/kubernetes/addon/perconapostgresqloperator/v1/iac/tf`

**Resources**:
- Kubernetes namespace
- Helm release for operator
- Configurable resource limits
- Output: namespace name

### CRDs Installed

The operator installs three Custom Resource Definitions:

1. **PerconaPGCluster** (`pgv2.percona.com/v1`)
   - Primary CRD for PostgreSQL cluster management
   - Defines cluster topology, replicas, storage
   - Configures backups, monitoring, security, connection pooling

2. **PerconaPGBackup** (`pgv2.percona.com/v1`)
   - On-demand backup resource
   - Triggers immediate backups
   - Specifies backup destination and retention

3. **PerconaPGRestore** (`pgv2.percona.com/v1`)
   - Database restore operations
   - Point-in-time recovery
   - Clone databases from backups

### Proto Definitions

**Complete API Structure**:
- `api.proto` - Main API resource definition with validations
- `spec.proto` - Operator specification with container resources
- `stack_input.proto` - Pulumi stack input structure
- `stack_outputs.proto` - Deployment outputs (namespace)

### Deployment Verification

After successful deployment, verification shows:

```bash
# Operator pod running
$ kubectl get pods -n percona-postgresql-operator
NAME                              READY   STATUS    RESTARTS   AGE
pg-operator-6dd9cbb89c-2jgf9      1/1     Running   0          66s

# Operator logs show successful initialization
$ kubectl logs -n percona-postgresql-operator pg-operator-xxx
INFO    setup    Manager starting up
INFO    server version    {"platform": "kubernetes", "version": "v1.30.14"}
INFO    Starting Controller    {"controller": "perconapgcluster-controller"}
INFO    Starting Controller    {"controller": "perconapgbackup-controller"}
INFO    Starting Controller    {"controller": "perconapgrestore-controller"}

# CRDs installed
$ kubectl get crds | grep percona
perconapgclusters.pgv2.percona.com
perconapgbackups.pgv2.percona.com
perconapgrestores.pgv2.percona.com
```

## Usage Examples

### Deploy Operator to Production Cluster

```bash
# Set local module path
export PERCONA_POSTGRESQL_OPERATOR_MODULE=~/scm/github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/addon/perconapostgresqloperator/v1/iac/pulumi

# Initialize Pulumi stack
planton pulumi init \
  --manifest percona-postgresql-operator.yaml \
  --module-dir ${PERCONA_POSTGRESQL_OPERATOR_MODULE}

# Preview changes
planton pulumi preview \
  --manifest percona-postgresql-operator.yaml \
  --module-dir ${PERCONA_POSTGRESQL_OPERATOR_MODULE}

# Deploy operator
planton pulumi up \
  --manifest percona-postgresql-operator.yaml \
  --module-dir ${PERCONA_POSTGRESQL_OPERATOR_MODULE}
```

### Basic PostgreSQL Cluster (Future)

Once the operator is deployed, PostgreSQL clusters can be created:

```yaml
apiVersion: pgv2.percona.com/v1
kind: PerconaPGCluster
metadata:
  name: production-db
  namespace: percona-postgresql-operator
spec:
  crVersion: 2.7.0
  image: percona/percona-postgresql-operator:2.7.0-ppg17-postgres
  postgresVersion: 17
  instances:
  - name: instance1
    replicas: 3
    dataVolumeClaimSpec:
      accessModes:
      - ReadWriteOnce
      resources:
        requests:
          storage: 100Gi
  backups:
    pgbackrest:
      repos:
      - name: repo1
        schedules:
          full: "0 2 * * 0"
          incremental: "0 2 * * 1-6"
        volume:
          volumeClaimSpec:
            accessModes:
            - ReadWriteOnce
            resources:
              requests:
                storage: 200Gi
  proxy:
    pgBouncer:
      replicas: 2
```

## Architecture

### Component Interaction

```
┌─────────────────────────────────────────┐
│  planton CLI                    │
│  + PerconaPostgresqlOperator manifest   │
└────────────┬────────────────────────────┘
             │ Deploys via Pulumi
             ▼
┌─────────────────────────────────────────┐
│  Kubernetes Cluster                     │
│                                         │
│  ┌───────────────────────────────────┐  │
│  │ Namespace: percona-postgresql-    │  │
│  │           operator                │  │
│  │                                   │  │
│  │  - Operator Pod (running)         │  │
│  │  - ServiceAccount                 │  │
│  │  - RBAC (ClusterRole, Bindings)   │  │
│  │  - CRDs (installed)               │  │
│  └───────────────────────────────────┘  │
└─────────────────────────────────────────┘
             │ Watches for
             ▼
┌─────────────────────────────────────────┐
│  PostgreSQL Cluster CRs                 │
│  (Created by users)                     │
│                                         │
│  - PerconaPGCluster                     │
│  - PerconaPGBackup                      │
│  - PerconaPGRestore                     │
└─────────────────────────────────────────┘
             │ Operator reconciles
             ▼
┌─────────────────────────────────────────┐
│  PostgreSQL Database Clusters           │
│  - StatefulSets (PostgreSQL instances)  │
│  - Services (primary, replicas)         │
│  - ConfigMaps (configuration)           │
│  - Secrets (credentials)                │
│  - PersistentVolumeClaims (data)        │
│  - Deployments (PgBouncer, monitoring)  │
└─────────────────────────────────────────┘
```

## Verification Commands

```bash
# Check operator status
kubectl get pods -n percona-postgresql-operator

# View operator logs
kubectl logs -n percona-postgresql-operator -l app.kubernetes.io/name=percona-postgresql-operator -f

# Verify CRDs
kubectl get crds | grep percona

# Check operator version
kubectl get deployment -n percona-postgresql-operator pg-operator -o yaml | grep image:

# Test cluster readiness
kubectl get deployment -n percona-postgresql-operator pg-operator \
  -o jsonpath='{.status.conditions[?(@.type=="Available")].status}'
# Should output: True
```

## Benefits

1. **Simplified Operations**: Declarative PostgreSQL cluster management via Kubernetes manifests
2. **Production Ready**: Enterprise-tested operator based on Crunchy Data's battle-tested PGO
3. **High Availability**: Automated replica management, failover, and streaming replication
4. **Backup & Recovery**: Built-in pgBackRest integration with PITR capabilities
5. **Security**: TLS encryption, RBAC, certificate management, and authentication
6. **Monitoring**: Native Prometheus integration and PostgreSQL Exporter
7. **Scalability**: Easy horizontal scaling via replica count changes
8. **Connection Pooling**: Integrated PgBouncer for high-performance applications
9. **Zero Downtime**: Rolling updates and version upgrades without downtime
10. **Multi-Tenancy**: Deploy multiple isolated PostgreSQL clusters
11. **Cloud Agnostic**: Works on any Kubernetes cluster (GKE, EKS, AKS, on-prem)

## Operator vs. Other PostgreSQL Deployment Methods

| Feature | Percona Operator | Zalando Operator | Helm Chart | Manual Deployment |
|---------|------------------|------------------|------------|-------------------|
| Automated Backups | ✅ pgBackRest | ✅ WAL-G | ⚠️ Manual setup | ❌ Fully manual |
| High Availability | ✅ Patroni-based | ✅ Patroni-based | ⚠️ Manual config | ❌ Manual config |
| Connection Pooling | ✅ PgBouncer | ✅ PgBouncer | ⚠️ Manual setup | ❌ DIY |
| Rolling Updates | ✅ Zero downtime | ✅ Zero downtime | ⚠️ Requires planning | ❌ Risky |
| Monitoring | ✅ Prometheus | ✅ Prometheus | ⚠️ Manual setup | ❌ DIY |
| Disaster Recovery | ✅ PITR support | ✅ PITR support | ❌ Not included | ❌ DIY |
| Scaling | ✅ Change replica count | ✅ Change replica count | ⚠️ Manual steps | ❌ Complex |
| Security | ✅ TLS, RBAC, Auth | ✅ TLS, RBAC, Auth | ⚠️ Manual setup | ❌ DIY |

## Percona Product Family

Percona provides operators for three major database systems:

1. **Percona Server for MongoDB Operator** ✅ **(Implemented - Oct 10, 2025)**
   - MongoDB replica sets and sharded clusters
   - Percona Server for MongoDB distribution

2. **Percona Distribution for PostgreSQL Operator** ✅ **(Implemented - Oct 11, 2025)**
   - PostgreSQL high availability clusters
   - Based on Crunchy Data's PGO
   - Patroni-based cluster management

3. **Percona XtraDB Cluster Operator** (Future)
   - MySQL high availability clusters
   - Galera-based synchronous replication

**Note**: With this implementation, Percona MongoDB and PostgreSQL operators are both available, providing comprehensive database management capabilities for the two most popular open-source databases.

## Comparison with Zalando PostgreSQL Operator

Both operators are production-ready but serve different needs:

| Aspect | Percona PG Operator | Zalando PG Operator |
|--------|---------------------|---------------------|
| **Base** | Crunchy Data PGO | Zalando's Spilo |
| **Backup Tool** | pgBackRest | WAL-G |
| **Community** | Percona Enterprise | Zalando Open Source |
| **Support** | Percona Enterprise Support | Community |
| **Maturity** | Mature (PGO lineage) | Very Mature (Zalando production) |
| **Features** | Enterprise focus | Production-tested |

**When to use Percona**:
- Need enterprise support from Percona
- Prefer pgBackRest for backups
- Using other Percona products (MongoDB, MySQL)
- Want Crunchy Data's PGO lineage

**When to use Zalando**:
- Already using Zalando operator successfully
- Prefer WAL-G for backups
- Want Zalando's production-proven approach
- Community support is sufficient

## Migration Guide

### From Zalando PostgreSQL Operator

Not recommended for existing deployments. Both operators manage PostgreSQL but with different CRDs and architectures. For new deployments, choose one operator and standardize on it.

### From Manual PostgreSQL Deployment

1. Deploy the Percona operator using the manifest
2. Create a `PerconaPGCluster` CR matching your current configuration
3. Backup existing data using pg_dump
4. Create new operator-managed cluster
5. Restore data using pg_restore
6. Update application connection strings
7. Decommission old deployment

## Security Considerations

- **Operator Permissions**: ClusterRole with broad Kubernetes API access (required for operator function)
- **Database Secrets**: Automatically managed by operator
- **TLS Certificates**: Can be auto-generated or externally provided
- **Backup Credentials**: Store in Kubernetes Secrets (not in manifests)
- **Network Policies**: Configure to restrict PostgreSQL cluster access
- **RBAC**: Enable PostgreSQL authentication and role-based access
- **Connection Pooling**: PgBouncer provides connection security and rate limiting

## Performance Considerations

- **Resource Allocation**: Operator uses minimal resources (100m CPU, 256Mi memory)
- **Database Performance**: Depends on PostgreSQL cluster configuration (not operator)
- **Backup Impact**: pgBackRest uses streaming backups (minimal impact)
- **Connection Pooling**: PgBouncer improves application performance
- **Storage**: Use high-performance storage classes for production databases

## Test Manifests Created

Two test manifests have been created for easy deployment:

1. **Planton Environment** (development/testing):
   - Path: `planton/ops/organizations/planton/infra-hub/cloud-resources/kubernetes/addon/percona-postgresql-operator.yaml`
   - Cluster: `k8scred_01k789v5ewezr0f45j5zht9ysj`

2. **App-Prod Environment** (production):
   - Path: `planton/ops/organizations/planton/infra-hub/cloud-resources/app-prod/kubernetes/addon/percona-postgresql-operator.yaml`
   - Cluster: `k8scred_01jp6qzdsj70s228htskj53214`

Both manifests include complete deployment commands in their respective README files.

## Future Enhancements

### Planned Features

1. **PostgreSQLKubernetes API Resource**: High-level abstraction for database deployment
2. **Backup Configuration**: Operator-level and per-database backup settings
3. **Monitoring Integration**: Built-in Prometheus ServiceMonitor
4. **Disaster Recovery**: Automated restore testing
5. **Multi-Cluster**: Deploy databases across multiple Kubernetes clusters
6. **Connection Pool Management**: Advanced PgBouncer configuration
7. **Version Management**: Automated PostgreSQL version upgrades
8. **Extensions Management**: Easy installation of PostgreSQL extensions

### Integration Opportunities

- **External Secrets**: Integration with external secret managers (Vault, AWS Secrets Manager)
- **Backup to R2**: Cloudflare R2 backup configuration (similar to Zalando operator)
- **Service Mesh**: Istio/Linkerd integration for traffic management
- **Cost Optimization**: Automated scaling based on load
- **Multi-Region**: Cross-region replication for disaster recovery

## Related Resources

- **Percona Operator Documentation**: https://docs.percona.com/percona-operator-for-postgresql/
- **Crunchy Data PGO**: https://access.crunchydata.com/documentation/postgres-operator/
- **Helm Chart**: https://github.com/percona/percona-helm-charts
- **Operator GitHub**: https://github.com/percona/percona-postgresql-operator
- **PostgreSQL CRD Reference**: https://docs.percona.com/percona-operator-for-postgresql/api.html

## Breaking Changes

None. This is a new feature with no impact on existing resources.

## Deployment Status

✅ **Module created and tested** - Ready for deployment  
✅ **Proto definitions compiled** - All Go stubs generated  
✅ **Pulumi module complete** - Helm-based deployment ready  
✅ **Terraform module complete** - Alternative IaC option available  
✅ **Documentation complete** - README, examples, and usage guides  
✅ **Test manifests ready** - Available for both environments

---

**Next Steps**: Deploy the operator to test/production clusters and create PostgreSQL database clusters using the `PerconaPGCluster` custom resource.

