---
title: "Presets"
description: "Ready-to-deploy configuration presets for Harbor"
type: "preset-list"
componentSlug: "harbor"
componentTitle: "Harbor"
provider: "kubernetes"
icon: "package"
order: 200
presets:
  - slug: "01-minimal"
    rank: "01"
    title: "Minimal Harbor Container Registry"
    excerpt: "This preset deploys Harbor with default settings and ingress access. Harbor is a cloud-native container registry with vulnerability scanning, content trust, replication, and RBAC."
  - slug: "02-production-with-s3"
    rank: "02"
    title: "Production Harbor with S3 Storage"
    excerpt: "This preset deploys Harbor with S3-compatible storage for container image layers. Provides durable, scalable storage independent of the Kubernetes cluster's local disks."
---

# Harbor Presets

Ready-to-deploy configuration presets for Harbor. Each preset is a complete manifest you can copy, customize, and deploy.
