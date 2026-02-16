---
title: "Presets"
description: "Ready-to-deploy configuration presets for Transit Gateway"
type: "preset-list"
componentSlug: "transit-gateway"
componentTitle: "Transit Gateway"
provider: "aws"
icon: "package"
order: 200
presets:
  - slug: "01-multi-vpc-hub"
    rank: "01"
    title: "Multi-VPC Hub"
    excerpt: "Production Transit Gateway connecting application and shared-services VPCs with full-mesh routing. This is the most common Transit Gateway pattern, replacing complex VPC peering meshes with a..."
  - slug: "02-single-vpc-development"
    rank: "02"
    title: "Single VPC Development"
    excerpt: "Minimal Transit Gateway with a single VPC attachment for development or testing. This is the simplest possible TGW setup, useful for validating connectivity patterns before scaling to production."
  - slug: "03-hub-and-spoke-firewall"
    rank: "03"
    title: "Hub-and-Spoke Firewall"
    excerpt: "Transit Gateway with a centralized inspection VPC running a virtual firewall appliance (e.g., Palo Alto, Fortinet, AWS Network Firewall). The inspection VPC uses appliance mode to ensure symmetric..."
---

# Transit Gateway Presets

Ready-to-deploy configuration presets for Transit Gateway. Each preset is a complete manifest you can copy, customize, and deploy.
