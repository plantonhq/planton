# Routing-Enabled VPC

This preset creates a Scaleway VPC with inter-Private-Network routing enabled. When routing is on, resources in different Private Networks attached to this VPC can communicate with each other. This is the standard production configuration for multi-tier architectures.

## When to Use

- Multi-tier architectures where a Kapsule cluster in one Private Network needs to reach an RDB instance in another
- Environments with separate Private Networks for application, database, and cache tiers
- Any setup requiring network segmentation with controlled cross-network traffic

## Key Configuration Choices

- **Paris region** (`region: fr-par`) -- Scaleway's primary region with the broadest service availability
- **Routing enabled** (`enableRouting: true`) -- allows resources in different Private Networks within this VPC to communicate; required for multi-tier architectures
- **Custom routes propagation disabled** (`enableCustomRoutesPropagation: false`) -- enable only if using VPN gateways or network appliances that need to advertise routes across Private Networks

## Placeholders to Replace

No placeholders -- this preset is ready to deploy as-is. Change `region` to `nl-ams` or `pl-waw` for Amsterdam or Warsaw.

## Related Presets

- **01-standard** -- Use instead when all resources reside in a single Private Network and cross-network routing is unnecessary
