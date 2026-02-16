---
title: "Presets"
description: "Ready-to-deploy configuration presets for Rook Ceph Operator"
type: "preset-list"
componentSlug: "rook-ceph-operator"
componentTitle: "Rook Ceph Operator"
provider: "kubernetes"
icon: "package"
order: 200
presets:
  - slug: "01-standard"
    rank: "01"
    title: "Standard Rook Ceph Operator"
    excerpt: "This preset deploys the Rook Ceph Operator with recommended default resources and default CSI settings. Rook enables Ceph distributed storage on Kubernetes, providing block, file, and object storage...."
  - slug: "02-production-with-csi"
    rank: "02"
    title: "Production Rook Ceph Operator with Explicit CSI Configuration"
    excerpt: "This preset deploys the Rook Ceph Operator with all CSI driver options explicitly configured. Use this when you need full control over which storage drivers are enabled and want every setting..."
---

# Rook Ceph Operator Presets

Ready-to-deploy configuration presets for Rook Ceph Operator. Each preset is a complete manifest you can copy, customize, and deploy.
