# Kubernetes StatefulSet - Technical Documentation

## Overview

KubernetesStatefulSet is a deployment component designed for stateful applications on Kubernetes. Unlike Deployments, StatefulSets provide guarantees about the ordering and uniqueness of pods, making them ideal for applications that require stable network identifiers, stable persistent storage, or ordered deployment and scaling.

## Architecture

### Components Created

1. **Namespace** (optional): Dedicated namespace for resource isolation
2. **ServiceAccount**: Pod identity for RBAC
3. **Headless Service**: Provides stable DNS for pod discovery
4. **ClusterIP Service**: Load-balanced access for clients
5. **StatefulSet**: The workload with ordered pod management
6. **PersistentVolumeClaims**: Per-pod persistent storage
7. **Secrets**: Environment secrets management
8. **Image Pull Secret**: Private registry authentication
9. **Ingress Resources**: External access (optional)
10. **PodDisruptionBudget**: Availability guarantees (optional)

### Network Identity

Each pod in a StatefulSet gets a stable DNS name following this pattern:

```
<pod-name>.<headless-service>.<namespace>.svc.cluster.local
```

For example, with a StatefulSet named `postgres-db` in namespace `database`:
- `postgres-db-0.postgres-db-headless.database.svc.cluster.local`
- `postgres-db-1.postgres-db-headless.database.svc.cluster.local`
- `postgres-db-2.postgres-db-headless.database.svc.cluster.local`

### Storage

Each pod gets its own PersistentVolumeClaim based on the volume claim templates. PVCs are named:

```
<volume-name>-<pod-name>
```

For example: `data-postgres-db-0`, `data-postgres-db-1`

## Comparison with KubernetesDeployment

| Feature | StatefulSet | Deployment |
|---------|-------------|------------|
| Pod naming | Stable, ordered (app-0, app-1) | Random suffix |
| Pod creation | Ordered by default | Parallel |
| Network identity | Stable via headless service | No stable identity |
| Storage | Per-pod PVCs | Shared or none |
| Scaling | Ordered | Parallel |
| Use case | Databases, clusters | Stateless apps |

## Pod Management Policies

### OrderedReady (Default)

- Pods are created one at a time, starting from 0
- Each pod must be Running and Ready before the next is created
- Pods are terminated in reverse order (N-1 to 0)
- Suitable for: Leader-follower databases, ZooKeeper, etcd

### Parallel

- All pods are created/deleted simultaneously
- No ordering guarantees
- Faster scaling but less control
- Suitable for: Peer-based systems, Redis clusters

## Production Considerations

### High Availability

```yaml
availability:
  replicas: 3
  podDisruptionBudget:
    enabled: true
    minAvailable: "2"
```

### Storage Classes

Choose appropriate storage classes for your workload:
- `standard`: General purpose
- `ssd` / `fast-ssd`: High IOPS for databases
- `regional-*`: Multi-zone durability

### Resource Sizing

Consider the following for databases:
- Memory: At least 2x the expected data size for caching
- CPU: Depends on query complexity
- Storage: Plan for 3-5x current data size for growth

## Troubleshooting

### Common Issues

1. **Pods stuck in Pending**: Check PVC binding and storage class
2. **Pod startup order issues**: Verify readiness probes
3. **DNS resolution failures**: Ensure headless service is created first
4. **Storage quota exceeded**: Check namespace resource quotas

### Useful Commands

```bash
# List pods with their ordinals
kubectl get pods -l app=<statefulset-name> -n <namespace>

# Check PVC status
kubectl get pvc -l app=<statefulset-name> -n <namespace>

# Describe StatefulSet
kubectl describe statefulset <name> -n <namespace>

# Check headless service endpoints
kubectl get endpoints <name>-headless -n <namespace>
```

## Best Practices

1. **Always use readiness probes** for proper ordering during scaling
2. **Set resource limits** to prevent noisy neighbor issues
3. **Use PodDisruptionBudgets** for production workloads
4. **Test failover scenarios** before going to production
5. **Implement proper shutdown hooks** in your application
6. **Use init containers** for cluster bootstrapping when needed
