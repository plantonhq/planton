---
title: "Presets"
description: "Ready-to-deploy configuration presets for VSwitch"
type: "preset-list"
componentSlug: "vswitch"
componentTitle: "VSwitch"
provider: "alicloud"
icon: "package"
order: 200
presets:
  - slug: "01-dev-single-zone"
    rank: "01"
    title: "Development Single-Zone VSwitch"
    excerpt: "This preset creates a minimal VSwitch in a single availability zone using a small /24 CIDR from the 192.168.x.x range. It omits tags and IPv6 configuration, keeping the setup simple and inexpensive..."
  - slug: "02-prod-app-tier"
    rank: "02"
    title: "Production Application Tier VSwitch"
    excerpt: "This preset creates a VSwitch sized for production application workloads using a /20 CIDR from the 10.x.x.x range. The larger address space (4,096 IPs) accommodates Kubernetes node pools, ECS fleets,..."
  - slug: "03-ipv6-enabled"
    rank: "03"
    title: "IPv6-Enabled Dual-Stack VSwitch"
    excerpt: "This preset creates a VSwitch with both IPv4 and IPv6 addressing enabled. The parent VPC must have IPv6 enabled (`enableIpv6: true` in AliCloudVpc) for IPv6 allocation to succeed. A /24 IPv4 CIDR is..."
---

# VSwitch Presets

Ready-to-deploy configuration presets for VSwitch. Each preset is a complete manifest you can copy, customize, and deploy.
