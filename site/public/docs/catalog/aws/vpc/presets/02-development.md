---
title: "Development VPC"
description: "This preset creates a minimal VPC in a single Availability Zone without a NAT gateway. This reduces costs significantly (NAT gateway charges ~$32/month + data transfer) while still providing a..."
type: "preset"
rank: "02"
presetSlug: "02-development"
componentSlug: "vpc"
componentTitle: "VPC"
provider: "aws"
icon: "package"
order: 2
---

# Development VPC

This preset creates a minimal VPC in a single Availability Zone without a NAT gateway. This reduces costs significantly (NAT gateway charges ~$32/month + data transfer) while still providing a functional networking environment for development and testing workloads.

## When to Use

- Development and testing environments where high availability is not required
- Cost-sensitive workloads that do not need private subnet internet access
- Quick sandbox environments for prototyping

## Key Configuration Choices

- **Single AZ** (`availabilityZones: [a]`) -- Minimal footprint; no cross-AZ redundancy
- **No NAT gateway** (`isNatGatewayEnabled: false`) -- Saves ~$32/month; instances in private subnets cannot reach the internet (use public subnets or VPC endpoints if needed)
- **DNS enabled** -- DNS hostnames and support remain on for service discovery compatibility

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<aws-region>` | AWS region code (e.g., `us-east-1`); appended with `a` for AZ suffix | Your deployment region |

## Related Presets

- **01-production-multi-az** -- Use instead for production deployments requiring multi-AZ redundancy and NAT gateway
