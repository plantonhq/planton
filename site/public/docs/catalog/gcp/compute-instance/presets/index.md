---
title: "Presets"
description: "Ready-to-deploy configuration presets for Compute Instance"
type: "preset-list"
componentSlug: "compute-instance"
componentTitle: "Compute Instance"
provider: "gcp"
icon: "package"
order: 200
presets:
  - slug: "01-standard-production"
    rank: "01"
    title: "Standard Production VM"
    excerpt: "This preset creates a production Compute Engine instance with an SSD boot disk, a dedicated service account, deletion protection, and no external IP (private-only networking). It follows GCP security..."
  - slug: "02-spot-development"
    rank: "02"
    title: "Spot VM for Development"
    excerpt: "This preset creates a cost-optimized Spot VM with SSH access for development and testing. Spot VMs cost 60-91% less than on-demand but can be preempted. The instance is configured to stop (not..."
---

# Compute Instance Presets

Ready-to-deploy configuration presets for Compute Instance. Each preset is a complete manifest you can copy, customize, and deploy.
