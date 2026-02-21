---
title: "Presets"
description: "Ready-to-deploy configuration presets for Subnet"
type: "preset-list"
componentSlug: "subnet"
componentTitle: "Subnet"
provider: "oci"
icon: "package"
order: 200
presets:
  - slug: "01-private"
    rank: "01"
    title: "Private Subnet"
    excerpt: "This preset creates a private subnet with inline route rules that direct internet-bound traffic through a NAT Gateway and OCI service traffic through a Service Gateway. No VNICs in this subnet can..."
  - slug: "02-public"
    rank: "02"
    title: "Public Subnet"
    excerpt: "This preset creates a public subnet with an inline route rule that sends all traffic through an Internet Gateway. VNICs in this subnet can be assigned public IP addresses and receive inbound internet..."
  - slug: "03-development"
    rank: "03"
    title: "Development Subnet"
    excerpt: "This preset creates a minimal public subnet with no custom route rules. The subnet inherits the VCN's default route table, which is the simplest possible configuration. This is suitable for..."
---

# Subnet Presets

Ready-to-deploy configuration presets for Subnet. Each preset is a complete manifest you can copy, customize, and deploy.
