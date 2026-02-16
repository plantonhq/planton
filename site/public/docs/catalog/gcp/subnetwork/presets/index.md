---
title: "Presets"
description: "Ready-to-deploy configuration presets for Subnetwork"
type: "preset-list"
componentSlug: "subnetwork"
componentTitle: "Subnetwork"
provider: "gcp"
icon: "package"
order: 200
presets:
  - slug: "01-gke-ready"
    rank: "01"
    title: "GKE-Ready Subnet"
    excerpt: "This preset creates a subnet with secondary IP ranges for GKE pod and service CIDRs, plus Private Google Access enabled. This is the standard subnet configuration required before creating a..."
  - slug: "02-general-purpose"
    rank: "02"
    title: "General-Purpose Subnet"
    excerpt: "This preset creates a simple subnet for Compute Engine VMs, Cloud Run with VPC access, or other non-GKE workloads. No secondary IP ranges are defined since alias IPs are not needed outside of GKE."
---

# Subnetwork Presets

Ready-to-deploy configuration presets for Subnetwork. Each preset is a complete manifest you can copy, customize, and deploy.
