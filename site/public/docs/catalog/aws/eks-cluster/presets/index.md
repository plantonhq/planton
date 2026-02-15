---
title: "Presets"
description: "Ready-to-deploy configuration presets for EKS Cluster"
type: "preset-list"
componentSlug: "eks-cluster"
componentTitle: "EKS Cluster"
provider: "aws"
icon: "package"
order: 200
presets:
  - slug: "01-standard"
    rank: "01"
    title: "Standard EKS Cluster"
    excerpt: "This preset creates an EKS cluster with a publicly accessible API endpoint and control plane logging enabled. The cluster spans two Availability Zones for high availability. This is the most common..."
  - slug: "02-private-endpoint"
    rank: "02"
    title: "Private Endpoint EKS Cluster"
    excerpt: "This preset creates an EKS cluster with the API server endpoint restricted to VPC-internal access only. The Kubernetes API is not reachable from the public internet. Use this for security-sensitive..."
---

# EKS Cluster Presets

Ready-to-deploy configuration presets for EKS Cluster. Each preset is a complete manifest you can copy, customize, and deploy.
