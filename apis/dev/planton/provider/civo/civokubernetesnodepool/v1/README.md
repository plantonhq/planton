# Civo Kubernetes Node Pool

Manage additional node pools for existing Civo Kubernetes clusters with independent sizing and autoscaling.

## Overview

The `CivoKubernetesNodePool` resource adds additional node pools to existing Civo Kubernetes clusters. Node pools allow you to segment workloads across different instance types, enabling cost optimization and workload isolation.

**Key features:**

- **Multiple node pools** - Run different workload types on appropriately-sized nodes
- **Independent scaling** - Scale each pool independently
- **Autoscaling** - Configure min/max bounds for automatic scaling
- **Flexible sizing** - Choose from small to xlarge instances per pool
- **Workload isolation** - Use taints/labels to isolate workloads
- **Cost optimization** - Right-size nodes for specific workloads

## Prerequisites

- Civo account with API access
- Civo API token
- Existing Civo Kubernetes cluster - use `CivoKubernetesCluster` resource
- Planton CLI installed
- `kubectl` configured for cluster access

## Quick Start

### 1. General-Purpose Worker Pool

Add a standard worker pool for general workloads:

```yaml
apiVersion: civo.planton.dev/v1
kind: CivoKubernetesNodePool
metadata:
  name: general-workers
spec:
  nodePoolName: general-workers
  cluster:
    value: "your-cluster-name"
  size: g4s.kube.medium
  nodeCount: 3
  autoScale: true
  minNodes: 2
  maxNodes: 8
  tags:
    - workload:general
    - tier:workers
```

### 2. Compute-Intensive Pool

Large nodes for CPU/memory-intensive workloads:

```yaml
apiVersion: civo.planton.dev/v1
kind: CivoKubernetesNodePool
metadata:
  name: compute-pool
spec:
  nodePoolName: compute-intensive
  cluster:
    value: "your-cluster-name"
  size: g4s.kube.xlarge
  nodeCount: 2
  tags:
    - workload:compute
    - team:data-science
```

### 3. Batch Processing Pool

Autoscaling pool for batch jobs (scale to zero):

```yaml
apiVersion: civo.planton.dev/v1
kind: CivoKubernetesNodePool
metadata:
  name: batch-pool
spec:
  nodePoolName: batch-workers
  cluster:
    value: "your-cluster-name"
  size: g4s.kube.large
  nodeCount: 0
  autoScale: true
  minNodes: 0
  maxNodes: 20
  tags:
    - workload:batch
    - autoscale:enabled
```

## Deploy with Planton CLI

```bash
# Create the node pool
planton apply -f nodepool.yaml

# Check status
planton get civokubernetesnodepools

# View outputs
planton outputs civokubernetesnodepools/general-workers

# Verify in cluster
kubectl get nodes
```

## Configuration Reference

### Spec Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `nodePoolName` | string | Yes | Unique name within cluster |
| `cluster` | StringValueOrRef | Yes | Reference to parent cluster |
| `size` | string | Yes | Node instance size |
| `nodeCount` | uint32 | Yes | Number of nodes (must be > 0) |
| `autoScale` | bool | No | Enable autoscaling (default: `false`) |
| `minNodes` | uint32 | No | Min nodes when autoscaling (required if `autoScale=true`) |
| `maxNodes` | uint32 | No | Max nodes when autoscaling (required if `autoScale=true`) |
| `tags` | array[string] | No | Resource tags |

### Available Node Sizes

| Size | CPU | RAM | Storage | Monthly Cost* |
|------|-----|-----|---------|---------------|
| `g4s.kube.small` | 1 | 2 GB | 30 GB | ~$11 |
| `g4s.kube.medium` | 2 | 4 GB | 50 GB | ~$22 |
| `g4s.kube.large` | 4 | 8 GB | 100 GB | ~$43 |
| `g4s.kube.xlarge` | 6 | 16 GB | 150 GB | ~$87 |

*Per node, approximate as of 2025

## Stack Outputs

After provisioning:

- `node_pool_id` - Civo node pool identifier

Access via:

```bash
planton outputs civokubernetesnodepools/your-pool-name
```

## Common Use Cases

### Workload Segmentation

Run different workload types on appropriate nodes:

```yaml
# General workloads - Medium nodes
- nodePoolName: general-workers
  size: g4s.kube.medium
  nodeCount: 5

# Memory-intensive - Large nodes  
- nodePoolName: memory-intensive
  size: g4s.kube.large
  nodeCount: 3

# Batch jobs - Autoscaling
- nodePoolName: batch
  size: g4s.kube.medium
  nodeCount: 0
  autoScale: true
  minNodes: 0
  maxNodes: 10
```

### Cost Optimization

Use smaller nodes for lightweight workloads:

```yaml
# API services (small footprint)
- nodePoolName: api-workers
  size: g4s.kube.small
  nodeCount: 5
  tags:
    - workload:api
```

### Autoscaling for Variable Load

```yaml
# Web traffic pool (scales with load)
- nodePoolName: web-workers
  size: g4s.kube.medium
  nodeCount: 3  # Initial size
  autoScale: true
  minNodes: 2   # Always at least 2
  maxNodes: 12  # Scale up to 12 during peak
```

## Best Practices

### 1. Pool Naming

Use descriptive names that indicate purpose:

```yaml
nodePoolName: "memory-intensive-workers"  # Clear purpose
nodePoolName: "workers-01"  # Less clear
```

### 2. Autoscaling Configuration

Set reasonable bounds:

```yaml
# Good: Room to scale, but bounded
autoScale: true
minNodes: 2   # Minimum for availability
maxNodes: 10  # Prevents runaway costs
```

```yaml
# Risky: Too wide range
autoScale: true
minNodes: 1
maxNodes: 100  # Could incur massive costs
```

### 3. Node Count vs Autoscaling

When `autoScale=true`:
- `nodeCount` is the initial size
- Pool scales between `minNodes` and `maxNodes`
- Autoscaler adjusts based on pod resource requests

### 4. Workload Isolation

Use node selectors and taints:

```bash
# Add taint to node pool nodes (after creation)
kubectl taint nodes -l pool=compute-intensive workload=compute:NoSchedule

# Deploy pods with toleration
spec:
  tolerations:
    - key: workload
      value: compute
      effect: NoSchedule
  nodeSelector:
    pool: compute-intensive
```

### 5. Cost Management

Monitor node pool costs:

```bash
# Check current node count
kubectl get nodes -l pool=workers

# Calculate costs
# nodes × hourly_rate × 730 hours = monthly cost
```

### 6. Right-Sizing Strategy

Start small, scale up:

1. Begin with `g4s.kube.medium`
2. Monitor resource usage
3. Scale up if CPU/memory regularly hits limits
4. Or add more nodes if capacity is the issue

## Autoscaling

### How It Works

Civo Kubernetes autoscaling:
1. Pod requests resources that don't fit on existing nodes
2. Cluster autoscaler detects pending pods
3. Autoscaler requests additional nodes (up to `maxNodes`)
4. Nodes are provisioned (~60 seconds)
5. Pods are scheduled on new nodes

**Scale-down:**
- Nodes with low utilization (<50%) for 10+ minutes are candidates
- Autoscaler drains and removes underutilized nodes
- Scales down to `minNodes`

### Configuration

```yaml
spec:
  nodeCount: 3      # Start with 3 nodes
  autoScale: true   # Enable autoscaling
  minNodes: 2       # Never go below 2
  maxNodes: 10      # Never exceed 10
```

**Important**: Pods must have resource requests defined:

```yaml
resources:
  requests:
    cpu: "500m"
    memory: "512Mi"
  limits:
    cpu: "1000m"
    memory: "1Gi"
```

Without requests, autoscaler cannot calculate capacity.

## Troubleshooting

### Nodes Not Adding to Cluster

```bash
# Check node pool status
planton get civokubernetesnodepools/your-pool

# Verify cluster is healthy
kubectl get nodes

# Check for errors
kubectl describe node <node-name>
```

### Autoscaling Not Working

**Symptom:** Pods remain pending despite autoscaling enabled.

**Common causes:**

1. **Missing resource requests**:

```bash
kubectl describe pod <pending-pod>
# Look for: "0/3 nodes are available: insufficient cpu"
```

Fix: Add resource requests to pod spec.

2. **Max nodes reached**:

```bash
kubectl get nodes | wc -l
# If at maxNodes, autoscaler won't add more
```

Fix: Increase `maxNodes` or scale manually.

3. **Autoscaler not installed**:

```bash
kubectl get pods -n kube-system | grep autoscaler
```

Fix: Install Civo cluster autoscaler.

### Nodes Not Draining

**Symptom:** Autoscaler won't remove nodes.

**Causes:**
- Pods without PodDisruptionBudget
- DaemonSets blocking drain
- Local storage preventing eviction

```bash
# Check what's preventing drain
kubectl describe node <node-name>
```

## More Information

- **Deep Dive** - See [docs/README.md](docs/README.md) for comprehensive research on node pool patterns and autoscaling strategies
- **Examples** - Check [examples.md](examples.md) for more node pool configurations
- **Pulumi Module** - See [iac/pulumi/README.md](iac/pulumi/README.md) for direct Pulumi usage
- **Civo API** - [Official documentation](https://www.civo.com/api/kubernetes)

## Support

- Issues: [GitHub](https://github.com/plantonhq/planton/issues)
- Civo Support: support@civo.com
- Community: [Planton Discord](#)

## Related Resources

- `CivoKubernetesCluster` - Create clusters before adding pools
- `CivoVpc` - Network for cluster
- `CivoFirewall` - Secure cluster access

