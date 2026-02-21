---
title: "Basic Site-to-Site VPN"
description: "This preset creates a VPN Gateway with a single IPsec connection to one remote site. This is the most common pattern for connecting an Alibaba Cloud VPC to an on-premises data center or branch office."
type: "preset"
rank: "01"
presetSlug: "01-basic-site-to-site"
componentSlug: "vpn-gateway"
componentTitle: "VPN Gateway"
provider: "alicloud"
icon: "package"
order: 1
---

# Basic Site-to-Site VPN

This preset creates a VPN Gateway with a single IPsec connection to one remote site. This is the most common pattern for connecting an Alibaba Cloud VPC to an on-premises data center or branch office.

## When to Use

- Single-office connectivity to Alibaba Cloud VPC
- Development or staging environments needing VPN access
- Quick setup for proof-of-concept hybrid connectivity

## Key Configuration Choices

- **10 Mbps bandwidth** -- suitable for light workloads and admin traffic; increase for production
- **PayAsYouGo billing** (default) -- no commitment, pay by the hour
- **Provider defaults for IKE/IPsec** -- IKEv2, AES encryption, SHA1 auth, DH group2
- **DPD and NAT traversal enabled** (defaults) -- compatible with most remote devices

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<alibaba-cloud-region>` | Alibaba Cloud region code (e.g., `cn-hangzhou`) | Your deployment region |
| `<your-vpc-id>` | VPC ID for the VPN Gateway | Alibaba Cloud VPC console or `AliCloudVpc` stack outputs |
| `<your-vswitch-id>` | VSwitch ID for VPN Gateway placement | Alibaba Cloud VPC console or `AliCloudVswitch` stack outputs |
| `<your-vpn-name>` | VPN Gateway name (2-128 chars) | Choose a descriptive name |
| `<connection-name>` | Connection identifier (e.g., "office-hq") | Choose a descriptive name |
| `<remote-device-public-ip>` | Public IP of the on-prem VPN router/firewall | Network admin for the remote site |
| `<vpc-cidr>` | VPC CIDR block to route through the tunnel (e.g., "10.0.0.0/8") | Your VPC configuration |
| `<remote-network-cidr>` | Remote network CIDR (e.g., "192.168.0.0/16") | Network admin for the remote site |

## Related Presets

- **02-production-multi-site** -- Use for production with multiple remote sites and strong encryption
- **03-ssl-enabled** -- Use when you also need SSL VPN for remote client access
