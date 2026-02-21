---
title: "Presets"
description: "Ready-to-deploy configuration presets for CEN Instance"
type: "preset-list"
componentSlug: "cen-instance"
componentTitle: "CEN Instance"
provider: "alicloud"
icon: "package"
order: 200
presets:
  - slug: "01-multi-vpc-same-region"
    rank: "01"
    title: "Multi-VPC Same Region"
    excerpt: "Connects two VPCs in the same Alibaba Cloud region via a Cloud Enterprise Network (CEN) instance. This is the most common CEN pattern -- isolating workloads (production, staging, shared-services)..."
  - slug: "02-cross-region-backbone"
    rank: "02"
    title: "Cross-Region Backbone"
    excerpt: "Connects VPCs across multiple Alibaba Cloud regions to form a global private backbone. Uses `protectionLevel: REDUCED` to allow overlapping CIDR blocks between regions, which is common in..."
---

# CEN Instance Presets

Ready-to-deploy configuration presets for CEN Instance. Each preset is a complete manifest you can copy, customize, and deploy.
