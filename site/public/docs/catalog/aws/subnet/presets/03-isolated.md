---
title: "Isolated Subnet"
description: "A subnet with no route rules of its own: it stays on the VPC main route table and has no path to the internet. This is the right placement for data stores and other resources that should only ever be..."
type: "preset"
rank: "03"
presetSlug: "03-isolated"
componentSlug: "subnet"
componentTitle: "Subnet"
provider: "aws"
icon: "package"
order: 3
---

# Isolated Subnet

A subnet with no route rules of its own: it stays on the VPC main route table and has no path to the internet. This is the right placement for data stores and other resources that should only ever be reached from within the VPC.

## When to Use

- Database and cache tiers (RDS, ElastiCache, MemoryDB) that only accept traffic from private application subnets
- Any resource that must have no inbound or outbound internet path
- Simple, single-tier dev/test networks that need no custom routing

## Key Configuration Choices

- **No `routes` and no `routeTableId`** — the subnet uses the VPC main route table. Its `route_table_id` output is empty, signalling "main table". Internal VPC routing (the VPC's local route) still works, so resources in this subnet reach peers in other subnets of the same VPC.
- **No public IP on launch** — `mapPublicIpOnLaunch` is left at its default (`false`).
- **CIDR** (`10.0.2.0/24`) — a third /24, separate from public (`10.0.0.0/24`) and private (`10.0.1.0/24`) tiers.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<vpc-id>` | The VPC to create the subnet in | `AwsVpc` status outputs (`vpc_id`), or reference an `AwsVpc` via `valueFrom` |

## Related Presets

- **01-private** — when the subnet needs outbound internet via a NAT gateway
- **02-public** — when the subnet hosts internet-facing resources
