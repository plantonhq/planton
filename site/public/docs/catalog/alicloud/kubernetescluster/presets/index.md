---
title: "Presets"
description: "Ready-to-deploy configuration presets for KubernetesCluster"
type: "preset-list"
componentSlug: "kubernetescluster"
componentTitle: "KubernetesCluster"
provider: "alicloud"
icon: "package"
order: 200
presets:
  - slug: "01-production-terway"
    rank: "01"
    title: "Production ACK Cluster with Terway ENI Networking"
    excerpt: "This preset creates a production-grade ACK Managed Kubernetes cluster using Terway ENI-based networking. Terway assigns VPC Elastic Network Interfaces directly to pods, giving each pod a VPC-routable..."
  - slug: "02-development-flannel"
    rank: "02"
    title: "Development ACK Cluster with Flannel Networking"
    excerpt: "This preset creates a minimal ACK Managed Kubernetes cluster for development and testing. It uses the free ack.standard tier with Flannel overlay networking, two availability zones, and only the..."
  - slug: "03-production-flannel"
    rank: "03"
    title: "Production ACK Cluster with Flannel Networking"
    excerpt: "This preset creates a production-grade ACK Managed Kubernetes cluster using Flannel overlay networking. It provides the same security and observability posture as the Terway production preset..."
---

# KubernetesCluster Presets

Ready-to-deploy configuration presets for KubernetesCluster. Each preset is a complete manifest you can copy, customize, and deploy.
