---
title: "Presets"
description: "Ready-to-deploy configuration presets for EKS Node Group"
type: "preset-list"
componentSlug: "eks-node-group"
componentTitle: "EKS Node Group"
provider: "aws"
icon: "package"
order: 200
presets:
  - slug: "01-on-demand-general"
    rank: "01"
    title: "On-Demand General Purpose Node Group"
    excerpt: "This preset creates an EKS managed node group using on-demand `t3.medium` instances across two Availability Zones. The group scales between 2 and 5 nodes with 100 GiB root disks. This is the standard..."
  - slug: "02-spot-cost-optimized"
    rank: "02"
    title: "Spot Cost-Optimized Node Group"
    excerpt: "This preset creates an EKS managed node group using Spot instances for up to 70% cost savings compared to on-demand. The `node-lifecycle: spot` label enables workload targeting via node selectors or..."
---

# EKS Node Group Presets

Ready-to-deploy configuration presets for EKS Node Group. Each preset is a complete manifest you can copy, customize, and deploy.
