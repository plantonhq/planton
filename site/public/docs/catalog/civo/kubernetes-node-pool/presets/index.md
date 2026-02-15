---
title: "Presets"
description: "Ready-to-deploy configuration presets for Kubernetes Node Pool"
type: "preset-list"
componentSlug: "kubernetes-node-pool"
componentTitle: "Kubernetes Node Pool"
provider: "civo"
icon: "package"
order: 200
presets:
  - slug: "01-autoscaling"
    rank: "01"
    title: "Autoscaling Node Pool"
    excerpt: "This preset creates an autoscaling node pool that dynamically adjusts between 2 and 5 nodes based on workload demand. Starting at 2 nodes ensures baseline capacity while allowing the cluster..."
  - slug: "02-fixed-size"
    rank: "02"
    title: "Fixed-Size Node Pool"
    excerpt: "This preset creates a static 3-node pool with no autoscaling. Suitable for workloads with predictable, steady resource requirements where the overhead of autoscaler decisions is unnecessary."
---

# Kubernetes Node Pool Presets

Ready-to-deploy configuration presets for Kubernetes Node Pool. Each preset is a complete manifest you can copy, customize, and deploy.
