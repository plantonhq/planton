---
title: "Bastion Host Security Group"
description: "This preset creates a security group for bastion (jump) hosts that serve as the single entry point into a VPC. It allows SSH inbound from a trusted network and restricts outbound to SSH and database..."
type: "preset"
rank: "03"
presetSlug: "03-bastion-host"
componentSlug: "security-group"
componentTitle: "Security Group"
provider: "alicloud"
icon: "package"
order: 3
---

# Bastion Host Security Group

This preset creates a security group for bastion (jump) hosts that serve as the single entry point into a VPC. It allows SSH inbound from a trusted network and restricts outbound to SSH and database ports within the VPC.

## When to Use

- Dedicated bastion/jump hosts for VPC access
- SSH gateways for operations teams
- Audit-friendly access points where all VPC access is funneled through a single, logged entry point

## Key Configuration Choices

- **SSH from trusted CIDR only** -- Limits SSH access to a known IP range (office network, VPN endpoint). Never use `0.0.0.0/0` for SSH in production.
- **Restricted egress** -- Only allows SSH and database connections within the VPC, preventing the bastion from becoming a pivot point for lateral movement to the internet.
- **No internet egress** -- The bastion has no general internet access. Add explicit rules if package updates are needed.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<alibaba-cloud-region>` | Alibaba Cloud region code (e.g., `cn-hangzhou`) | Your deployment region |
| `<your-vpc-id>` | VPC ID that this security group belongs to | Alibaba Cloud VPC console or `AliCloudVpc` stack outputs |
| `<your-sg-name>` | Security group name (2-128 chars) | Choose a descriptive name |
| `<your-trusted-cidr>` | Trusted source CIDR for SSH (e.g., `203.0.113.0/24`) | Your office/VPN IP range |
| `<your-vpc-cidr>` | VPC CIDR block (e.g., `10.0.0.0/8`) | Your AliCloudVpc spec.cidrBlock |

## Related Presets

- **01-web-tier** -- Use for public-facing web servers
- **02-database-tier** -- Use for database instances (the bastion connects to these)
