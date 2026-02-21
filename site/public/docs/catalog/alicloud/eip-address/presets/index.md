---
title: "Presets"
description: "Ready-to-deploy configuration presets for EIP Address"
type: "preset-list"
componentSlug: "eip-address"
componentTitle: "EIP Address"
provider: "alicloud"
icon: "package"
order: 200
presets:
  - slug: "01-standard"
    rank: "01"
    title: "Standard EIP"
    excerpt: "This preset creates an Elastic IP Address with all defaults: 5 Mbps bandwidth, PayByTraffic metering, and BGP multi-line ISP. This is the most common configuration for development, staging, and..."
  - slug: "02-high-bandwidth"
    rank: "02"
    title: "High-Bandwidth Production EIP"
    excerpt: "This preset creates a production-grade EIP with 100 Mbps guaranteed bandwidth, PayByBandwidth metering, and BGP_PRO premium routing. Use this for internet-facing load balancers, high-traffic NAT..."
---

# EIP Address Presets

Ready-to-deploy configuration presets for EIP Address. Each preset is a complete manifest you can copy, customize, and deploy.
