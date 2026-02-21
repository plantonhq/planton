---
title: "Multi-AZ Production NAT Gateway"
description: "This preset creates a production-grade NAT Gateway serving multiple VSwitches across availability zones. Deletion protection is enabled to prevent accidental removal. Resource tags support cost..."
type: "preset"
rank: "02"
presetSlug: "02-multi-az-production"
componentSlug: "nat-gateway"
componentTitle: "NAT Gateway"
provider: "alicloud"
icon: "package"
order: 2
---

# Multi-AZ Production NAT Gateway

This preset creates a production-grade NAT Gateway serving multiple VSwitches across availability zones. Deletion protection is enabled to prevent accidental removal. Resource tags support cost tracking and operational visibility.

## When to Use

- Production environments with workloads spread across multiple AZs
- Kubernetes (ACK) clusters with worker nodes in separate VSwitches
- Any multi-tier architecture where multiple subnets need outbound internet access

## Key Configuration Choices

- **Deletion protection enabled** (`deletionProtection: true`) -- prevents accidental deletion via console or API
- **Multiple SNAT entries** -- one per VSwitch/AZ for clear traffic attribution and independent management
- **Resource tags** -- team and cost-center tags for operational visibility and cost allocation

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<alibaba-cloud-region>` | Alibaba Cloud region code (e.g., `cn-hangzhou`) | Your deployment region |
| `<your-vpc-id>` | VPC ID the NAT Gateway belongs to | Alibaba Cloud VPC console or `AliCloudVpc` stack outputs |
| `<your-nat-vswitch-id>` | VSwitch ID for NAT Gateway placement | Alibaba Cloud VPC console or `AliCloudVswitch` stack outputs |
| `<your-eip-id>` | EIP allocation ID to associate with the NAT Gateway | Alibaba Cloud EIP console or `AliCloudEipAddress` stack outputs |
| `<your-app-vswitch-id-zone-a>` | Application VSwitch ID in availability zone A | `AliCloudVswitch` stack outputs for zone A |
| `<your-app-vswitch-id-zone-b>` | Application VSwitch ID in availability zone B | `AliCloudVswitch` stack outputs for zone B |
| `<your-nat-name>` | NAT Gateway name (2-128 chars) | Choose a descriptive name |
| `<your-team>` | Team name for cost-tracking tag | Your team |
| `<your-cost-center>` | Cost center for billing tag | Your cost center |

## Related Presets

- **01-single-vswitch** -- Simpler setup for dev/staging with one VSwitch
- **03-cidr-based-snat** -- Use when you need CIDR-level granularity instead of VSwitch-based SNAT
