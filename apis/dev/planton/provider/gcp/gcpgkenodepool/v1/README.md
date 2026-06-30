# GCP GKE Node Pool

**Deployment Component for Google Kubernetes Engine Node Pools**

## Overview

`GcpGkeNodePool` is a Planton deployment component that manages **GKE node pools**—groups of Compute Engine VMs that run your Kubernetes workloads. Node pools enable heterogeneous infrastructure within a single cluster, allowing you to match compute resources precisely to workload requirements.

Unlike the "one-size-fits-all" approach of homogeneous clusters, node pools let you create specialized compute tiers:
- **Cost-optimized pools** with Spot VMs for batch workloads (up to 91% savings)
- **GPU/TPU pools** for machine learning workloads
- **High-memory pools** for caches and in-memory databases
- **General-purpose pools** for standard web applications

Node pools are **separate resources** from their parent GKE cluster, enabling independent lifecycle management—you can create, update, and destroy node pools without touching the cluster control plane, a critical pattern for production stability.

## Key Features

### Infrastructure as Code
- **Declarative YAML API** following Kubernetes conventions
- Deploy via `planton apply` with state management and drift detection
- Supports both **Pulumi** and **Terraform/OpenTofu** backends

### Production-Ready Configuration
- **Autoscaling**: Scale-to-zero for cost savings or fixed size for predictable capacity
- **Spot VMs**: Up to 91% cost reduction for fault-tolerant workloads
- **Auto-upgrade and auto-repair**: Enabled by default with customizable maintenance windows
- **Custom service accounts**: Least-privilege IAM for security
- **Flexible disk options**: Standard, SSD, or Balanced persistent disks

### Planton Integration
- **Foreign key references** to parent `GcpGkeCluster` resources
- Automatic labeling with org, environment, and resource metadata
- Consistent patterns with other Planton components

## Quick Start

### Minimal Example: Dev/Test Pool with Spot VMs

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpGkeNodePool
metadata:
  name: dev-pool
spec:
  cluster_project_id:
    value: my-gcp-project
  cluster_name:
    value: dev-cluster
  machine_type: e2-medium
  disk_size_gb: 50
  spot: true
  autoscaling:
    min_nodes: 0      # Scale-to-zero for cost savings
    max_nodes: 3
    location_policy: ANY  # Hunt for capacity across zones
```

### Deploy

```bash
planton apply -f dev-pool.yaml
```

### Production Example: High-Availability General-Purpose Pool

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpGkeNodePool
metadata:
  name: prod-general
spec:
  cluster_project_id:
    value: my-gcp-project
  cluster_name:
    value: prod-cluster
  machine_type: n2-standard-4
  disk_size_gb: 100
  disk_type: pd-ssd
  service_account: gke-prod-sa@my-project.iam.gserviceaccount.com
  autoscaling:
    min_nodes: 3      # Multi-zone HA baseline
    max_nodes: 10
    location_policy: BALANCED  # Even spread for availability
  management:
    disable_auto_upgrade: false  # Keep security patches current
    disable_auto_repair: false   # Auto-fix unhealthy nodes
  node_labels:
    workload-tier: general
    environment: production
```

## Architecture

### Separate Resource Pattern

`GcpGkeNodePool` is an **independent resource** that references its parent `GcpGkeCluster`:

```
GcpGkeCluster (Control Plane)
  ↑
  └─ GcpGkeNodePool (Worker Nodes)
     ├─ VM Instance 1
     ├─ VM Instance 2
     └─ VM Instance 3
```

**Why this design?**
- Node pool updates never risk cluster control plane disruption
- Lifecycle independence: create/update/delete pools without touching the cluster
- Industry alignment: mirrors Terraform, Pulumi, and GKE API patterns

### Configuration Inheritance

Node pools automatically inherit:
- **Network configuration** from the parent cluster (VPC, subnetwork, firewall tags)
- **Planton metadata** (org, environment, resource labels)
- **OAuth scopes** for monitoring, logging, and storage

## Best Practices

### Multiple Specialized Pools vs. One Large Pool

**Anti-pattern**: A single, massive node pool with one machine type.

**Best practice**: Create **multiple, specialized node pools** aligned to workload profiles:
- **General-purpose pool**: E2 or N2 series for standard applications
- **High-memory pool**: M2/M3 series for Redis, in-memory databases
- **Compute-optimized pool**: C2 series for CPU-bound workloads
- **Cost-optimized pool**: E2 Spot VMs for batch jobs, CI/CD runners
- **Accelerator pool**: N1 with GPUs for ML inference and training

### Autoscaling Location Policies

- **BALANCED** (default): Spreads nodes evenly across zones. Best for high-availability production workloads.
- **ANY**: Adds nodes wherever capacity is available. **Critical for Spot VM pools**—allows the autoscaler to "hunt" for spare capacity across all zones.

### Spot VMs: Cost Savings with Constraints

Spot VMs offer up to 91% cost reduction but can be preempted (terminated) by Google with short notice.

**Production pattern**:
1. **Mixed pools**: *Never* run a cluster solely on Spot VMs. At least one reliable, on-demand pool must exist for critical system components.
2. **Use for fault-tolerant workloads**: Batch jobs, CI/CD runners, development environments.
3. **Set `location_policy: ANY`** to maximize Spot VM acquisition success.

### Auto-Upgrade and Auto-Repair

**Keep both enabled** (the GKE default). Use **Maintenance Windows** to control *when* upgrades occur, rather than disabling them entirely. This ensures nodes stay current with security patches while minimizing disruption.

## Related Components

- **[GcpGkeCluster](../gcpgkecluster/v1/README.md)**: Parent GKE cluster (control plane)
- **[GcpProject](../gcpproject/v1/README.md)**: GCP project for resource organization
- **[GcpVpc](../gcpvpc/v1/README.md)**: VPC network for cluster connectivity

## Documentation

- **[Examples](./examples.md)**: 5+ working YAML examples covering common patterns
- **[Research Document](./docs/README.md)**: Deep dive into GKE node pools, deployment methods, and the evolution from manual operations to GitOps
- **[Pulumi Module](./iac/pulumi/README.md)**: Standalone Pulumi usage
- **[Terraform Module](./iac/tf/README.md)**: Standalone Terraform/OpenTofu usage

## Support

For issues, questions, or contributions, visit the [Planton repository](https://github.com/plantonhq/planton).

---

**Note**: This component requires an existing `GcpGkeCluster` resource. Node pools cannot exist without a parent cluster.
