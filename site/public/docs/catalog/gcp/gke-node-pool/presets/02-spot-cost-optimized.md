---
title: "Spot VM Cost-Optimized Node Pool"
description: "This preset creates a GKE node pool using Spot VMs for significant cost savings (60-91% discount). Spot VMs can be preempted at any time, making this pool suitable for fault-tolerant batch jobs, CI..."
type: "preset"
rank: "02"
presetSlug: "02-spot-cost-optimized"
componentSlug: "gke-node-pool"
componentTitle: "GKE Node Pool"
provider: "gcp"
icon: "package"
order: 2
---

# Spot VM Cost-Optimized Node Pool

This preset creates a GKE node pool using Spot VMs for significant cost savings (60-91% discount). Spot VMs can be preempted at any time, making this pool suitable for fault-tolerant batch jobs, CI runners, or development workloads. The pool can scale to zero when idle.

## When to Use

- Batch processing, CI/CD runners, or other fault-tolerant workloads
- Development and testing environments where cost matters more than availability
- Supplementary capacity alongside an on-demand pool for burst workloads

## Key Configuration Choices

- **Spot VMs** (`spot: true`) -- 60-91% cost savings; GCP can reclaim these VMs at any time
- **e2-standard-2** (`machineType`) -- smaller instances to maximize cost efficiency
- **Standard disk** (`diskType: pd-standard`) -- lower-cost disk for non-latency-sensitive workloads
- **Scale-to-zero** (`minNodes: 0`) -- no cost when no pods are scheduled on spot nodes
- **Node label** (`node-type: spot`) -- enables targeted scheduling via `nodeSelector` or `nodeAffinity`
- **Max 10 nodes** -- higher ceiling to absorb burst workloads

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<gcp-project-id>` | GCP project ID hosting the GKE cluster | `GcpProject` outputs |
| `<gke-cluster-name>` | Name of the parent GKE cluster | `GcpGkeCluster` metadata name |
| `<gcp-region>` | Location of the GKE cluster (e.g., `us-central1`) | `GcpGkeCluster` spec location |
| `<your-node-pool-name>` | Name for this node pool (1-40 chars, lowercase) | Choose a descriptive name (e.g., `spot-pool`) |

## Related Presets

- **01-on-demand-autoscaling** -- Use for production workloads that need guaranteed availability
