---
title: "SSL VPN Enabled Gateway"
description: "This preset creates a VPN Gateway with SSL VPN enabled for remote client access, alongside a site-to-site IPsec connection. SSL VPN allows individual users (developers, admins) to connect to the VPC..."
type: "preset"
rank: "03"
presetSlug: "03-ssl-enabled"
componentSlug: "vpn-gateway"
componentTitle: "VPN Gateway"
provider: "alicloud"
icon: "package"
order: 3
---

# SSL VPN Enabled Gateway

This preset creates a VPN Gateway with SSL VPN enabled for remote client access, alongside a site-to-site IPsec connection. SSL VPN allows individual users (developers, admins) to connect to the VPC from their laptops.

## When to Use

- Teams need both site-to-site VPN and remote-access VPN on the same gateway
- Developers need direct VPC access from their workstations
- Operations teams need emergency access to VPC resources from anywhere

## Key Configuration Choices

- **50 Mbps bandwidth** -- balanced for both site-to-site and SSL VPN traffic
- **SSL VPN enabled** with 50 concurrent connections -- adjust based on team size
- **Site-to-site connection included** -- combines both VPN modes on one gateway
- **PayAsYouGo billing** (default) -- SSL VPN connections incur additional charges

## Important Notes

- This preset creates the VPN Gateway with SSL VPN capability, but you still need to create an **SSL VPN Server** (`alicloud_ssl_vpn_server`) and **Client Certificates** (`alicloud_ssl_vpn_client_cert`) separately to actually enable remote client access.
- The `sslVpnInternetIp` output provides the IP that SSL clients will connect to.
- SSL VPN connections are billed separately based on the `sslConnections` count.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<alibaba-cloud-region>` | Alibaba Cloud region code | Your deployment region |
| `<your-vpc-id>` | VPC ID | `AliCloudVpc` stack outputs |
| `<your-vswitch-id>` | VSwitch ID | `AliCloudVswitch` stack outputs |
| `<your-vpn-name>` | VPN Gateway name (2-128 chars) | Choose a descriptive name |
| `<site-connection-name>` | Connection name (e.g., "office-hq") | Choose a descriptive name |
| `<remote-device-public-ip>` | On-prem VPN device public IP | Network admin |
| `<vpc-cidr>` | VPC CIDR block | Your VPC configuration |
| `<remote-network-cidr>` | Remote network CIDR | Network admin |

## Related Presets

- **01-basic-site-to-site** -- Use when you only need site-to-site VPN without SSL
- **02-production-multi-site** -- Use for multi-site production setups
