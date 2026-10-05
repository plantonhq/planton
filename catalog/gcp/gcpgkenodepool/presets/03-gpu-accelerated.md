# GPU-Accelerated Pool

This preset creates a scale-to-zero GPU pool for ML workloads: NVIDIA L4 nodes with GKE-managed drivers, image streaming for large ML images, and an explicit GPU taint.

## When to Use

- Model inference serving (L4 is the price/performance choice)
- Fine-tuning and small-scale training (move to A100/H100 machine types for large training)
- Any workload that requests `nvidia.com/gpu` resources

## Key Configuration Choices

- **`g2-standard-8` + `nvidia-l4`** — on accelerator-optimized families (A2/A3/G2) the GPU is part of the machine type, and the `guestAccelerators` block must match it
- **`gpuDriverVersion: DEFAULT`** — GKE installs and upgrades the NVIDIA driver; no DaemonSet to maintain
- **Scale-to-zero** — GPU nodes are the most expensive compute in the cluster; the pool exists only while ML work is queued
- **Image streaming (`gcfsEnabled`)** — multi-GB ML images start minutes faster by pulling data on demand
- **Explicit GPU taint** — GKE taints GPU nodes automatically; declaring it in the manifest makes the scheduling fence visible and reviewable

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `my-gke-cluster` | Your `GcpGkeCluster` resource name | Your cluster manifest |

## Going Further

- **MIG partitioning** (`gpuPartitionSize`) slices A100/H100 GPUs into isolated instances for many small workloads
- **GPU time-sharing** (`gpuSharingConfig`) lets several pods share one L4 for low-utilization inference
- **Compact placement** (`placementPolicy`) plus `fastSocketEnabled` + `gvnicEnabled` speed up multi-node distributed training

## Related Presets

- **01-on-demand-autoscaling** — the general-purpose primary pool
- **02-spot-cost-optimized** — Spot GPU pools combine both discounts for interruptible training

## Related Kinds

- [GcpGkeCluster](/docs/catalog/gcp/gcpgkecluster) — the control plane this pool attaches to
