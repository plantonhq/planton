---
title: "Presets"
description: "Ready-to-deploy configuration presets for Container Engine Node Pool"
type: "preset-list"
componentSlug: "container-engine-node-pool"
componentTitle: "Container Engine Node Pool"
provider: "oci"
icon: "package"
order: 200
presets:
  - slug: "01-standard-production"
    rank: "01"
    title: "Standard Production OKE Node Pool"
    excerpt: "This preset creates a general-purpose OKE node pool with E4.Flex compute shapes, VCN-native pod networking, and zero-downtime rolling upgrade settings. It distributes 3 nodes across 3 availability..."
  - slug: "02-hardened-encrypted"
    rank: "02"
    title: "Hardened Encrypted OKE Node Pool"
    excerpt: "This preset creates a security-hardened OKE node pool with customer-managed KMS encryption for boot volumes, in-transit encryption for paravirtualized volume attachments, and explicit fault domain..."
  - slug: "03-preemptible-dev"
    rank: "03"
    title: "Preemptible Dev OKE Node Pool"
    excerpt: "This preset creates a cost-optimized OKE node pool using preemptible (spot) instances for development, testing, and experimentation. Preemptible nodes use the same shapes as on-demand nodes but are..."
---

# Container Engine Node Pool Presets

Ready-to-deploy configuration presets for Container Engine Node Pool. Each preset is a complete manifest you can copy, customize, and deploy.
