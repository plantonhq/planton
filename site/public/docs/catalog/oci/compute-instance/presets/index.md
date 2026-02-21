---
title: "Presets"
description: "Ready-to-deploy configuration presets for Compute Instance"
type: "preset-list"
componentSlug: "compute-instance"
componentTitle: "Compute Instance"
provider: "oci"
icon: "package"
order: 200
presets:
  - slug: "01-general-purpose-flex"
    rank: "01"
    title: "General-Purpose Flex Instance"
    excerpt: "This preset creates a general-purpose OCI compute instance using the VM.Standard.E4.Flex shape (AMD EPYC). It configures 1 OCPU with 16 GiB of memory, a 50 GiB boot volume, and a public IP for direct..."
  - slug: "02-private-backend"
    rank: "02"
    title: "Private Backend Instance"
    excerpt: "This preset creates a production-hardened OCI compute instance in a private subnet with no public IP. It enables in-transit encryption, disables legacy IMDS endpoints, configures live migration for..."
  - slug: "03-preemptible-dev"
    rank: "03"
    title: "Preemptible Dev Instance"
    excerpt: "This preset creates a cost-optimized preemptible (spot-like) OCI compute instance for development, testing, and CI workloads. Preemptible instances use the same shapes and images as on-demand..."
---

# Compute Instance Presets

Ready-to-deploy configuration presets for Compute Instance. Each preset is a complete manifest you can copy, customize, and deploy.
