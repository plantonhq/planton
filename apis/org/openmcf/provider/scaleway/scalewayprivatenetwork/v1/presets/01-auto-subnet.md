# Auto-Subnet Private Network

This preset creates a Scaleway Private Network with IPAM-managed automatic subnet allocation. Scaleway assigns an IPv4 CIDR from its default range, so you do not need to plan address space upfront. This is the fastest way to get a functional Private Network for attaching instances, Kapsule clusters, databases, and load balancers.

## When to Use

- Single-tier environments or quick-start setups where address planning is not a priority
- Development and testing environments
- When you only have one Private Network in the VPC and overlap is not a concern

## Key Configuration Choices

- **Auto-assigned subnet** (no `ipv4Subnet` specified) -- Scaleway IPAM allocates a CIDR automatically; the assigned range is available in `status.outputs.ipv4_subnet_cidr`
- **Paris region** (`region: fr-par`) -- must match the parent VPC's region
- **Default route propagation disabled** (`enableDefaultRoutePropagation: false`) -- enable only when VPC routing is active and this network needs to reach other Private Networks

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| ----------- | ----------- | ------------- |
| `<your-vpc-id>` | UUID of the parent VPC | Scaleway console or `ScalewayVpc` status outputs |

## Related Presets

- **02-explicit-subnet** -- Use instead when you need a specific CIDR for predictable addressing, VPN integration, or multi-network non-overlapping ranges
