# On-Demand Autoscaling Node Pool

This preset creates a GKE node pool with on-demand (non-preemptible) VMs, SSD boot disks, and cluster autoscaler enabled. It scales between 1 and 5 nodes per zone with balanced distribution, making it the standard choice for production workloads.

## When to Use

- Production GKE workloads that need reliable, non-interruptible compute
- General-purpose node pools for web services, APIs, and background workers
- Environments where autoscaling is preferred over fixed capacity

## Key Configuration Choices

- **e2-standard-4** (`machineType`) -- 4 vCPU, 16 GB RAM; cost-effective for most workloads
- **SSD boot disk** (`diskType: pd-ssd`) -- faster I/O for container image loading and ephemeral storage
- **100 GB disk** (`diskSizeGb: 100`) -- sufficient for container images and local storage
- **COS_CONTAINERD** (`imageType`) -- Container-Optimized OS, the GCP-recommended node image
- **Autoscaling 1-5** -- minimum 1 node per zone (always-on capacity), maximum 5 per zone
- **BALANCED location policy** -- distributes nodes evenly across zones for availability
- **Auto-upgrade and auto-repair** -- enabled by default (not disabled)

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<gcp-project-id>` | GCP project ID hosting the GKE cluster | `GcpProject` outputs |
| `<gke-cluster-name>` | Name of the parent GKE cluster | `GcpGkeCluster` metadata name |
| `<gcp-region>` | Location of the GKE cluster (e.g., `us-central1`) | `GcpGkeCluster` spec location |
| `<your-node-pool-name>` | Name for this node pool (1-40 chars, lowercase) | Choose a descriptive name (e.g., `general`) |

## Related Presets

- **02-spot-cost-optimized** -- Use for non-critical workloads where cost savings outweigh availability guarantees
