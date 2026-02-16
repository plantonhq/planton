---
title: "NAT Gateway"
description: "This preset creates a Scaleway Public Gateway that provides NAT masquerade for a Private Network. Resources in the attached network can reach the internet through the gateway's public IP without..."
type: "preset"
rank: "01"
presetSlug: "01-nat-gateway"
componentSlug: "public-gateway"
componentTitle: "Public Gateway"
provider: "scaleway"
icon: "package"
order: 1
---

# NAT Gateway

This preset creates a Scaleway Public Gateway that provides NAT masquerade for a Private Network. Resources in the attached network can reach the internet through the gateway's public IP without having their own public addresses. This is the most common gateway configuration and is required whenever private resources need outbound internet access.

## When to Use

- Kapsule worker nodes in a Private Network that need to pull container images from public registries
- RDB or Redis instances that need to reach external services for replication or webhooks
- Any Private Network workload requiring outbound internet access without individual public IPs

## Key Configuration Choices

- **Standard tier** (`type: VPC-GW-S`) -- sufficient bandwidth for most workloads; upgrade to `VPC-GW-XL` only for high-throughput requirements in Paris zones
- **NAT masquerade enabled** (`enableMasquerade: true`) -- all outbound traffic from the Private Network is NATed through the gateway's public IP
- **No SSH bastion** -- add the `bastion` block if you also need SSH jump-host access to private instances
- **No port forwarding** -- add `patRules` if specific services need to be reachable from the internet

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| ----------- | ----------- | ------------- |
| `<your-private-network-id>` | UUID of the Private Network to attach the gateway to | Scaleway console or `ScalewayPrivateNetwork` status outputs |

## Related Presets

- **02-bastion-enabled** -- Use instead when you also need SSH bastion access to instances in the Private Network
