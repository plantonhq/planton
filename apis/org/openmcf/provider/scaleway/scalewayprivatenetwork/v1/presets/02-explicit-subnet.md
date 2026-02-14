# Explicit-Subnet Private Network

This preset creates a Scaleway Private Network with a user-defined IPv4 CIDR block. Specifying the subnet gives you full control over address space, which is essential when multiple Private Networks in the same VPC must have non-overlapping ranges for routing to work correctly.

## When to Use

- Multi-tier architectures with separate app, database, and cache Private Networks that need routing between them
- VPN or hybrid-cloud integration where address ranges must not overlap with on-premises networks
- Production environments requiring predictable and documented IP addressing

## Key Configuration Choices

- **Explicit CIDR** (`ipv4Subnet: 10.0.1.0/24`) -- a /24 provides 254 usable addresses; adjust the CIDR to match your capacity needs and avoid overlap with other Private Networks in the same VPC
- **Paris region** (`region: fr-par`) -- must match the parent VPC's region
- **Default route propagation disabled** (`enableDefaultRoutePropagation: false`) -- enable when VPC routing is active and this network needs to reach other Private Networks

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| ----------- | ----------- | ------------- |
| `<your-vpc-id>` | UUID of the parent VPC | Scaleway console or `ScalewayVpc` status outputs |

## Related Presets

- **01-auto-subnet** -- Use instead when address planning is not needed and you want Scaleway to assign the CIDR automatically
