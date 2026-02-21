---
title: "Presets"
description: "Ready-to-deploy configuration presets for KubernetesNodePool"
type: "preset-list"
componentSlug: "kubernetesnodepool"
componentTitle: "KubernetesNodePool"
provider: "alicloud"
icon: "package"
order: 200
presets:
  - slug: "01-general-purpose-autoscaling"
    rank: "01"
    title: "General-Purpose Auto-Scaling Node Pool"
    excerpt: "This preset creates a production node pool with auto-scaling enabled, spanning three availability zones with balanced distribution. It uses general-purpose ECS g7 instances with managed lifecycle..."
  - slug: "02-fixed-size-development"
    rank: "02"
    title: "Fixed-Size Development Node Pool"
    excerpt: "This preset creates a small, fixed-size node pool for development and testing. Two nodes across two availability zones provide basic resilience without the complexity of auto-scaling or managed..."
  - slug: "03-cost-optimized-spot"
    rank: "03"
    title: "Cost-Optimized Spot Instance Node Pool"
    excerpt: "This preset creates a node pool using spot instances with price caps for significant cost savings (typically 60-90% off on-demand pricing). Four instance types across three AZs maximize spot pool..."
---

# KubernetesNodePool Presets

Ready-to-deploy configuration presets for KubernetesNodePool. Each preset is a complete manifest you can copy, customize, and deploy.
