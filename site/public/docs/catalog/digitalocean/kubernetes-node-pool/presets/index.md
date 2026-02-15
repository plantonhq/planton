---
title: "Presets"
description: "Ready-to-deploy configuration presets for Kubernetes Node Pool"
type: "preset-list"
componentSlug: "kubernetes-node-pool"
componentTitle: "Kubernetes Node Pool"
provider: "digitalocean"
icon: "package"
order: 200
presets:
  - slug: "01-autoscaling-production"
    rank: "01"
    title: "Autoscaling Production Node Pool"
    excerpt: "This preset creates an autoscaling node pool for a DigitalOcean Kubernetes cluster. It provisions general-purpose nodes with Kubernetes labels for workload scheduling and automatic scaling between 2..."
  - slug: "02-fixed-size"
    rank: "02"
    title: "Fixed-Size System Node Pool"
    excerpt: "This preset creates a fixed-size node pool dedicated to system workloads (ingress controllers, monitoring agents, cluster add-ons). It uses a Kubernetes taint to prevent application pods from..."
---

# Kubernetes Node Pool Presets

Ready-to-deploy configuration presets for Kubernetes Node Pool. Each preset is a complete manifest you can copy, customize, and deploy.
