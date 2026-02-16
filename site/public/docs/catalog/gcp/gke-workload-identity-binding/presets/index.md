---
title: "Presets"
description: "Ready-to-deploy configuration presets for GKE Workload Identity Binding"
type: "preset-list"
componentSlug: "gke-workload-identity-binding"
componentTitle: "GKE Workload Identity Binding"
provider: "gcp"
icon: "package"
order: 200
presets:
  - slug: "01-standard"
    rank: "01"
    title: "Standard Workload Identity Binding"
    excerpt: "This preset creates a Workload Identity binding that allows a Kubernetes ServiceAccount (KSA) to impersonate a Google ServiceAccount (GSA). This is the GCP-recommended way for GKE pods to..."
---

# GKE Workload Identity Binding Presets

Ready-to-deploy configuration presets for GKE Workload Identity Binding. Each preset is a complete manifest you can copy, customize, and deploy.
