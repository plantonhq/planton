---
title: "Cross-Region Backbone"
description: "Connects VPCs across multiple Alibaba Cloud regions to form a global private backbone. Uses `protectionLevel: REDUCED` to allow overlapping CIDR blocks between regions, which is common in..."
type: "preset"
rank: "02"
presetSlug: "02-cross-region-backbone"
componentSlug: "cen-instance"
componentTitle: "CEN Instance"
provider: "alicloud"
icon: "package"
order: 2
---

# Cross-Region Backbone

Connects VPCs across multiple Alibaba Cloud regions to form a global private backbone. Uses `protectionLevel: REDUCED` to allow overlapping CIDR blocks between regions, which is common in hub-and-spoke architectures where each region uses a standardized address plan and routing is controlled externally via route maps.

## When to Use

- You operate workloads in multiple Alibaba Cloud regions and need private, low-latency inter-region connectivity
- Your regional VPCs share overlapping CIDR ranges by design (e.g., each region uses `10.0.0.0/16`) and routing is managed through route maps
- You are building a multi-region disaster recovery or active-active architecture
- You need a global backbone as an alternative to region-to-region VPN tunnels

## Key Configuration Choices

- **REDUCED protection level** (`protectionLevel: REDUCED`) -- allows overlapping CIDR blocks between attached VPCs. Without this, CEN rejects attachments when VPC CIDRs overlap. Set this when your regions use standardized address plans and you control routing via route maps
- **API routing region** (`region`) -- CEN is a global service; this field only determines which Alibaba Cloud API endpoint handles the request. Choose any region convenient for management (typically your primary region)
- **Three cross-region attachments** -- demonstrates the multi-region pattern; each attachment specifies its own `childInstanceRegionId` because attached VPCs can be in any region worldwide
- **Tags for cost attribution** -- `purpose: global-connectivity` tag helps identify cross-region network costs in billing reports

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<api-routing-region>` | Any region for API routing (e.g., `cn-hangzhou`) | Your primary management region |
| `<your-cen-name>` | CEN instance name (2-128 characters) | Choose a descriptive name |
| `<your-team>` | Team or business unit tag | Your organizational structure |
| `<china-vpc-id>` | VPC ID in your primary China region | `AliCloudVpc` stack outputs |
| `<china-region>` | Primary China region (e.g., `cn-hangzhou`) | Your deployment topology |
| `<secondary-vpc-id>` | VPC ID in a secondary China region | `AliCloudVpc` stack outputs |
| `<secondary-region>` | Secondary China region (e.g., `cn-shanghai`) | Your deployment topology |
| `<international-vpc-id>` | VPC ID in an international region | `AliCloudVpc` stack outputs |
| `<international-region>` | International region (e.g., `ap-southeast-1`) | Your deployment topology |

## Related Presets

- **01-multi-vpc-same-region** -- Simpler setup connecting VPCs within a single region
