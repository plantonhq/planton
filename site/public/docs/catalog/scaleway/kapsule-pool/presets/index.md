---
title: "Presets"
description: "Ready-to-deploy configuration presets for Kapsule Pool"
type: "preset-list"
componentSlug: "kapsule-pool"
componentTitle: "Kapsule Pool"
provider: "scaleway"
icon: "package"
order: 200
presets:
  - slug: "01-general-purpose"
    rank: "01"
    title: "General-Purpose Node Pool"
    excerpt: "This preset creates a fixed-size node pool with GP1-XS instances (4 vCPU, 16 GB RAM) for general Kubernetes workloads. Autohealing is enabled and nodes have no public IPs. This is the standard..."
  - slug: "02-autoscaling-workers"
    rank: "02"
    title: "Autoscaling Worker Pool"
    excerpt: "This preset creates an autoscaling node pool with PRO2-M instances (4 vCPU, 16 GB RAM) that scales between 1 and 8 nodes based on workload demand. The upgrade policy uses 1 surge node for..."
---

# Kapsule Pool Presets

Ready-to-deploy configuration presets for Kapsule Pool. Each preset is a complete manifest you can copy, customize, and deploy.
