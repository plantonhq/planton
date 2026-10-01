# Subnet-Scoped Peering

This preset peers **named subnets only** instead of the networks' complete
address spaces. Set `peer_complete_virtual_networks_enabled` to false and
list the local and remote subnet names included in the peering.

Subnet-scoped peering limits which address ranges are reachable across the
link -- workloads in unlisted subnets on either side cannot reach each
other through this peering. NSGs and route tables still apply per subnet;
this preset only narrows the L3 reachability boundary.

Deploy the reciprocal direction on the remote network with the subnet lists
swapped (`local_subnet_names` on each side name that side's subnets,
`remote_subnet_names` name the far side's). Both directions must agree on
the same subnet pairing for traffic to flow.

For dual-stack networks, `only_ipv6_peering_enabled` restricts the peering
to IPv6 address space only (subnet-scoped). Changing it replaces the
peering.

## When to Use

- Shared-services VNets that should reach only an app subnet, not the whole
  app VNet
- Limiting blast radius between tiers that share a peering for one subnet
- Dual-stack designs that need IPv6-only peering between specific subnets

## Key Configuration Choices

- **`peerCompleteVirtualNetworksEnabled: false`** -- required; subnet name
  lists are rejected when complete-network peering is enabled
- **Subnet names, not prefixes** -- ARM resolves subnets by name within
  each referenced network; ensure the subnets exist before applying
- **Reciprocal direction** -- stamp a second resource on the remote network
  with local/remote networks and subnet lists reversed

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<local-vnet-resource-name>` | Planton metadata name of the network this peering is written on | Your network stack |
| `<remote-vnet-resource-name>` | Planton metadata name of the far network | Your network stack |
| `shared-to-app` (`name`) | Peering name within the local network | Your naming convention |
| `<local-subnet-name>` | Subnet name on the local network to include | `AzureSubnet.spec.name` on the local side |
| `<remote-subnet-name>` | Subnet name on the remote network to include | `AzureSubnet.spec.name` on the remote side |

## Pair With

A second subnet-scoped peering on the remote network with networks and
subnet lists reversed. For full-network reachability instead, use
`01-hub-to-spoke` and `02-spoke-to-hub` with default
`peerCompleteVirtualNetworksEnabled` (true).
