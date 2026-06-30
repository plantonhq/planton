# Percona Server MongoDB Operator Support

**Date**: October 10, 2025  
**Type**: Feature  
**Component**: PerconaServerMongodbOperator

## Summary

Added support for deploying and managing the Percona Server for MongoDB Operator on Kubernetes clusters. This operator enables automated deployment, management, and operations of production-ready MongoDB databases with built-in backup, monitoring, and high availability features.

## Motivation

Organizations running MongoDB on Kubernetes needed enterprise-grade database management capabilities including:
- Automated MongoDB cluster deployment and scaling
- Built-in backup and disaster recovery
- High availability with replica sets and sharding
- Monitoring and observability integration
- Security hardening and encryption
- Simplified database operations and upgrades

The Percona Server for MongoDB Operator provides these capabilities through Kubernetes-native declarative configuration, reducing operational complexity and improving database reliability.

## What's New

### 1. PerconaServerMongodbOperator API Resource

New Kubernetes cloud resource kind for deploying the Percona operator:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: PerconaServerMongodbOperator
metadata:
  name: mongodb-operator
spec:
  targetCluster:
    kubernetesProviderConfigId: k8s-cluster-01
  namespace: mongodb-operator
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

Added `PerconaServerMongodbOperator` to the cloud resource kind enum:

```protobuf
PerconaServerMongodbOperator = 834 [(kind_meta) = {
  provider: kubernetes, 
  version: v1, 
  id_prefix: "percmdbop", 
  kubernetes_meta: {category: addon}
}];
```

**Properties**:
- Enum value: 834
- ID prefix: `percmdbop` (Percona MongoDB Operator)
- Provider: Kubernetes
- Category: Addon (operator-level infrastructure)

### 3. Operator Capabilities

The Percona Server for MongoDB Operator manages:

**MongoDB Cluster Types**:
- Replica Sets (high availability)
- Sharded Clusters (horizontal scaling)
- Non-sharded deployments (simple use cases)

**Operational Features**:
- Automated backups to S3-compatible storage
- Point-in-Time Recovery (PITR)
- Scheduled and on-demand backups
- Automated failover and self-healing
- Rolling updates with zero downtime
- TLS/SSL encryption
- RBAC and authentication management

**Monitoring Integration**:
- Prometheus metrics export
- PMM (Percona Monitoring and Management) support
- Custom metric collection

## Implementation Details

### Pulumi Module

**Location**: `apis/project/planton/provider/kubernetes/addon/perconaservermongodboperator/v1/iac/pulumi`

**Key Files**:
- `main.go` - Main Pulumi program
- `namespace.go` - Namespace creation
- `operator.go` - Helm release for operator
- `outputs.go` - Stack outputs
- `vars.go` - Input variables

**Helm Chart**:
- Chart: `percona/psmdb-operator`
- Repository: `https://percona.github.io/percona-helm-charts`
- Version: Latest stable release
- Deployment: Single operator pod per cluster

### CRDs Installed

The operator installs three Custom Resource Definitions:

1. **PerconaServerMongoDB** (`psmdb.percona.com/v1`)
   - Primary CRD for MongoDB cluster management
   - Defines cluster topology, replicas, storage
   - Configures backups, monitoring, security

2. **PerconaServerMongoDBBackup** (`psmdb.percona.com/v1`)
   - On-demand backup resource
   - Triggers immediate backups
   - Specifies backup destination and retention

3. **PerconaServerMongoDBRestore** (`psmdb.percona.com/v1`)
   - Database restore operations
   - Point-in-time recovery
   - Clone databases from backups

### Deployment Verification

After successful deployment, verification shows:

```bash
# Operator pod running
$ kubectl get pods -n mongodb-operator
NAME                              READY   STATUS    RESTARTS   AGE
psmdb-operator-6dd9cbb89c-2jgf9   1/1     Running   0          66s

# Operator logs show successful initialization
$ kubectl logs -n mongodb-operator psmdb-operator-xxx
INFO    setup    Manager starting up
INFO    server version    {"platform": "kubernetes", "version": "v1.30.14"}
INFO    Starting Controller    {"controller": "psmdb-controller"}
INFO    Starting Controller    {"controller": "psmdbbackup-controller"}
INFO    Starting Controller    {"controller": "psmdbrestore-controller"}

# CRDs installed
$ kubectl get crds | grep percona
perconaservermongodbs.psmdb.percona.com
perconaservermongodbbackups.psmdb.percona.com
perconaservermongodbrestores.psmdb.percona.com
```

## Usage Examples

### Deploy Operator to Production Cluster

```bash
# Set local module path
export PERCONA_SERVER_MONGODB_OPERATOR_MODULE=~/scm/github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes/addon/perconaservermongodboperator/v1/iac/pulumi

# Initialize Pulumi stack
planton pulumi init \
  --manifest mongodb-operator.yaml \
  --module-dir ${PERCONA_SERVER_MONGODB_OPERATOR_MODULE}

# Preview changes
planton pulumi preview \
  --manifest mongodb-operator.yaml \
  --module-dir ${PERCONA_SERVER_MONGODB_OPERATOR_MODULE}

# Deploy operator
planton pulumi up \
  --manifest mongodb-operator.yaml \
  --module-dir ${PERCONA_SERVER_MONGODB_OPERATOR_MODULE}
```

### Basic MongoDB Cluster (Future)

Once the operator is deployed, MongoDB clusters can be created:

```yaml
apiVersion: psmdb.percona.com/v1
kind: PerconaServerMongoDB
metadata:
  name: my-cluster
  namespace: mongodb-operator
spec:
  crVersion: 1.20.1
  image: percona/percona-server-mongodb:8.0.4-3
  replsets:
  - name: rs0
    size: 3
    volumeSpec:
      persistentVolumeClaim:
        resources:
          requests:
            storage: 10Gi
  backup:
    enabled: true
    tasks:
    - name: daily-backup
      enabled: true
      schedule: "0 2 * * *"
```

## Architecture

### Component Interaction

```
┌─────────────────────────────────────────┐
│  planton CLI                    │
│  + PerconaServerMongodbOperator manifest│
└────────────┬────────────────────────────┘
             │ Deploys via Pulumi
             ▼
┌─────────────────────────────────────────┐
│  Kubernetes Cluster                     │
│                                         │
│  ┌───────────────────────────────────┐  │
│  │ Namespace: mongodb-operator       │  │
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
│  MongoDB Cluster CRs                    │
│  (Created by users)                     │
│                                         │
│  - PerconaServerMongoDB                 │
│  - PerconaServerMongoDBBackup           │
│  - PerconaServerMongoDBRestore          │
└─────────────────────────────────────────┘
             │ Operator reconciles
             ▼
┌─────────────────────────────────────────┐
│  MongoDB Database Clusters              │
│  - StatefulSets                         │
│  - Services                             │
│  - ConfigMaps                           │
│  - Secrets                              │
│  - PersistentVolumeClaims               │
└─────────────────────────────────────────┘
```

### Production Deployment

**Deployed to**: `app-prod` Kubernetes cluster  
**Namespace**: `mongodb-operator`  
**Cluster**: GKE cluster in production environment  
**Operator Version**: 1.20.1 (release-1-20-1 branch)  
**Git Commit**: `30d9ec941baf57619c8973249a3c5d3fd5cc08f4`

## Verification Commands

```bash
# Check operator status
kubectl get pods -n mongodb-operator

# View operator logs
kubectl logs -n mongodb-operator -l app.kubernetes.io/name=percona-server-mongodb-operator -f

# Verify CRDs
kubectl get crds | grep percona

# Check operator version
kubectl get deployment -n mongodb-operator psmdb-operator -o yaml | grep image:

# Test cluster readiness
kubectl get deployment -n mongodb-operator psmdb-operator \
  -o jsonpath='{.status.conditions[?(@.type=="Available")].status}'
# Should output: True
```

## Benefits

1. **Simplified Operations**: Declarative MongoDB cluster management via Kubernetes manifests
2. **Production Ready**: Battle-tested operator used by thousands of organizations
3. **High Availability**: Automated replica set management and failover
4. **Backup & Recovery**: Built-in backup to S3-compatible storage with PITR
5. **Security**: TLS encryption, RBAC, and authentication management
6. **Monitoring**: Native Prometheus integration and PMM support
7. **Scalability**: Easy horizontal scaling via replica count changes
8. **Zero Downtime**: Rolling updates and version upgrades
9. **Multi-Tenancy**: Deploy multiple isolated MongoDB clusters
10. **Cloud Agnostic**: Works on any Kubernetes cluster (GKE, EKS, AKS, on-prem)

## Operator vs. Other MongoDB Deployment Methods

| Feature | Percona Operator | Helm Chart | Manual Deployment |
|---------|------------------|------------|-------------------|
| Automated Backups | ✅ Built-in | ⚠️ Manual setup | ❌ Fully manual |
| High Availability | ✅ Automatic | ⚠️ Manual config | ❌ Manual config |
| Rolling Updates | ✅ Zero downtime | ⚠️ Requires planning | ❌ Risky |
| Monitoring | ✅ Prometheus/PMM | ⚠️ Manual setup | ❌ DIY |
| Disaster Recovery | ✅ PITR support | ❌ Not included | ❌ DIY |
| Scaling | ✅ Change replica count | ⚠️ Manual steps | ❌ Complex |
| Security | ✅ TLS, RBAC, Auth | ⚠️ Manual setup | ❌ DIY |

## Percona Product Family

Percona provides operators for three major database systems:

1. **Percona Server for MongoDB Operator** ✅ **(Implemented)**
   - MongoDB replica sets and sharded clusters
   - Percona Server for MongoDB distribution

2. **Percona XtraDB Cluster Operator** (Future)
   - MySQL high availability clusters
   - Galera-based synchronous replication

3. **Percona Server for PostgreSQL Operator** (Future)
   - PostgreSQL high availability
   - Patroni-based cluster management

**Note**: This implementation focuses specifically on MongoDB. The naming clarifies the scope and leaves room for future Percona operator integrations.

## Migration Guide

### From Manual MongoDB Deployment

1. Deploy the Percona operator using the manifest
2. Create a `PerconaServerMongoDB` CR matching your current configuration
3. Backup existing data
4. Create new operator-managed cluster
5. Migrate data using mongodump/mongorestore
6. Update application connection strings
7. Decommission old deployment

### From MongoDB Community Operator

The Percona operator uses similar CRD structures but with enhanced features:
- Review Percona CRD documentation
- Compare configurations
- Test in non-production environment
- Plan migration during maintenance window

## Security Considerations

- **Operator Permissions**: ClusterRole with broad Kubernetes API access (required for operator function)
- **Database Secrets**: Automatically managed by operator
- **TLS Certificates**: Can be auto-generated or externally provided
- **Backup Credentials**: Store in Kubernetes Secrets (not in manifests)
- **Network Policies**: Configure to restrict MongoDB cluster access
- **RBAC**: Enable MongoDB authentication and role-based access

## Performance Considerations

- **Resource Allocation**: Operator uses minimal resources (100m CPU, 256Mi memory)
- **Database Performance**: Depends on MongoDB cluster configuration (not operator)
- **Backup Impact**: Backups use replica set secondaries (no primary impact)
- **Storage**: Use high-performance storage classes for production databases

## Future Enhancements

### Planned Features

1. **MongoDBKubernetes API Resource**: High-level abstraction for database deployment
2. **Backup Configuration**: Operator-level and per-database backup settings
3. **Monitoring Integration**: Built-in Prometheus ServiceMonitor
4. **PMM Integration**: Automatic monitoring agent deployment
5. **Disaster Recovery**: Automated restore testing
6. **Multi-Cluster**: Deploy databases across multiple Kubernetes clusters
7. **Sharding Support**: Simplified sharded cluster configuration
8. **Version Management**: Automated MongoDB version upgrades

### Integration Opportunities

- **External Secrets**: Integration with external secret managers (Vault, AWS Secrets Manager)
- **Backup to R2**: Cloudflare R2 backup configuration (similar to PostgreSQL operator)
- **Service Mesh**: Istio/Linkerd integration for traffic management
- **Cost Optimization**: Automated scaling based on load

## Related Resources

- **Percona Operator Documentation**: https://docs.percona.com/percona-operator-for-mongodb/
- **Helm Chart**: https://github.com/percona/percona-helm-charts
- **Operator GitHub**: https://github.com/percona/percona-server-mongodb-operator
- **MongoDB CRD Reference**: https://docs.percona.com/percona-operator-for-mongodb/crd.html

## Breaking Changes

None. This is a new feature with no impact on existing resources.

## Deployment Status

✅ **Successfully deployed to production** (`app-prod` cluster)  
✅ **Operator pod running** and managing MongoDB CRDs  
✅ **CRDs installed** and ready for MongoDB cluster creation  
✅ **Verification complete** - all health checks passing

---

**Next Steps**: Deploy production MongoDB databases using the operator's `PerconaServerMongoDB` custom resource.

