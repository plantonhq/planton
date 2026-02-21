---
title: "Presets"
description: "Ready-to-deploy configuration presets for Container Engine Cluster"
type: "preset-list"
componentSlug: "container-engine-cluster"
componentTitle: "Container Engine Cluster"
provider: "oci"
icon: "package"
order: 200
presets:
  - slug: "01-standard-production"
    rank: "01"
    title: "Standard Production OKE Cluster"
    excerpt: "This preset creates an Enhanced OKE cluster with VCN-native pod networking and a public Kubernetes API endpoint protected by a Network Security Group. It configures the service load balancer subnet..."
  - slug: "02-private-cluster"
    rank: "02"
    title: "Private OKE Cluster"
    excerpt: "This preset creates a fully private Enhanced OKE cluster with no public API endpoint, customer-managed KMS encryption for Kubernetes secrets at rest, and container image signature verification. The..."
  - slug: "03-development"
    rank: "03"
    title: "Development OKE Cluster"
    excerpt: "This preset creates a minimal OKE cluster optimized for development, testing, and experimentation. It uses the Basic cluster type with flannel overlay networking to minimize setup complexity and..."
---

# Container Engine Cluster Presets

Ready-to-deploy configuration presets for Container Engine Cluster. Each preset is a complete manifest you can copy, customize, and deploy.
