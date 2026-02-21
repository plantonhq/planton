---
title: "Production Multi-Site VPN"
description: "This preset creates a VPN Gateway connecting to two remote sites with production-grade encryption (AES-256 + SHA-256 + DH group14) and health check monitoring on both tunnels."
type: "preset"
rank: "02"
presetSlug: "02-production-multi-site"
componentSlug: "vpn-gateway"
componentTitle: "VPN Gateway"
provider: "alicloud"
icon: "package"
order: 2
---

# Production Multi-Site VPN

This preset creates a VPN Gateway connecting to two remote sites with production-grade encryption (AES-256 + SHA-256 + DH group14) and health check monitoring on both tunnels.

## When to Use

- Production environments with multiple data centers or branch offices
- High-security requirements mandating strong encryption algorithms
- Environments needing health monitoring for automatic failure detection

## Key Configuration Choices

- **100 Mbps bandwidth** -- suitable for production workloads; adjust based on traffic volume
- **AES-256 encryption** -- strongest available symmetric encryption
- **SHA-256 authentication** -- modern hash algorithm, stronger than default SHA-1
- **DH Group 14** -- 2048-bit MODP for Perfect Forward Secrecy (stronger than default group2)
- **Health checks enabled** -- probes detect tunnel failures for operational alerting
- **Multiple connections** -- each connecting to a different remote site

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<alibaba-cloud-region>` | Alibaba Cloud region code | Your deployment region |
| `<your-vpc-id>` | VPC ID | `AliCloudVpc` stack outputs |
| `<your-vswitch-id>` | VSwitch ID for VPN Gateway | `AliCloudVswitch` stack outputs |
| `<your-vpn-name>` | VPN Gateway name (2-128 chars) | Choose a descriptive name |
| `<your-org>` | Organization name | Your organization |
| `<your-team>` | Owning team | Your team name |
| `<cost-center>` | Cost center tag | Your billing structure |
| `<primary-site-name>` | Primary site connection name | e.g., "datacenter-primary" |
| `<primary-site-ip>` | Primary site VPN device public IP | Network admin |
| `<secondary-site-name>` | Secondary site connection name | e.g., "datacenter-dr" |
| `<secondary-site-ip>` | Secondary site VPN device public IP | Network admin |
| `<vpc-cidr-1>`, `<vpc-cidr-2>` | VPC CIDR blocks | Your VPC configuration |
| `<primary-remote-cidr>`, `<secondary-remote-cidr>` | Remote network CIDRs | Network admin |
| `<pre-shared-key-1>`, `<pre-shared-key-2>` | Pre-shared keys (1-100 chars) | Generate securely |
| `<local-probe-ip>` | VPC IP for health probes | An IP routable within your VPC |
| `<remote-probe-ip>`, `<secondary-remote-probe-ip>` | Remote IPs for health probes | Network admin |

## Related Presets

- **01-basic-site-to-site** -- Use for simple single-site setups
- **03-ssl-enabled** -- Use when you also need SSL VPN for remote client access
