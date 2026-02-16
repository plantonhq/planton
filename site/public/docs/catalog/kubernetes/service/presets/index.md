---
title: "Presets"
description: "Ready-to-deploy configuration presets for Service"
type: "preset-list"
componentSlug: "service"
componentTitle: "Service"
provider: "kubernetes"
icon: "package"
order: 200
presets:
  - slug: "01-cluster-ip"
    rank: "01"
    title: "ClusterIP Service"
    excerpt: "This preset creates a standard ClusterIP service that exposes a deployment internally within the cluster. The most common Kubernetes Service type for inter-service communication."
  - slug: "02-load-balancer"
    rank: "02"
    title: "LoadBalancer Service"
    excerpt: "This preset creates a LoadBalancer service that provisions a cloud load balancer with a public IP. Suitable for services that need direct external access without an ingress controller."
  - slug: "03-headless"
    rank: "03"
    title: "Headless Service"
    excerpt: "This preset creates a headless service (ClusterIP: None) for direct pod-to-pod DNS resolution. Essential for StatefulSets and any application that needs to discover individual pod IPs rather than a..."
---

# Service Presets

Ready-to-deploy configuration presets for Service. Each preset is a complete manifest you can copy, customize, and deploy.
