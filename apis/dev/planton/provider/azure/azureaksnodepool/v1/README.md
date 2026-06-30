# AzureAksNodePool

Azure AKS node pool resource for adding and managing node pools within existing Azure Kubernetes Service (AKS) clusters. Provides configuration for VM sizing, autoscaling, availability zones, and specialized workload types (System, User, Spot, Windows, GPU).

## Spec Fields (80/20)

### Essential Fields (80% Use Case)
- **cluster_name**: Reference to the parent AKS cluster (must exist)
- **vm_size**: Azure VM SKU for nodes (e.g., `Standard_D4s_v3`, `Standard_D8s_v3`)
  - Determines CPU/memory capacity of each node
- **initial_node_count**: Starting number of nodes in the pool (minimum 1)
  - Without autoscaling, this is the fixed node count
- **availability_zones**: Azure availability zones for high availability (e.g., `["1", "2", "3"]`)
  - Minimum 2 zones recommended for production
  - If unspecified, nodes use region defaults

### Advanced Fields (20% Use Case)
- **autoscaling**: Enable cluster autoscaler for this pool
  - `min_nodes`: Minimum node count (can be 0 for user pools)
  - `max_nodes`: Maximum node count
- **os_type**: Operating system type:
  - `LINUX` (default): Standard Linux nodes
  - `WINDOWS`: Windows Server nodes (requires Windows-enabled cluster)
- **mode**: Node pool purpose:
  - `USER` (default): Application workloads, can scale to zero
  - `SYSTEM`: Critical cluster components (CoreDNS, metrics-server), must be Linux, cannot scale to zero
- **spot_enabled**: Use Spot (preemptible) VMs for cost savings (boolean)
  - Cannot be used with SYSTEM mode pools

## Stack Outputs

- **node_pool_name**: Name of the node pool in AKS (matches metadata.name)
- **agent_pool_resource_id**: Azure Resource Manager ID of the created node pool
- **max_pods_per_node**: Maximum number of pods per node (determined by AKS based on network configuration)

## How It Works

Planton provisions AKS node pools via Pulumi or Terraform modules defined in this repository. The API contract is protobuf-based (api.proto, spec.proto) and stack execution is orchestrated by the platform using the AzureAksNodePoolStackInput (includes Azure credentials and IaC metadata).

### Node Pool Modes

| Mode | Purpose | OS Types | Scale to Zero | Use Case |
|------|---------|----------|---------------|----------|
| **SYSTEM** | Critical components | Linux only | No | CoreDNS, metrics-server, cluster infrastructure |
| **USER** | Application workloads | Linux or Windows | Yes | Your applications, microservices |

**Recommendation**: Every AKS cluster needs at least one SYSTEM pool. Additional pools should be USER mode for applications.

### VM Size Selection

Common VM series for AKS node pools:

| Series | vCPU:RAM Ratio | Best For | Example SKU |
|--------|----------------|----------|-------------|
| **Dsv3** | General purpose | Most workloads | Standard_D4s_v3 (4 vCPU, 16 GB) |
| **Esv3** | Memory optimized | Databases, caches | Standard_E8s_v3 (8 vCPU, 64 GB) |
| **Fsv2** | Compute optimized | Batch processing | Standard_F8s_v2 (8 vCPU, 16 GB) |
| **NCv3** | GPU | ML/AI workloads | Standard_NC6s_v3 (6 vCPU, 112 GB, V100) |

**Cost Optimization**: Start with Standard_D4s_v3 for most workloads, scale to larger SKUs as needed.

### Availability Zones

Azure regions typically have 3 availability zones. Distributing nodes across zones provides:
- **Infrastructure resilience**: Survive datacenter failures
- **SLA improvement**: Higher uptime guarantees
- **Load distribution**: Spread application load geographically

**Recommendation**: Always specify at least 2 zones (preferably 3) for production node pools.

### Autoscaling Behavior

When autoscaling is enabled:
- **Scale Up**: Triggered when pods can't be scheduled due to insufficient resources
- **Scale Down**: Triggered after nodes are underutilized for ~10 minutes
- **Min Nodes**: Lower bound (0 is allowed for USER pools)
- **Max Nodes**: Upper bound to prevent runaway costs

**Production Pattern**: Set `min_nodes` to handle baseline load, `max_nodes` to 2-3x baseline for burst capacity.

### Spot VMs (Preemptible Instances)

Spot VMs offer significant cost savings (60-90% discount) but can be evicted when Azure needs the capacity:
- **Best For**: Batch jobs, CI/CD runners, fault-tolerant stateless workloads
- **Not Suitable For**: Databases, stateful apps, long-running processes
- **Cannot Use For**: SYSTEM mode pools

**Recommendation**: Use Spot for dev/test environments and batch workloads, not for production user-facing services.

## Multi-Environment Best Practice

Common pattern for production AKS clusters:

### Development Environment
- 1 SYSTEM pool: 2-4 Standard_D2s_v3 nodes
- 1 USER pool: 1-3 Standard_D2s_v3 nodes (can use Spot)

### Staging Environment
- 1 SYSTEM pool: 3-6 Standard_D4s_v3 nodes (autoscaling)
- 1 USER pool: 2-8 Standard_D4s_v3 nodes (autoscaling)

### Production Environment
- 1 SYSTEM pool: 3-6 Standard_D4s_v3 nodes (fixed, multi-AZ)
- 1 USER pool: 3-20 Standard_D8s_v3 nodes (autoscaling, multi-AZ)
- Additional pools for specialized workloads (GPU, memory-optimized, etc.)

## Common Use Cases

### Standard Application Pool
USER mode pool with autoscaling for running containerized applications.

### Additional System Pool
Dedicated SYSTEM pool for separating critical cluster components from applications.

### GPU Workload Pool
NCv3 or NCv4 series VMs for machine learning and AI inference workloads.

### Windows Container Pool
WINDOWS os_type for running .NET Framework applications that require Windows Server.

### Spot Instance Pool
Cost-optimized USER pool with spot_enabled for batch processing and dev/test.

### Memory-Intensive Pool
Esv3 series VMs for databases, caches, and in-memory processing.

## Cost Optimization

### VM Sizing
- Start with smaller VM sizes (Standard_D4s_v3) and scale up as needed
- Use `kubectl top nodes` to monitor actual resource usage
- Right-size based on actual workload requirements, not guesses

### Autoscaling
- Set `min_nodes` to handle baseline load (not peak)
- Set `max_nodes` to 2-3x baseline for burst capacity
- Enable scale-to-zero for USER pools in dev/test environments

### Spot VMs
- Use Spot for fault-tolerant workloads (60-90% savings)
- Set appropriate tolerations and affinity rules in your pods
- Have fallback to regular nodes for critical workloads

### Node Pool Strategy
- Use fewer, larger VMs instead of many small VMs (reduces overhead)
- Separate workload types into dedicated pools (isolate noisy neighbors)
- Delete unused node pools

## References

- Azure AKS Node Pools: https://learn.microsoft.com/en-us/azure/aks/use-multiple-node-pools
- VM Sizes: https://learn.microsoft.com/en-us/azure/virtual-machines/sizes
- Spot VMs on AKS: https://learn.microsoft.com/en-us/azure/aks/spot-node-pool
- Cluster Autoscaler: https://learn.microsoft.com/en-us/azure/aks/cluster-autoscaler
- Windows Node Pools: https://learn.microsoft.com/en-us/azure/aks/windows-aks-custom-image
- GPU Node Pools: https://learn.microsoft.com/en-us/azure/aks/gpu-cluster
- Terraform Provider: https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/resources/kubernetes_cluster_node_pool
- Pulumi Provider: https://www.pulumi.com/registry/packages/azure-native/api-docs/containerservice/agentpool/
