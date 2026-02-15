---
title: "Bastion Host Security Group"
description: "This preset creates a security group for a bastion (jump) host that only accepts SSH connections from trusted IP addresses. Never use `0.0.0.0/0` for bastion SSH access -- always restrict to your..."
type: "preset"
rank: "03"
presetSlug: "03-bastion"
componentSlug: "security-group"
componentTitle: "Security Group"
provider: "aws"
icon: "package"
order: 3
---

# Bastion Host Security Group

This preset creates a security group for a bastion (jump) host that only accepts SSH connections from trusted IP addresses. Never use `0.0.0.0/0` for bastion SSH access -- always restrict to your office, VPN, or specific engineer IP addresses.

## When to Use

- Bastion hosts or jump boxes used for SSH access to private instances
- Any EC2 instance that needs SSH access restricted to known IP ranges
- Temporary administrative access points in a VPC

## Key Configuration Choices

- **SSH from trusted IPs only** (`<trusted-ip-cidr>`) -- Restricts port 22 to specific CIDR blocks; use your office IP or VPN range
- **All outbound traffic** -- Permits SSH forwarding to internal instances and general internet access for updates

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<vpc-id>` | VPC ID where this security group will be created | AWS VPC console or `AwsVpc` status outputs |
| `<trusted-ip-cidr>` | CIDR block of trusted source IPs (e.g., `203.0.113.0/32` for a single IP or `10.0.0.0/8` for a VPN range) | Your network administrator or `curl ifconfig.me` for your current IP |

## Related Presets

- **01-web-tier** -- Use for internet-facing resources (ALBs, web servers)
- **02-database-tier** -- Use for databases that should only accept connections from the application tier
