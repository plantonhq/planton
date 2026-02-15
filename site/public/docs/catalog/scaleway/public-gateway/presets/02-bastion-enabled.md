---
title: "Bastion-Enabled Gateway"
description: "This preset creates a Scaleway Public Gateway with both NAT masquerade and an SSH bastion. In addition to providing outbound internet access for the Private Network, the gateway acts as a secure SSH..."
type: "preset"
rank: "02"
presetSlug: "02-bastion-enabled"
componentSlug: "public-gateway"
componentTitle: "Public Gateway"
provider: "scaleway"
icon: "package"
order: 2
---

# Bastion-Enabled Gateway

This preset creates a Scaleway Public Gateway with both NAT masquerade and an SSH bastion. In addition to providing outbound internet access for the Private Network, the gateway acts as a secure SSH jump host, allowing operators to reach private instances without assigning public IPs to them.

## When to Use

- Production environments where operators need SSH access to instances, Kapsule nodes, or other resources on a Private Network
- Security-conscious setups where direct public SSH access to individual machines is prohibited
- Environments that require an auditable single entry point for SSH connections

## Key Configuration Choices

- **Standard tier** (`type: VPC-GW-S`) -- sufficient for combined NAT and bastion workloads
- **NAT masquerade enabled** (`enableMasquerade: true`) -- outbound internet for all Private Network resources
- **SSH bastion on port 22** (`bastion.port: 22`) -- standard SSH port; change only if blocked by corporate firewalls
- **IP allowlist** (`bastion.allowedIpRanges`) -- restricts bastion access to specific source CIDRs; leaving this empty allows all IPs, which is not recommended for production

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| ----------- | ----------- | ------------- |
| `<your-private-network-id>` | UUID of the Private Network to attach the gateway to | Scaleway console or `ScalewayPrivateNetwork` status outputs |
| `<your-office-cidr>` | CIDR range of your office or VPN (e.g., `203.0.113.0/24`) | Your network administrator or VPN provider |

## Related Presets

- **01-nat-gateway** -- Use instead when only outbound internet access is needed and SSH bastion is not required
