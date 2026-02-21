---
title: "Presets"
description: "Ready-to-deploy configuration presets for Dynamic Routing Gateway"
type: "preset-list"
componentSlug: "dynamic-routing-gateway"
componentTitle: "Dynamic Routing Gateway"
provider: "oci"
icon: "package"
order: 200
presets:
  - slug: "01-single-vcn-attachment"
    rank: "01"
    title: "Single VCN Attachment"
    excerpt: "This preset creates a Dynamic Routing Gateway with a single VCN attachment. It is the most common DRG starting point, enabling inter-VCN routing and serving as the prerequisite for adding..."
  - slug: "02-hub-and-spoke"
    rank: "02"
    title: "Hub-and-Spoke"
    excerpt: "This preset creates a Dynamic Routing Gateway configured as a hub for multi-VCN networking. Two spoke VCNs are attached to a shared custom route table that imports routes from all VCN attachments via..."
---

# Dynamic Routing Gateway Presets

Ready-to-deploy configuration presets for Dynamic Routing Gateway. Each preset is a complete manifest you can copy, customize, and deploy.
