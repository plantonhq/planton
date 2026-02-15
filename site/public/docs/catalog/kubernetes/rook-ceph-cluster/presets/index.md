---
title: "Presets"
description: "Ready-to-deploy configuration presets for Rook Ceph Cluster"
type: "preset-list"
componentSlug: "rook-ceph-cluster"
componentTitle: "Rook Ceph Cluster"
provider: "kubernetes"
icon: "package"
order: 200
presets:
  - slug: "01-standard"
    rank: "01"
    title: "Standard Rook Ceph Cluster"
    excerpt: "This preset creates a Ceph cluster managed by the Rook operator with the toolbox and dashboard enabled. Requires the `KubernetesRookCephOperator` to be deployed first in the same namespace."
  - slug: "02-production-with-block-pool"
    rank: "02"
    title: "Production Ceph Cluster with Block Pool"
    excerpt: "This preset creates a production Ceph cluster with an explicit replicated block pool, a default Kubernetes StorageClass, and Prometheus monitoring enabled."
---

# Rook Ceph Cluster Presets

Ready-to-deploy configuration presets for Rook Ceph Cluster. Each preset is a complete manifest you can copy, customize, and deploy.
