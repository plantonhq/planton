---
title: "Presets"
description: "Ready-to-deploy configuration presets for Container Cluster Template"
type: "preset-list"
componentSlug: "container-cluster-template"
componentTitle: "Container Cluster Template"
provider: "openstack"
icon: "package"
order: 200
presets:
  - slug: "01-standard-kubernetes"
    rank: "01"
    title: "Standard Kubernetes Template"
    excerpt: "This preset creates a minimal Magnum cluster template for Kubernetes with Flannel networking and Google DNS. The template defines the base configuration shared by all clusters created from it -- COE..."
  - slug: "02-production-kubernetes"
    rank: "02"
    title: "Production Kubernetes Template"
    excerpt: "This preset creates a Magnum cluster template for production Kubernetes deployments with explicit network configuration, master load balancing, and floating IPs. Clusters created from this template..."
---

# Container Cluster Template Presets

Ready-to-deploy configuration presets for Container Cluster Template. Each preset is a complete manifest you can copy, customize, and deploy.
