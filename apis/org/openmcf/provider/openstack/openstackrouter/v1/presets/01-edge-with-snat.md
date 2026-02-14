# Edge Router with SNAT

This preset creates a router with an external gateway and Source NAT enabled. Tenant instances on connected subnets can reach the internet through this router without needing individual floating IPs. This is the standard production router configuration for most OpenStack deployments.

## When to Use

- Any tenant that needs outbound internet access from private subnets
- Standard application deployments where instances pull packages, reach APIs, or send notifications
- Base router before attaching subnets via OpenStackRouterInterface

## Key Configuration Choices

- **External gateway** (`externalNetworkId`) -- connects the router to the provider network for internet access
- **SNAT enabled** (`enableSnat: true`) -- tenant traffic is NATed to the router's external IP for outbound connectivity
- **Admin state up** -- default (true), router forwards traffic immediately

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| ----------- | ----------- | ------------- |
| `<external-network-id>` | ID of the external (provider) network | OpenStack admin or `OpenStackNetwork` (external) status outputs |

## Related Presets

- **02-internal-only** -- Use when subnets need inter-subnet routing but no external connectivity
