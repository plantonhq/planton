# KubernetesDaemonSet Technical Documentation

## Architecture Overview

The KubernetesDaemonSet component provides a declarative way to deploy DaemonSets to Kubernetes clusters. DaemonSets ensure that a copy of a pod runs on all (or a subset of) nodes in the cluster.

### How DaemonSets Work

1. **Node Scheduling**: When a DaemonSet is created, the DaemonSet controller creates pods on all matching nodes
2. **Automatic Scaling**: As new nodes are added to the cluster, pods are automatically scheduled
3. **Garbage Collection**: When nodes are removed, the associated pods are garbage collected
4. **Node Affinity**: Pods can be constrained to specific nodes using node selectors and tolerations

### Key Kubernetes Resources Created

| Resource | Description |
|----------|-------------|
| Namespace | Optional, created when `create_namespace: true` |
| DaemonSet | The main workload controller |
| Secret (env-secrets) | Contains environment secrets |
| Secret (image-pull) | Optional, for private registry authentication |

## Implementation Details

### Update Strategies

#### RollingUpdate (Default)
- Pods are updated progressively
- `maxUnavailable`: Maximum pods that can be unavailable during update (default: 1)
- `maxSurge`: Maximum pods that can be created above desired count

#### OnDelete
- Pods are only updated when manually deleted
- Useful for controlled maintenance windows

### Volume Mounts

DaemonSets commonly need access to host filesystems:

```yaml
volume_mounts:
  - name: varlog
    mount_path: /var/log
    host_path: /var/log
    host_path_type: DirectoryOrCreate
    read_only: true
```

Supported `host_path_type` values:
- `DirectoryOrCreate`: Create directory if it doesn't exist
- `Directory`: Directory must exist
- `FileOrCreate`: Create file if it doesn't exist
- `File`: File must exist
- `Socket`: Unix socket must exist
- `CharDevice`: Character device must exist
- `BlockDevice`: Block device must exist

### Tolerations

Tolerations allow DaemonSet pods to run on tainted nodes:

```yaml
tolerations:
  # Run on master nodes
  - key: node-role.kubernetes.io/master
    operator: Exists
    effect: NoSchedule
  # Run on all nodes regardless of taints
  - operator: Exists
```

### Security Context

For node-level operations, DaemonSets often need elevated privileges:

```yaml
security_context:
  privileged: true
  run_as_user: 0
  capabilities:
    add:
      - NET_ADMIN
      - SYS_PTRACE
    drop:
      - ALL
```

## Common Use Cases

### 1. Log Collection
- Collect logs from all nodes
- Mount `/var/log` and container log directories
- Examples: Fluentd, Fluent Bit, Filebeat

### 2. Node Monitoring
- Expose node metrics
- Mount `/proc`, `/sys`, and other system paths
- Examples: Prometheus Node Exporter, Datadog Agent

### 3. Network Plugins
- Implement CNI or network policies
- Require privileged access and host networking
- Examples: Calico, Cilium, Weave

### 4. Storage Daemons
- Provide distributed storage
- Mount block devices and storage paths
- Examples: Ceph, Longhorn

### 5. Security Agents
- Runtime security monitoring
- Require privileged access to monitor system calls
- Examples: Falco, Sysdig

## Best Practices

1. **Resource Limits**: Always set resource limits to prevent DaemonSet pods from consuming all node resources
2. **Node Selectors**: Use node selectors to target specific node types
3. **Read-Only Mounts**: Mount host paths as read-only when writes aren't needed
4. **Minimal Privileges**: Request only the capabilities actually needed
5. **Update Strategy**: Use RollingUpdate with appropriate maxUnavailable for zero-downtime updates

## Troubleshooting

### Pods Not Scheduling

Check if:
- Node selectors match node labels
- Tolerations match node taints
- Resource requests are satisfiable

```bash
kubectl describe daemonset <name> -n <namespace>
kubectl get events -n <namespace>
```

### Pods Crashing

Check container logs:
```bash
kubectl logs -n <namespace> -l app=daemonset --all-containers
```

Check security context and volume mount permissions.

