# AzureVirtualNetworkPeering Pulumi Module

## Overview

This Pulumi module provisions one direction of an Azure virtual network
peering using the Azure Classic provider (`pulumi-azure`). It creates a
single `network.VirtualNetworkPeering` -- private connectivity between two
networks over the Microsoft backbone, no gateways or public IPs involved.

One resource is ONE DIRECTION, exactly as ARM models it; traffic only
flows once the reciprocal peering exists on the remote network (Azure
retries internally while the far side catches up, so both directions can
deploy concurrently). The peering is an ARM child of its LOCAL network:
the module derives the resource group and network name from the local
network's ARM ID rather than asking for them separately.

The connectivity flags and subnet-name lists update in place. Name, the
two networks, and the complete-vs-subnet-scoped and IPv6-only choices are
the peering's identity -- changing any of them replaces it (a brief
connectivity gap for this direction only). Peerings are not tracked ARM
resources, so they carry no tags.

## Resources Created

- `network.VirtualNetworkPeering` -- one direction of the peering

## Inputs

The module receives an `AzureVirtualNetworkPeeringIacInput` containing:

- `target.spec.name` -- the peering's name, unique within the local network
- `target.spec.virtual_network_id` -- ARM ID of the LOCAL network (reference resolved to a literal by the platform); resource group and network name are parsed from it
- `target.spec.remote_virtual_network_id` -- ARM ID of the REMOTE network; cross-subscription and global (cross-region) peering work unchanged
- `target.spec.allow_virtual_network_access` / `allow_forwarded_traffic` / `allow_gateway_transit` / `use_remote_gateways` -- the four connectivity flags; defaults mirror Azure's (access on, the rest off)
- `target.spec.peer_complete_virtual_networks_enabled` -- Azure defaults to true; false enables subnet-scoped peering via `local_subnet_names` / `remote_subnet_names`
- `target.spec.only_ipv6_peering_enabled` -- peer only the IPv6 address space (dual-stack, subnet-scoped)
- `provider_config` -- Azure credentials (static client secret, keyless web identity, or ambient chain)

## Outputs

| Output | Description |
|--------|-------------|
| `peering_id` | Full ARM ID of the peering |
| `peering_name` | The peering's name within its local network |
| `virtual_network_name` | Local network name, derived from its ARM ID |
| `resource_group_name` | Resource group of the local network, derived from its ARM ID |

## Local Development

```bash
make build       # Build the module
make deps        # Download and tidy dependencies
make update-deps # Update to latest planton
```
