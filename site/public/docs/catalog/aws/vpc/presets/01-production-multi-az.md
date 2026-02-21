---
title: "Production Multi-AZ VPC"
description: "This preset creates a production-ready VPC spanning two Availability Zones with NAT gateway, DNS support, and DNS hostnames enabled. The `/16` CIDR provides 65,536 IP addresses, enough for most..."
type: "preset"
rank: "01"
presetSlug: "01-production-multi-az"
componentSlug: "vpc"
componentTitle: "VPC"
provider: "aws"
icon: "package"
order: 1
---

# Production Multi-AZ VPC

This preset creates a production-ready VPC spanning two Availability Zones with NAT gateway, DNS support, and DNS hostnames enabled. The `/16` CIDR provides 65,536 IP addresses, enough for most production workloads. This is the standard foundation for any production AWS deployment.

## When to Use

- Production workloads requiring high availability across multiple Availability Zones
- Any deployment that needs private subnets with internet access via NAT gateway
- Foundation for EKS clusters, ECS services, RDS instances, and other AWS resources

## Key Configuration Choices

- **Multi-AZ** (`availabilityZones: [a, b]`) -- Distributes resources across two AZs for high availability and fault tolerance
- **NAT gateway enabled** (`isNatGatewayEnabled: true`) -- Allows instances in private subnets to reach the internet for updates, API calls, and dependency downloads
- **DNS hostnames** (`isDnsHostnamesEnabled: true`) -- Required for Route53 private hosted zones, service discovery, and VPC endpoints
- **DNS support** (`isDnsSupportEnabled: true`) -- Enables Amazon-provided DNS resolution within the VPC
- **/16 CIDR** (`vpcCidr: 10.0.0.0/16`) -- 65,536 IPs; standard production sizing that accommodates growth

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<aws-region>` | AWS region code (e.g., `us-east-1`, `eu-west-1`); used for the `region` field and appended with `a` and `b` for AZ suffixes | Your deployment region |

## Related Presets

- **02-development** -- Use instead for development environments where cost savings are prioritized over high availability
