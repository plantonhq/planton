---
title: "Presets"
description: "Ready-to-deploy configuration presets for GKE Cluster"
type: "preset-list"
componentSlug: "gke-cluster"
componentTitle: "GKE Cluster"
provider: "gcp"
icon: "package"
order: 200
presets:
  - slug: "01-private-standard"
    rank: "01"
    title: "Private GKE Cluster -- Standard"
    excerpt: "This preset creates a private GKE cluster with no public node IPs, the REGULAR release channel, Workload Identity enabled, and network policy enforcement. Private clusters are the GCP-recommended..."
  - slug: "02-private-rapid"
    rank: "02"
    title: "Private GKE Cluster -- Rapid Channel"
    excerpt: "This preset creates a private GKE cluster on the RAPID release channel for development or staging environments that want early access to the latest Kubernetes versions and GKE features. Otherwise..."
---

# GKE Cluster Presets

Ready-to-deploy configuration presets for GKE Cluster. Each preset is a complete manifest you can copy, customize, and deploy.
