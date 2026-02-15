---
title: "Presets"
description: "Ready-to-deploy configuration presets for Kapsule Cluster"
type: "preset-list"
componentSlug: "kapsule-cluster"
componentTitle: "Kapsule Cluster"
provider: "scaleway"
icon: "package"
order: 200
presets:
  - slug: "01-dev-minimal"
    rank: "01"
    title: "Development Kubernetes Cluster"
    excerpt: "This preset creates a minimal Scaleway Kapsule cluster with a shared (mutualized) control plane and a small 2-node default pool using DEV1-M instances. It is the fastest path to a working Kubernetes..."
  - slug: "02-production-autoscaling"
    rank: "02"
    title: "Production Kubernetes Cluster with Autoscaling"
    excerpt: "This preset creates a production-grade Scaleway Kapsule cluster with autoscaling, automatic patch upgrades on Sunday mornings, private-only nodes, and autohealing. The default pool uses PRO2-S..."
---

# Kapsule Cluster Presets

Ready-to-deploy configuration presets for Kapsule Cluster. Each preset is a complete manifest you can copy, customize, and deploy.
