# Standard Tenant Network

This preset creates a basic tenant network with all OpenStack defaults. The network gets port security enabled (via deployment default), DHCP-ready behavior, and standard MTU. This is the starting point for virtually all OpenStack workloads.

## When to Use

- Any workload that needs its own isolated Layer 2 network
- Standard application deployments using private networking
- Base network before attaching subnets, routers, and ports

## Key Configuration Choices

- **Defaults only** -- relies on OpenStack deployment defaults for admin state (up), MTU, and port security
- **Not shared** -- visible only to the owning project (multi-tenant isolation)
- **Not external** -- this is a tenant network, not a provider/floating-IP network
- **No DNS domain** -- DNS integration can be added later if the Neutron deployment supports it

## Placeholders to Replace

No placeholders -- this preset is deployable as-is after setting `metadata.name`.
