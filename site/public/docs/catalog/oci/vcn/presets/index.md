---
title: "Presets"
description: "Ready-to-deploy configuration presets for VCN"
type: "preset-list"
componentSlug: "vcn"
componentTitle: "VCN"
provider: "oci"
icon: "package"
order: 200
presets:
  - slug: "01-standard-public-private"
    rank: "01"
    title: "Standard Public-Private VCN"
    excerpt: "This preset creates a production VCN with all three gateways enabled: Internet Gateway, NAT Gateway, and Service Gateway. This is the standard OCI networking topology for workloads that require both..."
  - slug: "02-private-only"
    rank: "02"
    title: "Private-Only VCN"
    excerpt: "This preset creates a security-hardened VCN with no Internet Gateway. NAT Gateway and Service Gateway provide outbound connectivity, but no resources in this VCN are directly reachable from the..."
  - slug: "03-development"
    rank: "03"
    title: "Development VCN"
    excerpt: "This preset creates a minimal-cost VCN with only an Internet Gateway. NAT Gateway and Service Gateway are omitted to avoid their hourly charges. This is the simplest VCN configuration, suitable for..."
---

# VCN Presets

Ready-to-deploy configuration presets for VCN. Each preset is a complete manifest you can copy, customize, and deploy.
