---
title: "Single VSwitch NAT Gateway"
description: "This preset creates a NAT Gateway that provides outbound internet access for a single VSwitch. This is the most common pattern for development or simple production environments."
type: "preset"
rank: "01"
presetSlug: "01-single-vswitch"
componentSlug: "nat-gateway"
componentTitle: "NAT Gateway"
provider: "alicloud"
icon: "package"
order: 1
---

# Single VSwitch NAT Gateway

This preset creates a NAT Gateway that provides outbound internet access for a single VSwitch. This is the most common pattern for development or simple production environments.

## When to Use

- Development or staging environments with a single application VSwitch
- Simple architectures where all workloads share one subnet
- Quick setup for proof-of-concept deployments

## Key Configuration Choices

- **Enhanced NAT type** (default) -- modern NAT gateway with higher performance and VSwitch placement support
- **PayByLcu billing** (default) -- pay only for actual capacity units consumed, no fixed specification needed
- **Single SNAT entry** -- all traffic from the specified VSwitch exits through the EIP

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<alibaba-cloud-region>` | Alibaba Cloud region code (e.g., `cn-hangzhou`) | Your deployment region |
| `<your-vpc-id>` | VPC ID the NAT Gateway belongs to | Alibaba Cloud VPC console or `AliCloudVpc` stack outputs |
| `<your-nat-vswitch-id>` | VSwitch ID for NAT Gateway placement | Alibaba Cloud VPC console or `AliCloudVswitch` stack outputs |
| `<your-eip-id>` | EIP allocation ID to associate with the NAT Gateway | Alibaba Cloud EIP console or `AliCloudEipAddress` stack outputs |
| `<your-app-vswitch-id>` | VSwitch ID whose traffic needs internet access | Alibaba Cloud VPC console or `AliCloudVswitch` stack outputs |
| `<your-nat-name>` | NAT Gateway name (2-128 chars) | Choose a descriptive name |

## Related Presets

- **02-multi-az-production** -- Use for multi-AZ production environments with multiple VSwitches
- **03-cidr-based-snat** -- Use when you need fine-grained CIDR-level NAT control
