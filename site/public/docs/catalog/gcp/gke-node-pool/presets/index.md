---
title: "Presets"
description: "Ready-to-deploy configuration presets for GKE Node Pool"
type: "preset-list"
componentSlug: "gke-node-pool"
componentTitle: "GKE Node Pool"
provider: "gcp"
icon: "package"
order: 200
presets:
  - slug: "01-on-demand-autoscaling"
    rank: "01"
    title: "On-Demand Autoscaling Node Pool"
    excerpt: "This preset creates a GKE node pool with on-demand (non-preemptible) VMs, SSD boot disks, and cluster autoscaler enabled. It scales between 1 and 5 nodes per zone with balanced distribution, making..."
  - slug: "02-spot-cost-optimized"
    rank: "02"
    title: "Spot VM Cost-Optimized Node Pool"
    excerpt: "This preset creates a GKE node pool using Spot VMs for significant cost savings (60-91% discount). Spot VMs can be preempted at any time, making this pool suitable for fault-tolerant batch jobs, CI..."
---

# GKE Node Pool Presets

Ready-to-deploy configuration presets for GKE Node Pool. Each preset is a complete manifest you can copy, customize, and deploy.
