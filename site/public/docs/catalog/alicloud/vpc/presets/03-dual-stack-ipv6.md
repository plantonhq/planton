---
title: "Dual-Stack IPv6 VPC"
description: "This preset creates a VPC with dual-stack networking enabled. When `enableIpv6` is set to `true`, Alibaba Cloud allocates a /56 IPv6 CIDR block to the VPC in addition to the IPv4 CIDR. VSwitches..."
type: "preset"
rank: "03"
presetSlug: "03-dual-stack-ipv6"
componentSlug: "vpc"
componentTitle: "VPC"
provider: "alicloud"
icon: "package"
order: 3
---

# Dual-Stack IPv6 VPC

This preset creates a VPC with dual-stack networking enabled. When `enableIpv6` is set to `true`, Alibaba Cloud allocates a /56 IPv6 CIDR block to the VPC in addition to the IPv4 CIDR. VSwitches created within this VPC can then be assigned IPv6 subnets, enabling workloads to communicate over both IPv4 and IPv6.

## When to Use

- Applications that serve IPv6 clients or need IPv6 egress connectivity
- China-region deployments where IPv6 adoption is mandated or incentivized by regulatory policy
- Modern microservice architectures planning for IPv6-native communication
- Prerequisite for creating IPv6-enabled VSwitches (see AliCloudVswitch `03-ipv6-enabled` preset)

## Key Configuration Choices

- **IPv6 enabled** (`enableIpv6: true`) -- Alibaba Cloud allocates a /56 IPv6 CIDR block automatically. This cannot be changed after VPC creation; you must create a new VPC to add or remove IPv6 support.
- **172.16.x range** (`cidrBlock: 172.16.0.0/12`) -- Uses the third RFC 1918 range to avoid overlapping with production VPCs (10.x) or development VPCs (192.168.x). The /12 mask provides over 1 million IPs, suitable for large-scale dual-stack deployments with many VSwitches.
- **Dual-stack tag** (`networkType: dual-stack`) -- Makes it easy to identify IPv6-enabled VPCs when filtering resources in the console or via API. Useful for network operations teams managing mixed IPv4-only and dual-stack environments.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|-------------|-------------|---------------|
| `<alibaba-cloud-region>` | Alibaba Cloud region code (e.g., `cn-hangzhou`, `cn-shanghai`, `ap-southeast-1`) | Your deployment region strategy; verify IPv6 availability in the target region |
| `<your-vpc-name>` | VPC name (1-128 chars; cannot start with `http://` or `https://`) | Choose a name following your naming convention (e.g., `ipv6-platform-vpc`) |

## Related Presets

- **01-standard-production** -- Use instead for IPv4-only production environments
- **02-development** -- Use instead for minimal IPv4-only development environments
