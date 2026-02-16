---
title: "Presets"
description: "Ready-to-deploy configuration presets for AKS Cluster"
type: "preset-list"
componentSlug: "aks-cluster"
componentTitle: "AKS Cluster"
provider: "azure"
icon: "package"
order: 200
presets:
  - slug: "01-standard"
    rank: "01"
    title: "Standard Production AKS Cluster"
    excerpt: "This preset deploys a production-ready AKS cluster with a public API endpoint, Azure CNI Overlay networking, a 3-zone system node pool, a general-purpose user node pool with autoscaling, and all..."
  - slug: "02-private"
    rank: "02"
    title: "Private AKS Cluster"
    excerpt: "This preset deploys a private AKS cluster with no public API server endpoint. The Kubernetes API is accessible only from within the VNet or via peered networks (VPN, ExpressRoute). All other..."
---

# AKS Cluster Presets

Ready-to-deploy configuration presets for AKS Cluster. Each preset is a complete manifest you can copy, customize, and deploy.
