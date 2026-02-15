---
title: "Presets"
description: "Ready-to-deploy configuration presets for Kubernetes Cluster"
type: "preset-list"
componentSlug: "kubernetes-cluster"
componentTitle: "Kubernetes Cluster"
provider: "civo"
icon: "package"
order: 200
presets:
  - slug: "01-production-ha"
    rank: "01"
    title: "Production HA Kubernetes Cluster"
    excerpt: "This preset creates a highly available Civo Kubernetes (K3s) cluster with 3 worker nodes, automatic patch upgrades, and VPC networking. This is the most common production configuration, providing..."
  - slug: "02-development"
    rank: "02"
    title: "Development Kubernetes Cluster"
    excerpt: "This preset creates a minimal, cost-effective single-node Kubernetes cluster for development and testing. No HA, no auto-upgrade, smallest node size. Ideal for local development, CI pipelines, and..."
---

# Kubernetes Cluster Presets

Ready-to-deploy configuration presets for Kubernetes Cluster. Each preset is a complete manifest you can copy, customize, and deploy.
