# On-Demand Autoscaling Pool

This preset creates the workhorse node pool most clusters run first: on-demand VMs, per-zone autoscaling, surge upgrades with zero capacity dip, a dedicated node service account, and secure boot on.

## When to Use

- The primary pool for production services that must always have capacity
- Steady-state workloads where Spot preemption is unacceptable
- The first pool added after creating a Standard GKE cluster

## Key Configuration Choices

- **Cluster by reference** — `clusterName` and `location` resolve from the cluster's outputs, so the pool addresses its parent exactly as GKE named it
- **Per-zone autoscaling 1-5** — a regional cluster in three zones runs 3-15 nodes; `BALANCED` keeps zones evenly sized
- **`n2-standard-4` on `pd-balanced`** — a real general-purpose shape; the `e2-medium` default is sandbox-sized
- **Surge upgrades (`maxSurge: 1, maxUnavailable: 0`)** — upgrades never reduce serving capacity
- **Dedicated service account + secure boot** — least-privilege node identity; workload permissions come from Workload Identity

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `my-gke-cluster` | Your `GcpGkeCluster` resource name | Your cluster manifest |
| `gke-nodes` | Your `GcpServiceAccount` resource name | Your service account manifest |

## Related Presets

- **02-spot-cost-optimized** — scale-to-zero Spot capacity for fault-tolerant batch
- **03-gpu-accelerated** — GPU nodes for ML workloads

## Related Kinds

- [GcpGkeCluster](/docs/catalog/gcp/gcpgkecluster) — the control plane this pool attaches to
- [GcpServiceAccount](/docs/catalog/gcp/gcpserviceaccount) — the node identity
