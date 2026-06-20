---
title: "Public Subnet"
description: "A public subnet whose default route points at an internet gateway, with launch-time public IP assignment enabled. This is where internet-facing resources live: application load balancers, bastion..."
type: "preset"
rank: "02"
presetSlug: "02-public"
componentSlug: "subnet"
componentTitle: "Subnet"
provider: "aws"
icon: "package"
order: 2
---

# Public Subnet

A public subnet whose default route points at an internet gateway, with launch-time public IP assignment enabled. This is where internet-facing resources live: application load balancers, bastion hosts, and the NAT gateways that private subnets route through.

## When to Use

- Internet-facing load balancers and reverse proxies
- Bastion / jump hosts
- NAT gateways (which must sit in a public subnet to provide outbound access to private subnets)

## Key Configuration Choices

- **Internet gateway default route** (`0.0.0.0/0` via `internet_gateway`) — what makes this subnet "public". The inline `routes` block creates a dedicated route table owned by the subnet.
- **Public IP on launch** (`mapPublicIpOnLaunch: true`) — instances receive a public IPv4 automatically. Combined with the IGW route, they are directly reachable.
- **CIDR** (`10.0.0.0/24`) — the first /24 of a /16 VPC; pair it with private subnets higher in the range.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<vpc-id>` | The VPC to create the subnet in | `AwsVpc` status outputs (`vpc_id`), or reference an `AwsVpc` via `valueFrom` |
| `<internet-gateway-id>` | The internet gateway attached to the VPC | The VPC's internet gateway |

## Related Presets

- **01-private** — for application/worker subnets that should not be publicly reachable
- **03-isolated** — for data-tier subnets with no internet path at all
